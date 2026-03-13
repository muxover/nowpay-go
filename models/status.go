package models

import "encoding/json"

// StatusResponse is the response from the API status/health endpoint.
type StatusResponse struct {
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

// AuthRequest is the request body for POST /v1/auth (email + password).
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the response from POST /v1/auth (JWT token).
type AuthResponse struct {
	Token string `json:"token"`
}

// BalanceResponse is the response from GET /v1/balance (balance per currency).
type BalanceResponse struct {
	Currencies map[string]BalanceCurrency `json:"-"` // keyed by currency code
}

// BalanceCurrency holds amount and pendingAmount for one currency.
type BalanceCurrency struct {
	Amount        float64 `json:"amount"`
	PendingAmount float64 `json:"pendingAmount"`
}

// UnmarshalJSON implements custom unmarshal for BalanceResponse (API returns flat object like {"eth":{"amount":0.1,"pendingAmount":0}}).
func (b *BalanceResponse) UnmarshalJSON(data []byte) error {
	var m map[string]struct {
		Amount        float64 `json:"amount"`
		PendingAmount float64 `json:"pendingAmount"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	b.Currencies = make(map[string]BalanceCurrency)
	for k, v := range m {
		b.Currencies[k] = BalanceCurrency{Amount: v.Amount, PendingAmount: v.PendingAmount}
	}
	return nil
}
