package invitationin

import "context"

type ExpireInvitationsUsecaseInt interface {
	Execute(ctx context.Context) error
}
