package model

import "gorm.io/gorm"

type Location struct {
	gorm.Model
	TownID    int64  `json:"town_id" gorm:"not null;index"`
	Name      string `json:"name"`
	Longitude string `json:"longitude"`
	Latitude  string `json:"latitude"`

	// 关联
	Town      Town          `json:"town" gorm:"foreignKey:TownID"`
	NPCs      []NPC         `json:"npcs" gorm:"foreignKey:location_id"`
	Schedules []NPCSchedule `json:"schedules" gorm:"foreignKey:location_id"`
}
