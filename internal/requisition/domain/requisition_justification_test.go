package domain

import (
	"strings"
	"testing"
)

func Test_NewRequisitionJustification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   error
	}{
		{
			name:      "valid justification",
			value:     "Required for the new hire",
			wantValue: "Required for the new hire",
		},
		{
			name:      "whitespace only is allowed",
			value:     " ",
			wantValue: " ",
		},
		{
			name:      "maximum length",
			value:     strings.Repeat("a", 500),
			wantValue: strings.Repeat("a", 500),
		},
		{
			name:    "empty justification",
			value:   "",
			wantErr: ErrInvalidRequisitionJustification,
		},
		{
			name:    "justification too long",
			value:   strings.Repeat("a", 501),
			wantErr: ErrRequisitionJustificationTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			justification, err := NewRequisitionJustification(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewRequisitionJustification() error = %v, want %v", err, tt.wantErr)
				}
				if justification != nil {
					t.Errorf("NewRequisitionJustification() = %v, want nil", justification)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewRequisitionJustification() unexpected error = %v", err)
			}
			if justification == nil {
				t.Fatal("NewRequisitionJustification() = nil, want justification")
			}
			if justification.Value() != tt.wantValue {
				t.Errorf("RequisitionJustification.Value() = %q, want %q", justification.Value(), tt.wantValue)
			}
		})
	}
}
