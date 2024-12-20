package models

import (
	"time"
)

type Like struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`                     // Идентификатор пользователя
	TopicID   uint      `json:"topic_id" gorm:"not null"`                    // Идентификатор темы
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"` // Дата создания лайка

	// Внешние ключи и связи
	User  User  `gorm:"foreignKey:UserID;references:ID;onDelete:CASCADE"`  // Связь с пользователем
	Topic Topic `gorm:"foreignKey:TopicID;references:ID;onDelete:CASCADE"` // Связь с темой
}

func (Like) TableName() string {
	return "likes" // Название таблицы в базе данных
}
