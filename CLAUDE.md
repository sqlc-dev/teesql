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
9. **Check if other skipped tests now pass** (see below)
10. **Update `skipped_tests_by_size.txt`** after enabling tests

## Checking for Newly Passing Skipped Tests

After implementing parser changes, run:

```bash
go test ./parser/... -only-skipped -v 2>&1 | grep "PASS:"
```

This shows any skipped tests that now pass. Enable those tests by setting `{"skip": false}` in their `metadata.json`.

Available test flags:
- `-only-skipped` - Run only skipped tests (find newly passing tests)
- `-run-skipped` - Run skipped tests along with normal tests

## Updating skipped_tests_by_size.txt

After enabling tests, regenerate the file. The script only includes tests that:
- Have `"skip": true` in metadata.json
- Do NOT have `"invalid_syntax"` in metadata.json (these can't be implemented)
- Have an `ast.json` file (tests without it are unparseable)

```bash
cd parser/testdata
ls -d */ | while read dir; do
  dir="${dir%/}"
  if [ -f "$dir/metadata.json" ] && [ -f "$dir/ast.json" ] && [ -f "$dir/query.sql" ]; then
    if grep -q '"skip": true' "$dir/metadata.json" 2>/dev/null; then
      if grep -qv '"invalid_syntax"' "$dir/metadata.json" 2>/dev/null; then
        size=$(wc -c < "$dir/query.sql")
        echo "$size $dir"
      fi
    fi
  fi
done | sort -n > ../../skipped_tests_by_size.txt
```

## Test Structure

Each test in `parser/testdata/` contains:
- `metadata.json` - `{"skip": true}` or `{"skip": false}`
- `query.sql` - T-SQL to parse
- `ast.json` - Expected AST output

## Important Rules

- **NEVER modify `ast.json` files** - These are golden files containing the expected output. If tests fail due to JSON mismatches, fix the Go code to match the expected output, not the other way around.

## Generating ast.json with TsqlAstParser

The `TsqlAstParser/` directory contains a C# tool that generates `ast.json` files using Microsoft's official T-SQL parser (ScriptDom).

### Prerequisites

1. Install .NET 8.0 SDK:
   ```bash
   curl -sSL https://dot.net/v1/dotnet-install.sh | bash /dev/stdin --channel 8.0 --install-dir ~/.dotnet
   ```

2. Download the NuGet package (if `packages/` directory is empty):
   ```bash
   mkdir -p packages
   curl -L -o packages/microsoft.sqlserver.transactsql.scriptdom.170.128.0.nupkg \
     "https://api.nuget.org/v3-flatcontainer/microsoft.sqlserver.transactsql.scriptdom/170.128.0/microsoft.sqlserver.transactsql.scriptdom.170.128.0.nupkg"
   ```

3. Build the tool:
   ```bash
   ~/.dotnet/dotnet build TsqlAstParser -c Release
   ```

### Usage

Generate `ast.json` for a single test:
```bash
~/.dotnet/dotnet run --project TsqlAstParser -c Release -- parser/testdata/TestName/query.sql parser/testdata/TestName/ast.json
```

Generate `ast.json` for all tests missing it:
```bash
for dir in parser/testdata/*/; do
  if [ -f "$dir/query.sql" ] && [ ! -f "$dir/ast.json" ]; then
    ~/.dotnet/dotnet run --project TsqlAstParser -c Release -- "$dir/query.sql" "$dir/ast.json"
  fi
done
```

### Limitations

TsqlAstParser uses TSql160Parser (SQL Server 2022) and cannot parse:
- SQL Server 170+ features (VECTOR indexes, AI functions, JSON enhancements)
- Fabric DW-specific syntax (CLONE TABLE, CLUSTER BY)
- Deprecated syntax removed in newer versions
- Intentionally invalid SQL (error test cases)

Tests for unsupported syntax will not have `ast.json` files generated.
