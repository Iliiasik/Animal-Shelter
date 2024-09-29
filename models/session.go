package models

import (
	"time"
)

type Session struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	SessionID string     `json:"session_id" gorm:"unique;not null"`
	UserID    int        `json:"user_id" gorm:"not null"`
	CreatedAt *time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt *time.Time `json:"expires_at" gorm:"not null"`

	// Relationships
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Session) TableName() string {
	return "sessions" // Custom table name
}
