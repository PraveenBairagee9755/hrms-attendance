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

// LeaveUsage represents approved leave usage for one leave type.
type LeaveUsage struct {
	LeaveTypeID    string  `json:"leaveTypeId"`
	LeaveTypeName  string  `json:"leaveTypeName"`
	MaxDaysPerYear int     `json:"maxDaysPerYear"`
	UsedDays       float64 `json:"usedDays"`
	LOPDays        float64 `json:"lopDays"`
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

// GetApprovedLeaveUsage returns approved leave usage grouped by leave type
// for the requested year.
func (r *Repository) GetApprovedLeaveUsage(
	ctx context.Context,
	employeeID string,
	year int,
) ([]LeaveUsage, error) {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	startDate := Date(
		year,
		time.January,
		1,
	)

	endDate := Date(
		year,
		time.December,
		31,
	)

	var usage []LeaveUsage

	stmt := SELECT(
		table.LeaveType.ID.AS("leaveTypeId"),
		table.LeaveType.Name.AS("leaveTypeName"),
		table.LeaveType.MaxDaysPerYear.AS("maxDaysPerYear"),

		COALESCE(
			SUM(table.LeaveApplication.TotalDays),
			Float(0),
		).AS("usedDays"),
	).FROM(
		table.LeaveType,
		table.LeaveApplication,
	).WHERE(
		table.LeaveType.ID.EQ(
			table.LeaveApplication.LeaveTypeId,
		).
			AND(
				table.LeaveApplication.EmployeeId.EQ(UUID(employeeUUID)),
			).
			AND(
				table.LeaveApplication.Status.EQ(String("Approved")),
			).
			AND(
				table.LeaveApplication.StartDate.LT_EQ(endDate),
			).
			AND(
				table.LeaveApplication.EndDate.GT_EQ(startDate),
			),
	).GROUP_BY(
		table.LeaveType.ID,
		table.LeaveType.Name,
		table.LeaveType.MaxDaysPerYear,
	).ORDER_BY(
		table.LeaveType.Name.ASC(),
	)

	err = stmt.QueryContext(
		ctx,
		r.DB,
		&usage,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to get approved leave usage: %w",
			err,
		)
	}

	return usage, nil
}
