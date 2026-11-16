package vo

import (
	"fmt"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
)

var (
	ErrNonPositiveAmount   = fmt.Errorf("%w: amount must be greater than zero", domain.ErrInvalidInput)
	ErrUnsupportedCurrency = fmt.Errorf("%w: unsupported currency", domain.ErrInvalidInput)
)

type Currency string

const (
	CurrencyBRL Currency = "BRL"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
)

var supportedCurrencies = map[Currency]struct{}{
	CurrencyBRL: {},
	CurrencyUSD: {},
	CurrencyEUR: {},
}

func ParseCurrency(raw string) (Currency, error) {
	currency := Currency(raw)
	if _, ok := supportedCurrencies[currency]; !ok {
		return "", ErrUnsupportedCurrency
	}
	return currency, nil
}

// Money keeps the amount in minor units (cents) to avoid floating point rounding.
type Money struct {
	cents    int64
	currency Currency
}

func NewMoney(cents int64, rawCurrency string) (Money, error) {
	if cents <= 0 {
		return Money{}, ErrNonPositiveAmount
	}

	currency, err := ParseCurrency(rawCurrency)
	if err != nil {
		return Money{}, err
	}

	return Money{cents: cents, currency: currency}, nil
}

func (m Money) Cents() int64       { return m.cents }
func (m Money) Currency() Currency { return m.currency }
