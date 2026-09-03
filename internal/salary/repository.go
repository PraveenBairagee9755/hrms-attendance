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

// GetSalaryStructure returns the active salary structure
// for an employee on the requested date.
func (r *Repository) GetSalaryStructure(
	ctx context.Context,
	employeeID string,
	date time.Time,
) (*model.SalaryStructure, error) {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	var salary model.SalaryStructure

	employeeDate := Date(
		date.Year(),
		date.Month(),
		date.Day(),
	)

	stmt := SELECT(
		table.SalaryStructure.AllColumns,
	).FROM(
		table.SalaryStructure,
	).WHERE(
		table.SalaryStructure.EmployeeId.EQ(UUID(employeeUUID)).
			AND(
				table.SalaryStructure.EffectiveFrom.LT_EQ(employeeDate),
			).
			AND(
				table.SalaryStructure.EffectiveTo.IS_NULL().
					OR(
						table.SalaryStructure.EffectiveTo.GT_EQ(employeeDate),
					),
			),
	).ORDER_BY(
		table.SalaryStructure.EffectiveFrom.DESC(),
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
