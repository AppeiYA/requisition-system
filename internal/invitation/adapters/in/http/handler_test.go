package invitationhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubAcceptInvitation struct {
	err error
	got invitationin.AcceptInvitationCommand
}

func (s *stubAcceptInvitation) Execute(ctx context.Context, cmd invitationin.AcceptInvitationCommand) error {
	s.got = cmd
	return s.err
}

type stubValidateInvitation struct {
	res *invitationin.ValidateInvitationResult
	err error
	got string
}

func (s *stubValidateInvitation) Execute(ctx context.Context, token string) (*invitationin.ValidateInvitationResult, error) {
	s.got = token
	return s.res, s.err
}

func newTestRouter(accept invitationin.AcceptInvitationUseCaseInt, validate invitationin.ValidateInvitationUseCaseInt) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	NewInvitationHandler(accept, validate).Register(router.Group("/api/v1"))

	return router
}

func doJSON(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func Test_AcceptInvitation_Success(t *testing.T) {
	t.Parallel()

	accept := &stubAcceptInvitation{}
	router := newTestRouter(accept, &stubValidateInvitation{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/invitations/token-123/accept", `{"phone":"+14155551234","password":"Password123$"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if accept.got.Token != "token-123" {
		t.Errorf("expected token %q, got %q", "token-123", accept.got.Token)
	}
	if accept.got.Phone != "+14155551234" {
		t.Errorf("expected phone to be forwarded, got %q", accept.got.Phone)
	}
}

func Test_AcceptInvitation_InvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubAcceptInvitation{}, &stubValidateInvitation{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/invitations/token-123/accept", `{"firstName":"Jane"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func Test_AcceptInvitation_UsecaseError(t *testing.T) {
	t.Parallel()

	accept := &stubAcceptInvitation{err: customerrors.NewForbiddenError("invitation has expired")}
	router := newTestRouter(accept, &stubValidateInvitation{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/invitations/token-123/accept", `{"firstName":"Jane","lastName":"Doe","password":"Password123$"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", rec.Code, rec.Body.String())
	}
}

func Test_ValidateInvitation_Valid(t *testing.T) {
	t.Parallel()

	validate := &stubValidateInvitation{res: &invitationin.ValidateInvitationResult{
		Valid:        true,
		Email:        "invitee@example.com",
		DepartmentID: "dpt-1",
		JobTitle:     "Engineer",
	}}
	router := newTestRouter(&stubAcceptInvitation{}, validate)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/invitations/token-123/validate", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if validate.got != "token-123" {
		t.Errorf("expected token %q, got %q", "token-123", validate.got)
	}

	var body struct {
		Data ValidateInvitationResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Data.Valid || body.Data.Email != "invitee@example.com" {
		t.Errorf("unexpected data: %+v", body.Data)
	}
	if body.Data.DepartmentID != "dpt-1" || body.Data.JobTitle != "Engineer" {
		t.Errorf("expected department and job title in response, got: %+v", body.Data)
	}
}

func Test_ValidateInvitation_Invalid(t *testing.T) {
	t.Parallel()

	validate := &stubValidateInvitation{res: &invitationin.ValidateInvitationResult{Valid: false}}
	router := newTestRouter(&stubAcceptInvitation{}, validate)

	rec := doJSON(t, router, http.MethodGet, "/api/v1/invitations/token-123/validate", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data ValidateInvitationResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Valid {
		t.Error("expected valid=false")
	}
}
