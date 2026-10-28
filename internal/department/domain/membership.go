package domain

import (
	"strings"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var UserIDPrefix = "usr"

type DepartmentMembership struct {
	userID       string
	departmentID string
	jobTitle     *MemberJobTitle
	status       MembershipStatus
	createdAt    time.Time
}

var (
	ErrInvalidUserID = customerrors.NewInternalServerError("Invalid user ID")
)

func NewDepartmentMembership(userID, departmentID string, jobTitle *MemberJobTitle) (*DepartmentMembership, error) {
	if userID == "" || !strings.HasPrefix(userID, UserIDPrefix) {
		return nil, ErrInvalidUserID
	}
	if departmentID == "" || !strings.HasPrefix(departmentID, DepartmentIDPrefix) {
		return nil, ErrInvalidDepartmentID
	}

	status, err := NewMembershipStatus("active")
	if err != nil {
		return nil, err
	}

	return &DepartmentMembership{
		userID:       userID,
		departmentID: departmentID,
		jobTitle:     jobTitle,
		status:       status,
		createdAt:    time.Now(),
	}, nil
}

func ReconstituteDepartmentMembership(userID, departmentID, jobTitle, status string, createdAt time.Time) (*DepartmentMembership, error) {
	if userID == "" || !strings.HasPrefix(userID, UserIDPrefix) {
		return nil, ErrInvalidUserID
	}
	if departmentID == "" || !strings.HasPrefix(departmentID, DepartmentIDPrefix) {
		return nil, ErrInvalidDepartmentID
	}

	validJobTitle, err := NewMemberJobTitle(jobTitle)
	if err != nil {
		return nil, err
	}
	validStatus, err := NewMembershipStatus(status)
	if err != nil {
		return nil, err
	}

	return &DepartmentMembership{
		userID:       userID,
		departmentID: departmentID,
		jobTitle:     validJobTitle,
		status:       validStatus,
		createdAt:    createdAt,
	}, nil
}

func (dm *DepartmentMembership) UserID() string {
	return dm.userID
}

func (dm *DepartmentMembership) DepartmentID() string {
	return dm.departmentID
}

func (dm *DepartmentMembership) JobTitle() *MemberJobTitle {
	return dm.jobTitle
}

func (dm *DepartmentMembership) Status() MembershipStatus {
	return dm.status
}

func (dm *DepartmentMembership) CreatedAt() time.Time {
	return dm.createdAt
}
