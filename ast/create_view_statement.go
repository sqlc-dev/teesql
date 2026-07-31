package ast

// CreateViewStatement represents a CREATE VIEW statement.
type CreateViewStatement struct {
	Fragment
	SchemaObjectName *SchemaObjectName `json:"SchemaObjectName,omitempty"`
	Columns          []*Identifier     `json:"Columns,omitempty"`
	SelectStatement  *SelectStatement  `json:"SelectStatement,omitempty"`
	WithCheckOption  bool              `json:"WithCheckOption"`
	ViewOptions      []ViewOption      `json:"ViewOptions,omitempty"`
	IsMaterialized   bool              `json:"IsMaterialized"`
}

func (c *CreateViewStatement) node()      {}
func (c *CreateViewStatement) statement() {}

// CreateOrAlterViewStatement represents a CREATE OR ALTER VIEW statement.
type CreateOrAlterViewStatement struct {
	Fragment
	SchemaObjectName *SchemaObjectName `json:"SchemaObjectName,omitempty"`
	Columns          []*Identifier     `json:"Columns,omitempty"`
	SelectStatement  *SelectStatement  `json:"SelectStatement,omitempty"`
	WithCheckOption  bool              `json:"WithCheckOption"`
	ViewOptions      []ViewOption      `json:"ViewOptions,omitempty"`
	IsMaterialized   bool              `json:"IsMaterialized"`
}

func (c *CreateOrAlterViewStatement) node()      {}
func (c *CreateOrAlterViewStatement) statement() {}

// AlterViewStatement represents an ALTER VIEW statement.
type AlterViewStatement struct {
	Fragment
	SchemaObjectName *SchemaObjectName `json:"SchemaObjectName,omitempty"`
	Columns          []*Identifier     `json:"Columns,omitempty"`
	SelectStatement  *SelectStatement  `json:"SelectStatement,omitempty"`
	WithCheckOption  bool              `json:"WithCheckOption"`
	ViewOptions      []ViewOption      `json:"ViewOptions,omitempty"`
	IsMaterialized   bool              `json:"IsMaterialized"`
	IsRebuild        bool              `json:"IsRebuild"`
	IsDisable        bool              `json:"IsDisable"`
}

func (a *AlterViewStatement) node()      {}
func (a *AlterViewStatement) statement() {}

// ViewOption is an interface for different view option types.
type ViewOption interface {
	viewOption()
}

// ViewStatementOption represents a simple view option like SCHEMABINDING.
type ViewStatementOption struct {
	Fragment
	OptionKind string `json:"OptionKind,omitempty"`
}

func (v *ViewStatementOption) viewOption() {}

// ViewDistributionPolicy is an interface for distribution policy types
type ViewDistributionPolicy interface {
	distributionPolicy()
}

// ViewDistributionOption represents a DISTRIBUTION option for materialized views.
type ViewDistributionOption struct {
	Fragment
	OptionKind string                 `json:"OptionKind,omitempty"`
	Value      ViewDistributionPolicy `json:"Value,omitempty"`
}

func (v *ViewDistributionOption) viewOption() {}

// ViewHashDistributionPolicy represents the hash distribution policy for materialized views.
type ViewHashDistributionPolicy struct {
	Fragment
	DistributionColumn  *Identifier   `json:"DistributionColumn,omitempty"`
	DistributionColumns []*Identifier `json:"DistributionColumns,omitempty"`
}

func (v *ViewHashDistributionPolicy) distributionPolicy() {}

// ViewRoundRobinDistributionPolicy represents the round robin distribution policy for materialized views.
type ViewRoundRobinDistributionPolicy struct {
	Fragment
}

func (v *ViewRoundRobinDistributionPolicy) distributionPolicy() {}

// ViewForAppendOption represents the FOR_APPEND option for materialized views.
type ViewForAppendOption struct {
	Fragment
	OptionKind string `json:"OptionKind,omitempty"`
}

func (v *ViewForAppendOption) viewOption() {}
