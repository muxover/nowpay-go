# Recorded responses

Real responses from the NOWPayments production API, captured by `live_test.go` on 2026-09-28 and 2026-10-08. Payment addresses, client IPs, account IDs and balances are replaced with placeholders.

Several differ from the examples in the official docs (`../*.json`): `/currencies?fixed_rate=true` returns objects, `/merchant/coins` uses `selectedCurrencies`, `/full-currencies` sends `network_precision` as a string, `/estimate` sends `estimated_amount` as a string, and `POST /subscriptions` wraps the new subscription in an array. `GET /payment/` fails with a 500 unless `sortBy` and `orderBy` are sent, the fiat-payout currency endpoints require a `provider`, `/sub-partner/payments` uses `status`, text prices, and string amounts instead of the usual payment fields, and `/sub-partner/balance/:id` sends an empty balance as `[]` instead of an object.
