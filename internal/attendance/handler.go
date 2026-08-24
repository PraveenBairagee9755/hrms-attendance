package attendance

import (
	"encoding/json"
	"net/http"
)

// Handler manages HTTP transport routes
type Handler struct {
	service *Service
}

// NewHandler creates a new instance of the Handler
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// AttendanceRequest maps incoming JSON inputs
type AttendanceRequest struct {
	EmployeeID string `json:"employeeId"`
}

// ClockInHandler handles HTTP requests to /api/attendance/clock-in
func (h *Handler) ClockInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err := h.service.ClockIn(r.Context(), req.EmployeeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Successfully clocked in"}`))
}

// ClockOutHandler handles HTTP requests to /api/attendance/clock-out
func (h *Handler) ClockOutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AttendanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err := h.service.ClockOut(r.Context(), req.EmployeeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Successfully clocked out"}`))
}
