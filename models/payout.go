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
