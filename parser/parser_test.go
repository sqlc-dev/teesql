package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

type testMetadata struct {
	Todo          bool `json:"todo"`
	InvalidSyntax bool `json:"invalid_syntax"`
}

// Test flag for running todo tests and auto-enabling passing ones
// Usage: go test ./parser/... -check-todo   # run todo tests and auto-update metadata.json for passing tests
var checkTodoTests = flag.Bool("check-todo", false, "run todo tests and auto-update metadata.json for passing tests")

func TestParse(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testName := entry.Name()
		t.Run(testName, func(t *testing.T) {
			testDir := filepath.Join("testdata", testName)

			// Check metadata.json for todo/invalid_syntax flags
			metadataPath := filepath.Join(testDir, "metadata.json")
			metadataData, err := os.ReadFile(metadataPath)
			if err != nil {
				t.Fatalf("failed to read metadata.json: %v", err)
			}

			var metadata testMetadata
			if err := json.Unmarshal(metadataData, &metadata); err != nil {
				t.Fatalf("failed to parse metadata.json: %v", err)
			}

			// Skip tests marked with todo or invalid_syntax unless running with -check-todo
			shouldSkip := metadata.Todo || metadata.InvalidSyntax
			if shouldSkip && !*checkTodoTests {
				t.Skip("skipped via metadata.json (todo or invalid_syntax)")
			}
			if !shouldSkip && *checkTodoTests {
				t.Skip("not a todo/invalid_syntax test")
			}

			// For -check-todo, track if the test passes to update metadata (only for todo, not invalid_syntax)
			checkTodoMode := *checkTodoTests && metadata.Todo && !metadata.InvalidSyntax

			// Read the test SQL file
			sqlPath := filepath.Join(testDir, "query.sql")
			sqlData, err := os.ReadFile(sqlPath)
			if err != nil {
				t.Fatalf("failed to read SQL file: %v", err)
			}

			// Parse the SQL
			ctx := context.Background()
			script, err := Parse(ctx, bytes.NewReader(sqlData))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Marshal to JSON
			gotJSON, err := MarshalScript(script)
			if err != nil {
				t.Fatalf("MarshalScript failed: %v", err)
			}

			// Read expected JSON
			expectedPath := filepath.Join(testDir, "ast.json")
			expectedJSON, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("failed to read expected JSON: %v", err)
			}

			// Normalize both JSONs for comparison
			var gotObj, expectedObj any
			if err := json.Unmarshal(gotJSON, &gotObj); err != nil {
				t.Fatalf("failed to unmarshal got JSON: %v", err)
			}
			if err := json.Unmarshal(expectedJSON, &expectedObj); err != nil {
				t.Fatalf("failed to unmarshal expected JSON: %v", err)
			}

			// Re-marshal for consistent formatting
			gotNormalized, _ := json.MarshalIndent(gotObj, "", "  ")
			expectedNormalized, _ := json.MarshalIndent(expectedObj, "", "  ")

			if string(gotNormalized) != string(expectedNormalized) {
				t.Errorf("JSON mismatch:\ngot:\n%s\n\nexpected:\n%s", gotNormalized, expectedNormalized)
			}

			// If running with -check-todo and the test passed, update metadata.json to remove todo flag
			if checkTodoMode && !t.Failed() {
				newMetadata := "{}\n"
				if err := os.WriteFile(metadataPath, []byte(newMetadata), 0644); err != nil {
					t.Errorf("failed to update metadata.json: %v", err)
				} else {
					t.Logf("ENABLED: updated %s (removed todo flag)", metadataPath)
				}
			}
		})
	}
}
