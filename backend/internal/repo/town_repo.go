package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type TownRepo struct {
	db *gorm.DB
}

func NewTownRepo(db *gorm.DB) *TownRepo {
	return &TownRepo{db: db}
}

func (r *TownRepo) Create(town *model.Town) error {
	return r.db.Create(town).Error
}

func (r *TownRepo) FindByID(id uint) (*model.Town, error) {
	var town model.Town
	err := r.db.
		Preload("Locations").
		Preload("NPCs").
		Preload("Events").
		First(&town, id).Error
	if err != nil {
		return nil, err
	}
	return &town, nil
}

func (r *TownRepo) FindFirst() (*model.Town, error) {
	var town model.Town
	err := r.db.
		Preload("Locations").
		Preload("NPCs").
		Preload("Events").
		First(&town).Error
	if err != nil {
		return nil, err
	}
	return &town, nil
}

func (r *TownRepo) FindByName(name string) (*model.Town, error) {
	var town model.Town
	err := r.db.Where("name = ?", name).
		Preload("Locations").
		Preload("NPCs").
		First(&town).Error
	if err != nil {
		return nil, err
	}
	return &town, nil
}

func (r *TownRepo) Update(town *model.Town) error {
	return r.db.Save(town).Error
}

func (r *TownRepo) Delete(id uint) error {
	return r.db.Delete(&model.Town{}, id).Error
}

func (r *TownRepo) List() ([]model.Town, error) {
	var towns []model.Town
	err := r.db.
		Preload("Locations").
		Preload("NPCs").
		Find(&towns).Error
	return towns, err
}

// AdvanceDay 增加当前天数并返回更新后的 Town。
func (r *TownRepo) AdvanceDay(townID uint) error {
	return r.db.Model(&model.Town{}).Where("id = ?", townID).
		UpdateColumn("current_day", gorm.Expr("current_day + 1")).Error
}

// AdvanceMinute 增加当前分钟数。
func (r *TownRepo) AdvanceMinute(townID uint) error {
	return r.db.Model(&model.Town{}).Where("id = ?", townID).
		UpdateColumn("current_minute", gorm.Expr("current_minute + 1")).Error
}
