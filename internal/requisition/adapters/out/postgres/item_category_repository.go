package requisitionpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres/models"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	sharedcache "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

var (
	ErrItemCategoryNotFoundRepo   = customerrors.NewNotFoundError("item category not found")
	ErrItemCategoryNameExistsRepo = customerrors.NewConflictError("item category with this name already exists")
)

const pqUniqueViolationCode = "23505"
const itemCategoryCacheTTL = 24 * time.Hour
const itemCategoryCachePrefix = "item_category"
const itemCategoryListCachePrefix = itemCategoryCachePrefix + ":list:"

func itemCategoryCacheKey(id string) string {
	return fmt.Sprintf("%s:%s", itemCategoryCachePrefix, id)
}

type ItemCategoryRepository struct {
	repository db.Repository
	cache      sharedcache.Cache
}

func NewItemCategoryRepository(repository db.Repository, cache sharedcache.Cache) *ItemCategoryRepository {
	return &ItemCategoryRepository{repository: repository, cache: cache}
}

var saveItemCategoryQuery = `
	INSERT INTO item_categories (id, name, description, status, created_at, updated_at)
	VALUES (:id, :name, :description, :status, :created_at, :updated_at)
`

func (r *ItemCategoryRepository) SaveItemCategory(ctx context.Context, category *requisitiondomain.ItemCategory) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveItemCategoryQuery, models.ItemCategoryFromDomain(category))
	if err != nil {
		return translateItemCategoryError(err)
	}

	if err := r.cache.DeleteByPrefix(ctx, itemCategoryListCachePrefix); err != nil {
		return err
	}

	return nil
}

var getItemCategoryByIDQuery = `
	SELECT id, name, description, status, created_at, updated_at
	FROM item_categories
	WHERE id = $1
`

func (r *ItemCategoryRepository) GetItemCategoryByID(ctx context.Context, id string) (*requisitiondomain.ItemCategory, error) {
	key := itemCategoryCacheKey(id)
	var model models.ItemCategoryModel

	found, err := r.cache.Get(ctx, key, &model)
	if err != nil {
		return nil, err
	}

	if found {
		return model.ToDomain()
	}

	if err := r.repository.Executor(ctx).GetContext(ctx, &model, getItemCategoryByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrItemCategoryNotFoundRepo
		}
		return nil, err
	}

	if err := r.cache.Set(ctx, key, model, itemCategoryCacheTTL); err != nil {
		return nil, err
	}

	return model.ToDomain()
}

var listItemCategoriesQuery = `
	SELECT id, name, description, status, created_at, updated_at
	FROM item_categories
	ORDER BY name ASC
	LIMIT $1 OFFSET $2
`

func (r *ItemCategoryRepository) ListItemCategories(ctx context.Context, limit int, offset int) ([]*requisitiondomain.ItemCategory, error) {
	key := fmt.Sprintf("%s%d:%d", itemCategoryListCachePrefix, limit, offset)
	var modelsList []models.ItemCategoryModel

	found, err := r.cache.Get(ctx, key, &modelsList)
	if err != nil {
		return nil, err
	}
	if found {
		categories := make([]*requisitiondomain.ItemCategory, 0, len(modelsList))
		for _, model := range modelsList {
			category, err := model.ToDomain()
			if err != nil {
				return nil, err
			}
			categories = append(categories, category)
		}
		return categories, nil
	}

	if err := r.repository.Executor(ctx).SelectContext(ctx, &modelsList, listItemCategoriesQuery, limit, offset); err != nil {
		return nil, err
	}

	categories := make([]*requisitiondomain.ItemCategory, 0, len(modelsList))
	for _, model := range modelsList {
		category, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err := r.cache.Set(ctx, key, modelsList, itemCategoryCacheTTL); err != nil {
		return nil, err
	}

	return categories, nil
}

var updateItemCategoryQuery = `
	UPDATE item_categories
	SET name = $2, description = $3, status = $4, updated_at = NOW()
	WHERE id = $1
`

func (r *ItemCategoryRepository) UpdateItemCategory(ctx context.Context, category *requisitiondomain.ItemCategory) error {
	res, err := r.repository.Executor(ctx).ExecContext(
		ctx,
		updateItemCategoryQuery,
		category.ID(),
		category.Name().Value(),
		category.Description().Value(),
		category.Status().String(),
	)
	if err != nil {
		return translateItemCategoryError(err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrItemCategoryNotFoundRepo
	}

	if err := r.cache.Delete(ctx, itemCategoryCacheKey(category.ID())); err != nil {
		return err
	}

	if err := r.cache.DeleteByPrefix(ctx, itemCategoryListCachePrefix); err != nil {
		return err
	}

	return nil
}

var listItemsInCategoryQuery = `
	SELECT id, name, description, category_id, unit_cost, status, created_at, updated_at
	FROM items
	WHERE category_id = $1
	ORDER BY name ASC
`

func (r *ItemCategoryRepository) ListItemsInCategory(ctx context.Context, categoryID string) ([]*requisitiondomain.Item, error) {
	var modelsList []models.ItemModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &modelsList, listItemsInCategoryQuery, categoryID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*requisitiondomain.Item{}, nil
		}
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

func translateItemCategoryError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && string(pqErr.Code) == pqUniqueViolationCode {
		return ErrItemCategoryNameExistsRepo
	}
	return err
}
