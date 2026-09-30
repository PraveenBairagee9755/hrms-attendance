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
	EmployeeID      string
	EmployeeName    string
	DateOfJoining   *time.Time
	CtcPerMonth     decimal.Decimal
	DaysPaid        decimal.Decimal
	Salary          decimal.Decimal
	Late            decimal.Decimal
	Incentive       decimal.Decimal
	Conv            decimal.Decimal
	Advance         decimal.Decimal
	DeductionAmount decimal.Decimal
	AccountNo       string
	Ifsc            string
	PfDeduction     decimal.Decimal
	PfUanNumber     string
	GrossSalary     decimal.Decimal
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

	if len(header) < 16 {
		return 0, 0, nil, fmt.Errorf(
			"invalid Excel format: expected columns EmployeeId, EmployeeName, DateOfJoining, CtcPerMonth, DaysPaid, Salary, Late, Incentive, Conv, Advance, DeductionAmount, AccountNo, Ifsc, PfDeduction, PfUanNumber, GrossSalary",
		)
	}

	totalRows := 0
	successRows := 0
	var errorsList []string

	for rowIndex, row := range rows[1:] {

		excelRowNumber := rowIndex + 2
		totalRows++

		if len(row) < 16 {
			errorsList = append(errorsList,
				fmt.Sprintf("row %d: missing required columns", excelRowNumber),
			)
			continue
		}

		employeeID := strings.TrimSpace(row[0])
		employeeName := strings.TrimSpace(row[1])
		dateOfJoiningValue := strings.TrimSpace(row[2])
		accountNo := strings.TrimSpace(row[11])
		ifsc := strings.TrimSpace(row[12])
		pfUanNumber := strings.TrimSpace(row[14])

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

		// Employee Name
		if employeeName == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: employeeName is required",
					excelRowNumber,
				),
			)
			continue
		}

		// Date Of Joining
		var dateOfJoining *time.Time

		if dateOfJoiningValue != "" {

			parsedDateOfJoining, err := parseSalaryExcelDate(dateOfJoiningValue)
			if err != nil {
				errorsList = append(
					errorsList,
					fmt.Sprintf(
						"row %d: invalid dateOfJoining",
						excelRowNumber,
					),
				)
				continue
			}

			dateOfJoining = &parsedDateOfJoining
		}

		salaryStructure := SalaryStructureExcelRow{
			EmployeeID:    employeeID,
			EmployeeName:  employeeName,
			DateOfJoining: dateOfJoining,
			AccountNo:     accountNo,
			Ifsc:          ifsc,
			PfUanNumber:   pfUanNumber,
		}

		// Amount columns
		amountColumns := []struct {
			name     string
			index    int
			required bool
			target   *decimal.Decimal
		}{
			{"ctcPerMonth", 3, true, &salaryStructure.CtcPerMonth},
			{"daysPaid", 4, false, &salaryStructure.DaysPaid},
			{"salary", 5, false, &salaryStructure.Salary},
			{"late", 6, false, &salaryStructure.Late},
			{"incentive", 7, false, &salaryStructure.Incentive},
			{"conv", 8, false, &salaryStructure.Conv},
			{"advance", 9, false, &salaryStructure.Advance},
			{"deductionAmount", 10, false, &salaryStructure.DeductionAmount},
			{"pfDeduction", 13, false, &salaryStructure.PfDeduction},
			{"grossSalary", 15, true, &salaryStructure.GrossSalary},
		}

		var amountErr string

		for _, column := range amountColumns {

			value := strings.TrimSpace(row[column.index])

			if value == "" {
				if column.required {
					amountErr = fmt.Sprintf("%s is required", column.name)
					break
				}
				*column.target = decimal.Zero
				continue
			}

			amount, err := decimal.NewFromString(value)
			if err != nil || amount.IsNegative() {
				amountErr = fmt.Sprintf("invalid %s", column.name)
				break
			}

			*column.target = amount
		}

		if amountErr != "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf(
					"row %d: %s",
					excelRowNumber,
					amountErr,
				),
			)
			continue
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
