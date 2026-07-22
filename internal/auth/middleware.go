package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/kevin-voss/htmx-go-postgresql/internal/platform/middleware"
)

type contextKey int

const (
	userContextKey contextKey = iota
	sessionContextKey
)

// UserFromContext returns the authenticated user loaded by LoadSession, if any.
func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userContextKey).(User)
	return u, ok
}

// SessionFromContext returns the active session loaded by LoadSession, if any.
func SessionFromContext(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionContextKey).(Session)
	return s, ok
}

// ContextWithUser returns a child context carrying user (tests / handlers).
func ContextWithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// LoadSession returns middleware that reads the session cookie, validates it,
// and stores the session + user on the request context when present.
// Missing or invalid sessions are treated as anonymous (no error response).
func LoadSession(service *Service, cookieSecure bool, logger *slog.Logger) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie(SessionCookieName(cookieSecure))
			if err != nil || c.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			sess, user, err := service.LoadSessionUser(r.Context(), c.Value)
			if err != nil {
				if !errors.Is(err, ErrInvalidSession) && !errors.Is(err, ErrNotFound) {
					if logger != nil {
						logger.Error("load session failed", "err", err)
					}
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), sessionContextKey, sess)
			ctx = context.WithValue(ctx, userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuthentication blocks unauthenticated requests.
// Full-page browser navigations redirect to /login (302 Found).
// HTMX requests receive 401 Unauthorized so clients can handle re-auth.
func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := UserFromContext(r.Context()); !ok {
			if isHTMXRequest(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isHTMXRequest(r *http.Request) bool {
	if r.Header.Get("HX-Request") != "" {
		return true
	}
	return r.Header.Get("HX-Request-Type") != ""
}
