package ast

// TableSampleClause represents a TABLESAMPLE clause in a table reference
type TableSampleClause struct {
	Fragment
	System                  bool             `json:"System"`
	SampleNumber            ScalarExpression `json:"SampleNumber,omitempty"`
	TableSampleClauseOption string           `json:"TableSampleClauseOption"` // "NotSpecified", "Percent", "Rows"
	RepeatSeed              ScalarExpression `json:"RepeatSeed,omitempty"`
}

func (*TableSampleClause) node() {}
