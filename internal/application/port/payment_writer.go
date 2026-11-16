package port

import (
	"context"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/entities"
)

//go:generate go tool mockgen -source=payment_writer.go -destination=mocks/payment_writer_mock.go -package=mocks

type PaymentWriter interface {
	Save(ctx context.Context, payment *entities.Payment) error
}
