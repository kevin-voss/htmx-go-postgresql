package member

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	invitationTTL      = 7 * 24 * time.Hour
	invitationTokenLen = 32
)

// InvitationTTL returns the invitation token lifetime.
func InvitationTTL() time.Duration {
	return invitationTTL
}

func generateInvitationToken() (string, error) {
	buf := make([]byte, invitationTokenLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("member: generate invitation token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashInvitationToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}
