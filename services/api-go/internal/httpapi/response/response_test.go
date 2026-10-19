package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccessResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	Success(rec, http.StatusOK, map[string]string{"status": "ok"}, "req_123")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var env SuccessEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if env.Meta == nil || env.Meta.RequestID != "req_123" {
		t.Errorf("expected requestId req_123, got %v", env.Meta)
	}
}

func TestErrorResponse(t *testing.T) {
	rec := httptest.NewRecorder()
	Error(rec, http.StatusBadRequest, "VALIDATION_ERROR", "invalid phone", []FieldError{
		{Field: "phone", Code: "INVALID_FORMAT"},
	}, "req_456")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var env ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if env.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got %s", env.Error.Code)
	}
	if len(env.Error.Fields) != 1 || env.Error.Fields[0].Field != "phone" {
		t.Errorf("unexpected field errors: %v", env.Error.Fields)
	}
	if env.Meta == nil || env.Meta.RequestID != "req_456" {
		t.Errorf("expected requestId req_456, got %v", env.Meta)
	}
}
