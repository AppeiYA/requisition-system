package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
)

// CreateMembershipUsecase links a freshly registered user to the department
// they were invited to. It runs inside the invitation-acceptance transaction,
// so a failure rolls back both the user and the membership.
type CreateMembershipUsecase struct {
	membershipRepo departmentout.MembershipRepository
}

func NewCreateMembershipUsecase(membershipRepo departmentout.MembershipRepository) *CreateMembershipUsecase {
	return &CreateMembershipUsecase{membershipRepo: membershipRepo}
}

func (uc *CreateMembershipUsecase) Execute(ctx context.Context, cmd departmentin.CreateMembershipCommand) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	jobTitle, err := domain.NewMemberJobTitle(cmd.JobTitle)
	if err != nil {
		return err
	}

	membership, err := domain.NewDepartmentMembership(cmd.UserID, cmd.DepartmentID, jobTitle)
	if err != nil {
		return err
	}

	return uc.membershipRepo.SaveMembership(ctx, membership)
}
