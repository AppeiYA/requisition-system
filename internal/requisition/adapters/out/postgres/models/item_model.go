package models

import (
	"time"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

type ItemModel struct {
	ID          string     `db:"id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CategoryID  string     `db:"category_id"`
	UnitCost    int64      `db:"unit_cost"`
	Status      string     `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

func ItemFromDomain(item *requisitiondomain.Item) *ItemModel {
	return &ItemModel{
		ID:          item.ID(),
		Name:        item.Name().Value(),
		Description: item.Description().Value(),
		CategoryID:  item.CategoryID(),
		UnitCost:    item.UnitCost().Amount(),
		Status:      item.Status().String(),
		CreatedAt:   item.CreatedAt(),
	}
}

func (m ItemModel) ToDomain() (*requisitiondomain.Item, error) {
	return requisitiondomain.ReconstituteItem(
		m.ID,
		m.Name,
		m.Description,
		m.CategoryID,
		m.UnitCost,
		m.Status,
		m.CreatedAt,
	)
}
