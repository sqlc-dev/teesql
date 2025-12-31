package ast

// SetOffsetsStatement represents SET OFFSETS statement
type SetOffsetsStatement struct {
	Options string
	IsOn    bool
}

func (s *SetOffsetsStatement) node()      {}
func (s *SetOffsetsStatement) statement() {}
