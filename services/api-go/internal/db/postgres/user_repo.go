package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/altafur-fuad/Aupa/services/api-go/internal/auth"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO public.users (id, name, email, phone, password_hash, role, is_active, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.FullName,
		strings.ToLower(user.Email),
		user.Phone,
		user.PasswordHash,
		string(user.Role),
		true,
		user.IsVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Unique violation
			return auth.ErrUserAlreadyExists
		}
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*auth.User, error) {
	query := `
		SELECT id, name, COALESCE(email, ''), COALESCE(phone, ''), password_hash, role, email_verified, created_at, updated_at
		FROM public.users
		WHERE id = $1
	`
	var u auth.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&roleStr,
		&u.IsVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	u.Role = auth.Role(roleStr)
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, name, COALESCE(email, ''), COALESCE(phone, ''), password_hash, role, email_verified, created_at, updated_at
		FROM public.users
		WHERE lower(email) = lower($1)
	`
	var u auth.User
	var roleStr string
	err := r.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&u.ID,
		&u.FullName,
		&u.Email,
		&u.Phone,
		&u.PasswordHash,
		&roleStr,
		&u.IsVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	u.Role = auth.Role(roleStr)
	return &u, nil
}

func (r *PostgresUserRepository) UpdateVerification(ctx context.Context, id string, verified bool) error {
	query := `
		UPDATE public.users
		SET email_verified = $2, updated_at = $3
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, id, verified, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update user verification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return auth.ErrUserNotFound
	}
	return nil
}

type PostgresSessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *PostgresSessionRepository {
	return &PostgresSessionRepository{pool: pool}
}

func (r *PostgresSessionRepository) Create(ctx context.Context, session *auth.Session) error {
	query := `
		INSERT INTO public.refresh_tokens (id, user_id, token_hash, family_id, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.FamilyID,
		session.ExpiresAt,
		session.IsRevoked,
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token session: %w", err)
	}
	return nil
}

func (r *PostgresSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.Session, error) {
	query := `
		SELECT id, user_id, token_hash, family_id, expires_at, revoked, created_at
		FROM public.refresh_tokens
		WHERE token_hash = $1
	`
	var s auth.Session
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.FamilyID,
		&s.ExpiresAt,
		&s.IsRevoked,
		&s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to query refresh token: %w", err)
	}
	return &s, nil
}

func (r *PostgresSessionRepository) Revoke(ctx context.Context, sessionID string) error {
	query := `UPDATE public.refresh_tokens SET revoked = true WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, sessionID)
	return err
}

func (r *PostgresSessionRepository) RevokeFamily(ctx context.Context, familyID string) error {
	query := `UPDATE public.refresh_tokens SET revoked = true WHERE family_id = $1`
	_, err := r.pool.Exec(ctx, query, familyID)
	return err
}

func (r *PostgresSessionRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `UPDATE public.refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
