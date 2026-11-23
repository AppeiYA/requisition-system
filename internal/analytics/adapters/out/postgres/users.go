package analyticspostgres

import (
	"context"
	"fmt"
	"time"

	analyticsdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
)

func (r *Repository) GetUserAnalytics(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.UserAnalytics, error) {
	result := &analyticsdomain.UserAnalytics{
		GeneratedAt: time.Now().UTC(),
		Range:       analyticsdomain.Range{From: filter.From, To: filter.To},
	}
	args := []any{filter.From, filter.To}

	trendQuery := fmt.Sprintf(`
		SELECT date_trunc('%s', created_at) AS bucket, COUNT(*)::int AS count
		FROM users WHERE created_at >= $1 AND created_at < $2
		GROUP BY bucket ORDER BY bucket`, filter.Interval)
	var growth []trendCountRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &growth, trendQuery, args...); err != nil {
		return nil, err
	}
	result.Growth = make([]analyticsdomain.TrendBucket, len(growth))
	for i, row := range growth {
		result.Growth[i] = analyticsdomain.TrendBucket{Bucket: row.Bucket, Count: row.Count}
	}

	var roles []nameCountRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &roles, `
		SELECT role AS name, COUNT(*)::int AS count FROM users GROUP BY role ORDER BY count DESC`); err != nil {
		return nil, err
	}
	result.ByRole = make([]analyticsdomain.CountValue, len(roles))
	for i, row := range roles {
		result.ByRole[i] = analyticsdomain.CountValue{Name: row.Name, Count: row.Count}
	}

	var statuses []nameCountRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &statuses, `
		SELECT status AS name, COUNT(*)::int AS count FROM users GROUP BY status ORDER BY count DESC`); err != nil {
		return nil, err
	}
	result.ByStatus = make([]analyticsdomain.CountValue, len(statuses))
	for i, row := range statuses {
		result.ByStatus[i] = analyticsdomain.CountValue{Name: row.Name, Count: row.Count}
	}

	var departments []departmentBreakdownRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &departments, `
		SELECT d.name,
		       COUNT(dm.id)::int AS members,
		       COUNT(dm.id) FILTER (WHERE dm.status='active')::int AS active_members
		FROM departments d
		LEFT JOIN department_memberships dm ON dm.department_id=d.id
		GROUP BY d.id,d.name
		ORDER BY members DESC LIMIT $3`, args[0], args[1], filter.Limit); err != nil {
		return nil, err
	}
	result.ByDepartment = make([]analyticsdomain.DepartmentBreakdown, len(departments))
	for i, row := range departments {
		result.ByDepartment[i] = analyticsdomain.DepartmentBreakdown{Name: row.Name, Members: row.Members, ActiveMembers: row.ActiveMembers}
	}

	var funnel invitationFunnelRow
	if err := r.repository.Executor(ctx).GetContext(ctx, &funnel, `
		SELECT COUNT(*)::int AS created,
		       COUNT(*) FILTER (WHERE status='delivered')::int AS delivered,
		       COUNT(*) FILTER (WHERE status='accepted')::int AS accepted,
		       COUNT(*) FILTER (WHERE status='expired')::int AS expired,
		       COUNT(*) FILTER (WHERE status='revoked')::int AS revoked,
		       COUNT(*) FILTER (WHERE status='failed')::int AS failed,
		       COALESCE(AVG(EXTRACT(EPOCH FROM (accepted_at-created_at))) FILTER (WHERE accepted_at IS NOT NULL),0) AS average
		FROM invitations WHERE created_at >= $1 AND created_at < $2`, args...); err != nil {
		return nil, err
	}
	rate := 0.0
	if funnel.Created > 0 {
		rate = float64(funnel.Accepted) / float64(funnel.Created)
	}
	result.InvitationFunnel = analyticsdomain.InvitationFunnel{
		Created: funnel.Created, Delivered: funnel.Delivered, Accepted: funnel.Accepted,
		Expired: funnel.Expired, Revoked: funnel.Revoked, Failed: funnel.Failed,
		AcceptanceRate: rate, AverageTimeToAcceptSeconds: funnel.Average,
	}

	topRequesters, err := breakdownNameRows(ctx, r, `
		SELECT TRIM(u.first_name || ' ' || u.last_name) AS name,
		       COUNT(DISTINCT r.id)::int AS count,
		       COALESCE(SUM(ri.estimated_unit_amount * ri.quantity),0)::bigint AS minor
		FROM requisitions r
		JOIN users u ON u.id=r.requested_by
		LEFT JOIN requisition_items ri ON ri.requisition_id=r.id
		WHERE r.created_at >= $1 AND r.created_at < $2
		GROUP BY u.id,u.first_name,u.last_name
		ORDER BY count DESC LIMIT $3`,
		[]any{filter.From, filter.To, filter.Limit})
	if err != nil {
		return nil, err
	}
	result.TopRequesters = topRequesters
	return result, nil
}
