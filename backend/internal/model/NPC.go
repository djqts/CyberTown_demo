package model

import "gorm.io/gorm"

type NPC struct {
	gorm.Model
	TownID       uint          ` json:"town_id" gorm:"not null;index"`
	LocationID   uint          `json:"location_id" gorm:"not null;index"`
	Name         string        `json:"name" gorm:"type:varchar(255);not null"`
	Role         string        `json:"role" gorm:"type:varchar(255)"`
	Personality  string        `json:"personality" gorm:"type:text"`
	Status       string        `json:"status" gorm:"type:varchar(100)"`
	CurrentGoal  string        `json:"current_goal" gorm:"type:varchar(255)"`
	Town         Town          `json:"town" gorm:"foreignKey:TownID"`
	Location     Location      `json:"location" gorm:"foreignKey:LocationID"`
	Schedules    []NPCSchedule `json:"schedules" gorm:"foreignKey:NPCID"`
	ChatMessages []ChatMessage `json:"chat_messages" gorm:"foreignKey:NPCID"`
}
