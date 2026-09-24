package attendance

import (
	"github.com/gofiber/fiber/v3"
)

// Handler manages HTTP transport routes.
type Handler struct {
	service *Service
}

// NewHandler creates a new instance of the Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// AttendanceRequest maps incoming JSON inputs.
type AttendanceRequest struct {
	EmployeeID string `json:"employeeId"`
}

// ClockInHandler handles POST /api/attendance/clock-in.
func (h *Handler) ClockInHandler(c fiber.Ctx) error {

	var req AttendanceRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	result, err := h.service.ClockIn(
		c.Context(),
		req.EmployeeID,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	response := fiber.Map{
		"employeeId": result.EmployeeID,
		"inTime":     result.InTime,
		"message":    "Successfully clocked in",
	}

	// Only return lateBy when employee is late.
	if result.LateBy != "" {
		response["lateBy"] = result.LateBy
	}


	return c.Status(fiber.StatusCreated).JSON(response)
}

// ClockOutHandler handles POST /api/attendance/clock-out.
func (h *Handler) ClockOutHandler(c fiber.Ctx) error {

	var req AttendanceRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	result, err := h.service.ClockOut(
		c.Context(),
		req.EmployeeID,
	)

	// Only return earlyBy when employee comes leave early.
	//if result.EarlyBy != "" {
	//	response["earlyBy"] = result.EarlyBy
	//}

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"employeeId":     result.EmployeeID,
		"outTime":        result.OutTime,
		"totalWorkHours": result.TotalWorkHours,
		"message":        "Successfully clocked out",
	})
}

// ImportAttendanceHandler handles POST /api/attendance/import.
func (h *Handler) ImportAttendanceHandler(c fiber.Ctx) error {

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel file is required"})
	}

	// Open uploaded file.
	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to open uploaded Excel file"})
	}
	defer src.Close()

	// Process Excel file through service layer.
	totalRows, successRows, errorsList, err :=
		h.service.ImportAttendanceExcel(
			c.Context(),
			src,
		)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Attendance Excel import completed",
		"totalRows":   totalRows,
		"successRows": successRows,
		"failedRows":  len(errorsList),
		"errors":      errorsList,
	})
}

func (h *Handler) GetEmployeeAttendanceHandler(c fiber.Ctx) error {

	employeeID := c.Params("employeeId")
	fromDate := c.Query("fromDate")
	toDate := c.Query("toDate")

	attendance, err := h.service.GetEmployeeAttendance(
		c.Context(),
		employeeID,
		fromDate,
		toDate,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(
		fiber.Map{
			"employeeId": employeeID,
			"fromDate":   fromDate,
			"toDate":     toDate,
			"attendance": attendance,
		},
	)
}
