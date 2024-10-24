package models

import (
	"time"
)

type User struct {
	ID                int       `json:"id" gorm:"primaryKey"`                 // Идентификатор пользователя (основной ключ)
	Username          string    `json:"username" gorm:"unique;not null"`      // Уникальное имя пользователя
	Password          string    `json:"password" gorm:"not null"`             // Пароль пользователя (хранится в зашифрованном виде)
	Email             string    `json:"email" gorm:"unique;not null"`         // Уникальный email пользователя
	Role              string    `json:"role" gorm:"default:'user'"`           // Роль пользователя (по умолчанию "user")
	IsAdmin           bool      `json:"is_admin" gorm:"default:false"`        // Флаг, указывающий, является ли пользователь администратором
	EmailConfirmed    bool      `json:"email_confirmed" gorm:"default:false"` // Подтвержден ли email
	ConfirmationToken string    `json:"confirmation_token" gorm:"not null"`   // Токен для подтверждения email
	FirstName         string    `json:"first_name"`                           // Имя пользователя
	LastName          string    `json:"last_name"`                            // Фамилия пользователя
	Bio               string    `json:"bio" gorm:"type:text"`                 // Биография пользователя (текстовое поле)
	ProfileImage      string    `json:"profile_image"`                        // Путь к изображению профиля
	ProfileBgImage    string    `json:"profile_bg_image"`                     // Путь к изображению фона профиля
	PhoneNumber       string    `json:"phone_number"`                         // Номер телефона пользователя
	DateOfBirth       time.Time `json:"date_of_birth" gorm:"type:date"`       // Дата рождения пользователя (формат даты)

}

func (User) TableName() string {
	return "users" // Переопределяем имя таблицы для модели User
}
