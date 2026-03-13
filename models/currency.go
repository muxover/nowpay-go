package models

// Currency represents a supported or available currency (full detail from /v1/full-currencies).
type Currency struct {
	ID               int     `json:"id,omitempty"`
	Code             string  `json:"code"`
	Name             string  `json:"name,omitempty"`
	Enable           bool    `json:"enable,omitempty"`
	WalletRegex      string  `json:"wallet_regex,omitempty"`
	Priority         int     `json:"priority,omitempty"`
	ExtraIDExists    bool    `json:"extra_id_exists,omitempty"`
	LogoURL          string  `json:"logo_url,omitempty"`
	Network          string  `json:"network,omitempty"`
	SmartContract    *string `json:"smart_contract,omitempty"`
	NetworkPrecision *int    `json:"network_precision,omitempty"`
}

// CurrenciesResponse is the response for supported/available currencies.
type CurrenciesResponse struct {
	Currencies []string `json:"currencies"`
}

// FullCurrenciesResponse returns full currency objects when available.
type FullCurrenciesResponse struct {
	Currencies []Currency `json:"currencies,omitempty"`
}
