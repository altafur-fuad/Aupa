package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/delivery"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOutboxRepository struct {
	pool *pgxpool.Pool
}

func NewOutboxRepository(pool *pgxpool.Pool) *PostgresOutboxRepository {
	return &PostgresOutboxRepository{pool: pool}
}

func (r *PostgresOutboxRepository) Save(ctx context.Context, event *delivery.OutboxEvent) error {
	query := `
		INSERT INTO public.outbox_events (id, aggregate_type, aggregate_id, event_type, payload, retry_count, created_at)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		event.ID,
		event.AggregateType,
		event.AggregateID,
		event.EventType,
		event.Payload,
		event.RetryCount,
		event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	return nil
}

func (r *PostgresOutboxRepository) GetPending(ctx context.Context, limit int) ([]delivery.OutboxEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id, aggregate_type, aggregate_id, event_type, payload::text, retry_count, created_at, processed_at
		FROM public.outbox_events
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox events: %w", err)
	}
	defer rows.Close()

	var events []delivery.OutboxEvent
	for rows.Next() {
		var evt delivery.OutboxEvent
		if err := rows.Scan(
			&evt.ID,
			&evt.AggregateType,
			&evt.AggregateID,
			&evt.EventType,
			&evt.Payload,
			&evt.RetryCount,
			&evt.CreatedAt,
			&evt.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}
		events = append(events, evt)
	}
	return events, nil
}

func (r *PostgresOutboxRepository) MarkProcessed(ctx context.Context, id string) error {
	query := `
		UPDATE public.outbox_events
		SET processed_at = $2
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to mark outbox event processed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("outbox event not found")
	}
	return nil
}
