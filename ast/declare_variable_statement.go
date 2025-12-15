package ast

// DeclareVariableStatement represents a DECLARE statement.
type DeclareVariableStatement struct {
	Declarations []*DeclareVariableElement `json:"Declarations,omitempty"`
}

func (d *DeclareVariableStatement) node()      {}
func (d *DeclareVariableStatement) statement() {}

// DeclareVariableElement represents a single variable declaration.
type DeclareVariableElement struct {
	VariableName *Identifier                   `json:"VariableName,omitempty"`
	DataType     *SqlDataTypeReference         `json:"DataType,omitempty"`
	Value        ScalarExpression              `json:"Value,omitempty"`
	Nullable     *NullableConstraintDefinition `json:"Nullable,omitempty"`
}

// SqlDataTypeReference represents a SQL data type.
type SqlDataTypeReference struct {
	SqlDataTypeOption string            `json:"SqlDataTypeOption,omitempty"`
	Parameters        []ScalarExpression `json:"Parameters,omitempty"`
	Name              *SchemaObjectName `json:"Name,omitempty"`
}

func (s *SqlDataTypeReference) node()              {}
func (s *SqlDataTypeReference) dataTypeReference() {}

// XmlDataTypeReference represents an XML data type with optional schema collection
type XmlDataTypeReference struct {
	XmlDataTypeOption   string            `json:"XmlDataTypeOption,omitempty"`
	XmlSchemaCollection *SchemaObjectName `json:"XmlSchemaCollection,omitempty"`
	Name                *SchemaObjectName `json:"Name,omitempty"`
}

func (x *XmlDataTypeReference) node()              {}
func (x *XmlDataTypeReference) dataTypeReference() {}
