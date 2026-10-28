package models

import (
	"time"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type RequisitionModel struct {
	ID                       string    `db:"id"`
	RequestedBy              string    `db:"requested_by"`
	DepartmentID             string    `db:"department_id"`
	ItemID                   *string   `db:"item_id"`
	CategoryID               string    `db:"category_id"`
	RequestedItemName        string    `db:"requested_item_name"`
	RequestedItemDescription string    `db:"requested_item_description"`
	Quantity                 int       `db:"quantity"`
	EstimatedCost            int64     `db:"estimated_cost"`
	Justification            string    `db:"justification"`
	Priority                 string    `db:"priority"`
	NeededBy                 time.Time `db:"needed_by"`
	Status                   string    `db:"status"`
	CreatedAt                time.Time `db:"created_at"`
	UpdatedAt                time.Time `db:"updated_at"`
}

func RequisitionFromDomain(requisition *requisitiondomain.Requisition) *RequisitionModel {
	return &RequisitionModel{
		ID: requisition.ID(), RequestedBy: requisition.RequestedBy(),
		DepartmentID: requisition.DepartmentID(), ItemID: nullableString(requisition.ItemID()),
		CategoryID: requisition.CategoryID(), RequestedItemName: requisition.RequestedItemName().Value(),
		RequestedItemDescription: requisition.RequestedItemDescription().Value(),
		Quantity:                 int(requisition.Quantity().Value()), EstimatedCost: requisition.EstimatedCost().Amount(),
		Justification: requisition.Justification().Value(), Priority: requisition.Priority().String(),
		NeededBy: requisition.NeededBy(), Status: requisition.Status().String(),
		CreatedAt: requisition.CreatedAt(), UpdatedAt: requisition.UpdatedAt(),
	}
}

func (m RequisitionModel) ToDomain() (*requisitiondomain.Requisition, error) {
	cost, err := shareddomain.NewNairaMoneyFromMinorUnits(m.EstimatedCost)
	if err != nil {
		return nil, err
	}
	return requisitiondomain.ReconstituteRequisition(
		m.ID, m.RequestedBy, m.DepartmentID, stringValue(m.ItemID), m.CategoryID,
		m.RequestedItemName, m.RequestedItemDescription, m.Quantity, cost,
		m.Justification, m.Priority, m.NeededBy, m.Status, m.CreatedAt, m.UpdatedAt,
	)
}

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
