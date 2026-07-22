package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateRegisterOK(t *testing.T) {
	t.Parallel()

	errs := ValidateRegister(RegisterInput{
		DisplayName:          "Ada",
		Email:                "ada@example.com",
		Password:             "correct-horse-battery",
		PasswordConfirmation: "correct-horse-battery",
		AcceptTerms:          true,
	})
	if errs.Any() {
		t.Fatalf("unexpected errors: %+v", errs)
	}
}

func TestValidateRegisterFieldErrors(t *testing.T) {
	t.Parallel()

	errs := ValidateRegister(RegisterInput{
		DisplayName:          "A",
		Email:                "not-an-email",
		Password:             "short",
		PasswordConfirmation: "different",
		AcceptTerms:          false,
	})

	if errs.DisplayName == "" {
		t.Fatal("want display name error")
	}
	if errs.Email == "" {
		t.Fatal("want email error")
	}
	if errs.Password == "" {
		t.Fatal("want password error")
	}
	if errs.PasswordConfirmation == "" {
		t.Fatal("want password confirmation error")
	}
	if errs.Terms == "" {
		t.Fatal("want terms error")
	}
}

func TestValidateRegisterDisplayNameBounds(t *testing.T) {
	t.Parallel()

	base := RegisterInput{
		Email:                "user@example.com",
		Password:             "long-enough-password",
		PasswordConfirmation: "long-enough-password",
		AcceptTerms:          true,
	}

	tooShort := base
	tooShort.DisplayName = "A"
	if errs := ValidateRegister(tooShort); errs.DisplayName == "" {
		t.Fatal("want error for display name length 1")
	}

	tooLong := base
	tooLong.DisplayName = strings.Repeat("x", 51)
	if errs := ValidateRegister(tooLong); errs.DisplayName == "" {
		t.Fatal("want error for display name length 51")
	}

	ok := base
	ok.DisplayName = strings.Repeat("x", 50)
	if errs := ValidateRegister(ok); errs.Any() {
		t.Fatalf("50-char display name should pass: %+v", errs)
	}
}

type stubUserStore struct {
	byEmail map[string]User
	create  func(ctx context.Context, email, displayName, passwordHash string) (User, error)
}

func (s *stubUserStore) Create(ctx context.Context, email, displayName, passwordHash string) (User, error) {
	if s.create != nil {
		return s.create(ctx, email, displayName, passwordHash)
	}
	u := User{ID: "u1", Email: email, DisplayName: displayName, PasswordHash: passwordHash}
	if s.byEmail == nil {
		s.byEmail = map[string]User{}
	}
	s.byEmail[email] = u
	return u, nil
}

func (s *stubUserStore) GetByEmail(_ context.Context, email string) (User, error) {
	if u, ok := s.byEmail[email]; ok {
		return u, nil
	}
	return User{}, ErrNotFound
}

type stubSessionStore struct {
	byHash map[string]Session
	create func(ctx context.Context, userID, tokenHash string, expiresAt time.Time, userAgent, ipAddress string) (Session, error)
}

func (s *stubSessionStore) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time, userAgent, ipAddress string) (Session, error) {
	if s.create != nil {
		return s.create(ctx, userID, tokenHash, expiresAt, userAgent, ipAddress)
	}
	sess := Session{
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
		s.byHash = map[string]Session{}
	}
	s.byHash[tokenHash] = sess
	return sess, nil
}

func (s *stubSessionStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (Session, error) {
	if sess, ok := s.byHash[tokenHash]; ok {
		return sess, nil
	}
	return Session{}, ErrNotFound
}

func (s *stubSessionStore) RevokeSessionByTokenHash(_ context.Context, tokenHash string) error {
	sess, ok := s.byHash[tokenHash]
	if !ok {
		return ErrNotFound
	}
	now := time.Now().UTC()
	sess.RevokedAt = &now
	s.byHash[tokenHash] = sess
	return nil
}

func (s *stubSessionStore) TouchSession(_ context.Context, id string, at time.Time) error {
	for hash, sess := range s.byHash {
		if sess.ID == id {
			if sess.RevokedAt != nil {
				return ErrNotFound
			}
			sess.LastSeenAt = at
			s.byHash[hash] = sess
			return nil
		}
	}
	return ErrNotFound
}

func TestServiceRegisterNormalizesEmailAndHashesPassword(t *testing.T) {
	t.Parallel()

	store := &stubUserStore{}
	svc := NewService(store, &stubSessionStore{})

	user, errs, err := svc.Register(context.Background(), RegisterInput{
		DisplayName:          "  Ada  ",
		Email:                "  Ada@Example.COM ",
		Password:             "correct-horse-battery",
		PasswordConfirmation: "correct-horse-battery",
		AcceptTerms:          true,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if errs.Any() {
		t.Fatalf("unexpected field errors: %+v", errs)
	}
	if user.Email != "ada@example.com" {
		t.Fatalf("email = %q, want normalized lowercase", user.Email)
	}
	if user.DisplayName != "Ada" {
		t.Fatalf("display name = %q, want trimmed Ada", user.DisplayName)
	}
	if user.PasswordHash == "" || user.PasswordHash == "correct-horse-battery" {
		t.Fatal("password was not hashed")
	}
	ok, err := Compare("correct-horse-battery", user.PasswordHash)
	if err != nil || !ok {
		t.Fatalf("hashed password did not verify: ok=%v err=%v", ok, err)
	}
}

func TestServiceRegisterDuplicateEmail(t *testing.T) {
	t.Parallel()

	store := &stubUserStore{
		byEmail: map[string]User{
			"ada@example.com": {ID: "existing", Email: "ada@example.com"},
		},
	}
	svc := NewService(store, &stubSessionStore{})

	_, errs, err := svc.Register(context.Background(), RegisterInput{
		DisplayName:          "Ada",
		Email:                "ADA@example.com",
		Password:             "correct-horse-battery",
		PasswordConfirmation: "correct-horse-battery",
		AcceptTerms:          true,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if errs.Email == "" {
		t.Fatal("want duplicate email field error")
	}
}

func TestServiceRegisterCreateRaceDuplicate(t *testing.T) {
	t.Parallel()

	store := &stubUserStore{
		create: func(context.Context, string, string, string) (User, error) {
			return User{}, ErrDuplicateEmail
		},
	}
	svc := NewService(store, &stubSessionStore{})

	_, errs, err := svc.Register(context.Background(), RegisterInput{
		DisplayName:          "Ada",
		Email:                "ada@example.com",
		Password:             "correct-horse-battery",
		PasswordConfirmation: "correct-horse-battery",
		AcceptTerms:          true,
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if errs.Email == "" {
		t.Fatal("want duplicate email field error from create race")
	}
}

func TestServiceRegisterPropagatesStoreErrors(t *testing.T) {
	t.Parallel()

	boom := errors.New("db down")
	store := &stubUserStore{
		create: func(context.Context, string, string, string) (User, error) {
			return User{}, boom
		},
	}
	svc := NewService(store, &stubSessionStore{})

	_, _, err := svc.Register(context.Background(), RegisterInput{
		DisplayName:          "Ada",
		Email:                "ada@example.com",
		Password:             "correct-horse-battery",
		PasswordConfirmation: "correct-horse-battery",
		AcceptTerms:          true,
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want %v", err, boom)
	}
}
