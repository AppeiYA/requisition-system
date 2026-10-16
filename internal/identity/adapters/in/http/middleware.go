package identityhttp

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

const userIDContextKey = "identity.user_id"

type AuthMiddleware struct {
	accessTokens identityout.AccessTokenService
}

func NewAuthMiddleware(accessTokens identityout.AccessTokenService) *AuthMiddleware {
	return &AuthMiddleware{
		accessTokens: accessTokens,
	}
}

// AuthMiddleware validates the bearer access token and stores the authenticated
// user id in the request context.
func(m *AuthMiddleware) Handle() gin.HandlerFunc {
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

		c.Set(userIDContextKey, model.ID)
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
