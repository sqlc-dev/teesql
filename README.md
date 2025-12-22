# teesql

A T-SQL parser for Go that produces JSON AST output compatible with Microsoft's [SqlScriptDOM](https://learn.microsoft.com/en-us/dotnet/api/microsoft.sqlserver.transactsql.scriptdom).

## Installation

```bash
go get github.com/sqlc-dev/teesql
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/sqlc-dev/teesql/parser"
)

func main() {
	sql := "SELECT id, name FROM users WHERE active = 1"

	script, err := parser.Parse(context.Background(), strings.NewReader(sql))
	if err != nil {
		panic(err)
	}

	json, err := parser.MarshalScript(script)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(json))
}
```

Output:

```json
{
  "Batches": [
    {
      "Statements": [
        {
          "QueryExpression": {
            "SelectElements": [
              {"ColumnType": 2, "Identifier": {"Value": "id"}},
              {"ColumnType": 2, "Identifier": {"Value": "name"}}
            ],
            "FromClause": {
              "TableReferences": [
                {"SchemaObject": {"BaseIdentifier": {"Value": "users"}}}
              ]
            },
            "WhereClause": {
              "SearchCondition": {
                "FirstExpression": {"ColumnType": 2, "Identifier": {"Value": "active"}},
                "ComparisonType": 0,
                "SecondExpression": {"Value": "1"}
              }
            }
          }
        }
      ]
    }
  ]
}
```

## Using encoding/json

You can also use the standard library directly:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sqlc-dev/teesql/parser"
)

func main() {
	sql := "SELECT 1"

	script, err := parser.Parse(context.Background(), strings.NewReader(sql))
	if err != nil {
		panic(err)
	}

	out, err := json.MarshalIndent(script, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(out))
}
```
