package cart

import (
	"context"
	"errors"
	"strings"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
)

var (
	ErrEmptyCart          = errors.New("cart cannot be empty")
	ErrInvalidQuantity    = errors.New("item quantity must be at least 1")
	ErrItemUnavailable    = errors.New("one or more items are currently unavailable")
	ErrOptionNotFound     = errors.New("invalid customization option selected")
	ErrRestaurantMismatch = errors.New("items must be from the same restaurant")
)

// ItemInput represents an individual item submitted in cart quote request.
type ItemInput struct {
	ItemID        string   `json:"itemId"`
	Quantity      int      `json:"quantity"`
	OptionItemIDs []string `json:"optionItemIds,omitempty"`
}

// QuoteRequest represents the client quote request payload.
type QuoteRequest struct {
	RestaurantID      string      `json:"restaurantId"`
	DeliveryAddressID string      `json:"deliveryAddressId"`
	PromoCode         string      `json:"promoCode,omitempty"`
	Items             []ItemInput `json:"items"`
}

// QuoteResponse represents the authoritative server-side price breakdown in poisha (BDT).
type QuoteResponse struct {
	RestaurantID      string `json:"restaurantId"`
	SubtotalPoisha    int64  `json:"subtotalPoisha"`
	DeliveryFeePoisha int64  `json:"deliveryFeePoisha"`
	DiscountPoisha    int64  `json:"discountPoisha"`
	VATPoisha         int64  `json:"vatPoisha"`
	TotalPoisha       int64  `json:"totalPoisha"`
	Currency          string `json:"currency"`
	ItemCount         int    `json:"itemCount"`
}

// Calculator performs authoritative server-side price calculation.
type Calculator struct {
	catalogService *catalog.Service
}

func NewCalculator(catalogService *catalog.Service) *Calculator {
	return &Calculator{catalogService: catalogService}
}

// CalculateQuote verifies items, calculates integer totals in poisha, applies promos, and calculates VAT.
func (c *Calculator) CalculateQuote(ctx context.Context, req QuoteRequest) (*QuoteResponse, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyCart
	}

	restaurant, err := c.catalogService.GetRestaurant(ctx, req.RestaurantID)
	if err != nil {
		return nil, err
	}

	menuItems, err := c.catalogService.GetMenu(ctx, req.RestaurantID)
	if err != nil {
		return nil, err
	}

	menuMap := make(map[string]catalog.MenuItem)
	for _, m := range menuItems {
		menuMap[m.ID] = m
	}

	var subtotalPoisha int64
	var totalItemCount int

	for _, line := range req.Items {
		if line.Quantity <= 0 {
			return nil, ErrInvalidQuantity
		}

		item, exists := menuMap[line.ItemID]
		if !exists {
			return nil, catalog.ErrMenuItemNotFound
		}
		if !item.IsAvailable {
			return nil, ErrItemUnavailable
		}

		unitPrice := item.PricePoisha

		// Validate and add option modifiers
		if len(line.OptionItemIDs) > 0 {
			optionMap := make(map[string]catalog.OptionItem)
			for _, group := range item.OptionGroups {
				for _, opt := range group.Items {
					optionMap[opt.ID] = opt
				}
			}

			for _, optID := range line.OptionItemIDs {
				opt, optExists := optionMap[optID]
				if !optExists {
					return nil, ErrOptionNotFound
				}
				unitPrice += opt.PriceModifierPoisha
			}
		}

		subtotalPoisha += unitPrice * int64(line.Quantity)
		totalItemCount += line.Quantity
	}

	deliveryFeePoisha := restaurant.DeliveryFeePoisha

	// Promo discount calculation (e.g. AUPA50 -> 50% discount up to ৳200 / 20000 poisha on orders >= ৳500 / 50000 poisha)
	var discountPoisha int64
	cleanPromo := strings.ToUpper(strings.TrimSpace(req.PromoCode))
	if cleanPromo == "AUPA50" {
		if subtotalPoisha >= 50000 { // ৳500 minimum
			discountPoisha = subtotalPoisha / 2
			if discountPoisha > 20000 { // ৳200 cap
				discountPoisha = 20000
			}
		}
	}

	// Standard restaurant VAT in BD is 5% on discounted food subtotal
	taxableAmount := subtotalPoisha - discountPoisha
	if taxableAmount < 0 {
		taxableAmount = 0
	}
	vatPoisha := (taxableAmount * 5) / 100

	totalPoisha := taxableAmount + deliveryFeePoisha + vatPoisha

	return &QuoteResponse{
		RestaurantID:      req.RestaurantID,
		SubtotalPoisha:    subtotalPoisha,
		DeliveryFeePoisha: deliveryFeePoisha,
		DiscountPoisha:    discountPoisha,
		VATPoisha:         vatPoisha,
		TotalPoisha:       totalPoisha,
		Currency:          "BDT",
		ItemCount:         totalItemCount,
	}, nil
}
