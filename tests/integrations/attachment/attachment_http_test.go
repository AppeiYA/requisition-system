package attachment_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

const (
	attachmentHTTPUserID  = "usr-attachment-http-user"
	attachmentHTTPOtherID = "usr-attachment-http-other"
	attachmentHTTPDeptID  = "dpt-attachment-http"
	attachmentHTTPCatID   = "cat-attachment-http"
)

var attachmentHTTPPDF = []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")

type attachmentHTTPFixture struct {
	fixture      *testutil.AttachmentHTTP
	userToken    string
	otherToken   string
	departmentID string
	categoryID   string
	neededBy     string
}

func setupAttachmentHTTPFixture(t *testing.T) *attachmentHTTPFixture {
	t.Helper()

	fixture := testutil.NewSuite(t, "../../../migrations").SetupAttachmentHTTP(t)
	ctx := context.Background()

	user := fixture.AddUser(
		t,
		attachmentHTTPUserID,
		"attachment.user@example.com",
		"Password123$",
		"Attachment",
		"User",
	)
	other := fixture.AddUser(
		t,
		attachmentHTTPOtherID,
		"attachment.other@example.com",
		"Password123$",
		"Attachment",
		"Other",
	)

	name, err := departmentdomain.NewDepartmentName("Attachments")
	require.NoError(t, err)
	description, err := departmentdomain.NewDepartmentDescription("Attachment HTTP integration test")
	require.NoError(t, err)
	code, err := departmentdomain.NewDepartmentCode("ATF")
	require.NoError(t, err)
	department, err := departmentdomain.NewDepartment(attachmentHTTPDeptID, name, description, code)
	require.NoError(t, err)
	require.NoError(t, fixture.DepartmentRepository.SaveDepartment(ctx, department))
	fixture.AddMembership(t, user.User.ID(), attachmentHTTPDeptID, "Engineer")

	categoryName, err := requisitiondomain.NewCategoryName("Office supplies")
	require.NoError(t, err)
	categoryDescription, err := requisitiondomain.NewCategoryDescription("Items used in the test")
	require.NoError(t, err)
	category, err := requisitiondomain.NewItemCategory(
		attachmentHTTPCatID,
		categoryName,
		categoryDescription,
	)
	require.NoError(t, err)
	require.NoError(t, fixture.ItemCategoryRepository.SaveItemCategory(ctx, category))

	return &attachmentHTTPFixture{
		fixture:      fixture,
		userToken:    attachmentHTTPToken(t, fixture, user.User.ID(), "user"),
		otherToken:   attachmentHTTPToken(t, fixture, other.User.ID(), "user"),
		departmentID: attachmentHTTPDeptID,
		categoryID:   attachmentHTTPCatID,
		neededBy:     time.Now().UTC().AddDate(0, 0, 7).Format("2006-01-02"),
	}
}

func attachmentHTTPToken(
	t *testing.T,
	fixture *testutil.AttachmentHTTP,
	userID string,
	role string,
) string {
	t.Helper()

	token, err := fixture.AccessTokens.GenerateAccessToken(context.Background(), &identityout.TokenModel{
		ID:   userID,
		Role: role,
	})
	require.NoError(t, err)
	return token.Value
}

func (f *attachmentHTTPFixture) createRequisition(t *testing.T) (string, string) {
	t.Helper()

	payload := map[string]any{
		"department_id": f.departmentID,
		"priority":      "medium",
		"needed_by":     f.neededBy,
		"justification": "Attachment HTTP round trip",
		"lines": []map[string]any{
			{
				"requested_item": map[string]any{
					"name":        "Notebook",
					"category_id": f.categoryID,
					"description": "Plain notebook",
				},
				"quantity":              1,
				"estimated_unit_amount": 1000,
			},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/requisitions", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder := httptest.NewRecorder()
	f.fixture.Router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	var envelope struct {
		Data requisitionhttp.RequisitionResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope), recorder.Body.String())
	require.NotEmpty(t, envelope.Data.ID)
	require.NotEmpty(t, envelope.Data.Lines)
	require.True(t, strings.HasPrefix(envelope.Data.Lines[0].ID, "reqitem-"), envelope.Data.Lines[0].ID)
	return envelope.Data.ID, envelope.Data.Lines[0].ID
}

func attachmentMultipartRequest(
	t *testing.T,
	path string,
	token string,
	refName string,
	refID string,
	fileName string,
	content []byte,
) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("ref_name", refName))
	require.NoError(t, writer.WriteField("ref_id", refID))
	part, err := writer.CreateFormFile("file", fileName)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, path, &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func stagedUploadMultipartRequest(
	t *testing.T,
	token string,
	fileName string,
	content []byte,
) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", fileName)
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request := httptest.NewRequest(http.MethodPost, "/api/v1/uploads", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	return request
}

func attachmentHTTPServe(f *attachmentHTTPFixture, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	f.fixture.Router.ServeHTTP(recorder, request)
	return recorder
}

type attachmentHTTPData struct {
	ID          string    `json:"id"`
	RefName     string    `json:"ref_name"`
	RefID       string    `json:"ref_id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
	URL         string    `json:"url"`
}

func decodeAttachmentHTTPData(t *testing.T, recorder *httptest.ResponseRecorder) attachmentHTTPData {
	t.Helper()

	var envelope struct {
		Data attachmentHTTPData `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope), recorder.Body.String())
	return envelope.Data
}

type stagedUploadHTTPData struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	Checksum    string    `json:"checksum"`
	Status      string    `json:"status"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func decodeStagedUploadHTTPData(t *testing.T, recorder *httptest.ResponseRecorder) stagedUploadHTTPData {
	t.Helper()

	var envelope struct {
		Data stagedUploadHTTPData `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope), recorder.Body.String())
	return envelope.Data
}

func TestStagedUploadHTTPRoundTrip(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	recorder := attachmentHTTPServe(f, stagedUploadMultipartRequest(
		t,
		f.userToken,
		"quote.pdf",
		attachmentHTTPPDF,
	))
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	created := decodeStagedUploadHTTPData(t, recorder)
	require.True(t, strings.HasPrefix(created.ID, "upl-"), created.ID)
	require.Equal(t, "quote.pdf", created.FileName)
	require.Equal(t, "application/pdf", created.ContentType)
	require.Equal(t, int64(len(attachmentHTTPPDF)), created.FileSize)
	require.Len(t, created.Checksum, 64)
	require.Equal(t, "pending", created.Status)
	require.Equal(t, attachmentHTTPUserID, created.UploadedBy)
	require.True(t, created.ExpiresAt.After(created.CreatedAt))
	require.NotContains(t, recorder.Body.String(), "staging/")
	require.NotContains(t, recorder.Body.String(), f.fixture.LocalRoot)

	stored, err := f.fixture.StagedUploadRepository.GetByID(context.Background(), created.ID)
	require.NoError(t, err)
	require.Equal(t, "staging/"+created.ID, stored.StoragePath())
	require.Equal(t, created.Checksum, stored.Checksum())
	require.False(t, stored.ExpiresAt().IsZero())
}

func TestStagedUploadHTTPFailures(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)

	t.Run("missing auth", func(t *testing.T) {
		recorder := attachmentHTTPServe(f, stagedUploadMultipartRequest(
			t, "", "quote.pdf", attachmentHTTPPDF,
		))
		require.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
	})

	t.Run("unsupported content type", func(t *testing.T) {
		recorder := attachmentHTTPServe(f, stagedUploadMultipartRequest(
			t, f.userToken, "page.html", []byte("<html><body>not a document</body></html>"),
		))
		require.Equal(t, http.StatusUnsupportedMediaType, recorder.Code, recorder.Body.String())
	})

	t.Run("oversize", func(t *testing.T) {
		recorder := attachmentHTTPServe(f, stagedUploadMultipartRequest(
			t, f.userToken, "large.pdf",
			append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("x"), 8<<10)...),
		))
		require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code, recorder.Body.String())
	})

	t.Run("missing file", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		require.NoError(t, writer.Close())
		request := httptest.NewRequest(http.MethodPost, "/api/v1/uploads", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("Authorization", "Bearer "+f.userToken)
		recorder := attachmentHTTPServe(f, request)
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	})
}

func TestAttachmentHTTPRoundTrip(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	requisitionID, _ := f.createRequisition(t)

	upload := attachmentMultipartRequest(
		t,
		"/api/v1/attachments",
		f.userToken,
		"requisition",
		requisitionID,
		"quote.pdf",
		attachmentHTTPPDF,
	)
	recorder := attachmentHTTPServe(f, upload)
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	created := decodeAttachmentHTTPData(t, recorder)
	require.True(t, strings.HasPrefix(created.ID, "att-"), created.ID)
	require.Equal(t, "requisition", created.RefName)
	require.Equal(t, requisitionID, created.RefID)
	require.Equal(t, "quote.pdf", created.FileName)
	require.Equal(t, "application/pdf", created.ContentType)
	require.Equal(t, int64(len(attachmentHTTPPDF)), created.FileSize)
	require.Len(t, created.Checksum, 64)
	require.Equal(t, attachmentHTTPUserID, created.UploadedBy)
	require.Equal(t, "/api/v1/attachments/"+created.ID+"/content", created.URL)
	require.False(t, created.CreatedAt.IsZero())

	// Metadata carries a freshly derived URL and no storage key.
	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/attachments/"+created.ID, nil)
	getRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, getRequest)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Equal(t, created.URL, decodeAttachmentHTTPData(t, recorder).URL)
	require.NotContains(t, recorder.Body.String(), "storage_key")

	listURL := "/api/v1/attachments?ref_name=requisition&ref_id=" + requisitionID + "&limit=5&offset=0"
	listRequest := httptest.NewRequest(http.MethodGet, listURL, nil)
	listRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, listRequest)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var listEnvelope struct {
		Data       []attachmentHTTPData `json:"data"`
		Pagination struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		} `json:"pagination"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &listEnvelope), recorder.Body.String())
	require.Len(t, listEnvelope.Data, 1)
	require.Equal(t, created.ID, listEnvelope.Data[0].ID)
	require.Equal(t, 5, listEnvelope.Pagination.Limit)
	require.Equal(t, 0, listEnvelope.Pagination.Offset)

	downloadRequest := httptest.NewRequest(http.MethodGet, created.URL, nil)
	downloadRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, downloadRequest)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)
	require.Equal(t, attachmentHTTPPDF, body)
	require.Equal(t, "application/pdf", recorder.Header().Get("Content-Type"))
	require.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
	require.Equal(t, "inline", recorder.Header().Get("Content-Disposition"))

	rangeRequest := httptest.NewRequest(http.MethodGet, created.URL, nil)
	rangeRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	rangeRequest.Header.Set("Range", "bytes=0-3")
	recorder = attachmentHTTPServe(f, rangeRequest)
	require.Equal(t, http.StatusPartialContent, recorder.Code, recorder.Body.String())
	body, err = io.ReadAll(recorder.Body)
	require.NoError(t, err)
	require.Equal(t, attachmentHTTPPDF[:4], body)

	// The fixture allows one attachment per parent, so a second upload to the
	// same requisition is rejected with 409.
	recorder = attachmentHTTPServe(f, attachmentMultipartRequest(
		t,
		"/api/v1/attachments",
		f.userToken,
		"requisition",
		requisitionID,
		"second.pdf",
		attachmentHTTPPDF,
	))
	require.Equal(t, http.StatusConflict, recorder.Code, recorder.Body.String())

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/attachments/"+created.ID, nil)
	deleteRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, deleteRequest)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	missingRequest := httptest.NewRequest(http.MethodGet, "/api/v1/attachments/"+created.ID, nil)
	missingRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, missingRequest)
	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.NotContains(t, recorder.Body.String(), f.fixture.LocalRoot)
}

func TestAttachmentHTTPRequisitionItemUpload(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)
	requisitionID, itemID := f.createRequisition(t)

	recorder := attachmentHTTPServe(f, attachmentMultipartRequest(
		t,
		"/api/v1/attachments",
		f.userToken,
		"requisition_item",
		itemID,
		"line-note.pdf",
		attachmentHTTPPDF,
	))
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	created := decodeAttachmentHTTPData(t, recorder)
	require.Equal(t, "requisition_item", created.RefName)
	require.Equal(t, itemID, created.RefID)
	require.NotEqual(t, requisitionID, created.RefID)
	require.Equal(t, "application/pdf", created.ContentType)
	require.Len(t, created.Checksum, 64)
	require.NotContains(t, recorder.Body.String(), "storage_key")

	listRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/attachments?ref_name=requisition_item&ref_id="+itemID,
		nil,
	)
	listRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, listRequest)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var listEnvelope struct {
		Data []attachmentHTTPData `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &listEnvelope), recorder.Body.String())
	require.Len(t, listEnvelope.Data, 1)
	require.Equal(t, created.ID, listEnvelope.Data[0].ID)

	// The two ref types are independent even though both are authorised
	// through the requisition: the parent's own list is still empty.
	parentListRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/attachments?ref_name=requisition&ref_id="+requisitionID,
		nil,
	)
	parentListRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, parentListRequest)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var parentEnvelope struct {
		Data []attachmentHTTPData `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &parentEnvelope), recorder.Body.String())
	require.Empty(t, parentEnvelope.Data)

	downloadRequest := httptest.NewRequest(http.MethodGet, created.URL, nil)
	downloadRequest.Header.Set("Authorization", "Bearer "+f.userToken)
	recorder = attachmentHTTPServe(f, downloadRequest)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)
	require.Equal(t, attachmentHTTPPDF, body)

	// A line attachment inherits the requisition owner's access rules.
	recorder = attachmentHTTPServe(f, attachmentMultipartRequest(
		t,
		"/api/v1/attachments",
		f.otherToken,
		"requisition_item",
		itemID,
		"line-note.pdf",
		attachmentHTTPPDF,
	))
	require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())

	recorder = attachmentHTTPServe(f, attachmentMultipartRequest(
		t,
		"/api/v1/attachments",
		f.userToken,
		"requisition_item",
		"reqitem-0000000000000000",
		"line-note.pdf",
		attachmentHTTPPDF,
	))
	require.Equal(t, http.StatusNotFound, recorder.Code, recorder.Body.String())
}

func TestAttachmentHTTPUploadFailures(t *testing.T) {
	f := setupAttachmentHTTPFixture(t)

	t.Run("unsupported content type", func(t *testing.T) {
		requisitionID, _ := f.createRequisition(t)
		recorder := attachmentHTTPServe(f, attachmentMultipartRequest(
			t,
			"/api/v1/attachments",
			f.userToken,
			"requisition",
			requisitionID,
			"page.html",
			[]byte("<html><body>not a document</body></html>"),
		))
		require.Equal(t, http.StatusUnsupportedMediaType, recorder.Code, recorder.Body.String())
	})

	t.Run("oversize", func(t *testing.T) {
		requisitionID, _ := f.createRequisition(t)
		recorder := attachmentHTTPServe(f, attachmentMultipartRequest(
			t,
			"/api/v1/attachments",
			f.userToken,
			"requisition",
			requisitionID,
			"large.pdf",
			append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("x"), 8<<10)...),
		))
		require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code, recorder.Body.String())
	})

	t.Run("forbidden parent", func(t *testing.T) {
		requisitionID, _ := f.createRequisition(t)
		recorder := attachmentHTTPServe(f, attachmentMultipartRequest(
			t,
			"/api/v1/attachments",
			f.otherToken,
			"requisition",
			requisitionID,
			"quote.pdf",
			attachmentHTTPPDF,
		))
		require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
	})

	t.Run("unknown parent", func(t *testing.T) {
		recorder := attachmentHTTPServe(f, attachmentMultipartRequest(
			t,
			"/api/v1/attachments",
			f.userToken,
			"requisition",
			"req-0000000000000000",
			"quote.pdf",
			attachmentHTTPPDF,
		))
		require.Equal(t, http.StatusNotFound, recorder.Code, recorder.Body.String())
	})

	t.Run("missing auth", func(t *testing.T) {
		request := attachmentMultipartRequest(
			t,
			"/api/v1/attachments",
			"",
			"requisition",
			"req-0000000000000000",
			"quote.pdf",
			attachmentHTTPPDF,
		)
		recorder := attachmentHTTPServe(f, request)
		require.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
	})

	t.Run("missing ref fields", func(t *testing.T) {
		recorder := attachmentHTTPServe(f, attachmentMultipartRequest(
			t,
			"/api/v1/attachments",
			f.userToken,
			"requisition",
			"",
			"quote.pdf",
			attachmentHTTPPDF,
		))
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	})
}
