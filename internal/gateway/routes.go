package gateway

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/events"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

func NewRouter(log logger.Logger, eventSvc events.Service) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestIDMiddleware)
	r.Use(RecoverMiddleware(log))
	r.Use(LoggingMiddleware(log))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Use(AuthMiddleware(log))

		h := NewEventHandler(log, eventSvc)
		api.Post("/events", h.HandleRESTIngest)
	})

	return r
}
