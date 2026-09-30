package salary

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"hrms-attendance/db_gen/public/model"
	"hrms-attendance/db_gen/public/table"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/go-jet/jet/v2/postgres"
)

type LeaveUsage struct {
	LeaveTypeID    uuid.UUID `alias:"leaveTypeId"`
	LeaveTypeName  string    `alias:"leaveTypeName"`
	MaxDaysPerYear int       `alias:"maxDaysPerYear"`
	UsedDays       float64   `alias:"usedDays"`
	RemainingDays  float64
	LOPDays        float64
}

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

// GetSalaryStructure returns the latest salary structure
// for an employee.
func (r *Repository) GetSalaryStructure(
	ctx context.Context,
	employeeID string,
) (*model.SalaryStructure, error) {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	var salary model.SalaryStructure

	stmt := SELECT(
		table.SalaryStructure.AllColumns,
	).FROM(
		table.SalaryStructure,
	).WHERE(
		table.SalaryStructure.EmployeeId.EQ(UUID(employeeUUID)),
	).ORDER_BY(
		table.SalaryStructure.CreatedAt.DESC(),
	).LIMIT(1)

	err = stmt.QueryContext(ctx, r.DB, &salary)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get salary structure: %w",
			err,
		)
	}

	return &salary, nil
}

func (r *Repository) GetApprovedLeaveUsage(
	ctx context.Context,
	employeeID string,
	year int,
	month time.Month,
) ([]LeaveUsage, error) {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	startOfMonth := time.Date(
		year,
		month,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)

	query := `
		SELECT
			lt.id,
			lt.name,
			lt."maxDaysPerYear",
			COALESCE(SUM(la."totalDays"), 0)
		FROM public."LeaveApplication" la
		INNER JOIN public."LeaveType" lt
			ON la."leaveTypeId" = lt.id
		WHERE la."employeeId" = $1
			AND la.status = 'Approved'
			AND la."startDate" < $2
			AND la."endDate" >= $3
		GROUP BY
			lt.id,
			lt.name,
			lt."maxDaysPerYear"
		ORDER BY lt.name ASC
	`

	rows, err := r.DB.QueryContext(
		ctx,
		query,
		employeeUUID,
		startOfNextMonth,
		startOfMonth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query approved leave usage: %w", err)
	}
	defer rows.Close()

	var usage []LeaveUsage

	for rows.Next() {

		var item LeaveUsage

		err := rows.Scan(
			&item.LeaveTypeID,
			&item.LeaveTypeName,
			&item.MaxDaysPerYear,
			&item.UsedDays,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan leave usage: %w", err)
		}

		usage = append(usage, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading leave usage: %w", err)
	}

	return usage, nil
}

// ImportSalaryStructure inserts one salary structure imported from Excel.
func (r *Repository) ImportSalaryStructure(
	ctx context.Context,
	data SalaryStructureExcelRow,
) error {

	employeeUUID, err := uuid.Parse(data.EmployeeID)
	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	now := time.Now()
	salaryStructureID := uuid.New()

	stmt := table.SalaryStructure.INSERT(
		table.SalaryStructure.ID,
		table.SalaryStructure.EmployeeId,
		table.SalaryStructure.EmployeeName,
		table.SalaryStructure.DateOfJoining,
		table.SalaryStructure.CtcPerMonth,
		table.SalaryStructure.DaysPaid,
		table.SalaryStructure.Salary,
		table.SalaryStructure.Late,
		table.SalaryStructure.Incentive,
		table.SalaryStructure.Conv,
		table.SalaryStructure.Advance,
		table.SalaryStructure.DeductionAmount,
		table.SalaryStructure.AccountNo,
		table.SalaryStructure.Ifsc,
		table.SalaryStructure.PfDeduction,
		table.SalaryStructure.PfUanNumber,
		table.SalaryStructure.GrossSalary,
		table.SalaryStructure.CreatedAt,
		table.SalaryStructure.UpdatedAt,
	).VALUES(
		salaryStructureID,
		employeeUUID,
		data.EmployeeName,
		data.DateOfJoining,
		data.CtcPerMonth,
		data.DaysPaid,
		data.Salary,
		data.Late,
		data.Incentive,
		data.Conv,
		data.Advance,
		data.DeductionAmount,
		data.AccountNo,
		data.Ifsc,
		data.PfDeduction,
		data.PfUanNumber,
		data.GrossSalary,
		now,
		now,
	)

	_, err = stmt.ExecContext(ctx, r.DB)

	if err != nil {
		return fmt.Errorf("failed to import salary structure: %w", err)
	}

	return nil
}

func (r *Repository) GetSalaryDetails(
    ctx context.Context,
    employeeID string,
) (*model.SalaryStructure, error) {

    // Validate that employeeId is a UUID.
    employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
        return nil, fmt.Errorf("invalid employee ID: %w", err)
    }

    var result model.SalaryStructure

    stmt := table.SalaryStructure.
        SELECT(
            table.SalaryStructure.ID,
            table.SalaryStructure.EmployeeId,
            table.SalaryStructure.EmployeeName,
            table.SalaryStructure.DateOfJoining,
            table.SalaryStructure.CtcPerMonth,
            table.SalaryStructure.DaysPaid,
            table.SalaryStructure.Salary,
            table.SalaryStructure.Late,
            table.SalaryStructure.Incentive,
            table.SalaryStructure.Conv,
            table.SalaryStructure.Advance,
            table.SalaryStructure.DeductionAmount,
            table.SalaryStructure.AccountNo,
            table.SalaryStructure.Ifsc,
            table.SalaryStructure.PfDeduction,
            table.SalaryStructure.PfUanNumber,
            table.SalaryStructure.GrossSalary,
            table.SalaryStructure.CreatedAt,
            table.SalaryStructure.UpdatedAt,
        ).
        FROM(table.SalaryStructure).
        WHERE(
			table.SalaryStructure.EmployeeId.EQ(
				postgres.UUID(employeeUUID),
            ),
        )

    err = stmt.QueryContext(ctx, r.DB, &result)

    if err != nil {
        return nil, fmt.Errorf("salary details not found: %w", err)
    }

    return &result, nil
}
