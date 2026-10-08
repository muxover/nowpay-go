// Package nowpay is a Go client for the NOWPayments crypto payment API:
// payments, invoices, currencies, estimates, recurring payments, custody
// customers, and the payout, conversion, and fiat-payout lookups.
//
// Create a client with NewClient and call the services on it:
//
//	client := nowpay.NewClient(nowpay.Config{APIKey: os.Getenv("NOWPAYMENTS_API_KEY")})
//	payment, err := client.Payments.Create(ctx, &models.CreatePaymentRequest{
//		PriceAmount:   10,
//		PriceCurrency: "usd",
//		PayCurrency:   "btc",
//	})
//
// Request and response types live in the models package; IPN callbacks are
// verified and parsed by the webhook package.
package nowpay
