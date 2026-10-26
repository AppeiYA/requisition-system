package app

import (
	"github.com/AppeiYA/requisition-system/internal/requisition"
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	requisitionpostgres "github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	requisitionusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

func (a *App) buildRequisition(
	baseRepo db.Repository,
	idGen requisitionout.IDGenerator,
	cache cache.Cache,
) *requisition.RequisitionModule {
	itemCategoryRepo := requisitionpostgres.NewItemCategoryRepository(baseRepo, cache)

	createItemCategoryUC := requisitionusecase.NewCreateItemCategoryUsecase(itemCategoryRepo, idGen)
	getItemCategoryUC := requisitionusecase.NewGetItemCategoryUsecase(itemCategoryRepo)
	listItemCategoriesUC := requisitionusecase.NewListItemCategoriesUsecase(itemCategoryRepo)
	updateItemCategoryUC := requisitionusecase.NewUpdateItemCategoryUsecase(itemCategoryRepo)
	activateItemCategoryUC := requisitionusecase.NewActivateItemCategoryUsecase(itemCategoryRepo)
	deactivateItemCategoryUC := requisitionusecase.NewDeactivateItemCategoryUsecase(itemCategoryRepo)

	itemCategoryPortIn := requisitionin.NewItemCategoryPortIn(
		createItemCategoryUC,
		getItemCategoryUC,
		listItemCategoriesUC,
		updateItemCategoryUC,
		activateItemCategoryUC,
		deactivateItemCategoryUC,
	)

	itemCategoryHandler := requisitionhttp.NewItemCategoryHandler(itemCategoryPortIn)

	return requisition.NewRequisitionModule(itemCategoryPortIn, itemCategoryHandler)
}
