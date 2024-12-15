package models

import "time"

type AdoptionStatus struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"`
}

func (AdoptionStatus) TableName() string {
	return "adoptionstatuses"
}

type Adoption struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	AnimalID     uint      `json:"animal_id" gorm:"index;not null"` // Внешний ключ к Animals
	UserID       uint      `json:"user_id" gorm:"index;not null"`   // Внешний ключ к Users (заменяет AdopterName)
	AdoptionDate time.Time `json:"adoption_date"`
	StatusID     uint      `json:"status_id" gorm:"not null"` // Статус заявки
	// Relationships
	Animal Animal         `gorm:"foreignKey:AnimalID;references:ID;constraint:OnDelete:CASCADE;"` // Связь с таблицей Animals
	User   User           `gorm:"foreignKey:UserID;references:ID;"`                               // Связь с таблицей Users (заменяет AdopterName и AdopterContact)
	Status AdoptionStatus `gorm:"foreignKey:StatusID;references:ID;"`                             // Связь с таблицей AdoptionStatus
}

func (Adoption) TableName() string {
	return "adoptions" // Название таблицы
}
