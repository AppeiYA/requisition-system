package item

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type CreateItemRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	CategoryID  string  `json:"category_id" binding:"required"`
	UnitCost    float64 `json:"unit_cost" binding:"gte=0"`
}

type UpdateItemRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	CategoryID  *string  `json:"category_id"`
	UnitCost    *float64 `json:"unit_cost" binding:"omitempty,gte=0"`
}

type ItemResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CategoryID  string    `json:"category_id"`
	UnitCost    float64   `json:"unit_cost"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type RequestedItemResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	CategoryID   string     `json:"category_id"`
	Description  string     `json:"description"`
	CreatedBy    string     `json:"created_by"`
	ReviewStatus string     `json:"review_status"`
	ReviewedBy   *string    `json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type PromoteRequestedItemRequest struct {
	UnitCost float64 `json:"unit_cost" binding:"gte=0"`
}

func RequestedItemResponseFromDomain(item *domain.RequestedItem) *RequestedItemResponse {
	if item == nil {
		return nil
	}
	return &RequestedItemResponse{
		ID: item.ID(), Name: item.Name().Value(), CategoryID: item.CategoryID(),
		Description: item.Description().Value(), CreatedBy: item.CreatedBy(),
		ReviewStatus: item.ReviewStatus().String(), ReviewedBy: item.ReviewedBy(),
		ReviewedAt: item.ReviewedAt(), CreatedAt: item.CreatedAt(),
	}
}

func ItemResponseFromDomain(item *domain.Item) *ItemResponse {
	if item == nil {
		return nil
	}

	return &ItemResponse{
		ID:          item.ID(),
		Name:        item.Name().Value(),
		Description: item.Description().Value(),
		CategoryID:  item.CategoryID(),
		UnitCost:    item.UnitCost().ToFloat64(),
		Currency:    item.UnitCost().Currency().String(),
		Status:      item.Status().String(),
		CreatedAt:   item.CreatedAt(),
	}
}
