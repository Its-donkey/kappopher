package gql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client.clientID != DefaultClientID {
		t.Errorf("expected default client ID %q, got %q", DefaultClientID, client.clientID)
	}

	if client.baseURL != TwitchGQLEndpoint {
		t.Errorf("expected default base URL %q, got %q", TwitchGQLEndpoint, client.baseURL)
	}
}

func TestClientOptions(t *testing.T) {
	customClientID := "custom-client-id"
	customToken := "oauth-token"
	customURL := "https://custom.example.com/gql"

	client := NewClient(
		WithClientID(customClientID),
		WithOAuthToken(customToken),
		WithBaseURL(customURL),
		WithRetry(false, 5),
		WithRetryDelay(2*time.Second),
	)

	if client.clientID != customClientID {
		t.Errorf("expected client ID %q, got %q", customClientID, client.clientID)
	}

	if client.oauthToken != customToken {
		t.Errorf("expected oauth token %q, got %q", customToken, client.oauthToken)
	}

	if client.baseURL != customURL {
		t.Errorf("expected base URL %q, got %q", customURL, client.baseURL)
	}

	if client.retryEnabled {
		t.Error("expected retry to be disabled")
	}

	if client.maxRetries != 5 {
		t.Errorf("expected max retries 5, got %d", client.maxRetries)
	}

	if client.baseRetryDelay != 2*time.Second {
		t.Errorf("expected retry delay 2s, got %v", client.baseRetryDelay)
	}
}

func TestClientExecute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Client-Id") != "test-client" {
			t.Errorf("expected Client-Id header 'test-client', got %q", r.Header.Get("Client-Id"))
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %q", r.Header.Get("Content-Type"))
		}

		// Parse request
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.OperationName != "TestQuery" {
			t.Errorf("expected operation name 'TestQuery', got %q", req.OperationName)
		}

		// Return response
		resp := Response{
			Data: json.RawMessage(`{"user": {"id": "123"}}`),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithClientID("test-client"),
		WithBaseURL(server.URL),
	)

	resp, err := client.Execute(context.Background(), &Request{
		OperationName: "TestQuery",
		Query:         "query TestQuery { user { id } }",
	})

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(resp.Errors) > 0 {
		t.Errorf("unexpected errors: %v", resp.Errors)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}

	user, ok := data["user"].(map[string]interface{})
	if !ok {
		t.Fatal("expected user in response")
	}

	if user["id"] != "123" {
		t.Errorf("expected user id '123', got %v", user["id"])
	}
}

func TestClientExecuteWithOAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "OAuth test-token" {
			t.Errorf("expected Authorization 'OAuth test-token', got %q", authHeader)
		}

		resp := Response{Data: json.RawMessage(`{}`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(
		WithClientID("test-client"),
		WithOAuthToken("test-token"),
		WithBaseURL(server.URL),
	)

	_, err := client.Execute(context.Background(), &Request{Query: "{ __typename }"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestClientExecuteWithHash(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}

		if req.OperationName != "TestOp" {
			t.Errorf("expected operation name 'TestOp', got %q", req.OperationName)
		}

		if req.Extensions == nil || req.Extensions.PersistedQuery == nil {
			t.Error("expected persisted query extension")
		} else {
			if req.Extensions.PersistedQuery.SHA256Hash != "abc123" {
				t.Errorf("expected hash 'abc123', got %q", req.Extensions.PersistedQuery.SHA256Hash)
			}
			if req.Extensions.PersistedQuery.Version != 1 {
				t.Errorf("expected version 1, got %d", req.Extensions.PersistedQuery.Version)
			}
		}

		resp := Response{Data: json.RawMessage(`{}`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	_, err := client.ExecuteWithHash(context.Background(), "TestOp", "abc123", nil)
	if err != nil {
		t.Fatalf("ExecuteWithHash failed: %v", err)
	}
}

func TestClientExecuteBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqs []*Request
		if err := json.NewDecoder(r.Body).Decode(&reqs); err != nil {
			t.Errorf("failed to decode batch request: %v", err)
		}

		if len(reqs) != 2 {
			t.Errorf("expected 2 requests, got %d", len(reqs))
		}

		responses := []*Response{
			{Data: json.RawMessage(`{"first": true}`)},
			{Data: json.RawMessage(`{"second": true}`)},
		}
		json.NewEncoder(w).Encode(responses)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	requests := []*Request{
		{OperationName: "First", Query: "query First { first }"},
		{OperationName: "Second", Query: "query Second { second }"},
	}

	responses, err := client.ExecuteBatch(context.Background(), requests)
	if err != nil {
		t.Fatalf("ExecuteBatch failed: %v", err)
	}

	if len(responses) != 2 {
		t.Errorf("expected 2 responses, got %d", len(responses))
	}
}

func TestClientErrorHandling(t *testing.T) {
	t.Run("API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
		}))
		defer server.Close()

		client := NewClient(WithBaseURL(server.URL), WithRetry(false, 0))

		_, err := client.Execute(context.Background(), &Request{Query: "{}"})
		if err == nil {
			t.Fatal("expected error")
		}

		if !IsAPIError(err) {
			t.Errorf("expected APIError, got %T", err)
		}
	})

	t.Run("Rate limit error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := NewClient(WithBaseURL(server.URL), WithRetry(false, 0))

		_, err := client.Execute(context.Background(), &Request{Query: "{}"})
		if err == nil {
			t.Fatal("expected error")
		}

		if !IsRateLimitError(err) {
			t.Errorf("expected RateLimitError, got %T", err)
		}
	})
}

func TestClientGetters(t *testing.T) {
	client := NewClient(
		WithClientID("my-client"),
		WithBaseURL("https://example.com"),
	)

	if client.GetClientID() != "my-client" {
		t.Errorf("GetClientID returned %q", client.GetClientID())
	}

	if client.GetBaseURL() != "https://example.com" {
		t.Errorf("GetBaseURL returned %q", client.GetBaseURL())
	}
}

func TestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(Response{})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Execute(ctx, &Request{Query: "{}"})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
