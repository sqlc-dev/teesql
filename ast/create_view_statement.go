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

// CreateOrAlterViewStatement represents a CREATE OR ALTER VIEW statement.
type CreateOrAlterViewStatement struct {
	SchemaObjectName  *SchemaObjectName    `json:"SchemaObjectName,omitempty"`
	Columns           []*Identifier        `json:"Columns,omitempty"`
	SelectStatement   *SelectStatement     `json:"SelectStatement,omitempty"`
	WithCheckOption   bool                 `json:"WithCheckOption"`
	ViewOptions       []ViewOption         `json:"ViewOptions,omitempty"`
	IsMaterialized    bool                 `json:"IsMaterialized"`
}

func (c *CreateOrAlterViewStatement) node()      {}
func (c *CreateOrAlterViewStatement) statement() {}

// ViewOption is an interface for different view option types.
type ViewOption interface {
	viewOption()
}

// ViewStatementOption represents a simple view option like SCHEMABINDING.
type ViewStatementOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
}

func (v *ViewStatementOption) viewOption() {}

// ViewDistributionOption represents a DISTRIBUTION option for materialized views.
type ViewDistributionOption struct {
	OptionKind string                        `json:"OptionKind,omitempty"`
	Value      *ViewHashDistributionPolicy `json:"Value,omitempty"`
}

func (v *ViewDistributionOption) viewOption() {}

// ViewHashDistributionPolicy represents the hash distribution policy for materialized views.
type ViewHashDistributionPolicy struct {
	DistributionColumn  *Identifier   `json:"DistributionColumn,omitempty"`
	DistributionColumns []*Identifier `json:"DistributionColumns,omitempty"`
}

// ViewForAppendOption represents the FOR_APPEND option for materialized views.
type ViewForAppendOption struct {
	OptionKind string `json:"OptionKind,omitempty"`
}

func (v *ViewForAppendOption) viewOption() {}
