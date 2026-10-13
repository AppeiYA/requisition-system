package app

import (
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	jwtadapter "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/jwt"
	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identityredis "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/redis"
	"github.com/AppeiYA/requisition-system/internal/shared/config"
	"github.com/AppeiYA/requisition-system/internal/shared/db"
	"github.com/AppeiYA/requisition-system/internal/shared/generator"
	"github.com/AppeiYA/requisition-system/internal/shared/hasher"
	"github.com/AppeiYA/requisition-system/internal/shared/redis"

	identityusecase "github.com/AppeiYA/requisition-system/internal/identity/usecase"
)

func (a *App) buildIdentityHandler(
	cfg *config.Config,
	database *db.DB,
	redisClient *redis.Redis,
) *identityhttp.Handler {
	userRepository := identitypostgres.NewUserRepository(db.NewRepository(database))

	accessTokenService := jwtadapter.NewJWTService(cfg.JWT)
	refreshSessionRepository := identityredis.NewRefreshSessionRepository(
		redisClient.Client(),
		cfg.JWT.RefreshTokenTTL,
	)
	refreshTokenGenerator := &generator.TokenGenerator{}
	passwordHasher := hasher.NewPasswordHasher()

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

	return identityhttp.NewHandler(
		loginUsecase,
		refreshUsecase,
		logoutUsecase,
		getMeUsecase,
		accessTokenService,
	)
}