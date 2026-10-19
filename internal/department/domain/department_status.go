package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type DepartmentStatus string 
const (
	DepartmentStatusActive DepartmentStatus = "active"
	DepartmentStatusInactive DepartmentStatus = "inactive"
)

var (
	ErrInvalidDepartmentStatus = customerrors.NewBadRequestError("Invalid department status")
)

func (ds DepartmentStatus) IsValid() bool {
	switch ds {
	case DepartmentStatusActive, DepartmentStatusInactive:
		return true
	default:
		return false
	}
}

func NewDepartmentStatus(value string) (DepartmentStatus, error) {
	status := DepartmentStatus(value)
	if !status.IsValid() {
		return "", ErrInvalidDepartmentStatus
	}
	return status, nil
}

func (ds DepartmentStatus) String() string {
	return string(ds)
}