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

type Gender struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Name string `json:"name"` // Пример: "Male", "Female"
}

func (Gender) TableName() string {
	return "genders" // Название таблицы
}

type Animal struct {
	ID           int          `json:"id" gorm:"primaryKey"`
	Name         string       `json:"name"`
	SpeciesID    int          `json:"species_id" gorm:"not null"`
	Species      AnimalType   `json:"species" gorm:"foreignKey:SpeciesID"`
	Breed        string       `json:"breed"`
	Age          int          `json:"age"`
	GenderID     int          `json:"gender_id" gorm:"not null"`
	Gender       Gender       `json:"gender" gorm:"foreignKey:GenderID"`
	StatusID     int          `json:"status_id" gorm:"not null"`
	Status       AnimalStatus `json:"status" gorm:"foreignKey:StatusID"`
	ArrivalDate  string       `json:"arrival_date"`
	Description  string       `json:"description"`
	Location     string       `json:"location"`
	Weight       int          `json:"weight"`
	Color        string       `json:"color"`
	Images       []PostImage  `gorm:"foreignKey:AnimalID"`
	IsSterilized bool         `json:"is_sterilized" gorm:"default:false"`
	HasPassport  bool         `json:"has_passport" gorm:"default:false"`
}

func (Animal) TableName() string {
	return "animals"
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

type PostImage struct {
	ID       int    `json:"id" gorm:"primaryKey"`
	AnimalID int    `json:"animal_id" gorm:"index;foreignKey:AnimalID;constraint:OnDelete:CASCADE;"`
	ImageURL string `json:"image_url"`
	Animal   Animal `gorm:"foreignKey:AnimalID;references:ID"`
}

func (PostImage) TableName() string {
	return "postimages"
}
