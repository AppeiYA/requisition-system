package identityin

type IdentityPortIn struct {
	RegisterUserUseCase            RegisterUserUseCaseInt
	GetUserByEmailUsecase          GetUserByEmailUsecaseInt
	LoginUserUsecase               LoginUserUsecaseInt
	LogoutUserUsecase              LogoutUsecaseInt
	GetMeUsecase                   GetMeUsecaseInt
	RefreshTokenUsecase            RefreshTokenUsecaseInt
}

func NewIdentityPortIn(
	registerUserUseCase RegisterUserUseCaseInt,
	getUserByEmailUsecase GetUserByEmailUsecaseInt,
	loginUserUsecase LoginUserUsecaseInt,
	logoutUserUsecase LogoutUsecaseInt,
	getMeUsecase GetMeUsecaseInt,
	refreshTokenUsecase RefreshTokenUsecaseInt,
) *IdentityPortIn {
	return &IdentityPortIn{
		RegisterUserUseCase:   registerUserUseCase,
		GetUserByEmailUsecase: getUserByEmailUsecase,
		LoginUserUsecase:      loginUserUsecase,
		LogoutUserUsecase:     logoutUserUsecase,
		GetMeUsecase:          getMeUsecase,
		RefreshTokenUsecase:   refreshTokenUsecase,
	}
}