package gql

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Discovery handles discovering GraphQL operations using multiple methods.
type Discovery struct {
	client       *Client
	introspector *Introspector
	registry     *OperationsRegistry
}

// NewDiscovery creates a new discovery instance.
func NewDiscovery(client *Client) *Discovery {
	return &Discovery{
		client:       client,
		introspector: NewIntrospector(client),
		registry:     NewOperationsRegistry(),
	}
}

// DiscoverAll uses all discovery methods and returns combined results.
func (d *Discovery) DiscoverAll(ctx context.Context) (*DiscoveryResult, error) {
	result := &DiscoveryResult{
		Operations:   make([]Operation, 0),
		DiscoveredAt: time.Now(),
		Sources:      make([]DiscoverySource, 0),
		Errors:       make([]string, 0),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// 1. Try introspection
	wg.Add(1)
	go func() {
		defer wg.Done()
		schema, err := d.introspector.IntrospectSchema(ctx)
		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			if IsIntrospectionDisabled(err) {
				result.IntrospectionEnabled = false
				result.Errors = append(result.Errors, "introspection disabled: "+err.Error())
			} else {
				result.Errors = append(result.Errors, "introspection failed: "+err.Error())
			}
		} else {
			result.Schema = schema
			result.IntrospectionEnabled = true
			result.Sources = append(result.Sources, SourceIntrospection)

			ops := d.introspector.GetAllOperations(schema)
			result.Operations = append(result.Operations, ops...)
		}
	}()

	// 2. Error probing
	wg.Add(1)
	go func() {
		defer wg.Done()
		probed, err := d.ProbeOperations(ctx)
		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			result.Errors = append(result.Errors, "error probing failed: "+err.Error())
		} else if len(probed) > 0 {
			result.Sources = append(result.Sources, SourceErrorProbing)
			result.Operations = append(result.Operations, probed...)
		}
	}()

	// 3. Known operations
	wg.Add(1)
	go func() {
		defer wg.Done()
		known := d.GetKnownOperations()
		mu.Lock()
		defer mu.Unlock()

		if len(known) > 0 {
			result.Sources = append(result.Sources, SourceKnownList)
			result.Operations = append(result.Operations, known...)
		}
	}()

	wg.Wait()

	// Deduplicate operations
	result.Operations = deduplicateOperations(result.Operations)

	return result, nil
}

// Introspect performs schema introspection only.
func (d *Discovery) Introspect(ctx context.Context) (*Schema, error) {
	return d.introspector.IntrospectSchema(ctx)
}

// ProbeOperations discovers operations by analyzing error responses.
func (d *Discovery) ProbeOperations(ctx context.Context) ([]Operation, error) {
	var discovered []Operation
	now := time.Now()

	// Probe patterns - these are designed to trigger informative error messages
	probes := []string{
		// Invalid operation names that might reveal valid ones
		`query { __invalid_probe_field_12345 }`,
		`mutation { __invalid_probe_mutation_12345 }`,
		// Empty queries to check behavior
		`query TestProbe { }`,
		// Field existence probes on common root types
		`query { user { __probe } }`,
		`query { channel { __probe } }`,
		`query { game { __probe } }`,
		`query { stream { __probe } }`,
		`query { video { __probe } }`,
		`query { clip { __probe } }`,
	}

	for _, probe := range probes {
		select {
		case <-ctx.Done():
			return discovered, ctx.Err()
		default:
		}

		resp, err := d.client.Execute(ctx, &Request{Query: probe})
		if err != nil {
			continue
		}

		// Extract operation names from error messages
		for _, gqlErr := range resp.Errors {
			ops := extractOperationNamesFromError(gqlErr.Message)
			for _, opName := range ops {
				discovered = append(discovered, Operation{
					Name:         opName,
					Type:         inferOperationType(opName),
					Source:       SourceErrorProbing,
					DiscoveredAt: now,
				})
			}

			// Also extract field names that might be operation hints
			fields := extractFieldNamesFromError(gqlErr.Message)
			for _, field := range fields {
				// Only add if it looks like an operation name
				if looksLikeOperation(field) {
					discovered = append(discovered, Operation{
						Name:         field,
						Type:         inferOperationType(field),
						Source:       SourceErrorProbing,
						DiscoveredAt: now,
					})
				}
			}
		}
	}

	return deduplicateOperations(discovered), nil
}

// ProbeWithPatterns sends specific probe queries to discover operations.
func (d *Discovery) ProbeWithPatterns(ctx context.Context, patterns []string) ([]Operation, error) {
	var discovered []Operation
	now := time.Now()

	for _, pattern := range patterns {
		select {
		case <-ctx.Done():
			return discovered, ctx.Err()
		default:
		}

		resp, err := d.client.Execute(ctx, &Request{Query: pattern})
		if err != nil {
			continue
		}

		for _, gqlErr := range resp.Errors {
			ops := extractOperationNamesFromError(gqlErr.Message)
			for _, opName := range ops {
				discovered = append(discovered, Operation{
					Name:         opName,
					Type:         inferOperationType(opName),
					Source:       SourceErrorProbing,
					DiscoveredAt: now,
				})
			}
		}
	}

	return deduplicateOperations(discovered), nil
}

// GetKnownOperations returns operations from the known list.
func (d *Discovery) GetKnownOperations() []Operation {
	return d.registry.List()
}

// GetRegistry returns the operations registry.
func (d *Discovery) GetRegistry() *OperationsRegistry {
	return d.registry
}

// ValidateOperation checks if an operation exists by executing it.
func (d *Discovery) ValidateOperation(ctx context.Context, op Operation) (bool, error) {
	// Try to execute the operation with minimal variables
	var req *Request

	if op.SHA256Hash != "" {
		// Use persisted query
		resp, err := d.client.ExecuteWithHash(ctx, op.Name, op.SHA256Hash, nil)
		if err != nil {
			return false, err
		}
		// Operation exists if we don't get a "persisted query not found" error
		for _, e := range resp.Errors {
			if strings.Contains(e.Message, "PersistedQueryNotFound") {
				return false, nil
			}
		}
		return true, nil
	}

	if op.Query != "" {
		req = &Request{
			OperationName: op.Name,
			Query:         op.Query,
		}
	} else {
		// Build a minimal query
		var queryType string
		switch op.Type {
		case OperationMutation:
			queryType = "mutation"
		case OperationSubscription:
			queryType = "subscription"
		default:
			queryType = "query"
		}
		req = &Request{
			Query: queryType + " " + op.Name + " { __typename }",
		}
	}

	resp, err := d.client.Execute(ctx, req)
	if err != nil {
		return false, err
	}

	// Check if the operation was recognized
	for _, e := range resp.Errors {
		msg := strings.ToLower(e.Message)
		if strings.Contains(msg, "unknown operation") ||
			strings.Contains(msg, "not found") ||
			strings.Contains(msg, "does not exist") {
			return false, nil
		}
	}

	return true, nil
}

// extractOperationNamesFromError parses error messages for operation names.
func extractOperationNamesFromError(msg string) []string {
	var names []string

	// Pattern: "Did you mean X, Y, or Z?"
	didYouMean := regexp.MustCompile(`Did you mean ["']?(\w+)["']?`)
	matches := didYouMean.FindAllStringSubmatch(msg, -1)
	for _, m := range matches {
		if len(m) > 1 {
			names = append(names, m[1])
		}
	}

	// Pattern: "Unknown operation 'X'"
	unknownOp := regexp.MustCompile(`Unknown operation ['"](\w+)['"]`)
	matches = unknownOp.FindAllStringSubmatch(msg, -1)
	for _, m := range matches {
		if len(m) > 1 {
			names = append(names, m[1])
		}
	}

	// Pattern: list of suggestions like "X, Y, Z"
	suggestions := regexp.MustCompile(`\b([A-Z][a-zA-Z0-9_]+(?:_[A-Z][a-zA-Z0-9_]+)*)\b`)
	matches = suggestions.FindAllStringSubmatch(msg, -1)
	for _, m := range matches {
		if len(m) > 1 && looksLikeOperation(m[1]) {
			names = append(names, m[1])
		}
	}

	return names
}

// extractFieldNamesFromError parses error messages for field names.
func extractFieldNamesFromError(msg string) []string {
	var names []string

	// Pattern: "Cannot query field 'X' on type 'Y'"
	fieldPattern := regexp.MustCompile(`Cannot query field ['"](\w+)['"]`)
	matches := fieldPattern.FindAllStringSubmatch(msg, -1)
	for _, m := range matches {
		if len(m) > 1 {
			names = append(names, m[1])
		}
	}

	return names
}

// looksLikeOperation checks if a name looks like a GraphQL operation.
func looksLikeOperation(name string) bool {
	if len(name) < 3 {
		return false
	}

	// Operations typically start with uppercase and contain underscores or camelCase
	if name[0] < 'A' || name[0] > 'Z' {
		return false
	}

	// Common operation name patterns
	opPatterns := []string{
		"Query", "Mutation", "Subscription",
		"Page", "Button", "Card", "List",
		"Get", "Create", "Update", "Delete",
		"User", "Channel", "Stream", "Video",
	}

	for _, pattern := range opPatterns {
		if strings.Contains(name, pattern) {
			return true
		}
	}

	// Check for underscore-separated components (like ChannelPage_Query)
	if strings.Contains(name, "_") {
		return true
	}

	return false
}

// inferOperationType guesses the operation type from its name.
func inferOperationType(name string) OperationType {
	lower := strings.ToLower(name)

	// Mutation indicators
	mutationPrefixes := []string{
		"create", "update", "delete", "remove", "add",
		"set", "unset", "follow", "unfollow", "block",
		"unblock", "ban", "unban", "report", "claim",
		"redeem", "send", "join", "leave", "mark",
		"clear", "timeout",
	}
	for _, prefix := range mutationPrefixes {
		if strings.HasPrefix(lower, prefix) || strings.Contains(lower, prefix) {
			return OperationMutation
		}
	}

	// Subscription indicators
	subscriptionPatterns := []string{
		"subscription", "subscribe", "listen", "watch",
		"onupdate", "onchange", "onevent", "onmessage",
	}
	for _, pattern := range subscriptionPatterns {
		if strings.Contains(lower, pattern) {
			return OperationSubscription
		}
	}

	// Check for "On" prefix pattern (common for event subscriptions)
	if strings.HasPrefix(name, "On") && len(name) > 2 && name[2] >= 'A' && name[2] <= 'Z' {
		return OperationSubscription
	}

	return OperationQuery
}

// deduplicateOperations removes duplicate operations, preferring ones with more info.
func deduplicateOperations(ops []Operation) []Operation {
	seen := make(map[string]Operation)

	for _, op := range ops {
		existing, exists := seen[op.Name]
		if !exists {
			seen[op.Name] = op
			continue
		}

		// Prefer operation with more information
		if op.Query != "" && existing.Query == "" {
			seen[op.Name] = op
		} else if op.SHA256Hash != "" && existing.SHA256Hash == "" {
			seen[op.Name] = op
		} else if op.Description != "" && existing.Description == "" {
			seen[op.Name] = op
		} else if len(op.Variables) > len(existing.Variables) {
			seen[op.Name] = op
		}
	}

	result := make([]Operation, 0, len(seen))
	for _, op := range seen {
		result = append(result, op)
	}
	return result
}

// DiscoveryOptions configures the discovery process.
type DiscoveryOptions struct {
	SkipIntrospection bool
	SkipProbing       bool
	SkipKnownList     bool
	ProbePatterns     []string
	Timeout           time.Duration
}

// DiscoverWithOptions performs discovery with custom options.
func (d *Discovery) DiscoverWithOptions(ctx context.Context, opts DiscoveryOptions) (*DiscoveryResult, error) {
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	result := &DiscoveryResult{
		Operations:   make([]Operation, 0),
		DiscoveredAt: time.Now(),
		Sources:      make([]DiscoverySource, 0),
		Errors:       make([]string, 0),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	if !opts.SkipIntrospection {
		wg.Add(1)
		go func() {
			defer wg.Done()
			schema, err := d.introspector.IntrospectSchema(ctx)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if IsIntrospectionDisabled(err) {
					result.IntrospectionEnabled = false
				}
				result.Errors = append(result.Errors, err.Error())
			} else {
				result.Schema = schema
				result.IntrospectionEnabled = true
				result.Sources = append(result.Sources, SourceIntrospection)
				result.Operations = append(result.Operations, d.introspector.GetAllOperations(schema)...)
			}
		}()
	}

	if !opts.SkipProbing {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var probed []Operation
			var err error

			if len(opts.ProbePatterns) > 0 {
				probed, err = d.ProbeWithPatterns(ctx, opts.ProbePatterns)
			} else {
				probed, err = d.ProbeOperations(ctx)
			}

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				result.Errors = append(result.Errors, err.Error())
			} else if len(probed) > 0 {
				result.Sources = append(result.Sources, SourceErrorProbing)
				result.Operations = append(result.Operations, probed...)
			}
		}()
	}

	if !opts.SkipKnownList {
		wg.Add(1)
		go func() {
			defer wg.Done()
			known := d.GetKnownOperations()
			mu.Lock()
			defer mu.Unlock()

			if len(known) > 0 {
				result.Sources = append(result.Sources, SourceKnownList)
				result.Operations = append(result.Operations, known...)
			}
		}()
	}

	wg.Wait()
	result.Operations = deduplicateOperations(result.Operations)

	return result, nil
}
