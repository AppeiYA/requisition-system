package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationmocks "github.com/AppeiYA/requisition-system/internal/invitation/mocks"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
)

func acceptCmd() invitationin.AcceptInvitationCommand {
	return invitationin.AcceptInvitationCommand{
		Token:    testInvitationToken,
		Phone:    "+14155551234",
		Password: "Password123$",
	}
}

func newAcceptUsecase(
	repo invitationout.InvitationRepository,
	identityPort invitationout.InvitationIdentityPort,
	departmentPort invitationout.InvitationDepartmentPort,
	tx invitationout.TransactionManager,
) (*AcceptInvitationUsecase, *invitationmocks.MockTelemetry) {
	telemetry := invitationmocks.NewMockTelemetry()
	return NewAcceptInvitationUsecase(repo, identityPort, departmentPort, tx, telemetry), telemetry
}

func Test_AcceptInvitationUsecase_Success(t *testing.T) {
	t.Parallel()

	var registered *invitationout.CreateIdentityRequest
	var membership *invitationout.CreateMembershipRequest
	var acceptedID string
	var txUsed, membershipInsideTx, measuring, txInsideMeasurement bool

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		if tokenHash != domain.HashInvitationToken(testInvitationToken) {
			t.Errorf("lookup used wrong token hash")
		}
		return pendingInvitation(), nil
	}
	repo.MarkInvitationAcceptedFunc = func(ctx context.Context, invitationID string) error {
		acceptedID = invitationID
		return nil
	}

	identityPort := invitationmocks.NewMockInvitationIdentityPort()
	identityPort.RegisterUserFunc = func(ctx context.Context, req *invitationout.CreateIdentityRequest) (string, error) {
		registered = req
		return testRegisteredUserID, nil
	}

	departmentPort := invitationmocks.NewMockInvitationDepartmentPort()
	departmentPort.CreateMembershipFunc = func(ctx context.Context, req *invitationout.CreateMembershipRequest) error {
		membership = req
		membershipInsideTx = txUsed
		return nil
	}

	tx := invitationmocks.NewMockTransactionManager()
	tx.WithinTransactionFunc = func(ctx context.Context, fn func(context.Context) error) error {
		txUsed = true
		txInsideMeasurement = measuring
		return fn(ctx)
	}

	uc, telemetry := newAcceptUsecase(repo, identityPort, departmentPort, tx)

	telemetry.MeasureFunc = func(ctx context.Context, operation string, fn func(context.Context) error) error {
		measuring = true
		defer func() { measuring = false }()
		return fn(ctx)
	}

	if err := uc.Execute(context.Background(), acceptCmd()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if !txUsed {
		t.Error("expected registration to run inside a transaction")
	}
	if !membershipInsideTx {
		t.Error("expected membership creation to run inside the same transaction")
	}
	if !txInsideMeasurement {
		t.Error("expected the transaction to run inside the measured operation")
	}
	if registered == nil {
		t.Fatal("expected identity registration")
	}
	if registered.Email != testInvitationEmail {
		t.Errorf("expected email %q, got %q", testInvitationEmail, registered.Email)
	}
	if registered.FirstName != testInvitationFirstName || registered.LastName != testInvitationLastName {
		t.Errorf("expected invitation prefill name %q %q, got %q %q", testInvitationFirstName, testInvitationLastName, registered.FirstName, registered.LastName)
	}
	if registered.Phone != "+14155551234" {
		t.Errorf("expected phone to be forwarded, got %q", registered.Phone)
	}
	if registered.StaffID != testInvitationStaffID {
		t.Errorf("expected staff id %q to be forwarded, got %q", testInvitationStaffID, registered.StaffID)
	}
	if registered.Password != "Password123$" {
		t.Error("expected password to be forwarded")
	}
	if membership == nil {
		t.Fatal("expected department membership creation")
	}
	if membership.UserID != testRegisteredUserID {
		t.Errorf("expected membership user id %q, got %q", testRegisteredUserID, membership.UserID)
	}
	if membership.DepartmentID != testDepartmentID {
		t.Errorf("expected membership department id %q, got %q", testDepartmentID, membership.DepartmentID)
	}
	if membership.JobTitle != testJobTitle {
		t.Errorf("expected membership job title %q, got %q", testJobTitle, membership.JobTitle)
	}
	if acceptedID != testInvitationID {
		t.Errorf("expected invitation %q to be marked accepted, got %q", testInvitationID, acceptedID)
	}

	calls := telemetry.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected the operation to be measured once, got %d measurements", len(calls))
	}
	if calls[0].Operation != invitationout.OperationAcceptInvitation {
		t.Errorf("expected operation %q, got %q", invitationout.OperationAcceptInvitation, calls[0].Operation)
	}
	if calls[0].Err != nil {
		t.Errorf("expected a successful measurement, got: %v", calls[0].Err)
	}
}

func Test_AcceptInvitationUsecase_UsesPrefillWhenOmitted(t *testing.T) {
	t.Parallel()

	var registered *invitationout.CreateIdentityRequest

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return pendingInvitation(), nil
	}

	identityPort := invitationmocks.NewMockInvitationIdentityPort()
	identityPort.RegisterUserFunc = func(ctx context.Context, req *invitationout.CreateIdentityRequest) (string, error) {
		registered = req
		return testRegisteredUserID, nil
	}

	uc, _ := newAcceptUsecase(repo, identityPort, invitationmocks.NewMockInvitationDepartmentPort(), invitationmocks.NewMockTransactionManager())

	cmd := invitationin.AcceptInvitationCommand{
		Token:    testInvitationToken,
		Password: "Password123$",
	}

	if err := uc.Execute(context.Background(), cmd); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if registered == nil {
		t.Fatal("expected identity registration")
	}
	if registered.FirstName != testInvitationFirstName || registered.LastName != testInvitationLastName {
		t.Errorf("expected prefill name %q %q, got %q %q", testInvitationFirstName, testInvitationLastName, registered.FirstName, registered.LastName)
	}
	if registered.Phone != testInvitationPhone {
		t.Errorf("expected prefill phone %q, got %q", testInvitationPhone, registered.Phone)
	}
	if registered.StaffID != testInvitationStaffID {
		t.Errorf("expected prefill staff id %q, got %q", testInvitationStaffID, registered.StaffID)
	}
}

func Test_AcceptInvitationUsecase_EmptyToken(t *testing.T) {
	t.Parallel()

	uc, _ := newAcceptUsecase(
		invitationmocks.NewMockInvitationRepository(),
		invitationmocks.NewMockInvitationIdentityPort(),
		invitationmocks.NewMockInvitationDepartmentPort(),
		invitationmocks.NewMockTransactionManager(),
	)

	err := uc.Execute(context.Background(), invitationin.AcceptInvitationCommand{})
	if err != ErrInvitationInvalidToken {
		t.Fatalf("expected ErrInvitationInvalidToken, got: %v", err)
	}
}

func Test_AcceptInvitationUsecase_NotFound(t *testing.T) {
	t.Parallel()

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return nil, domain.ErrInvitationNotFound
	}

	uc, _ := newAcceptUsecase(repo, invitationmocks.NewMockInvitationIdentityPort(), invitationmocks.NewMockInvitationDepartmentPort(), invitationmocks.NewMockTransactionManager())

	err := uc.Execute(context.Background(), acceptCmd())
	if err != ErrInvitationNotFound {
		t.Fatalf("expected ErrInvitationNotFound, got: %v", err)
	}
}

func Test_AcceptInvitationUsecase_InvalidStates(t *testing.T) {
	t.Parallel()

	past := time.Now().Add(-time.Hour)
	present := time.Now()

	tests := []struct {
		name        string
		inv         *domain.Invitation
		expectedErr error
	}{
		{name: "expired", inv: invitationWithState(past, nil, nil), expectedErr: ErrInvitationExpired},
		{name: "revoked", inv: invitationWithState(time.Now().Add(time.Hour), nil, &present), expectedErr: ErrInvitationRevoked},
		{name: "accepted", inv: invitationWithState(time.Now().Add(time.Hour), &present, nil), expectedErr: ErrInvitationAlreadyUsed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := invitationmocks.NewMockInvitationRepository()
			repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
				return tt.inv, nil
			}

			tx := invitationmocks.NewMockTransactionManager()
			tx.WithinTransactionFunc = func(ctx context.Context, fn func(context.Context) error) error {
				t.Fatal("transaction must not start for an invalid invitation")
				return nil
			}

			uc, _ := newAcceptUsecase(repo, invitationmocks.NewMockInvitationIdentityPort(), invitationmocks.NewMockInvitationDepartmentPort(), tx)

			err := uc.Execute(context.Background(), acceptCmd())
			if err != tt.expectedErr {
				t.Fatalf("expected %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func Test_AcceptInvitationUsecase_RegisterUserError(t *testing.T) {
	t.Parallel()

	regErr := errors.New("registration failed")
	markCalled := false
	membershipCalled := false

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return pendingInvitation(), nil
	}
	repo.MarkInvitationAcceptedFunc = func(ctx context.Context, invitationID string) error {
		markCalled = true
		return nil
	}

	identityPort := invitationmocks.NewMockInvitationIdentityPort()
	identityPort.RegisterUserFunc = func(ctx context.Context, req *invitationout.CreateIdentityRequest) (string, error) {
		return "", regErr
	}

	departmentPort := invitationmocks.NewMockInvitationDepartmentPort()
	departmentPort.CreateMembershipFunc = func(ctx context.Context, req *invitationout.CreateMembershipRequest) error {
		membershipCalled = true
		return nil
	}

	uc, _ := newAcceptUsecase(repo, identityPort, departmentPort, invitationmocks.NewMockTransactionManager())

	err := uc.Execute(context.Background(), acceptCmd())
	if !errors.Is(err, regErr) {
		t.Fatalf("expected regErr, got: %v", err)
	}
	if membershipCalled {
		t.Error("membership must not be created when registration fails")
	}
	if markCalled {
		t.Error("invitation must not be marked accepted when registration fails")
	}
}

func Test_AcceptInvitationUsecase_CreateMembershipError(t *testing.T) {
	t.Parallel()

	departmentErr := errors.New("membership failed")
	markCalled := false

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return pendingInvitation(), nil
	}
	repo.MarkInvitationAcceptedFunc = func(ctx context.Context, invitationID string) error {
		markCalled = true
		return nil
	}

	departmentPort := invitationmocks.NewMockInvitationDepartmentPort()
	departmentPort.CreateMembershipFunc = func(ctx context.Context, req *invitationout.CreateMembershipRequest) error {
		return departmentErr
	}

	uc, _ := newAcceptUsecase(repo, invitationmocks.NewMockInvitationIdentityPort(), departmentPort, invitationmocks.NewMockTransactionManager())

	err := uc.Execute(context.Background(), acceptCmd())
	if !errors.Is(err, departmentErr) {
		t.Fatalf("expected departmentErr, got: %v", err)
	}
	if markCalled {
		t.Error("invitation must not be marked accepted when membership creation fails")
	}
}

func Test_AcceptInvitationUsecase_MarkAcceptedError(t *testing.T) {
	t.Parallel()

	markErr := errors.New("mark failed")

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return pendingInvitation(), nil
	}
	repo.MarkInvitationAcceptedFunc = func(ctx context.Context, invitationID string) error {
		return markErr
	}

	uc, _ := newAcceptUsecase(repo, invitationmocks.NewMockInvitationIdentityPort(), invitationmocks.NewMockInvitationDepartmentPort(), invitationmocks.NewMockTransactionManager())

	err := uc.Execute(context.Background(), acceptCmd())
	if !errors.Is(err, markErr) {
		t.Fatalf("expected markErr, got: %v", err)
	}
}

func Test_AcceptInvitationUsecase_TransactionError(t *testing.T) {
	t.Parallel()

	txErr := errors.New("tx failed")

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return pendingInvitation(), nil
	}

	tx := invitationmocks.NewMockTransactionManager()
	tx.WithinTransactionFunc = func(ctx context.Context, fn func(context.Context) error) error {
		return txErr
	}

	uc, _ := newAcceptUsecase(repo, invitationmocks.NewMockInvitationIdentityPort(), invitationmocks.NewMockInvitationDepartmentPort(), tx)

	err := uc.Execute(context.Background(), acceptCmd())
	if !errors.Is(err, txErr) {
		t.Fatalf("expected txErr, got: %v", err)
	}
}

func Test_AcceptInvitationUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc, telemetry := newAcceptUsecase(
		invitationmocks.NewMockInvitationRepository(),
		invitationmocks.NewMockInvitationIdentityPort(),
		invitationmocks.NewMockInvitationDepartmentPort(),
		invitationmocks.NewMockTransactionManager(),
	)

	err := uc.Execute(ctx, acceptCmd())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if calls := telemetry.Calls(); len(calls) != 0 {
		t.Errorf("expected a cancelled request not to be measured, got %d measurements", len(calls))
	}
}

func Test_AcceptInvitationUsecase_MeasuresValidationFailure(t *testing.T) {
	t.Parallel()

	uc, telemetry := newAcceptUsecase(
		invitationmocks.NewMockInvitationRepository(),
		invitationmocks.NewMockInvitationIdentityPort(),
		invitationmocks.NewMockInvitationDepartmentPort(),
		invitationmocks.NewMockTransactionManager(),
	)

	err := uc.Execute(context.Background(), invitationin.AcceptInvitationCommand{})
	if err != ErrInvitationInvalidToken {
		t.Fatalf("expected ErrInvitationInvalidToken, got: %v", err)
	}

	calls := telemetry.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected the operation to be measured once, got %d measurements", len(calls))
	}
	if calls[0].Operation != invitationout.OperationAcceptInvitation {
		t.Errorf("expected operation %q, got %q", invitationout.OperationAcceptInvitation, calls[0].Operation)
	}
	if !errors.Is(calls[0].Err, ErrInvitationInvalidToken) {
		t.Errorf("expected the measured error to be ErrInvitationInvalidToken, got: %v", calls[0].Err)
	}
}

func Test_AcceptInvitationUsecase_MeasuresBusinessFailure(t *testing.T) {
	t.Parallel()

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return nil, domain.ErrInvitationNotFound
	}

	uc, telemetry := newAcceptUsecase(repo, invitationmocks.NewMockInvitationIdentityPort(), invitationmocks.NewMockInvitationDepartmentPort(), invitationmocks.NewMockTransactionManager())

	err := uc.Execute(context.Background(), acceptCmd())
	if err != ErrInvitationNotFound {
		t.Fatalf("expected ErrInvitationNotFound, got: %v", err)
	}

	calls := telemetry.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected the operation to be measured once, got %d measurements", len(calls))
	}
	if calls[0].Operation != invitationout.OperationAcceptInvitation {
		t.Errorf("expected operation %q, got %q", invitationout.OperationAcceptInvitation, calls[0].Operation)
	}
	if !errors.Is(calls[0].Err, ErrInvitationNotFound) {
		t.Errorf("expected the measured error to be ErrInvitationNotFound, got: %v", calls[0].Err)
	}
}

func Test_AcceptInvitationUsecase_MeasurementKeepsTheContext(t *testing.T) {
	t.Parallel()

	type contextKey string
	const key contextKey = "trace"

	var seen any
	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		seen = ctx.Value(key)
		return pendingInvitation(), nil
	}

	uc, _ := newAcceptUsecase(repo, invitationmocks.NewMockInvitationIdentityPort(), invitationmocks.NewMockInvitationDepartmentPort(), invitationmocks.NewMockTransactionManager())

	ctx := context.WithValue(context.Background(), key, "span-context")
	if err := uc.Execute(ctx, acceptCmd()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if seen != "span-context" {
		t.Errorf("expected the measured operation to receive the original context, got %v", seen)
	}
}

func Test_AcceptInvitationUsecase_InstrumentationFailureDoesNotChangeTheOutcome(t *testing.T) {
	t.Parallel()

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return pendingInvitation(), nil
	}

	telemetry := invitationmocks.NewMockTelemetry()
	telemetry.MeasureFunc = func(ctx context.Context, operation string, fn func(context.Context) error) error {
		_ = fn(ctx)
		return errors.New("telemetry unavailable")
	}

	uc := NewAcceptInvitationUsecase(
		repo,
		invitationmocks.NewMockInvitationIdentityPort(),
		invitationmocks.NewMockInvitationDepartmentPort(),
		invitationmocks.NewMockTransactionManager(),
		telemetry,
	)

	if err := uc.Execute(context.Background(), acceptCmd()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func Test_AcceptInvitationUsecase_InstrumentationFailureDoesNotMaskTheBusinessError(t *testing.T) {
	t.Parallel()

	regErr := errors.New("registration failed")

	repo := invitationmocks.NewMockInvitationRepository()
	repo.GetInvitationByTokenHashFunc = func(ctx context.Context, tokenHash string) (*domain.Invitation, error) {
		return pendingInvitation(), nil
	}

	identityPort := invitationmocks.NewMockInvitationIdentityPort()
	identityPort.RegisterUserFunc = func(ctx context.Context, req *invitationout.CreateIdentityRequest) (string, error) {
		return "", regErr
	}

	telemetry := invitationmocks.NewMockTelemetry()
	telemetry.MeasureFunc = func(ctx context.Context, operation string, fn func(context.Context) error) error {
		_ = fn(ctx)
		return errors.New("telemetry unavailable")
	}

	uc := NewAcceptInvitationUsecase(
		repo,
		identityPort,
		invitationmocks.NewMockInvitationDepartmentPort(),
		invitationmocks.NewMockTransactionManager(),
		telemetry,
	)

	err := uc.Execute(context.Background(), acceptCmd())
	if !errors.Is(err, regErr) {
		t.Fatalf("expected regErr, got: %v", err)
	}
}
