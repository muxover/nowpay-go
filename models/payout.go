package models

// Payout represents a payout record.
type Payout struct {
	ID            int64   `json:"id"`
	WithdrawalID  string  `json:"withdrawal_id"`
	PayoutHash    string  `json:"payout_hash,omitempty"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Status        string  `json:"status"`
	Address       string  `json:"address"`
	ExtraID       string  `json:"extra_id,omitempty"`
	CreatedAt     string  `json:"created_at,omitempty"`
	UpdatedAt     string  `json:"updated_at,omitempty"`
	BatchWithdrawalID string `json:"batch_withdrawal_id,omitempty"`
}

// CreatePayoutRequest is the request body for creating a payout.
type CreatePayoutRequest struct {
	WithdrawalID string  `json:"withdrawal_id"`
	Address      string  `json:"address"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	ExtraID      string  `json:"extra_id,omitempty"`
}

// BatchPayoutRequest is the request body for creating multiple payouts (mass payouts).
type BatchPayoutRequest struct {
	Withdrawals []CreatePayoutRequest `json:"withdrawals"`
}

// BatchPayoutResponse is the response from a batch payout request.
type BatchPayoutResponse struct {
	BatchWithdrawalID string   `json:"batch_withdrawal_id,omitempty"`
	Withdrawals       []Payout `json:"withdrawals,omitempty"`
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
