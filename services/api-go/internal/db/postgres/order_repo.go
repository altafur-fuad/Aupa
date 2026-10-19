package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/orders"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *orders.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start order transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO public.orders (
			id, customer_id, restaurant_id, delivery_address_id, status, payment_status,
			payment_method, subtotal_poisha, delivery_fee_poisha, discount_poisha, vat_poisha,
			total_amount_poisha, currency, special_instructions, created_at, updated_at
		)
		VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err = tx.Exec(ctx, orderQuery,
		o.ID,
		o.CustomerID,
		o.RestaurantID,
		o.DeliveryAddressID,
		string(o.Status),
		string(o.PaymentStatus),
		string(o.PaymentMethod),
		o.SubtotalPoisha,
		o.DeliveryFeePoisha,
		o.DiscountPoisha,
		o.VATPoisha,
		o.TotalAmountPoisha,
		o.Currency,
		o.SpecialInstructions,
		o.CreatedAt,
		o.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order record: %w", err)
	}

	// Insert items
	itemQuery := `
		INSERT INTO public.order_items (id, order_id, item_id, name, unit_price_poisha, quantity, total_poisha, selected_options_snapshot)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
	`
	for _, it := range o.Items {
		optsJSON, _ := json.Marshal(it.SelectedOptionNames)
		_, err := tx.Exec(ctx, itemQuery,
			o.ID,
			it.ItemID,
			it.Name,
			it.UnitPricePoisha,
			it.Quantity,
			it.TotalPoisha,
			optsJSON,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	// Record initial status in status history
	historyQuery := `
		INSERT INTO public.order_status_history (id, order_id, to_status, actor_id, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
	`
	_, _ = tx.Exec(ctx, historyQuery, o.ID, string(o.Status), o.CustomerID, o.CreatedAt)

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) GetByID(ctx context.Context, id string) (*orders.Order, error) {
	query := `
		SELECT id, customer_id, restaurant_id, COALESCE(delivery_address_id::text, ''),
		       status, payment_status, payment_method, subtotal_poisha, delivery_fee_poisha,
		       discount_poisha, vat_poisha, total_amount_poisha, currency,
		       COALESCE(special_instructions, ''), created_at, updated_at
		FROM public.orders
		WHERE id = $1
	`
	var o orders.Order
	var statusStr, payStatusStr, payMethodStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID,
		&o.CustomerID,
		&o.RestaurantID,
		&o.DeliveryAddressID,
		&statusStr,
		&payStatusStr,
		&payMethodStr,
		&o.SubtotalPoisha,
		&o.DeliveryFeePoisha,
		&o.DiscountPoisha,
		&o.VATPoisha,
		&o.TotalAmountPoisha,
		&o.Currency,
		&o.SpecialInstructions,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, orders.ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to query order: %w", err)
	}

	o.Status = orders.OrderStatus(statusStr)
	o.PaymentStatus = orders.PaymentStatus(payStatusStr)
	o.PaymentMethod = orders.PaymentMethod(payMethodStr)

	// Fetch items
	items, err := r.getOrderItems(ctx, o.ID)
	if err == nil {
		o.Items = items
	}

	return &o, nil
}

func (r *PostgresOrderRepository) ListByCustomer(ctx context.Context, customerID string) ([]orders.Order, error) {
	query := `
		SELECT id, customer_id, restaurant_id, COALESCE(delivery_address_id::text, ''),
		       status, payment_status, payment_method, subtotal_poisha, delivery_fee_poisha,
		       discount_poisha, vat_poisha, total_amount_poisha, currency,
		       COALESCE(special_instructions, ''), created_at, updated_at
		FROM public.orders
		WHERE customer_id = $1::uuid
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query customer orders: %w", err)
	}
	defer rows.Close()

	var list []orders.Order
	for rows.Next() {
		var o orders.Order
		var statusStr, payStatusStr, payMethodStr string
		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.RestaurantID,
			&o.DeliveryAddressID,
			&statusStr,
			&payStatusStr,
			&payMethodStr,
			&o.SubtotalPoisha,
			&o.DeliveryFeePoisha,
			&o.DiscountPoisha,
			&o.VATPoisha,
			&o.TotalAmountPoisha,
			&o.Currency,
			&o.SpecialInstructions,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		o.Status = orders.OrderStatus(statusStr)
		o.PaymentStatus = orders.PaymentStatus(payStatusStr)
		o.PaymentMethod = orders.PaymentMethod(payMethodStr)
		list = append(list, o)
	}

	// Fetch items for each order
	for i := range list {
		items, err := r.getOrderItems(ctx, list[i].ID)
		if err == nil {
			list[i].Items = items
		}
	}

	return list, nil
}

func (r *PostgresOrderRepository) ListByRestaurant(ctx context.Context, restaurantID string) ([]orders.Order, error) {
	query := `
		SELECT id, customer_id, restaurant_id, COALESCE(delivery_address_id::text, ''),
		       status, payment_status, payment_method, subtotal_poisha, delivery_fee_poisha,
		       discount_poisha, vat_poisha, total_amount_poisha, currency,
		       COALESCE(special_instructions, ''), created_at, updated_at
		FROM public.orders
		WHERE restaurant_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query restaurant orders: %w", err)
	}
	defer rows.Close()

	var list []orders.Order
	for rows.Next() {
		var o orders.Order
		var statusStr, payStatusStr, payMethodStr string
		if err := rows.Scan(
			&o.ID,
			&o.CustomerID,
			&o.RestaurantID,
			&o.DeliveryAddressID,
			&statusStr,
			&payStatusStr,
			&payMethodStr,
			&o.SubtotalPoisha,
			&o.DeliveryFeePoisha,
			&o.DiscountPoisha,
			&o.VATPoisha,
			&o.TotalAmountPoisha,
			&o.Currency,
			&o.SpecialInstructions,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		o.Status = orders.OrderStatus(statusStr)
		o.PaymentStatus = orders.PaymentStatus(payStatusStr)
		o.PaymentMethod = orders.PaymentMethod(payMethodStr)
		list = append(list, o)
	}

	for i := range list {
		items, err := r.getOrderItems(ctx, list[i].ID)
		if err == nil {
			list[i].Items = items
		}
	}

	return list, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status orders.OrderStatus, paymentStatus orders.PaymentStatus) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch previous status
	var prevStatus string
	_ = tx.QueryRow(ctx, "SELECT status FROM public.orders WHERE id = $1", id).Scan(&prevStatus)

	query := `
		UPDATE public.orders
		SET status = $2, payment_status = $3, updated_at = $4
		WHERE id = $1
	`
	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, query, id, string(status), string(paymentStatus), now)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return orders.ErrOrderNotFound
	}

	// Insert history
	historyQuery := `
		INSERT INTO public.order_status_history (id, order_id, from_status, to_status, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
	`
	_, _ = tx.Exec(ctx, historyQuery, id, prevStatus, string(status), now)

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) getOrderItems(ctx context.Context, orderID string) ([]orders.OrderItem, error) {
	query := `
		SELECT id, item_id, name, unit_price_poisha, quantity, total_poisha, selected_options_snapshot
		FROM public.order_items
		WHERE order_id = $1
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []orders.OrderItem
	for rows.Next() {
		var it orders.OrderItem
		var optsJSON []byte
		if err := rows.Scan(
			&it.ID,
			&it.ItemID,
			&it.Name,
			&it.UnitPricePoisha,
			&it.Quantity,
			&it.TotalPoisha,
			&optsJSON,
		); err == nil {
			if len(optsJSON) > 0 {
				_ = json.Unmarshal(optsJSON, &it.SelectedOptionNames)
			}
			items = append(items, it)
		}
	}
	return items, nil
}

type PostgresIdempotencyRepository struct {
	pool *pgxpool.Pool
}

func NewIdempotencyRepository(pool *pgxpool.Pool) *PostgresIdempotencyRepository {
	return &PostgresIdempotencyRepository{pool: pool}
}

func (r *PostgresIdempotencyRepository) Get(ctx context.Context, key string, customerID string) (*orders.IdempotencyRecord, error) {
	query := `
		SELECT key, customer_id, request_hash, COALESCE(order_id, ''), created_at
		FROM public.idempotency_records
		WHERE key = $1 AND customer_id = $2::uuid
	`
	var rec orders.IdempotencyRecord
	err := r.pool.QueryRow(ctx, query, key, customerID).Scan(
		&rec.Key,
		&rec.CustomerID,
		&rec.RequestHash,
		&rec.OrderID,
		&rec.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query idempotency record: %w", err)
	}
	return &rec, nil
}

func (r *PostgresIdempotencyRepository) Save(ctx context.Context, record *orders.IdempotencyRecord) error {
	query := `
		INSERT INTO public.idempotency_records (id, key, customer_id, request_hash, order_id, created_at)
		VALUES (gen_random_uuid(), $1, $2::uuid, $3, $4, $5)
		ON CONFLICT (key, customer_id) DO UPDATE SET
			request_hash = EXCLUDED.request_hash,
			order_id = EXCLUDED.order_id
	`
	_, err := r.pool.Exec(ctx, query,
		record.Key,
		record.CustomerID,
		record.RequestHash,
		record.OrderID,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save idempotency record: %w", err)
	}
	return nil
}
