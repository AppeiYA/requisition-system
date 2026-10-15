package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentmocks "github.com/AppeiYA/requisition-system/internal/department/mocks"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func Test_GetDepartmentByIDUsecase_Success(t *testing.T) {
	t.Parallel()

	var gotID string
	repo := departmentmocks.NewMockDepartmentRepository()
	repo.GetDepartmentByIDFunc = func(ctx context.Context, id string) (*domain.Department, error) {
		gotID = id
		return departmentmocks.ValidDepartment(), nil
	}

	uc := NewGetDepartmentByIDUsecase(repo)

	res, err := uc.Execute(context.Background(), departmentmocks.DefaultDepartmentID)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if gotID != departmentmocks.DefaultDepartmentID {
		t.Errorf("expected lookup of %q, got %q", departmentmocks.DefaultDepartmentID, gotID)
	}
	if res.ID != departmentmocks.DefaultDepartmentID {
		t.Errorf("expected id %q, got %q", departmentmocks.DefaultDepartmentID, res.ID)
	}
	if res.Name != departmentmocks.DefaultDepartmentName {
		t.Errorf("expected name %q, got %q", departmentmocks.DefaultDepartmentName, res.Name)
	}
	if res.Code != departmentmocks.DefaultDepartmentCode {
		t.Errorf("expected code %q, got %q", departmentmocks.DefaultDepartmentCode, res.Code)
	}
	if res.Status != string(domain.DepartmentStatusActive) {
		t.Errorf("expected status %q, got %q", domain.DepartmentStatusActive, res.Status)
	}
}

func Test_GetDepartmentByIDUsecase_EmptyID(t *testing.T) {
	t.Parallel()

	uc := NewGetDepartmentByIDUsecase(departmentmocks.NewMockDepartmentRepository())

	_, err := uc.Execute(context.Background(), "")
	if err != ErrDepartmentIDRequired {
		t.Fatalf("expected ErrDepartmentIDRequired, got: %v", err)
	}
}

func Test_GetDepartmentByIDUsecase_NotFound(t *testing.T) {
	t.Parallel()

	repo := departmentmocks.NewMockDepartmentRepository()
	repo.GetDepartmentByIDFunc = func(ctx context.Context, id string) (*domain.Department, error) {
		return nil, customerrors.NewNotFoundError("department not found")
	}

	uc := NewGetDepartmentByIDUsecase(repo)

	_, err := uc.Execute(context.Background(), "dpt-missing")
	if err != ErrDepartmentNotFound {
		t.Fatalf("expected ErrDepartmentNotFound, got: %v", err)
	}
}

func Test_GetDepartmentByIDUsecase_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db down")
	repo := departmentmocks.NewMockDepartmentRepository()
	repo.GetDepartmentByIDFunc = func(ctx context.Context, id string) (*domain.Department, error) {
		return nil, dbErr
	}

	uc := NewGetDepartmentByIDUsecase(repo)

	_, err := uc.Execute(context.Background(), departmentmocks.DefaultDepartmentID)
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func Test_GetDepartmentByIDUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewGetDepartmentByIDUsecase(departmentmocks.NewMockDepartmentRepository())

	_, err := uc.Execute(ctx, departmentmocks.DefaultDepartmentID)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}
