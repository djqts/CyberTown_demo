package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/event"
	"backend/internal/logger"
)

// NPCWorker 处理 NPC 移动相关的级联事件。
type NPCWorker struct {
	npcSvc    npcMoveService
	publisher eventPublisher
	eventRepo eventLogCreator
	appLog    *logger.AppLogger
}

// NewNPCWorker 创建 NPC Worker。
func NewNPCWorker(
	npcSvc npcMoveService,
	publisher eventPublisher,
	eventRepo eventLogCreator,
	appLog *logger.AppLogger,
) *NPCWorker {
	return &NPCWorker{
		npcSvc:    npcSvc,
		publisher: publisher,
		eventRepo: eventRepo,
		appLog:    appLog,
	}
}

// HandleTownTick 处理 town.tick → 查询日程 → 发布 npc.move.requested。
func (w *NPCWorker) HandleTownTick(ctx context.Context, e *event.Event) error {
	var tick struct {
		Minute int `json:"minute"`
	}
	if err := json.Unmarshal(e.Payload, &tick); err != nil {
		return fmt.Errorf("parse tick payload: %w", err)
	}

	moves, err := w.npcSvc.FindActiveMoves(e.TownID, tick.Minute)
	if err != nil {
		return fmt.Errorf("find active moves: %w", err)
	}

	for _, m := range moves {
		payload, _ := json.Marshal(map[string]any{
			"npc_id":           m.NPC.ID,
			"npc_name":         m.NPC.Name,
			"from_location_id": m.FromLocation,
			"to_location_id":   m.ToLocation,
			"schedule_id":      m.Schedule.ID,
		})

		me := &event.Event{
			EventID:   newEventID(),
			EventType: event.EventTypeNPCMoveRequest,
			TownID:    e.TownID,
			ActorType: event.ActorTypeNPC,
			ActorID:   fmt.Sprintf("npc_%d", m.NPC.ID),
			Payload:   payload,
			CreatedAt: e.CreatedAt,
		}

		if err := w.publisher.Publish(ctx, me); err != nil {
			w.appLog.Error(err, "发布 npc.move.requested 失败", "npc_id", m.NPC.ID)
		}
	}
	return nil
}

// HandleMoveRequested 处理 npc.move.requested → 移动 NPC → 发布 npc.moved。
func (w *NPCWorker) HandleMoveRequested(ctx context.Context, e *event.Event) error {
	var move struct {
		NPCID          uint `json:"npc_id"`
		FromLocationID uint `json:"from_location_id"`
		ToLocationID   uint `json:"to_location_id"`
	}
	if err := json.Unmarshal(e.Payload, &move); err != nil {
		return fmt.Errorf("parse move requested payload: %w", err)
	}

	if err := w.npcSvc.MoveNPC(move.NPCID, move.ToLocationID); err != nil {
		return fmt.Errorf("move npc: %w", err)
	}

	w.appLog.Info("NPC 已移动", "npc_id", move.NPCID, "to_location_id", move.ToLocationID)

	// 重新查找 NPC 以获取最新状态和位置
	// 发布 npc.moved
	payload, _ := json.Marshal(map[string]any{
		"npc_id":           move.NPCID,
		"from_location_id": move.FromLocationID,
		"to_location_id":   move.ToLocationID,
	})

	me := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeNPCMoved,
		TownID:    e.TownID,
		ActorType: event.ActorTypeNPC,
		ActorID:   e.ActorID,
		Payload:   payload,
		CreatedAt: e.CreatedAt,
	}

	return w.publisher.Publish(ctx, me)
}

// HandleMoved 处理 npc.moved → 写入 event_logs。
func (w *NPCWorker) HandleMoved(ctx context.Context, e *event.Event) error {
	return writeEventLog(w.eventRepo, e)
}
