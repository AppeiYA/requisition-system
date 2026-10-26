package app

import (
	"github.com/AppeiYA/requisition-system/internal/requisition"
	categoryhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/category"
	itemhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	requisitionpostgres "github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	categoryusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase/category"
	itemusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func (a *App) buildRequisition(
	baseRepo db.Repository,
	idGen requisitionout.IDGenerator,
	cache cache.Cache,
) *requisition.RequisitionModule {
	itemCategoryRepo := requisitionpostgres.NewItemCategoryRepository(baseRepo, cache)
	itemRepo := requisitionpostgres.NewItemRepository(baseRepo, cache)

	createItemCategoryUC := categoryusecase.NewCreateItemCategoryUsecase(itemCategoryRepo, idGen)
	getItemCategoryUC := categoryusecase.NewGetItemCategoryUsecase(itemCategoryRepo)
	listItemCategoriesUC := categoryusecase.NewListItemCategoriesUsecase(itemCategoryRepo)
	updateItemCategoryUC := categoryusecase.NewUpdateItemCategoryUsecase(itemCategoryRepo)
	activateItemCategoryUC := categoryusecase.NewActivateItemCategoryUsecase(itemCategoryRepo)
	deactivateItemCategoryUC := categoryusecase.NewDeactivateItemCategoryUsecase(itemCategoryRepo)

	itemCategoryPortIn := categoryin.NewItemCategoryPortIn(
		createItemCategoryUC,
		getItemCategoryUC,
		listItemCategoriesUC,
		updateItemCategoryUC,
		activateItemCategoryUC,
		deactivateItemCategoryUC,
	)

	itemCategoryHandler := categoryhttp.NewItemCategoryHandler(itemCategoryPortIn)

	createItemUC := itemusecase.NewCreateItemUsecase(itemRepo, idGen)
	getItemUC := itemusecase.NewGetItemUsecase(itemRepo)
	listItemsUC := itemusecase.NewListItemsUsecase(itemRepo)
	updateItemUC := itemusecase.NewUpdateItemUsecase(itemRepo)
	activateItemUC := itemusecase.NewActivateItemUsecase(itemRepo)
	deactivateItemUC := itemusecase.NewDeactivateItemUsecase(itemRepo)

	itemPortIn := itemin.NewItemPortIn(
		createItemUC,
		getItemUC,
		listItemsUC,
		updateItemUC,
		activateItemUC,
		deactivateItemUC,
	)

	itemHandler := itemhttp.NewItemHandler(itemPortIn)

	return requisition.NewRequisitionModule(
		itemCategoryPortIn,
		itemCategoryHandler,
		itemPortIn,
		itemHandler,
	)
}
