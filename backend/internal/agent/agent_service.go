package agent

import (
	"context"
	"fmt"

	"backend/internal/memory"
	"backend/internal/model"
	"backend/internal/repo"
)

// AgentService 对外统一的 Agent 入口。
type AgentService struct {
	npcRepo    *repo.NPCRepo
	chatRepo   *repo.ChatRepo
	einoRunner *EinoRunner
	memSvc     *memory.Service
}

// NewAgentService 创建 Agent 服务。
func NewAgentService(npcRepo *repo.NPCRepo, chatRepo *repo.ChatRepo, einoRunner *EinoRunner, memSvc *memory.Service) *AgentService {
	return &AgentService{
		npcRepo:    npcRepo,
		chatRepo:   chatRepo,
		einoRunner: einoRunner,
		memSvc:     memSvc,
	}
}

// GenerateReply 为指定 NPC 生成对话回复。
func (s *AgentService) GenerateReply(ctx context.Context, npcID uint, userMsg, userToken string) (string, error) {
	// 1. 查询 NPC
	npc, err := s.npcRepo.FindByID(npcID)
	if err != nil {
		return "", fmt.Errorf("find npc %d: %w", npcID, err)
	}

	// 2. 获取对话历史（从 PostgreSQL）
	history, err := s.chatRepo.FindByNPCAndUser(npcID, userToken, 20)
	if err != nil {
		return "", fmt.Errorf("find history: %w", err)
	}
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	// 3. 召回记忆（短期 + 长期 + 世界知识）
	memoryContext := ""
	if s.memSvc != nil {
		memoryContext = s.memSvc.Recall(ctx, npcID, userToken, userMsg)
	}

	// 4. 构建 SimpleAgent 并生成回复
	agent := NewSimpleAgent(npc, s.einoRunner)
	reply, err := agent.GenerateReply(ctx, userMsg, history, memoryContext)
	if err != nil {
		return "", fmt.Errorf("generate reply: %w", err)
	}

	// 5. 保存到 PostgreSQL
	if err := s.chatRepo.Save(&model.ChatMessage{
		NPCID:     npcID,
		UserToken: userToken,
		Role:      "user",
		Content:   userMsg,
	}); err != nil {
		return "", fmt.Errorf("save user message: %w", err)
	}
	if err := s.chatRepo.Save(&model.ChatMessage{
		NPCID:     npcID,
		UserToken: userToken,
		Role:      "npc",
		Content:   reply,
	}); err != nil {
		return "", fmt.Errorf("save npc reply: %w", err)
	}

	// 6. 写入记忆系统
	if s.memSvc != nil {
		if err := s.memSvc.Memorize(ctx, npcID, userToken, userMsg, reply); err != nil {
			return "", fmt.Errorf("memory write: %w", err)
		}
	}

	return reply, nil
}
