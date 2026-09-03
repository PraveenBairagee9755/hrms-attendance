package leave

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hrms-attendance/db_gen/public/model"

	"github.com/shopspring/decimal"
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
) ([]model.LeaveType, error) {

	return s.repo.GetLeaveTypes(ctx)
}

// ApplyLeave validates and creates a pending leave application.
func (s *Service) ApplyLeave(
	ctx context.Context,
	employeeID string,
	leaveTypeID string,
	startDate time.Time,
	endDate time.Time,
	reason string,
) error {

	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}

	if leaveTypeID == "" {
		return errors.New("leave type ID cannot be empty")
	}

	if startDate.IsZero() {
		return errors.New("start date is required")
	}

	if endDate.IsZero() {
		return errors.New("end date is required")
	}

	if endDate.Before(startDate) {
		return errors.New("end date cannot be before start date")
	}

	// Calculate inclusive leave days.
	totalDays := calculateLeaveDays(startDate, endDate)

	if totalDays <= 0 {
		return errors.New("leave duration must be greater than zero")
	}

	// Get employee leave balance for the year.
	year := startDate.Year()

	balance, err := s.repo.GetEmployeeLeaveBalance(
		ctx,
		employeeID,
		leaveTypeID,
		year,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to get leave balance: %w",
			err,
		)
	}

	// EmployeeLeaveBalance uses decimal.Decimal.
	requestedDays := decimal.NewFromFloat(totalDays)

	// Check available balance.
	if balance.RemainingDays.LessThan(requestedDays) {
		return fmt.Errorf(
			"insufficient leave balance: available %.2f days, requested %.2f days",
			balance.RemainingDays.InexactFloat64(),
			totalDays,
		)
	}

	// Create pending application.
	err = s.repo.ApplyLeave(
		ctx,
		employeeID,
		leaveTypeID,
		startDate,
		endDate,
		totalDays,
		reason,
	)
	if err != nil {
		return err
	}

	return nil
}

// calculateLeaveDays calculates inclusive leave days.
//
// Example:
//
// Start: 2026-08-10
// End:   2026-08-12
// Result: 3
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
		time.UTC,
	)

	end := time.Date(
		endDate.Year(),
		endDate.Month(),
		endDate.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	return end.Sub(start).Hours()/24 + 1
}

// GetEmployeeLeaveHistory returns leave applications
// for an employee.
func (s *Service) GetEmployeeLeaveHistory(
	ctx context.Context,
	employeeID string,
) ([]model.LeaveApplication, error) {

	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	return s.repo.GetEmployeeLeaveHistory(
		ctx,
		employeeID,
	)
}

// GetEmployeeLeaveBalances returns all leave balances
// for an employee for a specific year.
func (s *Service) GetEmployeeLeaveBalances(
	ctx context.Context,
	employeeID string,
	year int,
) ([]model.EmployeeLeaveBalance, error) {

	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	if year <= 0 {
		return nil, errors.New("invalid year")
	}

	return s.repo.GetEmployeeLeaveBalances(
		ctx,
		employeeID,
		year,
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

	if leaveApplicationID == "" {
		return errors.New("leave application ID cannot be empty")
	}

	return s.repo.CancelLeave(
		ctx,
		employeeID,
		leaveApplicationID,
	)
}

// ApproveLeave approves a pending leave application
// and updates the employee leave balance.
func (s *Service) ApproveLeave(
	ctx context.Context,
	leaveApplicationID string,
	approvedBy string,
) error {

	if leaveApplicationID == "" {
		return errors.New("leave application ID cannot be empty")
	}

	if approvedBy == "" {
		return errors.New("approvedBy cannot be empty")
	}

	// Get leave application.
	application, err := s.repo.GetLeaveApplication(
		ctx,
		leaveApplicationID,
	)
	if err != nil {
		return err
	}

	// Only Pending applications can be approved.
	if application.Status != "Pending" {
		return fmt.Errorf(
			"leave application is not pending, current status: %s",
			application.Status,
		)
	}

	// Get employee balance.
	balance, err := s.repo.GetEmployeeLeaveBalance(
		ctx,
		application.EmployeeId.String(),
		application.LeaveTypeId.String(),
		application.StartDate.Year(),
	)
	if err != nil {
		return fmt.Errorf("failed to get leave balance: %w", err)
	}

	// Check balance.
	if balance.RemainingDays.LessThan(application.TotalDays) {
		return fmt.Errorf(
			"insufficient leave balance: available %.2f days, requested %.2f days",
			balance.RemainingDays.InexactFloat64(),
			application.TotalDays.InexactFloat64(),
		)
	}

	// Calculate new balance.
	newUsedDays := balance.UsedDays.Add(
		application.TotalDays,
	)

	newRemainingDays := balance.RemainingDays.Sub(
		application.TotalDays,
	)

	// Update balance.
	err = s.repo.UpdateLeaveBalance(
		ctx,
		balance.ID.String(),
		newUsedDays.InexactFloat64(),
		newRemainingDays.InexactFloat64(),
	)
	if err != nil {
		return fmt.Errorf("failed to update leave balance: %w", err)
	}

	// Approve application.
	err = s.repo.ApproveLeave(
		ctx,
		leaveApplicationID,
		approvedBy,
	)
	if err != nil {
		return fmt.Errorf("failed to approve leave application: %w", err)
	}

	return nil
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

	if rejectedBy == "" {
		return errors.New("rejectedBy cannot be empty")
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
