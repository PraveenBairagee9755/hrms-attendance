package leave

import (
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

// -------------------------
// Request DTOs
// -------------------------

type ApplyLeaveRequest struct {
	EmployeeID  string `json:"employeeId"`
	LeaveTypeID string `json:"leaveTypeId"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Reason      string `json:"reason"`
}

type CancelLeaveRequest struct {
	EmployeeID string `json:"employeeId"`
}

type ApproveLeaveRequest struct {
	ApprovedBy string `json:"approvedBy"`
}

type RejectLeaveRequest struct {
	RejectedBy      string `json:"rejectedBy"`
	RejectionReason string `json:"rejectionReason"`
}

// -------------------------
// Get Leave Types
// GET /api/leave/types
// -------------------------

func (h *Handler) GetLeaveTypesHandler(c *fiber.Ctx) error {

	leaveTypes, err := h.service.GetLeaveTypes(
		c.UserContext(),
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": leaveTypes})
}

// -------------------------
// Employee Leave Balance
// GET /api/leave/balance/:employeeId?year=2026
// -------------------------

func (h *Handler) GetEmployeeLeaveBalancesHandler(c *fiber.Ctx) error {

	employeeID := c.Params("employeeId")

	if employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	year := c.QueryInt("year", time.Now().Year())

	if year <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid year"})
	}

	balances, err := h.service.GetEmployeeLeaveBalances(
		c.UserContext(),
		employeeID,
		year,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": balances})
}

// -------------------------
// Apply Leave
// POST /api/leave/apply
// -------------------------

func (h *Handler) ApplyLeaveHandler(c *fiber.Ctx) error {

	var req ApplyLeaveRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	if req.LeaveTypeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "leaveTypeId is required"})
	}

	if req.StartDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "startDate is required"})
	}

	if req.EndDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "endDate is required"})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "reason is required"})
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "startDate must use YYYY-MM-DD format"})
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "endDate must use YYYY-MM-DD format"})
	}

	err = h.service.ApplyLeave(
		c.UserContext(),
		req.EmployeeID,
		req.LeaveTypeID,
		startDate,
		endDate,
		req.Reason,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Leave application submitted successfully"})
}

// -------------------------
// Employee Leave History
// GET /api/leave/history/:employeeId
// -------------------------

func (h *Handler) GetEmployeeLeaveHistoryHandler(c *fiber.Ctx) error {

	employeeID := c.Params("employeeId")

	if employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	history, err := h.service.GetEmployeeLeaveHistory(
		c.UserContext(),
		employeeID,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": history})
}

// -------------------------
// Cancel Leave
// POST /api/leave/cancel/:id
// -------------------------

func (h *Handler) CancelLeaveHandler(c *fiber.Ctx) error {

	leaveApplicationID := c.Params("id")

	if leaveApplicationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "leave application ID is required"})
	}

	var req CancelLeaveRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.EmployeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "employeeId is required"})
	}

	err := h.service.CancelLeave(
		c.UserContext(),
		req.EmployeeID,
		leaveApplicationID,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Leave cancelled successfully"})
}

// -------------------------
// Approve Leave
// POST /api/leave/approve/:id
// -------------------------

func (h *Handler) ApproveLeaveHandler(c *fiber.Ctx) error {

	leaveApplicationID := c.Params("id")

	if leaveApplicationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "leave application ID is required"})
	}

	var req ApproveLeaveRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.ApprovedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "approvedBy is required"})
	}

	err := h.service.ApproveLeave(
		c.UserContext(),
		leaveApplicationID,
		req.ApprovedBy,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Leave approved successfully"})
}

// -------------------------
// Reject Leave
// POST /api/leave/reject/:id
// -------------------------

func (h *Handler) RejectLeaveHandler(c *fiber.Ctx) error {

	leaveApplicationID := c.Params("id")

	if leaveApplicationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "leave application ID is required"})
	}

	var req RejectLeaveRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if req.RejectedBy == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rejectedBy is required"})
	}

	if req.RejectionReason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "rejectionReason is required"})
	}

	err := h.service.RejectLeave(
		c.UserContext(),
		leaveApplicationID,
		req.RejectedBy,
		req.RejectionReason,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Leave rejected successfully"})
}
