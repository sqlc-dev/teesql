package ast

// ExistsPredicate represents EXISTS (subquery)
type ExistsPredicate struct {
	Fragment
	Subquery QueryExpression `json:"Subquery,omitempty"`
}

func (*ExistsPredicate) node()              {}
func (*ExistsPredicate) booleanExpression() {}
