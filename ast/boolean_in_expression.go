package ast

// BooleanInExpression represents an IN expression.
type BooleanInExpression struct {
	Fragment
	Expression ScalarExpression
	NotDefined bool
	Values     []ScalarExpression
	Subquery   QueryExpression
	// SubqueryFragment carries the source span of the parenthesized
	// subquery, which ScriptDom records on its synthesized ScalarSubquery.
	SubqueryFragment Fragment
}

func (b *BooleanInExpression) node()              {}
func (b *BooleanInExpression) booleanExpression() {}
