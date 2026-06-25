package broadcast

import (
	"encoding/json"

	hw "backend/internal/gateway/http"
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
	// Record trace for diagnostics — handle both map and raw JSON
	if m, ok := data.(map[string]any); ok {
		hw.GlobalDiag.RecordBroadcast(eventType, m)
	} else if raw, ok := data.(json.RawMessage); ok {
		var m map[string]any
		if json.Unmarshal(raw, &m) == nil {
			hw.GlobalDiag.RecordBroadcast(eventType, m)
		}
	}
	s.hub.Broadcast(&ws.Message{
		Type: eventType,
		Data: data,
	})
}
