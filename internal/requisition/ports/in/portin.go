package requisitionin

type RequisitionPortIn struct {
	CreateRequisitionUsecase      CreateRequisitionUsecaseInt
	ListRequisitionsByUserUsecase ListRequisitionsByUserUsecaseInt
	ListRequisitionsUsecase       ListRequisitionsUsecaseInt
}

func NewRequisitionPortIn(
	createRequisitionUsecase CreateRequisitionUsecaseInt,
	listRequisitionsByUserUsecase ListRequisitionsByUserUsecaseInt,
	listRequisitionsUsecase ListRequisitionsUsecaseInt,
) *RequisitionPortIn {
	return &RequisitionPortIn{
		CreateRequisitionUsecase:      createRequisitionUsecase,
		ListRequisitionsByUserUsecase: listRequisitionsByUserUsecase,
		ListRequisitionsUsecase:       listRequisitionsUsecase,
	}
}
