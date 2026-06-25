package worker

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"backend/internal/event"
	ws "backend/internal/gateway/websocket"
	"backend/internal/logger"
)

type fakeConsumer struct {
	queue string
}

func (c *fakeConsumer) Consume(_ context.Context, queue string, _ event.Handler) error {
	c.queue = queue
	return nil
}

type fakeAgentReplier struct {
	npcID     uint
	userMsg   string
	userToken string
	reply     string
}

func (a *fakeAgentReplier) GenerateReply(_ context.Context, npcID uint, userMsg, userToken string) (string, string, error) {
	a.npcID = npcID
	a.userMsg = userMsg
	a.userToken = userToken
	return "TestNPC", a.reply, nil
}

func TestAgentWorkerPublishesNPCRepliedEvent(t *testing.T) {
	ctx := context.Background()
	agentSvc := &fakeAgentReplier{reply: "The square is lively today."}
	publisher := &fakePublisher{}
	worker := NewAgentWorker(&fakeConsumer{}, agentSvc, publisher, logger.NewApp("error", false))

	payload, err := json.Marshal(ws.UserMessage{
		NPCID:     3,
		Content:   "Any news in town?",
		UserToken: "player1",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	err = worker.handleEvent(ctx, &event.Event{
		EventID:   "user-message-1",
		EventType: event.EventTypeUserMessageSent,
		TownID:    1,
		ActorType: event.ActorTypeUser,
		ActorID:   "player1",
		Payload:   payload,
		CreatedAt: time.Unix(123, 0),
	})
	if err != nil {
		t.Fatalf("handleEvent returned error: %v", err)
	}

	if agentSvc.npcID != 3 || agentSvc.userMsg != "Any news in town?" || agentSvc.userToken != "player1" {
		t.Fatalf("agent args = npc:%d msg:%q token:%q", agentSvc.npcID, agentSvc.userMsg, agentSvc.userToken)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("published %d events, want 1", len(publisher.events))
	}

	published := publisher.events[0]
	if published.EventType != event.EventTypeNPCReplied {
		t.Fatalf("event type %q, want %q", published.EventType, event.EventTypeNPCReplied)
	}

	var reply ws.NPCReplied
	if err := json.Unmarshal(published.Payload, &reply); err != nil {
		t.Fatalf("unmarshal reply payload: %v", err)
	}
	if reply.NPCID != 3 || reply.UserToken != "player1" || reply.Content != "The square is lively today." {
		t.Fatalf("reply payload = %+v", reply)
	}
	if reply.NPCName != "TestNPC" {
		t.Fatalf("reply NPCName = %q, want %q", reply.NPCName, "TestNPC")
	}
}
