package analytics

import (
	analyticshttp "github.com/AppeiYA/requisition-system/internal/analytics/adapters/in/http"
)

type Module struct {
	handler *analyticshttp.Handler
}

func NewModule(handler *analyticshttp.Handler) *Module {
	return &Module{handler: handler}
}

func (m *Module) Handler() *analyticshttp.Handler {
	return m.handler
}
