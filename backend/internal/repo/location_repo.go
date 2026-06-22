package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type LocationRepo struct {
	db *gorm.DB
}

func NewLocationRepo(db *gorm.DB) *LocationRepo {
	return &LocationRepo{db: db}
}

func (r *LocationRepo) Create(loc *model.Location) error {
	return r.db.Create(loc).Error
}

func (r *LocationRepo) FindByID(id uint) (*model.Location, error) {
	var loc model.Location
	err := r.db.
		Preload("NPCs").
		Preload("Schedules").
		First(&loc, id).Error
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *LocationRepo) FindByTownID(townID int64) ([]model.Location, error) {
	var locs []model.Location
	err := r.db.Where("town_id = ?", townID).
		Preload("NPCs").
		Find(&locs).Error
	return locs, err
}

func (r *LocationRepo) Update(loc *model.Location) error {
	return r.db.Save(loc).Error
}

func (r *LocationRepo) Delete(id uint) error {
	return r.db.Delete(&model.Location{}, id).Error
}
