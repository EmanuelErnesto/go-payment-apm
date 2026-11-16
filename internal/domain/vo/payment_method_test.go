package vo_test

import (
	"errors"
	"testing"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/vo"
)

func TestParsePaymentMethod(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    vo.PaymentMethod
		wantErr error
	}{
		{name: "credit card", raw: "credit_card", want: vo.PaymentMethodCreditCard},
		{name: "debit card", raw: "debit_card", want: vo.PaymentMethodDebitCard},
		{name: "pix", raw: "pix", want: vo.PaymentMethodPix},
		{name: "boleto", raw: "boleto", want: vo.PaymentMethodBoleto},
		{name: "unknown method", raw: "cash", wantErr: vo.ErrUnsupportedPaymentMethod},
		{name: "wrong casing", raw: "PIX", wantErr: vo.ErrUnsupportedPaymentMethod},
		{name: "empty", raw: "", wantErr: vo.ErrUnsupportedPaymentMethod},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vo.ParsePaymentMethod(tt.raw)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Errorf("method = %q, want %q", got, tt.want)
			}
		})
	}
}
