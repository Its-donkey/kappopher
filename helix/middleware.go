package helix

import (
	"context"
	"net/http"
	"time"
)

// Middleware is a function that wraps request execution.
// It receives the request context, the request, and a next function to call.
// Middleware can modify the request before calling next, and modify/inspect
// the response after next returns.
type Middleware func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error)

// MiddlewareNext is a function that continues the middleware chain.
type MiddlewareNext func(ctx context.Context, req *Request) (*MiddlewareResponse, error)

// MiddlewareResponse contains the response data available to middleware.
type MiddlewareResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Use adds middleware to the client. Middleware is executed in the order added.
func (c *Client) Use(mw ...Middleware) {
	c.middleware = append(c.middleware, mw...)
}

// LoggingMiddleware creates middleware that logs requests and responses.
func LoggingMiddleware(logger func(format string, args ...interface{})) Middleware {
	return func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		logger("-> %s %s", req.Method, req.Endpoint)

		resp, err := next(ctx, req)

		if err != nil {
			logger("<- %s %s: error: %v", req.Method, req.Endpoint, err)
		} else {
			logger("<- %s %s: %d", req.Method, req.Endpoint, resp.StatusCode)
		}

		return resp, err
	}
}

// RetryMiddleware creates middleware that retries failed requests.
// This is separate from the built-in rate limit retry and handles other transient errors.
func RetryMiddleware(maxRetries int, retryableStatuses ...int) Middleware {
	statusSet := make(map[int]bool)
	for _, s := range retryableStatuses {
		statusSet[s] = true
	}
	// Default retryable statuses if none provided
	if len(statusSet) == 0 {
		statusSet[502] = true // Bad Gateway
		statusSet[503] = true // Service Unavailable
		statusSet[504] = true // Gateway Timeout
	}

	return func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		var lastResp *MiddlewareResponse
		var lastErr error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			resp, err := next(ctx, req)
			if err != nil {
				lastErr = err
				continue
			}

			if !statusSet[resp.StatusCode] {
				return resp, nil
			}

			lastResp = resp
			lastErr = nil
		}

		if lastErr != nil {
			return nil, lastErr
		}
		return lastResp, nil
	}
}

// HeaderMiddleware creates middleware that adds custom headers to requests.
func HeaderMiddleware(headers map[string]string) Middleware {
	return func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		// Headers are added in the actual HTTP request building,
		// so we store them in context for the client to pick up
		ctx = contextWithHeaders(ctx, headers)
		return next(ctx, req)
	}
}

// headersContextKey is the context key for custom headers.
type headersContextKey struct{}

// contextWithHeaders adds custom headers to context.
func contextWithHeaders(ctx context.Context, headers map[string]string) context.Context {
	existing := headersFromContext(ctx)
	merged := make(map[string]string)
	for k, v := range existing {
		merged[k] = v
	}
	for k, v := range headers {
		merged[k] = v
	}
	return context.WithValue(ctx, headersContextKey{}, merged)
}

// headersFromContext retrieves custom headers from context.
func headersFromContext(ctx context.Context) map[string]string {
	if headers, ok := ctx.Value(headersContextKey{}).(map[string]string); ok {
		return headers
	}
	return nil
}

// MetricsMiddleware creates middleware that tracks request metrics.
type RequestMetrics struct {
	Method     string
	Endpoint   string
	StatusCode int
	Duration   int64 // milliseconds
	Error      error
}

func MetricsMiddleware(collector func(metrics RequestMetrics)) Middleware {
	return func(ctx context.Context, req *Request, next MiddlewareNext) (*MiddlewareResponse, error) {
		start := time.Now()

		resp, err := next(ctx, req)

		metrics := RequestMetrics{
			Method:   req.Method,
			Endpoint: req.Endpoint,
			Duration: time.Since(start).Milliseconds(),
			Error:    err,
		}
		if resp != nil {
			metrics.StatusCode = resp.StatusCode
		}

		collector(metrics)

		return resp, err
	}
}
