package models

import (
	"time"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

type RequisitionModel struct {
	ID            string    `db:"id"`
	ReferenceCode string    `db:"reference_code"`
	RequestedBy   string    `db:"requested_by"`
	DepartmentID  string    `db:"department_id"`
	Justification string    `db:"justification"`
	Priority      string    `db:"priority"`
	NeededBy      time.Time `db:"needed_by"`
	Status        string    `db:"status"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

type RequisitionItemModel struct {
	ID                  string    `db:"id"`
	RequisitionID       string    `db:"requisition_id"`
	LineNumber          int       `db:"line_number"`
	ItemID              *string   `db:"item_id"`
	RequestedItemID     *string   `db:"requested_item_id"`
	ItemNameSnapshot    string    `db:"item_name_snapshot"`
	Quantity            int       `db:"quantity"`
	EstimatedUnitAmount int64     `db:"estimated_unit_amount"`
	Notes               *string   `db:"notes"`
	CreatedAt           time.Time `db:"created_at"`
}

type RequestedItemModel struct {
	ID             string    `db:"id"`
	Name           string    `db:"name"`
	CategoryID     string    `db:"category_id"`
	Description    string    `db:"description"`
	CreatedBy      string    `db:"created_by"`
	ReviewStatus   string    `db:"review_status"`
	PromotedItemID *string   `db:"promoted_item_id"`
	CreatedAt      time.Time `db:"created_at"`
}

type RequisitionStatusHistoryModel struct {
	ID            string    `db:"id"`
	RequisitionID string    `db:"requisition_id"`
	Sequence      int       `db:"sequence"`
	FromStatus    *string   `db:"from_status"`
	ToStatus      string    `db:"to_status"`
	ActedBy       string    `db:"acted_by"`
	ActorRole     string    `db:"actor_role"`
	Reason        *string   `db:"reason"`
	CreatedAt     time.Time `db:"created_at"`
}

func RequisitionFromDomain(requisition *requisitiondomain.Requisition) *RequisitionModel {
	return &RequisitionModel{
		ID: requisition.ID(), ReferenceCode: requisition.ReferenceCode(),
		RequestedBy: requisition.RequestedBy(), DepartmentID: requisition.DepartmentID(),
		Justification: requisition.Justification().Value(),
		Priority:      requisition.Priority().String(), NeededBy: requisition.NeededBy(),
		Status: requisition.Status().String(), CreatedAt: requisition.CreatedAt(),
		UpdatedAt: requisition.UpdatedAt(),
	}
}

func RequisitionItemFromDomain(item *requisitiondomain.RequisitionItem) (*RequisitionItemModel, error) {
	itemID := nullableString(item.ItemID())
	requestedItemID := nullableString(item.RequestedItemID())
	return &RequisitionItemModel{
		ID: item.ID(), RequisitionID: item.RequisitionID(), LineNumber: item.LineNumber(),
		ItemID: itemID, RequestedItemID: requestedItemID,
		ItemNameSnapshot:    item.ItemNameSnapshot().Value(),
		Quantity:            int(item.Quantity().Value()),
		EstimatedUnitAmount: item.EstimatedUnitAmount().Amount(),
		Notes:               nullableString(item.Notes()), CreatedAt: item.CreatedAt(),
	}, nil
}

func (m RequisitionItemModel) ToDomain() (*requisitiondomain.RequisitionItem, error) {
	name, err := requisitiondomain.NewItemName(m.ItemNameSnapshot)
	if err != nil {
		return nil, err
	}
	quantity, err := requisitiondomain.NewRequisitionQuantity(m.Quantity)
	if err != nil {
		return nil, err
	}
	amount, err := shareddomain.NewNairaMoneyFromMinorUnits(m.EstimatedUnitAmount)
	if err != nil {
		return nil, err
	}
	return requisitiondomain.ReconstituteRequisitionItem(
		m.ID, m.RequisitionID, m.LineNumber, stringValue(m.ItemID),
		stringValue(m.RequestedItemID), name, quantity, amount, stringValue(m.Notes),
		m.CreatedAt,
	)
}

func (m RequestedItemModel) ToDomain() (*requisitiondomain.RequestedItem, error) {
	return requisitiondomain.ReconstituteRequestedItem(
		m.ID, m.Name, m.CategoryID, m.Description, m.CreatedBy,
		m.ReviewStatus, stringValue(m.PromotedItemID), m.CreatedAt,
	)
}

func RequestedItemFromDomain(item *requisitiondomain.RequestedItem) *RequestedItemModel {
	return &RequestedItemModel{
		ID: item.ID(), Name: item.Name().Value(), CategoryID: item.CategoryID(),
		Description: item.Description().Value(), CreatedBy: item.CreatedBy(),
		ReviewStatus:   item.ReviewStatus().String(),
		PromotedItemID: nullableString(item.PromotedItemID()), CreatedAt: item.CreatedAt(),
	}
}

func (m RequisitionStatusHistoryModel) ToDomain() (*requisitiondomain.RequisitionStatusHistory, error) {
	var fromStatus *requisitiondomain.RequisitionStatus
	if m.FromStatus != nil {
		value, err := requisitiondomain.NewRequisitionStatus(*m.FromStatus)
		if err != nil {
			return nil, err
		}
		fromStatus = &value
	}
	toStatus, err := requisitiondomain.NewRequisitionStatus(m.ToStatus)
	if err != nil {
		return nil, err
	}
	return requisitiondomain.NewRequisitionStatusHistory(
		m.ID, m.RequisitionID, m.Sequence, fromStatus,
		toStatus, m.ActedBy,
		requisitiondomain.RequisitionActorRole(m.ActorRole),
		stringValue(m.Reason), m.CreatedAt,
	)
}

func StatusHistoryFromDomain(history *requisitiondomain.RequisitionStatusHistory) *RequisitionStatusHistoryModel {
	var fromStatus *string
	if history.FromStatus() != nil {
		value := history.FromStatus().String()
		fromStatus = &value
	}
	return &RequisitionStatusHistoryModel{
		ID: history.ID(), RequisitionID: history.RequisitionID(), Sequence: history.Sequence(),
		FromStatus: fromStatus, ToStatus: history.ToStatus().String(),
		ActedBy: history.ActedBy(), ActorRole: history.ActorRole().String(),
		Reason: nullableString(history.Reason()), CreatedAt: history.CreatedAt(),
	}
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
