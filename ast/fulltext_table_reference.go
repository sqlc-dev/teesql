package ast

// FullTextTableReference represents CONTAINSTABLE or FREETEXTTABLE in a FROM clause
type FullTextTableReference struct {
	FullTextFunctionType string                      `json:"FullTextFunctionType,omitempty"` // Contains, FreeText
	TableName            *SchemaObjectName           `json:"TableName,omitempty"`
	Columns              []*ColumnReferenceExpression `json:"Columns,omitempty"`
	SearchCondition      ScalarExpression            `json:"SearchCondition,omitempty"`
	TopN                 ScalarExpression            `json:"TopN,omitempty"`
	Language             ScalarExpression            `json:"Language,omitempty"`
	PropertyName         ScalarExpression            `json:"PropertyName,omitempty"`
	Alias                *Identifier                 `json:"Alias,omitempty"`
	ForPath              bool                        `json:"ForPath"`
}

func (*FullTextTableReference) node()           {}
func (*FullTextTableReference) tableReference() {}
