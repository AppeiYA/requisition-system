package analyticspostgres

import "time"

type usersOverviewRow struct {
	Total      int `db:"total"`
	Active     int `db:"active"`
	Suspended  int `db:"suspended"`
	Inactive   int `db:"inactive"`
	Admins     int `db:"admins"`
	NewInRange int `db:"new_in_range"`
}

type invitationsOverviewRow struct {
	Total     int `db:"total"`
	Pending   int `db:"pending"`
	Delivered int `db:"delivered"`
	Accepted  int `db:"accepted"`
	Expired   int `db:"expired"`
	Revoked   int `db:"revoked"`
	Failed    int `db:"failed"`
}

type requisitionTotalsRow struct {
	Total   int   `db:"total"`
	Minor   int64 `db:"minor"`
	Overdue int   `db:"overdue"`
}

type statusCountRow struct {
	Status string `db:"status"`
	Count  int    `db:"count"`
}

type catalogueOverviewRow struct {
	ActiveItems      int `db:"active_items"`
	ActiveCategories int `db:"active_categories"`
	Unreviewed       int `db:"unreviewed"`
}

type breakdownRow struct {
	Name  string `db:"name"`
	Count int    `db:"count"`
	Minor int64  `db:"minor"`
}

type trendMinorRow struct {
	Bucket time.Time `db:"bucket"`
	Count  int       `db:"count"`
	Minor  int64     `db:"minor"`
}

type trendCountRow struct {
	Bucket time.Time `db:"bucket"`
	Count  int       `db:"count"`
}

type nameCountRow struct {
	Name  string `db:"name"`
	Count int    `db:"count"`
}

type departmentBreakdownRow struct {
	Name          string `db:"name"`
	Members       int    `db:"members"`
	ActiveMembers int    `db:"active_members"`
}

type invitationFunnelRow struct {
	Created   int     `db:"created"`
	Delivered int     `db:"delivered"`
	Accepted  int     `db:"accepted"`
	Expired   int     `db:"expired"`
	Revoked   int     `db:"revoked"`
	Failed    int     `db:"failed"`
	Average   float64 `db:"average"`
}

type catalogueCountsRow struct {
	ActiveItems        int `db:"active_items"`
	InactiveItems      int `db:"inactive_items"`
	ActiveCategories   int `db:"active_categories"`
	InactiveCategories int `db:"inactive_categories"`
	Unreviewed         int `db:"unreviewed"`
	Promoted           int `db:"promoted"`
	Dismissed          int `db:"dismissed"`
	Older7             int `db:"older7"`
	Older30            int `db:"older30"`
}
