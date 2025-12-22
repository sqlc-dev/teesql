# teesql

A T-SQL parser for Go that produces JSON AST output compatible with Microsoft's SqlScriptDOM.

## Installation

```bash
go get github.com/kyleconroy/teesql
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyleconroy/teesql/parser"
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

## License

MIT
