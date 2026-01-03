package ast

// SubqueryComparisonPredicate represents a comparison with a subquery using ANY/SOME/ALL.
// Example: col IS DISTINCT FROM SOME (SELECT ...), col > ALL (SELECT ...)
type SubqueryComparisonPredicate struct {
	Expression                      ScalarExpression
	ComparisonType                  string // "IsDistinctFrom", "IsNotDistinctFrom", "Equals", etc.
	Subquery                        *ScalarSubquery
	SubqueryComparisonPredicateType string // "Any", "All"
}

func (s *SubqueryComparisonPredicate) node()              {}
func (s *SubqueryComparisonPredicate) booleanExpression() {}
