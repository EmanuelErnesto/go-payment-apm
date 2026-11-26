package query

import "github.com/EmanuelErnesto/go-payment-apm/internal/application/port"

type ListPaymentsQuery struct {
	PageNumber int
	PageSize   int
}

type ListPaymentsResult struct {
	Items      []port.PaymentView
	PageNumber int
	PageSize   int
	TotalItems int
	TotalPages int
}
