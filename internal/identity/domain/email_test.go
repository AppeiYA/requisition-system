package domain

import (
	"testing"
)

func Test_NewEmail(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid email",
			value:     "john.doe@example.com",
			wantValue: "john.doe@example.com",
			wantErr:   false,
		},
		{
			name:  "empty email",
			value: "", wantErr: true,
		},
		{
			name:    "missing @",
			value:   "johndoe.example.com",
			wantErr: true,
		},
		{
			name:    "missing domain",
			value:   "johndoe@",
			wantErr: true,
		},
		{
			name:    "missing local part",
			value:   "@example.com",
			wantErr: true,
		},
		{
			name:    "missing top level domain",
			value:   "johndoe@example",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			value:   "john doe@example.com",
			wantErr: true,
		},
		{
			name:      "email with plus",
			value:     "john+test@example.com",
			wantValue: "john+test@example.com",
			wantErr:   false,
		},
		{
			name:      "email with subdomain",
			value:     "john@mail.example.com",
			wantValue: "john@mail.example.com",
			wantErr:   false,
		},
		{
			name:    "email exceeding maximum length",
			value:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa@example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}

				if email != nil {
					t.Error("expected email to be nil when validation fails")
				}

				return
			}

			if err != nil {
				t.Errorf("expected no error but go %v", err)
			}

			if email == nil {
				t.Error("expected email but got nil")
			}

			if email.Value() != tt.wantValue {
				t.Errorf("expected email: %s but got %s",
					tt.wantValue,
					email.Value(),
				)
			}
		})
	}
}
