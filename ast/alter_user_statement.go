package ast

// AlterUserStatement represents an ALTER USER statement.
type AlterUserStatement struct {
	Name    *Identifier  `json:"Name,omitempty"`
	Options []UserOption `json:"Options,omitempty"`
}

func (s *AlterUserStatement) node()      {}
func (s *AlterUserStatement) statement() {}
