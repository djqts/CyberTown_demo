package memory

import (
	"context"
	"fmt"
	"strings"
)

// Service 记忆统一入口。
type Service struct {
	short   *ShortMemory
	long    *LongMemory
	embedFn func(string) []float32
}

// NewService 创建记忆服务。
func NewService(short *ShortMemory, long *LongMemory) *Service {
	return &Service{
		short:   short,
		long:    long,
		embedFn: Embed,
	}
}

// Recall 召回短期和长期记忆，返回拼装好的记忆文本。
func (s *Service) Recall(ctx context.Context, npcID uint, userToken, currentMsg string) string {
	var parts []string

	// 短期记忆
	shortMsgs, err := s.short.GetRecent(ctx, npcID, userToken, 10)
	if err == nil && len(shortMsgs) > 0 {
		var sb strings.Builder
		sb.WriteString("【近期对话】\n")
		for _, m := range shortMsgs {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", m.Role, m.Content))
		}
		parts = append(parts, sb.String())
	}

	// 长期记忆
	queryVec := s.embedFn(currentMsg)
	longMem, err := s.long.Search(ctx, npcID, queryVec, 3)
	if err == nil && len(longMem) > 0 {
		var sb strings.Builder
		sb.WriteString("【长期记忆】\n")
		for _, m := range longMem {
			sb.WriteString(fmt.Sprintf("  - %s\n", m))
		}
		parts = append(parts, sb.String())
	}

	// 世界知识
	worldK, err := s.long.SearchWorldKnowledge(ctx, queryVec, 2)
	if err == nil && len(worldK) > 0 {
		var sb strings.Builder
		sb.WriteString("【世界知识】\n")
		for _, k := range worldK {
			sb.WriteString(fmt.Sprintf("  - %s\n", k))
		}
		parts = append(parts, sb.String())
	}

	return strings.Join(parts, "\n")
}

// Memorize 保存对话到短期和长期记忆。
func (s *Service) Memorize(ctx context.Context, npcID uint, userToken, userMsg, npcReply string) error {
	// 短期记忆
	if err := s.short.Save(ctx, npcID, userToken, "user", userMsg); err != nil {
		return fmt.Errorf("short memory user save: %w", err)
	}
	if err := s.short.Save(ctx, npcID, userToken, "npc", npcReply); err != nil {
		return fmt.Errorf("short memory npc save: %w", err)
	}

	// 长期记忆（合并为一条）
	content := fmt.Sprintf("用户: %s | NPC: %s", userMsg, npcReply)
	vec := s.embedFn(content)
	if err := s.long.Save(ctx, npcID, userToken, content, vec); err != nil {
		return fmt.Errorf("long memory save: %w", err)
	}
	return nil
}
