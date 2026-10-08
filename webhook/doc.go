// Package webhook verifies and parses NOWPayments IPN callbacks.
//
// Always call VerifySignature on the raw request body before trusting
// anything in it, then ParseEvent to read it.
package webhook
