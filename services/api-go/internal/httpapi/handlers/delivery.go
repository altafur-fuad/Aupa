package handlers

import (
	"errors"
	"net/http"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/delivery"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

type DeliveryHandler struct {
	deliveryService *delivery.Service
}

func NewDeliveryHandler(deliveryService *delivery.Service) *DeliveryHandler {
	return &DeliveryHandler{
		deliveryService: deliveryService,
	}
}

// GetTracking handles GET /v1/orders/{id}/tracking
func (h *DeliveryHandler) GetTracking(w http.ResponseWriter, r *http.Request) {
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

	info, err := h.deliveryService.GetTrackingInfo(r.Context(), orderID, claims.Subject, false)
	if err != nil {
		switch {
		case errors.Is(err, delivery.ErrTrackingNotFound):
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Tracking information not found for order", nil, reqID)
		case errors.Is(err, delivery.ErrUnauthorizedTracking):
			response.Error(w, http.StatusForbidden, "FORBIDDEN", "You do not have permission to track this order", nil, reqID)
		default:
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil, reqID)
		}
		return
	}

	response.Success(w, http.StatusOK, info, reqID)
}
