package telegram

import (
	"fmt"

	"github.com/muxover/nowpay-go/models"
)

// InvoiceMessage formats a payment or invoice for display in a Telegram message.
// It includes amount, currency, address (if present), and a short status.
func InvoiceMessage(payment *models.Payment) string {
	if payment == nil {
		return "No payment details."
	}
	msg := fmt.Sprintf("Amount: %.8f %s\nPrice: %.2f %s\nStatus: %s",
		payment.PayAmount, payment.PayCurrency,
		payment.PriceAmount, payment.PriceCurrency,
		payment.PaymentStatus)
	if payment.PayAddress != "" {
		msg += fmt.Sprintf("\nAddress: `%s`", payment.PayAddress)
	}
	if payment.OrderID != "" {
		msg += fmt.Sprintf("\nOrder: %s", payment.OrderID)
	}
	return msg
}

// InvoiceMessageFromInvoice formats an invoice for display in a Telegram message.
func InvoiceMessageFromInvoice(inv *models.Invoice) string {
	if inv == nil {
		return "No invoice details."
	}
	msg := fmt.Sprintf("Amount: %.2f %s\nStatus: %s",
		inv.PriceAmount, inv.PriceCurrency, inv.PaymentStatus)
	if inv.InvoiceURL != "" {
		msg += fmt.Sprintf("\nPay: %s", inv.InvoiceURL)
	}
	if inv.OrderID != "" {
		msg += fmt.Sprintf("\nOrder: %s", inv.OrderID)
	}
	return msg
}
