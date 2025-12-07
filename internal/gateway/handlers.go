package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/events"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/realtime"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/routing"
)

type EventHandler struct {
	log           logger.Logger
	eventSvc      events.Service
	rtBroadcaster realtime.Broadcaster
	router        *routing.Engine
}

func NewEventHandler(
	log logger.Logger,
	es events.Service,
	rt realtime.Broadcaster,
	router *routing.Engine,
) *EventHandler {
	return &EventHandler{
		log:           log,
		eventSvc:      es,
		rtBroadcaster: rt,
		router:        router,
	}
}

type restIngestRequest struct {
	Type     string         `json:"type"`
	Data     map[string]any `json:"data"`
	Metadata map[string]any `json:"metadata"`
}

type restIngestResponse struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

func (h *EventHandler) HandleRESTIngest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, _ := ctx.Value(ContextKeyTenantID).(string)

	var reqBody restIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	src := events.SourceInfo{
		Protocol:  "REST",
		Endpoint:  r.URL.Path,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
	}

	env, err := h.eventSvc.Ingest(ctx, tenantID, events.IngestRequest{
		Type:     reqBody.Type,
		Data:     reqBody.Data,
		Metadata: reqBody.Metadata,
		Source:   src,
	})
	if err != nil {
		if err == events.ErrMissingType {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.log.Error("failed to ingest event", "err", err)
		http.Error(w, "failed to ingest event", http.StatusInternalServerError)
		return
	}

	channels, err := h.resolveChannels(ctx, tenantID, env.Type)
	if err != nil {
		h.log.Warn("resolve channels failed; using default", "err", err)
	}

	if len(channels) == 0 {
		channels = []string{DefaultTenantChannel(tenantID)}
	}

	for _, ch := range channels {
		if err := h.rtBroadcaster.BroadcastEvent(ctx, ch, env); err != nil {
			h.log.Warn("failed to broadcast event",
				"err", err,
				"channel", ch,
				"event_id", env.ID,
			)
		}
	}

	resp := restIngestResponse{
		EventID: env.ID,
		Status:  "accepted",
	}
	writeJSON(w, http.StatusAccepted, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func DefaultTenantChannel(tenantID string) string {
	return fmt.Sprintf("tenant:%s:events", tenantID)
}

func (h *EventHandler) resolveChannels(ctx context.Context, tenantID, eventType string) ([]string, error) {
	if h.router == nil {
		return nil, nil
	}
	return h.router.ResolveChannels(ctx, tenantID, eventType)
}
