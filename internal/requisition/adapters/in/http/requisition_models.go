package requisitionhttp

import (
	"encoding/json"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
)

const dateOnlyLayout = "2006-01-02"

type Date string

type dateValidationError struct {
	message string
}

func (e dateValidationError) Error() string {
	return e.message
}

func (e dateValidationError) ValidationField() string {
	return "needed_by"
}

func (e dateValidationError) ValidationMessage() string {
	return e.message
}

func (d *Date) UnmarshalJSON(value []byte) error {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return dateValidationError{message: "must be a date in YYYY-MM-DD format"}
	}
	if _, err := time.Parse(dateOnlyLayout, raw); err != nil {
		return dateValidationError{message: "must be a date in YYYY-MM-DD format"}
	}
	*d = Date(raw)
	return nil
}

func (d Date) Time() time.Time {
	value, _ := time.ParseInLocation(dateOnlyLayout, string(d), time.UTC)
	return value
}

type CreateRequisitionRequest struct {
	CategoryID               string  `json:"category_id" binding:"required"`
	ItemID                   string  `json:"item_id"`
	RequestedItemName        string  `json:"requested_item_name" binding:"required"`
	RequestedItemDescription string  `json:"requested_item_description" binding:"required"`
	Quantity                 int     `json:"quantity" binding:"required,gte=1"`
	EstimatedCost            float64 `json:"estimated_cost" binding:"gte=0"`
	Justification            string  `json:"justification" binding:"required"`
	Priority                 string  `json:"priority" binding:"required"`
	NeededBy                 Date    `json:"needed_by" binding:"required"`
}

type RequisitionResponse struct {
	ID                       string    `json:"id"`
	RequestedBy              string    `json:"requested_by"`
	DepartmentID             string    `json:"department_id"`
	ItemID                   string    `json:"item_id,omitempty"`
	CategoryID               string    `json:"category_id"`
	RequestedItemName        string    `json:"requested_item_name"`
	RequestedItemDescription string    `json:"requested_item_description"`
	Quantity                 int64     `json:"quantity"`
	EstimatedCost            float64   `json:"estimated_cost"`
	Currency                 string    `json:"currency"`
	Justification            string    `json:"justification"`
	Priority                 string    `json:"priority"`
	NeededBy                 string    `json:"needed_by"`
	Status                   string    `json:"status"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

func RequisitionResponseFromDomain(requisition *domain.Requisition) *RequisitionResponse {
	if requisition == nil {
		return nil
	}
	return &RequisitionResponse{
		ID: requisition.ID(), RequestedBy: requisition.RequestedBy(), DepartmentID: requisition.DepartmentID(),
		ItemID: requisition.ItemID(), CategoryID: requisition.CategoryID(),
		RequestedItemName:        requisition.RequestedItemName().Value(),
		RequestedItemDescription: requisition.RequestedItemDescription().Value(),
		Quantity:                 requisition.Quantity().Value(), EstimatedCost: requisition.EstimatedCost().ToFloat64(),
		Currency: requisition.EstimatedCost().Currency().String(), Justification: requisition.Justification().Value(),
		Priority: requisition.Priority().String(), NeededBy: requisition.NeededBy().UTC().Format(dateOnlyLayout), Status: requisition.Status().String(),
		CreatedAt: requisition.CreatedAt(), UpdatedAt: requisition.UpdatedAt(),
	}
}
