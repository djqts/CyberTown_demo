package chat

import (
	"backend/internal/model"
	"backend/internal/repo"
)

type Service struct {
	chatRepo *repo.ChatRepo
}

func NewService(chatRepo *repo.ChatRepo) *Service {
	return &Service{chatRepo: chatRepo}
}

// SaveMessage 保存聊天消息。
func (s *Service) SaveMessage(npcID uint, role, content, userToken string) (*model.ChatMessage, error) {
	msg := &model.ChatMessage{
		NPCID:     npcID,
		UserToken: userToken,
		Role:      role,
		Content:   content,
	}
	if err := s.chatRepo.Save(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// GetHistory 获取对话历史，按时间正序返回。
func (s *Service) GetHistory(npcID uint, userToken string, limit int) ([]model.ChatMessage, error) {
	msgs, err := s.chatRepo.FindByNPCAndUser(npcID, userToken, limit)
	if err != nil {
		return nil, err
	}
	// 反转 slice 以按时间正序返回
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}
