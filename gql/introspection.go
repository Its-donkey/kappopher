package gql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// IntrospectionQuery is the standard GraphQL introspection query.
const IntrospectionQuery = `
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
    types {
      ...FullType
    }
    directives {
      name
      description
      locations
      args {
        ...InputValue
      }
    }
  }
}

fragment FullType on __Type {
  kind
  name
  description
  fields(includeDeprecated: true) {
    name
    description
    args {
      ...InputValue
    }
    type {
      ...TypeRef
    }
    isDeprecated
    deprecationReason
  }
  inputFields {
    ...InputValue
  }
  interfaces {
    ...TypeRef
  }
  enumValues(includeDeprecated: true) {
    name
    description
    isDeprecated
    deprecationReason
  }
  possibleTypes {
    ...TypeRef
  }
}

fragment InputValue on __InputValue {
  name
  description
  type {
    ...TypeRef
  }
  defaultValue
}

fragment TypeRef on __Type {
  kind
  name
  ofType {
    kind
    name
    ofType {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
              }
            }
          }
        }
      }
    }
  }
}
`

// TypeQuery queries a specific type by name.
const TypeQuery = `
query TypeQuery($name: String!) {
  __type(name: $name) {
    kind
    name
    description
    fields(includeDeprecated: true) {
      name
      description
      args {
        name
        description
        type {
          kind
          name
          ofType {
            kind
            name
          }
        }
        defaultValue
      }
      type {
        kind
        name
        ofType {
          kind
          name
        }
      }
      isDeprecated
      deprecationReason
    }
    inputFields {
      name
      description
      type {
        kind
        name
        ofType {
          kind
          name
        }
      }
      defaultValue
    }
    interfaces {
      name
    }
    enumValues(includeDeprecated: true) {
      name
      description
      isDeprecated
      deprecationReason
    }
    possibleTypes {
      name
    }
  }
}
`

// IntrospectionResponse wraps the introspection query result.
type IntrospectionResponse struct {
	Schema *Schema `json:"__schema"`
}

// TypeResponse wraps a type query result.
type TypeResponse struct {
	Type *FullType `json:"__type"`
}

// Introspector handles GraphQL schema introspection.
type Introspector struct {
	client *Client
}

// NewIntrospector creates a new introspector.
func NewIntrospector(client *Client) *Introspector {
	return &Introspector{client: client}
}

// IntrospectSchema performs a full schema introspection.
func (i *Introspector) IntrospectSchema(ctx context.Context) (*Schema, error) {
	req := &Request{
		OperationName: "IntrospectionQuery",
		Query:         IntrospectionQuery,
	}

	resp, err := i.client.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("introspection request failed: %w", err)
	}

	if len(resp.Errors) > 0 {
		// Check if introspection is disabled
		for _, e := range resp.Errors {
			if strings.Contains(strings.ToLower(e.Message), "introspection") {
				return nil, &IntrospectionDisabledError{Message: e.Message}
			}
		}
		return nil, fmt.Errorf("introspection query returned errors: %v", resp.Errors)
	}

	var result IntrospectionResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse introspection response: %w", err)
	}

	return result.Schema, nil
}

// IntrospectType queries a specific type by name.
func (i *Introspector) IntrospectType(ctx context.Context, typeName string) (*FullType, error) {
	req := &Request{
		OperationName: "TypeQuery",
		Query:         TypeQuery,
		Variables: map[string]interface{}{
			"name": typeName,
		},
	}

	resp, err := i.client.Execute(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("type query failed: %w", err)
	}

	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("type query returned errors: %v", resp.Errors)
	}

	var result TypeResponse
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse type response: %w", err)
	}

	if result.Type == nil {
		return nil, fmt.Errorf("type %q not found", typeName)
	}

	return result.Type, nil
}

// GetQueryOperations extracts query operations from the schema.
func (i *Introspector) GetQueryOperations(schema *Schema) []Operation {
	if schema == nil || schema.QueryType == nil {
		return nil
	}
	return i.extractOperationsFromType(schema, schema.QueryType.Name, OperationQuery)
}

// GetMutationOperations extracts mutation operations from the schema.
func (i *Introspector) GetMutationOperations(schema *Schema) []Operation {
	if schema == nil || schema.MutationType == nil {
		return nil
	}
	return i.extractOperationsFromType(schema, schema.MutationType.Name, OperationMutation)
}

// GetSubscriptionOperations extracts subscription operations from the schema.
func (i *Introspector) GetSubscriptionOperations(schema *Schema) []Operation {
	if schema == nil || schema.SubscriptionType == nil {
		return nil
	}
	return i.extractOperationsFromType(schema, schema.SubscriptionType.Name, OperationSubscription)
}

func (i *Introspector) extractOperationsFromType(schema *Schema, typeName string, opType OperationType) []Operation {
	var operations []Operation

	for _, t := range schema.Types {
		if t.Name == typeName {
			for _, field := range t.Fields {
				op := Operation{
					Name:        field.Name,
					Type:        opType,
					Description: field.Description,
					Deprecated:  field.IsDeprecated,
					Source:      SourceIntrospection,
					Variables:   extractVariables(field.Args),
				}
				operations = append(operations, op)
			}
			break
		}
	}

	return operations
}

func extractVariables(args []InputValue) []VariableDefinition {
	var vars []VariableDefinition
	for _, arg := range args {
		v := VariableDefinition{
			Name:     arg.Name,
			Type:     arg.Type.GetTypeString(),
			Required: arg.Type.Kind == TypeKindNonNull,
		}
		if arg.DefaultValue != "" {
			v.DefaultValue = arg.DefaultValue
		}
		vars = append(vars, v)
	}
	return vars
}

// IntrospectionDisabledError indicates introspection is disabled on the server.
type IntrospectionDisabledError struct {
	Message string
}

func (e *IntrospectionDisabledError) Error() string {
	return fmt.Sprintf("introspection disabled: %s", e.Message)
}

// IsIntrospectionDisabled checks if the error indicates disabled introspection.
func IsIntrospectionDisabled(err error) bool {
	_, ok := err.(*IntrospectionDisabledError)
	return ok
}

// GetAllOperations returns all operations from the schema.
func (i *Introspector) GetAllOperations(schema *Schema) []Operation {
	var ops []Operation
	ops = append(ops, i.GetQueryOperations(schema)...)
	ops = append(ops, i.GetMutationOperations(schema)...)
	ops = append(ops, i.GetSubscriptionOperations(schema)...)
	return ops
}

// GetTypeByName finds a type in the schema by name.
func GetTypeByName(schema *Schema, name string) *FullType {
	if schema == nil {
		return nil
	}
	for i := range schema.Types {
		if schema.Types[i].Name == name {
			return &schema.Types[i]
		}
	}
	return nil
}

// GetNonInternalTypes returns types that are not internal GraphQL types.
func GetNonInternalTypes(schema *Schema) []FullType {
	var types []FullType
	for _, t := range schema.Types {
		if !strings.HasPrefix(t.Name, "__") {
			types = append(types, t)
		}
	}
	return types
}
