package analyticspostgres

import (
	"context"
	"time"

	analyticsdomain "github.com/AppeiYA/requisition-system/internal/analytics/domain"
)

func (r *Repository) GetCatalogueAnalytics(ctx context.Context, filter analyticsdomain.Filter) (*analyticsdomain.CatalogueAnalytics, error) {
	result := &analyticsdomain.CatalogueAnalytics{
		GeneratedAt: time.Now().UTC(),
		Range:       analyticsdomain.Range{From: filter.From, To: filter.To},
	}
	var counts catalogueCountsRow
	if err := r.repository.Executor(ctx).GetContext(ctx, &counts, `
		SELECT
		  (SELECT COUNT(*) FROM items WHERE status='active')::int AS active_items,
		  (SELECT COUNT(*) FROM items WHERE status='inactive')::int AS inactive_items,
		  (SELECT COUNT(*) FROM item_categories WHERE status='active')::int AS active_categories,
		  (SELECT COUNT(*) FROM item_categories WHERE status='inactive')::int AS inactive_categories,
		  (SELECT COUNT(*) FROM requested_items WHERE review_status='unreviewed')::int AS unreviewed,
		  (SELECT COUNT(*) FROM requested_items WHERE review_status='promoted')::int AS promoted,
		  (SELECT COUNT(*) FROM requested_items WHERE review_status='dismissed')::int AS dismissed,
		  (SELECT COUNT(*) FROM requested_items WHERE review_status='unreviewed' AND created_at < NOW()-INTERVAL '7 days')::int AS older7,
		  (SELECT COUNT(*) FROM requested_items WHERE review_status='unreviewed' AND created_at < NOW()-INTERVAL '30 days')::int AS older30`); err != nil {
		return nil, err
	}
	result.ActiveItems = counts.ActiveItems
	result.InactiveItems = counts.InactiveItems
	result.ActiveCategories = counts.ActiveCategories
	result.InactiveCategories = counts.InactiveCategories
	result.RequestedItemReview = analyticsdomain.ReviewStatusCounts{Unreviewed: counts.Unreviewed, Promoted: counts.Promoted, Dismissed: counts.Dismissed}
	result.ReviewBacklogAging = analyticsdomain.ReviewBacklogAging{OlderThan7Days: counts.Older7, OlderThan30Days: counts.Older30}

	topItems, err := breakdownNameRows(ctx, r, `
		SELECT ri.item_name_snapshot AS name,
		       SUM(ri.quantity)::int AS count,
		       SUM(ri.estimated_unit_amount * ri.quantity)::bigint AS minor
		FROM requisition_items ri
		JOIN requisitions r ON r.id=ri.requisition_id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY ri.item_name_snapshot
		ORDER BY count DESC LIMIT $4`,
		[]any{filter.From, filter.To, filter.DepartmentID, filter.Limit})
	if err != nil {
		return nil, err
	}
	result.TopItems = topItems

	topCategories, err := breakdownNameRows(ctx, r, `
		SELECT c.name,
		       COUNT(DISTINCT r.id)::int AS count,
		       SUM(ri.estimated_unit_amount * ri.quantity)::bigint AS minor
		FROM requisition_items ri
		JOIN requisitions r ON r.id=ri.requisition_id
		JOIN (
		  SELECT ri2.id, COALESCE(i.category_id,rqi.category_id) AS category_id
		  FROM requisition_items ri2
		  LEFT JOIN items i ON i.id=ri2.item_id
		  LEFT JOIN requested_items rqi ON rqi.id=ri2.requested_item_id
		) src ON src.id=ri.id
		JOIN item_categories c ON c.id=src.category_id
		WHERE r.created_at >= $1 AND r.created_at < $2
		  AND ($3='' OR r.department_id=$3)
		GROUP BY c.name ORDER BY minor DESC LIMIT $4`,
		[]any{filter.From, filter.To, filter.DepartmentID, filter.Limit})
	if err != nil {
		return nil, err
	}
	result.TopCategories = topCategories
	return result, nil
}
