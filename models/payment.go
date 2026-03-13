package models

// Payment represents a NOWPayments payment.
type Payment struct {
	PaymentID           int64   `json:"payment_id"`
	PaymentStatus       string  `json:"payment_status"`
	PayAddress          string  `json:"pay_address"`
	PriceAmount         float64 `json:"price_amount"`
	PriceCurrency       string  `json:"price_currency"`
	PayAmount           float64 `json:"pay_amount"`
	PayCurrency         string  `json:"pay_currency"`
	OrderID             string  `json:"order_id"`
	OrderDescription    string  `json:"order_description"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	PurchaseID          string  `json:"purchase_id"`
	AmountReceived      float64 `json:"amount_received"`
	PayinExtraID        string  `json:"payin_extra_id"`
	SmartContract       string  `json:"smart_contract"`
	NetworkPrecision    int     `json:"network_precision"`
	TimeLimit           int     `json:"time_limit"`
	BurningPercent      *string `json:"burning_percent,omitempty"`
	ExpirationEstimate  string  `json:"expiration_estimate"`
	OutcomeAmount       float64 `json:"outcome_amount"`
	OutcomeCurrency     string  `json:"outcome_currency"`
}

// CreatePaymentRequest is the request body for creating a payment.
type CreatePaymentRequest struct {
	PriceAmount         float64 `json:"price_amount"`
	PriceCurrency       string  `json:"price_currency"`
	PayCurrency         string  `json:"pay_currency"`
	OrderID             string  `json:"order_id,omitempty"`
	OrderDescription    string  `json:"order_description,omitempty"`
	IPNCallbackURL      string  `json:"ipn_callback_url,omitempty"`
	SuccessURL          string  `json:"success_url,omitempty"`
	CancelURL           string  `json:"cancel_url,omitempty"`
	IsFixedRate         bool    `json:"is_fixed_rate,omitempty"`
}

// ListPaymentsParams holds optional query params for listing payments.
type ListPaymentsParams struct {
	Limit       int    `json:"limit,omitempty"`
	Page        int    `json:"page,omitempty"`
	SortBy      string `json:"sortBy,omitempty"`
	OrderBy     string `json:"orderBy,omitempty"`
	DateFrom    string `json:"dateFrom,omitempty"`
	DateTo      string `json:"dateTo,omitempty"`
	InvoiceID   string `json:"invoiceId,omitempty"`
}

// CreateInvoicePaymentRequest is the request for creating a payment from an invoice (POST /v1/invoice-payment).
type CreateInvoicePaymentRequest struct {
	IID               int64   `json:"iid"` // invoice id
	PayCurrency       string  `json:"pay_currency"`
	PurchaseID        string  `json:"purchase_id,omitempty"`
	OrderDescription  string  `json:"order_description,omitempty"`
	CustomerEmail     string  `json:"customer_email,omitempty"`
	PayoutAddress     string  `json:"payout_address,omitempty"`
	PayoutExtraID     string  `json:"payout_extra_id,omitempty"`
	PayoutCurrency    string  `json:"payout_currency,omitempty"`
}

// UpdateMerchantEstimateResponse is the response from POST /v1/payment/:id/update-merchant-estimate.
type UpdateMerchantEstimateResponse struct {
	ID                     int64   `json:"id,omitempty"`
	TokenID                int64   `json:"token_id,omitempty"`
	PayAmount              float64 `json:"pay_amount,omitempty"`
	ExpirationEstimateDate string  `json:"expiration_estimate_date,omitempty"`
}

// PaymentFlow holds detailed payment flow/processing info (e.g. confirmations).
type PaymentFlow struct {
	PaymentID     int64    `json:"payment_id,omitempty"`
	PaymentStatus string   `json:"payment_status,omitempty"`
	Confirmations []string `json:"confirmations,omitempty"`
	// Extend with API-specific fields as needed.
}

// RefundRequest is the request body for refunding a payment.
type RefundRequest struct {
	PaymentID int64   `json:"payment_id"`
	RefundType string `json:"refund_type,omitempty"` // e.g. "full" or "partial"
	Amount    float64 `json:"amount,omitempty"`     // for partial refund
}

// RefundResponse is the response from a refund request.
type RefundResponse struct {
	RefundID   string `json:"refund_id,omitempty"`
	PaymentID  int64  `json:"payment_id,omitempty"`
	Status     string `json:"status,omitempty"`
}
