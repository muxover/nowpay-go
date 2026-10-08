package models

import "encoding/json"

// Invoice is a hosted invoice (POST /v1/invoice).
type Invoice struct {
	ID               int64   `json:"id"`
	TokenID          string  `json:"token_id,omitempty"`
	InvoiceURL       string  `json:"invoice_url"`
	InvoiceID        string  `json:"invoice_id,omitempty"`
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	IPNCallbackURL   string  `json:"ipn_callback_url,omitempty"`
	SuccessURL       string  `json:"success_url,omitempty"`
	CancelURL        string  `json:"cancel_url,omitempty"`
	PartiallyPaidURL string  `json:"partially_paid_url,omitempty"`
	IsFixedRate      bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
	PaymentStatus    string  `json:"payment_status,omitempty"`
	PayAddress       string  `json:"pay_address,omitempty"`
	PayAmount        float64 `json:"pay_amount,omitempty"`
	ActuallyPaid     float64 `json:"actually_paid,omitempty"`
	OutcomeAmount    float64 `json:"outcome_amount,omitempty"`
	OutcomeCurrency  string  `json:"outcome_currency,omitempty"`
}

// UnmarshalJSON accepts id and amounts as numbers or strings; the API sends
// "id": "4522625843" and "price_amount": "1000".
func (i *Invoice) UnmarshalJSON(data []byte) error {
	type plain Invoice
	aux := struct {
		*plain
		ID           FlexInt   `json:"id"`
		TokenID      FlexInt   `json:"token_id"`
		InvoiceID    FlexInt   `json:"invoice_id"`
		PriceAmount  FlexFloat `json:"price_amount"`
		PayAmount    FlexFloat `json:"pay_amount"`
		ActuallyPaid FlexFloat `json:"actually_paid"`
		Outcome      FlexFloat `json:"outcome_amount"`
	}{plain: (*plain)(i)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	i.ID = aux.ID.Int64()
	if aux.TokenID != 0 {
		i.TokenID = formatFlexInt(aux.TokenID)
	}
	if aux.InvoiceID != 0 {
		i.InvoiceID = formatFlexInt(aux.InvoiceID)
	}
	i.PriceAmount = aux.PriceAmount.Float64()
	i.PayAmount = aux.PayAmount.Float64()
	i.ActuallyPaid = aux.ActuallyPaid.Float64()
	i.OutcomeAmount = aux.Outcome.Float64()
	return nil
}

// CreateInvoiceRequest is the body for POST /v1/invoice.
type CreateInvoiceRequest struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency,omitempty"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	SuccessURL       string  `json:"success_url,omitempty"`
	CancelURL        string  `json:"cancel_url,omitempty"`
	PartiallyPaidURL string  `json:"partially_paid_url,omitempty"`
	IPNCallbackURL   string  `json:"ipn_callback_url,omitempty"`
	IsFixedRate      bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user,omitempty"`
}
