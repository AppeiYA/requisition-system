-- Optional analytics materialisation. Direct analytics remains the default.
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_requisition_facts AS
SELECT
    r.id,
    r.requested_by,
    r.department_id,
    r.status,
    r.priority,
    r.created_at,
    date_trunc('day', r.created_at) AS created_day,
    r.needed_by,
    COALESCE(SUM(ri.estimated_unit_amount * ri.quantity), 0)::bigint AS estimated_minor,
    COUNT(ri.id)::int AS line_count
FROM requisitions r
LEFT JOIN requisition_items ri ON ri.requisition_id = r.id
GROUP BY r.id, r.requested_by, r.department_id, r.status, r.priority,
         r.created_at, r.needed_by
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS mv_requisition_facts_id
    ON mv_requisition_facts (id);
CREATE INDEX IF NOT EXISTS mv_requisition_facts_created_day
    ON mv_requisition_facts (created_day);
CREATE INDEX IF NOT EXISTS mv_requisition_facts_department_created_day
    ON mv_requisition_facts (department_id, created_day);
CREATE INDEX IF NOT EXISTS mv_requisition_facts_status_created_day
    ON mv_requisition_facts (status, created_day);

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_requisition_daily AS
SELECT
    created_day,
    department_id,
    status,
    priority,
    COUNT(*)::bigint AS requisition_count,
    SUM(estimated_minor)::bigint AS estimated_minor,
    SUM(line_count)::bigint AS line_count
FROM mv_requisition_facts
GROUP BY created_day, department_id, status, priority
WITH NO DATA;

CREATE UNIQUE INDEX IF NOT EXISTS mv_requisition_daily_grain
    ON mv_requisition_daily (created_day, department_id, status, priority);

CREATE TABLE IF NOT EXISTS analytics_refresh_state (
    view_name VARCHAR(64) PRIMARY KEY,
    refreshed_at TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL,
    duration_ms INTEGER NULL,
    row_count BIGINT NULL,
    last_error TEXT NULL
);
