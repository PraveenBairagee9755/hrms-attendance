package attendance

import (
	"github.com/gofiber/fiber/v2"
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
func (h *Handler) ClockInHandler(c *fiber.Ctx) error {

	var req AttendanceRequest

	// Parse JSON request body.
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// Validate employee ID.
	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "employeeId is required",
		})
	}

	// Call service layer.
	err := h.service.ClockIn(
		c.UserContext(),
		req.EmployeeID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "Successfully clocked in",
		"employeeId": req.EmployeeID,
	})
}

// ClockOutHandler handles POST /api/attendance/clock-out.
func (h *Handler) ClockOutHandler(c *fiber.Ctx) error {

	var req AttendanceRequest

	// Parse JSON request body.
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	// Validate employee ID.
	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "employeeId is required",
		})
	}

	// Call service layer.
	err := h.service.ClockOut(
		c.UserContext(),
		req.EmployeeID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Successfully clocked out",
		"employeeId": req.EmployeeID,
	})
}
