package domain

import customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"

type MemberJobTitle struct {
	value string
}

var (
	ErrEmptyJobTitle = customerrors.NewBadRequestError("Job title cannot be empty")
	ErrJobTitleTooLong = customerrors.NewBadRequestError("Job title cannot exceed 60 characters")
)

func NewMemberJobTitle(value string) (*MemberJobTitle, error) {
	if value == "" {
		return nil, ErrEmptyJobTitle
	}

	if len(value) > 60 {
		return nil, ErrJobTitleTooLong
	}

	return &MemberJobTitle{value: value}, nil
}

func (mjt *MemberJobTitle) Value() string {
	return mjt.value
}