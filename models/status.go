package models

import (
	"bytes"
	"encoding/json"
	"fmt"
)

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

type BalanceResponse struct {
	Currencies map[string]BalanceCurrency `json:"-"` // keyed by currency code
}

// BalanceCurrency holds amount and pendingAmount for one currency.
type BalanceCurrency struct {
	Amount        float64 `json:"amount"`
	PendingAmount float64 `json:"pendingAmount"`
}

// UnmarshalJSON accepts amounts as numbers or strings.
func (b *BalanceCurrency) UnmarshalJSON(data []byte) error {
	var aux struct {
		Amount        FlexFloat `json:"amount"`
		PendingAmount FlexFloat `json:"pendingAmount"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	b.Amount, b.PendingAmount = aux.Amount.Float64(), aux.PendingAmount.Float64()
	return nil
}

// UnmarshalJSON reads the flat {"eth": {"amount": ..., "pendingAmount": ...}} object the API returns.
func (b *BalanceResponse) UnmarshalJSON(data []byte) error {
	return decodeBalances(data, &b.Currencies)
}

// decodeBalances reads a currency-keyed balance object. The API sends an empty
// balance as [] rather than {}.
func decodeBalances(data []byte, out *map[string]BalanceCurrency) error {
	if t := bytes.TrimSpace(data); len(t) > 0 && t[0] == '[' {
		var list []json.RawMessage
		if err := json.Unmarshal(t, &list); err != nil {
			return err
		}
		if len(list) > 0 {
			return fmt.Errorf("nowpay: balances sent as a non-empty array: %s", t)
		}
		*out = map[string]BalanceCurrency{}
		return nil
	}
	return json.Unmarshal(data, out)
}
