package identityhttp

import (
	"github.com/gin-gonic/gin"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/ratelimit"
)

type Handler struct {
	identityPortIn       *identityin.IdentityPortIn
	authMiddleware       *AuthMiddleware
	rateLimiter          *ratelimit.Limiter
	rateLimit            *config.RateLimitConfig
	passwordResetEnabled bool
}

func NewHandler(identityPortIn *identityin.IdentityPortIn, authMiddleware *AuthMiddleware) *Handler {
	return &Handler{
		identityPortIn:       identityPortIn,
		authMiddleware:       authMiddleware,
		passwordResetEnabled: true,
	}
}

// SetPasswordResetEnabled gates the self-service routes so the backend can be
// deployed before the frontend reset page exists. It is called before Register.
func (h *Handler) SetPasswordResetEnabled(enabled bool) {
	h.passwordResetEnabled = enabled
}

func (h *Handler) SetRateLimiter(limiter *ratelimit.Limiter, rateLimit config.RateLimitConfig) {
	h.rateLimiter = limiter
	h.rateLimit = &rateLimit
}

func (h *Handler) Register(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	auth.POST("/login", h.loginMiddleware(), h.login)
	auth.POST("/refresh", h.refreshMiddleware(), h.refresh)
	auth.POST("/logout", h.logoutMiddleware(), h.logout)
	if h.passwordResetEnabled {
		auth.POST("/forgot-password", h.forgotPasswordMiddleware(), h.forgotPassword)
		auth.POST("/reset-password", h.resetPasswordMiddleware(), h.resetPassword)
	}

	auth.GET("/me", h.authMiddleware.Handle(), h.getMe)
}

func (h *Handler) loginMiddleware() gin.HandlerFunc {
	if h.rateLimiter == nil || h.rateLimit == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return h.rateLimiter.Middleware("login", h.rateLimit.LoginLimit, h.rateLimit.LoginWindow)
}

func (h *Handler) refreshMiddleware() gin.HandlerFunc {
	if h.rateLimiter == nil || h.rateLimit == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return h.rateLimiter.Middleware("refresh", h.rateLimit.RefreshLimit, h.rateLimit.RefreshWindow)
}

func (h *Handler) logoutMiddleware() gin.HandlerFunc {
	if h.rateLimiter == nil || h.rateLimit == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return h.rateLimiter.Middleware("logout", h.rateLimit.LogoutLimit, h.rateLimit.LogoutWindow)
}

func (h *Handler) forgotPasswordMiddleware() gin.HandlerFunc {
	if h.rateLimiter == nil || h.rateLimit == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return h.rateLimiter.Middleware("forgot-password", h.rateLimit.ForgotPasswordLimit, h.rateLimit.ForgotPasswordWindow)
}

func (h *Handler) resetPasswordMiddleware() gin.HandlerFunc {
	if h.rateLimiter == nil || h.rateLimit == nil {
		return func(c *gin.Context) { c.Next() }
	}
	return h.rateLimiter.Middleware("reset-password", h.rateLimit.ResetPasswordLimit, h.rateLimit.ResetPasswordWindow)
}
