package identityhttp

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

const (
	UserIDContextKey   = authcontext.UserIDKey
	UserRoleContextKey = authcontext.UserRoleKey
	userIDContextKey   = UserIDContextKey
)

// GetUserID retrieves the authenticated user's ID from the gin context.
func GetUserID(c *gin.Context) (string, bool) {
	return authcontext.GetUserID(c)
}

// GetUserRole retrieves the authenticated user's role from the gin context.
func GetUserRole(c *gin.Context) (string, bool) {
	return authcontext.GetUserRole(c)
}

type AuthMiddleware struct {
	accessTokens identityout.AccessTokenService
}

func NewAuthMiddleware(accessTokens identityout.AccessTokenService) *AuthMiddleware {
	return &AuthMiddleware{
		accessTokens: accessTokens,
	}
}

// Handle validates the bearer access token and stores the authenticated
// user id and role in the request context.
func (m *AuthMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			response.Error(c, "missing or invalid authorization header", http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		model, err := m.accessTokens.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}

		authcontext.SetUser(c, model.ID, model.Role)
		c.Next()
	}
}

// RequireRole checks that the authenticated user has one of the allowed roles.
// It assumes Handle() has already executed upstream.
func (m *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := authcontext.GetUserRole(c)
		if !ok || userRole == "" {
			response.Error(c, "forbidden: admin access required", http.StatusForbidden, nil)
			c.Abort()
			return
		}

		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		response.Error(c, "forbidden: admin access required", http.StatusForbidden, nil)
		c.Abort()
	}
}

// IsAdmin validates the bearer access token and ensures the user has the admin role.
// Kept for backward compatibility or direct single-middleware usage.
func (m *AuthMiddleware) IsAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			response.Error(c, "missing or invalid authorization header", http.StatusUnauthorized, nil)
			c.Abort()
			return
		}

		model, err := m.accessTokens.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			response.ErrorFrom(c, err)
			c.Abort()
			return
		}

		if model.Role != "admin" {
			response.Error(c, "forbidden: admin access required", http.StatusForbidden, nil)
			c.Abort()
			return
		}

		authcontext.SetUser(c, model.ID, model.Role)
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}

	return token, true
}
