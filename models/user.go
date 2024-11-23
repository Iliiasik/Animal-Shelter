package models

import "time"

type Role struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`                                                         // Идентификатор пользователя
	Username string `json:"username" gorm:"unique;not null"`                                              // Уникальное имя пользователя
	Password string `json:"password" gorm:"not null"`                                                     // Пароль пользователя
	Email    string `json:"email" gorm:"unique;not null"`                                                 // Уникальный email
	RoleID   uint   `json:"role_id"`                                                                      // Внешний ключ для роли
	Role     Role   `json:"role" gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // Связь с таблицей ролей
	IsAdmin  bool   `json:"is_admin" gorm:"default:false"`                                                // Флаг администратора
}

type UserDetail struct {
	UserID               uint      `json:"user_id" gorm:"primaryKey"` // Внешний ключ на таблицу пользователей и первичный ключ
	PhoneNumber          string    `json:"phone_number"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	Bio                  string    `json:"bio" gorm:"type:text"`           // Биография
	DateOfBirth          time.Time `json:"date_of_birth" gorm:"type:date"` // Дата рождения
	FormattedDateOfBirth string    `gorm:"-"`                              // Форматированная дата рождения
}

type UserPrivacy struct {
	UserID    uint `json:"user_id" gorm:"primaryKey"`      // Внешний ключ на таблицу пользователей и первичный ключ
	ShowEmail bool `json:"show_email" gorm:"default:true"` // Показывать email
	ShowPhone bool `json:"show_phone" gorm:"default:true"` // Показывать телефон
}

type UserImage struct {
	UserID         uint   `json:"user_id" gorm:"primaryKey"`                                        // Внешний ключ на пользователя
	ProfileImage   string `json:"profile_image" gorm:"default:'system_images/default_profile.jpg'"` // Изображение профиля
	ProfileBgImage string `json:"profile_bg_image" gorm:"default:'system_images/default_bg.jpg'"`   // Фоновое изображение
}

type UserEmailConfirmation struct {
	UserID            uint      `json:"user_id" gorm:"primaryKey"`            // Внешний ключ на пользователя
	EmailConfirmed    bool      `json:"email_confirmed" gorm:"default:false"` // Подтвержден ли email
	ConfirmationToken string    `json:"confirmation_token" gorm:"not null"`   // Токен подтверждения email
	CreatedAt         time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	ExpiresAt         time.Time `json:"expires_at"` // Время истечения токена
}

func (User) TableName() string {
	return "users" // Имя таблицы для пользователей
}

func (Role) TableName() string {
	return "roles" // Имя таблицы для ролей
}

func (UserDetail) TableName() string {
	return "user_details" // Имя таблицы для контактной информации и биографии
}

func (UserPrivacy) TableName() string {
	return "user_privacy" // Имя таблицы для конфиденциальности пользователя
}

func (UserImage) TableName() string {
	return "user_images" // Имя таблицы для изображений профиля и фона
}

func (UserEmailConfirmation) TableName() string {
	return "user_email_confirmations" // Имя таблицы для подтверждения email и токенов
}
