package models

// Currency represents a supported or available currency.
type Currency struct {
	Code        string `json:"code"`
	Name        string `json:"name,omitempty"`
	Logo        string `json:"logo_url,omitempty"`
	Network     string `json:"network,omitempty"`
	IsFiat      bool   `json:"is_fiat,omitempty"`
	HasExternal bool   `json:"has_external,omitempty"`
}

// CurrenciesResponse is the response for supported/available currencies.
type CurrenciesResponse struct {
	Currencies []string `json:"currencies"`
}

// FullCurrenciesResponse returns full currency objects when available.
type FullCurrenciesResponse struct {
	Currencies []Currency `json:"currencies,omitempty"`
}
