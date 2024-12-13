package models

import "time"

// Все модели о животных

type AnimalStatus struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	StatusName string `json:"status_name"`
}

func (AnimalStatus) TableName() string {
	return "animalstatus" // Custom table name
}

type AnimalType struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	TypeName string `json:"type_name"`
}

func (AnimalType) TableName() string {
	return "animaltypes" // Custom table name
}

type Gender struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name"` // Пример: "Male", "Female"
}

func (Gender) TableName() string {
	return "genders" // Название таблицы
}

type AnimalAge struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	AnimalID uint `json:"animal_id" gorm:"not null;constraint:OnDelete:CASCADE;"` // Foreign key для связи с животным
	Years    int  `json:"years" gorm:"not null;default:0"`
	Months   int  `json:"months" gorm:"not null;default:0"`
}

func (AnimalAge) TableName() string {
	return "animalages"
}

type Animal struct {
	ID              uint         `json:"id" gorm:"primaryKey"`
	Name            string       `json:"name"`
	SpeciesID       uint         `json:"species_id" gorm:"not null"`
	Species         AnimalType   `json:"species" gorm:"foreignKey:SpeciesID"`
	Breed           string       `json:"breed"`
	Age             AnimalAge    `json:"age" gorm:"foreignKey:AnimalID;constraint:OnDelete:CASCADE;"`
	GenderID        uint         `json:"gender_id" gorm:"not null"`
	Gender          Gender       `json:"gender" gorm:"foreignKey:GenderID"`
	StatusID        uint         `json:"status_id" gorm:"not null"`
	Status          AnimalStatus `json:"status" gorm:"foreignKey:StatusID"`
	PublicationDate time.Time    `json:"publication_date" gorm:"default:CURRENT_TIMESTAMP"`
	Description     string       `json:"description"`
	Location        string       `json:"location"`
	Weight          float64      `json:"weight"`
	Color           string       `json:"color"`
	Images          []PostImage  `gorm:"foreignKey:AnimalID;constraint:OnDelete:CASCADE;"`
	IsSterilized    bool         `json:"is_sterilized" gorm:"default:false"`
	HasPassport     bool         `json:"has_passport" gorm:"default:false"`
	Views           int          `json:"views" gorm:"default:0"`
	UserID          uint         `json:"user_id" gorm:"not null"`
	User            User         `json:"user" gorm:"foreignKey:UserID"`
}

func (Animal) TableName() string {
	return "animals"
}

type PostImage struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	AnimalID uint   `json:"animal_id" gorm:"index;foreignKey:AnimalID;constraint:OnDelete:CASCADE;"`
	ImageURL string `json:"image_url"`
	Animal   Animal `gorm:"foreignKey:AnimalID;references:ID"`
}

func (PostImage) TableName() string {
	return "postimages"
}

type MedicalRecord struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	AnimalID    uint   `json:"animal_id" gorm:"index;foreignKey:AnimalID;constraint:OnDelete:CASCADE;"`
	CheckupDate string `json:"checkup_date"`
	Notes       string `json:"notes"`
	VetID       uint   `json:"vet_id"` // Внешний ключ для связи с пользователем (ветеринаром)
	Animal      Animal `gorm:"foreignKey:AnimalID;references:ID"`
	User        User   `gorm:"foreignKey:VetID;references:ID"` // Связь с таблицей User через VetID
}

func (MedicalRecord) TableName() string {
	return "medicalrecords"
}
