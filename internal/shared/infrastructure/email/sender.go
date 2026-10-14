package email

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/config"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrSendFailed = customerrors.NewInternalServerError("failed to send email")
)

type Sender struct {
	cfg config.EmailConfig
}

func NewSender(cfg config.EmailConfig) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Send(ctx context.Context, to, subject, body string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	from := s.cfg.From
	if from == "" {
		from = "noreply@requisition-system.local"
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n", from, to, subject, body)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		return ErrSendFailed
	}
	return nil
}
