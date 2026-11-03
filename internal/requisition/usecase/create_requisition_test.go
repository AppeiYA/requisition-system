package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type createMembership struct {
	membership *departmentdomain.DepartmentMembership
}

func (m createMembership) GetMembershipByUserID(context.Context, string) (*departmentdomain.DepartmentMembership, error) {
	return m.membership, nil
}

type createCategoryRepo struct{}

func (createCategoryRepo) SaveItemCategory(context.Context, *requisitiondomain.ItemCategory) error {
	return nil
}
func (createCategoryRepo) GetItemCategoryByID(context.Context, string) (*requisitiondomain.ItemCategory, error) {
	name, _ := requisitiondomain.NewCategoryName("Office")
	description, _ := requisitiondomain.NewCategoryDescription("Office supplies")
	return requisitiondomain.NewItemCategory("cat-123", name, description)
}
func (createCategoryRepo) ListItemCategories(context.Context, int, int) ([]*requisitiondomain.ItemCategory, error) {
	return nil, nil
}
func (createCategoryRepo) ListItemsInCategory(context.Context, string) ([]*requisitiondomain.Item, error) {
	return nil, nil
}
func (createCategoryRepo) UpdateItemCategory(context.Context, *requisitiondomain.ItemCategory) error {
	return nil
}

type createItemRepo struct{}

func (createItemRepo) SaveItem(context.Context, *requisitiondomain.Item) error { return nil }
func (createItemRepo) GetItemByID(context.Context, string) (*requisitiondomain.Item, error) {
	name, _ := requisitiondomain.NewItemName("Laptop")
	description, _ := requisitiondomain.NewItemDescription("Business laptop")
	return requisitiondomain.NewItem("itm-123", name, description, "cat-123", 1500)
}
func (createItemRepo) ListItems(context.Context, int, int) ([]*requisitiondomain.Item, error) {
	return nil, nil
}
func (createItemRepo) UpdateItem(context.Context, *requisitiondomain.Item) error { return nil }

type createRequisitionRepo struct {
	saved *requisitiondomain.Requisition
	get   *requisitiondomain.Requisition
	err   error
}

func (r *createRequisitionRepo) SaveRequisition(_ context.Context, requisition *requisitiondomain.Requisition) error {
	r.saved = requisition
	return r.err
}
func (r *createRequisitionRepo) GetRequisitionByID(context.Context, string) (*requisitiondomain.Requisition, error) {
	return r.get, r.err
}
func (r *createRequisitionRepo) GetRequisitionByIDForUpdate(context.Context, string) (*requisitiondomain.Requisition, error) {
	return r.get, r.err
}
func (*createRequisitionRepo) UpdateRequisition(context.Context, *requisitiondomain.Requisition) error {
	return nil
}
func (*createRequisitionRepo) UpdateRequisitionAggregate(context.Context, *requisitiondomain.Requisition) error {
	return nil
}
func (*createRequisitionRepo) ListRequisitionsByUserID(context.Context, string, int, int) ([]*requisitiondomain.Requisition, error) {
	return nil, nil
}
func (*createRequisitionRepo) ListRequisitions(context.Context, int, int) ([]*requisitiondomain.Requisition, error) {
	return nil, nil
}

type createTransactionManager struct {
	called bool
	err    error
}

func (m *createTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	m.called = true
	if m.err != nil {
		return m.err
	}
	return fn(ctx)
}

type createIDGenerator struct {
	next int
}

func (g *createIDGenerator) GenerateID(prefix string) (string, error) {
	g.next++
	return prefix + "-" + string(rune('a'+g.next)), nil
}

var (
	_ requisitionout.MembershipLookup       = createMembership{}
	_ requisitionout.ItemCategoryRepository = createCategoryRepo{}
	_ requisitionout.ItemRepository         = createItemRepo{}
)

func validCreateMembership(t *testing.T) *departmentdomain.DepartmentMembership {
	t.Helper()
	jobTitle, err := departmentdomain.NewMemberJobTitle("Staff")
	if err != nil {
		t.Fatal(err)
	}
	membership, err := departmentdomain.NewDepartmentMembership("usr-123", "dpt-123", jobTitle)
	if err != nil {
		t.Fatal(err)
	}
	return membership
}

func TestCreateRequisitionUsecasePersistsAggregateTransactionally(t *testing.T) {
	repository := &createRequisitionRepo{}
	transaction := &createTransactionManager{}
	usecase := NewCreateRequisitionUsecase(
		repository, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: validCreateMembership(t)},
		&createIDGenerator{}, transaction,
	)

	requisition, err := usecase.Execute(context.Background(), requisitionin.CreateRequisitionCommand{
		UserID:       "usr-123",
		DepartmentID: "dpt-123",
		Lines: []requisitionin.CreateRequisitionLine{{
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Pens", CategoryID: "cat-123", Description: "Blue pens",
			},
			Quantity:            3,
			EstimatedUnitAmount: 2500,
		}},
		Justification: "Office supplies",
		Priority:      "medium",
		NeededBy:      time.Now().AddDate(0, 0, 2),
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if !transaction.called {
		t.Fatal("expected creation to execute inside a transaction")
	}
	if repository.saved == nil {
		t.Fatal("expected requisition to be saved")
	}
	if len(repository.saved.Items()) != 1 || len(repository.saved.RequestedItems()) != 1 {
		t.Fatalf("saved aggregate has %d lines and %d requested items", len(repository.saved.Items()), len(repository.saved.RequestedItems()))
	}
	if got := repository.saved.Items()[0].EstimatedUnitAmount().Amount(); got != 250000 {
		t.Fatalf("saved estimated unit amount = %d, want 250000 minor units", got)
	}
	if len(repository.saved.StatusHistory()) != 1 {
		t.Fatal("expected initial status history")
	}
	if repository.saved.StatusHistory()[0].ToStatus() != requisitiondomain.RequisitionStatusPending {
		t.Fatal("expected initial status to be pending")
	}
	if total, err := requisition.TotalEstimatedAmount(); err != nil || total == nil {
		t.Fatal("expected derived total")
	}
}

func TestCreateRequisitionUsecaseBuildsCatalogueAndRequestedLines(t *testing.T) {
	repository := &createRequisitionRepo{}
	usecase := NewCreateRequisitionUsecase(
		repository, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: validCreateMembership(t)},
		&createIDGenerator{}, &createTransactionManager{},
	)

	requisition, err := usecase.Execute(context.Background(), requisitionin.CreateRequisitionCommand{
		UserID:       "usr-123",
		DepartmentID: "dpt-123",
		Lines: []requisitionin.CreateRequisitionLine{
			{ItemID: "itm-123", Quantity: 2, EstimatedUnitAmount: 150000},
			{RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Laptop stand", CategoryID: "cat-123", Description: "Adjustable laptop stand",
			}, Quantity: 1, EstimatedUnitAmount: 850000},
		},
		Justification: "Office setup",
		Priority:      "urgent",
		NeededBy:      time.Now().AddDate(0, 0, 2),
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(requisition.Items()) != 2 {
		t.Fatalf("Items() length = %d, want 2", len(requisition.Items()))
	}
	if requisition.Items()[0].ItemID() != "itm-123" || requisition.Items()[0].RequestedItemID() != "" {
		t.Fatal("first line should reference the catalogue item")
	}
	if requisition.Items()[1].RequestedItemID() == "" || requisition.Items()[1].ItemID() != "" {
		t.Fatal("second line should reference a requested item")
	}
	if got := requisition.Items()[1].ItemNameSnapshot().Value(); got != "Laptop stand" {
		t.Fatalf("item name snapshot = %q, want Laptop stand", got)
	}
}

func TestCreateRequisitionUsecaseRollsBackOnRepositoryError(t *testing.T) {
	repository := &createRequisitionRepo{err: errors.New("save failed")}
	transaction := &createTransactionManager{}
	usecase := NewCreateRequisitionUsecase(
		repository, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: validCreateMembership(t)},
		&createIDGenerator{}, transaction,
	)

	_, err := usecase.Execute(context.Background(), requisitionin.CreateRequisitionCommand{
		UserID: "usr-123",
		Lines: []requisitionin.CreateRequisitionLine{{
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Pens", CategoryID: "cat-123", Description: "Blue pens",
			},
			Quantity: 1, EstimatedUnitAmount: 2500,
		}},
		Justification: "Office supplies",
		Priority:      "normal",
		NeededBy:      time.Now().AddDate(0, 0, 2),
	})
	if !errors.Is(err, repository.err) {
		t.Fatalf("Execute() error = %v, want %v", err, repository.err)
	}
	if !transaction.called {
		t.Fatal("expected transaction manager to be called")
	}
}

func TestCreateRequisitionUsecaseRejectsInactiveMembership(t *testing.T) {
	membership, err := departmentdomain.ReconstituteDepartmentMembership(
		"usr-123", "dpt-123", "Staff",
		departmentdomain.MembershipStatusInactive.String(), time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	usecase := NewCreateRequisitionUsecase(
		&createRequisitionRepo{}, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: membership}, &createIDGenerator{},
		&createTransactionManager{},
	)

	_, err = usecase.Execute(context.Background(), requisitionin.CreateRequisitionCommand{
		UserID: "usr-123",
		Lines: []requisitionin.CreateRequisitionLine{{
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Pens", CategoryID: "cat-123", Description: "Blue pens",
			},
			Quantity: 1, EstimatedUnitAmount: 2500,
		}},
		Justification: "Office supplies", NeededBy: time.Now().AddDate(0, 0, 2),
	})
	if err != ErrInactiveMembership {
		t.Fatalf("Execute() error = %v, want %v", err, ErrInactiveMembership)
	}
}

func TestCreateRequisitionUsecaseRejectsDepartmentMismatch(t *testing.T) {
	usecase := NewCreateRequisitionUsecase(
		&createRequisitionRepo{}, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: validCreateMembership(t)},
		&createIDGenerator{}, &createTransactionManager{},
	)

	_, err := usecase.Execute(context.Background(), requisitionin.CreateRequisitionCommand{
		UserID:       "usr-123",
		DepartmentID: "dpt-other",
		Lines: []requisitionin.CreateRequisitionLine{{
			ItemID: "itm-123", Quantity: 1, EstimatedUnitAmount: 150000,
		}},
		Justification: "Office supplies", NeededBy: time.Now().AddDate(0, 0, 2),
	})
	if err != ErrDepartmentMismatch {
		t.Fatalf("Execute() error = %v, want %v", err, ErrDepartmentMismatch)
	}
}
