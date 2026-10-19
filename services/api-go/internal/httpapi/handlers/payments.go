package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/payments"
)

type PaymentsHandler struct {
	paymentsService *payments.Service
}

func NewPaymentsHandler(paymentsService *payments.Service) *PaymentsHandler {
	return &PaymentsHandler{
		paymentsService: paymentsService,
	}
}

type InitiatePaymentRequest struct {
	Provider string `json:"provider,omitempty"`
}

// InitiatePayment handles POST /v1/orders/{id}/pay
func (h *PaymentsHandler) InitiatePayment(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	orderID := r.PathValue("id")
	if orderID == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_PARAM", "Missing order id in path", nil, reqID)
		return
	}

	var req InitiatePaymentRequest
	if r.Body != nil && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	session, err := h.paymentsService.InitiatePayment(r.Context(), claims.Subject, orderID, req.Provider)
	if err != nil {
		switch {
		case errors.Is(err, orders.ErrOrderNotFound):
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Order not found", nil, reqID)
		case errors.Is(err, orders.ErrUnauthorizedOrderAccess):
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to pay for this order", nil, reqID)
		case errors.Is(err, payments.ErrPaymentAlreadyCompleted):
			response.Error(w, http.StatusConflict, "ALREADY_PAID", "Order is already paid", nil, reqID)
		case errors.Is(err, payments.ErrInvalidOrderState):
			response.Error(w, http.StatusUnprocessableEntity, "INVALID_STATE", err.Error(), nil, reqID)
		case errors.Is(err, payments.ErrProviderNotFound):
			response.Error(w, http.StatusBadRequest, "INVALID_PROVIDER", err.Error(), nil, reqID)
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil, reqID)
		}
		return
	}

	response.Success(w, http.StatusOK, session, reqID)
}

// HandleWebhook processes provider webhooks via POST /v1/webhooks/payments/{provider}
func (h *PaymentsHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	providerName := r.PathValue("provider")
	if providerName == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_PARAM", "Missing provider in webhook path", nil, reqID)
		return
	}

	// Limit webhook payload to 1MB
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024*1024))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "PAYLOAD_TOO_LARGE", "Failed to read webhook body", nil, reqID)
		return
	}

	event, isDuplicate, err := h.paymentsService.ProcessWebhook(r.Context(), providerName, r, body)
	if err != nil {
		switch {
		case errors.Is(err, payments.ErrMissingSignatureHeader) || errors.Is(err, payments.ErrInvalidSignature):
			response.Error(w, http.StatusUnauthorized, "INVALID_SIGNATURE", err.Error(), nil, reqID)
		case errors.Is(err, payments.ErrProviderNotFound):
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Payment provider not found", nil, reqID)
		case errors.Is(err, payments.ErrAmountMismatch):
			response.Error(w, http.StatusUnprocessableEntity, "AMOUNT_MISMATCH", "Webhook amount does not match order total", nil, reqID)
		case errors.Is(err, orders.ErrOrderNotFound):
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Order not found for webhook event", nil, reqID)
		default:
			response.Error(w, http.StatusBadRequest, "BAD_REQUEST", err.Error(), nil, reqID)
		}
		return
	}

	response.Success(w, http.StatusOK, map[string]any{
		"received":    true,
		"eventId":     event.EventID,
		"status":      event.Status,
		"isDuplicate": isDuplicate,
	}, reqID)
}
