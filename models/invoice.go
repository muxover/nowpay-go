package models

// Invoice represents a NOWPayments invoice.
type Invoice struct {
	ID                  int64   `json:"id"`
	InvoiceURL          string  `json:"invoice_url"`
	InvoiceID            string  `json:"invoice_id"`
	PriceAmount         float64 `json:"price_amount"`
	PriceCurrency       string  `json:"price_currency"`
	PayCurrency         string  `json:"pay_currency"`
	OrderID             string  `json:"order_id,omitempty"`
	OrderDescription    string  `json:"order_description,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	PaymentStatus       string  `json:"payment_status,omitempty"`
	PayAddress          string  `json:"pay_address,omitempty"`
	PayAmount           float64 `json:"pay_amount,omitempty"`
	ActuallyPaid        float64 `json:"actually_paid,omitempty"`
	OutcomeAmount       float64 `json:"outcome_amount,omitempty"`
	OutcomeCurrency     string  `json:"outcome_currency,omitempty"`
}

// CreateInvoiceRequest is the request body for creating an invoice.
type CreateInvoiceRequest struct {
	PriceAmount       float64 `json:"price_amount"`
	PriceCurrency     string  `json:"price_currency"`
	PayCurrency       string  `json:"pay_currency,omitempty"`
	OrderID           string  `json:"order_id,omitempty"`
	OrderDescription  string  `json:"order_description,omitempty"`
	SuccessURL        string  `json:"success_url,omitempty"`
	CancelURL         string  `json:"cancel_url,omitempty"`
	PartiallyPaidURL  string  `json:"partially_paid_url,omitempty"`
	IPNCallbackURL    string  `json:"ipn_callback_url,omitempty"`
	IsFixedRate       bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser   bool    `json:"is_fee_paid_by_user,omitempty"`
}
