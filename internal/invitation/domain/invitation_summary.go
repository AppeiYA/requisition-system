package domain

import "time"

// InvitationListFilter is the validated set of filters for the admin invitation
// list. From and To are nil when the caller did not supply a date boundary.
type InvitationListFilter struct {
	Search       string
	Status       string
	DepartmentID string
	InvitedBy    string
	From         *time.Time
	To           *time.Time
	Limit        int
	Offset       int
}

// InvitationSummary is the admin read model for an invitation. It deliberately
// omits the token hash: only the raw token handed to the invitee may travel,
// and even that is never stored here.
type InvitationSummary struct {
	ID               string
	FirstName        string
	LastName         string
	Email            string
	DepartmentID     string
	DepartmentName   string
	JobTitle         string
	Status           string
	InvitedBy        string
	InviterName      string
	ExpiresAt        time.Time
	AcceptedAt       *time.Time
	RevokedAt        *time.Time
	DeliveredAt      *time.Time
	DeliveryFailedAt *time.Time
	CreatedAt        time.Time
}
