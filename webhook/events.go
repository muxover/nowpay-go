package webhook

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/muxover/nowpay-go/models"
)

// Event types set by ParseEvent. NOWPayments does not send a type; it is
// derived from payment_status.
const (
	EventPaymentCreated       = "payment_created"
	EventPaymentConfirmed     = "payment_confirmed"
	EventPaymentPartiallyPaid = "payment_partially_paid"
	EventPaymentFinished      = "payment_finished"
	EventPaymentFailed        = "payment_failed"
	EventPaymentRefunded      = "payment_refunded"
)

// Event is a parsed payment callback with the fields most handlers need.
// Use AsPaymentEvent or AsInvoiceEvent for every field.
type Event struct {
	Type string `json:"-"` // derived from PaymentStatus; empty for an unknown status

	PaymentID       int64   `json:"payment_id,omitempty"`
	ParentPaymentID int64   `json:"parent_payment_id,omitempty"`
	PaymentStatus   string  `json:"payment_status,omitempty"`
	PayAmount       float64 `json:"pay_amount,omitempty"`
	ActuallyPaid    float64 `json:"actually_paid,omitempty"`
	PayCurrency     string  `json:"pay_currency,omitempty"`
	PriceAmount     float64 `json:"price_amount,omitempty"`
	PriceCurrency   string  `json:"price_currency,omitempty"`
	OrderID         string  `json:"order_id,omitempty"`
	InvoiceID       string  `json:"invoice_id,omitempty"`
	OutcomeAmount   float64 `json:"outcome_amount,omitempty"`
	OutcomeCurrency string  `json:"outcome_currency,omitempty"`

	raw []byte
}

// UnmarshalJSON accepts IDs and amounts sent as numbers or strings.
func (e *Event) UnmarshalJSON(data []byte) error {
	type plain Event
	aux := struct {
		*plain
		PaymentID       models.FlexInt   `json:"payment_id"`
		ParentPaymentID models.FlexInt   `json:"parent_payment_id"`
		PayAmount       models.FlexFloat `json:"pay_amount"`
		ActuallyPaid    models.FlexFloat `json:"actually_paid"`
		PriceAmount     models.FlexFloat `json:"price_amount"`
		OrderID         flexString       `json:"order_id"`
		InvoiceID       flexString       `json:"invoice_id"`
		OutcomeAmount   models.FlexFloat `json:"outcome_amount"`
	}{plain: (*plain)(e)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.PaymentID = aux.PaymentID.Int64()
	e.ParentPaymentID = aux.ParentPaymentID.Int64()
	e.PayAmount = aux.PayAmount.Float64()
	e.ActuallyPaid = aux.ActuallyPaid.Float64()
	e.PriceAmount = aux.PriceAmount.Float64()
	e.OrderID = string(aux.OrderID)
	e.InvoiceID = string(aux.InvoiceID)
	e.OutcomeAmount = aux.OutcomeAmount.Float64()
	return nil
}

// PaymentFee is the fee breakdown sent with payment callbacks.
type PaymentFee struct {
	Currency      string           `json:"currency"`
	DepositFee    models.FlexFloat `json:"depositFee"`
	WithdrawalFee models.FlexFloat `json:"withdrawalFee"`
	ServiceFee    models.FlexFloat `json:"serviceFee"`
}

// PaymentEvent is the full body of a payment callback.
type PaymentEvent struct {
	PaymentID          int64       `json:"payment_id"`
	ParentPaymentID    int64       `json:"parent_payment_id,omitempty"`
	InvoiceID          int64       `json:"invoice_id,omitempty"`
	PaymentStatus      string      `json:"payment_status"`
	PayAddress         string      `json:"pay_address,omitempty"`
	PayinExtraID       string      `json:"payin_extra_id,omitempty"`
	PayAmount          float64     `json:"pay_amount,omitempty"`
	ActuallyPaid       float64     `json:"actually_paid,omitempty"`
	ActuallyPaidAtFiat float64     `json:"actually_paid_at_fiat,omitempty"`
	PayCurrency        string      `json:"pay_currency,omitempty"`
	PriceAmount        float64     `json:"price_amount,omitempty"`
	PriceCurrency      string      `json:"price_currency,omitempty"`
	OrderID            string      `json:"order_id,omitempty"`
	OrderDescription   string      `json:"order_description,omitempty"`
	PurchaseID         string      `json:"purchase_id,omitempty"`
	OutcomeAmount      float64     `json:"outcome_amount,omitempty"`
	OutcomeCurrency    string      `json:"outcome_currency,omitempty"`
	Fee                *PaymentFee `json:"fee,omitempty"`
	CreatedAt          string      `json:"created_at,omitempty"`
	UpdatedAt          string      `json:"updated_at,omitempty"`
}

// UnmarshalJSON accepts IDs and amounts sent as numbers or strings.
func (p *PaymentEvent) UnmarshalJSON(data []byte) error {
	type plain PaymentEvent
	aux := struct {
		*plain
		PaymentID          models.FlexInt   `json:"payment_id"`
		ParentPaymentID    models.FlexInt   `json:"parent_payment_id"`
		InvoiceID          models.FlexInt   `json:"invoice_id"`
		PayinExtraID       flexString       `json:"payin_extra_id"`
		PayAmount          models.FlexFloat `json:"pay_amount"`
		ActuallyPaid       models.FlexFloat `json:"actually_paid"`
		ActuallyPaidAtFiat models.FlexFloat `json:"actually_paid_at_fiat"`
		PriceAmount        models.FlexFloat `json:"price_amount"`
		OrderID            flexString       `json:"order_id"`
		PurchaseID         flexString       `json:"purchase_id"`
		OutcomeAmount      models.FlexFloat `json:"outcome_amount"`
	}{plain: (*plain)(p)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	p.PaymentID = aux.PaymentID.Int64()
	p.ParentPaymentID = aux.ParentPaymentID.Int64()
	p.InvoiceID = aux.InvoiceID.Int64()
	p.PayinExtraID = string(aux.PayinExtraID)
	p.PayAmount = aux.PayAmount.Float64()
	p.ActuallyPaid = aux.ActuallyPaid.Float64()
	p.ActuallyPaidAtFiat = aux.ActuallyPaidAtFiat.Float64()
	p.PriceAmount = aux.PriceAmount.Float64()
	p.OrderID = string(aux.OrderID)
	p.PurchaseID = string(aux.PurchaseID)
	p.OutcomeAmount = aux.OutcomeAmount.Float64()
	return nil
}

// InvoiceEvent holds the invoice-related fields of a callback.
type InvoiceEvent struct {
	InvoiceID     string  `json:"invoice_id,omitempty"`
	PaymentStatus string  `json:"payment_status,omitempty"`
	PriceAmount   float64 `json:"price_amount,omitempty"`
	PriceCurrency string  `json:"price_currency,omitempty"`
	PayAddress    string  `json:"pay_address,omitempty"`
	PayAmount     float64 `json:"pay_amount,omitempty"`
	ActuallyPaid  float64 `json:"actually_paid,omitempty"`
	OrderID       string  `json:"order_id,omitempty"`
}

// UnmarshalJSON accepts IDs and amounts sent as numbers or strings.
func (i *InvoiceEvent) UnmarshalJSON(data []byte) error {
	type plain InvoiceEvent
	aux := struct {
		*plain
		InvoiceID    flexString       `json:"invoice_id"`
		PriceAmount  models.FlexFloat `json:"price_amount"`
		PayAmount    models.FlexFloat `json:"pay_amount"`
		ActuallyPaid models.FlexFloat `json:"actually_paid"`
		OrderID      flexString       `json:"order_id"`
	}{plain: (*plain)(i)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	i.InvoiceID = string(aux.InvoiceID)
	i.PriceAmount = aux.PriceAmount.Float64()
	i.PayAmount = aux.PayAmount.Float64()
	i.ActuallyPaid = aux.ActuallyPaid.Float64()
	i.OrderID = string(aux.OrderID)
	return nil
}

// ParseEvent parses a raw IPN body. Verify the signature first.
func ParseEvent(body []byte) (*Event, error) {
	var e Event
	if err := json.Unmarshal(body, &e); err != nil {
		return nil, err
	}
	e.raw = append([]byte(nil), body...)
	e.Type = eventType(e.PaymentStatus)
	return &e, nil
}

func eventType(status string) string {
	switch status {
	case "waiting", "confirming":
		return EventPaymentCreated
	case "confirmed", "sending":
		return EventPaymentConfirmed
	case "partially_paid":
		return EventPaymentPartiallyPaid
	case "finished":
		return EventPaymentFinished
	case "failed", "expired":
		return EventPaymentFailed
	case "refunded":
		return EventPaymentRefunded
	}
	return ""
}

var errNoBody = errors.New("nowpay: event has no body; create it with ParseEvent")

// AsPaymentEvent decodes the whole callback as a PaymentEvent.
func (e *Event) AsPaymentEvent() (*PaymentEvent, error) {
	if len(e.raw) == 0 {
		return nil, errNoBody
	}
	var p PaymentEvent
	if err := json.Unmarshal(e.raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// AsInvoiceEvent decodes the invoice fields of the callback.
func (e *Event) AsInvoiceEvent() (*InvoiceEvent, error) {
	if len(e.raw) == 0 {
		return nil, errNoBody
	}
	var i InvoiceEvent
	if err := json.Unmarshal(e.raw, &i); err != nil {
		return nil, err
	}
	return &i, nil
}

// flexString decodes a JSON string, number, or null into a string.
type flexString string

func (s *flexString) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	switch {
	case len(b) == 0 || string(b) == "null":
		*s = ""
	case b[0] == '"':
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*s = flexString(v)
	default:
		*s = flexString(b)
	}
	return nil
}
