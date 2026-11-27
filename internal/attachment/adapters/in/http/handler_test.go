package attachmenthttp

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

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

const (
	testAttachmentID   = "att-0123456789abcdef"
	testRequisitionID  = "req-0123456789abcdef"
	testAttachmentUser = "usr-0123456789abcdef"
)

var testPDF = []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\n%%EOF\n")

type stubUploadAttachment struct {
	res    *domain.Attachment
	err    error
	got    attachmentin.UploadAttachmentCommand
	called bool
}

func (s *stubUploadAttachment) Execute(_ context.Context, cmd attachmentin.UploadAttachmentCommand) (*domain.Attachment, error) {
	s.called = true
	s.got = cmd
	if cmd.Body != nil {
		_, _ = io.Copy(io.Discard, cmd.Body)
	}
	return s.res, s.err
}

type stubStageUpload struct {
	res    *domain.StagedUpload
	err    error
	got    attachmentin.StageUploadCommand
	called bool
}

func (s *stubStageUpload) Execute(_ context.Context, cmd attachmentin.StageUploadCommand) (*domain.StagedUpload, error) {
	s.called = true
	s.got = cmd
	if cmd.Body != nil {
		_, _ = io.Copy(io.Discard, cmd.Body)
	}
	return s.res, s.err
}

type stubDownloadAttachment struct {
	res    *attachmentin.Download
	err    error
	got    attachmentin.DownloadAttachmentQuery
	called bool
}

func (s *stubDownloadAttachment) Execute(_ context.Context, query attachmentin.DownloadAttachmentQuery) (*attachmentin.Download, error) {
	s.called = true
	s.got = query
	return s.res, s.err
}

type stubGetAttachment struct {
	res *domain.Attachment
	err error
	got attachmentin.GetAttachmentQuery
}

func (s *stubGetAttachment) Execute(_ context.Context, query attachmentin.GetAttachmentQuery) (*domain.Attachment, error) {
	s.got = query
	return s.res, s.err
}

type stubListAttachments struct {
	res []*domain.Attachment
	err error
	got attachmentin.ListAttachmentsQuery
}

func (s *stubListAttachments) Execute(_ context.Context, query attachmentin.ListAttachmentsQuery) ([]*domain.Attachment, error) {
	s.got = query
	return s.res, s.err
}

type stubDeleteAttachment struct {
	err error
	got attachmentin.DeleteAttachmentCommand
}

func (s *stubDeleteAttachment) Execute(_ context.Context, cmd attachmentin.DeleteAttachmentCommand) error {
	s.got = cmd
	return s.err
}

type stubDeleteAttachmentsForRef struct{}

func (s *stubDeleteAttachmentsForRef) Execute(_ context.Context, _ attachmentin.DeleteForRefCommand) (int, error) {
	return 0, nil
}

// readSeekCloser lets a bytes.Reader exercise the ServeContent range path.
type readSeekCloser struct {
	*bytes.Reader
}

func (r *readSeekCloser) Close() error { return nil }

func newTestAttachment(t *testing.T) *domain.Attachment {
	t.Helper()

	ref, err := domain.NewAttachmentRef(domain.RefNameRequisition, testRequisitionID)
	if err != nil {
		t.Fatalf("build ref: %v", err)
	}
	fileName, err := domain.NewAttachmentFileName("quote.pdf")
	if err != nil {
		t.Fatalf("build filename: %v", err)
	}
	attachment, err := domain.NewAttachment(
		testAttachmentID,
		ref,
		"requisition/"+testRequisitionID+"/"+testAttachmentID,
		fileName,
		domain.ContentTypePDF,
		int64(len(testPDF)),
		strings.Repeat("a", domain.ChecksumLength),
		testAttachmentUser,
	)
	if err != nil {
		t.Fatalf("build attachment: %v", err)
	}
	return attachment
}

func withTestActor() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(identityhttp.UserIDContextKey, testAttachmentUser)
		c.Set(identityhttp.UserRoleContextKey, "staff")
		c.Next()
	}
}

func withoutTestActor() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

func newTestRouter(
	port *attachmentin.AttachmentPortIn,
	maxSizeBytes int64,
	baseURL string,
	auth gin.HandlerFunc,
) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	NewAttachmentHandler(port, maxSizeBytes, baseURL).Register(router.Group("/api/v1"), auth)
	return router
}

func testPortIn(
	upload attachmentin.UploadAttachmentUsecaseInt,
	stage attachmentin.StageUploadUsecaseInt,
	download attachmentin.DownloadAttachmentUsecaseInt,
	get attachmentin.GetAttachmentUsecaseInt,
	list attachmentin.ListAttachmentsUsecaseInt,
	deleteAttachment attachmentin.DeleteAttachmentUsecaseInt,
) *attachmentin.AttachmentPortIn {
	if stage == nil {
		stage = &stubStageUpload{}
	}
	return attachmentin.NewAttachmentPortIn(
		upload, stage, download, get, list, deleteAttachment, &stubDeleteAttachmentsForRef{},
	)
}

func multipartUploadRequest(t *testing.T, fields map[string]string, fileName string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}
	if fileName != "" {
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("write file part: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func serveAttachment(t *testing.T, router *gin.Engine, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func newTestStagedUpload(t *testing.T) *domain.StagedUpload {
	t.Helper()
	name, err := domain.NewAttachmentFileName("quote.pdf")
	if err != nil {
		t.Fatalf("build staged file name: %v", err)
	}
	now := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	upload, err := domain.NewStagedUpload(
		"upl-0123456789abcdef",
		testAttachmentUser,
		"staging/upl-0123456789abcdef",
		name,
		domain.ContentTypePDF,
		int64(len(testPDF)),
		strings.Repeat("a", domain.ChecksumLength),
		time.Hour,
		now,
	)
	if err != nil {
		t.Fatalf("build staged upload: %v", err)
	}
	return upload
}

func Test_StageUpload_Success(t *testing.T) {
	t.Parallel()

	stage := &stubStageUpload{res: newTestStagedUpload(t)}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, stage, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"http://localhost:8080/api/v1",
		withTestActor(),
	)

	request := multipartUploadRequest(t, nil, "quote.pdf", testPDF)
	request.URL.Path = "/api/v1/uploads"
	recorder := serveAttachment(t, router, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	if !stage.called {
		t.Fatal("expected stage upload use case to be called")
	}
	if stage.got.ActorID != testAttachmentUser || stage.got.FileName != "quote.pdf" {
		t.Errorf("stage fields not forwarded: %+v", stage.got)
	}

	var body struct {
		Data StagedUploadResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.ID != "upl-0123456789abcdef" {
		t.Errorf("id = %q, want staged upload ID", body.Data.ID)
	}
	if body.Data.Status != "pending" || body.Data.ExpiresAt.IsZero() {
		t.Errorf("unexpected staged metadata: %+v", body.Data)
	}
	if strings.Contains(recorder.Body.String(), "staging/") {
		t.Error("response exposed the staging storage path")
	}
}

func Test_StageUpload_RequiresActor(t *testing.T) {
	t.Parallel()

	stage := &stubStageUpload{}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, stage, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withoutTestActor(),
	)

	request := multipartUploadRequest(t, nil, "quote.pdf", testPDF)
	request.URL.Path = "/api/v1/uploads"
	recorder := serveAttachment(t, router, request)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
	if stage.called {
		t.Error("stage upload use case ran without an actor")
	}
}

func Test_UploadAttachment_Success(t *testing.T) {
	t.Parallel()

	upload := &stubUploadAttachment{res: newTestAttachment(t)}
	router := newTestRouter(
		testPortIn(upload, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"http://localhost:8080/api/v1",
		withTestActor(),
	)

	request := multipartUploadRequest(t, map[string]string{
		"ref_name": "requisition",
		"ref_id":   testRequisitionID,
	}, "quote.pdf", testPDF)
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body.String())
	}
	if !upload.called {
		t.Fatal("expected upload use case to be called")
	}
	if upload.got.ActorID != testAttachmentUser || upload.got.ActorRole != "staff" {
		t.Errorf("actor not forwarded: %+v", upload.got)
	}
	if upload.got.RefName != "requisition" || upload.got.RefID != testRequisitionID || upload.got.FileName != "quote.pdf" {
		t.Errorf("upload fields not forwarded: %+v", upload.got)
	}

	var body struct {
		Data AttachmentResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.ID != testAttachmentID {
		t.Errorf("id = %q, want %q", body.Data.ID, testAttachmentID)
	}
	if body.Data.RefName != "requisition" || body.Data.RefID != testRequisitionID {
		t.Errorf("unexpected ref in response: %+v", body.Data)
	}
	if body.Data.URL != "/api/v1/attachments/"+testAttachmentID+"/content" {
		t.Errorf("url = %q", body.Data.URL)
	}
	if body.Data.Status != domain.AttachmentStatusReady.String() {
		t.Errorf("status = %q, want ready", body.Data.Status)
	}
	if body.Data.ContentType != "application/pdf" || body.Data.FileSize != int64(len(testPDF)) {
		t.Errorf("unexpected file metadata: %+v", body.Data)
	}
}

func Test_UploadAttachment_RequiresActor(t *testing.T) {
	t.Parallel()

	upload := &stubUploadAttachment{}
	router := newTestRouter(
		testPortIn(upload, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withoutTestActor(),
	)

	request := multipartUploadRequest(t, map[string]string{
		"ref_name": "requisition",
		"ref_id":   testRequisitionID,
	}, "quote.pdf", testPDF)
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", recorder.Code, recorder.Body.String())
	}
	if upload.called {
		t.Error("upload use case must not run for an unauthenticated request")
	}
}

func Test_UploadAttachment_TooLarge(t *testing.T) {
	t.Parallel()

	upload := &stubUploadAttachment{res: newTestAttachment(t)}
	router := newTestRouter(
		testPortIn(upload, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		64,
		"/api/v1",
		withTestActor(),
	)

	request := multipartUploadRequest(t, map[string]string{
		"ref_name": "requisition",
		"ref_id":   testRequisitionID,
	}, "quote.pdf", bytes.Repeat([]byte("x"), 1024))
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413: %s", recorder.Code, recorder.Body.String())
	}
	if upload.called {
		t.Error("upload use case must not run for an oversized request")
	}
}

func Test_UploadAttachment_UnsupportedType(t *testing.T) {
	t.Parallel()

	upload := &stubUploadAttachment{err: domain.ErrInvalidContentType}
	router := newTestRouter(
		testPortIn(upload, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	request := multipartUploadRequest(t, map[string]string{
		"ref_name": "requisition",
		"ref_id":   testRequisitionID,
	}, "page.html", []byte("<html></html>"))
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("status = %d, want 415: %s", recorder.Code, recorder.Body.String())
	}
}

func Test_UploadAttachment_MissingFields(t *testing.T) {
	t.Parallel()

	upload := &stubUploadAttachment{}
	router := newTestRouter(
		testPortIn(upload, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	request := multipartUploadRequest(t, map[string]string{
		"ref_id": testRequisitionID,
	}, "quote.pdf", testPDF)
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", recorder.Code, recorder.Body.String())
	}
	if upload.called {
		t.Error("upload use case must not run when ref_name is missing")
	}
}

func Test_UploadAttachment_FileBeforeParentFields(t *testing.T) {
	t.Parallel()

	upload := &stubUploadAttachment{}
	router := newTestRouter(
		testPortIn(upload, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "quote.pdf")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(testPDF); err != nil {
		t.Fatalf("Write file: %v", err)
	}
	_ = writer.WriteField("ref_name", "requisition")
	_ = writer.WriteField("ref_id", testRequisitionID)
	_ = writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", recorder.Code, recorder.Body.String())
	}
	if upload.called {
		t.Error("upload use case must not run when file precedes parent fields")
	}
}

func Test_UploadAttachment_LyingContentLength(t *testing.T) {
	t.Parallel()

	upload := &stubUploadAttachment{res: newTestAttachment(t)}
	router := newTestRouter(
		testPortIn(upload, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		64,
		"/api/v1",
		withTestActor(),
	)

	request := multipartUploadRequest(t, map[string]string{
		"ref_name": "requisition",
		"ref_id":   testRequisitionID,
	}, "quote.pdf", bytes.Repeat([]byte("x"), 512))
	request.ContentLength = 10 // Lie about Content-Length
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413: %s", recorder.Code, recorder.Body.String())
	}
}

func Test_DownloadAttachment_StreamsWithRange(t *testing.T) {
	t.Parallel()

	download := &stubDownloadAttachment{res: &attachmentin.Download{
		Attachment: newTestAttachment(t),
		Stream:     &readSeekCloser{Reader: bytes.NewReader(testPDF)},
	}}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, download, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/attachments/"+testAttachmentID+"/content", nil)
	request.Header.Set("Range", "bytes=0-3")
	recorder := serveAttachment(t, router, request)

	if recorder.Code != http.StatusPartialContent {
		t.Fatalf("status = %d, want 206: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Body.String(); got != string(testPDF[:4]) {
		t.Errorf("range body = %q, want %q", got, testPDF[:4])
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/pdf" {
		t.Errorf("Content-Type = %q", got)
	}
	if got := recorder.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q", got)
	}
	if got := recorder.Header().Get("Content-Disposition"); got != "inline" {
		t.Errorf("Content-Disposition = %q", got)
	}
	if download.got.AttachmentID != testAttachmentID || download.got.ActorID != testAttachmentUser {
		t.Errorf("download query not forwarded: %+v", download.got)
	}
}

func Test_DownloadAttachment_IfModifiedSince(t *testing.T) {
	t.Parallel()

	att := newTestAttachment(t)
	download := &stubDownloadAttachment{res: &attachmentin.Download{
		Attachment: att,
		Stream:     &readSeekCloser{Reader: bytes.NewReader(testPDF)},
	}}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, download, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/attachments/"+testAttachmentID+"/content", nil)
	req.Header.Set("If-Modified-Since", att.CreatedAt().Add(time.Minute).UTC().Format(http.TimeFormat))
	recorder := serveAttachment(t, router, req)

	if recorder.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304: %s", recorder.Code, recorder.Body.String())
	}
}

func Test_DownloadAttachment_NotReady(t *testing.T) {
	t.Parallel()

	download := &stubDownloadAttachment{err: domain.ErrAttachmentNotReady}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, download, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	recorder := serveAttachment(t, router, httptest.NewRequest(
		http.MethodGet, "/api/v1/attachments/"+testAttachmentID+"/content", nil,
	))

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", recorder.Code, recorder.Body.String())
	}
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Error != "attachment_not_ready" {
		t.Errorf("error code = %q, want attachment_not_ready", body.Error)
	}
}

func Test_DownloadAttachment_Redirect(t *testing.T) {
	t.Parallel()

	download := &stubDownloadAttachment{res: &attachmentin.Download{
		Attachment: newTestAttachment(t),
		Location:   "https://storage.example.com/signed?token=abc",
	}}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, download, &stubGetAttachment{}, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	recorder := serveAttachment(t, router, httptest.NewRequest(
		http.MethodGet, "/api/v1/attachments/"+testAttachmentID+"/content", nil,
	))

	if recorder.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302: %s", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Location"); got != "https://storage.example.com/signed?token=abc" {
		t.Errorf("Location = %q", got)
	}
	if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}

func Test_GetAttachment_SuccessAndNotFound(t *testing.T) {
	t.Parallel()

	get := &stubGetAttachment{res: newTestAttachment(t)}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, &stubDownloadAttachment{}, get, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	recorder := serveAttachment(t, router, httptest.NewRequest(
		http.MethodGet, "/api/v1/attachments/"+testAttachmentID, nil,
	))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if get.got.AttachmentID != testAttachmentID {
		t.Errorf("get query = %+v", get.got)
	}
	var body struct {
		Data AttachmentResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.URL != "/api/v1/attachments/"+testAttachmentID+"/content" {
		t.Errorf("url = %q", body.Data.URL)
	}
	if body.Data.Status != domain.AttachmentStatusReady.String() {
		t.Errorf("status = %q, want ready", body.Data.Status)
	}

	missing := &stubGetAttachment{err: customerrors.NewNotFoundError("attachment not found")}
	missingRouter := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, &stubDownloadAttachment{}, missing, &stubListAttachments{}, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)
	recorder = serveAttachment(t, missingRouter, httptest.NewRequest(
		http.MethodGet, "/api/v1/attachments/att-missing", nil,
	))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", recorder.Code, recorder.Body.String())
	}
}

func Test_ListAttachments_SuccessAndValidation(t *testing.T) {
	t.Parallel()

	list := &stubListAttachments{res: []*domain.Attachment{newTestAttachment(t)}}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, list, &stubDeleteAttachment{}),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	recorder := serveAttachment(t, router, httptest.NewRequest(
		http.MethodGet,
		"/api/v1/attachments?ref_name=requisition&ref_id="+testRequisitionID+"&limit=5&offset=10",
		nil,
	))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if list.got.RefName != "requisition" || list.got.RefID != testRequisitionID ||
		list.got.Limit != 5 || list.got.Offset != 10 {
		t.Errorf("list query not forwarded: %+v", list.got)
	}

	var body struct {
		Data       []AttachmentResponse `json:"data"`
		Pagination struct {
			Limit  int `json:"limit"`
			Offset int `json:"offset"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data) != 1 || body.Data[0].ID != testAttachmentID {
		t.Errorf("unexpected data: %+v", body.Data)
	}
	if body.Pagination.Limit != 5 || body.Pagination.Offset != 10 {
		t.Errorf("unexpected pagination: %+v", body.Pagination)
	}

	recorder = serveAttachment(t, router, httptest.NewRequest(
		http.MethodGet, "/api/v1/attachments", nil,
	))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", recorder.Code, recorder.Body.String())
	}

	recorder = serveAttachment(t, router, httptest.NewRequest(
		http.MethodGet, "/api/v1/attachments?ref_name=requisition&ref_id="+testRequisitionID+"&limit=1000", nil,
	))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", recorder.Code, recorder.Body.String())
	}
}

func Test_DeleteAttachment_SuccessAndForbidden(t *testing.T) {
	t.Parallel()

	deleteAttachment := &stubDeleteAttachment{}
	router := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, deleteAttachment),
		1<<20,
		"/api/v1",
		withTestActor(),
	)

	recorder := serveAttachment(t, router, httptest.NewRequest(
		http.MethodDelete, "/api/v1/attachments/"+testAttachmentID, nil,
	))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if deleteAttachment.got.AttachmentID != testAttachmentID || deleteAttachment.got.ActorID != testAttachmentUser {
		t.Errorf("delete command not forwarded: %+v", deleteAttachment.got)
	}

	forbidden := &stubDeleteAttachment{err: customerrors.NewForbiddenError("download not allowed")}
	forbiddenRouter := newTestRouter(
		testPortIn(&stubUploadAttachment{}, nil, &stubDownloadAttachment{}, &stubGetAttachment{}, &stubListAttachments{}, forbidden),
		1<<20,
		"/api/v1",
		withTestActor(),
	)
	recorder = serveAttachment(t, forbiddenRouter, httptest.NewRequest(
		http.MethodDelete, "/api/v1/attachments/"+testAttachmentID, nil,
	))
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", recorder.Code, recorder.Body.String())
	}
}
