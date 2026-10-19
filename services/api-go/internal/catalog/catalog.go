package catalog

import (
	"context"
	"errors"
	"strings"
	"sync"
)

var (
	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrMenuItemNotFound   = errors.New("menu item not found")
)

// OptionItem represents a customization choice with integer minor unit price modifier.
type OptionItem struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	PriceModifierPoisha int64  `json:"priceModifierPoisha"`
}

// OptionGroup represents a set of customization options.
type OptionGroup struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	IsRequired   bool         `json:"isRequired"`
	MinSelection int          `json:"minSelection"`
	MaxSelection int          `json:"maxSelection"`
	Items        []OptionItem `json:"items"`
}

// MenuItem represents an available food item in a restaurant.
type MenuItem struct {
	ID           string        `json:"id"`
	RestaurantID string        `json:"restaurantId"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	PricePoisha  int64         `json:"pricePoisha"`
	Category     string        `json:"category"`
	IsAvailable  bool          `json:"isAvailable"`
	OptionGroups []OptionGroup `json:"optionGroups,omitempty"`
}

// Restaurant represents an active dining merchant in a service zone.
type Restaurant struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	Cuisine             string  `json:"cuisine"`
	Rating              float64 `json:"rating"`
	ReviewCount         int     `json:"reviewCount"`
	DeliveryTimeMinutes int     `json:"deliveryTimeMinutes"`
	DeliveryFeePoisha   int64   `json:"deliveryFeePoisha"`
	Area                string  `json:"area"`
	City                string  `json:"city"`
	ZoneID              string  `json:"zoneId"`
	IsOpen              bool    `json:"isOpen"`
}

// RestaurantFilter specifies search criteria.
type RestaurantFilter struct {
	Query    string
	Category string
	ZoneID   string
}

// RestaurantRepository defines restaurant catalog access.
type RestaurantRepository interface {
	List(ctx context.Context, filter RestaurantFilter) ([]Restaurant, error)
	GetByID(ctx context.Context, id string) (*Restaurant, error)
}

// MenuRepository defines menu catalog access.
type MenuRepository interface {
	GetByRestaurantID(ctx context.Context, restaurantID string) ([]MenuItem, error)
	GetItemByID(ctx context.Context, itemID string) (*MenuItem, error)
}

// Service provides catalog queries.
type Service struct {
	restaurantRepo RestaurantRepository
	menuRepo       MenuRepository
}

func NewService(rRepo RestaurantRepository, mRepo MenuRepository) *Service {
	return &Service{
		restaurantRepo: rRepo,
		menuRepo:       mRepo,
	}
}

func (s *Service) ListRestaurants(ctx context.Context, filter RestaurantFilter) ([]Restaurant, error) {
	return s.restaurantRepo.List(ctx, filter)
}

func (s *Service) GetRestaurant(ctx context.Context, id string) (*Restaurant, error) {
	return s.restaurantRepo.GetByID(ctx, id)
}

func (s *Service) GetMenu(ctx context.Context, restaurantID string) ([]MenuItem, error) {
	// Verify restaurant exists
	if _, err := s.restaurantRepo.GetByID(ctx, restaurantID); err != nil {
		return nil, err
	}
	return s.menuRepo.GetByRestaurantID(ctx, restaurantID)
}

// MemoryCatalogRepository provides in-memory seeded store for development and testing.
type MemoryCatalogRepository struct {
	mu          sync.RWMutex
	restaurants map[string]Restaurant
	menuItems   map[string]MenuItem
}

func NewMemoryCatalogRepository() *MemoryCatalogRepository {
	repo := &MemoryCatalogRepository{
		restaurants: make(map[string]Restaurant),
		menuItems:   make(map[string]MenuItem),
	}
	repo.seedInitialData()
	return repo
}

func (m *MemoryCatalogRepository) seedInitialData() {
	m.restaurants["rest_01"] = Restaurant{
		ID:                  "rest_01",
		Name:                "Burger Barn",
		Cuisine:             "American • Burgers",
		Rating:              4.8,
		ReviewCount:         340,
		DeliveryTimeMinutes: 25,
		DeliveryFeePoisha:   6000, // ৳60
		Area:                "Banani",
		City:                "Dhaka",
		ZoneID:              "zone_dhaka_central",
		IsOpen:              true,
	}

	m.restaurants["rest_02"] = Restaurant{
		ID:                  "rest_02",
		Name:                "Sultan's Dine",
		Cuisine:             "Traditional • Biryani",
		Rating:              4.9,
		ReviewCount:         1200,
		DeliveryTimeMinutes: 35,
		DeliveryFeePoisha:   8000, // ৳80
		Area:                "Gulshan 2",
		City:                "Dhaka",
		ZoneID:              "zone_dhaka_central",
		IsOpen:              true,
	}

	m.restaurants["rest_03"] = Restaurant{
		ID:                  "rest_03",
		Name:                "Chillox Banani",
		Cuisine:             "Fast Food • Burgers",
		Rating:              4.7,
		ReviewCount:         890,
		DeliveryTimeMinutes: 30,
		DeliveryFeePoisha:   5000, // ৳50
		Area:                "Banani",
		City:                "Dhaka",
		ZoneID:              "zone_dhaka_central",
		IsOpen:              true,
	}

	m.restaurants["rest_04"] = Restaurant{
		ID:                  "rest_04",
		Name:                "Pizza Guy",
		Cuisine:             "Italian • Pizza",
		Rating:              4.6,
		ReviewCount:         450,
		DeliveryTimeMinutes: 40,
		DeliveryFeePoisha:   7000, // ৳70
		Area:                "Gulshan 1",
		City:                "Dhaka",
		ZoneID:              "zone_dhaka_central",
		IsOpen:              true,
	}

	// Menu items for Burger Barn
	m.menuItems["item_01"] = MenuItem{
		ID:           "item_01",
		RestaurantID: "rest_01",
		Name:         "Classic Cheese Burger",
		Description:  "150g beef patty with cheddar cheese, lettuce, tomato, and secret sauce.",
		PricePoisha:  35000, // ৳350
		Category:     "Burgers",
		IsAvailable:  true,
		OptionGroups: []OptionGroup{
			{
				ID:           "og_patty",
				Name:         "Patty Size",
				IsRequired:   true,
				MinSelection: 1,
				MaxSelection: 1,
				Items: []OptionItem{
					{ID: "opt_single", Name: "Single Patty", PriceModifierPoisha: 0},
					{ID: "opt_double", Name: "Double Patty (+৳150)", PriceModifierPoisha: 15000},
				},
			},
			{
				ID:           "og_cheese",
				Name:         "Add-ons",
				IsRequired:   false,
				MinSelection: 0,
				MaxSelection: 2,
				Items: []OptionItem{
					{ID: "opt_extra_cheese", Name: "Extra Cheddar (+৳50)", PriceModifierPoisha: 5000},
					{ID: "opt_bacon", Name: "Beef Bacon (+৳80)", PriceModifierPoisha: 8000},
				},
			},
		},
	}

	m.menuItems["item_02"] = MenuItem{
		ID:           "item_02",
		RestaurantID: "rest_01",
		Name:         "Smoky BBQ Bacon Burger",
		Description:  "Charbroiled beef patty, smoked bacon, crispy onion rings and BBQ sauce.",
		PricePoisha:  42000, // ৳420
		Category:     "Burgers",
		IsAvailable:  true,
	}

	// Menu items for Sultan's Dine
	m.menuItems["item_03"] = MenuItem{
		ID:           "item_03",
		RestaurantID: "rest_02",
		Name:         "Kacchi Biryani (Basmati)",
		Description:  "Mouth-watering mutton kacchi served with aromatic basmati rice, potato and salad.",
		PricePoisha:  48000, // ৳480
		Category:     "Biryani",
		IsAvailable:  true,
		OptionGroups: []OptionGroup{
			{
				ID:           "og_portion",
				Name:         "Portion",
				IsRequired:   true,
				MinSelection: 1,
				MaxSelection: 1,
				Items: []OptionItem{
					{ID: "opt_half", Name: "Regular", PriceModifierPoisha: 0},
					{ID: "opt_full", Name: "Large (+৳200)", PriceModifierPoisha: 20000},
				},
			},
		},
	}
}

func (m *MemoryCatalogRepository) List(ctx context.Context, filter RestaurantFilter) ([]Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Restaurant
	q := strings.ToLower(strings.TrimSpace(filter.Query))
	cat := strings.ToLower(strings.TrimSpace(filter.Category))

	for _, r := range m.restaurants {
		if filter.ZoneID != "" && r.ZoneID != filter.ZoneID {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(r.Name), q) && !strings.Contains(strings.ToLower(r.Cuisine), q) {
			continue
		}
		if cat != "" && cat != "all" && !strings.Contains(strings.ToLower(r.Cuisine), cat) {
			continue
		}
		result = append(result, r)
	}

	return result, nil
}

func (m *MemoryCatalogRepository) GetByID(ctx context.Context, id string) (*Restaurant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, exists := m.restaurants[id]
	if !exists {
		return nil, ErrRestaurantNotFound
	}
	return &r, nil
}

func (m *MemoryCatalogRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]MenuItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []MenuItem
	for _, item := range m.menuItems {
		if item.RestaurantID == restaurantID {
			result = append(result, item)
		}
	}
	return result, nil
}

func (m *MemoryCatalogRepository) GetItemByID(ctx context.Context, itemID string) (*MenuItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.menuItems[itemID]
	if !exists {
		return nil, ErrMenuItemNotFound
	}
	return &item, nil
}
