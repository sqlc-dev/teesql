package ast

// SelectStarExpression represents SELECT *.
type SelectStarExpression struct {
	Fragment
	Qualifier *MultiPartIdentifier `json:"Qualifier,omitempty"`
}

func (*SelectStarExpression) node()          {}
func (*SelectStarExpression) selectElement() {}
