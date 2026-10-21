package adminexternal

import (
	"context"
	"errors"
	"testing"

	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type stubGetDepartmentByID struct {
	res   *departmentin.DepartmentResult
	err   error
	gotID string
}

func (s *stubGetDepartmentByID) Execute(ctx context.Context, id string) (*departmentin.DepartmentResult, error) {
	s.gotID = id
	return s.res, s.err
}

func Test_AdminDepartmentAdapter_CheckDepartmentExists_True(t *testing.T) {
	t.Parallel()

	get := &stubGetDepartmentByID{res: &departmentin.DepartmentResult{ID: "dpt-1"}}
	adapter := NewAdminDepartmentAdapter(get)

	exists, err := adapter.CheckDepartmentExists(context.Background(), "dpt-1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !exists {
		t.Fatal("expected department to exist")
	}
	if get.gotID != "dpt-1" {
		t.Errorf("expected lookup of %q, got %q", "dpt-1", get.gotID)
	}
}

func Test_AdminDepartmentAdapter_CheckDepartmentExists_NotFound(t *testing.T) {
	t.Parallel()

	adapter := NewAdminDepartmentAdapter(&stubGetDepartmentByID{err: customerrors.NewNotFoundError("department not found")})

	exists, err := adapter.CheckDepartmentExists(context.Background(), "dpt-missing")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if exists {
		t.Fatal("expected department not to exist")
	}
}

func Test_AdminDepartmentAdapter_CheckDepartmentExists_Error(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db down")
	adapter := NewAdminDepartmentAdapter(&stubGetDepartmentByID{err: dbErr})

	_, err := adapter.CheckDepartmentExists(context.Background(), "dpt-1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}

func Test_AdminDepartmentAdapter_GetDepartmentName(t *testing.T) {
	t.Parallel()

	get := &stubGetDepartmentByID{res: &departmentin.DepartmentResult{ID: "dpt-1", Name: "Finance"}}
	adapter := NewAdminDepartmentAdapter(get)

	name, err := adapter.GetDepartmentName(context.Background(), "dpt-1")
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if name != "Finance" {
		t.Errorf("expected name %q, got %q", "Finance", name)
	}
	if get.gotID != "dpt-1" {
		t.Errorf("expected lookup of %q, got %q", "dpt-1", get.gotID)
	}
}

func Test_AdminDepartmentAdapter_GetDepartmentName_Error(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("db down")
	adapter := NewAdminDepartmentAdapter(&stubGetDepartmentByID{err: dbErr})

	_, err := adapter.GetDepartmentName(context.Background(), "dpt-1")
	if !errors.Is(err, dbErr) {
		t.Fatalf("expected dbErr, got: %v", err)
	}
}
