package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func ConnectDB() *sql.DB {
	// PostgreSQL connection string
	connStr := "user=postgres password=gavno dbname=AShelter sslmode=disable"

	// Open a connection to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("Connected to PostgreSQL database")

	return db
}
