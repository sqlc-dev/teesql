package ast

// ExistsPredicate represents EXISTS (subquery)
type ExistsPredicate struct {
	Subquery QueryExpression `json:"Subquery,omitempty"`
}

func (*ExistsPredicate) node()              {}
func (*ExistsPredicate) booleanExpression() {}
