package departmentpostgres

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/department/adapters/out/postgres/models"
	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

type MembershipRepository struct {
	repository db.Repository
}

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
		return err
	}
	return nil
}
