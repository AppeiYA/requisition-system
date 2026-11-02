package requisitionhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/authcontext"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
	"github.com/gin-gonic/gin"
)

type stubGetRequisition struct {
	res *domain.Requisition
	err error
	got requisitionin.GetRequisitionQuery
}

func (s *stubGetRequisition) Execute(_ context.Context, query requisitionin.GetRequisitionQuery) (*domain.Requisition, error) {
	s.got = query
	return s.res, s.err
}

type stubTransitionRequisition struct {
	res *domain.Requisition
	err error
	got requisitionin.TransitionRequisitionCommand
}

func (s *stubTransitionRequisition) Execute(_ context.Context, cmd requisitionin.TransitionRequisitionCommand) (*domain.Requisition, error) {
	s.got = cmd
	return s.res, s.err
}

func newRequisitionActorRouter(
	userID string,
	role string,
	get requisitionin.GetRequisitionUsecaseInt,
	transition requisitionin.TransitionRequisitionUsecaseInt,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	port := &requisitionin.RequisitionPortIn{
		CreateRequisitionUsecase:      &stubCreateRequisition{},
		ListRequisitionsByUserUsecase: &stubListRequisitionsByUser{},
		ListRequisitionsUsecase:       &stubListRequisitions{},
		GetRequisitionUsecase:         get,
		TransitionRequisitionUsecase:  transition,
	}
	auth := func(c *gin.Context) {
		if userID != "" {
			c.Set(authcontext.UserIDKey, userID)
			c.Set(authcontext.UserRoleKey, role)
		}
		c.Next()
	}
	admin := func(c *gin.Context) {
		if role == "admin" {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusForbidden)
	}
	NewRequisitionHandler(port).Register(router.Group("/api/v1"), auth, admin)
	return router
}

func transitionedRequisition(t *testing.T, status domain.RequisitionStatus, reason string) *domain.Requisition {
	t.Helper()

	requisition := validRequisition()
	if _, err := requisition.ApplyStatusTransition(
		"rsh_2", status, "usr-admin", domain.ActorRoleAdmin, reason, time.Now().UTC(),
	); err != nil {
		t.Fatalf("ApplyStatusTransition() unexpected error = %v", err)
	}
	return requisition
}

func Test_GetRequisition_OwnerSuccess(t *testing.T) {
	t.Parallel()

	get := &stubGetRequisition{res: validRequisition()}
	router := newRequisitionActorRouter("usr-123", "user", get, &stubTransitionRequisition{})

	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/req_123", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if get.got.RequisitionID != "req_123" || get.got.ActorID != "usr-123" || get.got.ActorRole != "user" {
		t.Fatalf("query = %+v", get.got)
	}
	mapped := decodeCreatedRequisition(t, rec.Body.Bytes())
	if len(mapped.Lines) != 2 || mapped.TotalEstimatedAmount != 11500 {
		t.Fatalf("aggregate response = %+v", mapped)
	}
	if len(mapped.StatusHistory) != 1 || mapped.StatusHistory[0].ToStatus != "pending" {
		t.Fatalf("status history = %+v", mapped.StatusHistory)
	}
}

func Test_GetRequisition_AdminSuccess(t *testing.T) {
	t.Parallel()

	get := &stubGetRequisition{res: validRequisition()}
	router := newRequisitionActorRouter("usr-admin", "admin", get, &stubTransitionRequisition{})

	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/req_123", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if get.got.ActorRole != "admin" {
		t.Fatalf("actor role = %q, want admin", get.got.ActorRole)
	}
}

func Test_GetRequisition_Forbidden(t *testing.T) {
	t.Parallel()

	get := &stubGetRequisition{err: customerrors.NewForbiddenError("you are not allowed to access this requisition")}
	router := newRequisitionActorRouter("usr-other", "user", get, &stubTransitionRequisition{})

	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/req_123", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", rec.Code, rec.Body.String())
	}
}

func Test_GetRequisition_NotFound(t *testing.T) {
	t.Parallel()

	get := &stubGetRequisition{err: customerrors.NewNotFoundError("requisition not found")}
	router := newRequisitionActorRouter("usr-123", "user", get, &stubTransitionRequisition{})

	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/req_123", "")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func Test_GetRequisition_InvalidID(t *testing.T) {
	t.Parallel()

	get := &stubGetRequisition{res: validRequisition()}
	router := newRequisitionActorRouter("usr-123", "user", get, &stubTransitionRequisition{})

	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/invalid", "")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	assertValidationField(t, rec.Body.Bytes(), "id")
}

func Test_GetRequisition_Unauthenticated(t *testing.T) {
	t.Parallel()

	router := newRequisitionActorRouter("", "", &stubGetRequisition{}, &stubTransitionRequisition{})

	rec := doReqJSON(t, router, http.MethodGet, "/api/v1/requisitions/req_123", "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body.String())
	}
}

func Test_TransitionRequisition_Approved(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{res: transitionedRequisition(t, domain.RequisitionStatusApproved, "")}
	router := newRequisitionActorRouter("usr-admin", "admin", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status", `{"toStatus":"approved"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if transition.got.RequisitionID != "req_123" || transition.got.ActorID != "usr-admin" {
		t.Fatalf("command = %+v", transition.got)
	}
	if transition.got.ActorRole != "admin" || transition.got.ToStatus != "approved" {
		t.Fatalf("command = %+v", transition.got)
	}

	mapped := decodeCreatedRequisition(t, rec.Body.Bytes())
	if mapped.Status != "approved" {
		t.Fatalf("status = %q, want approved", mapped.Status)
	}
	if len(mapped.StatusHistory) != 2 {
		t.Fatalf("status history = %+v", mapped.StatusHistory)
	}
	last := mapped.StatusHistory[1]
	if last.Sequence != 2 || last.ToStatus != "approved" || last.FromStatus == nil || *last.FromStatus != "pending" {
		t.Fatalf("last history = %+v", last)
	}
	if last.ActedBy != "usr-admin" || last.ActorRole != "admin" || last.CreatedAt.IsZero() {
		t.Fatalf("last history = %+v", last)
	}
}

func Test_TransitionRequisition_CancelledWithReason(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{
		res: transitionedRequisition(t, domain.RequisitionStatusCancelled, "No longer required"),
	}
	router := newRequisitionActorRouter("usr-123", "user", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status",
		`{"to_status":"cancelled","reason":"No longer required"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
	}
	if transition.got.ToStatus != "cancelled" || transition.got.Reason != "No longer required" {
		t.Fatalf("command = %+v", transition.got)
	}
	mapped := decodeCreatedRequisition(t, rec.Body.Bytes())
	if mapped.Status != "cancelled" {
		t.Fatalf("status = %q, want cancelled", mapped.Status)
	}
	if mapped.StatusHistory[1].Reason != "No longer required" {
		t.Fatalf("history reason = %q", mapped.StatusHistory[1].Reason)
	}
}

func Test_TransitionRequisition_RequiresReason(t *testing.T) {
	t.Parallel()

	for _, status := range []string{"rejected", "cancelled"} {
		status := status
		t.Run(status, func(t *testing.T) {
			t.Parallel()
			transition := &stubTransitionRequisition{}
			router := newRequisitionActorRouter("usr-admin", "admin", &stubGetRequisition{}, transition)

			rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status",
				`{"to_status":"`+status+`"}`)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
			}
			assertValidationField(t, rec.Body.Bytes(), "reason")
			if transition.got.RequisitionID != "" {
				t.Fatal("usecase should not be called when validation fails")
			}
		})
	}
}

func Test_TransitionRequisition_RejectsInvalidStatus(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{}
	router := newRequisitionActorRouter("usr-admin", "admin", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status", `{"to_status":"unknown"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	assertValidationField(t, rec.Body.Bytes(), "to_status")
}

func Test_TransitionRequisition_InvalidID(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{}
	router := newRequisitionActorRouter("usr-admin", "admin", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/invalid/status", `{"to_status":"approved"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	assertValidationField(t, rec.Body.Bytes(), "id")
}

func Test_TransitionRequisition_InvalidTransitionError(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{
		err: domain.ErrCannotTransitionToStatus("pending", "completed"),
	}
	router := newRequisitionActorRouter("usr-admin", "admin", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status", `{"to_status":"completed"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Message != "cannot transition from pending to completed" {
		t.Fatalf("message = %q", body.Message)
	}
}

func Test_TransitionRequisition_Forbidden(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{err: customerrors.NewForbiddenError("not authorized")}
	router := newRequisitionActorRouter("usr-other", "user", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status",
		`{"to_status":"cancelled","reason":"x"}`)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", rec.Code, rec.Body.String())
	}
}

func Test_TransitionRequisition_NotFound(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{err: customerrors.NewNotFoundError("requisition not found")}
	router := newRequisitionActorRouter("usr-admin", "admin", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status", `{"to_status":"approved"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404: %s", rec.Code, rec.Body.String())
	}
}

func Test_TransitionRequisition_Conflict(t *testing.T) {
	t.Parallel()

	transition := &stubTransitionRequisition{err: customerrors.NewConflictError("concurrent status change")}
	router := newRequisitionActorRouter("usr-admin", "admin", &stubGetRequisition{}, transition)

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status", `{"to_status":"approved"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", rec.Code, rec.Body.String())
	}
}

func Test_TransitionRequisition_Unauthenticated(t *testing.T) {
	t.Parallel()

	router := newRequisitionActorRouter("", "", &stubGetRequisition{}, &stubTransitionRequisition{})

	rec := doReqJSON(t, router, http.MethodPost, "/api/v1/requisitions/req_123/status", `{"to_status":"approved"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401: %s", rec.Code, rec.Body.String())
	}
}
