// Example: HTTP server that verifies NOWPayments IPN webhooks and parses events.
// Run with: IPN_SECRET=your_ipn_secret go run .
// Set your IPN callback URL in NOWPayments dashboard to http://your-host/webhook
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/muxover/nowpay-go/webhook"
)

func main() {
	secret := os.Getenv("IPN_SECRET")
	if secret == "" {
		log.Fatal("Set IPN_SECRET to your NOWPayments IPN secret")
	}

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		sig := r.Header.Get(webhook.SignatureHeader)
		if !webhook.VerifySignature(body, sig, secret) {
			log.Println("invalid signature")
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		ev, err := webhook.ParseEvent(body)
		if err != nil {
			log.Printf("parse event: %v", err)
			http.Error(w, "bad payload", http.StatusBadRequest)
			return
		}

		log.Printf("event: payment_id=%d status=%s order_id=%s", ev.PaymentID, ev.PaymentStatus, ev.OrderID)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	fmt.Println("Webhook server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
