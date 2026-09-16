package leave

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"hrms-attendance/db_gen/public/model"
)

type PolicyRepository struct {
	DB *sql.DB
}

func NewPolicyRepository(db *sql.DB) *PolicyRepository {
	return &PolicyRepository{
		DB: db,
	}
}

// GetPolicies returns all leave policies from PostgreSQL.
func (r *PolicyRepository) GetPolicies(
	ctx context.Context,
) ([]model.LeavePolicy, error) {

	rows, err := r.DB.QueryContext(
		ctx,
		`
		SELECT
			id,
			"name",
			"leaveTypeId",
			"annualLimit",
			"maxConsecutiveDays",
			"carryForward",
			"maxCarryForwardDays",
			"requiresApproval",
			"effectiveFrom",
			"status",
			"createdAt",
			"updatedAt"
		FROM public."LeavePolicy"
		ORDER BY "effectiveFrom" DESC, "createdAt" DESC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get leave policies: %w", err)
	}
	defer rows.Close()

	policies := make([]model.LeavePolicy, 0)

	for rows.Next() {
		var policy model.LeavePolicy

		err := rows.Scan(
			&policy.ID,
			&policy.Name,
			&policy.LeaveTypeId,
			&policy.AnnualLimit,
			&policy.MaxConsecutiveDays,
			&policy.CarryForward,
			&policy.MaxCarryForwardDays,
			&policy.RequiresApproval,
			&policy.EffectiveFrom,
			&policy.Status,
			&policy.CreatedAt,
			&policy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan leave policy: %w", err)
		}

		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while reading leave policies: %w", err)
	}

	return policies, nil
}

// GetPolicy returns one leave policy by ID.
func (r *PolicyRepository) GetPolicy(
	ctx context.Context,
	id string,
) (*model.LeavePolicy, error) {

	policyID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid policy ID")
	}

	var policy model.LeavePolicy

	err = r.DB.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			"name",
			"leaveTypeId",
			"annualLimit",
			"maxConsecutiveDays",
			"carryForward",
			"maxCarryForwardDays",
			"requiresApproval",
			"effectiveFrom",
			"status",
			"createdAt",
			"updatedAt"
		FROM public."LeavePolicy"
		WHERE id = $1
		`,
		policyID,
	).Scan(
		&policy.ID,
		&policy.Name,
		&policy.LeaveTypeId,
		&policy.AnnualLimit,
		&policy.MaxConsecutiveDays,
		&policy.CarryForward,
		&policy.MaxCarryForwardDays,
		&policy.RequiresApproval,
		&policy.EffectiveFrom,
		&policy.Status,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("leave policy not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get leave policy: %w", err)
	}

	return &policy, nil
}

func (r *PolicyRepository) CreatePolicy(
	ctx context.Context,
	policy model.LeavePolicy,
) (*model.LeavePolicy, error) {

	err := r.DB.QueryRowContext(
		ctx,
		`
        INSERT INTO public."LeavePolicy" (
            id,
            "name",
            "leaveTypeId",
            "annualLimit",
            "maxConsecutiveDays",
            "carryForward",
            "maxCarryForwardDays",
            "requiresApproval",
            "effectiveFrom",
            "status",
            "createdAt",
            "updatedAt"
        )
        VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
        )
        RETURNING
            id,
            "name",
            "leaveTypeId",
            "annualLimit",
            "maxConsecutiveDays",
            "carryForward",
            "maxCarryForwardDays",
            "requiresApproval",
            "effectiveFrom",
            "status",
            "createdAt",
            "updatedAt"
        `,
		policy.ID,
		policy.Name,
		policy.LeaveTypeId,
		policy.AnnualLimit,
		policy.MaxConsecutiveDays,
		policy.CarryForward,
		policy.MaxCarryForwardDays,
		policy.RequiresApproval,
		policy.EffectiveFrom,
		policy.Status,
		policy.CreatedAt,
		policy.UpdatedAt,
	).Scan(
		&policy.ID,
		&policy.Name,
		&policy.LeaveTypeId,
		&policy.AnnualLimit,
		&policy.MaxConsecutiveDays,
		&policy.CarryForward,
		&policy.MaxCarryForwardDays,
		&policy.RequiresApproval,
		&policy.EffectiveFrom,
		&policy.Status,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create leave policy: %w", err)
	}

	return &policy, nil
}

func (r *PolicyRepository) UpdatePolicy(
	ctx context.Context,
	policy model.LeavePolicy,
) (*model.LeavePolicy, error) {

	err := r.DB.QueryRowContext(
		ctx,
		`
        UPDATE public."LeavePolicy"
        SET
            "name" = $1,
            "leaveTypeId" = $2,
            "annualLimit" = $3,
            "maxConsecutiveDays" = $4,
            "carryForward" = $5,
            "maxCarryForwardDays" = $6,
            "requiresApproval" = $7,
            "effectiveFrom" = $8,
            "status" = $9,
            "updatedAt" = $10
        WHERE id = $11
        RETURNING
            id,
            "name",
            "leaveTypeId",
            "annualLimit",
            "maxConsecutiveDays",
            "carryForward",
            "maxCarryForwardDays",
            "requiresApproval",
            "effectiveFrom",
            "status",
            "createdAt",
            "updatedAt"
        `,
		policy.Name,
		policy.LeaveTypeId,
		policy.AnnualLimit,
		policy.MaxConsecutiveDays,
		policy.CarryForward,
		policy.MaxCarryForwardDays,
		policy.RequiresApproval,
		policy.EffectiveFrom,
		policy.Status,
		policy.UpdatedAt,
		policy.ID,
	).Scan(
		&policy.ID,
		&policy.Name,
		&policy.LeaveTypeId,
		&policy.AnnualLimit,
		&policy.MaxConsecutiveDays,
		&policy.CarryForward,
		&policy.MaxCarryForwardDays,
		&policy.RequiresApproval,
		&policy.EffectiveFrom,
		&policy.Status,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("leave policy not found")
		}

		return nil, fmt.Errorf("failed to update leave policy: %w", err)
	}

	return &policy, nil
}

func (r *PolicyRepository) DeletePolicy(
	ctx context.Context,
	id string,
) error {

	policyID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid policy ID")
	}

	result, err := r.DB.ExecContext(
		ctx,
		`
        DELETE FROM public."LeavePolicy"
        WHERE id = $1
        `,
		policyID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete leave policy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify policy deletion: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("leave policy not found")
	}

	return nil
}
