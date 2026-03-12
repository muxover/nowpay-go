package models

// Estimate holds the result of a price/amount estimate (crypto or fiat).
type Estimate struct {
	EstimatedAmount   float64 `json:"estimated_amount"`
	PayAmount         float64 `json:"pay_amount,omitempty"`
	PayCurrency       string  `json:"pay_currency,omitempty"`
	PriceAmount       float64 `json:"price_amount,omitempty"`
	PriceCurrency     string  `json:"price_currency,omitempty"`
}

// EstimateRequest is used for estimate-by-price (e.g. 10 USD -> BTC).
type EstimateRequest struct {
	Amount        float64 `json:"amount"`
	CurrencyFrom  string  `json:"currency_from"`
	CurrencyTo    string  `json:"currency_to"`
}

// MinAmountResponse is the minimum payment amount for a currency.
type MinAmountResponse struct {
	MinAmount float64 `json:"min_amount"`
}
