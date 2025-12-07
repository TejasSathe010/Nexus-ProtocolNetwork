package realtime

import "sync"

type SSEClient chan []byte

type SSEBroker struct {
	mu       sync.RWMutex
	channels map[string]map[SSEClient]struct{}
}

func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		channels: make(map[string]map[SSEClient]struct{}),
	}
}

func (b *SSEBroker) Subscribe(channel string) SSEClient {
	client := make(SSEClient, 16)

	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.channels[channel]; !ok {
		b.channels[channel] = make(map[SSEClient]struct{})
	}
	b.channels[channel][client] = struct{}{}
	return client
}

func (b *SSEBroker) Unsubscribe(channel string, client SSEClient) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.channels[channel]; ok {
		delete(subs, client)
		close(client)
		if len(subs) == 0 {
			delete(b.channels, channel)
		}
	}
}

func (b *SSEBroker) Publish(channel string, msg []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.channels[channel]; ok {
		for client := range subs {
			select {
			case client <- msg:
			default:
				// Slow consumer; drop.
			}
		}
	}
}
