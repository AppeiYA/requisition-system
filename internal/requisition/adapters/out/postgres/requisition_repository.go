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
	"github.com/lib/pq"
)

var ErrRequisitionNotFound = customerrors.NewNotFoundError("requisition not found")

type RequisitionRepository struct {
	repository db.Repository
}

func NewRequisitionRepository(repository db.Repository) *RequisitionRepository {
	return &RequisitionRepository{repository: repository}
}

const requisitionColumns = `
	id, reference_code, requested_by, department_id, justification,
	priority, needed_by, status, created_at, updated_at
`

var saveRequisitionQuery = `
	INSERT INTO requisitions (
		id, reference_code, requested_by, department_id, justification,
		priority, needed_by, status, created_at, updated_at
	) VALUES (
		:id, :reference_code, :requested_by, :department_id, :justification,
		:priority, :needed_by, :status, :created_at, :updated_at
	)
`

func (r *RequisitionRepository) SaveRequisition(ctx context.Context, requisition *requisitiondomain.Requisition) error {
	executor := r.repository.Executor(ctx)
	if _, err := executor.NamedExecContext(ctx, saveRequisitionQuery, models.RequisitionFromDomain(requisition)); err != nil {
		return err
	}

	for _, requestedItem := range requisition.RequestedItems() {
		if _, err := executor.NamedExecContext(ctx, `
			INSERT INTO requested_items (
				id, name, category_id, description, created_by,
				review_status, promoted_item_id, created_at
			) VALUES (
				:id, :name, :category_id, :description, :created_by,
				:review_status, :promoted_item_id, :created_at
			)
		`, models.RequestedItemFromDomain(requestedItem)); err != nil {
			return err
		}
	}

	for _, item := range requisition.Items() {
		if item.RequestedItemID() != "" && !hasRequestedItem(requisition.RequestedItems(), item.RequestedItemID()) {
			return customerrors.NewBadRequestError("requested item is missing from requisition aggregate")
		}
		model, err := models.RequisitionItemFromDomain(item)
		if err != nil {
			return err
		}
		if _, err := executor.NamedExecContext(ctx, `
			INSERT INTO requisition_items (
				id, requisition_id, line_number, item_id, requested_item_id,
				item_name_snapshot, quantity, estimated_unit_amount, notes, created_at
			) VALUES (
				:id, :requisition_id, :line_number, :item_id, :requested_item_id,
				:item_name_snapshot, :quantity, :estimated_unit_amount, :notes, :created_at
			)
		`, model); err != nil {
			return err
		}
	}

	for _, history := range requisition.StatusHistory() {
		if _, err := executor.NamedExecContext(ctx, `
			INSERT INTO requisition_status_history (
				id, requisition_id, sequence, from_status, to_status,
				acted_by, actor_role, reason, created_at
			) VALUES (
				:id, :requisition_id, :sequence, :from_status, :to_status,
				:acted_by, :actor_role, :reason, :created_at
			)
		`, models.StatusHistoryFromDomain(history)); err != nil {
			return err
		}
	}
	return nil
}

func hasRequestedItem(items []*requisitiondomain.RequestedItem, id string) bool {
	for _, item := range items {
		if item != nil && item.ID() == id {
			return true
		}
	}
	return false
}

func (r *RequisitionRepository) GetRequisitionByID(ctx context.Context, id string) (*requisitiondomain.Requisition, error) {
	var model models.RequisitionModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model,
		`SELECT `+requisitionColumns+` FROM requisitions WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRequisitionNotFound
		}
		return nil, err
	}
	return r.loadAggregate(ctx, model)
}

func (r *RequisitionRepository) UpdateRequisition(ctx context.Context, requisition *requisitiondomain.Requisition) error {
	_, err := r.repository.Executor(ctx).NamedExecContext(ctx, `
		UPDATE requisitions
		SET justification = :justification, priority = :priority,
			needed_by = :needed_by, updated_at = :updated_at
		WHERE id = :id
	`, models.RequisitionFromDomain(requisition))
	return err
}

func (r *RequisitionRepository) UpdateRequisitionAggregate(ctx context.Context, requisition *requisitiondomain.Requisition) error {
	executor := r.repository.Executor(ctx)
	if _, err := executor.NamedExecContext(ctx, `
		UPDATE requisitions
		SET department_id = :department_id, justification = :justification,
			priority = :priority, needed_by = :needed_by, updated_at = :updated_at
		WHERE id = :id
	`, models.RequisitionFromDomain(requisition)); err != nil {
		return err
	}
	if _, err := executor.ExecContext(ctx,
		`DELETE FROM requisition_items WHERE requisition_id = $1`, requisition.ID(),
	); err != nil {
		return err
	}
	for _, requestedItem := range requisition.RequestedItems() {
		if _, err := executor.NamedExecContext(ctx, `
			INSERT INTO requested_items (
				id, name, category_id, description, created_by,
				review_status, promoted_item_id, created_at
			) VALUES (
				:id, :name, :category_id, :description, :created_by,
				:review_status, :promoted_item_id, :created_at
			)
		`, models.RequestedItemFromDomain(requestedItem)); err != nil {
			return err
		}
	}
	for _, item := range requisition.Items() {
		if item.RequestedItemID() != "" && !hasRequestedItem(requisition.RequestedItems(), item.RequestedItemID()) {
			return customerrors.NewBadRequestError("requested item is missing from requisition aggregate")
		}
		model, err := models.RequisitionItemFromDomain(item)
		if err != nil {
			return err
		}
		if _, err := executor.NamedExecContext(ctx, `
			INSERT INTO requisition_items (
				id, requisition_id, line_number, item_id, requested_item_id,
				item_name_snapshot, quantity, estimated_unit_amount, notes, created_at
			) VALUES (
				:id, :requisition_id, :line_number, :item_id, :requested_item_id,
				:item_name_snapshot, :quantity, :estimated_unit_amount, :notes, :created_at
			)
		`, model); err != nil {
			return err
		}
	}
	return nil
}

// GetRequisitionByIDForUpdate loads the aggregate while holding a row lock on
// the requisition header. Concurrent transitions therefore serialize before
// they calculate the next history sequence.
func (r *RequisitionRepository) GetRequisitionByIDForUpdate(ctx context.Context, id string) (*requisitiondomain.Requisition, error) {
	var model models.RequisitionModel
	if err := r.repository.Executor(ctx).GetContext(ctx, &model,
		`SELECT `+requisitionColumns+` FROM requisitions WHERE id = $1 FOR UPDATE`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRequisitionNotFound
		}
		return nil, err
	}
	return r.loadAggregate(ctx, model)
}

// UpdateRequisitionStatus writes the header status and the new history entry
// through the same executor. Callers run it inside a transaction so both rows
// commit or roll back together, and the append-only history triggers remain
// the final consistency check.
func (r *RequisitionRepository) UpdateRequisitionStatus(
	ctx context.Context,
	requisition *requisitiondomain.Requisition,
	history *requisitiondomain.RequisitionStatusHistory,
) error {
	if history == nil {
		return customerrors.NewBadRequestError("status history is required")
	}
	executor := r.repository.Executor(ctx)

	result, err := executor.NamedExecContext(ctx, `
		UPDATE requisitions
		SET status = :status, updated_at = :updated_at
		WHERE id = :id
	`, models.RequisitionFromDomain(requisition))
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrRequisitionNotFound
	}

	if _, err := executor.NamedExecContext(ctx, `
		INSERT INTO requisition_status_history (
			id, requisition_id, sequence, from_status, to_status,
			acted_by, actor_role, reason, created_at
		) VALUES (
			:id, :requisition_id, :sequence, :from_status, :to_status,
			:acted_by, :actor_role, :reason, :created_at
		)
	`, models.StatusHistoryFromDomain(history)); err != nil {
		return mapStatusHistoryWriteError(err)
	}
	return nil
}

func mapStatusHistoryWriteError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return customerrors.NewConflictError("requisition status changed concurrently; retry the transition")
	}
	return err
}

func (r *RequisitionRepository) ListRequisitionsByUserID(ctx context.Context, userID string, limit int, offset int) ([]*requisitiondomain.Requisition, error) {
	var rows []models.RequisitionModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &rows,
		`SELECT `+requisitionColumns+` FROM requisitions
		 WHERE requested_by = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset); err != nil {
		return nil, err
	}
	return r.loadAggregates(ctx, rows)
}

func (r *RequisitionRepository) ListRequisitions(ctx context.Context, limit int, offset int) ([]*requisitiondomain.Requisition, error) {
	var rows []models.RequisitionModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &rows,
		`SELECT `+requisitionColumns+` FROM requisitions
		 ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset); err != nil {
		return nil, err
	}
	return r.loadAggregates(ctx, rows)
}

func (r *RequisitionRepository) loadAggregates(ctx context.Context, rows []models.RequisitionModel) ([]*requisitiondomain.Requisition, error) {
	result := make([]*requisitiondomain.Requisition, 0, len(rows))
	for _, row := range rows {
		requisition, err := r.loadAggregate(ctx, row)
		if err != nil {
			return nil, err
		}
		result = append(result, requisition)
	}
	return result, nil
}

func (r *RequisitionRepository) loadAggregate(ctx context.Context, model models.RequisitionModel) (*requisitiondomain.Requisition, error) {
	executor := r.repository.Executor(ctx)
	var itemModels []models.RequisitionItemModel
	if err := executor.SelectContext(ctx, &itemModels, `
		SELECT id, requisition_id, line_number, item_id, requested_item_id,
			item_name_snapshot, quantity, estimated_unit_amount, notes, created_at
		FROM requisition_items WHERE requisition_id = $1 ORDER BY line_number
	`, model.ID); err != nil {
		return nil, err
	}
	items := make([]*requisitiondomain.RequisitionItem, 0, len(itemModels))
	for _, itemModel := range itemModels {
		item, err := itemModel.ToDomain()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	var historyModels []models.RequisitionStatusHistoryModel
	if err := executor.SelectContext(ctx, &historyModels, `
		SELECT id, requisition_id, sequence, from_status, to_status,
			acted_by, actor_role, reason, created_at
		FROM requisition_status_history WHERE requisition_id = $1 ORDER BY sequence
	`, model.ID); err != nil {
		return nil, err
	}
	history := make([]*requisitiondomain.RequisitionStatusHistory, 0, len(historyModels))
	for _, historyModel := range historyModels {
		entry, err := historyModel.ToDomain()
		if err != nil {
			return nil, err
		}
		history = append(history, entry)
	}
	var requestedItemModels []models.RequestedItemModel
	if err := executor.SelectContext(ctx, &requestedItemModels, `
		SELECT DISTINCT ri.requested_item_id AS id, req.name, req.category_id,
			req.description, req.created_by, req.review_status,
			req.promoted_item_id, req.created_at
		FROM requisition_items ri
		JOIN requested_items req ON req.id = ri.requested_item_id
		WHERE ri.requisition_id = $1 AND ri.requested_item_id IS NOT NULL
	`, model.ID); err != nil {
		return nil, err
	}
	requestedItems := make([]*requisitiondomain.RequestedItem, 0, len(requestedItemModels))
	for _, requestedItemModel := range requestedItemModels {
		requestedItem, err := requestedItemModel.ToDomain()
		if err != nil {
			return nil, err
		}
		requestedItems = append(requestedItems, requestedItem)
	}

	return requisitiondomain.ReconstituteRequisitionAggregateWithRequestedItems(
		model.ID, model.ReferenceCode, model.RequestedBy, model.DepartmentID,
		model.Priority, model.NeededBy, model.Justification, model.Status,
		model.CreatedAt, model.UpdatedAt, items, requestedItems, history,
	)
}

var _ requisitionout.RequisitionRepository = (*RequisitionRepository)(nil)
var _ requisitionout.RequisitionTransitionRepository = (*RequisitionRepository)(nil)
