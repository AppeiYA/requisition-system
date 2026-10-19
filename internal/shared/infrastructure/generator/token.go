package generator

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// refreshTokenBytes is the entropy of a generated refresh token. 32 bytes
// (256 bits) makes the token infeasible to guess.
const refreshTokenBytes = 32

// TokenGenerator produces opaque refresh tokens and their storage hashes. It
// intentionally implements the identity RefreshTokenGenerator port structurally
// so it can be wired in without importing the identity package.
type TokenGenerator struct{}

// Generate returns a cryptographically secure, URL-safe opaque token.
func (g *TokenGenerator) Generate() (string, error) {
	bytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// Hash returns the hex-encoded SHA-256 hash of the raw token. Only this hash is
// ever persisted; the raw token is returned to the client once.
func (g *TokenGenerator) Hash(token string) string {
	sum := sha256.Sum256([]byte(token))

	return hex.EncodeToString(sum[:])
}
