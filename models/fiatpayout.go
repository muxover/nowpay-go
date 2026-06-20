package models

// FiatProvider is a fiat-payout provider (GET /v1/fiat-payouts/providers).
type FiatProvider struct {
	Code    string `json:"code"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// FiatCurrency is a supported fiat currency for a provider.
type FiatCurrency struct {
	Provider     string `json:"provider"`
	CurrencyCode string `json:"currencyCode"`
	Enabled      bool   `json:"enabled"`
}

// FiatCryptoCurrency is a supported crypto currency for fiat payouts.
type FiatCryptoCurrency struct {
	Provider        string `json:"provider"`
	CurrencyCode    string `json:"currencyCode"`
	CurrencyNetwork string `json:"currencyNetwork"`
	Enabled         bool   `json:"enabled"`
}

// FiatPaymentMethodField describes one field required by a payment method.
type FiatPaymentMethodField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Mandatory   bool   `json:"mandatory"`
	Description string `json:"description"`
}

// FiatPaymentMethod is a fiat payout payment method.
type FiatPaymentMethod struct {
	Name        string                   `json:"name"`
	PaymentCode string                   `json:"paymentCode"`
	Provider    string                   `json:"provider"`
	Fields      []FiatPaymentMethodField `json:"fields"`
}

// CreateFiatAccountRequest is the body for POST /v1/fiat-payouts/account.
type CreateFiatAccountRequest struct {
	Currency    string            `json:"currency"`
	PaymentCode string            `json:"paymentCode"`
	Fields      map[string]string `json:"fields"`
	Provider    string            `json:"provider"`
	CountryCode string            `json:"countryCode"`
}

// FiatPayout represents a fiat payout (POST/GET /v1/fiat-payouts).
type FiatPayout struct {
	ID                   FlexInt `json:"id"`
	Provider             string  `json:"provider"`
	RequestID            string  `json:"requestId,omitempty"`
	Status               string  `json:"status"`
	FiatCurrencyCode     string  `json:"fiatCurrencyCode,omitempty"`
	FiatAmount           string  `json:"fiatAmount,omitempty"`
	CryptoCurrencyCode   string  `json:"cryptoCurrencyCode,omitempty"`
	CryptoCurrencyAmount string  `json:"cryptoCurrencyAmount,omitempty"`
	FiatAccountCode      string  `json:"fiatAccountCode,omitempty"`
	FiatAccountNumber    string  `json:"fiatAccountNumber,omitempty"`
	PayoutDescription    string  `json:"payoutDescription,omitempty"`
	Error                string  `json:"error,omitempty"`
	CreatedAt            string  `json:"createdAt,omitempty"`
	UpdatedAt            string  `json:"updatedAt,omitempty"`
}

// RequestFiatPayoutRequest is the body for POST /v1/fiat-payouts.
type RequestFiatPayoutRequest struct {
	FiatCurrency   string  `json:"fiatCurrency"`
	CryptoCurrency string  `json:"cryptoCurrency"`
	Amount         float64 `json:"amount"`
	Provider       string  `json:"provider"`
}

// FiatPayoutRows is the paginated envelope used by fiat-payout list endpoints ({"rows":[...]}).
type FiatPayoutRows struct {
	Rows []FiatPayout `json:"rows"`
}

// ListFiatPayoutsParams holds query params for GET /v1/fiat-payouts.
type ListFiatPayoutsParams struct {
	ID             string
	Provider       string
	RequestID      string
	FiatCurrency   string
	CryptoCurrency string
	Status         string
	Limit          int
	Page           int
	OrderBy        string
	SortBy         string
	DateFrom       string
	DateTo         string
}
