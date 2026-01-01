package ast

// MergeStatement represents a MERGE statement
type MergeStatement struct {
	MergeSpecification *MergeSpecification
}

func (s *MergeStatement) node()      {}
func (s *MergeStatement) statement() {}

// MergeSpecification represents the specification of a MERGE statement
type MergeSpecification struct {
	Target          TableReference    // The target table
	TableReference  TableReference    // The USING clause table reference
	SearchCondition BooleanExpression // The ON clause condition (may be GraphMatchPredicate)
	ActionClauses   []*MergeActionClause
	OutputClause    *OutputClause
}

func (s *MergeSpecification) node() {}

// MergeActionClause represents a WHEN clause in a MERGE statement
type MergeActionClause struct {
	Condition       string      // "Matched", "NotMatched", "NotMatchedBySource", "NotMatchedByTarget"
	SearchCondition BooleanExpression
	Action          MergeAction
}

func (c *MergeActionClause) node() {}

// MergeAction is an interface for merge actions
type MergeAction interface {
	Node
	mergeAction()
}

// DeleteMergeAction represents DELETE in a MERGE WHEN clause
type DeleteMergeAction struct{}

func (a *DeleteMergeAction) node()        {}
func (a *DeleteMergeAction) mergeAction() {}

// UpdateMergeAction represents UPDATE SET in a MERGE WHEN clause
type UpdateMergeAction struct {
	SetClauses []SetClause
}

func (a *UpdateMergeAction) node()        {}
func (a *UpdateMergeAction) mergeAction() {}

// InsertMergeAction represents INSERT in a MERGE WHEN clause
type InsertMergeAction struct {
	Columns []*ColumnReferenceExpression
	Values  []ScalarExpression
}

func (a *InsertMergeAction) node()        {}
func (a *InsertMergeAction) mergeAction() {}

// JoinParenthesisTableReference represents a parenthesized join table reference
type JoinParenthesisTableReference struct {
	Join TableReference // The join inside the parenthesis
}

func (j *JoinParenthesisTableReference) node()           {}
func (j *JoinParenthesisTableReference) tableReference() {}

// GraphMatchPredicate represents MATCH predicate in graph queries
type GraphMatchPredicate struct {
	Expression GraphMatchExpression
}

func (g *GraphMatchPredicate) node()              {}
func (g *GraphMatchPredicate) booleanExpression() {}

// GraphMatchExpression is an interface for graph match expressions
type GraphMatchExpression interface {
	Node
	graphMatchExpression()
}

// GraphMatchCompositeExpression represents a graph pattern like (Node1-(Edge)->Node2)
type GraphMatchCompositeExpression struct {
	LeftNode     *GraphMatchNodeExpression
	Edge         *Identifier
	RightNode    *GraphMatchNodeExpression
	ArrowOnRight bool // true if arrow is -> (left to right), false if <- (right to left)
}

func (g *GraphMatchCompositeExpression) node()                 {}
func (g *GraphMatchCompositeExpression) graphMatchExpression() {}

// GraphMatchNodeExpression represents a node in a graph match pattern
type GraphMatchNodeExpression struct {
	Node         *Identifier
	UsesLastNode bool
}

func (g *GraphMatchNodeExpression) node()                 {}
func (g *GraphMatchNodeExpression) graphMatchExpression() {}
