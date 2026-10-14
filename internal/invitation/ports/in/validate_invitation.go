package invitationin

import (
	"context"
)

type ValidateInvitationResult struct {
	Valid bool
	Email string
}

type ValidateInvitationUseCaseInt interface {
	Execute(ctx context.Context, token string) (*ValidateInvitationResult, error)
}
