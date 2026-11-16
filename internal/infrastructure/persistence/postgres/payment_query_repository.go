package postgres

import (
	"context"
	"database/sql"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port"
)

var _ port.PaymentReader = (*PaymentQueryRepository)(nil)

type PaymentQueryRepository struct {
	db *sql.DB
}

func NewPaymentQueryRepository(db *sql.DB) *PaymentQueryRepository {
	return &PaymentQueryRepository{db: db}
}

func (r *PaymentQueryRepository) List(ctx context.Context, limit, offset int) ([]port.PaymentView, error) {
	const selectPage = `
		SELECT id, amount_cents, currency, method, status, description, created_at
		FROM payments
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, selectPage, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	views := make([]port.PaymentView, 0, limit)
	for rows.Next() {
		var view port.PaymentView
		if err := rows.Scan(
			&view.ID,
			&view.AmountCents,
			&view.Currency,
			&view.Method,
			&view.Status,
			&view.Description,
			&view.CreatedAt,
		); err != nil {
			return nil, err
		}
		views = append(views, view)
	}
	return views, rows.Err()
}

func (r *PaymentQueryRepository) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM payments`).Scan(&total)
	return total, err
}
