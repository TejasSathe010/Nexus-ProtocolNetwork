package realtime

import "sync"

type WSHub struct {
	mu       sync.RWMutex
	channels map[string]map[*WSClient]struct{}
}

func NewWSHub() *WSHub {
	return &WSHub{
		channels: make(map[string]map[*WSClient]struct{}),
	}
}

func (h *WSHub) Register(channel string, c *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.channels[channel]; !ok {
		h.channels[channel] = make(map[*WSClient]struct{})
	}
	h.channels[channel][c] = struct{}{}
}

func (h *WSHub) Unregister(channel string, c *WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subs, ok := h.channels[channel]; ok {
		delete(subs, c)
		if len(subs) == 0 {
			delete(h.channels, channel)
		}
	}
}

func (h *WSHub) Publish(channel string, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if subs, ok := h.channels[channel]; ok {
		for client := range subs {
			select {
			case client.Send <- msg:
			default:
				// Slow consumer; drop message or close in future
			}
		}
	}
}
