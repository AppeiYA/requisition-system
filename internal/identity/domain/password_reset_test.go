package domain

import (
	"errors"
	"testing"
	"time"
)

func Test_NewPasswordResetToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.November, 25, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        string
		userID    string
		hash      string
		ttl       time.Duration
		expected  error
		wantError bool
	}{
		{name: "valid", id: "prt-1", userID: "usr-1", hash: "abc123", ttl: 30 * time.Minute},
		{name: "missing id", userID: "usr-1", hash: "abc123", ttl: time.Minute, expected: ErrPasswordResetTokenIDRequired, wantError: true},
		{name: "missing user", id: "prt-1", hash: "abc123", ttl: time.Minute, expected: ErrPasswordResetTokenUserRequired, wantError: true},
		{name: "missing hash", id: "prt-1", userID: "usr-1", ttl: time.Minute, expected: ErrPasswordResetTokenHashRequired, wantError: true},
		{name: "zero ttl", id: "prt-1", userID: "usr-1", hash: "abc123", wantError: true, expected: ErrPasswordResetTokenTTLInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewPasswordResetToken(tt.id, tt.userID, tt.hash, now, tt.ttl)
			if tt.wantError {
				if !errors.Is(err, tt.expected) {
					t.Fatalf("error = %v, want %v", err, tt.expected)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token.UserID() != tt.userID || token.TokenHash() != tt.hash {
				t.Fatalf("unexpected token: %+v", token)
			}
			if !token.ExpiresAt().Equal(now.Add(tt.ttl)) {
				t.Fatalf("expiresAt = %v, want %v", token.ExpiresAt(), now.Add(tt.ttl))
			}
			if !token.CreatedAt().Equal(now) {
				t.Fatalf("createdAt = %v, want %v", token.CreatedAt(), now)
			}
		})
	}
}

func Test_PasswordResetToken_State(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.November, 25, 12, 0, 0, 0, time.UTC)
	token, err := NewPasswordResetToken("prt-1", "usr-1", "hash", now, 30*time.Minute)
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	if token.IsUsed() || token.IsExpired(now) || !token.IsActive(now) {
		t.Fatalf("new token should be active at its creation time")
	}
	if !token.IsExpired(now.Add(30 * time.Minute)) {
		t.Fatalf("token should be expired exactly at expiresAt")
	}
	if !token.IsActive(now.Add(30*time.Minute - time.Nanosecond)) {
		t.Fatalf("token should still be active just before expiresAt")
	}

	usedAt := now.Add(2 * time.Minute)
	if err := token.MarkUsed(usedAt); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	if !token.IsUsed() || token.IsActive(now.Add(time.Minute)) {
		t.Fatalf("used token must not be active")
	}
	if token.UsedAt() == nil || !token.UsedAt().Equal(usedAt) {
		t.Fatalf("usedAt = %v, want %v", token.UsedAt(), usedAt)
	}
	if err := token.MarkUsed(now.Add(3 * time.Minute)); !errors.Is(err, ErrPasswordResetTokenAlreadyUsed) {
		t.Fatalf("second mark used error = %v, want %v", err, ErrPasswordResetTokenAlreadyUsed)
	}
}
