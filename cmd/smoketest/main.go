//go:build smoketest

// Quick smoke test against a live Kion instance.
//
// Targets the v3_16 sub-package specifically. The `smoketest` build tag
// keeps it out of the default `go build ./...` because the field access
// below is tied to the shape of v3_16's generated types — if v3.16 gets
// deprecated or its response schema changes, update this file (or bump
// to a newer version) and rebuild.
//
// Build with: make build-smoketest
// Run with:   KION_URL=... KION_API_KEY=... ./smoketest
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	kion "github.com/kionsoftware/kion-sdk-go"
	v3_16 "github.com/kionsoftware/kion-sdk-go/generated/v3_16"
)

func main() {
	baseURL := os.Getenv("KION_URL")
	apiKey := os.Getenv("KION_API_KEY")
	if baseURL == "" || apiKey == "" {
		log.Fatal("Set KION_URL and KION_API_KEY")
	}

	client, err := v3_16.New(baseURL, kion.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("v3_16.New: %v", err)
	}

	ctx := context.Background()

	// Try listing labels
	fmt.Println("=== GET /v3/label ===")
	labelsRes, err := client.GetLabelIndex(ctx, v3_16.GetLabelIndexParams{})
	if err != nil {
		fmt.Printf("GetLabelIndex error: %v\n", err)
		fmt.Printf("  StatusCode: %d\n", kion.StatusCode(err))
		return
	}

	switch v := labelsRes.(type) {
	case *v3_16.LabelListPaginatedResponse:
		data := v.Data.Value
		items := data.Items.Value
		fmt.Printf("OK: %d labels returned\n", len(items))
		for i, l := range items {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(items)-5)
				break
			}
			fmt.Printf("  [%d] %s:%s (color: #%s)\n", l.ID.Value, l.Key, l.Value, l.Color)
		}
	case *v3_16.UnauthorizedResponse:
		fmt.Println("ERROR: 401 Unauthorized")
	case *v3_16.BadRequestResponse:
		fmt.Println("ERROR: 400 Bad Request")
	case *v3_16.InternalServerErrorResponse:
		fmt.Println("ERROR: 500 Internal Server Error")
	default:
		fmt.Printf("Unexpected response type: %T\n", v)
	}
}
