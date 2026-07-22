package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	displayNameMin = 2
	displayNameMax = 50
	passwordMin    = 12
)

// UserStore is the persistence port used by Service.
type UserStore interface {
	Create(ctx context.Context, email, displayName, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
}

// RegisterInput is the public registration form payload.
type RegisterInput struct {
	DisplayName          string
	Email                string
	Password             string
	PasswordConfirmation string
	AcceptTerms          bool
}

// RegisterErrors holds per-field validation messages for the register form.
type RegisterErrors struct {
	DisplayName          string
	Email                string
	Password             string
	PasswordConfirmation string
	Terms                string
}

// Any reports whether any field error is set.
func (e RegisterErrors) Any() bool {
	return e.DisplayName != "" ||
		e.Email != "" ||
		e.Password != "" ||
		e.PasswordConfirmation != "" ||
		e.Terms != ""
}

// Service implements auth business rules (registration, login, sessions).
type Service struct {
	users    UserStore
	sessions SessionStore
}

// NewService constructs an auth service.
func NewService(users UserStore, sessions SessionStore) *Service {
	return &Service{users: users, sessions: sessions}
}

// Register validates input, hashes the password, and persists a new user.
// On validation or duplicate-email failure it returns field errors and a zero User.
func (s *Service) Register(ctx context.Context, in RegisterInput) (User, RegisterErrors, error) {
	normalized := normalizeRegisterInput(in)
	fieldErrs := ValidateRegister(normalized)
	if fieldErrs.Any() {
		return User{}, fieldErrs, nil
	}

	_, err := s.users.GetByEmail(ctx, normalized.Email)
	if err == nil {
		return User{}, RegisterErrors{Email: "Email is already registered."}, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return User{}, RegisterErrors{}, fmt.Errorf("auth: lookup email: %w", err)
	}

	hash, err := Hash(normalized.Password)
	if err != nil {
		return User{}, RegisterErrors{}, fmt.Errorf("auth: hash password: %w", err)
	}

	user, err := s.users.Create(ctx, normalized.Email, normalized.DisplayName, hash)
	if err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			return User{}, RegisterErrors{Email: "Email is already registered."}, nil
		}
		return User{}, RegisterErrors{}, err
	}
	return user, RegisterErrors{}, nil
}

// ValidateRegister applies registration field rules (no uniqueness check).
func ValidateRegister(in RegisterInput) RegisterErrors {
	var errs RegisterErrors

	nameLen := utf8.RuneCountInString(in.DisplayName)
	if nameLen < displayNameMin || nameLen > displayNameMax {
		errs.DisplayName = "Display name must be between 2 and 50 characters."
	}

	if in.Email == "" || !validEmail(in.Email) {
		errs.Email = "Enter a valid email address."
	}

	if utf8.RuneCountInString(in.Password) < passwordMin {
		errs.Password = "Password must be at least 12 characters."
	}

	if in.PasswordConfirmation != in.Password {
		errs.PasswordConfirmation = "Passwords do not match."
	}

	if !in.AcceptTerms {
		errs.Terms = "You must accept the terms."
	}

	return errs
}

func normalizeRegisterInput(in RegisterInput) RegisterInput {
	return RegisterInput{
		DisplayName:          strings.TrimSpace(in.DisplayName),
		Email:                strings.ToLower(strings.TrimSpace(in.Email)),
		Password:             in.Password,
		PasswordConfirmation: in.PasswordConfirmation,
		AcceptTerms:          in.AcceptTerms,
	}
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	// Reject "Name <email>" forms; registration expects a bare address.
	return addr.Address == email && strings.Contains(email, "@")
}
