package ast

// CreateFullTextStopListStatement represents CREATE FULLTEXT STOPLIST statement
type CreateFullTextStopListStatement struct {
	Name               *Identifier `json:"Name,omitempty"`
	IsSystemStopList   bool        `json:"IsSystemStopList"`
	DatabaseName       *Identifier `json:"DatabaseName,omitempty"`
	SourceStopListName *Identifier `json:"SourceStopListName,omitempty"`
	Owner              *Identifier `json:"Owner,omitempty"`
}

func (s *CreateFullTextStopListStatement) node()      {}
func (s *CreateFullTextStopListStatement) statement() {}

// AlterFullTextStopListStatement represents ALTER FULLTEXT STOPLIST statement
type AlterFullTextStopListStatement struct {
	Name   *Identifier           `json:"Name,omitempty"`
	Action *FullTextStopListAction `json:"Action,omitempty"`
}

func (s *AlterFullTextStopListStatement) node()      {}
func (s *AlterFullTextStopListStatement) statement() {}

// FullTextStopListAction represents an action in ALTER FULLTEXT STOPLIST
type FullTextStopListAction struct {
	IsAdd        bool                         `json:"IsAdd"`
	IsAll        bool                         `json:"IsAll"`
	StopWord     *StringLiteral               `json:"StopWord,omitempty"`
	LanguageTerm *IdentifierOrValueExpression `json:"LanguageTerm,omitempty"`
}

func (a *FullTextStopListAction) node() {}

// DropFullTextStopListStatement represents DROP FULLTEXT STOPLIST statement
type DropFullTextStopListStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (s *DropFullTextStopListStatement) node()      {}
func (s *DropFullTextStopListStatement) statement() {}

// DropFullTextCatalogStatement represents DROP FULLTEXT CATALOG statement
type DropFullTextCatalogStatement struct {
	Name       *Identifier `json:"Name,omitempty"`
	IsIfExists bool        `json:"IsIfExists"`
}

func (s *DropFullTextCatalogStatement) node()      {}
func (s *DropFullTextCatalogStatement) statement() {}

// DropFulltextIndexStatement represents DROP FULLTEXT INDEX statement
type DropFulltextIndexStatement struct {
	TableName *SchemaObjectName `json:"TableName,omitempty"`
}

func (s *DropFulltextIndexStatement) node()      {}
func (s *DropFulltextIndexStatement) statement() {}
