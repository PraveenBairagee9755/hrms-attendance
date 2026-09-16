package leave

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"hrms-attendance/db_gen/public/model"
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
	service    *PolicyService
	policyRepo *PolicyRepository
}

func NewPolicyHandler(service *PolicyService, policyRepo *PolicyRepository) *PolicyHandler {
	return &PolicyHandler{
		service:    service,
		policyRepo: policyRepo,
	}
}

// POST /api/leave/policies
func (h *PolicyHandler) CreatePolicyHandler(c *fiber.Ctx) error {
	var req CreateLeavePolicyRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	effectiveFrom, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "effectiveFrom must be in YYYY-MM-DD format"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	if req.LeaveTypeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "leaveTypeId is required"})
	}

	leaveTypeID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid leaveTypeId"})
	}

	if req.AnnualLimit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "annualLimit must be greater than 0"})
	}

	if req.MaxConsecutiveDays <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "maxConsecutiveDays must be greater than 0"})
	}

	if req.MaxConsecutiveDays > req.AnnualLimit {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "maxConsecutiveDays cannot exceed annualLimit"})
	}

	if req.CarryForward &&
		(req.MaxCarryForwardDays <= 0 ||
			req.MaxCarryForwardDays > req.AnnualLimit) {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid maxCarryForwardDays"})
	}

	status := req.Status
	if status == "" {
		status = "Active"
	}

	if status != "Active" && status != "Inactive" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be Active or Inactive"})
	}

	now := time.Now()

	policy := model.LeavePolicy{
		ID:                  uuid.New(),
		Name:                req.Name,
		LeaveTypeId:         leaveTypeID,
		AnnualLimit:         decimal.NewFromFloat(req.AnnualLimit),
		MaxConsecutiveDays:  decimal.NewFromFloat(req.MaxConsecutiveDays),
		CarryForward:        req.CarryForward,
		MaxCarryForwardDays: decimal.NewFromFloat(req.MaxCarryForwardDays),
		RequiresApproval:    req.RequiresApproval,
		EffectiveFrom:       effectiveFrom,
		Status:              status,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	createdPolicy, err := h.policyRepo.CreatePolicy(
		c.UserContext(),
		policy,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(createdPolicy)
}

// PUT /api/leave/policies/:id
func (h *PolicyHandler) UpdatePolicyHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	policyID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid policy ID"})
	}

	existingPolicy, err := h.policyRepo.GetPolicy(
		c.UserContext(),
		id,
	)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	var req CreateLeavePolicyRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	effectiveFrom, err := time.Parse("2006-01-02", req.EffectiveFrom)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "effectiveFrom must be in YYYY-MM-DD format"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	if req.LeaveTypeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "leaveTypeId is required"})
	}

	leaveTypeID, err := uuid.Parse(req.LeaveTypeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid leaveTypeId"})
	}

	if req.AnnualLimit <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "annualLimit must be greater than 0"})
	}

	if req.MaxConsecutiveDays <= 0 ||
		req.MaxConsecutiveDays > req.AnnualLimit {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid maxConsecutiveDays"})
	}

	if req.CarryForward &&
		(req.MaxCarryForwardDays <= 0 ||
			req.MaxCarryForwardDays > req.AnnualLimit) {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid maxCarryForwardDays"})
	}

	status := req.Status
	if status == "" {
		status = existingPolicy.Status
	}

	if status != "Active" && status != "Inactive" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status must be Active or Inactive"})
	}

	policy := model.LeavePolicy{
		ID:                  policyID,
		Name:                req.Name,
		LeaveTypeId:         leaveTypeID,
		AnnualLimit:         decimal.NewFromFloat(req.AnnualLimit),
		MaxConsecutiveDays:  decimal.NewFromFloat(req.MaxConsecutiveDays),
		CarryForward:        req.CarryForward,
		MaxCarryForwardDays: decimal.NewFromFloat(req.MaxCarryForwardDays),
		RequiresApproval:    req.RequiresApproval,
		EffectiveFrom:       effectiveFrom,
		Status:              status,
		CreatedAt:           existingPolicy.CreatedAt,
		UpdatedAt:           time.Now(),
	}

	updatedPolicy, err := h.policyRepo.UpdatePolicy(
		c.UserContext(),
		policy,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(updatedPolicy)
}

// DELETE /api/leave/policies/:id
func (h *PolicyHandler) DeletePolicyHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.policyRepo.DeletePolicy(
		c.UserContext(),
		id,
	); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Leave policy deleted successfully"})
}

func (h *PolicyHandler) GetPoliciesDBHandler(c *fiber.Ctx) error {
	policies, err := h.policyRepo.GetPolicies(c.UserContext())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"policies": policies})
}

func (h *PolicyHandler) GetPolicyDBHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	policy, err := h.policyRepo.GetPolicy(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(policy)
}
