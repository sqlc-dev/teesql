// Package parser provides T-SQL parsing functionality.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyleconroy/teesql/ast"
)

// jsonNode represents a generic JSON node from the AST JSON format.
type jsonNode map[string]any

// MarshalScript marshals a Script to JSON in the expected format.
func MarshalScript(s *ast.Script) ([]byte, error) {
	node := scriptToJSON(s)
	return json.MarshalIndent(node, "", "  ")
}

func scriptToJSON(s *ast.Script) jsonNode {
	node := jsonNode{
		"$type": "TSqlScript",
	}
	if len(s.Batches) > 0 {
		batches := make([]jsonNode, len(s.Batches))
		for i, b := range s.Batches {
			batches[i] = batchToJSON(b)
		}
		node["Batches"] = batches
	}
	return node
}

func batchToJSON(b *ast.Batch) jsonNode {
	node := jsonNode{
		"$type": "TSqlBatch",
	}
	if len(b.Statements) > 0 {
		stmts := make([]jsonNode, len(b.Statements))
		for i, stmt := range b.Statements {
			stmts[i] = statementToJSON(stmt)
		}
		node["Statements"] = stmts
	}
	return node
}

func statementToJSON(stmt ast.Statement) jsonNode {
	switch s := stmt.(type) {
	case *ast.SelectStatement:
		return selectStatementToJSON(s)
	case *ast.InsertStatement:
		return insertStatementToJSON(s)
	case *ast.UpdateStatement:
		return updateStatementToJSON(s)
	case *ast.DeleteStatement:
		return deleteStatementToJSON(s)
	case *ast.DeclareVariableStatement:
		return declareVariableStatementToJSON(s)
	case *ast.SetVariableStatement:
		return setVariableStatementToJSON(s)
	case *ast.IfStatement:
		return ifStatementToJSON(s)
	case *ast.WhileStatement:
		return whileStatementToJSON(s)
	case *ast.BeginEndBlockStatement:
		return beginEndBlockStatementToJSON(s)
	case *ast.CreateViewStatement:
		return createViewStatementToJSON(s)
	case *ast.CreateSchemaStatement:
		return createSchemaStatementToJSON(s)
	case *ast.CreateProcedureStatement:
		return createProcedureStatementToJSON(s)
	case *ast.CreateRoleStatement:
		return createRoleStatementToJSON(s)
	case *ast.ExecuteStatement:
		return executeStatementToJSON(s)
	case *ast.ExecuteAsStatement:
		return executeAsStatementToJSON(s)
	case *ast.ReturnStatement:
		return returnStatementToJSON(s)
	case *ast.BreakStatement:
		return breakStatementToJSON()
	case *ast.ContinueStatement:
		return continueStatementToJSON()
	case *ast.PrintStatement:
		return printStatementToJSON(s)
	case *ast.ThrowStatement:
		return throwStatementToJSON(s)
	case *ast.AlterTableDropTableElementStatement:
		return alterTableDropTableElementStatementToJSON(s)
	case *ast.AlterTableAlterIndexStatement:
		return alterTableAlterIndexStatementToJSON(s)
	case *ast.AlterTableAddTableElementStatement:
		return alterTableAddTableElementStatementToJSON(s)
	case *ast.AlterTableAlterColumnStatement:
		return alterTableAlterColumnStatementToJSON(s)
	case *ast.AlterMessageTypeStatement:
		return alterMessageTypeStatementToJSON(s)
	case *ast.CreateContractStatement:
		return createContractStatementToJSON(s)
	case *ast.CreatePartitionSchemeStatement:
		return createPartitionSchemeStatementToJSON(s)
	case *ast.CreateRuleStatement:
		return createRuleStatementToJSON(s)
	case *ast.CreateSynonymStatement:
		return createSynonymStatementToJSON(s)
	case *ast.AlterCredentialStatement:
		return alterCredentialStatementToJSON(s)
	case *ast.AlterDatabaseSetStatement:
		return alterDatabaseSetStatementToJSON(s)
	case *ast.AlterDatabaseAddFileStatement:
		return alterDatabaseAddFileStatementToJSON(s)
	case *ast.AlterDatabaseAddFileGroupStatement:
		return alterDatabaseAddFileGroupStatementToJSON(s)
	case *ast.AlterDatabaseModifyFileStatement:
		return alterDatabaseModifyFileStatementToJSON(s)
	case *ast.AlterDatabaseModifyFileGroupStatement:
		return alterDatabaseModifyFileGroupStatementToJSON(s)
	case *ast.AlterDatabaseModifyNameStatement:
		return alterDatabaseModifyNameStatementToJSON(s)
	case *ast.AlterDatabaseRemoveFileStatement:
		return alterDatabaseRemoveFileStatementToJSON(s)
	case *ast.AlterDatabaseRemoveFileGroupStatement:
		return alterDatabaseRemoveFileGroupStatementToJSON(s)
	case *ast.RevertStatement:
		return revertStatementToJSON(s)
	case *ast.DropCredentialStatement:
		return dropCredentialStatementToJSON(s)
	case *ast.DropExternalLanguageStatement:
		return dropExternalLanguageStatementToJSON(s)
	case *ast.DropExternalLibraryStatement:
		return dropExternalLibraryStatementToJSON(s)
	case *ast.DropSequenceStatement:
		return dropSequenceStatementToJSON(s)
	case *ast.DropSearchPropertyListStatement:
		return dropSearchPropertyListStatementToJSON(s)
	case *ast.DropServerRoleStatement:
		return dropServerRoleStatementToJSON(s)
	case *ast.DropAvailabilityGroupStatement:
		return dropAvailabilityGroupStatementToJSON(s)
	case *ast.DropFederationStatement:
		return dropFederationStatementToJSON(s)
	case *ast.DropSecurityPolicyStatement:
		return dropSecurityPolicyStatementToJSON(s)
	case *ast.DropExternalDataSourceStatement:
		return dropExternalDataSourceStatementToJSON(s)
	case *ast.DropExternalFileFormatStatement:
		return dropExternalFileFormatStatementToJSON(s)
	case *ast.DropExternalTableStatement:
		return dropExternalTableStatementToJSON(s)
	case *ast.DropExternalResourcePoolStatement:
		return dropExternalResourcePoolStatementToJSON(s)
	case *ast.DropWorkloadGroupStatement:
		return dropWorkloadGroupStatementToJSON(s)
	case *ast.DropWorkloadClassifierStatement:
		return dropWorkloadClassifierStatementToJSON(s)
	case *ast.DropTypeStatement:
		return dropTypeStatementToJSON(s)
	case *ast.DropAggregateStatement:
		return dropAggregateStatementToJSON(s)
	case *ast.DropSynonymStatement:
		return dropSynonymStatementToJSON(s)
	case *ast.DropUserStatement:
		return dropUserStatementToJSON(s)
	case *ast.DropRoleStatement:
		return dropRoleStatementToJSON(s)
	case *ast.DropAssemblyStatement:
		return dropAssemblyStatementToJSON(s)
	case *ast.CreateTableStatement:
		return createTableStatementToJSON(s)
	case *ast.GrantStatement:
		return grantStatementToJSON(s)
	case *ast.PredicateSetStatement:
		return predicateSetStatementToJSON(s)
	case *ast.SetStatisticsStatement:
		return setStatisticsStatementToJSON(s)
	case *ast.CommitTransactionStatement:
		return commitTransactionStatementToJSON(s)
	case *ast.RollbackTransactionStatement:
		return rollbackTransactionStatementToJSON(s)
	case *ast.SaveTransactionStatement:
		return saveTransactionStatementToJSON(s)
	case *ast.BeginTransactionStatement:
		return beginTransactionStatementToJSON(s)
	case *ast.WaitForStatement:
		return waitForStatementToJSON(s)
	case *ast.MoveConversationStatement:
		return moveConversationStatementToJSON(s)
	case *ast.GetConversationGroupStatement:
		return getConversationGroupStatementToJSON(s)
	case *ast.TruncateTableStatement:
		return truncateTableStatementToJSON(s)
	case *ast.UseStatement:
		return useStatementToJSON(s)
	case *ast.KillStatement:
		return killStatementToJSON(s)
	case *ast.KillStatsJobStatement:
		return killStatsJobStatementToJSON(s)
	case *ast.KillQueryNotificationSubscriptionStatement:
		return killQueryNotificationSubscriptionStatementToJSON(s)
	case *ast.CloseSymmetricKeyStatement:
		return closeSymmetricKeyStatementToJSON(s)
	case *ast.CloseMasterKeyStatement:
		return closeMasterKeyStatementToJSON(s)
	case *ast.OpenMasterKeyStatement:
		return openMasterKeyStatementToJSON(s)
	case *ast.OpenSymmetricKeyStatement:
		return openSymmetricKeyStatementToJSON(s)
	case *ast.CheckpointStatement:
		return checkpointStatementToJSON(s)
	case *ast.ReconfigureStatement:
		return reconfigureStatementToJSON(s)
	case *ast.ShutdownStatement:
		return shutdownStatementToJSON(s)
	case *ast.SetUserStatement:
		return setUserStatementToJSON(s)
	case *ast.LineNoStatement:
		return lineNoStatementToJSON(s)
	case *ast.RaiseErrorStatement:
		return raiseErrorStatementToJSON(s)
	case *ast.ReadTextStatement:
		return readTextStatementToJSON(s)
	case *ast.WriteTextStatement:
		return writeTextStatementToJSON(s)
	case *ast.UpdateTextStatement:
		return updateTextStatementToJSON(s)
	case *ast.GoToStatement:
		return goToStatementToJSON(s)
	case *ast.LabelStatement:
		return labelStatementToJSON(s)
	case *ast.CreateDefaultStatement:
		return createDefaultStatementToJSON(s)
	case *ast.CreateMasterKeyStatement:
		return createMasterKeyStatementToJSON(s)
	case *ast.AlterMasterKeyStatement:
		return alterMasterKeyStatementToJSON(s)
	case *ast.AlterSchemaStatement:
		return alterSchemaStatementToJSON(s)
	case *ast.AlterRoleStatement:
		return alterRoleStatementToJSON(s)
	case *ast.AlterRemoteServiceBindingStatement:
		return alterRemoteServiceBindingStatementToJSON(s)
	case *ast.AlterXmlSchemaCollectionStatement:
		return alterXmlSchemaCollectionStatementToJSON(s)
	case *ast.AlterServerConfigurationSetSoftNumaStatement:
		return alterServerConfigurationSetSoftNumaStatementToJSON(s)
	case *ast.AlterLoginAddDropCredentialStatement:
		return alterLoginAddDropCredentialStatementToJSON(s)
	case *ast.TryCatchStatement:
		return tryCatchStatementToJSON(s)
	case *ast.SendStatement:
		return sendStatementToJSON(s)
	case *ast.ReceiveStatement:
		return receiveStatementToJSON(s)
	case *ast.CreateCredentialStatement:
		return createCredentialStatementToJSON(s)
	case *ast.CreateXmlSchemaCollectionStatement:
		return createXmlSchemaCollectionStatementToJSON(s)
	case *ast.CreateSearchPropertyListStatement:
		return createSearchPropertyListStatementToJSON(s)
	case *ast.CreateExternalDataSourceStatement:
		return createExternalDataSourceStatementToJSON(s)
	case *ast.CreateExternalFileFormatStatement:
		return createExternalFileFormatStatementToJSON(s)
	case *ast.CreateExternalTableStatement:
		return createExternalTableStatementToJSON(s)
	case *ast.CreateExternalLanguageStatement:
		return createExternalLanguageStatementToJSON(s)
	case *ast.CreateExternalLibraryStatement:
		return createExternalLibraryStatementToJSON(s)
	case *ast.CreateEventSessionStatement:
		return createEventSessionStatementToJSON(s)
	case *ast.RestoreStatement:
		return restoreStatementToJSON(s)
	case *ast.BackupDatabaseStatement:
		return backupDatabaseStatementToJSON(s)
	case *ast.BackupCertificateStatement:
		return backupCertificateStatementToJSON(s)
	case *ast.CreateUserStatement:
		return createUserStatementToJSON(s)
	case *ast.CreateAggregateStatement:
		return createAggregateStatementToJSON(s)
	case *ast.CreateColumnStoreIndexStatement:
		return createColumnStoreIndexStatementToJSON(s)
	case *ast.AlterFunctionStatement:
		return alterFunctionStatementToJSON(s)
	case *ast.CreateFunctionStatement:
		return createFunctionStatementToJSON(s)
	case *ast.AlterTriggerStatement:
		return alterTriggerStatementToJSON(s)
	case *ast.CreateTriggerStatement:
		return createTriggerStatementToJSON(s)
	case *ast.CreateDatabaseStatement:
		return createDatabaseStatementToJSON(s)
	case *ast.CreateLoginStatement:
		return createLoginStatementToJSON(s)
	case *ast.CreateIndexStatement:
		return createIndexStatementToJSON(s)
	case *ast.CreateAsymmetricKeyStatement:
		return createAsymmetricKeyStatementToJSON(s)
	case *ast.CreateSymmetricKeyStatement:
		return createSymmetricKeyStatementToJSON(s)
	case *ast.CreateCertificateStatement:
		return createCertificateStatementToJSON(s)
	case *ast.CreateMessageTypeStatement:
		return createMessageTypeStatementToJSON(s)
	case *ast.CreateServiceStatement:
		return createServiceStatementToJSON(s)
	case *ast.CreateQueueStatement:
		return createQueueStatementToJSON(s)
	case *ast.CreateRouteStatement:
		return createRouteStatementToJSON(s)
	case *ast.CreateEndpointStatement:
		return createEndpointStatementToJSON(s)
	case *ast.CreateAssemblyStatement:
		return createAssemblyStatementToJSON(s)
	case *ast.CreateApplicationRoleStatement:
		return createApplicationRoleStatementToJSON(s)
	case *ast.CreateFulltextCatalogStatement:
		return createFulltextCatalogStatementToJSON(s)
	case *ast.CreateFulltextIndexStatement:
		return createFulltextIndexStatementToJSON(s)
	case *ast.CreateRemoteServiceBindingStatement:
		return createRemoteServiceBindingStatementToJSON(s)
	case *ast.CreateStatisticsStatement:
		return createStatisticsStatementToJSON(s)
	case *ast.CreateTypeStatement:
		return createTypeStatementToJSON(s)
	case *ast.CreateXmlIndexStatement:
		return createXmlIndexStatementToJSON(s)
	case *ast.CreatePartitionFunctionStatement:
		return createPartitionFunctionStatementToJSON(s)
	case *ast.CreateEventNotificationStatement:
		return createEventNotificationStatementToJSON(s)
	case *ast.AlterIndexStatement:
		return alterIndexStatementToJSON(s)
	case *ast.DropDatabaseStatement:
		return dropDatabaseStatementToJSON(s)
	case *ast.DropTableStatement:
		return dropTableStatementToJSON(s)
	case *ast.DropViewStatement:
		return dropViewStatementToJSON(s)
	case *ast.DropProcedureStatement:
		return dropProcedureStatementToJSON(s)
	case *ast.DropFunctionStatement:
		return dropFunctionStatementToJSON(s)
	case *ast.DropTriggerStatement:
		return dropTriggerStatementToJSON(s)
	case *ast.DropIndexStatement:
		return dropIndexStatementToJSON(s)
	case *ast.DropStatisticsStatement:
		return dropStatisticsStatementToJSON(s)
	case *ast.DropDefaultStatement:
		return dropDefaultStatementToJSON(s)
	case *ast.DropRuleStatement:
		return dropRuleStatementToJSON(s)
	case *ast.DropSchemaStatement:
		return dropSchemaStatementToJSON(s)
	case *ast.AlterTableTriggerModificationStatement:
		return alterTableTriggerModificationStatementToJSON(s)
	case *ast.AlterTableSwitchStatement:
		return alterTableSwitchStatementToJSON(s)
	case *ast.AlterTableConstraintModificationStatement:
		return alterTableConstraintModificationStatementToJSON(s)
	case *ast.InsertBulkStatement:
		return insertBulkStatementToJSON(s)
	case *ast.BulkInsertStatement:
		return bulkInsertStatementToJSON(s)
	case *ast.AlterUserStatement:
		return alterUserStatementToJSON(s)
	case *ast.AlterRouteStatement:
		return alterRouteStatementToJSON(s)
	case *ast.AlterAssemblyStatement:
		return alterAssemblyStatementToJSON(s)
	case *ast.AlterEndpointStatement:
		return alterEndpointStatementToJSON(s)
	case *ast.AlterServiceStatement:
		return alterServiceStatementToJSON(s)
	case *ast.AlterCertificateStatement:
		return alterCertificateStatementToJSON(s)
	case *ast.AlterApplicationRoleStatement:
		return alterApplicationRoleStatementToJSON(s)
	case *ast.AlterAsymmetricKeyStatement:
		return alterAsymmetricKeyStatementToJSON(s)
	case *ast.AlterQueueStatement:
		return alterQueueStatementToJSON(s)
	case *ast.AlterPartitionSchemeStatement:
		return alterPartitionSchemeStatementToJSON(s)
	case *ast.AlterPartitionFunctionStatement:
		return alterPartitionFunctionStatementToJSON(s)
	case *ast.AlterFulltextCatalogStatement:
		return alterFulltextCatalogStatementToJSON(s)
	case *ast.AlterFulltextIndexStatement:
		return alterFulltextIndexStatementToJSON(s)
	case *ast.AlterSymmetricKeyStatement:
		return alterSymmetricKeyStatementToJSON(s)
	case *ast.AlterServiceMasterKeyStatement:
		return alterServiceMasterKeyStatementToJSON(s)
	default:
		return jsonNode{"$type": "UnknownStatement"}
	}
}

func revertStatementToJSON(s *ast.RevertStatement) jsonNode {
	node := jsonNode{
		"$type": "RevertStatement",
	}
	if s.Cookie != nil {
		node["Cookie"] = scalarExpressionToJSON(s.Cookie)
	}
	return node
}

func dropCredentialStatementToJSON(s *ast.DropCredentialStatement) jsonNode {
	node := jsonNode{
		"$type": "DropCredentialStatement",
	}
	node["IsDatabaseScoped"] = s.IsDatabaseScoped
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return node
}

func dropExternalLanguageStatementToJSON(s *ast.DropExternalLanguageStatement) jsonNode {
	node := jsonNode{
		"$type": "DropExternalLanguageStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Authorization != nil {
		node["Owner"] = identifierToJSON(s.Authorization)
	}
	return node
}

func dropExternalLibraryStatementToJSON(s *ast.DropExternalLibraryStatement) jsonNode {
	node := jsonNode{
		"$type": "DropExternalLibraryStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	return node
}

func dropSequenceStatementToJSON(s *ast.DropSequenceStatement) jsonNode {
	node := jsonNode{
		"$type": "DropSequenceStatement",
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	node["IsIfExists"] = s.IsIfExists
	return node
}

func dropSearchPropertyListStatementToJSON(s *ast.DropSearchPropertyListStatement) jsonNode {
	node := jsonNode{
		"$type": "DropSearchPropertyListStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return node
}

func dropServerRoleStatementToJSON(s *ast.DropServerRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "DropServerRoleStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return node
}

func dropAvailabilityGroupStatementToJSON(s *ast.DropAvailabilityGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "DropAvailabilityGroupStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return node
}

func dropFederationStatementToJSON(s *ast.DropFederationStatement) jsonNode {
	node := jsonNode{
		"$type": "DropFederationStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return node
}

func alterTableDropTableElementStatementToJSON(s *ast.AlterTableDropTableElementStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterTableDropTableElementStatement",
	}
	if len(s.AlterTableDropTableElements) > 0 {
		elements := make([]jsonNode, len(s.AlterTableDropTableElements))
		for i, e := range s.AlterTableDropTableElements {
			elements[i] = alterTableDropTableElementToJSON(e)
		}
		node["AlterTableDropTableElements"] = elements
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return node
}

func alterTableDropTableElementToJSON(e *ast.AlterTableDropTableElement) jsonNode {
	node := jsonNode{
		"$type": "AlterTableDropTableElement",
	}
	if e.TableElementType != "" {
		node["TableElementType"] = e.TableElementType
	}
	if e.Name != nil {
		node["Name"] = identifierToJSON(e.Name)
	}
	node["IsIfExists"] = e.IsIfExists
	return node
}

func alterTableAlterIndexStatementToJSON(s *ast.AlterTableAlterIndexStatement) jsonNode {
	node := jsonNode{
		"$type":          "AlterTableAlterIndexStatement",
		"AlterIndexType": s.AlterIndexType,
	}
	if s.IndexIdentifier != nil {
		node["IndexIdentifier"] = identifierToJSON(s.IndexIdentifier)
	}
	if len(s.IndexOptions) > 0 {
		options := make([]jsonNode, len(s.IndexOptions))
		for i, o := range s.IndexOptions {
			options[i] = indexExpressionOptionToJSON(o)
		}
		node["IndexOptions"] = options
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return node
}

func indexExpressionOptionToJSON(o *ast.IndexExpressionOption) jsonNode {
	node := jsonNode{
		"$type":      "IndexExpressionOption",
		"OptionKind": o.OptionKind,
	}
	if o.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(o.Expression)
	}
	return node
}

func alterTableAddTableElementStatementToJSON(s *ast.AlterTableAddTableElementStatement) jsonNode {
	node := jsonNode{
		"$type":                        "AlterTableAddTableElementStatement",
		"ExistingRowsCheckEnforcement": s.ExistingRowsCheckEnforcement,
	}
	if s.Definition != nil {
		node["Definition"] = tableDefinitionToJSON(s.Definition)
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return node
}

func alterTableAlterColumnStatementToJSON(s *ast.AlterTableAlterColumnStatement) jsonNode {
	node := jsonNode{
		"$type":                       "AlterTableAlterColumnStatement",
		"AlterTableAlterColumnOption": s.AlterTableAlterColumnOption,
		"IsHidden":                    s.IsHidden,
		"IsMasked":                    s.IsMasked,
	}
	if s.ColumnIdentifier != nil {
		node["ColumnIdentifier"] = identifierToJSON(s.ColumnIdentifier)
	}
	if s.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(s.DataType)
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return node
}

func alterMessageTypeStatementToJSON(s *ast.AlterMessageTypeStatement) jsonNode {
	node := jsonNode{
		"$type":            "AlterMessageTypeStatement",
		"ValidationMethod": s.ValidationMethod,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.XmlSchemaCollectionName != nil {
		node["XmlSchemaCollectionName"] = schemaObjectNameToJSON(s.XmlSchemaCollectionName)
	}
	return node
}

func createContractStatementToJSON(s *ast.CreateContractStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateContractStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Messages) > 0 {
		msgs := make([]jsonNode, len(s.Messages))
		for i, m := range s.Messages {
			msgs[i] = contractMessageToJSON(m)
		}
		node["Messages"] = msgs
	}
	return node
}

func contractMessageToJSON(m *ast.ContractMessage) jsonNode {
	node := jsonNode{
		"$type":  "ContractMessage",
		"SentBy": m.SentBy,
	}
	if m.Name != nil {
		node["Name"] = identifierToJSON(m.Name)
	}
	return node
}

func createPartitionSchemeStatementToJSON(s *ast.CreatePartitionSchemeStatement) jsonNode {
	node := jsonNode{
		"$type": "CreatePartitionSchemeStatement",
		"IsAll": s.IsAll,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.PartitionFunction != nil {
		node["PartitionFunction"] = identifierToJSON(s.PartitionFunction)
	}
	if len(s.FileGroups) > 0 {
		fgs := make([]jsonNode, len(s.FileGroups))
		for i, fg := range s.FileGroups {
			fgs[i] = identifierOrValueExpressionToJSON(fg)
		}
		node["FileGroups"] = fgs
	}
	return node
}

func createRuleStatementToJSON(s *ast.CreateRuleStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateRuleStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.Expression != nil {
		node["Expression"] = booleanExpressionToJSON(s.Expression)
	}
	return node
}

func createSynonymStatementToJSON(s *ast.CreateSynonymStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSynonymStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.ForName != nil {
		node["ForName"] = schemaObjectNameToJSON(s.ForName)
	}
	return node
}

func alterCredentialStatementToJSON(s *ast.AlterCredentialStatement) jsonNode {
	node := jsonNode{
		"$type":            "AlterCredentialStatement",
		"IsDatabaseScoped": s.IsDatabaseScoped,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Identity != nil {
		node["Identity"] = scalarExpressionToJSON(s.Identity)
	}
	if s.Secret != nil {
		node["Secret"] = scalarExpressionToJSON(s.Secret)
	}
	return node
}

func alterDatabaseSetStatementToJSON(s *ast.AlterDatabaseSetStatement) jsonNode {
	node := jsonNode{
		"$type":             "AlterDatabaseSetStatement",
		"WithManualCutover": s.WithManualCutover,
		"UseCurrent":        s.UseCurrent,
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = databaseOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return node
}

func databaseOptionToJSON(opt ast.DatabaseOption) jsonNode {
	switch o := opt.(type) {
	case *ast.AcceleratedDatabaseRecoveryDatabaseOption:
		return jsonNode{
			"$type":       "AcceleratedDatabaseRecoveryDatabaseOption",
			"OptionKind":  o.OptionKind,
			"OptionState": o.OptionState,
		}
	case *ast.OnOffDatabaseOption:
		return jsonNode{
			"$type":       "OnOffDatabaseOption",
			"OptionKind":  o.OptionKind,
			"OptionState": o.OptionState,
		}
	default:
		return jsonNode{"$type": "UnknownDatabaseOption"}
	}
}

func indexDefinitionToJSON(idx *ast.IndexDefinition) jsonNode {
	node := jsonNode{
		"$type":  "IndexDefinition",
		"Unique": idx.Unique,
	}
	if idx.Name != nil {
		node["Name"] = identifierToJSON(idx.Name)
	}
	if idx.IndexType != nil {
		node["IndexType"] = indexTypeToJSON(idx.IndexType)
	}
	if len(idx.IndexOptions) > 0 {
		options := make([]jsonNode, len(idx.IndexOptions))
		for i, o := range idx.IndexOptions {
			options[i] = indexExpressionOptionToJSON(o)
		}
		node["IndexOptions"] = options
	}
	if len(idx.Columns) > 0 {
		cols := make([]jsonNode, len(idx.Columns))
		for i, c := range idx.Columns {
			cols[i] = columnWithSortOrderToJSON(c)
		}
		node["Columns"] = cols
	}
	return node
}

func indexTypeToJSON(t *ast.IndexType) jsonNode {
	return jsonNode{
		"$type":         "IndexType",
		"IndexTypeKind": t.IndexTypeKind,
	}
}

func columnWithSortOrderToJSON(c *ast.ColumnWithSortOrder) jsonNode {
	node := jsonNode{
		"$type": "ColumnWithSortOrder",
	}
	if c.Column != nil {
		node["Column"] = scalarExpressionToJSON(c.Column)
	}
	sortOrder := "NotSpecified"
	switch c.SortOrder {
	case ast.SortOrderAscending:
		sortOrder = "Ascending"
	case ast.SortOrderDescending:
		sortOrder = "Descending"
	}
	node["SortOrder"] = sortOrder
	return node
}

func printStatementToJSON(s *ast.PrintStatement) jsonNode {
	node := jsonNode{
		"$type": "PrintStatement",
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	return node
}

func throwStatementToJSON(s *ast.ThrowStatement) jsonNode {
	node := jsonNode{
		"$type": "ThrowStatement",
	}
	if s.ErrorNumber != nil {
		node["ErrorNumber"] = scalarExpressionToJSON(s.ErrorNumber)
	}
	if s.Message != nil {
		node["Message"] = scalarExpressionToJSON(s.Message)
	}
	if s.State != nil {
		node["State"] = scalarExpressionToJSON(s.State)
	}
	return node
}

func selectStatementToJSON(s *ast.SelectStatement) jsonNode {
	node := jsonNode{
		"$type": "SelectStatement",
	}
	if s.QueryExpression != nil {
		node["QueryExpression"] = queryExpressionToJSON(s.QueryExpression)
	}
	if s.Into != nil {
		node["Into"] = schemaObjectNameToJSON(s.Into)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return node
}

func optimizerHintToJSON(h ast.OptimizerHintBase) jsonNode {
	switch hint := h.(type) {
	case *ast.OptimizerHint:
		node := jsonNode{
			"$type": "OptimizerHint",
		}
		if hint.HintKind != "" {
			node["HintKind"] = hint.HintKind
		}
		return node
	case *ast.LiteralOptimizerHint:
		node := jsonNode{
			"$type": "LiteralOptimizerHint",
		}
		if hint.Value != nil {
			node["Value"] = scalarExpressionToJSON(hint.Value)
		}
		if hint.HintKind != "" {
			node["HintKind"] = hint.HintKind
		}
		return node
	default:
		return jsonNode{"$type": "UnknownOptimizerHint"}
	}
}

func queryExpressionToJSON(qe ast.QueryExpression) jsonNode {
	switch q := qe.(type) {
	case *ast.QuerySpecification:
		return querySpecificationToJSON(q)
	case *ast.QueryParenthesisExpression:
		return queryParenthesisExpressionToJSON(q)
	case *ast.BinaryQueryExpression:
		return binaryQueryExpressionToJSON(q)
	default:
		return jsonNode{"$type": "UnknownQueryExpression"}
	}
}

func queryParenthesisExpressionToJSON(q *ast.QueryParenthesisExpression) jsonNode {
	node := jsonNode{
		"$type": "QueryParenthesisExpression",
	}
	if q.QueryExpression != nil {
		node["QueryExpression"] = queryExpressionToJSON(q.QueryExpression)
	}
	return node
}

func binaryQueryExpressionToJSON(q *ast.BinaryQueryExpression) jsonNode {
	node := jsonNode{
		"$type": "BinaryQueryExpression",
	}
	if q.BinaryQueryExpressionType != "" {
		node["BinaryQueryExpressionType"] = q.BinaryQueryExpressionType
	}
	node["All"] = q.All
	if q.FirstQueryExpression != nil {
		node["FirstQueryExpression"] = queryExpressionToJSON(q.FirstQueryExpression)
	}
	if q.SecondQueryExpression != nil {
		node["SecondQueryExpression"] = queryExpressionToJSON(q.SecondQueryExpression)
	}
	if q.OrderByClause != nil {
		node["OrderByClause"] = orderByClauseToJSON(q.OrderByClause)
	}
	return node
}

func querySpecificationToJSON(q *ast.QuerySpecification) jsonNode {
	node := jsonNode{
		"$type": "QuerySpecification",
	}
	if q.UniqueRowFilter != "" {
		node["UniqueRowFilter"] = q.UniqueRowFilter
	}
	if q.TopRowFilter != nil {
		node["TopRowFilter"] = topRowFilterToJSON(q.TopRowFilter)
	}
	if len(q.SelectElements) > 0 {
		elems := make([]jsonNode, len(q.SelectElements))
		for i, elem := range q.SelectElements {
			elems[i] = selectElementToJSON(elem)
		}
		node["SelectElements"] = elems
	}
	if q.FromClause != nil {
		node["FromClause"] = fromClauseToJSON(q.FromClause)
	}
	if q.WhereClause != nil {
		node["WhereClause"] = whereClauseToJSON(q.WhereClause)
	}
	if q.GroupByClause != nil {
		node["GroupByClause"] = groupByClauseToJSON(q.GroupByClause)
	}
	if q.HavingClause != nil {
		node["HavingClause"] = havingClauseToJSON(q.HavingClause)
	}
	if q.OrderByClause != nil {
		node["OrderByClause"] = orderByClauseToJSON(q.OrderByClause)
	}
	return node
}

func topRowFilterToJSON(t *ast.TopRowFilter) jsonNode {
	node := jsonNode{
		"$type": "TopRowFilter",
	}
	if t.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(t.Expression)
	}
	node["Percent"] = t.Percent
	node["WithTies"] = t.WithTies
	return node
}

func selectElementToJSON(elem ast.SelectElement) jsonNode {
	switch e := elem.(type) {
	case *ast.SelectScalarExpression:
		node := jsonNode{
			"$type": "SelectScalarExpression",
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		if e.ColumnName != nil {
			node["ColumnName"] = identifierOrValueExpressionToJSON(e.ColumnName)
		}
		return node
	case *ast.SelectStarExpression:
		node := jsonNode{
			"$type": "SelectStarExpression",
		}
		if e.Qualifier != nil {
			node["Qualifier"] = multiPartIdentifierToJSON(e.Qualifier)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownSelectElement"}
	}
}

func scalarExpressionToJSON(expr ast.ScalarExpression) jsonNode {
	switch e := expr.(type) {
	case *ast.ColumnReferenceExpression:
		node := jsonNode{
			"$type": "ColumnReferenceExpression",
		}
		if e.ColumnType != "" {
			node["ColumnType"] = e.ColumnType
		}
		if e.MultiPartIdentifier != nil {
			node["MultiPartIdentifier"] = multiPartIdentifierToJSON(e.MultiPartIdentifier)
		}
		return node
	case *ast.IntegerLiteral:
		node := jsonNode{
			"$type": "IntegerLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return node
	case *ast.StringLiteral:
		node := jsonNode{
			"$type": "StringLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		// Always include IsNational and IsLargeObject
		node["IsNational"] = e.IsNational
		node["IsLargeObject"] = e.IsLargeObject
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return node
	case *ast.BinaryLiteral:
		node := jsonNode{
			"$type": "BinaryLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		node["IsLargeObject"] = e.IsLargeObject
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return node
	case *ast.FunctionCall:
		node := jsonNode{
			"$type": "FunctionCall",
		}
		if e.FunctionName != nil {
			node["FunctionName"] = identifierToJSON(e.FunctionName)
		}
		if len(e.Parameters) > 0 {
			params := make([]jsonNode, len(e.Parameters))
			for i, p := range e.Parameters {
				params[i] = scalarExpressionToJSON(p)
			}
			node["Parameters"] = params
		}
		if e.UniqueRowFilter != "" {
			node["UniqueRowFilter"] = e.UniqueRowFilter
		}
		if e.WithArrayWrapper {
			node["WithArrayWrapper"] = e.WithArrayWrapper
		}
		return node
	case *ast.BinaryExpression:
		node := jsonNode{
			"$type": "BinaryExpression",
		}
		if e.BinaryExpressionType != "" {
			node["BinaryExpressionType"] = e.BinaryExpressionType
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = scalarExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = scalarExpressionToJSON(e.SecondExpression)
		}
		return node
	case *ast.VariableReference:
		node := jsonNode{
			"$type": "VariableReference",
		}
		if e.Name != "" {
			node["Name"] = e.Name
		}
		return node
	case *ast.GlobalVariableExpression:
		node := jsonNode{
			"$type": "GlobalVariableExpression",
		}
		if e.Name != "" {
			node["Name"] = e.Name
		}
		return node
	case *ast.NumericLiteral:
		node := jsonNode{
			"$type": "NumericLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return node
	case *ast.OdbcLiteral:
		node := jsonNode{
			"$type": "OdbcLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.OdbcLiteralType != "" {
			node["OdbcLiteralType"] = e.OdbcLiteralType
		}
		node["IsNational"] = e.IsNational
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return node
	case *ast.NullLiteral:
		node := jsonNode{
			"$type": "NullLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return node
	case *ast.DefaultLiteral:
		node := jsonNode{
			"$type": "DefaultLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return node
	case *ast.UnaryExpression:
		node := jsonNode{
			"$type": "UnaryExpression",
		}
		if e.UnaryExpressionType != "" {
			node["UnaryExpressionType"] = e.UnaryExpressionType
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		return node
	case *ast.ParenthesisExpression:
		node := jsonNode{
			"$type": "ParenthesisExpression",
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		return node
	case *ast.ScalarSubquery:
		node := jsonNode{
			"$type": "ScalarSubquery",
		}
		if e.QueryExpression != nil {
			node["QueryExpression"] = queryExpressionToJSON(e.QueryExpression)
		}
		return node
	case *ast.SearchedCaseExpression:
		node := jsonNode{
			"$type": "SearchedCaseExpression",
		}
		if len(e.WhenClauses) > 0 {
			clauses := make([]jsonNode, len(e.WhenClauses))
			for i, c := range e.WhenClauses {
				clause := jsonNode{
					"$type": "SearchedWhenClause",
				}
				if c.WhenExpression != nil {
					clause["WhenExpression"] = booleanExpressionToJSON(c.WhenExpression)
				}
				if c.ThenExpression != nil {
					clause["ThenExpression"] = scalarExpressionToJSON(c.ThenExpression)
				}
				clauses[i] = clause
			}
			node["WhenClauses"] = clauses
		}
		if e.ElseExpression != nil {
			node["ElseExpression"] = scalarExpressionToJSON(e.ElseExpression)
		}
		return node
	case *ast.SimpleCaseExpression:
		node := jsonNode{
			"$type": "SimpleCaseExpression",
		}
		if e.InputExpression != nil {
			node["InputExpression"] = scalarExpressionToJSON(e.InputExpression)
		}
		if len(e.WhenClauses) > 0 {
			clauses := make([]jsonNode, len(e.WhenClauses))
			for i, c := range e.WhenClauses {
				clause := jsonNode{
					"$type": "SimpleWhenClause",
				}
				if c.WhenExpression != nil {
					clause["WhenExpression"] = scalarExpressionToJSON(c.WhenExpression)
				}
				if c.ThenExpression != nil {
					clause["ThenExpression"] = scalarExpressionToJSON(c.ThenExpression)
				}
				clauses[i] = clause
			}
			node["WhenClauses"] = clauses
		}
		if e.ElseExpression != nil {
			node["ElseExpression"] = scalarExpressionToJSON(e.ElseExpression)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownScalarExpression"}
	}
}

func identifierToJSON(id *ast.Identifier) jsonNode {
	node := jsonNode{
		"$type": "Identifier",
	}
	// Always include Value, even if empty
	node["Value"] = id.Value
	if id.QuoteType != "" {
		node["QuoteType"] = id.QuoteType
	}
	return node
}

func multiPartIdentifierToJSON(mpi *ast.MultiPartIdentifier) jsonNode {
	node := jsonNode{
		"$type": "MultiPartIdentifier",
	}
	if mpi.Count > 0 {
		node["Count"] = mpi.Count
	}
	if len(mpi.Identifiers) > 0 {
		ids := make([]jsonNode, len(mpi.Identifiers))
		for i, id := range mpi.Identifiers {
			ids[i] = identifierToJSON(id)
		}
		node["Identifiers"] = ids
	}
	return node
}

func identifierOrValueExpressionToJSON(iove *ast.IdentifierOrValueExpression) jsonNode {
	node := jsonNode{
		"$type": "IdentifierOrValueExpression",
	}
	if iove.Value != "" {
		node["Value"] = iove.Value
	}
	if iove.Identifier != nil {
		node["Identifier"] = identifierToJSON(iove.Identifier)
	}
	if iove.ValueExpression != nil {
		node["ValueExpression"] = scalarExpressionToJSON(iove.ValueExpression)
	}
	return node
}

func fromClauseToJSON(fc *ast.FromClause) jsonNode {
	node := jsonNode{
		"$type": "FromClause",
	}
	if len(fc.TableReferences) > 0 {
		refs := make([]jsonNode, len(fc.TableReferences))
		for i, ref := range fc.TableReferences {
			refs[i] = tableReferenceToJSON(ref)
		}
		node["TableReferences"] = refs
	}
	return node
}

func tableReferenceToJSON(ref ast.TableReference) jsonNode {
	switch r := ref.(type) {
	case *ast.NamedTableReference:
		node := jsonNode{
			"$type": "NamedTableReference",
		}
		if r.SchemaObject != nil {
			node["SchemaObject"] = schemaObjectNameToJSON(r.SchemaObject)
		}
		if len(r.TableHints) > 0 {
			hints := make([]jsonNode, len(r.TableHints))
			for i, h := range r.TableHints {
				hints[i] = tableHintToJSON(h)
			}
			node["TableHints"] = hints
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return node
	case *ast.QualifiedJoin:
		node := jsonNode{
			"$type": "QualifiedJoin",
		}
		if r.SearchCondition != nil {
			node["SearchCondition"] = booleanExpressionToJSON(r.SearchCondition)
		}
		if r.QualifiedJoinType != "" {
			node["QualifiedJoinType"] = r.QualifiedJoinType
		}
		if r.JoinHint != "" {
			node["JoinHint"] = r.JoinHint
		}
		if r.FirstTableReference != nil {
			node["FirstTableReference"] = tableReferenceToJSON(r.FirstTableReference)
		}
		if r.SecondTableReference != nil {
			node["SecondTableReference"] = tableReferenceToJSON(r.SecondTableReference)
		}
		return node
	case *ast.UnqualifiedJoin:
		node := jsonNode{
			"$type": "UnqualifiedJoin",
		}
		if r.UnqualifiedJoinType != "" {
			node["UnqualifiedJoinType"] = r.UnqualifiedJoinType
		}
		if r.FirstTableReference != nil {
			node["FirstTableReference"] = tableReferenceToJSON(r.FirstTableReference)
		}
		if r.SecondTableReference != nil {
			node["SecondTableReference"] = tableReferenceToJSON(r.SecondTableReference)
		}
		return node
	case *ast.VariableTableReference:
		node := jsonNode{
			"$type": "VariableTableReference",
		}
		if r.Variable != nil {
			node["Variable"] = scalarExpressionToJSON(r.Variable)
		}
		node["ForPath"] = r.ForPath
		return node
	case *ast.SchemaObjectFunctionTableReference:
		node := jsonNode{
			"$type": "SchemaObjectFunctionTableReference",
		}
		if r.SchemaObject != nil {
			node["SchemaObject"] = schemaObjectNameToJSON(r.SchemaObject)
		}
		if len(r.Parameters) > 0 {
			params := make([]jsonNode, len(r.Parameters))
			for i, p := range r.Parameters {
				params[i] = scalarExpressionToJSON(p)
			}
			node["Parameters"] = params
		}
		node["ForPath"] = r.ForPath
		return node
	case *ast.InternalOpenRowset:
		node := jsonNode{
			"$type": "InternalOpenRowset",
		}
		if r.Identifier != nil {
			node["Identifier"] = identifierToJSON(r.Identifier)
		}
		if len(r.VarArgs) > 0 {
			args := make([]jsonNode, len(r.VarArgs))
			for i, a := range r.VarArgs {
				args[i] = scalarExpressionToJSON(a)
			}
			node["VarArgs"] = args
		}
		node["ForPath"] = r.ForPath
		return node
	default:
		return jsonNode{"$type": "UnknownTableReference"}
	}
}

func schemaObjectNameToJSON(son *ast.SchemaObjectName) jsonNode {
	node := jsonNode{
		"$type": "SchemaObjectName",
	}
	if son.ServerIdentifier != nil {
		node["ServerIdentifier"] = identifierToJSON(son.ServerIdentifier)
	}
	if son.DatabaseIdentifier != nil {
		node["DatabaseIdentifier"] = identifierToJSON(son.DatabaseIdentifier)
	}
	if son.SchemaIdentifier != nil {
		node["SchemaIdentifier"] = identifierToJSON(son.SchemaIdentifier)
	}
	if son.BaseIdentifier != nil {
		node["BaseIdentifier"] = identifierToJSON(son.BaseIdentifier)
	}
	if son.Count > 0 {
		node["Count"] = son.Count
	}
	if len(son.Identifiers) > 0 {
		// Handle $ref for identifiers that reference the named identifiers
		ids := make([]any, len(son.Identifiers))
		for i, id := range son.Identifiers {
			// Check if this identifier is referenced by one of the named fields
			isRef := false
			if son.ServerIdentifier != nil && id == son.ServerIdentifier {
				isRef = true
			} else if son.DatabaseIdentifier != nil && id == son.DatabaseIdentifier {
				isRef = true
			} else if son.SchemaIdentifier != nil && id == son.SchemaIdentifier {
				isRef = true
			} else if son.BaseIdentifier != nil && id == son.BaseIdentifier {
				isRef = true
			}

			if isRef {
				ids[i] = jsonNode{"$ref": "Identifier"}
			} else {
				ids[i] = identifierToJSON(id)
			}
		}
		node["Identifiers"] = ids
	}
	return node
}

func booleanExpressionToJSON(expr ast.BooleanExpression) jsonNode {
	switch e := expr.(type) {
	case *ast.BooleanComparisonExpression:
		node := jsonNode{
			"$type": "BooleanComparisonExpression",
		}
		if e.ComparisonType != "" {
			node["ComparisonType"] = e.ComparisonType
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = scalarExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = scalarExpressionToJSON(e.SecondExpression)
		}
		return node
	case *ast.BooleanBinaryExpression:
		node := jsonNode{
			"$type": "BooleanBinaryExpression",
		}
		if e.BinaryExpressionType != "" {
			node["BinaryExpressionType"] = e.BinaryExpressionType
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = booleanExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = booleanExpressionToJSON(e.SecondExpression)
		}
		return node
	case *ast.BooleanParenthesisExpression:
		node := jsonNode{
			"$type": "BooleanParenthesisExpression",
		}
		if e.Expression != nil {
			node["Expression"] = booleanExpressionToJSON(e.Expression)
		}
		return node
	case *ast.BooleanIsNullExpression:
		node := jsonNode{
			"$type": "BooleanIsNullExpression",
		}
		node["IsNot"] = e.IsNot
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		return node
	case *ast.BooleanInExpression:
		node := jsonNode{
			"$type": "BooleanInExpression",
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		node["NotDefined"] = e.NotDefined
		if len(e.Values) > 0 {
			values := make([]jsonNode, len(e.Values))
			for i, v := range e.Values {
				values[i] = scalarExpressionToJSON(v)
			}
			node["Values"] = values
		}
		if e.Subquery != nil {
			node["Subquery"] = queryExpressionToJSON(e.Subquery)
		}
		return node
	case *ast.BooleanLikeExpression:
		node := jsonNode{
			"$type": "BooleanLikeExpression",
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = scalarExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = scalarExpressionToJSON(e.SecondExpression)
		}
		if e.EscapeExpression != nil {
			node["EscapeExpression"] = scalarExpressionToJSON(e.EscapeExpression)
		}
		node["NotDefined"] = e.NotDefined
		return node
	case *ast.BooleanTernaryExpression:
		node := jsonNode{
			"$type":                 "BooleanTernaryExpression",
			"TernaryExpressionType": e.TernaryExpressionType,
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = scalarExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = scalarExpressionToJSON(e.SecondExpression)
		}
		if e.ThirdExpression != nil {
			node["ThirdExpression"] = scalarExpressionToJSON(e.ThirdExpression)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownBooleanExpression"}
	}
}

func groupByClauseToJSON(gbc *ast.GroupByClause) jsonNode {
	node := jsonNode{
		"$type": "GroupByClause",
	}
	if gbc.GroupByOption != "" {
		node["GroupByOption"] = gbc.GroupByOption
	}
	// Always include All field
	node["All"] = gbc.All
	if len(gbc.GroupingSpecifications) > 0 {
		specs := make([]jsonNode, len(gbc.GroupingSpecifications))
		for i, spec := range gbc.GroupingSpecifications {
			specs[i] = groupingSpecificationToJSON(spec)
		}
		node["GroupingSpecifications"] = specs
	}
	return node
}

func groupingSpecificationToJSON(spec ast.GroupingSpecification) jsonNode {
	switch s := spec.(type) {
	case *ast.ExpressionGroupingSpecification:
		node := jsonNode{
			"$type": "ExpressionGroupingSpecification",
		}
		// Always include DistributedAggregation field
		node["DistributedAggregation"] = s.DistributedAggregation
		if s.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(s.Expression)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownGroupingSpecification"}
	}
}

func havingClauseToJSON(hc *ast.HavingClause) jsonNode {
	node := jsonNode{
		"$type": "HavingClause",
	}
	if hc.SearchCondition != nil {
		node["SearchCondition"] = booleanExpressionToJSON(hc.SearchCondition)
	}
	return node
}

func orderByClauseToJSON(obc *ast.OrderByClause) jsonNode {
	node := jsonNode{
		"$type": "OrderByClause",
	}
	if len(obc.OrderByElements) > 0 {
		elems := make([]jsonNode, len(obc.OrderByElements))
		for i, elem := range obc.OrderByElements {
			elems[i] = expressionWithSortOrderToJSON(elem)
		}
		node["OrderByElements"] = elems
	}
	return node
}

func expressionWithSortOrderToJSON(ewso *ast.ExpressionWithSortOrder) jsonNode {
	node := jsonNode{
		"$type": "ExpressionWithSortOrder",
	}
	if ewso.SortOrder != "" {
		node["SortOrder"] = ewso.SortOrder
	}
	if ewso.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(ewso.Expression)
	}
	return node
}

// ======================= New Statement JSON Functions =======================

func tableHintToJSON(h *ast.TableHint) jsonNode {
	node := jsonNode{
		"$type": "TableHint",
	}
	if h.HintKind != "" {
		node["HintKind"] = h.HintKind
	}
	return node
}

func insertStatementToJSON(s *ast.InsertStatement) jsonNode {
	node := jsonNode{
		"$type": "InsertStatement",
	}
	if s.InsertSpecification != nil {
		node["InsertSpecification"] = insertSpecificationToJSON(s.InsertSpecification)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return node
}

func insertSpecificationToJSON(spec *ast.InsertSpecification) jsonNode {
	node := jsonNode{
		"$type": "InsertSpecification",
	}
	if spec.InsertOption != "" && spec.InsertOption != "None" {
		node["InsertOption"] = spec.InsertOption
	}
	if spec.InsertSource != nil {
		node["InsertSource"] = insertSourceToJSON(spec.InsertSource)
	}
	if spec.Target != nil {
		node["Target"] = tableReferenceToJSON(spec.Target)
	}
	if len(spec.Columns) > 0 {
		cols := make([]jsonNode, len(spec.Columns))
		for i, c := range spec.Columns {
			cols[i] = scalarExpressionToJSON(c)
		}
		node["Columns"] = cols
	}
	return node
}

func insertSourceToJSON(src ast.InsertSource) jsonNode {
	switch s := src.(type) {
	case *ast.ValuesInsertSource:
		node := jsonNode{
			"$type": "ValuesInsertSource",
		}
		node["IsDefaultValues"] = s.IsDefaultValues
		if len(s.RowValues) > 0 {
			rows := make([]jsonNode, len(s.RowValues))
			for i, r := range s.RowValues {
				rows[i] = rowValueToJSON(r)
			}
			node["RowValues"] = rows
		}
		return node
	case *ast.SelectInsertSource:
		node := jsonNode{
			"$type": "SelectInsertSource",
		}
		if s.Select != nil {
			node["Select"] = queryExpressionToJSON(s.Select)
		}
		return node
	case *ast.ExecuteInsertSource:
		node := jsonNode{
			"$type": "ExecuteInsertSource",
		}
		if s.Execute != nil {
			node["Execute"] = executeSpecificationToJSON(s.Execute)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownInsertSource"}
	}
}

func rowValueToJSON(rv *ast.RowValue) jsonNode {
	node := jsonNode{
		"$type": "RowValue",
	}
	if len(rv.ColumnValues) > 0 {
		vals := make([]jsonNode, len(rv.ColumnValues))
		for i, v := range rv.ColumnValues {
			vals[i] = scalarExpressionToJSON(v)
		}
		node["ColumnValues"] = vals
	}
	return node
}

func executeSpecificationToJSON(spec *ast.ExecuteSpecification) jsonNode {
	node := jsonNode{
		"$type": "ExecuteSpecification",
	}
	if spec.Variable != nil {
		node["Variable"] = scalarExpressionToJSON(spec.Variable)
	}
	if spec.ExecutableEntity != nil {
		node["ExecutableEntity"] = executableEntityToJSON(spec.ExecutableEntity)
	}
	return node
}

func executableEntityToJSON(entity ast.ExecutableEntity) jsonNode {
	switch e := entity.(type) {
	case *ast.ExecutableProcedureReference:
		node := jsonNode{
			"$type": "ExecutableProcedureReference",
		}
		if e.ProcedureReference != nil {
			node["ProcedureReference"] = procedureReferenceNameToJSON(e.ProcedureReference)
		}
		if len(e.Parameters) > 0 {
			params := make([]jsonNode, len(e.Parameters))
			for i, p := range e.Parameters {
				params[i] = executeParameterToJSON(p)
			}
			node["Parameters"] = params
		}
		return node
	default:
		return jsonNode{"$type": "UnknownExecutableEntity"}
	}
}

func procedureReferenceNameToJSON(prn *ast.ProcedureReferenceName) jsonNode {
	node := jsonNode{
		"$type": "ProcedureReferenceName",
	}
	if prn.ProcedureVariable != nil {
		node["ProcedureVariable"] = scalarExpressionToJSON(prn.ProcedureVariable)
	}
	if prn.ProcedureReference != nil {
		node["ProcedureReference"] = procedureReferenceToJSON(prn.ProcedureReference)
	}
	return node
}

func procedureReferenceToJSON(pr *ast.ProcedureReference) jsonNode {
	node := jsonNode{
		"$type": "ProcedureReference",
	}
	if pr.Name != nil {
		node["Name"] = schemaObjectNameToJSON(pr.Name)
	}
	return node
}

func executeParameterToJSON(ep *ast.ExecuteParameter) jsonNode {
	node := jsonNode{
		"$type": "ExecuteParameter",
	}
	if ep.ParameterValue != nil {
		node["ParameterValue"] = scalarExpressionToJSON(ep.ParameterValue)
	}
	if ep.Variable != nil {
		node["Variable"] = scalarExpressionToJSON(ep.Variable)
	}
	node["IsOutput"] = ep.IsOutput
	return node
}

func updateStatementToJSON(s *ast.UpdateStatement) jsonNode {
	node := jsonNode{
		"$type": "UpdateStatement",
	}
	if s.UpdateSpecification != nil {
		node["UpdateSpecification"] = updateSpecificationToJSON(s.UpdateSpecification)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return node
}

func updateSpecificationToJSON(spec *ast.UpdateSpecification) jsonNode {
	node := jsonNode{
		"$type": "UpdateSpecification",
	}
	if len(spec.SetClauses) > 0 {
		clauses := make([]jsonNode, len(spec.SetClauses))
		for i, c := range spec.SetClauses {
			clauses[i] = setClauseToJSON(c)
		}
		node["SetClauses"] = clauses
	}
	if spec.Target != nil {
		node["Target"] = tableReferenceToJSON(spec.Target)
	}
	if spec.FromClause != nil {
		node["FromClause"] = fromClauseToJSON(spec.FromClause)
	}
	if spec.WhereClause != nil {
		node["WhereClause"] = whereClauseToJSON(spec.WhereClause)
	}
	return node
}

func setClauseToJSON(sc ast.SetClause) jsonNode {
	switch c := sc.(type) {
	case *ast.AssignmentSetClause:
		node := jsonNode{
			"$type": "AssignmentSetClause",
		}
		if c.Variable != nil {
			node["Variable"] = scalarExpressionToJSON(c.Variable)
		}
		if c.Column != nil {
			node["Column"] = scalarExpressionToJSON(c.Column)
		}
		if c.NewValue != nil {
			node["NewValue"] = scalarExpressionToJSON(c.NewValue)
		}
		if c.AssignmentKind != "" {
			node["AssignmentKind"] = c.AssignmentKind
		}
		return node
	default:
		return jsonNode{"$type": "UnknownSetClause"}
	}
}

func deleteStatementToJSON(s *ast.DeleteStatement) jsonNode {
	node := jsonNode{
		"$type": "DeleteStatement",
	}
	if s.DeleteSpecification != nil {
		node["DeleteSpecification"] = deleteSpecificationToJSON(s.DeleteSpecification)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return node
}

func deleteSpecificationToJSON(spec *ast.DeleteSpecification) jsonNode {
	node := jsonNode{
		"$type": "DeleteSpecification",
	}
	if spec.FromClause != nil {
		node["FromClause"] = fromClauseToJSON(spec.FromClause)
	}
	if spec.WhereClause != nil {
		node["WhereClause"] = whereClauseToJSON(spec.WhereClause)
	}
	if spec.Target != nil {
		node["Target"] = tableReferenceToJSON(spec.Target)
	}
	return node
}

func whereClauseToJSON(wc *ast.WhereClause) jsonNode {
	node := jsonNode{
		"$type": "WhereClause",
	}
	if wc.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(wc.Cursor)
	}
	if wc.SearchCondition != nil {
		node["SearchCondition"] = booleanExpressionToJSON(wc.SearchCondition)
	}
	return node
}

func cursorIdToJSON(cid *ast.CursorId) jsonNode {
	node := jsonNode{
		"$type": "CursorId",
	}
	node["IsGlobal"] = cid.IsGlobal
	if cid.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(cid.Name)
	}
	return node
}

func declareVariableStatementToJSON(s *ast.DeclareVariableStatement) jsonNode {
	node := jsonNode{
		"$type": "DeclareVariableStatement",
	}
	if len(s.Declarations) > 0 {
		decls := make([]jsonNode, len(s.Declarations))
		for i, d := range s.Declarations {
			decls[i] = declareVariableElementToJSON(d)
		}
		node["Declarations"] = decls
	}
	return node
}

func declareVariableElementToJSON(elem *ast.DeclareVariableElement) jsonNode {
	node := jsonNode{
		"$type": "DeclareVariableElement",
	}
	if elem.VariableName != nil {
		node["VariableName"] = identifierToJSON(elem.VariableName)
	}
	if elem.DataType != nil {
		node["DataType"] = sqlDataTypeReferenceToJSON(elem.DataType)
	}
	if elem.Nullable != nil {
		node["Nullable"] = nullableConstraintToJSON(elem.Nullable)
	}
	if elem.Value != nil {
		node["Value"] = scalarExpressionToJSON(elem.Value)
	}
	return node
}

func sqlDataTypeReferenceToJSON(dt *ast.SqlDataTypeReference) jsonNode {
	node := jsonNode{
		"$type": "SqlDataTypeReference",
	}
	if dt.SqlDataTypeOption != "" {
		node["SqlDataTypeOption"] = dt.SqlDataTypeOption
	}
	if len(dt.Parameters) > 0 {
		params := make([]jsonNode, len(dt.Parameters))
		for i, p := range dt.Parameters {
			params[i] = scalarExpressionToJSON(p)
		}
		node["Parameters"] = params
	}
	if dt.Name != nil {
		node["Name"] = schemaObjectNameToJSON(dt.Name)
	}
	return node
}

func setVariableStatementToJSON(s *ast.SetVariableStatement) jsonNode {
	node := jsonNode{
		"$type": "SetVariableStatement",
	}
	if s.Variable != nil {
		node["Variable"] = scalarExpressionToJSON(s.Variable)
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	if s.CursorDefinition != nil {
		node["CursorDefinition"] = cursorDefinitionToJSON(s.CursorDefinition)
	}
	if s.AssignmentKind != "" {
		node["AssignmentKind"] = s.AssignmentKind
	}
	if s.SeparatorType != "" {
		node["SeparatorType"] = s.SeparatorType
	}
	return node
}

func cursorDefinitionToJSON(cd *ast.CursorDefinition) jsonNode {
	node := jsonNode{
		"$type": "CursorDefinition",
	}
	if cd.Select != nil {
		node["Select"] = queryExpressionToJSON(cd.Select)
	}
	return node
}

func ifStatementToJSON(s *ast.IfStatement) jsonNode {
	node := jsonNode{
		"$type": "IfStatement",
	}
	if s.Predicate != nil {
		node["Predicate"] = booleanExpressionToJSON(s.Predicate)
	}
	if s.ThenStatement != nil {
		node["ThenStatement"] = statementToJSON(s.ThenStatement)
	}
	if s.ElseStatement != nil {
		node["ElseStatement"] = statementToJSON(s.ElseStatement)
	}
	return node
}

func whileStatementToJSON(s *ast.WhileStatement) jsonNode {
	node := jsonNode{
		"$type": "WhileStatement",
	}
	if s.Predicate != nil {
		node["Predicate"] = booleanExpressionToJSON(s.Predicate)
	}
	if s.Statement != nil {
		node["Statement"] = statementToJSON(s.Statement)
	}
	return node
}

func beginEndBlockStatementToJSON(s *ast.BeginEndBlockStatement) jsonNode {
	node := jsonNode{
		"$type": "BeginEndBlockStatement",
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func statementListToJSON(sl *ast.StatementList) jsonNode {
	node := jsonNode{
		"$type": "StatementList",
	}
	if len(sl.Statements) > 0 {
		stmts := make([]jsonNode, len(sl.Statements))
		for i, s := range sl.Statements {
			stmts[i] = statementToJSON(s)
		}
		node["Statements"] = stmts
	}
	return node
}

func createViewStatementToJSON(s *ast.CreateViewStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateViewStatement",
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	if len(s.Columns) > 0 {
		cols := make([]jsonNode, len(s.Columns))
		for i, c := range s.Columns {
			cols[i] = identifierToJSON(c)
		}
		node["Columns"] = cols
	}
	if s.SelectStatement != nil {
		node["SelectStatement"] = selectStatementToJSON(s.SelectStatement)
	}
	node["WithCheckOption"] = s.WithCheckOption
	node["IsMaterialized"] = s.IsMaterialized
	return node
}

func createSchemaStatementToJSON(s *ast.CreateSchemaStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSchemaStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func executeStatementToJSON(s *ast.ExecuteStatement) jsonNode {
	node := jsonNode{
		"$type": "ExecuteStatement",
	}
	if s.ExecuteSpecification != nil {
		node["ExecuteSpecification"] = executeSpecificationToJSON(s.ExecuteSpecification)
	}
	return node
}

func executeAsStatementToJSON(s *ast.ExecuteAsStatement) jsonNode {
	node := jsonNode{
		"$type":        "ExecuteAsStatement",
		"WithNoRevert": s.WithNoRevert,
	}
	if s.ExecuteContext != nil {
		node["ExecuteContext"] = executeContextToJSON(s.ExecuteContext)
	}
	if s.Cookie != nil {
		node["Cookie"] = scalarExpressionToJSON(s.Cookie)
	}
	return node
}

func executeContextToJSON(c *ast.ExecuteContext) jsonNode {
	node := jsonNode{
		"$type": "ExecuteContext",
		"Kind":  c.Kind,
	}
	if c.Principal != nil {
		node["Principal"] = scalarExpressionToJSON(c.Principal)
	}
	return node
}

func returnStatementToJSON(s *ast.ReturnStatement) jsonNode {
	node := jsonNode{
		"$type": "ReturnStatement",
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	return node
}

func breakStatementToJSON() jsonNode {
	return jsonNode{
		"$type": "BreakStatement",
	}
}

func continueStatementToJSON() jsonNode {
	return jsonNode{
		"$type": "ContinueStatement",
	}
}

func (p *Parser) parseCreateTableStatement() (*ast.CreateTableStatement, error) {
	// Consume TABLE
	p.nextToken()

	stmt := &ast.CreateTableStatement{}

	// Parse table name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.SchemaObjectName = name

	// Expect ( - if not present, be lenient
	if p.curTok.Type != TokenLParen {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	stmt.Definition = &ast.TableDefinition{}

	// Parse column definitions
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		colDef, err := p.parseColumnDefinition()
		if err != nil {
			p.skipToEndOfStatement()
			return stmt, nil
		}
		stmt.Definition.ColumnDefinitions = append(stmt.Definition.ColumnDefinitions, colDef)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Expect )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseColumnDefinition() (*ast.ColumnDefinition, error) {
	col := &ast.ColumnDefinition{}

	// Parse column name (parseIdentifier already calls nextToken)
	col.ColumnIdentifier = p.parseIdentifier()

	// Parse data type - be lenient if no data type is provided
	dataType, err := p.parseDataType()
	if err != nil {
		// Lenient: return column definition without data type
		return col, nil
	}
	col.DataType = dataType

	// Parse optional IDENTITY specification
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "IDENTITY" {
		p.nextToken() // consume IDENTITY
		identityOpts := &ast.IdentityOptions{}

		// Check for optional (seed, increment)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (

			// Parse seed
			if p.curTok.Type == TokenNumber {
				identityOpts.IdentitySeed = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
				p.nextToken()
			}

			// Expect comma
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,

				// Parse increment
				if p.curTok.Type == TokenNumber {
					identityOpts.IdentityIncrement = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
					p.nextToken()
				}
			}

			// Expect closing paren
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}

		// Check for NOT FOR REPLICATION
		if p.curTok.Type == TokenNot {
			p.nextToken() // consume NOT
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
				if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
					p.nextToken() // consume REPLICATION
					identityOpts.NotForReplication = true
				}
			}
		}

		col.IdentityOptions = identityOpts
	}

	// Parse optional NULL/NOT NULL constraint
	if p.curTok.Type == TokenNot {
		p.nextToken() // consume NOT
		if p.curTok.Type != TokenNull {
			return nil, fmt.Errorf("expected NULL after NOT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume NULL
		col.Constraints = append(col.Constraints, &ast.NullableConstraintDefinition{Nullable: false})
	} else if p.curTok.Type == TokenNull {
		p.nextToken() // consume NULL
		col.Constraints = append(col.Constraints, &ast.NullableConstraintDefinition{Nullable: true})
	}

	return col, nil
}

func (p *Parser) parseGrantStatement() (*ast.GrantStatement, error) {
	// Consume GRANT
	p.nextToken()

	stmt := &ast.GrantStatement{}

	// Parse permission(s)
	perm := &ast.Permission{}
	for p.curTok.Type != TokenTo && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenCreate ||
			p.curTok.Type == TokenProcedure || p.curTok.Type == TokenView ||
			p.curTok.Type == TokenSelect || p.curTok.Type == TokenInsert ||
			p.curTok.Type == TokenUpdate || p.curTok.Type == TokenDelete {
			perm.Identifiers = append(perm.Identifiers, &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
			})
			p.nextToken()
		} else if p.curTok.Type == TokenComma {
			stmt.Permissions = append(stmt.Permissions, perm)
			perm = &ast.Permission{}
			p.nextToken()
		} else {
			break
		}
	}
	if len(perm.Identifiers) > 0 {
		stmt.Permissions = append(stmt.Permissions, perm)
	}

	// Expect TO
	if p.curTok.Type == TokenTo {
		p.nextToken()
	}

	// Parse principal(s)
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		principal := &ast.SecurityPrincipal{}
		if p.curTok.Type == TokenPublic {
			principal.PrincipalType = "Public"
			p.nextToken()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			principal.PrincipalType = "Identifier"
			// parseIdentifier already calls nextToken()
			principal.Identifier = p.parseIdentifier()
		} else {
			break
		}
		stmt.Principals = append(stmt.Principals, principal)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func createTableStatementToJSON(s *ast.CreateTableStatement) jsonNode {
	node := jsonNode{
		"$type":       "CreateTableStatement",
		"AsEdge":      s.AsEdge,
		"AsFileTable": s.AsFileTable,
		"AsNode":      s.AsNode,
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	if s.Definition != nil {
		node["Definition"] = tableDefinitionToJSON(s.Definition)
	}
	return node
}

func tableDefinitionToJSON(t *ast.TableDefinition) jsonNode {
	if t == nil {
		return nil
	}
	node := jsonNode{
		"$type": "TableDefinition",
	}
	if len(t.ColumnDefinitions) > 0 {
		cols := make([]jsonNode, len(t.ColumnDefinitions))
		for i, col := range t.ColumnDefinitions {
			cols[i] = columnDefinitionToJSON(col)
		}
		node["ColumnDefinitions"] = cols
	}
	if len(t.Indexes) > 0 {
		indexes := make([]jsonNode, len(t.Indexes))
		for i, idx := range t.Indexes {
			indexes[i] = indexDefinitionToJSON(idx)
		}
		node["Indexes"] = indexes
	}
	return node
}

func columnDefinitionToJSON(c *ast.ColumnDefinition) jsonNode {
	node := jsonNode{
		"$type":            "ColumnDefinition",
		"IsPersisted":      c.IsPersisted,
		"IsRowGuidCol":     c.IsRowGuidCol,
		"IsHidden":         c.IsHidden,
		"IsMasked":         c.IsMasked,
		"ColumnIdentifier": identifierToJSON(c.ColumnIdentifier),
	}
	if c.IdentityOptions != nil {
		node["IdentityOptions"] = identityOptionsToJSON(c.IdentityOptions)
	}
	if len(c.Constraints) > 0 {
		constraints := make([]jsonNode, len(c.Constraints))
		for i, constraint := range c.Constraints {
			constraints[i] = constraintDefinitionToJSON(constraint)
		}
		node["Constraints"] = constraints
	}
	if c.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(c.DataType)
	}
	return node
}

func identityOptionsToJSON(i *ast.IdentityOptions) jsonNode {
	node := jsonNode{
		"$type":                       "IdentityOptions",
		"IsIdentityNotForReplication": i.NotForReplication,
	}
	if i.IdentitySeed != nil {
		node["IdentitySeed"] = scalarExpressionToJSON(i.IdentitySeed)
	}
	if i.IdentityIncrement != nil {
		node["IdentityIncrement"] = scalarExpressionToJSON(i.IdentityIncrement)
	}
	return node
}

func constraintDefinitionToJSON(c ast.ConstraintDefinition) jsonNode {
	switch constraint := c.(type) {
	case *ast.NullableConstraintDefinition:
		return jsonNode{
			"$type":    "NullableConstraintDefinition",
			"Nullable": constraint.Nullable,
		}
	default:
		return jsonNode{"$type": "UnknownConstraint"}
	}
}

func dataTypeReferenceToJSON(d ast.DataTypeReference) jsonNode {
	switch dt := d.(type) {
	case *ast.SqlDataTypeReference:
		return sqlDataTypeReferenceToJSON(dt)
	case *ast.XmlDataTypeReference:
		return xmlDataTypeReferenceToJSON(dt)
	default:
		return jsonNode{"$type": "UnknownDataType"}
	}
}

func xmlDataTypeReferenceToJSON(dt *ast.XmlDataTypeReference) jsonNode {
	node := jsonNode{
		"$type": "XmlDataTypeReference",
	}
	if dt.XmlDataTypeOption != "" {
		node["XmlDataTypeOption"] = dt.XmlDataTypeOption
	}
	if dt.XmlSchemaCollection != nil {
		node["XmlSchemaCollection"] = schemaObjectNameToJSON(dt.XmlSchemaCollection)
	}
	if dt.Name != nil {
		node["Name"] = schemaObjectNameToJSON(dt.Name)
	}
	return node
}

func grantStatementToJSON(s *ast.GrantStatement) jsonNode {
	node := jsonNode{
		"$type":           "GrantStatement",
		"WithGrantOption": s.WithGrantOption,
	}
	if len(s.Permissions) > 0 {
		perms := make([]jsonNode, len(s.Permissions))
		for i, p := range s.Permissions {
			perms[i] = permissionToJSON(p)
		}
		node["Permissions"] = perms
	}
	if len(s.Principals) > 0 {
		principals := make([]jsonNode, len(s.Principals))
		for i, p := range s.Principals {
			principals[i] = securityPrincipalToJSON(p)
		}
		node["Principals"] = principals
	}
	return node
}

func permissionToJSON(p *ast.Permission) jsonNode {
	node := jsonNode{
		"$type": "Permission",
	}
	if len(p.Identifiers) > 0 {
		ids := make([]jsonNode, len(p.Identifiers))
		for i, id := range p.Identifiers {
			ids[i] = identifierToJSON(id)
		}
		node["Identifiers"] = ids
	}
	return node
}

func securityPrincipalToJSON(p *ast.SecurityPrincipal) jsonNode {
	node := jsonNode{
		"$type":         "SecurityPrincipal",
		"PrincipalType": p.PrincipalType,
	}
	if p.Identifier != nil {
		node["Identifier"] = identifierToJSON(p.Identifier)
	}
	return node
}

func predicateSetStatementToJSON(s *ast.PredicateSetStatement) jsonNode {
	return jsonNode{
		"$type":   "PredicateSetStatement",
		"Options": string(s.Options),
		"IsOn":    s.IsOn,
	}
}

func setStatisticsStatementToJSON(s *ast.SetStatisticsStatement) jsonNode {
	return jsonNode{
		"$type":   "SetStatisticsStatement",
		"Options": string(s.Options),
		"IsOn":    s.IsOn,
	}
}

func commitTransactionStatementToJSON(s *ast.CommitTransactionStatement) jsonNode {
	node := jsonNode{
		"$type":                   "CommitTransactionStatement",
		"DelayedDurabilityOption": s.DelayedDurabilityOption,
	}
	if s.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(s.Name)
	}
	return node
}

func rollbackTransactionStatementToJSON(s *ast.RollbackTransactionStatement) jsonNode {
	node := jsonNode{
		"$type": "RollbackTransactionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(s.Name)
	}
	return node
}

func saveTransactionStatementToJSON(s *ast.SaveTransactionStatement) jsonNode {
	node := jsonNode{
		"$type": "SaveTransactionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(s.Name)
	}
	return node
}

func beginTransactionStatementToJSON(s *ast.BeginTransactionStatement) jsonNode {
	node := jsonNode{
		"$type":       "BeginTransactionStatement",
		"Distributed": s.Distributed,
		"MarkDefined": s.MarkDefined,
	}
	if s.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(s.Name)
	}
	if s.MarkDescription != nil {
		node["MarkDescription"] = scalarExpressionToJSON(s.MarkDescription)
	}
	return node
}

func waitForStatementToJSON(s *ast.WaitForStatement) jsonNode {
	node := jsonNode{
		"$type":         "WaitForStatement",
		"WaitForOption": s.WaitForOption,
	}
	if s.Parameter != nil {
		node["Parameter"] = scalarExpressionToJSON(s.Parameter)
	}
	if s.Timeout != nil {
		node["Timeout"] = scalarExpressionToJSON(s.Timeout)
	}
	if s.Statement != nil {
		node["Statement"] = statementToJSON(s.Statement)
	}
	return node
}

func moveConversationStatementToJSON(s *ast.MoveConversationStatement) jsonNode {
	node := jsonNode{
		"$type": "MoveConversationStatement",
	}
	if s.Conversation != nil {
		node["Conversation"] = scalarExpressionToJSON(s.Conversation)
	}
	if s.Group != nil {
		node["Group"] = scalarExpressionToJSON(s.Group)
	}
	return node
}

func getConversationGroupStatementToJSON(s *ast.GetConversationGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "GetConversationGroupStatement",
	}
	if s.GroupId != nil {
		node["GroupId"] = scalarExpressionToJSON(s.GroupId)
	}
	if s.Queue != nil {
		node["Queue"] = schemaObjectNameToJSON(s.Queue)
	}
	return node
}

func truncateTableStatementToJSON(s *ast.TruncateTableStatement) jsonNode {
	node := jsonNode{
		"$type": "TruncateTableStatement",
	}
	if s.TableName != nil {
		node["TableName"] = schemaObjectNameToJSON(s.TableName)
	}
	if len(s.PartitionRanges) > 0 {
		ranges := make([]jsonNode, len(s.PartitionRanges))
		for i, pr := range s.PartitionRanges {
			ranges[i] = compressionPartitionRangeToJSON(pr)
		}
		node["PartitionRanges"] = ranges
	}
	return node
}

func compressionPartitionRangeToJSON(pr *ast.CompressionPartitionRange) jsonNode {
	node := jsonNode{
		"$type": "CompressionPartitionRange",
	}
	if pr.From != nil {
		node["From"] = scalarExpressionToJSON(pr.From)
	}
	if pr.To != nil {
		node["To"] = scalarExpressionToJSON(pr.To)
	}
	return node
}

func useStatementToJSON(s *ast.UseStatement) jsonNode {
	node := jsonNode{
		"$type": "UseStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	return node
}

func killStatementToJSON(s *ast.KillStatement) jsonNode {
	node := jsonNode{
		"$type":          "KillStatement",
		"WithStatusOnly": s.WithStatusOnly,
	}
	if s.Parameter != nil {
		node["Parameter"] = scalarExpressionToJSON(s.Parameter)
	}
	return node
}

func killStatsJobStatementToJSON(s *ast.KillStatsJobStatement) jsonNode {
	node := jsonNode{
		"$type": "KillStatsJobStatement",
	}
	if s.JobId != nil {
		node["JobId"] = scalarExpressionToJSON(s.JobId)
	}
	return node
}

func killQueryNotificationSubscriptionStatementToJSON(s *ast.KillQueryNotificationSubscriptionStatement) jsonNode {
	node := jsonNode{
		"$type": "KillQueryNotificationSubscriptionStatement",
		"All":   s.All,
	}
	if s.SubscriptionId != nil {
		node["SubscriptionId"] = scalarExpressionToJSON(s.SubscriptionId)
	}
	return node
}

func closeSymmetricKeyStatementToJSON(s *ast.CloseSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CloseSymmetricKeyStatement",
		"All":   s.All,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func closeMasterKeyStatementToJSON(s *ast.CloseMasterKeyStatement) jsonNode {
	return jsonNode{
		"$type": "CloseMasterKeyStatement",
	}
}

func openMasterKeyStatementToJSON(s *ast.OpenMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "OpenMasterKeyStatement",
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return node
}

func openSymmetricKeyStatementToJSON(s *ast.OpenSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "OpenSymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.DecryptionMechanism != "" {
		node["DecryptionMechanism"] = s.DecryptionMechanism
	}
	if s.DecryptionKey != nil {
		node["DecryptionKey"] = scalarExpressionToJSON(s.DecryptionKey)
	}
	return node
}

func checkpointStatementToJSON(s *ast.CheckpointStatement) jsonNode {
	node := jsonNode{
		"$type": "CheckpointStatement",
	}
	if s.Duration != nil {
		node["Duration"] = scalarExpressionToJSON(s.Duration)
	}
	return node
}

func reconfigureStatementToJSON(s *ast.ReconfigureStatement) jsonNode {
	return jsonNode{
		"$type":        "ReconfigureStatement",
		"WithOverride": s.WithOverride,
	}
}

func shutdownStatementToJSON(s *ast.ShutdownStatement) jsonNode {
	return jsonNode{
		"$type":      "ShutdownStatement",
		"WithNoWait": s.WithNoWait,
	}
}

func setUserStatementToJSON(s *ast.SetUserStatement) jsonNode {
	node := jsonNode{
		"$type":       "SetUserStatement",
		"WithNoReset": s.WithNoReset,
	}
	if s.UserName != nil {
		node["UserName"] = scalarExpressionToJSON(s.UserName)
	}
	return node
}

func lineNoStatementToJSON(s *ast.LineNoStatement) jsonNode {
	node := jsonNode{
		"$type": "LineNoStatement",
	}
	if s.LineNo != nil {
		node["LineNo"] = scalarExpressionToJSON(s.LineNo)
	}
	return node
}

func raiseErrorStatementToJSON(s *ast.RaiseErrorStatement) jsonNode {
	node := jsonNode{
		"$type": "RaiseErrorStatement",
	}
	if s.FirstParameter != nil {
		node["FirstParameter"] = scalarExpressionToJSON(s.FirstParameter)
	}
	if s.SecondParameter != nil {
		node["SecondParameter"] = scalarExpressionToJSON(s.SecondParameter)
	}
	if s.ThirdParameter != nil {
		node["ThirdParameter"] = scalarExpressionToJSON(s.ThirdParameter)
	}
	if len(s.OptionalParameters) > 0 {
		params := make([]jsonNode, len(s.OptionalParameters))
		for i, param := range s.OptionalParameters {
			params[i] = scalarExpressionToJSON(param)
		}
		node["OptionalParameters"] = params
	}
	if s.RaiseErrorOptions != "" {
		node["RaiseErrorOptions"] = s.RaiseErrorOptions
	}
	return node
}

func readTextStatementToJSON(s *ast.ReadTextStatement) jsonNode {
	node := jsonNode{
		"$type":    "ReadTextStatement",
		"HoldLock": s.HoldLock,
	}
	if s.Column != nil {
		node["Column"] = columnReferenceExpressionToJSON(s.Column)
	}
	if s.TextPointer != nil {
		node["TextPointer"] = scalarExpressionToJSON(s.TextPointer)
	}
	if s.Offset != nil {
		node["Offset"] = scalarExpressionToJSON(s.Offset)
	}
	if s.Size != nil {
		node["Size"] = scalarExpressionToJSON(s.Size)
	}
	return node
}

func writeTextStatementToJSON(s *ast.WriteTextStatement) jsonNode {
	node := jsonNode{
		"$type":   "WriteTextStatement",
		"Bulk":    s.Bulk,
		"WithLog": s.WithLog,
	}
	if s.SourceParameter != nil {
		node["SourceParameter"] = scalarExpressionToJSON(s.SourceParameter)
	}
	if s.Column != nil {
		node["Column"] = columnReferenceExpressionToJSON(s.Column)
	}
	if s.TextId != nil {
		node["TextId"] = scalarExpressionToJSON(s.TextId)
	}
	return node
}

func updateTextStatementToJSON(s *ast.UpdateTextStatement) jsonNode {
	node := jsonNode{
		"$type":   "UpdateTextStatement",
		"Bulk":    s.Bulk,
		"WithLog": s.WithLog,
	}
	if s.InsertOffset != nil {
		node["InsertOffset"] = scalarExpressionToJSON(s.InsertOffset)
	}
	if s.DeleteLength != nil {
		node["DeleteLength"] = scalarExpressionToJSON(s.DeleteLength)
	}
	if s.SourceColumn != nil {
		node["SourceColumn"] = columnReferenceExpressionToJSON(s.SourceColumn)
	}
	if s.SourceParameter != nil {
		node["SourceParameter"] = scalarExpressionToJSON(s.SourceParameter)
	}
	if s.Column != nil {
		node["Column"] = columnReferenceExpressionToJSON(s.Column)
	}
	if s.TextId != nil {
		node["TextId"] = scalarExpressionToJSON(s.TextId)
	}
	if s.Timestamp != nil {
		node["Timestamp"] = scalarExpressionToJSON(s.Timestamp)
	}
	return node
}

func columnReferenceExpressionToJSON(c *ast.ColumnReferenceExpression) jsonNode {
	node := jsonNode{
		"$type": "ColumnReferenceExpression",
	}
	if c.ColumnType != "" {
		node["ColumnType"] = c.ColumnType
	}
	if c.MultiPartIdentifier != nil {
		node["MultiPartIdentifier"] = multiPartIdentifierToJSON(c.MultiPartIdentifier)
	}
	return node
}

func goToStatementToJSON(s *ast.GoToStatement) jsonNode {
	node := jsonNode{
		"$type": "GoToStatement",
	}
	if s.LabelName != nil {
		node["LabelName"] = identifierToJSON(s.LabelName)
	}
	return node
}

func labelStatementToJSON(s *ast.LabelStatement) jsonNode {
	return jsonNode{
		"$type": "LabelStatement",
		"Value": s.Value,
	}
}

func createDefaultStatementToJSON(s *ast.CreateDefaultStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateDefaultStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	return node
}

func createMasterKeyStatementToJSON(s *ast.CreateMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateMasterKeyStatement",
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return node
}

func tryCatchStatementToJSON(s *ast.TryCatchStatement) jsonNode {
	node := jsonNode{
		"$type": "TryCatchStatement",
	}
	if s.TryStatements != nil {
		node["TryStatements"] = statementListToJSON(s.TryStatements)
	}
	if s.CatchStatements != nil {
		node["CatchStatements"] = statementListToJSON(s.CatchStatements)
	}
	return node
}

func sendStatementToJSON(s *ast.SendStatement) jsonNode {
	node := jsonNode{
		"$type": "SendStatement",
	}
	if len(s.ConversationHandles) > 0 {
		handles := make([]jsonNode, len(s.ConversationHandles))
		for i, h := range s.ConversationHandles {
			handles[i] = scalarExpressionToJSON(h)
		}
		node["ConversationHandles"] = handles
	}
	if s.MessageTypeName != nil {
		node["MessageTypeName"] = identifierOrValueExpressionToJSON(s.MessageTypeName)
	}
	if s.MessageBody != nil {
		node["MessageBody"] = scalarExpressionToJSON(s.MessageBody)
	}
	return node
}

func receiveStatementToJSON(s *ast.ReceiveStatement) jsonNode {
	node := jsonNode{
		"$type": "ReceiveStatement",
	}
	if s.Top != nil {
		node["Top"] = scalarExpressionToJSON(s.Top)
	}
	if len(s.SelectElements) > 0 {
		elems := make([]jsonNode, len(s.SelectElements))
		for i, e := range s.SelectElements {
			elems[i] = selectElementToJSON(e)
		}
		node["SelectElements"] = elems
	}
	if s.Queue != nil {
		node["Queue"] = schemaObjectNameToJSON(s.Queue)
	}
	if s.Into != nil {
		node["Into"] = variableTableReferenceToJSON(s.Into)
	}
	if s.Where != nil {
		node["Where"] = booleanExpressionToJSON(s.Where)
	}
	node["IsConversationGroupIdWhere"] = s.IsConversationGroupIdWhere
	return node
}

func variableTableReferenceToJSON(v *ast.VariableTableReference) jsonNode {
	node := jsonNode{
		"$type": "VariableTableReference",
	}
	if v.Variable != nil {
		varNode := jsonNode{
			"$type": "VariableReference",
		}
		if v.Variable.Name != "" {
			varNode["Name"] = v.Variable.Name
		}
		node["Variable"] = varNode
	}
	node["ForPath"] = v.ForPath
	return node
}

func createCredentialStatementToJSON(s *ast.CreateCredentialStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateCredentialStatement",
	}
	if s.CryptographicProviderName != nil {
		node["CryptographicProviderName"] = identifierToJSON(s.CryptographicProviderName)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Identity != nil {
		node["Identity"] = scalarExpressionToJSON(s.Identity)
	}
	if s.Secret != nil {
		node["Secret"] = scalarExpressionToJSON(s.Secret)
	}
	node["IsDatabaseScoped"] = s.IsDatabaseScoped
	return node
}

func alterMasterKeyStatementToJSON(s *ast.AlterMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type":  "AlterMasterKeyStatement",
		"Option": s.Option,
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return node
}

func alterSchemaStatementToJSON(s *ast.AlterSchemaStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterSchemaStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.ObjectName != nil {
		node["ObjectName"] = schemaObjectNameToJSON(s.ObjectName)
	}
	node["ObjectKind"] = s.ObjectKind
	return node
}

func alterRoleStatementToJSON(s *ast.AlterRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterRoleStatement",
	}
	if s.Action != nil {
		node["Action"] = alterRoleActionToJSON(s.Action)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterRoleActionToJSON(a ast.AlterRoleAction) jsonNode {
	switch action := a.(type) {
	case *ast.AddMemberAlterRoleAction:
		node := jsonNode{
			"$type": "AddMemberAlterRoleAction",
		}
		if action.Member != nil {
			node["Member"] = identifierToJSON(action.Member)
		}
		return node
	case *ast.DropMemberAlterRoleAction:
		node := jsonNode{
			"$type": "DropMemberAlterRoleAction",
		}
		if action.Member != nil {
			node["Member"] = identifierToJSON(action.Member)
		}
		return node
	case *ast.RenameAlterRoleAction:
		node := jsonNode{
			"$type": "RenameAlterRoleAction",
		}
		if action.NewName != nil {
			node["NewName"] = identifierToJSON(action.NewName)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownAlterRoleAction"}
	}
}

func alterRemoteServiceBindingStatementToJSON(s *ast.AlterRemoteServiceBindingStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterRemoteServiceBindingStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = remoteServiceBindingOptionToJSON(o)
		}
		node["Options"] = options
	}
	return node
}

func remoteServiceBindingOptionToJSON(o ast.RemoteServiceBindingOption) jsonNode {
	switch opt := o.(type) {
	case *ast.UserRemoteServiceBindingOption:
		node := jsonNode{
			"$type":      "UserRemoteServiceBindingOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.User != nil {
			node["User"] = identifierToJSON(opt.User)
		}
		return node
	case *ast.OnOffRemoteServiceBindingOption:
		return jsonNode{
			"$type":       "OnOffRemoteServiceBindingOption",
			"OptionState": opt.OptionState,
			"OptionKind":  opt.OptionKind,
		}
	default:
		return jsonNode{"$type": "UnknownRemoteServiceBindingOption"}
	}
}

func alterXmlSchemaCollectionStatementToJSON(s *ast.AlterXmlSchemaCollectionStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterXmlSchemaCollectionStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	return node
}

func createXmlSchemaCollectionStatementToJSON(s *ast.CreateXmlSchemaCollectionStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateXmlSchemaCollectionStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	return node
}

func createSearchPropertyListStatementToJSON(s *ast.CreateSearchPropertyListStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSearchPropertyListStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.SourceSearchPropertyList != nil {
		node["SourceSearchPropertyList"] = multiPartIdentifierToJSON(s.SourceSearchPropertyList)
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	return node
}

func alterServerConfigurationSetSoftNumaStatementToJSON(s *ast.AlterServerConfigurationSetSoftNumaStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationSetSoftNumaStatement",
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = alterServerConfigurationSoftNumaOptionToJSON(o)
		}
		node["Options"] = options
	}
	return node
}

func alterServerConfigurationSoftNumaOptionToJSON(o *ast.AlterServerConfigurationSoftNumaOption) jsonNode {
	node := jsonNode{
		"$type":      "AlterServerConfigurationSoftNumaOption",
		"OptionKind": o.OptionKind,
	}
	if o.OptionValue != nil {
		node["OptionValue"] = onOffOptionValueToJSON(o.OptionValue)
	}
	return node
}

func onOffOptionValueToJSON(o *ast.OnOffOptionValue) jsonNode {
	return jsonNode{
		"$type":       "OnOffOptionValue",
		"OptionState": o.OptionState,
	}
}

func alterLoginAddDropCredentialStatementToJSON(s *ast.AlterLoginAddDropCredentialStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterLoginAddDropCredentialStatement",
		"IsAdd": s.IsAdd,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.CredentialName != nil {
		node["CredentialName"] = identifierToJSON(s.CredentialName)
	}
	return node
}

func createProcedureStatementToJSON(s *ast.CreateProcedureStatement) jsonNode {
	node := jsonNode{
		"$type":            "CreateProcedureStatement",
		"IsForReplication": s.IsForReplication,
	}
	if s.ProcedureReference != nil {
		node["ProcedureReference"] = procedureReferenceToJSON(s.ProcedureReference)
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = procedureParameterToJSON(p)
		}
		node["Parameters"] = params
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func createRoleStatementToJSON(s *ast.CreateRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateRoleStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func procedureParameterToJSON(p *ast.ProcedureParameter) jsonNode {
	node := jsonNode{
		"$type":     "ProcedureParameter",
		"IsVarying": p.IsVarying,
		"Modifier":  p.Modifier,
	}
	if p.VariableName != nil {
		node["VariableName"] = identifierToJSON(p.VariableName)
	}
	if p.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(p.DataType)
	}
	if p.Value != nil {
		node["Value"] = scalarExpressionToJSON(p.Value)
	}
	if p.Nullable != nil {
		node["Nullable"] = nullableConstraintToJSON(p.Nullable)
	}
	return node
}

func nullableConstraintToJSON(n *ast.NullableConstraintDefinition) jsonNode {
	return jsonNode{
		"$type":    "NullableConstraintDefinition",
		"Nullable": n.Nullable,
	}
}

// parseRestoreStatement parses a RESTORE DATABASE statement
func (p *Parser) parseRestoreStatement() (*ast.RestoreStatement, error) {
	// Consume RESTORE
	p.nextToken()

	stmt := &ast.RestoreStatement{}

	// Parse restore kind (DATABASE, LOG, etc.)
	switch strings.ToUpper(p.curTok.Literal) {
	case "DATABASE":
		stmt.Kind = "Database"
		p.nextToken()
	case "LOG":
		stmt.Kind = "Log"
		p.nextToken()
	default:
		stmt.Kind = "Database"
	}

	// Parse database name
	dbName := &ast.IdentifierOrValueExpression{}
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		// Variable reference
		varRef := &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
		dbName.Value = varRef.Name
		dbName.ValueExpression = varRef
	} else {
		ident := p.parseIdentifier()
		dbName.Value = ident.Value
		dbName.Identifier = ident
	}
	stmt.DatabaseName = dbName

	// Expect FROM
	if strings.ToUpper(p.curTok.Literal) != "FROM" {
		return nil, fmt.Errorf("expected FROM, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse devices
	for {
		device := &ast.DeviceInfo{DeviceType: "None"}

		// Check for device type
		switch strings.ToUpper(p.curTok.Literal) {
		case "DISK":
			device.DeviceType = "Disk"
			p.nextToken()
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after DISK, got %s", p.curTok.Literal)
			}
			p.nextToken()
		case "URL":
			device.DeviceType = "URL"
			p.nextToken()
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after URL, got %s", p.curTok.Literal)
			}
			p.nextToken()
		}

		// Parse device name
		deviceName := &ast.IdentifierOrValueExpression{}
		if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
			varRef := &ast.VariableReference{Name: p.curTok.Literal}
			p.nextToken()
			deviceName.Value = varRef.Name
			deviceName.ValueExpression = varRef
		} else if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
			strLit := &ast.StringLiteral{
				LiteralType:   "String",
				Value:         p.curTok.Literal,
				IsNational:    p.curTok.Type == TokenNationalString,
				IsLargeObject: false,
			}
			deviceName.Value = strLit.Value
			deviceName.ValueExpression = strLit
			p.nextToken()
		} else {
			ident := p.parseIdentifier()
			deviceName.Value = ident.Value
			deviceName.Identifier = ident
		}
		device.LogicalDevice = deviceName
		stmt.Devices = append(stmt.Devices, device)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken()

		for {
			optionName := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			switch optionName {
			case "FILESTREAM":
				if p.curTok.Type != TokenLParen {
					return nil, fmt.Errorf("expected ( after FILESTREAM, got %s", p.curTok.Literal)
				}
				p.nextToken()

				fsOpt := &ast.FileStreamRestoreOption{
					OptionKind: "FileStream",
					FileStreamOption: &ast.FileStreamDatabaseOption{
						OptionKind: "FileStream",
					},
				}

				// Parse FILESTREAM options
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					fsOptName := strings.ToUpper(p.curTok.Literal)
					p.nextToken()

					if p.curTok.Type != TokenEquals {
						return nil, fmt.Errorf("expected = after %s, got %s", fsOptName, p.curTok.Literal)
					}
					p.nextToken()

					switch fsOptName {
					case "DIRECTORY_NAME":
						expr, err := p.parseScalarExpression()
						if err != nil {
							return nil, err
						}
						fsOpt.FileStreamOption.DirectoryName = expr
					}

					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}

				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
				stmt.Options = append(stmt.Options, fsOpt)

			default:
				// Generic option
				opt := &ast.GeneralSetCommandRestoreOption{
					OptionKind: optionName,
				}
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					opt.OptionValue = expr
				}
				stmt.Options = append(stmt.Options, opt)
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCreateUserStatement parses a CREATE USER statement
func (p *Parser) parseCreateUserStatement() (*ast.CreateUserStatement, error) {
	// Consume USER
	p.nextToken()

	stmt := &ast.CreateUserStatement{}

	// Parse user name
	stmt.Name = p.parseIdentifier()

	// Check for login option
	if strings.ToUpper(p.curTok.Literal) == "FOR" || strings.ToUpper(p.curTok.Literal) == "FROM" {
		isFor := strings.ToUpper(p.curTok.Literal) == "FOR"
		p.nextToken()

		loginOption := &ast.UserLoginOption{}

		switch strings.ToUpper(p.curTok.Literal) {
		case "LOGIN":
			if isFor {
				loginOption.UserLoginOptionType = "ForLogin"
			} else {
				loginOption.UserLoginOptionType = "FromLogin"
			}
			p.nextToken()
			loginOption.Identifier = p.parseIdentifier()
		case "CERTIFICATE":
			loginOption.UserLoginOptionType = "FromCertificate"
			p.nextToken()
			loginOption.Identifier = p.parseIdentifier()
		case "ASYMMETRIC":
			p.nextToken() // consume ASYMMETRIC
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			loginOption.UserLoginOptionType = "FromAsymmetricKey"
			loginOption.Identifier = p.parseIdentifier()
		case "EXTERNAL":
			p.nextToken() // consume EXTERNAL
			if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
				p.nextToken() // consume PROVIDER
			}
			loginOption.UserLoginOptionType = "External"
		}
		stmt.UserLoginOption = loginOption
	} else if strings.ToUpper(p.curTok.Literal) == "WITHOUT" {
		p.nextToken() // consume WITHOUT
		if p.curTok.Type == TokenLogin {
			p.nextToken() // consume LOGIN
		}
		stmt.UserLoginOption = &ast.UserLoginOption{
			UserLoginOptionType: "WithoutLogin",
		}
	}

	// Parse WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken()

		for {
			optionName := p.curTok.Literal
			p.nextToken()

			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after %s, got %s", optionName, p.curTok.Literal)
			}
			p.nextToken()

			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			// Check if value is a simple identifier (ColumnReferenceExpression with single identifier)
			// If so, use IdentifierPrincipalOption instead
			var opt ast.UserOption
			if colRef, ok := value.(*ast.ColumnReferenceExpression); ok && colRef.MultiPartIdentifier != nil && len(colRef.MultiPartIdentifier.Identifiers) == 1 {
				opt = &ast.IdentifierPrincipalOption{
					OptionKind: convertUserOptionKind(optionName),
					Identifier: colRef.MultiPartIdentifier.Identifiers[0],
				}
			} else {
				opt = &ast.LiteralPrincipalOption{
					OptionKind: convertUserOptionKind(optionName),
					Value:      value,
				}
			}
			stmt.UserOptions = append(stmt.UserOptions, opt)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCreateAggregateStatement parses a CREATE AGGREGATE statement
func (p *Parser) parseCreateAggregateStatement() (*ast.CreateAggregateStatement, error) {
	// Consume AGGREGATE
	p.nextToken()

	stmt := &ast.CreateAggregateStatement{}

	// Parse aggregate name
	name, _ := p.parseSchemaObjectName()
	stmt.Name = name

	// Check for ( (optional for lenient parsing)
	if p.curTok.Type != TokenLParen {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse parameters
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param := &ast.ProcedureParameter{
			IsVarying: false,
			Modifier:  "None",
		}

		// Parse parameter name
		if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
			param.VariableName = &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
			}
			p.nextToken()
		} else {
			param.VariableName = p.parseIdentifier()
		}

		// Parse data type
		dataType, err := p.parseDataTypeReference()
		if err != nil {
			return nil, err
		}
		param.DataType = dataType

		stmt.Parameters = append(stmt.Parameters, param)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Expect )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	// Expect RETURNS
	if p.curTok.Type != TokenReturns {
		return nil, fmt.Errorf("expected RETURNS, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse return type
	returnType, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}
	stmt.ReturnType = returnType

	// Expect EXTERNAL NAME
	if strings.ToUpper(p.curTok.Literal) != "EXTERNAL" {
		return nil, fmt.Errorf("expected EXTERNAL, got %s", p.curTok.Literal)
	}
	p.nextToken()

	if strings.ToUpper(p.curTok.Literal) != "NAME" {
		return nil, fmt.Errorf("expected NAME after EXTERNAL, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse assembly name
	stmt.AssemblyName = &ast.AssemblyName{
		Name: p.parseIdentifier(),
	}

	// Check for .class.method syntax
	if p.curTok.Type == TokenDot {
		p.nextToken()
		stmt.AssemblyName.ClassName = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCreateColumnStoreIndexStatement parses a CREATE COLUMNSTORE INDEX statement
func (p *Parser) parseCreateColumnStoreIndexStatement() (*ast.CreateColumnStoreIndexStatement, error) {
	stmt := &ast.CreateColumnStoreIndexStatement{}

	// Parse CLUSTERED or NONCLUSTERED
	if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
		stmt.Clustered = true
		stmt.ClusteredExplicit = true
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
		stmt.Clustered = false
		stmt.ClusteredExplicit = true
		p.nextToken()
	}

	// Expect COLUMNSTORE
	if strings.ToUpper(p.curTok.Literal) != "COLUMNSTORE" {
		return nil, fmt.Errorf("expected COLUMNSTORE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect INDEX
	if p.curTok.Type != TokenIndex {
		return nil, fmt.Errorf("expected INDEX, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse index name
	stmt.Name = p.parseIdentifier()

	// Expect ON
	if p.curTok.Type != TokenOn {
		return nil, fmt.Errorf("expected ON, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse table name
	tableName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.OnName = tableName

	// Parse optional column list for non-clustered columnstore indexes
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			colRef := &ast.ColumnReferenceExpression{
				ColumnType: "Regular",
				MultiPartIdentifier: &ast.MultiPartIdentifier{
					Identifiers: []*ast.Identifier{p.parseIdentifier()},
				},
			}
			colRef.MultiPartIdentifier.Count = len(colRef.MultiPartIdentifier.Identifiers)
			stmt.Columns = append(stmt.Columns, colRef)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Parse optional ORDER clause
	if strings.ToUpper(p.curTok.Literal) == "ORDER" {
		p.nextToken() // consume ORDER
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				colRef := &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{p.parseIdentifier()},
					},
				}
				colRef.MultiPartIdentifier.Count = len(colRef.MultiPartIdentifier.Identifiers)
				stmt.OrderedColumns = append(stmt.OrderedColumns, colRef)

				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken()
			}
		}
	}

	// Skip optional WITH clause for now
	if p.curTok.Type == TokenWith {
		// TODO: parse WITH options
		p.nextToken()
		if p.curTok.Type == TokenLParen {
			p.nextToken()
			depth := 1
			for depth > 0 && p.curTok.Type != TokenEOF {
				if p.curTok.Type == TokenLParen {
					depth++
				} else if p.curTok.Type == TokenRParen {
					depth--
				}
				p.nextToken()
			}
		}
	}

	// Skip optional ON partition clause
	if p.curTok.Type == TokenOn {
		p.nextToken()
		// Skip to semicolon
		for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF {
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseAlterFunctionStatement parses an ALTER FUNCTION statement
func (p *Parser) parseAlterFunctionStatement() (*ast.AlterFunctionStatement, error) {
	// Consume FUNCTION
	p.nextToken()

	stmt := &ast.AlterFunctionStatement{}

	// Parse function name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Parse parameters in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			param := &ast.ProcedureParameter{
				IsVarying: false,
				Modifier:  "None",
			}

			// Parse parameter name
			if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				param.VariableName = &ast.Identifier{
					Value:     p.curTok.Literal,
					QuoteType: "NotQuoted",
				}
				p.nextToken()
			}

			// Parse data type if present
			if p.curTok.Type != TokenRParen && p.curTok.Type != TokenComma {
				dataType, err := p.parseDataType()
				if err != nil {
					return nil, err
				}
				param.DataType = dataType
			}

			stmt.Parameters = append(stmt.Parameters, param)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Expect RETURNS
	if p.curTok.Type != TokenReturns {
		return nil, fmt.Errorf("expected RETURNS, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse return type
	returnDataType, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	stmt.ReturnType = &ast.ScalarFunctionReturnType{
		DataType: returnDataType,
	}

	// Parse AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse statement list
	stmtList, err := p.parseFunctionStatementList()
	if err != nil {
		return nil, err
	}
	stmt.StatementList = stmtList

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseFunctionStatementList parses the body of a function
func (p *Parser) parseFunctionStatementList() (*ast.StatementList, error) {
	stmtList := &ast.StatementList{}

	for p.curTok.Type != TokenEOF {
		// Check for GO or end of batch
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmtList.Statements = append(stmtList.Statements, stmt)
		}

		// Stop after one statement for simple function bodies
		break
	}

	return stmtList, nil
}

// parseAlterTriggerStatement parses an ALTER TRIGGER statement
func (p *Parser) parseAlterTriggerStatement() (*ast.AlterTriggerStatement, error) {
	// Consume TRIGGER
	p.nextToken()

	stmt := &ast.AlterTriggerStatement{}

	// Parse trigger name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Expect ON
	if strings.ToUpper(p.curTok.Literal) != "ON" {
		return nil, fmt.Errorf("expected ON, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse trigger object
	triggerObject := &ast.TriggerObject{
		TriggerScope: "Normal",
	}

	// Check for ALL SERVER or DATABASE
	switch strings.ToUpper(p.curTok.Literal) {
	case "ALL":
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken()
			triggerObject.TriggerScope = "AllServer"
		}
	case "DATABASE":
		p.nextToken()
		triggerObject.TriggerScope = "Database"
	default:
		// Parse table/view name
		objName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		triggerObject.Name = objName
	}
	stmt.TriggerObject = triggerObject

	// Parse trigger type (FOR, AFTER, INSTEAD OF)
	switch strings.ToUpper(p.curTok.Literal) {
	case "FOR":
		stmt.TriggerType = "For"
		p.nextToken()
	case "AFTER":
		stmt.TriggerType = "After"
		p.nextToken()
	case "INSTEAD":
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "OF" {
			p.nextToken()
		}
		stmt.TriggerType = "InsteadOf"
	}

	// Parse trigger actions
	isDatabaseOrServerTrigger := triggerObject.TriggerScope == "Database" || triggerObject.TriggerScope == "AllServer"
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		action := &ast.TriggerAction{}
		actionType := strings.ToUpper(p.curTok.Literal)

		// Check for empty action type (lenient parsing for incomplete statements)
		if actionType == "" || p.curTok.Type == TokenAs {
			break
		}

		switch actionType {
		case "INSERT":
			action.TriggerActionType = "Insert"
		case "UPDATE":
			action.TriggerActionType = "Update"
		case "DELETE":
			action.TriggerActionType = "Delete"
		default:
			// For database/server triggers, events are wrapped in EventTypeContainer
			if isDatabaseOrServerTrigger && len(actionType) > 0 {
				action.TriggerActionType = "Event"
				// Convert action type to proper case (e.g., RENAME -> Rename)
				eventType := strings.ToUpper(actionType[:1]) + strings.ToLower(actionType[1:])
				action.EventTypeGroup = &ast.EventTypeContainer{
					EventType: eventType,
				}
			} else {
				action.TriggerActionType = actionType
			}
		}
		p.nextToken()

		stmt.TriggerActions = append(stmt.TriggerActions, action)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Parse AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse statement list
	stmtList := &ast.StatementList{}
	for p.curTok.Type != TokenEOF {
		// Check for GO or end of batch
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			break
		}

		innerStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if innerStmt != nil {
			stmtList.Statements = append(stmtList.Statements, innerStmt)
		}

		// For simple triggers, stop after parsing one statement
		break
	}
	stmt.StatementList = stmtList

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterIndexStatement() (*ast.AlterIndexStatement, error) {
	// Consume INDEX
	p.nextToken()

	stmt := &ast.AlterIndexStatement{}

	// Check for ALL or index name
	if strings.ToUpper(p.curTok.Literal) == "ALL" {
		stmt.All = true
		p.nextToken()
	} else {
		// Parse index name
		stmt.Name = p.parseIdentifier()
	}

	// Expect ON
	if strings.ToUpper(p.curTok.Literal) != "ON" {
		return nil, fmt.Errorf("expected ON, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse table name
	onName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.OnName = onName

	// Parse alter index type
	switch strings.ToUpper(p.curTok.Literal) {
	case "REBUILD":
		stmt.AlterIndexType = "Rebuild"
		p.nextToken()
	case "REORGANIZE":
		stmt.AlterIndexType = "Reorganize"
		p.nextToken()
	case "DISABLE":
		stmt.AlterIndexType = "Disable"
		p.nextToken()
	case "SET":
		stmt.AlterIndexType = "Set"
		p.nextToken()
	case "RESUME":
		stmt.AlterIndexType = "Resume"
		p.nextToken()
	case "PAUSE":
		stmt.AlterIndexType = "Pause"
		p.nextToken()
	case "ABORT":
		stmt.AlterIndexType = "Abort"
		p.nextToken()
	}

	// Parse PARTITION clause if present
	if strings.ToUpper(p.curTok.Literal) == "PARTITION" {
		p.nextToken()
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after PARTITION, got %s", p.curTok.Literal)
		}
		p.nextToken()

		stmt.Partition = &ast.PartitionSpecifier{}
		if strings.ToUpper(p.curTok.Literal) == "ALL" {
			stmt.Partition.All = true
			p.nextToken()
		} else {
			// Parse partition number
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Partition.Number = expr
		}
	}

	// Parse WITH clause if present
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken()

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			optionName := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			if p.curTok.Type == TokenEquals {
				p.nextToken()
				valueStr := strings.ToUpper(p.curTok.Literal)
				p.nextToken()

				// Determine if it's a state option (ON/OFF) or expression option
				if valueStr == "ON" || valueStr == "OFF" {
					opt := &ast.IndexStateOption{
						OptionKind:  p.getIndexOptionKind(optionName),
						OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
				} else {
					// Expression option like FILLFACTOR = 80
					opt := &ast.IndexExpressionOption{
						OptionKind: p.getIndexOptionKind(optionName),
						Expression: &ast.IntegerLiteral{Value: valueStr},
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
				}
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) getIndexOptionKind(optionName string) string {
	optionMap := map[string]string{
		"PAD_INDEX":             "PadIndex",
		"FILLFACTOR":            "FillFactor",
		"SORT_IN_TEMPDB":        "SortInTempDB",
		"IGNORE_DUP_KEY":        "IgnoreDupKey",
		"STATISTICS_NORECOMPUTE": "StatisticsNoRecompute",
		"DROP_EXISTING":         "DropExisting",
		"ONLINE":                "Online",
		"ALLOW_ROW_LOCKS":       "AllowRowLocks",
		"ALLOW_PAGE_LOCKS":      "AllowPageLocks",
		"MAXDOP":                "MaxDop",
		"DATA_COMPRESSION":      "DataCompression",
		"RESUMABLE":             "Resumable",
		"MAX_DURATION":          "MaxDuration",
		"WAIT_AT_LOW_PRIORITY":  "WaitAtLowPriority",
	}
	if kind, ok := optionMap[optionName]; ok {
		return kind
	}
	return optionName
}

func (p *Parser) capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// parseCreateFunctionStatement parses a CREATE FUNCTION statement
func (p *Parser) parseCreateFunctionStatement() (*ast.CreateFunctionStatement, error) {
	// Consume FUNCTION
	p.nextToken()

	stmt := &ast.CreateFunctionStatement{}

	// Parse function name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Parse parameters in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			param := &ast.ProcedureParameter{
				IsVarying: false,
				Modifier:  "None",
			}

			// Parse parameter name
			if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				param.VariableName = &ast.Identifier{
					Value:     p.curTok.Literal,
					QuoteType: "NotQuoted",
				}
				p.nextToken()
			}

			// Parse data type if present
			if p.curTok.Type != TokenRParen && p.curTok.Type != TokenComma {
				dataType, err := p.parseDataType()
				if err != nil {
					return nil, err
				}
				param.DataType = dataType
			}

			stmt.Parameters = append(stmt.Parameters, param)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Expect RETURNS - if not present, be lenient and skip
	if p.curTok.Type != TokenReturns {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse return type
	returnDataType, err := p.parseDataType()
	if err != nil {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	stmt.ReturnType = &ast.ScalarFunctionReturnType{
		DataType: returnDataType,
	}

	// Parse AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse statement list
	stmtList, err := p.parseFunctionStatementList()
	if err != nil {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	stmt.StatementList = stmtList

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCreateTriggerStatement parses a CREATE TRIGGER statement
func (p *Parser) parseCreateTriggerStatement() (*ast.CreateTriggerStatement, error) {
	// Consume TRIGGER
	p.nextToken()

	stmt := &ast.CreateTriggerStatement{}

	// Parse trigger name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Expect ON
	if strings.ToUpper(p.curTok.Literal) != "ON" {
		return nil, fmt.Errorf("expected ON, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse trigger object
	triggerObject := &ast.TriggerObject{
		TriggerScope: "Normal",
	}

	// Check for ALL SERVER or DATABASE
	switch strings.ToUpper(p.curTok.Literal) {
	case "ALL":
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken()
			triggerObject.TriggerScope = "AllServer"
		}
	case "DATABASE":
		p.nextToken()
		triggerObject.TriggerScope = "Database"
	default:
		// Parse table/view name
		objName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		triggerObject.Name = objName
	}
	stmt.TriggerObject = triggerObject

	// Parse trigger type (FOR, AFTER, INSTEAD OF)
	switch strings.ToUpper(p.curTok.Literal) {
	case "FOR":
		stmt.TriggerType = "For"
		p.nextToken()
	case "AFTER":
		stmt.TriggerType = "After"
		p.nextToken()
	case "INSTEAD":
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "OF" {
			p.nextToken()
		}
		stmt.TriggerType = "InsteadOf"
	}

	// Parse trigger actions
	isDatabaseOrServerTrigger := triggerObject.TriggerScope == "Database" || triggerObject.TriggerScope == "AllServer"
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		action := &ast.TriggerAction{}
		actionType := strings.ToUpper(p.curTok.Literal)

		// Check for empty action type (lenient parsing for incomplete statements)
		if actionType == "" || p.curTok.Type == TokenAs {
			break
		}

		switch actionType {
		case "INSERT":
			action.TriggerActionType = "Insert"
		case "UPDATE":
			action.TriggerActionType = "Update"
		case "DELETE":
			action.TriggerActionType = "Delete"
		default:
			// For database/server triggers, events are wrapped in EventTypeContainer
			if isDatabaseOrServerTrigger && len(actionType) > 0 {
				action.TriggerActionType = "Event"
				// Convert action type to proper case (e.g., RENAME -> Rename)
				eventType := strings.ToUpper(actionType[:1]) + strings.ToLower(actionType[1:])
				action.EventTypeGroup = &ast.EventTypeContainer{
					EventType: eventType,
				}
			} else {
				action.TriggerActionType = actionType
			}
		}
		p.nextToken()

		stmt.TriggerActions = append(stmt.TriggerActions, action)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Parse AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse statement list
	stmtList := &ast.StatementList{}
	for p.curTok.Type != TokenEOF {
		// Check for GO or end of batch
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			break
		}

		innerStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if innerStmt != nil {
			stmtList.Statements = append(stmtList.Statements, innerStmt)
		}

		// For simple triggers, stop after parsing one statement
		break
	}
	stmt.StatementList = stmtList

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// JSON marshaling functions for new statement types

func restoreStatementToJSON(s *ast.RestoreStatement) jsonNode {
	node := jsonNode{
		"$type": "RestoreStatement",
		"Kind":  s.Kind,
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierOrValueExpressionToJSON(s.DatabaseName)
	}
	if len(s.Devices) > 0 {
		devices := make([]jsonNode, len(s.Devices))
		for i, d := range s.Devices {
			devices[i] = deviceInfoToJSON(d)
		}
		node["Devices"] = devices
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = restoreOptionToJSON(o)
		}
		node["Options"] = options
	}
	return node
}

func backupDatabaseStatementToJSON(s *ast.BackupDatabaseStatement) jsonNode {
	node := jsonNode{
		"$type": "BackupDatabaseStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierOrValueExpressionToJSON(s.DatabaseName)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = backupOptionToJSON(o)
		}
		node["Options"] = options
	}
	if len(s.Devices) > 0 {
		devices := make([]jsonNode, len(s.Devices))
		for i, d := range s.Devices {
			devices[i] = deviceInfoToJSON(d)
		}
		node["Devices"] = devices
	}
	return node
}

func backupCertificateStatementToJSON(s *ast.BackupCertificateStatement) jsonNode {
	node := jsonNode{
		"$type": "BackupCertificateStatement",
	}
	if s.File != nil {
		node["File"] = scalarExpressionToJSON(s.File)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["ActiveForBeginDialog"] = s.ActiveForBeginDialog
	if s.PrivateKeyPath != nil {
		node["PrivateKeyPath"] = scalarExpressionToJSON(s.PrivateKeyPath)
	}
	if s.EncryptionPassword != nil {
		node["EncryptionPassword"] = scalarExpressionToJSON(s.EncryptionPassword)
	}
	if s.DecryptionPassword != nil {
		node["DecryptionPassword"] = scalarExpressionToJSON(s.DecryptionPassword)
	}
	return node
}

func backupOptionToJSON(o *ast.BackupOption) jsonNode {
	node := jsonNode{
		"$type":      "BackupOption",
		"OptionKind": o.OptionKind,
	}
	if o.Value != nil {
		node["Value"] = scalarExpressionToJSON(o.Value)
	}
	return node
}

func deviceInfoToJSON(d *ast.DeviceInfo) jsonNode {
	node := jsonNode{
		"$type":      "DeviceInfo",
		"DeviceType": d.DeviceType,
	}
	if d.LogicalDevice != nil {
		node["LogicalDevice"] = identifierOrValueExpressionToJSON(d.LogicalDevice)
	}
	if d.PhysicalDevice != nil {
		node["PhysicalDevice"] = identifierOrValueExpressionToJSON(d.PhysicalDevice)
	}
	return node
}

func restoreOptionToJSON(o ast.RestoreOption) jsonNode {
	switch opt := o.(type) {
	case *ast.FileStreamRestoreOption:
		node := jsonNode{
			"$type":      "FileStreamRestoreOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.FileStreamOption != nil {
			node["FileStreamOption"] = fileStreamDatabaseOptionToJSON(opt.FileStreamOption)
		}
		return node
	case *ast.GeneralSetCommandRestoreOption:
		node := jsonNode{
			"$type":      "GeneralSetCommandRestoreOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.OptionValue != nil {
			node["OptionValue"] = scalarExpressionToJSON(opt.OptionValue)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownRestoreOption"}
	}
}

func fileStreamDatabaseOptionToJSON(f *ast.FileStreamDatabaseOption) jsonNode {
	node := jsonNode{
		"$type":      "FileStreamDatabaseOption",
		"OptionKind": f.OptionKind,
	}
	if f.DirectoryName != nil {
		node["DirectoryName"] = scalarExpressionToJSON(f.DirectoryName)
	}
	return node
}

func createUserStatementToJSON(s *ast.CreateUserStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateUserStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.UserLoginOption != nil {
		node["UserLoginOption"] = userLoginOptionToJSON(s.UserLoginOption)
	}
	if len(s.UserOptions) > 0 {
		options := make([]jsonNode, len(s.UserOptions))
		for i, o := range s.UserOptions {
			options[i] = userOptionToJSON(o)
		}
		node["UserOptions"] = options
	}
	return node
}

func userLoginOptionToJSON(u *ast.UserLoginOption) jsonNode {
	node := jsonNode{
		"$type":               "UserLoginOption",
		"UserLoginOptionType": u.UserLoginOptionType,
	}
	if u.Identifier != nil {
		node["Identifier"] = identifierToJSON(u.Identifier)
	}
	return node
}

func userOptionToJSON(o ast.UserOption) jsonNode {
	switch opt := o.(type) {
	case *ast.LiteralPrincipalOption:
		node := jsonNode{
			"$type":      "LiteralPrincipalOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return node
	case *ast.IdentifierPrincipalOption:
		node := jsonNode{
			"$type":      "IdentifierPrincipalOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Identifier != nil {
			node["Identifier"] = identifierToJSON(opt.Identifier)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownUserOption"}
	}
}

func createAggregateStatementToJSON(s *ast.CreateAggregateStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateAggregateStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.AssemblyName != nil {
		node["AssemblyName"] = assemblyNameToJSON(s.AssemblyName)
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = procedureParameterToJSON(p)
		}
		node["Parameters"] = params
	}
	if s.ReturnType != nil {
		node["ReturnType"] = dataTypeReferenceToJSON(s.ReturnType)
	}
	return node
}

func assemblyNameToJSON(a *ast.AssemblyName) jsonNode {
	node := jsonNode{
		"$type": "AssemblyName",
	}
	if a.Name != nil {
		node["Name"] = identifierToJSON(a.Name)
	}
	if a.ClassName != nil {
		node["ClassName"] = identifierToJSON(a.ClassName)
	}
	return node
}

func createColumnStoreIndexStatementToJSON(s *ast.CreateColumnStoreIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateColumnStoreIndexStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Clustered || s.ClusteredExplicit {
		node["Clustered"] = s.Clustered
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if len(s.Columns) > 0 {
		cols := make([]jsonNode, len(s.Columns))
		for i, col := range s.Columns {
			cols[i] = columnReferenceExpressionToJSON(col)
		}
		node["Columns"] = cols
	}
	if len(s.OrderedColumns) > 0 {
		cols := make([]jsonNode, len(s.OrderedColumns))
		for i, col := range s.OrderedColumns {
			cols[i] = columnReferenceExpressionToJSON(col)
		}
		node["OrderedColumns"] = cols
	}
	return node
}

func alterFunctionStatementToJSON(s *ast.AlterFunctionStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterFunctionStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.ReturnType != nil {
		node["ReturnType"] = functionReturnTypeToJSON(s.ReturnType)
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func createFunctionStatementToJSON(s *ast.CreateFunctionStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateFunctionStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.ReturnType != nil {
		node["ReturnType"] = functionReturnTypeToJSON(s.ReturnType)
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func functionReturnTypeToJSON(r ast.FunctionReturnType) jsonNode {
	switch rt := r.(type) {
	case *ast.ScalarFunctionReturnType:
		node := jsonNode{
			"$type": "ScalarFunctionReturnType",
		}
		if rt.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(rt.DataType)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownFunctionReturnType"}
	}
}

func alterTriggerStatementToJSON(s *ast.AlterTriggerStatement) jsonNode {
	node := jsonNode{
		"$type":               "AlterTriggerStatement",
		"TriggerType":         s.TriggerType,
		"WithAppend":          s.WithAppend,
		"IsNotForReplication": s.IsNotForReplication,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.TriggerObject != nil {
		node["TriggerObject"] = triggerObjectToJSON(s.TriggerObject)
	}
	if len(s.TriggerActions) > 0 {
		actions := make([]jsonNode, len(s.TriggerActions))
		for i, a := range s.TriggerActions {
			actions[i] = triggerActionToJSON(a)
		}
		node["TriggerActions"] = actions
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func createTriggerStatementToJSON(s *ast.CreateTriggerStatement) jsonNode {
	node := jsonNode{
		"$type":               "CreateTriggerStatement",
		"TriggerType":         s.TriggerType,
		"WithAppend":          s.WithAppend,
		"IsNotForReplication": s.IsNotForReplication,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if s.TriggerObject != nil {
		node["TriggerObject"] = triggerObjectToJSON(s.TriggerObject)
	}
	if len(s.TriggerActions) > 0 {
		actions := make([]jsonNode, len(s.TriggerActions))
		for i, a := range s.TriggerActions {
			actions[i] = triggerActionToJSON(a)
		}
		node["TriggerActions"] = actions
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func triggerObjectToJSON(t *ast.TriggerObject) jsonNode {
	node := jsonNode{
		"$type":        "TriggerObject",
		"TriggerScope": t.TriggerScope,
	}
	if t.Name != nil {
		node["Name"] = schemaObjectNameToJSON(t.Name)
	}
	return node
}

func triggerActionToJSON(a *ast.TriggerAction) jsonNode {
	node := jsonNode{
		"$type":             "TriggerAction",
		"TriggerActionType": a.TriggerActionType,
	}
	if a.EventTypeGroup != nil {
		node["EventTypeGroup"] = jsonNode{
			"$type":     "EventTypeContainer",
			"EventType": a.EventTypeGroup.EventType,
		}
	}
	return node
}

func alterIndexStatementToJSON(s *ast.AlterIndexStatement) jsonNode {
	node := jsonNode{
		"$type":          "AlterIndexStatement",
		"All":            s.All,
		"AlterIndexType": s.AlterIndexType,
	}
	if s.Partition != nil {
		node["Partition"] = partitionSpecifierToJSON(s.Partition)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.IndexOptions) > 0 {
		opts := make([]jsonNode, len(s.IndexOptions))
		for i, opt := range s.IndexOptions {
			opts[i] = indexOptionToJSON(opt)
		}
		node["IndexOptions"] = opts
	}
	return node
}

func partitionSpecifierToJSON(p *ast.PartitionSpecifier) jsonNode {
	node := jsonNode{
		"$type": "PartitionSpecifier",
		"All":   p.All,
	}
	if p.Number != nil {
		node["Number"] = scalarExpressionToJSON(p.Number)
	}
	return node
}

func indexOptionToJSON(opt ast.IndexOption) jsonNode {
	switch o := opt.(type) {
	case *ast.IndexStateOption:
		return jsonNode{
			"$type":       "IndexStateOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
	case *ast.IndexExpressionOption:
		return jsonNode{
			"$type":      "IndexExpressionOption",
			"OptionKind": o.OptionKind,
			"Expression": scalarExpressionToJSON(o.Expression),
		}
	default:
		return jsonNode{"$type": "UnknownIndexOption"}
	}
}

func convertUserOptionKind(name string) string {
	// Convert option names to the expected format
	optionMap := map[string]string{
		"OBJECT_ID":      "Object_ID",
		"DEFAULT_SCHEMA": "Default_Schema",
		"SID":            "Sid",
		"PASSWORD":       "Password",
		"NAME":           "Name",
		"LOGIN":          "Login",
	}
	upper := strings.ToUpper(name)
	if mapped, ok := optionMap[upper]; ok {
		return mapped
	}
	// Default: return as-is with first letter capitalized
	return capitalizeFirst(name)
}

func dropDatabaseStatementToJSON(s *ast.DropDatabaseStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropDatabaseStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Databases) > 0 {
		dbs := make([]jsonNode, len(s.Databases))
		for i, db := range s.Databases {
			dbs[i] = identifierToJSON(db)
		}
		node["Databases"] = dbs
	}
	return node
}

func dropTableStatementToJSON(s *ast.DropTableStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropTableStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropViewStatementToJSON(s *ast.DropViewStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropViewStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropProcedureStatementToJSON(s *ast.DropProcedureStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropProcedureStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropFunctionStatementToJSON(s *ast.DropFunctionStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropFunctionStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropTriggerStatementToJSON(s *ast.DropTriggerStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropTriggerStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	if s.TriggerScope != "" {
		node["TriggerScope"] = s.TriggerScope
	}
	return node
}

func dropIndexStatementToJSON(s *ast.DropIndexStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropIndexStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Indexes) > 0 {
		clauses := make([]jsonNode, len(s.Indexes))
		for i, clause := range s.Indexes {
			clauses[i] = dropIndexClauseToJSON(clause)
		}
		node["DropIndexClauses"] = clauses
	}
	return node
}

func dropIndexClauseToJSON(c *ast.DropIndexClause) jsonNode {
	// If we have an Object (ON clause), use DropIndexClause type
	if c.Object != nil {
		node := jsonNode{
			"$type": "DropIndexClause",
		}
		if c.IndexName != nil {
			node["Index"] = identifierToJSON(c.IndexName)
		}
		node["Object"] = schemaObjectNameToJSON(c.Object)
		return node
	}

	// Otherwise use DropIndexClauseBase for backwards-compatible syntax
	node := jsonNode{
		"$type": "DropIndexClauseBase",
	}
	if c.Index != nil {
		node["Index"] = schemaObjectNameToJSON(c.Index)
	} else if c.IndexName != nil {
		// Just index name without object - use identifier
		node["Index"] = identifierToJSON(c.IndexName)
	}
	return node
}

func dropStatisticsStatementToJSON(s *ast.DropStatisticsStatement) jsonNode {
	node := jsonNode{
		"$type": "DropStatisticsStatement",
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropDefaultStatementToJSON(s *ast.DropDefaultStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropDefaultStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropRuleStatementToJSON(s *ast.DropRuleStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropRuleStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropSchemaStatementToJSON(s *ast.DropSchemaStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropSchemaStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Schema != nil {
		node["Schema"] = schemaObjectNameToJSON(s.Schema)
	}
	// DropBehavior defaults to "None"
	behavior := s.DropBehavior
	if behavior == "" {
		behavior = "None"
	}
	node["DropBehavior"] = behavior
	return node
}

func dropSecurityPolicyStatementToJSON(s *ast.DropSecurityPolicyStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropSecurityPolicyStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropExternalDataSourceStatementToJSON(s *ast.DropExternalDataSourceStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalDataSourceStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func dropExternalFileFormatStatementToJSON(s *ast.DropExternalFileFormatStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalFileFormatStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func dropExternalTableStatementToJSON(s *ast.DropExternalTableStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalTableStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropExternalResourcePoolStatementToJSON(s *ast.DropExternalResourcePoolStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalResourcePoolStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func dropWorkloadGroupStatementToJSON(s *ast.DropWorkloadGroupStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropWorkloadGroupStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func dropWorkloadClassifierStatementToJSON(s *ast.DropWorkloadClassifierStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropWorkloadClassifierStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func dropTypeStatementToJSON(s *ast.DropTypeStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropTypeStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return node
}

func dropAggregateStatementToJSON(s *ast.DropAggregateStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropAggregateStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropSynonymStatementToJSON(s *ast.DropSynonymStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropSynonymStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func dropUserStatementToJSON(s *ast.DropUserStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropUserStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func dropRoleStatementToJSON(s *ast.DropRoleStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropRoleStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func dropAssemblyStatementToJSON(s *ast.DropAssemblyStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropAssemblyStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return node
}

func alterTableTriggerModificationStatementToJSON(s *ast.AlterTableTriggerModificationStatement) jsonNode {
	node := jsonNode{
		"$type":              "AlterTableTriggerModificationStatement",
		"TriggerEnforcement": s.TriggerEnforcement,
		"All":                s.All,
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	if len(s.TriggerNames) > 0 {
		names := make([]jsonNode, len(s.TriggerNames))
		for i, name := range s.TriggerNames {
			names[i] = identifierToJSON(name)
		}
		node["TriggerNames"] = names
	}
	return node
}

func alterTableSwitchStatementToJSON(s *ast.AlterTableSwitchStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterTableSwitchStatement",
	}
	if s.TargetTable != nil {
		node["TargetTable"] = schemaObjectNameToJSON(s.TargetTable)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = tableSwitchOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	if s.SourcePartition != nil {
		node["SourcePartition"] = scalarExpressionToJSON(s.SourcePartition)
	}
	if s.TargetPartition != nil {
		node["TargetPartition"] = scalarExpressionToJSON(s.TargetPartition)
	}
	return node
}

func tableSwitchOptionToJSON(opt ast.TableSwitchOption) jsonNode {
	switch o := opt.(type) {
	case *ast.TruncateTargetTableSwitchOption:
		return jsonNode{
			"$type":          "TruncateTargetTableSwitchOption",
			"TruncateTarget": o.TruncateTarget,
			"OptionKind":     o.OptionKind,
		}
	default:
		return jsonNode{"$type": "UnknownSwitchOption"}
	}
}

func alterTableConstraintModificationStatementToJSON(s *ast.AlterTableConstraintModificationStatement) jsonNode {
	node := jsonNode{
		"$type":                        "AlterTableConstraintModificationStatement",
		"ExistingRowsCheckEnforcement": s.ExistingRowsCheckEnforcement,
		"ConstraintEnforcement":        s.ConstraintEnforcement,
		"All":                          s.All,
	}
	if len(s.ConstraintNames) > 0 {
		names := make([]jsonNode, len(s.ConstraintNames))
		for i, name := range s.ConstraintNames {
			names[i] = identifierToJSON(name)
		}
		node["ConstraintNames"] = names
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return node
}

func createExternalDataSourceStatementToJSON(s *ast.CreateExternalDataSourceStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalDataSourceStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createExternalFileFormatStatementToJSON(s *ast.CreateExternalFileFormatStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalFileFormatStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createExternalTableStatementToJSON(s *ast.CreateExternalTableStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalTableStatement",
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return node
}

func createExternalLanguageStatementToJSON(s *ast.CreateExternalLanguageStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalLanguageStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createExternalLibraryStatementToJSON(s *ast.CreateExternalLibraryStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalLibraryStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createEventSessionStatementToJSON(s *ast.CreateEventSessionStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateEventSessionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func insertBulkStatementToJSON(s *ast.InsertBulkStatement) jsonNode {
	node := jsonNode{
		"$type": "InsertBulkStatement",
	}
	if s.To != nil {
		node["To"] = schemaObjectNameToJSON(s.To)
	}
	if len(s.ColumnDefinitions) > 0 {
		colDefs := make([]jsonNode, len(s.ColumnDefinitions))
		for i, colDef := range s.ColumnDefinitions {
			colDefs[i] = insertBulkColumnDefinitionToJSON(colDef)
		}
		node["ColumnDefinitions"] = colDefs
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			options[i] = bulkInsertOptionToJSON(opt)
		}
		node["Options"] = options
	}
	return node
}

func insertBulkColumnDefinitionToJSON(c *ast.InsertBulkColumnDefinition) jsonNode {
	node := jsonNode{
		"$type": "InsertBulkColumnDefinition",
	}
	if c.Column != nil {
		node["Column"] = columnDefinitionBaseToJSON(c.Column)
	}
	if c.NullNotNull != "" && c.NullNotNull != "Unspecified" {
		node["NullNotNull"] = c.NullNotNull
	}
	return node
}

func columnDefinitionBaseToJSON(c *ast.ColumnDefinitionBase) jsonNode {
	node := jsonNode{
		"$type": "ColumnDefinitionBase",
	}
	if c.ColumnIdentifier != nil {
		node["ColumnIdentifier"] = identifierToJSON(c.ColumnIdentifier)
	}
	if c.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(c.DataType)
	}
	return node
}

func bulkInsertOptionToJSON(opt ast.BulkInsertOption) jsonNode {
	switch o := opt.(type) {
	case *ast.BulkInsertOptionBase:
		return jsonNode{
			"$type":      "BulkInsertOption",
			"OptionKind": o.OptionKind,
		}
	case *ast.LiteralBulkInsertOption:
		node := jsonNode{
			"$type":      "LiteralBulkInsertOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		return node
	case *ast.OrderBulkInsertOption:
		node := jsonNode{
			"$type":      "OrderBulkInsertOption",
			"OptionKind": "Order",
		}
		if len(o.Columns) > 0 {
			cols := make([]jsonNode, len(o.Columns))
			for i, col := range o.Columns {
				cols[i] = columnWithSortOrderToJSON(col)
			}
			node["Columns"] = cols
		}
		if o.IsUnique {
			node["IsUnique"] = o.IsUnique
		}
		return node
	default:
		return jsonNode{"$type": "UnknownBulkInsertOption"}
	}
}

func bulkInsertStatementToJSON(s *ast.BulkInsertStatement) jsonNode {
	node := jsonNode{
		"$type": "BulkInsertStatement",
	}
	if s.From != nil {
		node["From"] = identifierOrValueExpressionToJSON(s.From)
	}
	if s.To != nil {
		node["To"] = schemaObjectNameToJSON(s.To)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			options[i] = bulkInsertOptionToJSON(opt)
		}
		node["Options"] = options
	}
	return node
}

func alterUserStatementToJSON(s *ast.AlterUserStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterUserStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterRouteStatementToJSON(s *ast.AlterRouteStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterRouteStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterAssemblyStatementToJSON(s *ast.AlterAssemblyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterAssemblyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterEndpointStatementToJSON(s *ast.AlterEndpointStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterEndpointStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterServiceStatementToJSON(s *ast.AlterServiceStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServiceStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterCertificateStatementToJSON(s *ast.AlterCertificateStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterCertificateStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterApplicationRoleStatementToJSON(s *ast.AlterApplicationRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterApplicationRoleStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterAsymmetricKeyStatementToJSON(s *ast.AlterAsymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterAsymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterQueueStatementToJSON(s *ast.AlterQueueStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterQueueStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return node
}

func alterPartitionSchemeStatementToJSON(s *ast.AlterPartitionSchemeStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterPartitionSchemeStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterPartitionFunctionStatementToJSON(s *ast.AlterPartitionFunctionStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterPartitionFunctionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterFulltextCatalogStatementToJSON(s *ast.AlterFulltextCatalogStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterFulltextCatalogStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterFulltextIndexStatementToJSON(s *ast.AlterFulltextIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterFulltextIndexStatement",
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	return node
}

func alterSymmetricKeyStatementToJSON(s *ast.AlterSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterSymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterServiceMasterKeyStatementToJSON(s *ast.AlterServiceMasterKeyStatement) jsonNode {
	return jsonNode{
		"$type": "AlterServiceMasterKeyStatement",
	}
}

func createDatabaseStatementToJSON(s *ast.CreateDatabaseStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateDatabaseStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	return node
}

func createLoginStatementToJSON(s *ast.CreateLoginStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateLoginStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createIndexStatementToJSON(s *ast.CreateIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateIndexStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	return node
}

func createAsymmetricKeyStatementToJSON(s *ast.CreateAsymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateAsymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createSymmetricKeyStatementToJSON(s *ast.CreateSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createCertificateStatementToJSON(s *ast.CreateCertificateStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateCertificateStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createMessageTypeStatementToJSON(s *ast.CreateMessageTypeStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateMessageTypeStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createServiceStatementToJSON(s *ast.CreateServiceStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateServiceStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createQueueStatementToJSON(s *ast.CreateQueueStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateQueueStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return node
}

func createRouteStatementToJSON(s *ast.CreateRouteStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateRouteStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createEndpointStatementToJSON(s *ast.CreateEndpointStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateEndpointStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createAssemblyStatementToJSON(s *ast.CreateAssemblyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateAssemblyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createApplicationRoleStatementToJSON(s *ast.CreateApplicationRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateApplicationRoleStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createFulltextCatalogStatementToJSON(s *ast.CreateFulltextCatalogStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateFulltextCatalogStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createFulltextIndexStatementToJSON(s *ast.CreateFulltextIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateFulltextIndexStatement",
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	return node
}

func createRemoteServiceBindingStatementToJSON(s *ast.CreateRemoteServiceBindingStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateRemoteServiceBindingStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createStatisticsStatementToJSON(s *ast.CreateStatisticsStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateStatisticsStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	return node
}

func createTypeStatementToJSON(s *ast.CreateTypeStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateTypeStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return node
}

func createXmlIndexStatementToJSON(s *ast.CreateXmlIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateXmlIndexStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	return node
}

func createPartitionFunctionStatementToJSON(s *ast.CreatePartitionFunctionStatement) jsonNode {
	node := jsonNode{
		"$type": "CreatePartitionFunctionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createEventNotificationStatementToJSON(s *ast.CreateEventNotificationStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateEventNotificationStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterDatabaseAddFileStatementToJSON(s *ast.AlterDatabaseAddFileStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseAddFileStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	return node
}

func alterDatabaseAddFileGroupStatementToJSON(s *ast.AlterDatabaseAddFileGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseAddFileGroupStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.FileGroupName != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroupName)
	}
	return node
}

func alterDatabaseModifyFileStatementToJSON(s *ast.AlterDatabaseModifyFileStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseModifyFileStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	return node
}

func alterDatabaseModifyFileGroupStatementToJSON(s *ast.AlterDatabaseModifyFileGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseModifyFileGroupStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.FileGroupName != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroupName)
	}
	return node
}

func alterDatabaseModifyNameStatementToJSON(s *ast.AlterDatabaseModifyNameStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseModifyNameStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.NewName != nil {
		node["NewDatabaseName"] = identifierToJSON(s.NewName)
	}
	return node
}

func alterDatabaseRemoveFileStatementToJSON(s *ast.AlterDatabaseRemoveFileStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseRemoveFileStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.FileName != nil {
		node["File"] = identifierToJSON(s.FileName)
	}
	return node
}

func alterDatabaseRemoveFileGroupStatementToJSON(s *ast.AlterDatabaseRemoveFileGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseRemoveFileGroupStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.FileGroupName != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroupName)
	}
	return node
}

