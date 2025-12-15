package ast

// OptimizerHintBase is the interface for optimizer hints.
type OptimizerHintBase interface {
	Node
	optimizerHint()
}
