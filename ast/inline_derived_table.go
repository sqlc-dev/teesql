package ast

// InlineDerivedTable represents a VALUES clause used as a table reference
// Example: (VALUES ('a'), ('b')) AS x(col)
type InlineDerivedTable struct {
	RowValues []*RowValue   `json:"RowValues,omitempty"`
	Alias     *Identifier   `json:"Alias,omitempty"`
	Columns   []*Identifier `json:"Columns,omitempty"`
	ForPath   bool          `json:"ForPath"`
}

func (t *InlineDerivedTable) node()           {}
func (t *InlineDerivedTable) tableReference() {}
