package ast

// PivotedTableReference represents a table with PIVOT
type PivotedTableReference struct {
	TableReference              TableReference
	InColumns                   []*Identifier
	PivotColumn                 *ColumnReferenceExpression
	ValueColumns                []*ColumnReferenceExpression
	AggregateFunctionIdentifier *MultiPartIdentifier
	Alias                       *Identifier
	ForPath                     bool
}

func (p *PivotedTableReference) node()           {}
func (p *PivotedTableReference) tableReference() {}

// UnpivotedTableReference represents a table with UNPIVOT
type UnpivotedTableReference struct {
	TableReference       TableReference
	InColumns            []*ColumnReferenceExpression
	PivotColumn          *Identifier
	ValueColumn          *Identifier
	NullHandling         string // "None", "ExcludeNulls", "IncludeNulls"
	Alias                *Identifier
	ForPath              bool
}

func (u *UnpivotedTableReference) node()           {}
func (u *UnpivotedTableReference) tableReference() {}
