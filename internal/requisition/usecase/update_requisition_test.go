package usecase

import (
	"context"
	"testing"
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type updateRequisitionRepo struct {
	current     *requisitiondomain.Requisition
	updated     *requisitiondomain.Requisition
	getErr      error
	updateErr   error
	updateCalls int
}

func (r *updateRequisitionRepo) GetRequisitionByIDForUpdate(context.Context, string) (*requisitiondomain.Requisition, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.current, nil
}

func (r *updateRequisitionRepo) UpdateRequisitionAggregate(_ context.Context, requisition *requisitiondomain.Requisition) error {
	r.updateCalls++
	if r.updateErr != nil {
		return r.updateErr
	}
	r.updated = requisition
	r.current = requisition
	return nil
}

var _ requisitionout.RequisitionEditRepository = (*updateRequisitionRepo)(nil)

func updateTestItem(t *testing.T, id, requisitionID string, lineNumber int, itemID, requestedItemID string, name string, quantity int, amount int64) *requisitiondomain.RequisitionItem {
	t.Helper()

	itemName, err := requisitiondomain.NewItemName(name)
	if err != nil {
		t.Fatal(err)
	}
	validQuantity, err := requisitiondomain.NewRequisitionQuantity(quantity)
	if err != nil {
		t.Fatal(err)
	}
	money, err := shareddomain.NewNairaMoneyFromMinorUnits(amount)
	if err != nil {
		t.Fatal(err)
	}
	item, err := requisitiondomain.NewRequisitionItem(
		id, requisitionID, lineNumber, itemID, requestedItemID, itemName,
		validQuantity, money, "",
	)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func updateTestRequestedItem(t *testing.T, id, name, description, reviewStatus, promotedItemID string, createdAt time.Time) *requisitiondomain.RequestedItem {
	t.Helper()

	item, err := requisitiondomain.ReconstituteRequestedItem(
		id, name, "cat-123", description, "usr-123", reviewStatus,
		promotedItemID, createdAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func updateTestRequisition(t *testing.T, requestedItem *requisitiondomain.RequestedItem) *requisitiondomain.Requisition {
	t.Helper()

	items := []*requisitiondomain.RequisitionItem{
		updateTestItem(t, "reqitem-catalogue", "req-123", 1, "itm-123", "", "Laptop", 2, 150000),
	}
	requestedItems := []*requisitiondomain.RequestedItem(nil)
	if requestedItem != nil {
		items = append(items, updateTestItem(t, "reqitem-requested", "req-123", 2, "", requestedItem.ID(), requestedItem.Name().Value(), 1, 850000))
		requestedItems = []*requisitiondomain.RequestedItem{requestedItem}
	}

	justification, err := requisitiondomain.NewRequisitionJustification("Office setup")
	if err != nil {
		t.Fatal(err)
	}
	priority, err := requisitiondomain.NewRequisitionPriority("medium")
	if err != nil {
		t.Fatal(err)
	}
	requisition, err := requisitiondomain.NewSubmittedRequisitionWithRequestedItems(
		"req-123", "REQ-2026-10001", "usr-123", "dpt-123",
		priority, time.Now().UTC().AddDate(0, 0, 5), justification,
		items, requestedItems,
	)
	if err != nil {
		t.Fatal(err)
	}
	return requisition
}

func newUpdateUsecase(t *testing.T, repository *updateRequisitionRepo) *UpdateRequisitionUsecase {
	t.Helper()

	return NewUpdateRequisitionUsecase(
		repository, createCategoryRepo{}, createItemRepo{},
		createMembership{membership: validCreateMembership(t)},
		&createIDGenerator{}, &createTransactionManager{},
	)
}

func updateTestCommand(lines ...requisitionin.UpdateRequisitionLine) requisitionin.UpdateRequisitionCommand {
	return requisitionin.UpdateRequisitionCommand{
		RequisitionID: "req-123",
		UserID:        "usr-123",
		Lines:         lines,
		Justification: "Updated justification",
		Priority:      "high",
		NeededBy:      time.Now().UTC().AddDate(0, 0, 7),
	}
}

func TestUpdateRequisitionPreservesIdentityWhenLinesAreUnchanged(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Adjustable stand", "unreviewed", "", createdAt)
	repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
	usecase := newUpdateUsecase(t, repository)

	updated, err := usecase.Execute(context.Background(), updateTestCommand(
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
			Quantity: 2, EstimatedUnitAmount: 150000,
		},
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-requested",
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Laptop stand", CategoryID: "cat-123", Description: "Adjustable stand",
			},
			Quantity: 1, EstimatedUnitAmount: 850000,
		},
	))
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if repository.updateCalls != 1 {
		t.Fatalf("aggregate update calls = %d, want 1", repository.updateCalls)
	}

	items := updated.Items()
	if len(items) != 2 || items[0].ID() != "reqitem-catalogue" || items[1].ID() != "reqitem-requested" {
		t.Fatalf("line identities = %+v, want both original IDs", items)
	}
	requestedItems := updated.RequestedItems()
	if len(requestedItems) != 1 || requestedItems[0].ID() != "ritem-requested" {
		t.Fatalf("requested item identities = %+v, want original ID", requestedItems)
	}
	if !requestedItems[0].CreatedAt().Equal(createdAt) {
		t.Fatalf("requested item createdAt = %v, want %v", requestedItems[0].CreatedAt(), createdAt)
	}
	if updated.Priority() != requisitiondomain.PriorityHigh || updated.Justification().Value() != "Updated justification" {
		t.Fatalf("header was not updated: priority=%s justification=%s", updated.Priority(), updated.Justification().Value())
	}
}

func TestUpdateRequisitionUpdatesRequestedItemDetailsWithoutResettingWorkflow(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 9, 15, 8, 30, 0, 0, time.UTC)
	requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Old description", "promoted", "itm-999", createdAt)
	repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
	usecase := newUpdateUsecase(t, repository)

	updated, err := usecase.Execute(context.Background(), updateTestCommand(
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
			Quantity: 2, EstimatedUnitAmount: 150000,
		},
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-requested",
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Laptop stand", CategoryID: "cat-123", Description: "New description",
			},
			Quantity: 5, EstimatedUnitAmount: 900000,
		},
	))
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}

	requestedItems := updated.RequestedItems()
	if len(requestedItems) != 1 {
		t.Fatalf("requested items = %d, want 1", len(requestedItems))
	}
	got := requestedItems[0]
	if got.ID() != "ritem-requested" || got.Description().Value() != "New description" {
		t.Fatalf("requested item = %+v, want preserved ID with updated description", got)
	}
	if got.ReviewStatus() != requisitiondomain.RequestedItemPromoted || got.PromotedItemID() != "itm-999" {
		t.Fatalf("workflow fields were reset: status=%s promoted=%s", got.ReviewStatus(), got.PromotedItemID())
	}
	if !got.CreatedAt().Equal(createdAt) {
		t.Fatalf("createdAt = %v, want %v", got.CreatedAt(), createdAt)
	}
	if updated.Items()[1].ID() != "reqitem-requested" || updated.Items()[1].Quantity().Value() != 5 {
		t.Fatalf("requested line = %+v, want preserved ID and updated quantity", updated.Items()[1])
	}
}

func TestUpdateRequisitionAddsAndRemovesLines(t *testing.T) {
	t.Parallel()

	requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Adjustable stand", "unreviewed", "", time.Now().UTC())
	repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
	usecase := newUpdateUsecase(t, repository)

	// Remove the requested line and add a new requested line.
	updated, err := usecase.Execute(context.Background(), updateTestCommand(
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
			Quantity: 2, EstimatedUnitAmount: 150000,
		},
		requisitionin.UpdateRequisitionLine{
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Monitor", CategoryID: "cat-123", Description: "27 inch monitor",
			},
			Quantity: 1, EstimatedUnitAmount: 200000,
		},
	))
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(updated.Items()) != 2 {
		t.Fatalf("lines = %d, want 2", len(updated.Items()))
	}
	if updated.Items()[0].ID() != "reqitem-catalogue" {
		t.Fatalf("existing line ID = %q, want reqitem-catalogue", updated.Items()[0].ID())
	}
	if updated.Items()[1].ID() == "reqitem-requested" || updated.Items()[1].ID() == "" {
		t.Fatalf("new line ID = %q, want a fresh ID", updated.Items()[1].ID())
	}
	requestedItems := updated.RequestedItems()
	if len(requestedItems) != 1 || requestedItems[0].ID() == "ritem-requested" {
		t.Fatalf("requested items = %+v, want only the new requested item", requestedItems)
	}
}

func TestUpdateRequisitionTransitionsBetweenCatalogueAndRequested(t *testing.T) {
	t.Parallel()

	t.Run("requested to catalogue", func(t *testing.T) {
		t.Parallel()

		requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Adjustable stand", "unreviewed", "", time.Now().UTC())
		repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
		usecase := newUpdateUsecase(t, repository)

		updated, err := usecase.Execute(context.Background(), updateTestCommand(
			requisitionin.UpdateRequisitionLine{
				RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
				Quantity: 2, EstimatedUnitAmount: 150000,
			},
			requisitionin.UpdateRequisitionLine{
				RequisitionItemID: "reqitem-requested", ItemID: "itm-456",
				Quantity: 1, EstimatedUnitAmount: 850000,
			},
		))
		if err != nil {
			t.Fatalf("Execute() unexpected error = %v", err)
		}
		if updated.Items()[1].ID() != "reqitem-requested" || updated.Items()[1].RequestedItemID() != "" {
			t.Fatalf("transitioned line = %+v, want preserved ID and no requested item", updated.Items()[1])
		}
		if len(updated.RequestedItems()) != 0 {
			t.Fatalf("requested items = %+v, want none", updated.RequestedItems())
		}
	})

	t.Run("catalogue to requested", func(t *testing.T) {
		t.Parallel()

		repository := &updateRequisitionRepo{current: updateTestRequisition(t, nil)}
		usecase := newUpdateUsecase(t, repository)

		updated, err := usecase.Execute(context.Background(), updateTestCommand(
			requisitionin.UpdateRequisitionLine{
				RequisitionItemID: "reqitem-catalogue",
				RequestedItem: &requisitionin.CreateRequestedItem{
					Name: "Laptop stand", CategoryID: "cat-123", Description: "Adjustable stand",
				},
				Quantity: 1, EstimatedUnitAmount: 850000,
			},
		))
		if err != nil {
			t.Fatalf("Execute() unexpected error = %v", err)
		}
		if updated.Items()[0].ID() != "reqitem-catalogue" || updated.Items()[0].RequestedItemID() == "" {
			t.Fatalf("transitioned line = %+v, want preserved ID and a requested item", updated.Items()[0])
		}
		if len(updated.RequestedItems()) != 1 {
			t.Fatalf("requested items = %d, want 1", len(updated.RequestedItems()))
		}
	})
}

func TestUpdateRequisitionRejectsLineFromAnotherRequisition(t *testing.T) {
	t.Parallel()

	requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Adjustable stand", "unreviewed", "", time.Now().UTC())
	repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
	usecase := newUpdateUsecase(t, repository)

	_, err := usecase.Execute(context.Background(), updateTestCommand(
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-other", ItemID: "itm-123",
			Quantity: 1, EstimatedUnitAmount: 1,
		},
	))
	if err != requisitiondomain.ErrRequisitionItemNotOwned {
		t.Fatalf("error = %v, want %v", err, requisitiondomain.ErrRequisitionItemNotOwned)
	}
	if repository.updateCalls != 0 {
		t.Fatal("aggregate should not be updated when line ownership is invalid")
	}
}

func TestUpdateRequisitionRejectsDuplicateLineIDs(t *testing.T) {
	t.Parallel()

	requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Adjustable stand", "unreviewed", "", time.Now().UTC())
	repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
	usecase := newUpdateUsecase(t, repository)

	_, err := usecase.Execute(context.Background(), updateTestCommand(
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
			Quantity: 1, EstimatedUnitAmount: 1,
		},
		requisitionin.UpdateRequisitionLine{
			RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
			Quantity: 2, EstimatedUnitAmount: 1,
		},
	))
	if err != requisitiondomain.ErrDuplicateRequisitionItemID {
		t.Fatalf("error = %v, want %v", err, requisitiondomain.ErrDuplicateRequisitionItemID)
	}
}

func TestUpdateRequisitionRepeatedUpdatesKeepIdentity(t *testing.T) {
	t.Parallel()

	requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Adjustable stand", "unreviewed", "", time.Now().UTC())
	repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
	usecase := newUpdateUsecase(t, repository)

	lines := []requisitionin.UpdateRequisitionLine{
		{
			RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
			Quantity: 2, EstimatedUnitAmount: 150000,
		},
		{
			RequisitionItemID: "reqitem-requested",
			RequestedItem: &requisitionin.CreateRequestedItem{
				Name: "Laptop stand", CategoryID: "cat-123", Description: "Adjustable stand",
			},
			Quantity: 1, EstimatedUnitAmount: 850000,
		},
	}

	first, err := usecase.Execute(context.Background(), updateTestCommand(lines...))
	if err != nil {
		t.Fatalf("first Execute() unexpected error = %v", err)
	}
	second, err := usecase.Execute(context.Background(), updateTestCommand(lines...))
	if err != nil {
		t.Fatalf("second Execute() unexpected error = %v", err)
	}

	if first.Items()[0].ID() != second.Items()[0].ID() || first.Items()[1].ID() != second.Items()[1].ID() {
		t.Fatalf("line identities changed: first=%+v second=%+v", first.Items(), second.Items())
	}
	if first.RequestedItems()[0].ID() != second.RequestedItems()[0].ID() {
		t.Fatalf("requested item identity changed: first=%s second=%s", first.RequestedItems()[0].ID(), second.RequestedItems()[0].ID())
	}
}

func TestUpdateRequisitionRequiresOwnershipAndActiveMembership(t *testing.T) {
	t.Parallel()

	t.Run("inactive membership", func(t *testing.T) {
		t.Parallel()

		requestedItem := updateTestRequestedItem(t, "ritem-requested", "Laptop stand", "Adjustable stand", "unreviewed", "", time.Now().UTC())
		repository := &updateRequisitionRepo{current: updateTestRequisition(t, requestedItem)}
		membership, err := departmentdomain.ReconstituteDepartmentMembership(
			"usr-123", "dpt-123", "Staff",
			departmentdomain.MembershipStatusInactive.String(), time.Now(),
		)
		if err != nil {
			t.Fatal(err)
		}
		usecase := NewUpdateRequisitionUsecase(
			repository, createCategoryRepo{}, createItemRepo{},
			createMembership{membership: membership}, &createIDGenerator{},
			&createTransactionManager{},
		)

		_, err = usecase.Execute(context.Background(), updateTestCommand(
			requisitionin.UpdateRequisitionLine{
				RequisitionItemID: "reqitem-catalogue", ItemID: "itm-123",
				Quantity: 1, EstimatedUnitAmount: 1,
			},
		))
		if err != ErrInactiveMembership {
			t.Fatalf("error = %v, want %v", err, ErrInactiveMembership)
		}
	})
}
