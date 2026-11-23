package analyticspostgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	analyticsdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
)

type MaterializedRepository struct {
	direct     *Repository
	repository Repository
}

func NewMaterializedRepository(repository Repository) *MaterializedRepository {
	return &MaterializedRepository{direct: &repository, repository: repository}
}

func (r *MaterializedRepository) dataAsOf(ctx context.Context) (time.Time, error) {
	var result sql.NullTime
	err := r.repository.repository.Executor(ctx).GetContext(ctx, &result, `
		SELECT MIN(refreshed_at)
		FROM analytics_refresh_state
		WHERE view_name IN ('mv_requisition_facts', 'mv_requisition_daily')`)
	if err != nil {
		return time.Time{}, err
	}
	if !result.Valid {
		return time.Time{}, nil
	}
	return result.Time, nil
}

func (r *MaterializedRepository) setSource(ctx context.Context, generatedAt time.Time) time.Time {
	if dataAsOf, err := r.dataAsOf(ctx); err == nil && !dataAsOf.IsZero() {
		return dataAsOf
	}
	return generatedAt
}

func (r *MaterializedRepository) GetOverview(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.Overview, error) {
	now := time.Now().UTC()
	result := &analyticsdomain.Overview{
		GeneratedAt: now,
		Range:       analyticsdomain.Range{From: filter.From, To: filter.To},
		Requisitions: analyticsdomain.OverviewRequisitions{
			ByStatus: map[string]int{},
			Currency: "NGN",
		},
	}

	var users usersOverviewRow
	if err := r.repository.repository.Executor(ctx).GetContext(ctx, &users, `
		SELECT COUNT(*)::int AS total,
		       COUNT(*) FILTER (WHERE status='active')::int AS active,
		       COUNT(*) FILTER (WHERE status='suspended')::int AS suspended,
		       COUNT(*) FILTER (WHERE status='inactive')::int AS inactive,
		       COUNT(*) FILTER (WHERE role='admin')::int AS admins,
		       COUNT(*) FILTER (WHERE created_at >= $1 AND created_at < $2)::int AS new_in_range
		FROM users`, filter.From, filter.To); err != nil {
		return nil, err
	}
	result.Users = analyticsdomain.OverviewUsers{
		Total: users.Total, Active: users.Active, Suspended: users.Suspended,
		Inactive: users.Inactive, Admins: users.Admins, NewInRange: users.NewInRange,
	}

	var invitations invitationsOverviewRow
	if err := r.repository.repository.Executor(ctx).GetContext(ctx, &invitations, `
		SELECT COUNT(*)::int AS total,
		       COUNT(*) FILTER (WHERE status='pending')::int AS pending,
		       COUNT(*) FILTER (WHERE status='delivered')::int AS delivered,
		       COUNT(*) FILTER (WHERE status='accepted')::int AS accepted,
		       COUNT(*) FILTER (WHERE status='expired')::int AS expired,
		       COUNT(*) FILTER (WHERE status='revoked')::int AS revoked,
		       COUNT(*) FILTER (WHERE status='failed')::int AS failed
		FROM invitations
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3='' OR department_id=$3)`,
		filter.From, filter.To, filter.DepartmentID); err != nil {
		return nil, err
	}
	acceptanceRate := 0.0
	if invitations.Total > 0 {
		acceptanceRate = float64(invitations.Accepted) / float64(invitations.Total)
	}
	result.Invitations = analyticsdomain.OverviewInvitations{
		Total: invitations.Total, Pending: invitations.Pending, Delivered: invitations.Delivered,
		Accepted: invitations.Accepted, Expired: invitations.Expired, Revoked: invitations.Revoked,
		Failed: invitations.Failed, AcceptanceRate: acceptanceRate,
	}

	var catalogue catalogueOverviewRow
	if err := r.repository.repository.Executor(ctx).GetContext(ctx, &catalogue, `
		SELECT
		  (SELECT COUNT(*) FROM items WHERE status='active')::int AS active_items,
		  (SELECT COUNT(*) FROM item_categories WHERE status='active')::int AS active_categories,
		  (SELECT COUNT(*) FROM requested_items WHERE review_status='unreviewed')::int AS unreviewed`); err != nil {
		return nil, err
	}
	result.Catalogue = analyticsdomain.OverviewCatalogue{
		ActiveItems:              catalogue.ActiveItems,
		ActiveCategories:         catalogue.ActiveCategories,
		UnreviewedRequestedItems: catalogue.Unreviewed,
	}

	var totals requisitionTotalsRow
	if err := r.repository.repository.Executor(ctx).GetContext(ctx, &totals, `
		SELECT COUNT(*)::int AS total,
		       COALESCE(SUM(estimated_minor),0)::bigint AS minor,
		       COUNT(*) FILTER (WHERE status='pending' AND needed_by < CURRENT_DATE)::int AS overdue
		FROM mv_requisition_facts
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3='' OR department_id=$3)`,
		filter.From, filter.To, filter.DepartmentID); err != nil {
		return nil, err
	}
	result.Requisitions.Total = totals.Total
	result.Requisitions.TotalEstimatedAmount = majorMinor(totals.Minor)
	if totals.Total > 0 {
		result.Requisitions.AverageEstimatedAmount = majorMinor(totals.Minor / int64(totals.Total))
	} else {
		result.Requisitions.AverageEstimatedAmount = 0
	}
	result.Requisitions.PendingOverdue = totals.Overdue

	var statuses []statusCountRow
	if err := r.repository.repository.Executor(ctx).SelectContext(ctx, &statuses, `
		SELECT status, COUNT(*)::int AS count
		FROM mv_requisition_facts
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3='' OR department_id=$3)
		GROUP BY status`, filter.From, filter.To, filter.DepartmentID); err != nil {
		return nil, err
	}
	result.Requisitions.ByStatus = map[string]int{}
	for _, row := range statuses {
		result.Requisitions.ByStatus[row.Status] = row.Count
	}

	result.DataAsOf = r.setSource(ctx, result.GeneratedAt)
	result.Source = "materialized"
	return result, nil
}

func (r *MaterializedRepository) GetRequisitionAnalytics(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.RequisitionAnalytics, error) {
	result, err := r.direct.GetRequisitionAnalytics(ctx, filter)
	if err != nil {
		return nil, err
	}
	args := []any{filter.From, filter.To, filter.DepartmentID}

	status, err := breakdownNameRows(ctx, r.direct, `
		SELECT status AS name, COUNT(*)::int AS count,
		       SUM(estimated_minor)::bigint AS minor
		FROM mv_requisition_facts
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3='' OR department_id=$3)
		GROUP BY status ORDER BY count DESC`, args)
	if err != nil {
		return nil, err
	}
	result.ByStatus = status

	priority, err := breakdownNameRows(ctx, r.direct, `
		SELECT priority AS name, COUNT(*)::int AS count,
		       SUM(estimated_minor)::bigint AS minor
		FROM mv_requisition_facts
		WHERE created_at >= $1 AND created_at < $2
		  AND ($3='' OR department_id=$3)
		GROUP BY priority ORDER BY count DESC`, args)
	if err != nil {
		return nil, err
	}
	result.ByPriority = priority

	departments, err := breakdownNameRows(ctx, r.direct, `
		SELECT d.name, COUNT(*)::int AS count,
		       SUM(f.estimated_minor)::bigint AS minor
		FROM mv_requisition_facts f
		JOIN departments d ON d.id=f.department_id
		WHERE f.created_at >= $1 AND f.created_at < $2
		  AND ($3='' OR f.department_id=$3)
		GROUP BY d.name ORDER BY minor DESC`, args)
	if err != nil {
		return nil, err
	}
	result.ByDepartment = departments

	trendQuery := fmt.Sprintf(`
		SELECT date_trunc('%s', created_day) AS bucket,
		       SUM(requisition_count)::int AS count,
		       SUM(estimated_minor)::bigint AS minor
		FROM mv_requisition_daily
		WHERE created_day >= $1 AND created_day < $2
		  AND ($3='' OR department_id=$3)
		GROUP BY bucket ORDER BY bucket`, filter.Interval)
	var trendRows []trendMinorRow
	if err := r.repository.repository.Executor(ctx).SelectContext(ctx, &trendRows, trendQuery, args...); err != nil {
		return nil, err
	}
	result.Trend = make([]analyticsdomain.TrendBucket, len(trendRows))
	for i, row := range trendRows {
		result.Trend[i] = analyticsdomain.TrendBucket{
			Bucket: row.Bucket, Count: row.Count, EstimatedAmount: majorMinor(row.Minor),
		}
	}
	result.DataAsOf = r.setSource(ctx, result.GeneratedAt)
	result.Source = "materialized"
	return result, nil
}

func (r *MaterializedRepository) GetUserAnalytics(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.UserAnalytics, error) {
	result, err := r.direct.GetUserAnalytics(ctx, filter)
	if err != nil {
		return nil, err
	}
	result.DataAsOf = r.setSource(ctx, result.GeneratedAt)
	result.Source = "materialized"
	return result, nil
}

func (r *MaterializedRepository) GetCatalogueAnalytics(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.CatalogueAnalytics, error) {
	result, err := r.direct.GetCatalogueAnalytics(ctx, filter)
	if err != nil {
		return nil, err
	}
	result.DataAsOf = r.setSource(ctx, result.GeneratedAt)
	result.Source = "materialized"
	return result, nil
}
