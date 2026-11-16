package query_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/port/mocks"
	"github.com/EmanuelErnesto/go-payment-apm/internal/application/query"
)

func TestListPaymentsHandler_Handle(t *testing.T) {
	errDatabase := errors.New("connection refused")
	views := []port.PaymentView{
		{ID: "a", AmountCents: 100, Currency: "BRL", Method: "pix", Status: "approved", CreatedAt: time.Now()},
		{ID: "b", AmountCents: 200, Currency: "USD", Method: "boleto", Status: "approved", CreatedAt: time.Now()},
	}

	tests := []struct {
		name      string
		query     query.ListPaymentsQuery
		setupMock func(reader *mocks.MockPaymentReader)
		want      query.ListPaymentsResult
		wantErr   error
	}{
		{
			name:  "first page",
			query: query.ListPaymentsQuery{PageNumber: 1, PageSize: 2},
			setupMock: func(reader *mocks.MockPaymentReader) {
				reader.EXPECT().Count(gomock.Any()).Return(5, nil)
				reader.EXPECT().List(gomock.Any(), 2, 0).Return(views, nil)
			},
			want: query.ListPaymentsResult{Items: views, PageNumber: 1, PageSize: 2, TotalItems: 5, TotalPages: 3},
		},
		{
			name:  "offset is derived from page",
			query: query.ListPaymentsQuery{PageNumber: 3, PageSize: 10},
			setupMock: func(reader *mocks.MockPaymentReader) {
				reader.EXPECT().Count(gomock.Any()).Return(30, nil)
				reader.EXPECT().List(gomock.Any(), 10, 20).Return(views, nil)
			},
			want: query.ListPaymentsResult{Items: views, PageNumber: 3, PageSize: 10, TotalItems: 30, TotalPages: 3},
		},
		{
			name:  "empty table",
			query: query.ListPaymentsQuery{PageNumber: 1, PageSize: 10},
			setupMock: func(reader *mocks.MockPaymentReader) {
				reader.EXPECT().Count(gomock.Any()).Return(0, nil)
				reader.EXPECT().List(gomock.Any(), 10, 0).Return([]port.PaymentView{}, nil)
			},
			want: query.ListPaymentsResult{Items: []port.PaymentView{}, PageNumber: 1, PageSize: 10},
		},
		{name: "page zero", query: query.ListPaymentsQuery{PageNumber: 0, PageSize: 10}, wantErr: query.ErrInvalidPageNumber},
		{name: "negative page", query: query.ListPaymentsQuery{PageNumber: -1, PageSize: 10}, wantErr: query.ErrInvalidPageNumber},
		{name: "page size zero", query: query.ListPaymentsQuery{PageNumber: 1, PageSize: 0}, wantErr: query.ErrInvalidPageSize},
		{name: "page size above max", query: query.ListPaymentsQuery{PageNumber: 1, PageSize: query.MaxPageSize + 1}, wantErr: query.ErrInvalidPageSize},
		{
			name:  "count failure",
			query: query.ListPaymentsQuery{PageNumber: 1, PageSize: 10},
			setupMock: func(reader *mocks.MockPaymentReader) {
				reader.EXPECT().Count(gomock.Any()).Return(0, errDatabase)
			},
			wantErr: errDatabase,
		},
		{
			name:  "list failure",
			query: query.ListPaymentsQuery{PageNumber: 1, PageSize: 10},
			setupMock: func(reader *mocks.MockPaymentReader) {
				reader.EXPECT().Count(gomock.Any()).Return(1, nil)
				reader.EXPECT().List(gomock.Any(), 10, 0).Return(nil, errDatabase)
			},
			wantErr: errDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			reader := mocks.NewMockPaymentReader(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(reader)
			}

			got, err := query.NewListPaymentsHandler(reader).Handle(context.Background(), tt.query)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("result = %+v, want %+v", got, tt.want)
			}
		})
	}
}
