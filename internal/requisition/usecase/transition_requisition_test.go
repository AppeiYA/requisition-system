package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func transitionTestRequisition(t *testing.T, requestedBy string) *requisitiondomain.Requisition {
	t.Helper()

	name, err := requisitiondomain.NewItemName("Laptop")
	if err != nil {
		t.Fatal(err)
	}
	quantity, err := requisitiondomain.NewRequisitionQuantity(1)
	if err != nil {
		t.Fatal(err)
	}
	amount, err := shareddomain.NewNairaMoneyFromMinorUnits(150000)
	if err != nil {
		t.Fatal(err)
	}
	item, err := requisitiondomain.NewRequisitionItem(
		"reqitem-1", "req-1", 1, "itm-1", "", name, quantity, amount, "",
	)
	if err != nil {
		t.Fatal(err)
	}
	justification, err := requisitiondomain.NewRequisitionJustification("Office setup")
	if err != nil {
		t.Fatal(err)
	}
	requisition, err := requisitiondomain.NewSubmittedRequisition(
		"req-1", "REQ-2026-10001", requestedBy, "dpt-1",
		requisitiondomain.PriorityMedium, time.Now().UTC().AddDate(0, 0, 2),
		justification, []*requisitiondomain.RequisitionItem{item},
	)
	if err != nil {
		t.Fatal(err)
	}
	initial, err := requisition.InitialStatusHistory(
		"rsh-1", requestedBy, requisitiondomain.ActorRoleStaff, time.Now().UTC(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := requisition.RecordInitialStatusHistory(initial); err != nil {
		t.Fatal(err)
	}
	return requisition
}

type transitionRequisitionRepo struct {
	requisition *requisitiondomain.Requisition
	getErr      error
	updateErr   error
	locked      bool
	updated     *requisitiondomain.Requisition
	history     *requisitiondomain.RequisitionStatusHistory
	updateCalls int
}

func (r *transitionRequisitionRepo) GetRequisitionByIDForUpdate(context.Context, string) (*requisitiondomain.Requisition, error) {
	r.locked = true
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.requisition, nil
}

func (r *transitionRequisitionRepo) UpdateRequisitionStatus(
	_ context.Context,
	requisition *requisitiondomain.Requisition,
	history *requisitiondomain.RequisitionStatusHistory,
) error {
	r.updateCalls++
	r.updated = requisition
	r.history = history
	return r.updateErr
}

var _ requisitionout.RequisitionTransitionRepository = (*transitionRequisitionRepo)(nil)

func newTransitionUsecase(repository *transitionRequisitionRepo, transaction *createTransactionManager) *TransitionRequisitionUsecase {
	return NewTransitionRequisitionUsecase(repository, &createIDGenerator{}, transaction)
}

func TestTransitionRequisitionUsecaseApprovesAsAdmin(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	transaction := &createTransactionManager{}
	usecase := newTransitionUsecase(repository, transaction)

	updated, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1",
		ActorID:       "usr-admin",
		ActorRole:     "admin",
		ToStatus:      "approved",
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if !transaction.called || !repository.locked {
		t.Fatal("expected transition to lock the requisition inside a transaction")
	}
	if updated.Status() != requisitiondomain.RequisitionStatusApproved {
		t.Fatalf("status = %q, want approved", updated.Status())
	}
	if repository.history == nil {
		t.Fatal("expected a status history row to be persisted")
	}
	if repository.history.Sequence() != 2 {
		t.Fatalf("sequence = %d, want 2", repository.history.Sequence())
	}
	if repository.history.FromStatus() == nil || *repository.history.FromStatus() != requisitiondomain.RequisitionStatusPending {
		t.Fatal("history should record pending as the previous status")
	}
	if repository.history.ToStatus() != requisitiondomain.RequisitionStatusApproved {
		t.Fatalf("to status = %q, want approved", repository.history.ToStatus())
	}
	if repository.history.ActedBy() != "usr-admin" || repository.history.ActorRole() != requisitiondomain.ActorRoleAdmin {
		t.Fatalf("actor = %q/%q", repository.history.ActedBy(), repository.history.ActorRole())
	}
	if repository.history.CreatedAt().IsZero() {
		t.Fatal("history should carry the action timestamp")
	}
}

func TestTransitionRequisitionUsecaseRejectsWithReason(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1",
		ActorID:       "usr-admin",
		ActorRole:     "admin",
		ToStatus:      "rejected",
		Reason:        "No budget available",
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if repository.history == nil || repository.history.ToStatus() != requisitiondomain.RequisitionStatusRejected {
		t.Fatalf("history = %+v, want rejected", repository.history)
	}
	if repository.history.Reason() != "No budget available" {
		t.Fatalf("reason = %q", repository.history.Reason())
	}
}

func TestTransitionRequisitionUsecaseCancelsOwnPendingAsStaff(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	updated, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1",
		ActorID:       "usr-1",
		ActorRole:     "user",
		ToStatus:      "cancelled",
		Reason:        "No longer required",
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if updated.Status() != requisitiondomain.RequisitionStatusCancelled {
		t.Fatalf("status = %q, want cancelled", updated.Status())
	}
	if repository.history.ActorRole() != requisitiondomain.ActorRoleStaff {
		t.Fatalf("actor role = %q, want staff", repository.history.ActorRole())
	}
	if repository.history.Reason() != "No longer required" {
		t.Fatalf("reason = %q", repository.history.Reason())
	}
}

func TestTransitionRequisitionUsecaseRequiresReasonForRejectionAndCancellation(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"rejected", "cancelled"} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
			usecase := newTransitionUsecase(repository, &createTransactionManager{})

			_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
				RequisitionID: "req-1",
				ActorID:       "usr-admin",
				ActorRole:     "admin",
				ToStatus:      status,
			})
			if err != requisitiondomain.ErrStatusReasonRequired {
				t.Fatalf("Execute() error = %v, want %v", err, requisitiondomain.ErrStatusReasonRequired)
			}
			if repository.updateCalls != 0 {
				t.Fatal("invalid transition should not persist anything")
			}
		})
	}
}

func TestTransitionRequisitionUsecaseRejectsInvalidTransition(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1",
		ActorID:       "usr-admin",
		ActorRole:     "admin",
		ToStatus:      "completed",
	})
	if err == nil {
		t.Fatal("pending -> completed should be rejected")
	}
	if repository.updateCalls != 0 {
		t.Fatal("invalid transition should not persist anything")
	}
}

func TestTransitionRequisitionUsecasePreservesEarlierHistory(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	if _, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1", ActorID: "usr-admin", ActorRole: "admin", ToStatus: "approved",
	}); err != nil {
		t.Fatalf("approval unexpected error = %v", err)
	}
	if _, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1", ActorID: "usr-admin", ActorRole: "admin", ToStatus: "in_fulfillment",
	}); err != nil {
		t.Fatalf("fulfillment unexpected error = %v", err)
	}

	history := repository.requisition.StatusHistory()
	if len(history) != 3 {
		t.Fatalf("history length = %d, want 3", len(history))
	}
	if history[0].ToStatus() != requisitiondomain.RequisitionStatusPending ||
		history[1].ToStatus() != requisitiondomain.RequisitionStatusApproved ||
		history[2].ToStatus() != requisitiondomain.RequisitionStatusInFulfillment {
		t.Fatalf("history statuses = %q, %q, %q", history[0].ToStatus(), history[1].ToStatus(), history[2].ToStatus())
	}
	if history[2].Sequence() != 3 {
		t.Fatalf("sequence = %d, want 3", history[2].Sequence())
	}
}

func TestTransitionRequisitionUsecaseEnforcesOwnership(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1",
		ActorID:       "usr-other",
		ActorRole:     "user",
		ToStatus:      "cancelled",
		Reason:        "No longer required",
	})
	if !errors.Is(err, ErrRequisitionTransitionForbidden) {
		t.Fatalf("Execute() error = %v, want %v", err, ErrRequisitionTransitionForbidden)
	}
	if repository.updateCalls != 0 {
		t.Fatal("unauthorized transition should not persist anything")
	}
}

func TestTransitionRequisitionUsecaseStaffCannotApprove(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1",
		ActorID:       "usr-1",
		ActorRole:     "user",
		ToStatus:      "approved",
	})
	if !errors.Is(err, ErrRequisitionTransitionForbidden) {
		t.Fatalf("Execute() error = %v, want %v", err, ErrRequisitionTransitionForbidden)
	}
}

func TestTransitionRequisitionUsecasePropagatesPersistenceFailure(t *testing.T) {
	t.Parallel()

	updateErr := errors.New("history insert failed")
	repository := &transitionRequisitionRepo{
		requisition: transitionTestRequisition(t, "usr-1"),
		updateErr:   updateErr,
	}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1", ActorID: "usr-admin", ActorRole: "admin", ToStatus: "approved",
	})
	if !errors.Is(err, updateErr) {
		t.Fatalf("Execute() error = %v, want %v", err, updateErr)
	}
	if repository.updateCalls != 1 {
		t.Fatalf("update calls = %d, want 1", repository.updateCalls)
	}
}

func TestTransitionRequisitionUsecaseRejectsInvalidActorRole(t *testing.T) {
	t.Parallel()

	repository := &transitionRequisitionRepo{requisition: transitionTestRequisition(t, "usr-1")}
	usecase := newTransitionUsecase(repository, &createTransactionManager{})

	_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "req-1", ActorID: "usr-1", ActorRole: "superuser", ToStatus: "cancelled", Reason: "x",
	})
	if err != requisitiondomain.ErrInvalidActorRole {
		t.Fatalf("Execute() error = %v, want %v", err, requisitiondomain.ErrInvalidActorRole)
	}
}

func TestTransitionRequisitionUsecaseRejectsInvalidRequisitionID(t *testing.T) {
	t.Parallel()

	usecase := newTransitionUsecase(&transitionRequisitionRepo{}, &createTransactionManager{})

	_, err := usecase.Execute(context.Background(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: "invalid", ActorID: "usr-admin", ActorRole: "admin", ToStatus: "approved",
	})
	if err != requisitiondomain.ErrInvalidRequisitionID {
		t.Fatalf("Execute() error = %v, want %v", err, requisitiondomain.ErrInvalidRequisitionID)
	}
}
