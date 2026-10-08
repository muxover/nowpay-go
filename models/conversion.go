package models

// Conversion represents a currency conversion (/v1/conversion).
type Conversion struct {
	ID           FlexInt   `json:"id"`
	Status       string    `json:"status"`
	FromCurrency string    `json:"from_currency"`
	ToCurrency   string    `json:"to_currency"`
	FromAmount   FlexFloat `json:"from_amount"`
	ToAmount     FlexFloat `json:"to_amount"`
	CreatedAt    string    `json:"created_at,omitempty"`
	UpdatedAt    string    `json:"updated_at,omitempty"`
}

// ListConversionsParams holds query params for GET /v1/conversion.
type ListConversionsParams struct {
	ID           string
	Status       string
	FromCurrency string
	ToCurrency   string
	CreatedFrom  string
	CreatedTo    string
	Limit        int
	Offset       int
	Order        string
}
