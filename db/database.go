package db

import (
	"Animals_Shelter/models"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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
		&models.Topic{},
		&models.Post{},
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
