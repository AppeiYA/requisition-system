package shareddomain

import (
	"math"
	"strconv"
	"strings"

	customerrors "github.com/AppeiYA/requisition-system/internal/shared/infrastructure/errors"
)

var (
	ErrInvalidCurrency = customerrors.NewBadRequestError("invalid currency")
	ErrInvalidAmount = customerrors.NewBadRequestError("invalid monetary amount")

	ErrNegativeAmount = customerrors.NewBadRequestError(
		"monetary amount cannot be negative",
	)

	ErrCurrencyMismatch = customerrors.NewBadRequestError(
		"currency mismatch",
	)

	ErrMoneyOverflow = customerrors.NewBadRequestError("monetary arithmetic overflow")
)

type Currency string
const (
	CurrencyNaira Currency = "NGN"
)
func (c Currency) isValid() bool {
	switch c {
	case CurrencyNaira:
		return true
	default:
		return false
	}
}
func NewCurrency(value string) (Currency, error) {
	currency := Currency(value)
	if !currency.isValid() {
		return "", ErrInvalidCurrency
	}
	return currency, nil
}
func (c Currency) String() string {
	return string(c)
}

type Money struct {
	amount   int64 
	currency Currency
}

// NewMoney creates money from an amount expressed in naira (NGN) or other valid currency. The amount is expected to be in the major unit (e.g., naira), and it will be converted to the minor unit (e.g., kobo) by multiplying by 100.
func NewMoney(amount float64, currency string) (*Money, error) {
	validCurrency, err := NewCurrency(currency)
	if err != nil {
		return nil, err
	}

	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return nil, ErrInvalidAmount
	}
 
	if amount < 0 {
		return nil, ErrNegativeAmount
	}

	scaled := amount * 100

	if scaled > float64(math.MaxInt64) {
        return nil, ErrMoneyOverflow
    }

	rounded := math.Round(scaled)

    if math.Abs(scaled-rounded) > 1e-7 {
        return nil, ErrInvalidAmount
    }

	return &Money{
		amount:   int64(rounded),
		currency: validCurrency,
	}, nil
}

// NewNairaMoney creates money from an amount expressed in naira (NGN). The amount is expected to be in the major unit (naira), and it will be converted to the minor unit (kobo) by multiplying by 100.
func NewNairaMoney(amount float64) (*Money, error) {
	return NewMoney(amount, string(CurrencyNaira))
}

func NewMoneyFromDecimal(value string, currency string) (*Money, error) {
	validCurrency, err := NewCurrency(currency)
	if err != nil {
		return nil, err
	}

	value = strings.TrimSpace(value)

	if value == "" || strings.HasPrefix(value, "-") {
		return nil, ErrInvalidAmount
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return nil, ErrInvalidAmount
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return nil, ErrInvalidAmount
	}

	var fractional int64

	if len(parts) == 2 {
		decimal := parts[1]

		// NGN uses two decimal places.
		if len(decimal) == 0 || len(decimal) > 2 {
			return nil, ErrInvalidAmount
		}

		for _, r := range decimal {
			if r < '0' || r > '9' {
				return nil, ErrInvalidAmount
			}
		}

		if len(decimal) == 1 {
			decimal += "0"
		}

		fractional, err = strconv.ParseInt(decimal, 10, 64)
		if err != nil {
			return nil, ErrInvalidAmount
		}
	}

	if whole > (math.MaxInt64-fractional)/100 {
		return nil, ErrMoneyOverflow
	}

	return &Money{
		amount:   whole*100 + fractional,
		currency: validCurrency,
	}, nil
}

func (m *Money) IsZero() bool {
	return m != nil && m.amount == 0
}

func (m *Money) IsPositive() bool {
	return m != nil && m.amount > 0
}

func (m *Money) IsNegative() bool {
	return m != nil && m.amount < 0
}

func (m *Money) Equals(other *Money) bool {
	if m == nil || other == nil {
		return m == other
	}

	return m.amount == other.amount &&
		m.currency == other.currency
}

func (m *Money) Amount() int64 {
	if m == nil {
		return 0
	}

	return m.amount
}

func (m *Money) Currency() Currency {
	if m == nil {
		return ""
	}

	return m.currency
}

func (m *Money) ToFloat64() float64 {
	if m == nil {
		return 0
	}

	return float64(m.amount) / 100
}

// DecimalString returns a decimal representation suitable for display.
func (m *Money) DecimalString() string {
	if m == nil {
		return ""
	}

	whole := m.amount / 100
	fractional := m.amount % 100

	return strconv.FormatInt(whole, 10) + "." +
		leftPadTwoDigits(fractional)
}

func leftPadTwoDigits(value int64) string {
	if value < 10 {
		return "0" + strconv.FormatInt(value, 10)
	}

	return strconv.FormatInt(value, 10)
}

// Add adds two monetary values with the same currency.
func (m *Money) Add(other *Money) (*Money, error) {
	if m == nil || other == nil {
		return nil, ErrInvalidAmount
	}

	if m.currency != other.currency {
		return nil, ErrCurrencyMismatch
	}

	if other.amount > math.MaxInt64-m.amount {
		return nil, ErrMoneyOverflow
	}

	return &Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

func (m *Money) Subtract(other *Money) (*Money, error) {
	if m == nil || other == nil {
		return nil, ErrInvalidAmount
	}

	if m.currency != other.currency {
		return nil, ErrCurrencyMismatch
	}

	if other.amount > m.amount {
		return nil, ErrNegativeAmount
	}

	return &Money{
		amount:   m.amount - other.amount,
		currency: m.currency,
	}, nil
}

func (m *Money) Multiply(quantity int64) (*Money, error) {
	if m == nil || quantity < 0 {
		return nil, ErrInvalidAmount
	}

	if quantity != 0 && m.amount > math.MaxInt64/quantity {
		return nil, ErrMoneyOverflow
	}

	return &Money{
		amount:   m.amount * quantity,
		currency: m.currency,
	}, nil
}