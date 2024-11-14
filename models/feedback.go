package models

import "gorm.io/gorm"

type Feedback struct {
	gorm.Model
	UserID uint   `gorm:"not null"`          // ID пользователя, который оставил отзыв
	User   User   `gorm:"foreignKey:UserID"` // Связь с пользователем
	Text   string `gorm:"not null"`          // Текст отзыва
}

func (Feedback) TableName() string {
	return "feedback" // Название таблицы соответствует SQL-структуре
}
