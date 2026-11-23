package identitypostgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AppeiYA/requisition-system/internal/identity/adapters/out/postgres/models"
	"github.com/AppeiYA/requisition-system/internal/identity/domain"
)

// The summary columns are shared by list and detail so the two paths cannot
// drift. COALESCE keeps nullable columns scannable into plain strings; the
// read model uses "" for "no department" / "no phone".
const userSummaryColumns = `
	u.id, u.first_name, u.last_name, u.email,
	COALESCE(u.staff_id, '') AS staff_id,
	COALESCE(u.phone_number, '') AS phone,
	u.role, u.status,
	COALESCE(dm.department_id, '') AS department_id,
	COALESCE(d.name, '') AS department_name,
	COALESCE(dm.job_title, '') AS job_title,
	COALESCE(dm.status, '') AS membership_status,
	u.created_at, u.updated_at
`

const userSummaryJoins = `
	FROM users u
	LEFT JOIN department_memberships dm ON dm.user_id = u.id
	LEFT JOIN departments d ON d.id = dm.department_id
`

// The four filter parameters are always passed, so the SQL text is constant
// and only the values change. Search terms are parameters, never SQL text.
const userFilterPredicates = `
	WHERE ($1 = '' OR u.email ILIKE '%' || $1 || '%'
		OR u.first_name ILIKE '%' || $1 || '%'
		OR u.last_name ILIKE '%' || $1 || '%'
		OR COALESCE(u.staff_id, '') ILIKE '%' || $1 || '%')
	AND ($2 = '' OR u.role = $2)
	AND ($3 = '' OR u.status = $3)
	AND ($4 = '' OR COALESCE(dm.department_id, '') = $4)
`

// Sort clauses are allow-listed here and again in domain.UserListFilter. The
// map is the authoritative mapping: an unknown value simply does not exist.
var userSortClauses = map[string]string{
	"":            "u.created_at DESC, u.id ASC",
	"created_at":  "u.created_at ASC, u.id ASC",
	"-created_at": "u.created_at DESC, u.id ASC",
	"last_name":   "u.last_name ASC, u.id ASC",
	"email":       "u.email ASC, u.id ASC",
}

type userSummaryRow struct {
	ID               string     `db:"id"`
	FirstName        string     `db:"first_name"`
	LastName         string     `db:"last_name"`
	Email            string     `db:"email"`
	StaffID          string     `db:"staff_id"`
	Phone            string     `db:"phone"`
	Role             string     `db:"role"`
	Status           string     `db:"status"`
	DepartmentID     string     `db:"department_id"`
	DepartmentName   string     `db:"department_name"`
	JobTitle         string     `db:"job_title"`
	MembershipStatus string     `db:"membership_status"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        *time.Time `db:"updated_at"`
}

func (r userSummaryRow) toDomain() domain.UserSummary {
	return domain.UserSummary{
		ID:               r.ID,
		FirstName:        r.FirstName,
		LastName:         r.LastName,
		Email:            r.Email,
		StaffID:          r.StaffID,
		Phone:            r.Phone,
		Role:             r.Role,
		Status:           r.Status,
		DepartmentID:     r.DepartmentID,
		DepartmentName:   r.DepartmentName,
		JobTitle:         r.JobTitle,
		MembershipStatus: r.MembershipStatus,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func userFilterArgs(filter domain.UserListFilter) []any {
	return []any{filter.Search, filter.Role, filter.Status, filter.DepartmentID}
}

func (r *UserRepository) ListUsers(ctx context.Context, filter domain.UserListFilter) ([]domain.UserSummary, error) {
	sortClause, ok := userSortClauses[filter.Sort]
	if !ok {
		return nil, fmt.Errorf("invalid user sort %q", filter.Sort)
	}

	query := `SELECT ` + userSummaryColumns + ` ` + userSummaryJoins + ` ` +
		userFilterPredicates + ` ORDER BY ` + sortClause + ` LIMIT $5 OFFSET $6`
	args := append(userFilterArgs(filter), filter.Limit, filter.Offset)

	var rows []userSummaryRow
	if err := r.repository.Executor(ctx).SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, err
	}

	result := make([]domain.UserSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.toDomain())
	}
	return result, nil
}

func (r *UserRepository) CountUsers(ctx context.Context, filter domain.UserListFilter) (int, error) {
	query := `SELECT COUNT(*) ` + userSummaryJoins + ` ` + userFilterPredicates

	var count int
	if err := r.repository.Executor(ctx).GetContext(ctx, &count, query, userFilterArgs(filter)...); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *UserRepository) GetUserDetail(ctx context.Context, id string) (*domain.UserDetail, error) {
	query := `SELECT ` + userSummaryColumns + ` ` + userSummaryJoins + ` WHERE u.id = $1`

	var row userSummaryRow
	if err := r.repository.Executor(ctx).GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	detail := &domain.UserDetail{
		UserSummary:       row.toDomain(),
		RequisitionCounts: make(map[string]int, len(domain.RequisitionStatuses)),
	}
	for _, status := range domain.RequisitionStatuses {
		detail.RequisitionCounts[status] = 0
	}

	executor := r.repository.Executor(ctx)
	var counts []struct {
		Status string `db:"status"`
		Count  int    `db:"count"`
	}
	if err := executor.SelectContext(ctx, &counts,
		`SELECT status, COUNT(*) AS count FROM requisitions WHERE requested_by = $1 GROUP BY status`, id); err != nil {
		return nil, err
	}
	for _, count := range counts {
		if _, known := detail.RequisitionCounts[count.Status]; known {
			detail.RequisitionCounts[count.Status] = count.Count
		}
	}

	// requisitions no longer stores an estimated cost; the authoritative
	// amount is the sum of its item lines, stored in kobo.
	amountRow := struct {
		KoboTotal int64      `db:"kobo_total"`
		LastAt    *time.Time `db:"last_at"`
	}{}
	if err := executor.GetContext(ctx, &amountRow, `
		SELECT COALESCE(SUM(ri.quantity * ri.estimated_unit_amount), 0) AS kobo_total,
		       MAX(r.created_at) AS last_at
		FROM requisitions r
		LEFT JOIN requisition_items ri ON ri.requisition_id = r.id
		WHERE r.requested_by = $1
	`, id); err != nil {
		return nil, err
	}

	detail.TotalEstimatedAmount = float64(amountRow.KoboTotal) / 100
	detail.LastRequisitionAt = amountRow.LastAt

	return detail, nil
}

func (r *UserRepository) GetUserByIDForUpdate(ctx context.Context, id string) (*domain.User, error) {
	var model models.UserModel
	query := getUserByIDQuery + ` FOR UPDATE`
	if err := r.repository.Executor(ctx).GetContext(ctx, &model, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return model.ToDomain()
}

// CountOtherActiveAdminsForUpdate locks every active admin row in id order and
// returns the count excluding id. Callers must call this before locking the
// target row: locking the target first would let two admins demoting each other
// acquire the locks in opposite order and deadlock. With a stable order the
// second transaction waits, then sees the first one's committed result and
// refuses to remove the last admin.
func (r *UserRepository) CountOtherActiveAdminsForUpdate(ctx context.Context, id string) (int, error) {
	var ids []string
	query := `
		SELECT id FROM users
		WHERE role = 'admin' AND status = 'active'
		ORDER BY id
		FOR UPDATE
	`
	if err := r.repository.Executor(ctx).SelectContext(ctx, &ids, query); err != nil {
		return 0, err
	}

	count := 0
	for _, adminID := range ids {
		if adminID != id {
			count++
		}
	}
	return count, nil
}
