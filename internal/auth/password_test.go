package auth

import (
	"strings"
	"testing"
)

func TestHashCompareRoundTrip(t *testing.T) {
	const password = "correct horse battery staple"

	encoded, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if encoded == "" {
		t.Fatal("Hash returned empty string")
	}
	if !strings.HasPrefix(encoded, "$argon2id$") {
		t.Fatalf("encoded hash missing argon2id prefix: %q", encoded)
	}
	// Parameters must be stored with the hash (time / memory / threads).
	if !strings.Contains(encoded, "m=") || !strings.Contains(encoded, "t=") || !strings.Contains(encoded, "p=") {
		t.Fatalf("encoded hash missing parameters: %q", encoded)
	}

	ok, err := Compare(password, encoded)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if !ok {
		t.Fatal("correct password did not verify")
	}
}

func TestCompareWrongPassword(t *testing.T) {
	encoded, err := Hash("correct-password-123")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	ok, err := Compare("wrong-password-456", encoded)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if ok {
		t.Fatal("wrong password verified")
	}
}

func TestHashUsesUniqueSalt(t *testing.T) {
	const password = "same-password-twice"

	a, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash a: %v", err)
	}
	b, err := Hash(password)
	if err != nil {
		t.Fatalf("Hash b: %v", err)
	}
	if a == b {
		t.Fatal("two hashes of the same password were identical (salt not random)")
	}

	okA, err := Compare(password, a)
	if err != nil || !okA {
		t.Fatalf("Compare a: ok=%v err=%v", okA, err)
	}
	okB, err := Compare(password, b)
	if err != nil || !okB {
		t.Fatalf("Compare b: ok=%v err=%v", okB, err)
	}
}

func TestCompareInvalidHash(t *testing.T) {
	ok, err := Compare("password", "not-a-valid-hash")
	if err != ErrInvalidHash {
		t.Fatalf("expected ErrInvalidHash, got ok=%v err=%v", ok, err)
	}
}

func TestNeedsRehash(t *testing.T) {
	encoded, err := Hash("password-for-rehash-check")
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}

	needs, err := NeedsRehash(encoded)
	if err != nil {
		t.Fatalf("NeedsRehash: %v", err)
	}
	if needs {
		t.Fatal("fresh hash unexpectedly needs rehash")
	}

	// Same password material with weaker memory parameter should need rehash.
	weak := strings.Replace(encoded, "m=65536", "m=16384", 1)
	needs, err = NeedsRehash(weak)
	if err != nil {
		t.Fatalf("NeedsRehash weak: %v", err)
	}
	if !needs {
		t.Fatal("weak-parameter hash should need rehash")
	}
}
