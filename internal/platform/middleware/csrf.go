package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

const (
	// CSRFFieldName is the HTML form field carrying the CSRF token.
	CSRFFieldName = "csrf_token"
	// CSRFHeaderName is an alternate header for HTMX/JS clients.
	CSRFHeaderName = "X-CSRF-Token"

	csrfCookieDev    = "forgeboard_csrf"
	csrfCookieSecure = "__Host-forgeboard_csrf"
	csrfTokenBytes   = 32
)

type csrfContextKey struct{}

// CSRF returns middleware that issues a double-submit CSRF cookie and
// validates it on unsafe methods (POST, PUT, PATCH, DELETE).
func CSRF(cookieSecure bool) Middleware {
	cookieName := csrfCookieName(cookieSecure)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := csrfFromRequest(r, cookieName)
			if token == "" {
				var err error
				token, err = newCSRFToken()
				if err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: true,
					Secure:   cookieSecure,
					SameSite: http.SameSiteLaxMode,
				})
			}

			ctx := context.WithValue(r.Context(), csrfContextKey{}, token)
			r = r.WithContext(ctx)

			if isUnsafeMethod(r.Method) {
				submitted := submittedCSRFToken(r)
				if !csrfEqual(token, submitted) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CSRFToken returns the per-request CSRF token from context, or "".
func CSRFToken(ctx context.Context) string {
	token, _ := ctx.Value(csrfContextKey{}).(string)
	return token
}

func csrfCookieName(secure bool) string {
	if secure {
		return csrfCookieSecure
	}
	return csrfCookieDev
}

func csrfFromRequest(r *http.Request, cookieName string) string {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}

func submittedCSRFToken(r *http.Request) string {
	if v := r.Header.Get(CSRFHeaderName); v != "" {
		return v
	}
	// PostFormValue parses the body as needed for form posts.
	return r.PostFormValue(CSRFFieldName)
}

func isUnsafeMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func newCSRFToken() (string, error) {
	buf := make([]byte, csrfTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func csrfEqual(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
