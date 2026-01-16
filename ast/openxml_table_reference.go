package ast

// OpenXmlTableReference represents an OPENXML table-valued function
// Syntax: OPENXML(variable, rowpattern [, flags]) [WITH (schema) | WITH table_name | AS alias]
type OpenXmlTableReference struct {
	Variable               ScalarExpression         `json:"Variable,omitempty"`
	RowPattern             ScalarExpression         `json:"RowPattern,omitempty"`
	Flags                  ScalarExpression         `json:"Flags,omitempty"`
	SchemaDeclarationItems []*SchemaDeclarationItem `json:"SchemaDeclarationItems,omitempty"`
	TableName              *SchemaObjectName        `json:"TableName,omitempty"`
	Alias                  *Identifier              `json:"Alias,omitempty"`
	ForPath                bool                     `json:"ForPath"`
}

func (*OpenXmlTableReference) node()           {}
func (*OpenXmlTableReference) tableReference() {}
