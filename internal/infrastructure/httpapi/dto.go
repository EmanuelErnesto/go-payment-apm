package httpapi

import (
	"time"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/command"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/query"
)

type CreatePaymentRequest struct {
	AmountCents int64  `json:"amount_cents" binding:"required,gt=0" example:"1990"`
	Currency    string `json:"currency" binding:"required,oneof=BRL USD EUR" example:"BRL"`
	Method      string `json:"method" binding:"required,oneof=credit_card debit_card pix boleto" example:"pix"`
	Description string `json:"description" binding:"max=255" example:"Monthly subscription"`
}

func (r CreatePaymentRequest) toCommand() command.CreatePaymentCommand {
	return command.CreatePaymentCommand{
		AmountCents: r.AmountCents,
		Currency:    r.Currency,
		Method:      r.Method,
		Description: r.Description,
	}
}

type ListPaymentsRequest struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=10"`
}

type PaymentResponse struct {
	ID          string    `json:"id" example:"0193a1c2-7b7e-7c3a-9f3e-1d2c3b4a5f60"`
	AmountCents int64     `json:"amount_cents" example:"1990"`
	Currency    string    `json:"currency" example:"BRL"`
	Method      string    `json:"method" example:"pix"`
	Status      string    `json:"status" example:"approved"`
	Description string    `json:"description" example:"Monthly subscription"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaginationResponse struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"10"`
	TotalItems int `json:"total_items" example:"42"`
	TotalPages int `json:"total_pages" example:"5"`
}

type ListPaymentsResponse struct {
	Data       []PaymentResponse  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid request body"`
}

func fromCreateResult(r command.CreatePaymentResult) PaymentResponse {
	return PaymentResponse{
		ID:          r.ID,
		AmountCents: r.AmountCents,
		Currency:    r.Currency,
		Method:      r.Method,
		Status:      r.Status,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
	}
}

func fromView(v port.PaymentView) PaymentResponse {
	return PaymentResponse{
		ID:          v.ID,
		AmountCents: v.AmountCents,
		Currency:    v.Currency,
		Method:      v.Method,
		Status:      v.Status,
		Description: v.Description,
		CreatedAt:   v.CreatedAt,
	}
}

func fromListResult(r query.ListPaymentsResult) ListPaymentsResponse {
	data := make([]PaymentResponse, 0, len(r.Items))
	for _, item := range r.Items {
		data = append(data, fromView(item))
	}

	return ListPaymentsResponse{
		Data: data,
		Pagination: PaginationResponse{
			Page:       r.PageNumber,
			PageSize:   r.PageSize,
			TotalItems: r.TotalItems,
			TotalPages: r.TotalPages,
		},
	}
}
