package item

import (
	"net/http"
	"strings"

	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

func (h *ItemHandler) promoteRequested(c *gin.Context) {
	var request PromoteRequestedItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}
	requestedItemID := c.Param("id")
	if !strings.HasPrefix(requestedItemID, domain.RequestedItemPrefixID) {
		response.Error(c, "invalid requested item ID", http.StatusBadRequest, nil)
		c.Abort()
		return
	}
	reviewerID, ok := identityhttp.GetUserID(c)
	if !ok || reviewerID == "" {
		response.Error(c, "authenticated user not found", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}
	if err := h.itemPortIn.PromoteRequestedItemUsecase.Execute(c.Request.Context(), itemin.PromoteRequestedItemDTO{
		RequestedItemID: requestedItemID, UnitCost: request.UnitCost, ReviewerID: reviewerID,
	}); err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	response.Success(c, "requested item promoted", http.StatusOK, nil)
}
