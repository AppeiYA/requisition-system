package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentmocks "github.com/AppeiYA/requisition-system/internal/department/mocks"
)

func Test_GetDepartmentsUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := departmentmocks.NewMockDepartmentRepository()
	repo.GetDepartmentsFunc = func(ctx context.Context) ([]*domain.Department, error) {
		return []*domain.Department{departmentmocks.ValidDepartment()}, nil
	}

	uc := NewGetDepartmentsUsecase(repo)

	res, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 department, got: %d", len(res))
	}
	if res[0].ID() != departmentmocks.DefaultDepartmentID {
		t.Errorf("expected id %q, got %q", departmentmocks.DefaultDepartmentID, res[0].ID())
	}
	if res[0].Name().Value() != departmentmocks.DefaultDepartmentName {
		t.Errorf("expected name %q, got %q", departmentmocks.DefaultDepartmentName, res[0].Name().Value())
	}
	if res[0].Code().Value() != departmentmocks.DefaultDepartmentCode {
		t.Errorf("expected code %q, got %q", departmentmocks.DefaultDepartmentCode, res[0].Code().Value())
	}
	if res[0].Status().String() != string(domain.DepartmentStatusActive) {
		t.Errorf("expected status %q, got %q", domain.DepartmentStatusActive, res[0].Status().String())
	}
}

func Test_GetDepartmentsUsecase_EmptyResult(t *testing.T) {
	t.Parallel()

	repo := departmentmocks.NewMockDepartmentRepository()
	repo.GetDepartmentsFunc = func(ctx context.Context) ([]*domain.Department, error) {
		return []*domain.Department{}, nil
	}

	uc := NewGetDepartmentsUsecase(repo)

	res, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected 0 departments, got: %d", len(res))
	}
}

func Test_GetDepartmentsUsecase_RepoError(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db down")
	repo := departmentmocks.NewMockDepartmentRepository()
	repo.GetDepartmentsFunc = func(ctx context.Context) ([]*domain.Department, error) {
		return nil, dbErr
	}

	uc := NewGetDepartmentsUsecase(repo)

	_, err := uc.Execute(context.Background())
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func Test_GetDepartmentsUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewGetDepartmentsUsecase(departmentmocks.NewMockDepartmentRepository())

	_, err := uc.Execute(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}
