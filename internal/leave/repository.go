package leave

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

// GetLeaveTypes returns all active leave types.
func (r *Repository) GetLeaveTypes(ctx context.Context) ([]model.LeaveType, error) {

	var leaveTypes []model.LeaveType

	stmt := SELECT(
		table.LeaveType.AllColumns,
	).FROM(
		table.LeaveType,
	).WHERE(
		table.LeaveType.IsActive.EQ(Bool(true)),
	).ORDER_BY(
		table.LeaveType.Name.ASC(),
	)

	err := stmt.QueryContext(ctx, r.DB, &leaveTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave types: %w", err)
	}

	return leaveTypes, nil
}

// ApplyLeave creates a new leave application.
func (r *Repository) ApplyLeave(
	ctx context.Context,
	employeeID string,
	leaveTypeID string,
	startDate time.Time,
	endDate time.Time,
	totalDays float64,
	reason string,
) error {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	leaveTypeUUID, err := uuid.Parse(leaveTypeID)
	if err != nil {
		return fmt.Errorf("invalid leave type ID: %w", err)
	}

	stmt := table.LeaveApplication.INSERT(
		table.LeaveApplication.EmployeeId,
		table.LeaveApplication.LeaveTypeId,
		table.LeaveApplication.StartDate,
		table.LeaveApplication.EndDate,
		table.LeaveApplication.TotalDays,
		table.LeaveApplication.Reason,
	).VALUES(
		UUID(employeeUUID),
		UUID(leaveTypeUUID),
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
		totalDays,
		reason,
	)

	_, err = stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return fmt.Errorf("failed to apply leave: %w", err)
	}

	return nil
}

// GetEmployeeLeaveHistory returns leave applications
// for an employee.
func (r *Repository) GetEmployeeLeaveHistory(
	ctx context.Context,
	employeeID string,
) ([]model.LeaveApplication, error) {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return nil, fmt.Errorf("invalid employee ID: %w", err)
	}

	var applications []model.LeaveApplication

	stmt := SELECT(
		table.LeaveApplication.AllColumns,
	).FROM(
		table.LeaveApplication,
	).WHERE(
		table.LeaveApplication.EmployeeId.EQ(
			UUID(employeeUUID),
		),
	).ORDER_BY(
		table.LeaveApplication.AppliedAt.DESC(),
	)

	err = stmt.QueryContext(ctx, r.DB, &applications)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee leave history: %w", err)
	}

	return applications, nil
}

// CancelLeave cancels a pending leave application.
func (r *Repository) CancelLeave(
	ctx context.Context,
	employeeID string,
	leaveApplicationID string,
) error {

	employeeUUID, err := uuid.Parse(employeeID)
	if err != nil {
		return fmt.Errorf("invalid employee ID: %w", err)
	}

	applicationUUID, err := uuid.Parse(leaveApplicationID)
	if err != nil {
		return fmt.Errorf("invalid leave application ID: %w", err)
	}

	now := time.Now()

	stmt := table.LeaveApplication.UPDATE(
		table.LeaveApplication.Status,
		table.LeaveApplication.CancelledAt,
		table.LeaveApplication.CancelledBy,
	).SET(
		"Cancelled",
		now,
		UUID(employeeUUID),
	).WHERE(
		table.LeaveApplication.ID.EQ(
			UUID(applicationUUID),
		).
			AND(
				table.LeaveApplication.EmployeeId.EQ(
					UUID(employeeUUID),
				),
			).
			AND(
				table.LeaveApplication.Status.EQ(
					String("Pending"),
				),
			),
	)

	result, err := stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return fmt.Errorf("failed to cancel leave: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check cancellation result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("leave application not found or cannot be cancelled")
	}

	return nil
}

// ApproveLeave approves a pending leave application.
func (r *Repository) ApproveLeave(
	ctx context.Context,
	leaveApplicationID string,
	approvedBy string,
) error {

	applicationUUID, err := uuid.Parse(leaveApplicationID)
	if err != nil {
		return fmt.Errorf("invalid leave application ID: %w", err)
	}

	approverUUID, err := uuid.Parse(approvedBy)
	if err != nil {
		return fmt.Errorf("invalid approver ID: %w", err)
	}

	now := time.Now()

	stmt := table.LeaveApplication.UPDATE(
		table.LeaveApplication.Status,
		table.LeaveApplication.ApprovedBy,
		table.LeaveApplication.ApprovedAt,
	).SET(
		"Approved",
		UUID(approverUUID),
		now,
	).WHERE(
		table.LeaveApplication.ID.EQ(
			UUID(applicationUUID),
		).
			AND(
				table.LeaveApplication.Status.EQ(
					String("Pending"),
				),
			),
	)

	result, err := stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return fmt.Errorf("failed to approve leave: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check approval result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("leave application not found or is not pending")
	}

	return nil
}

// RejectLeave rejects a pending leave application.
func (r *Repository) RejectLeave(
	ctx context.Context,
	leaveApplicationID string,
	rejectedBy string,
	rejectionReason string,
) error {

	applicationUUID, err := uuid.Parse(leaveApplicationID)
	if err != nil {
		return fmt.Errorf("invalid leave application ID: %w", err)
	}

	rejectorUUID, err := uuid.Parse(rejectedBy)
	if err != nil {
		return fmt.Errorf("invalid rejector ID: %w", err)
	}

	stmt := table.LeaveApplication.UPDATE(
		table.LeaveApplication.Status,
		table.LeaveApplication.ApprovedBy,
		table.LeaveApplication.RejectionReason,
	).SET(
		"Rejected",
		UUID(rejectorUUID),
		rejectionReason,
	).WHERE(
		table.LeaveApplication.ID.EQ(
			UUID(applicationUUID),
		).
			AND(
				table.LeaveApplication.Status.EQ(
					String("Pending"),
				),
			),
	)

	result, err := stmt.ExecContext(ctx, r.DB)
	if err != nil {
		return fmt.Errorf("failed to reject leave: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rejection result: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("leave application not found or is not pending")
	}

	return nil
}
