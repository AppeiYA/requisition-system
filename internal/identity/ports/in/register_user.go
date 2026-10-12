package identityin

import "context"

type RegisterUserDTO struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type RegisterUserUseCaseInt interface {
	Execute(ctx context.Context, dto RegisterUserDTO) error
}