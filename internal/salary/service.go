package salary

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

    "hrms-attendance/db_gen/public/model"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// SalaryCalculation contains the final salary calculation.
type SalaryCalculation struct {
	EmployeeID       string       `json:"employeeId"`
	Year             int          `json:"year"`
	Month            int          `json:"month"`
	MonthlySalary    float64      `json:"monthlySalary"`
	TotalLOPDays     float64      `json:"totalLOPDays"`
	DailySalary      float64      `json:"dailySalary"`
	LOPDeduction     float64      `json:"lopDeduction"`
	CalculatedSalary float64      `json:"calculatedSalary"`
	LeaveUsage       []LeaveUsage `json:"leaveUsage"`
}

// CalculateSalary calculates salary after leave-limit deductions.
func (s *Service) CalculateSalary(
	ctx context.Context,
	employeeID string,
	year int,
	month time.Month,
) (*SalaryCalculation, error) {

	// Validate employee ID.
	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return nil, errors.New("invalid employee ID")
	}

	// Validate month.
	if month < time.January || month > time.December {
		return nil, errors.New("invalid month")
	}

	// Get employee salary structure.
	salaryStructure, err := s.repo.GetSalaryStructure(
		ctx,
		employeeID,
	)

	if err != nil {
		return nil, err
	}

	// Get approved leave usage for the year.
	leaveUsage, err := s.repo.GetApprovedLeaveUsage(
		ctx,
		employeeID,
		year,
		month,
	)

	if err != nil {
		return nil, err
	}

	// --------------------------------
	// Calculate LOP days
	// --------------------------------

	var totalLOPDays float64

	for i := range leaveUsage {
		usedDays := leaveUsage[i].UsedDays

		// Loss Of Pay leave is always unpaid.
		if leaveUsage[i].LeaveTypeName == "Loss Of Pay" {
			leaveUsage[i].LOPDays = usedDays
		} else {
			// Sick/Casual/Birthday leave within the
			// employee's balance is paid leave.
			leaveUsage[i].LOPDays = 0
		}

		totalLOPDays += leaveUsage[i].LOPDays
	}

	// --------------------------------
	// Get salary
	// --------------------------------

	// Jet generated GrossSalary as decimal.Decimal.
	monthlySalaryDecimal := salaryStructure.GrossSalary

	// If gross salary is zero, use CTC per month.
	if monthlySalaryDecimal.LessThanOrEqual(decimal.Zero) &&
		salaryStructure.CtcPerMonth != nil {
		monthlySalaryDecimal = *salaryStructure.CtcPerMonth
	}

	if monthlySalaryDecimal.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("salary structure does not contain a valid salary")
	}

	// Convert Decimal to float64 only for the final calculation.
	monthlySalary, _ := monthlySalaryDecimal.Float64()

	// --------------------------------
	// Days in month
	// --------------------------------

	daysInMonth := daysInMonth(year, month)

	// --------------------------------
	// Daily salary
	// --------------------------------

	dailySalary := monthlySalary / float64(daysInMonth)

	// --------------------------------
	// LOP deduction
	// --------------------------------

	lopDeduction := dailySalary * totalLOPDays

	// Don't allow deduction greater than salary.
	if lopDeduction > monthlySalary {
		lopDeduction = monthlySalary
	}

	// --------------------------------
	// Final salary
	// --------------------------------

	calculatedSalary := monthlySalary - lopDeduction

	return &SalaryCalculation{
		EmployeeID:       employeeID,
		Year:             year,
		Month:            int(month),
		MonthlySalary:    monthlySalary,
		TotalLOPDays:     totalLOPDays,
		DailySalary:      dailySalary,
		LOPDeduction:     lopDeduction,
		CalculatedSalary: calculatedSalary,
		LeaveUsage:       leaveUsage,
	}, nil
}

// daysInMonth returns the number of days in a month.
func daysInMonth(year int, month time.Month) int {

	firstDayNextMonth := time.Date(
		year,
		month+1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	lastDay := firstDayNextMonth.AddDate(
		0,
		0,
		-1,
	)

	return lastDay.Day()
}

// String representation for debugging.
func (s *SalaryCalculation) String() string {

	return fmt.Sprintf(
		"Employee=%s MonthlySalary=%.2f LOPDays=%.2f LOPDeduction=%.2f FinalSalary=%.2f",
		s.EmployeeID,
		s.MonthlySalary,
		s.TotalLOPDays,
		s.LOPDeduction,
		s.CalculatedSalary,
	)
}

func (s *Service) GetSalaryDetails(
    ctx context.Context,
    employeeID string,
) (*model.SalaryStructure, error) {

    if employeeID == "" {
        return nil, errors.New("employee ID cannot be empty")
    }

    if _, err := uuid.Parse(employeeID); err != nil {
        return nil, errors.New("invalid employee ID")
    }

    return s.repo.GetSalaryDetails(ctx, employeeID)
}