package event

import (
	"encoding/json"
	"time"
)

// Event 统一事件结构，在 RabbitMQ 上以 JSON 传输。
type Event struct {
	EventID   string          `json:"event_id"`
	EventType string          `json:"event_type"`
	TownID    uint            `json:"town_id"`
	ActorType string          `json:"actor_type"`
	ActorID   string          `json:"actor_id"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}
