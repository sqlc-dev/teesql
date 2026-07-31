package ast

// ForClause is an interface for different types of FOR clauses.
type ForClause interface {
	Node
	forClause()
}

// BrowseForClause represents a FOR BROWSE clause.
type BrowseForClause struct {
	Fragment
}

func (*BrowseForClause) node()      {}
func (*BrowseForClause) forClause() {}

// ReadOnlyForClause represents a FOR READ ONLY clause.
type ReadOnlyForClause struct {
	Fragment
}

func (*ReadOnlyForClause) node()      {}
func (*ReadOnlyForClause) forClause() {}

// UpdateForClause represents a FOR UPDATE [OF columns] clause.
type UpdateForClause struct {
	Fragment
	Columns []*ColumnReferenceExpression `json:"Columns,omitempty"`
}

func (*UpdateForClause) node()      {}
func (*UpdateForClause) forClause() {}

// XmlForClause represents a FOR XML clause with its options.
type XmlForClause struct {
	Fragment
	Options []*XmlForClauseOption `json:"Options,omitempty"`
}

func (*XmlForClause) node()      {}
func (*XmlForClause) forClause() {}

// XmlForClauseOption represents an option in a FOR XML clause.
type XmlForClauseOption struct {
	Fragment
	OptionKind string         `json:"OptionKind,omitempty"`
	Value      *StringLiteral `json:"Value,omitempty"`
}

func (*XmlForClauseOption) node() {}

// JsonForClause represents a FOR JSON clause with its options.
type JsonForClause struct {
	Fragment
	Options []*JsonForClauseOption `json:"Options,omitempty"`
}

func (*JsonForClause) node()      {}
func (*JsonForClause) forClause() {}

// JsonForClauseOption represents an option in a FOR JSON clause.
type JsonForClauseOption struct {
	Fragment
	OptionKind string         `json:"OptionKind,omitempty"`
	Value      *StringLiteral `json:"Value,omitempty"`
}

func (*JsonForClauseOption) node() {}
