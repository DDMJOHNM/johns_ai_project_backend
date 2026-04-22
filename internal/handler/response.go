package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

// RespondContextError writes a JSON error when err is a context deadline or cancellation (including
// values wrapped with fmt.Errorf("...: %w", err)). Returns true if the response was written.
func RespondContextError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		RespondJSON(w, http.StatusGatewayTimeout, ErrorResponse{
			Error:   "Gateway Timeout",
			Message: "The request took too long to complete",
		})
		return true
	}
	if errors.Is(err, context.Canceled) {
		RespondJSON(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   "Request Canceled",
			Message: "The request was canceled or the client disconnected",
		})
		return true
	}
	return false
}

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
