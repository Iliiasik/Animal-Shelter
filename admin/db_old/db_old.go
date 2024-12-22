package db_old

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func ConnectOldDB() *gorm.DB {
	err := godotenv.Load("configuration.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	dsn := fmt.Sprintf(
		"host=34.16.104.66 port=8080 user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	return db
}
