package app

import (
	"github.com/AppeiYA/requisition-system/internal/identity"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	jwtadapter "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/jwt"
	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identityredis "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/redis"
	identityworker "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/worker"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	identityusecase "github.com/AppeiYA/requisition-system/internal/identity/usecase"
	requisitionexternal "github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/external"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/audit"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/hasher"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/ratelimit"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/hibiken/asynq"
)

func (a *App) buildIdentity(
	cfg *config.Config,
	baseRepo db.Repository,
	cache cache.Cache,
	redisClient *redis.Redis,
	txManager *db.TransactionManager,
	accessTokenService *jwtadapter.JWTService,
	idGen *generator.IDGenerator,
	asynqClient *asynq.Client,
	operationTelemetry identityout.Telemetry,
) *identity.IdentityModule {
	userRepository := identitypostgres.NewUserRepository(baseRepo, cache)
	passwordResetTokenRepository := identitypostgres.NewPasswordResetTokenRepository(baseRepo)

	refreshSessionRepository := identityredis.NewRefreshSessionRepository(
		redisClient.Client(),
		cfg.JWT.RefreshTokenTTL,
	)
	refreshTokenGenerator := &generator.TokenGenerator{}
	passwordHasher := hasher.NewPasswordHasher()
	authMiddleware := identityhttp.NewAuthMiddleware(accessTokenService, userRepository)

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
		userRepository,
	)
	logoutUsecase := identityusecase.NewLogoutUsecase(
		refreshSessionRepository,
		refreshTokenGenerator,
	)
	getMeUsecase := identityusecase.NewGetMeUsecase(userRepository)
	registerUsecase := identityusecase.NewRegisterUserUsecase(userRepository, passwordHasher, idGen)
	getUserByEmailUsecase := identityusecase.NewGetUserByEmailUsecase(userRepository)

	passwordResetQueue := identityworker.NewPasswordResetQueue(asynqClient, cfg.Worker)
	requestPasswordResetUsecase := identityusecase.NewRequestPasswordResetUsecase(
		userRepository,
		passwordResetTokenRepository,
		passwordResetQueue,
		refreshTokenGenerator,
		idGen,
		txManager,
		cfg.PasswordReset,
		operationTelemetry,
	)
	resetPasswordUsecase := identityusecase.NewResetPasswordUsecase(
		userRepository,
		passwordResetTokenRepository,
		refreshSessionRepository,
		refreshTokenGenerator,
		passwordHasher,
		txManager,
		operationTelemetry,
	)

	auditRecorder := audit.NewPostgresRecorder(baseRepo, idGen)
	adminUsersPortIn := &identityin.AdminUsersPortIn{
		ListUsers:         identityusecase.NewListUsersUsecase(userRepository),
		GetUser:           identityusecase.NewGetUserUsecase(userRepository),
		UpdateUserStatus:  identityusecase.NewUpdateUserStatusUsecase(userRepository, refreshSessionRepository, txManager, auditRecorder),
		UpdateUserRole:    identityusecase.NewUpdateUserRoleUsecase(userRepository, refreshSessionRepository, txManager, auditRecorder),
		RevokeUserSession: identityusecase.NewRevokeUserSessionsUsecase(userRepository, refreshSessionRepository, auditRecorder),
	}

	identityPortIn := identityin.NewIdentityPortIn(
		registerUsecase,
		getUserByEmailUsecase,
		loginUsecase,
		logoutUsecase,
		getMeUsecase,
		refreshUsecase,
		requestPasswordResetUsecase,
		resetPasswordUsecase,
	)

	handler := identityhttp.NewHandler(
		identityPortIn,
		authMiddleware,
	)
	handler.SetPasswordResetEnabled(cfg.PasswordReset.Enabled)
	handler.SetRateLimiter(ratelimit.New(redisClient.Client()), cfg.RateLimit)
	userLookup := requisitionexternal.NewIdentityAdapter(userRepository)

	return identity.NewIdentityModule(
		identityPortIn,
		adminUsersPortIn,
		handler,
		authMiddleware,
		userLookup,
	)
}
