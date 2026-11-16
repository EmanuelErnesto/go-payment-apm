package postgres

import (
	"context"
	"database/sql"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/entities"
)

var _ port.PaymentWriter = (*PaymentCommandRepository)(nil)

type PaymentCommandRepository struct {
	db *sql.DB
}

func NewPaymentCommandRepository(db *sql.DB) *PaymentCommandRepository {
	return &PaymentCommandRepository{db: db}
}

func (r *PaymentCommandRepository) Save(ctx context.Context, payment *entities.Payment) error {
	const insert = `
		INSERT INTO payments (id, amount_cents, currency, method, status, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, insert,
		payment.ID(),
		payment.Amount().Cents(),
		string(payment.Amount().Currency()),
		string(payment.Method()),
		string(payment.Status()),
		string(payment.Description()),
		payment.CreatedAt(),
	)
	return err
}
