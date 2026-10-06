package domain

import (
	"strings"
	"testing"
)

func Test_NewUsername(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		usernameType UsernameType
		wantValue    string
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "valid first name",
			value:        "John",
			usernameType: FirstName,
			wantValue:    "John",
			wantErr:      false,
		},
		{
			name:         "valid last name",
			value:        "Doe",
			usernameType: LastName,
			wantValue:    "Doe",
			wantErr:      false,
		},
		{
			name:         "minimum valid length boundary of 2 characters",
			value:        "Al",
			usernameType: FirstName,
			wantValue:    "Al",
			wantErr:      false,
		},
		{
			name:         "maximum valid length boundary of 100 characters",
			value:        strings.Repeat("a", 100),
			usernameType: LastName,
			wantValue:    strings.Repeat("a", 100),
			wantErr:      false,
		},
		{
			name:         "empty value is too short",
			value:        "",
			usernameType: FirstName,
			wantErr:      true,
			expectedErr:  ErrUsernameTooShort(FirstName),
		},
		{
			name:         "single character value is too short",
			value:        "J",
			usernameType: LastName,
			wantErr:      true,
			expectedErr:  ErrUsernameTooShort(LastName),
		},
		{
			name:         "value exceeding 100 characters is too long",
			value:        strings.Repeat("a", 101),
			usernameType: FirstName,
			wantErr:      true,
			expectedErr:  ErrUsernameTooLong(FirstName),
		},
		{
			name:         "invalid username type",
			value:        "John",
			usernameType: UsernameType("middle_name"),
			wantErr:      true,
			expectedErr:  ErrInvalidUsername(UsernameType("middle_name")),
		},
		{
			name:         "empty username type",
			value:        "John",
			usernameType: UsernameType(""),
			wantErr:      true,
			expectedErr:  ErrInvalidUsername(UsernameType("")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, err := NewUsername(tt.value, tt.usernameType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUsername() error = nil, wantErr %v", tt.wantErr)
				}
				if username != nil {
					t.Errorf("NewUsername() = %v, want nil", username)
				}
				if tt.expectedErr != nil && err != nil {
					if err.Error() != tt.expectedErr.Error() {
						t.Errorf("NewUsername() error = %v, want %v", err, tt.expectedErr)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("NewUsername() unexpected error = %v", err)
			}
			if username == nil {
				t.Fatalf("NewUsername() = nil, want valid username")
			}
			if username.Value() != tt.wantValue {
				t.Errorf("NewUsername().Value() = %v, want %v", username.Value(), tt.wantValue)
			}
		})
	}
}
