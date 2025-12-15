package ast

// SetStatisticsStatement represents SET STATISTICS IO/PROFILE/TIME/XML statements
type SetStatisticsStatement struct {
	Options SetOptions
	IsOn    bool
}

func (s *SetStatisticsStatement) node()      {}
func (s *SetStatisticsStatement) statement() {}
