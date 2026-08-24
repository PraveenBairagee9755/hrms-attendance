package attendance

import (
	"context"
	"errors"
)

// Service handles business logic validations for the attendance system
type Service struct {
	repo *Repository
}

// NewService creates a new instance of the Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// ClockIn validates and processes an employee's sign-in
func (s *Service) ClockIn(ctx context.Context, employeeID string) error {
	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}
	
	// Add business rules here later (e.g., check if already clocked in today)
	return s.repo.ClockIn(ctx, employeeID)
}

// ClockOut validates and processes an employee's sign-out
func (s *Service) ClockOut(ctx context.Context, employeeID string) error {
	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}
	
	return s.repo.ClockOut(ctx, employeeID)
}
