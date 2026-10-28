package requisitionpostgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres/models"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	sharedcache "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

var (
	ErrItemNotFoundRepo   = customerrors.NewNotFoundError("item not found")
	ErrItemNameExistsRepo = customerrors.NewConflictError("item with this name already exists in this category")
)

const pqForeignKeyViolationCode = "23503"
const departmentCacheTTL = 10 * time.Minute
const itemCachePrefix = "item"
const itemListCachePrefix = itemCachePrefix + ":list:"

type ItemRepository struct {
	repository db.Repository
	cache      sharedcache.Cache
}

func NewItemRepository(repository db.Repository, cache sharedcache.Cache) *ItemRepository {
	return &ItemRepository{repository: repository, cache: cache}
}

var saveItemQuery = `
	INSERT INTO items (id, name, description, category_id, unit_cost, status, created_at, updated_at)
	VALUES (:id, :name, :description, :category_id, :unit_cost, :status, :created_at, :updated_at)
`

func (r *ItemRepository) SaveItem(ctx context.Context, item *requisitiondomain.Item) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveItemQuery, models.ItemFromDomain(item))
	if err != nil {
		return translateItemError(err)
	}
	if err := r.cache.DeleteByPrefix(ctx, itemListCachePrefix); err != nil {
		return err
	}
	return nil
}

var getItemByIDQuery = `
	SELECT id, name, description, category_id, unit_cost, status, created_at, updated_at
	FROM items
	WHERE id = $1
`

func (r *ItemRepository) GetItemByID(ctx context.Context, id string) (*requisitiondomain.Item, error) {
	key := itemCachePrefix + ":" + id
	var model models.ItemModel

	found, err := r.cache.Get(ctx, key, &model)
	if err != nil {
		return nil, err
	}

	if found {
		return model.ToDomain()
	}

	if err := r.repository.Executor(ctx).GetContext(ctx, &model, getItemByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrItemNotFoundRepo
		}
		return nil, err
	}

	if err := r.cache.Set(ctx, key, model, departmentCacheTTL); err != nil {
		return nil, err
	}

	return model.ToDomain()
}

var listItemsQuery = `
	SELECT id, name, description, category_id, unit_cost, status, created_at, updated_at
	FROM items
	ORDER BY name ASC
	LIMIT $1 OFFSET $2
`

func (r *ItemRepository) ListItems(ctx context.Context, limit int, offset int) ([]*requisitiondomain.Item, error) {
	var modelsList []models.ItemModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &modelsList, listItemsQuery, limit, offset); err != nil {
		return nil, err
	}

	items := make([]*requisitiondomain.Item, 0, len(modelsList))
	for _, model := range modelsList {
		item, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

var updateItemQuery = `
	UPDATE items
	SET name = $2, description = $3, category_id = $4, unit_cost = $5, status = $6, updated_at = NOW()
	WHERE id = $1
`

func (r *ItemRepository) UpdateItem(ctx context.Context, item *requisitiondomain.Item) error {
	key := itemCachePrefix + ":" + item.ID()
	res, err := r.repository.Executor(ctx).ExecContext(
		ctx,
		updateItemQuery,
		item.ID(),
		item.Name().Value(),
		item.Description().Value(),
		item.CategoryID(),
		item.UnitCost().Amount(),
		item.Status().String(),
	)
	if err != nil {
		return translateItemError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrItemNotFoundRepo
	}

	if err := r.cache.Delete(ctx, key); err != nil {
		return err
	}

	if err := r.cache.DeleteByPrefix(ctx, itemListCachePrefix); err != nil {
		return err
	}

	return nil
}

func translateItemError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch string(pqErr.Code) {
		case pqUniqueViolationCode:
			return ErrItemNameExistsRepo
		case pqForeignKeyViolationCode:
			return ErrItemCategoryNotFoundRepo
		}
	}
	return err
}
