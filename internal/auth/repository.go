package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when no user or session matches the lookup.
var ErrNotFound = errors.New("auth: user not found")

// ErrDuplicateEmail is returned when email already exists.
var ErrDuplicateEmail = errors.New("auth: duplicate email")

// Repository persists users and sessions in PostgreSQL.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository constructs a repository backed by pool.
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

// GetByID returns the user with the given id.
func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	const q = `
		SELECT id, email, display_name, password_hash, email_verified_at, created_at, updated_at
		FROM users
		WHERE id = $1`

	var u User
	err := r.db.QueryRow(ctx, q, id).Scan(
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
		return User{}, fmt.Errorf("auth: get user by id: %w", err)
	}
	return u, nil
}

// MarkEmailVerified sets email_verified_at when not already set.
func (r *Repository) MarkEmailVerified(ctx context.Context, userID string, at time.Time) error {
	const q = `
		UPDATE users
		SET email_verified_at = COALESCE(email_verified_at, $2),
		    updated_at = $2
		WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, userID, at)
	if err != nil {
		return fmt.Errorf("auth: mark email verified: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdatePasswordHash replaces the user's password hash.
func (r *Repository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	const q = `
		UPDATE users
		SET password_hash = $2,
		    updated_at = now()
		WHERE id = $1`

	tag, err := r.db.Exec(ctx, q, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("auth: update password hash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreatePasswordResetToken inserts a password-reset token (hash only).
func (r *Repository) CreatePasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (PasswordResetToken, error) {
	const q = `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, created_at, expires_at, used_at`

	var t PasswordResetToken
	err := r.db.QueryRow(ctx, q, userID, tokenHash, expiresAt).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.CreatedAt,
		&t.ExpiresAt,
		&t.UsedAt,
	)
	if err != nil {
		return PasswordResetToken{}, fmt.Errorf("auth: create password reset token: %w", err)
	}
	return t, nil
}

// GetPasswordResetTokenByHash returns the token with the given hash.
func (r *Repository) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (PasswordResetToken, error) {
	const q = `
		SELECT id, user_id, token_hash, created_at, expires_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1`

	var t PasswordResetToken
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.CreatedAt,
		&t.ExpiresAt,
		&t.UsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PasswordResetToken{}, ErrNotFound
		}
		return PasswordResetToken{}, fmt.Errorf("auth: get password reset token: %w", err)
	}
	return t, nil
}

// MarkPasswordResetTokenUsed sets used_at for an unused token.
func (r *Repository) MarkPasswordResetTokenUsed(ctx context.Context, id string, at time.Time) error {
	const q = `
		UPDATE password_reset_tokens
		SET used_at = $2
		WHERE id = $1 AND used_at IS NULL`

	tag, err := r.db.Exec(ctx, q, id, at)
	if err != nil {
		return fmt.Errorf("auth: mark password reset token used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateEmailVerificationToken inserts a verification token (hash only).
func (r *Repository) CreateEmailVerificationToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (EmailVerificationToken, error) {
	const q = `
		INSERT INTO email_verification_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, created_at, expires_at, used_at`

	var t EmailVerificationToken
	err := r.db.QueryRow(ctx, q, userID, tokenHash, expiresAt).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.CreatedAt,
		&t.ExpiresAt,
		&t.UsedAt,
	)
	if err != nil {
		return EmailVerificationToken{}, fmt.Errorf("auth: create email verification token: %w", err)
	}
	return t, nil
}

// GetEmailVerificationTokenByHash returns the token with the given hash.
func (r *Repository) GetEmailVerificationTokenByHash(ctx context.Context, tokenHash string) (EmailVerificationToken, error) {
	const q = `
		SELECT id, user_id, token_hash, created_at, expires_at, used_at
		FROM email_verification_tokens
		WHERE token_hash = $1`

	var t EmailVerificationToken
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&t.ID,
		&t.UserID,
		&t.TokenHash,
		&t.CreatedAt,
		&t.ExpiresAt,
		&t.UsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmailVerificationToken{}, ErrNotFound
		}
		return EmailVerificationToken{}, fmt.Errorf("auth: get email verification token: %w", err)
	}
	return t, nil
}

// MarkEmailVerificationTokenUsed sets used_at for an unused token.
func (r *Repository) MarkEmailVerificationTokenUsed(ctx context.Context, id string, at time.Time) error {
	const q = `
		UPDATE email_verification_tokens
		SET used_at = $2
		WHERE id = $1 AND used_at IS NULL`

	tag, err := r.db.Exec(ctx, q, id, at)
	if err != nil {
		return fmt.Errorf("auth: mark email verification token used: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSession inserts a session row (token_hash only — never the raw token).
func (r *Repository) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time, userAgent, ipAddress string) (Session, error) {
	const q = `
		INSERT INTO sessions (user_id, token_hash, expires_at, user_agent, ip_address)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
		RETURNING id, user_id, token_hash, created_at, last_seen_at, expires_at,
			COALESCE(user_agent, ''), COALESCE(ip_address, ''), revoked_at`

	var s Session
	err := r.db.QueryRow(ctx, q, userID, tokenHash, expiresAt, userAgent, ipAddress).Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.CreatedAt,
		&s.LastSeenAt,
		&s.ExpiresAt,
		&s.UserAgent,
		&s.IPAddress,
		&s.RevokedAt,
	)
	if err != nil {
		return Session{}, fmt.Errorf("auth: create session: %w", err)
	}
	return s, nil
}

// GetSessionByTokenHash returns the session with the given token hash.
func (r *Repository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	const q = `
		SELECT id, user_id, token_hash, created_at, last_seen_at, expires_at,
			COALESCE(user_agent, ''), COALESCE(ip_address, ''), revoked_at
		FROM sessions
		WHERE token_hash = $1`

	var s Session
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&s.ID,
		&s.UserID,
		&s.TokenHash,
		&s.CreatedAt,
		&s.LastSeenAt,
		&s.ExpiresAt,
		&s.UserAgent,
		&s.IPAddress,
		&s.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("auth: get session by token hash: %w", err)
	}
	return s, nil
}

// ListActiveSessionsByUserID returns non-revoked, non-expired sessions for a user.
func (r *Repository) ListActiveSessionsByUserID(ctx context.Context, userID string) ([]Session, error) {
	const q = `
		SELECT id, user_id, token_hash, created_at, last_seen_at, expires_at,
			COALESCE(user_agent, ''), COALESCE(ip_address, ''), revoked_at
		FROM sessions
		WHERE user_id = $1
		  AND revoked_at IS NULL
		  AND expires_at > now()
		ORDER BY last_seen_at DESC`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("auth: list sessions by user: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.TokenHash,
			&s.CreatedAt,
			&s.LastSeenAt,
			&s.ExpiresAt,
			&s.UserAgent,
			&s.IPAddress,
			&s.RevokedAt,
		); err != nil {
			return nil, fmt.Errorf("auth: scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("auth: list sessions by user: %w", err)
	}
	return sessions, nil
}

// RevokeSessionByTokenHash sets revoked_at for the matching session.
func (r *Repository) RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error {
	const q = `
		UPDATE sessions
		SET revoked_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL`

	tag, err := r.db.Exec(ctx, q, tokenHash)
	if err != nil {
		return fmt.Errorf("auth: revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// RevokeSessionByIDForUser revokes a session only when it belongs to userID.
func (r *Repository) RevokeSessionByIDForUser(ctx context.Context, sessionID, userID string) error {
	const q = `
		UPDATE sessions
		SET revoked_at = now()
		WHERE id = $1 AND user_id = $2 AND revoked_at IS NULL`

	tag, err := r.db.Exec(ctx, q, sessionID, userID)
	if err != nil {
		return fmt.Errorf("auth: revoke session by id: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// TouchSession updates last_seen_at for an active session.
func (r *Repository) TouchSession(ctx context.Context, id string, at time.Time) error {
	const q = `
		UPDATE sessions
		SET last_seen_at = $2
		WHERE id = $1 AND revoked_at IS NULL`

	tag, err := r.db.Exec(ctx, q, id, at)
	if err != nil {
		return fmt.Errorf("auth: touch session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
