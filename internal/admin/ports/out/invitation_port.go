package adminout

import (
	"context"
	"time"

	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

// SavedInvitation is the result of persisting an invitation. The token is the
// raw, single-use token that must be delivered to the invitee (only its hash is
// stored).
type SavedInvitation struct {
	ID        string
	Token     string
	ExpiresAt time.Time
}

type AdminInvitationPort interface {
	CheckActiveInvitationExists(ctx context.Context, email *shareddomain.Email) (bool, error)
	SaveInvitation(ctx context.Context, email *shareddomain.Email, inviterID string) (*SavedInvitation, error)
}
