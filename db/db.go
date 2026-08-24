package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectDB establishes a connection pool with the PostgreSQL database
func ConnectDB() (*sql.DB, error) {
	// Database connection credentials
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "Pravin@123"
	dbname := "hrms_db"

	// Create the connection string
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Open the database connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Verify the connection is active
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	log.Println("Successfully connected to the PostgreSQL database!")
	return db, nil
}
