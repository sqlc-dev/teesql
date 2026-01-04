package ast

// FullTextPredicate represents CONTAINS or FREETEXT predicates in WHERE clauses
type FullTextPredicate struct {
	FullTextFunctionType string                       `json:"FullTextFunctionType,omitempty"` // Contains, FreeText
	Columns              []*ColumnReferenceExpression `json:"Columns,omitempty"`
	Value                ScalarExpression             `json:"Value,omitempty"`
	PropertyName         ScalarExpression             `json:"PropertyName,omitempty"`
	LanguageTerm         ScalarExpression             `json:"LanguageTerm,omitempty"`
}

func (*FullTextPredicate) node()              {}
func (*FullTextPredicate) booleanExpression() {}
