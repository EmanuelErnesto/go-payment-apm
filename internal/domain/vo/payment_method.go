package vo

import (
	"fmt"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
)

var ErrUnsupportedPaymentMethod = fmt.Errorf("%w: unsupported payment method", domain.ErrInvalidInput)

type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodDebitCard  PaymentMethod = "debit_card"
	PaymentMethodPix        PaymentMethod = "pix"
	PaymentMethodBoleto     PaymentMethod = "boleto"
)

var supportedPaymentMethods = map[PaymentMethod]struct{}{
	PaymentMethodCreditCard: {},
	PaymentMethodDebitCard:  {},
	PaymentMethodPix:        {},
	PaymentMethodBoleto:     {},
}

func ParsePaymentMethod(raw string) (PaymentMethod, error) {
	method := PaymentMethod(raw)
	if _, ok := supportedPaymentMethods[method]; !ok {
		return "", ErrUnsupportedPaymentMethod
	}
	return method, nil
}
