package config

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode"
)

var weakSecrets = map[string]struct{}{
	"secret":          {},
	"password":        {},
	"changeme":        {},
	"change-me":       {},
	"your-secret":     {},
	"your-secret-key": {},
	"jwt-secret":      {},
	"jwt-secret-key":  {},
	"secret-key":      {},
	"development":     {},
	"test-secret":     {},
	"admin":           {},
}

type JWTConfig struct {
	Secret          string
	Issuer          string
	Audience        string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration // also the Redis refresh-session TTL
	Leeway          time.Duration
}

// LoadJWTConfig loads the JWT settings outside production, where the secret
// only has to be long enough.
func LoadJWTConfig() (*JWTConfig, error) {
	return loadJWTConfig(false)
}

// ProdLoadJWTConfig loads the JWT settings in production, where the secret
// must additionally be strong Base64-encoded random bytes.
func ProdLoadJWTConfig() (*JWTConfig, error) {
	return loadJWTConfig(true)
}

func loadJWTConfig(production bool) (*JWTConfig, error) {
	secret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	if production {
		if err = validateJWTSecret(secret, 32); err != nil {
			return nil, fmt.Errorf("invalid JWT Configuration: %w", err)
		}
	}

	if len(secret) < minJWTSecretLength {
		return nil, ErrJWTSecretTooShort
	}

	accessTokenTTL, err := getTimeDurationEnv("JWT_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshTokenTTL, err := getTimeDurationEnv("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	if refreshTokenTTL <= accessTokenTTL {
		return nil, ErrJWTRefreshTTLNotLongerThanAccess
	}

	leeway, err := getTimeDurationEnv("JWT_LEEWAY", 30*time.Second)
	if err != nil {
		return nil, err
	}

	return &JWTConfig{
		Secret:          secret,
		Issuer:          getEnv("JWT_ISSUER", "requisition-system"),
		Audience:        getEnv("JWT_AUDIENCE", "requisition-system"),
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
		Leeway:          leeway,
	}, nil
}

func validateJWTSecret(secret string, minBytes int) error {
	if minBytes < 32 {
		return fmt.Errorf("minimum bytes for secret must be at least 32, got %d", minBytes)
	}

	if secret == "" {
		return ErrJWTSecretTooShort
	}

	if strings.TrimSpace(secret) != secret {
		return fmt.Errorf("jwt secret contains leading or trailing whitespace")
	}

	normalized := strings.ToLower(secret)

	if _, exists := weakSecrets[normalized]; exists {
		return fmt.Errorf("jwt secret is too weak, please use a stronger secret")
	}

	var key []byte
	var err error

	encodings := []*base64.Encoding{
		base64.StdEncoding.Strict(),
		base64.RawStdEncoding.Strict(),
		base64.URLEncoding.Strict(),
		base64.RawURLEncoding.Strict(),
	}

	for _, encoding := range encodings {
		key, err = encoding.DecodeString(secret)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf(`JWT_SECRET must be valid Base64-encoded random bytes`)
	}

	if len(key) < minBytes {
		return fmt.Errorf(
			"JWT_SECRET must contain at least %d decoded bytes; got %d",
			minBytes,
			len(key),
		)
	}

	if isWeakKey(key) {
		return fmt.Errorf("JWT_SECRET contains a predictable or weak pattern")
	}

	return nil
}

func isWeakKey(key []byte) bool {
	if len(key) == 0 {
		return true
	}

	// Reject a key consisting entirely of the same byte. AND
	// Reject sequential byte patterns
	allSame := true
	increasing := true
	decreasing := true
	for i := 1; i < len(key); i++ {
		if key[i] != key[0] {
			allSame = false
		}

		if key[i] != key[i-1]+1 {
			increasing = false
		}

		if key[i] != key[i-1]-1 {
			decreasing = false
		}
	}

	if allSame {
		return true
	}

	if increasing || decreasing {
		return true
	}

	// Reject obvious repeated textual patterns.
	printable := true
	for _, b := range key {
		if !unicode.IsPrint(rune(b)) {
			printable = false
			break
		}
	}

	if printable {
		text := strings.ToLower(string(key))

		for weak := range weakSecrets {
			if strings.Contains(text, weak) {
				return true
			}
		}
	}

	return false
}
