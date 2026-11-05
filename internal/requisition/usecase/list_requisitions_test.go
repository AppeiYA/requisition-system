package usecase

import (
	"context"
	"errors"
	"testing"

	requisitiondomain "github.com/AppeiYA/requisition-system/internal/requisition/domain"
	requisitionin "github.com/AppeiYA/requisition-system/internal/requisition/ports/in"
)

func testSummary() *requisitiondomain.RequisitionSummary {
	return &requisitiondomain.RequisitionSummary{
		ID:                   "req-1",
		ReferenceCode:        "REQ-2026-10001",
		RequestedBy:          "usr-1",
		DepartmentID:         "dpt-1",
		Status:               "pending",
		Priority:             "medium",
		Justification:        "Office setup",
		TotalEstimatedAmount: 12.5,
	}
}

func TestListRequisitionsUsecaseReturnsSummariesAndClampsPagination(t *testing.T) {
	t.Parallel()

	repo := &createRequisitionRepo{summaries: []*requisitiondomain.RequisitionSummary{testSummary()}}
	uc := NewListRequisitionsUsecase(repo)

	got, err := uc.Execute(context.Background(), requisitionin.ListRequisitionsQuery{Limit: 500, Offset: -1})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(got) != 1 || got[0].ID != "req-1" {
		t.Fatalf("Execute() = %+v, want one summary", got)
	}
	if repo.summaryLimit != 100 || repo.summaryOffset != 0 {
		t.Fatalf("pagination = (%d, %d), want (100, 0)", repo.summaryLimit, repo.summaryOffset)
	}
}

func TestListRequisitionsUsecaseReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := &createRequisitionRepo{summaries: []*requisitiondomain.RequisitionSummary{testSummary()}}

	_, err := NewListRequisitionsUsecase(repo).Execute(ctx, requisitionin.ListRequisitionsQuery{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
	if repo.listCalled {
		t.Fatal("repository should not be called when the context is cancelled")
	}
}

func TestListRequisitionsByUserUsecaseReturnsSummariesAndDelegatesUser(t *testing.T) {
	t.Parallel()

	repo := &createRequisitionRepo{userSummaries: []*requisitiondomain.RequisitionSummary{testSummary()}}
	uc := NewListRequisitionsByUserUsecase(repo, nil)

	got, err := uc.Execute(context.Background(), requisitionin.ListRequisitionsByUserQuery{
		UserID: "usr-123", Limit: 0, Offset: 9,
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if len(got) != 1 || got[0].RequestedBy != "usr-1" {
		t.Fatalf("Execute() = %+v, want one summary", got)
	}
	if repo.summaryUserID != "usr-123" {
		t.Fatalf("user filter = %q, want usr-123", repo.summaryUserID)
	}
	if repo.summaryLimit != 100 || repo.summaryOffset != 9 {
		t.Fatalf("pagination = (%d, %d), want (100, 9)", repo.summaryLimit, repo.summaryOffset)
	}
}

func TestListRequisitionsByUserUsecaseReturnsContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := &createRequisitionRepo{userSummaries: []*requisitiondomain.RequisitionSummary{testSummary()}}

	_, err := NewListRequisitionsByUserUsecase(repo, nil).Execute(ctx, requisitionin.ListRequisitionsByUserQuery{
		UserID: "usr-123",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Execute() error = %v, want context.Canceled", err)
	}
	if repo.userListCalled {
		t.Fatal("repository should not be called when the context is cancelled")
	}
}
