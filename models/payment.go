package models

// Payment represents a NOWPayments payment (POST /v1/payment, GET /v1/payment/:id).
type Payment struct {
	PaymentID              FlexInt   `json:"payment_id"`
	InvoiceID              FlexInt   `json:"invoice_id,omitempty"`
	PaymentStatus          string    `json:"payment_status"`
	PayAddress             string    `json:"pay_address"`
	PayinExtraID           string    `json:"payin_extra_id,omitempty"`
	PriceAmount            float64   `json:"price_amount"`
	PriceCurrency          string    `json:"price_currency"`
	PayAmount              float64   `json:"pay_amount"`
	ActuallyPaid           float64   `json:"actually_paid,omitempty"`
	PayCurrency            string    `json:"pay_currency"`
	OrderID                string    `json:"order_id,omitempty"`
	OrderDescription       string    `json:"order_description,omitempty"`
	IPNCallbackURL         string    `json:"ipn_callback_url,omitempty"`
	CreatedAt              string    `json:"created_at"`
	UpdatedAt              string    `json:"updated_at"`
	PurchaseID             FlexInt   `json:"purchase_id,omitempty"`
	AmountReceived         float64   `json:"amount_received,omitempty"`
	SmartContract          string    `json:"smart_contract,omitempty"`
	Network                string    `json:"network,omitempty"`
	NetworkPrecision       int       `json:"network_precision,omitempty"`
	TimeLimit              *int      `json:"time_limit,omitempty"`
	BurningPercent         *string   `json:"burning_percent,omitempty"`
	ExpirationEstimateDate string    `json:"expiration_estimate_date,omitempty"`
	OutcomeAmount          float64   `json:"outcome_amount,omitempty"`
	OutcomeCurrency        string    `json:"outcome_currency,omitempty"`
	PayoutHash             string    `json:"payout_hash,omitempty"`
	PayinHash              string    `json:"payin_hash,omitempty"`
	Type                   string    `json:"type,omitempty"`
	PaymentExtraIDs        []FlexInt `json:"payment_extra_ids,omitempty"`
}

type CreatePaymentRequest struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	IPNCallbackURL   string  `json:"ipn_callback_url,omitempty"`
	IsFixedRate      bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser  bool    `json:"is_fee_paid_by_user,omitempty"`
}

type ListPaymentsParams struct {
	Limit     int    `json:"limit,omitempty"`
	Page      int    `json:"page,omitempty"`
	SortBy    string `json:"sortBy,omitempty"`
	OrderBy   string `json:"orderBy,omitempty"`
	DateFrom  string `json:"dateFrom,omitempty"`
	DateTo    string `json:"dateTo,omitempty"`
	InvoiceID string `json:"invoiceId,omitempty"`
}

// ListPaymentsResponse is the envelope returned by GET /v1/payment/ ({"data":[...]}).
type ListPaymentsResponse struct {
	Data []Payment `json:"data"`
}

// CreateInvoicePaymentRequest is the request for creating a payment from an invoice (POST /v1/invoice-payment).
type CreateInvoicePaymentRequest struct {
	IID              int64  `json:"iid"` // invoice id
	PayCurrency      string `json:"pay_currency"`
	PurchaseID       string `json:"purchase_id,omitempty"`
	OrderDescription string `json:"order_description,omitempty"`
	CustomerEmail    string `json:"customer_email,omitempty"`
	PayoutAddress    string `json:"payout_address,omitempty"`
	PayoutExtraID    string `json:"payout_extra_id,omitempty"`
	PayoutCurrency   string `json:"payout_currency,omitempty"`
}

// UpdateMerchantEstimateResponse is the response from POST /v1/payment/:id/update-merchant-estimate.
type UpdateMerchantEstimateResponse struct {
	ID                     FlexInt `json:"id,omitempty"`
	TokenID                FlexInt `json:"token_id,omitempty"`
	PayAmount              float64 `json:"pay_amount,omitempty"`
	ExpirationEstimateDate string  `json:"expiration_estimate_date,omitempty"`
}
