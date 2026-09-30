package attendance

import (
	"context"
	"fmt"
	"io"
	"time"
	"strconv"
    "strings"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// AttendanceExcelRow represents one row from the Excel file.
type AttendanceExcelRow struct {
    EmployeeID       string
    Date             time.Time
    CheckInTime      *time.Time
    CheckOutTime     *time.Time
    Status           string
    WorkHours        *time.Time
    LateByMinutes    int
    EarlyByMinutes   int
    OvertimeMinutes  int
    Shift            string
}

// ImportAttendanceExcel reads attendance data from an Excel file.
func (s *Service) ImportAttendanceExcel(
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
	// EmployeeId | Date | CheckInTime | CheckOutTime | Status
	header := rows[0]

	if len(header) < 5 {
		return 0, 0, nil, fmt.Errorf(
			"invalid Excel format: expected columns EmployeeId, Date, CheckInTime, CheckOutTime, Status",
		)
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
		dateValue := row[1]
		checkInValue := row[2]
		checkOutValue := row[3]
		status := row[4]

		if employeeID == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: employeeId is required", excelRowNumber),
			)
			continue
		}

		// Validate employee UUID.
		if _, err := uuid.Parse(employeeID); err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: invalid employeeId", excelRowNumber),
			)
			continue
		}

		attendanceDate, err := parseExcelDate(dateValue)
		if err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: invalid date", excelRowNumber),
			)
			continue
		}

		var checkInTime *time.Time
		if checkInValue != "" {
			value, err := parseExcelDateTime(checkInValue)
			if err != nil {
				errorsList = append(
					errorsList,
					fmt.Sprintf("row %d: invalid checkInTime", excelRowNumber),
				)
				continue
			}
			checkInTime = &value
		}

		var checkOutTime *time.Time
		if checkOutValue != "" {
			value, err := parseExcelDateTime(checkOutValue)
			if err != nil {
				errorsList = append(
					errorsList,
					fmt.Sprintf("row %d: invalid checkOutTime", excelRowNumber),
				)
				continue
			}
			checkOutTime = &value
		}

		if status == "" {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: status is required", excelRowNumber),
			)
			continue
		}

		durationValue := ""
        lateByValue := ""
        earlyByValue := ""
        overtimeValue := ""
        shift := ""

        if len(row) > 5 {
			durationValue = strings.TrimSpace(row[5])
        }

        if len(row) > 6 {
			lateByValue = strings.TrimSpace(row[6])
        }

        if len(row) > 7 {
			earlyByValue = strings.TrimSpace(row[7])
        }

        if len(row) > 8 {
			overtimeValue = strings.TrimSpace(row[8])
        }

        if len(row) > 9 {
			shift = strings.TrimSpace(row[9])
        }

		workHours, err := parseWorkHours(durationValue)
        if err != nil {
			errorsList = append(
				errorsList,
                fmt.Sprintf("row %d: invalid duration: %v", excelRowNumber, err),
            )
            continue
        }

        lateByMinutes, err := parseDurationMinutes(lateByValue)
        if err != nil {
			errorsList = append(
				errorsList,
                fmt.Sprintf("row %d: invalid lateBy: %v", excelRowNumber, err),
            )
            continue
        }

        earlyByMinutes, err := parseDurationMinutes(earlyByValue)
        if err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: invalid earlyBy: %v", excelRowNumber, err),
            )
            continue
        }

        overtimeMinutes, err := parseDurationMinutes(overtimeValue)
        if err != nil {
			errorsList = append(
				errorsList,
                fmt.Sprintf("row %d: invalid overtime: %v", excelRowNumber, err),
            )
            continue
        }

		attendance := AttendanceExcelRow{
			EmployeeID:       employeeID,
            Date:             attendanceDate,
            CheckInTime:      checkInTime,
            CheckOutTime:     checkOutTime,
            Status:           status,
            WorkHours:        workHours,
            LateByMinutes:    lateByMinutes,
            EarlyByMinutes:   earlyByMinutes,
            OvertimeMinutes:  overtimeMinutes,
            Shift:            shift,
        }

		// Save the row using the repository.
		if err := s.repo.ImportAttendance(ctx, attendance); err != nil {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: %v", excelRowNumber, err),
			)
			continue
		}

		successRows++
	}

	return totalRows, successRows, errorsList, nil
}

func parseWorkHours(value string) (*time.Time, error) {
    value = strings.TrimSpace(value)

    if value == "" || value == "-" {
        return nil, nil
    }

    t, err := time.Parse("15:04", value)
    if err != nil {
        return nil, fmt.Errorf("invalid duration %q: %w", value, err)
    }

    return &t, nil
}

func parseDurationMinutes(value string) (int, error) {
    value = strings.TrimSpace(value)

    if value == "" || value == "-" {
        return 0, nil
    }

    // Handle HH:MM
    parts := strings.Split(value, ":")

    if len(parts) != 2 {
        return 0, fmt.Errorf("invalid duration format: %s", value)
    }

    hours, err := strconv.Atoi(strings.TrimSpace(parts[0]))
    if err != nil {
        return 0, fmt.Errorf("invalid duration hours: %s", value)
    }

    minutes, err := strconv.Atoi(strings.TrimSpace(parts[1]))
    if err != nil {
        return 0, fmt.Errorf("invalid duration minutes: %s", value)
    }

    if minutes < 0 || minutes >= 60 {
        return 0, fmt.Errorf("invalid duration: %s", value)
    }

    return hours*60 + minutes, nil
}

func parseExcelDate(value string) (time.Time, error) {
	// Try common Excel date formats first.
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

func parseExcelDateTime(value string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"02/01/2006 15:04:05",
		"02/01/2006 15:04",
		"15:04:05",
		"15:04",
	}

	for _, format := range formats {
		if parsed, err := time.Parse(format, value); err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsupported datetime format: %s", value)
}
