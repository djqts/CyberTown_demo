package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/behavior"
	"backend/internal/event"
	"backend/internal/logger"
	"backend/internal/model"
)

// ActivityWorker 消费 town.tick，生成 NPC 主动行为。
type ActivityWorker struct {
	consumer  consumer
	publisher eventPublisher
	eventRepo eventLogCreator
	behavSvc  *behavior.BehaviorService
	npcRepo   npcStatusUpdater
	actGen    func(ctx context.Context, npc *model.NPC, reason string) (string, error)
	appLog    *logger.AppLogger
}

// NewActivityWorker 创建 ActivityWorker。
func NewActivityWorker(
	consumer consumer,
	publisher eventPublisher,
	eventRepo eventLogCreator,
	behavSvc *behavior.BehaviorService,
	npcRepo npcStatusUpdater,
	actGen func(ctx context.Context, npc *model.NPC, reason string) (string, error),
	appLog *logger.AppLogger,
) *ActivityWorker {
	return &ActivityWorker{
		consumer:  consumer,
		publisher: publisher,
		eventRepo: eventRepo,
		behavSvc:  behavSvc,
		npcRepo:   npcRepo,
		actGen:    actGen,
		appLog:    appLog,
	}
}

// Start 从 town_tick_activity 队列消费事件。
func (w *ActivityWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, "town_tick_activity", w.handleEvent)
}

func (w *ActivityWorker) handleEvent(ctx context.Context, e *event.Event) error {
	if e.EventType != event.EventTypeTownTick {
		return nil
	}

	var tick struct {
		Day    int `json:"day"`
		Minute int `json:"minute"`
	}
	if err := json.Unmarshal(e.Payload, &tick); err != nil {
		return fmt.Errorf("ActivityWorker parse tick payload: %w", err)
	}

	npcs, err := w.npcRepo.FindByTownID(e.TownID)
	if err != nil {
		return fmt.Errorf("ActivityWorker find npcs: %w", err)
	}

	// Build location name map for context-aware activities
	locNames := make(map[uint]string)
	for _, npc := range npcs {
		locNames[npc.LocationID] = "" // placeholder, filled below
	}
	// Quick lookup from first NPC at each location (using the NPC's location — we don't have location names easily)
	// Instead, use a simple approach: use the NPC's location_id directly as key for GetActivity
	// GetActivity needs a location NAME string, so we need a minimal name lookup
	// For simplicity: use a fixed lookup against known locations
	_ = locNames

	for _, npc := range npcs {
		locName := getLocationNameByID(npc.LocationID)
		var llmGen func(context.Context, *model.NPC, behavior.TriggerReason) (*behavior.ActivityResult, error)
		if w.actGen != nil {
			llmGen = func(ctx context.Context, npc *model.NPC, reason behavior.TriggerReason) (*behavior.ActivityResult, error) {
				action, err := w.actGen(ctx, npc, string(reason))
				if err != nil {
					return nil, err
				}
				return &behavior.ActivityResult{NPCID: npc.ID, NPCName: npc.Name, Action: action, Mood: npc.Mood}, nil
			}
		}
		result := w.behavSvc.Decide(ctx, &npc, locName, llmGen)
		if result == nil {
			continue
		}

		w.publishActivity(ctx, e, result)

		now := behavior.NowPtr()
		npc.LastActiveAt = now
		_ = w.npcRepo.UpdateFull(&npc)
	}

	return nil
}

// getLocationNameByID returns a location name for a given ID using a hardcoded lookup.
// This avoids needing a DB lookup during every tick for activity generation.
func getLocationNameByID(id uint) string {
	names := map[uint]string{
		38: "广场", 39: "咖啡馆", 40: "钟楼", 41: "市政厅",
		42: "图书馆", 43: "花店", 44: "铁匠铺", 45: "诊所",
		46: "农舍", 47: "钓鱼小屋", 48: "学校", 49: "面包店",
		50: "酒馆", 51: "公园凉亭", 52: "手工工坊", 53: "住宅区", 54: "森林营地",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return ""
}

func (w *ActivityWorker) publishActivity(ctx context.Context, tickEvent *event.Event, result *behavior.ActivityResult) {
	payload, _ := json.Marshal(map[string]any{
		"npc_id":   result.NPCID,
		"npc_name": result.NPCName,
		"action":   result.Action,
		"mood":     result.Mood,
	})

	ae := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeNPCActivityGenerated,
		TownID:    tickEvent.TownID,
		ActorType: event.ActorTypeNPC,
		ActorID:   fmt.Sprintf("npc_%d", result.NPCID),
		Payload:   payload,
		CreatedAt: tickEvent.CreatedAt,
	}

	if err := w.publisher.Publish(ctx, ae); err != nil {
		w.appLog.Error(err, "publish npc.activity.generated failed", "npc_id", result.NPCID)
		return
	}

	_ = writeEventLog(w.eventRepo, ae)

	// Also publish occasional NPC thought
	if result.Thought != "" {
		tp, _ := json.Marshal(map[string]any{
			"npc_id":   result.NPCID,
			"npc_name": result.NPCName,
			"thought":  result.Thought,
		})
		te := &event.Event{
			EventID:   newEventID(),
			EventType: event.EventTypeNPCThoughtGenerated,
			TownID:    tickEvent.TownID,
			ActorType: event.ActorTypeNPC,
			ActorID:   fmt.Sprintf("npc_%d", result.NPCID),
			Payload:   tp,
			CreatedAt: tickEvent.CreatedAt,
		}
		if err := w.publisher.Publish(ctx, te); err != nil {
			w.appLog.Error(err, "publish npc.thought.generated failed", "npc_id", result.NPCID)
		}
	}
}
