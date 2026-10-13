package jwtadapter

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/config"
)

const (
	testSubject  = "usr-1234"
	testIssuer   = "test-issuer"
	testAudience = "test-audience"
)

func testConfig() config.JWTConfig {
	return config.JWTConfig{
		Secret:         strings.Repeat("a", 32),
		Issuer:         testIssuer,
		Audience:       testAudience,
		AccessTokenTTL: time.Minute,
		Leeway:         0,
	}
}

func TestGenerateAccessToken_Success(t *testing.T) {
	t.Parallel()

	svc := NewJWTService(testConfig())

	access, err := svc.GenerateAccessToken(context.Background(), &identityout.TokenModel{ID: testSubject})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if access.Value == "" {
		t.Fatal("expected non-empty token")
	}
	if access.ExpiresIn != int64(time.Minute.Seconds()) {
		t.Errorf("ExpiresIn = %d, want %d", access.ExpiresIn, int64(time.Minute.Seconds()))
	}

	model, err := svc.ValidateAccessToken(context.Background(), access.Value)
	if err != nil {
		t.Fatalf("token should validate: %v", err)
	}
	if model.ID != testSubject {
		t.Errorf("subject = %q, want %q", model.ID, testSubject)
	}
}

func TestGenerateAccessToken_InvalidModel(t *testing.T) {
	t.Parallel()

	svc := NewJWTService(testConfig())

	cases := map[string]*identityout.TokenModel{
		"nil":      nil,
		"empty id": {},
	}

	for name, model := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := svc.GenerateAccessToken(context.Background(), model); !errors.Is(err, ErrInvalidTokenModel) {
				t.Fatalf("expected ErrInvalidTokenModel, got: %v", err)
			}
		})
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	t.Parallel()

	cfg := testConfig()
	cfg.AccessTokenTTL = -time.Minute
	svc := NewJWTService(cfg)

	access, err := svc.GenerateAccessToken(context.Background(), &identityout.TokenModel{ID: testSubject})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := svc.ValidateAccessToken(context.Background(), access.Value); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected ErrExpiredToken, got: %v", err)
	}
}

func TestValidateAccessToken_Malformed(t *testing.T) {
	t.Parallel()

	svc := NewJWTService(testConfig())

	if _, err := svc.ValidateAccessToken(context.Background(), "not-a-jwt"); !errors.Is(err, ErrMalformedToken) {
		t.Fatalf("expected ErrMalformedToken, got: %v", err)
	}
}

func TestValidateAccessToken_WrongSignature(t *testing.T) {
	t.Parallel()

	issuer := NewJWTService(testConfig())
	access, err := issuer.GenerateAccessToken(context.Background(), &identityout.TokenModel{ID: testSubject})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := testConfig()
	cfg.Secret = strings.Repeat("c", 32)
	other := NewJWTService(cfg)

	if _, err := other.ValidateAccessToken(context.Background(), access.Value); !errors.Is(err, ErrInvalidTokenSignature) {
		t.Fatalf("expected ErrInvalidTokenSignature, got: %v", err)
	}
}

func TestValidateAccessToken_WrongIssuer(t *testing.T) {
	t.Parallel()

	issuer := NewJWTService(testConfig())
	access, err := issuer.GenerateAccessToken(context.Background(), &identityout.TokenModel{ID: testSubject})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := testConfig()
	cfg.Issuer = "someone-else"
	other := NewJWTService(cfg)

	if _, err := other.ValidateAccessToken(context.Background(), access.Value); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestValidateAccessToken_WrongAudience(t *testing.T) {
	t.Parallel()

	issuer := NewJWTService(testConfig())
	access, err := issuer.GenerateAccessToken(context.Background(), &identityout.TokenModel{ID: testSubject})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := testConfig()
	cfg.Audience = "someone-else"
	other := NewJWTService(cfg)

	if _, err := other.ValidateAccessToken(context.Background(), access.Value); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got: %v", err)
	}
}
