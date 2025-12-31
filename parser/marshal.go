// Package parser provides T-SQL parsing functionality.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sqlc-dev/teesql/ast"
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
	case *ast.UpdateStatisticsStatement:
		return updateStatisticsStatementToJSON(s)
	case *ast.DeleteStatement:
		return deleteStatementToJSON(s)
	case *ast.DeclareVariableStatement:
		return declareVariableStatementToJSON(s)
	case *ast.DeclareTableVariableStatement:
		return declareTableVariableStatementToJSON(s)
	case *ast.SetVariableStatement:
		return setVariableStatementToJSON(s)
	case *ast.IfStatement:
		return ifStatementToJSON(s)
	case *ast.WhileStatement:
		return whileStatementToJSON(s)
	case *ast.BeginEndBlockStatement:
		return beginEndBlockStatementToJSON(s)
	case *ast.BeginEndAtomicBlockStatement:
		return beginEndAtomicBlockStatementToJSON(s)
	case *ast.CreateViewStatement:
		return createViewStatementToJSON(s)
	case *ast.CreateSchemaStatement:
		return createSchemaStatementToJSON(s)
	case *ast.CreateProcedureStatement:
		return createProcedureStatementToJSON(s)
	case *ast.AlterProcedureStatement:
		return alterProcedureStatementToJSON(s)
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
	case *ast.AlterDatabaseScopedConfigurationClearStatement:
		return alterDatabaseScopedConfigurationClearStatementToJSON(s)
	case *ast.AlterResourceGovernorStatement:
		return alterResourceGovernorStatementToJSON(s)
	case *ast.CreateCryptographicProviderStatement:
		return createCryptographicProviderStatementToJSON(s)
	case *ast.AlterCryptographicProviderStatement:
		return alterCryptographicProviderStatementToJSON(s)
	case *ast.DropCryptographicProviderStatement:
		return dropCryptographicProviderStatementToJSON(s)
	case *ast.UseFederationStatement:
		return useFederationStatementToJSON(s)
	case *ast.CreateFederationStatement:
		return createFederationStatementToJSON(s)
	case *ast.AlterFederationStatement:
		return alterFederationStatementToJSON(s)
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
	case *ast.AlterExternalDataSourceStatement:
		return alterExternalDataSourceStatementToJSON(s)
	case *ast.AlterExternalLanguageStatement:
		return alterExternalLanguageStatementToJSON(s)
	case *ast.AlterExternalLibraryStatement:
		return alterExternalLibraryStatementToJSON(s)
	case *ast.DropExternalFileFormatStatement:
		return dropExternalFileFormatStatementToJSON(s)
	case *ast.DropExternalTableStatement:
		return dropExternalTableStatementToJSON(s)
	case *ast.DropExternalResourcePoolStatement:
		return dropExternalResourcePoolStatementToJSON(s)
	case *ast.DropExternalModelStatement:
		return dropExternalModelStatementToJSON(s)
	case *ast.DropWorkloadGroupStatement:
		return dropWorkloadGroupStatementToJSON(s)
	case *ast.DropWorkloadClassifierStatement:
		return dropWorkloadClassifierStatementToJSON(s)
	case *ast.CreateWorkloadGroupStatement:
		return createWorkloadGroupStatementToJSON(s)
	case *ast.CreateWorkloadClassifierStatement:
		return createWorkloadClassifierStatementToJSON(s)
	case *ast.AlterWorkloadGroupStatement:
		return alterWorkloadGroupStatementToJSON(s)
	case *ast.AlterSequenceStatement:
		return alterSequenceStatementToJSON(s)
	case *ast.CreateSequenceStatement:
		return createSequenceStatementToJSON(s)
	case *ast.DbccStatement:
		return dbccStatementToJSON(s)
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
	case *ast.RevokeStatement:
		return revokeStatementToJSON(s)
	case *ast.DenyStatement:
		return denyStatementToJSON(s)
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
	case *ast.CreateServerRoleStatement:
		return createServerRoleStatementToJSON(s)
	case *ast.AlterServerRoleStatement:
		return alterServerRoleStatementToJSON(s)
	case *ast.CreateServerAuditStatement:
		return createServerAuditStatementToJSON(s)
	case *ast.AlterServerAuditStatement:
		return alterServerAuditStatementToJSON(s)
	case *ast.AlterRemoteServiceBindingStatement:
		return alterRemoteServiceBindingStatementToJSON(s)
	case *ast.AlterXmlSchemaCollectionStatement:
		return alterXmlSchemaCollectionStatementToJSON(s)
	case *ast.AlterServerConfigurationSetSoftNumaStatement:
		return alterServerConfigurationSetSoftNumaStatementToJSON(s)
	case *ast.AlterServerConfigurationSetExternalAuthenticationStatement:
		return alterServerConfigurationSetExternalAuthenticationStatementToJSON(s)
	case *ast.AlterServerConfigurationStatement:
		return alterServerConfigurationStatementToJSON(s)
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
	case *ast.BackupTransactionLogStatement:
		return backupTransactionLogStatementToJSON(s)
	case *ast.BackupCertificateStatement:
		return backupCertificateStatementToJSON(s)
	case *ast.BackupServiceMasterKeyStatement:
		return backupServiceMasterKeyStatementToJSON(s)
	case *ast.RestoreServiceMasterKeyStatement:
		return restoreServiceMasterKeyStatementToJSON(s)
	case *ast.RestoreMasterKeyStatement:
		return restoreMasterKeyStatementToJSON(s)
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
	case *ast.EnableDisableTriggerStatement:
		return enableDisableTriggerStatementToJSON(s)
	case *ast.CreateDatabaseStatement:
		return createDatabaseStatementToJSON(s)
	case *ast.CreateLoginStatement:
		return createLoginStatementToJSON(s)
	case *ast.CreateIndexStatement:
		return createIndexStatementToJSON(s)
	case *ast.CreateSpatialIndexStatement:
		return createSpatialIndexStatementToJSON(s)
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
	case *ast.CreateTypeUddtStatement:
		return createTypeUddtStatementToJSON(s)
	case *ast.CreateTypeUdtStatement:
		return createTypeUdtStatementToJSON(s)
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
	case *ast.AlterTableSetStatement:
		return alterTableSetStatementToJSON(s)
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
	case *ast.CreateFullTextCatalogStatement:
		return createFullTextCatalogStatementToJSON(s)
	case *ast.AlterFulltextIndexStatement:
		return alterFulltextIndexStatementToJSON(s)
	case *ast.AlterSymmetricKeyStatement:
		return alterSymmetricKeyStatementToJSON(s)
	case *ast.AlterServiceMasterKeyStatement:
		return alterServiceMasterKeyStatementToJSON(s)
	case *ast.RenameEntityStatement:
		return renameEntityStatementToJSON(s)
	case *ast.OpenCursorStatement:
		return openCursorStatementToJSON(s)
	case *ast.CloseCursorStatement:
		return closeCursorStatementToJSON(s)
	case *ast.DeallocateCursorStatement:
		return deallocateCursorStatementToJSON(s)
	case *ast.FetchCursorStatement:
		return fetchCursorStatementToJSON(s)
	case *ast.DeclareCursorStatement:
		return declareCursorStatementToJSON(s)
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
	if len(e.DropClusteredConstraintOptions) > 0 {
		options := make([]jsonNode, len(e.DropClusteredConstraintOptions))
		for i, o := range e.DropClusteredConstraintOptions {
			options[i] = dropClusteredConstraintOptionToJSON(o)
		}
		node["DropClusteredConstraintOptions"] = options
	}
	node["IsIfExists"] = e.IsIfExists
	return node
}

func dropClusteredConstraintOptionToJSON(o ast.DropClusteredConstraintOption) jsonNode {
	switch opt := o.(type) {
	case *ast.DropClusteredConstraintStateOption:
		return jsonNode{
			"$type":       "DropClusteredConstraintStateOption",
			"OptionState": opt.OptionState,
			"OptionKind":  opt.OptionKind,
		}
	case *ast.DropClusteredConstraintMoveOption:
		node := jsonNode{
			"$type":      "DropClusteredConstraintMoveOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.OptionValue != nil {
			node["OptionValue"] = fileGroupOrPartitionSchemeToJSON(opt.OptionValue)
		}
		return node
	case *ast.DropClusteredConstraintValueOption:
		node := jsonNode{
			"$type":      "DropClusteredConstraintValueOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.OptionValue != nil {
			node["OptionValue"] = scalarExpressionToJSON(opt.OptionValue)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownDropClusteredConstraintOption"}
	}
}

func fileGroupOrPartitionSchemeToJSON(fg *ast.FileGroupOrPartitionScheme) jsonNode {
	node := jsonNode{
		"$type": "FileGroupOrPartitionScheme",
	}
	if fg.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(fg.Name)
	}
	if len(fg.PartitionSchemeColumns) > 0 {
		cols := make([]jsonNode, len(fg.PartitionSchemeColumns))
		for i, c := range fg.PartitionSchemeColumns {
			cols[i] = identifierToJSON(c)
		}
		node["PartitionSchemeColumns"] = cols
	}
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
	case *ast.DelayedDurabilityDatabaseOption:
		return jsonNode{
			"$type":      "DelayedDurabilityDatabaseOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
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
	if len(idx.IncludeColumns) > 0 {
		cols := make([]jsonNode, len(idx.IncludeColumns))
		for i, c := range idx.IncludeColumns {
			cols[i] = scalarExpressionToJSON(c)
		}
		node["IncludeColumns"] = cols
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
	case *ast.OptimizeForOptimizerHint:
		node := jsonNode{
			"$type": "OptimizeForOptimizerHint",
		}
		if len(hint.Pairs) > 0 {
			pairs := make([]jsonNode, len(hint.Pairs))
			for i, pair := range hint.Pairs {
				pairs[i] = variableValuePairToJSON(pair)
			}
			node["Pairs"] = pairs
		}
		node["IsForUnknown"] = hint.IsForUnknown
		if hint.HintKind != "" {
			node["HintKind"] = hint.HintKind
		}
		return node
	default:
		return jsonNode{"$type": "UnknownOptimizerHint"}
	}
}

func variableValuePairToJSON(p *ast.VariableValuePair) jsonNode {
	node := jsonNode{
		"$type": "VariableValuePair",
	}
	if p.Variable != nil {
		node["Variable"] = scalarExpressionToJSON(p.Variable)
	}
	if p.Value != nil {
		node["Value"] = scalarExpressionToJSON(p.Value)
	}
	node["IsForUnknown"] = p.IsForUnknown
	return node
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
	case *ast.SelectSetVariable:
		node := jsonNode{
			"$type": "SelectSetVariable",
		}
		if e.Variable != nil {
			varNode := jsonNode{
				"$type": "VariableReference",
			}
			if e.Variable.Name != "" {
				varNode["Name"] = e.Variable.Name
			}
			node["Variable"] = varNode
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		if e.AssignmentKind != "" {
			node["AssignmentKind"] = e.AssignmentKind
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
		// Always include Value for StringLiteral, even if empty
		node["Value"] = e.Value
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
		if e.CallTarget != nil {
			node["CallTarget"] = callTargetToJSON(e.CallTarget)
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
		if e.OverClause != nil {
			node["OverClause"] = jsonNode{
				"$type": "OverClause",
			}
		}
		node["WithArrayWrapper"] = e.WithArrayWrapper
		return node
	case *ast.UserDefinedTypePropertyAccess:
		node := jsonNode{
			"$type": "UserDefinedTypePropertyAccess",
		}
		if e.CallTarget != nil {
			node["CallTarget"] = callTargetToJSON(e.CallTarget)
		}
		if e.PropertyName != nil {
			node["PropertyName"] = identifierToJSON(e.PropertyName)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return node
	case *ast.CastCall:
		node := jsonNode{
			"$type": "CastCall",
		}
		if e.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(e.DataType)
		}
		if e.Parameter != nil {
			node["Parameter"] = scalarExpressionToJSON(e.Parameter)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return node
	case *ast.ConvertCall:
		node := jsonNode{
			"$type": "ConvertCall",
		}
		if e.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(e.DataType)
		}
		if e.Parameter != nil {
			node["Parameter"] = scalarExpressionToJSON(e.Parameter)
		}
		if e.Style != nil {
			node["Style"] = scalarExpressionToJSON(e.Style)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return node
	case *ast.TryCastCall:
		node := jsonNode{
			"$type": "TryCastCall",
		}
		if e.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(e.DataType)
		}
		if e.Parameter != nil {
			node["Parameter"] = scalarExpressionToJSON(e.Parameter)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return node
	case *ast.TryConvertCall:
		node := jsonNode{
			"$type": "TryConvertCall",
		}
		if e.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(e.DataType)
		}
		if e.Parameter != nil {
			node["Parameter"] = scalarExpressionToJSON(e.Parameter)
		}
		if e.Style != nil {
			node["Style"] = scalarExpressionToJSON(e.Style)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
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
	case *ast.MaxLiteral:
		node := jsonNode{
			"$type": "MaxLiteral",
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
	case *ast.SourceDeclaration:
		node := jsonNode{
			"$type": "SourceDeclaration",
		}
		if e.Value != nil {
			node["Value"] = eventSessionObjectNameToJSON(e.Value)
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

func eventSessionObjectNameToJSON(e *ast.EventSessionObjectName) jsonNode {
	node := jsonNode{
		"$type": "EventSessionObjectName",
	}
	if e.MultiPartIdentifier != nil {
		node["MultiPartIdentifier"] = multiPartIdentifierToJSON(e.MultiPartIdentifier)
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
	case *ast.BulkOpenRowset:
		node := jsonNode{
			"$type": "BulkOpenRowset",
		}
		if len(r.DataFiles) > 0 {
			files := make([]jsonNode, len(r.DataFiles))
			for i, f := range r.DataFiles {
				files[i] = scalarExpressionToJSON(f)
			}
			node["DataFiles"] = files
		}
		if len(r.Options) > 0 {
			opts := make([]jsonNode, len(r.Options))
			for i, o := range r.Options {
				opts[i] = bulkInsertOptionToJSON(o)
			}
			node["Options"] = opts
		}
		if len(r.Columns) > 0 {
			cols := make([]jsonNode, len(r.Columns))
			for i, c := range r.Columns {
				cols[i] = identifierToJSON(c)
			}
			node["Columns"] = cols
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
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
			"$type": "LikePredicate",
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
		node["OdbcEscape"] = e.OdbcEscape
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
	if spec.LinkedServer != nil {
		node["LinkedServer"] = identifierToJSON(spec.LinkedServer)
	}
	if spec.ExecuteContext != nil {
		node["ExecuteContext"] = executeContextToJSON(spec.ExecuteContext)
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
	case *ast.ExecutableStringList:
		node := jsonNode{
			"$type": "ExecutableStringList",
		}
		if len(e.Strings) > 0 {
			strs := make([]jsonNode, len(e.Strings))
			for i, s := range e.Strings {
				strs[i] = scalarExpressionToJSON(s)
			}
			node["Strings"] = strs
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

func declareTableVariableStatementToJSON(s *ast.DeclareTableVariableStatement) jsonNode {
	node := jsonNode{
		"$type": "DeclareTableVariableStatement",
	}
	if s.Body != nil {
		node["Body"] = declareTableVariableBodyToJSON(s.Body)
	}
	return node
}

func declareTableVariableBodyToJSON(body *ast.DeclareTableVariableBody) jsonNode {
	node := jsonNode{
		"$type": "DeclareTableVariableBody",
	}
	if body.VariableName != nil {
		node["VariableName"] = identifierToJSON(body.VariableName)
	}
	node["AsDefined"] = body.AsDefined
	if body.Definition != nil {
		node["Definition"] = tableDefinitionToJSON(body.Definition)
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
	if s.SeparatorType != "" {
		node["SeparatorType"] = s.SeparatorType
	} else {
		node["SeparatorType"] = "NotSpecified"
	}
	if s.Identifier != nil {
		node["Identifier"] = identifierToJSON(s.Identifier)
	}
	node["FunctionCallExists"] = s.FunctionCallExists
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = scalarExpressionToJSON(p)
		}
		node["Parameters"] = params
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
	return node
}

func cursorDefinitionToJSON(cd *ast.CursorDefinition) jsonNode {
	node := jsonNode{
		"$type": "CursorDefinition",
	}
	if len(cd.Options) > 0 {
		opts := make([]jsonNode, len(cd.Options))
		for i, opt := range cd.Options {
			opts[i] = jsonNode{
				"$type":      "CursorOption",
				"OptionKind": opt.OptionKind,
			}
		}
		node["Options"] = opts
	}
	if cd.Select != nil {
		node["Select"] = selectStatementToJSON(cd.Select)
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

func beginEndAtomicBlockStatementToJSON(s *ast.BeginEndAtomicBlockStatement) jsonNode {
	node := jsonNode{
		"$type": "BeginEndAtomicBlockStatement",
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = atomicBlockOptionToJSON(o)
		}
		node["Options"] = options
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return node
}

func atomicBlockOptionToJSON(o ast.AtomicBlockOption) jsonNode {
	switch opt := o.(type) {
	case *ast.IdentifierAtomicBlockOption:
		node := jsonNode{
			"$type":      "IdentifierAtomicBlockOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = identifierToJSON(opt.Value)
		}
		return node
	case *ast.LiteralAtomicBlockOption:
		node := jsonNode{
			"$type":      "LiteralAtomicBlockOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownAtomicBlockOption"}
	}
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

	// Parse column definitions and table constraints
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		upperLit := strings.ToUpper(p.curTok.Literal)

		// Check for table-level constraints
		if upperLit == "CONSTRAINT" {
			constraint, err := p.parseNamedTableConstraint()
			if err != nil {
				p.skipToEndOfStatement()
				return stmt, nil
			}
			if constraint != nil {
				stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
			}
		} else if upperLit == "PRIMARY" || upperLit == "UNIQUE" || upperLit == "FOREIGN" || upperLit == "CHECK" {
			constraint, err := p.parseUnnamedTableConstraint()
			if err != nil {
				p.skipToEndOfStatement()
				return stmt, nil
			}
			if constraint != nil {
				stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
			}
		} else {
			// Parse column definition
			colDef, err := p.parseColumnDefinition()
			if err != nil {
				p.skipToEndOfStatement()
				return stmt, nil
			}
			stmt.Definition.ColumnDefinitions = append(stmt.Definition.ColumnDefinitions, colDef)
		}

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

	// Parse optional ON filegroup and TEXTIMAGE_ON filegroup clauses
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if p.curTok.Type == TokenOn {
			p.nextToken() // consume ON
			// Parse filegroup identifier
			ident := p.parseIdentifier()
			stmt.OnFileGroupOrPartitionScheme = &ast.FileGroupOrPartitionScheme{
				Name: &ast.IdentifierOrValueExpression{
					Value:      ident.Value,
					Identifier: ident,
				},
			}
		} else if upperLit == "TEXTIMAGE_ON" {
			p.nextToken() // consume TEXTIMAGE_ON
			// Parse filegroup identifier
			ident := p.parseIdentifier()
			stmt.TextImageOn = &ast.IdentifierOrValueExpression{
				Value:      ident.Value,
				Identifier: ident,
			}
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

func (p *Parser) parseColumnDefinition() (*ast.ColumnDefinition, error) {
	col := &ast.ColumnDefinition{}

	// Parse column name (parseIdentifier already calls nextToken)
	col.ColumnIdentifier = p.parseIdentifier()

	// Parse data type - be lenient if no data type is provided
	dataType, err := p.parseDataTypeReference()
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
			} else if p.curTok.Type == TokenNull {
				// NOT NULL after IDENTITY - handle it here since NOT was already consumed
				p.nextToken() // consume NULL
				col.Constraints = append(col.Constraints, &ast.NullableConstraintDefinition{Nullable: false})
			}
		}

		col.IdentityOptions = identityOpts
	}

	// Parse column constraints (NULL, NOT NULL, UNIQUE, PRIMARY KEY, DEFAULT, CHECK, CONSTRAINT)
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)

		if p.curTok.Type == TokenNot {
			p.nextToken() // consume NOT
			if p.curTok.Type == TokenNull {
				p.nextToken() // consume NULL
				col.Constraints = append(col.Constraints, &ast.NullableConstraintDefinition{Nullable: false})
			}
		} else if p.curTok.Type == TokenNull {
			p.nextToken() // consume NULL
			col.Constraints = append(col.Constraints, &ast.NullableConstraintDefinition{Nullable: true})
		} else if upperLit == "UNIQUE" {
			p.nextToken() // consume UNIQUE
			constraint := &ast.UniqueConstraintDefinition{
				IsPrimaryKey: false,
			}
			// Parse optional CLUSTERED/NONCLUSTERED
			if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
				constraint.Clustered = true
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
				constraint.Clustered = false
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
				p.nextToken()
			}
			col.Constraints = append(col.Constraints, constraint)
		} else if upperLit == "PRIMARY" {
			p.nextToken() // consume PRIMARY
			if p.curTok.Type == TokenKey {
				p.nextToken() // consume KEY
			}
			constraint := &ast.UniqueConstraintDefinition{
				IsPrimaryKey: true,
			}
			// Parse optional CLUSTERED/NONCLUSTERED
			if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
				constraint.Clustered = true
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
				constraint.Clustered = false
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
				p.nextToken()
			}
			col.Constraints = append(col.Constraints, constraint)
		} else if p.curTok.Type == TokenDefault {
			p.nextToken() // consume DEFAULT
			defaultConstraint := &ast.DefaultConstraintDefinition{}

			// Parse the default expression
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			defaultConstraint.Expression = expr
			col.DefaultConstraint = defaultConstraint
		} else if upperLit == "CHECK" {
			p.nextToken() // consume CHECK
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				cond, err := p.parseBooleanExpression()
				if err != nil {
					return nil, err
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
				col.Constraints = append(col.Constraints, &ast.CheckConstraintDefinition{
					CheckCondition: cond,
				})
			}
		} else if upperLit == "CONSTRAINT" {
			p.nextToken() // skip CONSTRAINT
			if p.curTok.Type == TokenIdent {
				p.nextToken() // skip constraint name
			}
			// Continue to parse actual constraint in next iteration
			continue
		} else {
			break
		}
	}

	return col, nil
}

// parseNamedTableConstraint parses a CONSTRAINT name ... table constraint
func (p *Parser) parseNamedTableConstraint() (ast.TableConstraint, error) {
	// Consume CONSTRAINT
	p.nextToken()

	// Parse constraint name
	constraintName := p.parseIdentifier()

	// Now parse the actual constraint type
	upperLit := strings.ToUpper(p.curTok.Literal)

	if upperLit == "PRIMARY" {
		constraint, err := p.parsePrimaryKeyConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return constraint, nil
	} else if upperLit == "UNIQUE" {
		constraint, err := p.parseUniqueConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return constraint, nil
	} else if upperLit == "FOREIGN" {
		constraint, err := p.parseForeignKeyConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return constraint, nil
	} else if upperLit == "CHECK" {
		constraint, err := p.parseCheckConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return constraint, nil
	}

	return nil, nil
}

// parseUnnamedTableConstraint parses an unnamed table constraint (PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK)
func (p *Parser) parseUnnamedTableConstraint() (ast.TableConstraint, error) {
	upperLit := strings.ToUpper(p.curTok.Literal)

	if upperLit == "PRIMARY" {
		return p.parsePrimaryKeyConstraint()
	} else if upperLit == "UNIQUE" {
		return p.parseUniqueConstraint()
	} else if upperLit == "FOREIGN" {
		return p.parseForeignKeyConstraint()
	} else if upperLit == "CHECK" {
		return p.parseCheckConstraint()
	}

	return nil, nil
}

// parsePrimaryKeyConstraint parses PRIMARY KEY CLUSTERED/NONCLUSTERED (columns)
func (p *Parser) parsePrimaryKeyConstraint() (*ast.UniqueConstraintDefinition, error) {
	// Consume PRIMARY
	p.nextToken()
	if p.curTok.Type == TokenKey {
		p.nextToken() // consume KEY
	}

	constraint := &ast.UniqueConstraintDefinition{
		IsPrimaryKey: true,
	}

	// Parse optional CLUSTERED/NONCLUSTERED
	if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
		constraint.Clustered = true
		constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
		constraint.Clustered = false
		constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
		p.nextToken()
	}

	// Parse column list
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			col := p.parseColumnWithSortOrder()
			constraint.Columns = append(constraint.Columns, col)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	return constraint, nil
}

// parseUniqueConstraint parses UNIQUE CLUSTERED/NONCLUSTERED (columns)
func (p *Parser) parseUniqueConstraint() (*ast.UniqueConstraintDefinition, error) {
	// Consume UNIQUE
	p.nextToken()

	constraint := &ast.UniqueConstraintDefinition{
		IsPrimaryKey: false,
	}

	// Parse optional CLUSTERED/NONCLUSTERED
	if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
		constraint.Clustered = true
		constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
		constraint.Clustered = false
		constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
		p.nextToken()
	}

	// Parse column list
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			col := p.parseColumnWithSortOrder()
			constraint.Columns = append(constraint.Columns, col)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	return constraint, nil
}

// parseForeignKeyConstraint parses FOREIGN KEY (columns) REFERENCES table (columns)
func (p *Parser) parseForeignKeyConstraint() (*ast.ForeignKeyConstraintDefinition, error) {
	// Consume FOREIGN
	p.nextToken()
	if p.curTok.Type == TokenKey {
		p.nextToken() // consume KEY
	}

	constraint := &ast.ForeignKeyConstraintDefinition{}

	// Parse column list
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			ident := p.parseIdentifier()
			constraint.Columns = append(constraint.Columns, ident)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse REFERENCES
	if strings.ToUpper(p.curTok.Literal) == "REFERENCES" {
		p.nextToken() // consume REFERENCES

		// Parse reference table name
		refTable, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		constraint.ReferenceTableName = refTable

		// Parse referenced column list
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				ident := p.parseIdentifier()
				constraint.ReferencedColumns = append(constraint.ReferencedColumns, ident)

				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	return constraint, nil
}

// parseCheckConstraint parses CHECK (expression)
func (p *Parser) parseCheckConstraint() (*ast.CheckConstraintDefinition, error) {
	// Consume CHECK
	p.nextToken()

	constraint := &ast.CheckConstraintDefinition{}

	// Parse condition
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		cond, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		constraint.CheckCondition = cond
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	return constraint, nil
}

// parseColumnWithSortOrder parses a column name with optional ASC/DESC sort order
func (p *Parser) parseColumnWithSortOrder() *ast.ColumnWithSortOrder {
	col := &ast.ColumnWithSortOrder{
		SortOrder: ast.SortOrderNotSpecified,
	}

	// Parse column name
	ident := p.parseIdentifier()
	col.Column = &ast.ColumnReferenceExpression{
		ColumnType: "Regular",
		MultiPartIdentifier: &ast.MultiPartIdentifier{
			Count:       1,
			Identifiers: []*ast.Identifier{ident},
		},
	}

	// Parse optional ASC/DESC
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "ASC" {
		col.SortOrder = ast.SortOrderAscending
		p.nextToken()
	} else if upperLit == "DESC" {
		col.SortOrder = ast.SortOrderDescending
		p.nextToken()
	}

	return col
}

func (p *Parser) parseGrantStatement() (*ast.GrantStatement, error) {
	// Consume GRANT
	p.nextToken()

	stmt := &ast.GrantStatement{}

	// Parse permission(s)
	perm := &ast.Permission{}
	for p.curTok.Type != TokenTo && p.curTok.Type != TokenOn && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenCreate ||
			p.curTok.Type == TokenProcedure || p.curTok.Type == TokenView ||
			p.curTok.Type == TokenSelect || p.curTok.Type == TokenInsert ||
			p.curTok.Type == TokenUpdate || p.curTok.Type == TokenDelete ||
			p.curTok.Type == TokenAlter || p.curTok.Type == TokenExecute ||
			p.curTok.Type == TokenDrop || p.curTok.Type == TokenExternal {
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

	// Check for ON clause (SecurityTargetObject)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON

		stmt.SecurityTargetObject = &ast.SecurityTargetObject{}
		stmt.SecurityTargetObject.ObjectKind = "NotSpecified"

		// Parse object kind and ::
		// Object kinds can be: SERVER ROLE, APPLICATION ROLE, ASYMMETRIC KEY, SYMMETRIC KEY, etc.
		objectKind := strings.ToUpper(p.curTok.Literal)
		switch objectKind {
		case "SERVER":
			p.nextToken() // consume SERVER
			if strings.ToUpper(p.curTok.Literal) == "ROLE" {
				p.nextToken() // consume ROLE
				stmt.SecurityTargetObject.ObjectKind = "ServerRole"
			} else {
				stmt.SecurityTargetObject.ObjectKind = "Server"
			}
		case "APPLICATION":
			p.nextToken() // consume APPLICATION
			if strings.ToUpper(p.curTok.Literal) == "ROLE" {
				p.nextToken() // consume ROLE
			}
			stmt.SecurityTargetObject.ObjectKind = "ApplicationRole"
		case "ASYMMETRIC":
			p.nextToken() // consume ASYMMETRIC
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			stmt.SecurityTargetObject.ObjectKind = "AsymmetricKey"
		case "SYMMETRIC":
			p.nextToken() // consume SYMMETRIC
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			stmt.SecurityTargetObject.ObjectKind = "SymmetricKey"
		case "REMOTE":
			p.nextToken() // consume REMOTE
			if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
				p.nextToken() // consume SERVICE
				if strings.ToUpper(p.curTok.Literal) == "BINDING" {
					p.nextToken() // consume BINDING
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "RemoteServiceBinding"
		case "FULLTEXT":
			p.nextToken() // consume FULLTEXT
			if strings.ToUpper(p.curTok.Literal) == "CATALOG" {
				p.nextToken() // consume CATALOG
			}
			stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
		case "MESSAGE":
			p.nextToken() // consume MESSAGE
			if strings.ToUpper(p.curTok.Literal) == "TYPE" {
				p.nextToken() // consume TYPE
			}
			stmt.SecurityTargetObject.ObjectKind = "MessageType"
		case "XML":
			p.nextToken() // consume XML
			if strings.ToUpper(p.curTok.Literal) == "SCHEMA" {
				p.nextToken() // consume SCHEMA
				if strings.ToUpper(p.curTok.Literal) == "COLLECTION" {
					p.nextToken() // consume COLLECTION
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "XmlSchemaCollection"
		case "SEARCH":
			p.nextToken() // consume SEARCH
			if strings.ToUpper(p.curTok.Literal) == "PROPERTY" {
				p.nextToken() // consume PROPERTY
				if strings.ToUpper(p.curTok.Literal) == "LIST" {
					p.nextToken() // consume LIST
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "SearchPropertyList"
		case "AVAILABILITY":
			p.nextToken() // consume AVAILABILITY
			if strings.ToUpper(p.curTok.Literal) == "GROUP" {
				p.nextToken() // consume GROUP
			}
			stmt.SecurityTargetObject.ObjectKind = "AvailabilityGroup"
		case "TYPE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Type"
		case "OBJECT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Object"
		case "ASSEMBLY":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Assembly"
		case "CERTIFICATE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Certificate"
		case "CONTRACT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Contract"
		case "DATABASE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Database"
		case "ENDPOINT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Endpoint"
		case "LOGIN":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Login"
		case "ROLE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Role"
		case "ROUTE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Route"
		case "SCHEMA":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Schema"
		case "SERVICE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Service"
		case "USER":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "User"
		}

		// Expect ::
		if p.curTok.Type == TokenColonColon {
			p.nextToken() // consume ::

			// Parse object name as multi-part identifier
			stmt.SecurityTargetObject.ObjectName = &ast.SecurityTargetObjectName{}
			multiPart := &ast.MultiPartIdentifier{}
			for {
				id := p.parseIdentifier()
				multiPart.Identifiers = append(multiPart.Identifiers, id)
				if p.curTok.Type == TokenDot {
					p.nextToken() // consume .
				} else {
					break
				}
			}
			multiPart.Count = len(multiPart.Identifiers)
			stmt.SecurityTargetObject.ObjectName.MultiPartIdentifier = multiPart
		}
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

func (p *Parser) parseRevokeStatement() (*ast.RevokeStatement, error) {
	// Consume REVOKE
	p.nextToken()

	stmt := &ast.RevokeStatement{}

	// Check for GRANT OPTION FOR
	if strings.ToUpper(p.curTok.Literal) == "GRANT" {
		p.nextToken() // consume GRANT
		if strings.ToUpper(p.curTok.Literal) == "OPTION" {
			p.nextToken() // consume OPTION
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
			}
		}
		stmt.GrantOptionFor = true
	}

	// Parse permission(s)
	perm := &ast.Permission{}
	for p.curTok.Type != TokenTo && p.curTok.Type != TokenOn && p.curTok.Type != TokenEOF && strings.ToUpper(p.curTok.Literal) != "FROM" {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenCreate ||
			p.curTok.Type == TokenProcedure || p.curTok.Type == TokenView ||
			p.curTok.Type == TokenSelect || p.curTok.Type == TokenInsert ||
			p.curTok.Type == TokenUpdate || p.curTok.Type == TokenDelete ||
			p.curTok.Type == TokenAlter || p.curTok.Type == TokenExecute ||
			p.curTok.Type == TokenDrop || p.curTok.Type == TokenExternal {
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

	// Check for ON clause (SecurityTargetObject)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON

		stmt.SecurityTargetObject = &ast.SecurityTargetObject{}
		stmt.SecurityTargetObject.ObjectKind = "NotSpecified"

		// Parse object kind and ::
		objectKind := strings.ToUpper(p.curTok.Literal)
		switch objectKind {
		case "SERVER":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "ROLE" {
				p.nextToken()
				stmt.SecurityTargetObject.ObjectKind = "ServerRole"
			} else {
				stmt.SecurityTargetObject.ObjectKind = "Server"
			}
		case "APPLICATION":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "ROLE" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "ApplicationRole"
		case "ASYMMETRIC":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "AsymmetricKey"
		case "SYMMETRIC":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "SymmetricKey"
		case "REMOTE":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "BINDING" {
					p.nextToken()
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "RemoteServiceBinding"
		case "FULLTEXT":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "CATALOG" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
		case "MESSAGE":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "TYPE" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "MessageType"
		case "XML":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "SCHEMA" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "COLLECTION" {
					p.nextToken()
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "XmlSchemaCollection"
		case "SEARCH":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "PROPERTY" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "LIST" {
					p.nextToken()
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "SearchPropertyList"
		case "AVAILABILITY":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "GROUP" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "AvailabilityGroup"
		case "TYPE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Type"
		case "OBJECT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Object"
		case "ASSEMBLY":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Assembly"
		case "CERTIFICATE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Certificate"
		case "CONTRACT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Contract"
		case "DATABASE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Database"
		case "ENDPOINT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Endpoint"
		case "LOGIN":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Login"
		case "ROLE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Role"
		case "ROUTE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Route"
		case "SCHEMA":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Schema"
		case "SERVICE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Service"
		case "USER":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "User"
		}

		// Expect ::
		if p.curTok.Type == TokenColonColon {
			p.nextToken() // consume ::

			// Parse object name as multi-part identifier
			stmt.SecurityTargetObject.ObjectName = &ast.SecurityTargetObjectName{}
			multiPart := &ast.MultiPartIdentifier{}
			for {
				id := p.parseIdentifier()
				multiPart.Identifiers = append(multiPart.Identifiers, id)
				if p.curTok.Type == TokenDot {
					p.nextToken() // consume .
				} else {
					break
				}
			}
			multiPart.Count = len(multiPart.Identifiers)
			stmt.SecurityTargetObject.ObjectName.MultiPartIdentifier = multiPart
		}
	}

	// Expect TO or FROM
	if p.curTok.Type == TokenTo || strings.ToUpper(p.curTok.Literal) == "FROM" {
		p.nextToken()
	}

	// Parse principal(s)
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && strings.ToUpper(p.curTok.Literal) != "CASCADE" {
		principal := &ast.SecurityPrincipal{}
		if p.curTok.Type == TokenPublic {
			principal.PrincipalType = "Public"
			p.nextToken()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			principal.PrincipalType = "Identifier"
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

	// Check for CASCADE
	if strings.ToUpper(p.curTok.Literal) == "CASCADE" {
		stmt.CascadeOption = true
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDenyStatement() (*ast.DenyStatement, error) {
	// Consume DENY
	p.nextToken()

	stmt := &ast.DenyStatement{}

	// Parse permission(s)
	perm := &ast.Permission{}
	for p.curTok.Type != TokenTo && p.curTok.Type != TokenOn && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenCreate ||
			p.curTok.Type == TokenProcedure || p.curTok.Type == TokenView ||
			p.curTok.Type == TokenSelect || p.curTok.Type == TokenInsert ||
			p.curTok.Type == TokenUpdate || p.curTok.Type == TokenDelete ||
			p.curTok.Type == TokenAlter || p.curTok.Type == TokenExecute ||
			p.curTok.Type == TokenDrop || p.curTok.Type == TokenExternal {
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

	// Check for ON clause (SecurityTargetObject)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON

		stmt.SecurityTargetObject = &ast.SecurityTargetObject{}
		stmt.SecurityTargetObject.ObjectKind = "NotSpecified"

		// Parse object kind and ::
		objectKind := strings.ToUpper(p.curTok.Literal)
		switch objectKind {
		case "SERVER":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "ROLE" {
				p.nextToken()
				stmt.SecurityTargetObject.ObjectKind = "ServerRole"
			} else {
				stmt.SecurityTargetObject.ObjectKind = "Server"
			}
		case "APPLICATION":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "ROLE" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "ApplicationRole"
		case "ASYMMETRIC":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "AsymmetricKey"
		case "SYMMETRIC":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "SymmetricKey"
		case "REMOTE":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "BINDING" {
					p.nextToken()
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "RemoteServiceBinding"
		case "FULLTEXT":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "CATALOG" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
		case "MESSAGE":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "TYPE" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "MessageType"
		case "XML":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "SCHEMA" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "COLLECTION" {
					p.nextToken()
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "XmlSchemaCollection"
		case "SEARCH":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "PROPERTY" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "LIST" {
					p.nextToken()
				}
			}
			stmt.SecurityTargetObject.ObjectKind = "SearchPropertyList"
		case "AVAILABILITY":
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "GROUP" {
				p.nextToken()
			}
			stmt.SecurityTargetObject.ObjectKind = "AvailabilityGroup"
		case "TYPE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Type"
		case "OBJECT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Object"
		case "ASSEMBLY":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Assembly"
		case "CERTIFICATE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Certificate"
		case "CONTRACT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Contract"
		case "DATABASE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Database"
		case "ENDPOINT":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Endpoint"
		case "LOGIN":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Login"
		case "ROLE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Role"
		case "ROUTE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Route"
		case "SCHEMA":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Schema"
		case "SERVICE":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "Service"
		case "USER":
			p.nextToken()
			stmt.SecurityTargetObject.ObjectKind = "User"
		}

		// Expect ::
		if p.curTok.Type == TokenColonColon {
			p.nextToken() // consume ::

			// Parse object name as multi-part identifier
			stmt.SecurityTargetObject.ObjectName = &ast.SecurityTargetObjectName{}
			multiPart := &ast.MultiPartIdentifier{}
			for {
				id := p.parseIdentifier()
				multiPart.Identifiers = append(multiPart.Identifiers, id)
				if p.curTok.Type == TokenDot {
					p.nextToken() // consume .
				} else {
					break
				}
			}
			multiPart.Count = len(multiPart.Identifiers)
			stmt.SecurityTargetObject.ObjectName.MultiPartIdentifier = multiPart
		}
	}

	// Expect TO
	if p.curTok.Type == TokenTo {
		p.nextToken()
	}

	// Parse principal(s)
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && strings.ToUpper(p.curTok.Literal) != "CASCADE" {
		principal := &ast.SecurityPrincipal{}
		if p.curTok.Type == TokenPublic {
			principal.PrincipalType = "Public"
			p.nextToken()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			principal.PrincipalType = "Identifier"
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

	// Check for CASCADE
	if strings.ToUpper(p.curTok.Literal) == "CASCADE" {
		stmt.CascadeOption = true
		p.nextToken()
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
	if s.OnFileGroupOrPartitionScheme != nil {
		node["OnFileGroupOrPartitionScheme"] = fileGroupOrPartitionSchemeToJSON(s.OnFileGroupOrPartitionScheme)
	}
	if s.TextImageOn != nil {
		node["TextImageOn"] = identifierOrValueExpressionToJSON(s.TextImageOn)
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
	if len(t.TableConstraints) > 0 {
		constraints := make([]jsonNode, len(t.TableConstraints))
		for i, constraint := range t.TableConstraints {
			constraints[i] = tableConstraintToJSON(constraint)
		}
		node["TableConstraints"] = constraints
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

func tableConstraintToJSON(c ast.TableConstraint) jsonNode {
	switch constraint := c.(type) {
	case *ast.UniqueConstraintDefinition:
		return uniqueConstraintToJSON(constraint)
	case *ast.CheckConstraintDefinition:
		return checkConstraintToJSON(constraint)
	case *ast.ForeignKeyConstraintDefinition:
		return foreignKeyConstraintToJSON(constraint)
	default:
		return jsonNode{"$type": "UnknownTableConstraint"}
	}
}

func foreignKeyConstraintToJSON(c *ast.ForeignKeyConstraintDefinition) jsonNode {
	node := jsonNode{
		"$type":             "ForeignKeyConstraintDefinition",
		"NotForReplication": c.NotForReplication,
	}
	if c.ConstraintIdentifier != nil {
		node["ConstraintIdentifier"] = identifierToJSON(c.ConstraintIdentifier)
	}
	if c.ReferenceTableName != nil {
		node["ReferenceTableName"] = schemaObjectNameToJSON(c.ReferenceTableName)
	}
	if len(c.Columns) > 0 {
		cols := make([]jsonNode, len(c.Columns))
		for i, col := range c.Columns {
			cols[i] = identifierToJSON(col)
		}
		node["Columns"] = cols
	}
	if len(c.ReferencedColumns) > 0 {
		cols := make([]jsonNode, len(c.ReferencedColumns))
		for i, col := range c.ReferencedColumns {
			cols[i] = identifierToJSON(col)
		}
		node["ReferencedTableColumns"] = cols
	}
	// Always include DeleteAction and UpdateAction with default value
	deleteAction := c.DeleteAction
	if deleteAction == "" {
		deleteAction = "NotSpecified"
	}
	node["DeleteAction"] = deleteAction
	updateAction := c.UpdateAction
	if updateAction == "" {
		updateAction = "NotSpecified"
	}
	node["UpdateAction"] = updateAction
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
	if c.DefaultConstraint != nil {
		node["DefaultConstraint"] = defaultConstraintToJSON(c.DefaultConstraint)
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

func defaultConstraintToJSON(d *ast.DefaultConstraintDefinition) jsonNode {
	node := jsonNode{
		"$type":      "DefaultConstraintDefinition",
		"WithValues": false,
	}
	if d.ConstraintIdentifier != nil {
		node["ConstraintIdentifier"] = identifierToJSON(d.ConstraintIdentifier)
	}
	if d.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(d.Expression)
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
	case *ast.UniqueConstraintDefinition:
		return uniqueConstraintToJSON(constraint)
	case *ast.CheckConstraintDefinition:
		return checkConstraintToJSON(constraint)
	default:
		return jsonNode{"$type": "UnknownConstraint"}
	}
}

func uniqueConstraintToJSON(c *ast.UniqueConstraintDefinition) jsonNode {
	node := jsonNode{
		"$type":        "UniqueConstraintDefinition",
		"Clustered":    c.Clustered,
		"IsPrimaryKey": c.IsPrimaryKey,
	}
	if c.ConstraintIdentifier != nil {
		node["ConstraintIdentifier"] = identifierToJSON(c.ConstraintIdentifier)
	}
	if c.IndexType != nil {
		node["IndexType"] = indexTypeToJSON(c.IndexType)
	}
	if len(c.Columns) > 0 {
		cols := make([]jsonNode, len(c.Columns))
		for i, col := range c.Columns {
			cols[i] = columnWithSortOrderToJSON(col)
		}
		node["Columns"] = cols
	}
	return node
}

func checkConstraintToJSON(c *ast.CheckConstraintDefinition) jsonNode {
	node := jsonNode{
		"$type":             "CheckConstraintDefinition",
		"NotForReplication": c.NotForReplication,
	}
	if c.ConstraintIdentifier != nil {
		node["ConstraintIdentifier"] = identifierToJSON(c.ConstraintIdentifier)
	}
	if c.CheckCondition != nil {
		node["CheckCondition"] = booleanExpressionToJSON(c.CheckCondition)
	}
	return node
}

func dataTypeReferenceToJSON(d ast.DataTypeReference) jsonNode {
	switch dt := d.(type) {
	case *ast.SqlDataTypeReference:
		return sqlDataTypeReferenceToJSON(dt)
	case *ast.XmlDataTypeReference:
		return xmlDataTypeReferenceToJSON(dt)
	case *ast.UserDataTypeReference:
		return userDataTypeReferenceToJSON(dt)
	default:
		return jsonNode{"$type": "UnknownDataType"}
	}
}

func userDataTypeReferenceToJSON(dt *ast.UserDataTypeReference) jsonNode {
	node := jsonNode{
		"$type": "UserDataTypeReference",
	}
	if dt.Name != nil {
		node["Name"] = schemaObjectNameToJSON(dt.Name)
	}
	return node
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
	if s.SecurityTargetObject != nil {
		node["SecurityTargetObject"] = securityTargetObjectToJSON(s.SecurityTargetObject)
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

func revokeStatementToJSON(s *ast.RevokeStatement) jsonNode {
	node := jsonNode{
		"$type":          "RevokeStatement",
		"GrantOptionFor": s.GrantOptionFor,
		"CascadeOption":  s.CascadeOption,
	}
	if len(s.Permissions) > 0 {
		perms := make([]jsonNode, len(s.Permissions))
		for i, p := range s.Permissions {
			perms[i] = permissionToJSON(p)
		}
		node["Permissions"] = perms
	}
	if s.SecurityTargetObject != nil {
		node["SecurityTargetObject"] = securityTargetObjectToJSON(s.SecurityTargetObject)
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

func denyStatementToJSON(s *ast.DenyStatement) jsonNode {
	node := jsonNode{
		"$type":         "DenyStatement",
		"CascadeOption": s.CascadeOption,
	}
	if len(s.Permissions) > 0 {
		perms := make([]jsonNode, len(s.Permissions))
		for i, p := range s.Permissions {
			perms[i] = permissionToJSON(p)
		}
		node["Permissions"] = perms
	}
	if s.SecurityTargetObject != nil {
		node["SecurityTargetObject"] = securityTargetObjectToJSON(s.SecurityTargetObject)
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

func securityTargetObjectToJSON(s *ast.SecurityTargetObject) jsonNode {
	node := jsonNode{
		"$type":      "SecurityTargetObject",
		"ObjectKind": s.ObjectKind,
	}
	if s.ObjectName != nil {
		node["ObjectName"] = securityTargetObjectNameToJSON(s.ObjectName)
	}
	return node
}

func securityTargetObjectNameToJSON(s *ast.SecurityTargetObjectName) jsonNode {
	node := jsonNode{
		"$type": "SecurityTargetObjectName",
	}
	if s.MultiPartIdentifier != nil {
		node["MultiPartIdentifier"] = multiPartIdentifierToJSON(s.MultiPartIdentifier)
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

func createServerRoleStatementToJSON(s *ast.CreateServerRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateServerRoleStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterServerRoleStatementToJSON(s *ast.AlterServerRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerRoleStatement",
	}
	if s.Action != nil {
		node["Action"] = alterRoleActionToJSON(s.Action)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func createServerAuditStatementToJSON(s *ast.CreateServerAuditStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateServerAuditStatement",
	}
	if s.AuditName != nil {
		node["AuditName"] = identifierToJSON(s.AuditName)
	}
	if s.AuditTarget != nil {
		node["AuditTarget"] = auditTargetToJSON(s.AuditTarget)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = auditOptionToJSON(o)
		}
		node["Options"] = options
	}
	if s.PredicateExpression != nil {
		node["PredicateExpression"] = booleanExpressionToJSON(s.PredicateExpression)
	}
	return node
}

func alterServerAuditStatementToJSON(s *ast.AlterServerAuditStatement) jsonNode {
	node := jsonNode{
		"$type":       "AlterServerAuditStatement",
		"RemoveWhere": s.RemoveWhere,
	}
	if s.AuditName != nil {
		node["AuditName"] = identifierToJSON(s.AuditName)
	}
	if s.AuditTarget != nil {
		node["AuditTarget"] = auditTargetToJSON(s.AuditTarget)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = auditOptionToJSON(o)
		}
		node["Options"] = options
	}
	if s.PredicateExpression != nil {
		node["PredicateExpression"] = booleanExpressionToJSON(s.PredicateExpression)
	}
	return node
}

func auditTargetToJSON(t *ast.AuditTarget) jsonNode {
	node := jsonNode{
		"$type":      "AuditTarget",
		"TargetKind": t.TargetKind,
	}
	if len(t.TargetOptions) > 0 {
		opts := make([]jsonNode, len(t.TargetOptions))
		for i, o := range t.TargetOptions {
			opts[i] = auditTargetOptionToJSON(o)
		}
		node["TargetOptions"] = opts
	}
	return node
}

func auditTargetOptionToJSON(o ast.AuditTargetOption) jsonNode {
	switch opt := o.(type) {
	case *ast.LiteralAuditTargetOption:
		node := jsonNode{
			"$type":      "LiteralAuditTargetOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownAuditTargetOption"}
	}
}

func auditOptionToJSON(o ast.AuditOption) jsonNode {
	switch opt := o.(type) {
	case *ast.OnFailureAuditOption:
		return jsonNode{
			"$type":           "OnFailureAuditOption",
			"OnFailureAction": opt.OnFailureAction,
			"OptionKind":      opt.OptionKind,
		}
	case *ast.QueueDelayAuditOption:
		node := jsonNode{
			"$type":      "QueueDelayAuditOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Delay != nil {
			node["Delay"] = scalarExpressionToJSON(opt.Delay)
		}
		return node
	case *ast.StateAuditOption:
		return jsonNode{
			"$type":      "StateAuditOption",
			"Value":      opt.Value,
			"OptionKind": opt.OptionKind,
		}
	case *ast.AuditGuidAuditOption:
		node := jsonNode{
			"$type":      "AuditGuidAuditOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Guid != nil {
			node["Guid"] = scalarExpressionToJSON(opt.Guid)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownAuditOption"}
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

func alterServerConfigurationSetExternalAuthenticationStatementToJSON(s *ast.AlterServerConfigurationSetExternalAuthenticationStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationSetExternalAuthenticationStatement",
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = alterServerConfigurationExternalAuthenticationContainerOptionToJSON(o)
		}
		node["Options"] = options
	}
	return node
}

func alterServerConfigurationExternalAuthenticationContainerOptionToJSON(o *ast.AlterServerConfigurationExternalAuthenticationContainerOption) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationExternalAuthenticationContainerOption",
	}
	if len(o.Suboptions) > 0 {
		suboptions := make([]jsonNode, len(o.Suboptions))
		for i, s := range o.Suboptions {
			suboptions[i] = alterServerConfigurationExternalAuthenticationOptionToJSON(s)
		}
		node["Suboptions"] = suboptions
	}
	node["OptionKind"] = o.OptionKind
	if o.OptionValue != nil {
		node["OptionValue"] = onOffOptionValueToJSON(o.OptionValue)
	}
	return node
}

func alterServerConfigurationExternalAuthenticationOptionToJSON(o *ast.AlterServerConfigurationExternalAuthenticationOption) jsonNode {
	node := jsonNode{
		"$type":      "AlterServerConfigurationExternalAuthenticationOption",
		"OptionKind": o.OptionKind,
	}
	if o.OptionValue != nil {
		node["OptionValue"] = literalOptionValueToJSON(o.OptionValue)
	}
	return node
}

func literalOptionValueToJSON(o *ast.LiteralOptionValue) jsonNode {
	node := jsonNode{
		"$type": "LiteralOptionValue",
	}
	if o.Value != nil {
		node["Value"] = scalarExpressionToJSON(o.Value)
	}
	return node
}

func alterServerConfigurationStatementToJSON(s *ast.AlterServerConfigurationStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationStatement",
	}
	if s.ProcessAffinity != "" {
		node["ProcessAffinity"] = s.ProcessAffinity
	}
	if len(s.ProcessAffinityRanges) > 0 {
		ranges := make([]jsonNode, len(s.ProcessAffinityRanges))
		for i, r := range s.ProcessAffinityRanges {
			ranges[i] = processAffinityRangeToJSON(r)
		}
		node["ProcessAffinityRanges"] = ranges
	}
	return node
}

func processAffinityRangeToJSON(r *ast.ProcessAffinityRange) jsonNode {
	node := jsonNode{
		"$type": "ProcessAffinityRange",
	}
	if r.From != nil {
		node["From"] = scalarExpressionToJSON(r.From)
	}
	if r.To != nil {
		node["To"] = scalarExpressionToJSON(r.To)
	}
	return node
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

// parseRestoreStatement parses a RESTORE statement
func (p *Parser) parseRestoreStatement() (ast.Statement, error) {
	// Consume RESTORE
	p.nextToken()

	// Check for SERVICE MASTER KEY
	if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
		return p.parseRestoreServiceMasterKeyStatement()
	}

	// Check for MASTER KEY
	if strings.ToUpper(p.curTok.Literal) == "MASTER" {
		return p.parseRestoreMasterKeyStatement()
	}

	stmt := &ast.RestoreStatement{}

	// Parse restore kind (DATABASE, LOG, etc.)
	switch strings.ToUpper(p.curTok.Literal) {
	case "DATABASE":
		stmt.Kind = "Database"
		p.nextToken()
	case "LOG":
		stmt.Kind = "TransactionLog"
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

	// Check for optional FROM clause
	if strings.ToUpper(p.curTok.Literal) != "FROM" {
		// No FROM clause - just the database name with no devices
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
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

func (p *Parser) parseRestoreServiceMasterKeyStatement() (*ast.RestoreServiceMasterKeyStatement, error) {
	// Consume SERVICE
	p.nextToken()

	// Expect MASTER
	if strings.ToUpper(p.curTok.Literal) != "MASTER" {
		return nil, fmt.Errorf("expected MASTER after SERVICE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect KEY
	if p.curTok.Type != TokenKey {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.RestoreServiceMasterKeyStatement{}

	// Expect FROM
	if strings.ToUpper(p.curTok.Literal) != "FROM" {
		return nil, fmt.Errorf("expected FROM after SERVICE MASTER KEY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect FILE
	if strings.ToUpper(p.curTok.Literal) != "FILE" {
		return nil, fmt.Errorf("expected FILE after FROM, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after FILE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse file path
	file, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.File = file

	// Parse DECRYPTION BY PASSWORD clause
	if strings.ToUpper(p.curTok.Literal) == "DECRYPTION" {
		p.nextToken() // consume DECRYPTION
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
			p.nextToken() // consume PASSWORD
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			pwd, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Password = pwd
		}
	}

	// Check for FORCE
	if strings.ToUpper(p.curTok.Literal) == "FORCE" {
		stmt.IsForce = true
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseRestoreMasterKeyStatement() (*ast.RestoreMasterKeyStatement, error) {
	// Consume MASTER
	p.nextToken()

	// Expect KEY
	if p.curTok.Type != TokenKey {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.RestoreMasterKeyStatement{}

	// Expect FROM
	if strings.ToUpper(p.curTok.Literal) != "FROM" {
		return nil, fmt.Errorf("expected FROM after MASTER KEY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect FILE
	if strings.ToUpper(p.curTok.Literal) != "FILE" {
		return nil, fmt.Errorf("expected FILE after FROM, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after FILE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse file path
	file, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.File = file

	// Parse DECRYPTION BY PASSWORD clause
	if strings.ToUpper(p.curTok.Literal) == "DECRYPTION" {
		p.nextToken() // consume DECRYPTION
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
			p.nextToken() // consume PASSWORD
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			pwd, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Password = pwd
		}
	}

	// Parse ENCRYPTION BY PASSWORD clause
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
		p.nextToken() // consume ENCRYPTION
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
			p.nextToken() // consume PASSWORD
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			pwd, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.EncryptionPassword = pwd
		}
	}

	// Check for FORCE
	if strings.ToUpper(p.curTok.Literal) == "FORCE" {
		stmt.IsForce = true
		p.nextToken()
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

	// Parse optional ORDER clause (Azure Synapse/DW syntax - ORDER directly after ON table)
	if p.curTok.Type == TokenOrder || strings.ToUpper(p.curTok.Literal) == "ORDER" {
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

	// Parse optional WHERE clause (filtered index)
	if p.curTok.Type == TokenWhere {
		p.nextToken() // consume WHERE
		pred, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		stmt.FilterClause = pred
	}

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				if p.curTok.Type == TokenComma {
					p.nextToken()
					continue
				}

				optName := strings.ToUpper(p.curTok.Literal)
				switch optName {
				case "COMPRESSION_DELAY":
					p.nextToken() // consume COMPRESSION_DELAY
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					opt := &ast.CompressionDelayIndexOption{
						Expression: expr,
						TimeUnit:   "Unitless",
						OptionKind: "CompressionDelay",
					}
					// Check for MINUTE/MINUTES
					if strings.ToUpper(p.curTok.Literal) == "MINUTE" {
						opt.TimeUnit = "Minute"
						p.nextToken()
					} else if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
						opt.TimeUnit = "Minutes"
						p.nextToken()
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)

				case "SORT_IN_TEMPDB":
					p.nextToken() // consume SORT_IN_TEMPDB
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					state := "NotSet"
					if p.curTok.Type == TokenOn {
						state = "On"
						p.nextToken()
					} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
						state = "Off"
						p.nextToken()
					}
					stmt.IndexOptions = append(stmt.IndexOptions, &ast.IndexStateOption{
						OptionKind:  "SortInTempDB",
						OptionState: state,
					})

				case "ORDER":
					p.nextToken() // consume ORDER
					if p.curTok.Type == TokenLParen {
						p.nextToken() // consume (
						orderOpt := &ast.OrderIndexOption{
							OptionKind: "Order",
						}
						for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
							colRef := &ast.ColumnReferenceExpression{
								ColumnType: "Regular",
								MultiPartIdentifier: &ast.MultiPartIdentifier{
									Identifiers: []*ast.Identifier{p.parseIdentifier()},
								},
							}
							colRef.MultiPartIdentifier.Count = len(colRef.MultiPartIdentifier.Identifiers)
							orderOpt.Columns = append(orderOpt.Columns, colRef)

							if p.curTok.Type == TokenComma {
								p.nextToken()
							} else {
								break
							}
						}
						if p.curTok.Type == TokenRParen {
							p.nextToken()
						}
						stmt.IndexOptions = append(stmt.IndexOptions, orderOpt)
					}

				default:
					// Skip unknown options
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken()
						p.nextToken() // skip value
					}
				}
			}
			if p.curTok.Type == TokenRParen {
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
			if p.curTok.Type != TokenRParen && p.curTok.Type != TokenComma && p.curTok.Type != TokenEquals {
				dataType, err := p.parseDataType()
				if err != nil {
					return nil, err
				}
				param.DataType = dataType
			}

			// Parse optional default value
			if p.curTok.Type == TokenEquals {
				p.nextToken()
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				param.Value = val
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

	// Check if RETURNS TABLE
	if strings.ToUpper(p.curTok.Literal) == "TABLE" {
		p.nextToken()

		// Parse optional WITH clause for options
		if strings.ToUpper(p.curTok.Literal) == "WITH" {
			p.nextToken()
			for {
				opt := &ast.FunctionOption{}
				switch strings.ToUpper(p.curTok.Literal) {
				case "SCHEMABINDING":
					opt.OptionKind = "SchemaBinding"
				case "ENCRYPTION":
					opt.OptionKind = "Encryption"
				case "NATIVE_COMPILATION":
					opt.OptionKind = "NativeCompilation"
				default:
					opt.OptionKind = capitalizeFirst(p.curTok.Literal)
				}
				p.nextToken()
				stmt.Options = append(stmt.Options, opt)

				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
		}

		// Parse AS
		if p.curTok.Type == TokenAs {
			p.nextToken()
		}

		// For inline table-valued functions, parse RETURN SELECT...
		if strings.ToUpper(p.curTok.Literal) == "RETURN" {
			p.nextToken()
			// Parse the SELECT statement
			selectStmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if sel, ok := selectStmt.(*ast.SelectStatement); ok {
				stmt.ReturnType = &ast.SelectFunctionReturnType{
					SelectStatement: sel,
				}
			}
		}
	} else {
		// Scalar function - parse return type
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
	}

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

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		for {
			optName := strings.ToUpper(p.curTok.Literal)
			switch optName {
			case "NATIVE_COMPILATION":
				stmt.Options = append(stmt.Options, &ast.TriggerOption{OptionKind: "NativeCompile"})
				p.nextToken()
			case "SCHEMABINDING":
				stmt.Options = append(stmt.Options, &ast.TriggerOption{OptionKind: "SchemaBinding"})
				p.nextToken()
			case "ENCRYPTION":
				stmt.Options = append(stmt.Options, &ast.TriggerOption{OptionKind: "Encryption"})
				p.nextToken()
			case "EXECUTE":
				p.nextToken() // consume EXECUTE
				if p.curTok.Type == TokenAs {
					p.nextToken() // consume AS
				}
				execAsClause := &ast.ExecuteAsClause{}
				switch strings.ToUpper(p.curTok.Literal) {
				case "CALLER":
					execAsClause.ExecuteAsOption = "Caller"
				case "SELF":
					execAsClause.ExecuteAsOption = "Self"
				case "OWNER":
					execAsClause.ExecuteAsOption = "Owner"
				default:
					// User name
					execAsClause.ExecuteAsOption = "User"
				}
				p.nextToken()
				stmt.Options = append(stmt.Options, &ast.ExecuteAsTriggerOption{
					OptionKind:      "ExecuteAsClause",
					ExecuteAsClause: execAsClause,
				})
			default:
				// Unknown option, skip it
				p.nextToken()
			}
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
	}

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

func backupTransactionLogStatementToJSON(s *ast.BackupTransactionLogStatement) jsonNode {
	node := jsonNode{
		"$type": "BackupTransactionLogStatement",
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

func backupServiceMasterKeyStatementToJSON(s *ast.BackupServiceMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "BackupServiceMasterKeyStatement",
	}
	if s.File != nil {
		node["File"] = scalarExpressionToJSON(s.File)
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return node
}

func restoreServiceMasterKeyStatementToJSON(s *ast.RestoreServiceMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type":   "RestoreServiceMasterKeyStatement",
		"IsForce": s.IsForce,
	}
	if s.File != nil {
		node["File"] = scalarExpressionToJSON(s.File)
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return node
}

func restoreMasterKeyStatementToJSON(s *ast.RestoreMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type":   "RestoreMasterKeyStatement",
		"IsForce": s.IsForce,
	}
	if s.EncryptionPassword != nil {
		node["EncryptionPassword"] = scalarExpressionToJSON(s.EncryptionPassword)
	}
	if s.File != nil {
		node["File"] = scalarExpressionToJSON(s.File)
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
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
		node["PhysicalDevice"] = scalarExpressionToJSON(d.PhysicalDevice)
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
	if s.FilterClause != nil {
		node["FilterPredicate"] = booleanExpressionToJSON(s.FilterClause)
	}
	if len(s.IndexOptions) > 0 {
		opts := make([]jsonNode, len(s.IndexOptions))
		for i, opt := range s.IndexOptions {
			opts[i] = columnStoreIndexOptionToJSON(opt)
		}
		node["IndexOptions"] = opts
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

func columnStoreIndexOptionToJSON(opt ast.IndexOption) jsonNode {
	switch o := opt.(type) {
	case *ast.CompressionDelayIndexOption:
		node := jsonNode{
			"$type":      "CompressionDelayIndexOption",
			"OptionKind": o.OptionKind,
			"TimeUnit":   o.TimeUnit,
		}
		if o.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(o.Expression)
		}
		return node
	case *ast.OrderIndexOption:
		node := jsonNode{
			"$type":      "OrderIndexOption",
			"OptionKind": o.OptionKind,
		}
		if len(o.Columns) > 0 {
			cols := make([]jsonNode, len(o.Columns))
			for i, col := range o.Columns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["Columns"] = cols
		}
		return node
	case *ast.IndexStateOption:
		return jsonNode{
			"$type":       "IndexStateOption",
			"OptionKind":  o.OptionKind,
			"OptionState": o.OptionState,
		}
	default:
		return jsonNode{"$type": "UnknownIndexOption"}
	}
}

func createSpatialIndexStatementToJSON(s *ast.CreateSpatialIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSpatialIndexStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Object != nil {
		node["Object"] = schemaObjectNameToJSON(s.Object)
	}
	if s.SpatialColumnName != nil {
		node["SpatialColumnName"] = identifierToJSON(s.SpatialColumnName)
	}
	if s.SpatialIndexingScheme != "" {
		node["SpatialIndexingScheme"] = s.SpatialIndexingScheme
	}
	if s.OnFileGroup != nil {
		node["OnFileGroup"] = identifierOrValueExpressionToJSON(s.OnFileGroup)
	}
	if len(s.SpatialIndexOptions) > 0 {
		opts := make([]jsonNode, len(s.SpatialIndexOptions))
		for i, opt := range s.SpatialIndexOptions {
			opts[i] = spatialIndexOptionToJSON(opt)
		}
		node["SpatialIndexOptions"] = opts
	}
	return node
}

func spatialIndexOptionToJSON(opt ast.SpatialIndexOption) jsonNode {
	switch o := opt.(type) {
	case *ast.SpatialIndexRegularOption:
		node := jsonNode{
			"$type": "SpatialIndexRegularOption",
		}
		if o.Option != nil {
			node["Option"] = indexOptionToJSON(o.Option)
		}
		return node
	case *ast.BoundingBoxSpatialIndexOption:
		node := jsonNode{
			"$type": "BoundingBoxSpatialIndexOption",
		}
		if len(o.BoundingBoxParameters) > 0 {
			params := make([]jsonNode, len(o.BoundingBoxParameters))
			for i, p := range o.BoundingBoxParameters {
				params[i] = boundingBoxParameterToJSON(p)
			}
			node["BoundingBoxParameters"] = params
		}
		return node
	case *ast.GridsSpatialIndexOption:
		node := jsonNode{
			"$type": "GridsSpatialIndexOption",
		}
		if len(o.GridParameters) > 0 {
			params := make([]jsonNode, len(o.GridParameters))
			for i, p := range o.GridParameters {
				params[i] = gridParameterToJSON(p)
			}
			node["GridParameters"] = params
		}
		return node
	case *ast.CellsPerObjectSpatialIndexOption:
		node := jsonNode{
			"$type": "CellsPerObjectSpatialIndexOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownSpatialIndexOption"}
	}
}

func boundingBoxParameterToJSON(p *ast.BoundingBoxParameter) jsonNode {
	node := jsonNode{
		"$type": "BoundingBoxParameter",
	}
	if p.Parameter != "" {
		node["Parameter"] = p.Parameter
	}
	if p.Value != nil {
		node["Value"] = scalarExpressionToJSON(p.Value)
	}
	return node
}

func gridParameterToJSON(p *ast.GridParameter) jsonNode {
	node := jsonNode{
		"$type": "GridParameter",
	}
	if p.Parameter != "" {
		node["Parameter"] = p.Parameter
	}
	if p.Value != "" {
		node["Value"] = p.Value
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
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			opts[i] = functionOptionToJSON(o)
		}
		node["Options"] = opts
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

func functionOptionToJSON(o *ast.FunctionOption) jsonNode {
	return jsonNode{
		"$type":      "FunctionOption",
		"OptionKind": o.OptionKind,
	}
}

func createFunctionStatementToJSON(s *ast.CreateFunctionStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateFunctionStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = procedureParameterToJSON(p)
		}
		node["Parameters"] = params
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
	case *ast.SelectFunctionReturnType:
		node := jsonNode{
			"$type": "SelectFunctionReturnType",
		}
		if rt.SelectStatement != nil {
			node["SelectStatement"] = selectStatementToJSON(rt.SelectStatement)
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
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = triggerOptionTypeToJSON(o)
		}
		node["Options"] = options
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

func triggerOptionTypeToJSON(o ast.TriggerOptionType) jsonNode {
	switch opt := o.(type) {
	case *ast.TriggerOption:
		node := jsonNode{
			"$type":      "TriggerOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.OptionState != "" {
			node["OptionState"] = opt.OptionState
		}
		return node
	case *ast.ExecuteAsTriggerOption:
		node := jsonNode{
			"$type":      "ExecuteAsTriggerOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.ExecuteAsClause != nil {
			node["ExecuteAsClause"] = jsonNode{
				"$type":           "ExecuteAsClause",
				"ExecuteAsOption": opt.ExecuteAsClause.ExecuteAsOption,
			}
		}
		return node
	default:
		return jsonNode{"$type": "UnknownTriggerOption"}
	}
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

func enableDisableTriggerStatementToJSON(s *ast.EnableDisableTriggerStatement) jsonNode {
	node := jsonNode{
		"$type":              "EnableDisableTriggerStatement",
		"TriggerEnforcement": s.TriggerEnforcement,
		"All":                s.All,
	}
	if len(s.TriggerNames) > 0 {
		names := make([]jsonNode, len(s.TriggerNames))
		for i, n := range s.TriggerNames {
			names[i] = schemaObjectNameToJSON(n)
		}
		node["TriggerNames"] = names
	}
	if s.TriggerObject != nil {
		node["TriggerObject"] = triggerObjectToJSON(s.TriggerObject)
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
	case *ast.DataCompressionOption:
		node := jsonNode{
			"$type":            "DataCompressionOption",
			"CompressionLevel": o.CompressionLevel,
			"OptionKind":       o.OptionKind,
		}
		if len(o.PartitionRanges) > 0 {
			ranges := make([]jsonNode, len(o.PartitionRanges))
			for i, r := range o.PartitionRanges {
				ranges[i] = compressionPartitionRangeToJSON(r)
			}
			node["PartitionRanges"] = ranges
		}
		return node
	case *ast.IgnoreDupKeyIndexOption:
		return jsonNode{
			"$type":       "IgnoreDupKeyIndexOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
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

func dropExternalModelStatementToJSON(s *ast.DropExternalModelStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalModelStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
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

func createWorkloadGroupStatementToJSON(s *ast.CreateWorkloadGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateWorkloadGroupStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.WorkloadGroupParameters) > 0 {
		params := make([]jsonNode, len(s.WorkloadGroupParameters))
		for i, p := range s.WorkloadGroupParameters {
			params[i] = workloadGroupParameterToJSON(p)
		}
		node["WorkloadGroupParameters"] = params
	}
	if s.PoolName != nil {
		node["PoolName"] = identifierToJSON(s.PoolName)
	}
	if s.ExternalPoolName != nil {
		node["ExternalPoolName"] = identifierToJSON(s.ExternalPoolName)
	}
	return node
}

func alterWorkloadGroupStatementToJSON(s *ast.AlterWorkloadGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterWorkloadGroupStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.WorkloadGroupParameters) > 0 {
		params := make([]jsonNode, len(s.WorkloadGroupParameters))
		for i, p := range s.WorkloadGroupParameters {
			params[i] = workloadGroupParameterToJSON(p)
		}
		node["WorkloadGroupParameters"] = params
	}
	if s.PoolName != nil {
		node["PoolName"] = identifierToJSON(s.PoolName)
	}
	if s.ExternalPoolName != nil {
		node["ExternalPoolName"] = identifierToJSON(s.ExternalPoolName)
	}
	return node
}

func createWorkloadClassifierStatementToJSON(s *ast.CreateWorkloadClassifierStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateWorkloadClassifierStatement",
	}
	if s.ClassifierName != nil {
		node["ClassifierName"] = identifierToJSON(s.ClassifierName)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = workloadClassifierOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return node
}

func stringLiteralToJSON(s *ast.StringLiteral) jsonNode {
	node := jsonNode{
		"$type": "StringLiteral",
	}
	if s.LiteralType != "" {
		node["LiteralType"] = s.LiteralType
	} else {
		node["LiteralType"] = "String"
	}
	node["IsNational"] = s.IsNational
	node["IsLargeObject"] = s.IsLargeObject
	// Always include Value for StringLiteral, even if empty
	node["Value"] = s.Value
	return node
}

func workloadClassifierOptionToJSON(opt ast.WorkloadClassifierOption) jsonNode {
	switch o := opt.(type) {
	case *ast.ClassifierWorkloadGroupOption:
		node := jsonNode{
			"$type":      "ClassifierWorkloadGroupOption",
			"OptionType": o.OptionType,
		}
		if o.WorkloadGroupName != nil {
			node["WorkloadGroupName"] = stringLiteralToJSON(o.WorkloadGroupName)
		}
		return node
	case *ast.ClassifierMemberNameOption:
		node := jsonNode{
			"$type":      "ClassifierMemberNameOption",
			"OptionType": o.OptionType,
		}
		if o.MemberName != nil {
			node["MemberName"] = stringLiteralToJSON(o.MemberName)
		}
		return node
	case *ast.ClassifierWlmContextOption:
		node := jsonNode{
			"$type":      "ClassifierWlmContextOption",
			"OptionType": o.OptionType,
		}
		if o.WlmContext != nil {
			node["WlmContext"] = stringLiteralToJSON(o.WlmContext)
		}
		return node
	case *ast.ClassifierStartTimeOption:
		node := jsonNode{
			"$type":      "ClassifierStartTimeOption",
			"OptionType": o.OptionType,
		}
		if o.Time != nil {
			node["Time"] = wlmTimeLiteralToJSON(o.Time)
		}
		return node
	case *ast.ClassifierEndTimeOption:
		node := jsonNode{
			"$type":      "ClassifierEndTimeOption",
			"OptionType": o.OptionType,
		}
		if o.Time != nil {
			node["Time"] = wlmTimeLiteralToJSON(o.Time)
		}
		return node
	case *ast.ClassifierWlmLabelOption:
		node := jsonNode{
			"$type":      "ClassifierWlmLabelOption",
			"OptionType": o.OptionType,
		}
		if o.WlmLabel != nil {
			node["WlmLabel"] = stringLiteralToJSON(o.WlmLabel)
		}
		return node
	case *ast.ClassifierImportanceOption:
		return jsonNode{
			"$type":      "ClassifierImportanceOption",
			"Importance": o.Importance,
			"OptionType": o.OptionType,
		}
	default:
		return jsonNode{}
	}
}

func wlmTimeLiteralToJSON(t *ast.WlmTimeLiteral) jsonNode {
	node := jsonNode{
		"$type": "WlmTimeLiteral",
	}
	if t.TimeString != nil {
		node["TimeString"] = stringLiteralToJSON(t.TimeString)
	}
	return node
}

func workloadGroupParameterToJSON(p interface{}) jsonNode {
	switch param := p.(type) {
	case *ast.WorkloadGroupResourceParameter:
		node := jsonNode{
			"$type":         "WorkloadGroupResourceParameter",
			"ParameterType": param.ParameterType,
		}
		if param.ParameterValue != nil {
			node["ParameterValue"] = scalarExpressionToJSON(param.ParameterValue)
		}
		return node
	case *ast.WorkloadGroupImportanceParameter:
		return jsonNode{
			"$type":          "WorkloadGroupImportanceParameter",
			"ParameterType":  param.ParameterType,
			"ParameterValue": param.ParameterValue,
		}
	default:
		return jsonNode{}
	}
}

func alterSequenceStatementToJSON(s *ast.AlterSequenceStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterSequenceStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if len(s.SequenceOptions) > 0 {
		opts := make([]jsonNode, len(s.SequenceOptions))
		for i, opt := range s.SequenceOptions {
			opts[i] = sequenceOptionToJSON(opt)
		}
		node["SequenceOptions"] = opts
	}
	return node
}

func createSequenceStatementToJSON(s *ast.CreateSequenceStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSequenceStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if len(s.SequenceOptions) > 0 {
		opts := make([]jsonNode, len(s.SequenceOptions))
		for i, opt := range s.SequenceOptions {
			opts[i] = sequenceOptionToJSON(opt)
		}
		node["SequenceOptions"] = opts
	}
	return node
}

func sequenceOptionToJSON(opt interface{}) jsonNode {
	switch o := opt.(type) {
	case *ast.SequenceOption:
		return jsonNode{
			"$type":      "SequenceOption",
			"OptionKind": o.OptionKind,
			"NoValue":    o.NoValue,
		}
	case *ast.ScalarExpressionSequenceOption:
		node := jsonNode{
			"$type":      "ScalarExpressionSequenceOption",
			"OptionKind": o.OptionKind,
			"NoValue":    o.NoValue,
		}
		if o.OptionValue != nil {
			node["OptionValue"] = scalarExpressionToJSON(o.OptionValue)
		}
		return node
	case *ast.DataTypeSequenceOption:
		node := jsonNode{
			"$type":      "DataTypeSequenceOption",
			"OptionKind": o.OptionKind,
			"NoValue":    o.NoValue,
		}
		if o.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(o.DataType)
		}
		return node
	default:
		return jsonNode{}
	}
}

func dbccStatementToJSON(s *ast.DbccStatement) jsonNode {
	node := jsonNode{
		"$type":               "DbccStatement",
		"Command":             s.Command,
		"ParenthesisRequired": s.ParenthesisRequired,
		"OptionsUseJoin":      s.OptionsUseJoin,
	}
	if s.DllName != "" {
		node["DllName"] = s.DllName
	}
	if len(s.Literals) > 0 {
		lits := make([]jsonNode, len(s.Literals))
		for i, lit := range s.Literals {
			lits[i] = dbccNamedLiteralToJSON(lit)
		}
		node["Literals"] = lits
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = dbccOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return node
}

func dbccNamedLiteralToJSON(l *ast.DbccNamedLiteral) jsonNode {
	node := jsonNode{
		"$type": "DbccNamedLiteral",
	}
	if l.Name != "" {
		node["Name"] = l.Name
	}
	if l.Value != nil {
		node["Value"] = scalarExpressionToJSON(l.Value)
	}
	return node
}

func dbccOptionToJSON(o *ast.DbccOption) jsonNode {
	return jsonNode{
		"$type":      "DbccOption",
		"OptionKind": o.OptionKind,
	}
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
		node["SourcePartitionNumber"] = scalarExpressionToJSON(s.SourcePartition)
	}
	if s.TargetPartition != nil {
		node["TargetPartitionNumber"] = scalarExpressionToJSON(s.TargetPartition)
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

func alterTableSetStatementToJSON(s *ast.AlterTableSetStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterTableSetStatement",
	}
	if len(s.Options) > 0 {
		var options []jsonNode
		for _, opt := range s.Options {
			options = append(options, tableOptionToJSON(opt))
		}
		node["Options"] = options
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return node
}

func tableOptionToJSON(opt ast.TableOption) jsonNode {
	switch o := opt.(type) {
	case *ast.SystemVersioningTableOption:
		return systemVersioningTableOptionToJSON(o)
	default:
		return jsonNode{"$type": "UnknownTableOption"}
	}
}

func systemVersioningTableOptionToJSON(o *ast.SystemVersioningTableOption) jsonNode {
	node := jsonNode{
		"$type":                   "SystemVersioningTableOption",
		"OptionState":             o.OptionState,
		"ConsistencyCheckEnabled": o.ConsistencyCheckEnabled,
	}
	if o.HistoryTable != nil {
		node["HistoryTable"] = schemaObjectNameToJSON(o.HistoryTable)
	}
	if o.RetentionPeriod != nil {
		node["RetentionPeriod"] = retentionPeriodDefinitionToJSON(o.RetentionPeriod)
	}
	node["OptionKind"] = o.OptionKind
	return node
}

func retentionPeriodDefinitionToJSON(r *ast.RetentionPeriodDefinition) jsonNode {
	node := jsonNode{
		"$type": "RetentionPeriodDefinition",
	}
	if r.Duration != nil {
		node["Duration"] = scalarExpressionToJSON(r.Duration)
	}
	node["Units"] = r.Units
	node["IsInfinity"] = r.IsInfinity
	return node
}

func createExternalDataSourceStatementToJSON(s *ast.CreateExternalDataSourceStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalDataSourceStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.DataSourceType != "" {
		node["DataSourceType"] = s.DataSourceType
	}
	if s.Location != nil {
		node["Location"] = stringLiteralToJSON(s.Location)
	}
	if len(s.ExternalDataSourceOptions) > 0 {
		var options []jsonNode
		for _, opt := range s.ExternalDataSourceOptions {
			options = append(options, externalDataSourceOptionToJSON(opt))
		}
		node["ExternalDataSourceOptions"] = options
	}
	return node
}

func externalDataSourceOptionToJSON(opt *ast.ExternalDataSourceLiteralOrIdentifierOption) jsonNode {
	node := jsonNode{
		"$type": "ExternalDataSourceLiteralOrIdentifierOption",
	}
	if opt.Value != nil {
		node["Value"] = identifierOrValueExpressionToJSON(opt.Value)
	}
	if opt.OptionKind != "" {
		node["OptionKind"] = opt.OptionKind
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
	if s.FormatType != "" {
		node["FormatType"] = s.FormatType
	}
	if len(s.ExternalFileFormatOptions) > 0 {
		var options []jsonNode
		for _, opt := range s.ExternalFileFormatOptions {
			options = append(options, externalFileFormatOptionToJSON(opt))
		}
		node["ExternalFileFormatOptions"] = options
	}
	return node
}

func externalFileFormatOptionToJSON(opt ast.ExternalFileFormatOption) jsonNode {
	switch o := opt.(type) {
	case *ast.ExternalFileFormatContainerOption:
		node := jsonNode{
			"$type":      "ExternalFileFormatContainerOption",
			"OptionKind": o.OptionKind,
		}
		if len(o.Suboptions) > 0 {
			var subs []jsonNode
			for _, sub := range o.Suboptions {
				subs = append(subs, externalFileFormatOptionToJSON(sub))
			}
			node["Suboptions"] = subs
		}
		return node
	case *ast.ExternalFileFormatLiteralOption:
		node := jsonNode{
			"$type":      "ExternalFileFormatLiteralOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = stringLiteralToJSON(o.Value)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownExternalFileFormatOption"}
	}
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
			"IsUnique":   o.IsUnique,
		}
		if len(o.Columns) > 0 {
			cols := make([]jsonNode, len(o.Columns))
			for i, col := range o.Columns {
				cols[i] = columnWithSortOrderToJSON(col)
			}
			node["Columns"] = cols
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
	if s.Kind != "" {
		node["Kind"] = s.Kind
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.ActiveForBeginDialog != "" {
		node["ActiveForBeginDialog"] = s.ActiveForBeginDialog
	}
	if s.PrivateKeyPath != nil {
		node["PrivateKeyPath"] = scalarExpressionToJSON(s.PrivateKeyPath)
	}
	if s.DecryptionPassword != nil {
		node["DecryptionPassword"] = scalarExpressionToJSON(s.DecryptionPassword)
	}
	if s.EncryptionPassword != nil {
		node["EncryptionPassword"] = scalarExpressionToJSON(s.EncryptionPassword)
	}
	if s.AttestedBy != nil {
		node["AttestedBy"] = scalarExpressionToJSON(s.AttestedBy)
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
	if len(s.ApplicationRoleOptions) > 0 {
		opts := make([]jsonNode, len(s.ApplicationRoleOptions))
		for i, opt := range s.ApplicationRoleOptions {
			opts[i] = applicationRoleOptionToJSON(opt)
		}
		node["ApplicationRoleOptions"] = opts
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
	if len(s.QueueOptions) > 0 {
		opts := make([]jsonNode, len(s.QueueOptions))
		for i, opt := range s.QueueOptions {
			opts[i] = queueOptionToJSON(opt)
		}
		node["QueueOptions"] = opts
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
	if s.FileGroup != nil {
		node["FileGroup"] = identifierOrValueExpressionToJSON(s.FileGroup)
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
	// Only include IsSplit when there was a SPLIT or MERGE action
	if s.HasAction {
		node["IsSplit"] = s.IsSplit
	}
	if s.Boundary != nil {
		node["Boundary"] = scalarExpressionToJSON(s.Boundary)
	}
	return node
}

func alterFulltextCatalogStatementToJSON(s *ast.AlterFulltextCatalogStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterFullTextCatalogStatement",
	}
	if s.Action != "" {
		node["Action"] = s.Action
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			optNode := jsonNode{
				"$type": "OnOffFullTextCatalogOption",
			}
			if opt.OptionState != "" {
				optNode["OptionState"] = opt.OptionState
			}
			if opt.OptionKind != "" {
				optNode["OptionKind"] = opt.OptionKind
			}
			opts[i] = optNode
		}
		node["Options"] = opts
	}
	return node
}

func createFullTextCatalogStatementToJSON(s *ast.CreateFullTextCatalogStatement) jsonNode {
	node := jsonNode{
		"$type":     "CreateFullTextCatalogStatement",
		"IsDefault": s.IsDefault,
	}
	if s.FileGroup != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroup)
	}
	if s.Path != nil {
		node["Path"] = scalarExpressionToJSON(s.Path)
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			optNode := jsonNode{
				"$type": "OnOffFullTextCatalogOption",
			}
			if opt.OptionState != "" {
				optNode["OptionState"] = opt.OptionState
			}
			if opt.OptionKind != "" {
				optNode["OptionKind"] = opt.OptionKind
			}
			opts[i] = optNode
		}
		node["Options"] = opts
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
	node := jsonNode{
		"$type": "AlterServiceMasterKeyStatement",
	}
	if s.Kind != "" {
		node["Kind"] = s.Kind
	}
	if s.Account != nil {
		node["Account"] = scalarExpressionToJSON(s.Account)
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return node
}

func renameEntityStatementToJSON(s *ast.RenameEntityStatement) jsonNode {
	node := jsonNode{
		"$type": "RenameEntityStatement",
	}
	if s.RenameEntityType != "" {
		node["RenameEntityType"] = s.RenameEntityType
	}
	if s.SeparatorType != "" {
		node["SeparatorType"] = s.SeparatorType
	}
	if s.OldName != nil {
		node["OldName"] = schemaObjectNameToJSON(s.OldName)
	}
	if s.NewName != nil {
		node["NewName"] = identifierToJSON(s.NewName)
	}
	return node
}

func createDatabaseStatementToJSON(s *ast.CreateDatabaseStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateDatabaseStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = createDatabaseOptionToJSON(opt)
		}
		node["Options"] = opts
		// Only output AttachMode when there are options
		if s.AttachMode != "" {
			node["AttachMode"] = s.AttachMode
		}
	}
	return node
}

func createDatabaseOptionToJSON(opt ast.CreateDatabaseOption) jsonNode {
	switch o := opt.(type) {
	case *ast.OnOffDatabaseOption:
		return jsonNode{
			"$type":       "OnOffDatabaseOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
	case *ast.IdentifierDatabaseOption:
		node := jsonNode{
			"$type":      "IdentifierDatabaseOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = identifierToJSON(o.Value)
		}
		return node
	default:
		return jsonNode{"$type": "CreateDatabaseOption"}
	}
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
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.ValidationMethod != "" {
		node["ValidationMethod"] = s.ValidationMethod
	}
	if s.XmlSchemaCollectionName != nil {
		node["XmlSchemaCollectionName"] = schemaObjectNameToJSON(s.XmlSchemaCollectionName)
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
	if len(s.QueueOptions) > 0 {
		opts := make([]jsonNode, len(s.QueueOptions))
		for i, opt := range s.QueueOptions {
			opts[i] = queueOptionToJSON(opt)
		}
		node["QueueOptions"] = opts
	}
	return node
}

func queueOptionToJSON(opt ast.QueueOption) jsonNode {
	switch o := opt.(type) {
	case *ast.QueueStateOption:
		node := jsonNode{
			"$type":       "QueueStateOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		return node
	case *ast.QueueOptionSimple:
		node := jsonNode{
			"$type":      "QueueOption",
			"OptionKind": o.OptionKind,
		}
		return node
	default:
		return jsonNode{"$type": "QueueOption"}
	}
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
	if len(s.ApplicationRoleOptions) > 0 {
		opts := make([]jsonNode, len(s.ApplicationRoleOptions))
		for i, opt := range s.ApplicationRoleOptions {
			opts[i] = applicationRoleOptionToJSON(opt)
		}
		node["ApplicationRoleOptions"] = opts
	}
	return node
}

func applicationRoleOptionToJSON(opt *ast.ApplicationRoleOption) jsonNode {
	node := jsonNode{
		"$type":      "ApplicationRoleOption",
		"OptionKind": opt.OptionKind,
	}
	if opt.Value != nil {
		node["Value"] = identifierOrValueExpressionToJSON(opt.Value)
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
	if len(s.Columns) > 0 {
		cols := make([]jsonNode, len(s.Columns))
		for i, c := range s.Columns {
			cols[i] = columnReferenceExpressionToJSON(c)
		}
		node["Columns"] = cols
	}
	if len(s.StatisticsOptions) > 0 {
		opts := make([]jsonNode, len(s.StatisticsOptions))
		for i, o := range s.StatisticsOptions {
			opts[i] = statisticsOptionToJSON(o)
		}
		node["StatisticsOptions"] = opts
	}
	if s.FilterPredicate != nil {
		node["FilterPredicate"] = booleanExpressionToJSON(s.FilterPredicate)
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

func createTypeUddtStatementToJSON(s *ast.CreateTypeUddtStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateTypeUddtStatement",
	}
	if s.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(s.DataType)
	}
	if s.NullableConstraint != nil {
		node["NullableConstraint"] = nullableConstraintToJSON(s.NullableConstraint)
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return node
}

func createTypeUdtStatementToJSON(s *ast.CreateTypeUdtStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateTypeUdtStatement",
	}
	if s.AssemblyName != nil {
		node["AssemblyName"] = assemblyNameToJSON(s.AssemblyName)
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
	if s.ParameterType != nil {
		node["ParameterType"] = partitionParameterTypeToJSON(s.ParameterType)
	}
	if s.Range != "" {
		node["Range"] = s.Range
	}
	if len(s.BoundaryValues) > 0 {
		values := make([]jsonNode, len(s.BoundaryValues))
		for i, v := range s.BoundaryValues {
			values[i] = scalarExpressionToJSON(v)
		}
		node["BoundaryValues"] = values
	}
	return node
}

func partitionParameterTypeToJSON(p *ast.PartitionParameterType) jsonNode {
	node := jsonNode{
		"$type": "PartitionParameterType",
	}
	if p.DataType != nil {
		node["DataType"] = sqlDataTypeReferenceToJSON(p.DataType)
	}
	if p.Collation != nil {
		node["Collation"] = identifierToJSON(p.Collation)
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
	if s.Scope != nil {
		node["Scope"] = eventNotificationObjectScopeToJSON(s.Scope)
		// Include WithFanIn when Scope is present
		node["WithFanIn"] = s.WithFanIn
	}
	if len(s.EventTypeGroups) > 0 {
		groups := make([]jsonNode, len(s.EventTypeGroups))
		for i, g := range s.EventTypeGroups {
			groups[i] = eventTypeGroupContainerToJSON(g)
		}
		node["EventTypeGroups"] = groups
	}
	if s.BrokerService != nil {
		node["BrokerService"] = stringLiteralToJSON(s.BrokerService)
	}
	if s.BrokerInstanceSpecifier != nil {
		node["BrokerInstanceSpecifier"] = stringLiteralToJSON(s.BrokerInstanceSpecifier)
	}
	return node
}

func eventNotificationObjectScopeToJSON(s *ast.EventNotificationObjectScope) jsonNode {
	node := jsonNode{
		"$type":  "EventNotificationObjectScope",
		"Target": s.Target,
	}
	if s.QueueName != nil {
		node["QueueName"] = schemaObjectNameToJSON(s.QueueName)
	}
	return node
}

func eventTypeGroupContainerToJSON(c ast.EventTypeGroupContainer) jsonNode {
	switch v := c.(type) {
	case *ast.EventTypeContainer:
		return jsonNode{
			"$type":     "EventTypeContainer",
			"EventType": v.EventType,
		}
	case *ast.EventGroupContainer:
		return jsonNode{
			"$type":      "EventGroupContainer",
			"EventGroup": v.EventGroup,
		}
	default:
		return jsonNode{"$type": "Unknown"}
	}
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
	node["ContainsFileStream"] = s.ContainsFileStream
	node["ContainsMemoryOptimizedData"] = s.ContainsMemoryOptimizedData
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.FileGroupName != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroupName)
	}
	node["UseCurrent"] = s.UseCurrent
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
	node["MakeDefault"] = s.MakeDefault
	node["UseCurrent"] = false
	if s.NewFileGroupName != nil {
		node["NewFileGroupName"] = identifierToJSON(s.NewFileGroupName)
	}
	if s.UpdatabilityOption != "" {
		node["UpdatabilityOption"] = s.UpdatabilityOption
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
	node["UseCurrent"] = s.UseCurrent
	return node
}

func alterDatabaseScopedConfigurationClearStatementToJSON(s *ast.AlterDatabaseScopedConfigurationClearStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseScopedConfigurationClearStatement",
	}
	if s.Option != nil {
		node["Option"] = databaseConfigurationClearOptionToJSON(s.Option)
	}
	node["Secondary"] = s.Secondary
	return node
}

func databaseConfigurationClearOptionToJSON(o *ast.DatabaseConfigurationClearOption) jsonNode {
	node := jsonNode{
		"$type": "DatabaseConfigurationClearOption",
	}
	if o.OptionKind != "" {
		node["OptionKind"] = o.OptionKind
	}
	if o.PlanHandle != nil {
		node["PlanHandle"] = scalarExpressionToJSON(o.PlanHandle)
	}
	return node
}

func alterResourceGovernorStatementToJSON(s *ast.AlterResourceGovernorStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterResourceGovernorStatement",
	}
	if s.Command != "" {
		node["Command"] = s.Command
	}
	if s.ClassifierFunction != nil {
		node["ClassifierFunction"] = schemaObjectNameToJSON(s.ClassifierFunction)
	}
	return node
}

func createCryptographicProviderStatementToJSON(s *ast.CreateCryptographicProviderStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateCryptographicProviderStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.File != nil {
		node["File"] = scalarExpressionToJSON(s.File)
	}
	return node
}

func alterCryptographicProviderStatementToJSON(s *ast.AlterCryptographicProviderStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterCryptographicProviderStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Option != "" {
		node["Option"] = s.Option
	}
	if s.File != nil {
		node["File"] = scalarExpressionToJSON(s.File)
	}
	return node
}

func dropCryptographicProviderStatementToJSON(s *ast.DropCryptographicProviderStatement) jsonNode {
	node := jsonNode{
		"$type": "DropCryptographicProviderStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return node
}

func useFederationStatementToJSON(s *ast.UseFederationStatement) jsonNode {
	node := jsonNode{
		"$type": "UseFederationStatement",
	}
	if s.FederationName != nil {
		node["FederationName"] = identifierToJSON(s.FederationName)
	}
	if s.DistributionName != nil {
		node["DistributionName"] = identifierToJSON(s.DistributionName)
	}
	if s.Value != nil {
		node["Value"] = scalarExpressionToJSON(s.Value)
	}
	node["Filtering"] = s.Filtering
	return node
}

func createFederationStatementToJSON(s *ast.CreateFederationStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateFederationStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.DistributionName != nil {
		node["DistributionName"] = identifierToJSON(s.DistributionName)
	}
	if s.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(s.DataType)
	}
	return node
}

func alterFederationStatementToJSON(s *ast.AlterFederationStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterFederationStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Kind != "" {
		node["Kind"] = s.Kind
	}
	if s.DistributionName != nil {
		node["DistributionName"] = identifierToJSON(s.DistributionName)
	}
	if s.Boundary != nil {
		node["Boundary"] = scalarExpressionToJSON(s.Boundary)
	}
	return node
}

func callTargetToJSON(ct ast.CallTarget) jsonNode {
	switch t := ct.(type) {
	case *ast.MultiPartIdentifierCallTarget:
		node := jsonNode{
			"$type": "MultiPartIdentifierCallTarget",
		}
		if t.MultiPartIdentifier != nil {
			node["MultiPartIdentifier"] = multiPartIdentifierToJSON(t.MultiPartIdentifier)
		}
		return node
	case *ast.ExpressionCallTarget:
		node := jsonNode{
			"$type": "ExpressionCallTarget",
		}
		if t.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(t.Expression)
		}
		return node
	case *ast.UserDefinedTypeCallTarget:
		node := jsonNode{
			"$type": "UserDefinedTypeCallTarget",
		}
		if t.SchemaObjectName != nil {
			node["SchemaObjectName"] = schemaObjectNameToJSON(t.SchemaObjectName)
		}
		return node
	default:
		return jsonNode{}
	}
}

func alterProcedureStatementToJSON(s *ast.AlterProcedureStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterProcedureStatement",
	}
	if s.ProcedureReference != nil {
		node["ProcedureReference"] = procedureReferenceToJSON(s.ProcedureReference)
	}
	node["IsForReplication"] = s.IsForReplication
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

func alterExternalDataSourceStatementToJSON(s *ast.AlterExternalDataSourceStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterExternalDataSourceStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ExternalDataSourceOptions) > 0 {
		opts := make([]jsonNode, len(s.ExternalDataSourceOptions))
		for i, o := range s.ExternalDataSourceOptions {
			opts[i] = externalDataSourceOptionToJSON(o)
		}
		node["ExternalDataSourceOptions"] = opts
	}
	return node
}

func alterExternalLanguageStatementToJSON(s *ast.AlterExternalLanguageStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterExternalLanguageStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func alterExternalLibraryStatementToJSON(s *ast.AlterExternalLibraryStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterExternalLibraryStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return node
}

func fetchTypeToJSON(f *ast.FetchType) jsonNode {
	node := jsonNode{
		"$type": "FetchType",
	}
	if f.Orientation != "" {
		node["Orientation"] = f.Orientation
	}
	if f.RowOffset != nil {
		node["RowOffset"] = scalarExpressionToJSON(f.RowOffset)
	}
	return node
}

func openCursorStatementToJSON(s *ast.OpenCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "OpenCursorStatement",
	}
	if s.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(s.Cursor)
	}
	return node
}

func closeCursorStatementToJSON(s *ast.CloseCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "CloseCursorStatement",
	}
	if s.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(s.Cursor)
	}
	return node
}

func deallocateCursorStatementToJSON(s *ast.DeallocateCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "DeallocateCursorStatement",
	}
	if s.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(s.Cursor)
	}
	return node
}

func fetchCursorStatementToJSON(s *ast.FetchCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "FetchCursorStatement",
	}
	if s.FetchType != nil {
		node["FetchType"] = fetchTypeToJSON(s.FetchType)
	}
	if s.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(s.Cursor)
	}
	if len(s.IntoVariables) > 0 {
		vars := make([]jsonNode, len(s.IntoVariables))
		for i, v := range s.IntoVariables {
			vars[i] = scalarExpressionToJSON(v)
		}
		node["IntoVariables"] = vars
	}
	return node
}

func updateStatisticsStatementToJSON(s *ast.UpdateStatisticsStatement) jsonNode {
	node := jsonNode{
		"$type": "UpdateStatisticsStatement",
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	if len(s.SubElements) > 0 {
		elems := make([]jsonNode, len(s.SubElements))
		for i, e := range s.SubElements {
			elems[i] = identifierToJSON(e)
		}
		node["SubElements"] = elems
	}
	if len(s.StatisticsOptions) > 0 {
		opts := make([]jsonNode, len(s.StatisticsOptions))
		for i, o := range s.StatisticsOptions {
			opts[i] = statisticsOptionToJSON(o)
		}
		node["StatisticsOptions"] = opts
	}
	return node
}

func statisticsOptionToJSON(opt ast.StatisticsOption) jsonNode {
	switch o := opt.(type) {
	case *ast.SimpleStatisticsOption:
		return simpleStatisticsOptionToJSON(o)
	case *ast.LiteralStatisticsOption:
		return literalStatisticsOptionToJSON(o)
	case *ast.OnOffStatisticsOption:
		return onOffStatisticsOptionToJSON(o)
	case *ast.ResampleStatisticsOption:
		return resampleStatisticsOptionToJSON(o)
	default:
		return jsonNode{"$type": "UnknownStatisticsOption"}
	}
}

func simpleStatisticsOptionToJSON(o *ast.SimpleStatisticsOption) jsonNode {
	node := jsonNode{
		"$type": "StatisticsOption",
	}
	if o.OptionKind != "" {
		node["OptionKind"] = o.OptionKind
	}
	return node
}

func literalStatisticsOptionToJSON(o *ast.LiteralStatisticsOption) jsonNode {
	node := jsonNode{
		"$type": "LiteralStatisticsOption",
	}
	if o.OptionKind != "" {
		node["OptionKind"] = o.OptionKind
	}
	if o.Literal != nil {
		node["Literal"] = scalarExpressionToJSON(o.Literal)
	}
	return node
}

func onOffStatisticsOptionToJSON(o *ast.OnOffStatisticsOption) jsonNode {
	node := jsonNode{
		"$type": "OnOffStatisticsOption",
	}
	if o.OptionKind != "" {
		node["OptionKind"] = o.OptionKind
	}
	if o.OptionState != "" {
		node["OptionState"] = o.OptionState
	}
	return node
}

func resampleStatisticsOptionToJSON(o *ast.ResampleStatisticsOption) jsonNode {
	node := jsonNode{
		"$type": "ResampleStatisticsOption",
	}
	if o.OptionKind != "" {
		node["OptionKind"] = o.OptionKind
	}
	if len(o.Partitions) > 0 {
		partitions := make([]jsonNode, len(o.Partitions))
		for i, p := range o.Partitions {
			partitions[i] = scalarExpressionToJSON(p)
		}
		node["Partitions"] = partitions
	}
	return node
}

func declareCursorStatementToJSON(s *ast.DeclareCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "DeclareCursorStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.CursorDefinition != nil {
		node["CursorDefinition"] = declareCursorDefinitionToJSON(s.CursorDefinition)
	}
	return node
}

func declareCursorDefinitionToJSON(d *ast.CursorDefinition) jsonNode {
	node := jsonNode{
		"$type": "CursorDefinition",
	}
	if len(d.Options) > 0 {
		opts := make([]jsonNode, len(d.Options))
		for i, o := range d.Options {
			opts[i] = jsonNode{
				"$type":      "CursorOption",
				"OptionKind": o.OptionKind,
			}
		}
		node["Options"] = opts
	}
	if d.Select != nil {
		node["Select"] = selectStatementToJSON(d.Select)
	}
	return node
}
