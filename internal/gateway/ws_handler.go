package gateway

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/logger"
	"github.com/tejassathe/Nexus-ProtocolNetwork/pkg/realtime"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewWSHandler(log logger.Logger, hub *realtime.WSHub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID, _ := r.Context().Value(ContextKeyTenantID).(string)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("websocket upgrade failed", "err", err)
			return
		}

		client := realtime.NewWSClient(conn, log, hub, tenantID)

		go client.WritePump()
		go client.ReadPump()
	}
}
