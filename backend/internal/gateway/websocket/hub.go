package websocket

import (
	"backend/internal/logger"
)

// Hub owns websocket client maps and serializes all access through Run.
type Hub struct {
	clients       map[*Client]bool
	clientsByUser map[string]*Client
	broadcast     chan *Message
	direct        chan directMessage
	register      chan *Client
	unregister    chan *Client
	appLog        *logger.AppLogger
}

type directMessage struct {
	userToken string
	msg       *Message
}

func NewHub(appLog *logger.AppLogger) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		clientsByUser: make(map[string]*Client),
		broadcast:     make(chan *Message, 256),
		direct:        make(chan directMessage, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		appLog:        appLog,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if client.UserToken != "" {
				h.clientsByUser[client.UserToken] = client
			}
			h.appLog.Info("WebSocket client connected", "total", len(h.clients), "user_token", client.UserToken)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.UserToken != "" && h.clientsByUser[client.UserToken] == client {
					delete(h.clientsByUser, client.UserToken)
				}
				close(client.send)
				h.appLog.Info("WebSocket client disconnected", "total", len(h.clients))
			}

		case msg := <-h.broadcast:
			for client := range h.clients {
				if err := client.Send(msg); err != nil {
					h.appLog.Error(err, "broadcast websocket message failed")
				}
			}

		case direct := <-h.direct:
			if client, ok := h.clientsByUser[direct.userToken]; ok {
				if err := client.Send(direct.msg); err != nil {
					h.appLog.Error(err, "send websocket message to user failed", "user_token", direct.userToken)
				}
			} else {
				h.appLog.Warn("websocket user is offline; dropping message", "user_token", direct.userToken)
			}
		}
	}
}

func (h *Hub) Broadcast(msg *Message) {
	select {
	case h.broadcast <- msg:
	default:
		h.appLog.Warn("websocket broadcast channel full; dropping message", "type", msg.Type)
	}
}

func (h *Hub) SendToUser(userToken string, msg *Message) {
	select {
	case h.direct <- directMessage{userToken: userToken, msg: msg}:
	default:
		h.appLog.Warn("websocket direct channel full; dropping message", "user_token", userToken)
	}
}
