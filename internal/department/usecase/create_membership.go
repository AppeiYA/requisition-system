package usecase

import (
	"context"
	"errors"
	"net/http"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrUserAlreadyHasDepartment = customerrors.NewConflictError("user already belongs to a department")

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

	existing, err := uc.membershipRepo.GetMembershipByUserID(ctx, cmd.UserID)
	if err != nil {
		var customErr *customerrors.CustomError
		if !errors.As(err, &customErr) || customErr.Code != http.StatusNotFound {
			return err
		}
	} else if existing != nil {
		return ErrUserAlreadyHasDepartment
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
