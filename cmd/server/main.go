package main

import (
	"log"
	"net/http"

	"hrms-attendance/db"
	"hrms-attendance/internal/attendance"
)

func main() {
	log.Println("Starting HRMS Attendance Server...")

	// 1. Initialize database connection pool
	database, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	// 2. Initialize layers (Dependency Injection)
	attendanceRepo := attendance.NewRepository(database)
	attendanceService := attendance.NewService(attendanceRepo)
	attendanceHandler := attendance.NewHandler(attendanceService)

	// 3. Register HTTP Routes
	http.HandleFunc("/api/attendance/clock-in", attendanceHandler.ClockInHandler)
	http.HandleFunc("/api/attendance/clock-out", attendanceHandler.ClockOutHandler)

	// Simple check route to confirm server is alive
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HRMS Attendance Service is running perfectly!"))
	})

	log.Println("Server starting on port :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
