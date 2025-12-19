// Package gql provides a client for Twitch's unofficial GraphQL API.
// This package enables discovery, documentation, and execution of GraphQL operations.
package gql

import (
	"encoding/json"
	"time"
)

// TwitchGQLEndpoint is the base URL for Twitch's GraphQL API.
const TwitchGQLEndpoint = "https://gql.twitch.tv/gql"

// DefaultClientID is the web client ID commonly used for unauthenticated requests.
const DefaultClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"

// Request represents a GraphQL request.
type Request struct {
	OperationName string                 `json:"operationName,omitempty"`
	Query         string                 `json:"query,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    *Extensions            `json:"extensions,omitempty"`
}

// PersistedRequest represents a persisted query request using SHA256 hash.
type PersistedRequest struct {
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    *Extensions            `json:"extensions,omitempty"`
}

// Extensions contains request extensions like persisted queries.
type Extensions struct {
	PersistedQuery *PersistedQuery `json:"persistedQuery,omitempty"`
}

// PersistedQuery identifies a query by its SHA256 hash.
type PersistedQuery struct {
	Version    int    `json:"version"`
	SHA256Hash string `json:"sha256Hash"`
}

// Response represents a GraphQL response.
type Response struct {
	Data       json.RawMessage `json:"data,omitempty"`
	Errors     []Error         `json:"errors,omitempty"`
	Extensions json.RawMessage `json:"extensions,omitempty"`
}

// Error represents a GraphQL error.
type Error struct {
	Message    string                 `json:"message"`
	Locations  []Location             `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Location represents a position in a GraphQL document.
type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// OperationType represents the type of GraphQL operation.
type OperationType string

const (
	OperationQuery        OperationType = "query"
	OperationMutation     OperationType = "mutation"
	OperationSubscription OperationType = "subscription"
)

// Operation represents a discovered GraphQL operation.
type Operation struct {
	Name        string                 `json:"name"`
	Type        OperationType          `json:"type"`
	Query       string                 `json:"query,omitempty"`
	SHA256Hash  string                 `json:"sha256Hash,omitempty"`
	Variables   []VariableDefinition   `json:"variables,omitempty"`
	Description string                 `json:"description,omitempty"`
	Deprecated  bool                   `json:"deprecated,omitempty"`
	Source      DiscoverySource        `json:"source"`
	DiscoveredAt time.Time             `json:"discoveredAt"`
}

// VariableDefinition describes a variable in an operation.
type VariableDefinition struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Required     bool        `json:"required"`
}

// DiscoverySource indicates how an operation was discovered.
type DiscoverySource string

const (
	SourceIntrospection DiscoverySource = "introspection"
	SourceErrorProbing  DiscoverySource = "error_probing"
	SourceKnownList     DiscoverySource = "known_list"
	SourceManual        DiscoverySource = "manual"
)

// Schema represents the full GraphQL introspection schema.
type Schema struct {
	QueryType        *TypeRef     `json:"queryType"`
	MutationType     *TypeRef     `json:"mutationType,omitempty"`
	SubscriptionType *TypeRef     `json:"subscriptionType,omitempty"`
	Types            []FullType   `json:"types"`
	Directives       []Directive  `json:"directives"`
}

// TypeRef is a reference to a type by name.
type TypeRef struct {
	Name string `json:"name"`
}

// FullType represents a complete GraphQL type from introspection.
type FullType struct {
	Kind          TypeKind     `json:"kind"`
	Name          string       `json:"name"`
	Description   string       `json:"description,omitempty"`
	Fields        []Field      `json:"fields,omitempty"`
	InputFields   []InputValue `json:"inputFields,omitempty"`
	Interfaces    []TypeRef    `json:"interfaces,omitempty"`
	EnumValues    []EnumValue  `json:"enumValues,omitempty"`
	PossibleTypes []TypeRef    `json:"possibleTypes,omitempty"`
}

// TypeKind represents the kind of a GraphQL type.
type TypeKind string

const (
	TypeKindScalar      TypeKind = "SCALAR"
	TypeKindObject      TypeKind = "OBJECT"
	TypeKindInterface   TypeKind = "INTERFACE"
	TypeKindUnion       TypeKind = "UNION"
	TypeKindEnum        TypeKind = "ENUM"
	TypeKindInputObject TypeKind = "INPUT_OBJECT"
	TypeKindList        TypeKind = "LIST"
	TypeKindNonNull     TypeKind = "NON_NULL"
)

// Field represents a field on a GraphQL type.
type Field struct {
	Name              string       `json:"name"`
	Description       string       `json:"description,omitempty"`
	Args              []InputValue `json:"args"`
	Type              TypeReference `json:"type"`
	IsDeprecated      bool         `json:"isDeprecated"`
	DeprecationReason string       `json:"deprecationReason,omitempty"`
}

// InputValue represents an argument or input field.
type InputValue struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	Type         TypeReference `json:"type"`
	DefaultValue string        `json:"defaultValue,omitempty"`
}

// TypeReference is a recursive type reference for nested types.
type TypeReference struct {
	Kind   TypeKind       `json:"kind"`
	Name   string         `json:"name,omitempty"`
	OfType *TypeReference `json:"ofType,omitempty"`
}

// EnumValue represents a value in a GraphQL enum.
type EnumValue struct {
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	IsDeprecated      bool   `json:"isDeprecated"`
	DeprecationReason string `json:"deprecationReason,omitempty"`
}

// Directive represents a GraphQL directive.
type Directive struct {
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Locations   []string     `json:"locations"`
	Args        []InputValue `json:"args"`
}

// DiscoveryResult contains all discovered operations and schema information.
type DiscoveryResult struct {
	Schema         *Schema               `json:"schema,omitempty"`
	Operations     []Operation           `json:"operations"`
	DiscoveredAt   time.Time             `json:"discoveredAt"`
	Sources        []DiscoverySource     `json:"sources"`
	Errors         []string              `json:"errors,omitempty"`
	IntrospectionEnabled bool            `json:"introspectionEnabled"`
}

// GetOperationsByType returns operations filtered by type.
func (dr *DiscoveryResult) GetOperationsByType(opType OperationType) []Operation {
	var result []Operation
	for _, op := range dr.Operations {
		if op.Type == opType {
			result = append(result, op)
		}
	}
	return result
}

// GetOperationByName returns an operation by name.
func (dr *DiscoveryResult) GetOperationByName(name string) *Operation {
	for i := range dr.Operations {
		if dr.Operations[i].Name == name {
			return &dr.Operations[i]
		}
	}
	return nil
}

// GetTypeString returns a human-readable string for a type reference.
func (tr *TypeReference) GetTypeString() string {
	if tr == nil {
		return ""
	}
	switch tr.Kind {
	case TypeKindNonNull:
		if tr.OfType != nil {
			return tr.OfType.GetTypeString() + "!"
		}
	case TypeKindList:
		if tr.OfType != nil {
			return "[" + tr.OfType.GetTypeString() + "]"
		}
	default:
		return tr.Name
	}
	return ""
}
