package worker

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"

	"backend/internal/event"
	"backend/internal/logger"
	"backend/internal/model"
)

// EventWorker 消费事件并分发到对应处理器。
type EventWorker struct {
	consumer  consumer
	eventRepo eventLogCreator
	npcWorker *NPCWorker
	appLog    *logger.AppLogger
}

// NewEventWorker 创建事件 Worker。
func NewEventWorker(
	consumer consumer,
	eventRepo eventLogCreator,
	npcWorker *NPCWorker,
	appLog *logger.AppLogger,
) *EventWorker {
	return &EventWorker{
		consumer:  consumer,
		eventRepo: eventRepo,
		npcWorker: npcWorker,
		appLog:    appLog,
	}
}

// Start 开始消费 town_events 队列中的事件。
func (w *EventWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, "town_events", w.handleEvent)
}

func (w *EventWorker) handleEvent(ctx context.Context, e *event.Event) error {
	switch e.EventType {
	case event.EventTypeTownTick:
		// 1) 写入 event_logs
		if err := writeEventLog(w.eventRepo, e); err != nil {
			return err
		}
		// 2) 触发 NPC 日程检查
		if w.npcWorker != nil {
			return w.npcWorker.HandleTownTick(ctx, e)
		}
		return nil

	case event.EventTypeNPCMoveRequest:
		if w.npcWorker != nil {
			return w.npcWorker.HandleMoveRequested(ctx, e)
		}
		return nil

	case event.EventTypeNPCMoved:
		if w.npcWorker != nil {
			return w.npcWorker.HandleMoved(ctx, e)
		}
		return nil

	default:
		w.appLog.Warn("未知事件类型", "event_type", e.EventType)
		return nil
	}
}

// ---- 共享工具 ----

// writeEventLog 将事件写入 event_logs 表。
func writeEventLog(eventRepo eventLogCreator, e *event.Event) error {
	payload, _ := json.Marshal(e.Payload)

	return eventRepo.Create(&model.EventLog{
		TownID:    e.TownID,
		EventType: e.EventType,
		Payload:   string(payload),
	})
}

// newEventID 生成简单的 UUID 格式标识。
func newEventID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
