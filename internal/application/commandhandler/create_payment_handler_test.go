package commandhandler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/command"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/commandhandler"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port/mocks"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/entities"
	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/vo"
)

const fixedID = "0193a1c2-7b7e-7c3a-9f3e-1d2c3b4a5f60"

var fixedNow = time.Date(2026, 11, 16, 13, 0, 0, 0, time.UTC)

func TestCreatePaymentHandler_Handle(t *testing.T) {
	errDatabase := errors.New("connection refused")

	validCommand := command.CreatePaymentCommand{
		AmountCents: 1990,
		Currency:    "BRL",
		Method:      "pix",
		Description: "  coffee  ",
	}

	tests := []struct {
		name       string
		cmd        command.CreatePaymentCommand
		setupMock  func(writer *mocks.MockPaymentWriter)
		want       command.CreatePaymentResult
		wantErr    error
		wantClient bool
	}{
		{
			name: "persists an approved payment",
			cmd:  validCommand,
			setupMock: func(writer *mocks.MockPaymentWriter) {
				writer.EXPECT().
					Save(gomock.Any(), gomock.Cond(func(p *entities.Payment) bool {
						return p.ID() == fixedID && p.Status() == vo.PaymentStatusApproved
					})).
					Return(nil)
			},
			want: command.CreatePaymentResult{
				ID:          fixedID,
				AmountCents: 1990,
				Currency:    "BRL",
				Method:      "pix",
				Status:      "approved",
				Description: "coffee",
				CreatedAt:   fixedNow,
			},
		},
		{
			name:       "rejects non positive amount",
			cmd:        command.CreatePaymentCommand{AmountCents: 0, Currency: "BRL", Method: "pix"},
			wantErr:    vo.ErrNonPositiveAmount,
			wantClient: true,
		},
		{
			name:       "rejects unsupported currency",
			cmd:        command.CreatePaymentCommand{AmountCents: 100, Currency: "XYZ", Method: "pix"},
			wantErr:    vo.ErrUnsupportedCurrency,
			wantClient: true,
		},
		{
			name:       "rejects unsupported method",
			cmd:        command.CreatePaymentCommand{AmountCents: 100, Currency: "BRL", Method: "cash"},
			wantErr:    vo.ErrUnsupportedPaymentMethod,
			wantClient: true,
		},
		{
			name: "propagates repository failure",
			cmd:  validCommand,
			setupMock: func(writer *mocks.MockPaymentWriter) {
				writer.EXPECT().Save(gomock.Any(), gomock.Any()).Return(errDatabase)
			},
			wantErr: errDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			writer := mocks.NewMockPaymentWriter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(writer)
			}

			handler := commandhandler.NewCreatePaymentHandler(
				writer,
				func() string { return fixedID },
				func() time.Time { return fixedNow },
			)

			got, err := handler.Handle(context.Background(), tt.cmd)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				if errors.Is(err, domain.ErrInvalidInput) != tt.wantClient {
					t.Fatalf("client error classification mismatch for %v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("result = %+v, want %+v", got, tt.want)
			}
		})
	}
}
