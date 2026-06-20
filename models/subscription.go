package models

// SubscriptionPlan is a recurring-payment plan (/v1/subscriptions/plans).
type SubscriptionPlan struct {
	ID               FlexInt `json:"id"`
	Title            string  `json:"title"`
	IntervalDay      FlexInt `json:"interval_day"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	IPNCallbackURL   *string `json:"ipn_callback_url,omitempty"`
	SuccessURL       *string `json:"success_url,omitempty"`
	CancelURL        *string `json:"cancel_url,omitempty"`
	PartiallyPaidURL *string `json:"partially_paid_url,omitempty"`
	CreatedAt        string  `json:"created_at,omitempty"`
	UpdatedAt        string  `json:"updated_at,omitempty"`
}

// CreatePlanRequest is the body for POST /v1/subscriptions/plans.
type CreatePlanRequest struct {
	Title            string  `json:"title"`
	IntervalDay      int     `json:"interval_day"`
	Amount           float64 `json:"amount"`
	Currency         string  `json:"currency"`
	IPNCallbackURL   string  `json:"ipn_callback_url,omitempty"`
	SuccessURL       string  `json:"success_url,omitempty"`
	CancelURL        string  `json:"cancel_url,omitempty"`
	PartiallyPaidURL string  `json:"partially_paid_url,omitempty"`
}

// UpdatePlanRequest is the body for PATCH /v1/subscriptions/plans/:id.
type UpdatePlanRequest struct {
	Title            string   `json:"title,omitempty"`
	IntervalDay      int      `json:"interval_day,omitempty"`
	Amount           *float64 `json:"amount,omitempty"`
	Currency         string   `json:"currency,omitempty"`
	IPNCallbackURL   string   `json:"ipn_callback_url,omitempty"`
	SuccessURL       string   `json:"success_url,omitempty"`
	CancelURL        string   `json:"cancel_url,omitempty"`
	PartiallyPaidURL string   `json:"partially_paid_url,omitempty"`
}

// ListPlansParams holds query params for GET /v1/subscriptions/plans.
type ListPlansParams struct {
	Limit  int
	Offset int
}

// Subscription is an email/customer recurring payment (/v1/subscriptions).
type Subscription struct {
	ID                 FlexInt                `json:"id"`
	SubscriptionPlanID FlexInt                `json:"subscription_plan_id"`
	IsActive           bool                   `json:"is_active"`
	Status             string                 `json:"status,omitempty"`
	ExpireDate         string                 `json:"expire_date,omitempty"`
	Subscriber         SubscriptionSubscriber `json:"subscriber"`
	CreatedAt          string                 `json:"created_at,omitempty"`
	UpdatedAt          string                 `json:"updated_at,omitempty"`
}

// SubscriptionSubscriber identifies a subscription's subscriber (email or custody sub-account).
type SubscriptionSubscriber struct {
	Email        string  `json:"email,omitempty"`
	SubPartnerID FlexInt `json:"sub_partner_id,omitempty"`
}

// CreateSubscriptionRequest is the body for POST /v1/subscriptions.
// Provide Email for an email subscription, or SubPartnerID for a custody subscriber.
type CreateSubscriptionRequest struct {
	SubscriptionPlanID int64  `json:"subscription_plan_id"`
	Email              string `json:"email,omitempty"`
	SubPartnerID       int64  `json:"sub_partner_id,omitempty"`
}

// ListSubscriptionsParams holds query params for GET /v1/subscriptions.
type ListSubscriptionsParams struct {
	Status             string
	SubscriptionPlanID int64
	IsActive           *bool
	Limit              int
	Offset             int
}
