package attachmenthttp

import (
	"net/http"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *AttachmentHandler) delete(c *gin.Context) {
	actorID, actorRole, ok := requireAttachmentActor(c)
	if !ok {
		return
	}

	err := h.attachmentPortIn.DeleteAttachmentUsecase.Execute(c.Request.Context(), attachmentin.DeleteAttachmentCommand{
		AttachmentID: c.Param("id"),
		ActorID:      actorID,
		ActorRole:    actorRole,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "attachment deleted", http.StatusOK, nil)
}
