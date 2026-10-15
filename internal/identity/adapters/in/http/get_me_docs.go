package identityhttp

import "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/response"

// The blank assignment keeps the imported type referenced so swaggo can resolve
// it from the annotations below; it has no runtime effect.
var _ = response.Response{}

// getMeDoc documents the getMe handler in get_me.go.
//
//	@Summary		Get the authenticated user
//	@Description	Returns the profile of the user identified by the bearer token.
//	@Tags			identity
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.Response{data=User}
//	@Failure		401	{object}	response.Response	"Missing or invalid access token"
//	@Router			/auth/me [get]
func getMeDoc() {}
