package postgres_test

import (
	"context"
	"testing"
	"time"

	identitypostgres "github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres"
	identitydomain "github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func newPasswordResetRepository(t *testing.T, suite *testutil.Suite) *identitypostgres.PasswordResetTokenRepository {
	t.Helper()

	return identitypostgres.NewPasswordResetTokenRepository(appdb.NewRepository(suite.AppDB))
}

func createResetToken(t *testing.T, repo *identitypostgres.PasswordResetTokenRepository, id, userID, hash string, now time.Time, ttl time.Duration) *identitydomain.PasswordResetToken {
	t.Helper()

	token, err := identitydomain.NewPasswordResetToken(id, userID, hash, now, ttl)
	require.NoError(t, err)
	require.NoError(t, repo.Create(context.Background(), token))
	return token
}

func TestPasswordResetTokenRepository_CreateAndLookup(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	users := newRepository(t, suite)
	user := newUser(t, "usr-prt-1", "prt.lookup@example.com")
	require.NoError(t, users.SaveUser(ctx, user))

	repo := newPasswordResetRepository(t, suite)
	token := createResetToken(t, repo, "prt-1", user.ID(), "sha256-hash", time.Now().UTC(), 30*time.Minute)

	found, err := repo.GetByTokenHash(ctx, token.TokenHash())
	require.NoError(t, err)
	require.Equal(t, token.ID(), found.ID())
	require.Equal(t, token.UserID(), found.UserID())
	require.Equal(t, "sha256-hash", found.TokenHash())
	require.False(t, found.IsUsed())

	var storedRaw int
	err = suite.SQLX.GetContext(ctx, &storedRaw,
		`SELECT COUNT(*) FROM password_reset_tokens WHERE token_hash = $1`, "raw-token")
	require.NoError(t, err)
	require.Zero(t, storedRaw)

	missing, err := repo.GetByTokenHash(ctx, "unknown-hash")
	require.ErrorIs(t, err, identityout.ErrPasswordResetTokenNotFound)
	require.Nil(t, missing)
}

func TestPasswordResetTokenRepository_ConsumeIsSingleUse(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	users := newRepository(t, suite)
	user := newUser(t, "usr-prt-consume", "prt.consume@example.com")
	require.NoError(t, users.SaveUser(ctx, user))

	repo := newPasswordResetRepository(t, suite)
	token := createResetToken(t, repo, "prt-consume", user.ID(), "consume-hash", time.Now().UTC(), 30*time.Minute)

	consumed, err := repo.Consume(ctx, token.TokenHash(), user.ID(), time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, token.ID(), consumed.ID())
	require.NotNil(t, consumed.UsedAt())

	_, err = repo.Consume(ctx, token.TokenHash(), user.ID(), time.Now().UTC())
	require.ErrorIs(t, err, identityout.ErrPasswordResetTokenNotFound)

	_, err = repo.Consume(ctx, token.TokenHash(), "usr-someone-else", time.Now().UTC())
	require.ErrorIs(t, err, identityout.ErrPasswordResetTokenNotFound)
}

func TestPasswordResetTokenRepository_ExpiredTokenCannotBeConsumed(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	users := newRepository(t, suite)
	user := newUser(t, "usr-prt-expired", "prt.expired@example.com")
	require.NoError(t, users.SaveUser(ctx, user))

	repo := newPasswordResetRepository(t, suite)
	old := time.Now().UTC().Add(-time.Hour)
	token, err := identitydomain.NewPasswordResetToken("prt-expired", user.ID(), "expired-hash", old, time.Minute)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, token))

	_, err = repo.Consume(ctx, token.TokenHash(), user.ID(), time.Now().UTC())
	require.ErrorIs(t, err, identityout.ErrPasswordResetTokenNotFound)
}

func TestPasswordResetTokenRepository_OneActiveTokenPerUser(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	users := newRepository(t, suite)
	user := newUser(t, "usr-prt-unique", "prt.unique@example.com")
	require.NoError(t, users.SaveUser(ctx, user))

	repo := newPasswordResetRepository(t, suite)
	createResetToken(t, repo, "prt-unique-1", user.ID(), "hash-1", time.Now().UTC(), 30*time.Minute)

	second, err := identitydomain.NewPasswordResetToken("prt-unique-2", user.ID(), "hash-2", time.Now().UTC(), 30*time.Minute)
	require.NoError(t, err)
	require.Error(t, repo.Create(ctx, second), "partial unique index must allow only one active token per user")

	// Making the first token terminal frees the slot for a new one.
	require.NoError(t, repo.InvalidateActiveForUser(ctx, user.ID(), time.Now().UTC()))
	require.NoError(t, repo.Create(ctx, second))
}

func TestPasswordResetTokenRepository_InvalidateAndDeleteOlderThan(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	users := newRepository(t, suite)
	user := newUser(t, "usr-prt-invalidate", "prt.invalidate@example.com")
	require.NoError(t, users.SaveUser(ctx, user))

	repo := newPasswordResetRepository(t, suite)
	old := time.Now().UTC().Add(-2 * time.Hour)
	// The partial unique index counts expired-but-not-invalidated rows too, so
	// the old row must be made terminal before a second row is inserted.
	oldToken, err := identitydomain.NewPasswordResetToken("prt-old", user.ID(), "old-hash", old, 4*time.Hour)
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, oldToken))
	require.NoError(t, repo.InvalidateActiveForUser(ctx, user.ID(), time.Now().UTC()))
	newToken := createResetToken(t, repo, "prt-new", user.ID(), "new-hash", time.Now().UTC(), 30*time.Minute)

	latest, err := repo.GetLatestActiveForUser(ctx, user.ID(), time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, newToken.ID(), latest.ID())

	require.NoError(t, repo.InvalidateActiveForUser(ctx, user.ID(), time.Now().UTC()))
	latest, err = repo.GetLatestActiveForUser(ctx, user.ID(), time.Now().UTC())
	require.NoError(t, err)
	require.Nil(t, latest)

	count, err := repo.CountCreatedSince(ctx, user.ID(), time.Now().UTC().Add(-time.Hour))
	require.NoError(t, err)
	require.Equal(t, 1, count)

	deleted, err := repo.DeleteOlderThan(ctx, time.Now().UTC().Add(-time.Hour))
	require.NoError(t, err)
	require.EqualValues(t, 1, deleted)

	missing, err := repo.GetByTokenHash(ctx, oldToken.TokenHash())
	require.ErrorIs(t, err, identityout.ErrPasswordResetTokenNotFound)
	require.Nil(t, missing)
}

func TestPasswordResetTokenRepository_InvalidateByIDOnlyTargetsOneToken(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	users := newRepository(t, suite)
	first := newUser(t, "usr-prt-byid-1", "prt.byid1@example.com")
	second := newUser(t, "usr-prt-byid-2", "prt.byid2@example.com")
	require.NoError(t, users.SaveUser(ctx, first))
	require.NoError(t, users.SaveUser(ctx, second))

	repo := newPasswordResetRepository(t, suite)
	tokenOne := createResetToken(t, repo, "prt-byid-1", first.ID(), "byid-hash-1", time.Now().UTC(), 30*time.Minute)
	tokenTwo := createResetToken(t, repo, "prt-byid-2", second.ID(), "byid-hash-2", time.Now().UTC(), 30*time.Minute)

	require.NoError(t, repo.InvalidateByID(ctx, tokenOne.ID(), time.Now().UTC()))

	one, err := repo.GetByTokenHash(ctx, tokenOne.TokenHash())
	require.NoError(t, err)
	require.True(t, one.IsUsed())
	two, err := repo.GetByTokenHash(ctx, tokenTwo.TokenHash())
	require.NoError(t, err)
	require.False(t, two.IsUsed())
}
