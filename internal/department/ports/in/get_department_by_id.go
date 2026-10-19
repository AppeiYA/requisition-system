package departmentin

import (
	"context"
	"time"
)

// DepartmentResult is the read model returned when looking up a department.
type DepartmentResult struct {
	ID          string
	Name        string
	Description string
	Code        string
	Status      string
	CreatedAt   time.Time
}

type GetDepartmentByIDUsecaseInt interface {
	Execute(ctx context.Context, id string) (*DepartmentResult, error)
}
