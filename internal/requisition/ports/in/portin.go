package requisitionin

type RequisitionPortIn struct {
	CreateRequisitionUsecase      CreateRequisitionUsecaseInt
	UpdateRequisitionUsecase      UpdateRequisitionUsecaseInt
	ListRequisitionsByUserUsecase ListRequisitionsByUserUsecaseInt
	ListRequisitionsUsecase       ListRequisitionsUsecaseInt
	GetRequisitionUsecase         GetRequisitionUsecaseInt
	GetRequisitionItemUsecase     GetRequisitionItemUsecaseInt
	TransitionRequisitionUsecase  TransitionRequisitionUsecaseInt
}

func NewRequisitionPortIn(
	createRequisitionUsecase CreateRequisitionUsecaseInt,
	updateRequisitionUsecase UpdateRequisitionUsecaseInt,
	listRequisitionsByUserUsecase ListRequisitionsByUserUsecaseInt,
	listRequisitionsUsecase ListRequisitionsUsecaseInt,
	getRequisitionUsecase GetRequisitionUsecaseInt,
	getRequisitionItemUsecase GetRequisitionItemUsecaseInt,
	transitionRequisitionUsecase TransitionRequisitionUsecaseInt,
) *RequisitionPortIn {
	return &RequisitionPortIn{
		CreateRequisitionUsecase:      createRequisitionUsecase,
		UpdateRequisitionUsecase:      updateRequisitionUsecase,
		ListRequisitionsByUserUsecase: listRequisitionsByUserUsecase,
		ListRequisitionsUsecase:       listRequisitionsUsecase,
		GetRequisitionUsecase:         getRequisitionUsecase,
		GetRequisitionItemUsecase:     getRequisitionItemUsecase,
		TransitionRequisitionUsecase:  transitionRequisitionUsecase,
	}
}
