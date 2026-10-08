package models

import "encoding/json"

// Currency is one entry from GET /v1/full-currencies.
type Currency struct {
	ID                       int     `json:"id,omitempty"`
	Code                     string  `json:"code"`
	Name                     string  `json:"name,omitempty"`
	Enable                   bool    `json:"enable,omitempty"`
	WalletRegex              string  `json:"wallet_regex,omitempty"`
	Priority                 int     `json:"priority,omitempty"`
	ExtraIDExists            bool    `json:"extra_id_exists,omitempty"`
	ExtraIDOptional          bool    `json:"extra_id_optional,omitempty"`
	ExtraIDRegex             *string `json:"extra_id_regex,omitempty"`
	LogoURL                  string  `json:"logo_url,omitempty"`
	Track                    bool    `json:"track,omitempty"`
	CgID                     string  `json:"cg_id,omitempty"`
	IsMaxlimit               bool    `json:"is_maxlimit,omitempty"`
	Network                  string  `json:"network,omitempty"`
	SmartContract            *string `json:"smart_contract,omitempty"`
	NetworkPrecision         *int    `json:"network_precision,omitempty"`
	Precision                int     `json:"precision,omitempty"`
	Ticker                   string  `json:"ticker,omitempty"`
	IsStable                 bool    `json:"is_stable,omitempty"`
	IsPopular                bool    `json:"is_popular,omitempty"`
	AvailableForPayment      bool    `json:"available_for_payment,omitempty"`
	AvailableForPayout       bool    `json:"available_for_payout,omitempty"`
	AvailableForToConversion bool    `json:"available_for_to_conversion,omitempty"`
}

// UnmarshalJSON accepts network_precision as a number or a string; the API
// sends "18".
func (c *Currency) UnmarshalJSON(data []byte) error {
	type plain Currency
	aux := struct {
		*plain
		NetworkPrecision *FlexInt `json:"network_precision"`
	}{plain: (*plain)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.NetworkPrecision = nil
	if aux.NetworkPrecision != nil {
		n := int(aux.NetworkPrecision.Int64())
		c.NetworkPrecision = &n
	}
	return nil
}

// CurrenciesResponse is a list of currency codes.
type CurrenciesResponse struct {
	Currencies []string `json:"currencies"`
}

// UnmarshalJSON reads the codes from every shape the currency endpoints use:
// {"currencies": ["btc"]}, {"currencies": [{"currency": "btc", ...}]} with
// fixed_rate=true, and {"selectedCurrencies": ["BTC"]} from /merchant/coins.
func (r *CurrenciesResponse) UnmarshalJSON(data []byte) error {
	var aux struct {
		Currencies         []json.RawMessage `json:"currencies"`
		SelectedCurrencies []string          `json:"selectedCurrencies"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Currencies = make([]string, 0, len(aux.Currencies)+len(aux.SelectedCurrencies))
	for _, raw := range aux.Currencies {
		var code string
		if err := json.Unmarshal(raw, &code); err == nil {
			r.Currencies = append(r.Currencies, code)
			continue
		}
		var obj struct {
			Currency string `json:"currency"`
			Code     string `json:"code"`
		}
		if err := json.Unmarshal(raw, &obj); err != nil {
			return err
		}
		if obj.Currency == "" {
			obj.Currency = obj.Code
		}
		r.Currencies = append(r.Currencies, obj.Currency)
	}
	r.Currencies = append(r.Currencies, aux.SelectedCurrencies...)
	return nil
}

// FullCurrenciesResponse is the response from GET /v1/full-currencies.
type FullCurrenciesResponse struct {
	Currencies []Currency `json:"currencies,omitempty"`
}
