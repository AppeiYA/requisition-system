package requisition

import (
	categoryhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/category"
	itemhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
)

type RequisitionModule struct {
	itemCategoryPortIn  *categoryin.ItemCategoryPortIn
	itemCategoryHandler *categoryhttp.ItemCategoryHandler
	itemPortIn          *itemin.ItemPortIn
	itemHandler         *itemhttp.ItemHandler
}

func NewRequisitionModule(
	itemCategoryPortIn *categoryin.ItemCategoryPortIn,
	itemCategoryHandler *categoryhttp.ItemCategoryHandler,
	itemPortIn *itemin.ItemPortIn,
	itemHandler *itemhttp.ItemHandler,
) *RequisitionModule {
	return &RequisitionModule{
		itemCategoryPortIn:  itemCategoryPortIn,
		itemCategoryHandler: itemCategoryHandler,
		itemPortIn:          itemPortIn,
		itemHandler:         itemHandler,
	}
}

func (m *RequisitionModule) ItemCategoryPortIn() *categoryin.ItemCategoryPortIn {
	return m.itemCategoryPortIn
}

func (m *RequisitionModule) ItemCategoryHandler() *categoryhttp.ItemCategoryHandler {
	return m.itemCategoryHandler
}

func (m *RequisitionModule) ItemPortIn() *itemin.ItemPortIn {
	return m.itemPortIn
}

func (m *RequisitionModule) ItemHandler() *itemhttp.ItemHandler {
	return m.itemHandler
}
