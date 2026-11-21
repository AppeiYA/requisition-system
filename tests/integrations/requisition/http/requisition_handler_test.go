package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	departmentdomain "github.com/AppeiYA/requisition-system/internal/department/domain"
	identityout "github.com/AppeiYA/requisition-system/internal/identity/ports/out"
	requisitionhttp "github.com/AppeiYA/requisition-system/internal/requisition/adapters/in/http"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	"github.com/AppeiYA/requisition-system/tests/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

const (
	requisitionHTTPUserID     = "usr-requisition-http-user"
	requisitionHTTPAdminID    = "usr-requisition-http-admin"
	requisitionHTTPOtherID    = "usr-requisition-http-other"
	requisitionHTTPDepartment = "dpt-requisition-http"
	requisitionHTTPCategory   = "cat-requisition-http"
)

type requisitionHTTPFixture struct {
	Requisition *testutil.RequisitionHTTP
	UserID      string
	AdminID     string
	OtherID     string
	Department  string
	Category    string
	NeededBy    string
	UserToken   string
	AdminToken  string
	OtherToken  string
}

func setupRequisitionHTTPFixture(t *testing.T) *requisitionHTTPFixture {
	t.Helper()

	requisition := testutil.NewSuite(t, "../../../../migrations").SetupRequisitionHTTP(t)
	ctx := context.Background()

	user := requisition.AddUser(
		t,
		requisitionHTTPUserID,
		"requisition.user@example.com",
		"Password123$",
		"Requisition",
		"User",
	)
	admin := requisition.AddAdmin(
		t,
		requisitionHTTPAdminID,
		"requisition.admin@example.com",
		"Password123$",
		"Requisition",
		"Admin",
	)
	other := requisition.AddUser(
		t,
		requisitionHTTPOtherID,
		"requisition.other@example.com",
		"Password123$",
		"Requisition",
		"Other",
	)

	name, err := departmentdomain.NewDepartmentName("Requisition")
	require.NoError(t, err)
	description, err := departmentdomain.NewDepartmentDescription("Requisition integration test department")
	require.NoError(t, err)
	code, err := departmentdomain.NewDepartmentCode("RQT")
	require.NoError(t, err)
	department, err := departmentdomain.NewDepartment(
		requisitionHTTPDepartment,
		name,
		description,
		code,
	)
	require.NoError(t, err)
	require.NoError(t, requisition.DepartmentRepository.SaveDepartment(ctx, department))

	requisition.AddMembership(t, user.User.ID(), requisitionHTTPDepartment, "Engineer")

	categoryName, err := requisitiondomain.NewCategoryName("Furniture")
	require.NoError(t, err)
	categoryDescription, err := requisitiondomain.NewCategoryDescription("Office furniture")
	require.NoError(t, err)
	category, err := requisitiondomain.NewItemCategory(
		requisitionHTTPCategory,
		categoryName,
		categoryDescription,
	)
	require.NoError(t, err)
	require.NoError(t, requisition.ItemCategoryRepository.SaveItemCategory(ctx, category))

	return &requisitionHTTPFixture{
		Requisition: requisition,
		UserID:      user.User.ID(),
		AdminID:     admin.User.ID(),
		OtherID:     other.User.ID(),
		Department:  requisitionHTTPDepartment,
		Category:    requisitionHTTPCategory,
		NeededBy:    time.Now().UTC().AddDate(0, 0, 7).Format("2006-01-02"),
		UserToken:   requisitionToken(t, requisition.AccessTokens, user.User.ID(), "user"),
		AdminToken:  requisitionToken(t, requisition.AccessTokens, admin.User.ID(), "admin"),
		OtherToken:  requisitionToken(t, requisition.AccessTokens, other.User.ID(), "user"),
	}
}

func requisitionToken(
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

func (f *requisitionHTTPFixture) payload() map[string]any {
	return map[string]any{
		"department_id": f.Department,
		"priority":      "medium",
		"needed_by":     f.NeededBy,
		"justification": "Replacement furniture for the new hires",
		"lines": []map[string]any{
			{
				"requested_item": map[string]any{
					"name":        "Standing desk",
					"category_id": f.Category,
					"description": "Adjustable standing desk",
				},
				"quantity":              2,
				"estimated_unit_amount": 5000,
				"notes":                 "Deliver to HQ",
			},
		},
	}
}

func performRequisitionRequest(
	t *testing.T,
	router *gin.Engine,
	method string,
	path string,
	token string,
	body any,
) *httptest.ResponseRecorder {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(payload)
	}

	request := httptest.NewRequest(method, path, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeRequisitionEnvelope(
	t *testing.T,
	recorder *httptest.ResponseRecorder,
	target any,
) {
	t.Helper()

	var envelope struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope), recorder.Body.String())
	require.True(t, envelope.Success, recorder.Body.String())
	require.NoError(t, json.Unmarshal(envelope.Data, target))
}

func createRequisitionThroughHTTP(
	t *testing.T,
	fixture *requisitionHTTPFixture,
) requisitionhttp.RequisitionResponse {
	t.Helper()

	recorder := performRequisitionRequest(
		t,
		fixture.Requisition.Router,
		http.MethodPost,
		"/api/v1/requisitions",
		fixture.UserToken,
		fixture.payload(),
	)
	require.Equal(t, http.StatusCreated, recorder.Code, recorder.Body.String())

	var created requisitionhttp.RequisitionResponse
	decodeRequisitionEnvelope(t, recorder, &created)
	return created
}

func TestRequisitionCreateGetAndListHandler(t *testing.T) {
	fixture := setupRequisitionHTTPFixture(t)
	created := createRequisitionThroughHTTP(t, fixture)

	require.True(t, strings.HasPrefix(created.ID, "req-"), created.ID)
	require.True(t, strings.HasPrefix(created.ReferenceCode, "REQ-"), created.ReferenceCode)
	require.Equal(t, fixture.UserID, created.RequestedBy)
	require.Equal(t, fixture.Department, created.DepartmentID)
	require.Equal(t, "pending", created.Status)
	require.Equal(t, "medium", created.Priority)
	require.Equal(t, fixture.NeededBy, created.NeededBy)
	require.Equal(t, "Replacement furniture for the new hires", created.Justification)
	require.Equal(t, "NGN", created.Currency)
	require.Equal(t, 10000.0, created.TotalEstimatedAmount)

	require.Len(t, created.Lines, 1)
	line := created.Lines[0]
	require.True(t, strings.HasPrefix(line.ID, "reqitem-"), line.ID)
	require.Equal(t, 1, line.LineNumber)
	require.Empty(t, line.ItemID)
	require.True(t, strings.HasPrefix(line.RequestedItemID, "ritem-"), line.RequestedItemID)
	require.Equal(t, "Standing desk", line.ItemName)
	require.Equal(t, int64(2), line.Quantity)
	require.Equal(t, 5000.0, line.EstimatedUnitAmount)
	require.Equal(t, 10000.0, line.LineTotal)
	require.Equal(t, "Deliver to HQ", line.Notes)

	require.Len(t, created.StatusHistory, 1)
	history := created.StatusHistory[0]
	require.Equal(t, 1, history.Sequence)
	require.Nil(t, history.FromStatus)
	require.Equal(t, "pending", history.ToStatus)
	require.Equal(t, fixture.UserID, history.ActedBy)
	require.Equal(t, "staff", history.ActorRole)

	persisted, err := fixture.Requisition.RequisitionRepository.GetRequisitionByID(
		context.Background(),
		created.ID,
	)
	require.NoError(t, err)
	require.Equal(t, requisitiondomain.RequisitionStatusPending, persisted.Status())
	require.Len(t, persisted.Items(), 1)
	require.Len(t, persisted.RequestedItems(), 1)
	require.Len(t, persisted.StatusHistory(), 1)

	t.Run("gets the requisition by id", func(t *testing.T) {
		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodGet,
			"/api/v1/requisitions/"+created.ID,
			fixture.UserToken,
			nil,
		)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

		var fetched requisitionhttp.RequisitionResponse
		decodeRequisitionEnvelope(t, recorder, &fetched)
		require.Equal(t, created.ID, fetched.ID)
		require.Equal(t, created.ReferenceCode, fetched.ReferenceCode)
		require.Len(t, fetched.Lines, 1)
		require.Equal(t, created.Lines[0].ID, fetched.Lines[0].ID)
	})

	t.Run("lists the authenticated user's requisitions", func(t *testing.T) {
		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodGet,
			"/api/v1/requisitions?limit=10",
			fixture.UserToken,
			nil,
		)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

		var envelope struct {
			Success    bool                                         `json:"success"`
			Data       []requisitionhttp.RequisitionSummaryResponse `json:"data"`
			Pagination struct {
				Page     int `json:"page"`
				PageSize int `json:"page_size"`
			} `json:"pagination"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		require.True(t, envelope.Success)
		require.Len(t, envelope.Data, 1)
		require.Equal(t, created.ID, envelope.Data[0].ID)
		require.Equal(t, 10000.0, envelope.Data[0].TotalEstimatedAmount)
		require.Equal(t, 1, envelope.Pagination.Page)
		require.Equal(t, 10, envelope.Pagination.PageSize)
	})

	t.Run("lists all requisitions for an administrator", func(t *testing.T) {
		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodGet,
			"/api/v1/requisitions/all",
			fixture.AdminToken,
			nil,
		)
		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())

		var envelope struct {
			Success bool                                         `json:"success"`
			Data    []requisitionhttp.RequisitionSummaryResponse `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &envelope))
		require.True(t, envelope.Success)
		require.Len(t, envelope.Data, 1)
		require.Equal(t, created.ID, envelope.Data[0].ID)
	})

	t.Run("forbids another user from reading the requisition", func(t *testing.T) {
		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodGet,
			"/api/v1/requisitions/"+created.ID,
			fixture.OtherToken,
			nil,
		)
		require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
	})
}

func TestRequisitionValidationAndTransitionHandler(t *testing.T) {
	fixture := setupRequisitionHTTPFixture(t)

	t.Run("requires authentication", func(t *testing.T) {
		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions",
			"",
			fixture.payload(),
		)
		require.Equal(t, http.StatusUnauthorized, recorder.Code, recorder.Body.String())
	})

	t.Run("rejects a request without lines", func(t *testing.T) {
		payload := fixture.payload()
		delete(payload, "lines")

		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions",
			fixture.UserToken,
			payload,
		)
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	})

	t.Run("rejects a past needed by date", func(t *testing.T) {
		payload := fixture.payload()
		payload["needed_by"] = "2020-01-01"

		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions",
			fixture.UserToken,
			payload,
		)
		require.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	})

	t.Run("rejects a department that differs from the membership", func(t *testing.T) {
		payload := fixture.payload()
		payload["department_id"] = "dpt-another-department"

		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions",
			fixture.UserToken,
			payload,
		)
		require.Equal(t, http.StatusForbidden, recorder.Code, recorder.Body.String())
	})

	t.Run("rejects a user without a department membership", func(t *testing.T) {
		recorder := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions",
			fixture.OtherToken,
			fixture.payload(),
		)
		require.Equal(t, http.StatusNotFound, recorder.Code, recorder.Body.String())
	})

	t.Run("administrator approves and the owner cancels", func(t *testing.T) {
		approvedCandidate := createRequisitionThroughHTTP(t, fixture)

		staffApproval := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions/"+approvedCandidate.ID+"/status",
			fixture.UserToken,
			map[string]any{
				"to_status": "approved",
				"reason":    "Approved by the requester",
			},
		)
		require.Equal(t, http.StatusForbidden, staffApproval.Code, staffApproval.Body.String())

		adminApproval := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions/"+approvedCandidate.ID+"/status",
			fixture.AdminToken,
			map[string]any{
				"to_status": "approved",
				"reason":    "Approved by the administrator",
			},
		)
		require.Equal(t, http.StatusOK, adminApproval.Code, adminApproval.Body.String())

		var approved requisitionhttp.RequisitionResponse
		decodeRequisitionEnvelope(t, adminApproval, &approved)
		require.Equal(t, "approved", approved.Status)
		require.Len(t, approved.StatusHistory, 2)
		require.Equal(t, 2, approved.StatusHistory[1].Sequence)
		require.NotNil(t, approved.StatusHistory[1].FromStatus)
		require.Equal(t, "pending", *approved.StatusHistory[1].FromStatus)
		require.Equal(t, "approved", approved.StatusHistory[1].ToStatus)
		require.Equal(t, fixture.AdminID, approved.StatusHistory[1].ActedBy)
		require.Equal(t, "admin", approved.StatusHistory[1].ActorRole)
		require.Equal(t, "Approved by the administrator", approved.StatusHistory[1].Reason)

		repeatApproval := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions/"+approvedCandidate.ID+"/status",
			fixture.AdminToken,
			map[string]any{"to_status": "approved"},
		)
		require.Equal(t, http.StatusBadRequest, repeatApproval.Code, repeatApproval.Body.String())

		cancelledCandidate := createRequisitionThroughHTTP(t, fixture)

		missingReason := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions/"+cancelledCandidate.ID+"/status",
			fixture.UserToken,
			map[string]any{"to_status": "cancelled"},
		)
		require.Equal(t, http.StatusBadRequest, missingReason.Code, missingReason.Body.String())

		cancelled := performRequisitionRequest(
			t,
			fixture.Requisition.Router,
			http.MethodPost,
			"/api/v1/requisitions/"+cancelledCandidate.ID+"/status",
			fixture.UserToken,
			map[string]any{
				"to_status": "cancelled",
				"reason":    "No longer required",
			},
		)
		require.Equal(t, http.StatusOK, cancelled.Code, cancelled.Body.String())

		var cancelledResponse requisitionhttp.RequisitionResponse
		decodeRequisitionEnvelope(t, cancelled, &cancelledResponse)
		require.Equal(t, "cancelled", cancelledResponse.Status)
		require.Len(t, cancelledResponse.StatusHistory, 2)
		require.Equal(t, "cancelled", cancelledResponse.StatusHistory[1].ToStatus)
		require.Equal(t, fixture.UserID, cancelledResponse.StatusHistory[1].ActedBy)
		require.Equal(t, "staff", cancelledResponse.StatusHistory[1].ActorRole)
	})
}
