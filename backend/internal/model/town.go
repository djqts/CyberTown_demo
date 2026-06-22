package model

import (
	"gorm.io/gorm"
)

type Town struct {
	gorm.Model
    Name 	string `json:"name" gorm:"not null;unique"`
	CurrentDay int    `json:"current_day" gorm:"not null;default:0"`
	CurrentMinute int    `json:"current_minute" gorm:"not null;default:0"`
	// 关联
	Locations []Location `json:"locations" gorm:"foreignKey:town_id"`
	NPCs 	[]NPC      `json:"npcs" gorm:"foreignKey:town_id"`
	Events 	[]EventLog    `json:"events" gorm:"foreignKey:town_id"`
}

