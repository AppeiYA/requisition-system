package postgres_test

import (
	"context"
	"testing"

	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identitydomain "github.com/AppeiYA/requisition-system/internal/identity/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func newRepository(t *testing.T, suite *testutil.Suite) *identitypostgres.UserRepository {
	t.Helper()

	suite.SetupRedis(t)
	cache := redisinfra.NewCache(suite.Redis.Client)
	repository := appdb.NewRepository(suite.AppDB)

	return identitypostgres.NewUserRepository(repository, cache)
}

func newUser(t *testing.T, id, emailAddress string) *identitydomain.User {
	t.Helper()

	email, err := shareddomain.NewEmail(emailAddress)
	require.NoError(t, err)

	passwordHash, err := identitydomain.NewPasswordHash("bcrypt-hash")
	require.NoError(t, err)

	user, err := identitydomain.NewUser(
		id,
		"Integration",
		"User",
		"",
		"",
		email,
		passwordHash,
	)
	require.NoError(t, err)

	return user
}

func TestUserRepositorySaveAndFindByEmail(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	repository := newRepository(t, suite)
	user := newUser(t, "usr-integration-1", "Integration.User@example.com")

	require.NoError(t, repository.SaveUser(ctx, user))

	found, err := repository.GetUserByEmail(ctx, user.Email())
	require.NoError(t, err)
	require.Equal(t, user.ID(), found.ID())
	require.Equal(t, "integration.user@example.com", found.Email().Value())
	require.Equal(t, identitydomain.ACTIVE, found.Status())
}

func TestUserRepositoryFindByIDUsesDatabaseThenCache(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	repository := newRepository(t, suite)
	user := newUser(t, "usr-integration-cache", "cache@example.com")
	require.NoError(t, repository.SaveUser(ctx, user))

	found, err := repository.GetUserByID(ctx, user.ID())
	require.NoError(t, err)
	require.Equal(t, user.ID(), found.ID())

	foundAgain, err := repository.GetUserByID(ctx, user.ID())
	require.NoError(t, err)
	require.Equal(t, found.ID(), foundAgain.ID())
}

func TestUserRepositoryRejectsDuplicateEmail(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	repository := newRepository(t, suite)
	first := newUser(t, "usr-integration-1", "duplicate@example.com")
	second := newUser(t, "usr-integration-2", "duplicate@example.com")

	require.NoError(t, repository.SaveUser(ctx, first))
	require.Error(t, repository.SaveUser(ctx, second))
}

func TestUserRepositoryUpdateStatus(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	repository := newRepository(t, suite)
	user := newUser(t, "usr-integration-status", "status@example.com")
	require.NoError(t, repository.SaveUser(ctx, user))

	require.NoError(t, user.UpdateStatus(identitydomain.SUSPENDED))
	require.NoError(t, repository.UpdateUserStatus(ctx, user))

	status, err := repository.CheckUserStatus(ctx, user.ID())
	require.NoError(t, err)
	require.Equal(t, identitydomain.SUSPENDED, status)

	found, err := repository.GetUserByID(ctx, user.ID())
	require.NoError(t, err)
	require.Equal(t, identitydomain.SUSPENDED, found.Status())
}
