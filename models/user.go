package models

import "time"

// Роли пользователей
type Role struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"unique;not null"`
}

// Пользователи
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`                                                         // Идентификатор пользователя
	Username string `json:"username" gorm:"unique;not null"`                                              // Уникальное имя пользователя
	Password string `json:"password" gorm:"not null"`                                                     // Пароль пользователя
	Email    string `json:"email" gorm:"unique;not null"`                                                 // Уникальный email
	RoleID   uint   `json:"role_id"`                                                                      // Внешний ключ для роли
	Role     Role   `json:"role" gorm:"foreignKey:RoleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // Связь с таблицей ролей
}

// Детали о пользователях
type UserDetail struct {
	UserID               uint      `json:"user_id" gorm:"primaryKey"`                                                   // Внешний ключ на таблицу пользователей и первичный ключ
	User                 User      `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Связь с таблицей пользователей
	PhoneNumber          string    `json:"phone_number"`                                                                // Номер телефона
	FirstName            string    `json:"first_name"`                                                                  // Имя
	LastName             string    `json:"last_name"`                                                                   // Фамилия
	Bio                  string    `json:"bio" gorm:"type:text"`                                                        // Биография
	DateOfBirth          time.Time `json:"date_of_birth" gorm:"type:date"`                                              // Дата рождения
	FormattedDateOfBirth string    `gorm:"-"`                                                                           // Форматированная дата рождения (не сохраняется в БД)
}

type UserPrivacy struct {
	UserID    uint `json:"user_id" gorm:"primaryKey"`                                                   // Внешний ключ на таблицу пользователей и первичный ключ
	User      User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Связь с таблицей пользователей
	ShowEmail bool `json:"show_email" gorm:"default:true"`                                              // Показывать email
	ShowPhone bool `json:"show_phone" gorm:"default:true"`                                              // Показывать телефон
}

type UserImage struct {
	UserID         uint   `json:"user_id" gorm:"primaryKey"`                                                   // Внешний ключ на таблицу пользователей
	User           User   `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Связь с таблицей пользователей
	ProfileImage   string `json:"profile_image" gorm:"default:'system_images/default_profile.jpg'"`            // Изображение профиля
	ProfileBgImage string `json:"profile_bg_image" gorm:"default:'system_images/default_bg.jpg'"`              // Фоновое изображение
}

type UserEmailConfirmation struct {
	UserID            uint      `json:"user_id" gorm:"primaryKey"`                                                   // Внешний ключ на пользователя
	User              User      `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Связь с таблицей пользователей
	EmailConfirmed    bool      `json:"email_confirmed" gorm:"default:false"`                                        // Подтвержден ли email
	ConfirmationToken string    `json:"confirmation_token" gorm:"not null"`                                          // Токен подтверждения email
	CreatedAt         time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`                                 // Дата создания
}

func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (UserDetail) TableName() string {
	return "user_details"
}

func (UserPrivacy) TableName() string {
	return "user_privacy"
}

func (UserImage) TableName() string {
	return "user_images"
}

func (UserEmailConfirmation) TableName() string {
	return "user_email_confirmations"
}
