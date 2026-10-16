package adminin

type AdminPortIn struct {
	InviteMember InviteMemberUseCase
}

func NewAdminPortIn(inviteMember InviteMemberUseCase) *AdminPortIn {
	return &AdminPortIn{
		InviteMember: inviteMember,
	}
}

