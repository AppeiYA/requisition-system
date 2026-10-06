package domain

import "testing"

func Test_NewUsernameType(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid username type first_name",
			value:     "first_name",
			wantValue: "first_name",
			wantErr:   false,
		},
		{
			name:      "valid username type last_name",
			value:     "last_name",
			wantValue: "last_name",
			wantErr:   false,
		},
		{
			name:    "empty username type",
			value:   "",
			wantErr: true,
		},
		{
			name:    "invalid username type",
			value:   "middle_name",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usernameType, err := NewUsernameType(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUsernameType() error = %v, wantErr %v", err, tt.wantErr)
				}
				if usernameType != "" {
					t.Errorf("NewUsernameType() = %v, want empty string", usernameType)
				}
			} else {
				if err != nil {
					t.Errorf("NewUsernameType() error = %v, wantErr %v", err, tt.wantErr)
				}
				if usernameType != UsernameType(tt.wantValue) {
					t.Errorf("NewUsernameType() = %v, want %v", usernameType, tt.wantValue)
				}
			}
		})
	}
}
