package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ConnectDB establishes a connection pool with the PostgreSQL database.
func ConnectDB() (*sql.DB, error) {
	var psqlInfo string

	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL != "" {
		psqlInfo = databaseURL
	} else {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		if host == "" {
			host = "localhost"
		}

		if port == "" {
			port = "5432"
		}

		if user == "" {
			user = "postgres"
		}

		if dbname == "" {
			dbname = "hrms_db"
		}

		if password == "" {
			return nil, fmt.Errorf("DB_PASSWORD environment variable is required")
		}

		psqlInfo = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host,
			port,
			user,
			password,
			dbname,
		)
	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	log.Println("Successfully connected to the PostgreSQL database!")

	return db, nil
}
