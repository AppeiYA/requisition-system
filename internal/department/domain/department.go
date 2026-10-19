package domain

import (
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var DepartmentIDPrefix = "dpt"

type Department struct {
	id          string
	name        *DepartmentName
	description *DepartmentDescription
	code        *DepartmentCode
	status      DepartmentStatus
	updatedAt   *time.Time
	createdAt   time.Time
}

var (
	ErrInvalidDepartmentID = customerrors.NewInternalServerError("Invalid department ID")
)

func NewDepartment(id string, name *DepartmentName, description *DepartmentDescription, code *DepartmentCode) (*Department, error) {
	if id == "" {
		return nil, ErrInvalidDepartmentID
	}

	status := DepartmentStatusActive

	return &Department{
		id:          id,
		name:        name,
		description: description,
		code:        code,
		status:      status,
		createdAt:   time.Now(),
	}, nil
}

// ReconstituteDepartment rebuilds a Department from persisted data, restoring
// the stored status and timestamps instead of deriving them.
func ReconstituteDepartment(
	id string,
	name *DepartmentName,
	description *DepartmentDescription,
	code *DepartmentCode,
	status DepartmentStatus,
	updatedAt *time.Time,
	createdAt time.Time,
) (*Department, error) {
	if id == "" {
		return nil, ErrInvalidDepartmentID
	}
	if !status.IsValid() {
		return nil, ErrInvalidDepartmentStatus
	}

	return &Department{
		id:          id,
		name:        name,
		description: description,
		code:        code,
		status:      status,
		updatedAt:   updatedAt,
		createdAt:   createdAt,
	}, nil
}

func (d *Department) ID() string {
	return d.id
}

func (d *Department) Name() *DepartmentName {
	return d.name
}

func (d *Department) Description() *DepartmentDescription {
	return d.description
}

func (d *Department) Code() *DepartmentCode {
	return d.code
}

func (d *Department) Status() DepartmentStatus {
	return d.status
}

func (d *Department) UpdatedAt() *time.Time {
	return d.updatedAt
}

func (d *Department) CreatedAt() time.Time {
	return d.createdAt
}
