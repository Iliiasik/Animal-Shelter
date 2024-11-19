package db

import (
	"Animals_Shelter/models"
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
		&models.AnimalStatus{},
		&models.AnimalType{},
		&models.Animal{},
		&models.MedicalRecord{},
		&models.PostImage{},
		&models.Session{},
		&models.Adoption{},
		&models.Topic{},
		&models.Post{},
		&models.Like{},
		&models.User{},
	)

	return db
}
