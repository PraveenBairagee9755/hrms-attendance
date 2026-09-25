package salary

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CalculateSalaryRequest maps incoming JSON for salary calculation.
type CalculateSalaryRequest struct {
	EmployeeID string `json:"employeeId"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
}

// CalculateSalaryHandler calculates an employee's salary
// after applying leave-limit and LOP deductions.
func (h *Handler) CalculateSalaryHandler(c fiber.Ctx) error {

	var req CalculateSalaryRequest

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// --------------------------------
	// Get employee ID
	// --------------------------------

	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employee ID is required"})
	}

	// --------------------------------
	// Get year
	// --------------------------------

	year := req.Year

	if year == 0 {
		year = time.Now().Year()
	}

	// --------------------------------
	// Get month
	// --------------------------------

	monthNumber := req.Month

	if monthNumber == 0 {
		monthNumber = int(time.Now().Month())
	}

	if monthNumber < 1 || monthNumber > 12 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "month must be between 1 and 12"})
	}

	month := time.Month(monthNumber)
	employeeID := req.EmployeeID

	// --------------------------------
	// Call service
	// --------------------------------

	result, err := h.service.CalculateSalary(
		c.Context(),
		employeeID,
		year,
		month,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// --------------------------------
	// Return response
	// --------------------------------

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": result})
}

// ImportSalaryStructureExcelHandler handles salary structure Excel imports.
func (h *Handler) ImportSalaryStructureExcelHandler(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Excel file is required",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "failed to open uploaded Excel file",
		})
	}
	defer file.Close()

	successRows, failedRows, errorsList, err :=
		h.service.ImportSalaryStructureExcel(c.Context(), file)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":     true,
		"message":     fmt.Sprintf("%d salary structure row(s) imported successfully", successRows),
		"successRows": successRows,
		"failedRows":  failedRows,
		"errors":      errorsList,
	})
}
