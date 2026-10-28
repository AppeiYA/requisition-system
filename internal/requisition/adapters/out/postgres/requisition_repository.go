package requisitionpostgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/AppeiYA/requisition-system/internal/requisition/adapters/out/postgres/models"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionout "github.com/AppeiYA/requisition-system/internal/requisition/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var ErrRequisitionNotFound = customerrors.NewNotFoundError("requisition not found")

type RequisitionRepository struct {
	repository db.Repository
}

func NewRequisitionRepository(repository db.Repository) *RequisitionRepository {
	return &RequisitionRepository{repository: repository}
}

var saveRequisitionQuery = `
	INSERT INTO requisitions (
		id, requested_by, department_id, item_id, category_id, requested_item_name,
		requested_item_description, quantity, estimated_cost, justification,
		priority, needed_by, status, created_at, updated_at
	) VALUES (
		:id, :requested_by, :department_id, :item_id, :category_id, :requested_item_name,
		:requested_item_description, :quantity, :estimated_cost, :justification,
		:priority, :needed_by, :status, :created_at, :updated_at
	)
`

func (r *RequisitionRepository) SaveRequisition(ctx context.Context, requisition *requisitiondomain.Requisition) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, saveRequisitionQuery, models.RequisitionFromDomain(requisition))
	return err
}

var getRequisitionByIDQuery = `
	SELECT id, requested_by, department_id, item_id, category_id, requested_item_name,
		requested_item_description, quantity, estimated_cost, justification,
		priority, needed_by, status, created_at, updated_at
	FROM requisitions WHERE id = $1
`

func (r *RequisitionRepository) GetRequisitionByID(ctx context.Context, id string) (*requisitiondomain.Requisition, error) {
	var model models.RequisitionModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, getRequisitionByIDQuery, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRequisitionNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

func (r *RequisitionRepository) UpdateRequisition(ctx context.Context, requisition *requisitiondomain.Requisition) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, `
		UPDATE requisitions SET quantity = :quantity, estimated_cost = :estimated_cost,
			justification = :justification, priority = :priority, needed_by = :needed_by,
			status = :status, updated_at = :updated_at WHERE id = :id
	`, models.RequisitionFromDomain(requisition))
	return err
}

var listRequisitionsByUserIDQuery = `
	SELECT id, requested_by, department_id, item_id, category_id, requested_item_name,
		requested_item_description, quantity, estimated_cost, justification,
		priority, needed_by, status, created_at, updated_at
	FROM requisitions
	WHERE requested_by = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
`

func (r *RequisitionRepository) ListRequisitionsByUserID(ctx context.Context, userID string, limit int, offset int) ([]*requisitiondomain.Requisition, error) {
	var modelsList []models.RequisitionModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &modelsList, listRequisitionsByUserIDQuery, userID, limit, offset); err != nil {
		return nil, err
	}

	requisitions := make([]*requisitiondomain.Requisition, 0, len(modelsList))
	for _, model := range modelsList {
		requisition, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		requisitions = append(requisitions, requisition)
	}
	return requisitions, nil
}

var listRequisitionsQuery = `
	SELECT id, requested_by, department_id, item_id, category_id, requested_item_name,
		requested_item_description, quantity, estimated_cost, justification,
		priority, needed_by, status, created_at, updated_at
	FROM requisitions
	ORDER BY created_at DESC
	LIMIT $1 OFFSET $2
`

func (r *RequisitionRepository) ListRequisitions(ctx context.Context, limit int, offset int) ([]*requisitiondomain.Requisition, error) {
	var modelsList []models.RequisitionModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &modelsList, listRequisitionsQuery, limit, offset); err != nil {
		return nil, err
	}

	requisitions := make([]*requisitiondomain.Requisition, 0, len(modelsList))
	for _, model := range modelsList {
		requisition, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		requisitions = append(requisitions, requisition)
	}
	return requisitions, nil
}

var _ requisitionout.RequisitionRepository = (*RequisitionRepository)(nil)
