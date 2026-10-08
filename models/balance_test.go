package models

import (
	"encoding/json"
	"testing"
)

func TestBalancesDecode(t *testing.T) {
	var empty BalanceResponse
	if err := json.Unmarshal([]byte(` [ ] `), &empty); err != nil || empty.Currencies == nil || len(empty.Currencies) != 0 {
		t.Errorf("empty balance: %+v, %v", empty, err)
	}
	var full BalanceResponse
	if err := json.Unmarshal([]byte(`{"btc":{"amount":"0.5","pendingAmount":0}}`), &full); err != nil || full.Currencies["btc"].Amount != 0.5 {
		t.Errorf("balance: %+v, %v", full, err)
	}
	var odd CustomerBalance
	if err := json.Unmarshal([]byte(`{"subPartnerId":1,"balances":[{"amount":1}]}`), &odd); err == nil {
		t.Error("a non-empty balances array should be an error, not silently dropped")
	}
}
