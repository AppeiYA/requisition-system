package domain

import (
	"strings"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type RequisitionActorRole string

const (
	ActorRoleAdmin RequisitionActorRole = "admin"
	ActorRoleStaff RequisitionActorRole = "staff"
)

var (
	ErrInvalidStatusHistoryID = customerrors.NewBadRequestError("invalid status history ID")
	ErrInvalidStatusSequence  = customerrors.NewBadRequestError("status history sequence must be greater than zero")
	ErrStatusReasonRequired   = customerrors.NewBadRequestError("reason is required for rejected or cancelled status")
	ErrStatusMustChange       = customerrors.NewBadRequestError("status must change")
	ErrInvalidActorRole	  = customerrors.NewBadRequestError("invalid actor role")
)

const StatusHistoryPrefixID = "rsh"

type RequisitionStatusHistory struct {
	id            string
	requisitionID string
	sequence      int
	fromStatus    *RequisitionStatus
	toStatus      RequisitionStatus
	actedBy       string
	actorRole     RequisitionActorRole
	reason        string
	createdAt     time.Time
}

func NewRequisitionStatusHistory(
	id string,
	requisitionID string,
	sequence int,
	fromStatus *RequisitionStatus,
	toStatus RequisitionStatus,
	actedBy string,
	actorRole RequisitionActorRole,
	reason string,
	createdAt time.Time,
) (*RequisitionStatusHistory, error) {
	if id == "" || !strings.HasPrefix(id, StatusHistoryPrefixID) {
		return nil, ErrInvalidStatusHistoryID
	}
	if requisitionID == "" || !strings.HasPrefix(requisitionID, RequisitionPrefixID) {
		return nil, ErrInvalidRequisitionID
	}
	if sequence <= 0 {
		return nil, ErrInvalidStatusSequence
	}
	if fromStatus != nil && !fromStatus.IsValid() {
		return nil, ErrInvalidRequisitionStatus
	}
	if !toStatus.IsValid() {
		return nil, ErrInvalidRequisitionStatus
	}
	if fromStatus != nil && *fromStatus == toStatus {
		return nil, ErrStatusMustChange
	}
	if actedBy == "" {
		return nil, ErrRequestedIDCannotBeEmpty
	}
	if !actorRole.IsValid() {
		return nil, ErrInvalidActorRole
	}
	if (toStatus == RequisitionStatusRejected || toStatus == RequisitionStatusCancelled) && strings.TrimSpace(reason) == "" {
		return nil, ErrStatusReasonRequired
	}
	return &RequisitionStatusHistory{
		id: id, requisitionID: requisitionID, sequence: sequence,
		fromStatus: fromStatus, toStatus: toStatus, actedBy: actedBy,
		actorRole: actorRole, reason: reason, createdAt: createdAt,
	}, nil
}

func (r RequisitionActorRole) IsValid() bool {
	return r == ActorRoleAdmin || r == ActorRoleStaff
}

func (r RequisitionActorRole) String() string                       { return string(r) }
func (h *RequisitionStatusHistory) ID() string                      { return h.id }
func (h *RequisitionStatusHistory) RequisitionID() string           { return h.requisitionID }
func (h *RequisitionStatusHistory) Sequence() int                   { return h.sequence }
func (h *RequisitionStatusHistory) FromStatus() *RequisitionStatus  { return h.fromStatus }
func (h *RequisitionStatusHistory) ToStatus() RequisitionStatus     { return h.toStatus }
func (h *RequisitionStatusHistory) ActedBy() string                 { return h.actedBy }
func (h *RequisitionStatusHistory) ActorRole() RequisitionActorRole { return h.actorRole }
func (h *RequisitionStatusHistory) Reason() string                  { return h.reason }
func (h *RequisitionStatusHistory) CreatedAt() time.Time            { return h.createdAt }
