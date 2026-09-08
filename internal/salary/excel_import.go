package salary

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

type SalaryStructureExcelRow struct {
	EmployeeID    string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	BasicSalary   decimal.Decimal
	HRA           decimal.Decimal
	Allowances    string
	Deductions    string
	GrossSalary   decimal.Decimal
	NetSalary     decimal.Decimal
	Currency      string
	CreatedBy     *uuid.UUID
}

// ImportSalaryStructureExcel imports salary structures from Excel.
func (s *Service) ImportSalaryStructureExcel(
	ctx context.Context,
	reader io.Reader,
) (int, int, []string, error) {

	file, err := excelize.OpenReader(reader)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer file.Close()

	sheets := file.GetSheetList()

	if len(sheets) == 0 {
		return 0, 0, nil, fmt.Errorf("Excel file contains no sheets")
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to read Excel rows: %w", err)
	}

	if len(rows) < 2 {
		return 0, 0, nil, fmt.Errorf(
			"Excel file must contain a header and at least one data row",
		)
	}

	header := rows[0]

	if len(header) < 11 {
		return 0, 0, nil, fmt.Errorf(
			"invalid Excel format: expected columns EmployeeId, EffectiveFrom, EffectiveTo, BasicSalary, Hra, Allowances, Deductions, GrossSalary, NetSalary, Currency, CreatedBy",
		)
	}

	totalRows := 0
	successRows := 0
	var errorsList []string

	for rowIndex, row := range rows[1:] {

		excelRowNumber := rowIndex + 2
		totalRows++

		if len(row) < 10 {
			errorsList = append(errorsList,
				fmt.Sprintf("row %d: missing required columns", excelRowNumber),
			)
			continue
		}

		employeeID := strings.TrimSpace(row[0])
		effectiveFromValue := strings.TrimSpace(row[1])
		effectiveToValue := strings.TrimSpace(row[2])
		basicSalaryValue := strings.TrimSpace(row[3])
		hraValue := strings.TrimSpace(row[4])
		allowances := strings.TrimSpace(row[5])
		deductions := strings.TrimSpace(row[6])
		grossSalaryValue := strings.TrimSpace(row[7])
		netSalaryValue := strings.TrimSpace(row[8])
		currency := strings.TrimSpace(row[9])

		createdByValue := ""

		if len(row) > 10 {
			createdByValue = strings.TrimSpace(row[10])
		}

		// Employee ID
		if employeeID == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: employeeId is required",
					excelRowNumber,
				),
			)
			continue
		}

		if _, err := uuid.Parse(employeeID); err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: invalid employeeId",
					excelRowNumber,
				),
			)
			continue
		}

		// Effective From
		if effectiveFromValue == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: effectiveFrom is required",
					excelRowNumber,
				),
			)
			continue
		}

		effectiveFrom, err := parseSalaryExcelDate(effectiveFromValue)
		if err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: invalid effectiveFrom",
					excelRowNumber,
				),
			)
			continue
		}

		// Effective To
		var effectiveTo *time.Time

		if effectiveToValue != "" {

			parsedEffectiveTo, err := parseSalaryExcelDate(effectiveToValue)
			if err != nil {
				errorsList = append(
					errorsList,
					fmt.Sprintf(
						"row %d: invalid effectiveTo",
						excelRowNumber,
					),
				)
				continue
			}

			effectiveTo = &parsedEffectiveTo

			if effectiveTo.Before(effectiveFrom) {
				errorsList = append(
					errorsList,
					fmt.Sprintf(
						"row %d: effectiveTo cannot be before effectiveFrom",
						excelRowNumber,
					),
				)
				continue
			}
		}

		// Basic Salary
		basicSalary, err := decimal.NewFromString(basicSalaryValue)
		if err != nil || basicSalary.IsNegative() {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: invalid basicSalary",
					excelRowNumber,
				),
			)
			continue
		}

		// HRA
		hra, err := decimal.NewFromString(hraValue)
		if err != nil || hra.IsNegative() {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: invalid hra",
					excelRowNumber,
				),
			)
			continue
		}

		// Gross Salary
		grossSalary, err := decimal.NewFromString(grossSalaryValue)
		if err != nil || grossSalary.IsNegative() {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: invalid grossSalary",
					excelRowNumber,
				),
			)
			continue
		}

		// Net Salary
		netSalary, err := decimal.NewFromString(netSalaryValue)
		if err != nil || netSalary.IsNegative() {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: invalid netSalary",
					excelRowNumber,
				),
			)
			continue
		}

		// Currency
		if currency == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: currency is required",
					excelRowNumber,
				),
			)
			continue
		}

		// Created By
		var createdBy *uuid.UUID

		if createdByValue != "" {
			createdByUUID, err := uuid.Parse(createdByValue)
			if err != nil {
				errorsList = append(
					errorsList,
					fmt.Sprintf(
						"row %d: invalid createdBy",
						excelRowNumber,
					),
				)
				continue
			}

			createdBy = &createdByUUID
		}

		salaryStructure := SalaryStructureExcelRow{
			EmployeeID:    employeeID,
			EffectiveFrom: effectiveFrom,
			EffectiveTo:   effectiveTo,
			BasicSalary:   basicSalary,
			HRA:           hra,
			Allowances:    allowances,
			Deductions:    deductions,
			GrossSalary:   grossSalary,
			NetSalary:     netSalary,
			Currency:      currency,
			CreatedBy:     createdBy,
		}

		err = s.repo.ImportSalaryStructure(
			ctx,
			salaryStructure,
		)

		if err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: %v",
					excelRowNumber,
					err,
				),
			)
			continue
		}

		successRows++
	}

	failedRows := totalRows - successRows

	return successRows, failedRows, errorsList, nil
}

func parseSalaryExcelDate(value string) (time.Time, error) {

	formats := []string{
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
		"02-01-2006",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf(
		"unsupported date format: %s",
		value,
	)
}
