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

// UnmarshalJSON accepts both to_status and toStatus.
func (r *TransitionRequisitionStatusRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		ToStatus      *string `json:"to_status"`
		ToStatusCamel *string `json:"toStatus"`
		Reason        *string `json:"reason"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = TransitionRequisitionStatusRequest{
		ToStatus: firstString(raw.ToStatus, raw.ToStatusCamel),
		Reason:   firstString(raw.Reason, nil),
	}
	return nil
}

// UnmarshalJSON accepts both the API's snake_case field names and their
// camelCase equivalents so older clients can migrate without a flag day.
func (r *CreateRequisitionRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		DepartmentID      *string                        `json:"department_id"`
		DepartmentIDCamel *string                        `json:"departmentId"`
		Priority          *string                        `json:"priority"`
		NeededBy          *Date                          `json:"needed_by"`
		NeededByCamel     *Date                          `json:"neededBy"`
		Justification     *string                        `json:"justification"`
		Lines             []CreateRequisitionLineRequest `json:"lines"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = CreateRequisitionRequest{
		DepartmentID:  firstString(raw.DepartmentID, raw.DepartmentIDCamel),
		Priority:      firstString(raw.Priority, nil),
		NeededBy:      firstDate(raw.NeededBy, raw.NeededByCamel),
		Justification: firstString(raw.Justification, nil),
		Lines:         raw.Lines,
	}
	return nil
}

// UnmarshalJSON accepts both item_id/requested_item/estimated_unit_amount and
// itemId/requestedItem/estimatedUnitAmount.
func (l *CreateRequisitionLineRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		ItemID                   *string                     `json:"item_id"`
		ItemIDCamel              *string                     `json:"itemId"`
		RequestedItem            *CreateRequestedItemRequest `json:"requested_item"`
		RequestedItemCamel       *CreateRequestedItemRequest `json:"requestedItem"`
		Quantity                 *int                        `json:"quantity"`
		EstimatedUnitAmount      *float64                    `json:"estimated_unit_amount"`
		EstimatedUnitAmountCamel *float64                    `json:"estimatedUnitAmount"`
		Notes                    *string                     `json:"notes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	requestedItem := raw.RequestedItem
	if requestedItem == nil {
		requestedItem = raw.RequestedItemCamel
	}
	*l = CreateRequisitionLineRequest{
		ItemID:              firstString(raw.ItemID, raw.ItemIDCamel),
		RequestedItem:       requestedItem,
		Quantity:            firstInt(raw.Quantity),
		EstimatedUnitAmount: firstFloat64(raw.EstimatedUnitAmount, raw.EstimatedUnitAmountCamel),
		Notes:               firstString(raw.Notes, nil),
	}
	return nil
}

// UnmarshalJSON accepts both category_id and categoryId.
func (i *CreateRequestedItemRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		Name            *string `json:"name"`
		CategoryID      *string `json:"category_id"`
		CategoryIDCamel *string `json:"categoryId"`
		Description     *string `json:"description"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*i = CreateRequestedItemRequest{
		Name:        firstString(raw.Name, nil),
		CategoryID:  firstString(raw.CategoryID, raw.CategoryIDCamel),
		Description: firstString(raw.Description, nil),
	}
	return nil
}

func firstString(values ...*string) string {
	for _, value := range values {
		if value != nil {
			return *value
		}
	}
	return ""
}

func firstDate(values ...*Date) Date {
	for _, value := range values {
		if value != nil {
			return *value
		}
	}
	return ""
}

func firstInt(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func firstFloat64(values ...*float64) float64 {
	for _, value := range values {
		if value != nil {
			return *value
		}
	}
	return 0
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
		NeededBy:             requisition.NeededBy().UTC().Format(dateOnlyLayout),
		Justification:        requisition.Justification().Value(),
		Currency:             currency,
		Lines:                lines,
		TotalEstimatedAmount: total.ToFloat64(),
		StatusHistory:        history,
		CreatedAt:            requisition.CreatedAt(),
		UpdatedAt:            requisition.UpdatedAt(),
	}, nil
}
