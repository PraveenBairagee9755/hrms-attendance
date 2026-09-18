package attendance

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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
// ClockIn inserts today's attendance record.
func (r *Repository) ClockIn(
	ctx context.Context,
	employeeID string,
) (*ClockInResult, error) {

	loc, err := getAttendanceLocation()
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	// Get today's date according to office timezone.
	today := now.Format("2006-01-02")

	var existingID int

	err = r.DB.QueryRowContext(
		ctx,
		`
        SELECT id
        FROM public."Attendance"
        WHERE "employeeId" = $1
		 AND "date" = $2::date
        ORDER BY id DESC
        LIMIT 1
        `,
		employeeUUID,
		today,
	).Scan(&existingID)

	if err == nil {
		return nil, fmt.Errorf("employee has already clocked in today")
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check today's attendance: %w", err)
	}

	// Read shift start time.
	shiftStartText := os.Getenv("SHIFT_START_TIME")
	if shiftStartText == "" {
		shiftStartText = "10:00"
	}

	shiftStartTime, err := time.ParseInLocation(
		"15:04",
		shiftStartText,
		loc,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid SHIFT_START_TIME: %w", err)
	}

	// Create today's shift start timestamp.
	shiftStart := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		shiftStartTime.Hour(),
		shiftStartTime.Minute(),
		0,
		0,
		loc,
	)

	// Calculate ONLY one of lateBy or earlyBy.
	var lateBy string
	var earlyBy string

	if now.After(shiftStart) {
		lateBy = formatDuration(now.Sub(shiftStart))
	} else if now.Before(shiftStart) {
		earlyBy = formatDuration(shiftStart.Sub(now))
	}

	attendanceDate := Date(
		now.Year(),
		now.Month(),
		now.Day(),
	)

	// Insert attendance.
	stmt := table.Attendance.INSERT(
		table.Attendance.EmployeeId,
		table.Attendance.Date,
		table.Attendance.CheckInTime,
		table.Attendance.Status,
		table.Attendance.CreatedAt,
		table.Attendance.UpdatedAt,
	).VALUES(
		employeeUUID,
		attendanceDate,
		now,
		"Present",
		now,
		now,
	)

	_, err = stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return nil, fmt.Errorf("clock-in failed: %w", err)
	}

	return &ClockInResult{
		EmployeeID: employeeID,
		InTime:     now.Format("15:04:05"),
		LateBy:     lateBy,
		EarlyBy:    earlyBy,
	}, nil
}

// ClockOut updates today's attendance record,
// calculates total work hours, and stores checkout time.
func (r *Repository) ClockOut(
	ctx context.Context,
	employeeID string,
) (*ClockOutResult, error) {

	loc, err := getAttendanceLocation()
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	today := now.Format("2006-01-02")

	var (
		attendanceID int
		checkInTime  time.Time
		checkOutTime *time.Time
	)

	err = r.DB.QueryRowContext(
		ctx,
		`
        SELECT id, "checkInTime"
        FROM public."Attendance"
        WHERE "employeeId" = $1
		 AND date = $2::date
        ORDER BY id DESC
        LIMIT 1
        `,
		employeeUUID,
		today,
	).Scan(&attendanceID, &checkInTime)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no attendance record found for employee %s today", employeeID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to find today's attendance: %w", err)
	}

	if checkOutTime != nil {
		return nil, fmt.Errorf("employee has already clocked out today")
	}

	if now.Before(checkInTime) {
		return nil, fmt.Errorf("clock-out time cannot be before clock-in time")
	}

	// Calculate total working duration.
	workDuration := now.Sub(checkInTime)

	// Convert duration to decimal hours for database.
	workHours := workDuration.Hours()

	// Update attendance.
	stmt := table.Attendance.UPDATE(
		table.Attendance.CheckOutTime,
		table.Attendance.WorkHours,
		table.Attendance.UpdatedAt,
	).SET(
		now,
		workHours,
		now,
	).WHERE(
		table.Attendance.ID.EQ(
			Int32(int32(attendanceID)),
		),
	)

	result, err := stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return nil, fmt.Errorf("clock-out failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check clock-out result: %w", err)
	}

	if rows == 0 {
		return nil, fmt.Errorf("attendance record was not updated")
	}

	return &ClockOutResult{
		EmployeeID:     employeeID,
		OutTime:        now.Format("15:04:05"),
		TotalWorkHours: formatDuration(workDuration),
	}, nil
}

// getAttendanceLocation returns the configured office timezone.
func getAttendanceLocation() (*time.Location, error) {

	timezone := os.Getenv("APP_TIMEZONE")

	if timezone == "" {
		timezone = "Asia/Kolkata"
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid APP_TIMEZONE %q: %w",
			timezone,
			err,
		)
	}

	return loc, nil
}

// formatDuration formats a duration as HH:MM:SS.
func formatDuration(d time.Duration) string {

	if d < 0 {
		d = -d
	}

	totalSeconds := int64(d.Seconds())

	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	return fmt.Sprintf(
		"%02d:%02d:%02d",
		hours,
		minutes,
		seconds,
	)
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

func (r *Repository) GetEmployeeAttendance(
	ctx context.Context,
	employeeID string,
	fromDate string,
	toDate string,
) ([]model.Attendance, error) {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	query := `
        SELECT
            id,
            "employeeId",
            "date",
            "checkInTime",
            "checkOutTime",
            status,
            "workHours",
            remarks,
            "createdAt",
            "updatedAt",
            "markedBy"
        FROM public."Attendance"
        WHERE "employeeId" = $1
    `

	args := []interface{}{employeeUUID}

	if fromDate != "" {
		query += ` AND "date" >= $2::date`
		args = append(args, fromDate)
	}

	if toDate != "" {
		if fromDate != "" {
			query += ` AND "date" <= $3::date`
		} else {
			query += ` AND "date" <= $2::date`
		}
		args = append(args, toDate)
	}

	query += ` ORDER BY "date" DESC, id DESC`

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch attendance: %w", err)
	}
	defer rows.Close()

	var attendance []model.Attendance

	for rows.Next() {
		var a model.Attendance

		err := rows.Scan(
			&a.ID,
			&a.EmployeeId,
			&a.Date,
			&a.CheckInTime,
			&a.CheckOutTime,
			&a.Status,
			&a.WorkHours,
			&a.Remarks,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.MarkedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attendance: %w", err)
		}

		attendance = append(attendance, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading attendance: %w", err)
	}

	return attendance, nil
}
