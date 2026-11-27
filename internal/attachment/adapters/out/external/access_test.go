package external

import (
	"context"
	"testing"

	attachmentout "github.com/AppeiYA/requisition-system/internal/attachment/ports/out"
	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

type stubGetRequisition struct {
	got requisitionin.GetRequisitionQuery
	err error
}

func (s *stubGetRequisition) Execute(_ context.Context, q requisitionin.GetRequisitionQuery) (*requisitiondomain.Requisition, error) {
	s.got = q
	return nil, s.err
}

type stubGetRequisitionItem struct {
	got requisitionin.GetRequisitionItemQuery
	err error
}

func (s *stubGetRequisitionItem) Execute(_ context.Context, q requisitionin.GetRequisitionItemQuery) (*requisitiondomain.Requisition, error) {
	s.got = q
	return nil, s.err
}

func TestCanAccessDelegatesToRequisitionPortIn(t *testing.T) {
	for _, tc := range []struct {
		refName string
		refID   string
	}{
		{refName: "requisition", refID: "req-0123456789abcdef"},
		{refName: "requisition_item", refID: "reqitem-0123456789abcdef"},
	} {
		t.Run(tc.refName, func(t *testing.T) {
			get := &stubGetRequisition{}
			getItem := &stubGetRequisitionItem{}
			checker := NewRequisitionAccessChecker(&requisitionin.RequisitionPortIn{
				GetRequisitionUsecase:     get,
				GetRequisitionItemUsecase: getItem,
			})

			if err := checker.CanAccess(context.Background(), tc.refName, tc.refID, "usr-1", "employee", attachmentout.ActionUpload); err != nil {
				t.Fatalf("CanAccess() error = %v", err)
			}
			if tc.refName == "requisition" && get.got.RequisitionID != tc.refID {
				t.Errorf("RequisitionID = %q, want %q", get.got.RequisitionID, tc.refID)
			}
			if tc.refName == "requisition_item" && getItem.got.RequisitionItemID != tc.refID {
				t.Errorf("RequisitionItemID = %q, want %q", getItem.got.RequisitionItemID, tc.refID)
			}
		})
	}
}

func TestCanAccessPropagatesParentErrors(t *testing.T) {
	denied := customerrors.NewForbiddenError("denied")
	checker := NewRequisitionAccessChecker(&requisitionin.RequisitionPortIn{
		GetRequisitionUsecase: &stubGetRequisition{err: denied},
	})

	err := checker.CanAccess(context.Background(), "requisition", "req-1", "usr-1", "employee", attachmentout.ActionRead)
	if err != denied {
		t.Errorf("error = %v, want %v", err, denied)
	}
}

func TestCanAccessRejectsUnknownRef(t *testing.T) {
	checker := NewRequisitionAccessChecker(&requisitionin.RequisitionPortIn{})
	if err := checker.CanAccess(context.Background(), "invoice", "inv-1", "usr-1", "employee", attachmentout.ActionRead); err == nil {
		t.Error("expected an error for an unknown ref name")
	}
}
