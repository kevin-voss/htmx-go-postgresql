package auth

import (
	"sync"
	"time"
)

const (
	loginRateLimitMax     = 5
	loginRateLimitWindow  = 15 * time.Minute
)

// LoginRateLimiter is a simple in-memory sliding-window limiter for login attempts.
type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
	now      func() time.Time
}

// NewLoginRateLimiter constructs a limiter with the default login thresholds.
func NewLoginRateLimiter() *LoginRateLimiter {
	return &LoginRateLimiter{
		attempts: map[string][]time.Time{},
		limit:    loginRateLimitMax,
		window:   loginRateLimitWindow,
		now:      time.Now,
	}
}

// Allow records an attempt for key and reports whether it is within the limit.
// When false, the attempt is still recorded (so repeated abuse stays blocked).
func (l *LoginRateLimiter) Allow(key string) bool {
	if key == "" {
		key = "unknown"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now().UTC()
	cutoff := now.Add(-l.window)
	old := l.attempts[key]
	recent := make([]time.Time, 0, len(old)+1)
	for _, t := range old {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	allowed := len(recent) < l.limit
	recent = append(recent, now)
	l.attempts[key] = recent
	return allowed
}

// Reset clears recorded attempts for key (e.g. after a successful login).
func (l *LoginRateLimiter) Reset(key string) {
	if key == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, key)
}

// LoginRateLimitMax returns the max attempts per window (for handoff/docs).
func LoginRateLimitMax() int {
	return loginRateLimitMax
}

// LoginRateLimitWindow returns the sliding window duration (for handoff/docs).
func LoginRateLimitWindow() time.Duration {
	return loginRateLimitWindow
}
