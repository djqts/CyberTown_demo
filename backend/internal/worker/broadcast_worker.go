package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/broadcast"
	"backend/internal/event"
	hw "backend/internal/gateway/http"
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
	case event.EventTypeNPCIdleAction:
		return w.handleNPCIdleAction(ctx, e)

	case event.EventTypeNPCReplied:
		return w.handleNPCReplied(ctx, e)
	case event.EventTypeNPCActivityGenerated:
		return w.handleActivityGenerated(ctx, e)
	case event.EventTypeNPCMoodChanged:
		return w.handleMoodChanged(ctx, e)
	case event.EventTypeNPCInteractionGenerated:
		return w.handleInteractionGenerated(ctx, e)
	case event.EventTypeStoryEventTriggered:
		return w.handleStoryEvent(ctx, e)
	case event.EventTypeNPCGoalChanged:
		return w.handleGoalChanged(ctx, e)
	case event.EventTypeTownNewsGenerated:
		return w.handleTownNews(ctx, e)
	default:
		return nil
	}
}

func (w *BroadcastWorker) handleNPCMoved(_ context.Context, e *event.Event) error {
	defer func() {
		if r := recover(); r != nil {
			hw.GlobalDiag.RecordError()
		}
	}()
	var payload struct {
		NPCID          uint   `json:"npc_id"`
		NPCName        string `json:"npc_name"`
		FromLocationID uint   `json:"from_location_id"`
		ToLocationID   uint   `json:"to_location_id"`
		FromLocation   string `json:"from_location"`
		ToLocation     string `json:"to_location"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse npc.moved payload: %w", err)
	}

	npcID := payload.NPCID
	npcName := payload.NPCName
	if npc, err := w.npcRepo.FindByID(payload.NPCID); err == nil {
		npcName = npc.Name
	}

	fromID := payload.FromLocationID
	toID := payload.ToLocationID
	fromName := payload.FromLocation
	toName := payload.ToLocation

	// Resolve IDs from names (demo events use string names)
	if fromID == 0 && fromName != "" {
		if loc, err := w.locRepo.FindByName(int64(e.TownID), fromName); err == nil {
			fromID = loc.ID
		}
	}
	if toID == 0 && toName != "" {
		if loc, err := w.locRepo.FindByName(int64(e.TownID), toName); err == nil {
			toID = loc.ID
		}
	}

	// Resolve names from IDs (scheduled events use numeric IDs)
	if fromName == "" && fromID > 0 {
		if loc, err := w.locRepo.FindByID(fromID); err == nil {
			fromName = loc.Name
		}
	}
	if toName == "" && toID > 0 {
		if loc, err := w.locRepo.FindByID(toID); err == nil {
			toName = loc.Name
		}
	}
	if fromName == "" {
		fromName = "未知"
	}
	if toName == "" {
		toName = "未知"
	}

	w.bcast.Push(e.EventType, map[string]any{
		"npc_id":          npcID,
		"npc_name":        npcName,
		"from_location":   fromName,
		"to_location":     toName,
		"from_location_id": fromID,
		"to_location_id":   toID,
	})

	writeEventLog(w.eventRepo, &event.Event{
		EventID: e.EventID, EventType: event.EventTypeBroadcast,
		TownID: e.TownID, ActorType: e.ActorType, ActorID: e.ActorID,
		Payload: e.Payload, CreatedAt: e.CreatedAt,
	})

	w.appLog.Info("事件已广播", "event_type", e.EventType, "npc_name", npcName, "from", fromName, "to", toName)
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

func (w *BroadcastWorker) handleNPCIdleAction(_ context.Context, e *event.Event) error {
	var payload struct {
		NPCID   uint   `json:"npc_id"`
		NPCName string `json:"npc_name"`
		Action  string `json:"action"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse npc.idle.action payload: %w", err)
	}

	w.bcast.Push(e.EventType, map[string]any{
		"npc_id":   payload.NPCID,
		"npc_name": payload.NPCName,
		"action":   payload.Action,
	})

	w.appLog.Info("NPC idle action broadcast",
		"npc_name", payload.NPCName,
		"action", payload.Action,
	)
	return nil
}

func (w *BroadcastWorker) handleActivityGenerated(_ context.Context, e *event.Event) error {
	var payload struct {
		NPCID   uint   `json:"npc_id"`
		NPCName string `json:"npc_name"`
		Action  string `json:"action"`
		Mood    string `json:"mood"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse npc.activity.generated payload: %w", err)
	}

	w.bcast.Push(e.EventType, map[string]any{
		"npc_id":   payload.NPCID,
		"npc_name": payload.NPCName,
		"action":   payload.Action,
		"mood":     payload.Mood,
	})

	w.appLog.Info("activity broadcast", "npc_name", payload.NPCName, "action", payload.Action)
	return nil
}

func (w *BroadcastWorker) handleMoodChanged(_ context.Context, e *event.Event) error {
	var payload struct {
		NPCID   uint   `json:"npc_id"`
		NPCName string `json:"npc_name"`
		OldMood string `json:"old_mood"`
		NewMood string `json:"new_mood"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse npc.mood.changed payload: %w", err)
	}

	w.bcast.Push(e.EventType, map[string]any{
		"npc_id":   payload.NPCID,
		"npc_name": payload.NPCName,
		"old_mood": payload.OldMood,
		"new_mood": payload.NewMood,
	})
	return nil
}

func (w *BroadcastWorker) handleInteractionGenerated(_ context.Context, e *event.Event) error {
	w.bcast.Push(e.EventType, e.Payload)
	w.appLog.Info("interaction broadcast", "event_id", e.EventID)
	return nil
}

func (w *BroadcastWorker) handleStoryEvent(_ context.Context, e *event.Event) error {
	w.bcast.Push(e.EventType, e.Payload)
	w.appLog.Info("story event broadcast", "event_id", e.EventID)
	return nil
}

func (w *BroadcastWorker) handleGoalChanged(_ context.Context, e *event.Event) error {
	var payload struct {
		NPCID   uint   `json:"npc_id"`
		NPCName string `json:"npc_name"`
		OldGoal string `json:"old_goal"`
		NewGoal string `json:"new_goal"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse npc.goal.changed payload: %w", err)
	}
	w.bcast.Push(e.EventType, map[string]any{
		"npc_id":   payload.NPCID,
		"npc_name": payload.NPCName,
		"old_goal": payload.OldGoal,
		"new_goal": payload.NewGoal,
		"reason":   payload.Reason,
	})
	return nil
}

func (w *BroadcastWorker) handleTownNews(_ context.Context, e *event.Event) error {
	var payload struct {
		StoryTitle string `json:"story_title"`
		Message    string `json:"message"`
	}
	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		return fmt.Errorf("parse town.news.generated payload: %w", err)
	}
	w.bcast.Push(e.EventType, map[string]any{
		"story_title": payload.StoryTitle,
		"message":     payload.Message,
	})
	w.appLog.Info("town news broadcast", "title", payload.StoryTitle)
	return nil
}
