package attendance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// CreateRegularization creates a Pending regularization request.
func (s *Service) CreateRegularization(
	ctx context.Context,
	employeeID string,
	attendanceID int32,
	requestedCheckIn *time.Time,
	requestedCheckOut *time.Time,
	reason string,
) error {

	if employeeID == "" {
		return errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return errors.New("invalid employee ID")
	}

	if attendanceID <= 0 {
		return errors.New("attendance ID must be greater than 0")
	}

	if requestedCheckIn == nil && requestedCheckOut == nil {
		return errors.New("requested check-in or requested check-out is required")
	}

	if requestedCheckIn != nil && requestedCheckOut != nil && !requestedCheckOut.After(*requestedCheckIn) {
		return errors.New("requested check-out must be after requested check-in")
	}

	if strings.TrimSpace(reason) == "" {
		return errors.New("reason cannot be empty")
	}

	return s.repo.CreateRegularization(
		ctx,
		employeeID,
		attendanceID,
		requestedCheckIn,
		requestedCheckOut,
		strings.TrimSpace(reason),
	)
}

// GetRegularizationHistory returns all regularizations for an employee.
func (s *Service) GetRegularizationHistory(
	ctx context.Context,
	employeeID string,
) ([]interface{}, error) {

	if employeeID == "" {
		return nil, errors.New("employee ID cannot be empty")
	}

	if _, err := uuid.Parse(employeeID); err != nil {
		return nil, errors.New("invalid employee ID")
	}

	records, err := s.repo.GetRegularizationHistory(
		ctx,
		employeeID,
	)

	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(records))

	for i := range records {
		result[i] = records[i]
	}

	return result, nil
}

// ApproveRegularization approves a request and updates Attendance atomically.
func (s *Service) ApproveRegularization(
	ctx context.Context,
	regularizationID string,
	approvedBy string,
) error {

	if regularizationID == "" {
		return errors.New("regularization ID cannot be empty")
	}

	if _, err := uuid.Parse(regularizationID); err != nil {
		return errors.New("invalid regularization ID")
	}

	if approvedBy == "" {
		return errors.New("approvedBy cannot be empty")
	}

	if _, err := uuid.Parse(approvedBy); err != nil {
		return errors.New("invalid approvedBy UUID")
	}

	return s.repo.ApproveRegularization(
		ctx,
		regularizationID,
		approvedBy,
	)
}

// RejectRegularization rejects a pending request.
func (s *Service) RejectRegularization(
	ctx context.Context,
	regularizationID string,
	rejectedBy string,
	rejectionReason string,
) error {

	if regularizationID == "" {
		return errors.New("regularization ID cannot be empty")
	}

	if _, err := uuid.Parse(regularizationID); err != nil {
		return errors.New("invalid regularization ID")
	}

	if rejectedBy == "" {
		return errors.New("rejectedBy cannot be empty")
	}

	if _, err := uuid.Parse(rejectedBy); err != nil {
		return errors.New("invalid rejectedBy UUID")
	}

	if strings.TrimSpace(rejectionReason) == "" {
		return errors.New("rejection reason cannot be empty")
	}

	return s.repo.RejectRegularization(
		ctx,
		regularizationID,
		rejectedBy,
		strings.TrimSpace(rejectionReason),
	)
}

// ImportRegularizationExcel imports regularization requests.
// Every imported request starts as Pending.
func (s *Service) ImportRegularizationExcel(
	ctx context.Context,
	reader io.Reader,
) (int, int, []string, error) {

	data, err := io.ReadAll(reader)

	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to read Excel file: %w", err)
	}

	file, err := excelize.OpenReader(bytes.NewReader(data))

	if err != nil {
		return 0, 0, nil, fmt.Errorf(
			"failed to open Excel file: %w",
			err,
		)
	}

	defer file.Close()

	sheets := file.GetSheetList()

	if len(sheets) == 0 {
		return 0, 0, nil, errors.New("Excel file has no sheets")
	}

	rows, err := file.GetRows(sheets[0])

	if err != nil {
		return 0, 0, nil, fmt.Errorf("failed to read Excel rows: %w", err)
	}

	if len(rows) <= 1 {
		return 0, 0, nil, errors.New("Excel file contains no data rows")
	}

	headerMap := make(map[string]int)

	for i, header := range rows[0] {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	requiredHeaders := []string{
		"employeeid",
		"attendanceid",
		"reason",
	}

	for _, header := range requiredHeaders {
		if _, exists := headerMap[header]; !exists {
			return 0, 0, nil, fmt.Errorf("missing required Excel column: %s", header)
		}
	}

	totalRows := len(rows) - 1
	successRows := 0
	var errorsList []string

	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {

		row := rows[rowIndex]
		excelRowNumber := rowIndex + 1

		getValue := func(column string) string {

			index, exists := headerMap[column]

			if !exists || index >= len(row) {
				return ""
			}

			return strings.TrimSpace(row[index])
		}

		employeeID := getValue("employeeid")
		attendanceIDString := getValue("attendanceid")
		checkInString := getValue("requestedcheckin")
		checkOutString := getValue("requestedcheckout")
		reason := getValue("reason")

		if employeeID == "" {
			errorsList = append(errorsList, fmt.Sprintf("row %d: employeeId is required", excelRowNumber))
			continue
		}

		if _, err := uuid.Parse(employeeID); err != nil {
			errorsList = append(errorsList, fmt.Sprintf("row %d: invalid employeeId", excelRowNumber))
			continue
		}

		var attendanceID int32

		if _, err := fmt.Sscanf(
			attendanceIDString,
			"%d",
			&attendanceID,
		); err != nil || attendanceID <= 0 {
			errorsList = append(errorsList, fmt.Sprintf("row %d: invalid attendanceId", excelRowNumber))
			continue
		}

		var requestedCheckIn *time.Time
		var requestedCheckOut *time.Time

		if checkInString != "" {

			t, err := parseRegularizationTime(checkInString)

			if err != nil {
				errorsList = append(errorsList, fmt.Sprintf("row %d: invalid requestedCheckIn", excelRowNumber))
				continue
			}

			requestedCheckIn = &t
		}

		if checkOutString != "" {

			t, err := parseRegularizationTime(checkOutString)

			if err != nil {
				errorsList = append(errorsList, fmt.Sprintf("row %d: invalid requestedCheckOut", excelRowNumber))
				continue
			}

			requestedCheckOut = &t
		}

		if requestedCheckIn == nil && requestedCheckOut == nil {
			errorsList = append(errorsList, fmt.Sprintf("row %d: requestedCheckIn or requestedCheckOut is required", excelRowNumber))
			continue
		}

		if requestedCheckIn != nil && requestedCheckOut != nil && !requestedCheckOut.After(*requestedCheckIn) {
			errorsList = append(
				errorsList,
				fmt.Sprintf("row %d: requestedCheckOut must be after requestedCheckIn", excelRowNumber))
			continue
		}

		if strings.TrimSpace(reason) == "" {
			errorsList = append(errorsList, fmt.Sprintf("row %d: reason is required", excelRowNumber))
			continue
		}

		err = s.CreateRegularization(
			ctx,
			employeeID,
			attendanceID,
			requestedCheckIn,
			requestedCheckOut,
			reason,
		)

		if err != nil {
			errorsList = append(errorsList, fmt.Sprintf("row %d: %s", excelRowNumber, err.Error()))
			continue
		}

		successRows++
	}

	return totalRows, successRows, errorsList, nil
}
