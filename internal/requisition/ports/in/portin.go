package requisitionin

type RequisitionPortIn struct {
	CreateRequisitionUsecase      CreateRequisitionUsecaseInt
	UpdateRequisitionUsecase      UpdateRequisitionUsecaseInt
	ListRequisitionsByUserUsecase ListRequisitionsByUserUsecaseInt
	ListRequisitionsUsecase       ListRequisitionsUsecaseInt
	GetRequisitionUsecase         GetRequisitionUsecaseInt
	TransitionRequisitionUsecase  TransitionRequisitionUsecaseInt
}

func NewRequisitionPortIn(
	createRequisitionUsecase CreateRequisitionUsecaseInt,
	updateRequisitionUsecase UpdateRequisitionUsecaseInt,
	listRequisitionsByUserUsecase ListRequisitionsByUserUsecaseInt,
	listRequisitionsUsecase ListRequisitionsUsecaseInt,
	getRequisitionUsecase GetRequisitionUsecaseInt,
	transitionRequisitionUsecase TransitionRequisitionUsecaseInt,
) *RequisitionPortIn {
	return &RequisitionPortIn{
		CreateRequisitionUsecase:      createRequisitionUsecase,
		UpdateRequisitionUsecase:      updateRequisitionUsecase,
		ListRequisitionsByUserUsecase: listRequisitionsByUserUsecase,
		ListRequisitionsUsecase:       listRequisitionsUsecase,
		GetRequisitionUsecase:         getRequisitionUsecase,
		TransitionRequisitionUsecase:  transitionRequisitionUsecase,
	}
}
