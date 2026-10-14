package domain

import "testing"

func Test_NewPassword(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid password",
			value:   "ValidPass1!",
			wantErr: false,
		},
		{
			name:    "empty password",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid password - too short",
			value:   "Short1!",
			wantErr: true,
		},
		{
			name:    "invalid password - no uppercase",
			value:   "lowercase1!",
			wantErr: true,
		},
		{
			name:    "invalid password - no lowercase",
			value:   "UPPERCASE1!",
			wantErr: true,
		},
		{
			name:    "invalid password - no number",
			value:   "NoNumber!",
			wantErr: true,
		},
		{
			name:    "invalid password - no special character",
			value:   "NoSpecial1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := NewPassword(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPassword() expected error, got nil")
				}

				if password != nil {
					t.Errorf("NewPassword() expected nil password, got %v", password)
				}
			} else {
				if err != nil {
					t.Errorf("NewPassword() unexpected error: %v", err)
				}

				if password == nil {
					t.Errorf("NewPassword() expected non-nil password, got nil")
				}
			}
		})
	}
}
