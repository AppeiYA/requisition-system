package domain

import (
	"strings"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const RequestedItemPrefixID = "ritem"
const reviewerIDPrefix = "usr"

type RequestedItemReviewStatus string

const (
	RequestedItemUnreviewed RequestedItemReviewStatus = "unreviewed"
	RequestedItemPromoted   RequestedItemReviewStatus = "promoted"
	RequestedItemDismissed  RequestedItemReviewStatus = "dismissed"
)

var validRequestedItemReviewStatusTransition = map[RequestedItemReviewStatus][]RequestedItemReviewStatus{
	RequestedItemUnreviewed: {RequestedItemPromoted, RequestedItemDismissed},
	RequestedItemPromoted:   {},
	RequestedItemDismissed:  {},
}

var (
	ErrInvalidRequestedItemID          = customerrors.NewBadRequestError("invalid requested item ID")
	ErrInvalidRequestedItemStatus      = customerrors.NewBadRequestError("invalid requested item review status")
	ErrPromotedItemIDMustBeEmpty       = customerrors.NewBadRequestError("promoted item ID must be empty unless the requested item is promoted")
	ErrPromotedItemIDRequired          = customerrors.NewBadRequestError("promoted item ID is required when the requested item is promoted")
	ErrRequestedItemNotReferenced      = customerrors.NewBadRequestError("requested item must be referenced by exactly one requisition line")
	ErrReviewMetadataForUnreviewedItem = customerrors.NewBadRequestError("reviewedBy and reviewedAt must be nil for unreviewed requested items")
	ErrReviewMetadataForReviewedItem   = customerrors.NewBadRequestError("reviewedBy and reviewedAt must be non-nil for reviewed requested items")
	ErrReviewedAtBeforeCreatedAt       = customerrors.NewBadRequestError("reviewedAt must be after or equal to createdAt")
	ErrInvalidReviewerID               = customerrors.NewBadRequestError("reviewedBy must be a valid user ID")
)

type RequestedItem struct {
	id             string
	name           *ItemName
	categoryID     string
	description    *ItemDescription
	createdBy      string
	reviewStatus   RequestedItemReviewStatus
	promotedItemID string
	reviewedBy     *string
	reviewedAt     *time.Time
	createdAt      time.Time
}

func NewRequestedItem(
	id string,
	name *ItemName,
	categoryID string,
	description *ItemDescription,
	createdBy string,
) (*RequestedItem, error) {
	return newRequestedItem(id, name, categoryID, description, createdBy, RequestedItemUnreviewed, nil, nil, "", time.Now())
}

func ReconstituteRequestedItem(
	id string,
	name string,
	categoryID string,
	description string,
	createdBy string,
	reviewStatus string,
	reviewedBy *string,
	reviewedAt *time.Time,
	promotedItemID string,
	createdAt time.Time,
) (*RequestedItem, error) {
	itemName, err := NewItemName(name)
	if err != nil {
		return nil, err
	}
	itemDescription, err := NewItemDescription(description)
	if err != nil {
		return nil, err
	}
	status, err := NewRequestedItemReviewStatus(reviewStatus)
	if err != nil {
		return nil, err
	}
	return newRequestedItem(id, itemName, categoryID, itemDescription, createdBy, status, reviewedBy, reviewedAt, promotedItemID, createdAt)
}

func newRequestedItem(
	id string,
	name *ItemName,
	categoryID string,
	description *ItemDescription,
	createdBy string,
	reviewStatus RequestedItemReviewStatus,
	reviewedBy *string,
	reviewedAt *time.Time,
	promotedItemID string,
	createdAt time.Time,
) (*RequestedItem, error) {
	if id == "" || !strings.HasPrefix(id, RequestedItemPrefixID) {
		return nil, ErrInvalidRequestedItemID
	}
	if name == nil {
		return nil, ErrNilFields("name")
	}
	if categoryID == "" || !strings.HasPrefix(categoryID, CategoryPrefixID) {
		return nil, ErrInvalidCategoryID
	}
	if description == nil {
		return nil, ErrNilFields("description")
	}
	if createdBy == "" {
		return nil, ErrRequestedIDCannotBeEmpty
	}
	if promotedItemID != "" && !strings.HasPrefix(promotedItemID, ItemPrefixID) {
		return nil, ErrInvalidItemID
	}
	if reviewStatus == RequestedItemPromoted && promotedItemID == "" {
		return nil, ErrPromotedItemIDRequired
	}
	if reviewStatus != RequestedItemPromoted && promotedItemID != "" {
		return nil, ErrPromotedItemIDMustBeEmpty
	}
	if reviewStatus == RequestedItemUnreviewed && (reviewedBy != nil || reviewedAt != nil) {
		return nil, ErrReviewMetadataForUnreviewedItem
	}
	if reviewStatus != RequestedItemUnreviewed && (reviewedBy == nil || reviewedAt == nil) {
		return nil, ErrReviewMetadataForReviewedItem
	}
	if reviewedAt != nil && !dateIsAfterOrEqualTo(createdAt, *reviewedAt) {
		return nil, ErrReviewedAtBeforeCreatedAt
	}
	if reviewedBy != nil && (*reviewedBy == "" || !strings.HasPrefix(*reviewedBy, reviewerIDPrefix)) {
		return nil, ErrInvalidReviewerID
	}
	return &RequestedItem{
		id: id, name: name, categoryID: categoryID, description: description,
		createdBy: createdBy, reviewStatus: reviewStatus, reviewedBy: reviewedBy, reviewedAt: reviewedAt, promotedItemID: promotedItemID,
		createdAt: createdAt,
	}, nil
}

func NewRequestedItemReviewStatus(value string) (RequestedItemReviewStatus, error) {
	status := RequestedItemReviewStatus(value)
	if !status.IsValid() {
		return "", ErrInvalidRequestedItemStatus
	}
	return status, nil
}

func (s RequestedItemReviewStatus) IsValid() bool {
	return s == RequestedItemUnreviewed || s == RequestedItemPromoted || s == RequestedItemDismissed
}

func dateIsAfterOrEqualTo(baseTime time.Time, value time.Time) bool {
	return value.After(baseTime) || value.Equal(baseTime)
}

func (i *RequestedItem) IsUnreviewed() bool {
	return i.reviewStatus == RequestedItemUnreviewed
}

func (i *RequestedItem) IsPromoted() bool {
	return i.reviewStatus == RequestedItemPromoted
}

func (i *RequestedItem) IsDismissed() bool {
	return i.reviewStatus == RequestedItemDismissed
}

func (i *RequestedItem) MarkAsDismissed(reviewedBy string, reviewedAt time.Time) error {
	if err := i.validateReviewTransition(RequestedItemDismissed, reviewedBy, reviewedAt); err != nil {
		return err
	}
	i.reviewStatus = RequestedItemDismissed
	i.reviewedBy = &reviewedBy
	i.reviewedAt = &reviewedAt
	return nil
}

func (i *RequestedItem) MarkAsPromoted(promotedItemID, reviewedBy string, reviewedAt time.Time) error {
	if promotedItemID == "" || !strings.HasPrefix(promotedItemID, ItemPrefixID) {
		return ErrInvalidItemID
	}
	if err := i.validateReviewTransition(RequestedItemPromoted, reviewedBy, reviewedAt); err != nil {
		return err
	}
	i.reviewStatus = RequestedItemPromoted
	i.promotedItemID = promotedItemID
	i.reviewedBy = &reviewedBy
	i.reviewedAt = &reviewedAt
	return nil
}

func (s RequestedItemReviewStatus) String() string               { return string(s) }
func (i *RequestedItem) ID() string                              { return i.id }
func (i *RequestedItem) Name() *ItemName                         { return i.name }
func (i *RequestedItem) CategoryID() string                      { return i.categoryID }
func (i *RequestedItem) Description() *ItemDescription           { return i.description }
func (i *RequestedItem) CreatedBy() string                       { return i.createdBy }
func (i *RequestedItem) ReviewStatus() RequestedItemReviewStatus { return i.reviewStatus }
func (i *RequestedItem) PromotedItemID() string                  { return i.promotedItemID }
func (i *RequestedItem) ReviewedBy() *string                     { return i.reviewedBy }
func (i *RequestedItem) ReviewedAt() *time.Time                  { return i.reviewedAt }
func (i *RequestedItem) CreatedAt() time.Time                    { return i.createdAt }

func (i *RequestedItem) canTransitionTo(newStatus RequestedItemReviewStatus) bool {
	validTransitions, exists := validRequestedItemReviewStatusTransition[i.reviewStatus]
	if !exists {
		return false
	}
	for _, validStatus := range validTransitions {
		if newStatus == validStatus {
			return true
		}
	}
	return false
}

func (i *RequestedItem) validateReviewTransition(newStatus RequestedItemReviewStatus, reviewedBy string, reviewedAt time.Time) error {
	if !i.canTransitionTo(newStatus) {
		return ErrInvalidRequestedItemStatus
	}
	if reviewedBy == "" || !strings.HasPrefix(reviewedBy, reviewerIDPrefix) {
		return ErrInvalidReviewerID
	}
	if !dateIsAfterOrEqualTo(i.createdAt, reviewedAt) {
		return ErrReviewedAtBeforeCreatedAt
	}
	return nil
}
