package interaction

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/internal/model"
)

// SocialMemory 一条社交传播的记忆——NPC之间的"八卦"
type SocialMemory struct {
	Content    string    `json:"content"`
	FromNPCID  uint      `json:"from_npc_id"`
	FromName   string    `json:"from_name"`
	ToNPCID    uint      `json:"to_npc_id"`
	ToName     string    `json:"to_name"`
	LocationID uint      `json:"location_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// SocialPropagation 社交传播引擎——NPC之间的信息流动
// 模拟"镇上的八卦"：两个NPC的对话可以被其他NPC听到和传播
type SocialPropagation struct {
	mu       sync.RWMutex
	memories []SocialMemory
	maxSize  int
}

func NewSocialPropagation() *SocialPropagation {
	return &SocialPropagation{
		memories: make([]SocialMemory, 0, 200),
		maxSize:  200,
	}
}

// Spread 记录一次社交互动，使内容可供其他NPC"听到"
func (s *SocialPropagation) Spread(ctx context.Context, fromNPC, toNPC *model.NPC, content string, locationID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.memories = append(s.memories, SocialMemory{
		Content:    content,
		FromNPCID:  fromNPC.ID,
		FromName:   fromNPC.Name,
		ToNPCID:    toNPC.ID,
		ToName:     toNPC.Name,
		LocationID: locationID,
		CreatedAt:  time.Now(),
	})

	if len(s.memories) > s.maxSize {
		s.memories = s.memories[len(s.memories)-s.maxSize:]
	}
}

// HearGossip 一个NPC"听到"附近最近的社交信息
// 返回符合该NPC视角的社交上下文
func (s *SocialPropagation) HearGossip(npcID uint, locationID uint, limit int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var heard []string
	count := 0
	for i := len(s.memories) - 1; i >= 0 && count < limit; i-- {
		m := s.memories[i]
		// NPC能听到：同地点的对话，或者关于自己的对话
		if m.LocationID == locationID || m.ToNPCID == npcID || m.FromNPCID == npcID {
			msg := fmt.Sprintf("%s和%s聊到：%s", m.FromName, m.ToName, m.Content)
			heard = append(heard, msg)
			count++
		}
	}

	if len(heard) == 0 {
		return ""
	}

	result := "【小镇传闻】\n"
	for _, h := range heard {
		result += "  - " + h + "\n"
	}
	return result
}

// GetRecentGossip 获取最近的社交动态（供前端展示）
func (s *SocialPropagation) GetRecentGossip(limit int) []SocialMemory {
	s.mu.RLock()
	defer s.mu.RUnlock()

	start := len(s.memories) - limit
	if start < 0 {
		start = 0
	}
	result := make([]SocialMemory, len(s.memories)-start)
	copy(result, s.memories[start:])
	return result
}
