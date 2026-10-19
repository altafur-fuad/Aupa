package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/catalog"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCatalogRepository struct {
	pool *pgxpool.Pool
}

func NewCatalogRepository(pool *pgxpool.Pool) *PostgresCatalogRepository {
	return &PostgresCatalogRepository{pool: pool}
}

func (r *PostgresCatalogRepository) List(ctx context.Context, filter catalog.RestaurantFilter) ([]catalog.Restaurant, error) {
	query := `
		SELECT id, name, cuisine, rating, review_count, 30 as delivery_time, 5000 as delivery_fee,
		       'Banani' as area, 'Dhaka' as city, zone_id, is_open
		FROM public.restaurants
		WHERE ($1 = '' OR zone_id = $1)
		  AND ($2 = '' OR lower(cuisine) = lower($2))
		  AND ($3 = '' OR lower(name) LIKE '%' || lower($3) || '%' OR lower(cuisine) LIKE '%' || lower($3) || '%')
		ORDER BY rating DESC, review_count DESC
	`
	rows, err := r.pool.Query(ctx, query, filter.ZoneID, filter.Category, filter.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to query restaurants: %w", err)
	}
	defer rows.Close()

	var list []catalog.Restaurant
	for rows.Next() {
		var rest catalog.Restaurant
		if err := rows.Scan(
			&rest.ID,
			&rest.Name,
			&rest.Cuisine,
			&rest.Rating,
			&rest.ReviewCount,
			&rest.DeliveryTimeMinutes,
			&rest.DeliveryFeePoisha,
			&rest.Area,
			&rest.City,
			&rest.ZoneID,
			&rest.IsOpen,
		); err != nil {
			return nil, fmt.Errorf("failed to scan restaurant row: %w", err)
		}
		list = append(list, rest)
	}
	return list, nil
}

func (r *PostgresCatalogRepository) GetByID(ctx context.Context, id string) (*catalog.Restaurant, error) {
	query := `
		SELECT id, name, cuisine, rating, review_count, 30 as delivery_time, 5000 as delivery_fee,
		       'Banani' as area, 'Dhaka' as city, zone_id, is_open
		FROM public.restaurants
		WHERE id = $1
	`
	var rest catalog.Restaurant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rest.ID,
		&rest.Name,
		&rest.Cuisine,
		&rest.Rating,
		&rest.ReviewCount,
		&rest.DeliveryTimeMinutes,
		&rest.DeliveryFeePoisha,
		&rest.Area,
		&rest.City,
		&rest.ZoneID,
		&rest.IsOpen,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, catalog.ErrRestaurantNotFound
		}
		return nil, fmt.Errorf("failed to query restaurant: %w", err)
	}
	return &rest, nil
}

func (r *PostgresCatalogRepository) GetByRestaurantID(ctx context.Context, restaurantID string) ([]catalog.MenuItem, error) {
	query := `
		SELECT m.id, m.restaurant_id, m.name, COALESCE(m.description, ''), m.price_poisha,
		       COALESCE(c.name, 'Main Course'), m.is_available
		FROM public.menu_items m
		LEFT JOIN public.menu_categories c ON m.category_id = c.id
		WHERE m.restaurant_id = $1
		ORDER BY m.price_poisha ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu items: %w", err)
	}
	defer rows.Close()

	var items []catalog.MenuItem
	for rows.Next() {
		var it catalog.MenuItem
		if err := rows.Scan(
			&it.ID,
			&it.RestaurantID,
			&it.Name,
			&it.Description,
			&it.PricePoisha,
			&it.Category,
			&it.IsAvailable,
		); err != nil {
			return nil, fmt.Errorf("failed to scan menu item: %w", err)
		}
		items = append(items, it)
	}

	// Fetch option items
	for i := range items {
		optQuery := `
			SELECT id, name, price_delta_poisha
			FROM public.menu_item_options
			WHERE item_id = $1
		`
		optRows, err := r.pool.Query(ctx, optQuery, items[i].ID)
		if err == nil {
			var optionItems []catalog.OptionItem
			for optRows.Next() {
				var opt catalog.OptionItem
				if err := optRows.Scan(&opt.ID, &opt.Name, &opt.PriceModifierPoisha); err == nil {
					optionItems = append(optionItems, opt)
				}
			}
			optRows.Close()

			if len(optionItems) > 0 {
				items[i].OptionGroups = []catalog.OptionGroup{
					{
						ID:           "grp_" + items[i].ID,
						Name:         "Customization",
						IsRequired:   false,
						MinSelection: 0,
						MaxSelection: len(optionItems),
						Items:        optionItems,
					},
				}
			}
		}
	}

	return items, nil
}

func (r *PostgresCatalogRepository) GetItemByID(ctx context.Context, itemID string) (*catalog.MenuItem, error) {
	query := `
		SELECT m.id, m.restaurant_id, m.name, COALESCE(m.description, ''), m.price_poisha,
		       COALESCE(c.name, 'Main Course'), m.is_available
		FROM public.menu_items m
		LEFT JOIN public.menu_categories c ON m.category_id = c.id
		WHERE m.id = $1
	`
	var it catalog.MenuItem
	err := r.pool.QueryRow(ctx, query, itemID).Scan(
		&it.ID,
		&it.RestaurantID,
		&it.Name,
		&it.Description,
		&it.PricePoisha,
		&it.Category,
		&it.IsAvailable,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, catalog.ErrMenuItemNotFound
		}
		return nil, fmt.Errorf("failed to query menu item: %w", err)
	}

	// Options
	optQuery := `SELECT id, name, price_delta_poisha FROM public.menu_item_options WHERE item_id = $1`
	optRows, err := r.pool.Query(ctx, optQuery, it.ID)
	if err == nil {
		var optionItems []catalog.OptionItem
		for optRows.Next() {
			var opt catalog.OptionItem
			if err := optRows.Scan(&opt.ID, &opt.Name, &opt.PriceModifierPoisha); err == nil {
				optionItems = append(optionItems, opt)
			}
		}
		optRows.Close()
		if len(optionItems) > 0 {
			it.OptionGroups = []catalog.OptionGroup{
				{
					ID:           "grp_" + it.ID,
					Name:         "Customization",
					IsRequired:   false,
					MinSelection: 0,
					MaxSelection: len(optionItems),
					Items:        optionItems,
				},
			}
		}
	}

	return &it, nil
}
