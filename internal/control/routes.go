package control

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tejassathe/Nexus-ProtocolNetwork/internal/gateway"
	ctl "github.com/tejassathe/Nexus-ProtocolNetwork/pkg/control"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

func NewRouter(log logger.Logger, store *ctl.Store) http.Handler {
	r := chi.NewRouter()

	r.Use(gateway.RequestIDMiddleware)
	r.Use(gateway.RecoverMiddleware(log))
	r.Use(gateway.LoggingMiddleware(log))

	h := NewHandler(log, store)

	r.Route("/control", func(cr chi.Router) {
		cr.Post("/tenants", h.CreateTenant)
		cr.Get("/tenants", h.ListTenants)
		cr.Post("/tenants/{tenant_id}/api-keys", h.CreateAPIKey)
		cr.Get("/tenants/{tenant_id}/routes", h.ListRoutes)
		cr.Post("/tenants/{tenant_id}/routes", h.CreateRoute)
	})

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return r
}
