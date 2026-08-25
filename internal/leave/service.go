package leave

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetLeaveTypes returns all active leave types.
func (s *Service) GetLeaveTypes(
	ctx context.Context,
) (interface{}, error) {

	return s.repo.GetLeaveTypes(ctx)
}

// ApplyLeave validates and creates a leave application.
func (s *Service) ApplyLeave(
	ctx context.Context,
	employeeID string,
	leaveTypeID string,
	startDate time.Time,
	endDate time.Time,
	reason string,
) error {

	// Validate employee ID.
	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return errors.New("invalid employee ID")
	}

	// Validate leave type ID.
	if leaveTypeID == "" {
		return errors.New("leave type ID cannot be empty")
	}

	if _, err := uuid.Parse(leaveTypeID); err != nil {
		return errors.New("invalid leave type ID")
	}

	// Validate dates.
	if startDate.IsZero() {
		return errors.New("start date is required")
	}

	if endDate.IsZero() {
		return errors.New("end date is required")
	}

	if endDate.Before(startDate) {
		return errors.New("end date cannot be before start date")
	}

	// Calculate number of leave days.
	totalDays := calculateLeaveDays(startDate, endDate)

	if totalDays <= 0 {
		return errors.New("leave duration must be greater than zero")
	}

	// Validate reason.
	if reason == "" {
		return errors.New("leave reason cannot be empty")
	}

	// Create leave application.
	return s.repo.ApplyLeave(
		ctx,
		employeeID,
		leaveTypeID,
		startDate,
		endDate,
		totalDays,
		reason,
	)
}

// GetEmployeeLeaveHistory returns an employee's leave applications.
func (s *Service) GetEmployeeLeaveHistory(
	ctx context.Context,
	employeeID string,
) (interface{}, error) {

	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return nil, errors.New("invalid employee ID")
	}

	return s.repo.GetEmployeeLeaveHistory(
		ctx,
		employeeID,
	)
}

// CancelLeave cancels a pending leave application.
func (s *Service) CancelLeave(
	ctx context.Context,
	employeeID string,
	leaveApplicationID string,
) error {

	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return errors.New("invalid employee ID")
	}

	if leaveApplicationID == "" {
		return errors.New("leave application ID cannot be empty")
	}

	if _, err := uuid.Parse(leaveApplicationID); err != nil {
		return errors.New("invalid leave application ID")
	}

	return s.repo.CancelLeave(
		ctx,
		employeeID,
		leaveApplicationID,
	)
}

// ApproveLeave approves a pending leave application.
func (s *Service) ApproveLeave(
	ctx context.Context,
	leaveApplicationID string,
	approvedBy string,
) error {

	if leaveApplicationID == "" {
		return errors.New("leave application ID cannot be empty")
	}

	if _, err := uuid.Parse(leaveApplicationID); err != nil {
		return errors.New("invalid leave application ID")
	}

	if approvedBy == "" {
		return errors.New("approver ID cannot be empty")
	}

	if _, err := uuid.Parse(approvedBy); err != nil {
		return errors.New("invalid approver ID")
	}

	return s.repo.ApproveLeave(
		ctx,
		leaveApplicationID,
		approvedBy,
	)
}

// RejectLeave rejects a pending leave application.
func (s *Service) RejectLeave(
	ctx context.Context,
	leaveApplicationID string,
	rejectedBy string,
	rejectionReason string,
) error {

	if leaveApplicationID == "" {
		return errors.New("leave application ID cannot be empty")
	}

	if _, err := uuid.Parse(leaveApplicationID); err != nil {
		return errors.New("invalid leave application ID")
	}

	if rejectedBy == "" {
		return errors.New("rejector ID cannot be empty")
	}

	if _, err := uuid.Parse(rejectedBy); err != nil {
		return errors.New("invalid rejector ID")
	}

	if rejectionReason == "" {
		return errors.New("rejection reason cannot be empty")
	}

	return s.repo.RejectLeave(
		ctx,
		leaveApplicationID,
		rejectedBy,
		rejectionReason,
	)
}

// calculateLeaveDays calculates the number of calendar days
// between start date and end date, including both dates.
func calculateLeaveDays(
	startDate time.Time,
	endDate time.Time,
) float64 {

	start := time.Date(
		startDate.Year(),
		startDate.Month(),
		startDate.Day(),
		0,
		0,
		0,
		0,
		startDate.Location(),
	)

	end := time.Date(
		endDate.Year(),
		endDate.Month(),
		endDate.Day(),
		0,
		0,
		0,
		0,
		endDate.Location(),
	)

	duration := end.Sub(start)

	return duration.Hours()/24 + 1
}
