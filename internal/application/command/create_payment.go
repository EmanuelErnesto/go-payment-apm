package command

import (
	"context"
	"fmt"
	"time"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/entities"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/vo"
)

type CreatePaymentCommand struct {
	AmountCents int64
	Currency    string
	Method      string
	Description string
}

type CreatePaymentResult struct {
	ID          string
	AmountCents int64
	Currency    string
	Method      string
	Status      string
	Description string
	CreatedAt   time.Time
}

type (
	IDGenerator func() string
	Clock       func() time.Time
)

type CreatePaymentHandler struct {
	writer port.PaymentWriter
	newID  IDGenerator
	now    Clock
}

func NewCreatePaymentHandler(writer port.PaymentWriter, newID IDGenerator, now Clock) *CreatePaymentHandler {
	return &CreatePaymentHandler{writer: writer, newID: newID, now: now}
}

func (h *CreatePaymentHandler) Handle(ctx context.Context, cmd CreatePaymentCommand) (CreatePaymentResult, error) {
	payment, err := h.buildPayment(cmd)
	if err != nil {
		return CreatePaymentResult{}, err
	}

	if err := payment.Approve(); err != nil {
		return CreatePaymentResult{}, fmt.Errorf("approving payment: %w", err)
	}

	if err := h.writer.Save(ctx, payment); err != nil {
		return CreatePaymentResult{}, fmt.Errorf("saving payment: %w", err)
	}

	return toResult(payment), nil
}

func (h *CreatePaymentHandler) buildPayment(cmd CreatePaymentCommand) (*entities.Payment, error) {
	amount, err := vo.NewMoney(cmd.AmountCents, cmd.Currency)
	if err != nil {
		return nil, err
	}

	method, err := vo.ParsePaymentMethod(cmd.Method)
	if err != nil {
		return nil, err
	}

	description, err := vo.NewDescription(cmd.Description)
	if err != nil {
		return nil, err
	}

	return entities.NewPayment(h.newID(), amount, method, description, h.now())
}

func toResult(p *entities.Payment) CreatePaymentResult {
	return CreatePaymentResult{
		ID:          p.ID(),
		AmountCents: p.Amount().Cents(),
		Currency:    string(p.Amount().Currency()),
		Method:      string(p.Method()),
		Status:      string(p.Status()),
		Description: string(p.Description()),
		CreatedAt:   p.CreatedAt(),
	}
}
