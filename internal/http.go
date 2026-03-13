package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const (
	DefaultBaseURL   = "https://api.nowpayments.io/v1"
	HeaderAPIKey     = "x-api-key"
	HeaderAuthBearer = "Authorization"
)

// RetryConfig holds retry policy (0 = no retries).
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// Client is the low-level HTTP client for NOWPayments API.
type Client struct {
	baseURL    string
	apiKey     string
	token      string
	httpClient *http.Client
	retry      RetryConfig
}

// NewClient creates an internal HTTP client. token is optional; retry is used for 5xx and 429.
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
		baseURL:    baseURL,
		apiKey:     apiKey,
		token:      token,
		httpClient: httpClient,
		retry:      retry,
	}
}

// Do sends an HTTP request and decodes the JSON response into result.
// Retries on 429 and 5xx when RetryConfig is set (exponential backoff).
func (c *Client) Do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode body: %w", err)
		}
	}

	url := c.baseURL + path
	maxAttempts := 1
	if c.retry.MaxRetries > 0 {
		maxAttempts += c.retry.MaxRetries
	}
	initialBackoff := c.retry.InitialBackoff
	if initialBackoff <= 0 {
		initialBackoff = 1 * time.Second
	}
	maxBackoff := c.retry.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 30 * time.Second
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var bodyReader io.Reader
		if len(bodyBytes) > 0 {
			bodyReader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(HeaderAPIKey, c.apiKey)
		if c.token != "" {
			req.Header.Set(HeaderAuthBearer, "Bearer "+c.token)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		if resp.StatusCode < 400 {
			if result != nil && len(respBody) > 0 {
				if err := json.Unmarshal(respBody, result); err != nil {
					return fmt.Errorf("decode response: %w", err)
				}
			}
			return nil
		}

		lastErr = mapStatusToError(resp.StatusCode, respBody)
		retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
		if !retryable || attempt >= maxAttempts-1 {
			return lastErr
		}
		backoff := time.Duration(float64(initialBackoff) * math.Pow(2, float64(attempt)))
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		select {
		case <-time.After(backoff):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return lastErr
}

// mapStatusToError returns an error based on status code and optional body.
func mapStatusToError(status int, body []byte) error {
	msg := string(body)
	if msg == "" {
		msg = http.StatusText(status)
	}
	err := &APIError{StatusCode: status, Message: msg}
	switch status {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %v", ErrInvalidAPIKey, err)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %v", ErrForbidden, err)
	case http.StatusNotFound:
		return fmt.Errorf("%w: %v", ErrNotFound, err)
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %v", ErrBadRequest, err)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w: %v", ErrValidation, err)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %v", ErrRateLimited, err)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return fmt.Errorf("%w: %v", ErrServerError, err)
	default:
		return err
	}
}
