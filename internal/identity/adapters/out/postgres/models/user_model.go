package models

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

type UserModel struct {
	ID           string     `db:"id"`
	FirstName    string     `db:"first_name"`
	LastName     string     `db:"last_name"`
	PhoneNumber  *string    `db:"phone_number"`
	StaffID      *string    `db:"staff_id"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Status       string     `db:"status"`
	Role		 string     `db:"role"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

func (m UserModel) ToDomain() (*domain.User, error) {
	var phone string
	if m.PhoneNumber != nil {
		phone = *m.PhoneNumber
	}

	var staffID string
	if m.StaffID != nil {
		staffID = *m.StaffID
	}

	return domain.ReconstituteUser(
		m.ID,
		m.FirstName,
		m.LastName,
		phone,
		staffID,
		m.Email,
		m.PasswordHash,
		m.Status,
		m.Role,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func FromDomain(user *domain.User) *UserModel {
	var phone *string
	if value := user.Phone().Value(); value != "" {
		phone = &value
	}

	var staffID *string
	if value := user.StaffID(); value != "" {
		staffID = &value
	}

	return &UserModel{
		ID:           user.ID(),
		FirstName:    user.FirstName().Value(),
		LastName:     user.LastName().Value(),
		PhoneNumber:  phone,
		StaffID:      staffID,
		Email:        user.Email().Value(),
		PasswordHash: user.PasswordHash().Value(),
		Status:       user.Status().Value(),
		Role:         user.Role().String(),
		CreatedAt:    user.CreatedAt(),
		UpdatedAt:    user.UpdatedAt(),
	}
}
