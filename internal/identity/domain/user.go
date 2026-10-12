package domain

import (
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

var UserIDPrefix = "usr"

var (
	ErrInvalidID     = customerrors.NewInternalServerError("empty id in create user")
)

type User struct {
	id string
	firstName    *Username
	lastName     *Username
	email        *Email
	passwordHash *PasswordHash
	status       Status
	createdAt    time.Time
	updatedAt    *time.Time
}

func NewUser(id, firstName, lastName string, email *Email, passwordHash *PasswordHash) (*User, error) {
	if id == "" {
		return nil, ErrInvalidID
	}
	status, err := NewStatus("active")
	if err != nil {
		return nil, err
	}
	first_name, err := NewUsername(firstName, FirstName)
	if err != nil {
		return nil, err
	}
	last_name, err := NewUsername(lastName, LastName)
	if err != nil {
		return nil, err
	}
	return &User{
		id:           id,
		firstName:    first_name,
		lastName:     last_name,
		email:        email,
		passwordHash: passwordHash,
		status:       status,
		createdAt:    time.Now(),
	}, nil
}

func ReconstituteUser(id, firstName, lastName, email, passwordHash, status string, createdAt time.Time, updatedAt *time.Time) (*User, error) {
	first_name, err := NewUsername(firstName, FirstName)
	if err != nil {
		return nil, err
	}

	last_name, err := NewUsername(lastName, LastName)
	if err != nil {
		return nil, err
	}

	emailAddr, err := NewEmail(email)
	if err != nil {
		return nil, err
	}

	password, err := NewPasswordHash(passwordHash)
	if err != nil {
		return nil, err
	}

	statusVal, err := NewStatus(status)
	if err != nil {
		return nil, err
	}

	return &User{
		id:           id,
		firstName:    first_name,
		lastName:     last_name,
		email:        emailAddr,
		passwordHash: password,
		status:       statusVal,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

func (u *User) ID() string {
	return u.id
}

func (u *User) FirstName() *Username {
	return u.firstName
}

func (u *User) LastName() *Username {
	return u.lastName
}

func (u *User) Email() *Email {
	return u.email
}

func (u *User) PasswordHash() *PasswordHash {
	return u.passwordHash
}

func (u *User) Status() Status {
	return u.status
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() *time.Time {
	return u.updatedAt
}

func (u *User) UpdateStatus(newStatus Status) error {
	if !u.status.CanTransitionTo(newStatus) {
		return ErrInvalidStatus
	}
	u.status = newStatus
	now := time.Now()
	u.updatedAt = &now
	return nil
}
