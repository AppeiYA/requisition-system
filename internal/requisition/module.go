package requisition

import (
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
)

type RequisitionModule struct {
	itemCategoryPortIn  *requisitionin.ItemCategoryPortIn
	itemCategoryHandler *requisitionhttp.ItemCategoryHandler
}

func NewRequisitionModule(
	itemCategoryPortIn *requisitionin.ItemCategoryPortIn,
	itemCategoryHandler *requisitionhttp.ItemCategoryHandler,
) *RequisitionModule {
	return &RequisitionModule{
		itemCategoryPortIn:  itemCategoryPortIn,
		itemCategoryHandler: itemCategoryHandler,
	}
}

func (m *RequisitionModule) ItemCategoryPortIn() *requisitionin.ItemCategoryPortIn {
	return m.itemCategoryPortIn
}

func (m *RequisitionModule) ItemCategoryHandler() *requisitionhttp.ItemCategoryHandler {
	return m.itemCategoryHandler
}
