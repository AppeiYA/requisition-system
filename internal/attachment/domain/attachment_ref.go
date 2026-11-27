package domain

import (
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvalidAttachmentRefName = customerrors.NewBadRequestError("invalid attachment ref name")
	ErrInvalidAttachmentRefID   = customerrors.NewBadRequestError("invalid attachment ref ID")
)

type AttachmentRefName string

const (
	RefNameRequisition     AttachmentRefName = "requisition"
	RefNameRequisitionItem AttachmentRefName = "requisition_item"
)

func (r AttachmentRefName) IsValid() bool {
	switch r {
	case RefNameRequisition, RefNameRequisitionItem:
		return true
	default:
		return false
	}
}

func (r AttachmentRefName) String() string { return string(r) }

var RefNameToIDSyntax = map[AttachmentRefName]string{
	RefNameRequisition:     "req",
	RefNameRequisitionItem: "reqitem",
}

type AttachmentRef struct {
	refName AttachmentRefName
	id      string
}

func NewAttachmentRef(refName AttachmentRefName, id string) (*AttachmentRef, error) {
	if !refName.IsValid() {
		return nil, ErrInvalidAttachmentRefName
	}
	prefix, ok := RefNameToIDSyntax[refName]
	if !ok || !strings.HasPrefix(id, prefix+"-") || id == prefix+"-" {
		return nil, ErrInvalidAttachmentRefID
	}

	return &AttachmentRef{refName: refName, id: id}, nil
}

func (r *AttachmentRef) RefName() AttachmentRefName { return r.refName }
func (r *AttachmentRef) ID() string                 { return r.id }
