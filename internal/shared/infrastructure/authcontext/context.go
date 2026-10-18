package authcontext

import (
	"context"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey   = "identity.user_id"
	UserRoleKey = "identity.user_role"
)

type userContextKey struct{}

// UserInfo holds the authenticated identity recovered from a token.
type UserInfo struct {
	ID   string
	Role string
}

// SetUser stores the user information in both the gin.Context and the underlying http.Request context.
func SetUser(c *gin.Context, id, role string) {
	c.Set(UserIDKey, id)
	c.Set(UserRoleKey, role)
	if c.Request != nil {
		ctx := context.WithValue(c.Request.Context(), userContextKey{}, UserInfo{ID: id, Role: role})
		c.Request = c.Request.WithContext(ctx)
	}
}

// GetUserID retrieves the user ID from gin.Context or its underlying Request context.
func GetUserID(c *gin.Context) (string, bool) {
	if val, ok := c.Get(UserIDKey); ok {
		if id, ok := val.(string); ok && id != "" {
			return id, true
		}
	}
	if c.Request != nil {
		if u, ok := FromContext(c.Request.Context()); ok && u.ID != "" {
			return u.ID, true
		}
	}
	return "", false
}

// GetUserRole retrieves the user role from gin.Context or its underlying Request context.
func GetUserRole(c *gin.Context) (string, bool) {
	if val, ok := c.Get(UserRoleKey); ok {
		if role, ok := val.(string); ok && role != "" {
			return role, true
		}
	}
	if c.Request != nil {
		if u, ok := FromContext(c.Request.Context()); ok && u.Role != "" {
			return u.Role, true
		}
	}
	return "", false
}

// FromContext extracts UserInfo from a standard Go context.Context.
func FromContext(ctx context.Context) (UserInfo, bool) {
	u, ok := ctx.Value(userContextKey{}).(UserInfo)
	return u, ok
}

// WithUser creates a new context.Context carrying UserInfo.
func WithUser(ctx context.Context, id, role string) context.Context {
	return context.WithValue(ctx, userContextKey{}, UserInfo{ID: id, Role: role})
}

