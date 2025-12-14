package ast

// LabelStatement represents a label definition (e.g., "start:").
type LabelStatement struct {
	Value string `json:"Value"`
}

func (l *LabelStatement) node()      {}
func (l *LabelStatement) statement() {}
