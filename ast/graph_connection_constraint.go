package ast

// GraphConnectionConstraintDefinition represents a CONNECTION constraint for graph edge tables
type GraphConnectionConstraintDefinition struct {
	ConstraintIdentifier *Identifier
	FromNodeToNodeList   []*GraphConnectionBetweenNodes
	DeleteAction         string // "NotSpecified", "Cascade", "NoAction", etc.
}

func (g *GraphConnectionConstraintDefinition) node()            {}
func (g *GraphConnectionConstraintDefinition) tableConstraint() {}

// GraphConnectionBetweenNodes represents a FROM node TO node specification in a CONNECTION constraint
type GraphConnectionBetweenNodes struct {
	FromNode *SchemaObjectName
	ToNode   *SchemaObjectName
}

func (g *GraphConnectionBetweenNodes) node() {}
