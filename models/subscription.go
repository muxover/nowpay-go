package models

type Subscription struct {
	ID               string  `json:"id"`
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency,omitempty"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	SubscriptionID   string  `json:"subscription_id,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
	Status           string  `json:"status,omitempty"`
}

type CreateSubscriptionRequest struct {
	PriceAmount      float64 `json:"price_amount"`
	PriceCurrency    string  `json:"price_currency"`
	PayCurrency      string  `json:"pay_currency,omitempty"`
	OrderID          string  `json:"order_id,omitempty"`
	OrderDescription string  `json:"order_description,omitempty"`
	Period           string  `json:"period,omitempty"`
	CallbackURL      string  `json:"callback_url,omitempty"`
}

type UpdateSubscriptionRequest struct {
	PriceAmount      *float64 `json:"price_amount,omitempty"`
	PriceCurrency    string   `json:"price_currency,omitempty"`
	PayCurrency      string   `json:"pay_currency,omitempty"`
	OrderDescription string   `json:"order_description,omitempty"`
	Period           string   `json:"period,omitempty"`
	CallbackURL      string   `json:"callback_url,omitempty"`
	Status           string   `json:"status,omitempty"`
}
