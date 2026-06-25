package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/event"
	"backend/internal/interaction"
	"backend/internal/logger"
	"backend/internal/model"
)

type InteractionWorker struct {
	consumer  consumer
	publisher eventPublisher
	eventRepo eventLogCreator
	interSvc  *interaction.Service
	interGen  func(ctx context.Context, a, b *model.NPC) (*interaction.InteractionResult, error)
	appLog    *logger.AppLogger
}

func NewInteractionWorker(
	consumer consumer,
	publisher eventPublisher,
	eventRepo eventLogCreator,
	interSvc *interaction.Service,
	interGen func(ctx context.Context, a, b *model.NPC) (*interaction.InteractionResult, error),
	appLog *logger.AppLogger,
) *InteractionWorker {
	return &InteractionWorker{
		consumer:  consumer,
		publisher: publisher,
		eventRepo: eventRepo,
		interSvc:  interSvc,
		interGen:  interGen,
		appLog:    appLog,
	}
}

func (w *InteractionWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, "town_tick_interaction", w.handleEvent)
}

func (w *InteractionWorker) handleEvent(ctx context.Context, e *event.Event) error {
	if e.EventType != event.EventTypeTownTick {
		return nil
	}

	pairs, err := w.interSvc.FindInteractions(e.TownID)
	if err != nil {
		return fmt.Errorf("find interactions: %w", err)
	}

	for _, pair := range pairs {
		result := w.interSvc.GenerateInteraction(ctx, &pair.A, &pair.B, w.interGen)
		if result == nil {
			continue
		}
		w.publishInteraction(ctx, e, &pair.A, &pair.B, result)
		w.publishRelationshipChanges(ctx, e, &pair.A, &pair.B, result)
		// Social propagation: spread gossip from this interaction
		if len(result.Dialogue) > 0 {
			summary := result.Dialogue[0].Speech
			if len(summary) > 60 {
				summary = summary[:60]
			}
			w.interSvc.SpreadGossip(ctx, &pair.A, &pair.B, summary, pair.A.LocationID)
		}
		w.interSvc.MarkDone(pair.A.ID, pair.B.ID)
	}

	return nil
}

func (w *InteractionWorker) publishInteraction(
	ctx context.Context,
	tickEvent *event.Event,
	a, b *model.NPC,
	result *interaction.InteractionResult,
) {
	payload, _ := json.Marshal(map[string]any{
		"npc_a": map[string]any{
			"id":   a.ID,
			"name": a.Name,
		},
		"npc_b": map[string]any{
			"id":   b.ID,
			"name": b.Name,
		},
		"dialogue":     result.Dialogue,
		"mood_changes": result.MoodChanges,
		"rel_deltas":   result.RelDeltas,
	})

	ie := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeNPCInteractionGenerated,
		TownID:    tickEvent.TownID,
		ActorType: event.ActorTypeNPC,
		ActorID:   fmt.Sprintf("npc_%d", a.ID),
		Payload:   payload,
		CreatedAt: tickEvent.CreatedAt,
	}

	if err := w.publisher.Publish(ctx, ie); err != nil {
		w.appLog.Error(err, "publish npc.interaction.generated failed")
		return
	}
	_ = writeEventLog(w.eventRepo, ie)
}

func (w *InteractionWorker) publishRelationshipChanges(
	ctx context.Context,
	tickEvent *event.Event,
	a, b *model.NPC,
	result *interaction.InteractionResult,
) {
	for _, rd := range result.RelDeltas {
		relPayload, _ := json.Marshal(map[string]any{
			"from_npc_id":   rd.FromNPCID,
			"from_npc_name": getNPCName(a, b, rd.FromNPCID),
			"to_npc_id":     rd.ToNPCID,
			"to_npc_name":   getNPCName(a, b, rd.ToNPCID),
			"delta":         rd.Delta,
			"reason":        rd.Reason,
		})
		re := &event.Event{
			EventID:   newEventID(),
			EventType: event.EventTypeNPCRelationshipChanged,
			TownID:    tickEvent.TownID,
			ActorType: event.ActorTypeNPC,
			ActorID:   fmt.Sprintf("npc_%d", rd.FromNPCID),
			Payload:   relPayload,
			CreatedAt: tickEvent.CreatedAt,
		}
		if err := w.publisher.Publish(ctx, re); err != nil {
			w.appLog.Error(err, "publish npc.relationship.changed failed")
		}
		_ = writeEventLog(w.eventRepo, re)
	}
}

func getNPCName(a, b *model.NPC, id uint) string {
	if id == a.ID {
		return a.Name
	}
	if id == b.ID {
		return b.Name
	}
	return ""
}
