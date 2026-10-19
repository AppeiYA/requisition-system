package adminout

import "context"

type EmailSender interface {
	SendInvitation(ctx context.Context, to string, invitedBy string, token string, expiresAt string) error
}
