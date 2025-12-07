package events

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

var (
	ErrMissingType = errors.New("event type is required")
)

type IngestRequest struct {
	Type     string
	Data     map[string]any
	Metadata map[string]any
	Source   SourceInfo
}

type Service interface {
	Ingest(ctx context.Context, tenantID string, req IngestRequest) (EventEnvelope, error)
}

type logService struct {
	log logger.Logger
}

func NewLogService(log logger.Logger) Service {
	return &logService{log: log}
}

func (s *logService) Ingest(ctx context.Context, tenantID string, req IngestRequest) (EventEnvelope, error) {
	if req.Type == "" {
		return EventEnvelope{}, ErrMissingType
	}

	env := EventEnvelope{
		ID:       uuid.NewString(),
		TenantID: tenantID,
		Type:     req.Type,
		Source:   req.Source,
		Data:     req.Data,
		Metadata: req.Metadata,
		Status: EventStatus{
			IngestedAt:    time.Now().UTC(),
			DeliveryState: "PENDING",
		},
	}

	s.log.Info("event_ingested",
		"event_id", env.ID,
		"tenant_id", env.TenantID,
		"type", env.Type,
		"protocol", env.Source.Protocol,
	)

	return env, nil
}
