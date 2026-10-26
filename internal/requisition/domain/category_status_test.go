package domain

import "testing"

func Test_NewCategoryStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   error
	}{
		{name: "active status", value: "active", wantValue: "active"},
		{name: "inactive status", value: "inactive", wantValue: "inactive"},
		{name: "empty status", value: "", wantErr: ErrInvalidCategoryStatus},
		{name: "unknown status", value: "archived", wantErr: ErrInvalidCategoryStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := NewCategoryStatus(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewCategoryStatus() error = %v, want %v", err, tt.wantErr)
				}
				if status != nil {
					t.Errorf("NewCategoryStatus() = %v, want nil", status)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewCategoryStatus() unexpected error = %v", err)
			}
			if status == nil {
				t.Fatal("NewCategoryStatus() = nil, want category status")
			}
			if status.String() != tt.wantValue {
				t.Errorf("CategoryStatus.String() = %q, want %q", status.String(), tt.wantValue)
			}
		})
	}
}

func Test_CategoryStatus_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  CategoryStatus
		isValid bool
	}{
		{name: "active is valid", status: CategoryStatusActive, isValid: true},
		{name: "inactive is valid", status: CategoryStatusInactive, isValid: true},
		{name: "empty is invalid", status: CategoryStatus(""), isValid: false},
		{name: "unknown is invalid", status: CategoryStatus("archived"), isValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.IsValid(); got != tt.isValid {
				t.Errorf("CategoryStatus.IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func Test_CategoryStatus_CanTransitionTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		from          CategoryStatus
		to            CategoryStatus
		canTransition bool
	}{
		{
			name:          "active to inactive is allowed",
			from:          CategoryStatusActive,
			to:            CategoryStatusInactive,
			canTransition: true,
		},
		{
			name:          "inactive to active is allowed",
			from:          CategoryStatusInactive,
			to:            CategoryStatusActive,
			canTransition: true,
		},
		{
			name:          "active to active is not allowed",
			from:          CategoryStatusActive,
			to:            CategoryStatusActive,
			canTransition: false,
		},
		{
			name:          "inactive to inactive is not allowed",
			from:          CategoryStatusInactive,
			to:            CategoryStatusInactive,
			canTransition: false,
		},
		{
			name:          "unknown status cannot transition",
			from:          CategoryStatus("archived"),
			to:            CategoryStatusActive,
			canTransition: false,
		},
		{
			name:          "cannot transition to unknown status",
			from:          CategoryStatusActive,
			to:            CategoryStatus("archived"),
			canTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.from.CanTransitionTo(tt.to); got != tt.canTransition {
				t.Errorf("CategoryStatus.CanTransitionTo() = %v, want %v", got, tt.canTransition)
			}
		})
	}
}

func Test_CategoryStatus_String(t *testing.T) {
	t.Parallel()

	if got := CategoryStatusActive.String(); got != "active" {
		t.Errorf("CategoryStatus.String() = %q, want %q", got, "active")
	}
	if got := CategoryStatusInactive.String(); got != "inactive" {
		t.Errorf("CategoryStatus.String() = %q, want %q", got, "inactive")
	}
}
