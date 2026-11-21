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

func (h *ItemHandler) dismissRequested(c *gin.Context) {
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
	if err := h.itemPortIn.DismissRequestedItemUsecase.Execute(c.Request.Context(), itemin.DismissRequestedItemDTO{
		RequestedItemID: requestedItemID,
		ReviewerID:      reviewerID,
	}); err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	response.Success(c, "requested item dismissed", http.StatusOK, nil)
}
