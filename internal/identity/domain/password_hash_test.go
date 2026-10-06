package domain

import "testing"

func Test_NewPasswordHash(t *testing.T) {
	test := []struct {
		name      string
		value     string
		wantValue string
		wantErr   bool
	}{
		{
			name:      "valid password hash",
			value:     "hashed_password",
			wantValue: "hashed_password",
			wantErr:   false,
		},
		{
			name:    "empty password hash",
			value:   "",
			wantErr: true,
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := NewPasswordHash(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				}
				if hash != nil {
					t.Errorf("NewPasswordHash() = %v, want nil", hash)
				}
			} else {
				if err != nil {
					t.Errorf("NewPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
				}
				if hash.Value() != tt.wantValue {
					t.Errorf("NewPasswordHash() = %v, want %v", hash, tt.wantValue)
				}
			}
		})
	}
}
