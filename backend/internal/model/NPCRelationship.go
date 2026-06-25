package model

import "gorm.io/gorm"

type NPCRelationship struct {
	gorm.Model
	NPCID       uint   `json:"npc_id" gorm:"not null;index"`
	TargetNPCID uint   `json:"target_npc_id" gorm:"not null;index"`
	Affinity    int    `json:"affinity" gorm:"default:50"`
	Trust       int    `json:"trust" gorm:"default:50"`
	Tag         string `json:"tag" gorm:"type:varchar(100)"`
}
