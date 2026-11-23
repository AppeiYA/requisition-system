package identityhttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	identitydomain "github.com/AppeiYA/requisition-system/internal/identity/domain"
	identitymocks "github.com/AppeiYA/requisition-system/internal/identity/mocks"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_Handle_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockTokens := identitymocks.NewMockAccessTokenService()
	mockTokens.ValidateAccessTokenFunc = func(ctx context.Context, token string) (*identityout.TokenModel, error) {
		if token != "valid-token" {
			t.Fatalf("unexpected token: %s", token)
		}
		return &identityout.TokenModel{ID: "usr-1", Role: "admin", Email: "test@example.com"}, nil
	}

	mw := newTestAuthMiddleware(t, mockTokens, shareddomain.RoleAdmin)

	router := gin.New()
	var capturedID, capturedRole string
	router.GET("/test", mw.Handle(), func(c *gin.Context) {
		id, _ := GetUserID(c)
		role, _ := GetUserRole(c)
		capturedID = id
		capturedRole = role
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if capturedID != "usr-1" {
		t.Errorf("captured ID = %q, want %q", capturedID, "usr-1")
	}
	if capturedRole != "admin" {
		t.Errorf("captured Role = %q, want %q", capturedRole, "admin")
	}
}

func TestAuthMiddleware_Handle_MissingOrInvalidHeader(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockTokens := identitymocks.NewMockAccessTokenService()
	mw := newTestAuthMiddleware(t, mockTokens, shareddomain.RoleUser)

	cases := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"wrong prefix", "Basic xyz"},
		{"empty bearer", "Bearer "},
		{"only whitespace bearer", "Bearer    "},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/test", mw.Handle(), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthMiddleware_Handle_ValidationError(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockTokens := identitymocks.NewMockAccessTokenService()
	mockTokens.ValidateAccessTokenFunc = func(ctx context.Context, token string) (*identityout.TokenModel, error) {
		return nil, customerrors.NewUnauthorizedError("token expired")
	}

	mw := newTestAuthMiddleware(t, mockTokens, shareddomain.RoleUser)

	router := gin.New()
	router.GET("/test", mw.Handle(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_RequireRole_Success(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockTokens := identitymocks.NewMockAccessTokenService()
	mockTokens.ValidateAccessTokenFunc = func(ctx context.Context, token string) (*identityout.TokenModel, error) {
		return &identityout.TokenModel{ID: "usr-1", Role: "admin"}, nil
	}

	mw := newTestAuthMiddleware(t, mockTokens, shareddomain.RoleAdmin)

	router := gin.New()
	router.GET("/admin-only", mw.Handle(), mw.RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddleware_RequireRole_Forbidden(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mockTokens := identitymocks.NewMockAccessTokenService()
	mockTokens.ValidateAccessTokenFunc = func(ctx context.Context, token string) (*identityout.TokenModel, error) {
		return &identityout.TokenModel{ID: "usr-1", Role: "user"}, nil
	}

	mw := newTestAuthMiddleware(t, mockTokens, shareddomain.RoleUser)

	router := gin.New()
	router.GET("/admin-only", mw.Handle(), mw.RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAuthMiddleware_IsAdmin(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	t.Run("admin allows", func(t *testing.T) {
		mockTokens := identitymocks.NewMockAccessTokenService()
		mockTokens.ValidateAccessTokenFunc = func(ctx context.Context, token string) (*identityout.TokenModel, error) {
			return &identityout.TokenModel{ID: "usr-1", Role: "admin"}, nil
		}

		mw := newTestAuthMiddleware(t, mockTokens, shareddomain.RoleAdmin)

		router := gin.New()
		router.GET("/admin", mw.IsAdmin(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("non-admin forbids", func(t *testing.T) {
		mockTokens := identitymocks.NewMockAccessTokenService()
		mockTokens.ValidateAccessTokenFunc = func(ctx context.Context, token string) (*identityout.TokenModel, error) {
			return &identityout.TokenModel{ID: "usr-1", Role: "user"}, nil
		}

		mw := newTestAuthMiddleware(t, mockTokens, shareddomain.RoleUser)

		router := gin.New()
		router.GET("/admin", mw.IsAdmin(), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		req.Header.Set("Authorization", "Bearer user-token")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

func newTestAuthMiddleware(t *testing.T, accessTokens identityout.AccessTokenService, role shareddomain.Role) *AuthMiddleware {
	t.Helper()
	return newTestAuthMiddlewareWithUser(t, accessTokens, newTestDBUser(t, role, identitydomain.ACTIVE.Value()))
}

func newTestAuthMiddlewareWithUser(t *testing.T, accessTokens identityout.AccessTokenService, user *identitydomain.User) *AuthMiddleware {
	t.Helper()

	userRepo := identitymocks.NewMockUsersRepository()
	userRepo.GetUserByIDFunc = func(ctx context.Context, id string) (*identitydomain.User, error) {
		return user, nil
	}
	return NewAuthMiddleware(accessTokens, userRepo)
}

// newTestDBUser builds a persisted user with an explicit role and status. The
// role and status deliberately come from here rather than from the token, as
// they would in production.
func newTestDBUser(t *testing.T, role shareddomain.Role, status string) *identitydomain.User {
	t.Helper()

	base := identitymocks.ValidUser()
	user, err := identitydomain.ReconstituteUser(
		base.ID(),
		base.FirstName().Value(),
		base.LastName().Value(),
		base.Phone().Value(),
		base.StaffID(),
		base.Email().Value(),
		base.PasswordHash().Value(),
		status,
		role.String(),
		base.CreatedAt(),
		base.UpdatedAt(),
	)
	if err != nil {
		t.Fatalf("build test user: %v", err)
	}

	return user
}
