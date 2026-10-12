package mock

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

const (
	DefaultUserID       = "usr-12345678"
	DefaultFirstName    = "John"
	DefaultLastName     = "Doe"
	DefaultEmail        = "john.doe@example.com"
	DefaultPassword     = "SecurePassword123!"
	DefaultPasswordHash = "hashed_SecurePassword123!"
)

func ValidEmail() *domain.Email {
	email, err := domain.NewEmail(DefaultEmail)
	if err != nil {
		panic(err)
	}
	return email
}

func ValidPassword() *domain.Password {
	pass, err := domain.NewPassword(DefaultPassword)
	if err != nil {
		panic(err)
	}
	return pass
}

func ValidPasswordHash() *domain.PasswordHash {
	hash, err := domain.NewPasswordHash(DefaultPasswordHash)
	if err != nil {
		panic(err)
	}
	return hash
}

func ValidUser() *domain.User {
	return NewUserBuilder().MustBuild()
}

func ValidUserWithEmail(emailStr string) *domain.User {
	return NewUserBuilder().WithEmail(emailStr).MustBuild()
}

type UserBuilder struct {
	id           string
	firstName    string
	lastName     string
	email        string
	passwordHash string
	status       string
	createdAt    time.Time
	updatedAt    *time.Time
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		id:           DefaultUserID,
		firstName:    DefaultFirstName,
		lastName:     DefaultLastName,
		email:        DefaultEmail,
		passwordHash: DefaultPasswordHash,
		status:       string(domain.ACTIVE),
		createdAt:    time.Now(),
		updatedAt:    nil,
	}
}

func (b *UserBuilder) WithID(id string) *UserBuilder {
	b.id = id
	return b
}

func (b *UserBuilder) WithFirstName(firstName string) *UserBuilder {
	b.firstName = firstName
	return b
}

func (b *UserBuilder) WithLastName(lastName string) *UserBuilder {
	b.lastName = lastName
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

func (b *UserBuilder) WithPasswordHash(hash string) *UserBuilder {
	b.passwordHash = hash
	return b
}

func (b *UserBuilder) WithStatus(status string) *UserBuilder {
	b.status = status
	return b
}

func (b *UserBuilder) WithCreatedAt(createdAt time.Time) *UserBuilder {
	b.createdAt = createdAt
	return b
}

func (b *UserBuilder) WithUpdatedAt(updatedAt *time.Time) *UserBuilder {
	b.updatedAt = updatedAt
	return b
}

func (b *UserBuilder) Build() (*domain.User, error) {
	return domain.ReconstituteUser(
		b.id,
		b.firstName,
		b.lastName,
		b.email,
		b.passwordHash,
		b.status,
		b.createdAt,
		b.updatedAt,
	)
}

func (b *UserBuilder) MustBuild() *domain.User {
	user, err := b.Build()
	if err != nil {
		panic(err)
	}
	return user
}

