package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/broadcast"
	"backend/internal/event"
	"backend/internal/logger"
	"backend/internal/repo"

	ws "backend/internal/gateway/websocket"
)

// BroadcastWorker 消费 npc.moved / npc.replied，广播到 WebSocket。
type BroadcastWorker struct {
	consumer  consumer
	bcast     broadcastPusher
	eventRepo eventLogCreator
	npcRepo   npcFinder
	locRepo   locationFinder
	hub       userMessenger
	appLog    *logger.AppLogger
}

// NewBroadcastWorker 创建广播 Worker。
func NewBroadcastWorker(
	consumer *event.Consumer,
	bcast *broadcast.Service,
	eventRepo *repo.EventRepo,
	npcRepo *repo.NPCRepo,
	locRepo *repo.LocationRepo,
	hub *ws.Hub,
	appLog *logger.AppLogger,
) *BroadcastWorker {
	return &BroadcastWorker{
		consumer:  consumer,
		bcast:     bcast,
		eventRepo: eventRepo,
		npcRepo:   npcRepo,
		locRepo:   locRepo,
		hub:       hub,
		appLog:    appLog,
	}
}

// Start 从 town_broadcast 队列消费事件。
func (w *BroadcastWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, "town_broadcast", w.handleEvent)
}

func (w *BroadcastWorker) handleEvent(ctx context.Context, e *event.Event) error {
	switch e.EventType {
	case event.EventTypeNPCMoved:
		return w.handleNPCMoved(ctx, e)
	case event.EventTypeNPCReplied:
		return w.handleNPCReplied(ctx, e)
	default:
		return nil
	}
}

func (w *BroadcastWorker) handleNPCMoved(_ context.Context, e *event.Event) error {
	var payload struct {
		NPCID          uint `json:"npc_id"`
		FromLocationID uint `json:"from_location_id"`
		ToLocationID   uint `json:"to_location_id"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse npc.moved payload: %w", err)
	}

	npc, err := w.npcRepo.FindByID(payload.NPCID)
	if err != nil {
		return fmt.Errorf("find npc: %w", err)
	}

	fromLoc, err := w.locRepo.FindByID(payload.FromLocationID)
	if err != nil {
		return fmt.Errorf("find previous location: %w", err)
	}

	toLoc, err := w.locRepo.FindByID(payload.ToLocationID)
	if err != nil {
		return fmt.Errorf("find location: %w", err)
	}

	// 推送到 WebSocket
	w.bcast.Push(e.EventType, map[string]any{
		"npc_id":        npc.ID,
		"npc_name":      npc.Name,
		"from_location": fromLoc.Name,
		"to_location":   toLoc.Name,
	})

	// 写入 town.event.broadcast 日志
	writeEventLog(w.eventRepo, &event.Event{
		EventID:   e.EventID,
		EventType: event.EventTypeBroadcast,
		TownID:    e.TownID,
		ActorType: e.ActorType,
		ActorID:   e.ActorID,
		Payload:   e.Payload,
		CreatedAt: e.CreatedAt,
	})

	w.appLog.Info("事件已广播",
		"event_type", e.EventType,
		"npc_name", npc.Name,
		"from", fromLoc.Name,
		"to", toLoc.Name,
	)
	return nil
}

func (w *BroadcastWorker) handleNPCReplied(_ context.Context, e *event.Event) error {
	var payload ws.NPCReplied
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse npc.replied payload: %w", err)
	}

	// 定向推送给发送消息的用户
	w.hub.SendToUser(payload.UserToken, &ws.Message{
		Type: ws.MsgTypeNPCReplied,
		Data: payload,
	})

	w.appLog.Info("NPC 回复已定向推送",
		"npc_id", payload.NPCID,
		"user_token", payload.UserToken,
	)
	return nil
}
