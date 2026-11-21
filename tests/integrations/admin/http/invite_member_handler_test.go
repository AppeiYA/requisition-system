package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

func TestInviteMemberHandler(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	admin := suite.SetupAdminHTTP(t)
	ctx := context.Background()

	adminFixture := admin.AddAdmin(
		t,
		"usr-admin-id",
		"admin@example.com",
		"Password123$",
		"Admin",
		"Login",
	)
	userFixture := admin.AddUser(
		t,
		"usr-user-id",
		"user@example.com",
		"Password123$",
		"Regular",
		"User",
	)

	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO departments (id, name, description, code, status)
		VALUES ('dpt-admin-test', 'Finance', 'Finance department', 'FIN', 'active')
	`)
	require.NoError(t, err)

	adminToken := accessToken(t, admin.AccessTokens, adminFixture.User.ID(), "admin")
	userToken := accessToken(t, admin.AccessTokens, userFixture.User.ID(), "user")

	var queued tasks.InvitationEmailPayload
	admin.TaskQueue.EnqueueInvitationEmailFunc = func(_ context.Context, payload tasks.InvitationEmailPayload) error {
		queued = payload
		return nil
	}

	request := func(token, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/admin/invitations",
			bytes.NewBufferString(body),
		)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		admin.Router.ServeHTTP(rec, req)
		return rec
	}

	t.Run("creates invitation and queues email", func(t *testing.T) {
		rec := request(adminToken, `{
			"email":"invitee@example.com",
			"first_name":"Jane",
			"last_name":"Doe",
			"phone":"+14155551234",
			"staff_id":"STF-1234",
			"department_id":"dpt-admin-test",
			"job_title":"Software Engineer"
		}`)

		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		var response struct {
			Success bool `json:"success"`
			Data    struct {
				InvitationID string `json:"InvitationID"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
		require.True(t, response.Success)
		require.NotEmpty(t, response.Data.InvitationID)

		invitation, err := admin.InvitationRepository.GetInvitationByEmail(ctx, "invitee@example.com")
		require.NoError(t, err)
		require.Equal(t, response.Data.InvitationID, invitation.ID())
		require.Equal(t, adminFixture.User.ID(), invitation.InvitedBy())
		require.Equal(t, "dpt-admin-test", invitation.DepartmentID())
		require.Equal(t, "Software Engineer", invitation.JobTitle())

		require.Equal(t, invitation.ID(), queued.InvitationID)
		require.Equal(t, "invitee@example.com", queued.To)
		require.Equal(t, adminFixture.User.ID(), queued.InvitedBy)
		require.Equal(t, "Finance", queued.DepartmentName)
		require.Equal(t, "Software Engineer", queued.JobTitle)
		require.NotEmpty(t, queued.Token)
	})

	t.Run("rejects non-admin user", func(t *testing.T) {
		rec := request(userToken, `{
			"email":"another@example.com",
			"first_name":"Another",
			"last_name":"User",
			"department_id":"dpt-admin-test",
			"job_title":"Analyst"
		}`)

		require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})

	t.Run("rejects invalid request", func(t *testing.T) {
		rec := request(adminToken, `{"email":"not-an-email"}`)

		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	})

	t.Run("rejects unknown department", func(t *testing.T) {
		rec := request(adminToken, `{
			"email":"unknown-department@example.com",
			"first_name":"Unknown",
			"last_name":"Department",
			"department_id":"dpt-missing",
			"job_title":"Analyst"
		}`)

		require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	})

	t.Run("rejects existing user", func(t *testing.T) {
		rec := request(adminToken, `{
			"email":"user@example.com",
			"first_name":"Existing",
			"last_name":"User",
			"department_id":"dpt-admin-test",
			"job_title":"Analyst"
		}`)

		require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	})
}

func accessToken(
	t *testing.T,
	service identityout.AccessTokenService,
	userID string,
	role string,
) string {
	t.Helper()

	token, err := service.GenerateAccessToken(context.Background(), &identityout.TokenModel{
		ID:   userID,
		Role: role,
	})
	require.NoError(t, err)
	return token.Value
}
