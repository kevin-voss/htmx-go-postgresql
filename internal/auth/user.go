package auth

import "time"

// User is a persisted account row.
type User struct {
	ID              string
	Email           string
	DisplayName     string
	PasswordHash    string
	EmailVerifiedAt *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
