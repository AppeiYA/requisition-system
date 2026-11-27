package attachmenthttp

import (
	"errors"
	"net/http"
	"strings"

	"github.com/AppeiYA/requisition-system/internal/attachment/domain"
	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *AttachmentHandler) upload(c *gin.Context) {
	actorID, actorRole, ok := requireAttachmentActor(c)
	if !ok {
		return
	}

	parsed, ok := parseAttachmentMultipart(c, h.maxSizeBytes, true)
	if !ok {
		return
	}
	defer parsed.Body.Close()

	attachment, err := h.attachmentPortIn.UploadAttachmentUsecase.Execute(c.Request.Context(), attachmentin.UploadAttachmentCommand{
		ActorID:   actorID,
		ActorRole: actorRole,
		RefName:   parsed.RefName,
		RefID:     parsed.RefID,
		FileName:  parsed.FileName,
		Body:      parsed.Body,
	})
	if err != nil {
		writeUploadError(c, err)
		return
	}

	response.Success(c, "attachment uploaded", http.StatusCreated, attachmentResponse(attachment, h.baseURL))
}

// writeUploadError maps the use case's user-input failures to the HTTP codes
// documented for uploads. Content type is detected server-side, so an
// unsupported type is a 415 even though the domain reports a bad request.
func writeUploadError(c *gin.Context, err error) {
	if isAttachmentTooLarge(err) {
		response.ErrorFrom(c, customerrors.NewCustomError(
			"attachment exceeds the maximum allowed size",
			http.StatusRequestEntityTooLarge,
		))
		c.Abort()
		return
	}

	if errors.Is(err, domain.ErrInvalidContentType) {
		response.ErrorFrom(c, customerrors.NewCustomError(err.Error(), http.StatusUnsupportedMediaType))
		c.Abort()
		return
	}

	var customErr *customerrors.CustomError
	if errors.As(err, &customErr) && customErr.Code == http.StatusBadRequest {
		message := strings.ToLower(customErr.Message)
		switch {
		case strings.Contains(message, "content type"), strings.Contains(message, "unsupported file"):
			response.ErrorFrom(c, customerrors.NewCustomError(customErr.Message, http.StatusUnsupportedMediaType))
			c.Abort()
			return
		case strings.Contains(message, "too large"), strings.Contains(message, "maximum size"):
			response.ErrorFrom(c, customerrors.NewCustomError(customErr.Message, http.StatusRequestEntityTooLarge))
			c.Abort()
			return
		}
	}

	response.ErrorFrom(c, err)
	c.Abort()
}
