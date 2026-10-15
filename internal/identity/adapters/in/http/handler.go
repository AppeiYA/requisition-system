package identityhttp

import (
	"github.com/gin-gonic/gin"

	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
)

type Handler struct {
	loginUsecase   identityin.LoginUserUsecaseInt
	refreshUsecase identityin.RefreshTokenUsecaseInt
	logoutUsecase  identityin.LogoutUsecaseInt
	getMeUsecase   identityin.GetMeUsecaseInt
	accessTokens   identityout.AccessTokenService
}

func NewHandler(
	loginUsecase identityin.LoginUserUsecaseInt,
	refreshUsecase identityin.RefreshTokenUsecaseInt,
	logoutUsecase identityin.LogoutUsecaseInt,
	getMeUsecase identityin.GetMeUsecaseInt,
	accessTokens identityout.AccessTokenService,
) *Handler {
	return &Handler{
		loginUsecase:   loginUsecase,
		refreshUsecase: refreshUsecase,
		logoutUsecase:  logoutUsecase,
		getMeUsecase:   getMeUsecase,
		accessTokens:   accessTokens,
	}
}

func (h *Handler) Register(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	auth.POST("/login", h.login)
	auth.POST("/refresh", h.refresh)
	auth.POST("/logout", h.logout)

	auth.GET("/me", AuthMiddleware(h.accessTokens), h.getMe)
}
