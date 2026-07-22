package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"
)

func TestCreateEmailVerificationTokenStoresHashOnly(t *testing.T) {
	t.Parallel()

	store := &stubVerificationStore{}
	svc := NewService(&stubUserStore{}, &stubSessionStore{}, store, &stubPasswordResetStore{})

	raw, err := svc.CreateEmailVerificationToken(context.Background(), "u1")
	if err != nil {
		t.Fatalf("CreateEmailVerificationToken: %v", err)
	}
	if raw == "" {
		t.Fatal("raw token must be non-empty")
	}

	sum := sha256.Sum256([]byte(raw))
	wantHash := hex.EncodeToString(sum[:])
	tok, ok := store.byHash[wantHash]
	if !ok {
		t.Fatalf("token hash %q not stored", wantHash)
	}
	if _, storedRaw := store.byHash[raw]; storedRaw {
		t.Fatal("raw token must not be stored as a key")
	}
	if tok.UserID != "u1" {
		t.Fatalf("user_id = %q, want u1", tok.UserID)
	}
	if !tok.ExpiresAt.After(time.Now().UTC()) {
		t.Fatal("token should expire in the future")
	}
	if got := EmailVerificationTTL(); got != 24*time.Hour {
		t.Fatalf("EmailVerificationTTL = %v, want 24h", got)
	}
}

func TestVerifyEmailMarksUserAndConsumesToken(t *testing.T) {
	t.Parallel()

	users := &stubUserStore{byEmail: map[string]User{
		"ada@example.com": {ID: "u1", Email: "ada@example.com", DisplayName: "Ada"},
	}}
	verifications := &stubVerificationStore{}
	svc := NewService(users, &stubSessionStore{}, verifications, &stubPasswordResetStore{})

	raw, err := svc.CreateEmailVerificationToken(context.Background(), "u1")
	if err != nil {
		t.Fatalf("CreateEmailVerificationToken: %v", err)
	}

	if err := svc.VerifyEmail(context.Background(), raw); err != nil {
		t.Fatalf("VerifyEmail: %v", err)
	}

	u := users.byEmail["ada@example.com"]
	if u.EmailVerifiedAt == nil {
		t.Fatal("expected email_verified_at to be set")
	}

	tok := verifications.byHash[hashVerificationToken(raw)]
	if tok.UsedAt == nil {
		t.Fatal("expected token used_at to be set")
	}

	if err := svc.VerifyEmail(context.Background(), raw); !errors.Is(err, ErrInvalidVerificationToken) {
		t.Fatalf("second VerifyEmail err = %v, want ErrInvalidVerificationToken", err)
	}
}

func TestVerifyEmailRejectsExpiredAndMissing(t *testing.T) {
	t.Parallel()

	users := &stubUserStore{byEmail: map[string]User{
		"ada@example.com": {ID: "u1", Email: "ada@example.com"},
	}}
	verifications := &stubVerificationStore{}
	svc := NewService(users, &stubSessionStore{}, verifications, &stubPasswordResetStore{})

	if err := svc.VerifyEmail(context.Background(), ""); !errors.Is(err, ErrInvalidVerificationToken) {
		t.Fatalf("empty token err = %v, want ErrInvalidVerificationToken", err)
	}
	if err := svc.VerifyEmail(context.Background(), "nope"); !errors.Is(err, ErrInvalidVerificationToken) {
		t.Fatalf("missing token err = %v, want ErrInvalidVerificationToken", err)
	}

	raw, err := svc.CreateEmailVerificationToken(context.Background(), "u1")
	if err != nil {
		t.Fatalf("CreateEmailVerificationToken: %v", err)
	}
	hash := hashVerificationToken(raw)
	tok := verifications.byHash[hash]
	tok.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	verifications.byHash[hash] = tok

	if err := svc.VerifyEmail(context.Background(), raw); !errors.Is(err, ErrInvalidVerificationToken) {
		t.Fatalf("expired token err = %v, want ErrInvalidVerificationToken", err)
	}
}
