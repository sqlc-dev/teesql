package ast

// ChangeTableChangesTableReference represents CHANGETABLE(CHANGES ...) table reference
type ChangeTableChangesTableReference struct {
	Target       *SchemaObjectName  `json:"Target,omitempty"`
	SinceVersion ScalarExpression   `json:"SinceVersion,omitempty"`
	ForceSeek    bool               `json:"ForceSeek"`
	Columns      []*Identifier      `json:"Columns,omitempty"`
	Alias        *Identifier        `json:"Alias,omitempty"`
	ForPath      bool               `json:"ForPath"`
}

func (c *ChangeTableChangesTableReference) node()           {}
func (c *ChangeTableChangesTableReference) tableReference() {}

// ChangeTableVersionTableReference represents CHANGETABLE(VERSION ...) table reference
type ChangeTableVersionTableReference struct {
	Target            *SchemaObjectName  `json:"Target,omitempty"`
	PrimaryKeyColumns []*Identifier      `json:"PrimaryKeyColumns,omitempty"`
	PrimaryKeyValues  []ScalarExpression `json:"PrimaryKeyValues,omitempty"`
	ForceSeek         bool               `json:"ForceSeek"`
	Columns           []*Identifier      `json:"Columns,omitempty"`
	Alias             *Identifier        `json:"Alias,omitempty"`
	ForPath           bool               `json:"ForPath"`
}

func (c *ChangeTableVersionTableReference) node()           {}
func (c *ChangeTableVersionTableReference) tableReference() {}
