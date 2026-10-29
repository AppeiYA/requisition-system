package domain

import (
	"strings"
	"time"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

const RequestedItemPrefixID = "ritem"

type RequestedItemReviewStatus string

const (
	RequestedItemUnreviewed RequestedItemReviewStatus = "unreviewed"
	RequestedItemPromoted   RequestedItemReviewStatus = "promoted"
	RequestedItemDismissed  RequestedItemReviewStatus = "dismissed"
)

var (
	ErrInvalidRequestedItemID     = customerrors.NewBadRequestError("invalid requested item ID")
	ErrInvalidRequestedItemStatus = customerrors.NewBadRequestError("invalid requested item review status")
	ErrPromotedItemIDMustBeEmpty  = customerrors.NewBadRequestError("promoted item ID must be empty unless the requested item is promoted")
	ErrPromotedItemIDRequired	 = customerrors.NewBadRequestError("promoted item ID is required when the requested item is promoted")
)

type RequestedItem struct {
	id             string
	name           *ItemName
	categoryID     string
	description    *ItemDescription
	createdBy      string
	reviewStatus   RequestedItemReviewStatus
	promotedItemID string
	createdAt      time.Time
}

func NewRequestedItem(
	id string,
	name *ItemName,
	categoryID string,
	description *ItemDescription,
	createdBy string,
) (*RequestedItem, error) {
	return newRequestedItem(id, name, categoryID, description, createdBy, RequestedItemUnreviewed, "", time.Now())
}

func ReconstituteRequestedItem(
	id string,
	name string,
	categoryID string,
	description string,
	createdBy string,
	reviewStatus string,
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
	return newRequestedItem(id, itemName, categoryID, itemDescription, createdBy, status, promotedItemID, createdAt)
}

func newRequestedItem(
	id string,
	name *ItemName,
	categoryID string,
	description *ItemDescription,
	createdBy string,
	reviewStatus RequestedItemReviewStatus,
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
	return &RequestedItem{
		id: id, name: name, categoryID: categoryID, description: description,
		createdBy: createdBy, reviewStatus: reviewStatus, promotedItemID: promotedItemID,
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

func (s RequestedItemReviewStatus) String() string               { return string(s) }
func (i *RequestedItem) ID() string                              { return i.id }
func (i *RequestedItem) Name() *ItemName                         { return i.name }
func (i *RequestedItem) CategoryID() string                      { return i.categoryID }
func (i *RequestedItem) Description() *ItemDescription           { return i.description }
func (i *RequestedItem) CreatedBy() string                       { return i.createdBy }
func (i *RequestedItem) ReviewStatus() RequestedItemReviewStatus { return i.reviewStatus }
func (i *RequestedItem) PromotedItemID() string                  { return i.promotedItemID }
func (i *RequestedItem) CreatedAt() time.Time                    { return i.createdAt }
