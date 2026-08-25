package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"hrms-attendance/db"
	"hrms-attendance/internal/attendance"
	"hrms-attendance/internal/leave"
)

func main() {
	log.Println("Starting HRMS Attendance Server...")

	// 1. Initialize database connection
	database, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	// 2. Create Fiber application
	app := fiber.New()

	// =========================================================
	// ATTENDANCE MODULE
	// =========================================================

	attendanceRepo := attendance.NewRepository(database)
	attendanceService := attendance.NewService(attendanceRepo)
	attendanceHandler := attendance.NewHandler(attendanceService)

	// Attendance routes
	app.Post("/api/attendance/clock-in", attendanceHandler.ClockInHandler)

	app.Post("/api/attendance/clock-out", attendanceHandler.ClockOutHandler)

	// =========================================================
	// LEAVE MODULE
	// =========================================================

	leaveRepo := leave.NewRepository(database)
	leaveService := leave.NewService(leaveRepo)
	leaveHandler := leave.NewHandler(leaveService)

	// Get available leave types
	app.Get("/api/leave/types", leaveHandler.GetLeaveTypesHandler)

	// Apply for leave
	app.Post("/api/leave/apply", leaveHandler.ApplyLeaveHandler)

	// Employee leave history
	app.Get("/api/leave/history/:employeeId", leaveHandler.GetEmployeeLeaveHistoryHandler)

	// Cancel leave
	app.Post("/api/leave/cancel/:id", leaveHandler.CancelLeaveHandler)

	// Approve leave
	app.Post("/api/leave/approve/:id", leaveHandler.ApproveLeaveHandler)

	// Reject leave
	app.Post("/api/leave/reject/:id", leaveHandler.RejectLeaveHandler)

	// =========================================================
	// HEALTH CHECK
	// =========================================================

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("HRMS Attendance Service is running perfectly!")
	})

	// =========================================================
	// START SERVER
	// =========================================================

	log.Println("Server starting on port :8080")

	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
