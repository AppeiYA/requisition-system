package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionmocks "github.com/AppeiYA/requisition-system/internal/requisition/mocks"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

func strPtr(value string) *string { return &value }

func Test_CreateItemCategoryUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	idGen := requisitionmocks.NewMockIDGenerator()

	var saved *domain.ItemCategory
	repo.SaveItemCategoryFunc = func(ctx context.Context, category *domain.ItemCategory) error {
		saved = category
		return nil
	}

	uc := NewCreateItemCategoryUsecase(repo, idGen)
	category, err := uc.Execute(context.Background(), categoryin.CreateItemCategoryCommand{
		Name:        requisitionmocks.DefaultItemCategoryName,
		Description: requisitionmocks.DefaultItemCategoryDescription,
	})

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if saved == nil {
		t.Fatal("expected the category to be saved")
	}
	if category.ID() != "cat-1234" {
		t.Errorf("ID() = %q, want %q", category.ID(), "cat-1234")
	}
	if category.Name().Value() != requisitionmocks.DefaultItemCategoryName {
		t.Errorf("Name().Value() = %q, want %q", category.Name().Value(), requisitionmocks.DefaultItemCategoryName)
	}
	if category.Status() != domain.CategoryStatusActive {
		t.Errorf("Status() = %q, want %q", category.Status(), domain.CategoryStatusActive)
	}
}

func Test_CreateItemCategoryUsecase_InvalidName(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	repo.SaveItemCategoryFunc = func(ctx context.Context, category *domain.ItemCategory) error {
		t.Fatal("SaveItemCategory() should not be called")
		return nil
	}

	uc := NewCreateItemCategoryUsecase(repo, requisitionmocks.NewMockIDGenerator())
	_, err := uc.Execute(context.Background(), categoryin.CreateItemCategoryCommand{Name: ""})

	if err != domain.ErrCategoryNameEmpty {
		t.Fatalf("Execute() error = %v, want %v", err, domain.ErrCategoryNameEmpty)
	}
}

func Test_CreateItemCategoryUsecase_IDGeneratorError(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	idGen := requisitionmocks.NewMockIDGenerator()
	wantErr := errors.New("id generation failed")
	idGen.GenerateIDFunc = func(prefix string) (string, error) {
		if prefix != domain.CategoryPrefixID {
			t.Errorf("GenerateID() prefix = %q, want %q", prefix, domain.CategoryPrefixID)
		}
		return "", wantErr
	}

	uc := NewCreateItemCategoryUsecase(repo, idGen)
	_, err := uc.Execute(context.Background(), categoryin.CreateItemCategoryCommand{
		Name: requisitionmocks.DefaultItemCategoryName,
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}

func Test_ListItemCategoriesUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	var gotLimit, gotOffset int

	repo.ListItemCategoriesFunc = func(ctx context.Context, limit int, offset int) ([]*domain.ItemCategory, error) {
		gotLimit = limit
		gotOffset = offset
		return []*domain.ItemCategory{requisitionmocks.ValidItemCategory()}, nil
	}

	uc := NewListItemCategoriesUsecase(repo)
	categories, err := uc.Execute(context.Background(), categoryin.ListItemCategoriesQuery{Limit: 25, Offset: 5})

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(categories) != 1 {
		t.Fatalf("Execute() returned %d categories, want 1", len(categories))
	}
	if gotLimit != 25 || gotOffset != 5 {
		t.Errorf("repository received limit=%d offset=%d, want limit=25 offset=5", gotLimit, gotOffset)
	}
}

func Test_GetItemCategoryUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	var gotID string
	repo.GetItemCategoryByIDFunc = func(ctx context.Context, id string) (*domain.ItemCategory, error) {
		gotID = id
		return requisitionmocks.ValidItemCategory(), nil
	}

	uc := NewGetItemCategoryUsecase(repo)
	category, err := uc.Execute(context.Background(), requisitionmocks.DefaultItemCategoryID)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if category.ID() != requisitionmocks.DefaultItemCategoryID {
		t.Errorf("ID() = %q, want %q", category.ID(), requisitionmocks.DefaultItemCategoryID)
	}
	if gotID != requisitionmocks.DefaultItemCategoryID {
		t.Errorf("repository received id=%q, want %q", gotID, requisitionmocks.DefaultItemCategoryID)
	}
}

func Test_GetItemCategoryUsecase_NotFound(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	wantErr := customerrors.NewNotFoundError("item category not found")
	repo.GetItemCategoryByIDFunc = func(ctx context.Context, id string) (*domain.ItemCategory, error) {
		return nil, wantErr
	}

	uc := NewGetItemCategoryUsecase(repo)
	_, err := uc.Execute(context.Background(), "cat-missing")

	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}

func Test_UpdateItemCategoryUsecase_PartialFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cmd      categoryin.UpdateItemCategoryCommand
		wantName string
		wantDesc string
	}{
		{
			name:     "name only",
			cmd:      categoryin.UpdateItemCategoryCommand{Name: strPtr("Hardware")},
			wantName: "Hardware",
			wantDesc: requisitionmocks.DefaultItemCategoryDescription,
		},
		{
			name:     "description only",
			cmd:      categoryin.UpdateItemCategoryCommand{Description: strPtr("Updated description")},
			wantName: requisitionmocks.DefaultItemCategoryName,
			wantDesc: "Updated description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := requisitionmocks.NewMockItemCategoryRepository()
			var updated *domain.ItemCategory
			repo.UpdateItemCategoryFunc = func(ctx context.Context, category *domain.ItemCategory) error {
				updated = category
				return nil
			}

			tt.cmd.ID = requisitionmocks.DefaultItemCategoryID
			uc := NewUpdateItemCategoryUsecase(repo)
			category, err := uc.Execute(context.Background(), tt.cmd)

			if err != nil {
				t.Fatalf("Execute() unexpected error = %v", err)
			}
			if updated == nil {
				t.Fatal("expected the category to be updated")
			}
			if category.Name().Value() != tt.wantName {
				t.Errorf("Name().Value() = %q, want %q", category.Name().Value(), tt.wantName)
			}
			if category.Description().Value() != tt.wantDesc {
				t.Errorf("Description().Value() = %q, want %q", category.Description().Value(), tt.wantDesc)
			}
		})
	}
}

func Test_UpdateItemCategoryUsecase_NoFields(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	repo.UpdateItemCategoryFunc = func(ctx context.Context, category *domain.ItemCategory) error {
		t.Fatal("UpdateItemCategory() should not be called")
		return nil
	}

	uc := NewUpdateItemCategoryUsecase(repo)
	_, err := uc.Execute(context.Background(), categoryin.UpdateItemCategoryCommand{
		ID: requisitionmocks.DefaultItemCategoryID,
	})

	if err != domain.ErrNoCategoryUpdateFields {
		t.Fatalf("Execute() error = %v, want %v", err, domain.ErrNoCategoryUpdateFields)
	}
}

func Test_UpdateItemCategoryUsecase_InvalidFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cmd     categoryin.UpdateItemCategoryCommand
		wantErr error
	}{
		{
			name:    "empty name",
			cmd:     categoryin.UpdateItemCategoryCommand{Name: strPtr("")},
			wantErr: domain.ErrCategoryNameEmpty,
		},
		{
			name:    "name too long",
			cmd:     categoryin.UpdateItemCategoryCommand{Name: strPtr(strings.Repeat("a", 101))},
			wantErr: domain.ErrCategoryNameTooLong,
		},
		{
			name:    "description too long",
			cmd:     categoryin.UpdateItemCategoryCommand{Description: strPtr(strings.Repeat("a", 201))},
			wantErr: domain.ErrCategoryDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := requisitionmocks.NewMockItemCategoryRepository()
			repo.UpdateItemCategoryFunc = func(ctx context.Context, category *domain.ItemCategory) error {
				t.Fatal("UpdateItemCategory() should not be called")
				return nil
			}

			tt.cmd.ID = requisitionmocks.DefaultItemCategoryID
			uc := NewUpdateItemCategoryUsecase(repo)
			_, err := uc.Execute(context.Background(), tt.cmd)

			if err != tt.wantErr {
				t.Fatalf("Execute() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func Test_UpdateItemCategoryUsecase_NotFound(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	wantErr := customerrors.NewNotFoundError("item category not found")
	repo.GetItemCategoryByIDFunc = func(ctx context.Context, id string) (*domain.ItemCategory, error) {
		return nil, wantErr
	}

	uc := NewUpdateItemCategoryUsecase(repo)
	_, err := uc.Execute(context.Background(), categoryin.UpdateItemCategoryCommand{
		ID:   "cat-missing",
		Name: strPtr("Hardware"),
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}

func Test_ActivateItemCategoryUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	category := requisitionmocks.ValidItemCategory()
	category.Deactivate()
	repo.GetItemCategoryByIDFunc = func(ctx context.Context, id string) (*domain.ItemCategory, error) {
		return category, nil
	}

	var updated *domain.ItemCategory
	repo.UpdateItemCategoryFunc = func(ctx context.Context, category *domain.ItemCategory) error {
		updated = category
		return nil
	}

	uc := NewActivateItemCategoryUsecase(repo)
	result, err := uc.Execute(context.Background(), requisitionmocks.DefaultItemCategoryID)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if updated == nil {
		t.Fatal("expected the category to be updated")
	}
	if result.Status() != domain.CategoryStatusActive {
		t.Errorf("Status() = %q, want %q", result.Status(), domain.CategoryStatusActive)
	}
}

func Test_DeactivateItemCategoryUsecase_Success(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	var updated *domain.ItemCategory
	repo.UpdateItemCategoryFunc = func(ctx context.Context, category *domain.ItemCategory) error {
		updated = category
		return nil
	}

	uc := NewDeactivateItemCategoryUsecase(repo)
	result, err := uc.Execute(context.Background(), requisitionmocks.DefaultItemCategoryID)

	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if updated == nil {
		t.Fatal("expected the category to be updated")
	}
	if result.Status() != domain.CategoryStatusInactive {
		t.Errorf("Status() = %q, want %q", result.Status(), domain.CategoryStatusInactive)
	}
}

func Test_ActivateItemCategoryUsecase_NotFound(t *testing.T) {
	t.Parallel()

	repo := requisitionmocks.NewMockItemCategoryRepository()
	wantErr := customerrors.NewNotFoundError("item category not found")
	repo.GetItemCategoryByIDFunc = func(ctx context.Context, id string) (*domain.ItemCategory, error) {
		return nil, wantErr
	}

	uc := NewActivateItemCategoryUsecase(repo)
	_, err := uc.Execute(context.Background(), "cat-missing")

	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}
