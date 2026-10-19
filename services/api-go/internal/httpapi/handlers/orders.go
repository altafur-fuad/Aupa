package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/cart"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
)

type OrdersHandler struct {
	orderService *orders.Service
}

func NewOrdersHandler(orderService *orders.Service) *OrdersHandler {
	return &OrdersHandler{orderService: orderService}
}

type createOrderRequest struct {
	RestaurantID        string               `json:"restaurantId"`
	DeliveryAddressID   string               `json:"deliveryAddressId"`
	PaymentMethod       orders.PaymentMethod `json:"paymentMethod"`
	PromoCode           string               `json:"promoCode,omitempty"`
	SpecialInstructions string               `json:"specialInstructions,omitempty"`
	Items               []cart.ItemInput     `json:"items"`
}

// CreateOrder handles POST /v1/orders
func (h *OrdersHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		response.Error(w, http.StatusBadRequest, "MISSING_HEADER", "Idempotency-Key header is required for order creation.", nil, reqID)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "MALFORMED_REQUEST", "Failed to read request body.", nil, reqID)
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "MALFORMED_REQUEST", "Invalid JSON payload.", nil, reqID)
		return
	}

	input := orders.CreateOrderInput{
		CustomerID:          claims.Subject,
		RestaurantID:        req.RestaurantID,
		DeliveryAddressID:   req.DeliveryAddressID,
		PaymentMethod:       req.PaymentMethod,
		PromoCode:           req.PromoCode,
		SpecialInstructions: req.SpecialInstructions,
		Items:               req.Items,
		IdempotencyKey:      idempotencyKey,
		RawPayloadHash:      orders.HashPayload(bodyBytes),
	}

	order, isReplay, err := h.orderService.CreateOrder(r.Context(), input)
	if err != nil {
		if errors.Is(err, orders.ErrIdempotencyConflict) {
			response.Error(w, http.StatusConflict, "IDEMPOTENCY_CONFLICT", err.Error(), nil, reqID)
			return
		}
		response.Error(w, http.StatusBadRequest, "ORDER_CREATION_FAILED", err.Error(), nil, reqID)
		return
	}

	statusCode := http.StatusCreated
	if isReplay {
		statusCode = http.StatusOK
	}

	response.Success(w, statusCode, order, reqID)
}

// ListOrders handles GET /v1/orders
func (h *OrdersHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	list, err := h.orderService.ListOrders(r.Context(), claims.Subject)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve order history.", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, list, reqID)
}

// GetOrder handles GET /v1/orders/{id}
func (h *OrdersHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	orderID := r.PathValue("id")
	isPrivileged := claims.Role == "admin" || claims.Role == "merchant"

	order, err := h.orderService.GetOrder(r.Context(), orderID, claims.Subject, isPrivileged)
	if err != nil {
		if errors.Is(err, orders.ErrOrderNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Order not found.", nil, reqID)
			return
		}
		if errors.Is(err, orders.ErrUnauthorizedOrderAccess) {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error(), nil, reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch order details.", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, order, reqID)
}

// CancelOrder handles POST /v1/orders/{id}/cancel
func (h *OrdersHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil, reqID)
		return
	}

	orderID := r.PathValue("id")
	isPrivileged := claims.Role == "admin"

	order, err := h.orderService.CancelOrder(r.Context(), orderID, claims.Subject, isPrivileged)
	if err != nil {
		if errors.Is(err, orders.ErrCannotCancelOrder) {
			response.Error(w, http.StatusBadRequest, "CANNOT_CANCEL", err.Error(), nil, reqID)
			return
		}
		if errors.Is(err, orders.ErrUnauthorizedOrderAccess) {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error(), nil, reqID)
			return
		}
		if errors.Is(err, orders.ErrOrderNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "Order not found.", nil, reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel order.", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, order, reqID)
}
