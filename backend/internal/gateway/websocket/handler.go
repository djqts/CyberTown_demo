package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// ServeWS 处理 WebSocket 升级请求，从 query 读取 user_token。
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, onMessage OnMessageFunc) {
	userToken := r.URL.Query().Get("user_token")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{
		hub:       hub,
		conn:      conn,
		send:      make(chan []byte, 256),
		UserToken: userToken,
		onMessage: onMessage,
	}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}
