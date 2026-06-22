package model

import "gorm.io/gorm"

type NPCSchedule struct {
	gorm.Model
	NPCID     uint   `json:"npcId" gorm:"index"`
	StartTime string `json:"startTime" gorm:"not null"` 
	EndTime   string `json:"endTime" gorm:"not null"` 
	LocationID uint   `json:"locationId" gorm:"not null;index"`
	Action   string `json:"action" gorm:"not null"`
	
	NPC 	NPC      `json:"npc" gorm:"foreignKey:NPCID"`
	Location Location `json:"location" gorm:"foreignKey:LocationID"`
}