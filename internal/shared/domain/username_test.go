package shareddomain

import (
	"strings"
	"testing"
)

func Test_NewUsername(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		value        string
		usernameType UsernameType
		wantValue    string
		wantErr      string
	}{
		{
			name:         "valid first name",
			value:        "John",
			usernameType: FirstName,
			wantValue:    "John",
		},
		{
			name:         "valid last name",
			value:        "Doe",
			usernameType: LastName,
			wantValue:    "Doe",
		},
		{
			name:         "minimum length",
			value:        "Al",
			usernameType: FirstName,
			wantValue:    "Al",
		},
		{
			name:         "maximum length",
			value:        strings.Repeat("a", 100),
			usernameType: LastName,
			wantValue:    strings.Repeat("a", 100),
		},
		{
			name:         "empty value",
			value:        "",
			usernameType: FirstName,
			wantErr:      ErrUsernameTooShort(FirstName).Error(),
		},
		{
			name:         "single character value",
			value:        "J",
			usernameType: LastName,
			wantErr:      ErrUsernameTooShort(LastName).Error(),
		},
		{
			name:         "value too long",
			value:        strings.Repeat("a", 101),
			usernameType: FirstName,
			wantErr:      ErrUsernameTooLong(FirstName).Error(),
		},
		{
			name:         "invalid username type",
			value:        "John",
			usernameType: UsernameType("middle_name"),
			wantErr:      ErrInvalidUsername(UsernameType("middle_name")).Error(),
		},
		{
			name:         "empty username type",
			value:        "John",
			usernameType: UsernameType(""),
			wantErr:      ErrInvalidUsername(UsernameType("")).Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			username, err := NewUsername(tt.value, tt.usernameType)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("NewUsername() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("NewUsername() error = %q, want %q", err.Error(), tt.wantErr)
				}
				if username != nil {
					t.Errorf("NewUsername() = %v, want nil", username)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewUsername() unexpected error = %v", err)
			}
			if username == nil {
				t.Fatal("NewUsername() = nil, want username")
			}
			if username.Value() != tt.wantValue {
				t.Errorf("Username.Value() = %q, want %q", username.Value(), tt.wantValue)
			}
		})
	}
}
