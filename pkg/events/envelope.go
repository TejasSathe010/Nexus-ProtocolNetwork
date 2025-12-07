package events

import "time"

type SourceInfo struct {
	Protocol  string            `json:"protocol"`
	Endpoint  string            `json:"endpoint"`
	IP        string            `json:"ip,omitempty"`
	UserAgent string            `json:"user_agent,omitempty"`
	Extra     map[string]string `json:"extra,omitempty"`
}

type EventStatus struct {
	IngestedAt    time.Time `json:"ingested_at"`
	DeliveryState string    `json:"delivery_state"`
}

type EventEnvelope struct {
	ID       string `json:"id"`
	TenantID string `json:"tenant_id"`
	Type     string `json:"type"`

	Source   SourceInfo     `json:"source"`
	Data     map[string]any `json:"data"`
	Metadata map[string]any `json:"metadata"`

	Status EventStatus `json:"status"`
}
