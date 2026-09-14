package leave

import "time"

// LeavePolicy represents the rules that apply to a leave type.
//
// This is currently an in-memory model.
// Database persistence can be added after the requirements
// are confirmed by the senior/team.
type LeavePolicy struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	LeaveTypeID         string    `json:"leaveTypeId"`
	AnnualLimit         float64   `json:"annualLimit"`
	MaxConsecutiveDays  float64   `json:"maxConsecutiveDays"`
	CarryForward        bool      `json:"carryForward"`
	MaxCarryForwardDays float64   `json:"maxCarryForwardDays"`
	RequiresApproval    bool      `json:"requiresApproval"`
	EffectiveFrom       time.Time `json:"effectiveFrom"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}
