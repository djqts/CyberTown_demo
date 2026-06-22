package worker

import (
	"context"

	"backend/internal/event"
	ws "backend/internal/gateway/websocket"
	"backend/internal/model"
	"backend/internal/service"
)

type consumer interface {
	Consume(ctx context.Context, queue string, handler event.Handler) error
}

type eventPublisher interface {
	Publish(ctx context.Context, e *event.Event) error
}

type eventLogCreator interface {
	Create(event *model.EventLog) error
}

type npcMoveService interface {
	FindActiveMoves(townID uint, minuteOfDay int) ([]service.NPCMove, error)
	MoveNPC(npcID, newLocationID uint) error
}

type agentReplier interface {
	GenerateReply(ctx context.Context, npcID uint, userMsg, userToken string) (string, error)
}

type broadcastPusher interface {
	Push(eventType string, data any)
}

type npcFinder interface {
	FindByID(id uint) (*model.NPC, error)
}

type locationFinder interface {
	FindByID(id uint) (*model.Location, error)
}

type userMessenger interface {
	SendToUser(userToken string, msg *ws.Message)
}
