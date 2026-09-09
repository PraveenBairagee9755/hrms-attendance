package attendance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"hrms-attendance/db_gen/public/model"
	"hrms-attendance/db_gen/public/table"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

// CreateRegularization inserts a new Pending request.
func (r *Repository) CreateRegularization(
	ctx context.Context,
	employeeID string,
	attendanceID int32,
	requestedCheckIn *time.Time,
	requestedCheckOut *time.Time,
	reason string,
) error {

	employeeUUID, err := uuid.Parse(employeeID)

	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	// Make sure Attendance exists and belongs to employee.
	var attendanceEmployeeID uuid.UUID

	err = r.DB.QueryRowContext(
		ctx,
		`
		SELECT "employeeId"
		FROM public."Attendance"
		WHERE id = $1
		`,
		attendanceID,
	).Scan(&attendanceEmployeeID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("attendance record %d not found", attendanceID)
	}

	if err != nil {
		return fmt.Errorf("failed to validate attendance: %w", err)
	}

	if attendanceEmployeeID != employeeUUID {
		return fmt.Errorf("attendance record does not belong to employee %s", employeeID)
	}

	// Prevent another Pending or Approved request
	// for the same Attendance record.
	var existingID uuid.UUID

	err = r.DB.QueryRowContext(
		ctx,
		`
		SELECT id
		FROM public."Regularization"
		WHERE "attendanceId" = $1
		  AND status IN ('Pending', 'Approved')
		LIMIT 1
		`,
		attendanceID,
	).Scan(&existingID)

	if err == nil {
		return fmt.Errorf("a Pending or Approved regularization already exists for attendance %d", attendanceID)
	}

	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check duplicate regularization: %w", err)
	}

	now := time.Now()

	stmt := table.Regularization.INSERT(
		table.Regularization.EmployeeId,
		table.Regularization.AttendanceId,
		table.Regularization.RequestedCheckIn,
		table.Regularization.RequestedCheckOut,
		table.Regularization.Reason,
		table.Regularization.Status,
		table.Regularization.CreatedAt,
		table.Regularization.UpdatedAt,
	).VALUES(
		employeeUUID,
		attendanceID,
		requestedCheckIn,
		requestedCheckOut,
		reason,
		"Pending",
		now,
		now,
	)

	_, err = stmt.ExecContext(ctx, r.DB)

	if err != nil {
		return fmt.Errorf("failed to create regularization: %w", err)
	}

	return nil
}

// GetRegularizationHistory returns an employee's regularizations.
func (r *Repository) GetRegularizationHistory(
	ctx context.Context,
	employeeID string,
) ([]model.Regularization, error) {

	employeeUUID, err := uuid.Parse(employeeID)

	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	var records []model.Regularization

	stmt := SELECT(
		table.Regularization.AllColumns,
	).FROM(
		table.Regularization,
	).WHERE(
		table.Regularization.EmployeeId.EQ(
			UUID(employeeUUID),
		),
	).ORDER_BY(
		table.Regularization.CreatedAt.DESC(),
	)

	err = stmt.QueryContext(
		ctx,
		r.DB,
		&records,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get regularization history: %w", err)
	}

	return records, nil
}

// ApproveRegularization updates Regularization and Attendance
// in one database transaction.
func (r *Repository) ApproveRegularization(
	ctx context.Context,
	regularizationID string,
	approvedBy string,
) error {

	regID, err := uuid.Parse(regularizationID)

	if err != nil {
		return fmt.Errorf("invalid regularization ID: %w", err)
	}

	approvedByUUID, err := uuid.Parse(approvedBy)

	if err != nil {
		return fmt.Errorf("invalid approvedBy UUID: %w", err)
	}

	tx, err := r.DB.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("failed to begin approval transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	var (
		employeeID        uuid.UUID
		attendanceID      int32
		requestedCheckIn  *time.Time
		requestedCheckOut *time.Time
		status            string
	)

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT
			"employeeId",
			"attendanceId",
			"requestedCheckIn",
			"requestedCheckOut",
			status
		FROM public."Regularization"
		WHERE id = $1
		FOR UPDATE
		`,
		regID,
	).Scan(
		&employeeID,
		&attendanceID,
		&requestedCheckIn,
		&requestedCheckOut,
		&status,
	)

	if err == sql.ErrNoRows {
		return fmt.Errorf("regularization %s not found", regularizationID)
	}

	if err != nil {
		return fmt.Errorf("failed to get regularization: %w", err)
	}

	if status != "Pending" {
		return fmt.Errorf("regularization cannot be approved because current status is %s", status)
	}

	// Lock and validate the Attendance record.
	var attendanceEmployeeID uuid.UUID

	err = tx.QueryRowContext(
		ctx,
		`
		SELECT "employeeId"
		FROM public."Attendance"
		WHERE id = $1
		FOR UPDATE
		`,
		attendanceID,
	).Scan(&attendanceEmployeeID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("attendance record %d not found", attendanceID)
	}

	if err != nil {
		return fmt.Errorf("failed to get attendance record: %w", err)
	}

	if attendanceEmployeeID != employeeID {
		return errors.New("attendance employee does not match regularization employee")
	}

	now := time.Now()

	// Update Attendance only with the values provided
	// in the regularization request.
	_, err = tx.ExecContext(
		ctx,
		`
		UPDATE public."Attendance"
		SET
			"checkInTime" = COALESCE($1, "checkInTime"),
			"checkOutTime" = COALESCE($2, "checkOutTime"),
			status = 'Present',
			"updatedAt" = $3
		WHERE id = $4
		`,
		requestedCheckIn,
		requestedCheckOut,
		now,
		attendanceID,
	)

	if err != nil {
		return fmt.Errorf("failed to update attendance: %w", err)
	}

	// Mark Regularization as Approved.
	_, err = tx.ExecContext(
		ctx,
		`
	UPDATE public."Regularization"
	SET
		status = 'Approved',
		"approvedBy" = $1,
		"approvedAt" = $2,
		"updatedAt" = $3
	WHERE id = $4
	`,
		approvedByUUID,
		now,
		now,
		regID,
	)

	if err != nil {
		return fmt.Errorf("failed to approve regularization: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit regularization approval: %w", err)
	}

	return nil
}

// RejectRegularization rejects a Pending request.
func (r *Repository) RejectRegularization(
	ctx context.Context,
	regularizationID string,
	rejectedBy string,
	rejectionReason string,
) error {

	regID, err := uuid.Parse(regularizationID)

	if err != nil {
		return fmt.Errorf("invalid regularization ID: %w", err)
	}

	// rejectedBy is currently validated for API consistency.
	// The current table has no rejectedBy column, so it is not stored.
	if _, err := uuid.Parse(rejectedBy); err != nil {
		return fmt.Errorf("invalid rejectedBy UUID: %w", err)
	}

	now := time.Now()

	result, err := r.DB.ExecContext(
		ctx,
		`
		UPDATE public."Regularization"
		SET
			status = 'Rejected',
			"rejectionReason" = $1,
			"updatedAt" = $2
		WHERE id = $3
		  AND status = 'Pending'
		`,
		rejectionReason,
		now,
		regID,
	)

	if err != nil {
		return fmt.Errorf("failed to reject regularization: %w", err)
	}

	rows, err := result.RowsAffected()

	if err != nil {
		return fmt.Errorf("failed to check rejection result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("regularization not found or is not Pending")
	}

	return nil
}
