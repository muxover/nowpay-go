package models

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// FlexInt is an integer that decodes from either a JSON number or a quoted
// string. NOWPayments returns IDs inconsistently (e.g. "5745459419" on create
// vs 6249365965 on read), so identifier fields use this type.
type FlexInt int64

// Int64 returns the underlying value.
func (f FlexInt) Int64() int64 { return int64(f) }

func (f FlexInt) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(f), 10)), nil
}

func (f *FlexInt) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if s == "" {
			*f = 0
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// Fall back to float-formatted strings like "200.0".
			fl, ferr := strconv.ParseFloat(s, 64)
			if ferr != nil {
				return err
			}
			*f = FlexInt(int64(fl))
			return nil
		}
		*f = FlexInt(n)
		return nil
	}
	var n int64
	if err := json.Unmarshal(b, &n); err == nil {
		*f = FlexInt(n)
		return nil
	}
	var fl float64
	if err := json.Unmarshal(b, &fl); err != nil {
		return err
	}
	*f = FlexInt(int64(fl))
	return nil
}

// FlexFloat is a float that decodes from either a JSON number or a quoted
// string. NOWPayments returns payout amounts as strings (e.g. "94.088939").
type FlexFloat float64

// Float64 returns the underlying value.
func (f FlexFloat) Float64() float64 { return float64(f) }

func (f FlexFloat) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(f), 'f', -1, 64)), nil
}

func (f *FlexFloat) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = 0
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		if s == "" {
			*f = 0
			return nil
		}
		fl, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		*f = FlexFloat(fl)
		return nil
	}
	var fl float64
	if err := json.Unmarshal(b, &fl); err != nil {
		return err
	}
	*f = FlexFloat(fl)
	return nil
}
