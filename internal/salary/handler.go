package salary

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// CalculateSalaryHandler calculates an employee's salary
// after applying leave-limit and LOP deductions.
func (h *Handler) CalculateSalaryHandler(c *fiber.Ctx) error {

	// --------------------------------
	// Get employee ID from URL
	// --------------------------------

	employeeID := c.Params("employeeId")

	if employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employee ID is required"})
	}

	// --------------------------------
	// Get year
	// --------------------------------

	yearString := c.Query("year")

	if yearString == "" {
		yearString = strconv.Itoa(time.Now().Year())
	}

	year, err := strconv.Atoi(yearString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid year"})
	}

	// --------------------------------
	// Get month
	// --------------------------------

	monthString := c.Query("month")

	if monthString == "" {
		monthString = strconv.Itoa(int(time.Now().Month()))
	}

	monthNumber, err := strconv.Atoi(monthString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid month"})
	}

	if monthNumber < 1 || monthNumber > 12 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "month must be between 1 and 12"})
	}

	month := time.Month(monthNumber)

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
