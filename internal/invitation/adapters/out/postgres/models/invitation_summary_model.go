package models

import (
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
)

// InvitationSummaryModel is the row shape returned by the admin invitation
// queries. The joins supply the department and inviter names.
type InvitationSummaryModel struct {
	ID               string     `db:"id"`
	FirstName        string     `db:"first_name"`
	LastName         string     `db:"last_name"`
	Email            string     `db:"email"`
	DepartmentID     string     `db:"department_id"`
	DepartmentName   string     `db:"department_name"`
	JobTitle         string     `db:"job_title"`
	Status           string     `db:"status"`
	InvitedBy        string     `db:"invited_by"`
	InviterName      string     `db:"inviter_name"`
	ExpiresAt        time.Time  `db:"expires_at"`
	AcceptedAt       *time.Time `db:"accepted_at"`
	RevokedAt        *time.Time `db:"revoked_at"`
	DeliveredAt      *time.Time `db:"delivered_at"`
	DeliveryFailedAt *time.Time `db:"delivery_failed_at"`
	CreatedAt        time.Time  `db:"created_at"`
}

func (m InvitationSummaryModel) ToDomain() invitationdomain.InvitationSummary {
	return invitationdomain.InvitationSummary{
		ID:               m.ID,
		FirstName:        m.FirstName,
		LastName:         m.LastName,
		Email:            m.Email,
		DepartmentID:     m.DepartmentID,
		DepartmentName:   m.DepartmentName,
		JobTitle:         m.JobTitle,
		Status:           m.Status,
		InvitedBy:        m.InvitedBy,
		InviterName:      m.InviterName,
		ExpiresAt:        m.ExpiresAt,
		AcceptedAt:       m.AcceptedAt,
		RevokedAt:        m.RevokedAt,
		DeliveredAt:      m.DeliveredAt,
		DeliveryFailedAt: m.DeliveryFailedAt,
		CreatedAt:        m.CreatedAt,
	}
}
