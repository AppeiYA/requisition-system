package models

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

type UserModel struct {
	ID           string     `db:"id"`
	FirstName    string     `db:"first_name"`
	LastName     string     `db:"last_name"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	Status       string     `db:"status"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

func (m UserModel) ToDomain() (*domain.User, error) {
	return domain.ReconstituteUser(
		m.ID,
		m.FirstName,
		m.LastName,
		m.Email,
		m.PasswordHash,
		m.Status,
		m.CreatedAt,
		m.UpdatedAt,
	)
}

func FromDomain(user *domain.User) *UserModel {
	return &UserModel{
		ID:           user.ID(),
		FirstName:    user.FirstName().Value(),
		LastName:     user.LastName().Value(),
		Email:        user.Email().Value(),
		PasswordHash: user.PasswordHash().Value(),
		Status:       user.Status().Value(),
		CreatedAt:    user.CreatedAt(),
		UpdatedAt:    user.UpdatedAt(),
	}
}
