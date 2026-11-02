package requisitionhttp

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
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
	router.PATCH("/requisitions/:id", authMiddleware, h.update)
	router.GET("/requisitions", authMiddleware, h.listMy)
	router.GET("/requisitions/all", authMiddleware, admin, h.listAll)
	router.GET("/requisitions/:id", authMiddleware, h.getByID)
	router.POST("/requisitions/:id/status", authMiddleware, h.transitionStatus)
}

func (h *RequisitionHandler) update(c *gin.Context) {
	var request CreateRequisitionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}
	if details := validateCreateRequisitionRequest(request); len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, details)
		c.Abort()
		return
	}
	requisitionID := c.Param("id")
	if details := validateRequisitionID(requisitionID); len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid request", http.StatusBadRequest, details)
		c.Abort()
		return
	}
	userID, ok := identityhttp.GetUserID(c)
	if !ok || userID == "" {
		response.Error(c, "authenticated user not found", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	lines := make([]requisitionin.CreateRequisitionLine, 0, len(request.Lines))
	for _, line := range request.Lines {
		commandLine := requisitionin.CreateRequisitionLine{
			ItemID: line.ItemID, Quantity: line.Quantity,
			EstimatedUnitAmount: line.EstimatedUnitAmount, Notes: line.Notes,
		}
		if line.RequestedItem != nil {
			commandLine.RequestedItem = &requisitionin.CreateRequestedItem{
				Name: line.RequestedItem.Name, CategoryID: line.RequestedItem.CategoryID,
				Description: line.RequestedItem.Description,
			}
		}
		lines = append(lines, commandLine)
	}
	updated, err := h.portIn.UpdateRequisitionUsecase.Execute(c.Request.Context(), requisitionin.UpdateRequisitionCommand{
		RequisitionID: requisitionID, UserID: userID, DepartmentID: request.DepartmentID,
		Lines: lines, Justification: request.Justification, Priority: request.Priority,
		NeededBy: request.NeededBy.Time(),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	mapped, err := RequisitionResponseFromDomainWithError(updated)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	response.Success(c, "requisition updated", http.StatusOK, mapped)
}

func (h *RequisitionHandler) create(c *gin.Context) {
	var request CreateRequisitionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}
	if details := validateCreateRequisitionRequest(request); len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, details)
		c.Abort()
		return
	}

	userID, ok := identityhttp.GetUserID(c)
	if !ok || userID == "" {
		response.Error(c, "authenticated user not found", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}

	lines := make([]requisitionin.CreateRequisitionLine, 0, len(request.Lines))
	for _, line := range request.Lines {
		commandLine := requisitionin.CreateRequisitionLine{
			ItemID:              line.ItemID,
			Quantity:            line.Quantity,
			EstimatedUnitAmount: line.EstimatedUnitAmount,
			Notes:               line.Notes,
		}
		if line.RequestedItem != nil {
			commandLine.RequestedItem = &requisitionin.CreateRequestedItem{
				Name:        line.RequestedItem.Name,
				CategoryID:  line.RequestedItem.CategoryID,
				Description: line.RequestedItem.Description,
			}
		}
		lines = append(lines, commandLine)
	}

	requisition, err := h.portIn.CreateRequisitionUsecase.Execute(c.Request.Context(), requisitionin.CreateRequisitionCommand{
		UserID:        userID,
		DepartmentID:  request.DepartmentID,
		Lines:         lines,
		Justification: request.Justification,
		Priority:      request.Priority,
		NeededBy:      request.NeededBy.Time(),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	created, err := RequisitionResponseFromDomainWithError(requisition)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	response.Success(c, "requisition created", http.StatusCreated, created)
}

func (h *RequisitionHandler) getByID(c *gin.Context) {
	userID, ok := identityhttp.GetUserID(c)
	if !ok || userID == "" {
		response.Error(c, "authenticated user not found", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}
	role, _ := identityhttp.GetUserRole(c)

	requisitionID := c.Param("id")
	if details := validateRequisitionID(requisitionID); len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid request", http.StatusBadRequest, details)
		c.Abort()
		return
	}

	requisition, err := h.portIn.GetRequisitionUsecase.Execute(c.Request.Context(), requisitionin.GetRequisitionQuery{
		RequisitionID: requisitionID,
		ActorID:       userID,
		ActorRole:     role,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	mapped, err := RequisitionResponseFromDomainWithError(requisition)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	response.Success(c, "requisition fetched successfully", http.StatusOK, mapped)
}

func (h *RequisitionHandler) transitionStatus(c *gin.Context) {
	var request TransitionRequisitionStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, response.ValidationErrors(err))
		c.Abort()
		return
	}

	requisitionID := c.Param("id")
	if details := validateTransitionRequisitionRequest(requisitionID, request); len(details) > 0 {
		response.ErrorWithDetails(c, "validation_error", "invalid request body", http.StatusBadRequest, details)
		c.Abort()
		return
	}

	userID, ok := identityhttp.GetUserID(c)
	if !ok || userID == "" {
		response.Error(c, "authenticated user not found", http.StatusUnauthorized, nil)
		c.Abort()
		return
	}
	role, _ := identityhttp.GetUserRole(c)

	requisition, err := h.portIn.TransitionRequisitionUsecase.Execute(c.Request.Context(), requisitionin.TransitionRequisitionCommand{
		RequisitionID: requisitionID,
		ActorID:       userID,
		ActorRole:     role,
		ToStatus:      request.ToStatus,
		Reason:        request.Reason,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	mapped, err := RequisitionResponseFromDomainWithError(requisition)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}
	response.Success(c, "requisition status updated", http.StatusOK, mapped)
}

func validateRequisitionID(requisitionID string) []response.FieldError {
	if strings.HasPrefix(requisitionID, domain.RequisitionPrefixID) {
		return nil
	}
	return []response.FieldError{{
		Field: "id", Message: "must be a valid requisition ID",
	}}
}

func validateTransitionRequisitionRequest(requisitionID string, request TransitionRequisitionStatusRequest) []response.FieldError {
	details := validateRequisitionID(requisitionID)

	status, err := domain.NewRequisitionStatus(request.ToStatus)
	if err != nil {
		details = append(details, response.FieldError{
			Field:   "to_status",
			Message: "must be one of pending, approved, rejected, in_fulfillment, completed, cancelled",
		})
		return details
	}
	if (status == domain.RequisitionStatusRejected || status == domain.RequisitionStatusCancelled) &&
		strings.TrimSpace(request.Reason) == "" {
		details = append(details, response.FieldError{
			Field:   "reason",
			Message: "reason is required when transitioning to rejected or cancelled",
		})
	}
	return details
}

func validateCreateRequisitionRequest(request CreateRequisitionRequest) []response.FieldError {
	details := make([]response.FieldError, 0)

	if len(request.Lines) == 0 {
		details = append(details, response.FieldError{
			Field: "lines", Message: "at least one requisition line is required",
		})
	}
	if !strings.HasPrefix(request.DepartmentID, domain.DepartmentPrefixID) {
		details = append(details, response.FieldError{
			Field: "department_id", Message: "must be a valid department ID",
		})
	}
	if request.NeededBy != "" && isNotFutureDate(request.NeededBy.Time()) {
		details = append(details, response.FieldError{
			Field: "needed_by", Message: "must be a date in the future",
		})
	}
	if len(request.Justification) > 500 {
		details = append(details, response.FieldError{
			Field: "justification", Message: "must be at most 500 characters long",
		})
	}

	seenItems := make(map[string]int, len(request.Lines))
	for index, line := range request.Lines {
		field := fmt.Sprintf("lines[%d]", index)
		hasItem := strings.TrimSpace(line.ItemID) != ""
		hasRequestedItem := line.RequestedItem != nil

		switch {
		case hasItem && hasRequestedItem:
			details = append(details, response.FieldError{
				Field: field, Message: "exactly one of item_id or requested_item must be provided",
			})
		case !hasItem && !hasRequestedItem:
			details = append(details, response.FieldError{
				Field: field, Message: "exactly one of item_id or requested_item must be provided",
			})
		}

		if hasItem {
			if !strings.HasPrefix(line.ItemID, domain.ItemPrefixID) {
				details = append(details, response.FieldError{
					Field: field + ".item_id", Message: "must be a valid item ID",
				})
			}
			if previousLine, exists := seenItems[line.ItemID]; exists {
				details = append(details, response.FieldError{
					Field:   field + ".item_id",
					Message: fmt.Sprintf("catalogue item already appears on line %d", previousLine),
				})
			} else {
				seenItems[line.ItemID] = index + 1
			}
		}

		if hasRequestedItem {
			requestedItem := line.RequestedItem
			if strings.TrimSpace(requestedItem.Name) == "" {
				details = append(details, response.FieldError{
					Field: field + ".requested_item.name", Message: "this field is required",
				})
			} else if len(requestedItem.Name) > 100 {
				details = append(details, response.FieldError{
					Field: field + ".requested_item.name", Message: "must be at most 100 characters long",
				})
			}
			if strings.TrimSpace(requestedItem.CategoryID) == "" {
				details = append(details, response.FieldError{
					Field: field + ".requested_item.category_id", Message: "this field is required",
				})
			} else if !strings.HasPrefix(requestedItem.CategoryID, domain.CategoryPrefixID) {
				details = append(details, response.FieldError{
					Field: field + ".requested_item.category_id", Message: "must be a valid category ID",
				})
			}
			if strings.TrimSpace(requestedItem.Description) == "" {
				details = append(details, response.FieldError{
					Field: field + ".requested_item.description", Message: "this field is required",
				})
			} else if len(requestedItem.Description) > 200 {
				details = append(details, response.FieldError{
					Field: field + ".requested_item.description", Message: "must be at most 200 characters long",
				})
			}
		}

		if line.Quantity <= 0 {
			details = append(details, response.FieldError{
				Field: field + ".quantity", Message: "must be greater than or equal to 1",
			})
		} else if line.Quantity > 1000 {
			details = append(details, response.FieldError{
				Field: field + ".quantity", Message: "must be less than or equal to 1000",
			})
		}
		if line.EstimatedUnitAmount < 0 {
			details = append(details, response.FieldError{
				Field: field + ".estimated_unit_amount", Message: "must be greater than or equal to 0",
			})
		}
	}

	return details
}

func isNotFutureDate(value time.Time) bool {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return !value.UTC().After(today)
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

	result, err := requisitionResponsesFromDomain(requisitions)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
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

	result, err := requisitionResponsesFromDomain(requisitions)
	if err != nil {
		response.ErrorFrom(c, err)
		c.Abort()
		return
	}

	response.SuccessWithPagination(c, "requisitions fetched successfully", http.StatusOK, result, response.NewPaginationMeta(q.Limit, q.Offset))
}

func requisitionResponsesFromDomain(requisitions []*domain.Requisition) ([]*RequisitionResponse, error) {
	result := make([]*RequisitionResponse, 0, len(requisitions))
	for _, requisition := range requisitions {
		mapped, err := RequisitionResponseFromDomainWithError(requisition)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}
