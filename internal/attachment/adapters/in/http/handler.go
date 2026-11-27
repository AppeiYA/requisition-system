package attachmenthttp

import (
	"net/http"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

// AttachmentHandler exposes the attachment use cases over HTTP. Storage
// details never reach this layer: downloads are streamed or redirected based
// on what the download use case returns.
type AttachmentHandler struct {
	attachmentPortIn *attachmentin.AttachmentPortIn
	maxSizeBytes     int64
	baseURL          string
}

func NewAttachmentHandler(
	attachmentPortIn *attachmentin.AttachmentPortIn,
	maxSizeBytes int64,
	baseURL string,
) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentPortIn: attachmentPortIn,
		maxSizeBytes:     maxSizeBytes,
		baseURL:          baseURL,
	}
}

func (h *AttachmentHandler) Register(router *gin.RouterGroup, auth gin.HandlerFunc) {
	uploads := router.Group("/uploads")
	uploads.POST("", auth, h.stageUpload)

	attachments := router.Group("/attachments")
	attachments.POST("", auth, h.upload)
	attachments.GET("", auth, h.list)
	attachments.GET("/:id", auth, h.get)
	attachments.GET("/:id/content", auth, h.download)
	attachments.DELETE("/:id", auth, h.delete)
}

// attachmentActor reads the identity installed by the auth middleware. The
// actor is required on all routes: permanent attachments are authorised
// against their parent, while staged uploads are authorised by ownership.
func attachmentActor(c *gin.Context) (string, string, bool) {
	actorID, ok := identityhttp.GetUserID(c)
	if !ok || actorID == "" {
		return "", "", false
	}
	actorRole, ok := identityhttp.GetUserRole(c)
	if !ok || actorRole == "" {
		return "", "", false
	}
	return actorID, actorRole, true
}

func requireAttachmentActor(c *gin.Context) (string, string, bool) {
	actorID, actorRole, ok := attachmentActor(c)
	if !ok {
		response.Error(c, "missing or invalid authorization header", http.StatusUnauthorized, nil)
		c.Abort()
		return "", "", false
	}
	return actorID, actorRole, true
}
