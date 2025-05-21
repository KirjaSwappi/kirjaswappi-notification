package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/service"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func NewHandler(b *service.Broadcaster) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("userId")
		if userID == "" {
			http.Error(w, "userId query param required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		ch := b.Subscribe(userID)
		defer b.Unsubscribe(userID, ch)

		for msg := range ch {
			if err := conn.WriteJSON(msg); err != nil {
				log.Println("WebSocket write error:", err)
				return
			}
		}
	}
}
