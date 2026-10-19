package departmentmocks

import (
	"context"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
)

var (
	DefaultDepartmentID          = "dpt-1234"
	DefaultDepartmentName        = "Finance"
	DefaultDepartmentDescription = "Handles payments and budgets"
	DefaultDepartmentCode        = "FIN"
	DefaultUserID                = "usr-5678"
	DefaultJobTitle              = "Financial Analyst"
)

// ValidDepartment returns a minimal active department.
func ValidDepartment() *departmentdomain.Department {
	name, _ := departmentdomain.NewDepartmentName(DefaultDepartmentName)
	description, _ := departmentdomain.NewDepartmentDescription(DefaultDepartmentDescription)
	code, _ := departmentdomain.NewDepartmentCode(DefaultDepartmentCode)

	department, _ := departmentdomain.NewDepartment(DefaultDepartmentID, name, description, code)
	return department
}

// ValidMembership returns a minimal active department membership.
func ValidMembership() *departmentdomain.DepartmentMembership {
	jobTitle, _ := departmentdomain.NewMemberJobTitle(DefaultJobTitle)

	membership, _ := departmentdomain.NewDepartmentMembership(DefaultUserID, DefaultDepartmentID, jobTitle)
	return membership
}

type MockDepartmentRepository struct {
	SaveDepartmentFunc    func(ctx context.Context, department *departmentdomain.Department) error
	GetDepartmentByIDFunc func(ctx context.Context, id string) (*departmentdomain.Department, error)
}

func NewMockDepartmentRepository() *MockDepartmentRepository {
	return &MockDepartmentRepository{
		SaveDepartmentFunc: func(ctx context.Context, department *departmentdomain.Department) error {
			return nil
		},
		GetDepartmentByIDFunc: func(ctx context.Context, id string) (*departmentdomain.Department, error) {
			return ValidDepartment(), nil
		},
	}
}

func (m *MockDepartmentRepository) SaveDepartment(ctx context.Context, department *departmentdomain.Department) error {
	return m.SaveDepartmentFunc(ctx, department)
}

func (m *MockDepartmentRepository) GetDepartmentByID(ctx context.Context, id string) (*departmentdomain.Department, error) {
	return m.GetDepartmentByIDFunc(ctx, id)
}

var _ departmentout.DepartmentRepository = (*MockDepartmentRepository)(nil)

type MockMembershipRepository struct {
	SaveMembershipFunc func(ctx context.Context, membership *departmentdomain.DepartmentMembership) error
}

func NewMockMembershipRepository() *MockMembershipRepository {
	return &MockMembershipRepository{
		SaveMembershipFunc: func(ctx context.Context, membership *departmentdomain.DepartmentMembership) error {
			return nil
		},
	}
}

func (m *MockMembershipRepository) SaveMembership(ctx context.Context, membership *departmentdomain.DepartmentMembership) error {
	return m.SaveMembershipFunc(ctx, membership)
}

var _ departmentout.MembershipRepository = (*MockMembershipRepository)(nil)

type MockIDGenerator struct {
	GenerateIDFunc func(prefix string) (string, error)
}

func NewMockIDGenerator() *MockIDGenerator {
	return &MockIDGenerator{
		GenerateIDFunc: func(prefix string) (string, error) {
			return prefix + "123", nil
		},
	}
}

func (m *MockIDGenerator) GenerateID(prefix string) (string, error) {
	return m.GenerateIDFunc(prefix)
}

var _ departmentout.IDGenerator = (*MockIDGenerator)(nil)
