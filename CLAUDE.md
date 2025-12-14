# Claude Development Guide

## Next Steps

To continue implementing parser support for skipped tests, consult:

```
skipped_tests_by_size.txt
```

This file lists all skipped tests ordered by query file size (smallest first). Smaller tests are generally simpler to implement.

## Workflow

1. Pick tests from `skipped_tests_by_size.txt` starting from the top
2. Check the test's `query.sql` to understand what SQL needs parsing
3. Check the test's `ast.json` to understand the expected output format
4. Implement the necessary AST types in `ast/`
5. Add parser logic in `parser/parser.go`
6. Add JSON marshaling functions in `parser/parser.go`
7. Enable the test by setting `{"skip": false}` in its `metadata.json`
8. Run `go test ./parser/...` to verify
9. **Update `skipped_tests_by_size.txt`** after enabling tests

## Updating skipped_tests_by_size.txt

After enabling tests, regenerate the file:

```bash
cd parser/testdata
for dir in */; do
  if [ -f "$dir/metadata.json" ] && grep -q '"skip": true' "$dir/metadata.json" 2>/dev/null; then
    if [ -f "$dir/query.sql" ]; then
      size=$(wc -c < "$dir/query.sql")
      name="${dir%/}"
      echo "$size $name"
    fi
  fi
done | sort -n > ../../skipped_tests_by_size.txt
```

## Test Structure

Each test in `parser/testdata/` contains:
- `metadata.json` - `{"skip": true}` or `{"skip": false}`
- `query.sql` - T-SQL to parse
- `ast.json` - Expected AST output
