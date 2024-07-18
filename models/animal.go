package models

// All models about animals

type AnimalStatus struct {
	ID         int    `json:"id"`
	StatusName string `json:"status_name"`
}

type AnimalType struct {
	ID       int    `json:"id"`
	TypeName string `json:"type_name"`
}

type Animal struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Species     int    `json:"species"`
	Breed       string `json:"breed"`
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	StatusID    int    `json:"status_id"`
	ArrivalDate string `json:"arrival_date"`
	Description string `json:"description"`
	Images      []PostImage
}

type MedicalRecord struct {
	ID          int    `json:"id"`
	AnimalID    int    `json:"animal_id"`
	CheckupDate string `json:"checkup_date"`
	Notes       string `json:"notes"`
	VetName     string `json:"vet_name"`
}

type PostImage struct {
	ID       int    `json:"id"`
	AnimalID int    `json:"animal_id"`
	ImageURL string `json:"image_url"`
}
