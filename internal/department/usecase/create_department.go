package usecase

import (
	"context"

	"github.com/AppeiYA/requisition-system/internal/department/domain"
	departmentin "github.com/AppeiYA/requisition-system/internal/department/ports/in"
	departmentout "github.com/AppeiYA/requisition-system/internal/department/ports/out"
)

type CreateDepartmentUsecase struct {
	departmentRepo departmentout.DepartmentRepository
	idGenerator    departmentout.IDGenerator
}

func NewCreateDepartmentUsecase(
	departmentRepo departmentout.DepartmentRepository,
	idGenerator departmentout.IDGenerator,
) *CreateDepartmentUsecase {
	return &CreateDepartmentUsecase{
		departmentRepo: departmentRepo,
		idGenerator:    idGenerator,
	}
}

func (uc *CreateDepartmentUsecase) Execute(ctx context.Context, cmd departmentin.CreateDepartmentCommand) (*departmentin.CreateDepartmentResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	name, err := domain.NewDepartmentName(cmd.Name)
	if err != nil {
		return nil, err
	}

	description, err := domain.NewDepartmentDescription(cmd.Description)
	if err != nil {
		return nil, err
	}

	code, err := domain.NewDepartmentCode(cmd.Code)
	if err != nil {
		return nil, err
	}

	departmentID, err := uc.idGenerator.GenerateID(domain.DepartmentIDPrefix)
	if err != nil {
		return nil, err
	}

	department, err := domain.NewDepartment(departmentID, name, description, code)
	if err != nil {
		return nil, err
	}

	if err := uc.departmentRepo.SaveDepartment(ctx, department); err != nil {
		return nil, err
	}

	return &departmentin.CreateDepartmentResult{
		ID:          department.ID(),
		Name:        department.Name().Value(),
		Description: department.Description().Value(),
		Code:        department.Code().Value(),
		Status:      department.Status().String(),
		CreatedAt:   department.CreatedAt(),
	}, nil
}
