package ast

// CreateSchemaStatement represents a CREATE SCHEMA statement.
type CreateSchemaStatement struct {
	Name          *Identifier   `json:"Name,omitempty"`
	Owner         *Identifier   `json:"Owner,omitempty"`
	StatementList *StatementList `json:"StatementList,omitempty"`
}

func (c *CreateSchemaStatement) node()      {}
func (c *CreateSchemaStatement) statement() {}
