package identityemail

import (
	"context"
	"fmt"
	"net/url"
	"time"

	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/email"
)

// PasswordResetSender builds the reset link and delivers the email through
// the same Resend sender the invitation flow uses.
type PasswordResetSender struct {
	sender   *email.Sender
	frontend config.FrontendConfig
	tokenTTL time.Duration
}

func NewPasswordResetSender(sender *email.Sender, frontend config.FrontendConfig, cfg config.PasswordResetConfig) *PasswordResetSender {
	ttl := cfg.TokenTTL
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	return &PasswordResetSender{sender: sender, frontend: frontend, tokenTTL: ttl}
}

var _ identityout.PasswordResetEmailSender = (*PasswordResetSender)(nil)

func (s *PasswordResetSender) SendPasswordResetEmail(ctx context.Context, to, token string, expiresAt time.Time) error {
	base := s.frontend.BaseURL
	if base == "" {
		base = "http://localhost:3000"
	}
	resetPath := s.frontend.PasswordResetPath
	if resetPath == "" {
		resetPath = "/reset-password"
	}

	link, err := url.JoinPath(base, resetPath)
	if err != nil {
		return fmt.Errorf("build password reset link: %w", err)
	}
	u, err := url.Parse(link)
	if err != nil {
		return fmt.Errorf("parse password reset link: %w", err)
	}
	query := u.Query()
	query.Set("token", token)
	u.RawQuery = query.Encode()

	subject := "Reset your Requisition System password"
	body := fmt.Sprintf(`
		<p>A password reset was requested for your Requisition System account.</p>
		<p>Reset your password here: <a href="%s">%s</a></p>
		<p>This link expires in %s (at %s) and can only be used once.</p>
		<p>If you requested more than one reset, the most recent email is the valid one.</p>
		<p>If you did not request a password reset, you can ignore this email.</p>
	`, u.String(), u.String(), s.tokenTTL, expiresAt.UTC().Format(time.RFC3339))

	return s.sender.Send(ctx, to, subject, body)
}
