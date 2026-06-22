package agent

import (
	"context"
	"fmt"

	"backend/internal/model"
)

type SimpleAgent struct {
	npc        *model.NPC
	einoRunner *EinoRunner
}

func NewSimpleAgent(npc *model.NPC, einoRunner *EinoRunner) *SimpleAgent {
	return &SimpleAgent{
		npc:        npc,
		einoRunner: einoRunner,
	}
}

func (a *SimpleAgent) GenerateReply(ctx context.Context, userMsg, userToken string, history []model.ChatMessage, memoryContext string) (string, error) {
	resp, err := a.einoRunner.Run(ctx, PromptInput{
		NPC:           a.npc,
		UserToken:     userToken,
		UserInput:     userMsg,
		History:       history,
		MemoryContext: memoryContext,
	})
	if err != nil {
		return "", fmt.Errorf("eino generate: %w", err)
	}
	if resp == nil || resp.Content == "" {
		return "", fmt.Errorf("empty model response")
	}
	return resp.Content, nil
}
