package gql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDiscovery(t *testing.T) {
	client := NewClient()
	discovery := NewDiscovery(client)

	if discovery.client != client {
		t.Error("client not set correctly")
	}

	if discovery.introspector == nil {
		t.Error("introspector not initialized")
	}

	if discovery.registry == nil {
		t.Error("registry not initialized")
	}
}

func TestGetKnownOperations(t *testing.T) {
	client := NewClient()
	discovery := NewDiscovery(client)

	ops := discovery.GetKnownOperations()
	if len(ops) == 0 {
		t.Error("expected known operations")
	}

	// Check for some expected operations
	foundChannel := false
	foundFollow := false
	for _, op := range ops {
		if op.Name == "ChannelPage_Query" {
			foundChannel = true
		}
		if op.Name == "FollowButton_FollowUser" {
			foundFollow = true
			if op.Type != OperationMutation {
				t.Errorf("expected FollowButton_FollowUser to be mutation, got %v", op.Type)
			}
		}
	}

	if !foundChannel {
		t.Error("expected ChannelPage_Query in known operations")
	}
	if !foundFollow {
		t.Error("expected FollowButton_FollowUser in known operations")
	}
}

func TestDiscoveryIntrospect(t *testing.T) {
	schemaResponse := map[string]interface{}{
		"data": map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType":    map[string]interface{}{"name": "Query"},
				"mutationType": map[string]interface{}{"name": "Mutation"},
				"types": []interface{}{
					map[string]interface{}{
						"kind": "OBJECT",
						"name": "Query",
						"fields": []interface{}{
							map[string]interface{}{
								"name":        "user",
								"description": "Get a user by login",
								"args":        []interface{}{},
								"type":        map[string]interface{}{"kind": "OBJECT", "name": "User"},
							},
						},
					},
					map[string]interface{}{
						"kind": "OBJECT",
						"name": "User",
						"fields": []interface{}{
							map[string]interface{}{
								"name": "id",
								"type": map[string]interface{}{"kind": "SCALAR", "name": "ID"},
							},
						},
					},
				},
				"directives": []interface{}{},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(schemaResponse)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	discovery := NewDiscovery(client)

	schema, err := discovery.Introspect(context.Background())
	if err != nil {
		t.Fatalf("Introspect failed: %v", err)
	}

	if schema.QueryType == nil {
		t.Error("expected QueryType")
	}

	if schema.QueryType.Name != "Query" {
		t.Errorf("expected QueryType name 'Query', got %q", schema.QueryType.Name)
	}

	if len(schema.Types) != 2 {
		t.Errorf("expected 2 types, got %d", len(schema.Types))
	}
}

func TestDiscoveryIntrospectDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Errors: []Error{
				{Message: "GraphQL introspection is not allowed"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	discovery := NewDiscovery(client)

	_, err := discovery.Introspect(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	if !IsIntrospectionDisabled(err) {
		t.Errorf("expected IntrospectionDisabledError, got %T", err)
	}
}

func TestDiscoveryProbeOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			Errors: []Error{
				{Message: "Cannot query field '__invalid_probe_field_12345' on type 'Query'. Did you mean 'user', 'channel', or 'game'?"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	discovery := NewDiscovery(client)

	ops, err := discovery.ProbeOperations(context.Background())
	if err != nil {
		t.Fatalf("ProbeOperations failed: %v", err)
	}

	// Should discover some operations from error messages
	// The exact count depends on the probe patterns
	if ops == nil {
		t.Error("expected operations slice")
	}
}

func TestDiscoverAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return introspection disabled error
		resp := Response{
			Errors: []Error{
				{Message: "Introspection is disabled"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	discovery := NewDiscovery(client)

	result, err := discovery.DiscoverAll(context.Background())
	if err != nil {
		t.Fatalf("DiscoverAll failed: %v", err)
	}

	// Should have known operations even if introspection fails
	if len(result.Operations) == 0 {
		t.Error("expected operations from known list")
	}

	// Should include SourceKnownList
	hasKnownSource := false
	for _, s := range result.Sources {
		if s == SourceKnownList {
			hasKnownSource = true
			break
		}
	}
	if !hasKnownSource {
		t.Error("expected SourceKnownList in sources")
	}

	// Should have discovery time set
	if result.DiscoveredAt.IsZero() {
		t.Error("DiscoveredAt should be set")
	}
}

func TestDiscoverWithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{Data: json.RawMessage(`{}`)}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	discovery := NewDiscovery(client)

	opts := DiscoveryOptions{
		SkipIntrospection: true,
		SkipProbing:       true,
		SkipKnownList:     false,
		Timeout:           5 * time.Second,
	}

	result, err := discovery.DiscoverWithOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("DiscoverWithOptions failed: %v", err)
	}

	// Should only have known list source
	if len(result.Sources) != 1 || result.Sources[0] != SourceKnownList {
		t.Errorf("expected only SourceKnownList, got %v", result.Sources)
	}
}

func TestInferOperationType(t *testing.T) {
	tests := []struct {
		name     string
		expected OperationType
	}{
		{"CreateClip", OperationMutation},
		{"UpdateUser", OperationMutation},
		{"DeleteVideo", OperationMutation},
		{"FollowButton_FollowUser", OperationMutation},
		{"UnfollowUser", OperationMutation},
		{"BlockUser", OperationMutation},
		{"GetUser", OperationQuery},
		{"ChannelPage_Query", OperationQuery},
		{"StreamMetadata", OperationQuery},
		{"OnMessageReceived", OperationSubscription},
		{"SubscribeToChat", OperationSubscription},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferOperationType(tt.name)
			if result != tt.expected {
				t.Errorf("inferOperationType(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestLooksLikeOperation(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"ChannelPage_Query", true},
		{"GetUser", true},
		{"CreateClip", true},
		{"UserCard", true},
		{"id", false},
		{"a", false},
		{"login", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := looksLikeOperation(tt.name)
			if result != tt.expected {
				t.Errorf("looksLikeOperation(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestDeduplicateOperations(t *testing.T) {
	ops := []Operation{
		{Name: "Op1", Type: OperationQuery},
		{Name: "Op2", Type: OperationQuery, Query: "query Op2 { user }"},
		{Name: "Op1", Type: OperationQuery, Description: "Better description"},
		{Name: "Op2", Type: OperationQuery}, // Less info, should not replace
	}

	result := deduplicateOperations(ops)

	if len(result) != 2 {
		t.Errorf("expected 2 operations, got %d", len(result))
	}

	// Find Op1 - should have description
	for _, op := range result {
		if op.Name == "Op1" && op.Description != "Better description" {
			t.Error("Op1 should have description from second occurrence")
		}
		if op.Name == "Op2" && op.Query == "" {
			t.Error("Op2 should have query from first occurrence")
		}
	}
}

func TestDiscoveryResultGetOperationsByType(t *testing.T) {
	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "Query1", Type: OperationQuery},
			{Name: "Query2", Type: OperationQuery},
			{Name: "Mutation1", Type: OperationMutation},
			{Name: "Sub1", Type: OperationSubscription},
		},
	}

	queries := result.GetOperationsByType(OperationQuery)
	if len(queries) != 2 {
		t.Errorf("expected 2 queries, got %d", len(queries))
	}

	mutations := result.GetOperationsByType(OperationMutation)
	if len(mutations) != 1 {
		t.Errorf("expected 1 mutation, got %d", len(mutations))
	}
}

func TestDiscoveryResultGetOperationByName(t *testing.T) {
	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "Query1", Type: OperationQuery},
			{Name: "Query2", Type: OperationQuery},
		},
	}

	op := result.GetOperationByName("Query1")
	if op == nil {
		t.Fatal("expected to find Query1")
	}
	if op.Name != "Query1" {
		t.Errorf("expected Query1, got %s", op.Name)
	}

	op = result.GetOperationByName("NonExistent")
	if op != nil {
		t.Error("expected nil for non-existent operation")
	}
}
