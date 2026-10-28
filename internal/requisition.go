package app

import (
	"github.com/AppeiYA/requisition-system/internal/requisition"
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	categoryhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/category"
	itemhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	requisitionpostgres "github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	requisitionusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase"
	categoryusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase/category"
	itemusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func (a *App) buildRequisition(
	baseRepo db.Repository,
	idGen requisitionout.IDGenerator,
	cache cache.Cache,
	membershipRepo requisitionout.MembershipLookup,
) *requisition.RequisitionModule {
	itemCategoryRepo := requisitionpostgres.NewItemCategoryRepository(baseRepo, cache)
	itemRepo := requisitionpostgres.NewItemRepository(baseRepo, cache)

	createItemCategoryUC := categoryusecase.NewCreateItemCategoryUsecase(itemCategoryRepo, idGen)
	getItemCategoryUC := categoryusecase.NewGetItemCategoryUsecase(itemCategoryRepo)
	listItemCategoriesUC := categoryusecase.NewListItemCategoriesUsecase(itemCategoryRepo)
	updateItemCategoryUC := categoryusecase.NewUpdateItemCategoryUsecase(itemCategoryRepo)
	activateItemCategoryUC := categoryusecase.NewActivateItemCategoryUsecase(itemCategoryRepo)
	deactivateItemCategoryUC := categoryusecase.NewDeactivateItemCategoryUsecase(itemCategoryRepo)
	listItemsInCategoryUC := categoryusecase.NewListItemsInCategoryUsecase(itemCategoryRepo)

	itemCategoryPortIn := categoryin.NewItemCategoryPortIn(
		createItemCategoryUC,
		getItemCategoryUC,
		listItemCategoriesUC,
		updateItemCategoryUC,
		activateItemCategoryUC,
		deactivateItemCategoryUC,
		listItemsInCategoryUC,
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

	requisitionRepo := requisitionpostgres.NewRequisitionRepository(baseRepo)
	createRequisitionUC := requisitionusecase.NewCreateRequisitionUsecase(
		requisitionRepo, itemCategoryRepo, itemRepo, membershipRepo, idGen,
	)
	listRequisitionsByUserUC := requisitionusecase.NewListRequisitionsByUserUsecase(requisitionRepo, nil)
	listRequisitionsUC := requisitionusecase.NewListRequisitionsUsecase(requisitionRepo)
	requisitionPortIn := requisitionin.NewRequisitionPortIn(createRequisitionUC, listRequisitionsByUserUC, listRequisitionsUC)
	requisitionHandler := requisitionhttp.NewRequisitionHandler(requisitionPortIn)

	return requisition.NewRequisitionModule(
		itemCategoryPortIn,
		itemCategoryHandler,
		itemPortIn,
		itemHandler,
		requisitionPortIn,
		requisitionHandler,
	)
}
