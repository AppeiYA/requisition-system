package invitationhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// The blank assignment keeps the imported type referenced so swaggo can resolve
// it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// validateInvitationDoc documents the ValidateInvitation handler in
// validate_invitation.go.
//
//	@Summary		Validate an invitation token
//	@Description	Checks whether an invitation token is still valid and returns the prefilled registration details for the invitee.
//	@Tags			invitation
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		200		{object}	response.Response{data=ValidateInvitationResponse}
//	@Failure		400		{object}	response.ErrorResponse	"Token is missing"
//	@Failure		404		{object}	response.Response		"Invitation not found"
//	@Router			/v1/invitations/{token}/validate [get]
func validateInvitationDoc() {}
