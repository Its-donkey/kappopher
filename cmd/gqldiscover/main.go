// Command gqldiscover discovers and documents Twitch GraphQL operations.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Its-donkey/helix/gql"
)

func main() {
	// Flags
	clientID := flag.String("client-id", gql.DefaultClientID, "Twitch Client-ID header")
	oauthToken := flag.String("token", "", "OAuth token for authenticated requests")
	outputJSON := flag.String("json", "", "Export results to JSON file")
	outputMD := flag.String("markdown", "", "Export results to Markdown file")
	outputGo := flag.String("go", "", "Export Go types to file")
	goPackage := flag.String("package", "gql", "Package name for Go types export")
	skipIntrospection := flag.Bool("skip-introspection", false, "Skip schema introspection")
	skipProbing := flag.Bool("skip-probing", false, "Skip error probing discovery")
	skipKnown := flag.Bool("skip-known", false, "Skip known operations list")
	timeout := flag.Duration("timeout", 30*time.Second, "Request timeout")
	verbose := flag.Bool("v", false, "Verbose output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "gqldiscover - Discover Twitch GraphQL operations\n\n")
		fmt.Fprintf(os.Stderr, "Usage: gqldiscover [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  gqldiscover -json schema.json -markdown docs.md\n")
		fmt.Fprintf(os.Stderr, "  gqldiscover -v -go types.go -package generated\n")
		fmt.Fprintf(os.Stderr, "  gqldiscover -skip-introspection -json ops.json\n")
	}

	flag.Parse()

	// Create client
	opts := []gql.Option{
		gql.WithClientID(*clientID),
	}
	if *oauthToken != "" {
		opts = append(opts, gql.WithOAuthToken(*oauthToken))
	}

	client := gql.NewClient(opts...)

	// Create discovery
	discovery := gql.NewDiscovery(client)

	// Run discovery
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if *verbose {
		fmt.Println("Starting Twitch GQL discovery...")
		fmt.Printf("  Client-ID: %s\n", *clientID)
		if *oauthToken != "" {
			fmt.Println("  OAuth: enabled")
		}
		fmt.Println()
	}

	discoveryOpts := gql.DiscoveryOptions{
		SkipIntrospection: *skipIntrospection,
		SkipProbing:       *skipProbing,
		SkipKnownList:     *skipKnown,
		Timeout:           *timeout,
	}

	result, err := discovery.DiscoverWithOptions(ctx, discoveryOpts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Discovery failed: %v\n", err)
		os.Exit(1)
	}

	// Print summary
	queries := result.GetOperationsByType(gql.OperationQuery)
	mutations := result.GetOperationsByType(gql.OperationMutation)
	subscriptions := result.GetOperationsByType(gql.OperationSubscription)

	fmt.Println("Discovery Results")
	fmt.Println("=================")
	fmt.Printf("Total Operations: %d\n", len(result.Operations))
	fmt.Printf("  Queries:        %d\n", len(queries))
	fmt.Printf("  Mutations:      %d\n", len(mutations))
	fmt.Printf("  Subscriptions:  %d\n", len(subscriptions))
	fmt.Printf("Introspection:    %v\n", result.IntrospectionEnabled)
	fmt.Printf("Sources:          %v\n", result.Sources)

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors (%d):\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	if result.Schema != nil {
		types := gql.GetNonInternalTypes(result.Schema)
		fmt.Printf("\nSchema Types: %d\n", len(types))
	}

	// Verbose: list operations
	if *verbose && len(result.Operations) > 0 {
		fmt.Println("\nOperations:")
		for _, op := range result.Operations {
			fmt.Printf("  [%s] %s", op.Type, op.Name)
			if op.Description != "" {
				fmt.Printf(" - %s", truncate(op.Description, 50))
			}
			fmt.Println()
		}
	}

	// Export
	exporter := gql.NewExporter(result)

	if *outputJSON != "" {
		if err := exporter.ExportJSON(*outputJSON); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\nExported JSON to: %s\n", *outputJSON)
	}

	if *outputMD != "" {
		if err := exporter.ExportMarkdown(*outputMD); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export Markdown: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported Markdown to: %s\n", *outputMD)
	}

	if *outputGo != "" {
		if err := exporter.ExportGoTypes(*outputGo, *goPackage); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to export Go types: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Exported Go types to: %s\n", *outputGo)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
