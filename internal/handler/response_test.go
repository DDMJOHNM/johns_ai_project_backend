package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondContextError_deadlineWrapped(t *testing.T) {
	w := httptest.NewRecorder()
	err := fmt.Errorf("downstream: %w", context.DeadlineExceeded)
	if !RespondContextError(w, err) {
		t.Fatal("expected true")
	}
	if w.Code != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusGatewayTimeout)
	}
	var body ErrorResponse
	if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "Gateway Timeout" {
		t.Fatalf("error field = %q", body.Error)
	}
}

func TestRespondContextError_canceledWrapped(t *testing.T) {
	w := httptest.NewRecorder()
	err := fmt.Errorf("wrap: %w", context.Canceled)
	if !RespondContextError(w, err) {
		t.Fatal("expected true")
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusServiceUnavailable)
	}
	var body ErrorResponse
	if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Error != "Request Canceled" {
		t.Fatalf("error field = %q", body.Error)
	}
}

func TestRespondContextError_nilReturnsFalse(t *testing.T) {
	if RespondContextError(httptest.NewRecorder(), nil) {
		t.Fatal("nil err should return false")
	}
}

func TestRespondContextError_otherErrorReturnsFalse(t *testing.T) {
	w := httptest.NewRecorder()
	if RespondContextError(w, fmt.Errorf("not a context error")) {
		t.Fatal("generic error should return false")
	}
}
