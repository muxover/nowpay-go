package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://api.nowpayments.io/v1"
	HeaderAPIKey   = "x-api-key"
)

// Client is the low-level HTTP client for NOWPayments API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates an internal HTTP client.
func NewClient(baseURL, apiKey string, timeout time.Duration, httpClient *http.Client) *Client {
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
		httpClient: httpClient,
	}
}

// Do sends an HTTP request and decodes the JSON response into result.
// If result is nil, the response body is read and discarded.
func (c *Client) Do(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(HeaderAPIKey, c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return mapStatusToError(resp.StatusCode, respBody)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
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
