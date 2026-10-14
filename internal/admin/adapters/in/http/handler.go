package adminhttp

import (
	"net/http"

	"github.com/gin-gonic/gin"

	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

type Handler struct {
	inviteUsecase adminin.InviteMemberUseCase
}

func NewHandler(
	inviteUsecase adminin.InviteMemberUseCase,
) *Handler {
	return &Handler{
		inviteUsecase: inviteUsecase,
	}
}

func (h *Handler) invite(c *gin.Context) {
	userID, ok := c.Get("identity.user_id")
	if !ok {
		response.Error(c, "unauthorized", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	res, err := h.inviteUsecase.Execute(c.Request.Context(), adminin.InviteMemberCommand{
		InviterID: userID.(string),
		Email:     req.Email,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "invitation sent", http.StatusCreated, res)
}

func (h *Handler) Register(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	admin := router.Group("/admin")
	admin.POST("/invitations", authMiddleware, h.invite)
}
