package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/payments"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresPaymentAttemptRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentAttemptRepository(pool *pgxpool.Pool) *PostgresPaymentAttemptRepository {
	return &PostgresPaymentAttemptRepository{pool: pool}
}

func (r *PostgresPaymentAttemptRepository) Create(ctx context.Context, attempt *payments.PaymentAttempt) error {
	query := `
		INSERT INTO public.payment_attempts (id, order_id, customer_id, provider, amount_poisha, currency, status, transaction_id, redirect_url, created_at, updated_at)
		VALUES ($1, $2, $3::uuid, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		attempt.ID,
		attempt.OrderID,
		attempt.CustomerID,
		attempt.Provider,
		attempt.AmountPoisha,
		attempt.Currency,
		string(attempt.Status),
		attempt.TransactionID,
		attempt.RedirectURL,
		attempt.CreatedAt,
		attempt.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment attempt: %w", err)
	}
	return nil
}

func (r *PostgresPaymentAttemptRepository) GetByID(ctx context.Context, id string) (*payments.PaymentAttempt, error) {
	query := `
		SELECT id, order_id, customer_id, provider, amount_poisha, currency, status,
		       COALESCE(transaction_id, ''), COALESCE(redirect_url, ''), created_at, updated_at
		FROM public.payment_attempts
		WHERE id = $1
	`
	var att payments.PaymentAttempt
	var statusStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&att.ID,
		&att.OrderID,
		&att.CustomerID,
		&att.Provider,
		&att.AmountPoisha,
		&att.Currency,
		&statusStr,
		&att.TransactionID,
		&att.RedirectURL,
		&att.CreatedAt,
		&att.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, payments.ErrAttemptNotFound
		}
		return nil, fmt.Errorf("failed to query payment attempt: %w", err)
	}
	att.Status = payments.PaymentAttemptStatus(statusStr)
	return &att, nil
}

func (r *PostgresPaymentAttemptRepository) GetByOrderID(ctx context.Context, orderID string) ([]payments.PaymentAttempt, error) {
	query := `
		SELECT id, order_id, customer_id, provider, amount_poisha, currency, status,
		       COALESCE(transaction_id, ''), COALESCE(redirect_url, ''), created_at, updated_at
		FROM public.payment_attempts
		WHERE order_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query payment attempts by order: %w", err)
	}
	defer rows.Close()

	var list []payments.PaymentAttempt
	for rows.Next() {
		var att payments.PaymentAttempt
		var statusStr string
		if err := rows.Scan(
			&att.ID,
			&att.OrderID,
			&att.CustomerID,
			&att.Provider,
			&att.AmountPoisha,
			&att.Currency,
			&statusStr,
			&att.TransactionID,
			&att.RedirectURL,
			&att.CreatedAt,
			&att.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan payment attempt: %w", err)
		}
		att.Status = payments.PaymentAttemptStatus(statusStr)
		list = append(list, att)
	}
	return list, nil
}

func (r *PostgresPaymentAttemptRepository) UpdateStatus(ctx context.Context, id string, status payments.PaymentAttemptStatus, txID string) error {
	query := `
		UPDATE public.payment_attempts
		SET status = $2, transaction_id = COALESCE(NULLIF($3, ''), transaction_id), updated_at = $4
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, string(status), txID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update payment attempt status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return payments.ErrAttemptNotFound
	}
	return nil
}

type PostgresWebhookAuditRepository struct {
	pool *pgxpool.Pool
}

func NewWebhookAuditRepository(pool *pgxpool.Pool) *PostgresWebhookAuditRepository {
	return &PostgresWebhookAuditRepository{pool: pool}
}

func (r *PostgresWebhookAuditRepository) HasProcessed(ctx context.Context, provider string, eventID string) (bool, error) {
	query := `SELECT 1 FROM public.webhook_audits WHERE provider = $1 AND event_id = $2`
	var dummy int
	err := r.pool.QueryRow(ctx, query, provider, eventID).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to query webhook audit: %w", err)
	}
	return true, nil
}

func (r *PostgresWebhookAuditRepository) MarkProcessed(ctx context.Context, provider string, eventID string, orderID string) error {
	query := `
		INSERT INTO public.webhook_audits (id, provider, event_id, order_id, processed_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		ON CONFLICT (provider, event_id) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, provider, eventID, orderID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to record webhook audit: %w", err)
	}
	return nil
}
