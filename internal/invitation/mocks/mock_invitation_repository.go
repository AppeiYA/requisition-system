package invitationmocks

import (
	"context"
	"time"

	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationout "github.com/AppeiYA/requisition-system/internal/invitation/ports/out"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
)

var (
	DefaultInvitationID = "inv-1234"
	DefaultInviterID    = "usr-1234"
	DefaultInviteeEmail = "invitee@example.com"
	DefaultFirstName    = "Jane"
	DefaultLastName     = "Doe"
	DefaultPhone        = "+14155551234"
	DefaultStaffID      = "STF-1234"
	DefaultDepartmentID = "dpt-1234"
	DefaultJobTitle     = "Software Engineer"
	DefaultToken        = "raw-invitation-token"
)

// ValidInvitation returns a pending invitation whose token hash matches
// DefaultToken, so tests can look it up via domain.HashInvitationToken.
func ValidInvitation() *invitationdomain.Invitation {
	email, _ := shareddomain.NewEmail(DefaultInviteeEmail)
	firstName, _ := shareddomain.NewUsername(DefaultFirstName, shareddomain.FirstName)
	lastName, _ := shareddomain.NewUsername(DefaultLastName, shareddomain.LastName)
	phone, _ := shareddomain.NewPhoneNumber(DefaultPhone)
	invitation, err := invitationdomain.NewInvitation(
		DefaultInvitationID,
		firstName,
		lastName,
		phone,
		DefaultStaffID,
		DefaultDepartmentID,
		DefaultJobTitle,
		email,
		invitationdomain.HashInvitationToken(DefaultToken),
		DefaultInviterID,
	)
	if err != nil {
		panic(err)
	}
	return invitation
}

type MockInvitationRepository struct {
	SaveInvitationFunc               func(ctx context.Context, invitation *invitationdomain.Invitation) error
	GetInvitationByTokenHashFunc     func(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error)
	LockInvitationByTokenHashFunc    func(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error)
	GetActiveInvitationByEmailFunc   func(ctx context.Context, email string) (*invitationdomain.Invitation, error)
	MarkInvitationAcceptedFunc       func(ctx context.Context, invitationID string) error
	RevokeInvitationFunc             func(ctx context.Context, invitationID string) error
	MarkInvitationDeliveredFunc      func(ctx context.Context, invitationID string) error
	MarkInvitationDeliveryFailedFunc func(ctx context.Context, invitationID string) error
	GetInvitationByIDFunc            func(ctx context.Context, invitationID string) (*invitationdomain.Invitation, error)
	GetInvitationByEmailFunc         func(ctx context.Context, email string) (*invitationdomain.Invitation, error)
	UpdateInvitationStatusFunc       func(ctx context.Context, invitationID string, status invitationdomain.InvitationStatus) error
	ExpireDueInvitationsFunc         func(ctx context.Context, now time.Time) error
}

func NewMockInvitationRepository() *MockInvitationRepository {
	return &MockInvitationRepository{
		SaveInvitationFunc: func(ctx context.Context, invitation *invitationdomain.Invitation) error {
			return nil
		},
		GetInvitationByTokenHashFunc: func(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error) {
			return nil, invitationdomain.ErrInvitationNotFound
		},
		GetActiveInvitationByEmailFunc: func(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
			return nil, invitationdomain.ErrInvitationNotFound
		},
		MarkInvitationAcceptedFunc: func(ctx context.Context, invitationID string) error {
			return nil
		},
		RevokeInvitationFunc: func(ctx context.Context, invitationID string) error {
			return nil
		},
		MarkInvitationDeliveredFunc: func(ctx context.Context, invitationID string) error {
			return nil
		},
		MarkInvitationDeliveryFailedFunc: func(ctx context.Context, invitationID string) error {
			return nil
		},
		GetInvitationByIDFunc: func(ctx context.Context, invitationID string) (*invitationdomain.Invitation, error) {
			return nil, invitationdomain.ErrInvitationNotFound
		},
		GetInvitationByEmailFunc: func(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
			return nil, invitationdomain.ErrInvitationNotFound
		},
		UpdateInvitationStatusFunc: func(ctx context.Context, invitationID string, status invitationdomain.InvitationStatus) error {
			return nil
		},
		ExpireDueInvitationsFunc: func(ctx context.Context, now time.Time) error {
			return nil
		},
	}
}

func (m *MockInvitationRepository) SaveInvitation(ctx context.Context, invitation *invitationdomain.Invitation) error {
	return m.SaveInvitationFunc(ctx, invitation)
}

func (m *MockInvitationRepository) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error) {
	return m.GetInvitationByTokenHashFunc(ctx, tokenHash)
}

func (m *MockInvitationRepository) LockInvitationByTokenHash(ctx context.Context, tokenHash string) (*invitationdomain.Invitation, error) {
	if m.LockInvitationByTokenHashFunc == nil {
		return m.GetInvitationByTokenHashFunc(ctx, tokenHash)
	}
	return m.LockInvitationByTokenHashFunc(ctx, tokenHash)
}

func (m *MockInvitationRepository) GetActiveInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
	return m.GetActiveInvitationByEmailFunc(ctx, email)
}

func (m *MockInvitationRepository) MarkInvitationAccepted(ctx context.Context, invitationID string) error {
	return m.MarkInvitationAcceptedFunc(ctx, invitationID)
}

func (m *MockInvitationRepository) RevokeInvitation(ctx context.Context, invitationID string) error {
	return m.RevokeInvitationFunc(ctx, invitationID)
}

func (m *MockInvitationRepository) MarkInvitationDelivered(ctx context.Context, invitationID string) error {
	return m.MarkInvitationDeliveredFunc(ctx, invitationID)
}

func (m *MockInvitationRepository) MarkInvitationDeliveryFailed(ctx context.Context, invitationID string) error {
	return m.MarkInvitationDeliveryFailedFunc(ctx, invitationID)
}

func (m *MockInvitationRepository) GetInvitationByID(ctx context.Context, invitationID string) (*invitationdomain.Invitation, error) {
	return m.GetInvitationByIDFunc(ctx, invitationID)
}

func (m *MockInvitationRepository) GetInvitationByEmail(ctx context.Context, email string) (*invitationdomain.Invitation, error) {
	return m.GetInvitationByEmailFunc(ctx, email)
}

func (m *MockInvitationRepository) UpdateInvitationStatus(ctx context.Context, invitationID string, status invitationdomain.InvitationStatus) error {
	return m.UpdateInvitationStatusFunc(ctx, invitationID, status)
}

func (m *MockInvitationRepository) ExpireDueInvitations(ctx context.Context, now time.Time) error {
	return m.ExpireDueInvitationsFunc(ctx, now)
}

var _ invitationout.InvitationRepository = (*MockInvitationRepository)(nil)
