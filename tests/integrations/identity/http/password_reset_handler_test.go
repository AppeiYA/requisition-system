package http_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/stretchr/testify/require"
)

type passwordResetEnvelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type authTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func postPasswordResetJSON(t *testing.T, identity *testutil.IdentityHTTP, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	identity.Router.ServeHTTP(recorder, request)
	return recorder
}

func decodePasswordResetEnvelope(t *testing.T, recorder *httptest.ResponseRecorder) passwordResetEnvelope {
	t.Helper()

	var envelope passwordResetEnvelope
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope), recorder.Body.String())
	return envelope
}

func loginForPasswordReset(t *testing.T, identity *testutil.IdentityHTTP, email, password string) *httptest.ResponseRecorder {
	t.Helper()

	return postPasswordResetJSON(t, identity, "/api/v1/auth/login", map[string]string{
		"email": email, "password": password,
	})
}

func requestResetEmail(t *testing.T, identity *testutil.IdentityHTTP, email string) *httptest.ResponseRecorder {
	t.Helper()

	return postPasswordResetJSON(t, identity, "/api/v1/auth/forgot-password", map[string]string{"email": email})
}

func resetWithToken(t *testing.T, identity *testutil.IdentityHTTP, token, password string) *httptest.ResponseRecorder {
	t.Helper()

	return postPasswordResetJSON(t, identity, "/api/v1/auth/reset-password", map[string]string{
		"token": token, "password": password,
	})
}

func TestForgotPasswordHandler_NeverRevealsAccountState(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	active := identity.AddUser(t, "usr-forgot-active", "forgot.active@example.com", "Password123$", "Forgot", "Active")
	suspended := identity.AddSuspendedUser(t, "usr-forgot-suspended", "forgot.suspended@example.com", "Password123$", "Forgot", "Suspended")

	const genericMessage = "if an account with that email exists, a password reset link has been sent"

	tests := []struct {
		name      string
		email     string
		wantEmail bool
	}{
		{name: "known active", email: active.Email, wantEmail: true},
		{name: "unknown", email: "nobody@example.com"},
		{name: "suspended", email: suspended.Email},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := len(identity.PasswordResetQueue.Payloads)
			recorder := requestResetEmail(t, identity, tt.email)
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			body := decodePasswordResetEnvelope(t, recorder)
			require.True(t, body.Success)
			require.Equal(t, genericMessage, body.Message)
			require.NotContains(t, recorder.Body.String(), tt.email)
			require.NotContains(t, recorder.Body.String(), "token")

			after := len(identity.PasswordResetQueue.Payloads)
			if tt.wantEmail {
				require.Equal(t, before+1, after)
				payload := identity.PasswordResetQueue.Payloads[after-1]
				require.Equal(t, tt.email, payload.To)
				require.NotEmpty(t, payload.Token)
				require.NotContains(t, recorder.Body.String(), payload.Token)
			} else {
				require.Equal(t, before, after)
			}
		})
	}

	payload := identity.PasswordResetQueue.Payloads[0]
	sum := sha256.Sum256([]byte(payload.Token))
	hash := hex.EncodeToString(sum[:])

	stored, err := identity.PasswordResetTokenRepository.GetByTokenHash(context.Background(), hash)
	require.NoError(t, err)
	require.Equal(t, hash, stored.TokenHash())
	require.NotEqual(t, payload.Token, stored.TokenHash())
	require.Equal(t, active.User.ID(), stored.UserID())

	var rawInDatabase int
	err = suite.SQLX.GetContext(context.Background(), &rawInDatabase,
		`SELECT COUNT(*) FROM password_reset_tokens WHERE token_hash = $1`, payload.Token)
	require.NoError(t, err)
	require.Zero(t, rawInDatabase)
}

func TestForgotPasswordHandler_RateLimit(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(t, "usr-forgot-limit", "forgot.limit@example.com", "Password123$", "Forgot", "Limit")

	for i := 0; i < 5; i++ {
		recorder := requestResetEmail(t, identity, fixture.Email)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	}

	recorder := requestResetEmail(t, identity, fixture.Email)
	require.Equal(t, http.StatusTooManyRequests, recorder.Code, recorder.Body.String())
	require.NotEmpty(t, recorder.Header().Get("Retry-After"))
}

func TestResetPasswordHandler_EndToEnd(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(t, "usr-reset-e2e", "reset.e2e@example.com", "Password123$", "Reset", "E2E")

	login := loginForPasswordReset(t, identity, fixture.Email, fixture.Password)
	require.Equal(t, http.StatusOK, login.Code, login.Body.String())
	var before authTokens
	require.NoError(t, json.Unmarshal(decodePasswordResetEnvelope(t, login).Data, &before))
	require.NotEmpty(t, before.RefreshToken)

	recorder := requestResetEmail(t, identity, fixture.Email)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Len(t, identity.PasswordResetQueue.Payloads, 1)
	token := identity.PasswordResetQueue.Payloads[0].Token

	recorder = resetWithToken(t, identity, token, "NewPassword456$")
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	body := decodePasswordResetEnvelope(t, recorder)
	require.Equal(t, "password reset successful", body.Message)
	require.Empty(t, body.Data)
	require.NotContains(t, recorder.Body.String(), token)

	recorder = loginForPasswordReset(t, identity, fixture.Email, fixture.Password)
	require.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())

	recorder = loginForPasswordReset(t, identity, fixture.Email, "NewPassword456$")
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

	refresh := postPasswordResetJSON(t, identity, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": before.RefreshToken,
	})
	require.Equal(t, http.StatusUnauthorized, refresh.Code, refresh.Body.String())

	reuse := resetWithToken(t, identity, token, "AnotherPassword789$")
	require.Equal(t, http.StatusBadRequest, reuse.Code, reuse.Body.String())
	require.Equal(t, "invalid or expired password reset token", decodePasswordResetEnvelope(t, reuse).Message)
}

func TestResetPasswordHandler_ExpiredAndSupersededTokens(t *testing.T) {
	t.Run("expired", func(t *testing.T) {
		suite := testutil.NewSuite(t, "../../../../migrations")
		identity := suite.SetupIdentityHTTP(t)
		fixture := identity.AddUser(t, "usr-reset-expired", "reset.expired@example.com", "Password123$", "Reset", "Expired")

		recorder := requestResetEmail(t, identity, fixture.Email)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		payload := identity.PasswordResetQueue.Payloads[0]
		sum := sha256.Sum256([]byte(payload.Token))
		hash := hex.EncodeToString(sum[:])

		_, err := suite.SQLX.ExecContext(context.Background(),
			`UPDATE password_reset_tokens SET expires_at = NOW() - INTERVAL '1 minute' WHERE token_hash = $1`, hash)
		require.NoError(t, err)

		recorder = resetWithToken(t, identity, payload.Token, "NewPassword456$")
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
		require.Equal(t, "invalid or expired password reset token", decodePasswordResetEnvelope(t, recorder).Message)
	})

	t.Run("superseded", func(t *testing.T) {
		suite := testutil.NewSuite(t, "../../../../migrations")
		identity := suite.SetupIdentityHTTP(t)
		fixture := identity.AddUser(t, "usr-reset-superseded", "reset.superseded@example.com", "Password123$", "Reset", "Superseded")

		require.Equal(t, http.StatusOK, requestResetEmail(t, identity, fixture.Email).Code)
		require.Equal(t, http.StatusOK, requestResetEmail(t, identity, fixture.Email).Code)
		require.Len(t, identity.PasswordResetQueue.Payloads, 2)
		oldToken := identity.PasswordResetQueue.Payloads[0].Token
		newToken := identity.PasswordResetQueue.Payloads[1].Token

		recorder := resetWithToken(t, identity, oldToken, "NewPassword456$")
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
		require.Equal(t, "invalid or expired password reset token", decodePasswordResetEnvelope(t, recorder).Message)

		recorder = resetWithToken(t, identity, newToken, "NewPassword456$")
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	})
}

func TestResetPasswordHandler_WeakPasswordLeavesTokenUsable(t *testing.T) {
	suite := testutil.NewSuite(t, "../../../../migrations")
	identity := suite.SetupIdentityHTTP(t)
	fixture := identity.AddUser(t, "usr-reset-weak", "reset.weak@example.com", "Password123$", "Reset", "Weak")

	require.Equal(t, http.StatusOK, requestResetEmail(t, identity, fixture.Email).Code)
	token := identity.PasswordResetQueue.Payloads[0].Token

	recorder := resetWithToken(t, identity, token, "weak")
	require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())

	recorder = resetWithToken(t, identity, token, "NewPassword456$")
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
}
