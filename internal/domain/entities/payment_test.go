package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/entities"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/vo"
)

func mustMoney(t *testing.T, cents int64, currency string) vo.Money {
	t.Helper()
	money, err := vo.NewMoney(cents, currency)
	if err != nil {
		t.Fatalf("building money: %v", err)
	}
	return money
}

func TestNewPayment(t *testing.T) {
	saoPaulo := time.FixedZone("BRT", -3*60*60)
	createdAt := time.Date(2026, 11, 16, 10, 0, 0, 0, saoPaulo)

	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{name: "valid payment starts pending", id: "0193a1c2-7b7e-7c3a-9f3e-1d2c3b4a5f60"},
		{name: "missing id", id: "", wantErr: entities.ErrEmptyPaymentID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := mustMoney(t, 1990, "BRL")

			payment, err := entities.NewPayment(tt.id, amount, vo.PaymentMethodPix, "coffee", createdAt)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if payment.ID() != tt.id {
				t.Errorf("id = %q, want %q", payment.ID(), tt.id)
			}
			if payment.Amount() != amount {
				t.Errorf("amount = %+v, want %+v", payment.Amount(), amount)
			}
			if payment.Method() != vo.PaymentMethodPix {
				t.Errorf("method = %q, want %q", payment.Method(), vo.PaymentMethodPix)
			}
			if payment.Status() != vo.PaymentStatusPending {
				t.Errorf("status = %q, want %q", payment.Status(), vo.PaymentStatusPending)
			}
			if payment.Description() != "coffee" {
				t.Errorf("description = %q, want %q", payment.Description(), "coffee")
			}
			if payment.CreatedAt().Location() != time.UTC || !payment.CreatedAt().Equal(createdAt) {
				t.Errorf("created_at = %v, want %v in UTC", payment.CreatedAt(), createdAt)
			}
		})
	}
}

func TestPaymentApprove(t *testing.T) {
	tests := []struct {
		name          string
		approvalsDone int
		wantErr       error
	}{
		{name: "pending payment is approved", approvalsDone: 0},
		{name: "approved payment cannot be approved again", approvalsDone: 1, wantErr: entities.ErrPaymentNotPending},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payment, err := entities.NewPayment("id", mustMoney(t, 100, "USD"), vo.PaymentMethodCreditCard, "", time.Now())
			if err != nil {
				t.Fatalf("building payment: %v", err)
			}
			for range tt.approvalsDone {
				if err := payment.Approve(); err != nil {
					t.Fatalf("arranging approval: %v", err)
				}
			}

			err = payment.Approve()

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if payment.Status() != vo.PaymentStatusApproved {
				t.Errorf("status = %q, want %q", payment.Status(), vo.PaymentStatusApproved)
			}
		})
	}
}

func TestNewPayment_CreatedAtHasMicrosecondPrecision(t *testing.T) {
	withNanos := time.Date(2026, 11, 16, 10, 0, 0, 876222709, time.UTC)

	payment, err := entities.NewPayment("id", mustMoney(t, 100, "USD"), vo.PaymentMethodPix, "", withNanos)
	if err != nil {
		t.Fatal(err)
	}

	want := time.Date(2026, 11, 16, 10, 0, 0, 876222000, time.UTC)
	if !payment.CreatedAt().Equal(want) {
		t.Fatalf("created_at = %v, want %v", payment.CreatedAt(), want)
	}
}
