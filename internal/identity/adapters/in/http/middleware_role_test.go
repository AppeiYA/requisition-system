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
	"github.com/gin-gonic/gin"
)

// tokenServiceWithRole models a stale access token: its role claim is fixed at
// issuance while the database user can change independently.
func tokenServiceWithRole(role string) *identitymocks.MockAccessTokenService {
	svc := identitymocks.NewMockAccessTokenService()
	svc.ValidateAccessTokenFunc = func(_ context.Context, _ string) (*identityout.TokenModel, error) {
		return &identityout.TokenModel{ID: "usr-1", Role: role}, nil
	}
	return svc
}

func serveAuthenticated(router *gin.Engine, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

// A demoted admin still carries role=admin in an unexpired token. Authorising
// from the token would keep granting access until it expired.
func TestAuthMiddleware_DemotedAdminLosesAccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mw := newTestAuthMiddleware(t, tokenServiceWithRole("admin"), shareddomain.RoleUser)
	router := gin.New()
	router.GET("/admin", mw.Handle(), mw.RequireRole("admin"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := serveAuthenticated(router, "stale-admin-token")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

// The inverse: a promotion takes effect immediately, without waiting for the
// access token to be reissued.
func TestAuthMiddleware_PromotedAdminGainsAccess(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mw := newTestAuthMiddleware(t, tokenServiceWithRole("user"), shareddomain.RoleAdmin)
	router := gin.New()
	var capturedRole string
	router.GET("/admin", mw.Handle(), mw.RequireRole("admin"), func(c *gin.Context) {
		capturedRole, _ = GetUserRole(c)
		c.Status(http.StatusOK)
	})

	rec := serveAuthenticated(router, "stale-user-token")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusOK, rec.Body.String())
	}
	if capturedRole != "admin" {
		t.Errorf("captured role = %q, want %q", capturedRole, "admin")
	}
}

func TestAuthMiddleware_InactiveUserIsRejected(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	user := newTestDBUser(t, shareddomain.RoleAdmin, identitydomain.SUSPENDED.Value())
	mw := newTestAuthMiddlewareWithUser(t, tokenServiceWithRole("admin"), user)
	router := gin.New()
	router.GET("/admin", mw.Handle(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := serveAuthenticated(router, "suspended-admin-token")

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

// IsAdmin is a separate middleware used by some routes, so it must apply the
// same database-role checks as Handle + RequireRole.
func TestIsAdmin_UsesDatabaseRole(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name      string
		tokenRole string
		dbRole    shareddomain.Role
		status    string
		want      int
	}{
		{
			name:      "demoted admin loses access",
			tokenRole: "admin",
			dbRole:    shareddomain.RoleUser,
			status:    identitydomain.ACTIVE.Value(),
			want:      http.StatusForbidden,
		},
		{
			name:      "promoted admin gains access",
			tokenRole: "user",
			dbRole:    shareddomain.RoleAdmin,
			status:    identitydomain.ACTIVE.Value(),
			want:      http.StatusOK,
		},
		{
			name:      "inactive admin loses access",
			tokenRole: "admin",
			dbRole:    shareddomain.RoleAdmin,
			status:    identitydomain.INACTIVE.Value(),
			want:      http.StatusForbidden,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			user := newTestDBUser(t, tc.dbRole, tc.status)
			mw := newTestAuthMiddlewareWithUser(t, tokenServiceWithRole(tc.tokenRole), user)
			router := gin.New()
			router.GET("/admin", mw.IsAdmin(), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			rec := serveAuthenticated(router, "stale-token")

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d (%s)", rec.Code, tc.want, rec.Body.String())
			}
		})
	}
}

func TestIsAdmin_MissingToken(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	mw := newTestAuthMiddleware(t, tokenServiceWithRole("admin"), shareddomain.RoleAdmin)
	router := gin.New()
	router.GET("/admin", mw.IsAdmin(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	rec := serveAuthenticated(router, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d (%s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}
