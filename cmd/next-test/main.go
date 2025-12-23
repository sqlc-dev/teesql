package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type testMetadata struct {
	Todo bool `json:"todo"`
}

type testInfo struct {
	Name     string
	QueryLen int64
}

func main() {
	testdataDir := "parser/testdata"
	entries, err := os.ReadDir(testdataDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading testdata directory: %v\n", err)
		os.Exit(1)
	}

	var todoTests []testInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		testDir := filepath.Join(testdataDir, entry.Name())
		metadataPath := filepath.Join(testDir, "metadata.json")

		metadataData, err := os.ReadFile(metadataPath)
		if err != nil {
			continue
		}

		var metadata testMetadata
		if err := json.Unmarshal(metadataData, &metadata); err != nil {
			continue
		}

		if !metadata.Todo {
			continue
		}

		queryPath := filepath.Join(testDir, "query.sql")
		info, err := os.Stat(queryPath)
		if err != nil {
			continue
		}

		todoTests = append(todoTests, testInfo{
			Name:     entry.Name(),
			QueryLen: info.Size(),
		})
	}

	if len(todoTests) == 0 {
		fmt.Println("No todo tests found!")
		return
	}

	sort.Slice(todoTests, func(i, j int) bool {
		return todoTests[i].QueryLen < todoTests[j].QueryLen
	})

	next := todoTests[0]
	fmt.Printf("%s (%d bytes)\n", next.Name, next.QueryLen)
}
