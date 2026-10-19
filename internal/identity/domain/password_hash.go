package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

var (
	ErrInvalidPasswordHash = customerrors.NewBadRequestError("Invalid password hash")
)

type PasswordHash struct {
	value string
}

func NewPasswordHash(value string) (*PasswordHash, error) {
	if value == "" {
		return nil, ErrInvalidPasswordHash
	}
	return &PasswordHash{value: value}, nil
}

func (p *PasswordHash) Value() string {
	return p.value
}
