package command

import "time"

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
