package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	adminmocks "github.com/AppeiYA/requisition-system/internal/admin/mocks"
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	adminout "github.com/AppeiYA/requisition-system/internal/admin/ports/out"
	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

func inviteCmd() adminin.InviteMemberCommand {
	return adminin.InviteMemberCommand{
		InviterID:    adminmocks.DefaultInviterID,
		Email:        adminmocks.DefaultEmail,
		FirstName:    adminmocks.DefaultFirstName,
		LastName:     adminmocks.DefaultLastName,
		Phone:        adminmocks.DefaultPhone,
		StaffID:      adminmocks.DefaultStaffID,
		DepartmentID: adminmocks.DefaultDepartmentID,
		JobTitle:     adminmocks.DefaultJobTitle,
	}
}

func newInviteUsecase(
	taskQueue adminout.TaskQueue,
	identity adminout.AdminIdentityPort,
	invitations adminout.AdminInvitationPort,
	department adminout.AdminDepartmentPort,
	tx adminout.TransactionManager,
) (*InviteMemberUsecase, *adminmocks.MockTelemetry) {
	telemetry := adminmocks.NewMockTelemetry()
	return NewInviteMemberUsecase(taskQueue, identity, invitations, department, tx, telemetry), telemetry
}

func Test_InviteMemberUsecase_Success(t *testing.T) {
	t.Parallel()

	var savedEmail string
	var savedInviter string
	var savedFirstName, savedLastName, savedPhone, savedStaffID string
	var savedDepartmentID, savedJobTitle string
	var lookedUpDepartmentID string
	var queuedTo, queuedInvitedBy, queuedToken, queuedExpiresAt string
	var queuedDepartmentName, queuedJobTitle string
	var txUsed bool

	identity := adminmocks.NewMockAdminIdentityPort()
	identity.CheckUserExistsFunc = func(ctx context.Context, email *shareddomain.Email) (bool, error) {
		return false, nil
	}

	department := adminmocks.NewMockAdminDepartmentPort()
	department.GetDepartmentByIDFunc = func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
		lookedUpDepartmentID = departmentID
		return &adminout.DepartmentResponse{ID: departmentID, Name: adminmocks.DefaultDepartmentName}, nil
	}

	invitations := adminmocks.NewMockAdminInvitationPort()
	invitations.SaveInvitationFunc = func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
		savedEmail = params.Email.Value()
		savedInviter = params.InviterID
		savedFirstName = params.FirstName
		savedLastName = params.LastName
		savedPhone = params.Phone
		savedStaffID = params.StaffID
		savedDepartmentID = params.DepartmentID
		savedJobTitle = params.JobTitle
		return &adminout.SavedInvitation{
			ID:        adminmocks.DefaultInvitationID,
			Token:     adminmocks.DefaultToken,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}, nil
	}

	queue := adminmocks.NewMockTaskQueue()
	queue.EnqueueInvitationEmailFunc = func(ctx context.Context, payload tasks.InvitationEmailPayload) error {
		queuedTo = payload.To
		queuedInvitedBy = payload.InvitedBy
		queuedToken = payload.Token
		queuedExpiresAt = payload.ExpiresAt
		queuedDepartmentName = payload.DepartmentName
		queuedJobTitle = payload.JobTitle
		return nil
	}

	tx := adminmocks.NewMockTransactionManager()
	tx.WithinTransactionFunc = func(ctx context.Context, fn func(context.Context) error) error {
		txUsed = true
		return fn(ctx)
	}

	uc, telemetry := newInviteUsecase(queue, identity, invitations, department, tx)

	res, err := uc.Execute(context.Background(), inviteCmd())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if !txUsed {
		t.Error("expected invitation to be saved inside a transaction")
	}
	if lookedUpDepartmentID != adminmocks.DefaultDepartmentID {
		t.Errorf("expected department lookup for %q, got %q", adminmocks.DefaultDepartmentID, lookedUpDepartmentID)
	}
	if res.InvitationID != adminmocks.DefaultInvitationID {
		t.Errorf("expected invitation id %q, got %q", adminmocks.DefaultInvitationID, res.InvitationID)
	}
	if savedEmail != adminmocks.DefaultEmail {
		t.Errorf("expected saved email %q, got %q", adminmocks.DefaultEmail, savedEmail)
	}
	if savedInviter != adminmocks.DefaultInviterID {
		t.Errorf("expected inviter %q, got %q", adminmocks.DefaultInviterID, savedInviter)
	}
	if savedFirstName != adminmocks.DefaultFirstName || savedLastName != adminmocks.DefaultLastName {
		t.Errorf("expected saved name %q %q, got %q %q", adminmocks.DefaultFirstName, adminmocks.DefaultLastName, savedFirstName, savedLastName)
	}
	if savedPhone != adminmocks.DefaultPhone {
		t.Errorf("expected saved phone %q, got %q", adminmocks.DefaultPhone, savedPhone)
	}
	if savedStaffID != adminmocks.DefaultStaffID {
		t.Errorf("expected saved staff id %q, got %q", adminmocks.DefaultStaffID, savedStaffID)
	}
	if savedDepartmentID != adminmocks.DefaultDepartmentID {
		t.Errorf("expected saved department id %q, got %q", adminmocks.DefaultDepartmentID, savedDepartmentID)
	}
	if savedJobTitle != adminmocks.DefaultJobTitle {
		t.Errorf("expected saved job title %q, got %q", adminmocks.DefaultJobTitle, savedJobTitle)
	}
	if queuedTo != adminmocks.DefaultEmail {
		t.Errorf("expected email queued for %q, got %q", adminmocks.DefaultEmail, queuedTo)
	}
	if queuedInvitedBy != adminmocks.DefaultInviterID {
		t.Errorf("expected queued inviter %q, got %q", adminmocks.DefaultInviterID, queuedInvitedBy)
	}
	if queuedToken != adminmocks.DefaultToken {
		t.Errorf("expected queued token %q, got %q", adminmocks.DefaultToken, queuedToken)
	}
	if queuedExpiresAt == "" {
		t.Error("expected invitation expiry to be queued with the task")
	}
	if queuedDepartmentName != adminmocks.DefaultDepartmentName {
		t.Errorf("expected queued department name %q, got %q", adminmocks.DefaultDepartmentName, queuedDepartmentName)
	}
	if queuedJobTitle != adminmocks.DefaultJobTitle {
		t.Errorf("expected queued job title %q, got %q", adminmocks.DefaultJobTitle, queuedJobTitle)
	}

	calls := telemetry.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected the operation to be measured once, got %d measurements", len(calls))
	}
	if calls[0].Operation != adminout.OperationInviteMember {
		t.Errorf("expected operation %q, got %q", adminout.OperationInviteMember, calls[0].Operation)
	}
	if calls[0].Err != nil {
		t.Errorf("expected a successful measurement, got: %v", calls[0].Err)
	}
}

func Test_InviteMemberUsecase_InvalidEmail(t *testing.T) {
	t.Parallel()

	uc, _ := newInviteUsecase(
		adminmocks.NewMockTaskQueue(),
		adminmocks.NewMockAdminIdentityPort(),
		adminmocks.NewMockAdminInvitationPort(),
		adminmocks.NewMockAdminDepartmentPort(),
		adminmocks.NewMockTransactionManager(),
	)

	_, err := uc.Execute(context.Background(), adminin.InviteMemberCommand{InviterID: adminmocks.DefaultInviterID, Email: "not-an-email"})
	if err != shareddomain.ErrInvalidEmailFormat {
		t.Fatalf("expected ErrInvalidEmailFormat, got: %v", err)
	}
}

func Test_InviteMemberUsecase_DepartmentNotFound(t *testing.T) {
	t.Parallel()

	saveCalled := false
	department := adminmocks.NewMockAdminDepartmentPort()
	department.GetDepartmentByIDFunc = func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
		return nil, nil
	}

	invitations := adminmocks.NewMockAdminInvitationPort()
	invitations.SaveInvitationFunc = func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
		saveCalled = true
		return &adminout.SavedInvitation{ID: adminmocks.DefaultInvitationID, Token: adminmocks.DefaultToken, ExpiresAt: time.Now()}, nil
	}

	uc, _ := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), invitations, department, adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if err != ErrDepartmentNotFound {
		t.Fatalf("expected ErrDepartmentNotFound, got: %v", err)
	}
	if saveCalled {
		t.Error("invitation must not be saved when the department does not exist")
	}
}

func Test_InviteMemberUsecase_DepartmentLookupError(t *testing.T) {
	t.Parallel()

	lookupErr := errors.New("department lookup failed")
	department := adminmocks.NewMockAdminDepartmentPort()
	department.GetDepartmentByIDFunc = func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
		return nil, lookupErr
	}

	saveCalled := false
	invitations := adminmocks.NewMockAdminInvitationPort()
	invitations.SaveInvitationFunc = func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
		saveCalled = true
		return &adminout.SavedInvitation{ID: adminmocks.DefaultInvitationID, Token: adminmocks.DefaultToken, ExpiresAt: time.Now()}, nil
	}

	uc, _ := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), invitations, department, adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookupErr, got: %v", err)
	}
	if saveCalled {
		t.Error("invitation must not be saved when the department cannot be resolved")
	}
}

func Test_InviteMemberUsecase_UserAlreadyExists(t *testing.T) {
	t.Parallel()

	identity := adminmocks.NewMockAdminIdentityPort()
	identity.CheckUserExistsFunc = func(ctx context.Context, email *shareddomain.Email) (bool, error) {
		return true, nil
	}

	uc, _ := newInviteUsecase(adminmocks.NewMockTaskQueue(), identity, adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func Test_InviteMemberUsecase_IdentityCheckError(t *testing.T) {
	t.Parallel()

	checkErr := errors.New("identity check failed")
	identity := adminmocks.NewMockAdminIdentityPort()
	identity.CheckUserExistsFunc = func(ctx context.Context, email *shareddomain.Email) (bool, error) {
		return false, checkErr
	}

	uc, _ := newInviteUsecase(adminmocks.NewMockTaskQueue(), identity, adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, checkErr) {
		t.Fatalf("expected checkErr, got: %v", err)
	}
}

func Test_InviteMemberUsecase_SaveError(t *testing.T) {
	t.Parallel()

	saveErr := errors.New("save failed")
	invitations := adminmocks.NewMockAdminInvitationPort()
	invitations.SaveInvitationFunc = func(ctx context.Context, params adminout.SaveInvitationParams) (*adminout.SavedInvitation, error) {
		return nil, saveErr
	}

	taskQueued := false
	queue := adminmocks.NewMockTaskQueue()
	queue.EnqueueInvitationEmailFunc = func(ctx context.Context, payload tasks.InvitationEmailPayload) error {
		taskQueued = true
		return nil
	}

	uc, _ := newInviteUsecase(queue, adminmocks.NewMockAdminIdentityPort(), invitations, adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
	if taskQueued {
		t.Error("no task must be queued when the invitation cannot be saved")
	}
}

func Test_InviteMemberUsecase_TransactionError(t *testing.T) {
	t.Parallel()

	txErr := errors.New("tx failed")
	tx := adminmocks.NewMockTransactionManager()
	tx.WithinTransactionFunc = func(ctx context.Context, fn func(context.Context) error) error {
		return txErr
	}

	uc, _ := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), tx)

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, txErr) {
		t.Fatalf("expected txErr, got: %v", err)
	}
}

func Test_InviteMemberUsecase_EnqueueError(t *testing.T) {
	t.Parallel()

	enqueueErr := errors.New("enqueue failed")
	queue := adminmocks.NewMockTaskQueue()
	queue.EnqueueInvitationEmailFunc = func(ctx context.Context, payload tasks.InvitationEmailPayload) error {
		return enqueueErr
	}

	uc, _ := newInviteUsecase(queue, adminmocks.NewMockAdminIdentityPort(), adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, enqueueErr) {
		t.Fatalf("expected enqueueErr, got: %v", err)
	}
}

func Test_InviteMemberUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc, telemetry := newInviteUsecase(
		adminmocks.NewMockTaskQueue(),
		adminmocks.NewMockAdminIdentityPort(),
		adminmocks.NewMockAdminInvitationPort(),
		adminmocks.NewMockAdminDepartmentPort(),
		adminmocks.NewMockTransactionManager(),
	)

	_, err := uc.Execute(ctx, inviteCmd())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
	if calls := telemetry.Calls(); len(calls) != 0 {
		t.Errorf("expected a cancelled request not to be measured, got %d measurements", len(calls))
	}
}

func Test_InviteMemberUsecase_MeasuresValidationFailure(t *testing.T) {
	t.Parallel()

	uc, telemetry := newInviteUsecase(
		adminmocks.NewMockTaskQueue(),
		adminmocks.NewMockAdminIdentityPort(),
		adminmocks.NewMockAdminInvitationPort(),
		adminmocks.NewMockAdminDepartmentPort(),
		adminmocks.NewMockTransactionManager(),
	)

	_, err := uc.Execute(context.Background(), adminin.InviteMemberCommand{InviterID: adminmocks.DefaultInviterID, Email: "not-an-email"})
	if err != shareddomain.ErrInvalidEmailFormat {
		t.Fatalf("expected ErrInvalidEmailFormat, got: %v", err)
	}

	calls := telemetry.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected the operation to be measured once, got %d measurements", len(calls))
	}
	if calls[0].Operation != adminout.OperationInviteMember {
		t.Errorf("expected operation %q, got %q", adminout.OperationInviteMember, calls[0].Operation)
	}
	if !errors.Is(calls[0].Err, shareddomain.ErrInvalidEmailFormat) {
		t.Errorf("expected the measured error to be ErrInvalidEmailFormat, got: %v", calls[0].Err)
	}
}

func Test_InviteMemberUsecase_MeasuresBusinessFailure(t *testing.T) {
	t.Parallel()

	department := adminmocks.NewMockAdminDepartmentPort()
	department.GetDepartmentByIDFunc = func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
		return nil, nil
	}

	uc, telemetry := newInviteUsecase(
		adminmocks.NewMockTaskQueue(),
		adminmocks.NewMockAdminIdentityPort(),
		adminmocks.NewMockAdminInvitationPort(),
		department,
		adminmocks.NewMockTransactionManager(),
	)

	_, err := uc.Execute(context.Background(), inviteCmd())
	if err != ErrDepartmentNotFound {
		t.Fatalf("expected ErrDepartmentNotFound, got: %v", err)
	}

	calls := telemetry.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected the operation to be measured once, got %d measurements", len(calls))
	}
	if calls[0].Operation != adminout.OperationInviteMember {
		t.Errorf("expected operation %q, got %q", adminout.OperationInviteMember, calls[0].Operation)
	}
	if !errors.Is(calls[0].Err, ErrDepartmentNotFound) {
		t.Errorf("expected the measured error to be ErrDepartmentNotFound, got: %v", calls[0].Err)
	}
}

func Test_InviteMemberUsecase_MeasurementKeepsTheContext(t *testing.T) {
	t.Parallel()

	type contextKey string
	const key contextKey = "trace"

	var seen any
	department := adminmocks.NewMockAdminDepartmentPort()
	department.GetDepartmentByIDFunc = func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
		seen = ctx.Value(key)
		return &adminout.DepartmentResponse{ID: departmentID, Name: adminmocks.DefaultDepartmentName}, nil
	}

	uc, _ := newInviteUsecase(
		adminmocks.NewMockTaskQueue(),
		adminmocks.NewMockAdminIdentityPort(),
		adminmocks.NewMockAdminInvitationPort(),
		department,
		adminmocks.NewMockTransactionManager(),
	)

	ctx := context.WithValue(context.Background(), key, "span-context")
	if _, err := uc.Execute(ctx, inviteCmd()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if seen != "span-context" {
		t.Errorf("expected the measured operation to receive the original context, got %v", seen)
	}
}

func Test_InviteMemberUsecase_InstrumentationFailureDoesNotChangeTheOutcome(t *testing.T) {
	t.Parallel()

	telemetry := adminmocks.NewMockTelemetry()
	telemetry.MeasureFunc = func(ctx context.Context, operation string, fn func(context.Context) error) error {
		_ = fn(ctx)
		return errors.New("telemetry unavailable")
	}

	uc := NewInviteMemberUsecase(
		adminmocks.NewMockTaskQueue(),
		adminmocks.NewMockAdminIdentityPort(),
		adminmocks.NewMockAdminInvitationPort(),
		adminmocks.NewMockAdminDepartmentPort(),
		adminmocks.NewMockTransactionManager(),
		telemetry,
	)

	res, err := uc.Execute(context.Background(), inviteCmd())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if res == nil || res.InvitationID != adminmocks.DefaultInvitationID {
		t.Errorf("expected the invitation id %q, got %+v", adminmocks.DefaultInvitationID, res)
	}
}

func Test_InviteMemberUsecase_InstrumentationFailureDoesNotMaskTheBusinessError(t *testing.T) {
	t.Parallel()

	lookupErr := errors.New("department lookup failed")
	department := adminmocks.NewMockAdminDepartmentPort()
	department.GetDepartmentByIDFunc = func(ctx context.Context, departmentID string) (*adminout.DepartmentResponse, error) {
		return nil, lookupErr
	}

	telemetry := adminmocks.NewMockTelemetry()
	telemetry.MeasureFunc = func(ctx context.Context, operation string, fn func(context.Context) error) error {
		_ = fn(ctx)
		return errors.New("telemetry unavailable")
	}

	uc := NewInviteMemberUsecase(
		adminmocks.NewMockTaskQueue(),
		adminmocks.NewMockAdminIdentityPort(),
		adminmocks.NewMockAdminInvitationPort(),
		department,
		adminmocks.NewMockTransactionManager(),
		telemetry,
	)

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, lookupErr) {
		t.Fatalf("expected lookupErr, got: %v", err)
	}
}
