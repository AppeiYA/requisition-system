package invitationhttp

import (
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	acceptInvitationUC   invitationin.AcceptInvitationUseCaseInt
	validateInvitationUC invitationin.ValidateInvitationUseCaseInt
}

func NewInvitationHandler(
	acceptInvitationUC invitationin.AcceptInvitationUseCaseInt,
	validateInvitationUC invitationin.ValidateInvitationUseCaseInt,
) *InvitationHandler {
	return &InvitationHandler{
		acceptInvitationUC:   acceptInvitationUC,
		validateInvitationUC: validateInvitationUC,
	}
}

func (h *InvitationHandler) Register(router *gin.RouterGroup) {
	inv := router.Group("/invitations")
	inv.POST("/:token/accept", h.AcceptInvitation)
	inv.GET("/:token/validate", h.ValidateInvitation)
}
