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

func TestLoginHandlerSuccess(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(
		t,
		"usr-login-handler",
		"login.handler@example.com",
		"Password123$",
		"Login",
		"Handler",
	)
	suspendedUser := identity.AddSuspendedUser(
		t,
		"usr-suspended",
		"suspended@example.com",
		"Password123$",
		"Suspended",
		"User",
	)

	tests := []struct {
		name            string
		email           string
		password        string
		expectedStatus  int
		wantErr         bool
		expectedMessage string
	}{
		{
			name:           "Successful login",
			email:          fixture.Email,
			password:       fixture.Password,
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:            "Invalid password",
			email:           fixture.Email,
			password:        "WrongPassword123$",
			expectedStatus:  http.StatusUnauthorized,
			wantErr:         true,
			expectedMessage: "invalid email or password",
		},
		{
			name:            "Non-existent user",
			email:           "non.existent@example.com",
			password:        "Password123$",
			expectedStatus:  http.StatusUnauthorized,
			wantErr:         true,
			expectedMessage: "invalid email or password",
		},
		{
			name:            "Suspended user",
			email:           suspendedUser.Email,
			password:        suspendedUser.Password,
			expectedStatus:  http.StatusUnauthorized,
			wantErr:         true,
			expectedMessage: "invalid email or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]string{
				"email":    tt.email,
				"password": tt.password,
			})
			require.NoError(t, err)

			request := httptest.NewRequest(
				http.MethodPost,
				"/api/v1/auth/login",
				bytes.NewReader(body),
			)
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			identity.Router.ServeHTTP(recorder, request)

			require.Equal(t, tt.expectedStatus, recorder.Code, recorder.Body.String())

			if !tt.wantErr {
				var response struct {
					Success bool   `json:"success"`
					Message string `json:"message"`
					Data    struct {
						AccessToken  string `json:"access_token"`
						RefreshToken string `json:"refresh_token"`
						TokenType    string `json:"token_type"`
						ExpiresIn    int64  `json:"expires_in"`
					} `json:"data"`
				}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &response))
				require.True(t, response.Success)
				require.Equal(t, "login successful", response.Message)
				require.NotEmpty(t, response.Data.AccessToken)
				require.NotEmpty(t, response.Data.RefreshToken)
				require.Equal(t, "Bearer", response.Data.TokenType)
				require.Positive(t, response.Data.ExpiresIn)
			} else {
				var errorResponse struct {
					Success bool   `json:"success"`
					Message string `json:"message"`
				}
				require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &errorResponse))
				require.False(t, errorResponse.Success)
				require.NotEmpty(t, errorResponse.Message)
				if tt.expectedMessage != "" {
					require.Equal(t, tt.expectedMessage, errorResponse.Message)
				}
			}
		})
	}
}

func TestLoginHandlerRateLimit(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(
		t,
		"usr-login-rate-limit",
		"login.rate.limit@example.com",
		"Password123$",
		"Login",
		"RateLimit",
	)

	body, err := json.Marshal(map[string]string{
		"email":    fixture.Email,
		"password": "WrongPassword123$",
	})
	require.NoError(t, err)

	for attempt := 1; attempt <= 6; attempt++ {
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/auth/login",
			bytes.NewReader(body),
		)
		request.Header.Set("Content-Type", "application/json")
		request.RemoteAddr = "203.0.113.10:12345"
		recorder := httptest.NewRecorder()

		identity.Router.ServeHTTP(recorder, request)

		if attempt <= 5 {
			require.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
		} else {
			require.Equal(t, http.StatusTooManyRequests, recorder.Code, recorder.Body.String())
			require.NotEmpty(t, recorder.Header().Get("Retry-After"))
		}
	}
}
