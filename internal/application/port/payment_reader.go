package port

import (
	"context"
	"time"
)

//go:generate go tool mockgen -source=payment_reader.go -destination=mocks/payment_reader_mock.go -package=mocks

// PaymentView is the read model: the query side returns flat projections
// instead of rehydrating domain entities.
type PaymentView struct {
	ID          string
	AmountCents int64
	Currency    string
	Method      string
	Status      string
	Description string
	CreatedAt   time.Time
}

type PaymentReader interface {
	List(ctx context.Context, limit, offset int) ([]PaymentView, error)
	Count(ctx context.Context) (int, error)
}
