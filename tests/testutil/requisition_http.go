package testutil

import (
	"context"
	"testing"

	departmentpostgres "github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres"
	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	categoryhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/category"
	itemhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	requisitionexternal "github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/external"
	requisitionpostgres "github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
	requisitionusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase"
	categoryusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase/category"
	itemusecase "github.com/AppeiYA/requisition-system/internal/requisition/usecase/item"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/generator"
	redisinfra "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/redis"
	"github.com/gin-gonic/gin"
)

type RequisitionHTTP struct {
	*DepartmentHTTP
	Router                  *gin.Engine
	MembershipRepository    *departmentpostgres.MembershipRepository
	RequisitionRepository   *requisitionpostgres.RequisitionRepository
	ItemRepository          *requisitionpostgres.ItemRepository
	ItemCategoryRepository  *requisitionpostgres.ItemCategoryRepository
	RequestedItemRepository *requisitionpostgres.RequestedItemRepository
}

func (s *Suite) SetupRequisitionHTTP(t *testing.T) *RequisitionHTTP {
	t.Helper()

	department := s.SetupDepartmentHTTP(t)
	baseRepository := db.NewRepository(s.AppDB)
	cache := redisinfra.NewCache(s.Redis.Client)
	idGenerator := &generator.IDGenerator{}
	transactionManager := db.NewTransactionManager(s.AppDB)

	membershipRepository := departmentpostgres.NewMembershipRepository(baseRepository)
	requisitionRepository := requisitionpostgres.NewRequisitionRepository(baseRepository)
	itemCategoryRepository := requisitionpostgres.NewItemCategoryRepository(baseRepository, cache)
	itemRepository := requisitionpostgres.NewItemRepository(baseRepository, cache)
	requestedItemRepository := requisitionpostgres.NewRequestedItemRepository(baseRepository)

	userLookup := requisitionexternal.NewIdentityAdapter(department.UserRepository)

	createItemCategory := categoryusecase.NewCreateItemCategoryUsecase(itemCategoryRepository, idGenerator)
	getItemCategory := categoryusecase.NewGetItemCategoryUsecase(itemCategoryRepository)
	listItemCategories := categoryusecase.NewListItemCategoriesUsecase(itemCategoryRepository)
	updateItemCategory := categoryusecase.NewUpdateItemCategoryUsecase(itemCategoryRepository)
	activateItemCategory := categoryusecase.NewActivateItemCategoryUsecase(itemCategoryRepository)
	deactivateItemCategory := categoryusecase.NewDeactivateItemCategoryUsecase(itemCategoryRepository)
	listItemsInCategory := categoryusecase.NewListItemsInCategoryUsecase(itemCategoryRepository)
	itemCategoryPortIn := categoryin.NewItemCategoryPortIn(
		createItemCategory,
		getItemCategory,
		listItemCategories,
		updateItemCategory,
		activateItemCategory,
		deactivateItemCategory,
		listItemsInCategory,
	)

	createItem := itemusecase.NewCreateItemUsecase(itemRepository, idGenerator)
	getItem := itemusecase.NewGetItemUsecase(itemRepository)
	listItems := itemusecase.NewListItemsUsecase(itemRepository)
	listUnreviewedItems := itemusecase.NewListUnreviewedItemsUsecase(requestedItemRepository)
	promoteRequestedItem := itemusecase.NewPromoteRequestedItemUsecase(
		requestedItemRepository,
		itemRepository,
		userLookup,
		transactionManager,
		idGenerator,
	)
	dismissRequestedItem := itemusecase.NewDismissRequestedItemUsecase(
		requestedItemRepository,
		userLookup,
		transactionManager,
	)
	updateItem := itemusecase.NewUpdateItemUsecase(itemRepository)
	activateItem := itemusecase.NewActivateItemUsecase(itemRepository)
	deactivateItem := itemusecase.NewDeactivateItemUsecase(itemRepository)
	itemPortIn := itemin.NewItemPortIn(
		createItem,
		getItem,
		listItems,
		listUnreviewedItems,
		promoteRequestedItem,
		dismissRequestedItem,
		updateItem,
		activateItem,
		deactivateItem,
	)

	createRequisition := requisitionusecase.NewCreateRequisitionUsecase(
		requisitionRepository,
		itemCategoryRepository,
		itemRepository,
		membershipRepository,
		idGenerator,
		transactionManager,
	)
	updateRequisition := requisitionusecase.NewUpdateRequisitionUsecase(
		requisitionRepository,
		itemCategoryRepository,
		itemRepository,
		membershipRepository,
		idGenerator,
		transactionManager,
	)
	listRequisitionsByUser := requisitionusecase.NewListRequisitionsByUserUsecase(requisitionRepository, nil)
	listRequisitions := requisitionusecase.NewListRequisitionsUsecase(requisitionRepository)
	getRequisition := requisitionusecase.NewGetRequisitionUsecase(requisitionRepository)
	transitionRequisition := requisitionusecase.NewTransitionRequisitionUsecase(
		requisitionRepository,
		idGenerator,
		transactionManager,
	)
	requisitionPortIn := requisitionin.NewRequisitionPortIn(
		createRequisition,
		updateRequisition,
		listRequisitionsByUser,
		listRequisitions,
		getRequisition,
		transitionRequisition,
	)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := identityhttp.NewAuthMiddleware(
		department.AccessTokens,
		department.UserRepository,
	)
	api := router.Group("/api/v1")
	categoryhttp.NewItemCategoryHandler(itemCategoryPortIn).Register(
		api,
		authMiddleware.Handle(),
		authMiddleware.IsAdmin(),
	)
	itemhttp.NewItemHandler(itemPortIn).Register(
		api,
		authMiddleware.Handle(),
		authMiddleware.IsAdmin(),
	)
	requisitionhttp.NewRequisitionHandler(requisitionPortIn).Register(
		api,
		authMiddleware.Handle(),
		authMiddleware.IsAdmin(),
	)

	return &RequisitionHTTP{
		DepartmentHTTP:          department,
		Router:                  router,
		MembershipRepository:    membershipRepository,
		RequisitionRepository:   requisitionRepository,
		ItemRepository:          itemRepository,
		ItemCategoryRepository:  itemCategoryRepository,
		RequestedItemRepository: requestedItemRepository,
	}
}

// AddMembership creates an active department membership for the user so
// requisition use cases can resolve the user's department.
func (r *RequisitionHTTP) AddMembership(
	t *testing.T,
	userID,
	departmentID,
	jobTitle string,
) *departmentdomain.DepartmentMembership {
	t.Helper()

	title, err := departmentdomain.NewMemberJobTitle(jobTitle)
	if err != nil {
		t.Fatalf("create test job title: %v", err)
	}
	membership, err := departmentdomain.NewDepartmentMembership(userID, departmentID, title)
	if err != nil {
		t.Fatalf("create test membership: %v", err)
	}
	if err := r.MembershipRepository.SaveMembership(context.Background(), membership); err != nil {
		t.Fatalf("save test membership: %v", err)
	}
	return membership
}
