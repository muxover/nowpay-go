package webhook

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type signatureCase struct {
	Name string `json:"name"`
	Body string `json:"body"`
	Keep string `json:"keep"`
	Docs string `json:"docs"`
}

// ../testdata/ipn_signatures.json holds bodies signed by Node.js with
// crypto.createHmac and JSON.stringify, the way NOWPayments signs IPNs.
func loadSignatures(t *testing.T) (string, []signatureCase) {
	t.Helper()
	b, err := os.ReadFile("../testdata/ipn_signatures.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Secret string          `json:"secret"`
		Cases  []signatureCase `json:"cases"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f.Secret, f.Cases
}

func TestVerifySignatureMatchesNode(t *testing.T) {
	secret, cases := loadSignatures(t)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			body := []byte(c.Body)
			if !VerifySignature(body, c.Keep, secret) {
				t.Error("rejected the JSON.stringify signature")
			}
			if !VerifySignature(body, c.Docs, secret) {
				t.Error("rejected the signature from the docs' sortObject")
			}
			if !VerifySignature(body, strings.ToUpper(c.Keep), secret) {
				t.Error("rejected an upper-case signature")
			}
			if !VerifySignature(body, c.Keep, " "+secret+"\n") {
				t.Error("rejected a secret with surrounding whitespace")
			}
			if VerifySignature(body, c.Keep, secret+"x") {
				t.Error("accepted the wrong secret")
			}
			got, err := Sign(body, secret)
			if err != nil || got != c.Keep {
				t.Errorf("Sign = %s, %v; want %s", got, err, c.Keep)
			}
		})
	}
}

func TestVerifySignatureIgnoresFormatting(t *testing.T) {
	secret, cases := loadSignatures(t)
	c := cases[0]
	var v any
	if err := json.Unmarshal([]byte(c.Body), &v); err != nil {
		t.Fatal(err)
	}
	pretty, _ := json.MarshalIndent(v, "", "    ")
	if !VerifySignature(pretty, c.Keep, secret) {
		t.Error("rejected the same payload with different whitespace and key order")
	}
}

func TestVerifySignatureRejects(t *testing.T) {
	secret, cases := loadSignatures(t)
	c := cases[0]
	tampered := strings.Replace(c.Body, `"pay_amount":15`, `"pay_amount":16`, 1)
	for name, tc := range map[string]struct{ body, sig, secret string }{
		"tampered body": {tampered, c.Keep, secret},
		"empty secret":  {c.Body, c.Keep, ""},
		"empty sig":     {c.Body, "", secret},
		"non-hex sig":   {c.Body, "zz" + c.Keep[2:], secret},
		"short sig":     {c.Body, c.Keep[:64], secret},
		"invalid json":  {"{", c.Keep, secret},
		"empty body":    {"", c.Keep, secret},
	} {
		if VerifySignature([]byte(tc.body), tc.sig, tc.secret) {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestNegativeZeroSerializesLikeJavaScript(t *testing.T) {
	a, err := Sign([]byte(`{"a":-0.0}`), "s")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Sign([]byte(`{"a":0}`), "s")
	if a != b {
		t.Error("-0 and 0 sign differently; JSON.stringify writes both as 0")
	}
}

func TestParsePaymentEvent(t *testing.T) {
	body, err := os.ReadFile("../testdata/ipn_payment.json")
	if err != nil {
		t.Fatal(err)
	}
	ev, err := ParseEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Type != EventPaymentFinished || ev.PaymentID != 123456789 || ev.ParentPaymentID != 987654321 || ev.OutcomeAmount != 14.8106 {
		t.Errorf("event = %+v", ev)
	}
	p, err := ev.AsPaymentEvent()
	if err != nil {
		t.Fatal(err)
	}
	if p.PurchaseID != "123456789" || p.Fee == nil || p.Fee.Currency != "btc" || p.PayAmount != 15 {
		t.Errorf("payment event = %+v", p)
	}
}

func TestParseEventAcceptsNumbersOrStrings(t *testing.T) {
	body := []byte(`{"payment_id":"5745459419","invoice_id":4522625843,"order_id":12345,"purchase_id":5837122679,"payment_status":"partially_paid","pay_amount":"0.5","actually_paid":"0.25"}`)
	ev, err := ParseEvent(body)
	if err != nil {
		t.Fatal(err)
	}
	if ev.PaymentID != 5745459419 || ev.InvoiceID != "4522625843" || ev.OrderID != "12345" || ev.PayAmount != 0.5 {
		t.Errorf("event = %+v", ev)
	}
	if ev.Type != EventPaymentPartiallyPaid {
		t.Errorf("type = %q", ev.Type)
	}
	p, err := ev.AsPaymentEvent()
	if err != nil {
		t.Fatal(err)
	}
	if p.InvoiceID != 4522625843 || p.PurchaseID != "5837122679" || p.ActuallyPaid != 0.25 {
		t.Errorf("payment event = %+v", p)
	}
	inv, err := ev.AsInvoiceEvent()
	if err != nil {
		t.Fatal(err)
	}
	if inv.InvoiceID != "4522625843" || inv.OrderID != "12345" {
		t.Errorf("invoice event = %+v", inv)
	}
}

func TestEventTypes(t *testing.T) {
	for status, want := range map[string]string{
		"waiting":        EventPaymentCreated,
		"confirming":     EventPaymentCreated,
		"confirmed":      EventPaymentConfirmed,
		"sending":        EventPaymentConfirmed,
		"partially_paid": EventPaymentPartiallyPaid,
		"finished":       EventPaymentFinished,
		"failed":         EventPaymentFailed,
		"expired":        EventPaymentFailed,
		"refunded":       EventPaymentRefunded,
		"something_new":  "",
	} {
		ev, err := ParseEvent([]byte(`{"payment_id":1,"payment_status":"` + status + `"}`))
		if err != nil {
			t.Fatal(err)
		}
		if ev.Type != want {
			t.Errorf("%s: type = %q, want %q", status, ev.Type, want)
		}
	}
}

func TestParseEventInvalid(t *testing.T) {
	for _, body := range []string{"", "{", "[1,2]", `{"payment_id":"abc"}`} {
		if _, err := ParseEvent([]byte(body)); err == nil {
			t.Errorf("%q: want an error", body)
		}
	}
}

func TestEventWithoutBody(t *testing.T) {
	var ev Event
	if _, err := ev.AsPaymentEvent(); err == nil {
		t.Error("AsPaymentEvent: want an error for an Event not made by ParseEvent")
	}
	if _, err := ev.AsInvoiceEvent(); err == nil {
		t.Error("AsInvoiceEvent: want an error for an Event not made by ParseEvent")
	}
}
