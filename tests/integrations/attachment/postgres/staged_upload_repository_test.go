package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	attachmentpostgres "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/postgres"
	attachmentdomain "github.com/AppeiYA/requisition-system/internal/attachment/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

const (
	stagedRepoOwnerID = "usr-staged-repository-owner"
	stagedRepoOtherID = "usr-staged-repository-other"
)

type stagedUploadRepositoryFixture struct {
	Suite      *testutil.Suite
	Repository *attachmentpostgres.StagedUploadRepository
}

func newStagedUploadRepositoryFixture(t *testing.T) *stagedUploadRepositoryFixture {
	t.Helper()

	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.Reset(ctx))

	for _, userID := range []string{stagedRepoOwnerID, stagedRepoOtherID} {
		_, err := suite.DB.Pool.Exec(ctx, `
			INSERT INTO users (id, first_name, last_name, email, password_hash, status, role)
			VALUES ($1, 'Staged', 'Repository', $2, 'integration-password-hash', 'active', 'user')
		`, userID, userID+"@example.com")
		require.NoError(t, err)
	}

	return &stagedUploadRepositoryFixture{
		Suite:      suite,
		Repository: attachmentpostgres.NewStagedUploadRepository(appdb.NewRepository(suite.AppDB)),
	}
}

func newStagedUpload(
	t *testing.T,
	id, actorID string,
	expiresAt time.Time,
) *attachmentdomain.StagedUpload {
	t.Helper()

	name, err := attachmentdomain.NewAttachmentFileName("quote.pdf")
	require.NoError(t, err)
	key, err := attachmentdomain.NewStagingUploadKey(id)
	require.NoError(t, err)
	upload, err := attachmentdomain.NewStagedUpload(
		id,
		actorID,
		key,
		name,
		attachmentdomain.ContentTypePDF,
		42,
		strings.Repeat("a", attachmentdomain.ChecksumLength),
		time.Hour,
		expiresAt.Add(-time.Hour),
	)
	require.NoError(t, err)
	return upload
}

func TestStagedUploadRepositoryRoundTripAndClaimQuery(t *testing.T) {
	f := newStagedUploadRepositoryFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()

	claimable := newStagedUpload(t, "upl-claimable", stagedRepoOwnerID, now.Add(time.Hour))
	expired := newStagedUpload(t, "upl-expired", stagedRepoOwnerID, now.Add(-time.Minute))
	foreign := newStagedUpload(t, "upl-foreign", stagedRepoOtherID, now.Add(time.Hour))

	for _, upload := range []*attachmentdomain.StagedUpload{claimable, expired, foreign} {
		require.NoError(t, f.Repository.Save(ctx, upload))
	}

	got, err := f.Repository.GetByID(ctx, claimable.ID())
	require.NoError(t, err)
	require.Equal(t, claimable.ID(), got.ID())
	require.Equal(t, claimable.StoragePath(), got.StoragePath())
	require.Equal(t, claimable.Checksum(), got.Checksum())
	require.Equal(t, attachmentdomain.StagedUploadStatusPending, got.Status())
	require.WithinDuration(t, claimable.ExpiresAt(), got.ExpiresAt(), time.Microsecond)

	claimedRows, err := f.Repository.ClaimByIDs(
		ctx,
		[]string{claimable.ID(), expired.ID(), foreign.ID()},
		stagedRepoOwnerID,
		now,
	)
	require.NoError(t, err)
	require.Len(t, claimedRows, 1)
	require.Equal(t, claimable.ID(), claimedRows[0].ID())
	require.Equal(t, attachmentdomain.StagedUploadStatusClaimed, claimedRows[0].Status())
	require.NotNil(t, claimedRows[0].ClaimedAt())

	// Second claim on the same ID must return 0 rows (already claimed).
	secondClaim, err := f.Repository.ClaimByIDs(
		ctx,
		[]string{claimable.ID()},
		stagedRepoOwnerID,
		now,
	)
	require.NoError(t, err)
	require.Empty(t, secondClaim)

	empty, err := f.Repository.ClaimByIDs(ctx, nil, stagedRepoOwnerID, now)
	require.NoError(t, err)
	require.Empty(t, empty)
}

func TestStagedUploadRepositoryUniqueStoragePath(t *testing.T) {
	f := newStagedUploadRepositoryFixture(t)
	ctx := context.Background()

	upload := newStagedUpload(t, "upl-duplicate", stagedRepoOwnerID, time.Now().UTC().Add(time.Hour))
	require.NoError(t, f.Repository.Save(ctx, upload))
	require.ErrorIs(t, f.Repository.Save(ctx, upload), attachmentpostgres.ErrStagedUploadExistsRepo)
}
