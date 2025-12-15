package ast

// SearchedCaseExpression represents a CASE WHEN ... THEN ... ELSE ... END expression.
type SearchedCaseExpression struct {
	WhenClauses    []*SearchedWhenClause
	ElseExpression ScalarExpression
}

func (s *SearchedCaseExpression) node()             {}
func (s *SearchedCaseExpression) scalarExpression() {}

// SearchedWhenClause represents a WHEN ... THEN clause.
type SearchedWhenClause struct {
	WhenExpression BooleanExpression
	ThenExpression ScalarExpression
}

// SimpleCaseExpression represents a CASE expression WHEN value THEN result END.
type SimpleCaseExpression struct {
	InputExpression ScalarExpression
	WhenClauses     []*SimpleWhenClause
	ElseExpression  ScalarExpression
}

func (s *SimpleCaseExpression) node()             {}
func (s *SimpleCaseExpression) scalarExpression() {}

// SimpleWhenClause represents a WHEN value THEN result clause.
type SimpleWhenClause struct {
	WhenExpression ScalarExpression
	ThenExpression ScalarExpression
}
