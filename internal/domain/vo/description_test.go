package vo_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain/vo"
)

func TestNewDescription(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    vo.Description
		wantErr error
	}{
		{name: "empty is allowed", raw: "", want: ""},
		{name: "trims surrounding spaces", raw: "  coffee  ", want: "coffee"},
		{name: "exactly the limit", raw: strings.Repeat("a", vo.MaxDescriptionLength), want: vo.Description(strings.Repeat("a", vo.MaxDescriptionLength))},
		{name: "limit counts runes, not bytes", raw: strings.Repeat("é", vo.MaxDescriptionLength), want: vo.Description(strings.Repeat("é", vo.MaxDescriptionLength))},
		{name: "above the limit", raw: strings.Repeat("a", vo.MaxDescriptionLength+1), wantErr: vo.ErrDescriptionTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vo.NewDescription(tt.raw)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Errorf("description = %q, want %q", got, tt.want)
			}
		})
	}
}
