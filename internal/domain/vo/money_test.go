package vo_test

import (
	"errors"
	"testing"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/vo"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name         string
		cents        int64
		currency     string
		wantErr      error
		wantCurrency vo.Currency
	}{
		{name: "valid BRL amount", cents: 1990, currency: "BRL", wantCurrency: vo.CurrencyBRL},
		{name: "valid USD amount", cents: 1, currency: "USD", wantCurrency: vo.CurrencyUSD},
		{name: "valid EUR amount", cents: 50_00, currency: "EUR", wantCurrency: vo.CurrencyEUR},
		{name: "zero amount", cents: 0, currency: "BRL", wantErr: vo.ErrNonPositiveAmount},
		{name: "negative amount", cents: -10, currency: "BRL", wantErr: vo.ErrNonPositiveAmount},
		{name: "unsupported currency", cents: 100, currency: "JPY", wantErr: vo.ErrUnsupportedCurrency},
		{name: "lowercase currency", cents: 100, currency: "brl", wantErr: vo.ErrUnsupportedCurrency},
		{name: "empty currency", cents: 100, currency: "", wantErr: vo.ErrUnsupportedCurrency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			money, err := vo.NewMoney(tt.cents, tt.currency)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if !errors.Is(err, domain.ErrInvalidInput) {
					t.Fatalf("expected error to wrap domain.ErrInvalidInput, got %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if money.Cents() != tt.cents {
				t.Errorf("cents = %d, want %d", money.Cents(), tt.cents)
			}
			if money.Currency() != tt.wantCurrency {
				t.Errorf("currency = %s, want %s", money.Currency(), tt.wantCurrency)
			}
		})
	}
}
