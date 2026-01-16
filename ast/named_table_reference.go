package ast

// NamedTableReference represents a named table reference.
type NamedTableReference struct {
	SchemaObject      *SchemaObjectName  `json:"SchemaObject,omitempty"`
	TableSampleClause *TableSampleClause `json:"TableSampleClause,omitempty"`
	TemporalClause    *TemporalClause    `json:"TemporalClause,omitempty"`
	Alias             *Identifier        `json:"Alias,omitempty"`
	TableHints        []TableHintType    `json:"TableHints,omitempty"`
	ForPath           bool               `json:"ForPath,omitempty"`
}

func (*NamedTableReference) node()           {}
func (*NamedTableReference) tableReference() {}

// TemporalClause represents a FOR SYSTEM_TIME clause for temporal tables.
type TemporalClause struct {
	TemporalClauseType string           `json:"TemporalClauseType,omitempty"`
	StartTime          ScalarExpression `json:"StartTime,omitempty"`
	EndTime            ScalarExpression `json:"EndTime,omitempty"`
}

func (*TemporalClause) node() {}
