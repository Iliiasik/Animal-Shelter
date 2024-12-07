package models

import "time"

type AdoptionStatus struct {
	ID     int    `json:"id" gorm:"primaryKey"`
	Status string `json:"status"` // Пример: "Pending", "Approved", "Rejected"
}

func (AdoptionStatus) TableName() string {
	return "adoptionstatuses"
}

type Adoption struct {
	ID           int       `json:"id" gorm:"primaryKey"`
	AnimalID     int       `json:"animal_id" gorm:"index;not null"` // Внешний ключ к Animals
	UserID       int       `json:"user_id" gorm:"index;not null"`   // Внешний ключ к Users (заменяет AdopterName)
	AdoptionDate time.Time `json:"adoption_date"`
	StatusID     int       `json:"status_id" gorm:"not null"` // Статус заявки
	// Relationships
	Animal Animal         `gorm:"foreignKey:AnimalID;references:ID;constraint:OnDelete:CASCADE;"` // Связь с таблицей Animals
	User   User           `gorm:"foreignKey:UserID;references:ID;"`                               // Связь с таблицей Users (заменяет AdopterName и AdopterContact)
	Status AdoptionStatus `gorm:"foreignKey:StatusID;references:ID;"`                             // Связь с таблицей AdoptionStatus
}

func (Adoption) TableName() string {
	return "adoptions" // Название таблицы
}
