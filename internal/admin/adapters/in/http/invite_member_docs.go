package adminhttp

import (
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"
)

// The blank assignments keep the imported types referenced so swaggo can
// resolve them from the annotations below; they have no runtime effect.
var (
	_ = response.Response{}
	_ = adminin.InviteMemberResult{}
)

// inviteMemberDoc documents the invite handler in handler.go.
//
//	@Summary		Invite a member
//	@Description	Creates a pending invitation for the given email and department and emails the invitation link. Requires the admin role.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		InviteRequest	true	"Invitation details"
//	@Success		201		{object}	response.Response{data=adminin.InviteMemberResult}
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Authenticated user is not an admin"
//	@Failure		409		{object}	response.Response		"Invitation already exists"
//	@Router			/admin/invitations [post]
func inviteMemberDoc() {}
