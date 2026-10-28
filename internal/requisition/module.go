package requisition

import (
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	categoryhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/category"
	itemhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http/item"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	categoryin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/category"
	itemin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in/item"
)

type RequisitionModule struct {
	itemCategoryPortIn  *categoryin.ItemCategoryPortIn
	itemCategoryHandler *categoryhttp.ItemCategoryHandler
	itemPortIn          *itemin.ItemPortIn
	itemHandler         *itemhttp.ItemHandler
	requisitionPortIn   *requisitionin.RequisitionPortIn
	requisitionHandler  *requisitionhttp.RequisitionHandler
}

func NewRequisitionModule(
	itemCategoryPortIn *categoryin.ItemCategoryPortIn,
	itemCategoryHandler *categoryhttp.ItemCategoryHandler,
	itemPortIn *itemin.ItemPortIn,
	itemHandler *itemhttp.ItemHandler,
	requisitionPortIn *requisitionin.RequisitionPortIn,
	requisitionHandler *requisitionhttp.RequisitionHandler,
) *RequisitionModule {
	return &RequisitionModule{
		itemCategoryPortIn:  itemCategoryPortIn,
		itemCategoryHandler: itemCategoryHandler,
		itemPortIn:          itemPortIn,
		itemHandler:         itemHandler,
		requisitionPortIn:   requisitionPortIn,
		requisitionHandler:  requisitionHandler,
	}
}

func (m *RequisitionModule) RequisitionHandler() *requisitionhttp.RequisitionHandler {
	return m.requisitionHandler
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
