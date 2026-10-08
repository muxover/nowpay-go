package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// SignatureHeader is the header name NOWPayments sends for IPN signature.
const SignatureHeader = "x-nowpayments-sig"

// VerifySignature reports whether signature (the x-nowpayments-sig header) is
// the HMAC-SHA512 of body, signed with the IPN secret.
//
// NOWPayments signs the body with its keys sorted, serialized the way
// JavaScript's JSON.stringify does it. The body is re-serialized the same way
// here, so the check works on the raw request body as received.
func VerifySignature(body []byte, signature string, secret string) bool {
	secret = strings.TrimSpace(secret)
	want, err := hex.DecodeString(strings.TrimSpace(signature))
	if secret == "" || err != nil || len(want) != sha512.Size {
		return false
	}
	v, err := decode(body)
	if err != nil {
		return false
	}

	// the Node.js sample in the docs turns arrays into index-keyed objects; accept that form too
	forms := []bool{false}
	if hasArray(v) {
		forms = append(forms, true)
	}
	for _, arraysAsObjects := range forms {
		mac := hmac.New(sha512.New, []byte(secret))
		mac.Write(canonical(v, arraysAsObjects))
		if hmac.Equal(mac.Sum(nil), want) {
			return true
		}
	}
	return false
}

// Sign returns the signature NOWPayments would send for body. It is useful
// for testing a webhook handler locally.
func Sign(body []byte, secret string) (string, error) {
	v, err := decode(body)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha512.New, []byte(strings.TrimSpace(secret)))
	mac.Write(canonical(v, false))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

func decode(body []byte) (any, error) {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return nil, errors.New("nowpay: empty webhook body")
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func hasArray(v any) bool {
	switch t := v.(type) {
	case []any:
		return true
	case map[string]any:
		for _, x := range t {
			if hasArray(x) {
				return true
			}
		}
	}
	return false
}

// canonical serializes v with sorted object keys, matching JSON.stringify.
func canonical(v any, arraysAsObjects bool) []byte {
	var buf bytes.Buffer
	writeValue(&buf, v, arraysAsObjects)
	return buf.Bytes()
}

func writeValue(buf *bytes.Buffer, v any, arraysAsObjects bool) {
	switch t := v.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		buf.WriteString(strconv.FormatBool(t))
	case float64:
		writeNumber(buf, t)
	case string:
		writeString(buf, t)
	case []any:
		if arraysAsObjects {
			m := make(map[string]any, len(t))
			for i, x := range t {
				m[strconv.Itoa(i)] = x
			}
			writeObject(buf, m, arraysAsObjects)
			return
		}
		buf.WriteByte('[')
		for i, x := range t {
			if i > 0 {
				buf.WriteByte(',')
			}
			writeValue(buf, x, arraysAsObjects)
		}
		buf.WriteByte(']')
	case map[string]any:
		writeObject(buf, t, arraysAsObjects)
	}
}

func writeObject(buf *bytes.Buffer, m map[string]any, arraysAsObjects bool) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		writeString(buf, k)
		buf.WriteByte(':')
		writeValue(buf, m[k], arraysAsObjects)
	}
	buf.WriteByte('}')
}

// writeNumber matches Number#toString; encoding/json already does, except for -0.
func writeNumber(buf *bytes.Buffer, f float64) {
	if f == 0 {
		buf.WriteByte('0')
		return
	}
	b, _ := json.Marshal(f)
	buf.Write(b)
}

// writeString quotes like JSON.stringify, which leaves <, >, &, U+2028 and U+2029 unescaped.
func writeString(buf *bytes.Buffer, s string) {
	const hexDigits = "0123456789abcdef"
	buf.WriteByte('"')
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == '"':
			buf.WriteString(`\"`)
		case r == '\\':
			buf.WriteString(`\\`)
		case r == '\b':
			buf.WriteString(`\b`)
		case r == '\f':
			buf.WriteString(`\f`)
		case r == '\n':
			buf.WriteString(`\n`)
		case r == '\r':
			buf.WriteString(`\r`)
		case r == '\t':
			buf.WriteString(`\t`)
		case r < 0x20:
			buf.WriteString(`\u00`)
			buf.WriteByte(hexDigits[r>>4])
			buf.WriteByte(hexDigits[r&0xf])
		default:
			buf.WriteString(s[i : i+size])
		}
		i += size
	}
	buf.WriteByte('"')
}
