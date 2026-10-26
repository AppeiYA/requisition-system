package domain

import (
	"fmt"
	"slices"
	"testing"
)

func Test_NewRequisitionStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue RequisitionStatus
		wantErr   error
	}{
		{name: "pending status", value: "pending", wantValue: RequisitionStatusPending},
		{name: "approved status", value: "approved", wantValue: RequisitionStatusApproved},
		{name: "rejected status", value: "rejected", wantValue: RequisitionStatusRejected},
		{name: "completed status", value: "completed", wantValue: RequisitionStatusCompleted},
		{name: "cancelled status", value: "cancelled", wantValue: RequisitionStatusCancelled},
		{name: "in fulfillment status", value: "in_fulfillment", wantValue: RequisitionStatusInFulfillment},
		{name: "empty status", value: "", wantErr: ErrInvalidRequisitionStatus},
		{name: "unknown status", value: "draft", wantErr: ErrInvalidRequisitionStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := NewRequisitionStatus(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewRequisitionStatus() error = %v, want %v", err, tt.wantErr)
				}
				if status != "" {
					t.Errorf("NewRequisitionStatus() = %q, want empty", status)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewRequisitionStatus() unexpected error = %v", err)
			}
			if status != tt.wantValue {
				t.Errorf("NewRequisitionStatus() = %q, want %q", status, tt.wantValue)
			}
			if status.String() != tt.value {
				t.Errorf("RequisitionStatus.String() = %q, want %q", status.String(), tt.value)
			}
		})
	}
}

func Test_RequisitionStatus_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  RequisitionStatus
		isValid bool
	}{
		{name: "pending is valid", status: RequisitionStatusPending, isValid: true},
		{name: "approved is valid", status: RequisitionStatusApproved, isValid: true},
		{name: "rejected is valid", status: RequisitionStatusRejected, isValid: true},
		{name: "completed is valid", status: RequisitionStatusCompleted, isValid: true},
		{name: "cancelled is valid", status: RequisitionStatusCancelled, isValid: true},
		{name: "in fulfillment is valid", status: RequisitionStatusInFulfillment, isValid: true},
		{name: "empty is invalid", status: RequisitionStatus(""), isValid: false},
		{name: "unknown is invalid", status: RequisitionStatus("draft"), isValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.IsValid(); got != tt.isValid {
				t.Errorf("RequisitionStatus.IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func Test_RequisitionStatus_CanTransitionTo(t *testing.T) {
	t.Parallel()

	allowedTransitions := map[RequisitionStatus][]RequisitionStatus{
		RequisitionStatusPending: {
			RequisitionStatusApproved,
			RequisitionStatusRejected,
			RequisitionStatusCancelled,
		},
		RequisitionStatusApproved: {
			RequisitionStatusInFulfillment,
			RequisitionStatusCancelled,
		},
		RequisitionStatusInFulfillment: {
			RequisitionStatusCompleted,
			RequisitionStatusCancelled,
		},
	}

	statuses := []RequisitionStatus{
		RequisitionStatusPending,
		RequisitionStatusApproved,
		RequisitionStatusRejected,
		RequisitionStatusCompleted,
		RequisitionStatusCancelled,
		RequisitionStatusInFulfillment,
	}

	for _, from := range statuses {
		for _, to := range statuses {
			t.Run(fmt.Sprintf("%s_to_%s", from, to), func(t *testing.T) {
				t.Parallel()

				want := slices.Contains(allowedTransitions[from], to)
				if got := from.CanTransitionTo(to); got != want {
					t.Errorf("RequisitionStatus.CanTransitionTo() from %q to %q = %v, want %v", from, to, got, want)
				}
			})
		}
	}

	t.Run("unknown status cannot transition", func(t *testing.T) {
		t.Parallel()

		from := RequisitionStatus("draft")
		if from.CanTransitionTo(RequisitionStatusPending) {
			t.Error("unknown status should not transition")
		}
	})

	t.Run("cannot transition to unknown status", func(t *testing.T) {
		t.Parallel()

		if RequisitionStatusPending.CanTransitionTo(RequisitionStatus("draft")) {
			t.Error("should not transition to an unknown status")
		}
	})
}

func Test_RequisitionStatus_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status RequisitionStatus
		want   string
	}{
		{status: RequisitionStatusPending, want: "pending"},
		{status: RequisitionStatusApproved, want: "approved"},
		{status: RequisitionStatusRejected, want: "rejected"},
		{status: RequisitionStatusCompleted, want: "completed"},
		{status: RequisitionStatusCancelled, want: "cancelled"},
		{status: RequisitionStatusInFulfillment, want: "in_fulfillment"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.String(); got != tt.want {
				t.Errorf("RequisitionStatus.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
