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
	FindByTownID(townID uint) ([]model.NPC, error)
}

type agentReplier interface {
	GenerateReply(ctx context.Context, npcID uint, userMsg, userToken string) (npcName string, reply string, err error)
}

type broadcastPusher interface {
	Push(eventType string, data any)
}

type npcFinder interface {
	FindByID(id uint) (*model.NPC, error)
}

type locationFinder interface {
	FindByID(id uint) (*model.Location, error)
	FindByName(townID int64, name string) (*model.Location, error)
}

type userMessenger interface {
	SendToUser(userToken string, msg *ws.Message)
}

type npcStatusUpdater interface {
	FindByTownID(townID uint) ([]model.NPC, error)
	UpdateFull(npc *model.NPC) error
	UpdateMood(npcID uint, mood string) error
	UpdateGoal(npcID uint, goal string) error
}
