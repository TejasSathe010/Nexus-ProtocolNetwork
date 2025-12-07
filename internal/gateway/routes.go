package gateway

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/events"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/realtime"
)

func NewRouter(
	log logger.Logger,
	eventSvc events.Service,
	wsHub *realtime.WSHub,
	sseBroker *realtime.SSEBroker,
	rtBroadcaster realtime.Broadcaster,
) http.Handler {
	r := chi.NewRouter()

	r.Use(RequestIDMiddleware)
	r.Use(RecoverMiddleware(log))
	r.Use(LoggingMiddleware(log))
	r.Use(AuthMiddleware(log))

	r.Group(func(r chi.Router) {
		r.Use()
		r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
	})

	r.Route("/api/v1", func(api chi.Router) {
		h := NewEventHandler(log, eventSvc, rtBroadcaster)
		api.Post("/events", h.HandleRESTIngest)
	})

	r.Get("/ws", NewWSHandler(log, wsHub))

	r.Get("/sse/stream", NewSSEHandler(log, sseBroker))

	return r
}
