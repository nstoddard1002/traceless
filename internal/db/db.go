package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Secret represents a stored encrypted secret.
type Secret struct {
	ID              string    `json:"id"`
	Ciphertext      string    `json:"ciphertext"`
	CreatedAt       time.Time `json:"created_at"`
	ExpiresAt       time.Time `json:"expires_at"`
	RemainingViews  int       `json:"remaining_views"`
	PasscodeEnabled bool      `json:"passcode_enabled"`
	Salt            string    `json:"salt,omitempty"`
}

// DB wraps the pgxpool.Pool.
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a new DB connection pool.
func NewDB(ctx context.Context, connStr string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Ping the database to ensure connection is established
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() {
	db.Pool.Close()
}

// CreateSecret inserts a new secret into the database.
func (db *DB) CreateSecret(ctx context.Context, s *Secret) error {
	query := `
		INSERT INTO secrets (id, ciphertext, expires_at, remaining_views, passcode_enabled, salt)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Pool.Exec(ctx, query, s.ID, s.Ciphertext, s.ExpiresAt, s.RemainingViews, s.PasscodeEnabled, s.Salt)
	if err != nil {
		return fmt.Errorf("failed to insert secret: %w", err)
	}
	return nil
}

// GetSecret retrieves a secret by its ID.
func (db *DB) GetSecret(ctx context.Context, id string) (*Secret, error) {
	query := `
		SELECT id, ciphertext, created_at, expires_at, remaining_views, passcode_enabled, salt
		FROM secrets
		WHERE id = $1
	`
	var s Secret
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Ciphertext, &s.CreatedAt, &s.ExpiresAt, &s.RemainingViews, &s.PasscodeEnabled, &s.Salt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query secret: %w", err)
	}
	return &s, nil
}

// GetSecretMeta retrieves secret metadata without the ciphertext.
func (db *DB) GetSecretMeta(ctx context.Context, id string) (*Secret, error) {
	query := `
		SELECT id, created_at, expires_at, remaining_views, passcode_enabled, salt
		FROM secrets
		WHERE id = $1
	`
	var s Secret
	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.CreatedAt, &s.ExpiresAt, &s.RemainingViews, &s.PasscodeEnabled, &s.Salt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query secret meta: %w", err)
	}
	return &s, nil
}

// DecrementRemainingViews reduces the remaining views by 1 and returns the new count.
func (db *DB) DecrementRemainingViews(ctx context.Context, id string) (int, error) {
	query := `
		UPDATE secrets
		SET remaining_views = remaining_views - 1
		WHERE id = $1
		RETURNING remaining_views
	`
	var remainingViews int
	err := db.Pool.QueryRow(ctx, query, id).Scan(&remainingViews)
	if err != nil {
		return 0, fmt.Errorf("failed to decrement remaining views: %w", err)
	}
	return remainingViews, nil
}

// DeleteSecret removes a secret from the database.
func (db *DB) DeleteSecret(ctx context.Context, id string) error {
	query := `DELETE FROM secrets WHERE id = $1`
	_, err := db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// DeleteExpiredSecrets removes all secrets that have passed their expiration time.
func (db *DB) DeleteExpiredSecrets(ctx context.Context) (int64, error) {
	query := `DELETE FROM secrets WHERE expires_at <= $1`
	tag, err := db.Pool.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired secrets: %w", err)
	}
	return tag.RowsAffected(), nil
}
