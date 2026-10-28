package requisitionhttp

import (
	"net/http"

	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/pagination"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
	"github.com/gin-gonic/gin"
)

type RequisitionHandler struct {
	portIn *requisitionin.RequisitionPortIn
}

func NewRequisitionHandler(portIn *requisitionin.RequisitionPortIn) *RequisitionHandler {
	return &RequisitionHandler{portIn: portIn}
}

func (h *RequisitionHandler) Register(router *gin.RouterGroup, authMiddleware gin.HandlerFunc, adminOnlyMiddleware ...gin.HandlerFunc) {
	var admin gin.HandlerFunc
	if len(adminOnlyMiddleware) > 0 && adminOnlyMiddleware[0] != nil {
		admin = adminOnlyMiddleware[0]
	} else {
		admin = authMiddleware
	}
	router.POST("/requisitions", authMiddleware, h.create)
	router.GET("/requisitions", authMiddleware, h.listMy)
	router.GET("/requisitions/all", authMiddleware, admin, h.listAll)
}

func (h *RequisitionHandler) create(c *gin.Context) {
	var request CreateRequisitionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	userID, ok := identityhttp.GetUserID(c)
	if !ok || userID == "" {
		response.Error(c, "authenticated user not found", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	requisition, err := h.portIn.CreateRequisitionUsecase.Execute(c.Request.Context(), requisitionin.CreateRequisitionCommand{
		UserID: userID, CategoryID: request.CategoryID, ItemID: request.ItemID,
		RequestedItemName: request.RequestedItemName, RequestedItemDescription: request.RequestedItemDescription,
		Quantity: request.Quantity, EstimatedCost: request.EstimatedCost, Justification: request.Justification,
		Priority: request.Priority, NeededBy: request.NeededBy.Time(),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.Success(c, "requisition created", http.StatusCreated, RequisitionResponseFromDomain(requisition))
}

func (h *RequisitionHandler) listMy(c *gin.Context) {
	userID, ok := identityhttp.GetUserID(c)
	if !ok || userID == "" {
		response.Error(c, "authenticated user not found", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	q, ok := pagination.Parse(c)
	if !ok {
		return
	}

	requisitions, err := h.portIn.ListRequisitionsByUserUsecase.Execute(c.Request.Context(), requisitionin.ListRequisitionsByUserQuery{
		UserID: userID,
		Limit:  q.Limit,
		Offset: q.Offset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*RequisitionResponse, 0, len(requisitions))
	for _, req := range requisitions {
		result = append(result, RequisitionResponseFromDomain(req))
	}

	response.SuccessWithPagination(c, "requisitions fetched successfully", http.StatusOK, result, response.NewPaginationMeta(q.Limit, q.Offset))
}

func (h *RequisitionHandler) listAll(c *gin.Context) {
	q, ok := pagination.Parse(c)
	if !ok {
		return
	}

	requisitions, err := h.portIn.ListRequisitionsUsecase.Execute(c.Request.Context(), requisitionin.ListRequisitionsQuery{
		Limit:  q.Limit,
		Offset: q.Offset,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	result := make([]*RequisitionResponse, 0, len(requisitions))
	for _, req := range requisitions {
		result = append(result, RequisitionResponseFromDomain(req))
	}

	response.SuccessWithPagination(c, "requisitions fetched successfully", http.StatusOK, result, response.NewPaginationMeta(q.Limit, q.Offset))
}
