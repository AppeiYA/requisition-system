package identity

import (
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type IdentityModule struct {
	identityPortIn   *identityin.IdentityPortIn
	adminUsersPortIn *identityin.AdminUsersPortIn
	identityHandler  *identityhttp.Handler
	authMiddleware   *identityhttp.AuthMiddleware
	userLookup       requisitionout.UserLookup
}

func NewIdentityModule(
	identityPortIn *identityin.IdentityPortIn,
	adminUsersPortIn *identityin.AdminUsersPortIn,
	identityHandler *identityhttp.Handler,
	authMiddleware *identityhttp.AuthMiddleware,
	userLookup requisitionout.UserLookup,
) *IdentityModule {
	return &IdentityModule{
		identityPortIn:   identityPortIn,
		adminUsersPortIn: adminUsersPortIn,
		identityHandler:  identityHandler,
		authMiddleware:   authMiddleware,
		userLookup:       userLookup,
	}
}

// AdminUsers exposes the identity use cases behind the admin API.
func (m *IdentityModule) AdminUsers() *identityin.AdminUsersPortIn {
	return m.adminUsersPortIn
}

func (m *IdentityModule) UserLookup() requisitionout.UserLookup {
	return m.userLookup
}

func (m *IdentityModule) PortIn() *identityin.IdentityPortIn {
	return m.identityPortIn
}

func (m *IdentityModule) Handler() *identityhttp.Handler {
	return m.identityHandler
}

func (m *IdentityModule) AuthMiddleware() *identityhttp.AuthMiddleware {
	return m.authMiddleware
}
