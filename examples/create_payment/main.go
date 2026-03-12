// Example: create a payment using NowPay Go.
// Run with: API_KEY=your_key go run .
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/muxover/nowpay-go"
	"github.com/muxover/nowpay-go/models"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Set API_KEY to your NOWPayments API key")
		os.Exit(1)
	}

	client := nowpay.NewClient(nowpay.Config{APIKey: apiKey})
	ctx := context.Background()

	req := &models.CreatePaymentRequest{
		PriceAmount:   10.00,
		PriceCurrency: "usd",
		PayCurrency:   "btc",
		OrderID:       "example-order-001",
	}

	payment, err := client.Payments.Create(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create payment: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Payment ID: %d\n", payment.PaymentID)
	fmt.Printf("Status: %s\n", payment.PaymentStatus)
	fmt.Printf("Pay amount: %f %s\n", payment.PayAmount, payment.PayCurrency)
	fmt.Printf("Address: %s\n", payment.PayAddress)
}
