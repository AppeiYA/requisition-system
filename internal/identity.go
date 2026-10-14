package app

import (
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	jwtadapter "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/jwt"
	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identityredis "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/redis"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identityusecase "github.com/AppeiYA/requisition-system/internal/identity/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/hasher"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
)

func (a *App) buildIdentity(
	cfg *config.Config,
	database *db.DB,
	redisClient *redis.Redis,
	txManager *db.TransactionManager,
) (
	*identityhttp.Handler,
	identityin.RegisterUserUseCaseInt,
	identityin.GetUserByEmailUsecaseInt,
	identityout.AccessTokenService,
	identityout.IDGenerator,
) {
	userRepository := identitypostgres.NewUserRepository(db.NewRepository(database))

	accessTokenService := jwtadapter.NewJWTService(cfg.JWT)
	refreshSessionRepository := identityredis.NewRefreshSessionRepository(
		redisClient.Client(),
		cfg.JWT.RefreshTokenTTL,
	)
	refreshTokenGenerator := &generator.TokenGenerator{}
	passwordHasher := hasher.NewPasswordHasher()
	idGen := &generator.IDGenerator{}

	loginUsecase := identityusecase.NewLoginUserUsecase(
		userRepository,
		accessTokenService,
		refreshSessionRepository,
		refreshTokenGenerator,
		passwordHasher,
	)
	refreshUsecase := identityusecase.NewRefreshTokenUsecase(
		accessTokenService,
		refreshSessionRepository,
		refreshTokenGenerator,
	)
	logoutUsecase := identityusecase.NewLogoutUsecase(
		refreshSessionRepository,
		refreshTokenGenerator,
	)
	getMeUsecase := identityusecase.NewGetMeUsecase(userRepository)
	registerUsecase := identityusecase.NewRegisterUserUsecase(userRepository, passwordHasher, idGen)
	getUserByEmailUsecase := identityusecase.NewGetUserByEmailUsecase(userRepository)

	handler := identityhttp.NewHandler(
		loginUsecase,
		refreshUsecase,
		logoutUsecase,
		getMeUsecase,
		accessTokenService,
	)

	return handler, registerUsecase, getUserByEmailUsecase, accessTokenService, idGen
}
