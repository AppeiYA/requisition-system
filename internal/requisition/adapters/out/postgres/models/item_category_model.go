package models

import (
	"time"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ItemCategoryModel struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

func ItemCategoryFromDomain(category *requisitiondomain.ItemCategory) *ItemCategoryModel {
	return &ItemCategoryModel{
		ID:          category.ID(),
		Name:        category.Name().Value(),
		Description: category.Description().Value(),
		Status:      category.Status().String(),
		CreatedAt:   category.CreatedAt(),
	}
}

func (m ItemCategoryModel) ToDomain() (*requisitiondomain.ItemCategory, error) {
	return requisitiondomain.ReconstituteItemCategory(
		m.ID,
		m.Name,
		m.Description,
		m.Status,
		m.CreatedAt,
	)
}
