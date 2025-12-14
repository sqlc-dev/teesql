package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParse_ZeroLengthFile(t *testing.T) {
	// Read the test SQL file
	sqlPath := filepath.Join("testdata", "ZeroLengthFile", "query.sql")
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
	expectedPath := filepath.Join("testdata", "ZeroLengthFile", "ast.json")
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
}
