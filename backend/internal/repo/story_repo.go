package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type StoryRepo struct {
	db *gorm.DB
}

func NewStoryRepo(db *gorm.DB) *StoryRepo {
	return &StoryRepo{db: db}
}

func (r *StoryRepo) Create(event *model.StoryEvent) error {
	return r.db.Create(event).Error
}

func (r *StoryRepo) FindByTownID(townID uint) ([]model.StoryEvent, error) {
	var events []model.StoryEvent
	err := r.db.Where("town_id = ?", townID).Find(&events).Error
	return events, err
}

func (r *StoryRepo) FindByID(id uint) (*model.StoryEvent, error) {
	var event model.StoryEvent
	err := r.db.First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *StoryRepo) MarkTriggered(id uint) error {
	return r.db.Model(&model.StoryEvent{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":            "cooldown",
			"last_triggered_at": gorm.Expr("NOW()"),
		}).Error
}
