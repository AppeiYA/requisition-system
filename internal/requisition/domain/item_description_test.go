package domain

import (
	"strings"
	"testing"
)

func Test_NewItemDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   error
	}{
		{
			name:      "valid description",
			value:     "Business laptop",
			wantValue: "Business laptop",
		},
		{
			name:      "maximum length",
			value:     strings.Repeat("a", 200),
			wantValue: strings.Repeat("a", 200),
		},
		{
			name:    "empty description",
			value:   "",
			wantErr: ErrItemDescriptionCannotBeEmpty,
		},
		{
			name:    "description too long",
			value:   strings.Repeat("a", 201),
			wantErr: ErrItemDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			description, err := NewItemDescription(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewItemDescription() error = %v, want %v", err, tt.wantErr)
				}
				if description != nil {
					t.Errorf("NewItemDescription() = %v, want nil", description)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewItemDescription() unexpected error = %v", err)
			}
			if description == nil {
				t.Fatal("NewItemDescription() = nil, want item description")
			}
			if description.Value() != tt.wantValue {
				t.Errorf("ItemDescription.Value() = %q, want %q", description.Value(), tt.wantValue)
			}
		})
	}
}
