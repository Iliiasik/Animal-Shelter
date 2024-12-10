package models

import (
	"time"
)

type Post struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	TopicID   int       `json:"topic_id"`                      // Внешний ключ к таблице Topic
	UserID    int       `json:"user_id" gorm:"index;not null"` // Внешний ключ к таблице Users
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	ParentID  *int      `json:"parent_id"` // ID родительского поста (если это ответ)

	// Рейтинг поста (+ или -)
	Rating int `json:"rating"` // Общий рейтинг (+ или -)

	// Связь с таблицей Topic
	Topic Topic `gorm:"foreignKey:TopicID;references:ID"`

	// Связь с таблицей Users
	User User `gorm:"foreignKey:UserID;references:ID"` // Связь с таблицей Users через поле UserID
}

func (Post) TableName() string {
	return "posts"
}
