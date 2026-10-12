package usecase

import (
	"context"
	"errors"
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/errors"
)

var (
	ErrUserAlreadyExists = customerrors.NewConflictError("user with this email already exists")
)

type RegisterUserDTO struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
}

type RegisterUserUsecase struct {
	userRepo       identityout.UsersRepository
	passwordHasher identityout.PasswordHasher
	idGenerator    identityout.IDGenerator
}

func NewRegisterUserUsecase(
	userRepo identityout.UsersRepository,
	passwordHasher identityout.PasswordHasher,
	idGenerator identityout.IDGenerator,
) *RegisterUserUsecase {
	return &RegisterUserUsecase{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
		idGenerator:    idGenerator,
	}
}

func (u *RegisterUserUsecase) Execute(ctx context.Context, dto RegisterUserDTO) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	email, err := domain.NewEmail(dto.Email)
	if err != nil {
		return err
	}

	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		var custErr *customerrors.CustomError
		if !errors.As(err, &custErr) || custErr.Code != http.StatusNotFound {
			return err
		}
	}
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	userID, err := u.idGenerator.GenerateID(domain.UserIDPrefix)
	if err != nil {
		return err
	}

	validPass, err := domain.NewPassword(dto.Password)
	if err != nil {
		return err
	}

	hashedPassword, err := u.passwordHasher.HashPassword(validPass.Value())
	if err != nil {
		return err
	}

	passHash, err := domain.NewPasswordHash(hashedPassword)
	if err != nil {
		return err
	}

	user, err := domain.NewUser(userID, dto.FirstName, dto.LastName, email, passHash)
	if err != nil {
		return err
	}

	return u.userRepo.SaveUser(ctx, user)
}
