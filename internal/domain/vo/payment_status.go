package vo

type PaymentStatus string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusApproved PaymentStatus = "approved"
)
