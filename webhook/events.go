package webhook

import (
	"encoding/json"
)

// EventType constants for webhook events.
const (
	EventPaymentCreated   = "payment_created"
	EventPaymentConfirmed  = "payment_confirmed"
	EventPaymentFinished  = "payment_finished"
	EventPaymentFailed     = "payment_failed"
	EventPaymentRefunded   = "payment_refunded"
	EventInvoiceCreated   = "invoice_created"
	EventInvoicePaid      = "invoice_paid"
	EventPayoutCreated    = "payout_created"
	EventPayoutFinished   = "payout_finished"
	EventPayoutFailed     = "payout_failed"
	EventSubscriptionCreated = "subscription_created"
	EventSubscriptionCancelled = "subscription_cancelled"
)

// Event is a parsed webhook event. Use Type to switch and cast Payload.
type Event struct {
	Type    string          `json:"type,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	// Raw body fields often sent by NOWPayments (payment_id, status, etc.)
	PaymentID     int64   `json:"payment_id,omitempty"`
	PaymentStatus string  `json:"payment_status,omitempty"`
	PayAmount     float64 `json:"pay_amount,omitempty"`
	PayCurrency   string  `json:"pay_currency,omitempty"`
	PriceAmount   float64 `json:"price_amount,omitempty"`
	PriceCurrency string  `json:"price_currency,omitempty"`
	OrderID       string  `json:"order_id,omitempty"`
	InvoiceID     string  `json:"invoice_id,omitempty"`
	OutcomeAmount float64 `json:"outcome_amount,omitempty"`
	OutcomeCurrency string `json:"outcome_currency,omitempty"`
}

// PaymentEvent is a payment-related webhook payload.
type PaymentEvent struct {
	PaymentID     int64   `json:"payment_id"`
	PaymentStatus string  `json:"payment_status"`
	PayAddress    string  `json:"pay_address,omitempty"`
	PayAmount     float64 `json:"pay_amount,omitempty"`
	PayCurrency   string  `json:"pay_currency,omitempty"`
	PriceAmount   float64 `json:"price_amount,omitempty"`
	PriceCurrency string  `json:"price_currency,omitempty"`
	OrderID       string  `json:"order_id,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
}

// InvoiceEvent is an invoice-related webhook payload.
type InvoiceEvent struct {
	InvoiceID      string  `json:"invoice_id,omitempty"`
	PaymentStatus  string  `json:"payment_status,omitempty"`
	PriceAmount    float64 `json:"price_amount,omitempty"`
	PriceCurrency  string  `json:"price_currency,omitempty"`
	PayAddress     string  `json:"pay_address,omitempty"`
	PayAmount      float64 `json:"pay_amount,omitempty"`
	ActuallyPaid   float64 `json:"actually_paid,omitempty"`
	OrderID        string  `json:"order_id,omitempty"`
}

// ParseEvent parses the webhook body into an Event. The body is the raw JSON POST body.
// Caller can use e.PaymentStatus or e.Type to determine event type and e.PaymentID, e.OrderID, etc. for details.
func ParseEvent(body []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}
	// If type not set, infer from payment_status
	if e.Type == "" && e.PaymentStatus != "" {
		e.Type = statusToEventType(e.PaymentStatus)
	}
	return &e, nil
}

func statusToEventType(status string) string {
	switch status {
	case "waiting", "confirming":
		return EventPaymentCreated
	case "confirmed":
		return EventPaymentConfirmed
	case "sending", "finished":
		return EventPaymentFinished
	case "failed", "expired", "refunded":
		if status == "refunded" {
			return EventPaymentRefunded
		}
		if status == "failed" || status == "expired" {
			return EventPaymentFailed
		}
		return EventPaymentFinished
	default:
		return EventPaymentCreated
	}
}

// AsPaymentEvent parses the event payload as PaymentEvent (for payment_* events).
func (e *Event) AsPaymentEvent() (*PaymentEvent, error) {
	var p PaymentEvent
	raw := e.Payload
	if len(raw) == 0 {
		// Use top-level fields
		p.PaymentID = e.PaymentID
		p.PaymentStatus = e.PaymentStatus
		p.PayAmount = e.PayAmount
		p.PayCurrency = e.PayCurrency
		p.PriceAmount = e.PriceAmount
		p.PriceCurrency = e.PriceCurrency
		p.OrderID = e.OrderID
		return &p, nil
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// AsInvoiceEvent parses the event payload as InvoiceEvent.
func (e *Event) AsInvoiceEvent() (*InvoiceEvent, error) {
	var i InvoiceEvent
	raw := e.Payload
	if len(raw) == 0 {
		i.InvoiceID = e.InvoiceID
		i.PaymentStatus = e.PaymentStatus
		i.OrderID = e.OrderID
		return &i, nil
	}
	if err := json.Unmarshal(raw, &i); err != nil {
		return nil, err
	}
	return &i, nil
}
