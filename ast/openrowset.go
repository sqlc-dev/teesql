package ast

// OpenRowsetCosmos represents an OPENROWSET with PROVIDER = ..., CONNECTION = ..., OBJECT = ... syntax.
type OpenRowsetCosmos struct {
	Options     []OpenRowsetCosmosOption     `json:"Options,omitempty"`
	WithColumns []*OpenRowsetColumnDefinition `json:"WithColumns,omitempty"`
	Alias       *Identifier                  `json:"Alias,omitempty"`
	ForPath     bool                         `json:"ForPath"`
}

func (o *OpenRowsetCosmos) node()           {}
func (o *OpenRowsetCosmos) tableReference() {}

// OpenRowsetCosmosOption is the interface for OpenRowset Cosmos options.
type OpenRowsetCosmosOption interface {
	openRowsetCosmosOption()
}

// LiteralOpenRowsetCosmosOption represents an option with a literal value.
type LiteralOpenRowsetCosmosOption struct {
	Value      ScalarExpression `json:"Value,omitempty"`
	OptionKind string           `json:"OptionKind,omitempty"`
}

func (l *LiteralOpenRowsetCosmosOption) openRowsetCosmosOption() {}

// OpenRowsetTableReference represents OPENROWSET with various syntaxes:
// - OPENROWSET('provider', 'connstr', object)
// - OPENROWSET('provider', 'server'; 'user'; 'password', 'query')
type OpenRowsetTableReference struct {
	ProviderName   ScalarExpression              `json:"ProviderName,omitempty"`
	ProviderString ScalarExpression              `json:"ProviderString,omitempty"`
	DataSource     ScalarExpression              `json:"DataSource,omitempty"`
	UserId         ScalarExpression              `json:"UserId,omitempty"`
	Password       ScalarExpression              `json:"Password,omitempty"`
	Query          ScalarExpression              `json:"Query,omitempty"`
	Object         *SchemaObjectName             `json:"Object,omitempty"`
	WithColumns    []*OpenRowsetColumnDefinition `json:"WithColumns,omitempty"`
	Alias          *Identifier                   `json:"Alias,omitempty"`
	ForPath        bool                          `json:"ForPath"`
}

func (o *OpenRowsetTableReference) node()           {}
func (o *OpenRowsetTableReference) tableReference() {}

// OpenRowsetColumnDefinition represents a column definition in WITH clause.
type OpenRowsetColumnDefinition struct {
	ColumnOrdinal    ScalarExpression  `json:"ColumnOrdinal,omitempty"`
	JsonPath         ScalarExpression  `json:"JsonPath,omitempty"`
	ColumnIdentifier *Identifier       `json:"ColumnIdentifier,omitempty"`
	DataType         DataTypeReference `json:"DataType,omitempty"`
	Collation        *Identifier       `json:"Collation,omitempty"`
}
