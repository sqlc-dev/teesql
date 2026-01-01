package ast

// ForClause is an interface for different types of FOR clauses.
type ForClause interface {
	Node
	forClause()
}

// BrowseForClause represents a FOR BROWSE clause.
type BrowseForClause struct{}

func (*BrowseForClause) node()      {}
func (*BrowseForClause) forClause() {}

// ReadOnlyForClause represents a FOR READ ONLY clause.
type ReadOnlyForClause struct{}

func (*ReadOnlyForClause) node()      {}
func (*ReadOnlyForClause) forClause() {}

// UpdateForClause represents a FOR UPDATE [OF columns] clause.
type UpdateForClause struct {
	Columns []*ColumnReferenceExpression `json:"Columns,omitempty"`
}

func (*UpdateForClause) node()      {}
func (*UpdateForClause) forClause() {}

// XmlForClause represents a FOR XML clause with its options.
type XmlForClause struct {
	Options []*XmlForClauseOption `json:"Options,omitempty"`
}

func (*XmlForClause) node()      {}
func (*XmlForClause) forClause() {}

// XmlForClauseOption represents an option in a FOR XML clause.
type XmlForClauseOption struct {
	OptionKind string           `json:"OptionKind,omitempty"`
	Value      *StringLiteral   `json:"Value,omitempty"`
}

func (*XmlForClauseOption) node() {}
