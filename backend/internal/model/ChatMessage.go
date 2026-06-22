package model

import "gorm.io/gorm"

type ChatMessage struct {
	gorm.Model
	NPCID     uint   `json:"npc_id" gorm:"not null;index"`
	UserToken string `json:"user_token" gorm:"type:varchar(255);not null;index"`
	Role      string `json:"role" gorm:"type:varchar(50);not null"`
	Content   string `json:"content" gorm:"type:text;not null"`

	NPC NPC `json:"npc" gorm:"foreignKey:NPCID"`
}
