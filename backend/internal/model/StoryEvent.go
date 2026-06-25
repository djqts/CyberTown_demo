package model

import (
	"time"

	"gorm.io/gorm"
)

type StoryEvent struct {
	gorm.Model
	TownID           uint       `json:"town_id" gorm:"not null;index"`
	Title            string     `json:"title" gorm:"type:varchar(255)"`
	Description      string     `json:"description" gorm:"type:text"`
	Status           string     `json:"status" gorm:"type:varchar(50);default:'ready'"`
	TriggerCondition string     `json:"trigger_condition" gorm:"type:json"`
	Effects          string     `json:"effects" gorm:"type:json"`
	LastTriggeredAt  *time.Time `json:"last_triggered_at"`
}
