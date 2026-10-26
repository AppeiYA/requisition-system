package domain

import (
	"strings"
	"testing"
)

func Test_NewCategoryName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   error
	}{
		{
			name:      "valid category name",
			value:     "Office Supplies",
			wantValue: "Office Supplies",
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
			wantErr: ErrCategoryNameEmpty,
		},
		{
			name:    "name too long",
			value:   strings.Repeat("a", 101),
			wantErr: ErrCategoryNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			name, err := NewCategoryName(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewCategoryName() error = %v, want %v", err, tt.wantErr)
				}
				if name != nil {
					t.Errorf("NewCategoryName() = %v, want nil", name)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewCategoryName() unexpected error = %v", err)
			}
			if name == nil {
				t.Fatal("NewCategoryName() = nil, want category name")
			}
			if name.Value() != tt.wantValue {
				t.Errorf("CategoryName.Value() = %q, want %q", name.Value(), tt.wantValue)
			}
		})
	}
}
