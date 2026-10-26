package domain

import "testing"

func Test_NewRequisitionQuantity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		quantity  int
		wantValue int64
		wantErr   error
	}{
		{name: "minimum quantity", quantity: 1, wantValue: 1},
		{name: "valid quantity", quantity: 5, wantValue: 5},
		{name: "maximum quantity", quantity: 1000, wantValue: 1000},
		{name: "zero quantity", quantity: 0, wantErr: ErrInvalidRequisitionQuantity},
		{name: "negative quantity", quantity: -1, wantErr: ErrInvalidRequisitionQuantity},
		{name: "quantity exceeds limit", quantity: 1001, wantErr: ErrRequisitionQuantityExceedsLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			quantity, err := NewRequisitionQuantity(tt.quantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewRequisitionQuantity() error = %v, want %v", err, tt.wantErr)
				}
				if quantity != nil {
					t.Errorf("NewRequisitionQuantity() = %v, want nil", quantity)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewRequisitionQuantity() unexpected error = %v", err)
			}
			if quantity == nil {
				t.Fatal("NewRequisitionQuantity() = nil, want quantity")
			}
			if quantity.Value() != tt.wantValue {
				t.Errorf("RequisitionQuantity.Value() = %d, want %d", quantity.Value(), tt.wantValue)
			}
		})
	}
}
