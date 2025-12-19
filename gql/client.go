package gql

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Client is a GraphQL client for Twitch's GQL API.
type Client struct {
	clientID   string
	oauthToken string
	httpClient *http.Client
	baseURL    string

	// Rate limiting
	rateLimitRemaining int
	rateLimitReset     time.Time
	rateMu             sync.Mutex

	// Retry configuration
	maxRetries     int
	retryEnabled   bool
	baseRetryDelay time.Duration
}

// Option configures a Client.
type Option func(*Client)

// NewClient creates a new GQL client with the given options.
func NewClient(opts ...Option) *Client {
	c := &Client{
		clientID:       DefaultClientID,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
		baseURL:        TwitchGQLEndpoint,
		maxRetries:     3,
		retryEnabled:   true,
		baseRetryDelay: time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithClientID sets the Client-ID header value.
func WithClientID(clientID string) Option {
	return func(c *Client) {
		c.clientID = clientID
	}
}

// WithOAuthToken sets the OAuth token for authenticated requests.
func WithOAuthToken(token string) Option {
	return func(c *Client) {
		c.oauthToken = token
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets the base URL for the GQL endpoint.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithRetry configures retry behavior.
func WithRetry(enabled bool, maxRetries int) Option {
	return func(c *Client) {
		c.retryEnabled = enabled
		c.maxRetries = maxRetries
	}
}

// WithRetryDelay sets the base delay between retries.
func WithRetryDelay(delay time.Duration) Option {
	return func(c *Client) {
		c.baseRetryDelay = delay
	}
}

// Execute sends a GraphQL request and returns the response.
func (c *Client) Execute(ctx context.Context, req *Request) (*Response, error) {
	return c.doRequest(ctx, req)
}

// ExecutePersisted sends a persisted query request.
func (c *Client) ExecutePersisted(ctx context.Context, req *PersistedRequest) (*Response, error) {
	gqlReq := &Request{
		OperationName: req.OperationName,
		Variables:     req.Variables,
		Extensions:    req.Extensions,
	}
	if gqlReq.Extensions == nil {
		gqlReq.Extensions = &Extensions{}
	}
	return c.doRequest(ctx, gqlReq)
}

// ExecuteWithHash sends a request using a persisted query hash.
func (c *Client) ExecuteWithHash(ctx context.Context, operationName, sha256Hash string, variables map[string]interface{}) (*Response, error) {
	req := &Request{
		OperationName: operationName,
		Variables:     variables,
		Extensions: &Extensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: sha256Hash,
			},
		},
	}
	return c.doRequest(ctx, req)
}

// ExecuteBatch sends multiple requests in a single HTTP call.
func (c *Client) ExecuteBatch(ctx context.Context, requests []*Request) ([]*Response, error) {
	body, err := json.Marshal(requests)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var responses []*Response
	if err := json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return responses, nil
}

func (c *Client) doRequest(ctx context.Context, req *Request) (*Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 && c.retryEnabled {
			delay := c.baseRetryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		resp, err := c.doRequestOnce(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !c.retryEnabled || !isRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) doRequestOnce(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	c.updateRateLimits(resp)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{
			ResetAt:   c.rateLimitReset,
			Remaining: c.rateLimitRemaining,
		}
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(bodyBytes),
		}
	}

	var gqlResp Response
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &gqlResp, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Client-Id", c.clientID)
	if c.oauthToken != "" {
		req.Header.Set("Authorization", "OAuth "+c.oauthToken)
	}
}

func (c *Client) updateRateLimits(resp *http.Response) {
	c.rateMu.Lock()
	defer c.rateMu.Unlock()
	// Twitch GQL doesn't expose rate limit headers like Helix,
	// but we track them in case they add them in the future
}

func isRetryableError(err error) bool {
	if _, ok := err.(*RateLimitError); ok {
		return true
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode >= 500
	}
	return false
}

// APIError represents an HTTP error from the API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("gql api error %d: %s", e.StatusCode, e.Message)
}

// RateLimitError indicates the rate limit has been exceeded.
type RateLimitError struct {
	ResetAt   time.Time
	Remaining int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, resets at %v", e.ResetAt)
}

// IsRateLimitError checks if an error is a rate limit error.
func IsRateLimitError(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// IsAPIError checks if an error is an API error.
func IsAPIError(err error) bool {
	_, ok := err.(*APIError)
	return ok
}

// GetClientID returns the client ID being used.
func (c *Client) GetClientID() string {
	return c.clientID
}

// GetBaseURL returns the base URL being used.
func (c *Client) GetBaseURL() string {
	return c.baseURL
}
