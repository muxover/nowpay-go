package models

import "encoding/json"

// Estimate is the response from GET /v1/estimate.
type Estimate struct {
	CurrencyFrom    string  `json:"currency_from,omitempty"`
	AmountFrom      float64 `json:"amount_from,omitempty"`
	CurrencyTo      string  `json:"currency_to,omitempty"`
	EstimatedAmount float64 `json:"estimated_amount"`
}

// UnmarshalJSON accepts amounts as numbers or strings; the API sends
// "estimated_amount": "0.00118682".
func (e *Estimate) UnmarshalJSON(data []byte) error {
	type plain Estimate
	aux := struct {
		*plain
		AmountFrom      FlexFloat `json:"amount_from"`
		EstimatedAmount FlexFloat `json:"estimated_amount"`
	}{plain: (*plain)(e)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	e.AmountFrom = aux.AmountFrom.Float64()
	e.EstimatedAmount = aux.EstimatedAmount.Float64()
	return nil
}

// MinAmountResponse is the response from GET /v1/min-amount.
type MinAmountResponse struct {
	CurrencyFrom   string  `json:"currency_from,omitempty"`
	CurrencyTo     string  `json:"currency_to,omitempty"`
	MinAmount      float64 `json:"min_amount"`
	FiatEquivalent float64 `json:"fiat_equivalent,omitempty"`
}

// UnmarshalJSON accepts amounts as numbers or strings.
func (m *MinAmountResponse) UnmarshalJSON(data []byte) error {
	type plain MinAmountResponse
	aux := struct {
		*plain
		MinAmount      FlexFloat `json:"min_amount"`
		FiatEquivalent FlexFloat `json:"fiat_equivalent"`
	}{plain: (*plain)(m)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	m.MinAmount = aux.MinAmount.Float64()
	m.FiatEquivalent = aux.FiatEquivalent.Float64()
	return nil
}

// MinAmountParams holds optional query params for GET /v1/min-amount.
type MinAmountParams struct {
	CurrencyFrom    string
	CurrencyTo      string
	FiatEquivalent  string // e.g. "usd"
	IsFixedRate     bool
	IsFeePaidByUser bool
}
