package db

import (
	"Animals_Shelter/models"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func ConnectDB() *gorm.DB {
	dsn := "user=postgres password=root dbname=AShelter sslmode=disable client_encoding=UTF8"
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
	)

	return db
}
