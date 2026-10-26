package shareddomain

import "testing"

func Test_NewUsernameType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue UsernameType
		wantErr   error
	}{
		{name: "first name type", value: "first_name", wantValue: FirstName},
		{name: "last name type", value: "last_name", wantValue: LastName},
		{name: "empty type", value: "", wantErr: ErrInvalidUsernameType},
		{name: "unknown type", value: "middle_name", wantErr: ErrInvalidUsernameType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			usernameType, err := NewUsernameType(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewUsernameType() error = %v, want %v", err, tt.wantErr)
				}
				if usernameType != "" {
					t.Errorf("NewUsernameType() = %q, want empty", usernameType)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewUsernameType() unexpected error = %v", err)
			}
			if usernameType != tt.wantValue {
				t.Errorf("NewUsernameType() = %q, want %q", usernameType, tt.wantValue)
			}
		})
	}
}

func Test_UsernameType_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		usernameType UsernameType
		valid        bool
	}{
		{name: "first name is valid", usernameType: FirstName, valid: true},
		{name: "last name is valid", usernameType: LastName, valid: true},
		{name: "empty is invalid", usernameType: UsernameType(""), valid: false},
		{name: "unknown is invalid", usernameType: UsernameType("middle_name"), valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.usernameType.isValid(); got != tt.valid {
				t.Errorf("UsernameType.isValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}
