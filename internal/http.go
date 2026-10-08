package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultBaseURL   = "https://api.nowpayments.io/v1"
	HeaderAPIKey     = "x-api-key"
	HeaderAuthBearer = "Authorization"
	userAgent        = "nowpay-go"
)

// RetryConfig holds retry policy (0 = no retries).
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// Client is the low-level HTTP client for the NOWPayments API.
type Client struct {
	baseURL    string
	apiKey     string
	token      string
	httpClient *http.Client
	retry      RetryConfig
}

// NewClient creates an internal HTTP client. token is optional.
func NewClient(baseURL, apiKey, token string, timeout time.Duration, httpClient *http.Client, retry RetryConfig) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		token:      token,
		httpClient: httpClient,
		retry:      retry,
	}
}

// Do sends a request and decodes the JSON response into result, if non-nil.
// A 5xx is retried only for idempotent methods: a POST may already have
// created a payment or payout.
func (c *Client) Do(ctx context.Context, method, path string, body, result any) error {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("nowpay: encode request: %w", err)
		}
	}

	attempts := 1 + max(c.retry.MaxRetries, 0)
	for attempt := 1; ; attempt++ {
		status, header, respBody, err := c.send(ctx, method, path, payload)
		if err != nil {
			return err
		}
		if status < 400 {
			if result != nil && len(bytes.TrimSpace(respBody)) > 0 {
				if err := json.Unmarshal(respBody, result); err != nil {
					return fmt.Errorf("nowpay: decode %s %s response: %w", method, path, err)
				}
			}
			return nil
		}

		apiErr := newAPIError(status, respBody)
		if attempt >= attempts || !retryable(method, status) {
			return apiErr
		}
		if err := sleep(ctx, c.backoff(attempt, header)); err != nil {
			return err
		}
	}
}

func (c *Client) send(ctx context.Context, method, path string, payload []byte) (int, http.Header, []byte, error) {
	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set(HeaderAPIKey, c.apiKey)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set(HeaderAuthBearer, "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, err
	}
	return resp.StatusCode, resp.Header, respBody, nil
}

func retryable(method string, status int) bool {
	if status == http.StatusTooManyRequests {
		return true
	}
	if status < 500 {
		return false
	}
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	}
	return false
}

// backoff honors Retry-After, otherwise doubles from InitialBackoff up to MaxBackoff.
func (c *Client) backoff(attempt int, header http.Header) time.Duration {
	initial := c.retry.InitialBackoff
	if initial <= 0 {
		initial = time.Second
	}
	limit := c.retry.MaxBackoff
	if limit <= 0 {
		limit = 30 * time.Second
	}

	if secs, err := strconv.Atoi(header.Get("Retry-After")); err == nil && secs >= 0 {
		return min(time.Duration(secs)*time.Second, limit)
	}
	d := initial
	for i := 1; i < attempt && d < limit; i++ {
		d *= 2
	}
	return min(d, limit)
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
