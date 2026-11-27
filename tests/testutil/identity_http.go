package testutil

import (
	"context"
	"testing"
	"time"

	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	jwtadapter "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/jwt"
	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identityredis "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/redis"
	identitydomain "github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identitytasks "github.com/AppeiYA/requisition-system/internal/identity/tasks"
	identityusecase "github.com/AppeiYA/requisition-system/internal/identity/usecase"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/hasher"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/ratelimit"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/gin-gonic/gin"
)

type IdentityHTTP struct {
	Router                       *gin.Engine
	UserRepository               *identitypostgres.UserRepository
	PasswordResetTokenRepository *identitypostgres.PasswordResetTokenRepository
	AccessTokens                 *jwtadapter.JWTService
	PasswordHasher               *hasher.PasswordHasher
	RefreshSessionRepository     *identityredis.RefreshSessionRepository
	RefreshTokenGenerator        *generator.TokenGenerator
	PasswordResetQueue           *RecordingPasswordResetQueue
	PasswordResetConfig          config.PasswordResetConfig
}

// RecordingPasswordResetQueue captures queued reset emails so integration
// tests can read the raw token without contacting Resend or asynq.
type RecordingPasswordResetQueue struct {
	Payloads []identitytasks.PasswordResetEmailPayload
}

func (q *RecordingPasswordResetQueue) EnqueuePasswordResetEmail(_ context.Context, payload identitytasks.PasswordResetEmailPayload) error {
	q.Payloads = append(q.Payloads, payload)
	return nil
}

type noopIdentityTelemetry struct{}

func (noopIdentityTelemetry) Measure(ctx context.Context, _ string, fn func(context.Context) error) error {
	return fn(ctx)
}

type IdentityUserFixture struct {
	User     *identitydomain.User
	Email    string
	Password string
}

func (i *IdentityHTTP) AddUser(
	t *testing.T,
	id,
	emailAddress,
	password,
	firstName,
	lastName string,
) IdentityUserFixture {
	t.Helper()

	return i.addUser(
		t,
		id,
		emailAddress,
		password,
		firstName,
		lastName,
		identitydomain.ACTIVE,
		shareddomain.RoleUser,
	)
}

func (i *IdentityHTTP) AddSuspendedUser(
	t *testing.T,
	id,
	emailAddress,
	password,
	firstName,
	lastName string,
) IdentityUserFixture {
	t.Helper()

	return i.addUser(
		t,
		id,
		emailAddress,
		password,
		firstName,
		lastName,
		identitydomain.SUSPENDED,
		shareddomain.RoleUser,
	)
}

func (i *IdentityHTTP) AddAdmin(
	t *testing.T,
	id,
	emailAddress,
	password,
	firstName,
	lastName string,
) IdentityUserFixture {
	t.Helper()

	return i.addUser(
		t,
		id,
		emailAddress,
		password,
		firstName,
		lastName,
		identitydomain.ACTIVE,
		shareddomain.RoleAdmin,
	)
}

func (i *IdentityHTTP) addUser(
	t *testing.T,
	id,
	emailAddress,
	password,
	firstName,
	lastName string,
	status identitydomain.Status,
	role shareddomain.Role,
) IdentityUserFixture {
	t.Helper()

	passwordHash, err := i.PasswordHasher.HashPassword(password)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}

	email, err := shareddomain.NewEmail(emailAddress)
	if err != nil {
		t.Fatalf("create test email: %v", err)
	}
	passwordHashValue, err := identitydomain.NewPasswordHash(passwordHash)
	if err != nil {
		t.Fatalf("create test password hash: %v", err)
	}

	user, err := identitydomain.NewUser(
		id,
		firstName,
		lastName,
		"",
		"",
		email,
		passwordHashValue,
	)
	if err != nil {
		t.Fatalf("create test user: %v", err)
	}

	if status == identitydomain.SUSPENDED {
		if err := user.UpdateStatus(identitydomain.SUSPENDED); err != nil {
			t.Fatalf("suspend test user: %v", err)
		}
	}

	if role == shareddomain.RoleAdmin {
		reconstituted, err := identitydomain.ReconstituteUser(
			user.ID(),
			user.FirstName().Value(),
			user.LastName().Value(),
			user.Phone().Value(),
			user.StaffID(),
			user.Email().Value(),
			user.PasswordHash().Value(),
			user.Status().Value(),
			role.String(),
			user.CreatedAt(),
			user.UpdatedAt(),
		)
		if err != nil {
			t.Fatalf("create admin test user: %v", err)
		}
		user = reconstituted
	}

	if err := i.UserRepository.SaveUser(context.Background(), user); err != nil {
		t.Fatalf("save test user: %v", err)
	}

	return IdentityUserFixture{
		User:     user,
		Email:    emailAddress,
		Password: password,
	}
}

func (s *Suite) SetupIdentityHTTP(t *testing.T) *IdentityHTTP {
	t.Helper()

	redisResource := s.SetupRedis(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Reset(ctx); err != nil {
		t.Fatalf("reset identity integration state: %v", err)
	}

	cache := redisinfra.NewCache(redisResource.Client)
	baseRepository := appdb.NewRepository(s.AppDB)
	userRepository := identitypostgres.NewUserRepository(baseRepository, cache)
	passwordResetTokenRepository := identitypostgres.NewPasswordResetTokenRepository(baseRepository)
	passwordResetConfig := config.PasswordResetConfig{
		TokenTTL:         30 * time.Minute,
		RequestCooldown:  0, // tests control timing through fixtures
		MaxPerHour:       0,
		CleanupRetention: 7 * 24 * time.Hour,
	}
	passwordResetQueue := &RecordingPasswordResetQueue{}

	jwtConfig := config.JWTConfig{
		Secret:          "integration-test-secret-long-enough",
		Issuer:          "requisition-system-test",
		Audience:        "requisition-system-test",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 24 * time.Hour,
		Leeway:          30 * time.Second,
	}
	accessTokens := jwtadapter.NewJWTService(jwtConfig)
	refreshSessionRepository := identityredis.NewRefreshSessionRepository(
		redisResource.Client,
		jwtConfig.RefreshTokenTTL,
	)
	refreshTokenGenerator := &generator.TokenGenerator{}
	passwordHasher := hasher.NewPasswordHasher()

	loginUsecase := identityusecase.NewLoginUserUsecase(
		userRepository,
		accessTokens,
		refreshSessionRepository,
		refreshTokenGenerator,
		passwordHasher,
	)
	refreshUsecase := identityusecase.NewRefreshTokenUsecase(
		accessTokens,
		refreshSessionRepository,
		refreshTokenGenerator,
		userRepository,
	)
	logoutUsecase := identityusecase.NewLogoutUsecase(
		refreshSessionRepository,
		refreshTokenGenerator,
	)
	getMeUsecase := identityusecase.NewGetMeUsecase(userRepository)
	registerUsecase := identityusecase.NewRegisterUserUsecase(
		userRepository,
		passwordHasher,
		&generator.IDGenerator{},
	)
	getUserByEmailUsecase := identityusecase.NewGetUserByEmailUsecase(userRepository)
	authMiddleware := identityhttp.NewAuthMiddleware(accessTokens, userRepository)

	txManager := appdb.NewTransactionManager(s.AppDB)
	requestPasswordResetUsecase := identityusecase.NewRequestPasswordResetUsecase(
		userRepository,
		passwordResetTokenRepository,
		passwordResetQueue,
		refreshTokenGenerator,
		&generator.IDGenerator{},
		txManager,
		passwordResetConfig,
		noopIdentityTelemetry{},
	)
	resetPasswordUsecase := identityusecase.NewResetPasswordUsecase(
		userRepository,
		passwordResetTokenRepository,
		refreshSessionRepository,
		refreshTokenGenerator,
		passwordHasher,
		txManager,
		noopIdentityTelemetry{},
	)

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
	handler := identityhttp.NewHandler(identityPortIn, authMiddleware)
	handler.SetRateLimiter(ratelimit.New(redisResource.Client), config.RateLimitConfig{
		LoginLimit:           5,
		LoginWindow:          15 * time.Minute,
		RefreshLimit:         10,
		RefreshWindow:        time.Minute,
		LogoutLimit:          20,
		LogoutWindow:         time.Minute,
		ForgotPasswordLimit:  5,
		ForgotPasswordWindow: 15 * time.Minute,
		ResetPasswordLimit:   10,
		ResetPasswordWindow:  15 * time.Minute,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	return &IdentityHTTP{
		Router:                       router,
		UserRepository:               userRepository,
		PasswordResetTokenRepository: passwordResetTokenRepository,
		AccessTokens:                 accessTokens,
		PasswordHasher:               passwordHasher,
		RefreshSessionRepository:     refreshSessionRepository,
		RefreshTokenGenerator:        refreshTokenGenerator,
		PasswordResetQueue:           passwordResetQueue,
		PasswordResetConfig:          passwordResetConfig,
	}
}
