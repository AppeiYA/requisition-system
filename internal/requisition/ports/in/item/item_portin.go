package itemin

type ItemPortIn struct {
	CreateItemUsecase           CreateItemUsecaseInt
	GetItemUsecase              GetItemUsecaseInt
	ListItemsUsecase            ListItemsUsecaseInt
	ListUnreviewedItemsUsecase  ListUnreviewedItemsUsecaseInt
	PromoteRequestedItemUsecase PromoteRequestedItemUsecaseInt
	DismissRequestedItemUsecase DismissRequestedItemUsecaseInt
	UpdateItemUsecase           UpdateItemUsecaseInt
	ActivateItemUsecase         ActivateItemUsecaseInt
	DeactivateItemUsecase       DeactivateItemUsecaseInt
}

func NewItemPortIn(
	createItemUsecase CreateItemUsecaseInt,
	getItemUsecase GetItemUsecaseInt,
	listItemsUsecase ListItemsUsecaseInt,
	listUnreviewedItemsUsecase ListUnreviewedItemsUsecaseInt,
	promoteRequestedItemUsecase PromoteRequestedItemUsecaseInt,
	dismissRequestedItemUsecase DismissRequestedItemUsecaseInt,
	updateItemUsecase UpdateItemUsecaseInt,
	activateItemUsecase ActivateItemUsecaseInt,
	deactivateItemUsecase DeactivateItemUsecaseInt,
) *ItemPortIn {
	return &ItemPortIn{
		CreateItemUsecase:           createItemUsecase,
		GetItemUsecase:              getItemUsecase,
		ListItemsUsecase:            listItemsUsecase,
		ListUnreviewedItemsUsecase:  listUnreviewedItemsUsecase,
		PromoteRequestedItemUsecase: promoteRequestedItemUsecase,
		DismissRequestedItemUsecase: dismissRequestedItemUsecase,
		UpdateItemUsecase:           updateItemUsecase,
		ActivateItemUsecase:         activateItemUsecase,
		DeactivateItemUsecase:       deactivateItemUsecase,
	}
}
