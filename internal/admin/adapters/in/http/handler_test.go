package adminhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubInvite struct {
	res *adminin.InviteMemberResult
	err error
	got adminin.InviteMemberCommand
}

func (s *stubInvite) Execute(ctx context.Context, cmd adminin.InviteMemberCommand) (*adminin.InviteMemberResult, error) {
	s.got = cmd
	return s.res, s.err
}

func withUserID(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("identity.user_id", userID)
		c.Next()
	}
}

func newTestRouter(invite adminin.InviteMemberUseCase, auth gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	port := &adminin.AdminPortIn{
		InviteMember: invite,
	}
	NewHandler(port).Register(router.Group("/api/v1"), auth)

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

func Test_Invite_Success(t *testing.T) {
	t.Parallel()

	invite := &stubInvite{res: &adminin.InviteMemberResult{InvitationID: "inv-1"}}
	router := newTestRouter(invite, withUserID("usr-1"))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/admin/invitations", `{"email":"invitee@example.com","firstName":"Jane","lastName":"Doe","phone":"+14155551234","staffId":"STF-1234","departmentId":"dpt-1","jobTitle":"Software Engineer"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if invite.got.InviterID != "usr-1" {
		t.Errorf("expected inviter %q, got %q", "usr-1", invite.got.InviterID)
	}
	if invite.got.Email != "invitee@example.com" {
		t.Errorf("expected email %q, got %q", "invitee@example.com", invite.got.Email)
	}
	if invite.got.FirstName != "Jane" || invite.got.LastName != "Doe" || invite.got.Phone != "+14155551234" {
		t.Errorf("expected prefill fields to be forwarded, got %+v", invite.got)
	}
	if invite.got.StaffID != "STF-1234" {
		t.Errorf("expected staff id to be forwarded, got %q", invite.got.StaffID)
	}
	if invite.got.DepartmentID != "dpt-1" {
		t.Errorf("expected department id to be forwarded, got %q", invite.got.DepartmentID)
	}
	if invite.got.JobTitle != "Software Engineer" {
		t.Errorf("expected job title to be forwarded, got %q", invite.got.JobTitle)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			InvitationID string `json:"InvitationID"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success {
		t.Error("expected success=true")
	}
}

func Test_Invite_Unauthorized(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubInvite{}, func(c *gin.Context) { c.Next() })

	rec := doJSON(t, router, http.MethodPost, "/api/v1/admin/invitations", `{"email":"invitee@example.com"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body.String())
	}
}

func Test_Invite_InvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(&stubInvite{}, withUserID("usr-1"))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/admin/invitations", `{"email":"not-an-email"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
}

func Test_Invite_Conflict(t *testing.T) {
	t.Parallel()

	invite := &stubInvite{err: customerrors.NewConflictError("user with this email already exists")}
	router := newTestRouter(invite, withUserID("usr-1"))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/admin/invitations", `{"email":"invitee@example.com","firstName":"Jane","lastName":"Doe","departmentId":"dpt-1","jobTitle":"Software Engineer"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", rec.Code, rec.Body.String())
	}
}
