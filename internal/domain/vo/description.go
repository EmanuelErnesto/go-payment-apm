package vo

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/EmanuelErnesto/go-payment-apm/internal/domain"
)

const MaxDescriptionLength = 255

var ErrDescriptionTooLong = fmt.Errorf("%w: description must have at most %d characters", domain.ErrInvalidInput, MaxDescriptionLength)

type Description string

func NewDescription(raw string) (Description, error) {
	trimmed := strings.TrimSpace(raw)
	if utf8.RuneCountInString(trimmed) > MaxDescriptionLength {
		return "", ErrDescriptionTooLong
	}
	return Description(trimmed), nil
}
