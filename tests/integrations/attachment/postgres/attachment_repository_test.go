package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	attachmentpostgres "github.com/AppeiYA/requisition-system/internal/attachment/adapters/out/postgres"
	attachmentdomain "github.com/AppeiYA/requisition-system/internal/attachment/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

const (
	attachmentTestUserID = "usr-attachment-repository"
	attachmentReqID      = "req-attachment-repository"
	attachmentItemRefID  = "reqitem-attachment-repository"
)

type attachmentRepositoryFixture struct {
	Suite      *testutil.Suite
	Repository *attachmentpostgres.AttachmentRepository
	TxManager  *appdb.TransactionManager
}

func newAttachmentRepositoryFixture(t *testing.T) *attachmentRepositoryFixture {
	t.Helper()

	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.Reset(ctx))

	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, status, role)
		VALUES ($1, 'Attachment', 'Repository', 'attachment.repository@example.com',
		        'integration-password-hash', 'active', 'user')
	`, attachmentTestUserID)
	require.NoError(t, err)

	baseRepository := appdb.NewRepository(suite.AppDB)
	return &attachmentRepositoryFixture{
		Suite:      suite,
		Repository: attachmentpostgres.NewAttachmentRepository(baseRepository),
		TxManager:  appdb.NewTransactionManager(suite.AppDB),
	}
}

func newTestAttachment(
	t *testing.T,
	id string,
	refName attachmentdomain.AttachmentRefName,
	refID string,
	fileName string,
	storageKey string,
	checksum string,
	fileSize int64,
) *attachmentdomain.Attachment {
	t.Helper()

	ref, err := attachmentdomain.NewAttachmentRef(refName, refID)
	require.NoError(t, err)
	name, err := attachmentdomain.NewAttachmentFileName(fileName)
	require.NoError(t, err)
	attachment, err := attachmentdomain.NewAttachment(
		id,
		ref,
		storageKey,
		name,
		attachmentdomain.ContentTypePDF,
		fileSize,
		checksum,
		attachmentTestUserID,
	)
	require.NoError(t, err)
	return attachment
}

func requisitionStorageKey(id string) string {
	return "requisition/" + attachmentReqID + "/" + id
}

func (f *attachmentRepositoryFixture) save(t *testing.T, attachments ...*attachmentdomain.Attachment) {
	t.Helper()
	for _, attachment := range attachments {
		require.NoError(t, f.Repository.Save(context.Background(), attachment))
	}
}

func (f *attachmentRepositoryFixture) setCreatedAt(
	t *testing.T,
	attachment *attachmentdomain.Attachment,
	createdAt time.Time,
) {
	t.Helper()
	_, err := f.Suite.DB.Pool.Exec(
		context.Background(),
		"UPDATE attachments SET created_at = $1 WHERE id = $2",
		createdAt,
		attachment.ID(),
	)
	require.NoError(t, err)
}

func TestAttachmentRepositorySaveGetListAndCount(t *testing.T) {
	fixture := newAttachmentRepositoryFixture(t)
	ctx := context.Background()

	baseCreatedAt := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)
	attachments := []*attachmentdomain.Attachment{
		newTestAttachment(t, "att-lifecycle-1", attachmentdomain.RefNameRequisition, attachmentReqID,
			"quote-one.pdf", requisitionStorageKey("att-lifecycle-1"), fmt.Sprintf("%064x", 1), 101),
		newTestAttachment(t, "att-lifecycle-2", attachmentdomain.RefNameRequisition, attachmentReqID,
			"quote-two.pdf", requisitionStorageKey("att-lifecycle-2"), fmt.Sprintf("%064x", 2), 202),
		newTestAttachment(t, "att-lifecycle-3", attachmentdomain.RefNameRequisition, attachmentReqID,
			"quote-three.pdf", requisitionStorageKey("att-lifecycle-3"), fmt.Sprintf("%064x", 3), 303),
	}
	fixture.save(t, attachments...)
	for i, attachment := range attachments {
		fixture.setCreatedAt(t, attachment, baseCreatedAt.Add(time.Duration(i)*time.Hour))
	}

	otherRefAttachment := newTestAttachment(t, "att-lifecycle-item", attachmentdomain.RefNameRequisitionItem,
		attachmentItemRefID, "item-photo.pdf",
		"requisition_item/"+attachmentItemRefID+"/att-lifecycle-item", fmt.Sprintf("%064x", 4), 404)
	fixture.save(t, otherRefAttachment)

	found, err := fixture.Repository.GetByID(ctx, attachments[0].ID())
	require.NoError(t, err)
	require.Equal(t, attachments[0].ID(), found.ID())
	require.Equal(t, attachmentdomain.RefNameRequisition.String(), found.Ref().RefName().String())
	require.Equal(t, attachmentReqID, found.Ref().ID())
	require.Equal(t, attachments[0].StorageKey(), found.StorageKey())
	require.Equal(t, "quote-one.pdf", found.FileName().String())
	require.Equal(t, attachmentdomain.ContentTypePDF, found.ContentType())
	require.Equal(t, int64(101), found.FileSize())
	require.Equal(t, fmt.Sprintf("%064x", 1), found.Checksum())
	require.Equal(t, attachmentTestUserID, found.UploadedBy())
	require.True(t, baseCreatedAt.Equal(found.CreatedAt()))

	_, err = fixture.Repository.GetByID(ctx, "att-missing")
	require.ErrorIs(t, err, attachmentpostgres.ErrAttachmentNotFoundRepo)

	page, total, err := fixture.Repository.ListByRef(ctx, "requisition", attachmentReqID, 2, 0)
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, page, 2)
	require.Equal(t, attachments[2].ID(), page[0].ID())
	require.Equal(t, attachments[1].ID(), page[1].ID())

	nextPage, total, err := fixture.Repository.ListByRef(ctx, "requisition", attachmentReqID, 2, 2)
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, nextPage, 1)
	require.Equal(t, attachments[0].ID(), nextPage[0].ID())

	all, total, err := fixture.Repository.ListByRef(ctx, "requisition", attachmentReqID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 3, total)
	require.Len(t, all, 3)

	otherRef, total, err := fixture.Repository.ListByRef(ctx, "requisition_item", attachmentItemRefID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, otherRef, 1)
	require.Equal(t, otherRefAttachment.ID(), otherRef[0].ID())

	count, err := fixture.Repository.CountByRef(ctx, "requisition", attachmentReqID)
	require.NoError(t, err)
	require.Equal(t, 3, count)

	count, err = fixture.Repository.CountByRef(ctx, "requisition_item", attachmentItemRefID)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	count, err = fixture.Repository.CountByRef(ctx, "requisition", attachmentItemRefID)
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestAttachmentRepositoryUniqueStorageKey(t *testing.T) {
	fixture := newAttachmentRepositoryFixture(t)
	ctx := context.Background()

	storageKey := "requisition/" + attachmentReqID + "/shared-object"
	original := newTestAttachment(t, "att-conflict-1", attachmentdomain.RefNameRequisition, attachmentReqID,
		"quote.pdf", storageKey, fmt.Sprintf("%064x", 1), 101)
	fixture.save(t, original)

	duplicate := newTestAttachment(t, "att-conflict-2", attachmentdomain.RefNameRequisition, attachmentReqID,
		"quote-copy.pdf", storageKey, fmt.Sprintf("%064x", 2), 202)
	err := fixture.Repository.Save(ctx, duplicate)
	require.ErrorIs(t, err, attachmentpostgres.ErrAttachmentExistsRepo)

	count, err := fixture.Repository.CountByRef(ctx, "requisition", attachmentReqID)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestAttachmentRepositoryDelete(t *testing.T) {
	fixture := newAttachmentRepositoryFixture(t)
	ctx := context.Background()

	baseCreatedAt := time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC)
	attachments := []*attachmentdomain.Attachment{
		newTestAttachment(t, "att-delete-1", attachmentdomain.RefNameRequisition, attachmentReqID,
			"quote-one.pdf", requisitionStorageKey("att-delete-1"), fmt.Sprintf("%064x", 1), 101),
		newTestAttachment(t, "att-delete-2", attachmentdomain.RefNameRequisition, attachmentReqID,
			"quote-two.pdf", requisitionStorageKey("att-delete-2"), fmt.Sprintf("%064x", 2), 202),
		newTestAttachment(t, "att-delete-3", attachmentdomain.RefNameRequisition, attachmentReqID,
			"quote-three.pdf", requisitionStorageKey("att-delete-3"), fmt.Sprintf("%064x", 3), 303),
	}
	fixture.save(t, attachments...)
	for i, attachment := range attachments {
		fixture.setCreatedAt(t, attachment, baseCreatedAt.Add(time.Duration(i)*time.Hour))
	}

	require.NoError(t, fixture.Repository.Delete(ctx, attachments[1].ID()))
	_, err := fixture.Repository.GetByID(ctx, attachments[1].ID())
	require.ErrorIs(t, err, attachmentpostgres.ErrAttachmentNotFoundRepo)
	require.ErrorIs(t, fixture.Repository.Delete(ctx, "att-missing"), attachmentpostgres.ErrAttachmentNotFoundRepo)

	var storageKeys []string
	err = fixture.TxManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		var txErr error
		storageKeys, txErr = fixture.Repository.DeleteByRef(txCtx, "requisition", attachmentReqID)
		return txErr
	})
	require.NoError(t, err)
	require.Equal(t,
		[]string{attachments[2].StorageKey(), attachments[0].StorageKey()},
		storageKeys,
	)

	count, err := fixture.Repository.CountByRef(ctx, "requisition", attachmentReqID)
	require.NoError(t, err)
	require.Equal(t, 0, count)

	list, total, err := fixture.Repository.ListByRef(ctx, "requisition", attachmentReqID, 10, 0)
	require.NoError(t, err)
	require.Equal(t, 0, total)
	require.Empty(t, list)

	storageKeys, err = fixture.Repository.DeleteByRef(ctx, "requisition", attachmentReqID)
	require.NoError(t, err)
	require.Empty(t, storageKeys)
}
