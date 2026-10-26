package shareddomain

import "testing"

func Test_NewRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue Role
		wantErr   error
	}{
		{name: "admin role", value: "admin", wantValue: RoleAdmin},
		{name: "user role", value: "user", wantValue: RoleUser},
		{name: "empty role", value: "", wantErr: ErrInvalidRole},
		{name: "unknown role", value: "superuser", wantErr: ErrInvalidRole},
		{name: "uppercase role", value: "ADMIN", wantErr: ErrInvalidRole},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			role, err := NewRole(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewRole() error = %v, want %v", err, tt.wantErr)
				}
				if role != "" {
					t.Errorf("NewRole() = %q, want empty", role)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewRole() unexpected error = %v", err)
			}
			if role != tt.wantValue {
				t.Errorf("NewRole() = %q, want %q", role, tt.wantValue)
			}
			if role.String() != tt.value {
				t.Errorf("Role.String() = %q, want %q", role.String(), tt.value)
			}
		})
	}
}

func Test_IsValidRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		role  string
		valid bool
	}{
		{name: "admin is valid", role: "admin", valid: true},
		{name: "user is valid", role: "user", valid: true},
		{name: "empty is invalid", role: "", valid: false},
		{name: "unknown is invalid", role: "superuser", valid: false},
		{name: "uppercase is invalid", role: "ADMIN", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsValidRole(tt.role); got != tt.valid {
				t.Errorf("IsValidRole() = %v, want %v", got, tt.valid)
			}
		})
	}
}
