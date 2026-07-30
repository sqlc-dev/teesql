package ast

// AlterIndexStatement represents ALTER INDEX statement
type AlterIndexStatement struct {
	Fragment
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
	Fragment
	Name           *Identifier
	Path           *StringLiteral
	XQueryDataType *StringLiteral
	SQLDataType    *SqlDataTypeReference
	MaxLength      *IntegerLiteral
	IsSingleton    bool
}

func (s *SelectiveXmlIndexPromotedPath) node() {}

// CreateSelectiveXmlIndexStatement represents CREATE SELECTIVE XML INDEX statement
type CreateSelectiveXmlIndexStatement struct {
	Fragment
	Name              *Identifier
	OnName            *SchemaObjectName
	XmlColumn         *Identifier
	IsSecondary       bool
	UsingXmlIndexName *Identifier     // For secondary indexes
	PathName          *Identifier     // For secondary indexes
	PromotedPaths     []*SelectiveXmlIndexPromotedPath
	XmlNamespaces     *XmlNamespaces
	IndexOptions      []IndexOption
}

func (s *CreateSelectiveXmlIndexStatement) statement() {}
func (s *CreateSelectiveXmlIndexStatement) node()      {}

// XmlNamespaces represents a WITH XMLNAMESPACES clause
type XmlNamespaces struct {
	Fragment
	XmlNamespacesElements []XmlNamespacesElement
}

func (x *XmlNamespaces) node() {}

// XmlNamespacesElement is an interface for XML namespace elements
type XmlNamespacesElement interface {
	xmlNamespacesElement()
}

// XmlNamespacesAliasElement represents an alias element in XMLNAMESPACES
type XmlNamespacesAliasElement struct {
	Fragment
	Identifier *Identifier
	String     *StringLiteral
}

func (x *XmlNamespacesAliasElement) node()                 {}
func (x *XmlNamespacesAliasElement) xmlNamespacesElement() {}

// XmlNamespacesDefaultElement represents a default element in XMLNAMESPACES
type XmlNamespacesDefaultElement struct {
	Fragment
	String *StringLiteral
}

func (x *XmlNamespacesDefaultElement) node()                 {}
func (x *XmlNamespacesDefaultElement) xmlNamespacesElement() {}

// PartitionSpecifier represents a partition specifier
type PartitionSpecifier struct {
	Fragment
	All     bool
	Number  ScalarExpression
	Numbers []ScalarExpression
}

func (p *PartitionSpecifier) node() {}
