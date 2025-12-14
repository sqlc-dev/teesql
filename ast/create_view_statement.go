package ast

// CreateViewStatement represents a CREATE VIEW statement.
type CreateViewStatement struct {
	SchemaObjectName  *SchemaObjectName    `json:"SchemaObjectName,omitempty"`
	Columns           []*Identifier        `json:"Columns,omitempty"`
	SelectStatement   *SelectStatement     `json:"SelectStatement,omitempty"`
	WithCheckOption   bool                 `json:"WithCheckOption"`
	ViewOptions       []ViewOption         `json:"ViewOptions,omitempty"`
	IsMaterialized    bool                 `json:"IsMaterialized"`
}

func (c *CreateViewStatement) node()      {}
func (c *CreateViewStatement) statement() {}

// ViewOption represents a view option like SCHEMABINDING.
type ViewOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
}
