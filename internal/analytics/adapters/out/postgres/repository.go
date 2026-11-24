package analyticspostgres

import (
	"context"
	"fmt"
	"time"

	analyticsdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
	analyticsout "github.com/AppeiYA/requisition-system/internal/analytics/ports/out"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

type Repository struct {
	repository db.Repository
}

func NewRepository(repository db.Repository) *Repository {
	return &Repository{repository: repository}
}

var _ analyticsout.Repository = (*Repository)(nil)

func majorMinor(value int64) float64 {
	return float64(value) / 100
}

func (r *Repository) GetOverview(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.Overview, error) {
	now := time.Now().UTC()
	result := &analyticsdomain.Overview{
		GeneratedAt:  now,
		Range:        analyticsdomain.Range{From: filter.From, To: filter.To},
		Requisitions: analyticsdomain.OverviewRequisitions{ByStatus: map[string]int{}, Currency: "NGN"},
	}

	var users usersOverviewRow
	err := r.repository.Executor(ctx).GetContext(ctx, &users, `
		SELECT COUNT(*)::int AS total,
		       COUNT(*) FILTER (WHERE status='active')::int AS active,
		       COUNT(*) FILTER (WHERE status='suspended')::int AS suspended,
		       COUNT(*) FILTER (WHERE status='inactive')::int AS inactive,
		       COUNT(*) FILTER (WHERE role='admin')::int AS admins,
		       COUNT(*) FILTER (WHERE created_at >= $1 AND created_at < $2)::int AS new_in_range
		FROM users`, filter.From, filter.To)
	if err != nil {
		return nil, err
	}
	result.Users = analyticsdomain.OverviewUsers{
		Total: users.Total, Active: users.Active, Suspended: users.Suspended,
		Inactive: users.Inactive, Admins: users.Admins, NewInRange: users.NewInRange,
	}

	var invitations invitationsOverviewRow
	dept := filter.DepartmentID
	err = r.repository.Executor(ctx).GetContext(ctx, &invitations, `
		SELECT COUNT(*)::int AS total,
		       COUNT(*) FILTER (WHERE status='pending')::int AS pending,
		       COUNT(*) FILTER (WHERE status='delivered')::int AS delivered,
		       COUNT(*) FILTER (WHERE status='accepted')::int AS accepted,
		       COUNT(*) FILTER (WHERE status='expired')::int AS expired,
		       COUNT(*) FILTER (WHERE status='revoked')::int AS revoked,
		       COUNT(*) FILTER (WHERE status='failed')::int AS failed
		FROM invitations
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3='' OR department_id=$3)`, filter.From, filter.To, dept)
	if err != nil {
		return nil, err
	}
	rate := 0.0
	if invitations.Total > 0 {
		rate = float64(invitations.Accepted) / float64(invitations.Total)
	}
	result.Invitations = analyticsdomain.OverviewInvitations{
		Total: invitations.Total, Pending: invitations.Pending, Delivered: invitations.Delivered,
		Accepted: invitations.Accepted, Expired: invitations.Expired, Revoked: invitations.Revoked,
		Failed: invitations.Failed, AcceptanceRate: rate,
	}

	var requisitionTotals requisitionTotalsRow
	err = r.repository.Executor(ctx).GetContext(ctx, &requisitionTotals, `
		WITH totals AS (
			SELECT r.id, r.needed_by, r.status,
			       COALESCE(SUM(ri.estimated_unit_amount * ri.quantity), 0)::bigint AS minor
			FROM requisitions r
			LEFT JOIN requisition_items ri ON ri.requisition_id=r.id
			WHERE r.created_at >= $1 AND r.created_at < $2
			  AND ($3='' OR r.department_id=$3)
			GROUP BY r.id
		)
		SELECT COUNT(*)::int AS total,
		       COALESCE(SUM(minor),0)::bigint AS minor,
		       COUNT(*) FILTER (WHERE status='pending' AND needed_by < CURRENT_DATE)::int AS overdue
		FROM totals`, filter.From, filter.To, dept)
	if err != nil {
		return nil, err
	}
	result.Requisitions.Total = requisitionTotals.Total
	result.Requisitions.TotalEstimatedAmount = majorMinor(requisitionTotals.Minor)
	if requisitionTotals.Total > 0 {
		result.Requisitions.AverageEstimatedAmount = majorMinor(requisitionTotals.Minor / int64(requisitionTotals.Total))
	}
	result.Requisitions.PendingOverdue = requisitionTotals.Overdue

	var statuses []statusCountRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &statuses, `
		SELECT status, COUNT(*)::int AS count
		FROM requisitions
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3='' OR department_id=$3)
		GROUP BY status`, filter.From, filter.To, dept); err != nil {
		return nil, err
	}
	for _, row := range statuses {
		result.Requisitions.ByStatus[row.Status] = row.Count
	}

	var catalogue catalogueOverviewRow
	err = r.repository.Executor(ctx).GetContext(ctx, &catalogue, `
		SELECT
		  (SELECT COUNT(*) FROM items WHERE status='active')::int AS active_items,
		  (SELECT COUNT(*) FROM item_categories WHERE status='active')::int AS active_categories,
		  (SELECT COUNT(*) FROM requested_items WHERE review_status='unreviewed')::int AS unreviewed`)
	if err != nil {
		return nil, err
	}
	result.Catalogue = analyticsdomain.OverviewCatalogue{
		ActiveItems: catalogue.ActiveItems, ActiveCategories: catalogue.ActiveCategories,
		UnreviewedRequestedItems: catalogue.Unreviewed,
	}
	return result, nil
}

func breakdownNameRows(ctx context.Context, r *Repository, query string, args []any) ([]analyticsdomain.Breakdown, error) {
	var rows []breakdownRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	result := make([]analyticsdomain.Breakdown, len(rows))
	for i, item := range rows {
		result[i] = analyticsdomain.Breakdown{Name: item.Name, Count: item.Count, EstimatedAmount: majorMinor(item.Minor)}
	}
	return result, nil
}

func (r *Repository) GetRequisitionAnalytics(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.RequisitionAnalytics, error) {
	now := time.Now().UTC()
	result := &analyticsdomain.RequisitionAnalytics{
		GeneratedAt: now,
		Range:       analyticsdomain.Range{From: filter.From, To: filter.To},
	}
	args := []any{filter.From, filter.To, filter.DepartmentID}

	status, err := breakdownNameRows(ctx, r, `
		SELECT r.status AS name, COUNT(DISTINCT r.id)::int AS count,
		       COALESCE(SUM(ri.estimated_unit_amount * ri.quantity),0)::bigint AS minor
		FROM requisitions r
		LEFT JOIN requisition_items ri ON ri.requisition_id=r.id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY r.status ORDER BY count DESC`, args)
	if err != nil {
		return nil, err
	}
	result.ByStatus = status

	priority, err := breakdownNameRows(ctx, r, `
		SELECT r.priority AS name, COUNT(DISTINCT r.id)::int AS count,
		       COALESCE(SUM(ri.estimated_unit_amount * ri.quantity),0)::bigint AS minor
		FROM requisitions r
		LEFT JOIN requisition_items ri ON ri.requisition_id=r.id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY r.priority ORDER BY count DESC`, args)
	if err != nil {
		return nil, err
	}
	result.ByPriority = priority

	departments, err := breakdownNameRows(ctx, r, `
		SELECT d.name, COUNT(DISTINCT r.id)::int AS count,
		       COALESCE(SUM(ri.estimated_unit_amount * ri.quantity),0)::bigint AS minor
		FROM requisitions r
		JOIN departments d ON d.id=r.department_id
		LEFT JOIN requisition_items ri ON ri.requisition_id=r.id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY d.name ORDER BY minor DESC`, args)
	if err != nil {
		return nil, err
	}
	result.ByDepartment = departments

	categories, err := breakdownNameRows(ctx, r, `
		SELECT c.name, COUNT(DISTINCT r.id)::int AS count,
		       SUM(ri.estimated_unit_amount * ri.quantity)::bigint AS minor
		FROM requisition_items ri
		JOIN requisitions r ON r.id=ri.requisition_id
		JOIN (
		  SELECT ri2.id, COALESCE(i.category_id, rqi.category_id) AS category_id
		  FROM requisition_items ri2
		  LEFT JOIN items i ON i.id=ri2.item_id
		  LEFT JOIN requested_items rqi ON rqi.id=ri2.requested_item_id
		) src ON src.id=ri.id
		JOIN item_categories c ON c.id=src.category_id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY c.name ORDER BY minor DESC`, args)
	if err != nil {
		return nil, err
	}
	result.ByCategory = categories

	trendQuery := fmt.Sprintf(`
		SELECT date_trunc('%s', r.created_at) AS bucket,
		       COUNT(DISTINCT r.id)::int AS count,
		       COALESCE(SUM(ri.estimated_unit_amount * ri.quantity),0)::bigint AS minor
		FROM requisitions r
		LEFT JOIN requisition_items ri ON ri.requisition_id=r.id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY bucket ORDER BY bucket`, filter.Interval)
	var trendRows []trendMinorRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &trendRows, trendQuery, args...); err != nil {
		return nil, err
	}
	result.Trend = make([]analyticsdomain.TrendBucket, len(trendRows))
	for i, row := range trendRows {
		result.Trend[i] = analyticsdomain.TrendBucket{Bucket: row.Bucket, Count: row.Count, EstimatedAmount: majorMinor(row.Minor)}
	}

	var pendingToApproved, approvedToCompleted, creationToClosed float64
	_ = r.repository.Executor(ctx).GetContext(ctx, &pendingToApproved, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (h.created_at-r.created_at))),0)
		FROM requisitions r
		JOIN LATERAL (
		  SELECT MIN(created_at) AS created_at FROM requisition_status_history
		  WHERE requisition_id=r.id AND to_status='approved'
		) h ON true
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)`, args...)
	_ = r.repository.Executor(ctx).GetContext(ctx, &approvedToCompleted, `
		WITH approved AS (
		  SELECT requisition_id, MIN(created_at) AS approved_at
		  FROM requisition_status_history WHERE to_status='approved'
		  GROUP BY requisition_id
		), completed AS (
		  SELECT requisition_id, MIN(created_at) AS completed_at
		  FROM requisition_status_history WHERE to_status='completed'
		  GROUP BY requisition_id
		)
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (c.completed_at-a.approved_at))),0)
		FROM approved a JOIN completed c ON c.requisition_id=a.requisition_id
		JOIN requisitions r ON r.id=a.requisition_id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)`, args...)
	_ = r.repository.Executor(ctx).GetContext(ctx, &creationToClosed, `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (
		  COALESCE(h.closed_at,r.updated_at)-r.created_at))),0)
		FROM requisitions r
		LEFT JOIN LATERAL (
		  SELECT MIN(created_at) AS closed_at
		  FROM requisition_status_history
		  WHERE requisition_id=r.id AND to_status IN ('rejected','cancelled')
		) h ON true
		WHERE r.status IN ('rejected','cancelled')
		  AND r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)`, args...)
	result.CycleTime = analyticsdomain.CycleTime{
		PendingToApprovedSeconds:   pendingToApproved,
		ApprovedToCompletedSeconds: approvedToCompleted,
		CreationToClosedSeconds:    creationToClosed,
	}

	topItems, err := breakdownNameRows(ctx, r, `
		SELECT ri.item_name_snapshot AS name,
		       SUM(ri.quantity)::int AS count,
		       SUM(ri.estimated_unit_amount * ri.quantity)::bigint AS minor
		FROM requisition_items ri
		JOIN requisitions r ON r.id=ri.requisition_id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY ri.item_name_snapshot
		ORDER BY count DESC LIMIT $4`, []any{filter.From, filter.To, filter.DepartmentID, filter.Limit})
	if err != nil {
		return nil, err
	}
	result.TopItems = topItems

	topRequesters, err := breakdownNameRows(ctx, r, `
		SELECT TRIM(u.first_name || ' ' || u.last_name) AS name,
		       COUNT(DISTINCT r.id)::int AS count,
		       COALESCE(SUM(ri.estimated_unit_amount * ri.quantity),0)::bigint AS minor
		FROM requisitions r
		JOIN users u ON u.id=r.requested_by
		LEFT JOIN requisition_items ri ON ri.requisition_id=r.id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY u.id,u.first_name,u.last_name
		ORDER BY count DESC LIMIT $4`, []any{filter.From, filter.To, filter.DepartmentID, filter.Limit})
	if err != nil {
		return nil, err
	}
	result.TopRequesters = topRequesters
	return result, nil
}
