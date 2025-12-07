package routing

import (
	"context"
	"fmt"

	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/control"
)

type Engine struct {
	store *control.Store
}

func NewEngine(store *control.Store) *Engine {
	return &Engine{store: store}
}

func (e *Engine) ResolveChannels(ctx context.Context, tenantID, eventType string) ([]string, error) {
	routes, err := e.store.FindRoutesForEvent(ctx, tenantID, eventType)
	if err != nil {
		return nil, fmt.Errorf("resolve channels: %w", err)
	}

	channels := make([]string, 0, len(routes))
	for _, r := range routes {
		channels = append(channels, r.TargetChannel)
	}
	return channels, nil
}
