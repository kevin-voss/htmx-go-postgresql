package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters (OWASP-oriented defaults). Stored in the encoded hash
// so Compare can verify historical hashes and NeedsRehash can detect upgrades.
const (
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // KiB (64 MiB)
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

const argon2idPrefix = "argon2id"

// ErrInvalidHash is returned when an encoded password hash cannot be parsed.
var ErrInvalidHash = errors.New("auth: invalid password hash")

// Hash returns an Argon2id-encoded hash of password with a fresh random salt.
// Format: $argon2id$v=19$m=<mem>,t=<time>,p=<threads>$<salt_b64>$<hash_b64>
func Hash(password string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("auth: generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return encodeHash(argon2Time, argon2Memory, argon2Threads, salt, key), nil
}

// Compare reports whether password matches the Argon2id-encoded hash.
// Comparison of derived keys uses constant-time equality.
func Compare(password, encoded string) (bool, error) {
	time, memory, threads, salt, want, err := decodeHash(encoded)
	if err != nil {
		return false, err
	}

	got := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(want)))
	if subtle.ConstantTimeCompare(got, want) == 1 {
		return true, nil
	}
	return false, nil
}

// NeedsRehash reports whether encoded was produced with parameters that differ
// from the current defaults (so callers can upgrade hashes after a successful login).
func NeedsRehash(encoded string) (bool, error) {
	time, memory, threads, _, key, err := decodeHash(encoded)
	if err != nil {
		return false, err
	}
	if time != argon2Time || memory != argon2Memory || threads != argon2Threads || len(key) != argon2KeyLen {
		return true, nil
	}
	return false, nil
}

func encodeHash(time, memory uint32, threads uint8, salt, key []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf(
		"$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2idPrefix,
		argon2.Version,
		memory,
		time,
		threads,
		b64Salt,
		b64Key,
	)
}

func decodeHash(encoded string) (time, memory uint32, threads uint8, salt, key []byte, err error) {
	parts := strings.Split(encoded, "$")
	// "", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash
	if len(parts) != 6 || parts[1] != argon2idPrefix {
		return 0, 0, 0, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return 0, 0, 0, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return 0, 0, 0, nil, nil, ErrInvalidHash
	}

	var t, m uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return 0, 0, 0, nil, nil, ErrInvalidHash
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return 0, 0, 0, nil, nil, ErrInvalidHash
	}
	key, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return 0, 0, 0, nil, nil, ErrInvalidHash
	}
	if len(salt) == 0 || len(key) == 0 {
		return 0, 0, 0, nil, nil, ErrInvalidHash
	}

	return t, m, p, salt, key, nil
}
