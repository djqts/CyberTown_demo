package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/event"
	ws "backend/internal/gateway/websocket"
	"backend/internal/logger"
)

// AgentWorker consumes user messages and publishes NPC replies.
type AgentWorker struct {
	consumer  consumer
	agentSvc  agentReplier
	publisher eventPublisher
	appLog    *logger.AppLogger
}

func NewAgentWorker(
	consumer consumer,
	agentSvc agentReplier,
	publisher eventPublisher,
	appLog *logger.AppLogger,
) *AgentWorker {
	return &AgentWorker{
		consumer:  consumer,
		agentSvc:  agentSvc,
		publisher: publisher,
		appLog:    appLog,
	}
}

func (w *AgentWorker) Start(ctx context.Context) error {
	return w.consumer.Consume(ctx, "user_events", w.handleEvent)
}

func (w *AgentWorker) handleEvent(ctx context.Context, e *event.Event) error {
	switch e.EventType {
	case event.EventTypeUserMessageSent:
		return w.handleUserMessageSent(ctx, e)
	default:
		return nil
	}
}

func (w *AgentWorker) handleUserMessageSent(ctx context.Context, e *event.Event) error {
	var userMsg ws.UserMessage
	if err := json.Unmarshal(e.Payload, &userMsg); err != nil {
		return fmt.Errorf("parse user.message.sent payload: %w", err)
	}

	w.appLog.Info("processing user message",
		"npc_id", userMsg.NPCID,
		"user_token", userMsg.UserToken,
	)

	reply, err := w.agentSvc.GenerateReply(ctx, userMsg.NPCID, userMsg.Content, userMsg.UserToken)
	if err != nil {
		w.appLog.Error(err, "agent reply generation failed", "npc_id", userMsg.NPCID)
		return fmt.Errorf("agent generate reply: %w", err)
	}

	payload, _ := json.Marshal(ws.NPCReplied{
		NPCID:     userMsg.NPCID,
		NPCName:   "",
		Content:   reply,
		UserToken: userMsg.UserToken,
	})

	npcEvent := &event.Event{
		EventID:   newEventID(),
		EventType: event.EventTypeNPCReplied,
		TownID:    e.TownID,
		ActorType: event.ActorTypeNPC,
		ActorID:   fmt.Sprintf("npc_%d", userMsg.NPCID),
		Payload:   payload,
		CreatedAt: e.CreatedAt,
	}

	if err := w.publisher.Publish(ctx, npcEvent); err != nil {
		return fmt.Errorf("publish npc.replied: %w", err)
	}

	w.appLog.Info("npc reply event published",
		"npc_id", userMsg.NPCID,
		"user_token", userMsg.UserToken,
	)
	return nil
}
