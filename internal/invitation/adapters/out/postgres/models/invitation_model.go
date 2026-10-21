package models

import (
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
)

type InvitationModel struct {
	ID           string     `db:"id"`
	FirstName    string     `db:"first_name"`
	LastName     string     `db:"last_name"`
	PhoneNumber  *string    `db:"phone_number"`
	StaffID      *string    `db:"staff_id"`
	DepartmentID string     `db:"department_id"`
	JobTitle     string     `db:"job_title"`
	Email        string     `db:"email"`
	TokenHash    string     `db:"token_hash"`
	InvitedBy    string     `db:"invited_by"`
	Status       string     `db:"status"`
	ExpiresAt    time.Time  `db:"expires_at"`
	AcceptedAt   *time.Time `db:"accepted_at"`
	RevokedAt    *time.Time `db:"revoked_at"`
	CreatedAt    time.Time  `db:"created_at"`
}

func FromDomain(inv *invitationdomain.Invitation) *InvitationModel {
	var phone *string
	if value := inv.Phone().Value(); value != "" {
		phone = &value
	}

	var staffID *string
	if value := inv.StaffID(); value != "" {
		staffID = &value
	}

	return &InvitationModel{
		ID:           inv.ID(),
		FirstName:    inv.FirstName().Value(),
		LastName:     inv.LastName().Value(),
		PhoneNumber:  phone,
		StaffID:      staffID,
		DepartmentID: inv.DepartmentID(),
		JobTitle:     inv.JobTitle(),
		Email:        inv.Email().Value(),
		TokenHash:    inv.TokenHash(),
		InvitedBy:    inv.InvitedBy(),
		Status:       inv.Status().String(),
		ExpiresAt:    inv.ExpiresAt(),
		AcceptedAt:   inv.AcceptedAt(),
		RevokedAt:    inv.RevokedAt(),
		CreatedAt:    inv.CreatedAt(),
	}
}

func (m InvitationModel) ToDomain() (*invitationdomain.Invitation, error) {
	var phone string
	if m.PhoneNumber != nil {
		phone = *m.PhoneNumber
	}

	var staffID string
	if m.StaffID != nil {
		staffID = *m.StaffID
	}

	return invitationdomain.ReconstituteInvitationWithStatus(
		m.ID,
		m.FirstName,
		m.LastName,
		phone,
		staffID,
		m.DepartmentID,
		m.JobTitle,
		m.Email,
		m.TokenHash,
		m.InvitedBy,
		m.Status,
		m.ExpiresAt,
		m.AcceptedAt,
		m.RevokedAt,
		m.CreatedAt,
	)
}
