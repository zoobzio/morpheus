// Package postmark provides a client for the Postmark transactional email API.
// It wraps all outbound calls in a resilience pipeline (timeout, backoff, circuit breaker).
package postmark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zoobzio/pipz"
)

// Resilience configuration.
const (
	apiBaseURL          = "https://api.postmarkapp.com"
	apiTimeout          = 30 * time.Second
	apiMaxAttempts      = 3
	apiBackoffDelay     = 500 * time.Millisecond
	apiFailureThreshold = 5
	apiResetTimeout     = 30 * time.Second
)

// Pipeline identities.
var (
	sendProcessorID = pipz.NewIdentity("postmark.send.call", "Postmark send email API call")
	sendTimeoutID   = pipz.NewIdentity("postmark.send.timeout", "Timeout for Postmark send email")
	sendBackoffID   = pipz.NewIdentity("postmark.send.backoff", "Backoff retry for Postmark send email")
	sendBreakerID   = pipz.NewIdentity("postmark.send.breaker", "Circuit breaker for Postmark send email")
)

// sendCall carries a send email request and its response through the pipeline.
type sendCall struct {
	serverToken string
	request     EmailRequest
	response    *EmailResponse
}

// Clone returns a deep copy of the call. Required by pipz.
func (c *sendCall) Clone() *sendCall {
	clone := *c
	if c.response != nil {
		r := *c.response
		clone.response = &r
	}
	return &clone
}

// Client sends transactional email via the Postmark API.
type Client struct {
	serverToken string
	defaultFrom string
	httpClient  *http.Client
	pipeline    pipz.Chainable[*sendCall]
}

// NewClient creates a new Postmark client with a resilience pipeline.
func NewClient(serverToken, defaultFrom string) *Client {
	c := &Client{
		serverToken: serverToken,
		defaultFrom: defaultFrom,
		httpClient:  &http.Client{},
	}
	c.pipeline = c.buildPipeline()
	return c
}

// buildPipeline constructs the resilient processing pipeline for send operations.
func (c *Client) buildPipeline() pipz.Chainable[*sendCall] {
	processor := pipz.Apply(sendProcessorID, func(ctx context.Context, call *sendCall) (*sendCall, error) {
		body, err := json.Marshal(call.request)
		if err != nil {
			return nil, fmt.Errorf("postmark: marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBaseURL+"/email", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("postmark: create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("X-Postmark-Server-Token", call.serverToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("postmark: send request: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("postmark: read response: %w", err)
		}

		var emailResp EmailResponse
		if err := json.Unmarshal(respBody, &emailResp); err != nil {
			return nil, fmt.Errorf("postmark: unmarshal response: %w", err)
		}

		if resp.StatusCode >= 500 {
			return nil, fmt.Errorf("postmark: server error %d: %s", resp.StatusCode, emailResp.Message)
		}

		call.response = &emailResp
		return call, nil
	})

	return pipz.NewCircuitBreaker(sendBreakerID,
		pipz.NewBackoff(sendBackoffID,
			pipz.NewTimeout(sendTimeoutID, processor, apiTimeout),
			apiMaxAttempts, apiBackoffDelay,
		),
		apiFailureThreshold, apiResetTimeout,
	)
}

// SendEmail sends a single transactional email via Postmark.
// If req.From is empty, the client's DefaultFrom address is used.
func (c *Client) SendEmail(ctx context.Context, req EmailRequest) (*EmailResponse, error) {
	if req.From == "" {
		req.From = c.defaultFrom
	}

	call := &sendCall{
		serverToken: c.serverToken,
		request:     req,
	}

	result, err := c.pipeline.Process(ctx, call)
	if err != nil {
		return nil, err
	}

	return result.response, nil
}

// Close shuts down the pipeline and releases resources.
func (c *Client) Close() error {
	if c.pipeline != nil {
		return c.pipeline.Close()
	}
	return nil
}
