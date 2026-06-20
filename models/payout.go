package models

// Payout represents a single withdrawal item as returned by the payout endpoints.
type Payout struct {
	ID                FlexInt   `json:"id"`
	BatchWithdrawalID FlexInt   `json:"batch_withdrawal_id,omitempty"`
	Address           string    `json:"address"`
	Currency          string    `json:"currency"`
	Amount            FlexFloat `json:"amount"`
	Status            string    `json:"status"`
	ExtraID           string    `json:"extra_id,omitempty"`
	Hash              string    `json:"hash,omitempty"`
	Error             string    `json:"error,omitempty"`
	IPNCallbackURL    string    `json:"ipn_callback_url,omitempty"`
	PayoutDescription string    `json:"payout_description,omitempty"`
	UniqueExternalID  string    `json:"unique_external_id,omitempty"`
	IsRequestPayouts  bool      `json:"is_request_payouts,omitempty"`
	CreatedAt         string    `json:"created_at,omitempty"`
	RequestedAt       string    `json:"requested_at,omitempty"`
	UpdatedAt         string    `json:"updated_at,omitempty"`
}

// PayoutWithdrawal is a single withdrawal entry in a CreatePayoutRequest.
type PayoutWithdrawal struct {
	Address           string  `json:"address"`
	Currency          string  `json:"currency"`
	Amount            float64 `json:"amount"`
	ExtraID           string  `json:"extra_id,omitempty"`
	IPNCallbackURL    string  `json:"ipn_callback_url,omitempty"`
	FiatAmount        float64 `json:"fiat_amount,omitempty"`
	FiatCurrency      string  `json:"fiat_currency,omitempty"`
	UniqueExternalID  string  `json:"unique_external_id,omitempty"`
	PayoutDescription string  `json:"payout_description,omitempty"`
}

// CreatePayoutRequest is the body for POST /v1/payout. At least one withdrawal
// is required. ExecuteAt + Interval enable scheduled payouts (interval "ONETIME").
type CreatePayoutRequest struct {
	PayoutDescription string             `json:"payout_description,omitempty"`
	IPNCallbackURL    string             `json:"ipn_callback_url,omitempty"`
	ExecuteAt         string             `json:"execute_at,omitempty"` // RFC3339, e.g. 2026-02-29T10:35:00.000Z
	Interval          string             `json:"interval,omitempty"`   // currently only "ONETIME"
	Withdrawals       []PayoutWithdrawal `json:"withdrawals"`
}

// PayoutBatch is the response from POST /v1/payout: a batch id plus its withdrawals.
type PayoutBatch struct {
	ID          FlexInt  `json:"id"`
	Withdrawals []Payout `json:"withdrawals"`
}

// ListPayoutsResponse is the envelope returned by GET /v1/payout ({"payouts":[...]}).
type ListPayoutsResponse struct {
	Payouts []Payout `json:"payouts"`
}

// VerifyPayoutRequest is the body for POST /v1/payout/:batch-withdrawal-id/verify.
type VerifyPayoutRequest struct {
	VerificationCode string `json:"verification_code"`
}

// ValidateAddressRequest is the request for POST /v1/payout/validate-address.
type ValidateAddressRequest struct {
	Address  string `json:"address"`
	Currency string `json:"currency"`
	ExtraID  string `json:"extra_id,omitempty"`
}

// ListPayoutsParams holds query params for GET /v1/payout (list payouts).
type ListPayoutsParams struct {
	BatchID  string
	Status   string
	OrderBy  string
	Order    string
	DateFrom string
	DateTo   string
	Limit    int
	Page     int
}
