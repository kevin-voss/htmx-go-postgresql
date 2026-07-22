package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	sessionTTL       = 7 * 24 * time.Hour
	sessionTokenLen  = 32
	cookieNameDev    = "forgeboard_session"
	cookieNameSecure = "__Host-forgeboard_session"
)

// ErrInvalidCredentials is returned for failed login attempts (generic).
var ErrInvalidCredentials = errors.New("auth: invalid email or password")

// ErrInvalidSession is returned when a session token is missing, expired, or revoked.
var ErrInvalidSession = errors.New("auth: invalid session")

// Session is a persisted server-side session row.
type Session struct {
	ID         string
	UserID     string
	TokenHash  string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
	UserAgent  string
	IPAddress  string
	RevokedAt  *time.Time
}

// CreateSessionInput holds metadata captured when issuing a session.
type CreateSessionInput struct {
	UserID    string
	UserAgent string
	IPAddress string
}

// SessionStore is the persistence port for sessions.
type SessionStore interface {
	CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time, userAgent, ipAddress string) (Session, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	ListActiveSessionsByUserID(ctx context.Context, userID string) ([]Session, error)
	RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error
	RevokeSessionByIDForUser(ctx context.Context, sessionID, userID string) error
	TouchSession(ctx context.Context, id string, at time.Time) error
}

// SessionInfo is a safe view of a session for account UI (no token material).
type SessionInfo struct {
	ID         string
	UserAgent  string
	IPAddress  string
	LastSeenAt time.Time
	CreatedAt  time.Time
	Current    bool
}

// LoginInput is the public login form payload.
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// Login authenticates credentials and issues a new session.
// On failure it returns ErrInvalidCredentials without revealing whether the email exists.
// The returned rawToken must be set on the cookie; only its hash is stored.
func (s *Service) Login(ctx context.Context, in LoginInput) (Session, string, error) {
	email := strings.ToLower(strings.TrimSpace(in.Email))
	if email == "" || in.Password == "" {
		return Session{}, "", ErrInvalidCredentials
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, "", ErrInvalidCredentials
		}
		return Session{}, "", fmt.Errorf("auth: login lookup: %w", err)
	}

	ok, err := Compare(in.Password, user.PasswordHash)
	if err != nil || !ok {
		return Session{}, "", ErrInvalidCredentials
	}

	return s.CreateSession(ctx, CreateSessionInput{
		UserID:    user.ID,
		UserAgent: in.UserAgent,
		IPAddress: in.IPAddress,
	})
}

// CreateSession generates a raw token, stores sha256(token), and returns both.
func (s *Service) CreateSession(ctx context.Context, in CreateSessionInput) (Session, string, error) {
	rawToken, err := generateSessionToken()
	if err != nil {
		return Session{}, "", err
	}
	hash := hashSessionToken(rawToken)
	expiresAt := time.Now().UTC().Add(sessionTTL)

	sess, err := s.sessions.CreateSession(ctx, in.UserID, hash, expiresAt, in.UserAgent, in.IPAddress)
	if err != nil {
		return Session{}, "", err
	}
	return sess, rawToken, nil
}

// LoadSession validates a raw cookie token and returns the active session.
// Expired and revoked sessions are rejected.
func (s *Service) LoadSession(ctx context.Context, rawToken string) (Session, error) {
	if strings.TrimSpace(rawToken) == "" {
		return Session{}, ErrInvalidSession
	}

	sess, err := s.sessions.GetSessionByTokenHash(ctx, hashSessionToken(rawToken))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, ErrInvalidSession
		}
		return Session{}, fmt.Errorf("auth: load session: %w", err)
	}

	now := time.Now().UTC()
	if sess.RevokedAt != nil || !sess.ExpiresAt.After(now) {
		return Session{}, ErrInvalidSession
	}

	if err := s.sessions.TouchSession(ctx, sess.ID, now); err != nil {
		return Session{}, fmt.Errorf("auth: touch session: %w", err)
	}
	sess.LastSeenAt = now
	return sess, nil
}

// LoadSessionUser validates the session cookie token and loads the owning user.
func (s *Service) LoadSessionUser(ctx context.Context, rawToken string) (Session, User, error) {
	sess, err := s.LoadSession(ctx, rawToken)
	if err != nil {
		return Session{}, User{}, err
	}
	user, err := s.users.GetByID(ctx, sess.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Session{}, User{}, ErrInvalidSession
		}
		return Session{}, User{}, fmt.Errorf("auth: load session user: %w", err)
	}
	return sess, user, nil
}

// Logout revokes the session identified by the raw cookie token.
func (s *Service) Logout(ctx context.Context, rawToken string) error {
	if strings.TrimSpace(rawToken) == "" {
		return nil
	}
	err := s.sessions.RevokeSessionByTokenHash(ctx, hashSessionToken(rawToken))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
}

// ListSessions returns non-revoked, non-expired sessions for userID.
// currentSessionID marks the caller's active session when present.
// Token hashes are never exposed.
func (s *Service) ListSessions(ctx context.Context, userID, currentSessionID string) ([]SessionInfo, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, nil
	}
	sessions, err := s.sessions.ListActiveSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("auth: list sessions: %w", err)
	}
	out := make([]SessionInfo, 0, len(sessions))
	for _, sess := range sessions {
		out = append(out, SessionInfo{
			ID:         sess.ID,
			UserAgent:  sess.UserAgent,
			IPAddress:  sess.IPAddress,
			LastSeenAt: sess.LastSeenAt,
			CreatedAt:  sess.CreatedAt,
			Current:    currentSessionID != "" && sess.ID == currentSessionID,
		})
	}
	return out, nil
}

// RevokeSession revokes a session owned by userID. Other users' sessions are not found.
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(sessionID) == "" {
		return ErrNotFound
	}
	if err := s.sessions.RevokeSessionByIDForUser(ctx, sessionID, userID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("auth: revoke session: %w", err)
	}
	return nil
}

func generateSessionToken() (string, error) {
	buf := make([]byte, sessionTokenLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("auth: generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashSessionToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

// SessionCookieName returns the cookie name for the given Secure flag.
func SessionCookieName(secure bool) string {
	if secure {
		return cookieNameSecure
	}
	return cookieNameDev
}

// SetSessionCookie writes the HttpOnly session cookie.
func SetSessionCookie(w http.ResponseWriter, rawToken string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName(secure),
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

// ClearSessionCookie expires the session cookie.
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName(secure),
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// ClientIP extracts a best-effort client IP from the request.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}
