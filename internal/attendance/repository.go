package attendance

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

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		DB: db,
	}
}

// ClockIn inserts today's attendance record.
func (r *Repository) ClockIn(
	ctx context.Context,
	employeeID string,
) error {

	now := time.Now()

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	stmt := table.Attendance.INSERT(
		table.Attendance.EmployeeId,
		table.Attendance.Date,
		table.Attendance.CheckInTime,
		table.Attendance.Status,
		table.Attendance.CreatedAt,
		table.Attendance.UpdatedAt,
	).VALUES(
		employeeUUID,
		Date(
			now.Year(),
			now.Month(),
			now.Day(),
		),
		now,
		"Present",
		now,
		now,
	)

	_, err = stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return fmt.Errorf("clock-in failed: %w", err)
	}

	return nil
}

// ClockOut updates today's attendance record
// with checkout time and updatedAt.
func (r *Repository) ClockOut(
	ctx context.Context,
	employeeID string,
) error {

	now := time.Now()

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	stmt := table.Attendance.UPDATE(
		table.Attendance.CheckOutTime,
		table.Attendance.UpdatedAt,
	).SET(
		now,
		now,
	).WHERE(
		table.Attendance.EmployeeId.EQ(
			UUID(employeeUUID),
		).
			AND(
				table.Attendance.Date.EQ(
					Date(
						now.Year(),
						now.Month(),
						now.Day(),
					),
				),
			),
	)

	result, err := stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return fmt.Errorf("clock-out failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check clock-out result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf(
			"no attendance record found for employee %s today",
			employeeID,
		)
	}

	return nil
}

// GetEmployeeHistory returns attendance records for an employee
// between the given start and end dates.
func (r *Repository) GetEmployeeHistory(
	ctx context.Context,
	employeeID string,
	startDate time.Time,
	endDate time.Time,
) ([]model.Attendance, error) {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	var records []model.Attendance

	stmt := SELECT(
		table.Attendance.AllColumns,
	).FROM(
		table.Attendance,
	).WHERE(
		table.Attendance.EmployeeId.EQ(
			UUID(employeeUUID),
		).
			AND(
				table.Attendance.Date.BETWEEN(
					Date(
						startDate.Year(),
						startDate.Month(),
						startDate.Day(),
					),
					Date(
						endDate.Year(),
						endDate.Month(),
						endDate.Day(),
					),
				),
			),
	).ORDER_BY(
		table.Attendance.Date.DESC(),
	)

	err = stmt.QueryContext(ctx, r.DB, &records)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get employee attendance history: %w",
			err,
		)
	}

	return records, nil
}

// ImportAttendance inserts one attendance record imported from Excel.
// ImportAttendance inserts or updates an attendance record imported from Excel.
func (r *Repository) ImportAttendance(
	ctx context.Context,
	data AttendanceExcelRow,
) error {

	employeeUUID, err := uuid.Parse(data.EmployeeID)
	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	now := time.Now()

	stmt := table.Attendance.INSERT(
		table.Attendance.EmployeeId,
		table.Attendance.Date,
		table.Attendance.CheckInTime,
		table.Attendance.CheckOutTime,
		table.Attendance.Status,
		table.Attendance.CreatedAt,
		table.Attendance.UpdatedAt,
	).VALUES(
		employeeUUID,
		Date(
			data.Date.Year(),
			data.Date.Month(),
			data.Date.Day(),
		),
		data.CheckInTime,
		data.CheckOutTime,
		data.Status,
		now,
		now,
	).ON_CONFLICT(
		table.Attendance.EmployeeId,
		table.Attendance.Date,
	).DO_UPDATE(
		SET(
			table.Attendance.CheckInTime.SET(
				table.Attendance.EXCLUDED.CheckInTime,
			),
			table.Attendance.CheckOutTime.SET(
				table.Attendance.EXCLUDED.CheckOutTime,
			),
			table.Attendance.Status.SET(
				table.Attendance.EXCLUDED.Status,
			),
			table.Attendance.UpdatedAt.SET(
				table.Attendance.EXCLUDED.UpdatedAt,
			),
		),
	)

	_, err = stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return fmt.Errorf("failed to import attendance: %w", err)
	}

	return nil
}
