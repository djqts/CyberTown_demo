package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"

	"backend/internal/event"
	"backend/internal/logger"
)

// roleActionPool maps NPC roles to idle actions they can perform.
var roleActionPool = map[string][]string{
	"咖啡师": {"擦拭吧台", "冲洗咖啡杯", "整理咖啡豆", "跟路过的顾客打招呼"},
	"钟表匠": {"放大镜检查齿轮", "校准钟摆", "擦拭钟面玻璃", "记录时间误差"},
	"邮差":  {"整理信件", "查看地址簿", "系紧邮包带子", "收拢信袋"},
}

// npcMoveKey tracks an NPC's last scheduled move for a given day.
type npcMoveKey struct {
	npcID  uint
	day    int
	fromID uint
	toID   uint
}

// NPCWorker handles NPC movement and idle actions.
type NPCWorker struct {
	npcSvc    npcMoveService
	publisher eventPublisher
	eventRepo eventLogCreator
	appLog    *logger.AppLogger
	onMoved   func(npcID uint) // callback for interaction tracking

	mu             sync.Mutex
	currentDay     int
	moved          map[npcMoveKey]bool
	lastIdleMinute map[uint]int // npcID -> last game-minute idle action was emitted
}

// NewNPCWorker creates a new NPCWorker.
func NewNPCWorker(
	npcSvc npcMoveService,
	publisher eventPublisher,
	eventRepo eventLogCreator,
	onMoved func(npcID uint),
	appLog *logger.AppLogger,
) *NPCWorker {
	return &NPCWorker{
		npcSvc:         npcSvc,
		publisher:      publisher,
		eventRepo:      eventRepo,
		onMoved:        onMoved,
		appLog:         appLog,
		currentDay:     -1,
		moved:          make(map[npcMoveKey]bool),
		lastIdleMinute: make(map[uint]int),
	}
}

// HandleTownTick processes town.tick: finds active moves, publishes npc.move.requested,
// and emits npc.idle.action for NPCs that are at their correct location.
func (w *NPCWorker) HandleTownTick(ctx context.Context, e *event.Event) error {
	var tick struct {
		Day    int `json:"day"`
		Minute int `json:"minute"`
	}
	if err := json.Unmarshal(e.Payload, &tick); err != nil {
		return fmt.Errorf("parse tick payload: %w", err)
	}

	w.mu.Lock()
	if tick.Day != w.currentDay {
		w.moved = make(map[npcMoveKey]bool)
		w.currentDay = tick.Day
	}
	w.mu.Unlock()

	moves, err := w.npcSvc.FindActiveMoves(e.TownID, tick.Minute)
	if err != nil {
		return fmt.Errorf("find active moves: %w", err)
	}

	// Deduplicate moves within the same tick
	seenThisTick := make(map[uint]bool)

	for _, m := range moves {
		if seenThisTick[m.NPC.ID] {
			continue
		}
		seenThisTick[m.NPC.ID] = true

		// Skip if this exact route was already published today
		key := npcMoveKey{
			npcID:  m.NPC.ID,
			day:    tick.Day,
			fromID: m.FromLocation,
			toID:   m.ToLocation,
		}
		w.mu.Lock()
		if w.moved[key] {
			w.mu.Unlock()
			continue
		}
		w.moved[key] = true
		w.mu.Unlock()

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

		// Process move synchronously — ensures correct event order: goal → move → activity
		if err := w.HandleMoveRequested(ctx, me); err != nil {
			w.appLog.Error(err, "process npc.move.requested failed", "npc_id", m.NPC.ID)
		}
		// Don't re-publish to RabbitMQ — already handled synchronously above
	}

	// Emit idle actions for NPCs at rest
	allNPCs, err := w.npcSvc.FindByTownID(e.TownID)
	if err != nil {
		w.appLog.Error(err, "find town npcs for idle action")
		return nil // non-fatal; moves already processed
	}

	for _, npc := range allNPCs {
		if seenThisTick[npc.ID] {
			continue // NPC is already moving this tick
		}

		w.mu.Lock()
		lastMinute := w.lastIdleMinute[npc.ID]
		w.mu.Unlock()

		// Rate-limit: at least 30 game-minutes between idle actions per NPC
		if tick.Minute-lastMinute < 30 && lastMinute > 0 {
			continue
		}

		// ~20% chance per resting NPC per tick
		if rand.Intn(100) >= 20 {
			continue
		}

		w.mu.Lock()
		w.lastIdleMinute[npc.ID] = tick.Minute
		w.mu.Unlock()

		w.emitIdleAction(ctx, e, npc.ID, npc.Name, npc.Role, npc.Status)
	}

	return nil
}

// emitIdleAction picks a random idle action for the NPC and publishes an npc.idle.action event.
func (w *NPCWorker) emitIdleAction(ctx context.Context, tickEvent *event.Event, npcID uint, npcName, role, scheduleAction string) {
	actions := roleActionPool[role]
	if len(actions) == 0 {
		actions = []string{scheduleAction}
	}
	action := actions[rand.Intn(len(actions))]

	payload, _ := json.Marshal(map[string]any{
		"npc_id":   npcID,
		"npc_name": npcName,
		"action":   action,
	})

	ie := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeNPCIdleAction,
		TownID:    tickEvent.TownID,
		ActorType: event.ActorTypeNPC,
		ActorID:   fmt.Sprintf("npc_%d", npcID),
		Payload:   payload,
		CreatedAt: tickEvent.CreatedAt,
	}

	if err := w.publisher.Publish(ctx, ie); err != nil {
		w.appLog.Error(err, "publish npc.idle.action failed", "npc_id", npcID)
	}
}

// HandleMoveRequested processes npc.move.requested: moves the NPC, publishes npc.moved.
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

	w.appLog.Info("NPC moved", "npc_id", move.NPCID, "to_location_id", move.ToLocationID)

	// Notify interaction system that this NPC just arrived
	if w.onMoved != nil {
		w.onMoved(move.NPCID)
	}

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

// HandleMoved processes npc.moved: writes to event_logs.
func (w *NPCWorker) HandleMoved(ctx context.Context, e *event.Event) error {
	return writeEventLog(w.eventRepo, e)
}
