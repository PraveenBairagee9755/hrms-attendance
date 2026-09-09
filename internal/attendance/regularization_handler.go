package attendance

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// RegularizationRequest maps incoming JSON for regularization submission.
type RegularizationRequest struct {
	EmployeeID        string `json:"employeeId"`
	AttendanceID      int32  `json:"attendanceId"`
	RequestedCheckIn  string `json:"requestedCheckIn"`
	RequestedCheckOut string `json:"requestedCheckOut"`
	Reason            string `json:"reason"`
}

// ApproveRegularizationRequest maps approval request.
type ApproveRegularizationRequest struct {
	ApprovedBy string `json:"approvedBy"`
}

// RejectRegularizationRequest maps rejection request.
type RejectRegularizationRequest struct {
	RejectedBy      string `json:"rejectedBy"`
	RejectionReason string `json:"rejectionReason"`
}

// RegularizationHandler handles Regularization HTTP requests.
type RegularizationHandler struct {
	service *Service
}

// NewRegularizationHandler creates a Regularization handler.
func NewRegularizationHandler(service *Service) *RegularizationHandler {
	return &RegularizationHandler{
		service: service,
	}
}

// CreateRegularizationHandler handles:
//
// POST /api/attendance/regularization
func (h *RegularizationHandler) CreateRegularizationHandler(c *fiber.Ctx) error {

	var req RegularizationRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	if req.AttendanceID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "attendanceId must be greater than 0"})
	}

	if req.RequestedCheckIn == "" && req.RequestedCheckOut == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "requestedCheckIn or requestedCheckOut is required"})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "reason is required"})
	}

	var requestedCheckIn *time.Time
	var requestedCheckOut *time.Time

	var err error

	if req.RequestedCheckIn != "" {
		t, parseErr := parseRegularizationTime(req.RequestedCheckIn)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid requestedCheckIn format"})
		}

		requestedCheckIn = &t
	}

	if req.RequestedCheckOut != "" {
		t, parseErr := parseRegularizationTime(req.RequestedCheckOut)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid requestedCheckOut format"})
		}

		requestedCheckOut = &t
	}

	if requestedCheckIn != nil && requestedCheckOut != nil && !requestedCheckOut.After(*requestedCheckIn) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "requestedCheckOut must be after requestedCheckIn"})
	}

	err = h.service.CreateRegularization(
		c.UserContext(),
		req.EmployeeID,
		req.AttendanceID,
		requestedCheckIn,
		requestedCheckOut,
		req.Reason,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":      "Regularization request submitted successfully",
		"employeeId":   req.EmployeeID,
		"attendanceId": req.AttendanceID,
		"status":       "Pending",
	})
}

// GetRegularizationHistoryHandler handles:
//
// GET /api/attendance/regularization/:employeeId
func (h *RegularizationHandler) GetRegularizationHistoryHandler(c *fiber.Ctx) error {

	employeeID := c.Params("employeeId")

	if employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	records, err := h.service.GetRegularizationHistory(
		c.UserContext(),
		employeeID,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": records})
}

// ApproveRegularizationHandler handles:
//
// POST /api/attendance/regularization/approve/:id
func (h *RegularizationHandler) ApproveRegularizationHandler(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "regularization ID is required"})
	}

	var req ApproveRegularizationRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.ApprovedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "approvedBy is required"})
	}

	if err := h.service.ApproveRegularization(
		c.UserContext(),
		id,
		req.ApprovedBy,
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Regularization approved successfully"})
}

// RejectRegularizationHandler handles:
//
// POST /api/attendance/regularization/reject/:id
func (h *RegularizationHandler) RejectRegularizationHandler(c *fiber.Ctx) error {

	id := c.Params("id")

	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "regularization ID is required"})
	}

	var req RejectRegularizationRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.RejectedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rejectedBy is required"})
	}

	if req.RejectionReason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rejectionReason is required"})
	}

	if err := h.service.RejectRegularization(
		c.UserContext(),
		id,
		req.RejectedBy,
		req.RejectionReason,
	); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Regularization rejected successfully"})
}

// ImportRegularizationHandler handles:
//
// POST /api/attendance/regularization/import
func (h *RegularizationHandler) ImportRegularizationHandler(c *fiber.Ctx) error {

	fileHeader, err := c.FormFile("file")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Excel file is required"})
	}

	file, err := fileHeader.Open()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open uploaded Excel file"})
	}

	defer file.Close()

	totalRows, successRows, errorsList, err := h.service.ImportRegularizationExcel(
		c.UserContext(),
		file,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Regularization Excel import completed",
		"totalRows":   totalRows,
		"successRows": successRows,
		"failedRows":  len(errorsList),
		"errors":      errorsList,
	})
}

// parseRegularizationTime accepts common timestamp formats.
func parseRegularizationTime(value string) (time.Time, error) {

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
	}

	for _, format := range formats {

		t, err := time.Parse(format, value)

		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported time format")
}
