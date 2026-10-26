package requisitionhttp

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

// CreateItemCategoryRequest is the admin's payload for creating an item
// category.
type CreateItemCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateItemCategoryRequest is the admin's payload for a partial category
// update. Every field is optional, but the domain requires at least one of
// them to be provided. Status changes use the activate and deactivate routes.
type UpdateItemCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type ItemCategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func ItemCategoryResponseFromDomain(category *domain.ItemCategory) *ItemCategoryResponse {
	if category == nil {
		return nil
	}

	return &ItemCategoryResponse{
		ID:          category.ID(),
		Name:        category.Name().Value(),
		Description: category.Description().Value(),
		Status:      category.Status().String(),
		CreatedAt:   category.CreatedAt(),
	}
}
