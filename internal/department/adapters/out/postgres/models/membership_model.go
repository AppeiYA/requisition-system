package models

import (
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
)

type MembershipModel struct {
	UserID       string    `db:"user_id"`
	DepartmentID string    `db:"department_id"`
	JobTitle     string    `db:"job_title"`
	Status       string    `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
}

func MembershipFromDomain(membership *departmentdomain.DepartmentMembership) *MembershipModel {
	return &MembershipModel{
		UserID:       membership.UserID(),
		DepartmentID: membership.DepartmentID(),
		JobTitle:     membership.JobTitle().Value(),
		Status:       membership.Status().String(),
		CreatedAt:    membership.CreatedAt(),
	}
}

func (m MembershipModel) ToDomain() (*departmentdomain.DepartmentMembership, error) {
	return departmentdomain.ReconstituteDepartmentMembership(
		m.UserID,
		m.DepartmentID,
		m.JobTitle,
		m.Status,
		m.CreatedAt,
	)
}
