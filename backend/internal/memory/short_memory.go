package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	shortMemoryMaxLen = 10
	shortMemoryTTL    = 24 * time.Hour
)

// ShortMsg 短期记忆中的一条消息。
type ShortMsg struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"ts"`
}

// ShortMemory Redis 短期记忆。
type ShortMemory struct {
	client *redis.Client
}

// NewShortMemory 创建短期记忆存储。
func NewShortMemory(client *redis.Client) *ShortMemory {
	return &ShortMemory{client: client}
}

func shortKey(npcID uint, userToken string) string {
	return fmt.Sprintf("chat:short:%d:%s", npcID, userToken)
}

// Save 保存一条消息。
func (s *ShortMemory) Save(ctx context.Context, npcID uint, userToken, role, content string) error {
	msg := ShortMsg{
		Role:      role,
		Content:   content,
		Timestamp: time.Now().Unix(),
	}
	data, _ := json.Marshal(msg)

	key := shortKey(npcID, userToken)
	pipe := s.client.Pipeline()
	pipe.LPush(ctx, key, data)
	pipe.LTrim(ctx, key, 0, shortMemoryMaxLen-1)
	pipe.Expire(ctx, key, shortMemoryTTL)
	cmds, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	for _, cmd := range cmds {
		if cmd.Err() != nil {
			return cmd.Err()
		}
	}
	return nil
}

// GetRecent 获取最近 N 条消息（按时间正序）。
func (s *ShortMemory) GetRecent(ctx context.Context, npcID uint, userToken string, limit int) ([]ShortMsg, error) {
	key := shortKey(npcID, userToken)
	vals, err := s.client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	msgs := make([]ShortMsg, len(vals))
	for i, v := range vals {
		json.Unmarshal([]byte(v), &msgs[i])
	}

	// Redis LPUSH 是倒序，反转回正序
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
