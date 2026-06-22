package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type ChatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) *ChatRepo {
	return &ChatRepo{db: db}
}

// Save 保存聊天消息。
func (r *ChatRepo) Save(msg *model.ChatMessage) error {
	return r.db.Create(msg).Error
}

// FindByNPCAndUser 获取某用户与某 NPC 的对话历史（最近 N 条）。
func (r *ChatRepo) FindByNPCAndUser(npcID uint, userToken string, limit int) ([]model.ChatMessage, error) {
	var msgs []model.ChatMessage
	err := r.db.
		Where("npc_id = ? AND user_token = ?", npcID, userToken).
		Order("created_at DESC").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

// FindRecentByNPC 获取 NPC 的最近对话（不限用户）。
func (r *ChatRepo) FindRecentByNPC(npcID uint, limit int) ([]model.ChatMessage, error) {
	var msgs []model.ChatMessage
	err := r.db.
		Where("npc_id = ?", npcID).
		Order("created_at DESC").
		Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
