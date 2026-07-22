package auth

import (
	"testing"
	"time"
)

func TestLoginRateLimiterAllowsThenBlocks(t *testing.T) {
	t.Parallel()

	lim := NewLoginRateLimiter()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	lim.now = func() time.Time { return now }

	for i := 0; i < LoginRateLimitMax(); i++ {
		if !lim.Allow("1.2.3.4") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if lim.Allow("1.2.3.4") {
		t.Fatal("attempt beyond limit should be blocked")
	}
	if !lim.Allow("9.9.9.9") {
		t.Fatal("different key should still be allowed")
	}
}

func TestLoginRateLimiterWindowExpiry(t *testing.T) {
	t.Parallel()

	lim := NewLoginRateLimiter()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	lim.now = func() time.Time { return now }

	for i := 0; i < LoginRateLimitMax(); i++ {
		if !lim.Allow("ip") {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
	}
	if lim.Allow("ip") {
		t.Fatal("should be blocked inside window")
	}

	now = now.Add(LoginRateLimitWindow() + time.Second)
	if !lim.Allow("ip") {
		t.Fatal("should allow after window expires")
	}
}

func TestLoginRateLimiterReset(t *testing.T) {
	t.Parallel()

	lim := NewLoginRateLimiter()
	for i := 0; i < LoginRateLimitMax(); i++ {
		_ = lim.Allow("ip")
	}
	if lim.Allow("ip") {
		t.Fatal("should be blocked before reset")
	}
	lim.Reset("ip")
	if !lim.Allow("ip") {
		t.Fatal("should allow after reset")
	}
}
