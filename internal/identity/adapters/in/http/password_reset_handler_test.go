package identityhttp

import (
	"context"
	"net/http"
	"testing"

	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubForgotPassword struct {
	email string
	err   error
	calls int
}

func (s *stubForgotPassword) Execute(ctx context.Context, email string) error {
	s.email = email
	s.calls++
	return s.err
}

type stubResetPassword struct {
	token    string
	password string
	err      error
	calls    int
}

func (s *stubResetPassword) Execute(ctx context.Context, token, password string) error {
	s.token = token
	s.password = password
	s.calls++
	return s.err
}

func newPasswordResetRouter(forgot *stubForgotPassword, reset *stubResetPassword) *gin.Engine {
	port := &identityin.IdentityPortIn{
		RequestPasswordResetUsecase: forgot,
		ResetPasswordUsecase:        reset,
	}
	auth := NewAuthMiddleware(identitymocks.NewMockAccessTokenService(), identitymocks.NewMockUsersRepository())
	return newTestRouter(NewHandler(port, auth))
}

func Test_ForgotPasswordHandler_GenericResponse(t *testing.T) {
	t.Parallel()

	forgot := &stubForgotPassword{}
	router := newPasswordResetRouter(forgot, &stubResetPassword{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/forgot-password",
		`{"email":"user@example.com"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	body := decodeEnvelope(t, rec)
	if !body.Success || body.Message != forgotPasswordMessage {
		t.Fatalf("unexpected response: %+v", body)
	}
	if forgot.calls != 1 || forgot.email != "user@example.com" {
		t.Fatalf("usecase calls = %d, email = %q", forgot.calls, forgot.email)
	}
}

func Test_ForgotPasswordHandler_InvalidInput(t *testing.T) {
	t.Parallel()

	forgot := &stubForgotPassword{}
	router := newPasswordResetRouter(forgot, &stubResetPassword{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/forgot-password",
		`{"email":"not-an-email"}`, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	if forgot.calls != 0 {
		t.Fatalf("usecase was called for invalid input")
	}
}

func Test_ForgotPasswordHandler_UsecaseError(t *testing.T) {
	t.Parallel()

	forgot := &stubForgotPassword{err: customerrors.NewInternalServerError("database unavailable")}
	router := newPasswordResetRouter(forgot, &stubResetPassword{})

	rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/forgot-password",
		`{"email":"user@example.com"}`, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", rec.Code, rec.Body.String())
	}
}

func Test_ResetPasswordHandler_SuccessAndGenericError(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		reset := &stubResetPassword{}
		router := newPasswordResetRouter(&stubForgotPassword{}, reset)

		rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/reset-password",
			`{"token":"raw-token","password":"NewPassword123$"}`, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
		}
		body := decodeEnvelope(t, rec)
		if !body.Success || body.Message != "password reset successful" {
			t.Fatalf("unexpected response: %+v", body)
		}
		if reset.token != "raw-token" || reset.password != "NewPassword123$" {
			t.Fatalf("unexpected usecase args: %+v", reset)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		reset := &stubResetPassword{err: customerrors.NewBadRequestError("invalid or expired password reset token")}
		router := newPasswordResetRouter(&stubForgotPassword{}, reset)

		rec := doJSON(t, router, http.MethodPost, "/api/v1/auth/reset-password",
			`{"token":"bad","password":"NewPassword123$"}`, nil)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
		}
		body := decodeEnvelope(t, rec)
		if body.Success || body.Message != "invalid or expired password reset token" {
			t.Fatalf("unexpected response: %+v", body)
		}
	})
}
