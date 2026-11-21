package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type responseEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func loginIdentity(t *testing.T, identity *testutil.IdentityHTTP, email, password string) tokenResponse {
	t.Helper()

	body, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	identity.Router.ServeHTTP(recorder, request)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	var envelope struct {
		Success bool          `json:"success"`
		Data    tokenResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
	require.True(t, envelope.Success)
	require.NotEmpty(t, envelope.Data.AccessToken)
	require.NotEmpty(t, envelope.Data.RefreshToken)

	return envelope.Data
}

func TestRefreshHandler(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(
		t,
		"usr-refresh-handler",
		"refresh.handler@example.com",
		"Password123$",
		"Refresh",
		"Handler",
	)
	loginTokens := loginIdentity(t, identity, fixture.Email, fixture.Password)

	tests := []struct {
		name           string
		refreshToken   string
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "Successful refresh",
			refreshToken:   loginTokens.RefreshToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing refresh token",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "Unknown refresh token",
			refreshToken:   "unknown-refresh-token",
			expectedStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{
				"refresh_token": tt.refreshToken,
			})
			require.NoError(t, err)

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/refresh",
				bytes.NewReader(body),
			)
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			identity.Router.ServeHTTP(recorder, request)

			require.Equal(t, tt.expectedStatus, recorder.Code, recorder.Body.String())

			var envelope responseEnvelope
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			require.Equal(t, !tt.wantErr, envelope.Success)

			if !tt.wantErr {
				var success struct {
					Data tokenResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &success))
				require.NotEmpty(t, success.Data.AccessToken)
				require.NotEmpty(t, success.Data.RefreshToken)
				require.NotEqual(t, loginTokens.RefreshToken, success.Data.RefreshToken)
			} else {
				require.NotEmpty(t, envelope.Message)
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(
		t,
		"usr-logout-handler",
		"logout.handler@example.com",
		"Password123$",
		"Logout",
		"Handler",
	)
	loginTokens := loginIdentity(t, identity, fixture.Email, fixture.Password)

	tests := []struct {
		name           string
		refreshToken   string
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "Successful logout",
			refreshToken:   loginTokens.RefreshToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing refresh token",
			expectedStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{
				"refresh_token": tt.refreshToken,
			})
			require.NoError(t, err)

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/logout",
				bytes.NewReader(body),
			)
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			identity.Router.ServeHTTP(recorder, request)

			require.Equal(t, tt.expectedStatus, recorder.Code, recorder.Body.String())

			var envelope responseEnvelope
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			require.Equal(t, !tt.wantErr, envelope.Success)

			if !tt.wantErr {
				require.Equal(t, "logout successful", envelope.Message)
			} else {
				require.NotEmpty(t, envelope.Message)
			}
		})
	}

	refreshBody, err := json.Marshal(map[string]string{
		"refresh_token": loginTokens.RefreshToken,
	})
	require.NoError(t, err)
	refreshRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/refresh",
		bytes.NewReader(refreshBody),
	)
	refreshRequest.Header.Set("Content-Type", "application/json")
	refreshRecorder := httptest.NewRecorder()
	identity.Router.ServeHTTP(refreshRecorder, refreshRequest)
	require.Equal(t, http.StatusUnauthorized, refreshRecorder.Code, refreshRecorder.Body.String())
}

func TestGetMeHandler(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(
		t,
		"usr-me-handler",
		"me.handler@example.com",
		"Password123$",
		"Me",
		"Handler",
	)
	loginTokens := loginIdentity(t, identity, fixture.Email, fixture.Password)

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
		wantErr        bool
	}{
		{
			name:           "Successful get me",
			authorization:  "Bearer " + loginTokens.AccessToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing authorization",
			expectedStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
		{
			name:           "Malformed bearer token",
			authorization:  "Bearer malformed-token",
			expectedStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
			if tt.authorization != "" {
				request.Header.Set("Authorization", tt.authorization)
			}
			recorder := httptest.NewRecorder()
			identity.Router.ServeHTTP(recorder, request)

			require.Equal(t, tt.expectedStatus, recorder.Code, recorder.Body.String())

			var envelope responseEnvelope
			require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
			require.Equal(t, !tt.wantErr, envelope.Success)

			if !tt.wantErr {
				var success struct {
					Data struct {
						ID        string `json:"id"`
						Email     string `json:"email"`
						FirstName string `json:"first_name"`
						LastName  string `json:"last_name"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &success))
				require.Equal(t, "usr-me-handler", success.Data.ID)
				require.Equal(t, fixture.Email, success.Data.Email)
				require.Equal(t, "Me", success.Data.FirstName)
				require.Equal(t, "Handler", success.Data.LastName)
			} else {
				require.NotEmpty(t, envelope.Message)
			}
		})
	}
}
