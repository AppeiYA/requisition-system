package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func insertAdminInvitation(t *testing.T, suite *testutil.Suite, id, email, status, token string, expiresAt string) {
	t.Helper()
	_, err := suite.DB.Pool.Exec(context.Background(), `
		INSERT INTO invitations (
			id, first_name, last_name, email, department_id, job_title,
			token_hash, invited_by, status, expires_at, created_at
		) VALUES ($1, 'Jane', 'Doe', $2, 'dpt-inv-test', 'Engineer', $3, 'usr-inv-admin', $4, $5, NOW())
	`, id, email, invitationdomain.HashInvitationToken(token), status, expiresAt)
	require.NoError(t, err)
}

func invitationValidateRouter(admin *testutil.AdminHTTP) *gin.Engine {
	ports := invitationin.NewInvitationPortIn(
		nil,
		invitationusecase.NewValidateInvitationUsecase(admin.InvitationRepository),
		nil,
		nil,
	)
	handler := invitationhttp.NewInvitationHandler(ports)
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.Register(router.Group("/api/v1"))
	return router
}

func invitationValidateStatus(t *testing.T, router *gin.Engine, token string) int {
	t.Helper()
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/invitations/"+token+"/validate", nil))
	var body struct {
		Data struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	if body.Data.Valid {
		return http.StatusOK
	}
	return http.StatusForbidden
}

func TestAdminInvitationHandlers(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	admin := suite.SetupAdminHTTP(t)
	ctx := context.Background()

	admin.AddAdmin(
		t, "usr-inv-admin", "inv.admin@example.com", "Password123$", "Invite", "Admin",
	)
	admin.AddUser(
		t, "usr-inv-user", "inv.user@example.com", "Password123$", "Invite", "User",
	)
	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO departments (id, name, description, code, status)
		VALUES ('dpt-inv-test', 'Finance', 'Finance department', 'FIN', 'active')
	`)
	require.NoError(t, err)

	insertAdminInvitation(t, suite, "inv-list-one", "one@example.com", "pending", "token-one", "2099-01-01T00:00:00Z")
	insertAdminInvitation(t, suite, "inv-list-two", "two@example.com", "accepted", "token-two", "2099-01-01T00:00:00Z")

	adminToken := accessToken(t, admin.AccessTokens, "usr-inv-admin", "admin")
	userToken := accessToken(t, admin.AccessTokens, "usr-inv-user", "user")
	validateRouter := invitationValidateRouter(admin)
	var rotatedToken string

	t.Run("requires admin authentication", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations", "", "")
		require.Equal(t, http.StatusUnauthorized, rec.Code, rec.Body.String())

		rec = adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations", userToken, "")
		require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
	})

	t.Run("lists invitations with pagination total", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations?limit=1&offset=0", adminToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var body struct {
			Success bool `json:"success"`
			Data    []struct {
				ID             string `json:"id"`
				Email          string `json:"email"`
				DepartmentName string `json:"department_name"`
			} `json:"data"`
			Pagination struct {
				Total int `json:"total"`
			} `json:"pagination"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.True(t, body.Success)
		require.Equal(t, 2, body.Pagination.Total)
		require.Len(t, body.Data, 1)
		require.Equal(t, "Finance", body.Data[0].DepartmentName)
	})

	t.Run("filters by status and rejects invalid filters", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations?status=accepted", adminToken, "")
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
		require.Equal(t, "inv-list-two", body.Data[0].ID)

		rec = adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations?status=unknown", adminToken, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())

		rec = adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations?from=not-a-date", adminToken, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	})

	t.Run("returns invitation detail", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations/inv-list-one", adminToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var body struct {
			Data struct {
				ID          string `json:"id"`
				Status      string `json:"status"`
				InviterName string `json:"inviter_name"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		require.Equal(t, "inv-list-one", body.Data.ID)
		require.Equal(t, "pending", body.Data.Status)
		require.Equal(t, "Invite Admin", body.Data.InviterName)

		rec = adminRequest(admin, http.MethodGet, "/api/v1/admin/invitations/inv-missing", adminToken, "")
		require.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
	})

	t.Run("rotates the token and queues a new email", func(t *testing.T) {
		type queuedEmail struct {
			InvitationID string
			To           string
			Token        string
		}
		var queued []queuedEmail
		admin.TaskQueue.EnqueueInvitationEmailFunc = func(_ context.Context, payload tasks.InvitationEmailPayload) error {
			queued = append(queued, queuedEmail{payload.InvitationID, payload.To, payload.Token})
			return nil
		}

		require.Equal(t, http.StatusOK, invitationValidateStatus(t, validateRouter, "token-one"))

		rec := adminRequest(admin, http.MethodPost, "/api/v1/admin/invitations/inv-list-one/resend", adminToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Len(t, queued, 1)
		require.Equal(t, "inv-list-one", queued[0].InvitationID)
		require.Equal(t, "one@example.com", queued[0].To)
		require.NotEqual(t, "token-one", queued[0].Token)
		rotatedToken = queued[0].Token

		require.Equal(t, http.StatusForbidden, invitationValidateStatus(t, validateRouter, "token-one"))
		require.Equal(t, http.StatusOK, invitationValidateStatus(t, validateRouter, queued[0].Token))

		require.Equal(t, 1, adminAuditCount(t, suite, "invitation.resend", "inv-list-one"))
	})

	t.Run("rejects resending an accepted invitation", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodPost, "/api/v1/admin/invitations/inv-list-two/resend", adminToken, "")
		require.Equal(t, http.StatusConflict, rec.Code, rec.Body.String())
	})

	t.Run("rejects an oversized revoke reason", func(t *testing.T) {
		reason := strings.Repeat("x", 501)
		rec := adminRequest(admin, http.MethodPost, "/api/v1/admin/invitations/inv-list-one/revoke?reason="+reason, adminToken, "")
		require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	})

	t.Run("revokes an invitation idempotently", func(t *testing.T) {
		rec := adminRequest(admin, http.MethodPost, "/api/v1/admin/invitations/inv-list-one/revoke?reason=duplicate", adminToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

		var status string
		err := suite.DB.Pool.QueryRow(ctx, `SELECT status FROM invitations WHERE id = $1`, "inv-list-one").Scan(&status)
		require.NoError(t, err)
		require.Equal(t, "revoked", status)
		require.Equal(t, http.StatusForbidden, invitationValidateStatus(t, validateRouter, rotatedToken))
		require.Equal(t, 1, adminAuditCount(t, suite, "invitation.revoke", "inv-list-one"))

		rec = adminRequest(admin, http.MethodPost, "/api/v1/admin/invitations/inv-list-one/revoke", adminToken, "")
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		require.Equal(t, 1, adminAuditCount(t, suite, "invitation.revoke", "inv-list-one"))
	})
}
