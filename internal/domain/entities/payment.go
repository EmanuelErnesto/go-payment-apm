package entities

import (
	"errors"
	"fmt"
	"time"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/vo"
)

var (
	ErrEmptyPaymentID    = fmt.Errorf("%w: payment id is required", domain.ErrInvalidInput)
	ErrPaymentNotPending = errors.New("only pending payments can be approved")
)

type Payment struct {
	id          string
	amount      vo.Money
	method      vo.PaymentMethod
	status      vo.PaymentStatus
	description vo.Description
	createdAt   time.Time
}

func NewPayment(id string, amount vo.Money, method vo.PaymentMethod, description vo.Description, createdAt time.Time) (*Payment, error) {
	if id == "" {
		return nil, ErrEmptyPaymentID
	}

	return &Payment{
		id:          id,
		amount:      amount,
		method:      method,
		status:      vo.PaymentStatusPending,
		description: description,
		// Microsecond precision is what every store we target keeps (Postgres
		// TIMESTAMPTZ). Normalising here keeps the create response identical
		// to what is later read back.
		createdAt: createdAt.UTC().Truncate(time.Microsecond),
	}, nil
}

func (p *Payment) Approve() error {
	if p.status != vo.PaymentStatusPending {
		return ErrPaymentNotPending
	}
	p.status = vo.PaymentStatusApproved
	return nil
}

func (p *Payment) ID() string                  { return p.id }
func (p *Payment) Amount() vo.Money            { return p.amount }
func (p *Payment) Method() vo.PaymentMethod    { return p.method }
func (p *Payment) Status() vo.PaymentStatus    { return p.status }
func (p *Payment) Description() vo.Description { return p.description }
func (p *Payment) CreatedAt() time.Time        { return p.createdAt }
