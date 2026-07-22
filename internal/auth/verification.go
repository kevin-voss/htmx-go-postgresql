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
)

const (
	emailVerificationTTL      = 24 * time.Hour
	emailVerificationTokenLen = 32
)

// ErrInvalidVerificationToken is returned for missing, used, or expired tokens.
var ErrInvalidVerificationToken = errors.New("auth: invalid verification token")

// EmailVerificationToken is a persisted verification token row.
type EmailVerificationToken struct {
	ID        string
	UserID    string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
}

// VerificationStore is the persistence port for email verification tokens.
type VerificationStore interface {
	CreateEmailVerificationToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (EmailVerificationToken, error)
	GetEmailVerificationTokenByHash(ctx context.Context, tokenHash string) (EmailVerificationToken, error)
	MarkEmailVerificationTokenUsed(ctx context.Context, id string, at time.Time) error
}

// CreateEmailVerificationToken generates a raw token, stores sha256(token), and returns the raw value.
func (s *Service) CreateEmailVerificationToken(ctx context.Context, userID string) (string, error) {
	rawToken, err := generateVerificationToken()
	if err != nil {
		return "", err
	}
	hash := hashVerificationToken(rawToken)
	expiresAt := time.Now().UTC().Add(emailVerificationTTL)

	if _, err := s.verifications.CreateEmailVerificationToken(ctx, userID, hash, expiresAt); err != nil {
		return "", err
	}
	return rawToken, nil
}

// VerifyEmail consumes a raw verification token and marks the user's email verified.
func (s *Service) VerifyEmail(ctx context.Context, rawToken string) error {
	if strings.TrimSpace(rawToken) == "" {
		return ErrInvalidVerificationToken
	}

	tok, err := s.verifications.GetEmailVerificationTokenByHash(ctx, hashVerificationToken(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrInvalidVerificationToken
		}
		return fmt.Errorf("auth: load verification token: %w", err)
	}

	now := time.Now().UTC()
	if tok.UsedAt != nil || !tok.ExpiresAt.After(now) {
		return ErrInvalidVerificationToken
	}

	if err := s.users.MarkEmailVerified(ctx, tok.UserID, now); err != nil {
		return fmt.Errorf("auth: mark email verified: %w", err)
	}
	if err := s.verifications.MarkEmailVerificationTokenUsed(ctx, tok.ID, now); err != nil {
		return fmt.Errorf("auth: consume verification token: %w", err)
	}
	return nil
}

// EmailVerificationTTL returns the token lifetime (for handoff/docs).
func EmailVerificationTTL() time.Duration {
	return emailVerificationTTL
}

func generateVerificationToken() (string, error) {
	buf := make([]byte, emailVerificationTokenLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: generate verification token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashVerificationToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
