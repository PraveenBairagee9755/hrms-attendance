package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"

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

	app.Post("/api/attendance/clock-out", attendanceHandler.ClockOutHandler)

	app.Post("/api/attendance/import", attendanceHandler.ImportAttendanceHandler)

	app.Post("/api/attendance/get", attendanceHandler.GetEmployeeAttendanceHandler)

	// =========================================================
	// ATTENDANCE REGULARIZATION MODULE
	// =========================================================

	regularizationHandler := attendance.NewRegularizationHandler(attendanceService)

	app.Post("/api/attendance/regularization", regularizationHandler.CreateRegularizationHandler)

	app.Post("/api/attendance/regularization/get", regularizationHandler.GetRegularizationHistoryHandler)

	app.Post("/api/attendance/regularization/approve/:id", regularizationHandler.ApproveRegularizationHandler)

	app.Post("/api/attendance/regularization/reject/:id", regularizationHandler.RejectRegularizationHandler)

	app.Post("/api/attendance/regularization/import", regularizationHandler.ImportRegularizationHandler)

	// =========================================================
	// LEAVE MODULE
	// =========================================================

	leaveRepo := leave.NewRepository(database)
	leaveService := leave.NewService(leaveRepo)
	leaveHandler := leave.NewHandler(leaveService)

	leavePolicyService := leave.NewPolicyService()
	leavePolicyRepo := leave.NewPolicyRepository(database)
	leavePolicyHandler := leave.NewPolicyHandler(leavePolicyService, leavePolicyRepo)

	app.Get("/api/leave/types", leaveHandler.GetLeaveTypesHandler)

	app.Post("/api/leave/apply", leaveHandler.ApplyLeaveHandler)

	app.Get("/api/leave/history/:employeeId", leaveHandler.GetEmployeeLeaveHistoryHandler)

	app.Post("/api/leave/cancel/:id", leaveHandler.CancelLeaveHandler)

	app.Post("/api/leave/approve/:id", leaveHandler.ApproveLeaveHandler)

	app.Post("/api/leave/reject/:id", leaveHandler.RejectLeaveHandler)

	app.Get("/api/leave/balance/:employeeId", leaveHandler.GetEmployeeLeaveBalancesHandler)

	app.Post("/api/leave/import", leaveHandler.ImportLeaveApplicationExcelHandler)

	app.Get("/api/leave/approvals/:leaveApplicationId",leaveHandler.GetLeaveApprovals)

	// LEAVE POLICY ROUTES
	app.Post("/api/leave/policies", leavePolicyHandler.CreatePolicyHandler)

	app.Get("/api/leave/policies", leavePolicyHandler.GetPoliciesDBHandler)

	app.Get("/api/leave/policies/:id", leavePolicyHandler.GetPolicyDBHandler)

	app.Put("/api/leave/policies/:id", leavePolicyHandler.UpdatePolicyHandler)

	app.Delete("/api/leave/policies/:id", leavePolicyHandler.DeletePolicyHandler)

	// =========================================================
	// SALARY MODULE
	// =========================================================

	salaryRepo := salary.NewRepository(database)
	salaryService := salary.NewService(salaryRepo)
	salaryHandler := salary.NewHandler(salaryService)

	app.Post("/api/salary/calculate", salaryHandler.CalculateSalaryHandler)

	app.Post("/api/salary/import", salaryHandler.ImportSalaryStructureExcelHandler)

	// =========================================================
	// HEALTH CHECK
	// =========================================================

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("HRMS Attendance Service is running perfectly!")
	})

	// =========================================================
	// START SERVER
	// =========================================================

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port :%s", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
