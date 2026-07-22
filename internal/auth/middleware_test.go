package auth_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

func TestRequireAuthenticationRedirectsBrowser(t *testing.T) {
	t.Parallel()

	h := auth.RequireAuthentication(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not run")
	}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/app", nil))

	if rr.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusFound)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Fatalf("Location = %q, want /login", loc)
	}
}

func TestRequireAuthenticationUnauthorizedForHTMX(t *testing.T) {
	t.Parallel()

	h := auth.RequireAuthentication(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not run")
	}))

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	req.Header.Set("HX-Request", "true")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthenticationAllowsAuthenticated(t *testing.T) {
	t.Parallel()

	h := auth.RequireAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := auth.UserFromContext(r.Context())
		if !ok || u.ID != "u1" {
			t.Fatalf("user missing or unexpected: ok=%v user=%+v", ok, u)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	ctx := contextWithUser(req.Context(), auth.User{ID: "u1", Email: "a@example.com"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req.WithContext(ctx))

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestLoadSessionSetsUserAndSessionInContext(t *testing.T) {
	t.Parallel()

	hash, err := auth.Hash("correct-horse-battery")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	users := &middlewareUserStore{byID: map[string]auth.User{
		"u1": {ID: "u1", Email: "ada@example.com", DisplayName: "Ada", PasswordHash: hash},
	}}
	sessions := &middlewareSessionStore{}
	svc := auth.NewService(users, sessions, &middlewareVerificationStore{})

	_, rawToken, err := svc.CreateSession(context.Background(), auth.CreateSessionInput{UserID: "u1"})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	var sawUser bool
	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := auth.UserFromContext(r.Context())
			if !ok || u.ID != "u1" {
				t.Fatalf("user missing: ok=%v %+v", ok, u)
			}
			sess, ok := auth.SessionFromContext(r.Context())
			if !ok || sess.UserID != "u1" {
				t.Fatalf("session missing: ok=%v %+v", ok, sess)
			}
			sawUser = true
			w.WriteHeader(http.StatusOK)
		}),
		auth.LoadSession(svc, false, slog.New(slog.NewTextHandler(io.Discard, nil))),
	)

	req := httptest.NewRequest(http.MethodGet, "/app", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName(false), Value: rawToken})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !sawUser {
		t.Fatal("handler did not observe authenticated context")
	}
}

func TestLoadSessionIgnoresInvalidCookie(t *testing.T) {
	t.Parallel()

	svc := auth.NewService(&middlewareUserStore{}, &middlewareSessionStore{}, &middlewareVerificationStore{})
	h := middleware.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := auth.UserFromContext(r.Context()); ok {
				t.Fatal("expected anonymous context")
			}
			w.WriteHeader(http.StatusOK)
		}),
		auth.LoadSession(svc, false, slog.New(slog.NewTextHandler(io.Discard, nil))),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: auth.SessionCookieName(false), Value: "not-a-real-token"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

// contextWithUser injects a user using LoadSession's private key via a tiny
// round-trip: RequireAuthentication only needs UserFromContext, so we use
// the exported helper path through a test-only wrapper in the auth package.
func contextWithUser(ctx context.Context, user auth.User) context.Context {
	return auth.ContextWithUser(ctx, user)
}

type middlewareUserStore struct {
	byID    map[string]auth.User
	byEmail map[string]auth.User
}

func (s *middlewareUserStore) Create(_ context.Context, email, displayName, passwordHash string) (auth.User, error) {
	u := auth.User{ID: "u-new", Email: email, DisplayName: displayName, PasswordHash: passwordHash}
	return u, nil
}

func (s *middlewareUserStore) GetByEmail(_ context.Context, email string) (auth.User, error) {
	if u, ok := s.byEmail[email]; ok {
		return u, nil
	}
	return auth.User{}, auth.ErrNotFound
}

func (s *middlewareUserStore) GetByID(_ context.Context, id string) (auth.User, error) {
	if u, ok := s.byID[id]; ok {
		return u, nil
	}
	return auth.User{}, auth.ErrNotFound
}

func (s *middlewareUserStore) MarkEmailVerified(_ context.Context, userID string, at time.Time) error {
	if u, ok := s.byID[userID]; ok {
		u.EmailVerifiedAt = &at
		s.byID[userID] = u
		return nil
	}
	return auth.ErrNotFound
}

type middlewareVerificationStore struct{}

func (middlewareVerificationStore) CreateEmailVerificationToken(context.Context, string, string, time.Time) (auth.EmailVerificationToken, error) {
	return auth.EmailVerificationToken{}, nil
}

func (middlewareVerificationStore) GetEmailVerificationTokenByHash(context.Context, string) (auth.EmailVerificationToken, error) {
	return auth.EmailVerificationToken{}, auth.ErrNotFound
}

func (middlewareVerificationStore) MarkEmailVerificationTokenUsed(context.Context, string, time.Time) error {
	return auth.ErrNotFound
}

type middlewareSessionStore struct {
	byHash map[string]auth.Session
}

func (s *middlewareSessionStore) CreateSession(_ context.Context, userID, tokenHash string, expiresAt time.Time, userAgent, ipAddress string) (auth.Session, error) {
	sess := auth.Session{
		ID:         "s1",
		UserID:     userID,
		TokenHash:  tokenHash,
		CreatedAt:  time.Now().UTC(),
		LastSeenAt: time.Now().UTC(),
		ExpiresAt:  expiresAt,
		UserAgent:  userAgent,
		IPAddress:  ipAddress,
	}
	if s.byHash == nil {
		s.byHash = map[string]auth.Session{}
	}
	s.byHash[tokenHash] = sess
	return sess, nil
}

func (s *middlewareSessionStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (auth.Session, error) {
	if sess, ok := s.byHash[tokenHash]; ok {
		return sess, nil
	}
	return auth.Session{}, auth.ErrNotFound
}

func (s *middlewareSessionStore) RevokeSessionByTokenHash(_ context.Context, tokenHash string) error {
	sess, ok := s.byHash[tokenHash]
	if !ok {
		return auth.ErrNotFound
	}
	now := time.Now().UTC()
	sess.RevokedAt = &now
	s.byHash[tokenHash] = sess
	return nil
}

func (s *middlewareSessionStore) TouchSession(_ context.Context, id string, at time.Time) error {
	for hash, sess := range s.byHash {
		if sess.ID == id {
			sess.LastSeenAt = at
			s.byHash[hash] = sess
			return nil
		}
	}
	return auth.ErrNotFound
}
