package attachment_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	"github.com/stretchr/testify/require"
)

func TestCleanupExpiresPendingStagedUploadAndDeletesBytes(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	ctx := context.Background()
	uploadID := stageUploadForCreation(t, f, f.userToken, "expired.pdf", attachmentHTTPPDF)
	key := "staging/" + uploadID

	_, err := f.fixture.FileStorage.Stat(ctx, key)
	require.NoError(t, err)
	_, err = f.fixture.AppDB.Conn().ExecContext(
		ctx,
		"UPDATE staged_uploads SET expires_at = $1 WHERE id = $2",
		time.Now().Add(-time.Hour), uploadID,
	)
	require.NoError(t, err)

	require.NoError(t, f.fixture.Cleanup.Execute(ctx))

	staged, err := f.fixture.StagedUploadRepository.GetByID(ctx, uploadID)
	require.NoError(t, err)
	require.Equal(t, "expired", staged.Status().String())
	_, err = f.fixture.FileStorage.Stat(ctx, key)
	require.True(t, attachmentout.IsObjectNotFound(err), err)
}

func TestCleanupLeavesClaimedStagedUploadUntilReconciliation(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	ctx := context.Background()
	uploadID := stageUploadForCreation(t, f, f.userToken, "claimed.pdf", attachmentHTTPPDF)
	status, created := createRequisitionWithUploads(t, f, []string{uploadID}, nil)
	require.Equal(t, http.StatusCreated, status)
	require.Len(t, created.Lines, 1)

	_, err := f.fixture.AppDB.Conn().ExecContext(
		ctx,
		"UPDATE staged_uploads SET expires_at = $1 WHERE id = $2",
		time.Now().Add(-time.Hour), uploadID,
	)
	require.NoError(t, err)

	require.NoError(t, f.fixture.Cleanup.Execute(ctx))

	staged, err := f.fixture.StagedUploadRepository.GetByID(ctx, uploadID)
	require.NoError(t, err)
	require.Equal(t, "claimed", staged.Status().String())
	_, err = f.fixture.FileStorage.Stat(ctx, "staging/"+uploadID)
	require.NoError(t, err)
}

func TestCleanupRecoversStuckProcessingAttachmentAndReconcilesIt(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	ctx := context.Background()
	uploadID := stageUploadForCreation(t, f, f.userToken, "stuck.pdf", attachmentHTTPPDF)
	status, created := createRequisitionWithUploads(t, f, []string{uploadID}, nil)
	require.Equal(t, http.StatusCreated, status)

	attachments, _, err := f.fixture.AttachmentRepository.ListByRef(
		ctx, "requisition", created.ID, 10, 0,
	)
	require.NoError(t, err)
	require.Len(t, attachments, 1)
	attachment := attachments[0]
	stuckAt := time.Now().Add(-time.Hour)

	_, err = f.fixture.AppDB.Conn().ExecContext(
		ctx,
		"UPDATE attachments SET status = 'processing', updated_at = $1 WHERE id = $2",
		stuckAt, attachment.ID(),
	)
	require.NoError(t, err)
	_, err = f.fixture.AppDB.Conn().ExecContext(
		ctx,
		`UPDATE attachment_reconciliation_outbox
		 SET status = 'processing', next_attempt_at = $1, updated_at = $1
		 WHERE attachment_id = $2`,
		stuckAt, attachment.ID(),
	)
	require.NoError(t, err)

	require.NoError(t, f.fixture.Cleanup.Execute(ctx))

	recovered, err := f.fixture.AttachmentRepository.GetByID(ctx, attachment.ID())
	require.NoError(t, err)
	require.Equal(t, "pending", recovered.Status().String())
	event, err := f.fixture.OutboxRepository.GetByAttachmentID(ctx, attachment.ID())
	require.NoError(t, err)
	require.Equal(t, "pending", event.Status().String())

	require.NoError(t, f.fixture.Reconcile.Execute(ctx))
	ready, err := f.fixture.AttachmentRepository.GetByID(ctx, attachment.ID())
	require.NoError(t, err)
	require.Equal(t, "ready", ready.Status().String())

	request := httptest.NewRequest(
		http.MethodGet, "/api/v1/attachments/"+attachment.ID()+"/content", nil,
	)
	request.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder := attachmentHTTPServe(f, request)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)
	require.True(t, bytes.Equal(attachmentHTTPPDF, body))
}
