package email

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/resendlabs/resend-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	ErrSendFailed = customerrors.NewInternalServerError("failed to send email")
)

// Sender delivers transactional email through the Resend API:
// https://resend.com/docs/api-reference/emails/send-email
type Sender struct {
	emails resend.EmailsSvc
	cfg    config.EmailConfig
}

func NewSender(cfg config.EmailConfig) *Sender {
	client := &http.Client{
		Transport: otelhttp.NewTransport(nil),
	}

	return &Sender{
		emails: resend.NewCustomClient(client, cfg.APIKey).Emails,
		cfg:    cfg,
	}
}

func (s *Sender) Send(ctx context.Context, to, subject, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := s.emails.Send(&resend.SendEmailRequest{
		From:    s.from(),
		To:      []string{to},
		Subject: subject,
		Html:    body,
	})
	if err != nil {
		// Keep the API error in the chain for logging while presenting the
		// generic message to callers.
		return fmt.Errorf("%w: %v", ErrSendFailed, err)
	}

	return nil
}

// from renders the configured sender the way the Resend API expects it,
// "Name <address>". The address must be a verified sender in Resend.
func (s *Sender) from() string {
	if s.cfg.FromName == "" {
		return s.cfg.From
	}
	return fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.From)
}
