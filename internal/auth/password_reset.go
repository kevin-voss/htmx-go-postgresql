package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	passwordResetTTL      = time.Hour
	passwordResetTokenLen = 32
)

// ErrInvalidResetToken is returned for missing, used, or expired reset tokens.
var ErrInvalidResetToken = errors.New("auth: invalid password reset token")

// PasswordResetToken is a persisted password-reset token row.
type PasswordResetToken struct {
	ID        string
	UserID    string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
}

// PasswordResetStore is the persistence port for password-reset tokens.
type PasswordResetStore interface {
	CreatePasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (PasswordResetToken, error)
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, id string, at time.Time) error
}

// ResetPasswordInput is the public reset-password form payload.
type ResetPasswordInput struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

// ResetPasswordErrors holds per-field validation messages for the reset form.
type ResetPasswordErrors struct {
	Password             string
	PasswordConfirmation string
	Token                string
}

// Any reports whether any field error is set.
func (e ResetPasswordErrors) Any() bool {
	return e.Password != "" || e.PasswordConfirmation != "" || e.Token != ""
}

// RequestPasswordReset looks up the email and, when a user exists, creates a
// reset token. Missing accounts are not an error so callers can show a generic
// acknowledgment without revealing account existence.
// Returns the raw token when an email should be sent; empty string otherwise.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) (rawToken string, user User, err error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || !validEmail(normalized) {
		return "", User{}, nil
	}

	user, err = s.users.GetByEmail(ctx, normalized)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", User{}, nil
		}
		return "", User{}, fmt.Errorf("auth: lookup email for reset: %w", err)
	}

	rawToken, err = s.CreatePasswordResetToken(ctx, user.ID)
	if err != nil {
		return "", User{}, err
	}
	return rawToken, user, nil
}

// CreatePasswordResetToken generates a raw token, stores sha256(token), and returns the raw value.
func (s *Service) CreatePasswordResetToken(ctx context.Context, userID string) (string, error) {
	rawToken, err := generatePasswordResetToken()
	if err != nil {
		return "", err
	}
	hash := hashPasswordResetToken(rawToken)
	expiresAt := time.Now().UTC().Add(passwordResetTTL)

	if _, err := s.resets.CreatePasswordResetToken(ctx, userID, hash, expiresAt); err != nil {
		return "", err
	}
	return rawToken, nil
}

// ResetPassword consumes a raw reset token and updates the user's password.
// On validation failure it returns field errors without changing state.
func (s *Service) ResetPassword(ctx context.Context, in ResetPasswordInput) (ResetPasswordErrors, error) {
	var fieldErrs ResetPasswordErrors

	if strings.TrimSpace(in.Token) == "" {
		fieldErrs.Token = "This reset link is invalid or has expired."
		return fieldErrs, nil
	}
	if utf8.RuneCountInString(in.Password) < passwordMin {
		fieldErrs.Password = "Password must be at least 12 characters."
	}
	if in.PasswordConfirmation != in.Password {
		fieldErrs.PasswordConfirmation = "Passwords do not match."
	}
	if fieldErrs.Any() {
		return fieldErrs, nil
	}

	tok, err := s.resets.GetPasswordResetTokenByHash(ctx, hashPasswordResetToken(in.Token))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			fieldErrs.Token = "This reset link is invalid or has expired."
			return fieldErrs, nil
		}
		return ResetPasswordErrors{}, fmt.Errorf("auth: load reset token: %w", err)
	}

	now := time.Now().UTC()
	if tok.UsedAt != nil || !tok.ExpiresAt.After(now) {
		fieldErrs.Token = "This reset link is invalid or has expired."
		return fieldErrs, nil
	}

	hash, err := Hash(in.Password)
	if err != nil {
		return ResetPasswordErrors{}, fmt.Errorf("auth: hash password: %w", err)
	}
	if err := s.users.UpdatePasswordHash(ctx, tok.UserID, hash); err != nil {
		return ResetPasswordErrors{}, fmt.Errorf("auth: update password: %w", err)
	}
	if err := s.resets.MarkPasswordResetTokenUsed(ctx, tok.ID, now); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ResetPasswordErrors{}, ErrInvalidResetToken
		}
		return ResetPasswordErrors{}, fmt.Errorf("auth: consume reset token: %w", err)
	}
	return ResetPasswordErrors{}, nil
}

// PasswordResetTTL returns the token lifetime (for handoff/docs).
func PasswordResetTTL() time.Duration {
	return passwordResetTTL
}

func generatePasswordResetToken() (string, error) {
	buf := make([]byte, passwordResetTokenLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: generate password reset token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashPasswordResetToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
