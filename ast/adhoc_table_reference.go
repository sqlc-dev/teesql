package ast

// AdHocTableReference represents a table accessed via OPENDATASOURCE
// Syntax: OPENDATASOURCE('provider', 'connstr').'object'
// Uses AdHocDataSource from execute_statement.go
type AdHocTableReference struct {
	DataSource *AdHocDataSource                   `json:"DataSource,omitempty"`
	Object     *SchemaObjectNameOrValueExpression `json:"Object,omitempty"`
	Alias      *Identifier                        `json:"Alias,omitempty"`
	ForPath    bool                               `json:"ForPath"`
}

func (*AdHocTableReference) node()           {}
func (*AdHocTableReference) tableReference() {}

// SchemaObjectNameOrValueExpression represents either a schema object name or a value expression
type SchemaObjectNameOrValueExpression struct {
	SchemaObjectName *SchemaObjectName `json:"SchemaObjectName,omitempty"`
	ValueExpression  ScalarExpression  `json:"ValueExpression,omitempty"`
}
