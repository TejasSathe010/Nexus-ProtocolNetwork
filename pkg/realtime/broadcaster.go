package realtime

import (
	"context"
	"encoding/json"

	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/events"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

type Broadcaster interface {
	BroadcastEvent(ctx context.Context, channel string, env events.EventEnvelope) error
}

type rtBroadcaster struct {
	log   logger.Logger
	wsHub *WSHub
	sse   *SSEBroker
}

func NewBroadcaster(log logger.Logger, wsHub *WSHub, sse *SSEBroker) Broadcaster {
	return &rtBroadcaster{
		log:   log,
		wsHub: wsHub,
		sse:   sse,
	}
}

func (b *rtBroadcaster) BroadcastEvent(ctx context.Context, channel string, env events.EventEnvelope) error {
	payload, err := json.Marshal(map[string]any{
		"channel": channel,
		"event":   env,
	})
	if err != nil {
		b.log.Error("failed to marshal event for broadcast", "err", err)
		return err
	}

	b.wsHub.Publish(channel, payload)
	b.sse.Publish(channel, payload)

	return nil
}
