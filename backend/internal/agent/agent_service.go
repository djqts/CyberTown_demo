package agent

import (
	"context"
	"fmt"

	"backend/internal/logger"
	"backend/internal/memory"
	"backend/internal/model"
	"backend/internal/repo"
)

// AgentService coordinates NPC profile, chat history, memory, and the LLM runner.
type AgentService struct {
	npcRepo    *repo.NPCRepo
	chatRepo   *repo.ChatRepo
	einoRunner *EinoRunner
	memSvc     *memory.Service
	appLog     *logger.AppLogger
}

func NewAgentService(
	npcRepo *repo.NPCRepo,
	chatRepo *repo.ChatRepo,
	einoRunner *EinoRunner,
	memSvc *memory.Service,
	appLog *logger.AppLogger,
) *AgentService {
	return &AgentService{
		npcRepo:    npcRepo,
		chatRepo:   chatRepo,
		einoRunner: einoRunner,
		memSvc:     memSvc,
		appLog:     appLog,
	}
}

func (s *AgentService) GenerateReply(ctx context.Context, npcID uint, userMsg, userToken string) (string, string, error) {
	npc, err := s.npcRepo.FindByID(npcID)
	if err != nil {
		return "", "", fmt.Errorf("find npc %d: %w", npcID, err)
	}

	history, err := s.chatRepo.FindByNPCAndUser(npcID, userToken, 20)
	if err != nil {
		return "", "", fmt.Errorf("find history: %w", err)
	}
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	simpleAgent := NewSimpleAgent(npc, s.einoRunner)
	reply, err := simpleAgent.GenerateReply(ctx, userMsg, userToken, history, "")
	if err != nil {
		return "", "", fmt.Errorf("generate reply: %w", err)
	}

	// Async: save chat history and memory — don't block the reply
	go func() {
		_ = s.chatRepo.Save(&model.ChatMessage{
			NPCID: npcID, UserToken: userToken, Role: "user", Content: userMsg,
		})
		_ = s.chatRepo.Save(&model.ChatMessage{
			NPCID: npcID, UserToken: userToken, Role: "npc", Content: reply,
		})
		if s.memSvc != nil {
			if err := s.memSvc.Memorize(context.Background(), npcID, userToken, userMsg, reply); err != nil && s.appLog != nil {
				s.appLog.Error(err, "async memory write failed", "npc_id", npcID)
			}
		}
	}()

	return npc.Name, reply, nil
}
