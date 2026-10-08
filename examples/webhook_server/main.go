// Run with: IPN_SECRET=your_ipn_secret go run .
// Set your IPN callback URL in NOWPayments dashboard to http://your-host/webhook
package main

import (
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

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

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

		switch ev.Type {
		case webhook.EventPaymentFinished:
			log.Printf("payment %d finished: fulfil order %s", ev.PaymentID, ev.OrderID)
		case webhook.EventPaymentPartiallyPaid:
			log.Printf("payment %d partially paid: %v of %v %s", ev.PaymentID, ev.ActuallyPaid, ev.PayAmount, ev.PayCurrency)
		case webhook.EventPaymentFailed:
			log.Printf("payment %d %s: order %s not paid", ev.PaymentID, ev.PaymentStatus, ev.OrderID)
		default:
			log.Printf("payment %d is %s", ev.PaymentID, ev.PaymentStatus)
		}

		w.WriteHeader(http.StatusOK)
	})

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	fmt.Println("Webhook server listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
