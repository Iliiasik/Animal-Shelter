package models

import (
	"time"
)

type Session struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	SessionID string     `json:"session_id" gorm:"unique;not null"`
	UserID    uint       `json:"user_id" gorm:"not null"`
	CreatedAt *time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	ExpiresAt *time.Time `json:"expires_at" gorm:"not null"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Session) TableName() string {
	return "sessions" // Custom table name
}
