package categoryin

type ItemCategoryPortIn struct {
	CreateItemCategoryUsecase     CreateItemCategoryUsecaseInt
	GetItemCategoryUsecase        GetItemCategoryUsecaseInt
	ListItemCategoriesUsecase     ListItemCategoriesUsecaseInt
	UpdateItemCategoryUsecase     UpdateItemCategoryUsecaseInt
	ActivateItemCategoryUsecase   ActivateItemCategoryUsecaseInt
	DeactivateItemCategoryUsecase DeactivateItemCategoryUsecaseInt
	ListItemsInCategoryUsecase    ListItemsInCategoryUsecaseInt
}

func NewItemCategoryPortIn(
	createItemCategoryUsecase CreateItemCategoryUsecaseInt,
	getItemCategoryUsecase GetItemCategoryUsecaseInt,
	listItemCategoriesUsecase ListItemCategoriesUsecaseInt,
	updateItemCategoryUsecase UpdateItemCategoryUsecaseInt,
	activateItemCategoryUsecase ActivateItemCategoryUsecaseInt,
	deactivateItemCategoryUsecase DeactivateItemCategoryUsecaseInt,
	listItemsInCategoryUsecase ListItemsInCategoryUsecaseInt,
) *ItemCategoryPortIn {
	return &ItemCategoryPortIn{
		CreateItemCategoryUsecase:     createItemCategoryUsecase,
		GetItemCategoryUsecase:        getItemCategoryUsecase,
		ListItemCategoriesUsecase:     listItemCategoriesUsecase,
		UpdateItemCategoryUsecase:     updateItemCategoryUsecase,
		ActivateItemCategoryUsecase:   activateItemCategoryUsecase,
		DeactivateItemCategoryUsecase: deactivateItemCategoryUsecase,
		ListItemsInCategoryUsecase:    listItemsInCategoryUsecase,
	}
}
