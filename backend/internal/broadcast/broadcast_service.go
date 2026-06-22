package broadcast

import (
	ws "backend/internal/gateway/websocket"
)

// Service 统一推送服务，封装 WebSocket Hub。
type Service struct {
	hub *ws.Hub
}

// NewService 创建广播服务。
func NewService(hub *ws.Hub) *Service {
	return &Service{hub: hub}
}

// Push 推送小镇事件到所有 WebSocket 客户端。
func (s *Service) Push(eventType string, data any) {
	s.hub.Broadcast(&ws.Message{
		Type: eventType,
		Data: data,
	})
}
