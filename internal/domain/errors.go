package domain

import "errors"

// ErrInvalidInput is the sentinel every validation error wraps, so adapters can
// translate the whole family into a client error with a single errors.Is check.
var ErrInvalidInput = errors.New("invalid input")
