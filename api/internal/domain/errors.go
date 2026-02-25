package domain

import "errors"

var (
	ErrInvalidOrder        = errors.New("invalid order")
	ErrInsufficientCredits = errors.New("insufficient credits")
)
