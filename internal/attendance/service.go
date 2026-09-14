package attendance

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"hrms-attendance/db_gen/public/model"
)

// ClockInResult contains the result of a successful clock-in.
type ClockInResult struct {
	EmployeeID string `json:"employeeId"`
	InTime     string `json:"inTime"`
	LateBy     string `json:"lateBy,omitempty"`
	EarlyBy    string `json:"earlyBy,omitempty"`
}

// ClockOutResult contains the result of a successful clock-out.
type ClockOutResult struct {
	EmployeeID     string `json:"employeeId"`
	OutTime        string `json:"outTime"`
	TotalWorkHours string `json:"totalWorkHours"`
}

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
func (s *Service) ClockIn(
	ctx context.Context,
	employeeID string,
) (*ClockInResult, error) {

	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return nil, errors.New("invalid employee ID")
	}

	return s.repo.ClockIn(ctx, employeeID)
}

// ClockOut validates and processes an employee's clock-out.
func (s *Service) ClockOut(
	ctx context.Context,
	employeeID string,
) (*ClockOutResult, error) {

	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return nil, errors.New("invalid employee ID")
	}

	return s.repo.ClockOut(ctx, employeeID)
}

func (s *Service) GetEmployeeAttendance(
	ctx context.Context,
	employeeID string,
	fromDate string,
	toDate string,
) ([]model.Attendance, error) {

	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return nil, errors.New("invalid employee ID")
	}

	if fromDate != "" {
		if _, err := time.Parse("2006-01-02", fromDate); err != nil {
			return nil, errors.New("invalid fromDate, use YYYY-MM-DD")
		}
	}

	if toDate != "" {
		if _, err := time.Parse("2006-01-02", toDate); err != nil {
			return nil, errors.New("invalid toDate, use YYYY-MM-DD")
		}
	}

	if fromDate != "" && toDate != "" && fromDate > toDate {
		return nil, errors.New("fromDate cannot be after toDate")
	}

	return s.repo.GetEmployeeAttendance(
		ctx,
		employeeID,
		fromDate,
		toDate,
	)
}
