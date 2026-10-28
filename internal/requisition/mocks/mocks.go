package requisitionmocks

import (
	"context"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

var (
	DefaultItemCategoryID          = "cat-1234"
	DefaultItemCategoryName        = "Office Supplies"
	DefaultItemCategoryDescription = "General office supplies"
)

func ValidItemCategory() *requisitiondomain.ItemCategory {
	name, _ := requisitiondomain.NewCategoryName(DefaultItemCategoryName)
	description, _ := requisitiondomain.NewCategoryDescription(DefaultItemCategoryDescription)

	category, _ := requisitiondomain.NewItemCategory(DefaultItemCategoryID, name, description)
	return category
}

type MockItemCategoryRepository struct {
	SaveItemCategoryFunc    func(ctx context.Context, category *requisitiondomain.ItemCategory) error
	GetItemCategoryByIDFunc func(ctx context.Context, id string) (*requisitiondomain.ItemCategory, error)
	ListItemCategoriesFunc  func(ctx context.Context, limit int, offset int) ([]*requisitiondomain.ItemCategory, error)
	ListItemsInCategoryFunc func(ctx context.Context, categoryID string) ([]*requisitiondomain.Item, error)
	UpdateItemCategoryFunc  func(ctx context.Context, category *requisitiondomain.ItemCategory) error
}

func NewMockItemCategoryRepository() *MockItemCategoryRepository {
	return &MockItemCategoryRepository{
		SaveItemCategoryFunc: func(ctx context.Context, category *requisitiondomain.ItemCategory) error {
			return nil
		},
		GetItemCategoryByIDFunc: func(ctx context.Context, id string) (*requisitiondomain.ItemCategory, error) {
			return ValidItemCategory(), nil
		},
		ListItemCategoriesFunc: func(ctx context.Context, limit int, offset int) ([]*requisitiondomain.ItemCategory, error) {
			return []*requisitiondomain.ItemCategory{ValidItemCategory()}, nil
		},
		ListItemsInCategoryFunc: func(ctx context.Context, categoryID string) ([]*requisitiondomain.Item, error) {
			return []*requisitiondomain.Item{}, nil
		},
		UpdateItemCategoryFunc: func(ctx context.Context, category *requisitiondomain.ItemCategory) error {
			return nil
		},
	}
}

func (m *MockItemCategoryRepository) SaveItemCategory(ctx context.Context, category *requisitiondomain.ItemCategory) error {
	return m.SaveItemCategoryFunc(ctx, category)
}

func (m *MockItemCategoryRepository) GetItemCategoryByID(ctx context.Context, id string) (*requisitiondomain.ItemCategory, error) {
	return m.GetItemCategoryByIDFunc(ctx, id)
}

func (m *MockItemCategoryRepository) ListItemCategories(ctx context.Context, limit int, offset int) ([]*requisitiondomain.ItemCategory, error) {
	return m.ListItemCategoriesFunc(ctx, limit, offset)
}

func (m *MockItemCategoryRepository) ListItemsInCategory(ctx context.Context, categoryID string) ([]*requisitiondomain.Item, error) {
	return m.ListItemsInCategoryFunc(ctx, categoryID)
}

func (m *MockItemCategoryRepository) UpdateItemCategory(ctx context.Context, category *requisitiondomain.ItemCategory) error {
	return m.UpdateItemCategoryFunc(ctx, category)
}

var _ requisitionout.ItemCategoryRepository = (*MockItemCategoryRepository)(nil)

type MockIDGenerator struct {
	GenerateIDFunc func(prefix string) (string, error)
}

func NewMockIDGenerator() *MockIDGenerator {
	return &MockIDGenerator{
		GenerateIDFunc: func(prefix string) (string, error) {
			return prefix + "-1234", nil
		},
	}
}

func (m *MockIDGenerator) GenerateID(prefix string) (string, error) {
	return m.GenerateIDFunc(prefix)
}

var _ requisitionout.IDGenerator = (*MockIDGenerator)(nil)
