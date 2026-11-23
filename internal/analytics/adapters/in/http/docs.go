package analyticshttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

var _ = response.Response{}

// overviewDoc documents the analytics overview endpoint.
//
//	@Summary		Get analytics overview
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query	string	false	"Start time (RFC3339 or YYYY-MM-DD)"
//	@Param			to	query	string	false	"End time (RFC3339 or YYYY-MM-DD)"
//	@Param			department_id	query	string	false	"Department ID"
//	@Param			interval	query	string	false	"day, week or month"
//	@Success		200	{object}	response.Response
//	@Failure		400	{object}	response.ErrorResponse
//	@Router			/admin/analytics/overview [get]
func overviewDoc() {}

// requisitionsDoc documents requisition analytics.
//
//	@Summary		Get requisition analytics
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query	string	false	"Start time"
//	@Param			to	query	string	false	"End time"
//	@Param			department_id	query	string	false	"Department ID"
//	@Param			interval	query	string	false	"day, week or month"
//	@Param			limit	query	int	false	"Top-list limit"
//	@Success		200	{object}	response.Response
//	@Router			/admin/analytics/requisitions [get]
func requisitionsDoc() {}

// usersDoc documents user analytics.
//
//	@Summary		Get user analytics
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query	string	false	"Start time"
//	@Param			to	query	string	false	"End time"
//	@Param			interval	query	string	false	"day, week or month"
//	@Param			limit	query	int	false	"Top-list limit"
//	@Success		200	{object}	response.Response
//	@Router			/admin/analytics/users [get]
func usersDoc() {}

// catalogueDoc documents catalogue analytics.
//
//	@Summary		Get catalogue analytics
//	@Tags			admin
//	@Produce		json
//	@Security		BearerAuth
//	@Param			from	query	string	false	"Start time"
//	@Param			to	query	string	false	"End time"
//	@Param			department_id	query	string	false	"Department ID"
//	@Param			limit	query	int	false	"Top-list limit"
//	@Success		200	{object}	response.Response
//	@Router			/admin/analytics/catalogue [get]
func catalogueDoc() {}
