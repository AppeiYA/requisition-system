package identity

import (
	identityhttp "github.com/AppeiYA/requisition-system/internal/identity/adapters/in/http"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
)

type IdentityModule struct {
	identityPortIn  *identityin.IdentityPortIn
	identityHandler *identityhttp.Handler
	authMiddleware  *identityhttp.AuthMiddleware
	userLookup      requisitionout.UserLookup
}

func NewIdentityModule(
	identityPortIn *identityin.IdentityPortIn,
	identityHandler *identityhttp.Handler,
	authMiddleware *identityhttp.AuthMiddleware,
	userLookup requisitionout.UserLookup,
) *IdentityModule {
	return &IdentityModule{
		identityPortIn:  identityPortIn,
		identityHandler: identityHandler,
		authMiddleware:  authMiddleware,
		userLookup:      userLookup,
	}
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
