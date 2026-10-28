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

type MembershipRepository struct {
	repository db.Repository
}

var (
	ErrMembershipNotFound = customerrors.NewNotFoundError("department membership not found")
	ErrMembershipExists   = customerrors.NewConflictError("user already belongs to a department")
)

func NewMembershipRepository(repository db.Repository) *MembershipRepository {
	return &MembershipRepository{repository: repository}
}

var saveMembershipQuery = `
	INSERT INTO department_memberships (user_id, department_id, job_title, status, created_at)
	VALUES (:user_id, :department_id, :job_title, :status, :created_at)
`

// SaveMembership participates in the ambient transaction when one is present
// on the context (invitation acceptance), so the membership is only committed
// together with the user it belongs to.
func (r *MembershipRepository) SaveMembership(ctx context.Context, membership *departmentdomain.DepartmentMembership) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveMembershipQuery, models.MembershipFromDomain(membership))
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && string(pqErr.Code) == "23505" {
			return ErrMembershipExists
		}
		return err
	}
	return nil
}

var getMembershipByUserIDQuery = `
	SELECT user_id, department_id, job_title, status, created_at
	FROM department_memberships
	WHERE user_id = $1
`

func (r *MembershipRepository) GetMembershipByUserID(ctx context.Context, userID string) (*departmentdomain.DepartmentMembership, error) {
	var model models.MembershipModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, getMembershipByUserIDQuery, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrMembershipNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}
