package domain

import "time"

// RequisitionOwner identifies the user who created a requisition.
type RequisitionOwner struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
}

// RequisitionSummary is the lightweight read model used by list endpoints. It
// carries only the requisition header values and the derived line total, so
// listing does not need to reconstruct the full aggregate.
type RequisitionSummary struct {
	ID                   string
	ReferenceCode        string
	RequestedBy          string
	Owner                RequisitionOwner
	DepartmentID         string
	Status               string
	Priority             string
	NeededBy             time.Time
	Justification        string
	TotalEstimatedAmount float64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
