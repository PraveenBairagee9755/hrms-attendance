package leave

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type PolicyService struct {
	mu       sync.RWMutex
	policies map[string]LeavePolicy
}

func NewPolicyService() *PolicyService {
	return &PolicyService{
		policies: make(map[string]LeavePolicy),
	}
}

// CreatePolicy creates a new leave policy.
func (s *PolicyService) CreatePolicy(
	ctx context.Context,
	policy LeavePolicy,
) (*LeavePolicy, error) {

	if err := validateLeavePolicy(policy); err != nil {
		return nil, err
	}

	now := time.Now()

	policy.ID = uuid.New().String()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	if policy.Status == "" {
		policy.Status = "Active"
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.policies[policy.ID] = policy

	result := policy

	return &result, nil
}

// GetPolicies returns all leave policies.
func (s *PolicyService) GetPolicies(
	ctx context.Context,
) []LeavePolicy {

	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]LeavePolicy, 0, len(s.policies))

	for _, policy := range s.policies {
		policies = append(policies, policy)
	}

	return policies
}

// GetPolicy returns a leave policy by ID.
func (s *PolicyService) GetPolicy(
	ctx context.Context,
	id string,
) (*LeavePolicy, error) {

	if id == "" {
		return nil, errors.New("policy ID is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[id]

	if !exists {
		return nil, errors.New("leave policy not found")
	}

	result := policy

	return &result, nil
}

// UpdatePolicy updates an existing leave policy.
func (s *PolicyService) UpdatePolicy(
	ctx context.Context,
	id string,
	policy LeavePolicy,
) (*LeavePolicy, error) {

	if id == "" {
		return nil, errors.New("policy ID is required")
	}

	if err := validateLeavePolicy(policy); err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.policies[id]

	if !exists {
		return nil, errors.New("leave policy not found")
	}

	policy.ID = existing.ID
	policy.CreatedAt = existing.CreatedAt
	policy.UpdatedAt = time.Now()

	if policy.Status == "" {
		policy.Status = existing.Status
	}

	s.policies[id] = policy

	result := policy

	return &result, nil
}

// DeletePolicy removes a leave policy.
func (s *PolicyService) DeletePolicy(
	ctx context.Context,
	id string,
) error {

	if id == "" {
		return errors.New("policy ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[id]; !exists {
		return errors.New("leave policy not found")
	}

	delete(s.policies, id)

	return nil
}

func validateLeavePolicy(policy LeavePolicy) error {

	if strings.TrimSpace(policy.Name) == "" {
		return errors.New("policy name is required")
	}

	if strings.TrimSpace(policy.LeaveTypeID) == "" {
		return errors.New("leaveTypeId is required")
	}

	if policy.AnnualLimit <= 0 {
		return errors.New("annualLimit must be greater than zero")
	}

	if policy.MaxConsecutiveDays <= 0 {
		return errors.New("maxConsecutiveDays must be greater than zero")
	}

	if policy.MaxConsecutiveDays > policy.AnnualLimit {
		return errors.New(
			"maxConsecutiveDays cannot be greater than annualLimit",
		)
	}

	if policy.CarryForward {
		if policy.MaxCarryForwardDays <= 0 {
			return errors.New("maxCarryForwardDays must be greater than zero when carryForward is enabled")
		}

		if policy.MaxCarryForwardDays > policy.AnnualLimit {
			return errors.New("maxCarryForwardDays cannot be greater than annualLimit")
		}
	} else {
		policy.MaxCarryForwardDays = 0
	}

	if policy.EffectiveFrom.IsZero() {
		return errors.New("effectiveFrom is required")
	}

	if policy.Status == "" {
		policy.Status = "Active"
	}

	if policy.Status != "Active" &&
		policy.Status != "Inactive" {

		return fmt.Errorf("invalid policy status: %s", policy.Status)
	}

	return nil
}
