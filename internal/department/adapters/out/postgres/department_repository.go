package departmentpostgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres/models"
	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

var (
	ErrDepartmentNotFoundRepo   = customerrors.NewNotFoundError("department not found")
	ErrDepartmentCodeExistsRepo = customerrors.NewConflictError("department with this code already exists")
)

const pqUniqueViolationCode = "23505"

type DepartmentRepository struct {
	repository db.Repository
}

func NewDepartmentRepository(repository db.Repository) *DepartmentRepository {
	return &DepartmentRepository{repository: repository}
}

var saveDepartmentQuery = `
	INSERT INTO departments (id, name, description, code, status, created_at, updated_at)
	VALUES (:id, :name, :description, :code, :status, :created_at, :updated_at)
`

func (r *DepartmentRepository) SaveDepartment(ctx context.Context, department *departmentdomain.Department) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveDepartmentQuery, models.DepartmentFromDomain(department))
	if err != nil {
		// Translate the unique-code index violation into an application error.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == pqUniqueViolationCode {
			return ErrDepartmentCodeExistsRepo
		}
		return err
	}
	return nil
}

var getByIDQuery = `
	SELECT id, name, description, code, status, created_at, updated_at
	FROM departments
	WHERE id = $1
`

func (r *DepartmentRepository) GetDepartmentByID(ctx context.Context, id string) (*departmentdomain.Department, error) {
	var model models.DepartmentModel
	err := r.repository.Executor(ctx).GetContext(ctx, &model, getByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDepartmentNotFoundRepo
		}
		return nil, err
	}
	return model.ToDomain()
}
