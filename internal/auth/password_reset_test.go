package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestCreatePasswordResetTokenStoresHashOnly(t *testing.T) {
	t.Parallel()

	store := &stubPasswordResetStore{}
	svc := NewService(&stubUserStore{}, &stubSessionStore{}, &stubVerificationStore{}, store)

	raw, err := svc.CreatePasswordResetToken(context.Background(), "u1")
	if err != nil {
		t.Fatalf("CreatePasswordResetToken: %v", err)
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
	if got := PasswordResetTTL(); got != time.Hour {
		t.Fatalf("PasswordResetTTL = %v, want 1h", got)
	}
}

func TestResetPasswordUpdatesHashAndConsumesTokenOnce(t *testing.T) {
	t.Parallel()

	hash, err := Hash("old-password-here")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	users := &stubUserStore{byEmail: map[string]User{
		"ada@example.com": {ID: "u1", Email: "ada@example.com", PasswordHash: hash},
	}}
	resets := &stubPasswordResetStore{}
	svc := NewService(users, &stubSessionStore{}, &stubVerificationStore{}, resets)

	raw, err := svc.CreatePasswordResetToken(context.Background(), "u1")
	if err != nil {
		t.Fatalf("CreatePasswordResetToken: %v", err)
	}

	fieldErrs, err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:                raw,
		Password:             "new-password-here",
		PasswordConfirmation: "new-password-here",
	})
	if err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if fieldErrs.Any() {
		t.Fatalf("unexpected field errors: %+v", fieldErrs)
	}

	u := users.byEmail["ada@example.com"]
	ok, err := Compare("new-password-here", u.PasswordHash)
	if err != nil || !ok {
		t.Fatalf("new password did not verify: ok=%v err=%v", ok, err)
	}
	tok := resets.byHash[hashPasswordResetToken(raw)]
	if tok.UsedAt == nil {
		t.Fatal("expected token used_at to be set")
	}

	fieldErrs, err = svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:                raw,
		Password:             "another-password",
		PasswordConfirmation: "another-password",
	})
	if err != nil {
		t.Fatalf("second ResetPassword err: %v", err)
	}
	if fieldErrs.Token == "" {
		t.Fatal("want token error on second consume")
	}
}

func TestResetPasswordRejectsExpiredAndMissing(t *testing.T) {
	t.Parallel()

	users := &stubUserStore{byEmail: map[string]User{
		"ada@example.com": {ID: "u1", Email: "ada@example.com"},
	}}
	resets := &stubPasswordResetStore{}
	svc := NewService(users, &stubSessionStore{}, &stubVerificationStore{}, resets)

	fieldErrs, err := svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:                "",
		Password:             "long-enough-pass",
		PasswordConfirmation: "long-enough-pass",
	})
	if err != nil {
		t.Fatalf("empty token err: %v", err)
	}
	if fieldErrs.Token == "" {
		t.Fatal("want token error for empty token")
	}

	fieldErrs, err = svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:                "missing",
		Password:             "long-enough-pass",
		PasswordConfirmation: "long-enough-pass",
	})
	if err != nil {
		t.Fatalf("missing token err: %v", err)
	}
	if fieldErrs.Token == "" {
		t.Fatal("want token error for missing token")
	}

	raw, err := svc.CreatePasswordResetToken(context.Background(), "u1")
	if err != nil {
		t.Fatalf("CreatePasswordResetToken: %v", err)
	}
	hash := hashPasswordResetToken(raw)
	tok := resets.byHash[hash]
	tok.ExpiresAt = time.Now().UTC().Add(-time.Minute)
	resets.byHash[hash] = tok

	fieldErrs, err = svc.ResetPassword(context.Background(), ResetPasswordInput{
		Token:                raw,
		Password:             "long-enough-pass",
		PasswordConfirmation: "long-enough-pass",
	})
	if err != nil {
		t.Fatalf("expired token err: %v", err)
	}
	if fieldErrs.Token == "" {
		t.Fatal("want token error for expired token")
	}
}

func TestRequestPasswordResetHidesMissingAccount(t *testing.T) {
	t.Parallel()

	users := &stubUserStore{byEmail: map[string]User{
		"ada@example.com": {ID: "u1", Email: "ada@example.com", DisplayName: "Ada"},
	}}
	resets := &stubPasswordResetStore{}
	svc := NewService(users, &stubSessionStore{}, &stubVerificationStore{}, resets)

	raw, user, err := svc.RequestPasswordReset(context.Background(), "nobody@example.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset missing: %v", err)
	}
	if raw != "" || user.ID != "" {
		t.Fatalf("missing account should yield empty token/user, got token=%q user=%+v", raw, user)
	}

	raw, user, err = svc.RequestPasswordReset(context.Background(), "Ada@Example.com")
	if err != nil {
		t.Fatalf("RequestPasswordReset existing: %v", err)
	}
	if raw == "" || user.ID != "u1" {
		t.Fatalf("existing account should yield token and user, got token=%q user=%+v", raw, user)
	}
}
