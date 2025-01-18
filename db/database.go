package db

import (
	"Animals_Shelter/models"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func ConnectDB() *gorm.DB {
	err := godotenv.Load("configuration.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=%s sslmode=disable client_encoding=UTF8",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Автоматическое создание таблиц на основе ваших моделей
	db.AutoMigrate(
		&models.Animal{},
		&models.AnimalStatus{},
		&models.AnimalType{},
		&models.Gender{},
		&models.AnimalAge{},
		&models.MedicalRecord{},
		&models.PostImage{},
		&models.Adoption{},
		&models.AdoptionStatus{},
		&models.AdoptionStatistic{},
		&models.Topic{},
		&models.Post{},
		&models.PostLike{},
		&models.Like{},
		&models.Feedback{},
		&models.User{},
		&models.Role{},
		&models.UserDetail{},
		&models.UserPrivacy{},
		&models.UserImage{},
		&models.UserEmailConfirmation{},
		&models.Session{},
	)

	// Инициализация ролей
	initializeRoles(db)
	initializeGenders(db)
	initializeAnimalStatus(db)
	initializeAnimalTypes(db)
	initializeAdoptionStatus(db)
	return db
}

// Инициализация ролей в базе данных
func initializeRoles(db *gorm.DB) {
	roles := []models.Role{
		{Name: "User"},
		{Name: "Veterinarian"},
		{Name: "Moderator"},
		{Name: "Admin"},
	}

	for _, role := range roles {
		var existingRole models.Role
		if err := db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&role).Error; err != nil {
					log.Printf("Error creating role %s: %v\n", role.Name, err)
				} else {
					log.Printf("Role %s created successfully.\n", role.Name)
				}
			} else {
				log.Printf("Error fetching role %s: %v\n", role.Name, err)
			}
		}
	}
}

func initializeGenders(db *gorm.DB) {
	genders := []models.Gender{
		{Name: "Male"},
		{Name: "Female"},
	}
	for _, gender := range genders {
		var existingGender models.Gender
		if err := db.Where("name = ?", gender.Name).First(&existingGender).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&gender).Error; err != nil {
					log.Printf("Error creating gender %s: %v\n", gender.Name, err)
				} else {
					log.Printf("Gender %s created successfully.\n", gender.Name)
				}
			} else {
				log.Printf("Error fetching gender %s: %v\n", gender.Name, err)
			}
		}
	}
}

func initializeAnimalStatus(db *gorm.DB) {
	statuses := []models.AnimalStatus{
		{Name: "Available"},
		{Name: "Booked"},
		{Name: "In treatment"},
	}

	for _, status := range statuses {
		var existingStatus models.AnimalStatus
		if err := db.Where("name = ?", status.Name).First(&existingStatus).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&status).Error; err != nil {
					log.Printf("Error creating animal status %s: %v\n", status.Name, err)
				} else {
					log.Printf("Animal status %s created successfully.\n", status.Name)
				}
			} else {
				log.Printf("Error fetching animal status %s: %v\n", status.Name, err)
			}
		}
	}
}

func initializeAnimalTypes(db *gorm.DB) {
	types := []models.AnimalType{
		{Name: "Dog"},
		{Name: "Cat"},
		{Name: "Bird"},
	}

	for _, types2 := range types {
		var existingTypes models.AnimalType
		if err := db.Where("name = ?", types2.Name).First(&existingTypes).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&types2).Error; err != nil {
					log.Printf("Error creating animal type %s: %v\n", types2.Name, err)
				} else {
					log.Printf("Animal type %s created successfully.\n", types2.Name)
				}
			} else {
				log.Printf("Error fetching animal type %s: %v\n", types2.Name, err)
			}
		}
	}
}

func initializeAdoptionStatus(db *gorm.DB) {
	statuses := []models.AdoptionStatus{
		{Name: "Under review"},
		{Name: "Approved"},
	}

	for _, status := range statuses {
		var existingStatus models.AdoptionStatus
		if err := db.Where("name = ?", status.Name).First(&existingStatus).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&status).Error; err != nil {
					log.Printf("Error creating adoption status %s: %v\n", status.Name, err)
				} else {
					log.Printf("Adoption status %s created successfully.\n", status.Name)
				}
			} else {
				log.Printf("Error fetching adoption status %s: %v\n", status.Name, err)
			}
		}
	}
}
