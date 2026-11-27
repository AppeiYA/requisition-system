package usecase

import (
	"context"
	"errors"
	"testing"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
)

type itemLookupRepo struct {
	*createRequisitionRepo
	parentID string
	itemErr  error
}

func (r *itemLookupRepo) GetRequisitionIDByItemID(context.Context, string) (string, error) {
	return r.parentID, r.itemErr
}

func TestGetRequisitionItemUsecaseAuthorisesThroughParent(t *testing.T) {
	t.Parallel()
	requisition := transitionTestRequisition(t, "usr-1")
	repo := &itemLookupRepo{
		createRequisitionRepo: &createRequisitionRepo{get: requisition},
		parentID:              "req-1",
	}
	uc := NewGetRequisitionItemUsecase(repo)

	got, err := uc.Execute(context.Background(), requisitionin.GetRequisitionItemQuery{
		RequisitionItemID: "reqitem-1", ActorID: "usr-1", ActorRole: "user",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got != requisition {
		t.Error("expected the parent requisition")
	}

	if _, err := uc.Execute(context.Background(), requisitionin.GetRequisitionItemQuery{
		RequisitionItemID: "reqitem-1", ActorID: "usr-2", ActorRole: "user",
	}); !errors.Is(err, ErrRequisitionAccessDenied) {
		t.Errorf("error = %v, want access denied", err)
	}
}

func TestGetRequisitionItemUsecaseValidatesIDAndPropagatesLookup(t *testing.T) {
	t.Parallel()
	uc := NewGetRequisitionItemUsecase(&itemLookupRepo{createRequisitionRepo: &createRequisitionRepo{}})

	if _, err := uc.Execute(context.Background(), requisitionin.GetRequisitionItemQuery{
		RequisitionItemID: "req-1", ActorID: "usr-1", ActorRole: "user",
	}); !errors.Is(err, requisitiondomain.ErrInvalidRequisitionItemID) {
		t.Errorf("error = %v, want invalid item ID", err)
	}

	notFound := errors.New("item missing")
	uc = NewGetRequisitionItemUsecase(&itemLookupRepo{
		createRequisitionRepo: &createRequisitionRepo{}, itemErr: notFound,
	})
	if _, err := uc.Execute(context.Background(), requisitionin.GetRequisitionItemQuery{
		RequisitionItemID: "reqitem-1", ActorID: "usr-1", ActorRole: "user",
	}); !errors.Is(err, notFound) {
		t.Errorf("error = %v, want lookup error", err)
	}
}
