package relationship

import (
	"backend/internal/logger"
	"backend/internal/repo"
)

type Service struct {
	relRepo *repo.RelationshipRepo
	appLog  *logger.AppLogger
}

func NewService(relRepo *repo.RelationshipRepo, appLog *logger.AppLogger) *Service {
	return &Service{relRepo: relRepo, appLog: appLog}
}

func (s *Service) AdjustAffinity(npcID, targetNPCID uint, delta int) error {
	if delta == 0 {
		return nil
	}
	return s.relRepo.UpdateAffinity(npcID, targetNPCID, delta)
}
