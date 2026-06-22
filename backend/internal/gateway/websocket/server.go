package websocket

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend/internal/event"
	"backend/internal/logger"
)

// Server WebSocket HTTP 服务。
type Server struct {
	Hub       *Hub
	appLog    *logger.AppLogger
	publisher eventPublisher
}

// NewServer 创建 WebSocket 服务。
func NewServer(appLog *logger.AppLogger, publisher eventPublisher) *Server {
	return &Server{
		Hub:       NewHub(appLog),
		appLog:    appLog,
		publisher: publisher,
	}
}

// Start 启动 Hub 并监听 WebSocket 连接。
func (s *Server) Start(addr string) error {
	go s.Hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWS(s.Hub, w, r, s.onUserMessage)
	})

	s.appLog.Info("WebSocket 网关已启动", "addr", addr)
	return http.ListenAndServe(addr, nil)
}

// onUserMessage 接收用户消息并发布 user.message.sent 事件。
func (s *Server) onUserMessage(userToken string, raw []byte) {
	var msg Message
	if err := json.Unmarshal(raw, &msg); err != nil {
		s.appLog.Warn("WebSocket 消息解析失败", "user_token", userToken, "raw", string(raw))
		return
	}

	if msg.Type != MsgTypeUserMessage {
		s.appLog.Warn("未知消息类型", "type", msg.Type)
		return
	}

	data, _ := json.Marshal(msg.Data)
	var userMsg UserMessage
	if err := json.Unmarshal(data, &userMsg); err != nil {
		s.appLog.Warn("用户消息数据解析失败", "user_token", userToken)
		return
	}
	userMsg.UserToken = userToken

	payload, _ := json.Marshal(userMsg)
	e := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeUserMessageSent,
		TownID:    1,
		ActorType: event.ActorTypeUser,
		ActorID:   userToken,
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	if err := s.publisher.Publish(context.Background(), e); err != nil {
		s.appLog.Error(err, "发布 user.message.sent 失败", "user_token", userToken)
	}
}

func newEventID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
