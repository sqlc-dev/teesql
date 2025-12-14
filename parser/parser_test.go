package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type testMetadata struct {
	Skip bool `json:"skip"`
}

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

			// Check metadata.json for skip flag
			metadataPath := filepath.Join(testDir, "metadata.json")
			metadataData, err := os.ReadFile(metadataPath)
			if err != nil {
				t.Fatalf("failed to read metadata.json: %v", err)
			}

			var metadata testMetadata
			if err := json.Unmarshal(metadataData, &metadata); err != nil {
				t.Fatalf("failed to parse metadata.json: %v", err)
			}

			if metadata.Skip {
				t.Skip("skipped via metadata.json")
			}

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
		})
	}
}
