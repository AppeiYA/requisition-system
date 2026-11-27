package attachmenthttp

import (
	"net/http"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *AttachmentHandler) stageUpload(c *gin.Context) {
	actorID, _, ok := requireAttachmentActor(c)
	if !ok {
		return
	}

	parsed, ok := parseAttachmentMultipart(c, h.maxSizeBytes, false)
	if !ok {
		return
	}
	defer parsed.Body.Close()

	upload, err := h.attachmentPortIn.StageUploadUsecase.Execute(c.Request.Context(), attachmentin.StageUploadCommand{
		ActorID:  actorID,
		FileName: parsed.FileName,
		Body:     parsed.Body,
	})
	if err != nil {
		writeUploadError(c, err)
		return
	}

	response.Success(c, "file staged", http.StatusCreated, stagedUploadResponse(upload))
}
