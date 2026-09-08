package leave

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// LeaveApplicationExcelRow represents one LeaveApplication row from Excel.
type LeaveApplicationExcelRow struct {
	EmployeeID  string
	LeaveTypeID string
	StartDate   time.Time
	EndDate     time.Time
	Reason      string
}

// ImportLeaveApplicationExcel reads leave applications from an Excel file.
func (s *Service) ImportLeaveApplicationExcel(
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
		return 0, 0, nil, fmt.Errorf("Excel file must contain a header and at least one data row")
	}

	// Expected columns:
	// EmployeeId | LeaveTypeId | StartDate | EndDate | Reason
	header := rows[0]

	if len(header) < 5 {
		return 0, 0, nil, fmt.Errorf("invalid Excel format: expected columns EmployeeId, LeaveTypeId, StartDate, EndDate, Reason")
	}

	totalRows := 0
	successRows := 0
	var errorsList []string

	for rowIndex, row := range rows[1:] {

		excelRowNumber := rowIndex + 2
		totalRows++

		if len(row) < 5 {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: missing required columns", excelRowNumber),
			)
			continue
		}

		employeeID := row[0]
		leaveTypeID := row[1]
		startDateValue := row[2]
		endDateValue := row[3]
		reason := row[4]

		// Validate employee ID.
		if employeeID == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: employeeId is required", excelRowNumber),
			)
			continue
		}

		if _, err := uuid.Parse(employeeID); err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: invalid employeeId", excelRowNumber),
			)
			continue
		}

		// Validate leave type ID.
		if leaveTypeID == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: leaveTypeId is required", excelRowNumber),
			)
			continue
		}

		if _, err := uuid.Parse(leaveTypeID); err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: invalid leaveTypeId", excelRowNumber),
			)
			continue
		}

		// Parse start date.
		startDate, err := parseLeaveExcelDate(startDateValue)
		if err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: invalid startDate", excelRowNumber),
			)
			continue
		}

		// Parse end date.
		endDate, err := parseLeaveExcelDate(endDateValue)
		if err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: invalid endDate", excelRowNumber),
			)
			continue
		}

		// End date cannot be before start date.
		if endDate.Before(startDate) {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: endDate cannot be before startDate", excelRowNumber),
			)
			continue
		}

		leaveApplication := LeaveApplicationExcelRow{
			EmployeeID:  employeeID,
			LeaveTypeID: leaveTypeID,
			StartDate:   startDate,
			EndDate:     endDate,
			Reason:      reason,
		}

		err = s.repo.ImportLeaveApplication(
			ctx,
			leaveApplication,
		)

		if err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: %v", excelRowNumber, err),
			)
			continue
		}

		successRows++
	}

	failedRows := totalRows - successRows

	return successRows, failedRows, errorsList, nil
}

// parseLeaveExcelDate parses common Excel date formats.
func parseLeaveExcelDate(value string) (time.Time, error) {

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

	return time.Time{}, fmt.Errorf("unsupported date format: %s", value)
}
