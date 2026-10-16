package invitationhttp

import (
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	invitationPortIn   *invitationin.InvitationPortIn
}

func NewInvitationHandler(
	invitationPortIn *invitationin.InvitationPortIn,
) *InvitationHandler {
	return &InvitationHandler{
		invitationPortIn:   invitationPortIn,
	}
}

func (h *InvitationHandler) Register(router *gin.RouterGroup) {
	inv := router.Group("/invitations")
	inv.POST("/:token/accept", h.AcceptInvitation)
	inv.GET("/:token/validate", h.ValidateInvitation)
}
