package ast

// BeginEndBlockStatement represents a BEGIN...END block.
type BeginEndBlockStatement struct {
	StatementList *StatementList `json:"StatementList,omitempty"`
}

func (b *BeginEndBlockStatement) node()      {}
func (b *BeginEndBlockStatement) statement() {}

// BeginEndAtomicBlockStatement represents a BEGIN ATOMIC...END block (for Hekaton/In-Memory OLTP).
type BeginEndAtomicBlockStatement struct {
	Options       []AtomicBlockOption
	StatementList *StatementList
}

func (b *BeginEndAtomicBlockStatement) node()      {}
func (b *BeginEndAtomicBlockStatement) statement() {}

// AtomicBlockOption is an interface for atomic block options.
type AtomicBlockOption interface {
	atomicBlockOption()
}

// IdentifierAtomicBlockOption represents an atomic block option with an identifier value.
type IdentifierAtomicBlockOption struct {
	OptionKind string
	Value      *Identifier
}

func (o *IdentifierAtomicBlockOption) atomicBlockOption() {}

// LiteralAtomicBlockOption represents an atomic block option with a literal value.
type LiteralAtomicBlockOption struct {
	OptionKind string
	Value      ScalarExpression
}

func (o *LiteralAtomicBlockOption) atomicBlockOption() {}

// StatementList is a list of statements.
type StatementList struct {
	Statements []Statement `json:"Statements,omitempty"`
}
