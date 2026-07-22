package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestHashSessionTokenIsSHA256Hex(t *testing.T) {
	t.Parallel()

	raw := "example-raw-session-token"
	got := hashSessionToken(raw)
	sum := sha256.Sum256([]byte(raw))
	want := hex.EncodeToString(sum[:])
	if got != want {
		t.Fatalf("hashSessionToken = %q, want %q", got, want)
	}
	if got == raw {
		t.Fatal("hash must not equal raw token")
	}
}

func TestGenerateSessionTokenUniqueAndNotStoredRaw(t *testing.T) {
	t.Parallel()

	a, err := generateSessionToken()
	if err != nil {
		t.Fatalf("generateSessionToken: %v", err)
	}
	b, err := generateSessionToken()
	if err != nil {
		t.Fatalf("generateSessionToken: %v", err)
	}
	if a == "" || b == "" {
		t.Fatal("tokens must be non-empty")
	}
	if a == b {
		t.Fatal("tokens should be unique")
	}
	if hashSessionToken(a) == a {
		t.Fatal("raw token must not equal its hash")
	}
}

func TestCreateSessionStoresHashOnly(t *testing.T) {
	t.Parallel()

	sessions := &stubSessionStore{}
	svc := NewService(&stubUserStore{}, sessions, &stubVerificationStore{}, &stubPasswordResetStore{})

	sess, rawToken, err := svc.CreateSession(context.Background(), CreateSessionInput{
		UserID:    "user-1",
		UserAgent: "test-agent",
		IPAddress: "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if rawToken == "" {
		t.Fatal("expected raw token for cookie")
	}
	if sess.TokenHash == "" || sess.TokenHash == rawToken {
		t.Fatalf("session must store hash only; hash=%q raw=%q", sess.TokenHash, rawToken)
	}
	if sess.TokenHash != hashSessionToken(rawToken) {
		t.Fatalf("stored hash mismatch: got %q want %q", sess.TokenHash, hashSessionToken(rawToken))
	}
	if stored, ok := sessions.byHash[sess.TokenHash]; !ok {
		t.Fatal("session not stored under hash key")
	} else if stored.TokenHash == rawToken {
		t.Fatal("store must not persist raw token")
	}
}

func TestLoadSessionRejectsExpiredAndRevoked(t *testing.T) {
	t.Parallel()

	sessions := &stubSessionStore{byHash: map[string]Session{}}
	svc := NewService(&stubUserStore{}, sessions, &stubVerificationStore{}, &stubPasswordResetStore{})

	rawOK := "active-token"
	hashOK := hashSessionToken(rawOK)
	sessions.byHash[hashOK] = Session{
		ID:        "active",
		UserID:    "u1",
		TokenHash: hashOK,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}

	if _, err := svc.LoadSession(context.Background(), rawOK); err != nil {
		t.Fatalf("active session should load: %v", err)
	}

	rawExpired := "expired-token"
	hashExpired := hashSessionToken(rawExpired)
	sessions.byHash[hashExpired] = Session{
		ID:        "expired",
		UserID:    "u1",
		TokenHash: hashExpired,
		ExpiresAt: time.Now().UTC().Add(-time.Minute),
	}
	if _, err := svc.LoadSession(context.Background(), rawExpired); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("expired err = %v, want ErrInvalidSession", err)
	}

	rawRevoked := "revoked-token"
	hashRevoked := hashSessionToken(rawRevoked)
	now := time.Now().UTC()
	sessions.byHash[hashRevoked] = Session{
		ID:        "revoked",
		UserID:    "u1",
		TokenHash: hashRevoked,
		ExpiresAt: time.Now().UTC().Add(time.Hour),
		RevokedAt: &now,
	}
	if _, err := svc.LoadSession(context.Background(), rawRevoked); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("revoked err = %v, want ErrInvalidSession", err)
	}
}

func TestLoginGenericErrorAndSessionIssue(t *testing.T) {
	t.Parallel()

	hash, err := Hash("correct-horse-battery")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	users := &stubUserStore{
		byEmail: map[string]User{
			"ada@example.com": {ID: "u1", Email: "ada@example.com", PasswordHash: hash},
		},
	}
	sessions := &stubSessionStore{}
	svc := NewService(users, sessions, &stubVerificationStore{}, &stubPasswordResetStore{})

	_, _, err = svc.Login(context.Background(), LoginInput{
		Email:    "missing@example.com",
		Password: "correct-horse-battery",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("missing user err = %v, want ErrInvalidCredentials", err)
	}

	_, _, err = svc.Login(context.Background(), LoginInput{
		Email:    "ada@example.com",
		Password: "wrong-password-here",
	})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("bad password err = %v, want ErrInvalidCredentials", err)
	}

	sess, rawToken, err := svc.Login(context.Background(), LoginInput{
		Email:     "  Ada@Example.COM ",
		Password:  "correct-horse-battery",
		UserAgent: "ua",
		IPAddress: "10.0.0.1",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if rawToken == "" || sess.TokenHash == rawToken {
		t.Fatal("login must return raw cookie token and store hash only")
	}
	if sess.UserID != "u1" {
		t.Fatalf("user id = %q, want u1", sess.UserID)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	t.Parallel()

	sessions := &stubSessionStore{}
	svc := NewService(&stubUserStore{}, sessions, &stubVerificationStore{}, &stubPasswordResetStore{})

	_, rawToken, err := svc.CreateSession(context.Background(), CreateSessionInput{UserID: "u1"})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if err := svc.Logout(context.Background(), rawToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if _, err := svc.LoadSession(context.Background(), rawToken); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("after logout LoadSession err = %v, want ErrInvalidSession", err)
	}
}

func TestSessionCookieName(t *testing.T) {
	t.Parallel()

	if got := SessionCookieName(false); got != "forgeboard_session" {
		t.Fatalf("dev cookie = %q", got)
	}
	if got := SessionCookieName(true); got != "__Host-forgeboard_session" {
		t.Fatalf("secure cookie = %q", got)
	}
	if !strings.HasPrefix(SessionCookieName(true), "__Host-") {
		t.Fatal("production cookie should use __Host- prefix")
	}
}
