package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/cart"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

type CartHandler struct {
	calculator *cart.Calculator
}

func NewCartHandler(calculator *cart.Calculator) *CartHandler {
	return &CartHandler{calculator: calculator}
}

// CalculateQuote handles POST /v1/checkout/quote
func (h *CartHandler) CalculateQuote(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	var req cart.QuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "MALFORMED_REQUEST", "invalid JSON request payload", nil, reqID)
		return
	}

	if req.RestaurantID == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "restaurantId is required", nil, reqID)
		return
	}

	quote, err := h.calculator.CalculateQuote(r.Context(), req)
	if err != nil {
		if errors.Is(err, catalog.ErrRestaurantNotFound) || errors.Is(err, catalog.ErrMenuItemNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil, reqID)
			return
		}
		if errors.Is(err, cart.ErrEmptyCart) || errors.Is(err, cart.ErrInvalidQuantity) || errors.Is(err, cart.ErrOptionNotFound) {
			response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil, reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError, "CALCULATION_FAILED", err.Error(), nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, quote, reqID)
}
