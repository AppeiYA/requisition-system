package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	invitationhttp "github.com/AppeiYA/requisition-system/internal/invitation/adapters/in/http"
	invitationpostgres "github.com/AppeiYA/requisition-system/internal/invitation/adapters/out/postgres"
	invitationdomain "github.com/AppeiYA/requisition-system/internal/invitation/domain"
	invitationin "github.com/AppeiYA/requisition-system/internal/invitation/ports/in"
	invitationusecase "github.com/AppeiYA/requisition-system/internal/invitation/usecase"
	shareddomain "github.com/AppeiYA/requisition-system/internal/shared/domain"
	appdb "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestValidateInvitationHandler(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	ctx := context.Background()
	require.NoError(t, suite.ResetDatabase(ctx))

	_, err := suite.DB.Pool.Exec(ctx, `
		INSERT INTO users (
			id, first_name, last_name, email, password_hash, status, role
		) VALUES (
			'usr-admin', 'Admin', 'User', 'admin@example.com',
			'integration-password-hash', 'active', 'admin'
		);
		INSERT INTO departments (id, name, description, code, status)
		VALUES ('dpt-integration', 'Integration', 'Integration department', 'INT', 'active');
	`)
	require.NoError(t, err)

	baseRepository := appdb.NewRepository(suite.AppDB)
	invitationRepository := invitationpostgres.NewInvitationRepository(baseRepository)
	validateUsecase := invitationusecase.NewValidateInvitationUsecase(invitationRepository)
	ports := invitationin.NewInvitationPortIn(nil, validateUsecase, nil, nil)
	handler := invitationhttp.NewInvitationHandler(ports)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.Register(router.Group("/api/v1"))

	rawToken := "valid-integration-token"
	email, err := shareddomain.NewEmail("invitee@example.com")
	require.NoError(t, err)
	firstName, err := shareddomain.NewUsername("Invitee", shareddomain.FirstName)
	require.NoError(t, err)
	lastName, err := shareddomain.NewUsername("User", shareddomain.LastName)
	require.NoError(t, err)
	phone, err := shareddomain.NewOptionalPhoneNumber("")
	require.NoError(t, err)
	invitation, err := invitationdomain.NewInvitation(
		"inv-integration",
		firstName,
		lastName,
		phone,
		"",
		"dpt-integration",
		"Engineer",
		email,
		invitationdomain.HashInvitationToken(rawToken),
		"usr-admin",
	)
	require.NoError(t, err)
	require.NoError(t, invitationRepository.SaveInvitation(context.Background(), invitation))

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedValid  bool
	}{
		{name: "valid", token: rawToken, expectedStatus: http.StatusOK, expectedValid: true},
		{name: "unknown", token: "unknown-token", expectedStatus: http.StatusOK, expectedValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodGet,
				"/api/v1/invitations/"+tt.token+"/validate",
				nil,
			)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			require.Equal(t, tt.expectedStatus, recorder.Code, recorder.Body.String())
			var envelope struct {
				Success bool `json:"success"`
				Data    struct {
					Valid bool `json:"valid"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			require.True(t, envelope.Success)
			require.Equal(t, tt.expectedValid, envelope.Data.Valid)
		})
	}
}
