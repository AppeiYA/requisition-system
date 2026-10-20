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
) *InviteMemberUsecase {
	return NewInviteMemberUsecase(taskQueue, identity, invitations, department, tx)
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
	invitations.CheckActiveInvitationExistsFunc = func(ctx context.Context, email *shareddomain.Email) (bool, error) {
		return false, nil
	}
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

	uc := newInviteUsecase(queue, identity, invitations, department, tx)

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
}

func Test_InviteMemberUsecase_InvalidEmail(t *testing.T) {
	t.Parallel()

	uc := newInviteUsecase(
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

	uc := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), invitations, department, adminmocks.NewMockTransactionManager())

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

	uc := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), invitations, department, adminmocks.NewMockTransactionManager())

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

	uc := newInviteUsecase(adminmocks.NewMockTaskQueue(), identity, adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if err != ErrUserAlreadyExists {
		t.Fatalf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func Test_InviteMemberUsecase_PendingInvitation(t *testing.T) {
	t.Parallel()

	invitations := adminmocks.NewMockAdminInvitationPort()
	invitations.CheckActiveInvitationExistsFunc = func(ctx context.Context, email *shareddomain.Email) (bool, error) {
		return true, nil
	}

	uc := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), invitations, adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if err != ErrPendingInvitation {
		t.Fatalf("expected ErrPendingInvitation, got: %v", err)
	}
}

func Test_InviteMemberUsecase_IdentityCheckError(t *testing.T) {
	t.Parallel()

	checkErr := errors.New("identity check failed")
	identity := adminmocks.NewMockAdminIdentityPort()
	identity.CheckUserExistsFunc = func(ctx context.Context, email *shareddomain.Email) (bool, error) {
		return false, checkErr
	}

	uc := newInviteUsecase(adminmocks.NewMockTaskQueue(), identity, adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, checkErr) {
		t.Fatalf("expected checkErr, got: %v", err)
	}
}

func Test_InviteMemberUsecase_InvitationCheckError(t *testing.T) {
	t.Parallel()

	checkErr := errors.New("invitation check failed")
	invitations := adminmocks.NewMockAdminInvitationPort()
	invitations.CheckActiveInvitationExistsFunc = func(ctx context.Context, email *shareddomain.Email) (bool, error) {
		return false, checkErr
	}

	uc := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), invitations, adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

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

	uc := newInviteUsecase(queue, adminmocks.NewMockAdminIdentityPort(), invitations, adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

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

	uc := newInviteUsecase(adminmocks.NewMockTaskQueue(), adminmocks.NewMockAdminIdentityPort(), adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), tx)

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

	uc := newInviteUsecase(queue, adminmocks.NewMockAdminIdentityPort(), adminmocks.NewMockAdminInvitationPort(), adminmocks.NewMockAdminDepartmentPort(), adminmocks.NewMockTransactionManager())

	_, err := uc.Execute(context.Background(), inviteCmd())
	if !errors.Is(err, enqueueErr) {
		t.Fatalf("expected enqueueErr, got: %v", err)
	}
}

func Test_InviteMemberUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := newInviteUsecase(
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
}
