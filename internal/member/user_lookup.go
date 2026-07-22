package member

import (
	"context"
	"errors"

	"github.com/kevin-voss/htmx-go-postgresql/internal/auth"
)

// AuthUserLookup adapts auth.UserStore to UserEmailLookup.
type AuthUserLookup struct {
	Users auth.UserStore
}

// GetUserIDByEmail returns the user id for email, or ErrNotFound.
func (l AuthUserLookup) GetUserIDByEmail(ctx context.Context, email string) (string, error) {
	if l.Users == nil {
		return "", ErrNotFound
	}
	u, err := l.Users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, auth.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return u.ID, nil
}
