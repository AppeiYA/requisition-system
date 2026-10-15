package invitationhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// The blank assignment keeps the imported type referenced so swaggo can resolve
// it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// acceptInvitationDoc documents the AcceptInvitation handler in
// accept_invitation.go.
//
//	@Summary		Accept an invitation
//	@Description	Completes registration for the invitee using the invitation token.
//	@Tags			invitation
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string					true	"Invitation token"
//	@Param			request	body		AcceptInvitationRequest	true	"Registration details"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		404		{object}	response.Response		"Invitation not found"
//	@Router			/v1/invitations/{token}/accept [post]
func acceptInvitationDoc() {}
