package adminhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// Blank assignments keep the types referenced for swaggo without runtime use.
var (
	_ = response.Response{}
	_ = UserSummaryResponse{}
	_ = UserDetailResponse{}
	_ = UpdateUserStatusRequest{}
	_ = UpdateUserRoleRequest{}
)

// listUsersDoc documents the list users handler.
//
//	@Summary		List users
//	@Description	Search, filter and paginate users. Requires the admin role.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			q				query		string	false	"Search by name, email or staff ID"
//	@Param			role			query		string	false	"admin or user"
//	@Param			status			query		string	false	"active, suspended or inactive"
//	@Param			department_id	query		string	false	"Department ID"
//	@Param			sort			query		string	false	"created_at, -created_at, last_name or email"
//	@Param			limit			query		int		false	"Page size (max 100)"
//	@Param			offset			query		int		false	"Offset"
//	@Success		200				{object}	response.Response{data=[]UserSummaryResponse}
//	@Failure		400				{object}	response.ErrorResponse	"Validation error"
//	@Failure		401				{object}	response.Response		"Missing or invalid access token"
//	@Failure		403				{object}	response.Response		"Authenticated user is not an admin"
//	@Router			/admin/users [get]
func listUsersDoc() {}

// getUserDoc documents the user detail handler.
//
//	@Summary		Get user
//	@Description	Get a user with department, requisition counts and spend. Requires the admin role.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	response.Response{data=UserDetailResponse}
//	@Failure		400	{object}	response.ErrorResponse	"Invalid user ID"
//	@Failure		401	{object}	response.Response		"Missing or invalid access token"
//	@Failure		403	{object}	response.Response		"Authenticated user is not an admin"
//	@Failure		404	{object}	response.Response		"User not found"
//	@Router			/admin/users/{id} [get]
func getUserDoc() {}

// updateUserStatusDoc documents the status change handler.
//
//	@Summary		Update user status
//	@Description	Activate, suspend or deactivate a user. Requires the admin role. Self-changes and removing the last active admin are rejected.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"User ID"
//	@Param			request	body		UpdateUserStatusRequest	true	"New status and reason"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Authenticated user is not an admin or is the target"
//	@Failure		404		{object}	response.Response		"User not found"
//	@Failure		409		{object}	response.Response		"Target is the last active admin"
//	@Router			/admin/users/{id}/status [patch]
func updateUserStatusDoc() {}

// updateUserRoleDoc documents the role change handler.
//
//	@Summary		Update user role
//	@Description	Promote or demote a user. Requires the admin role. Self-changes and demoting the last active admin are rejected.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string				true	"User ID"
//	@Param			request	body		UpdateUserRoleRequest	true	"New role and reason"
//	@Success		200		{object}	response.Response
//	@Failure		400		{object}	response.ErrorResponse	"Validation error"
//	@Failure		401		{object}	response.Response		"Missing or invalid access token"
//	@Failure		403		{object}	response.Response		"Authenticated user is not an admin or is the target"
//	@Failure		404		{object}	response.Response		"User not found"
//	@Failure		409		{object}	response.Response		"Target is the last active admin"
//	@Router			/admin/users/{id}/role [patch]
func updateUserRoleDoc() {}

// revokeUserSessionsDoc documents the force logout handler.
//
//	@Summary		Revoke user sessions
//	@Description	Invalidate every refresh session for a user. Access tokens expire within their TTL. Requires the admin role.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"User ID"
//	@Param			reason	query		string	false	"Optional reason for the audit log"
//	@Success		200		{object}	response.Response
//	@Failure		401		{object}	response.Response	"Missing or invalid access token"
//	@Failure		403		{object}	response.Response	"Authenticated user is not an admin"
//	@Failure		404		{object}	response.Response	"User not found"
//	@Router			/admin/users/{id}/sessions/revoke [post]
func revokeUserSessionsDoc() {}
