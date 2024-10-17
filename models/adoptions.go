package models

import (
	"time"
)

type Adoption struct {
	ID             int       `json:"id" gorm:"primaryKey"`
	AnimalID       int       `json:"animal_id" gorm:"index;not null"` // Внешний ключ к Animals
	AdopterName    string    `json:"adopter_name" gorm:"size:100"`
	AdopterContact string    `json:"adopter_contact" gorm:"size:100"`
	AdoptionDate   time.Time `json:"adoption_date"`

	// Relationship
	Animal Animal `gorm:"foreignKey:AnimalID;references:ID;constraint:OnDelete:CASCADE;"` // Связь с таблицей Animals
}

func (Adoption) TableName() string {
	return "adoptions" // Название таблицы соответствует SQL-структуре
}
