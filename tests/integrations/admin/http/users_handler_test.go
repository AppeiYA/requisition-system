package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	identitydomain "github.com/AppeiYA/requisition-system/internal/identity/domain"
	identityin "github.com/AppeiYA/requisition-system/internal/identity/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

type adminSessionTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func adminRequest(admin *testutil.AdminHTTP, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	admin.Router.ServeHTTP(rec, req)
	return rec
}

func loginForAdminTest(t *testing.T, admin *testutil.AdminHTTP, email, password string) adminSessionTokens {
	t.Helper()

	body, err := json.Marshal(map[string]string{"email": email, "password": password})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	admin.IdentityHTTP.Router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var envelope struct {
		Data adminSessionTokens `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope))
	require.NotEmpty(t, envelope.Data.RefreshToken)
	return envelope.Data
}

func adminRefreshStatus(t *testing.T, admin *testutil.AdminHTTP, refreshToken string) int {
	t.Helper()

	body, err := json.Marshal(map[string]string{"refresh_token": refreshToken})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	admin.IdentityHTTP.Router.ServeHTTP(rec, req)
	return rec.Code
}

func adminAuditCount(t *testing.T, suite *testutil.Suite, action, targetID string) int {
	t.Helper()

	var count int
	err := suite.DB.Pool.QueryRow(
		context.Background(),
		`SELECT COUNT(*) FROM admin_audit_log WHERE action = $1 AND target_id = $2`,
		action,
		targetID,
	).Scan(&count)
	require.NoError(t, err)
	return count
}

func TestAdminUserReadHandlers(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	admin := suite.SetupAdminHTTP(t)
	ctx := context.Background()

	adminFixture := admin.AddAdmin(
		t, "usr-read-admin", "read.admin@example.com", "Password123$", "Read", "Admin",
	)
	userFixture := admin.AddUser(
		t, "usr-read-user", "read.user@example.com", "Password123$", "Read", "User",
	)

	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO departments (id, name, description, code, status)
		VALUES ('dpt-read-test', 'Finance', 'Finance department', 'FIN', 'active')
	`)
	require.NoError(t, err)
	_, err = suite.DB.Pool.Exec(ctx, `
		INSERT INTO department_memberships (user_id, department_id, job_title)
		VALUES ($1, 'dpt-read-test', 'Analyst')
	`, userFixture.User.ID())
	require.NoError(t, err)

	adminToken := accessToken(t, admin.AccessTokens, adminFixture.User.ID(), "admin")
	userToken := accessToken(t, admin.AccessTokens, userFixture.User.ID(), "user")

	t.Run("rejects a missing token", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/users", "", "")
		require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())
	})

	t.Run("rejects a non-admin token", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/users", userToken, "")
		require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})

	t.Run("lists users with pagination total", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/users?limit=1&offset=0", adminToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var body struct {
			Success bool `json:"success"`
			Data    []struct {
				ID             string `json:"id"`
				Email          string `json:"email"`
				DepartmentName string `json:"department_name"`
			} `json:"data"`
			Pagination struct {
				Total  int `json:"total"`
				Limit  int `json:"limit"`
				Offset int `json:"offset"`
			} `json:"pagination"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.True(t, body.Success)
		require.Equal(t, 2, body.Pagination.Total)
		require.Equal(t, 1, body.Pagination.Limit)
		require.Equal(t, 0, body.Pagination.Offset)
		require.Len(t, body.Data, 1)
	})

	t.Run("filters by department id", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodGet,
			"/api/v1/admin/users?department_id=dpt-read-test",
			adminToken,
			"",
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var body struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
			Pagination struct {
				Total int `json:"total"`
			} `json:"pagination"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Equal(t, 1, body.Pagination.Total)
		require.Len(t, body.Data, 1)
		require.Equal(t, userFixture.User.ID(), body.Data[0].ID)
	})

	t.Run("rejects an invalid filter with field details", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/users?role=manager", adminToken, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())

		var body struct {
			Error   string `json:"error"`
			Details []struct {
				Field   string `json:"field"`
				Message string `json:"message"`
			} `json:"details"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.NotEmpty(t, body.Error)
		require.Len(t, body.Details, 1)
		require.Equal(t, "role", body.Details[0].Field)
	})

	t.Run("returns 404 for an unknown user", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/users/usr-missing", adminToken, "")
		require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	})

	t.Run("returns user detail with department and requisition counts", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/users/"+userFixture.User.ID(), adminToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var body struct {
			Data struct {
				ID                string         `json:"id"`
				Email             string         `json:"email"`
				DepartmentName    string         `json:"department_name"`
				JobTitle          string         `json:"job_title"`
				RequisitionCounts map[string]int `json:"requisition_counts"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Equal(t, userFixture.User.ID(), body.Data.ID)
		require.Equal(t, userFixture.Email, body.Data.Email)
		require.Equal(t, "Finance", body.Data.DepartmentName)
		require.Equal(t, "Analyst", body.Data.JobTitle)
		require.Equal(t, 0, body.Data.RequisitionCounts["pending"])
	})
}

func TestAdminUserStatusHandlers(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	admin := suite.SetupAdminHTTP(t)
	ctx := context.Background()

	adminFixture := admin.AddAdmin(
		t, "usr-status-admin", "status.admin@example.com", "Password123$", "Status", "Admin",
	)
	userFixture := admin.AddUser(
		t, "usr-status-user", "status.user@example.com", "Password123$", "Status", "User",
	)
	admin.AddUser(
		t, "usr-status-noreason", "status.noreason@example.com", "Password123$", "No", "Reason",
	)

	userTokens := loginForAdminTest(t, admin, userFixture.Email, userFixture.Password)
	require.Equal(t, http.StatusOK, adminRefreshStatus(t, admin, userTokens.RefreshToken))

	adminToken := accessToken(t, admin.AccessTokens, adminFixture.User.ID(), "admin")

	t.Run("suspends a user, revokes refresh sessions and writes audit", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPatch,
			"/api/v1/admin/users/"+userFixture.User.ID()+"/status",
			adminToken,
			`{"status":"suspended","reason":"policy violation"}`,
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		require.Equal(t, http.StatusUnauthorized, adminRefreshStatus(t, admin, userTokens.RefreshToken))

		user, err := admin.UserRepository.GetUserByID(ctx, userFixture.User.ID())
		require.NoError(t, err)
		require.Equal(t, identitydomain.SUSPENDED, user.Status())
		require.Equal(t, 1, adminAuditCount(t, suite, "user.status.update", userFixture.User.ID()))
	})

	t.Run("rejects a self status change", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPatch,
			"/api/v1/admin/users/"+adminFixture.User.ID()+"/status",
			adminToken,
			`{"status":"suspended","reason":"self"}`,
		)
		require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())

		user, err := admin.UserRepository.GetUserByID(ctx, adminFixture.User.ID())
		require.NoError(t, err)
		require.Equal(t, identitydomain.ACTIVE, user.Status())
	})

	t.Run("rejects suspending the last active admin", func(t *testing.T) {
		// An authenticated caller is always an active admin, so this guard
		// cannot be reached over HTTP. Calling the identity use case directly
		// verifies the 409 outcome of the same code path.
		err := admin.IdentityAdminUsers.UpdateUserStatus.Execute(ctx, identityin.UpdateUserStatusCommand{
			UserID:  adminFixture.User.ID(),
			Status:  "suspended",
			Reason:  "last admin",
			ActorID: "usr-another-actor",
		})
		var customErr *customerrors.CustomError
		require.ErrorAs(t, err, &customErr)
		require.Equal(t, http.StatusConflict, customErr.Code)
		require.Equal(t, 0, adminAuditCount(t, suite, "user.status.update", adminFixture.User.ID()))
	})

	t.Run("requires a reason when suspending", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPatch,
			"/api/v1/admin/users/usr-status-noreason/status",
			adminToken,
			`{"status":"suspended"}`,
		)
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	})
}

func TestAdminUserRoleHandlers(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	admin := suite.SetupAdminHTTP(t)
	ctx := context.Background()

	actorFixture := admin.AddAdmin(
		t, "usr-role-actor", "role.actor@example.com", "Password123$", "Role", "Actor",
	)
	targetAdmin := admin.AddAdmin(
		t, "usr-role-target", "role.target@example.com", "Password123$", "Role", "Target",
	)
	userFixture := admin.AddUser(
		t, "usr-role-user", "role.user@example.com", "Password123$", "Role", "User",
	)
	admin.AddSuspendedUser(
		t, "usr-role-suspended", "role.suspended@example.com", "Password123$", "Role", "Suspended",
	)

	userTokens := loginForAdminTest(t, admin, userFixture.Email, userFixture.Password)
	require.Equal(t, http.StatusOK, adminRefreshStatus(t, admin, userTokens.RefreshToken))

	adminToken := accessToken(t, admin.AccessTokens, actorFixture.User.ID(), "admin")
	targetAdminToken := accessToken(t, admin.AccessTokens, targetAdmin.User.ID(), "admin")

	t.Run("promotes an active user and revokes sessions", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPatch,
			"/api/v1/admin/users/"+userFixture.User.ID()+"/role",
			adminToken,
			`{"role":"admin","reason":"promotion"}`,
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		require.Equal(t, http.StatusUnauthorized, adminRefreshStatus(t, admin, userTokens.RefreshToken))

		user, err := admin.UserRepository.GetUserByID(ctx, userFixture.User.ID())
		require.NoError(t, err)
		require.Equal(t, "admin", user.Role().String())
		require.Equal(t, 1, adminAuditCount(t, suite, "user.role.update", userFixture.User.ID()))

		// Authorization reads the database role, not the claim in the old
		// token, so the newly promoted user can immediately use admin routes.
		oldUserToken := accessToken(t, admin.AccessTokens, userFixture.User.ID(), "user")
		rec = adminRequest(admin, http.MethodGet, "/api/v1/admin/users", oldUserToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	})

	t.Run("an old admin token loses access after demotion", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPatch,
			"/api/v1/admin/users/"+targetAdmin.User.ID()+"/role",
			adminToken,
			`{"role":"user","reason":"demotion"}`,
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Equal(t, 1, adminAuditCount(t, suite, "user.role.update", targetAdmin.User.ID()))

		rec = adminRequest(admin, http.MethodGet, "/api/v1/admin/users", targetAdminToken, "")
		require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})

	t.Run("rejects promoting a suspended user", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPatch,
			"/api/v1/admin/users/usr-role-suspended/role",
			adminToken,
			`{"role":"admin","reason":"promotion"}`,
		)
		require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	})

	t.Run("rejects a self role change", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPatch,
			"/api/v1/admin/users/"+actorFixture.User.ID()+"/role",
			adminToken,
			`{"role":"user","reason":"self"}`,
		)
		require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})
}

func TestConcurrentAdminDemotions(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	admin := suite.SetupAdminHTTP(t)
	ctx := context.Background()

	adminA := admin.AddAdmin(
		t, "usr-concurrent-a", "concurrent.a@example.com", "Password123$", "Concurrent", "Alpha",
	)
	adminB := admin.AddAdmin(
		t, "usr-concurrent-b", "concurrent.b@example.com", "Password123$", "Concurrent", "Bravo",
	)

	type outcome struct {
		target string
		err    error
	}
	results := make(chan outcome, 2)

	var wg sync.WaitGroup
	for _, cmd := range []identityin.UpdateUserRoleCommand{
		{UserID: adminB.User.ID(), Role: "user", Reason: "rotation", ActorID: adminA.User.ID()},
		{UserID: adminA.User.ID(), Role: "user", Reason: "rotation", ActorID: adminB.User.ID()},
	} {
		cmd := cmd
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- outcome{
				target: cmd.UserID,
				err:    admin.IdentityAdminUsers.UpdateUserRole.Execute(ctx, cmd),
			}
		}()
	}
	wg.Wait()
	close(results)

	successes, conflicts := 0, 0
	for result := range results {
		if result.err == nil {
			successes++
			continue
		}
		var customErr *customerrors.CustomError
		require.ErrorAs(t, result.err, &customErr, "unexpected error for %s: %v", result.target, result.err)
		require.Equal(t, http.StatusConflict, customErr.Code)
		conflicts++
	}
	require.Equal(t, 1, successes)
	require.Equal(t, 1, conflicts)

	var activeAdmins int
	err := suite.DB.Pool.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM users WHERE role = 'admin' AND status = 'active'`,
	).Scan(&activeAdmins)
	require.NoError(t, err)
	require.Equal(t, 1, activeAdmins)
}

func TestAdminRevokeUserSessionsHandler(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	admin := suite.SetupAdminHTTP(t)

	adminFixture := admin.AddAdmin(
		t, "usr-revoke-admin", "revoke.admin@example.com", "Password123$", "Revoke", "Admin",
	)
	userFixture := admin.AddUser(
		t, "usr-revoke-user", "revoke.user@example.com", "Password123$", "Revoke", "User",
	)

	userTokens := loginForAdminTest(t, admin, userFixture.Email, userFixture.Password)
	require.Equal(t, http.StatusOK, adminRefreshStatus(t, admin, userTokens.RefreshToken))

	adminToken := accessToken(t, admin.AccessTokens, adminFixture.User.ID(), "admin")

	t.Run("revokes all refresh sessions and writes audit", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPost,
			"/api/v1/admin/users/"+userFixture.User.ID()+"/sessions/revoke?reason=incident",
			adminToken,
			"",
		)
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Equal(t, http.StatusUnauthorized, adminRefreshStatus(t, admin, userTokens.RefreshToken))
		require.Equal(t, 1, adminAuditCount(t, suite, "user.sessions.revoke", userFixture.User.ID()))
	})

	t.Run("returns 404 for an unknown user", func(t *testing.T) {
		rec := adminRequest(
			admin,
			http.MethodPost,
			"/api/v1/admin/users/usr-revoke-missing/sessions/revoke",
			adminToken,
			"",
		)
		require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	})
}
