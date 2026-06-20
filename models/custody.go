package models

// Customer is a custody sub-account / customer (/v1/sub-partner).
type Customer struct {
	ID        FlexInt `json:"id"`
	Name      string  `json:"name"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

// CreateCustomerRequest is the body for POST /v1/sub-partner/balance.
type CreateCustomerRequest struct {
	Name string `json:"name"`
}

// CustomerBalance holds a customer's per-currency balances (GET /v1/sub-partner/balance/:id).
type CustomerBalance struct {
	SubPartnerID FlexInt                    `json:"subPartnerId"`
	Balances     map[string]BalanceCurrency `json:"balances"`
}

// ListCustomersParams holds query params for GET /v1/sub-partner.
type ListCustomersParams struct {
	ID     string
	Offset int
	Limit  int
	Order  string
}

// CustodyTransfer is a transfer/deposit/write-off between custody balances.
type CustodyTransfer struct {
	ID        FlexInt   `json:"id"`
	FromSubID FlexInt   `json:"from_sub_id,omitempty"`
	ToSubID   FlexInt   `json:"to_sub_id,omitempty"`
	Status    string    `json:"status"`
	Amount    FlexFloat `json:"amount"`
	Currency  string    `json:"currency"`
	CreatedAt string    `json:"created_at,omitempty"`
	UpdatedAt string    `json:"updated_at,omitempty"`
}

// TransferRequest is the body for POST /v1/sub-partner/transfer.
type TransferRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	FromID   int64   `json:"from_id"`
	ToID     int64   `json:"to_id"`
}

// CustodyMoveRequest is the body for POST /v1/sub-partner/deposit and /write-off.
type CustodyMoveRequest struct {
	Currency     string  `json:"currency"`
	Amount       float64 `json:"amount"`
	SubPartnerID int64   `json:"sub_partner_id"`
}

// CustodyDepositRequest is the body for POST /v1/sub-partner/payment (deposit with payment).
type CustodyDepositRequest struct {
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	SubPartnerID    string  `json:"sub_partner_id"`
	IsFixedRate     bool    `json:"is_fixed_rate,omitempty"`
	IsFeePaidByUser bool    `json:"is_fee_paid_by_user,omitempty"`
	IPNCallbackURL  string  `json:"ipn_callback_url,omitempty"`
}

// ListTransfersParams holds query params for GET /v1/sub-partner/transfers.
type ListTransfersParams struct {
	ID     string
	Status string
	Limit  int
	Offset int
	Order  string
}
