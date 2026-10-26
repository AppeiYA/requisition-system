package domain

import (
	"strings"
	"testing"
)

func Test_NewItemName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   error
	}{
		{
			name:      "valid item name",
			value:     "Laptop",
			wantValue: "Laptop",
		},
		{
			name:      "whitespace only is allowed",
			value:     " ",
			wantValue: " ",
		},
		{
			name:      "maximum length",
			value:     strings.Repeat("a", 100),
			wantValue: strings.Repeat("a", 100),
		},
		{
			name:    "empty name",
			value:   "",
			wantErr: ErrItemNameCannotBeEmpty,
		},
		{
			name:    "name too long",
			value:   strings.Repeat("a", 101),
			wantErr: ErrItemNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := NewItemName(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewItemName() error = %v, want %v", err, tt.wantErr)
				}
				if name != nil {
					t.Errorf("NewItemName() = %v, want nil", name)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewItemName() unexpected error = %v", err)
			}
			if name == nil {
				t.Fatal("NewItemName() = nil, want item name")
			}
			if name.Value() != tt.wantValue {
				t.Errorf("ItemName.Value() = %q, want %q", name.Value(), tt.wantValue)
			}
		})
	}
}
