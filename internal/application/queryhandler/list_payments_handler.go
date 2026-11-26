package queryhandler

import (
	"context"
	"fmt"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/query"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
)

const MaxPageSize = 100

var (
	ErrInvalidPageNumber = fmt.Errorf("%w: page must be greater than zero", domain.ErrInvalidInput)
	ErrInvalidPageSize   = fmt.Errorf("%w: page size must be between 1 and %d", domain.ErrInvalidInput, MaxPageSize)
)

type ListPaymentsHandler struct {
	reader port.PaymentReader
}

func NewListPaymentsHandler(reader port.PaymentReader) *ListPaymentsHandler {
	return &ListPaymentsHandler{reader: reader}
}

func (h *ListPaymentsHandler) Handle(ctx context.Context, q query.ListPaymentsQuery) (query.ListPaymentsResult, error) {
	if q.PageNumber < 1 {
		return query.ListPaymentsResult{}, ErrInvalidPageNumber
	}
	if q.PageSize < 1 || q.PageSize > MaxPageSize {
		return query.ListPaymentsResult{}, ErrInvalidPageSize
	}

	total, err := h.reader.Count(ctx)
	if err != nil {
		return query.ListPaymentsResult{}, fmt.Errorf("counting payments: %w", err)
	}

	items, err := h.reader.List(ctx, q.PageSize, (q.PageNumber-1)*q.PageSize)
	if err != nil {
		return query.ListPaymentsResult{}, fmt.Errorf("listing payments: %w", err)
	}

	return query.ListPaymentsResult{
		Items:      items,
		PageNumber: q.PageNumber,
		PageSize:   q.PageSize,
		TotalItems: total,
		TotalPages: ceilDiv(total, q.PageSize),
	}, nil
}

func ceilDiv(a, b int) int {
	return (a + b - 1) / b
}
