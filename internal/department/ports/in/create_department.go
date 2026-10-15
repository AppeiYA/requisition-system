package departmentin

import (
	"context"
	"time"
)

// CreateDepartmentCommand carries the admin-provided details of a new
// department.
type CreateDepartmentCommand struct {
	Name        string
	Description string
	Code        string
}

type CreateDepartmentResult struct {
	ID          string
	Name        string
	Description string
	Code        string
	Status      string
	CreatedAt   time.Time
}

type CreateDepartmentUsecaseInt interface {
	Execute(ctx context.Context, cmd CreateDepartmentCommand) (*CreateDepartmentResult, error)
}
