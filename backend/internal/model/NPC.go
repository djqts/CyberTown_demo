package model

import (
	"time"

	"gorm.io/gorm"
)

type NPC struct {
	gorm.Model
	TownID       uint          `json:"town_id" gorm:"not null;index"`
	LocationID   uint          `json:"location_id" gorm:"not null;index"`
	Name         string        `json:"name" gorm:"type:varchar(255);not null"`
	Role         string        `json:"role" gorm:"type:varchar(255)"`
	Personality  string        `json:"personality" gorm:"type:text"`
	Status       string        `json:"status" gorm:"type:varchar(100)"`
	CurrentGoal  string        `json:"current_goal" gorm:"type:varchar(255)"`
	Mood         string        `json:"mood" gorm:"type:varchar(50);default:'content'"`
	Energy       int           `json:"energy" gorm:"default:80"`
	Gender       string        `json:"gender" gorm:"type:varchar(10)"`
	AgeGroup     string        `json:"age_group" gorm:"type:varchar(20)"`
	Catchphrase  string        `json:"catchphrase" gorm:"type:varchar(255)"`
	Appearance   string        `json:"appearance" gorm:"type:text"`
	LastActiveAt *time.Time    `json:"last_active_at"`
	Town         Town          `json:"town" gorm:"foreignKey:TownID"`
	Location     Location      `json:"location" gorm:"foreignKey:LocationID"`
	Schedules    []NPCSchedule `json:"schedules" gorm:"foreignKey:NPCID"`
	ChatMessages []ChatMessage `json:"chat_messages" gorm:"foreignKey:NPCID"`
}
