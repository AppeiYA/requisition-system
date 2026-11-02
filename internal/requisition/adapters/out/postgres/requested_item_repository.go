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
	var model models.RequestedItemModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, `
		SELECT id, name, category_id, description, created_by,
			review_status, promoted_item_id, created_at
		FROM requested_items WHERE id = $1
	`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRequestedItemNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

var _ requisitionout.RequestedItemRepository = (*RequestedItemRepository)(nil)
