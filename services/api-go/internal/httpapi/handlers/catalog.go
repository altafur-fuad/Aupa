package handlers

import (
	"errors"
	"net/http"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/middleware"
	"github.com/altafur-fuad/Aupa/services/api-go/internal/httpapi/response"
)

// CatalogHandler handles restaurant and menu queries.
type CatalogHandler struct {
	catalogService *catalog.Service
}

func NewCatalogHandler(catalogService *catalog.Service) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

// ListRestaurants handles GET /v1/restaurants?q=&category=&zoneId=
func (h *CatalogHandler) ListRestaurants(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())

	filter := catalog.RestaurantFilter{
		Query:    r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
		ZoneID:   r.URL.Query().Get("zoneId"),
	}

	restaurants, err := h.catalogService.ListRestaurants(r.Context(), filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list restaurants", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, restaurants, reqID)
}

// GetRestaurant handles GET /v1/restaurants/{id}
func (h *CatalogHandler) GetRestaurant(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	id := r.PathValue("id")

	if id == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "restaurant id is required", nil, reqID)
		return
	}

	restaurant, err := h.catalogService.GetRestaurant(r.Context(), id)
	if err != nil {
		if errors.Is(err, catalog.ErrRestaurantNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "restaurant not found", nil, reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch restaurant", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, restaurant, reqID)
}

// GetMenu handles GET /v1/restaurants/{id}/menu
func (h *CatalogHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	reqID := middleware.GetRequestID(r.Context())
	id := r.PathValue("id")

	if id == "" {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "restaurant id is required", nil, reqID)
		return
	}

	menu, err := h.catalogService.GetMenu(r.Context(), id)
	if err != nil {
		if errors.Is(err, catalog.ErrRestaurantNotFound) {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", "restaurant not found", nil, reqID)
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch menu", nil, reqID)
		return
	}

	response.Success(w, http.StatusOK, menu, reqID)
}
