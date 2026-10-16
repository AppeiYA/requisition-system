package app

import (
	"github.com/AppeiYA/requisition-system/internal/identity"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	jwtadapter "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/jwt"
	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identityredis "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/redis"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
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
	accessTokenService *jwtadapter.JWTService,
	idGen *generator.IDGenerator,
) (
	*identity.IdentityModule,
) {
	userRepository := identitypostgres.NewUserRepository(db.NewRepository(database))

	refreshSessionRepository := identityredis.NewRefreshSessionRepository(
		redisClient.Client(),
		cfg.JWT.RefreshTokenTTL,
	)
	refreshTokenGenerator := &generator.TokenGenerator{}
	passwordHasher := hasher.NewPasswordHasher()
	authMiddleware := identityhttp.NewAuthMiddleware(accessTokenService)

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

	identityPortIn := identityin.NewIdentityPortIn(
		registerUsecase,
		getUserByEmailUsecase,
		loginUsecase,
		logoutUsecase,
		getMeUsecase,
		refreshUsecase,
	)

	handler := identityhttp.NewHandler(
		identityPortIn,
		authMiddleware,
	)

	return identity.NewIdentityModule(
		identityPortIn,
		handler,
		authMiddleware,
	)
}
