package departmentin

import "context"

// CreateMembershipCommand links a registered user to a department with an
// immutable job title. It is executed when an invitation is accepted, i.e.
// once the user exists.
type CreateMembershipCommand struct {
	UserID       string
	DepartmentID string
	JobTitle     string
}

type CreateMembershipUsecaseInt interface {
	Execute(ctx context.Context, cmd CreateMembershipCommand) error
}
