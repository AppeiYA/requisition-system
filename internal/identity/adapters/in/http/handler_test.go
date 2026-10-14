package identityhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubLogin struct {
	resp *identityin.TokenResponse
	err  error
	got  identityin.LoginUserRequest
}

func (s *stubLogin) Execute(ctx context.Context, req identityin.LoginUserRequest) (*identityin.TokenResponse, error) {
	s.got = req
	return s.resp, s.err
}

type stubRefresh struct {
	resp *identityin.TokenResponse
	err  error
	got  identityin.RefreshTokenRequest
}

func (s *stubRefresh) Execute(ctx context.Context, req identityin.RefreshTokenRequest) (*identityin.TokenResponse, error) {
	s.got = req
	return s.resp, s.err
}

type stubLogout struct {
	err error
	got identityin.LogoutRequest
}

func (s *stubLogout) Execute(ctx context.Context, req identityin.LogoutRequest) error {
	s.got = req
	return s.err
}

type stubGetMe struct {
	resp *identityin.UserResponse
	err  error
	got  string
}

func (s *stubGetMe) Execute(ctx context.Context, userID string) (*identityin.UserResponse, error) {
	s.got = userID
	return s.resp, s.err
}

func newTestRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	h.Register(router.Group("/api/v1"))

	return router
}

func doJSON(t *testing.T, router *gin.Engine, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

type envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) envelope {
	t.Helper()

	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}

	return env
}

func newHandler(login *stubLogin, refresh *stubRefresh, logout *stubLogout, me *stubGetMe, accessTokens identityout.AccessTokenService) *Handler {
	return NewHandler(login, refresh, logout, me, accessTokens)
}

func TestHandler_Login_Success(t *testing.T) {
	t.Parallel()

	login := &stubLogin{resp: &identityin.TokenResponse{AccessToken: "access", RefreshToken: "refresh"}}
	router := newTestRouter(newHandler(login, &stubRefresh{}, &stubLogout{}, &stubGetMe{}, identitymocks.NewMockAccessTokenService()))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/login", `{"email":"user@example.com","password":"Password123$"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if login.got.Email != "user@example.com" {
		t.Errorf("Email = %q, want user@example.com", login.got.Email)
	}

	env := decodeEnvelope(t, rec)
	if !env.Success {
		t.Error("expected success=true")
	}
}

func TestHandler_Login_InvalidBody(t *testing.T) {
	t.Parallel()

	router := newTestRouter(newHandler(&stubLogin{}, &stubRefresh{}, &stubLogout{}, &stubGetMe{}, identitymocks.NewMockAccessTokenService()))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/login", `{"email":"not-an-email"}`, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestHandler_Login_ErrorMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"unauthorized", customerrors.NewUnauthorizedError("bad credentials"), http.StatusUnauthorized},
		{"internal", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			login := &stubLogin{err: tt.err}
			router := newTestRouter(newHandler(login, &stubRefresh{}, &stubLogout{}, &stubGetMe{}, identitymocks.NewMockAccessTokenService()))

			rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/login", `{"email":"user@example.com","password":"Password123$"}`, nil)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestHandler_Refresh_Success(t *testing.T) {
	t.Parallel()

	refresh := &stubRefresh{resp: &identityin.TokenResponse{AccessToken: "access", RefreshToken: "refresh"}}
	router := newTestRouter(newHandler(&stubLogin{}, refresh, &stubLogout{}, &stubGetMe{}, identitymocks.NewMockAccessTokenService()))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"old-token"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if refresh.got.RefreshToken != "old-token" {
		t.Errorf("RefreshToken = %q, want old-token", refresh.got.RefreshToken)
	}
}

func TestHandler_Refresh_InvalidToken(t *testing.T) {
	t.Parallel()

	refresh := &stubRefresh{err: identityout.ErrRefreshSessionNotFound}
	router := newTestRouter(newHandler(&stubLogin{}, refresh, &stubLogout{}, &stubGetMe{}, identitymocks.NewMockAccessTokenService()))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/refresh", `{"refresh_token":"stale"}`, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestHandler_Logout_Success(t *testing.T) {
	t.Parallel()

	logout := &stubLogout{}
	router := newTestRouter(newHandler(&stubLogin{}, &stubRefresh{}, logout, &stubGetMe{}, identitymocks.NewMockAccessTokenService()))

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/logout", `{"refresh_token":"token"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if logout.got.RefreshToken != "token" {
		t.Errorf("RefreshToken = %q, want token", logout.got.RefreshToken)
	}
}

func TestHandler_Me_RequiresValidToken(t *testing.T) {
	t.Parallel()

	accessTokens := identitymocks.NewMockAccessTokenService()
	accessTokens.ValidateAccessTokenFunc = func(ctx context.Context, token string) (*identityout.TokenModel, error) {
		if token == "good-token" {
			return &identityout.TokenModel{ID: identitymocks.DefaultID}, nil
		}
		return nil, identityout.ErrRefreshSessionNotFound
	}

	me := &stubGetMe{resp: &identityin.UserResponse{ID: identitymocks.DefaultID}}
	router := newTestRouter(newHandler(&stubLogin{}, &stubRefresh{}, &stubLogout{}, me, accessTokens))

	// missing header
	if rec := doJSON(t, router, http.MethodGet, "/api/v1/me", "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing header: status = %d, want 401", rec.Code)
	}

	// invalid token
	rec := doJSON(t, router, http.MethodGet, "/api/v1/me", "", map[string]string{"Authorization": "Bearer bad"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token: status = %d, want 401", rec.Code)
	}

	// valid token
	rec = doJSON(t, router, http.MethodGet, "/api/v1/me", "", map[string]string{"Authorization": "Bearer good-token"})
	if rec.Code != http.StatusOK {
		t.Fatalf("valid token: status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if me.got != identitymocks.DefaultID {
		t.Errorf("userID = %q, want %q", me.got, identitymocks.DefaultID)
	}
}
