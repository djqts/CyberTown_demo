package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

// ScheduleRepo NPC 日程数据访问。
type ScheduleRepo struct {
	db *gorm.DB
}

// NewScheduleRepo 创建日程 Repo。
func NewScheduleRepo(db *gorm.DB) *ScheduleRepo {
	return &ScheduleRepo{db: db}
}

// FindByNPCID 查找指定 NPC 的所有日程。
func (r *ScheduleRepo) FindByNPCID(npcID uint) ([]model.NPCSchedule, error) {
	var scheds []model.NPCSchedule
	err := r.db.Where("npc_id = ?", npcID).Find(&scheds).Error
	return scheds, err
}

// FindByTownID 查找城镇中所有 NPC 的日程（JOIN npcs）。
func (r *ScheduleRepo) FindByTownID(townID uint) ([]model.NPCSchedule, error) {
	var scheds []model.NPCSchedule
	err := r.db.
		Joins("JOIN t_npcs ON t_npcs.id = t_npc_schedules.npc_id").
		Where("t_npcs.town_id = ?", townID).
		Find(&scheds).Error
	return scheds, err
}
