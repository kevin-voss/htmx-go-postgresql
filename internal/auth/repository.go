package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when no user matches the lookup.
var ErrNotFound = errors.New("auth: user not found")

// ErrDuplicateEmail is returned when email already exists.
var ErrDuplicateEmail = errors.New("auth: duplicate email")

// Repository persists users in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a user repository backed by pool.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Create inserts a user and returns the stored row.
func (r *Repository) Create(ctx context.Context, email, displayName, passwordHash string) (User, error) {
	const q = `
		INSERT INTO users (email, display_name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, display_name, password_hash, email_verified_at, created_at, updated_at`

	var u User
	err := r.db.QueryRow(ctx, q, email, displayName, passwordHash).Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.PasswordHash,
		&u.EmailVerifiedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrDuplicateEmail
		}
		return User{}, fmt.Errorf("auth: create user: %w", err)
	}
	return u, nil
}

// GetByEmail returns the user with the given normalized email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	const q = `
		SELECT id, email, display_name, password_hash, email_verified_at, created_at, updated_at
		FROM users
		WHERE email = $1`

	var u User
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&u.DisplayName,
		&u.PasswordHash,
		&u.EmailVerifiedAt,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("auth: get user by email: %w", err)
	}
	return u, nil
}
