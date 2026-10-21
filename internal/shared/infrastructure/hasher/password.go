package hasher

import (
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordMismatch = customerrors.NewUnauthorizedError("Account and password credentials do not match")
)

type PasswordHasher struct {
	cost int
}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{cost: bcrypt.DefaultCost}
}

func NewPasswordHasherWithCost(cost int) *PasswordHasher {
	return &PasswordHasher{cost: cost}
}

func (h *PasswordHasher) HashPassword(password string) (string, error) {
	cost := h.cost
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *PasswordHasher) ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	switch err {
	case nil:
		return nil
	case bcrypt.ErrMismatchedHashAndPassword:
		return ErrPasswordMismatch
	default:
		return err
	}
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	return h.HashPassword(password)
}

func (h *PasswordHasher) Compare(hashedPassword, password string) error {
	return h.ComparePassword(hashedPassword, password)
}
