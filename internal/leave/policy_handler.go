package leave

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CreateLeavePolicyRequest struct {
	Name                string  `json:"name"`
	LeaveTypeID         string  `json:"leaveTypeId"`
	AnnualLimit         float64 `json:"annualLimit"`
	MaxConsecutiveDays  float64 `json:"maxConsecutiveDays"`
	CarryForward        bool    `json:"carryForward"`
	MaxCarryForwardDays float64 `json:"maxCarryForwardDays"`
	RequiresApproval    bool    `json:"requiresApproval"`
	EffectiveFrom       string  `json:"effectiveFrom"`
	Status              string  `json:"status"`
}

type PolicyHandler struct {
	service *PolicyService
}

func NewPolicyHandler(service *PolicyService) *PolicyHandler {
	return &PolicyHandler{
		service: service,
	}
}

// POST /api/leave/policies
func (h *PolicyHandler) CreatePolicyHandler(c *fiber.Ctx) error {

	var req CreateLeavePolicyRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	effectiveFrom, err := time.Parse(
		"2006-01-02",
		req.EffectiveFrom,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "effectiveFrom must use YYYY-MM-DD format"})
	}

	policy := LeavePolicy{
		Name:                req.Name,
		LeaveTypeID:         req.LeaveTypeID,
		AnnualLimit:         req.AnnualLimit,
		MaxConsecutiveDays:  req.MaxConsecutiveDays,
		CarryForward:        req.CarryForward,
		MaxCarryForwardDays: req.MaxCarryForwardDays,
		RequiresApproval:    req.RequiresApproval,
		EffectiveFrom:       effectiveFrom,
		Status:              req.Status,
	}

	result, err := h.service.CreatePolicy(
		c.UserContext(),
		policy,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(
		fiber.Map{
			"message": "Leave policy created successfully",
			"data":    result,
		})
}

// GET /api/leave/policies
func (h *PolicyHandler) GetPoliciesHandler(c *fiber.Ctx) error {

	policies := h.service.GetPolicies(
		c.UserContext(),
	)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": policies})
}

// GET /api/leave/policies/:id
func (h *PolicyHandler) GetPolicyHandler(c *fiber.Ctx) error {

	id := c.Params("id")

	policy, err := h.service.GetPolicy(
		c.UserContext(),
		id,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": policy})
}

// PUT /api/leave/policies/:id
func (h *PolicyHandler) UpdatePolicyHandler(c *fiber.Ctx) error {

	id := c.Params("id")

	var req CreateLeavePolicyRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	effectiveFrom, err := time.Parse(
		"2006-01-02",
		req.EffectiveFrom,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "effectiveFrom must use YYYY-MM-DD format"})
	}

	policy := LeavePolicy{
		Name:                req.Name,
		LeaveTypeID:         req.LeaveTypeID,
		AnnualLimit:         req.AnnualLimit,
		MaxConsecutiveDays:  req.MaxConsecutiveDays,
		CarryForward:        req.CarryForward,
		MaxCarryForwardDays: req.MaxCarryForwardDays,
		RequiresApproval:    req.RequiresApproval,
		EffectiveFrom:       effectiveFrom,
		Status:              req.Status,
	}

	result, err := h.service.UpdatePolicy(
		context.Background(),
		id,
		policy,
	)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(
		fiber.Map{
			"message": "Leave policy updated successfully",
			"data":    result,
		},
	)
}

// DELETE /api/leave/policies/:id
func (h *PolicyHandler) DeletePolicyHandler(c *fiber.Ctx) error {

	id := c.Params("id")

	err := h.service.DeletePolicy(
		c.UserContext(),
		id,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Leave policy deleted successfully"})
}
