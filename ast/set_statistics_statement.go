package ast

// SetStatisticsStatement represents SET STATISTICS IO/PROFILE/TIME/XML statements
// Options can contain multiple comma-separated values like "IO, Profile, Time"
type SetStatisticsStatement struct {
	Options string
	IsOn    bool
}

func (s *SetStatisticsStatement) node()      {}
func (s *SetStatisticsStatement) statement() {}
