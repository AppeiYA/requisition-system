package adminhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// Blank assignments keep the adapter models referenced for swaggo without
// runtime use. JSON tags live on these models, never on the port DTOs.
var (
	_ = response.Response{}
	_ = InvitationSummaryResponse{}
	_ = InvitationDetailResponse{}
	_ = ResendInvitationResponse{}
)

// listInvitationsDoc documents the list invitations handler.
//
//	@Summary		List invitations
//	@Description	Search, filter and paginate invitations. Requires the admin role. Token hashes are never returned.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q				query		string	false	"Search by email or name"
//	@Param			status			query		string	false	"pending, delivered, failed, accepted, revoked or expired"
//	@Param			department_id	query		string	false	"Department ID"
//	@Param			invited_by		query		string	false	"Inviter user ID"
//	@Param			from			query		string	false	"Created at from (RFC3339 or YYYY-MM-DD, UTC)"
//	@Param			to				query		string	false	"Created at before (RFC3339 or YYYY-MM-DD, UTC)"
//	@Param			limit			query		int		false	"Page size (max 100)"
//	@Param			offset			query		int		false	"Offset"
//	@Success		200				{object}	response.Response{data=[]InvitationSummaryResponse}
//	@Failure		400				{object}	response.ErrorResponse	"Validation error"
//	@Failure		401				{object}	response.Response		"Missing or invalid access token"
//	@Failure		403				{object}	response.Response		"Authenticated user is not an admin"
//	@Router			/admin/invitations [get]
func listInvitationsDoc() {}

// getInvitationDoc documents the invitation detail handler.
//
//	@Summary		Get invitation
//	@Description	Get one invitation with its department and inviter names. Requires the admin role.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Invitation ID"
//	@Success		200	{object}	response.Response{data=InvitationDetailResponse}
//	@Failure		400	{object}	response.ErrorResponse	"Invalid invitation ID"
//	@Failure		401	{object}	response.Response		"Missing or invalid access token"
//	@Failure		403	{object}	response.Response		"Authenticated user is not an admin"
//	@Failure		404	{object}	response.Response		"Invitation not found"
//	@Router			/admin/invitations/{id} [get]
func getInvitationDoc() {}

// resendInvitationDoc documents the resend handler.
//
//	@Summary		Resend invitation
//	@Description	Rotate the invitation token, audit the change and queue a replacement email. Accepted and revoked invitations cannot be resent. Requires the admin role.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Invitation ID"
//	@Success		200	{object}	response.Response{data=ResendInvitationResponse}
//	@Failure		400	{object}	response.ErrorResponse	"Invalid invitation ID"
//	@Failure		401	{object}	response.Response		"Missing or invalid access token"
//	@Failure		403	{object}	response.Response		"Authenticated user is not an admin"
//	@Failure		404	{object}	response.Response		"Invitation not found"
//	@Failure		409	{object}	response.Response		"Invitation cannot be resent or another active invitation exists"
//	@Router			/admin/invitations/{id}/resend [post]
func resendInvitationDoc() {}

// revokeInvitationDoc documents the revoke handler.
//
//	@Summary		Revoke invitation
//	@Description	Invalidate the invitation link immediately. Revoking an already revoked invitation is a no-op. Requires the admin role.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string	true	"Invitation ID"
//	@Param			reason	query	string	false	"Optional reason for the audit log"
//	@Success		200		{object}	response.Response
//	@Failure		400	{object}	response.ErrorResponse	"Invalid invitation ID"
//	@Failure		401	{object}	response.Response		"Missing or invalid access token"
//	@Failure		403	{object}	response.Response		"Authenticated user is not an admin"
//	@Failure		404	{object}	response.Response		"Invitation not found"
//	@Failure		409	{object}	response.Response		"Invitation has been accepted"
//	@Router			/admin/invitations/{id}/revoke [post]
func revokeInvitationDoc() {}
