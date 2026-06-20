package models

type Estimate struct {
	CurrencyFrom    string  `json:"currency_from,omitempty"`
	AmountFrom      float64 `json:"amount_from,omitempty"`
	CurrencyTo      string  `json:"currency_to,omitempty"`
	EstimatedAmount float64 `json:"estimated_amount"`
}

type MinAmountResponse struct {
	CurrencyFrom   string  `json:"currency_from,omitempty"`
	CurrencyTo     string  `json:"currency_to,omitempty"`
	MinAmount      float64 `json:"min_amount"`
	FiatEquivalent float64 `json:"fiat_equivalent,omitempty"`
}

// MinAmountParams holds optional query params for GET /v1/min-amount.
type MinAmountParams struct {
	CurrencyFrom    string
	CurrencyTo      string
	FiatEquivalent  string // e.g. "usd"
	IsFixedRate     bool
	IsFeePaidByUser bool
}
