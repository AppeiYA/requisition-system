package attachmenthttp

import (
	"net/http"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *AttachmentHandler) get(c *gin.Context) {
	actorID, actorRole, ok := requireAttachmentActor(c)
	if !ok {
		return
	}

	attachment, err := h.attachmentPortIn.GetAttachmentUsecase.Execute(c.Request.Context(), attachmentin.GetAttachmentQuery{
		AttachmentID: c.Param("id"),
		ActorID:      actorID,
		ActorRole:    actorRole,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "attachment retrieved", http.StatusOK, attachmentResponse(attachment, h.baseURL))
}
