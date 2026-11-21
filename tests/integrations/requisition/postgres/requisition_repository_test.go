package postgres_test

import (
	"context"
	"testing"
	"time"

	requisitionpostgres "github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

const (
	repositoryUserID       = "usr-repository-user"
	repositoryAdminID      = "usr-repository-admin"
	repositoryDepartmentID = "dpt-repository"
	repositoryCategoryID   = "cat-repository"
	repositoryItemID       = "itm-repository"
)

type requisitionRepositoryFixture struct {
	Suite                   *testutil.Suite
	TransactionManager      *appdb.TransactionManager
	RequisitionRepository   *requisitionpostgres.RequisitionRepository
	RequestedItemRepository *requisitionpostgres.RequestedItemRepository
	ItemCategoryRepository  *requisitionpostgres.ItemCategoryRepository
	ItemRepository          *requisitionpostgres.ItemRepository
}

func newRequisitionRepositoryFixture(t *testing.T) *requisitionRepositoryFixture {
	t.Helper()

	suite := testutil.NewSuite(t, "../../../../migrations")
	redisResource := suite.SetupRedis(t)
	require.NoError(t, suite.Reset(context.Background()))

	ctx := context.Background()
	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO users (id, first_name, last_name, email, password_hash, status, role)
		VALUES
			('`+repositoryUserID+`', 'Repository', 'User', 'repository.user@example.com',
			 'integration-password-hash', 'active', 'user'),
			('`+repositoryAdminID+`', 'Repository', 'Admin', 'repository.admin@example.com',
			 'integration-password-hash', 'active', 'admin');
		INSERT INTO departments (id, name, description, code, status)
		VALUES ('`+repositoryDepartmentID+`', 'Repository', 'Repository department', 'REP', 'active');
	`)
	require.NoError(t, err)

	baseRepository := appdb.NewRepository(suite.AppDB)
	cache := redisinfra.NewCache(redisResource.Client)

	categoryRepository := requisitionpostgres.NewItemCategoryRepository(baseRepository, cache)
	categoryName, err := requisitiondomain.NewCategoryName("Furniture")
	require.NoError(t, err)
	categoryDescription, err := requisitiondomain.NewCategoryDescription("Office furniture")
	require.NoError(t, err)
	category, err := requisitiondomain.NewItemCategory(
		repositoryCategoryID,
		categoryName,
		categoryDescription,
	)
	require.NoError(t, err)
	require.NoError(t, categoryRepository.SaveItemCategory(ctx, category))

	_, err = suite.DB.Pool.Exec(ctx, `
		INSERT INTO items (id, name, description, category_id, unit_cost, status)
		VALUES ('`+repositoryItemID+`', 'Laptop', '14 inch laptop', '`+repositoryCategoryID+`', 500000, 'active');
	`)
	require.NoError(t, err)

	return &requisitionRepositoryFixture{
		Suite:                   suite,
		TransactionManager:      appdb.NewTransactionManager(suite.AppDB),
		RequisitionRepository:   requisitionpostgres.NewRequisitionRepository(baseRepository),
		RequestedItemRepository: requisitionpostgres.NewRequestedItemRepository(baseRepository),
		ItemCategoryRepository:  categoryRepository,
		ItemRepository:          requisitionpostgres.NewItemRepository(baseRepository, cache),
	}
}

func (f *requisitionRepositoryFixture) newPendingAggregate(
	t *testing.T,
	suffix string,
) *requisitiondomain.Requisition {
	t.Helper()

	requisitionID := "req-" + suffix
	catalogueName, err := requisitiondomain.NewItemName("Laptop")
	require.NoError(t, err)

	requestedName, err := requisitiondomain.NewItemName("Standing desk")
	require.NoError(t, err)
	requestedDescription, err := requisitiondomain.NewItemDescription("Adjustable standing desk")
	require.NoError(t, err)
	requestedItem, err := requisitiondomain.NewRequestedItem(
		"ritem-"+suffix,
		requestedName,
		repositoryCategoryID,
		requestedDescription,
		repositoryUserID,
	)
	require.NoError(t, err)

	catalogueQuantity, err := requisitiondomain.NewRequisitionQuantity(2)
	require.NoError(t, err)
	catalogueAmount, err := shareddomain.NewNairaMoney(5000)
	require.NoError(t, err)
	catalogueLine, err := requisitiondomain.NewRequisitionItem(
		"reqitem-"+suffix+"-1",
		requisitionID,
		1,
		repositoryItemID,
		"",
		catalogueName,
		catalogueQuantity,
		catalogueAmount,
		"Catalogue line",
	)
	require.NoError(t, err)

	requestedQuantity, err := requisitiondomain.NewRequisitionQuantity(3)
	require.NoError(t, err)
	requestedAmount, err := shareddomain.NewNairaMoney(2500)
	require.NoError(t, err)
	requestedLine, err := requisitiondomain.NewRequisitionItem(
		"reqitem-"+suffix+"-2",
		requisitionID,
		2,
		"",
		requestedItem.ID(),
		requestedName,
		requestedQuantity,
		requestedAmount,
		"Requested line",
	)
	require.NoError(t, err)

	priority, err := requisitiondomain.NewRequisitionPriority("medium")
	require.NoError(t, err)
	justification, err := requisitiondomain.NewRequisitionJustification("Integration test requisition")
	require.NoError(t, err)

	requisition, err := requisitiondomain.NewSubmittedRequisitionWithRequestedItems(
		requisitionID,
		"REQ-2026-"+suffix,
		repositoryUserID,
		repositoryDepartmentID,
		priority,
		time.Now().UTC().AddDate(0, 0, 7),
		justification,
		[]*requisitiondomain.RequisitionItem{catalogueLine, requestedLine},
		[]*requisitiondomain.RequestedItem{requestedItem},
	)
	require.NoError(t, err)

	history, err := requisition.InitialStatusHistory(
		"rsh-"+suffix,
		repositoryUserID,
		requisitiondomain.ActorRoleStaff,
		time.Now().UTC(),
	)
	require.NoError(t, err)
	require.NoError(t, requisition.RecordInitialStatusHistory(history))

	return requisition
}

func (f *requisitionRepositoryFixture) saveAggregate(
	t *testing.T,
	requisition *requisitiondomain.Requisition,
) {
	t.Helper()

	require.NoError(t, f.TransactionManager.WithinTransaction(
		context.Background(),
		func(txCtx context.Context) error {
			return f.RequisitionRepository.SaveRequisition(txCtx, requisition)
		},
	))
}

func TestRequisitionRepositoryAggregateLifecycle(t *testing.T) {
	fixture := newRequisitionRepositoryFixture(t)
	ctx := context.Background()

	t.Run("saves and loads the aggregate", func(t *testing.T) {
		requisition := fixture.newPendingAggregate(t, "aggregate")
		fixture.saveAggregate(t, requisition)

		found, err := fixture.RequisitionRepository.GetRequisitionByID(ctx, requisition.ID())
		require.NoError(t, err)
		require.Equal(t, requisition.ID(), found.ID())
		require.Equal(t, "REQ-2026-aggregate", found.ReferenceCode())
		require.Equal(t, repositoryUserID, found.RequestedBy())
		require.Equal(t, repositoryDepartmentID, found.DepartmentID())
		require.Equal(t, requisitiondomain.RequisitionStatusPending, found.Status())
		require.Equal(t, requisitiondomain.PriorityMedium, found.Priority())
		require.Equal(t, "Integration test requisition", found.Justification().Value())
		require.Equal(t, requisition.NeededBy().Value().UTC().Format("2006-01-02"),
			found.NeededBy().Value().UTC().Format("2006-01-02"))

		items := found.Items()
		require.Len(t, items, 2)
		require.Equal(t, "reqitem-aggregate-1", items[0].ID())
		require.Equal(t, repositoryItemID, items[0].ItemID())
		require.Empty(t, items[0].RequestedItemID())
		require.Equal(t, "Laptop", items[0].ItemNameSnapshot().Value())
		require.Equal(t, int64(2), items[0].Quantity().Value())
		require.Equal(t, "reqitem-aggregate-2", items[1].ID())
		require.Empty(t, items[1].ItemID())
		require.Equal(t, "ritem-aggregate", items[1].RequestedItemID())
		require.Equal(t, "Standing desk", items[1].ItemNameSnapshot().Value())

		requestedItems := found.RequestedItems()
		require.Len(t, requestedItems, 1)
		require.Equal(t, "ritem-aggregate", requestedItems[0].ID())
		require.Equal(t, "Standing desk", requestedItems[0].Name().Value())
		require.Equal(t, requisitiondomain.RequestedItemUnreviewed, requestedItems[0].ReviewStatus())

		history := found.StatusHistory()
		require.Len(t, history, 1)
		require.Equal(t, 1, history[0].Sequence())
		require.Nil(t, history[0].FromStatus())
		require.Equal(t, requisitiondomain.RequisitionStatusPending, history[0].ToStatus())
		require.Equal(t, repositoryUserID, history[0].ActedBy())
		require.Equal(t, requisitiondomain.ActorRoleStaff, history[0].ActorRole())

		total, err := found.TotalEstimatedAmount()
		require.NoError(t, err)
		require.Equal(t, 17500.0, total.ToFloat64())

		summaries, err := fixture.RequisitionRepository.ListRequisitionSummaries(ctx, 10, 0)
		require.NoError(t, err)
		require.Len(t, summaries, 1)
		require.Equal(t, requisition.ID(), summaries[0].ID)
		require.Equal(t, 17500.0, summaries[0].TotalEstimatedAmount)

		byUser, err := fixture.RequisitionRepository.ListRequisitionSummariesByUserID(
			ctx,
			repositoryUserID,
			10,
			0,
		)
		require.NoError(t, err)
		require.Len(t, byUser, 1)
		require.Equal(t, requisition.ID(), byUser[0].ID)
	})

	t.Run("updates the aggregate while preserving line identity", func(t *testing.T) {
		requisition := fixture.newPendingAggregate(t, "update")
		fixture.saveAggregate(t, requisition)

		require.NoError(t, fixture.TransactionManager.WithinTransaction(
			ctx,
			func(txCtx context.Context) error {
				loaded, err := fixture.RequisitionRepository.GetRequisitionByIDForUpdate(
					txCtx,
					requisition.ID(),
				)
				if err != nil {
					return err
				}

				items := loaded.Items()
				updatedQuantity, err := requisitiondomain.NewRequisitionQuantity(5)
				if err != nil {
					return err
				}
				updatedLine, err := requisitiondomain.ReconstituteRequisitionItem(
					items[0].ID(),
					items[0].RequisitionID(),
					items[0].LineNumber(),
					items[0].ItemID(),
					items[0].RequestedItemID(),
					items[0].ItemNameSnapshot(),
					updatedQuantity,
					items[0].EstimatedUnitAmount(),
					"Updated notes",
					items[0].CreatedAt(),
				)
				if err != nil {
					return err
				}

				justification, err := requisitiondomain.NewRequisitionJustification(
					"Updated justification",
				)
				if err != nil {
					return err
				}
				priority, err := requisitiondomain.NewRequisitionPriority("high")
				if err != nil {
					return err
				}
				if err := loaded.UpdateRequisition(&requisitiondomain.UpdateRequisitionInput{
					Justification: justification,
					Priority:      &priority,
				}); err != nil {
					return err
				}
				if err := loaded.ReplaceItems(
					[]*requisitiondomain.RequisitionItem{updatedLine, items[1]},
					loaded.RequestedItems(),
				); err != nil {
					return err
				}

				return fixture.RequisitionRepository.UpdateRequisitionAggregate(txCtx, loaded)
			},
		))

		found, err := fixture.RequisitionRepository.GetRequisitionByID(ctx, requisition.ID())
		require.NoError(t, err)
		require.Equal(t, requisitiondomain.PriorityHigh, found.Priority())
		require.Equal(t, "Updated justification", found.Justification().Value())

		items := found.Items()
		require.Len(t, items, 2)
		require.Equal(t, "reqitem-update-1", items[0].ID())
		require.Equal(t, int64(5), items[0].Quantity().Value())
		require.Equal(t, "Updated notes", items[0].Notes())
		require.Len(t, found.RequestedItems(), 1)
		require.Equal(t, "ritem-update", found.RequestedItems()[0].ID())

		total, err := found.TotalEstimatedAmount()
		require.NoError(t, err)
		require.Equal(t, 32500.0, total.ToFloat64())
	})

	t.Run("appends status history on transition", func(t *testing.T) {
		requisition := fixture.newPendingAggregate(t, "transition")
		fixture.saveAggregate(t, requisition)

		require.NoError(t, fixture.TransactionManager.WithinTransaction(
			ctx,
			func(txCtx context.Context) error {
				loaded, err := fixture.RequisitionRepository.GetRequisitionByIDForUpdate(
					txCtx,
					requisition.ID(),
				)
				if err != nil {
					return err
				}
				history, err := loaded.ApplyStatusTransition(
					"rsh-transition-2",
					requisitiondomain.RequisitionStatusApproved,
					repositoryAdminID,
					requisitiondomain.ActorRoleAdmin,
					"Approved by the repository test",
					time.Now().UTC(),
				)
				if err != nil {
					return err
				}
				return fixture.RequisitionRepository.UpdateRequisitionStatus(
					txCtx,
					loaded,
					history,
				)
			},
		))

		found, err := fixture.RequisitionRepository.GetRequisitionByID(ctx, requisition.ID())
		require.NoError(t, err)
		require.Equal(t, requisitiondomain.RequisitionStatusApproved, found.Status())

		history := found.StatusHistory()
		require.Len(t, history, 2)
		require.Equal(t, 2, history[1].Sequence())
		require.NotNil(t, history[1].FromStatus())
		require.Equal(t, requisitiondomain.RequisitionStatusPending, *history[1].FromStatus())
		require.Equal(t, requisitiondomain.RequisitionStatusApproved, history[1].ToStatus())
		require.Equal(t, repositoryAdminID, history[1].ActedBy())
		require.Equal(t, requisitiondomain.ActorRoleAdmin, history[1].ActorRole())
		require.Equal(t, "Approved by the repository test", history[1].Reason())

		_, err = fixture.Suite.DB.Pool.Exec(ctx, `
			UPDATE requisition_status_history
			SET reason = 'tampered'
			WHERE id = 'rsh-transition-2'
		`)
		require.Error(t, err)
	})
}

func TestRequestedItemRepositoryLifecycle(t *testing.T) {
	fixture := newRequisitionRepositoryFixture(t)
	ctx := context.Background()

	name, err := requisitiondomain.NewItemName("Ergonomic chair")
	require.NoError(t, err)
	description, err := requisitiondomain.NewItemDescription("Mesh back chair")
	require.NoError(t, err)
	item, err := requisitiondomain.NewRequestedItem(
		"ritem-requested-lifecycle",
		name,
		repositoryCategoryID,
		description,
		repositoryUserID,
	)
	require.NoError(t, err)

	require.NoError(t, fixture.RequestedItemRepository.SaveRequestedItem(ctx, item))

	found, err := fixture.RequestedItemRepository.GetRequestedItemByID(ctx, item.ID())
	require.NoError(t, err)
	require.Equal(t, item.ID(), found.ID())
	require.Equal(t, "Ergonomic chair", found.Name().Value())
	require.Equal(t, repositoryCategoryID, found.CategoryID())
	require.Equal(t, requisitiondomain.RequestedItemUnreviewed, found.ReviewStatus())

	unreviewed, err := fixture.RequestedItemRepository.ListUnreviewedRequestedItems(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, unreviewed, 1)
	require.Equal(t, item.ID(), unreviewed[0].ID())

	require.NoError(t, item.MarkAsDismissed(repositoryAdminID, time.Now().UTC()))
	require.NoError(t, fixture.RequestedItemRepository.UpdateRequestedItemStatus(ctx, item))

	found, err = fixture.RequestedItemRepository.GetRequestedItemByID(ctx, item.ID())
	require.NoError(t, err)
	require.Equal(t, requisitiondomain.RequestedItemDismissed, found.ReviewStatus())
	require.NotNil(t, found.ReviewedBy())
	require.Equal(t, repositoryAdminID, *found.ReviewedBy())
	require.NotNil(t, found.ReviewedAt())

	unreviewed, err = fixture.RequestedItemRepository.ListUnreviewedRequestedItems(ctx, 10, 0)
	require.NoError(t, err)
	require.Empty(t, unreviewed)

	require.Error(t, fixture.RequestedItemRepository.UpdateRequestedItemStatus(ctx, item))
}

func TestCatalogRepositories(t *testing.T) {
	fixture := newRequisitionRepositoryFixture(t)
	ctx := context.Background()

	t.Run("item category repository", func(t *testing.T) {
		category, err := fixture.ItemCategoryRepository.GetItemCategoryByID(ctx, repositoryCategoryID)
		require.NoError(t, err)
		require.Equal(t, "Furniture", category.Name().Value())
		require.Equal(t, requisitiondomain.CategoryStatusActive, category.Status())

		categories, err := fixture.ItemCategoryRepository.ListItemCategories(ctx, 10, 0)
		require.NoError(t, err)
		require.Len(t, categories, 1)
		require.Equal(t, repositoryCategoryID, categories[0].ID())

		items, err := fixture.ItemCategoryRepository.ListItemsInCategory(ctx, repositoryCategoryID)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, repositoryItemID, items[0].ID())

		updatedName := "Office Furniture"
		updatedDescription := "Updated office furniture"
		require.NoError(t, category.Update(&updatedName, &updatedDescription))
		require.NoError(t, fixture.ItemCategoryRepository.UpdateItemCategory(ctx, category))

		category, err = fixture.ItemCategoryRepository.GetItemCategoryByID(ctx, repositoryCategoryID)
		require.NoError(t, err)
		require.Equal(t, updatedName, category.Name().Value())
		require.Equal(t, updatedDescription, category.Description().Value())

		category.Deactivate()
		require.NoError(t, fixture.ItemCategoryRepository.UpdateItemCategory(ctx, category))

		category, err = fixture.ItemCategoryRepository.GetItemCategoryByID(ctx, repositoryCategoryID)
		require.NoError(t, err)
		require.Equal(t, requisitiondomain.CategoryStatusInactive, category.Status())
	})

	t.Run("item repository", func(t *testing.T) {
		item, err := fixture.ItemRepository.GetItemByID(ctx, repositoryItemID)
		require.NoError(t, err)
		require.Equal(t, "Laptop", item.Name().Value())
		require.Equal(t, requisitiondomain.ItemStatusActive, item.Status())

		items, err := fixture.ItemRepository.ListItems(ctx, 10, 0)
		require.NoError(t, err)
		require.Len(t, items, 1)
		require.Equal(t, repositoryItemID, items[0].ID())

		updatedName := "Business laptop"
		updatedDescription := "15 inch business laptop"
		updatedUnitCost := 7500.0
		require.NoError(t, item.Update(&updatedName, &updatedDescription, nil, &updatedUnitCost))
		require.NoError(t, fixture.ItemRepository.UpdateItem(ctx, item))

		item, err = fixture.ItemRepository.GetItemByID(ctx, repositoryItemID)
		require.NoError(t, err)
		require.Equal(t, updatedName, item.Name().Value())
		require.Equal(t, updatedDescription, item.Description().Value())
		require.Equal(t, 7500.0, item.UnitCost().ToFloat64())

		require.NoError(t, item.Deactivate())
		require.NoError(t, fixture.ItemRepository.UpdateItem(ctx, item))

		item, err = fixture.ItemRepository.GetItemByID(ctx, repositoryItemID)
		require.NoError(t, err)
		require.Equal(t, requisitiondomain.ItemStatusInactive, item.Status())
	})
}
