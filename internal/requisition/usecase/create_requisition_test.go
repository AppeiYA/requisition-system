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
	saved          *requisitiondomain.Requisition
	get            *requisitiondomain.Requisition
	summaries      []*requisitiondomain.RequisitionSummary
	userSummaries  []*requisitiondomain.RequisitionSummary
	summaryUserID  string
	summaryLimit   int
	summaryOffset  int
	listCalled     bool
	userListCalled bool
	err            error
}

func (r *createRequisitionRepo) SaveRequisition(_ context.Context, requisition *requisitiondomain.Requisition) error {
	r.saved = requisition
	return r.err
}
func (r *createRequisitionRepo) GetRequisitionByID(context.Context, string) (*requisitiondomain.Requisition, error) {
	return r.get, r.err
}
func (r *createRequisitionRepo) GetRequisitionIDByItemID(context.Context, string) (string, error) {
	return "", nil
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
func (r *createRequisitionRepo) ListRequisitionSummaries(_ context.Context, limit int, offset int) ([]*requisitiondomain.RequisitionSummary, error) {
	r.listCalled = true
	r.summaryLimit = limit
	r.summaryOffset = offset
	return r.summaries, r.err
}
func (r *createRequisitionRepo) ListRequisitionSummariesByUserID(_ context.Context, userID string, limit int, offset int) ([]*requisitiondomain.RequisitionSummary, error) {
	r.userListCalled = true
	r.summaryUserID = userID
	r.summaryLimit = limit
	r.summaryOffset = offset
	return r.userSummaries, r.err
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
		&createIDGenerator{}, transaction, nil,
	)

	requisition, err := usecase.Execute(context.Background(), requisitionin.CreateRequisitionCommand{
		UserID: "usr-123",
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
		&createIDGenerator{}, &createTransactionManager{}, nil,
	)

	requisition, err := usecase.Execute(context.Background(), requisitionin.CreateRequisitionCommand{
		UserID: "usr-123",
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
		&createIDGenerator{}, transaction, nil,
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
		Priority:      "medium",
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
		&createTransactionManager{}, nil,
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

type createRegistrar struct {
	targets []requisitionout.StagedAttachmentTarget
	err     error
}

func (r *createRegistrar) RegisterStagedAttachments(
	_ context.Context,
	targets []requisitionout.StagedAttachmentTarget,
) error {
	r.targets = append(r.targets, targets...)
	return r.err
}

func createCommandWithUploads() requisitionin.CreateRequisitionCommand {
	return requisitionin.CreateRequisitionCommand{
		UserID: "usr-123",
		Lines: []requisitionin.CreateRequisitionLine{{
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Pens", CategoryID: "cat-123", Description: "Blue pens",
			},
			Quantity: 1, EstimatedUnitAmount: 2500,
			UploadIDs: []string{"upl-line"},
		}},
		Justification: "Office supplies",
		Priority:      "medium",
		NeededBy:      time.Now().AddDate(0, 0, 2),
		UploadIDs:     []string{"upl-header"},
	}
}

func TestCreateRequisitionRegistersStagedAttachmentsAfterIDsExist(t *testing.T) {
	registrar := &createRegistrar{}
	usecase := NewCreateRequisitionUsecase(
		&createRequisitionRepo{}, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: validCreateMembership(t)},
		&createIDGenerator{}, &createTransactionManager{}, registrar,
	)

	created, err := usecase.Execute(context.Background(), createCommandWithUploads())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(registrar.targets) != 2 {
		t.Fatalf("registered %d targets, want 2", len(registrar.targets))
	}
	if registrar.targets[0].RefName != "requisition" || registrar.targets[0].RefID != created.ID() {
		t.Errorf("header target = %+v", registrar.targets[0])
	}
	if registrar.targets[1].RefName != "requisition_item" ||
		registrar.targets[1].RefID != created.Items()[0].ID() {
		t.Errorf("line target = %+v", registrar.targets[1])
	}
	for _, target := range registrar.targets {
		if target.ActorID != "usr-123" {
			t.Errorf("target actor = %q", target.ActorID)
		}
	}
}

func TestCreateRequisitionRollsBackWhenClaimFails(t *testing.T) {
	registrar := &createRegistrar{err: errors.New("claim failed")}
	usecase := NewCreateRequisitionUsecase(
		&createRequisitionRepo{}, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: validCreateMembership(t)},
		&createIDGenerator{}, &createTransactionManager{}, registrar,
	)

	_, err := usecase.Execute(context.Background(), createCommandWithUploads())
	if !errors.Is(err, registrar.err) {
		t.Fatalf("Execute() error = %v, want %v", err, registrar.err)
	}
}

func TestCreateRequisitionRejectsInvalidStagedUploadIDs(t *testing.T) {
	for _, uploads := range [][]string{{"req-1234"}, {"upl-1234", "upl-1234"}} {
		cmd := createCommandWithUploads()
		cmd.UploadIDs = uploads
		registrar := &createRegistrar{}
		usecase := NewCreateRequisitionUsecase(
			&createRequisitionRepo{}, createCategoryRepo{}, createItemRepo{},
			createMembership{membership: validCreateMembership(t)},
			&createIDGenerator{}, &createTransactionManager{}, registrar,
		)
		if _, err := usecase.Execute(context.Background(), cmd); err != ErrInvalidStagedUpload {
			t.Fatalf("Execute() error = %v, want %v", err, ErrInvalidStagedUpload)
		}
		if len(registrar.targets) != 0 {
			t.Errorf("registrar was called with %+v", registrar.targets)
		}
	}
}
