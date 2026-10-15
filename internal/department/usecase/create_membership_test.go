package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentmocks "github.com/AppeiYA/requisition-system/internal/department/mocks"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
)

func membershipCmd() departmentin.CreateMembershipCommand {
	return departmentin.CreateMembershipCommand{
		UserID:       departmentmocks.DefaultUserID,
		DepartmentID: departmentmocks.DefaultDepartmentID,
		JobTitle:     departmentmocks.DefaultJobTitle,
	}
}

func Test_CreateMembershipUsecase_Success(t *testing.T) {
	t.Parallel()

	var saved *domain.DepartmentMembership
	repo := departmentmocks.NewMockMembershipRepository()
	repo.SaveMembershipFunc = func(ctx context.Context, membership *domain.DepartmentMembership) error {
		saved = membership
		return nil
	}

	uc := NewCreateMembershipUsecase(repo)

	if err := uc.Execute(context.Background(), membershipCmd()); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if saved == nil {
		t.Fatal("expected membership to be saved")
	}
	if saved.UserID() != departmentmocks.DefaultUserID {
		t.Errorf("expected user id %q, got %q", departmentmocks.DefaultUserID, saved.UserID())
	}
	if saved.DepartmentID() != departmentmocks.DefaultDepartmentID {
		t.Errorf("expected department id %q, got %q", departmentmocks.DefaultDepartmentID, saved.DepartmentID())
	}
	if saved.JobTitle().Value() != departmentmocks.DefaultJobTitle {
		t.Errorf("expected job title %q, got %q", departmentmocks.DefaultJobTitle, saved.JobTitle().Value())
	}
	if saved.Status() != domain.MembershipStatusActive {
		t.Errorf("expected status %q, got %q", domain.MembershipStatusActive, saved.Status())
	}
}

func Test_CreateMembershipUsecase_EmptyJobTitle(t *testing.T) {
	t.Parallel()

	saveCalled := false
	repo := departmentmocks.NewMockMembershipRepository()
	repo.SaveMembershipFunc = func(ctx context.Context, membership *domain.DepartmentMembership) error {
		saveCalled = true
		return nil
	}

	uc := NewCreateMembershipUsecase(repo)

	cmd := membershipCmd()
	cmd.JobTitle = ""

	err := uc.Execute(context.Background(), cmd)
	if err != domain.ErrEmptyJobTitle {
		t.Fatalf("expected ErrEmptyJobTitle, got: %v", err)
	}
	if saveCalled {
		t.Error("membership must not be saved with an invalid job title")
	}
}

func Test_CreateMembershipUsecase_JobTitleTooLong(t *testing.T) {
	t.Parallel()

	uc := NewCreateMembershipUsecase(departmentmocks.NewMockMembershipRepository())

	cmd := membershipCmd()
	cmd.JobTitle = ""
	for i := 0; i < 61; i++ {
		cmd.JobTitle += "a"
	}

	err := uc.Execute(context.Background(), cmd)
	if err != domain.ErrJobTitleTooLong {
		t.Fatalf("expected ErrJobTitleTooLong, got: %v", err)
	}
}

func Test_CreateMembershipUsecase_InvalidUserID(t *testing.T) {
	t.Parallel()

	uc := NewCreateMembershipUsecase(departmentmocks.NewMockMembershipRepository())

	cmd := membershipCmd()
	cmd.UserID = "not-a-user-id"

	err := uc.Execute(context.Background(), cmd)
	if err != domain.ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got: %v", err)
	}
}

func Test_CreateMembershipUsecase_InvalidDepartmentID(t *testing.T) {
	t.Parallel()

	uc := NewCreateMembershipUsecase(departmentmocks.NewMockMembershipRepository())

	cmd := membershipCmd()
	cmd.DepartmentID = "not-a-department-id"

	err := uc.Execute(context.Background(), cmd)
	if err != domain.ErrInvalidDepartmentID {
		t.Fatalf("expected ErrInvalidDepartmentID, got: %v", err)
	}
}

func Test_CreateMembershipUsecase_SaveError(t *testing.T) {
	t.Parallel()

	saveErr := errors.New("save failed")
	repo := departmentmocks.NewMockMembershipRepository()
	repo.SaveMembershipFunc = func(ctx context.Context, membership *domain.DepartmentMembership) error {
		return saveErr
	}

	uc := NewCreateMembershipUsecase(repo)

	err := uc.Execute(context.Background(), membershipCmd())
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
}

func Test_CreateMembershipUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewCreateMembershipUsecase(departmentmocks.NewMockMembershipRepository())

	err := uc.Execute(ctx, membershipCmd())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}
