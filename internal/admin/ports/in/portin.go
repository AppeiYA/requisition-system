package adminin

type AdminPortIn struct {
	InviteMember       InviteMemberUseCase
	ListUsers          ListUsersUseCase
	GetUser            GetUserUseCase
	UpdateUserStatus   UpdateUserStatusUseCase
	UpdateUserRole     UpdateUserRoleUseCase
	RevokeUserSessions RevokeUserSessionsUseCase
	ListInvitations    ListInvitationsUseCase
	GetInvitation      GetInvitationUseCase
	ResendInvitation   ResendInvitationUseCase
	RevokeInvitation   RevokeInvitationUseCase
}

func NewAdminPortIn(inviteMember InviteMemberUseCase) *AdminPortIn {
	return &AdminPortIn{
		InviteMember: inviteMember,
	}
}

// NewAdminPortInWithAll wires the full admin surface. The original constructor
// is kept for callers that only exercise invitations.
func NewAdminPortInWithAll(
	inviteMember InviteMemberUseCase,
	listUsers ListUsersUseCase,
	getUser GetUserUseCase,
	updateUserStatus UpdateUserStatusUseCase,
	updateUserRole UpdateUserRoleUseCase,
	revokeUserSessions RevokeUserSessionsUseCase,
	listInvitations ListInvitationsUseCase,
	getInvitation GetInvitationUseCase,
	resendInvitation ResendInvitationUseCase,
	revokeInvitation RevokeInvitationUseCase,
) *AdminPortIn {
	return &AdminPortIn{
		InviteMember:       inviteMember,
		ListUsers:          listUsers,
		GetUser:            getUser,
		UpdateUserStatus:   updateUserStatus,
		UpdateUserRole:     updateUserRole,
		RevokeUserSessions: revokeUserSessions,
		ListInvitations:    listInvitations,
		GetInvitation:      getInvitation,
		ResendInvitation:   resendInvitation,
		RevokeInvitation:   revokeInvitation,
	}
}
