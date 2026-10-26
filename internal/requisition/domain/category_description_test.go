package domain

import (
	"strings"
	"testing"
)

func Test_NewCategoryDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   error
	}{
		{
			name:      "valid description",
			value:     "General office supplies",
			wantValue: "General office supplies",
		},
		{
			name:      "empty description is allowed",
			value:     "",
			wantValue: "",
		},
		{
			name:      "maximum length",
			value:     strings.Repeat("a", 200),
			wantValue: strings.Repeat("a", 200),
		},
		{
			name:    "description too long",
			value:   strings.Repeat("a", 201),
			wantErr: ErrCategoryDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description, err := NewCategoryDescription(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewCategoryDescription() error = %v, want %v", err, tt.wantErr)
				}
				if description != nil {
					t.Errorf("NewCategoryDescription() = %v, want nil", description)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewCategoryDescription() unexpected error = %v", err)
			}
			if description == nil {
				t.Fatal("NewCategoryDescription() = nil, want category description")
			}
			if description.Value() != tt.wantValue {
				t.Errorf("CategoryDescription.Value() = %q, want %q", description.Value(), tt.wantValue)
			}
		})
	}
}
