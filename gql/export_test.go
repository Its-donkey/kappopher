package gql

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewExporter(t *testing.T) {
	result := &DiscoveryResult{}
	exporter := NewExporter(result)

	if exporter.result != result {
		t.Error("result not set correctly")
	}
}

func TestExportJSONBytes(t *testing.T) {
	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "TestOp", Type: OperationQuery},
		},
		DiscoveredAt: time.Now(),
		Sources:      []DiscoverySource{SourceKnownList},
	}

	exporter := NewExporter(result)
	data, err := exporter.ExportJSONBytes()
	if err != nil {
		t.Fatalf("ExportJSONBytes failed: %v", err)
	}

	var parsed DiscoveryResult
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(parsed.Operations) != 1 {
		t.Errorf("expected 1 operation, got %d", len(parsed.Operations))
	}

	if parsed.Operations[0].Name != "TestOp" {
		t.Errorf("expected TestOp, got %s", parsed.Operations[0].Name)
	}
}

func TestExportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "TestOp", Type: OperationQuery},
		},
	}

	exporter := NewExporter(result)
	if err := exporter.ExportJSON(filePath); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty file")
	}
}

func TestExportMarkdownString(t *testing.T) {
	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "GetUser", Type: OperationQuery, Description: "Get user details"},
			{Name: "CreateClip", Type: OperationMutation, Description: "Create a clip"},
		},
		DiscoveredAt:         time.Now(),
		Sources:              []DiscoverySource{SourceKnownList},
		IntrospectionEnabled: true,
	}

	exporter := NewExporter(result)
	md := exporter.ExportMarkdownString()

	// Check for expected content
	if !strings.Contains(md, "# Twitch GraphQL Operations") {
		t.Error("missing title")
	}

	if !strings.Contains(md, "## Summary") {
		t.Error("missing summary section")
	}

	if !strings.Contains(md, "## Queries") {
		t.Error("missing queries section")
	}

	if !strings.Contains(md, "## Mutations") {
		t.Error("missing mutations section")
	}

	if !strings.Contains(md, "GetUser") {
		t.Error("missing GetUser operation")
	}

	if !strings.Contains(md, "CreateClip") {
		t.Error("missing CreateClip operation")
	}

	if !strings.Contains(md, "Get user details") {
		t.Error("missing operation description")
	}
}

func TestExportMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "subdir", "test.md")

	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "TestOp", Type: OperationQuery},
		},
	}

	exporter := NewExporter(result)
	if err := exporter.ExportMarkdown(filePath); err != nil {
		t.Fatalf("ExportMarkdown failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(data), "TestOp") {
		t.Error("expected TestOp in markdown")
	}
}

func TestExportGoTypesString(t *testing.T) {
	result := &DiscoveryResult{
		Schema: &Schema{
			Types: []FullType{
				{
					Kind:        TypeKindObject,
					Name:        "User",
					Description: "A Twitch user",
					Fields: []Field{
						{
							Name: "id",
							Type: TypeReference{Kind: TypeKindNonNull, OfType: &TypeReference{Kind: TypeKindScalar, Name: "ID"}},
						},
						{
							Name: "login",
							Type: TypeReference{Kind: TypeKindScalar, Name: "String"},
						},
						{
							Name:         "displayName",
							Type:         TypeReference{Kind: TypeKindScalar, Name: "String"},
							IsDeprecated: true,
							DeprecationReason: "Use display_name instead",
						},
					},
				},
				{
					Kind:        TypeKindEnum,
					Name:        "UserType",
					Description: "Type of user account",
					EnumValues: []EnumValue{
						{Name: "NORMAL", Description: "Regular user"},
						{Name: "STAFF", Description: "Twitch staff"},
						{Name: "ADMIN", Description: "Twitch admin", IsDeprecated: true},
					},
				},
				{
					Kind: TypeKindScalar,
					Name: "__Schema", // Internal type, should be skipped
				},
			},
		},
		DiscoveredAt: time.Now(),
	}

	exporter := NewExporter(result)
	code := exporter.ExportGoTypesString("gql")

	// Check for expected content
	if !strings.Contains(code, "package gql") {
		t.Error("missing package declaration")
	}

	if !strings.Contains(code, "type User struct") {
		t.Error("missing User struct")
	}

	if !strings.Contains(code, "type UserType string") {
		t.Error("missing UserType enum")
	}

	if !strings.Contains(code, "UserTypeNORMAL") {
		t.Error("missing enum constant")
	}

	if !strings.Contains(code, "json:\"id\"") {
		t.Error("missing json tag for id")
	}

	if !strings.Contains(code, "Deprecated:") {
		t.Error("missing deprecation comment")
	}
}

func TestExportGoTypesNoSchema(t *testing.T) {
	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "TestOp", Type: OperationQuery},
		},
	}

	exporter := NewExporter(result)
	code := exporter.ExportGoTypesString("gql")

	if !strings.Contains(code, "No schema available") {
		t.Error("expected 'no schema' message")
	}
}

func TestExportGoTypes(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "types.go")

	result := &DiscoveryResult{
		Schema: &Schema{
			Types: []FullType{
				{
					Kind: TypeKindObject,
					Name: "User",
					Fields: []Field{
						{Name: "id", Type: TypeReference{Kind: TypeKindScalar, Name: "ID"}},
					},
				},
			},
		},
	}

	exporter := NewExporter(result)
	if err := exporter.ExportGoTypes(filePath, "generated"); err != nil {
		t.Fatalf("ExportGoTypes failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(data), "package generated") {
		t.Error("expected package declaration")
	}
}

func TestToGoName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_id", "UserID"},
		{"display_name", "DisplayName"},
		{"url", "URL"},
		{"html_content", "HTMLContent"},
		{"api_key", "APIKey"},
		{"", ""},
		{"ID", "ID"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toGoName(tt.input)
			if result != tt.expected {
				t.Errorf("toGoName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTypeRefToGo(t *testing.T) {
	tests := []struct {
		name     string
		input    TypeReference
		expected string
	}{
		{
			name:     "scalar string",
			input:    TypeReference{Kind: TypeKindScalar, Name: "String"},
			expected: "*string",
		},
		{
			name:     "non-null string",
			input:    TypeReference{Kind: TypeKindNonNull, OfType: &TypeReference{Kind: TypeKindScalar, Name: "String"}},
			expected: "string",
		},
		{
			name:     "list of strings",
			input:    TypeReference{Kind: TypeKindList, OfType: &TypeReference{Kind: TypeKindScalar, Name: "String"}},
			expected: "[]*string",
		},
		{
			name: "non-null list",
			input: TypeReference{
				Kind: TypeKindNonNull,
				OfType: &TypeReference{
					Kind:   TypeKindList,
					OfType: &TypeReference{Kind: TypeKindScalar, Name: "Int"},
				},
			},
			expected: "[]*int",
		},
		{
			name:     "custom type",
			input:    TypeReference{Kind: TypeKindObject, Name: "User"},
			expected: "*User",
		},
		{
			name:     "ID scalar",
			input:    TypeReference{Kind: TypeKindScalar, Name: "ID"},
			expected: "*string",
		},
		{
			name:     "Boolean",
			input:    TypeReference{Kind: TypeKindScalar, Name: "Boolean"},
			expected: "*bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typeRefToGo(tt.input)
			if result != tt.expected {
				t.Errorf("typeRefToGo = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSortOperations(t *testing.T) {
	ops := []Operation{
		{Name: "Zebra"},
		{Name: "Apple"},
		{Name: "Mango"},
	}

	sortOperations(ops)

	if ops[0].Name != "Apple" {
		t.Errorf("expected Apple first, got %s", ops[0].Name)
	}
	if ops[1].Name != "Mango" {
		t.Errorf("expected Mango second, got %s", ops[1].Name)
	}
	if ops[2].Name != "Zebra" {
		t.Errorf("expected Zebra third, got %s", ops[2].Name)
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"hello", 5, "hello"},
		{"", 5, ""},
	}

	for _, tt := range tests {
		result := truncate(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestExportSchemaJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "schema.json")

	result := &DiscoveryResult{
		Schema: &Schema{
			QueryType: &TypeRef{Name: "Query"},
			Types: []FullType{
				{Kind: TypeKindObject, Name: "Query"},
			},
		},
	}

	exporter := NewExporter(result)
	if err := exporter.ExportSchemaJSON(filePath); err != nil {
		t.Fatalf("ExportSchemaJSON failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var schema Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	if schema.QueryType.Name != "Query" {
		t.Errorf("expected QueryType 'Query', got %q", schema.QueryType.Name)
	}
}

func TestExportSchemaJSONNoSchema(t *testing.T) {
	result := &DiscoveryResult{}
	exporter := NewExporter(result)

	err := exporter.ExportSchemaJSON("/tmp/test.json")
	if err == nil {
		t.Error("expected error for nil schema")
	}
}

func TestExportOperationsJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "ops.json")

	result := &DiscoveryResult{
		Operations: []Operation{
			{Name: "Op1", Type: OperationQuery},
			{Name: "Op2", Type: OperationMutation},
		},
	}

	exporter := NewExporter(result)
	if err := exporter.ExportOperationsJSON(filePath); err != nil {
		t.Fatalf("ExportOperationsJSON failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var ops []Operation
	if err := json.Unmarshal(data, &ops); err != nil {
		t.Fatalf("failed to parse operations: %v", err)
	}

	if len(ops) != 2 {
		t.Errorf("expected 2 operations, got %d", len(ops))
	}
}

func TestFormatComment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Simple comment", "Simple comment"},
		{"  Trimmed  ", "Trimmed"},
		{"Multi\nline\ncomment", "Multi line comment"},
	}

	for _, tt := range tests {
		result := formatComment(tt.input)
		if result != tt.expected {
			t.Errorf("formatComment(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
