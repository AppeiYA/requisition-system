package shareddomain

import (
	"math"
	"strings"
	"testing"
)

func Test_NewMoney(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		amount       float64
		currency     string
		wantAmount   int64
		wantCurrency Currency
		wantErr      error
	}{
		{
			name:         "valid amount",
			amount:       123.45,
			currency:     "NGN",
			wantAmount:   12345,
			wantCurrency: CurrencyNaira,
		},
		{
			name:         "zero amount",
			amount:       0,
			currency:     "NGN",
			wantAmount:   0,
			wantCurrency: CurrencyNaira,
		},
		{
			name:         "single decimal place",
			amount:       10.5,
			currency:     "NGN",
			wantAmount:   1050,
			wantCurrency: CurrencyNaira,
		},
		{
			name:         "smallest amount",
			amount:       0.01,
			currency:     "NGN",
			wantAmount:   1,
			wantCurrency: CurrencyNaira,
		},
		{
			name:     "invalid currency",
			amount:   10,
			currency: "USD",
			wantErr:  ErrInvalidCurrency,
		},
		{
			name:     "empty currency",
			amount:   10,
			currency: "",
			wantErr:  ErrInvalidCurrency,
		},
		{
			name:     "negative amount",
			amount:   -0.01,
			currency: "NGN",
			wantErr:  ErrNegativeAmount,
		},
		{
			name:     "NaN amount",
			amount:   math.NaN(),
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "positive infinity",
			amount:   math.Inf(1),
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "negative infinity",
			amount:   math.Inf(-1),
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "more than two decimal places",
			amount:   1.234,
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "overflow",
			amount:   1e18,
			currency: "NGN",
			wantErr:  ErrMoneyOverflow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			money, err := NewMoney(tt.amount, tt.currency)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewMoney() error = %v, want %v", err, tt.wantErr)
				}
				if money != nil {
					t.Errorf("NewMoney() = %v, want nil", money)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewMoney() unexpected error = %v", err)
			}
			if money == nil {
				t.Fatal("NewMoney() = nil, want money")
			}
			if money.Amount() != tt.wantAmount {
				t.Errorf("Money.Amount() = %d, want %d", money.Amount(), tt.wantAmount)
			}
			if money.Currency() != tt.wantCurrency {
				t.Errorf("Money.Currency() = %q, want %q", money.Currency(), tt.wantCurrency)
			}
		})
	}
}

func Test_NewNairaMoney(t *testing.T) {
	t.Parallel()

	money, err := NewNairaMoney(50)
	if err != nil {
		t.Fatalf("NewNairaMoney() unexpected error = %v", err)
	}
	if money == nil {
		t.Fatal("NewNairaMoney() = nil, want money")
	}
	if money.Amount() != 5000 {
		t.Errorf("Money.Amount() = %d, want %d", money.Amount(), 5000)
	}
	if money.Currency() != CurrencyNaira {
		t.Errorf("Money.Currency() = %q, want %q", money.Currency(), CurrencyNaira)
	}

	if _, err := NewNairaMoney(-1); err != ErrNegativeAmount {
		t.Errorf("NewNairaMoney(-1) error = %v, want %v", err, ErrNegativeAmount)
	}
}

func Test_NewMoneyFromDecimal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		value       string
		currency    string
		wantAmount  int64
		wantDecimal string
		wantErr     error
	}{
		{
			name:        "whole amount",
			value:       "123",
			currency:    "NGN",
			wantAmount:  12300,
			wantDecimal: "123.00",
		},
		{
			name:        "zero",
			value:       "0",
			currency:    "NGN",
			wantAmount:  0,
			wantDecimal: "0.00",
		},
		{
			name:        "two decimal places",
			value:       "123.45",
			currency:    "NGN",
			wantAmount:  12345,
			wantDecimal: "123.45",
		},
		{
			name:        "one decimal place is padded",
			value:       "123.4",
			currency:    "NGN",
			wantAmount:  12340,
			wantDecimal: "123.40",
		},
		{
			name:        "leading zeroes",
			value:       "001.05",
			currency:    "NGN",
			wantAmount:  105,
			wantDecimal: "1.05",
		},
		{
			name:        "surrounding whitespace",
			value:       "  42.07  ",
			currency:    "NGN",
			wantAmount:  4207,
			wantDecimal: "42.07",
		},
		{
			name:        "maximum amount",
			value:       "92233720368547758.07",
			currency:    "NGN",
			wantAmount:  math.MaxInt64,
			wantDecimal: "92233720368547758.07",
		},
		{
			name:     "empty value",
			value:    "",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "whitespace only",
			value:    "   ",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "negative value",
			value:    "-1.00",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "non numeric value",
			value:    "abc",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "too many decimal places",
			value:    "1.234",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "trailing decimal point",
			value:    "1.",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "missing whole part",
			value:    ".5",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "multiple decimal points",
			value:    "1..2",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "comma separator",
			value:    "1,00",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "non digit fractional part",
			value:    "1.5x",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "scientific notation",
			value:    "1e2",
			currency: "NGN",
			wantErr:  ErrInvalidAmount,
		},
		{
			name:     "overflow",
			value:    "92233720368547758.08",
			currency: "NGN",
			wantErr:  ErrMoneyOverflow,
		},
		{
			name:     "invalid currency",
			value:    "10.00",
			currency: "USD",
			wantErr:  ErrInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			money, err := NewMoneyFromDecimal(tt.value, tt.currency)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewMoneyFromDecimal() error = %v, want %v", err, tt.wantErr)
				}
				if money != nil {
					t.Errorf("NewMoneyFromDecimal() = %v, want nil", money)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewMoneyFromDecimal() unexpected error = %v", err)
			}
			if money == nil {
				t.Fatal("NewMoneyFromDecimal() = nil, want money")
			}
			if money.Amount() != tt.wantAmount {
				t.Errorf("Money.Amount() = %d, want %d", money.Amount(), tt.wantAmount)
			}
			if money.Currency() != CurrencyNaira {
				t.Errorf("Money.Currency() = %q, want %q", money.Currency(), CurrencyNaira)
			}
			if money.DecimalString() != tt.wantDecimal {
				t.Errorf("Money.DecimalString() = %q, want %q", money.DecimalString(), tt.wantDecimal)
			}
		})
	}
}

func Test_Money_NilReceiver(t *testing.T) {
	t.Parallel()

	var money *Money

	if money.IsZero() {
		t.Error("nil Money.IsZero() = true, want false")
	}
	if money.IsPositive() {
		t.Error("nil Money.IsPositive() = true, want false")
	}
	if money.IsNegative() {
		t.Error("nil Money.IsNegative() = true, want false")
	}
	if money.Amount() != 0 {
		t.Errorf("nil Money.Amount() = %d, want 0", money.Amount())
	}
	if money.Currency() != "" {
		t.Errorf("nil Money.Currency() = %q, want empty", money.Currency())
	}
	if money.ToFloat64() != 0 {
		t.Errorf("nil Money.ToFloat64() = %f, want 0", money.ToFloat64())
	}
	if money.DecimalString() != "" {
		t.Errorf("nil Money.DecimalString() = %q, want empty", money.DecimalString())
	}
}

func Test_Money_State(t *testing.T) {
	t.Parallel()

	zero, err := NewNairaMoney(0)
	if err != nil {
		t.Fatalf("NewNairaMoney() unexpected error = %v", err)
	}
	positive, err := NewNairaMoney(10)
	if err != nil {
		t.Fatalf("NewNairaMoney() unexpected error = %v", err)
	}
	negative := &Money{amount: -100, currency: CurrencyNaira}

	tests := []struct {
		name         string
		money        *Money
		wantZero     bool
		wantPositive bool
		wantNegative bool
	}{
		{
			name:         "nil",
			money:        nil,
			wantZero:     false,
			wantPositive: false,
			wantNegative: false,
		},
		{
			name:         "zero",
			money:        zero,
			wantZero:     true,
			wantPositive: false,
			wantNegative: false,
		},
		{
			name:         "positive",
			money:        positive,
			wantZero:     false,
			wantPositive: true,
			wantNegative: false,
		},
		{
			name:         "negative",
			money:        negative,
			wantZero:     false,
			wantPositive: false,
			wantNegative: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.money.IsZero(); got != tt.wantZero {
				t.Errorf("Money.IsZero() = %v, want %v", got, tt.wantZero)
			}
			if got := tt.money.IsPositive(); got != tt.wantPositive {
				t.Errorf("Money.IsPositive() = %v, want %v", got, tt.wantPositive)
			}
			if got := tt.money.IsNegative(); got != tt.wantNegative {
				t.Errorf("Money.IsNegative() = %v, want %v", got, tt.wantNegative)
			}
		})
	}
}

func Test_Money_Getters(t *testing.T) {
	t.Parallel()

	money, err := NewNairaMoney(123.45)
	if err != nil {
		t.Fatalf("NewNairaMoney() unexpected error = %v", err)
	}

	if money.Amount() != 12345 {
		t.Errorf("Money.Amount() = %d, want 12345", money.Amount())
	}
	if money.Currency() != CurrencyNaira {
		t.Errorf("Money.Currency() = %q, want %q", money.Currency(), CurrencyNaira)
	}
	if money.ToFloat64() != 123.45 {
		t.Errorf("Money.ToFloat64() = %f, want 123.45", money.ToFloat64())
	}
	if money.DecimalString() != "123.45" {
		t.Errorf("Money.DecimalString() = %q, want %q", money.DecimalString(), "123.45")
	}
}

func Test_Money_DecimalString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{name: "zero", amount: 0, want: "0.00"},
		{name: "single kobo", amount: 0.01, want: "0.01"},
		{name: "less than ten kobo", amount: 0.05, want: "0.05"},
		{name: "fifty kobo", amount: 0.5, want: "0.50"},
		{name: "whole naira", amount: 1, want: "1.00"},
		{name: "trailing zero fractional", amount: 123.4, want: "123.40"},
		{name: "full fractional", amount: 123.45, want: "123.45"},
		{name: "large amount", amount: 1000, want: "1000.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			money, err := NewNairaMoney(tt.amount)
			if err != nil {
				t.Fatalf("NewNairaMoney() unexpected error = %v", err)
			}
			if got := money.DecimalString(); got != tt.want {
				t.Errorf("Money.DecimalString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_Money_Equals(t *testing.T) {
	t.Parallel()

	var nilMoney *Money

	ten, err := NewNairaMoney(10)
	if err != nil {
		t.Fatalf("NewNairaMoney() unexpected error = %v", err)
	}
	tenAgain, err := NewNairaMoney(10)
	if err != nil {
		t.Fatalf("NewNairaMoney() unexpected error = %v", err)
	}
	eleven, err := NewNairaMoney(11)
	if err != nil {
		t.Fatalf("NewNairaMoney() unexpected error = %v", err)
	}
	foreign := &Money{amount: 1000, currency: Currency("USD")}

	tests := []struct {
		name string
		a    *Money
		b    *Money
		want bool
	}{
		{name: "both nil", a: nilMoney, b: nilMoney, want: true},
		{name: "left nil", a: nilMoney, b: ten, want: false},
		{name: "right nil", a: ten, b: nilMoney, want: false},
		{name: "same amount and currency", a: ten, b: tenAgain, want: true},
		{name: "different amount", a: ten, b: eleven, want: false},
		{name: "different currency", a: ten, b: foreign, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.a.Equals(tt.b); got != tt.want {
				t.Errorf("Money.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Money_Add(t *testing.T) {
	t.Parallel()

	naira := func(amount float64) *Money {
		money, err := NewNairaMoney(amount)
		if err != nil {
			t.Fatalf("NewNairaMoney(%v) unexpected error = %v", amount, err)
		}
		return money
	}

	var nilMoney *Money
	maxMoney := &Money{amount: math.MaxInt64, currency: CurrencyNaira}

	tests := []struct {
		name       string
		a          *Money
		b          *Money
		wantAmount int64
		wantErr    error
	}{
		{name: "valid addition", a: naira(10), b: naira(2.5), wantAmount: 1250},
		{name: "add zero", a: naira(10), b: naira(0), wantAmount: 1000},
		{name: "add zero to zero", a: naira(0), b: naira(0), wantAmount: 0},
		{name: "nil left operand", a: nilMoney, b: naira(1), wantErr: ErrInvalidAmount},
		{name: "nil right operand", a: naira(1), b: nilMoney, wantErr: ErrInvalidAmount},
		{name: "both nil", a: nilMoney, b: nilMoney, wantErr: ErrInvalidAmount},
		{name: "currency mismatch", a: naira(1), b: &Money{amount: 100, currency: Currency("USD")}, wantErr: ErrCurrencyMismatch},
		{name: "overflow", a: maxMoney, b: naira(0.01), wantErr: ErrMoneyOverflow},
		{name: "maximum amount plus zero", a: maxMoney, b: naira(0), wantAmount: math.MaxInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.a.Add(tt.b)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("Money.Add() error = %v, want %v", err, tt.wantErr)
				}
				if result != nil {
					t.Errorf("Money.Add() = %v, want nil", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("Money.Add() unexpected error = %v", err)
			}
			if result == nil {
				t.Fatal("Money.Add() = nil, want money")
			}
			if result.Amount() != tt.wantAmount {
				t.Errorf("Money.Add().Amount() = %d, want %d", result.Amount(), tt.wantAmount)
			}
			if result.Currency() != CurrencyNaira {
				t.Errorf("Money.Add().Currency() = %q, want %q", result.Currency(), CurrencyNaira)
			}
		})
	}
}

func Test_Money_Subtract(t *testing.T) {
	t.Parallel()

	naira := func(amount float64) *Money {
		money, err := NewNairaMoney(amount)
		if err != nil {
			t.Fatalf("NewNairaMoney(%v) unexpected error = %v", amount, err)
		}
		return money
	}

	var nilMoney *Money

	tests := []struct {
		name       string
		a          *Money
		b          *Money
		wantAmount int64
		wantErr    error
	}{
		{name: "valid subtraction", a: naira(10), b: naira(2.5), wantAmount: 750},
		{name: "equal amounts", a: naira(2.5), b: naira(2.5), wantAmount: 0},
		{name: "subtract zero", a: naira(10), b: naira(0), wantAmount: 1000},
		{name: "nil left operand", a: nilMoney, b: naira(1), wantErr: ErrInvalidAmount},
		{name: "nil right operand", a: naira(1), b: nilMoney, wantErr: ErrInvalidAmount},
		{name: "both nil", a: nilMoney, b: nilMoney, wantErr: ErrInvalidAmount},
		{name: "currency mismatch", a: naira(1), b: &Money{amount: 100, currency: Currency("USD")}, wantErr: ErrCurrencyMismatch},
		{name: "negative result", a: naira(1), b: naira(2), wantErr: ErrNegativeAmount},
		{name: "maximum minus maximum", a: &Money{amount: math.MaxInt64, currency: CurrencyNaira}, b: &Money{amount: math.MaxInt64, currency: CurrencyNaira}, wantAmount: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.a.Subtract(tt.b)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("Money.Subtract() error = %v, want %v", err, tt.wantErr)
				}
				if result != nil {
					t.Errorf("Money.Subtract() = %v, want nil", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("Money.Subtract() unexpected error = %v", err)
			}
			if result == nil {
				t.Fatal("Money.Subtract() = nil, want money")
			}
			if result.Amount() != tt.wantAmount {
				t.Errorf("Money.Subtract().Amount() = %d, want %d", result.Amount(), tt.wantAmount)
			}
			if result.Currency() != CurrencyNaira {
				t.Errorf("Money.Subtract().Currency() = %q, want %q", result.Currency(), CurrencyNaira)
			}
		})
	}
}

func Test_Money_Multiply(t *testing.T) {
	t.Parallel()

	naira := func(amount float64) *Money {
		money, err := NewNairaMoney(amount)
		if err != nil {
			t.Fatalf("NewNairaMoney(%v) unexpected error = %v", amount, err)
		}
		return money
	}

	var nilMoney *Money
	maxMoney := &Money{amount: math.MaxInt64, currency: CurrencyNaira}

	tests := []struct {
		name       string
		money      *Money
		quantity   int64
		wantAmount int64
		wantErr    error
	}{
		{name: "valid multiplication", money: naira(10.5), quantity: 3, wantAmount: 3150},
		{name: "multiply by one", money: naira(1), quantity: 1, wantAmount: 100},
		{name: "multiply by zero", money: naira(10), quantity: 0, wantAmount: 0},
		{name: "maximum amount by zero", money: maxMoney, quantity: 0, wantAmount: 0},
		{name: "nil money", money: nilMoney, quantity: 1, wantErr: ErrInvalidAmount},
		{name: "negative quantity", money: naira(10), quantity: -1, wantErr: ErrInvalidAmount},
		{name: "overflow", money: maxMoney, quantity: 2, wantErr: ErrMoneyOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tt.money.Multiply(tt.quantity)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("Money.Multiply() error = %v, want %v", err, tt.wantErr)
				}
				if result != nil {
					t.Errorf("Money.Multiply() = %v, want nil", result)
				}
				return
			}

			if err != nil {
				t.Fatalf("Money.Multiply() unexpected error = %v", err)
			}
			if result == nil {
				t.Fatal("Money.Multiply() = nil, want money")
			}
			if result.Amount() != tt.wantAmount {
				t.Errorf("Money.Multiply().Amount() = %d, want %d", result.Amount(), tt.wantAmount)
			}
			if result.Currency() != CurrencyNaira {
				t.Errorf("Money.Multiply().Currency() = %q, want %q", result.Currency(), CurrencyNaira)
			}
		})
	}
}

func Test_Currency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		wantValue Currency
		wantErr   error
	}{
		{name: "valid naira", value: "NGN", wantValue: CurrencyNaira},
		{name: "empty currency", value: "", wantErr: ErrInvalidCurrency},
		{name: "lowercase currency", value: "ngn", wantErr: ErrInvalidCurrency},
		{name: "unsupported currency", value: "USD", wantErr: ErrInvalidCurrency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			currency, err := NewCurrency(tt.value)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("NewCurrency() error = %v, want %v", err, tt.wantErr)
				}
				if currency != "" {
					t.Errorf("NewCurrency() = %q, want empty", currency)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewCurrency() unexpected error = %v", err)
			}
			if currency != tt.wantValue {
				t.Errorf("NewCurrency() = %q, want %q", currency, tt.wantValue)
			}
			if currency.String() != tt.value {
				t.Errorf("Currency.String() = %q, want %q", currency.String(), tt.value)
			}
		})
	}
}

func Test_Money_MaximumDecimalInput(t *testing.T) {
	t.Parallel()

	money, err := NewMoneyFromDecimal(strings.TrimSpace("92233720368547758.07"), "NGN")
	if err != nil {
		t.Fatalf("NewMoneyFromDecimal() unexpected error = %v", err)
	}
	if money.Amount() != math.MaxInt64 {
		t.Errorf("Money.Amount() = %d, want %d", money.Amount(), int64(math.MaxInt64))
	}
}
