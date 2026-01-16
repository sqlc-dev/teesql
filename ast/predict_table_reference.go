package ast

// PredictTableReference represents PREDICT(...) in a FROM clause
type PredictTableReference struct {
	ModelVariable          ScalarExpression         `json:"ModelVariable,omitempty"`
	DataSource             *NamedTableReference     `json:"DataSource,omitempty"`
	RunTime                *Identifier              `json:"RunTime,omitempty"`
	SchemaDeclarationItems []*SchemaDeclarationItem `json:"SchemaDeclarationItems,omitempty"`
	Alias                  *Identifier              `json:"Alias,omitempty"`
	ForPath                bool                     `json:"ForPath,omitempty"`
}

func (*PredictTableReference) node()           {}
func (*PredictTableReference) tableReference() {}

// SchemaDeclarationItem represents a column definition in PREDICT/OPENXML WITH clause
type SchemaDeclarationItem struct {
	ColumnDefinition *ColumnDefinitionBase `json:"ColumnDefinition,omitempty"`
	Mapping          ScalarExpression      `json:"Mapping,omitempty"` // Optional XPath mapping for OPENXML
}

func (*SchemaDeclarationItem) node() {}
