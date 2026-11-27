package attachmenthttp

import (
	"errors"
	"mime"
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *AttachmentHandler) download(c *gin.Context) {
	actorID, actorRole, ok := requireAttachmentActor(c)
	if !ok {
		return
	}

	download, err := h.attachmentPortIn.DownloadAttachmentUsecase.Execute(c.Request.Context(), attachmentin.DownloadAttachmentQuery{
		AttachmentID: c.Param("id"),
		ActorID:      actorID,
		ActorRole:    actorRole,
	})
	if err != nil {
		if errors.Is(err, domain.ErrAttachmentNotReady) {
			response.ErrorWithDetails(c, "attachment_not_ready", err.Error(), http.StatusConflict, nil)
			c.Abort()
			return
		}
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	if download == nil {
		response.Error(c, "attachment not found", http.StatusNotFound, nil)
		c.Abort()
		return
	}

	if download.Location != "" {
		c.Header("Cache-Control", "no-store")
		c.Redirect(http.StatusFound, download.Location)
		return
	}
	if download.Stream == nil || download.Attachment == nil {
		response.Error(c, "attachment content unavailable", http.StatusInternalServerError, nil)
		c.Abort()
		return
	}
	defer download.Stream.Close()

	attachment := download.Attachment
	fileName := attachment.FileName().String()
	c.Header("Content-Type", attachment.ContentType().String())
	c.Header("X-Content-Type-Options", "nosniff")
	if attachment.ContentType().IsInlineSafe() {
		c.Header("Content-Disposition", "inline")
	} else {
		c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
			"filename": fileName,
		}))
	}

	http.ServeContent(c.Writer, c.Request, fileName, attachment.CreatedAt(), download.Stream)
}
