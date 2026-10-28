package category

import (
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type CreateItemCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

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
