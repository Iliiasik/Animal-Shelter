package models

import (
	"time"
)

type Topic struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Title       string    `json:"title" gorm:"size:255"`
	Description string    `json:"description" gorm:"size:500"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	User User `gorm:"foreignKey:UserID"`
}

func (Topic) TableName() string {
	return "topics"
}
