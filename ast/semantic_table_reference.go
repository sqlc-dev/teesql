package ast

// SemanticTableReference represents SEMANTICKEYPHRASETABLE, SEMANTICSIMILARITYTABLE, or SEMANTICSIMILARITYDETAILSTABLE in a FROM clause
type SemanticTableReference struct {
	SemanticFunctionType string                       `json:"SemanticFunctionType,omitempty"` // SemanticKeyPhraseTable, SemanticSimilarityTable, SemanticSimilarityDetailsTable
	TableName            *SchemaObjectName            `json:"TableName,omitempty"`
	Columns              []*ColumnReferenceExpression `json:"Columns,omitempty"`
	SourceKey            ScalarExpression             `json:"SourceKey,omitempty"`
	MatchedColumn        *ColumnReferenceExpression   `json:"MatchedColumn,omitempty"`
	MatchedKey           ScalarExpression             `json:"MatchedKey,omitempty"`
	Alias                *Identifier                  `json:"Alias,omitempty"`
	ForPath              bool                         `json:"ForPath"`
}

func (*SemanticTableReference) node()           {}
func (*SemanticTableReference) tableReference() {}
