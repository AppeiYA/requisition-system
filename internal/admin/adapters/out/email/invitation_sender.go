package adminemail

import (
	"context"
	"fmt"
	"net/url"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/email"
)

type InvitationSender struct {
	sender *email.Sender
	cfg    config.FrontendConfig
}

func NewInvitationSender(sender *email.Sender, cfg config.FrontendConfig) *InvitationSender {
	return &InvitationSender{sender: sender, cfg: cfg}
}

func (s *InvitationSender) SendInvitation(ctx context.Context, to string, invitedBy string, token string, expiresAt string) error {
	base := s.cfg.BaseURL
	if base == "" {
		base = "http://localhost:3000"
	}
	u, err := url.Parse(base)
	if err != nil {
		u = &url.URL{Scheme: "http", Host: "localhost:3000"}
	}
	u.Path = "/register/invitation"
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	subject := "You're invited to join Requisition System"
	body := fmt.Sprintf(`
		<p>You have been invited to join the Requisition System.</p>
		<p>Invited by: %s</p>
		<p>Complete your registration here: <a href="%s">%s</a></p>
		<p>This invitation expires after 24 hours (expires at: %s).</p>
		<p>This link can only be used once.</p>
		<p>If you were not expecting this invitation, please ignore this email.</p>
	`, invitedBy, u.String(), u.String(), expiresAt)

	return s.sender.Send(ctx, to, subject, body)
}
