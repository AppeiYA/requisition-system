package admin

import (
	adminhttp "github.com/AppeiYA/requisition-system/internal/admin/adapters/in/http"
	adminin "github.com/AppeiYA/requisition-system/internal/admin/ports/in"
)

type AdminModule struct {
	adminPortIn *adminin.AdminPortIn
	adminHandler *adminhttp.Handler
}

func NewAdminModule(portIn *adminin.AdminPortIn, adminHandler *adminhttp.Handler) *AdminModule {
	return &AdminModule{
		adminPortIn: portIn,
		adminHandler: adminHandler,
	}
}

func (m *AdminModule) PortIn() *adminin.AdminPortIn {
	return m.adminPortIn
}

func (m *AdminModule) Handler() *adminhttp.Handler {
	return m.adminHandler
}