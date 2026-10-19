package catalog

import (
	"context"
	"testing"
)

func TestCatalogService(t *testing.T) {
	repo := NewMemoryCatalogRepository()
	svc := NewService(repo, repo)
	ctx := context.Background()

	t.Run("list all restaurants", func(t *testing.T) {
		list, err := svc.ListRestaurants(ctx, RestaurantFilter{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) < 4 {
			t.Errorf("expected at least 4 restaurants, got %d", len(list))
		}
	})

	t.Run("filter by query string", func(t *testing.T) {
		list, err := svc.ListRestaurants(ctx, RestaurantFilter{Query: "Pizza"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 || list[0].Name != "Pizza Guy" {
			t.Errorf("expected 1 result 'Pizza Guy', got %v", list)
		}
	})

	t.Run("filter by category", func(t *testing.T) {
		list, err := svc.ListRestaurants(ctx, RestaurantFilter{Category: "Biryani"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 || list[0].Name != "Sultan's Dine" {
			t.Errorf("expected 'Sultan's Dine', got %v", list)
		}
	})

	t.Run("get existing restaurant by ID", func(t *testing.T) {
		r, err := svc.GetRestaurant(ctx, "rest_01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.Name != "Burger Barn" {
			t.Errorf("expected Burger Barn, got %s", r.Name)
		}
		if r.DeliveryFeePoisha != 6000 {
			t.Errorf("expected delivery fee 6000 poisha, got %d", r.DeliveryFeePoisha)
		}
	})

	t.Run("get non-existent restaurant returns ErrRestaurantNotFound", func(t *testing.T) {
		_, err := svc.GetRestaurant(ctx, "non_existent_id")
		if err != ErrRestaurantNotFound {
			t.Errorf("expected ErrRestaurantNotFound, got %v", err)
		}
	})

	t.Run("get restaurant menu", func(t *testing.T) {
		menu, err := svc.GetMenu(ctx, "rest_01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(menu) == 0 {
			t.Fatalf("expected menu items for rest_01, got 0")
		}

		var found *MenuItem
		for _, it := range menu {
			if it.ID == "item_01" {
				found = &it
				break
			}
		}
		if found == nil || len(found.OptionGroups) == 0 {
			t.Errorf("expected option groups on item_01")
		}
	})
}
