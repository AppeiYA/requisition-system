package models

import (
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
)

type DepartmentModel struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Code        string     `db:"code"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

func DepartmentFromDomain(department *departmentdomain.Department) *DepartmentModel {
	return &DepartmentModel{
		ID:          department.ID(),
		Name:        department.Name().Value(),
		Description: department.Description().Value(),
		Code:        department.Code().Value(),
		Status:      department.Status().String(),
		CreatedAt:   department.CreatedAt(),
		UpdatedAt:   department.UpdatedAt(),
	}
}

func (m DepartmentModel) ToDomain() (*departmentdomain.Department, error) {
	name, err := departmentdomain.NewDepartmentName(m.Name)
	if err != nil {
		return nil, err
	}

	description, err := departmentdomain.NewDepartmentDescription(m.Description)
	if err != nil {
		return nil, err
	}

	code, err := departmentdomain.NewDepartmentCode(m.Code)
	if err != nil {
		return nil, err
	}

	status, err := departmentdomain.NewDepartmentStatus(m.Status)
	if err != nil {
		return nil, err
	}

	return departmentdomain.ReconstituteDepartment(
		m.ID,
		name,
		description,
		code,
		status,
		m.UpdatedAt,
		m.CreatedAt,
	)
}
