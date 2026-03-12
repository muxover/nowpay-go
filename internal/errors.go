package internal

import (
	"errors"
	"fmt"
)

// Sentinel errors for API and client failures. Use errors.Is to check.
var (
	ErrInvalidAPIKey      = errors.New("nowpay: invalid API key")
	ErrPaymentNotFound    = errors.New("nowpay: payment not found")
	ErrInvoiceNotFound   = errors.New("nowpay: invoice not found")
	ErrPayoutNotFound    = errors.New("nowpay: payout not found")
	ErrSubscriptionNotFound = errors.New("nowpay: subscription not found")
	ErrRateLimited       = errors.New("nowpay: rate limited")
	ErrInvalidSignature   = errors.New("nowpay: invalid webhook signature")
	ErrBadRequest         = errors.New("nowpay: bad request")
	ErrValidation        = errors.New("nowpay: validation error")
	ErrServerError       = errors.New("nowpay: server error")
	ErrUnauthorized      = errors.New("nowpay: unauthorized")
	ErrForbidden         = errors.New("nowpay: forbidden")
	ErrNotFound          = errors.New("nowpay: not found")
)

// APIError wraps an API response error with status and optional message.
type APIError struct {
	StatusCode int
	Message    string
	Err        error
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("nowpay: %s (status %d)", e.Message, e.StatusCode)
	}
	if e.Err != nil {
		return fmt.Sprintf("nowpay: %v (status %d)", e.Err, e.StatusCode)
	}
	return fmt.Sprintf("nowpay: API error (status %d)", e.StatusCode)
}

func (e *APIError) Unwrap() error { return e.Err }

// IsAPIError returns true if err is or wraps APIError.
func IsAPIError(err error) bool {
	var ae *APIError
	return errors.As(err, &ae)
}
