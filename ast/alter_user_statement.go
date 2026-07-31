package ast

// AlterUserStatement represents an ALTER USER statement.
type AlterUserStatement struct {
	Fragment
	Name        *Identifier  `json:"Name,omitempty"`
	UserOptions []UserOption `json:"UserOptions,omitempty"`
}

func (s *AlterUserStatement) node()      {}
func (s *AlterUserStatement) statement() {}
