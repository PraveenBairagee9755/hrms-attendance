package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

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

	// 3. Create Fiber application
	app := fiber.New()

	// 4. Attendance routes
	app.Post(
		"/api/attendance/clock-in",
		attendanceHandler.ClockInHandler,
	)

	app.Post(
		"/api/attendance/clock-out",
		attendanceHandler.ClockOutHandler,
	)

	// 5. Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString(
			"HRMS Attendance Service is running perfectly!",
		)
	})

	// 6. Start server
	log.Println("Server starting on port :8080")

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
