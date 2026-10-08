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

// ListPayoutsResponse is the envelope returned by GET /v1/payout ({"payouts":[...]}).
type ListPayoutsResponse struct {
	Payouts []Payout `json:"payouts"`
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
