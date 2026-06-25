package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"backend/internal/event"
	"backend/internal/logger"
	"backend/internal/model"
	"backend/internal/service"
)

type fakeNPCMoveService struct {
	movedNPCID uint
	movedToID  uint
}

func (s *fakeNPCMoveService) FindActiveMoves(uint, int) ([]service.NPCMove, error) {
	return nil, nil
}

func (s *fakeNPCMoveService) MoveNPC(npcID, newLocationID uint) error {
	s.movedNPCID = npcID
	s.movedToID = newLocationID
	return nil
}

func (s *fakeNPCMoveService) FindByTownID(uint) ([]model.NPC, error) {
	return nil, nil
}

type fakePublisher struct {
	events []*event.Event
}

func (p *fakePublisher) Publish(_ context.Context, e *event.Event) error {
	p.events = append(p.events, e)
	return nil
}

type fakeEventLogRepo struct{}

func (r *fakeEventLogRepo) Create(*model.EventLog) error {
	return nil
}

func TestNPCWorkerHandleMoveRequestedPublishesMovedWithFromAndTo(t *testing.T) {
	ctx := context.Background()
	npcSvc := &fakeNPCMoveService{}
	publisher := &fakePublisher{}
	worker := NewNPCWorker(npcSvc, publisher, &fakeEventLogRepo{}, nil, logger.NewApp("error", false))

	payload, err := json.Marshal(map[string]any{
		"npc_id":           7,
		"from_location_id": 2,
		"to_location_id":   3,
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	err = worker.HandleMoveRequested(ctx, &event.Event{
		EventID:   "move-requested-1",
		EventType: event.EventTypeNPCMoveRequest,
		TownID:    1,
		ActorType: event.ActorTypeNPC,
		ActorID:   "npc_7",
		Payload:   payload,
		CreatedAt: time.Unix(123, 0),
	})
	if err != nil {
		t.Fatalf("HandleMoveRequested returned error: %v", err)
	}

	if npcSvc.movedNPCID != 7 || npcSvc.movedToID != 3 {
		t.Fatalf("MoveNPC called with npc=%d to=%d, want npc=7 to=3", npcSvc.movedNPCID, npcSvc.movedToID)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published %d events, want 1", len(publisher.events))
	}

	published := publisher.events[0]
	if published.EventType != event.EventTypeNPCMoved {
		t.Fatalf("published event type %q, want %q", published.EventType, event.EventTypeNPCMoved)
	}

	var movedPayload struct {
		NPCID          uint `json:"npc_id"`
		FromLocationID uint `json:"from_location_id"`
		ToLocationID   uint `json:"to_location_id"`
	}
	if err := json.Unmarshal(published.Payload, &movedPayload); err != nil {
		t.Fatalf("unmarshal moved payload: %v", err)
	}
	if movedPayload.NPCID != 7 || movedPayload.FromLocationID != 2 || movedPayload.ToLocationID != 3 {
		t.Fatalf("moved payload = %+v, want npc=7 from=2 to=3", movedPayload)
	}
}
