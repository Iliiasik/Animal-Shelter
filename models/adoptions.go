package models

import (
	"time"
)

type Adoption struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	AnimalID       int       `json:"animal_id" gorm:"index;not null"`    // Внешний ключ к Animals
	AdopterName    string    `json:"adopter_name" gorm:"index;not null"` // Имя пользователя
	AdopterContact string    `json:"adopter_contact"`
	AdoptionDate   time.Time `json:"adoption_date"`

	// Relationships
	Animal   Animal `gorm:"foreignKey:AnimalID;references:ID;constraint:OnDelete:CASCADE;"`           // Связь с таблицей Animals
	UserName User   `gorm:"foreignKey:AdopterName;references:username;constraint:OnDelete:SET NULL;"` // Связь с таблицей Users по полю username
}

func (Adoption) TableName() string {
	return "adoptions" // Название таблицы соответствует SQL-структуре
}
