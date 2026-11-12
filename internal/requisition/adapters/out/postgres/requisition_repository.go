package requisitionpostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

type persistedRequisitionLine struct {
	ID              string  `db:"id"`
	LineNumber      int     `db:"line_number"`
	ItemID          *string `db:"item_id"`
	RequestedItemID *string `db:"requested_item_id"`
}

// maxSmallintLineNumber is the largest value PostgreSQL can store in the
// requisition_items.line_number SMALLINT column.
const maxSmallintLineNumber = 32767

func requisitionItemKey(itemID, requestedItemID string) string {
	switch {
	case itemID != "":
		return "catalogue:" + itemID
	case requestedItemID != "":
		return "requested:" + requestedItemID
	default:
		return ""
	}
}

func (l persistedRequisitionLine) key() string {
	itemID, requestedItemID := "", ""
	if l.ItemID != nil {
		itemID = *l.ItemID
	}
	if l.RequestedItemID != nil {
		requestedItemID = *l.RequestedItemID
	}
	return requisitionItemKey(itemID, requestedItemID)
}

// UpdateRequisitionAggregate reconciles the persisted aggregate against the
// supplied one instead of deleting and recreating its children. Existing line
// and requested-item identifiers are preserved, and workflow-managed requested
// item fields are never overwritten.
//
// The caller must invoke it with a context bound to a transaction and must
// already hold the requisition header lock, normally obtained through
// GetRequisitionByIDForUpdate in the same transaction. The reconciliation
// reads the current line snapshot and must not race with another writer.
func (r *RequisitionRepository) UpdateRequisitionAggregate(ctx context.Context, requisition *requisitiondomain.Requisition) error {
	executor := r.repository.Executor(ctx)

	var persistedLines []persistedRequisitionLine
	if err := executor.SelectContext(ctx, &persistedLines, `
		SELECT id, line_number, item_id, requested_item_id
		FROM requisition_items
		WHERE requisition_id = $1
		ORDER BY line_number
	`, requisition.ID()); err != nil {
		return err
	}

	persistedByID := make(map[string]persistedRequisitionLine, len(persistedLines))
	persistedRequestedItemIDs := make(map[string]struct{}, len(persistedLines))
	for _, line := range persistedLines {
		persistedByID[line.ID] = line
		if line.RequestedItemID != nil {
			persistedRequestedItemIDs[*line.RequestedItemID] = struct{}{}
		}
	}

	desiredItems := requisition.Items()
	desiredByID := make(map[string]*requisitiondomain.RequisitionItem, len(desiredItems))
	for _, item := range desiredItems {
		if item == nil {
			return requisitiondomain.ErrNilFields("requisition_item")
		}
		if item.RequisitionID() != requisition.ID() {
			return requisitiondomain.ErrRequisitionItemNotOwned
		}
		if item.LineNumber() < 1 || item.LineNumber() > maxSmallintLineNumber {
			return customerrors.NewBadRequestError(fmt.Sprintf(
				"requisition line number %d is outside the supported range 1..%d",
				item.LineNumber(), maxSmallintLineNumber,
			))
		}
		if item.RequestedItemID() != "" && !hasRequestedItem(requisition.RequestedItems(), item.RequestedItemID()) {
			return customerrors.NewBadRequestError("requested item is missing from requisition aggregate")
		}
		desiredByID[item.ID()] = item
	}

	if _, err := executor.NamedExecContext(ctx, `
		UPDATE requisitions
		SET department_id = :department_id, justification = :justification,
			priority = :priority, needed_by = :needed_by, updated_at = :updated_at
		WHERE id = :id
	`, models.RequisitionFromDomain(requisition)); err != nil {
		return err
	}

	desiredRequestedItemIDs := make(map[string]struct{}, len(requisition.RequestedItems()))
	for _, requestedItem := range requisition.RequestedItems() {
		if requestedItem == nil {
			return requisitiondomain.ErrNilFields("requested_item")
		}
		model := models.RequestedItemFromDomain(requestedItem)
		if _, exists := persistedRequestedItemIDs[requestedItem.ID()]; exists {
			result, err := executor.NamedExecContext(ctx, `
				UPDATE requested_items
				SET name = :name, category_id = :category_id, description = :description
				WHERE id = :id
			`, model)
			if err != nil {
				return err
			}
			affected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return customerrors.NewInternalServerError("requested item disappeared during requisition update")
			}
		} else {
			if _, err := executor.NamedExecContext(ctx, `
				INSERT INTO requested_items (
					id, name, category_id, description, created_by,
					review_status, promoted_item_id, created_at
				) VALUES (
					:id, :name, :category_id, :description, :created_by,
					:review_status, :promoted_item_id, :created_at
				)
			`, model); err != nil {
				return err
			}
		}
		desiredRequestedItemIDs[requestedItem.ID()] = struct{}{}
	}

	// Removed lines are deleted first so their line numbers, catalogue items,
	// and requested items are released before the remaining lines are updated.
	removedLineIDs := make([]string, 0)
	for _, line := range persistedLines {
		if _, keep := desiredByID[line.ID]; !keep {
			removedLineIDs = append(removedLineIDs, line.ID)
		}
	}
	if len(removedLineIDs) > 0 {
		if _, err := executor.ExecContext(ctx, `
			DELETE FROM requisition_items
			WHERE requisition_id = $1 AND id = ANY($2)
		`, requisition.ID(), pq.Array(removedLineIDs)); err != nil {
			return err
		}
	}

	// Line numbers and item keys are unique per requisition and checked per
	// statement. Only final line numbers in the SMALLINT range are ever
	// written; changes are applied in dependency order and a cycle is broken
	// by temporarily parking a line, never by computing an intermediate offset.
	existingDesired := make(map[string]*requisitiondomain.RequisitionItem, len(persistedByID))
	for _, item := range desiredItems {
		if _, exists := persistedByID[item.ID()]; exists {
			existingDesired[item.ID()] = item
		}
	}

	updateLine := func(item *requisitiondomain.RequisitionItem) error {
		model, err := models.RequisitionItemFromDomain(item)
		if err != nil {
			return err
		}
		result, err := executor.NamedExecContext(ctx, `
			UPDATE requisition_items
			SET line_number = :line_number, item_id = :item_id,
				requested_item_id = :requested_item_id,
				item_name_snapshot = :item_name_snapshot, quantity = :quantity,
				estimated_unit_amount = :estimated_unit_amount, notes = :notes
			WHERE id = :id AND requisition_id = :requisition_id
		`, model)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return customerrors.NewInternalServerError("requisition line disappeared during update")
		}
		return nil
	}

	insertLine := func(item *requisitiondomain.RequisitionItem) error {
		model, err := models.RequisitionItemFromDomain(item)
		if err != nil {
			return err
		}
		_, err = executor.NamedExecContext(ctx, `
			INSERT INTO requisition_items (
				id, requisition_id, line_number, item_id, requested_item_id,
				item_name_snapshot, quantity, estimated_unit_amount, notes, created_at
			) VALUES (
				:id, :requisition_id, :line_number, :item_id, :requested_item_id,
				:item_name_snapshot, :quantity, :estimated_unit_amount, :notes, :created_at
			)
		`, model)
		return err
	}

	// The catalogue-item/requested-item unique indexes and the line-number
	// constraint are all checked per statement. Update a line only when both
	// its target key and its target line number are free. A cycle (for example
	// two lines swapping items and line numbers) is broken by temporarily
	// parking one line; it is reinserted with its original identity afterwards.
	keyHolders := make(map[string]string, len(persistedLines))
	lineNumberHolders := make(map[int]string, len(persistedLines))
	for _, line := range persistedLines {
		if _, keep := desiredByID[line.ID]; !keep {
			continue
		}
		if key := line.key(); key != "" {
			keyHolders[key] = line.ID
		}
		lineNumberHolders[line.LineNumber] = line.ID
	}

	pending := make(map[string]*requisitiondomain.RequisitionItem, len(existingDesired))
	for id, item := range existingDesired {
		pending[id] = item
	}
	parked := make([]*requisitiondomain.RequisitionItem, 0)

	for len(pending) > 0 {
		progressed := false
		for id, item := range pending {
			newKey := requisitionItemKey(item.ItemID(), item.RequestedItemID())
			if holder, held := keyHolders[newKey]; held && holder != id {
				if _, blockedByPending := pending[holder]; blockedByPending {
					continue
				}
			}
			newLineNumber := item.LineNumber()
			if holder, held := lineNumberHolders[newLineNumber]; held && holder != id {
				if _, blockedByPending := pending[holder]; blockedByPending {
					continue
				}
			}

			if err := updateLine(item); err != nil {
				return err
			}
			oldKey := persistedByID[id].key()
			if oldKey != "" && oldKey != newKey && keyHolders[oldKey] == id {
				delete(keyHolders, oldKey)
			}
			if newKey != "" {
				keyHolders[newKey] = id
			}
			oldLineNumber := persistedByID[id].LineNumber
			if oldLineNumber != newLineNumber && lineNumberHolders[oldLineNumber] == id {
				delete(lineNumberHolders, oldLineNumber)
			}
			lineNumberHolders[newLineNumber] = id
			delete(pending, id)
			progressed = true
		}
		if progressed {
			continue
		}

		for id, item := range pending {
			oldKey := persistedByID[id].key()
			if oldKey != "" && keyHolders[oldKey] == id {
				delete(keyHolders, oldKey)
			}
			oldLineNumber := persistedByID[id].LineNumber
			if lineNumberHolders[oldLineNumber] == id {
				delete(lineNumberHolders, oldLineNumber)
			}
			if _, err := executor.ExecContext(ctx, `
				DELETE FROM requisition_items
				WHERE id = $1 AND requisition_id = $2
			`, id, requisition.ID()); err != nil {
				return err
			}
			parked = append(parked, item)
			delete(pending, id)
			break
		}
	}

	for _, item := range parked {
		if err := insertLine(item); err != nil {
			return err
		}
	}
	for _, item := range desiredItems {
		if _, exists := persistedByID[item.ID()]; exists {
			continue
		}
		if err := insertLine(item); err != nil {
			return err
		}
	}

	for requestedItemID := range persistedRequestedItemIDs {
		if _, keep := desiredRequestedItemIDs[requestedItemID]; keep {
			continue
		}
		// Unreviewed requested items are safe to remove once their line is
		// gone. Reviewed/promoted rows are retained as workflow history even
		// though they no longer have a requisition line.
		if _, err := executor.ExecContext(ctx, `
			DELETE FROM requested_items
			WHERE id = $1
				AND review_status = 'unreviewed'
				AND promoted_item_id IS NULL
				AND NOT EXISTS (
					SELECT 1 FROM requisition_items
					WHERE requested_item_id = $1
				)
		`, requestedItemID); err != nil {
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

var listRequisitionSummariesQuery = `
	SELECT
		r.id,
		r.reference_code,
		r.requested_by,
		r.department_id,
		r.status,
		r.priority,
		r.needed_by,
		r.justification,
		COALESCE(SUM(ri.estimated_unit_amount * ri.quantity), 0)::bigint AS total_estimated_amount,
		r.created_at,
		r.updated_at
	FROM requisitions r
	LEFT JOIN requisition_items ri
		ON ri.requisition_id = r.id
	GROUP BY
		r.id,
		r.reference_code,
		r.requested_by,
		r.department_id,
		r.status,
		r.priority,
		r.needed_by,
		r.justification,
		r.created_at,
		r.updated_at
	ORDER BY r.created_at DESC
	LIMIT $1 OFFSET $2
`

var listRequisitionSummariesByUserIDQuery = `
	SELECT
		r.id,
		r.reference_code,
		r.requested_by,
		r.department_id,
		r.status,
		r.priority,
		r.needed_by,
		r.justification,
		COALESCE(SUM(ri.estimated_unit_amount * ri.quantity), 0)::bigint AS total_estimated_amount,
		r.created_at,
		r.updated_at
	FROM requisitions r
	LEFT JOIN requisition_items ri
		ON ri.requisition_id = r.id
	WHERE r.requested_by = $1
	GROUP BY
		r.id,
		r.reference_code,
		r.requested_by,
		r.department_id,
		r.status,
		r.priority,
		r.needed_by,
		r.justification,
		r.created_at,
		r.updated_at
	ORDER BY r.created_at DESC
	LIMIT $2 OFFSET $3
`

func (r *RequisitionRepository) ListRequisitionSummariesByUserID(ctx context.Context, userID string, limit int, offset int) ([]*requisitiondomain.RequisitionSummary, error) {
	var rows []models.RequisitionSummaryModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &rows,
		listRequisitionSummariesByUserIDQuery, userID, limit, offset); err != nil {
		return nil, err
	}
	return requisitionSummariesFromModels(rows)
}

func (r *RequisitionRepository) ListRequisitionSummaries(ctx context.Context, limit int, offset int) ([]*requisitiondomain.RequisitionSummary, error) {
	var rows []models.RequisitionSummaryModel
	if err := r.repository.Executor(ctx).SelectContext(ctx, &rows,
		listRequisitionSummariesQuery, limit, offset); err != nil {
		return nil, err
	}
	return requisitionSummariesFromModels(rows)
}

func requisitionSummariesFromModels(rows []models.RequisitionSummaryModel) ([]*requisitiondomain.RequisitionSummary, error) {
	result := make([]*requisitiondomain.RequisitionSummary, 0, len(rows))
	for _, row := range rows {
		summary, err := row.ToDomain()
		if err != nil {
			return nil, err
		}
		result = append(result, summary)
	}
	return result, nil
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
