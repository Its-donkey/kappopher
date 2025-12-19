package gql

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

// Exporter handles exporting discovery results to various formats.
type Exporter struct {
	result *DiscoveryResult
}

// NewExporter creates a new exporter for the given discovery result.
func NewExporter(result *DiscoveryResult) *Exporter {
	return &Exporter{result: result}
}

// ExportJSON exports the discovery result to a JSON file.
func (e *Exporter) ExportJSON(filePath string) error {
	data, err := json.MarshalIndent(e.result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportJSONBytes returns the discovery result as JSON bytes.
func (e *Exporter) ExportJSONBytes() ([]byte, error) {
	return json.MarshalIndent(e.result, "", "  ")
}

// ExportMarkdown exports the discovery result to a Markdown file.
func (e *Exporter) ExportMarkdown(filePath string) error {
	content := e.generateMarkdown()

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportMarkdownString returns the discovery result as Markdown string.
func (e *Exporter) ExportMarkdownString() string {
	return e.generateMarkdown()
}

func (e *Exporter) generateMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Twitch GraphQL Operations\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", e.result.DiscoveredAt.Format(time.RFC3339)))

	// Summary
	sb.WriteString("## Summary\n\n")
	queries := e.result.GetOperationsByType(OperationQuery)
	mutations := e.result.GetOperationsByType(OperationMutation)
	subscriptions := e.result.GetOperationsByType(OperationSubscription)

	sb.WriteString(fmt.Sprintf("- **Total Operations**: %d\n", len(e.result.Operations)))
	sb.WriteString(fmt.Sprintf("- **Queries**: %d\n", len(queries)))
	sb.WriteString(fmt.Sprintf("- **Mutations**: %d\n", len(mutations)))
	sb.WriteString(fmt.Sprintf("- **Subscriptions**: %d\n", len(subscriptions)))
	sb.WriteString(fmt.Sprintf("- **Introspection Enabled**: %v\n", e.result.IntrospectionEnabled))
	sb.WriteString(fmt.Sprintf("- **Discovery Sources**: %v\n\n", e.result.Sources))

	// Queries
	if len(queries) > 0 {
		sb.WriteString("## Queries\n\n")
		sortOperations(queries)
		for _, op := range queries {
			e.writeOperation(&sb, op)
		}
	}

	// Mutations
	if len(mutations) > 0 {
		sb.WriteString("## Mutations\n\n")
		sortOperations(mutations)
		for _, op := range mutations {
			e.writeOperation(&sb, op)
		}
	}

	// Subscriptions
	if len(subscriptions) > 0 {
		sb.WriteString("## Subscriptions\n\n")
		sortOperations(subscriptions)
		for _, op := range subscriptions {
			e.writeOperation(&sb, op)
		}
	}

	// Schema types (if available)
	if e.result.Schema != nil {
		types := GetNonInternalTypes(e.result.Schema)
		if len(types) > 0 {
			sb.WriteString("## Types\n\n")
			sb.WriteString(fmt.Sprintf("Total types: %d\n\n", len(types)))

			// Group by kind
			byKind := make(map[TypeKind][]FullType)
			for _, t := range types {
				byKind[t.Kind] = append(byKind[t.Kind], t)
			}

			for _, kind := range []TypeKind{TypeKindObject, TypeKindEnum, TypeKindInterface, TypeKindUnion, TypeKindInputObject, TypeKindScalar} {
				if list, ok := byKind[kind]; ok && len(list) > 0 {
					sb.WriteString(fmt.Sprintf("### %s (%d)\n\n", kind, len(list)))
					for _, t := range list {
						sb.WriteString(fmt.Sprintf("- `%s`", t.Name))
						if t.Description != "" {
							sb.WriteString(fmt.Sprintf(" - %s", truncate(t.Description, 80)))
						}
						sb.WriteString("\n")
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	// Errors
	if len(e.result.Errors) > 0 {
		sb.WriteString("## Discovery Errors\n\n")
		for _, err := range e.result.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (e *Exporter) writeOperation(sb *strings.Builder, op Operation) {
	sb.WriteString(fmt.Sprintf("### %s\n\n", op.Name))

	if op.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", op.Description))
	}

	if op.Deprecated {
		sb.WriteString("**DEPRECATED**\n\n")
	}

	sb.WriteString(fmt.Sprintf("- **Type**: %s\n", op.Type))
	sb.WriteString(fmt.Sprintf("- **Source**: %s\n", op.Source))

	if op.SHA256Hash != "" {
		sb.WriteString(fmt.Sprintf("- **Hash**: `%s`\n", op.SHA256Hash))
	}

	if len(op.Variables) > 0 {
		sb.WriteString("- **Variables**:\n")
		for _, v := range op.Variables {
			required := ""
			if v.Required {
				required = " (required)"
			}
			sb.WriteString(fmt.Sprintf("  - `%s`: %s%s\n", v.Name, v.Type, required))
		}
	}

	if op.Query != "" {
		sb.WriteString("\n```graphql\n")
		sb.WriteString(op.Query)
		sb.WriteString("\n```\n")
	}

	sb.WriteString("\n")
}

// ExportGoTypes exports discovered types as Go structs.
func (e *Exporter) ExportGoTypes(filePath string, packageName string) error {
	content := e.generateGoTypes(packageName)

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportGoTypesString returns discovered types as Go code string.
func (e *Exporter) ExportGoTypesString(packageName string) string {
	return e.generateGoTypes(packageName)
}

func (e *Exporter) generateGoTypes(packageName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("// Code generated by gql discovery. DO NOT EDIT.\n"))
	sb.WriteString(fmt.Sprintf("// Generated: %s\n\n", e.result.DiscoveredAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	if e.result.Schema == nil {
		sb.WriteString("// No schema available for type generation.\n")
		sb.WriteString("// Run introspection to generate types.\n")
		return sb.String()
	}

	types := GetNonInternalTypes(e.result.Schema)
	if len(types) == 0 {
		sb.WriteString("// No types found in schema.\n")
		return sb.String()
	}

	// Generate enums first
	for _, t := range types {
		if t.Kind == TypeKindEnum {
			e.writeGoEnum(&sb, t)
		}
	}

	// Generate structs for object types
	for _, t := range types {
		if t.Kind == TypeKindObject || t.Kind == TypeKindInputObject {
			e.writeGoStruct(&sb, t)
		}
	}

	// Generate interfaces
	for _, t := range types {
		if t.Kind == TypeKindInterface {
			e.writeGoInterface(&sb, t)
		}
	}

	return sb.String()
}

func (e *Exporter) writeGoEnum(sb *strings.Builder, t FullType) {
	goName := toGoName(t.Name)

	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s %s\n", goName, formatComment(t.Description)))
	}
	sb.WriteString(fmt.Sprintf("type %s string\n\n", goName))

	if len(t.EnumValues) > 0 {
		sb.WriteString("const (\n")
		for _, v := range t.EnumValues {
			constName := fmt.Sprintf("%s%s", goName, toGoName(v.Name))
			if v.Description != "" {
				sb.WriteString(fmt.Sprintf("\t// %s\n", formatComment(v.Description)))
			}
			if v.IsDeprecated {
				sb.WriteString(fmt.Sprintf("\t// Deprecated: %s\n", v.DeprecationReason))
			}
			sb.WriteString(fmt.Sprintf("\t%s %s = %q\n", constName, goName, v.Name))
		}
		sb.WriteString(")\n\n")
	}
}

func (e *Exporter) writeGoStruct(sb *strings.Builder, t FullType) {
	goName := toGoName(t.Name)

	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s %s\n", goName, formatComment(t.Description)))
	}
	sb.WriteString(fmt.Sprintf("type %s struct {\n", goName))

	fields := t.Fields
	if t.Kind == TypeKindInputObject {
		fields = make([]Field, len(t.InputFields))
		for i, f := range t.InputFields {
			fields[i] = Field{
				Name:        f.Name,
				Description: f.Description,
				Type:        f.Type,
			}
		}
	}

	for _, f := range fields {
		fieldName := toGoName(f.Name)
		fieldType := typeRefToGo(f.Type)
		jsonTag := f.Name

		if f.Description != "" {
			sb.WriteString(fmt.Sprintf("\t// %s\n", formatComment(f.Description)))
		}
		if f.IsDeprecated {
			sb.WriteString(fmt.Sprintf("\t// Deprecated: %s\n", f.DeprecationReason))
		}

		// Add omitempty for nullable types
		omitempty := ""
		if f.Type.Kind != TypeKindNonNull {
			omitempty = ",omitempty"
		}

		sb.WriteString(fmt.Sprintf("\t%s %s `json:\"%s%s\"`\n", fieldName, fieldType, jsonTag, omitempty))
	}

	sb.WriteString("}\n\n")
}

func (e *Exporter) writeGoInterface(sb *strings.Builder, t FullType) {
	goName := toGoName(t.Name)

	if t.Description != "" {
		sb.WriteString(fmt.Sprintf("// %s %s\n", goName, formatComment(t.Description)))
	}
	sb.WriteString(fmt.Sprintf("type %s interface {\n", goName))
	sb.WriteString(fmt.Sprintf("\tIs%s()\n", goName))
	sb.WriteString("}\n\n")
}

// typeRefToGo converts a GraphQL type reference to a Go type.
func typeRefToGo(tr TypeReference) string {
	switch tr.Kind {
	case TypeKindNonNull:
		if tr.OfType != nil {
			inner := typeRefToGo(*tr.OfType)
			// Remove pointer for non-null types
			return strings.TrimPrefix(inner, "*")
		}
	case TypeKindList:
		if tr.OfType != nil {
			return "[]" + typeRefToGo(*tr.OfType)
		}
	default:
		return "*" + scalarToGo(tr.Name)
	}
	return "interface{}"
}

// scalarToGo maps GraphQL scalar types to Go types.
func scalarToGo(name string) string {
	switch name {
	case "String", "ID":
		return "string"
	case "Int":
		return "int"
	case "Float":
		return "float64"
	case "Boolean":
		return "bool"
	case "DateTime", "Time":
		return "time.Time"
	default:
		return toGoName(name)
	}
}

// toGoName converts a GraphQL name to a Go-style exported name.
func toGoName(name string) string {
	if name == "" {
		return ""
	}

	// Handle common acronyms
	acronyms := map[string]string{
		"id":   "ID",
		"url":  "URL",
		"uri":  "URI",
		"html": "HTML",
		"json": "JSON",
		"api":  "API",
		"ui":   "UI",
	}

	// Split by underscores and capitalize each part
	parts := strings.Split(name, "_")
	var result strings.Builder

	for _, part := range parts {
		if part == "" {
			continue
		}

		lower := strings.ToLower(part)
		if acronym, ok := acronyms[lower]; ok {
			result.WriteString(acronym)
		} else {
			// Capitalize first letter
			runes := []rune(part)
			runes[0] = unicode.ToUpper(runes[0])
			result.WriteString(string(runes))
		}
	}

	return result.String()
}

// formatComment formats a description for use as a Go comment.
func formatComment(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// truncate shortens a string to the given length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// sortOperations sorts operations alphabetically by name.
func sortOperations(ops []Operation) {
	sort.Slice(ops, func(i, j int) bool {
		return ops[i].Name < ops[j].Name
	})
}

// ExportSchemaJSON exports just the schema as JSON.
func (e *Exporter) ExportSchemaJSON(filePath string) error {
	if e.result.Schema == nil {
		return fmt.Errorf("no schema available")
	}

	data, err := json.MarshalIndent(e.result.Schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// ExportOperationsJSON exports just the operations as JSON.
func (e *Exporter) ExportOperationsJSON(filePath string) error {
	data, err := json.MarshalIndent(e.result.Operations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal operations: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}
