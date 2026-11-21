package requisitionpostgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres/models"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrRequestedItemNotFound = customerrors.NewNotFoundError("requested item not found")

type RequestedItemRepository struct {
	repository db.Repository
}

func NewRequestedItemRepository(repository db.Repository) *RequestedItemRepository {
	return &RequestedItemRepository{repository: repository}
}

func (r *RequestedItemRepository) SaveRequestedItem(ctx context.Context, item *requisitiondomain.RequestedItem) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, `
		INSERT INTO requested_items (
			id, name, category_id, description, created_by,
			review_status, promoted_item_id, created_at
		) VALUES (
			:id, :name, :category_id, :description, :created_by,
			:review_status, :promoted_item_id, :created_at
		)
	`, models.RequestedItemFromDomain(item))
	return err
}

func (r *RequestedItemRepository) GetRequestedItemByID(ctx context.Context, id string) (*requisitiondomain.RequestedItem, error) {
	return r.getRequestedItemByID(ctx, id, false)
}

func (r *RequestedItemRepository) GetRequestedItemByIDForUpdate(ctx context.Context, id string) (*requisitiondomain.RequestedItem, error) {
	return r.getRequestedItemByID(ctx, id, true)
}

func (r *RequestedItemRepository) getRequestedItemByID(ctx context.Context, id string, forUpdate bool) (*requisitiondomain.RequestedItem, error) {
	var model models.RequestedItemModel
	lock := ""
	if forUpdate {
		lock = " FOR UPDATE"
	}
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, `
		SELECT id, name, category_id, description, created_by,
			review_status, promoted_item_id, reviewed_by, reviewed_at, created_at
		FROM requested_items WHERE id = $1`+lock, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRequestedItemNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *RequestedItemRepository) ListUnreviewedRequestedItems(ctx context.Context, limit int, offset int) ([]*requisitiondomain.RequestedItem, error) {
	var itemModels []models.RequestedItemModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &itemModels, `
		SELECT id, name, category_id, description, created_by,
			review_status, promoted_item_id, reviewed_by, reviewed_at, created_at
		FROM requested_items
		WHERE review_status = 'unreviewed'
		ORDER BY created_at ASC, id ASC
		LIMIT $1 OFFSET $2
	`, limit, offset); err != nil {
		return nil, err
	}

	items := make([]*requisitiondomain.RequestedItem, 0, len(itemModels))
	for _, model := range itemModels {
		item, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *RequestedItemRepository) UpdateRequestedItemStatus(ctx context.Context, item *requisitiondomain.RequestedItem) error {
	result, err := r.repository.Executor(ctx).ExecContext(ctx, `
		UPDATE requested_items
		SET review_status = $2, promoted_item_id = NULLIF($3, ''),
			reviewed_by = $4, reviewed_at = $5
		WHERE id = $1 AND review_status = 'unreviewed'
	`, item.ID(), item.ReviewStatus().String(), item.PromotedItemID(),
		item.ReviewedBy(), item.ReviewedAt())
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return customerrors.NewConflictError("requested item has already been reviewed")
	}
	return nil
}

var _ requisitionout.RequestedItemRepository = (*RequestedItemRepository)(nil)
