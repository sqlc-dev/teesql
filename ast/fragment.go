package ast

// Fragment holds source position information for an AST node, mirroring
// the positional properties exposed by ScriptDom's TSqlFragment.
//
// StartOffset and FragmentLength are measured in UTF-16 code units (to match
// ScriptDom, whose offsets are .NET string indices). StartLine and
// StartColumn are 1-based; StartColumn also counts UTF-16 code units from
// the beginning of the line.
//
// A zero Fragment (StartLine == 0) means the node has no source tokens
// (for example, synthesized nodes such as empty identifiers in "db..table"),
// matching ScriptDom fragments whose StartOffset is -1.
type Fragment struct {
	StartOffset    int
	FragmentLength int
	StartLine      int
	StartColumn    int
}

// SetSpan records the source span of a node.
func (f *Fragment) SetSpan(startOffset, fragmentLength, startLine, startColumn int) {
	f.StartOffset = startOffset
	f.FragmentLength = fragmentLength
	f.StartLine = startLine
	f.StartColumn = startColumn
}

// Frag returns the fragment itself. It exists so that any node embedding
// Fragment satisfies lightweight accessor interfaces in other packages.
func (f *Fragment) Frag() *Fragment {
	return f
}

// HasSpan reports whether position information has been recorded.
func (f *Fragment) HasSpan() bool {
	return f.StartLine > 0
}

// EndOffset returns the offset one past the last UTF-16 code unit of the span.
func (f *Fragment) EndOffset() int {
	return f.StartOffset + f.FragmentLength
}
