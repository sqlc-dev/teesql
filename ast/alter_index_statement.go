package ast

// AlterIndexStatement represents ALTER INDEX statement
type AlterIndexStatement struct {
	Name           *Identifier
	All            bool
	OnName         *SchemaObjectName
	AlterIndexType string // "Rebuild", "Reorganize", "Disable", "Set", "UpdateSelectiveXmlPaths", etc.
	Partition      *PartitionSpecifier
	IndexOptions   []IndexOption
	PromotedPaths  []*SelectiveXmlIndexPromotedPath
	XmlNamespaces  *XmlNamespaces
}

func (s *AlterIndexStatement) statement() {}
func (s *AlterIndexStatement) node()      {}

// SelectiveXmlIndexPromotedPath represents a path in a selective XML index
type SelectiveXmlIndexPromotedPath struct {
	Name           *Identifier
	Path           *StringLiteral
	XQueryDataType *StringLiteral
	MaxLength      *IntegerLiteral
	IsSingleton    bool
}

func (s *SelectiveXmlIndexPromotedPath) node() {}

// XmlNamespaces represents a WITH XMLNAMESPACES clause
type XmlNamespaces struct {
	XmlNamespacesElements []*XmlNamespacesAliasElement
}

func (x *XmlNamespaces) node() {}

// XmlNamespacesAliasElement represents an alias element in XMLNAMESPACES
type XmlNamespacesAliasElement struct {
	Identifier *Identifier
	String     *StringLiteral
}

func (x *XmlNamespacesAliasElement) node() {}

// PartitionSpecifier represents a partition specifier
type PartitionSpecifier struct {
	All     bool
	Number  ScalarExpression
	Numbers []ScalarExpression
}

func (p *PartitionSpecifier) node() {}
