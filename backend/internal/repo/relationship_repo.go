package repo

import (
	"backend/internal/model"

	"gorm.io/gorm"
)

type RelationshipRepo struct {
	db *gorm.DB
}

func NewRelationshipRepo(db *gorm.DB) *RelationshipRepo {
	return &RelationshipRepo{db: db}
}

func (r *RelationshipRepo) Create(rel *model.NPCRelationship) error {
	return r.db.Create(rel).Error
}

func (r *RelationshipRepo) FindByNPCID(npcID uint) ([]model.NPCRelationship, error) {
	var rels []model.NPCRelationship
	err := r.db.Where("npc_id = ?", npcID).Find(&rels).Error
	return rels, err
}

func (r *RelationshipRepo) FindBetween(npcID, targetNPCID uint) (*model.NPCRelationship, error) {
	var rel model.NPCRelationship
	err := r.db.Where("npc_id = ? AND target_npc_id = ?", npcID, targetNPCID).First(&rel).Error
	if err != nil {
		return nil, err
	}
	return &rel, nil
}

func (r *RelationshipRepo) UpdateAffinity(npcID, targetNPCID uint, delta int) error {
	rel, err := r.FindBetween(npcID, targetNPCID)
	if err != nil {
		return err
	}
	rel.Affinity += delta
	if rel.Affinity < 0 {
		rel.Affinity = 0
	}
	if rel.Affinity > 100 {
		rel.Affinity = 100
	}
	return r.db.Save(rel).Error
}

func (r *RelationshipRepo) Delete(id uint) error {
	return r.db.Delete(&model.NPCRelationship{}, id).Error
}
