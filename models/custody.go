package models

import "encoding/json"

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

// UnmarshalJSON accepts an empty balance sent as [].
func (c *CustomerBalance) UnmarshalJSON(data []byte) error {
	var aux struct {
		SubPartnerID FlexInt         `json:"subPartnerId"`
		Balances     json.RawMessage `json:"balances"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.SubPartnerID = aux.SubPartnerID
	c.Balances = nil
	if len(aux.Balances) == 0 || string(aux.Balances) == "null" {
		return nil
	}
	return decodeBalances(aux.Balances, &c.Balances)
}

// ListCustomersParams holds query params for GET /v1/sub-partner.
type ListCustomersParams struct {
	ID     string
	Offset int
	Limit  int
	Order  string
}

// CustodyTransfer is a movement between custody balances (GET /v1/sub-partner/transfers).
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
