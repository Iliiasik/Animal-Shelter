package models

import (
	"time"
)

type Post struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	TopicID   uint      `json:"topic_id"`                      // Внешний ключ к таблице Topic
	UserID    uint      `json:"user_id" gorm:"index;not null"` // Внешний ключ к таблице Users
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	ParentID  *int      `json:"parent_id"`               // ID родительского поста (если это ответ)
	Rating    int       `json:"rating" gorm:"default:0"` // Общий рейтинг (+ или -)
	ImageURL  string    `json:"image_url"`               // URL изображения (если есть)

	Topic Topic `gorm:"foreignKey:TopicID;references:ID;constraint:OnDelete:CASCADE"`
	User  User  `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Post) TableName() string {
	return "posts"
}

type PostLike struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	PostID     uint      `json:"post_id" gorm:"not null;index"` // внешний ключ для поста
	UserID     uint      `json:"user_id" gorm:"not null;index"` // внешний ключ для пользователя
	LikeStatus bool      `json:"like_status" gorm:"not null"`   // true для лайка, false для дизлайка
	CreatedAt  time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`

	Post Post `gorm:"foreignKey:PostID;references:ID;constraint:OnDelete:CASCADE;" json:"post"` // Связь с постом
	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE;" json:"user"` // Связь с пользователем
}

func (PostLike) TableName() string {
	return "post_likes"
}
