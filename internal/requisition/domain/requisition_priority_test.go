package domain

import "testing"

func Test_NewRequisitionPriority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue RequisitionPriority
		wantErr   error
	}{
		{name: "low priority", value: "low", wantValue: PriorityLow},
		{name: "normal priority", value: "normal", wantValue: PriorityNormal},
		{name: "medium priority", value: "medium", wantValue: PriorityMedium},
		{name: "high priority", value: "high", wantValue: PriorityHigh},
		{name: "urgent priority", value: "urgent", wantValue: PriorityUrgent},
		{name: "empty priority", value: "", wantErr: ErrInvalidRequisitionPriority},
		{name: "unknown priority", value: "critical", wantErr: ErrInvalidRequisitionPriority},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			priority, err := NewRequisitionPriority(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewRequisitionPriority() error = %v, want %v", err, tt.wantErr)
				}
				if priority != "" {
					t.Errorf("NewRequisitionPriority() = %q, want empty", priority)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewRequisitionPriority() unexpected error = %v", err)
			}
			if priority != tt.wantValue {
				t.Errorf("NewRequisitionPriority() = %q, want %q", priority, tt.wantValue)
			}
			if priority.String() != tt.value {
				t.Errorf("RequisitionPriority.String() = %q, want %q", priority.String(), tt.value)
			}
		})
	}
}

func Test_RequisitionPriority_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priority RequisitionPriority
		isValid  bool
	}{
		{name: "low is valid", priority: PriorityLow, isValid: true},
		{name: "normal is valid", priority: PriorityNormal, isValid: true},
		{name: "medium is valid", priority: PriorityMedium, isValid: true},
		{name: "high is valid", priority: PriorityHigh, isValid: true},
		{name: "urgent is valid", priority: PriorityUrgent, isValid: true},
		{name: "empty is invalid", priority: RequisitionPriority(""), isValid: false},
		{name: "unknown is invalid", priority: RequisitionPriority("critical"), isValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.priority.IsValid(); got != tt.isValid {
				t.Errorf("RequisitionPriority.IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}
