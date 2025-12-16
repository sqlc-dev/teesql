package ast

// DropDatabaseStatement represents a DROP DATABASE statement
type DropDatabaseStatement struct {
	IsIfExists bool
	Databases  []*Identifier
}

func (s *DropDatabaseStatement) statement() {}
func (s *DropDatabaseStatement) node()      {}

// DropTableStatement represents a DROP TABLE statement
type DropTableStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropTableStatement) statement() {}
func (s *DropTableStatement) node()      {}

// DropViewStatement represents a DROP VIEW statement
type DropViewStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropViewStatement) statement() {}
func (s *DropViewStatement) node()      {}

// DropProcedureStatement represents a DROP PROCEDURE statement
type DropProcedureStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropProcedureStatement) statement() {}
func (s *DropProcedureStatement) node()      {}

// DropFunctionStatement represents a DROP FUNCTION statement
type DropFunctionStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropFunctionStatement) statement() {}
func (s *DropFunctionStatement) node()      {}

// DropTriggerStatement represents a DROP TRIGGER statement
type DropTriggerStatement struct {
	IsIfExists   bool
	Objects      []*SchemaObjectName
	TriggerScope string // "Normal", "Database", "AllServer"
}

func (s *DropTriggerStatement) statement() {}
func (s *DropTriggerStatement) node()      {}

// DropIndexStatement represents a DROP INDEX statement
type DropIndexStatement struct {
	IsIfExists bool
	Indexes    []*DropIndexClause
}

func (s *DropIndexStatement) statement() {}
func (s *DropIndexStatement) node()      {}

// DropIndexClause represents a single index to drop
type DropIndexClause struct {
	Index *SchemaObjectName
	// Could have additional options like ON table, WITH options
}

// DropStatisticsStatement represents a DROP STATISTICS statement
type DropStatisticsStatement struct {
	Objects []*SchemaObjectName
}

func (s *DropStatisticsStatement) statement() {}
func (s *DropStatisticsStatement) node()      {}

// DropDefaultStatement represents a DROP DEFAULT statement
type DropDefaultStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropDefaultStatement) statement() {}
func (s *DropDefaultStatement) node()      {}

// DropRuleStatement represents a DROP RULE statement
type DropRuleStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropRuleStatement) statement() {}
func (s *DropRuleStatement) node()      {}

// DropSchemaStatement represents a DROP SCHEMA statement
type DropSchemaStatement struct {
	IsIfExists bool
	Schema     *SchemaObjectName
}

func (s *DropSchemaStatement) statement() {}
func (s *DropSchemaStatement) node()      {}

// DropSecurityPolicyStatement represents a DROP SECURITY POLICY statement
type DropSecurityPolicyStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropSecurityPolicyStatement) statement() {}
func (s *DropSecurityPolicyStatement) node()      {}

// DropExternalDataSourceStatement represents a DROP EXTERNAL DATA SOURCE statement
type DropExternalDataSourceStatement struct {
	IsIfExists bool
	Name       *Identifier
}

func (s *DropExternalDataSourceStatement) statement() {}
func (s *DropExternalDataSourceStatement) node()      {}

// DropExternalFileFormatStatement represents a DROP EXTERNAL FILE FORMAT statement
type DropExternalFileFormatStatement struct {
	IsIfExists bool
	Name       *Identifier
}

func (s *DropExternalFileFormatStatement) statement() {}
func (s *DropExternalFileFormatStatement) node()      {}

// DropExternalTableStatement represents a DROP EXTERNAL TABLE statement
type DropExternalTableStatement struct {
	IsIfExists bool
	Objects    []*SchemaObjectName
}

func (s *DropExternalTableStatement) statement() {}
func (s *DropExternalTableStatement) node()      {}

// DropExternalResourcePoolStatement represents a DROP EXTERNAL RESOURCE POOL statement
type DropExternalResourcePoolStatement struct {
	IsIfExists bool
	Name       *Identifier
}

func (s *DropExternalResourcePoolStatement) statement() {}
func (s *DropExternalResourcePoolStatement) node()      {}

// DropWorkloadGroupStatement represents a DROP WORKLOAD GROUP statement
type DropWorkloadGroupStatement struct {
	IsIfExists bool
	Name       *Identifier
}

func (s *DropWorkloadGroupStatement) statement() {}
func (s *DropWorkloadGroupStatement) node()      {}

// DropWorkloadClassifierStatement represents a DROP WORKLOAD CLASSIFIER statement
type DropWorkloadClassifierStatement struct {
	IsIfExists bool
	Name       *Identifier
}

func (s *DropWorkloadClassifierStatement) statement() {}
func (s *DropWorkloadClassifierStatement) node()      {}
