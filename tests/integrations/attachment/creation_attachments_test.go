package attachment_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	"github.com/stretchr/testify/require"
)

func stageUploadForCreation(
	t *testing.T,
	f *attachmentHTTPFixture,
	token string,
	fileName string,
	content []byte,
) string {
	t.Helper()
	recorder := attachmentHTTPServe(f, stagedUploadMultipartRequest(t, token, fileName, content))
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())
	upload := decodeStagedUploadHTTPData(t, recorder)
	require.True(t, strings.HasPrefix(upload.ID, "upl-"), upload.ID)
	return upload.ID
}

func createRequisitionWithUploads(
	t *testing.T,
	f *attachmentHTTPFixture,
	headerUploads []string,
	lineUploads []string,
) (int, requisitionhttp.RequisitionResponse) {
	t.Helper()

	payload := map[string]any{
		"department_id": f.departmentID,
		"priority":      "medium",
		"needed_by":     f.neededBy,
		"justification": "Creation-time attachment round trip",
		"upload_ids":    headerUploads,
		"lines": []map[string]any{
			{
				"requested_item": map[string]any{
					"name":        "Notebook",
					"category_id": f.categoryID,
					"description": "Plain notebook",
				},
				"quantity":              1,
				"estimated_unit_amount": 1000,
				"upload_ids":            lineUploads,
			},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/requisitions", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder := attachmentHTTPServe(f, request)

	var envelope struct {
		Data requisitionhttp.RequisitionResponse `json:"data"`
	}
	if recorder.Code == http.StatusCreated {
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope), recorder.Body.String())
	}
	return recorder.Code, envelope.Data
}

func requisitionCount(t *testing.T, f *attachmentHTTPFixture) int {
	t.Helper()
	summaries, err := f.fixture.RequisitionRepository.ListRequisitionSummariesByUserID(
		context.Background(), attachmentHTTPUserID, 100, 0,
	)
	require.NoError(t, err)
	return len(summaries)
}

func TestRequisitionCreationClaimsStagedUploads(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	headerUploadID := stageUploadForCreation(
		t, f, f.userToken, "header.pdf", attachmentHTTPPDF,
	)
	lineUploadID := stageUploadForCreation(
		t, f, f.userToken, "line.pdf", attachmentHTTPPDF,
	)

	status, created := createRequisitionWithUploads(
		t, f, []string{headerUploadID}, []string{lineUploadID},
	)
	require.Equal(t, http.StatusCreated, status)
	require.Len(t, created.Lines, 1)

	ctx := context.Background()
	checks := []struct {
		refName  string
		refID    string
		uploadID string
		fileName string
	}{
		{"requisition", created.ID, headerUploadID, "header.pdf"},
		{"requisition_item", created.Lines[0].ID, lineUploadID, "line.pdf"},
	}
	for _, check := range checks {
		attachments, _, err := f.fixture.AttachmentRepository.ListByRef(
			ctx, check.refName, check.refID, 10, 0,
		)
		require.NoError(t, err)
		require.Len(t, attachments, 1)
		attachment := attachments[0]
		require.Equal(t, "pending", attachment.Status().String())
		require.Equal(t, check.uploadID, attachment.StagedUploadID())
		require.Equal(t, check.fileName, attachment.FileName().String())
		require.True(t, strings.HasPrefix(
			attachment.StorageKey(), check.refName+"/"+check.refID+"/att-",
		), attachment.StorageKey())

		event, err := f.fixture.OutboxRepository.GetByAttachmentID(ctx, attachment.ID())
		require.NoError(t, err)
		require.Equal(t, "aevt-", event.ID()[:5])
		require.Equal(t, check.uploadID, event.StagedUploadID())
		require.Equal(t, "staging/"+check.uploadID, event.FromKey())
		require.Equal(t, attachment.StorageKey(), event.ToKey())
		require.Equal(t, "pending", event.Status().String())

		staged, err := f.fixture.StagedUploadRepository.GetByID(ctx, check.uploadID)
		require.NoError(t, err)
		require.Equal(t, "claimed", staged.Status().String())

		// Pending files have metadata rows but no downloadable bytes yet.
		request := httptest.NewRequest(
			http.MethodGet, "/api/v1/attachments/"+attachment.ID()+"/content", nil,
		)
		request.Header.Set("Authorization", "Bearer "+f.userToken)
		recorder := attachmentHTTPServe(f, request)
		require.Equal(t, http.StatusConflict, recorder.Code, recorder.Body.String())
		require.NotContains(t, recorder.Body.String(), check.fileName)
	}
}

func TestRequisitionCreationReconcilesStagedUploadsIntoDownloadableAttachments(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	headerUploadID := stageUploadForCreation(
		t, f, f.userToken, "header.pdf", attachmentHTTPPDF,
	)
	lineUploadID := stageUploadForCreation(
		t, f, f.userToken, "line.pdf", attachmentHTTPPDF,
	)

	status, created := createRequisitionWithUploads(
		t, f, []string{headerUploadID}, []string{lineUploadID},
	)
	require.Equal(t, http.StatusCreated, status)
	require.Len(t, created.Lines, 1)

	ctx := context.Background()
	checks := []struct {
		refName  string
		refID    string
		uploadID string
	}{
		{"requisition", created.ID, headerUploadID},
		{"requisition_item", created.Lines[0].ID, lineUploadID},
	}

	// Before reconciliation only the staged bytes exist; the permanent keys
	// are absent and download is refused even though metadata has been saved.
	for _, check := range checks {
		attachments, _, err := f.fixture.AttachmentRepository.ListByRef(
			ctx, check.refName, check.refID, 10, 0,
		)
		require.NoError(t, err)
		require.Len(t, attachments, 1)
		attachment := attachments[0]
		require.Equal(t, "pending", attachment.Status().String())

		_, err = f.fixture.FileStorage.Stat(ctx, "staging/"+check.uploadID)
		require.NoError(t, err)
		_, err = f.fixture.FileStorage.Stat(ctx, attachment.StorageKey())
		require.True(t, attachmentout.IsObjectNotFound(err), err)

		request := httptest.NewRequest(
			http.MethodGet, "/api/v1/attachments/"+attachment.ID()+"/content", nil,
		)
		request.Header.Set("Authorization", "Bearer "+f.userToken)
		recorder := attachmentHTTPServe(f, request)
		require.Equal(t, http.StatusConflict, recorder.Code, recorder.Body.String())
	}

	require.NoError(t, f.fixture.Reconcile.Execute(ctx))

	for _, check := range checks {
		attachments, _, err := f.fixture.AttachmentRepository.ListByRef(
			ctx, check.refName, check.refID, 10, 0,
		)
		require.NoError(t, err)
		require.Len(t, attachments, 1)
		attachment := attachments[0]
		require.Equal(t, "ready", attachment.Status().String())

		_, err = f.fixture.FileStorage.Stat(ctx, "staging/"+check.uploadID)
		require.True(t, attachmentout.IsObjectNotFound(err), err)
		info, err := f.fixture.FileStorage.Stat(ctx, attachment.StorageKey())
		require.NoError(t, err)
		require.Equal(t, attachment.Checksum(), info.Checksum)

		event, err := f.fixture.OutboxRepository.GetByAttachmentID(ctx, attachment.ID())
		require.NoError(t, err)
		require.Equal(t, "done", event.Status().String())

		request := httptest.NewRequest(
			http.MethodGet, "/api/v1/attachments/"+attachment.ID()+"/content", nil,
		)
		request.Header.Set("Authorization", "Bearer "+f.userToken)
		recorder := attachmentHTTPServe(f, request)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		body, err := io.ReadAll(recorder.Body)
		require.NoError(t, err)
		require.Equal(t, attachmentHTTPPDF, body)
	}

	// A second sweep has nothing to do: completed events stay done and no
	// bytes are moved again.
	require.NoError(t, f.fixture.Reconcile.Execute(ctx))
}

func TestRequisitionCreationRejectsForeignStagedUpload(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	otherUploadID := stageUploadForCreation(
		t, f, f.otherToken, "other.pdf", attachmentHTTPPDF,
	)
	before := requisitionCount(t, f)

	status, _ := createRequisitionWithUploads(t, f, []string{otherUploadID}, nil)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, before, requisitionCount(t, f))

	staged, err := f.fixture.StagedUploadRepository.GetByID(context.Background(), otherUploadID)
	require.NoError(t, err)
	require.Equal(t, "pending", staged.Status().String())
}

func TestRequisitionCreationRejectsDuplicateStagedUpload(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	uploadID := stageUploadForCreation(t, f, f.userToken, "quote.pdf", attachmentHTTPPDF)
	before := requisitionCount(t, f)

	status, _ := createRequisitionWithUploads(t, f, []string{uploadID}, []string{uploadID})
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, before, requisitionCount(t, f))

	staged, err := f.fixture.StagedUploadRepository.GetByID(context.Background(), uploadID)
	require.NoError(t, err)
	require.Equal(t, "pending", staged.Status().String())
}

func TestRequisitionCreationRejectsExpiredStagedUpload(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	uploadID := stageUploadForCreation(t, f, f.userToken, "quote.pdf", attachmentHTTPPDF)
	_, err := f.fixture.AppDB.Conn().ExecContext(
		context.Background(),
		"UPDATE staged_uploads SET expires_at = $1 WHERE id = $2",
		time.Now().Add(-time.Hour), uploadID,
	)
	require.NoError(t, err)
	before := requisitionCount(t, f)

	status, _ := createRequisitionWithUploads(t, f, []string{uploadID}, nil)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, before, requisitionCount(t, f))
}

func TestRequisitionCreationRejectsUnknownStagedUpload(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	before := requisitionCount(t, f)

	status, _ := createRequisitionWithUploads(t, f, []string{"upl-does-not-exist"}, nil)
	require.Equal(t, http.StatusBadRequest, status)
	require.Equal(t, before, requisitionCount(t, f))
}
