package telegram

// InlineKeyboardButton represents a single inline keyboard button (e.g. for payment link).
// Compatible with Telegram Bot API: { "text": "...", "url": "..." } or callback_data.
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	CallbackData string `json:"callback_data,omitempty"`
}

// InlineKeyboardMarkup is a slice of rows of buttons for Telegram.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// PaymentButton returns a single URL button for use in an inline keyboard.
// label is the button text; invoiceURL is the payment or invoice URL.
func PaymentButton(label, invoiceURL string) InlineKeyboardButton {
	return InlineKeyboardButton{Text: label, URL: invoiceURL}
}

// PaymentKeyboard returns a full markup with one row containing the payment button.
func PaymentKeyboard(label, invoiceURL string) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{PaymentButton(label, invoiceURL)},
		},
	}
}
