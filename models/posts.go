package models

import (
	"time"
)

type Post struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	TopicID   int       `json:"topic_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`

	// Связь с таблицей Topic
	Topic Topic `gorm:"foreignKey:TopicID"`
}

func (Post) TableName() string {
	return "posts"
}
