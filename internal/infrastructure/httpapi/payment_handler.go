package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/command"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/query"
)

// PaymentHandler is a driving adapter: it only translates HTTP <-> use case.
// Tracing, logging, error mapping and the simulated delay live in middlewares;
// the request context (carrying the APM transaction and the correlated
// logger) is simply passed down to the application layer.
type PaymentHandler struct {
	createPayment *command.CreatePaymentHandler
	listPayments  *query.ListPaymentsHandler
}

func NewPaymentHandler(
	createPayment *command.CreatePaymentHandler,
	listPayments *query.ListPaymentsHandler,
) *PaymentHandler {
	return &PaymentHandler{
		createPayment: createPayment,
		listPayments:  listPayments,
	}
}

// List godoc
//
//	@Summary	List payments
//	@Tags		payments
//	@Produce	json
//	@Param		page		query		int	false	"Page number"		default(1)	minimum(1)
//	@Param		page_size	query		int	false	"Items per page"	default(10)	minimum(1)	maximum(100)
//	@Success	200			{object}	ListPaymentsResponse
//	@Failure	400			{object}	ErrorResponse
//	@Failure	500			{object}	ErrorResponse
//	@Router		/payments [get]
func (h *PaymentHandler) List(c *gin.Context) {
	var req ListPaymentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		respondBindError(c, err, "invalid query parameters")
		return
	}

	result, err := h.listPayments.Handle(c.Request.Context(), query.ListPaymentsQuery{
		PageNumber: req.Page,
		PageSize:   req.PageSize,
	})
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, fromListResult(result))
}

// Create godoc
//
//	@Summary	Create a payment
//	@Tags		payments
//	@Accept		json
//	@Produce	json
//	@Param		payment	body		CreatePaymentRequest	true	"Payment data (amount in cents)"
//	@Success	201		{object}	PaymentResponse
//	@Failure	400		{object}	ErrorResponse
//	@Failure	413		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/payments [post]
func (h *PaymentHandler) Create(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondBindError(c, err, "invalid request body")
		return
	}

	result, err := h.createPayment.Handle(c.Request.Context(), req.toCommand())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, fromCreateResult(result))
}

// CreateWithDelay godoc
//
//	@Summary		Create a payment with a simulated bottleneck
//	@Description	Same flow as POST /payments with an intentional delay, recorded as its own APM span.
//	@Tags			payments
//	@Accept			json
//	@Produce		json
//	@Param			payment	body		CreatePaymentRequest	true	"Payment data (amount in cents)"
//	@Success		201		{object}	PaymentResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		413		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/payments/delay [post]
//
// The delay itself is applied by the simulatedLatency route middleware.
func (h *PaymentHandler) CreateWithDelay(c *gin.Context) {
	h.Create(c)
}

// respondBindError answers 413 when the body hit the limitBody cap and 400 for
// any other decoding or validation failure.
func respondBindError(c *gin.Context, err error, message string) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: "request body too large"})
		return
	}
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: message})
}
