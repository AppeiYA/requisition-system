package requisitionhttp

import (
	"encoding/json"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
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

// CreateRequisitionRequest is the requisition create payload. Amounts are
// expressed in major currency units (naira for NGN).
type CreateRequisitionRequest struct {
	DepartmentID  string                         `json:"department_id" binding:"required" minLength:"1" example:"dpt_123"`
	Priority      string                         `json:"priority" binding:"omitempty,oneof=low normal medium high urgent" example:"medium"`
	NeededBy      Date                           `json:"needed_by" binding:"required" example:"2026-11-15"`
	Justification string                         `json:"justification" binding:"required" minLength:"1" maxLength:"500" example:"Replacement laptops for the new hires"`
	Lines         []CreateRequisitionLineRequest `json:"lines" binding:"required,min=1,dive"`
}

// CreateRequisitionLineRequest is one line in a requisition. Exactly one of
// ItemID and RequestedItem must be supplied.
type CreateRequisitionLineRequest struct {
	ItemID              string                      `json:"item_id" example:"itm_123"`
	RequestedItem       *CreateRequestedItemRequest `json:"requested_item"`
	Quantity            int                         `json:"quantity" minimum:"1" maximum:"1000" example:"2"`
	EstimatedUnitAmount float64                     `json:"estimated_unit_amount" example:"50000"`
	Notes               string                      `json:"notes" example:"Optional note"`
}

// CreateRequestedItemRequest is the free-filled item payload for a line that
// does not reference a catalogue item.
type CreateRequestedItemRequest struct {
	Name        string `json:"name" minLength:"1" maxLength:"100" example:"Laptop stand"`
	CategoryID  string `json:"category_id" minLength:"1" example:"cat_123"`
	Description string `json:"description" minLength:"1" maxLength:"200" example:"Adjustable laptop stand"`
}

// TransitionRequisitionStatusRequest is the status-transition payload. A
// reason is required when moving to rejected or cancelled.
type TransitionRequisitionStatusRequest struct {
	ToStatus string `json:"to_status" binding:"required" example:"approved"`
	Reason   string `json:"reason" example:"Approved by department administrator"`
}

type RequisitionResponse struct {
	ID                   string                             `json:"id"`
	ReferenceCode        string                             `json:"reference_code"`
	RequestedBy          string                             `json:"requested_by"`
	DepartmentID         string                             `json:"department_id"`
	Status               string                             `json:"status"`
	Priority             string                             `json:"priority"`
	NeededBy             string                             `json:"needed_by"`
	Justification        string                             `json:"justification"`
	Currency             string                             `json:"currency"`
	Lines                []RequisitionLineResponse          `json:"lines"`
	TotalEstimatedAmount float64                            `json:"total_estimated_amount"`
	StatusHistory        []RequisitionStatusHistoryResponse `json:"status_history"`
	CreatedAt            time.Time                          `json:"created_at"`
	UpdatedAt            time.Time                          `json:"updated_at"`
}

// RequisitionSummaryResponse is the list projection. It intentionally omits
// lines and status history so list responses stay header-oriented.
type RequisitionSummaryResponse struct {
	ID                   string    `json:"id"`
	ReferenceCode        string    `json:"reference_code"`
	RequestedBy          string    `json:"requested_by"`
	DepartmentID         string    `json:"department_id"`
	Status               string    `json:"status"`
	Priority             string    `json:"priority"`
	NeededBy             string    `json:"needed_by"`
	Justification        string    `json:"justification"`
	TotalEstimatedAmount float64   `json:"total_estimated_amount"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type RequisitionLineResponse struct {
	ID                  string  `json:"id"`
	LineNumber          int     `json:"line_number"`
	ItemID              string  `json:"item_id,omitempty"`
	RequestedItemID     string  `json:"requested_item_id,omitempty"`
	ItemName            string  `json:"item_name"`
	Quantity            int64   `json:"quantity"`
	EstimatedUnitAmount float64 `json:"estimated_unit_amount"`
	Notes               string  `json:"notes"`
	LineTotal           float64 `json:"line_total"`
}

type RequisitionStatusHistoryResponse struct {
	Sequence   int       `json:"sequence"`
	FromStatus *string   `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	ActedBy    string    `json:"acted_by"`
	ActorRole  string    `json:"actor_role"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

func RequisitionSummaryResponseFromDomain(summary *domain.RequisitionSummary) *RequisitionSummaryResponse {
	if summary == nil {
		return nil
	}
	return &RequisitionSummaryResponse{
		ID:                   summary.ID,
		ReferenceCode:        summary.ReferenceCode,
		RequestedBy:          summary.RequestedBy,
		DepartmentID:         summary.DepartmentID,
		Status:               summary.Status,
		Priority:             summary.Priority,
		NeededBy:             summary.NeededBy.UTC().Format(dateOnlyLayout),
		Justification:        summary.Justification,
		TotalEstimatedAmount: summary.TotalEstimatedAmount,
		CreatedAt:            summary.CreatedAt,
		UpdatedAt:            summary.UpdatedAt,
	}
}

func RequisitionResponseFromDomain(requisition *domain.Requisition) *RequisitionResponse {
	response, err := RequisitionResponseFromDomainWithError(requisition)
	if err != nil {
		return nil
	}
	return response
}

func RequisitionResponseFromDomainWithError(requisition *domain.Requisition) (*RequisitionResponse, error) {
	if requisition == nil {
		return nil, nil
	}

	lines := make([]RequisitionLineResponse, 0, len(requisition.Items()))
	for _, item := range requisition.Items() {
		lineTotal, err := item.LineTotal()
		if err != nil {
			return nil, err
		}
		lines = append(lines, RequisitionLineResponse{
			ID:                  item.ID(),
			LineNumber:          item.LineNumber(),
			ItemID:              item.ItemID(),
			RequestedItemID:     item.RequestedItemID(),
			ItemName:            item.ItemNameSnapshot().Value(),
			Quantity:            item.Quantity().Value(),
			EstimatedUnitAmount: item.EstimatedUnitAmount().ToFloat64(),
			Notes:               item.Notes(),
			LineTotal:           lineTotal.ToFloat64(),
		})
	}

	total, err := requisition.TotalEstimatedAmount()
	if err != nil {
		return nil, err
	}

	history := make([]RequisitionStatusHistoryResponse, 0, len(requisition.StatusHistory()))
	for _, entry := range requisition.StatusHistory() {
		var fromStatus *string
		if entry.FromStatus() != nil {
			value := entry.FromStatus().String()
			fromStatus = &value
		}
		history = append(history, RequisitionStatusHistoryResponse{
			Sequence:   entry.Sequence(),
			FromStatus: fromStatus,
			ToStatus:   entry.ToStatus().String(),
			ActedBy:    entry.ActedBy(),
			ActorRole:  entry.ActorRole().String(),
			Reason:     entry.Reason(),
			CreatedAt:  entry.CreatedAt(),
		})
	}

	currency := shareddomain.CurrencyNaira.String()
	if items := requisition.Items(); len(items) > 0 {
		currency = items[0].EstimatedUnitAmount().Currency().String()
	}

	return &RequisitionResponse{
		ID:                   requisition.ID(),
		ReferenceCode:        requisition.ReferenceCode(),
		RequestedBy:          requisition.RequestedBy(),
		DepartmentID:         requisition.DepartmentID(),
		Status:               requisition.Status().String(),
		Priority:             requisition.Priority().String(),
		NeededBy:             requisition.NeededBy().Value().UTC().Format(dateOnlyLayout),
		Justification:        requisition.Justification().Value(),
		Currency:             currency,
		Lines:                lines,
		TotalEstimatedAmount: total.ToFloat64(),
		StatusHistory:        history,
		CreatedAt:            requisition.CreatedAt(),
		UpdatedAt:            requisition.UpdatedAt(),
	}, nil
}
