package itemin

type ItemPortIn struct {
	CreateItemUsecase     CreateItemUsecaseInt
	GetItemUsecase        GetItemUsecaseInt
	ListItemsUsecase      ListItemsUsecaseInt
	UpdateItemUsecase     UpdateItemUsecaseInt
	ActivateItemUsecase   ActivateItemUsecaseInt
	DeactivateItemUsecase DeactivateItemUsecaseInt
}

func NewItemPortIn(
	createItemUsecase CreateItemUsecaseInt,
	getItemUsecase GetItemUsecaseInt,
	listItemsUsecase ListItemsUsecaseInt,
	updateItemUsecase UpdateItemUsecaseInt,
	activateItemUsecase ActivateItemUsecaseInt,
	deactivateItemUsecase DeactivateItemUsecaseInt,
) *ItemPortIn {
	return &ItemPortIn{
		CreateItemUsecase:     createItemUsecase,
		GetItemUsecase:        getItemUsecase,
		ListItemsUsecase:      listItemsUsecase,
		UpdateItemUsecase:     updateItemUsecase,
		ActivateItemUsecase:   activateItemUsecase,
		DeactivateItemUsecase: deactivateItemUsecase,
	}
}
