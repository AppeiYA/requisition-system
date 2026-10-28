package departmentpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres/models"
	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	sharedcache "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/cache"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/lib/pq"
)

var (
	ErrDepartmentNotFoundRepo   = customerrors.NewNotFoundError("department not found")
	ErrDepartmentCodeExistsRepo = customerrors.NewConflictError("department with this code already exists")
)

const pqUniqueViolationCode = "23505"
const departmentCacheTTL = 10 * time.Minute

type DepartmentRepository struct {
	repository db.Repository
	cache      sharedcache.Cache
}

func NewDepartmentRepository(repository db.Repository, cache sharedcache.Cache) *DepartmentRepository {
	return &DepartmentRepository{repository: repository, cache: cache}
}

var saveDepartmentQuery = `
	INSERT INTO departments (id, name, description, code, status, created_at, updated_at)
	VALUES (:id, :name, :description, :code, :status, :created_at, :updated_at)
`

func (r *DepartmentRepository) SaveDepartment(ctx context.Context, department *departmentdomain.Department) error {
	key := "department:all"
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveDepartmentQuery, models.DepartmentFromDomain(department))
	if err != nil {
		// Translate the unique-code index violation into an application error.
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == pqUniqueViolationCode {
			return ErrDepartmentCodeExistsRepo
		}
		return err
	}

	err = r.cache.Delete(ctx, key)
	if err != nil {
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
	key := fmt.Sprintf("department:%s", id)
	var model models.DepartmentModel

	found, err := r.cache.Get(ctx, key, &model)
	if err != nil {
		return nil, err
	}

	if found {
		return model.ToDomain()
	}

	err = r.repository.Executor(ctx).GetContext(ctx, &model, getByIDQuery, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDepartmentNotFoundRepo
		}
		return nil, err
	}

	err = r.cache.Set(ctx, key, model, departmentCacheTTL)
	if err != nil {
		return nil, err
	}

	return model.ToDomain()
}

var getDepartmentsQuery = `
	SELECT id, name, description, code, status, created_at, updated_at
	FROM departments
	ORDER BY name ASC
	LIMIT 1000
`

func (r *DepartmentRepository) GetDepartments(ctx context.Context) ([]*departmentdomain.Department, error) {
	key := "department:all"
	var modelsList []models.DepartmentModel

	found, err := r.cache.Get(ctx, key, &modelsList)
	if err != nil {
		return nil, err
	}
	if found {
		departments := make([]*departmentdomain.Department, 0, len(modelsList))
		for _, model := range modelsList {
			department, err := model.ToDomain()
			if err != nil {
				return nil, err
			}
			departments = append(departments, department)
		}
		return departments, nil
	}

	err = r.repository.Executor(ctx).SelectContext(ctx, &modelsList, getDepartmentsQuery)
	if err != nil {
		return nil, err
	}

	departments := make([]*departmentdomain.Department, 0, len(modelsList))

	for _, model := range modelsList {
		department, err := model.ToDomain()
		if err != nil {
			return nil, err
		}

		departments = append(departments, department)
	}

	err = r.cache.Set(ctx, key, modelsList, departmentCacheTTL)
	if err != nil {
		return nil, err
	}

	return departments, nil
}
