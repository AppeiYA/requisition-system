package invitationin

type InvitationPortIn struct {
	SaveInvitationUsecase SaveInvitationUsecaseInt
	ValidateInvitationUsecase ValidateInvitationUseCaseInt
	AcceptInvitationUsecase AcceptInvitationUseCaseInt
	GetInvitationUsecase GetInvitationUsecaseInt
}

func NewInvitationPortIn(
	saveInvitationUsecase SaveInvitationUsecaseInt,
	validateInvitationUsecase ValidateInvitationUseCaseInt,
	acceptInvitationUsecase AcceptInvitationUseCaseInt,
	getInvitationUsecase GetInvitationUsecaseInt,
) *InvitationPortIn {
	return &InvitationPortIn{
		SaveInvitationUsecase: saveInvitationUsecase,
		ValidateInvitationUsecase: validateInvitationUsecase,
		AcceptInvitationUsecase: acceptInvitationUsecase,
		GetInvitationUsecase: getInvitationUsecase,
	}
}