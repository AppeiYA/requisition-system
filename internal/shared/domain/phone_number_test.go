package shareddomain

import "testing"

func Test_NewPhoneNumber(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantValue string
		wantErr   bool
	}{
		{name: "valid E.164 number", value: "+14155551234", wantValue: "+14155551234"},
		{name: "valid with separators", value: "+1 (415) 555-1234", wantValue: "+14155551234"},
		{name: "valid with surrounding whitespace", value: "  +14155551234  ", wantValue: "+14155551234"},
		{name: "minimum length", value: "+12345678", wantValue: "+12345678"},
		{name: "missing plus", value: "14155551234", wantErr: true},
		{name: "letters", value: "+1abc5551234", wantErr: true},
		{name: "too short", value: "+1234567", wantErr: true},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := NewPhoneNumber(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("NewPhoneNumber() error = nil, wantErr %v", tt.wantErr)
				}
				if phone != nil {
					t.Errorf("NewPhoneNumber() = %v, want nil", phone)
				}
				if err != ErrInvalidPhoneNumber {
					t.Errorf("NewPhoneNumber() error = %v, want %v", err, ErrInvalidPhoneNumber)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPhoneNumber() unexpected error = %v", err)
			}
			if phone == nil {
				t.Fatal("NewPhoneNumber() = nil, want valid phone")
			}
			if phone.Value() != tt.wantValue {
				t.Errorf("NewPhoneNumber().Value() = %v, want %v", phone.Value(), tt.wantValue)
			}
		})
	}
}

func Test_NewOptionalPhoneNumber(t *testing.T) {
	t.Parallel()

	phone, err := NewOptionalPhoneNumber("")
	if err != nil {
		t.Fatalf("NewOptionalPhoneNumber(\"\") unexpected error = %v", err)
	}
	if phone != nil {
		t.Errorf("NewOptionalPhoneNumber(\"\") = %v, want nil", phone)
	}

	phone, err = NewOptionalPhoneNumber("+14155551234")
	if err != nil {
		t.Fatalf("NewOptionalPhoneNumber() unexpected error = %v", err)
	}
	if phone == nil || phone.Value() != "+14155551234" {
		t.Errorf("NewOptionalPhoneNumber() = %v, want %q", phone, "+14155551234")
	}

	if _, err := NewOptionalPhoneNumber("not-a-phone"); err != ErrInvalidPhoneNumber {
		t.Errorf("NewOptionalPhoneNumber() error = %v, want %v", err, ErrInvalidPhoneNumber)
	}
}

func Test_PhoneNumber_IsZeroAndEquals(t *testing.T) {
	t.Parallel()

	var nilPhone *PhoneNumber
	if !nilPhone.IsZero() {
		t.Error("nil PhoneNumber should be zero")
	}
	if nilPhone.Value() != "" {
		t.Errorf("nil PhoneNumber Value() = %q, want empty", nilPhone.Value())
	}

	a, _ := NewPhoneNumber("+14155551234")
	b, _ := NewPhoneNumber("+1 (415) 555-1234")
	c, _ := NewPhoneNumber("+442071838750")

	if !a.Equals(b) {
		t.Error("expected normalized numbers to be equal")
	}
	if a.Equals(c) {
		t.Error("expected different numbers not to be equal")
	}
	if !a.Equals(a) {
		t.Error("expected a number to equal itself")
	}
}

func Test_PhoneNumber_EqualsNilCases(t *testing.T) {
	t.Parallel()

	var nilPhone *PhoneNumber
	phone, err := NewPhoneNumber("+14155551234")
	if err != nil {
		t.Fatalf("NewPhoneNumber() unexpected error = %v", err)
	}

	if !nilPhone.Equals(nil) {
		t.Error("nil PhoneNumber should equal nil")
	}
	if nilPhone.Equals(phone) {
		t.Error("nil PhoneNumber should not equal a value")
	}
	if phone.Equals(nil) {
		t.Error("PhoneNumber should not equal nil")
	}
}
