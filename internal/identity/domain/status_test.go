package domain

import "testing"

func Test_NewStatus(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid status active",
			value:     "active",
			wantValue: "active",
			wantErr:   false,
		},
		{
			name:      "valid status suspended",
			value:     "suspended",
			wantValue: "suspended",
			wantErr:   false,
		},
		{
			name:    "invalid status",
			value:   "invalid_status",
			wantErr: true,
		},
		{
			name:    "empty status",
			value:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewStatus(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewStatus() error = %v, wantErr %v", err, tt.wantErr)
				}
				if status != "" {
					t.Errorf("NewStatus() = %v, want empty string", status)
				}
			} else {
				if err != nil {
					t.Errorf("NewStatus() error = %v, wantErr %v", err, tt.wantErr)
				}
				if status.Value() != tt.wantValue {
					t.Errorf("NewStatus() = %v, want %v", status, tt.wantValue)
				}
			}
		})
	}
}

func Test_Status_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name          string
		from          Status
		to            Status
		canTransition bool
	}{
		{
			name:          "active to suspended is allowed",
			from:          ACTIVE,
			to:            SUSPENDED,
			canTransition: true,
		},
		{
			name:          "suspended to active is allowed",
			from:          SUSPENDED,
			to:            ACTIVE,
			canTransition: true,
		},
		{
			name:          "active to active is not allowed",
			from:          ACTIVE,
			to:            ACTIVE,
			canTransition: false,
		},
		{
			name:          "suspended to suspended is not allowed",
			from:          SUSPENDED,
			to:            SUSPENDED,
			canTransition: false,
		},
		{
			name:          "unregistered status cannot transition",
			from:          Status("unknown"),
			to:            ACTIVE,
			canTransition: false,
		},
		{
			name:          "active to unknown status is not allowed",
			from:          ACTIVE,
			to:            Status("unknown"),
			canTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.canTransition {
				t.Errorf("CanTransitionTo() from %v to %v = %v, want %v", tt.from, tt.to, got, tt.canTransition)
			}
		})
	}
}

func Test_Status_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		status  Status
		isValid bool
	}{
		{
			name:    "active is valid",
			status:  ACTIVE,
			isValid: true,
		},
		{
			name:    "suspended is valid",
			status:  SUSPENDED,
			isValid: true,
		},
		{
			name:    "empty status is invalid",
			status:  Status(""),
			isValid: false,
		},
		{
			name:    "unknown status is invalid",
			status:  Status("pending"),
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.IsValid()
			if got != tt.isValid {
				t.Errorf("IsValid() for %v = %v, want %v", tt.status, got, tt.isValid)
			}
		})
	}
}
