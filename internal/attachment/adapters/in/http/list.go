package attachmenthttp

import (
	"net/http"

	attachmentin "github.com/AppeiYA/requisition-system/internal/attachment/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *AttachmentHandler) list(c *gin.Context) {
	actorID, actorRole, ok := requireAttachmentActor(c)
	if !ok {
		return
	}

	refName := c.Query("ref_name")
	refID := c.Query("ref_id")
	if refName == "" || refID == "" {
		response.Error(c, "ref_name and ref_id are required", http.StatusBadRequest, nil)
		c.Abort()
		return
	}

	q, ok := pagination.Parse(c)
	if !ok {
		c.Abort()
		return
	}

	attachments, err := h.attachmentPortIn.ListAttachmentsUsecase.Execute(c.Request.Context(), attachmentin.ListAttachmentsQuery{
		ActorID:   actorID,
		ActorRole: actorRole,
		RefName:   refName,
		RefID:     refID,
		Limit:     q.Limit,
		Offset:    q.Offset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*AttachmentResponse, 0, len(attachments))
	for _, attachment := range attachments {
		result = append(result, attachmentResponse(attachment, h.baseURL))
	}

	response.SuccessWithPagination(
		c,
		"attachments retrieved",
		http.StatusOK,
		result,
		response.NewPaginationMeta(q.Limit, q.Offset),
	)
}
