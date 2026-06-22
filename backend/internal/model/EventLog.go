package model

import "gorm.io/gorm"

type EventLog struct {
	gorm.Model
	TownID    uint   `json:"town_id" gorm:"not null;index"`
	EventType string `json:"event_type" gorm:"type:varchar(100);not null"`
	Payload   string `json:"payload" gorm:"type:text"`

	Town Town `json:"town" gorm:"foreignKey:TownID"`
}
