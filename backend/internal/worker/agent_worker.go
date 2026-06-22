package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/agent"
	"backend/internal/event"
	ws "backend/internal/gateway/websocket"
	"backend/internal/logger"
)

// AgentWorker 消费 user.message.sent 事件，调用 Agent 生成 NPC 回复。
type AgentWorker struct {
	consumer  consumer
	agentSvc  agentReplier
	publisher eventPublisher
	hub       userMessenger
	appLog    *logger.AppLogger
}

// NewAgentWorker 创建 Agent Worker。
func NewAgentWorker(
	consumer *event.Consumer,
	agentSvc *agent.AgentService,
	publisher *event.Publisher,
	hub *ws.Hub,
	appLog *logger.AppLogger,
) *AgentWorker {
	return &AgentWorker{
		consumer:  consumer,
		agentSvc:  agentSvc,
		publisher: publisher,
		hub:       hub,
		appLog:    appLog,
	}
}

// Start 从 user_events 队列消费事件。
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

	w.appLog.Info("处理用户消息",
		"npc_id", userMsg.NPCID,
		"user_token", userMsg.UserToken,
	)

	// 调用 Agent 生成回复
	reply, err := w.agentSvc.GenerateReply(ctx, userMsg.NPCID, userMsg.Content, userMsg.UserToken)
	if err != nil {
		w.appLog.Error(err, "Agent 生成回复失败", "npc_id", userMsg.NPCID)
		return fmt.Errorf("agent generate reply: %w", err)
	}

	// 发布 npc.replied 事件
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

	// 同时直接推送给 WebSocket 用户
	w.hub.SendToUser(userMsg.UserToken, &ws.Message{
		Type: ws.MsgTypeNPCReplied,
		Data: ws.NPCReplied{
			NPCID:     userMsg.NPCID,
			Content:   reply,
			UserToken: userMsg.UserToken,
		},
	})

	w.appLog.Info("NPC 回复已推送",
		"npc_id", userMsg.NPCID,
		"user_token", userMsg.UserToken,
	)
	return nil
}
