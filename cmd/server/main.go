package main

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"hrms-attendance/db"
	"hrms-attendance/internal/attendance"
	"hrms-attendance/internal/leave"
	"hrms-attendance/internal/salary"
)

func main() {
	log.Println("Starting HRMS Attendance Server...")

	// =========================================================
	// DATABASE
	// =========================================================

	database, err := db.ConnectDB()
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer database.Close()

	// =========================================================
	// FIBER
	// =========================================================

	app := fiber.New()

	// =========================================================
	// ATTENDANCE MODULE
	// =========================================================

	attendanceRepo := attendance.NewRepository(database)
	attendanceService := attendance.NewService(attendanceRepo)
	attendanceHandler := attendance.NewHandler(attendanceService)

	app.Post("/api/attendance/clock-in", attendanceHandler.ClockInHandler)

	app.Post(
		"/api/attendance/clock-out", attendanceHandler.ClockOutHandler)

	// =========================================================
	// LEAVE MODULE
	// =========================================================

	leaveRepo := leave.NewRepository(database)
	leaveService := leave.NewService(leaveRepo)
	leaveHandler := leave.NewHandler(leaveService)

	app.Get("/api/leave/types", leaveHandler.GetLeaveTypesHandler)

	app.Post("/api/leave/apply", leaveHandler.ApplyLeaveHandler)

	app.Get("/api/leave/history/:employeeId", leaveHandler.GetEmployeeLeaveHistoryHandler)

	app.Post("/api/leave/cancel/:id", leaveHandler.CancelLeaveHandler)

	app.Post("/api/leave/approve/:id", leaveHandler.ApproveLeaveHandler)

	app.Post("/api/leave/reject/:id", leaveHandler.RejectLeaveHandler)

	app.Get("/api/leave/balance/:employeeId", leaveHandler.GetEmployeeLeaveBalancesHandler)

	// =========================================================
	// SALARY MODULE
	// =========================================================

	salaryRepo := salary.NewRepository(database)
	salaryService := salary.NewService(salaryRepo)
	salaryHandler := salary.NewHandler(salaryService)

	app.Get("/api/salary/calculate/:employeeId", salaryHandler.CalculateSalaryHandler)

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
