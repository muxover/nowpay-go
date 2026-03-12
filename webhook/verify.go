package webhook

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// SignatureHeader is the header name NOWPayments sends for IPN signature.
const SignatureHeader = "x-nowpayments-sig"

// VerifySignature verifies the webhook body against the given signature using the IPN secret.
// Body is the raw request body; signature is the value of x-nowpayments-sig header.
// Returns true only if the HMAC-SHA512 of the sorted JSON body matches the signature.
func VerifySignature(body []byte, signature string, secret string) bool {
	if secret == "" || signature == "" {
		return false
	}
	computed := computeSignature(body, secret)
	return hmac.Equal([]byte(computed), []byte(signature))
}

func computeSignature(body []byte, secret string) string {
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		return ""
	}
	sorted := canonicalJSON(m)
	h := hmac.New(sha512.New, []byte(secret))
	h.Write(sorted)
	return hex.EncodeToString(h.Sum(nil))
}

// canonicalJSON marshals m with keys sorted alphabetically (NOWPayments IPN verification).
func canonicalJSON(m map[string]interface{}) []byte {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf := make([]byte, 0, 256)
	buf = append(buf, '{')
	for i, k := range keys {
		if i > 0 {
			buf = append(buf, ',')
		}
		kb, _ := json.Marshal(k)
		buf = append(buf, kb...)
		buf = append(buf, ':')
		v := m[k]
		vb, err := json.Marshal(v)
		if err != nil {
			vb = []byte("null")
		}
		buf = append(buf, vb...)
	}
	buf = append(buf, '}')
	return buf
}
