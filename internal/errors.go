package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Sentinel errors for API and client failures. Use errors.Is to check.
var (
	ErrInvalidAPIKey        = errors.New("nowpay: invalid API key")
	ErrPaymentNotFound      = errors.New("nowpay: payment not found")
	ErrInvoiceNotFound      = errors.New("nowpay: invoice not found")
	ErrSubscriptionNotFound = errors.New("nowpay: subscription not found")
	ErrRateLimited          = errors.New("nowpay: rate limited")
	ErrBadRequest           = errors.New("nowpay: bad request")
	ErrValidation           = errors.New("nowpay: validation error")
	ErrServerError          = errors.New("nowpay: server error")
	ErrUnauthorized         = errors.New("nowpay: unauthorized")
	ErrForbidden            = errors.New("nowpay: forbidden")
	ErrNotFound             = errors.New("nowpay: not found")
)

// APIError is returned for any response with a 4xx or 5xx status. It unwraps
// to the sentinel errors that match its status, so errors.Is works on it.
type APIError struct {
	StatusCode int    // HTTP status
	Code       string // NOWPayments error code, e.g. "INVALID_REQUEST_PARAMS"
	Message    string // NOWPayments error message, or the raw body

	sentinels []error
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.StatusCode)
	}
	if e.Code != "" {
		return fmt.Sprintf("nowpay: %s (status %d, %s)", msg, e.StatusCode, e.Code)
	}
	return fmt.Sprintf("nowpay: %s (status %d)", msg, e.StatusCode)
}

// Unwrap returns the sentinel errors that match the status code.
func (e *APIError) Unwrap() []error { return e.sentinels }

// IsAPIError reports whether err is or wraps an *APIError.
func IsAPIError(err error) bool {
	var ae *APIError
	return errors.As(err, &ae)
}

const maxMessageLen = 512

func newAPIError(status int, body []byte) *APIError {
	e := &APIError{StatusCode: status}

	var payload struct {
		Code    string          `json:"code"`
		Message json.RawMessage `json:"message"`
	}
	if json.Unmarshal(body, &payload) == nil && (payload.Code != "" || len(payload.Message) > 0) {
		e.Code = payload.Code
		var s string
		if json.Unmarshal(payload.Message, &s) == nil {
			e.Message = s
		} else {
			e.Message = string(payload.Message)
		}
	} else {
		e.Message = strings.TrimSpace(string(body))
	}
	if len(e.Message) > maxMessageLen {
		e.Message = strings.ToValidUTF8(e.Message[:maxMessageLen], "") + "…"
	}

	switch {
	case status == http.StatusUnauthorized:
		e.sentinels = []error{ErrUnauthorized, ErrInvalidAPIKey}
	case status == http.StatusForbidden:
		e.sentinels = []error{ErrForbidden}
	case status == http.StatusNotFound:
		e.sentinels = []error{ErrNotFound}
	case status == http.StatusBadRequest:
		e.sentinels = []error{ErrBadRequest}
	case status == http.StatusUnprocessableEntity:
		e.sentinels = []error{ErrValidation}
	case status == http.StatusTooManyRequests:
		e.sentinels = []error{ErrRateLimited}
	case status >= 500:
		e.sentinels = []error{ErrServerError}
	}
	return e
}
