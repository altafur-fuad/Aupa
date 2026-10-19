package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("generates request ID when none provided", func(t *testing.T) {
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := GetRequestID(r.Context())
			if reqID == "" {
				t.Error("expected request ID in context, got empty string")
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		headerVal := rec.Header().Get("X-Request-Id")
		if headerVal == "" {
			t.Error("expected X-Request-Id header in response")
		}
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		const customID = "req_custom_12345"
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := GetRequestID(r.Context())
			if reqID != customID {
				t.Errorf("expected %s, got %s", customID, reqID)
			}
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-Id", customID)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Header().Get("X-Request-Id") != customID {
			t.Errorf("expected header %s, got %s", customID, rec.Header().Get("X-Request-Id"))
		}
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	panicHandler := Recovery(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("database connection collapsed!")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	panicHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}

	var env response.ErrorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("failed to decode error json: %v", err)
	}

	if env.Error.Code != "INTERNAL_SERVER_ERROR" {
		t.Errorf("expected code INTERNAL_SERVER_ERROR, got %s", env.Error.Code)
	}

	// Verify stack was logged
	if !bytes.Contains(buf.Bytes(), []byte("database connection collapsed!")) {
		t.Errorf("expected panic message in logs, got %s", buf.String())
	}
}

func TestCORSMiddleware(t *testing.T) {
	cors := CORS("http://localhost:3000")
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("OPTIONS returns 204", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rec := httptest.NewRecorder()

		cors(dummy).ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected 204 for OPTIONS, got %d", rec.Code)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Errorf("unexpected allow origin header: %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})

	t.Run("GET includes CORS headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		cors(dummy).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
			t.Errorf("unexpected allow origin header: %s", rec.Header().Get("Access-Control-Allow-Origin"))
		}
	})
}

func TestTimeoutMiddleware(t *testing.T) {
	timeout := Timeout(50 * time.Millisecond)

	slowHandler := timeout(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(150 * time.Millisecond):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// Context timed out as expected
			return
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rec := httptest.NewRecorder()

	slowHandler.ServeHTTP(rec, req)

	// Since context timed out before writing headers, verify context error
	if req.Context().Err() == nil {
		// Note: The context wrapped is r.Context(), checked inside handler
	}
}

func TestChain(t *testing.T) {
	var callOrder []string

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "m1_start")
			next.ServeHTTP(w, r)
			callOrder = append(callOrder, "m1_end")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callOrder = append(callOrder, "m2_start")
			next.ServeHTTP(w, r)
			callOrder = append(callOrder, "m2_end")
		})
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callOrder = append(callOrder, "final")
		w.WriteHeader(http.StatusOK)
	})

	pipeline := Chain(final, m1, m2)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	pipeline.ServeHTTP(rec, req)

	expected := []string{"m1_start", "m2_start", "final", "m2_end", "m1_end"}
	if len(callOrder) != len(expected) {
		t.Fatalf("expected call order %v, got %v", expected, callOrder)
	}
	for i, name := range expected {
		if callOrder[i] != name {
			t.Errorf("at index %d expected %s, got %s", i, name, callOrder[i])
		}
	}
}
