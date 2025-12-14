package ast

// InternalOpenRowset represents an OPENROWSET table reference.
type InternalOpenRowset struct {
	Identifier *Identifier        `json:"Identifier,omitempty"`
	VarArgs    []ScalarExpression `json:"VarArgs,omitempty"`
	ForPath    bool               `json:"ForPath"`
}

func (i *InternalOpenRowset) node()           {}
func (i *InternalOpenRowset) tableReference() {}
