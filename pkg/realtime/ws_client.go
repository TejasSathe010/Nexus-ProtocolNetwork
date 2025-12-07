package realtime

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
)

type WSClient struct {
	Conn   *websocket.Conn
	Send   chan []byte
	Log    logger.Logger
	Hub    *WSHub
	Tenant string
}

type WSSubscribeMessage struct {
	Action  string `json:"action"`
	Channel string `json:"channel"`
}

func NewWSClient(conn *websocket.Conn, log logger.Logger, hub *WSHub, tenant string) *WSClient {
	return &WSClient{
		Conn:   conn,
		Send:   make(chan []byte, 32),
		Log:    log,
		Hub:    hub,
		Tenant: tenant,
	}
}

func (c *WSClient) ReadPump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			c.Log.Warn("ws read error", "err", err)
			return
		}

		var m WSSubscribeMessage
		if err := json.Unmarshal(msg, &m); err != nil {
			c.Log.Warn("invalid ws message", "err", err)
			continue
		}

		switch m.Action {
		case "subscribe":
			c.Hub.Register(m.Channel, c)
			c.Log.Info("ws subscribed", "tenant", c.Tenant, "channel", m.Channel)

		case "unsubscribe":
			c.Hub.Unregister(m.Channel, c)
			c.Log.Info("ws unsubscribed", "tenant", c.Tenant, "channel", m.Channel)

		default:
			c.Log.Warn("unknown ws action", "action", m.Action)
		}
	}
}

func (c *WSClient) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			c.Log.Warn("ws write error", "err", err)
			return
		}
	}
}
