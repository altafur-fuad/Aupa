package handlers

import (
	"net/http"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

// HealthHandler provides liveness and readiness endpoints.
type HealthHandler struct {
	isReadyFunc func() bool
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(isReadyFunc func() bool) *HealthHandler {
	return &HealthHandler{isReadyFunc: isReadyFunc}
}

// Healthz responds with 200 OK if service is alive.
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

// Readyz responds with 200 OK if dependencies are ready, or 503 if not.
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	if h.isReadyFunc != nil && !h.isReadyFunc() {
		response.Error(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "dependency check failed", nil, "")
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"status":   "ready",
		"database": "connected",
	})
}
