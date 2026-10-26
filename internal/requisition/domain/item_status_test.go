package domain

import "testing"

func Test_NewItemStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue ItemStatus
		wantErr   error
	}{
		{name: "active status", value: "active", wantValue: ItemStatusActive},
		{name: "inactive status", value: "inactive", wantValue: ItemStatusInactive},
		{name: "empty status", value: "", wantErr: ErrInvalidItemStatus},
		{name: "unknown status", value: "archived", wantErr: ErrInvalidItemStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, err := NewItemStatus(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewItemStatus() error = %v, want %v", err, tt.wantErr)
				}
				if status != "" {
					t.Errorf("NewItemStatus() = %q, want empty", status)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewItemStatus() unexpected error = %v", err)
			}
			if status != tt.wantValue {
				t.Errorf("NewItemStatus() = %q, want %q", status, tt.wantValue)
			}
			if status.String() != tt.value {
				t.Errorf("ItemStatus.String() = %q, want %q", status.String(), tt.value)
			}
		})
	}
}

func Test_ItemStatus_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  ItemStatus
		isValid bool
	}{
		{name: "active is valid", status: ItemStatusActive, isValid: true},
		{name: "inactive is valid", status: ItemStatusInactive, isValid: true},
		{name: "empty is invalid", status: ItemStatus(""), isValid: false},
		{name: "unknown is invalid", status: ItemStatus("archived"), isValid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.status.IsValid(); got != tt.isValid {
				t.Errorf("ItemStatus.IsValid() = %v, want %v", got, tt.isValid)
			}
		})
	}
}

func Test_ItemStatus_CanTransitionTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		from          ItemStatus
		to            ItemStatus
		canTransition bool
	}{
		{
			name:          "active to inactive is allowed",
			from:          ItemStatusActive,
			to:            ItemStatusInactive,
			canTransition: true,
		},
		{
			name:          "inactive to active is allowed",
			from:          ItemStatusInactive,
			to:            ItemStatusActive,
			canTransition: true,
		},
		{
			name:          "active to active is not allowed",
			from:          ItemStatusActive,
			to:            ItemStatusActive,
			canTransition: false,
		},
		{
			name:          "inactive to inactive is not allowed",
			from:          ItemStatusInactive,
			to:            ItemStatusInactive,
			canTransition: false,
		},
		{
			name:          "unknown status cannot transition",
			from:          ItemStatus("archived"),
			to:            ItemStatusActive,
			canTransition: false,
		},
		{
			name:          "cannot transition to unknown status",
			from:          ItemStatusActive,
			to:            ItemStatus("archived"),
			canTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.from.CanTransitionTo(tt.to); got != tt.canTransition {
				t.Errorf("ItemStatus.CanTransitionTo() = %v, want %v", got, tt.canTransition)
			}
		})
	}
}

func Test_ItemStatus_String(t *testing.T) {
	t.Parallel()

	if got := ItemStatusActive.String(); got != "active" {
		t.Errorf("ItemStatus.String() = %q, want %q", got, "active")
	}
	if got := ItemStatusInactive.String(); got != "inactive" {
		t.Errorf("ItemStatus.String() = %q, want %q", got, "inactive")
	}
}
