package ast

// OpenJsonTableReference represents an OPENJSON table reference in the FROM clause.
type OpenJsonTableReference struct {
	Variable               ScalarExpression             `json:"Variable,omitempty"`
	RowPattern             ScalarExpression             `json:"RowPattern,omitempty"`
	SchemaDeclarationItems []*SchemaDeclarationItemOpenjson `json:"SchemaDeclarationItems,omitempty"`
	Alias                  *Identifier                  `json:"Alias,omitempty"`
	ForPath                bool                         `json:"ForPath,omitempty"`
}

func (*OpenJsonTableReference) node()           {}
func (*OpenJsonTableReference) tableReference() {}

// SchemaDeclarationItemOpenjson represents a column definition in OPENJSON WITH clause.
type SchemaDeclarationItemOpenjson struct {
	AsJson           bool                  `json:"AsJson,omitempty"`
	ColumnDefinition *ColumnDefinitionBase `json:"ColumnDefinition,omitempty"`
	Mapping          ScalarExpression      `json:"Mapping,omitempty"`
}

func (*SchemaDeclarationItemOpenjson) node() {}
