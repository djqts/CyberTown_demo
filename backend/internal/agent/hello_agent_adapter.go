package agent

import (
	"context"
	"fmt"

	"backend/internal/model"
)

// SimpleAgent 基于 Eino 的 NPC 对话 Agent。
type SimpleAgent struct {
	npc        *model.NPC
	einoRunner *EinoRunner
}

// NewSimpleAgent 创建一个绑定到特定 NPC 的 SimpleAgent。
func NewSimpleAgent(npc *model.NPC, einoRunner *EinoRunner) *SimpleAgent {
	return &SimpleAgent{
		npc:        npc,
		einoRunner: einoRunner,
	}
}

// GenerateReply 生成 NPC 回复。history 为对话历史，userMsg 为当前用户输入。
func (a *SimpleAgent) GenerateReply(ctx context.Context, userMsg string, history []model.ChatMessage, memoryContext string) (string, error) {
	messages := BuildMessages(a.npc, userMsg, history, memoryContext)

	resp, err := a.einoRunner.Run(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("eino generate: %w", err)
	}

	if resp == nil {
		return "", fmt.Errorf("模型返回空回复")
	}

	return resp.Content, nil
}
