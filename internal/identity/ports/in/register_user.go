package identityin

import "context"

type RegisterUserDTO struct {
	FirstName string
	LastName  string
	Phone     string
	StaffID   string
	Email     string
	Password  string
}

type RegisterUserUseCaseInt interface {
	// Execute registers the user and returns the generated user ID.
	Execute(ctx context.Context, dto RegisterUserDTO) (string, error)
}
