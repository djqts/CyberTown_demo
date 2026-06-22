package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Create(event *model.EventLog) error {
	return r.db.Create(event).Error
}

func (r *EventRepo) FindByID(id uint) (*model.EventLog, error) {
	var event model.EventLog
	err := r.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepo) FindByTownID(townID uint) ([]model.EventLog, error) {
	var events []model.EventLog
	err := r.db.Where("town_id = ?", townID).
		Order("created_at DESC").
		Find(&events).Error
	return events, err
}

func (r *EventRepo) FindByType(townID uint, eventType string) ([]model.EventLog, error) {
	var events []model.EventLog
	err := r.db.Where("town_id = ? AND event_type = ?", townID, eventType).
		Order("created_at DESC").
		Find(&events).Error
	return events, err
}

func (r *EventRepo) Delete(id uint) error {
	return r.db.Delete(&model.EventLog{}, id).Error
}

func (r *EventRepo) DeleteByTownID(townID uint) error {
	return r.db.Where("town_id = ?", townID).Delete(&model.EventLog{}).Error
}
