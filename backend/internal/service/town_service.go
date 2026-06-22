package service

import (
	"fmt"

	"backend/internal/model"
	"backend/internal/repo"
)

// TownService 小镇业务逻辑。
type TownService struct {
	townRepo *repo.TownRepo
}

// NewTownService 创建小镇服务。
func NewTownService(townRepo *repo.TownRepo) *TownService {
	return &TownService{townRepo: townRepo}
}

// AdvanceTime 推进小镇时间 1 分钟。返回更新后的小镇状态和是否跨天。
func (s *TownService) AdvanceTime(townID uint) (*model.Town, bool, error) {
	town, err := s.townRepo.FindByID(townID)
	if err != nil {
		return nil, false, fmt.Errorf("find town: %w", err)
	}

	town.CurrentMinute++
	dayRolled := false

	if town.CurrentMinute >= 1440 {
		town.CurrentMinute = 0
		town.CurrentDay++
		dayRolled = true
	}

	// 一次性同时更新 day 和 minute
	if err := s.townRepo.Update(town); err != nil {
		return nil, false, fmt.Errorf("update town time: %w", err)
	}

	return town, dayRolled, nil
}
