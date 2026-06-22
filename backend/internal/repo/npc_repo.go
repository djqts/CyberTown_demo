package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type NPCRepo struct {
	db *gorm.DB
}

func NewNPCRepo(db *gorm.DB) *NPCRepo {
	return &NPCRepo{db: db}
}

// ---- NPC ----

func (r *NPCRepo) Create(npc *model.NPC) error {
	return r.db.Create(npc).Error
}

func (r *NPCRepo) FindByID(id uint) (*model.NPC, error) {
	var npc model.NPC
	err := r.db.
		Preload("Location").
		Preload("Schedules").
		Preload("ChatMessages").
		First(&npc, id).Error
	if err != nil {
		return nil, err
	}
	return &npc, nil
}

func (r *NPCRepo) FindByTownID(townID uint) ([]model.NPC, error) {
	var npcs []model.NPC
	err := r.db.Where("town_id = ?", townID).
		Preload("Location").
		Find(&npcs).Error
	return npcs, err
}

func (r *NPCRepo) FindByLocationID(locationID uint) ([]model.NPC, error) {
	var npcs []model.NPC
	err := r.db.Where("location_id = ?", locationID).Find(&npcs).Error
	return npcs, err
}

func (r *NPCRepo) Update(npc *model.NPC) error {
	return r.db.Save(npc).Error
}

func (r *NPCRepo) UpdateLocation(npcID, locationID uint) error {
	return r.db.Model(&model.NPC{}).Where("id = ?", npcID).
		Update("location_id", locationID).Error
}

func (r *NPCRepo) UpdateStatus(npcID uint, status string) error {
	return r.db.Model(&model.NPC{}).Where("id = ?", npcID).
		Update("status", status).Error
}

func (r *NPCRepo) Delete(id uint) error {
	return r.db.Delete(&model.NPC{}, id).Error
}

// ---- NPCSchedule ----

func (r *NPCRepo) CreateSchedule(s *model.NPCSchedule) error {
	return r.db.Create(s).Error
}

func (r *NPCRepo) FindSchedulesByNPCID(npcID uint) ([]model.NPCSchedule, error) {
	var scheds []model.NPCSchedule
	err := r.db.Where("npc_id = ?", npcID).Find(&scheds).Error
	return scheds, err
}

func (r *NPCRepo) UpdateSchedule(s *model.NPCSchedule) error {
	return r.db.Save(s).Error
}

func (r *NPCRepo) DeleteSchedule(id uint) error {
	return r.db.Delete(&model.NPCSchedule{}, id).Error
}

// ---- ChatMessage ----

func (r *NPCRepo) CreateMessage(m *model.ChatMessage) error {
	return r.db.Create(m).Error
}

func (r *NPCRepo) FindMessagesByNPCID(npcID uint) ([]model.ChatMessage, error) {
	var msgs []model.ChatMessage
	err := r.db.Where("npc_id = ?", npcID).
		Order("created_at ASC").
		Find(&msgs).Error
	return msgs, err
}

func (r *NPCRepo) DeleteMessage(id uint) error {
	return r.db.Delete(&model.ChatMessage{}, id).Error
}
