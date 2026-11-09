package domain

import (
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvalidAttachmentRefName = customerrors.NewInternalServerError("invalid attachment ref name")
	ErrInvalidAttachmentRefID = customerrors.NewInternalServerError("invalid attachment ref ID")
)

type AttachmentRefName string
const (
	RefNameRequisition AttachmentRefName = "requisition"
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
func (r AttachmentRefName) String() string {
	return string(r)
}

var RefNameToIDSyntax = map[AttachmentRefName]string{
	RefNameRequisition: "req",
	RefNameRequisitionItem: "reqitem",
}

type AttachmentRef struct {
	refName AttachmentRefName
	id string
}

func NewAttachmentRef(refName AttachmentRefName, id string) (*AttachmentRef, error) {
	if !refName.IsValid() {
		return nil, ErrInvalidAttachmentRefName
	}
	if id == "" {
		return nil, ErrInvalidAttachmentRefID
	}

	if value, ok := RefNameToIDSyntax[refName]; !ok {
		return nil, ErrInvalidAttachmentRefName
	}else {
		if !strings.HasPrefix(id, value) {
			return nil, ErrInvalidAttachmentRefID
		}
	}

	return &AttachmentRef{
		refName: refName,
		id: id,
	}, nil
}

func ReconstructAttachmentRef(id string) (*AttachmentRef, error) {
	for refName, prefix := range RefNameToIDSyntax {
		if strings.HasPrefix(id, prefix) {
			return &AttachmentRef{
				refName: refName,
				id: id,
			}, nil
		}
	}
	return nil, ErrInvalidAttachmentRefID
}

func (r *AttachmentRef) RefName() AttachmentRefName {
	return r.refName
}

func (r *AttachmentRef) ID() string {
	return r.id
}
