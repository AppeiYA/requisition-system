package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentmocks "github.com/AppeiYA/requisition-system/internal/department/mocks"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
)

func createCmd() departmentin.CreateDepartmentCommand {
	return departmentin.CreateDepartmentCommand{
		Name:        departmentmocks.DefaultDepartmentName,
		Description: departmentmocks.DefaultDepartmentDescription,
		Code:        departmentmocks.DefaultDepartmentCode,
	}
}

func Test_CreateDepartmentUsecase_Success(t *testing.T) {
	t.Parallel()

	var saved *domain.Department
	repo := departmentmocks.NewMockDepartmentRepository()
	repo.SaveDepartmentFunc = func(ctx context.Context, department *domain.Department) error {
		saved = department
		return nil
	}

	uc := NewCreateDepartmentUsecase(repo, departmentmocks.NewMockIDGenerator())

	res, err := uc.Execute(context.Background(), createCmd())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if res.ID != "dpt123" {
		t.Errorf("expected id %q, got %q", "dpt123", res.ID)
	}
	if saved == nil {
		t.Fatal("expected department to be saved")
	}
	if saved.Name().Value() != departmentmocks.DefaultDepartmentName {
		t.Errorf("expected name %q, got %q", departmentmocks.DefaultDepartmentName, saved.Name().Value())
	}
	if saved.Description().Value() != departmentmocks.DefaultDepartmentDescription {
		t.Errorf("expected description %q, got %q", departmentmocks.DefaultDepartmentDescription, saved.Description().Value())
	}
	if saved.Code().Value() != departmentmocks.DefaultDepartmentCode {
		t.Errorf("expected code %q, got %q", departmentmocks.DefaultDepartmentCode, saved.Code().Value())
	}
	if saved.Status() != domain.DepartmentStatusActive {
		t.Errorf("expected status %q, got %q", domain.DepartmentStatusActive, saved.Status())
	}
}

func Test_CreateDepartmentUsecase_NormalizesCode(t *testing.T) {
	t.Parallel()

	var saved *domain.Department
	repo := departmentmocks.NewMockDepartmentRepository()
	repo.SaveDepartmentFunc = func(ctx context.Context, department *domain.Department) error {
		saved = department
		return nil
	}

	uc := NewCreateDepartmentUsecase(repo, departmentmocks.NewMockIDGenerator())

	cmd := createCmd()
	cmd.Code = "fin"

	_, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if saved.Code().Value() != "FIN" {
		t.Errorf("expected uppercased code %q, got %q", "FIN", saved.Code().Value())
	}
}

func Test_CreateDepartmentUsecase_DomainValidation(t *testing.T) {
	t.Parallel()

	uc := NewCreateDepartmentUsecase(departmentmocks.NewMockDepartmentRepository(), departmentmocks.NewMockIDGenerator())

	tests := []struct {
		name        string
		cmd         departmentin.CreateDepartmentCommand
		expectedErr error
	}{
		{
			name:        "empty name",
			cmd:         departmentin.CreateDepartmentCommand{Name: "", Description: departmentmocks.DefaultDepartmentDescription, Code: departmentmocks.DefaultDepartmentCode},
			expectedErr: domain.ErrEmptyDepartmentName,
		},
		{
			name:        "empty description",
			cmd:         departmentin.CreateDepartmentCommand{Name: departmentmocks.DefaultDepartmentName, Description: "", Code: departmentmocks.DefaultDepartmentCode},
			expectedErr: domain.ErrEmptyDepartmentDescription,
		},
		{
			name:        "empty code",
			cmd:         departmentin.CreateDepartmentCommand{Name: departmentmocks.DefaultDepartmentName, Description: departmentmocks.DefaultDepartmentDescription, Code: ""},
			expectedErr: domain.ErrEmptyDepartmentCode,
		},
		{
			name:        "code too long",
			cmd:         departmentin.CreateDepartmentCommand{Name: departmentmocks.DefaultDepartmentName, Description: departmentmocks.DefaultDepartmentDescription, Code: "FINAN"},
			expectedErr: domain.ErrDepartmentCodeTooLong,
		},
		{
			name:        "code with digits",
			cmd:         departmentin.CreateDepartmentCommand{Name: departmentmocks.DefaultDepartmentName, Description: departmentmocks.DefaultDepartmentDescription, Code: "F1N"},
			expectedErr: domain.ErrDepartmentCodeInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := uc.Execute(context.Background(), tt.cmd)
			if err != tt.expectedErr {
				t.Fatalf("expected %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func Test_CreateDepartmentUsecase_IDGeneratorError(t *testing.T) {
	t.Parallel()

	idErr := errors.New("id generation failed")
	idGen := departmentmocks.NewMockIDGenerator()
	idGen.GenerateIDFunc = func(prefix string) (string, error) {
		return "", idErr
	}

	uc := NewCreateDepartmentUsecase(departmentmocks.NewMockDepartmentRepository(), idGen)

	_, err := uc.Execute(context.Background(), createCmd())
	if !errors.Is(err, idErr) {
		t.Fatalf("expected idErr, got: %v", err)
	}
}

func Test_CreateDepartmentUsecase_SaveError(t *testing.T) {
	t.Parallel()

	saveErr := errors.New("save failed")
	repo := departmentmocks.NewMockDepartmentRepository()
	repo.SaveDepartmentFunc = func(ctx context.Context, department *domain.Department) error {
		return saveErr
	}

	uc := NewCreateDepartmentUsecase(repo, departmentmocks.NewMockIDGenerator())

	_, err := uc.Execute(context.Background(), createCmd())
	if !errors.Is(err, saveErr) {
		t.Fatalf("expected saveErr, got: %v", err)
	}
}

func Test_CreateDepartmentUsecase_ContextCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewCreateDepartmentUsecase(departmentmocks.NewMockDepartmentRepository(), departmentmocks.NewMockIDGenerator())

	_, err := uc.Execute(ctx, createCmd())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}
