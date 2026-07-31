package ast

// BeginEndBlockStatement represents a BEGIN...END block.
type BeginEndBlockStatement struct {
	Fragment
	StatementList *StatementList `json:"StatementList,omitempty"`
}

func (b *BeginEndBlockStatement) node()      {}
func (b *BeginEndBlockStatement) statement() {}

// BeginEndAtomicBlockStatement represents a BEGIN ATOMIC...END block (for Hekaton/In-Memory OLTP).
type BeginEndAtomicBlockStatement struct {
	Fragment
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
	Fragment
	OptionKind string
	Value      *Identifier
}

func (o *IdentifierAtomicBlockOption) atomicBlockOption() {}

// LiteralAtomicBlockOption represents an atomic block option with a literal value.
type LiteralAtomicBlockOption struct {
	Fragment
	OptionKind string
	Value      ScalarExpression
}

func (o *LiteralAtomicBlockOption) atomicBlockOption() {}

// OnOffAtomicBlockOption represents an atomic block option with an ON/OFF value.
type OnOffAtomicBlockOption struct {
	Fragment
	OptionKind  string
	OptionState string // "On" or "Off"
}

func (o *OnOffAtomicBlockOption) atomicBlockOption() {}

// StatementList is a list of statements.
type StatementList struct {
	Fragment
	Statements []Statement `json:"Statements,omitempty"`
}
