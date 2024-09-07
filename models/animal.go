package models

// All models about animals

type AnimalStatus struct {
	ID         int    `json:"id" gorm:"primaryKey"`
	StatusName string `json:"status_name"`
}

func (AnimalStatus) TableName() string {
	return "animalstatus" // Custom table name
}

type AnimalType struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	TypeName string `json:"type_name"`
}

func (AnimalType) TableName() string {
	return "animaltypes" // Custom table name
}

type Animal struct {
	ID          int         `json:"id" gorm:"primaryKey"`
	Name        string      `json:"name"`
	Species     int         `json:"species"`
	Breed       string      `json:"breed"`
	Age         int         `json:"age"`
	Gender      string      `json:"gender"`
	StatusID    int         `json:"status_id" gorm:"not null"`
	ArrivalDate string      `json:"arrival_date"`
	Description string      `json:"description"`
	Images      []PostImage `gorm:"foreignKey:AnimalID"`
}

func (Animal) TableName() string {
	return "animals" // Custom table name
}

type MedicalRecord struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	AnimalID    int    `json:"animal_id" gorm:"index;foreignKey:AnimalID;constraint:OnDelete:CASCADE;"`
	CheckupDate string `json:"checkup_date"`
	Notes       string `json:"notes"`
	VetName     string `json:"vet_name"`

	// Relationships
	Animal Animal `gorm:"foreignKey:AnimalID;references:ID"`
}

func (MedicalRecord) TableName() string {
	return "medicalrecords" // Custom table name
}

type PostImage struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	AnimalID int    `json:"animal_id" gorm:"index;foreignKey:AnimalID;constraint:OnDelete:CASCADE;"`
	ImageURL string `json:"image_url"`

	// Relationships
	Animal Animal `gorm:"foreignKey:AnimalID;references:ID"`
}

func (PostImage) TableName() string {
	return "postimages" // Custom table name
}
