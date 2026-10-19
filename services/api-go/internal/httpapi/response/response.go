package response

import (
	"encoding/json"
	"net/http"
)

// Meta contains metadata for the response envelope.
type Meta struct {
	RequestID string `json:"requestId,omitempty"`
}

// SuccessEnvelope wraps successful response payloads.
type SuccessEnvelope struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// FieldError represents a specific field-level validation issue.
type FieldError struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}

// ErrorDetail contains machine-readable and user-friendly error info.
type ErrorDetail struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

// ErrorEnvelope wraps error response payloads.
type ErrorEnvelope struct {
	Error ErrorDetail `json:"error"`
	Meta  *Meta       `json:"meta,omitempty"`
}

// JSON writes an arbitrary payload as application/json.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// Success writes a standard success envelope.
func Success(w http.ResponseWriter, status int, data any, requestID string) {
	var meta *Meta
	if requestID != "" {
		meta = &Meta{RequestID: requestID}
	}
	JSON(w, status, SuccessEnvelope{
		Data: data,
		Meta: meta,
	})
}

// Error writes a standard error envelope.
func Error(w http.ResponseWriter, status int, code, message string, fields []FieldError, requestID string) {
	var meta *Meta
	if requestID != "" {
		meta = &Meta{RequestID: requestID}
	}
	JSON(w, status, ErrorEnvelope{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
		Meta: meta,
	})
}
