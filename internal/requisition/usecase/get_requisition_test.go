package usecase

import (
	"context"
	"errors"
	"testing"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
)

func TestGetRequisitionUsecaseReturnsOwnRequisition(t *testing.T) {
	t.Parallel()

	requisition := transitionTestRequisition(t, "usr-1")
	usecase := NewGetRequisitionUsecase(&createRequisitionRepo{get: requisition})

	got, err := usecase.Execute(context.Background(), requisitionin.GetRequisitionQuery{
		RequisitionID: "req-1", ActorID: "usr-1", ActorRole: "user",
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if got != requisition {
		t.Fatal("expected the stored requisition")
	}
}

func TestGetRequisitionUsecaseAllowsAdminToReadAnotherUsersRequisition(t *testing.T) {
	t.Parallel()

	requisition := transitionTestRequisition(t, "usr-1")
	usecase := NewGetRequisitionUsecase(&createRequisitionRepo{get: requisition})

	if _, err := usecase.Execute(context.Background(), requisitionin.GetRequisitionQuery{
		RequisitionID: "req-1", ActorID: "usr-admin", ActorRole: "admin",
	}); err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
}

func TestGetRequisitionUsecaseRejectsAnotherUsersRequisition(t *testing.T) {
	t.Parallel()

	requisition := transitionTestRequisition(t, "usr-1")
	usecase := NewGetRequisitionUsecase(&createRequisitionRepo{get: requisition})

	_, err := usecase.Execute(context.Background(), requisitionin.GetRequisitionQuery{
		RequisitionID: "req-1", ActorID: "usr-other", ActorRole: "user",
	})
	if !errors.Is(err, ErrRequisitionAccessDenied) {
		t.Fatalf("Execute() error = %v, want %v", err, ErrRequisitionAccessDenied)
	}
}

func TestGetRequisitionUsecaseRejectsInvalidID(t *testing.T) {
	t.Parallel()

	usecase := NewGetRequisitionUsecase(&createRequisitionRepo{})

	_, err := usecase.Execute(context.Background(), requisitionin.GetRequisitionQuery{
		RequisitionID: "invalid", ActorID: "usr-1", ActorRole: "user",
	})
	if err != requisitiondomain.ErrInvalidRequisitionID {
		t.Fatalf("Execute() error = %v, want %v", err, requisitiondomain.ErrInvalidRequisitionID)
	}
}

func TestGetRequisitionUsecasePropagatesNotFound(t *testing.T) {
	t.Parallel()

	notFound := errors.New("not found")
	usecase := NewGetRequisitionUsecase(&createRequisitionRepo{err: notFound})

	_, err := usecase.Execute(context.Background(), requisitionin.GetRequisitionQuery{
		RequisitionID: "req-1", ActorID: "usr-1", ActorRole: "user",
	})
	if !errors.Is(err, notFound) {
		t.Fatalf("Execute() error = %v, want %v", err, notFound)
	}
}
