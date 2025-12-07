package gateway

import (
	"fmt"
	"net/http"
	"time"

	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/realtime"
)

func NewSSEHandler(log logger.Logger, broker *realtime.SSEBroker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tenantID, _ := ctx.Value(ContextKeyTenantID).(string)

		channel := r.URL.Query().Get("channel")
		if channel == "" {
			channel = DefaultTenantChannel(tenantID)
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		client := broker.Subscribe(channel)
		defer broker.Unsubscribe(channel, client)

		log.Info("sse client subscribed", "tenant_id", tenantID, "channel", channel)

		fmt.Fprintf(w, ": connected\n\n")
		flusher.Flush()

		for {
			select {
			case msg, ok := <-client:
				if !ok {
					log.Info("sse client channel closed", "tenant_id", tenantID, "channel", channel)
					return
				}

				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()

			case <-ctx.Done():
				log.Info("sse client context done", "tenant_id", tenantID, "channel", channel)
				return

			case <-time.After(30 * time.Second):
				// Keep-alive.
				fmt.Fprintf(w, ": keep-alive\n\n")
				flusher.Flush()
			}
		}
	}
}
