package attendance

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Service handles business logic validations for the attendance system.
type Service struct {
	repo *Repository
}

// NewService creates a new instance of the Service.
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// ClockIn validates and processes an employee's clock-in.
func (s *Service) ClockIn(ctx context.Context, employeeID string) error {

	// Validate employee ID is not empty.
	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}

	// Validate employee ID is a valid UUID.
	if _, err := uuid.Parse(employeeID); err != nil {
		return errors.New("invalid employee ID")
	}

	// Business rules can be added here later:
	// - Check employee exists
	// - Check employee is active
	// - Check employee hasn't already clocked in today
	// - Check employee's shift

	return s.repo.ClockIn(ctx, employeeID)
}

// ClockOut validates and processes an employee's clock-out.
func (s *Service) ClockOut(ctx context.Context, employeeID string) error {

	// Validate employee ID is not empty.
	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}

	// Validate employee ID is a valid UUID.
	if _, err := uuid.Parse(employeeID); err != nil {
		return errors.New("invalid employee ID")
	}

	// Business rules can be added here later:
	// - Check employee has clocked in
	// - Check employee hasn't already clocked out
	// - Calculate work hours
	// - Update attendance status

	return s.repo.ClockOut(ctx, employeeID)
}
