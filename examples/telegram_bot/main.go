// Example: minimal Telegram bot flow using NowPay Go (create payment + send message with button).
// Run with: API_KEY=your_key BOT_TOKEN=your_bot_token go run .
// This example does not run a full bot; it demonstrates creating a payment and building the message + keyboard.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/muxover/nowpay-go"
	"github.com/muxover/nowpay-go/models"
	"github.com/muxover/nowpay-go/telegram"
)

func main() {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Set API_KEY to your NOWPayments API key")
		os.Exit(1)
	}
	_ = os.Getenv("BOT_TOKEN") // optional for this demo

	client := nowpay.NewClient(nowpay.Config{APIKey: apiKey})
	ctx := context.Background()

	req := &models.CreatePaymentRequest{
		PriceAmount:   5.00,
		PriceCurrency: "usd",
		PayCurrency:   "btc",
		OrderID:       "tg-order-001",
	}

	payment, err := client.Payments.Create(ctx, req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Create payment: %v\n", err)
		os.Exit(1)
	}

	msg := telegram.InvoiceMessage(payment)
	fmt.Println("Send this message to the user:")
	fmt.Println(msg)

	// Build payment URL (e.g. success_url or a custom checkout page). Here we use the pay address as link.
	payURL := "https://nowpayments.io/payment/?i=" + fmt.Sprint(payment.PaymentID)
	kb := telegram.PaymentKeyboard("Pay with crypto", payURL)
	fmt.Println("\nInline keyboard (JSON):")
	fmt.Printf("%+v\n", kb)
}
