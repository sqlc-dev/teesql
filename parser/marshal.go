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

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool {
	return &b
}

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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(b))
}

func statementToJSON(stmt ast.Statement) jsonNode {
	switch s := stmt.(type) {
	case *ast.SelectStatement:
		return addSpan(selectStatementToJSON(s), frag(stmt))
	case *ast.InsertStatement:
		return addSpan(insertStatementToJSON(s), frag(stmt))
	case *ast.UpdateStatement:
		return addSpan(updateStatementToJSON(s), frag(stmt))
	case *ast.UpdateStatisticsStatement:
		return addSpan(updateStatisticsStatementToJSON(s), frag(stmt))
	case *ast.DeleteStatement:
		return addSpan(deleteStatementToJSON(s), frag(stmt))
	case *ast.MergeStatement:
		return addSpan(mergeStatementToJSON(s), frag(stmt))
	case *ast.DeclareVariableStatement:
		return addSpan(declareVariableStatementToJSON(s), frag(stmt))
	case *ast.DeclareTableVariableStatement:
		return addSpan(declareTableVariableStatementToJSON(s), frag(stmt))
	case *ast.SetVariableStatement:
		return addSpan(setVariableStatementToJSON(s), frag(stmt))
	case *ast.IfStatement:
		return addSpan(ifStatementToJSON(s), frag(stmt))
	case *ast.WhileStatement:
		return addSpan(whileStatementToJSON(s), frag(stmt))
	case *ast.BeginEndBlockStatement:
		return addSpan(beginEndBlockStatementToJSON(s), frag(stmt))
	case *ast.BeginEndAtomicBlockStatement:
		return addSpan(beginEndAtomicBlockStatementToJSON(s), frag(stmt))
	case *ast.BeginDialogStatement:
		return addSpan(beginDialogStatementToJSON(s), frag(stmt))
	case *ast.BeginConversationTimerStatement:
		return addSpan(beginConversationTimerStatementToJSON(s), frag(stmt))
	case *ast.CreateViewStatement:
		return addSpan(createViewStatementToJSON(s), frag(stmt))
	case *ast.CreateOrAlterViewStatement:
		return addSpan(createOrAlterViewStatementToJSON(s), frag(stmt))
	case *ast.AlterViewStatement:
		return addSpan(alterViewStatementToJSON(s), frag(stmt))
	case *ast.CreateSchemaStatement:
		return addSpan(createSchemaStatementToJSON(s), frag(stmt))
	case *ast.CreateProcedureStatement:
		return addSpan(createProcedureStatementToJSON(s), frag(stmt))
	case *ast.CreateOrAlterProcedureStatement:
		return addSpan(createOrAlterProcedureStatementToJSON(s), frag(stmt))
	case *ast.AlterProcedureStatement:
		return addSpan(alterProcedureStatementToJSON(s), frag(stmt))
	case *ast.CreateRoleStatement:
		return addSpan(createRoleStatementToJSON(s), frag(stmt))
	case *ast.ExecuteStatement:
		return addSpan(executeStatementToJSON(s), frag(stmt))
	case *ast.ExecuteAsStatement:
		return addSpan(executeAsStatementToJSON(s), frag(stmt))
	case *ast.ReturnStatement:
		return addSpan(returnStatementToJSON(s), frag(stmt))
	case *ast.BreakStatement:
		return addSpan(breakStatementToJSON(), frag(stmt))
	case *ast.ContinueStatement:
		return addSpan(continueStatementToJSON(), frag(stmt))
	case *ast.PrintStatement:
		return addSpan(printStatementToJSON(s), frag(stmt))
	case *ast.ThrowStatement:
		return addSpan(throwStatementToJSON(s), frag(stmt))
	case *ast.AlterTableDropTableElementStatement:
		return addSpan(alterTableDropTableElementStatementToJSON(s), frag(stmt))
	case *ast.AlterTableAlterIndexStatement:
		return addSpan(alterTableAlterIndexStatementToJSON(s), frag(stmt))
	case *ast.AlterTableAddTableElementStatement:
		return addSpan(alterTableAddTableElementStatementToJSON(s), frag(stmt))
	case *ast.AlterTableAlterColumnStatement:
		return addSpan(alterTableAlterColumnStatementToJSON(s), frag(stmt))
	case *ast.AlterMessageTypeStatement:
		return addSpan(alterMessageTypeStatementToJSON(s), frag(stmt))
	case *ast.CreateContractStatement:
		return addSpan(createContractStatementToJSON(s), frag(stmt))
	case *ast.CreatePartitionSchemeStatement:
		return addSpan(createPartitionSchemeStatementToJSON(s), frag(stmt))
	case *ast.CreateRuleStatement:
		return addSpan(createRuleStatementToJSON(s), frag(stmt))
	case *ast.CreateSynonymStatement:
		return addSpan(createSynonymStatementToJSON(s), frag(stmt))
	case *ast.AlterCredentialStatement:
		return addSpan(alterCredentialStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseSetStatement:
		return addSpan(alterDatabaseSetStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseAddFileStatement:
		return addSpan(alterDatabaseAddFileStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseAddFileGroupStatement:
		return addSpan(alterDatabaseAddFileGroupStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseModifyFileStatement:
		return addSpan(alterDatabaseModifyFileStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseModifyFileGroupStatement:
		return addSpan(alterDatabaseModifyFileGroupStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseModifyNameStatement:
		return addSpan(alterDatabaseModifyNameStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseRemoveFileStatement:
		return addSpan(alterDatabaseRemoveFileStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseRemoveFileGroupStatement:
		return addSpan(alterDatabaseRemoveFileGroupStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseCollateStatement:
		return addSpan(alterDatabaseCollateStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseRebuildLogStatement:
		return addSpan(alterDatabaseRebuildLogStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseScopedConfigurationClearStatement:
		return addSpan(alterDatabaseScopedConfigurationClearStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseScopedConfigurationSetStatement:
		return addSpan(alterDatabaseScopedConfigurationSetStatementToJSON(s), frag(stmt))
	case *ast.AlterResourceGovernorStatement:
		return addSpan(alterResourceGovernorStatementToJSON(s), frag(stmt))
	case *ast.CreateResourcePoolStatement:
		return addSpan(createResourcePoolStatementToJSON(s), frag(stmt))
	case *ast.AlterResourcePoolStatement:
		return addSpan(alterResourcePoolStatementToJSON(s), frag(stmt))
	case *ast.DropResourcePoolStatement:
		return addSpan(dropResourcePoolStatementToJSON(s), frag(stmt))
	case *ast.AlterExternalResourcePoolStatement:
		return addSpan(alterExternalResourcePoolStatementToJSON(s), frag(stmt))
	case *ast.CreateExternalResourcePoolStatement:
		return addSpan(createExternalResourcePoolStatementToJSON(s), frag(stmt))
	case *ast.CreateCryptographicProviderStatement:
		return addSpan(createCryptographicProviderStatementToJSON(s), frag(stmt))
	case *ast.CreateColumnMasterKeyStatement:
		return addSpan(createColumnMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.DropColumnMasterKeyStatement:
		return addSpan(dropColumnMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.CreateColumnEncryptionKeyStatement:
		return addSpan(createColumnEncryptionKeyStatementToJSON(s), frag(stmt))
	case *ast.AlterColumnEncryptionKeyStatement:
		return addSpan(alterColumnEncryptionKeyStatementToJSON(s), frag(stmt))
	case *ast.DropColumnEncryptionKeyStatement:
		return addSpan(dropColumnEncryptionKeyStatementToJSON(s), frag(stmt))
	case *ast.AlterCryptographicProviderStatement:
		return addSpan(alterCryptographicProviderStatementToJSON(s), frag(stmt))
	case *ast.DropCryptographicProviderStatement:
		return addSpan(dropCryptographicProviderStatementToJSON(s), frag(stmt))
	case *ast.CreateBrokerPriorityStatement:
		return addSpan(createBrokerPriorityStatementToJSON(s), frag(stmt))
	case *ast.AlterBrokerPriorityStatement:
		return addSpan(alterBrokerPriorityStatementToJSON(s), frag(stmt))
	case *ast.DropBrokerPriorityStatement:
		return addSpan(dropBrokerPriorityStatementToJSON(s), frag(stmt))
	case *ast.UseFederationStatement:
		return addSpan(useFederationStatementToJSON(s), frag(stmt))
	case *ast.CreateFederationStatement:
		return addSpan(createFederationStatementToJSON(s), frag(stmt))
	case *ast.AlterFederationStatement:
		return addSpan(alterFederationStatementToJSON(s), frag(stmt))
	case *ast.RevertStatement:
		return addSpan(revertStatementToJSON(s), frag(stmt))
	case *ast.DropCredentialStatement:
		return addSpan(dropCredentialStatementToJSON(s), frag(stmt))
	case *ast.DropExternalLanguageStatement:
		return addSpan(dropExternalLanguageStatementToJSON(s), frag(stmt))
	case *ast.DropExternalLibraryStatement:
		return addSpan(dropExternalLibraryStatementToJSON(s), frag(stmt))
	case *ast.DropSequenceStatement:
		return addSpan(dropSequenceStatementToJSON(s), frag(stmt))
	case *ast.DropSearchPropertyListStatement:
		return addSpan(dropSearchPropertyListStatementToJSON(s), frag(stmt))
	case *ast.DropServerRoleStatement:
		return addSpan(dropServerRoleStatementToJSON(s), frag(stmt))
	case *ast.DropServerAuditStatement:
		return addSpan(dropServerAuditStatementToJSON(s), frag(stmt))
	case *ast.DropServerAuditSpecificationStatement:
		return addSpan(dropServerAuditSpecificationStatementToJSON(s), frag(stmt))
	case *ast.DropDatabaseAuditSpecificationStatement:
		return addSpan(dropDatabaseAuditSpecificationStatementToJSON(s), frag(stmt))
	case *ast.DropAvailabilityGroupStatement:
		return addSpan(dropAvailabilityGroupStatementToJSON(s), frag(stmt))
	case *ast.DropFederationStatement:
		return addSpan(dropFederationStatementToJSON(s), frag(stmt))
	case *ast.DropSecurityPolicyStatement:
		return addSpan(dropSecurityPolicyStatementToJSON(s), frag(stmt))
	case *ast.DropExternalDataSourceStatement:
		return addSpan(dropExternalDataSourceStatementToJSON(s), frag(stmt))
	case *ast.AlterExternalDataSourceStatement:
		return addSpan(alterExternalDataSourceStatementToJSON(s), frag(stmt))
	case *ast.AlterExternalLanguageStatement:
		return addSpan(alterExternalLanguageStatementToJSON(s), frag(stmt))
	case *ast.AlterExternalLibraryStatement:
		return addSpan(alterExternalLibraryStatementToJSON(s), frag(stmt))
	case *ast.DropExternalFileFormatStatement:
		return addSpan(dropExternalFileFormatStatementToJSON(s), frag(stmt))
	case *ast.DropExternalTableStatement:
		return addSpan(dropExternalTableStatementToJSON(s), frag(stmt))
	case *ast.DropExternalResourcePoolStatement:
		return addSpan(dropExternalResourcePoolStatementToJSON(s), frag(stmt))
	case *ast.DropExternalModelStatement:
		return addSpan(dropExternalModelStatementToJSON(s), frag(stmt))
	case *ast.DropWorkloadGroupStatement:
		return addSpan(dropWorkloadGroupStatementToJSON(s), frag(stmt))
	case *ast.DropWorkloadClassifierStatement:
		return addSpan(dropWorkloadClassifierStatementToJSON(s), frag(stmt))
	case *ast.CreateWorkloadGroupStatement:
		return addSpan(createWorkloadGroupStatementToJSON(s), frag(stmt))
	case *ast.CreateWorkloadClassifierStatement:
		return addSpan(createWorkloadClassifierStatementToJSON(s), frag(stmt))
	case *ast.AlterWorkloadGroupStatement:
		return addSpan(alterWorkloadGroupStatementToJSON(s), frag(stmt))
	case *ast.AlterSequenceStatement:
		return addSpan(alterSequenceStatementToJSON(s), frag(stmt))
	case *ast.CreateSequenceStatement:
		return addSpan(createSequenceStatementToJSON(s), frag(stmt))
	case *ast.DbccStatement:
		return addSpan(dbccStatementToJSON(s), frag(stmt))
	case *ast.DropTypeStatement:
		return addSpan(dropTypeStatementToJSON(s), frag(stmt))
	case *ast.DropAggregateStatement:
		return addSpan(dropAggregateStatementToJSON(s), frag(stmt))
	case *ast.DropSynonymStatement:
		return addSpan(dropSynonymStatementToJSON(s), frag(stmt))
	case *ast.DropUserStatement:
		return addSpan(dropUserStatementToJSON(s), frag(stmt))
	case *ast.DropRoleStatement:
		return addSpan(dropRoleStatementToJSON(s), frag(stmt))
	case *ast.DropAssemblyStatement:
		return addSpan(dropAssemblyStatementToJSON(s), frag(stmt))
	case *ast.DropAsymmetricKeyStatement:
		return addSpan(dropAsymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.DropSymmetricKeyStatement:
		return addSpan(dropSymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.CreateTableStatement:
		return addSpan(createTableStatementToJSON(s), frag(stmt))
	case *ast.GrantStatement:
		return addSpan(grantStatementToJSON(s), frag(stmt))
	case *ast.RevokeStatement:
		return addSpan(revokeStatementToJSON(s), frag(stmt))
	case *ast.DenyStatement:
		return addSpan(denyStatementToJSON(s), frag(stmt))
	case *ast.PredicateSetStatement:
		return addSpan(predicateSetStatementToJSON(s), frag(stmt))
	case *ast.SetStatisticsStatement:
		return addSpan(setStatisticsStatementToJSON(s), frag(stmt))
	case *ast.SetRowCountStatement:
		return addSpan(setRowCountStatementToJSON(s), frag(stmt))
	case *ast.SetOffsetsStatement:
		return addSpan(setOffsetsStatementToJSON(s), frag(stmt))
	case *ast.SetCommandStatement:
		return addSpan(setCommandStatementToJSON(s), frag(stmt))
	case *ast.SetTransactionIsolationLevelStatement:
		return addSpan(setTransactionIsolationLevelStatementToJSON(s), frag(stmt))
	case *ast.SetTextSizeStatement:
		return addSpan(setTextSizeStatementToJSON(s), frag(stmt))
	case *ast.SetIdentityInsertStatement:
		return addSpan(setIdentityInsertStatementToJSON(s), frag(stmt))
	case *ast.SetErrorLevelStatement:
		return addSpan(setErrorLevelStatementToJSON(s), frag(stmt))
	case *ast.CommitTransactionStatement:
		return addSpan(commitTransactionStatementToJSON(s), frag(stmt))
	case *ast.RollbackTransactionStatement:
		return addSpan(rollbackTransactionStatementToJSON(s), frag(stmt))
	case *ast.SaveTransactionStatement:
		return addSpan(saveTransactionStatementToJSON(s), frag(stmt))
	case *ast.BeginTransactionStatement:
		return addSpan(beginTransactionStatementToJSON(s), frag(stmt))
	case *ast.WaitForStatement:
		return addSpan(waitForStatementToJSON(s), frag(stmt))
	case *ast.MoveConversationStatement:
		return addSpan(moveConversationStatementToJSON(s), frag(stmt))
	case *ast.GetConversationGroupStatement:
		return addSpan(getConversationGroupStatementToJSON(s), frag(stmt))
	case *ast.TruncateTableStatement:
		return addSpan(truncateTableStatementToJSON(s), frag(stmt))
	case *ast.UseStatement:
		return addSpan(useStatementToJSON(s), frag(stmt))
	case *ast.KillStatement:
		return addSpan(killStatementToJSON(s), frag(stmt))
	case *ast.KillStatsJobStatement:
		return addSpan(killStatsJobStatementToJSON(s), frag(stmt))
	case *ast.KillQueryNotificationSubscriptionStatement:
		return addSpan(killQueryNotificationSubscriptionStatementToJSON(s), frag(stmt))
	case *ast.CloseSymmetricKeyStatement:
		return addSpan(closeSymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.CloseMasterKeyStatement:
		return addSpan(closeMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.OpenMasterKeyStatement:
		return addSpan(openMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.OpenSymmetricKeyStatement:
		return addSpan(openSymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.CheckpointStatement:
		return addSpan(checkpointStatementToJSON(s), frag(stmt))
	case *ast.ReconfigureStatement:
		return addSpan(reconfigureStatementToJSON(s), frag(stmt))
	case *ast.ShutdownStatement:
		return addSpan(shutdownStatementToJSON(s), frag(stmt))
	case *ast.SetUserStatement:
		return addSpan(setUserStatementToJSON(s), frag(stmt))
	case *ast.LineNoStatement:
		return addSpan(lineNoStatementToJSON(s), frag(stmt))
	case *ast.RaiseErrorStatement:
		return addSpan(raiseErrorStatementToJSON(s), frag(stmt))
	case *ast.ReadTextStatement:
		return addSpan(readTextStatementToJSON(s), frag(stmt))
	case *ast.WriteTextStatement:
		return addSpan(writeTextStatementToJSON(s), frag(stmt))
	case *ast.UpdateTextStatement:
		return addSpan(updateTextStatementToJSON(s), frag(stmt))
	case *ast.GoToStatement:
		return addSpan(goToStatementToJSON(s), frag(stmt))
	case *ast.LabelStatement:
		return addSpan(labelStatementToJSON(s), frag(stmt))
	case *ast.CreateDefaultStatement:
		return addSpan(createDefaultStatementToJSON(s), frag(stmt))
	case *ast.CreateMasterKeyStatement:
		return addSpan(createMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.AlterMasterKeyStatement:
		return addSpan(alterMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.AlterSchemaStatement:
		return addSpan(alterSchemaStatementToJSON(s), frag(stmt))
	case *ast.AlterRoleStatement:
		return addSpan(alterRoleStatementToJSON(s), frag(stmt))
	case *ast.CreateServerRoleStatement:
		return addSpan(createServerRoleStatementToJSON(s), frag(stmt))
	case *ast.AlterServerRoleStatement:
		return addSpan(alterServerRoleStatementToJSON(s), frag(stmt))
	case *ast.CreateAvailabilityGroupStatement:
		return addSpan(createAvailabilityGroupStatementToJSON(s), frag(stmt))
	case *ast.AlterAvailabilityGroupStatement:
		return addSpan(alterAvailabilityGroupStatementToJSON(s), frag(stmt))
	case *ast.CreateServerAuditStatement:
		return addSpan(createServerAuditStatementToJSON(s), frag(stmt))
	case *ast.AlterServerAuditStatement:
		return addSpan(alterServerAuditStatementToJSON(s), frag(stmt))
	case *ast.CreateServerAuditSpecificationStatement:
		return addSpan(createServerAuditSpecificationStatementToJSON(s), frag(stmt))
	case *ast.AlterServerAuditSpecificationStatement:
		return addSpan(alterServerAuditSpecificationStatementToJSON(s), frag(stmt))
	case *ast.CreateDatabaseAuditSpecificationStatement:
		return addSpan(createDatabaseAuditSpecificationStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseAuditSpecificationStatement:
		return addSpan(alterDatabaseAuditSpecificationStatementToJSON(s), frag(stmt))
	case *ast.AlterRemoteServiceBindingStatement:
		return addSpan(alterRemoteServiceBindingStatementToJSON(s), frag(stmt))
	case *ast.AlterXmlSchemaCollectionStatement:
		return addSpan(alterXmlSchemaCollectionStatementToJSON(s), frag(stmt))
	case *ast.AlterServerConfigurationSetSoftNumaStatement:
		return addSpan(alterServerConfigurationSetSoftNumaStatementToJSON(s), frag(stmt))
	case *ast.AlterServerConfigurationSetExternalAuthenticationStatement:
		return addSpan(alterServerConfigurationSetExternalAuthenticationStatementToJSON(s), frag(stmt))
	case *ast.AlterServerConfigurationSetDiagnosticsLogStatement:
		return addSpan(alterServerConfigurationSetDiagnosticsLogStatementToJSON(s), frag(stmt))
	case *ast.AlterServerConfigurationSetFailoverClusterPropertyStatement:
		return addSpan(alterServerConfigurationSetFailoverClusterPropertyStatementToJSON(s), frag(stmt))
	case *ast.AlterServerConfigurationSetBufferPoolExtensionStatement:
		return addSpan(alterServerConfigurationSetBufferPoolExtensionStatementToJSON(s), frag(stmt))
	case *ast.AlterServerConfigurationSetHadrClusterStatement:
		return addSpan(alterServerConfigurationSetHadrClusterStatementToJSON(s), frag(stmt))
	case *ast.AlterServerConfigurationStatement:
		return addSpan(alterServerConfigurationStatementToJSON(s), frag(stmt))
	case *ast.AlterLoginAddDropCredentialStatement:
		return addSpan(alterLoginAddDropCredentialStatementToJSON(s), frag(stmt))
	case *ast.TryCatchStatement:
		return addSpan(tryCatchStatementToJSON(s), frag(stmt))
	case *ast.SendStatement:
		return addSpan(sendStatementToJSON(s), frag(stmt))
	case *ast.ReceiveStatement:
		return addSpan(receiveStatementToJSON(s), frag(stmt))
	case *ast.CreateCredentialStatement:
		return addSpan(createCredentialStatementToJSON(s), frag(stmt))
	case *ast.CreateXmlSchemaCollectionStatement:
		return addSpan(createXmlSchemaCollectionStatementToJSON(s), frag(stmt))
	case *ast.CreateSearchPropertyListStatement:
		return addSpan(createSearchPropertyListStatementToJSON(s), frag(stmt))
	case *ast.CreateExternalDataSourceStatement:
		return addSpan(createExternalDataSourceStatementToJSON(s), frag(stmt))
	case *ast.CreateExternalFileFormatStatement:
		return addSpan(createExternalFileFormatStatementToJSON(s), frag(stmt))
	case *ast.CreateExternalTableStatement:
		return addSpan(createExternalTableStatementToJSON(s), frag(stmt))
	case *ast.CreateExternalLanguageStatement:
		return addSpan(createExternalLanguageStatementToJSON(s), frag(stmt))
	case *ast.CreateExternalLibraryStatement:
		return addSpan(createExternalLibraryStatementToJSON(s), frag(stmt))
	case *ast.CreateEventSessionStatement:
		return addSpan(createEventSessionStatementToJSON(s), frag(stmt))
	case *ast.RestoreStatement:
		return addSpan(restoreStatementToJSON(s), frag(stmt))
	case *ast.BackupDatabaseStatement:
		return addSpan(backupDatabaseStatementToJSON(s), frag(stmt))
	case *ast.BackupTransactionLogStatement:
		return addSpan(backupTransactionLogStatementToJSON(s), frag(stmt))
	case *ast.BackupCertificateStatement:
		return addSpan(backupCertificateStatementToJSON(s), frag(stmt))
	case *ast.BackupServiceMasterKeyStatement:
		return addSpan(backupServiceMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.BackupMasterKeyStatement:
		return addSpan(backupMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.RestoreServiceMasterKeyStatement:
		return addSpan(restoreServiceMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.RestoreMasterKeyStatement:
		return addSpan(restoreMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.CreateUserStatement:
		return addSpan(createUserStatementToJSON(s), frag(stmt))
	case *ast.CreateAggregateStatement:
		return addSpan(createAggregateStatementToJSON(s), frag(stmt))
	case *ast.CreateColumnStoreIndexStatement:
		return addSpan(createColumnStoreIndexStatementToJSON(s), frag(stmt))
	case *ast.AlterFunctionStatement:
		return addSpan(alterFunctionStatementToJSON(s), frag(stmt))
	case *ast.CreateFunctionStatement:
		return addSpan(createFunctionStatementToJSON(s), frag(stmt))
	case *ast.CreateOrAlterFunctionStatement:
		return addSpan(createOrAlterFunctionStatementToJSON(s), frag(stmt))
	case *ast.AlterTriggerStatement:
		return addSpan(alterTriggerStatementToJSON(s), frag(stmt))
	case *ast.CreateTriggerStatement:
		return addSpan(createTriggerStatementToJSON(s), frag(stmt))
	case *ast.CreateOrAlterTriggerStatement:
		return addSpan(createOrAlterTriggerStatementToJSON(s), frag(stmt))
	case *ast.EnableDisableTriggerStatement:
		return addSpan(enableDisableTriggerStatementToJSON(s), frag(stmt))
	case *ast.EndConversationStatement:
		return addSpan(endConversationStatementToJSON(s), frag(stmt))
	case *ast.CreateDatabaseStatement:
		return addSpan(createDatabaseStatementToJSON(s), frag(stmt))
	case *ast.CreateDatabaseEncryptionKeyStatement:
		return addSpan(createDatabaseEncryptionKeyStatementToJSON(s), frag(stmt))
	case *ast.AlterDatabaseEncryptionKeyStatement:
		return addSpan(alterDatabaseEncryptionKeyStatementToJSON(s), frag(stmt))
	case *ast.DropDatabaseEncryptionKeyStatement:
		return addSpan(dropDatabaseEncryptionKeyStatementToJSON(s), frag(stmt))
	case *ast.CreateLoginStatement:
		return addSpan(createLoginStatementToJSON(s), frag(stmt))
	case *ast.AlterLoginEnableDisableStatement:
		return addSpan(alterLoginEnableDisableStatementToJSON(s), frag(stmt))
	case *ast.AlterLoginOptionsStatement:
		return addSpan(alterLoginOptionsStatementToJSON(s), frag(stmt))
	case *ast.DropLoginStatement:
		return addSpan(dropLoginStatementToJSON(s), frag(stmt))
	case *ast.CreateIndexStatement:
		return addSpan(createIndexStatementToJSON(s), frag(stmt))
	case *ast.CreateSpatialIndexStatement:
		return addSpan(createSpatialIndexStatementToJSON(s), frag(stmt))
	case *ast.CreateAsymmetricKeyStatement:
		return addSpan(createAsymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.CreateSymmetricKeyStatement:
		return addSpan(createSymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.CreateCertificateStatement:
		return addSpan(createCertificateStatementToJSON(s), frag(stmt))
	case *ast.CreateMessageTypeStatement:
		return addSpan(createMessageTypeStatementToJSON(s), frag(stmt))
	case *ast.CreateServiceStatement:
		return addSpan(createServiceStatementToJSON(s), frag(stmt))
	case *ast.CreateQueueStatement:
		return addSpan(createQueueStatementToJSON(s), frag(stmt))
	case *ast.CreateRouteStatement:
		return addSpan(createRouteStatementToJSON(s), frag(stmt))
	case *ast.CreateEndpointStatement:
		return addSpan(createEndpointStatementToJSON(s), frag(stmt))
	case *ast.CreateAssemblyStatement:
		return addSpan(createAssemblyStatementToJSON(s), frag(stmt))
	case *ast.CreateApplicationRoleStatement:
		return addSpan(createApplicationRoleStatementToJSON(s), frag(stmt))
	case *ast.CreateFulltextCatalogStatement:
		return addSpan(createFulltextCatalogStatementToJSON(s), frag(stmt))
	case *ast.CreateFulltextIndexStatement:
		return addSpan(createFulltextIndexStatementToJSON(s), frag(stmt))
	case *ast.CreateRemoteServiceBindingStatement:
		return addSpan(createRemoteServiceBindingStatementToJSON(s), frag(stmt))
	case *ast.CreateStatisticsStatement:
		return addSpan(createStatisticsStatementToJSON(s), frag(stmt))
	case *ast.CreateTypeStatement:
		return addSpan(createTypeStatementToJSON(s), frag(stmt))
	case *ast.CreateTypeUddtStatement:
		return addSpan(createTypeUddtStatementToJSON(s), frag(stmt))
	case *ast.CreateTypeUdtStatement:
		return addSpan(createTypeUdtStatementToJSON(s), frag(stmt))
	case *ast.CreateTypeTableStatement:
		return addSpan(createTypeTableStatementToJSON(s), frag(stmt))
	case *ast.CreateXmlIndexStatement:
		return addSpan(createXmlIndexStatementToJSON(s), frag(stmt))
	case *ast.CreateSelectiveXmlIndexStatement:
		return addSpan(createSelectiveXmlIndexStatementToJSON(s), frag(stmt))
	case *ast.CreatePartitionFunctionStatement:
		return addSpan(createPartitionFunctionStatementToJSON(s), frag(stmt))
	case *ast.CreateEventNotificationStatement:
		return addSpan(createEventNotificationStatementToJSON(s), frag(stmt))
	case *ast.CreateSecurityPolicyStatement:
		return addSpan(createSecurityPolicyStatementToJSON(s), frag(stmt))
	case *ast.AlterSecurityPolicyStatement:
		return addSpan(alterSecurityPolicyStatementToJSON(s), frag(stmt))
	case *ast.AlterIndexStatement:
		return addSpan(alterIndexStatementToJSON(s), frag(stmt))
	case *ast.DropDatabaseStatement:
		return addSpan(dropDatabaseStatementToJSON(s), frag(stmt))
	case *ast.DropTableStatement:
		return addSpan(dropTableStatementToJSON(s), frag(stmt))
	case *ast.DropViewStatement:
		return addSpan(dropViewStatementToJSON(s), frag(stmt))
	case *ast.DropProcedureStatement:
		return addSpan(dropProcedureStatementToJSON(s), frag(stmt))
	case *ast.DropFunctionStatement:
		return addSpan(dropFunctionStatementToJSON(s), frag(stmt))
	case *ast.DropTriggerStatement:
		return addSpan(dropTriggerStatementToJSON(s), frag(stmt))
	case *ast.DropIndexStatement:
		return addSpan(dropIndexStatementToJSON(s), frag(stmt))
	case *ast.DropStatisticsStatement:
		return addSpan(dropStatisticsStatementToJSON(s), frag(stmt))
	case *ast.DropDefaultStatement:
		return addSpan(dropDefaultStatementToJSON(s), frag(stmt))
	case *ast.DropRuleStatement:
		return addSpan(dropRuleStatementToJSON(s), frag(stmt))
	case *ast.DropSchemaStatement:
		return addSpan(dropSchemaStatementToJSON(s), frag(stmt))
	case *ast.DropPartitionFunctionStatement:
		return addSpan(dropPartitionFunctionStatementToJSON(s), frag(stmt))
	case *ast.DropPartitionSchemeStatement:
		return addSpan(dropPartitionSchemeStatementToJSON(s), frag(stmt))
	case *ast.DropApplicationRoleStatement:
		return addSpan(dropApplicationRoleStatementToJSON(s), frag(stmt))
	case *ast.DropCertificateStatement:
		return addSpan(dropCertificateStatementToJSON(s), frag(stmt))
	case *ast.DropMasterKeyStatement:
		return addSpan(dropMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.DropXmlSchemaCollectionStatement:
		return addSpan(dropXmlSchemaCollectionStatementToJSON(s), frag(stmt))
	case *ast.DropContractStatement:
		return addSpan(dropContractStatementToJSON(s), frag(stmt))
	case *ast.DropEndpointStatement:
		return addSpan(dropEndpointStatementToJSON(s), frag(stmt))
	case *ast.DropMessageTypeStatement:
		return addSpan(dropMessageTypeStatementToJSON(s), frag(stmt))
	case *ast.DropQueueStatement:
		return addSpan(dropQueueStatementToJSON(s), frag(stmt))
	case *ast.DropRemoteServiceBindingStatement:
		return addSpan(dropRemoteServiceBindingStatementToJSON(s), frag(stmt))
	case *ast.DropRouteStatement:
		return addSpan(dropRouteStatementToJSON(s), frag(stmt))
	case *ast.DropServiceStatement:
		return addSpan(dropServiceStatementToJSON(s), frag(stmt))
	case *ast.DropEventNotificationStatement:
		return addSpan(dropEventNotificationStatementToJSON(s), frag(stmt))
	case *ast.DropEventSessionStatement:
		return addSpan(dropEventSessionStatementToJSON(s), frag(stmt))
	case *ast.AlterTableTriggerModificationStatement:
		return addSpan(alterTableTriggerModificationStatementToJSON(s), frag(stmt))
	case *ast.AlterTableFileTableNamespaceStatement:
		return addSpan(alterTableFileTableNamespaceStatementToJSON(s), frag(stmt))
	case *ast.AlterTableSwitchStatement:
		return addSpan(alterTableSwitchStatementToJSON(s), frag(stmt))
	case *ast.AlterTableConstraintModificationStatement:
		return addSpan(alterTableConstraintModificationStatementToJSON(s), frag(stmt))
	case *ast.AlterTableSetStatement:
		return addSpan(alterTableSetStatementToJSON(s), frag(stmt))
	case *ast.AlterTableRebuildStatement:
		return addSpan(alterTableRebuildStatementToJSON(s), frag(stmt))
	case *ast.AlterTableAlterPartitionStatement:
		return addSpan(alterTableAlterPartitionStatementToJSON(s), frag(stmt))
	case *ast.AlterTableChangeTrackingModificationStatement:
		return addSpan(alterTableChangeTrackingStatementToJSON(s), frag(stmt))
	case *ast.InsertBulkStatement:
		return addSpan(insertBulkStatementToJSON(s), frag(stmt))
	case *ast.BulkInsertStatement:
		return addSpan(bulkInsertStatementToJSON(s), frag(stmt))
	case *ast.CopyStatement:
		return addSpan(copyStatementToJSON(s), frag(stmt))
	case *ast.AlterUserStatement:
		return addSpan(alterUserStatementToJSON(s), frag(stmt))
	case *ast.AlterRouteStatement:
		return addSpan(alterRouteStatementToJSON(s), frag(stmt))
	case *ast.AlterSearchPropertyListStatement:
		return addSpan(alterSearchPropertyListStatementToJSON(s), frag(stmt))
	case *ast.AlterAssemblyStatement:
		return addSpan(alterAssemblyStatementToJSON(s), frag(stmt))
	case *ast.AlterEndpointStatement:
		return addSpan(alterEndpointStatementToJSON(s), frag(stmt))
	case *ast.AlterEventSessionStatement:
		return addSpan(alterEventSessionStatementToJSON(s), frag(stmt))
	case *ast.AlterAuthorizationStatement:
		return addSpan(alterAuthorizationStatementToJSON(s), frag(stmt))
	case *ast.AlterServiceStatement:
		return addSpan(alterServiceStatementToJSON(s), frag(stmt))
	case *ast.AlterCertificateStatement:
		return addSpan(alterCertificateStatementToJSON(s), frag(stmt))
	case *ast.AlterApplicationRoleStatement:
		return addSpan(alterApplicationRoleStatementToJSON(s), frag(stmt))
	case *ast.AlterAsymmetricKeyStatement:
		return addSpan(alterAsymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.AlterQueueStatement:
		return addSpan(alterQueueStatementToJSON(s), frag(stmt))
	case *ast.AlterPartitionSchemeStatement:
		return addSpan(alterPartitionSchemeStatementToJSON(s), frag(stmt))
	case *ast.AlterPartitionFunctionStatement:
		return addSpan(alterPartitionFunctionStatementToJSON(s), frag(stmt))
	case *ast.AlterFulltextCatalogStatement:
		return addSpan(alterFulltextCatalogStatementToJSON(s), frag(stmt))
	case *ast.CreateFullTextCatalogStatement:
		return addSpan(createFullTextCatalogStatementToJSON(s), frag(stmt))
	case *ast.AlterFulltextIndexStatement:
		return addSpan(alterFulltextIndexStatementToJSON(s), frag(stmt))
	case *ast.CreateFullTextStopListStatement:
		return addSpan(createFullTextStopListStatementToJSON(s), frag(stmt))
	case *ast.AlterFullTextStopListStatement:
		return addSpan(alterFullTextStopListStatementToJSON(s), frag(stmt))
	case *ast.DropFullTextStopListStatement:
		return addSpan(dropFullTextStopListStatementToJSON(s), frag(stmt))
	case *ast.DropFullTextCatalogStatement:
		return addSpan(dropFullTextCatalogStatementToJSON(s), frag(stmt))
	case *ast.DropFulltextIndexStatement:
		return addSpan(dropFulltextIndexStatementToJSON(s), frag(stmt))
	case *ast.AlterSymmetricKeyStatement:
		return addSpan(alterSymmetricKeyStatementToJSON(s), frag(stmt))
	case *ast.AlterServiceMasterKeyStatement:
		return addSpan(alterServiceMasterKeyStatementToJSON(s), frag(stmt))
	case *ast.RenameEntityStatement:
		return addSpan(renameEntityStatementToJSON(s), frag(stmt))
	case *ast.OpenCursorStatement:
		return addSpan(openCursorStatementToJSON(s), frag(stmt))
	case *ast.CloseCursorStatement:
		return addSpan(closeCursorStatementToJSON(s), frag(stmt))
	case *ast.DeallocateCursorStatement:
		return addSpan(deallocateCursorStatementToJSON(s), frag(stmt))
	case *ast.FetchCursorStatement:
		return addSpan(fetchCursorStatementToJSON(s), frag(stmt))
	case *ast.DeclareCursorStatement:
		return addSpan(declareCursorStatementToJSON(s), frag(stmt))
	case *ast.AddSignatureStatement:
		return addSpan(addSignatureStatementToJSON(s), frag(stmt))
	case *ast.DropSignatureStatement:
		return addSpan(dropSignatureStatementToJSON(s), frag(stmt))
	case *ast.AddSensitivityClassificationStatement:
		return addSpan(addSensitivityClassificationStatementToJSON(s), frag(stmt))
	case *ast.DropSensitivityClassificationStatement:
		return addSpan(dropSensitivityClassificationStatementToJSON(s), frag(stmt))
	default:
		return addSpan(jsonNode{"$type": "UnknownStatement"}, frag(stmt))
	}
}

func revertStatementToJSON(s *ast.RevertStatement) jsonNode {
	node := jsonNode{
		"$type": "RevertStatement",
	}
	if s.Cookie != nil {
		node["Cookie"] = scalarExpressionToJSON(s.Cookie)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func dropSearchPropertyListStatementToJSON(s *ast.DropSearchPropertyListStatement) jsonNode {
	node := jsonNode{
		"$type": "DropSearchPropertyListStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func dropServerRoleStatementToJSON(s *ast.DropServerRoleStatement) jsonNode {
	node := jsonNode{
		"$type": "DropServerRoleStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func dropAvailabilityGroupStatementToJSON(s *ast.DropAvailabilityGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "DropAvailabilityGroupStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func dropFederationStatementToJSON(s *ast.DropFederationStatement) jsonNode {
	node := jsonNode{
		"$type": "DropFederationStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(e))
}

func dropClusteredConstraintOptionToJSON(o ast.DropClusteredConstraintOption) jsonNode {
	switch opt := o.(type) {
	case *ast.DropClusteredConstraintStateOption:
		return addSpan(jsonNode{
			"$type":       "DropClusteredConstraintStateOption",
			"OptionState": opt.OptionState,
			"OptionKind":  opt.OptionKind,
		}, frag(o))
	case *ast.DropClusteredConstraintMoveOption:
		node := jsonNode{
			"$type":      "DropClusteredConstraintMoveOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.OptionValue != nil {
			node["OptionValue"] = fileGroupOrPartitionSchemeToJSON(opt.OptionValue)
		}
		return addSpan(node, frag(o))
	case *ast.DropClusteredConstraintValueOption:
		node := jsonNode{
			"$type":      "DropClusteredConstraintValueOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.OptionValue != nil {
			node["OptionValue"] = scalarExpressionToJSON(opt.OptionValue)
		}
		return addSpan(node, frag(o))
	case *ast.DropClusteredConstraintWaitAtLowPriorityLockOption:
		node := jsonNode{
			"$type":      "DropClusteredConstraintWaitAtLowPriorityLockOption",
			"OptionKind": opt.OptionKind,
		}
		if len(opt.Options) > 0 {
			options := make([]jsonNode, len(opt.Options))
			for i, o := range opt.Options {
				options[i] = lowPriorityLockWaitOptionToJSON(o)
			}
			node["Options"] = options
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownDropClusteredConstraintOption"}, frag(o))
	}
}

func lowPriorityLockWaitOptionToJSON(o ast.LowPriorityLockWaitOption) jsonNode {
	switch opt := o.(type) {
	case *ast.LowPriorityLockWaitMaxDurationOption:
		node := jsonNode{
			"$type":      "LowPriorityLockWaitMaxDurationOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.MaxDuration != nil {
			node["MaxDuration"] = scalarExpressionToJSON(opt.MaxDuration)
		}
		if opt.Unit != "" {
			node["Unit"] = opt.Unit
		}
		return addSpan(node, frag(o))
	case *ast.LowPriorityLockWaitAbortAfterWaitOption:
		return addSpan(jsonNode{
			"$type":          "LowPriorityLockWaitAbortAfterWaitOption",
			"OptionKind":     opt.OptionKind,
			"AbortAfterWait": opt.AbortAfterWait,
		}, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownLowPriorityLockWaitOption"}, frag(o))
	}
}

func onlineIndexLowPriorityLockWaitOptionToJSON(o *ast.OnlineIndexLowPriorityLockWaitOption) jsonNode {
	node := jsonNode{
		"$type": "OnlineIndexLowPriorityLockWaitOption",
	}
	if len(o.Options) > 0 {
		options := make([]jsonNode, len(o.Options))
		for i, opt := range o.Options {
			options[i] = lowPriorityLockWaitOptionToJSON(opt)
		}
		node["Options"] = options
	}
	return addSpan(node, frag(o))
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
	return addSpan(node, frag(fg))
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
	return addSpan(node, frag(s))
}

func indexExpressionOptionToJSON(o *ast.IndexExpressionOption) jsonNode {
	node := jsonNode{
		"$type":      "IndexExpressionOption",
		"OptionKind": o.OptionKind,
	}
	if o.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(o.Expression)
	}
	return addSpan(node, frag(o))
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
	return addSpan(node, frag(s))
}

func alterTableAlterColumnStatementToJSON(s *ast.AlterTableAlterColumnStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterTableAlterColumnStatement",
	}
	if s.ColumnIdentifier != nil {
		node["ColumnIdentifier"] = identifierToJSON(s.ColumnIdentifier)
	}
	if s.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(s.DataType)
	}
	node["AlterTableAlterColumnOption"] = s.AlterTableAlterColumnOption
	if s.StorageOptions != nil {
		node["StorageOptions"] = columnStorageOptionsToJSON(s.StorageOptions)
	}
	node["IsHidden"] = s.IsHidden
	if s.Encryption != nil {
		node["Encryption"] = columnEncryptionDefinitionToJSON(s.Encryption)
	}
	if s.Collation != nil {
		node["Collation"] = identifierToJSON(s.Collation)
	}
	node["IsMasked"] = s.IsMasked
	if s.MaskingFunction != nil {
		node["MaskingFunction"] = scalarExpressionToJSON(s.MaskingFunction)
	}
	if s.GeneratedAlways != "" {
		node["GeneratedAlways"] = s.GeneratedAlways
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = indexOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return addSpan(node, frag(s))
}

func columnEncryptionDefinitionToJSON(e *ast.ColumnEncryptionDefinition) jsonNode {
	node := jsonNode{
		"$type": "ColumnEncryptionDefinition",
	}
	if len(e.Parameters) > 0 {
		params := make([]jsonNode, len(e.Parameters))
		for i, p := range e.Parameters {
			params[i] = columnEncryptionParameterToJSON(p)
		}
		node["Parameters"] = params
	}
	return addSpan(node, frag(e))
}

func columnEncryptionParameterToJSON(p ast.ColumnEncryptionParameter) jsonNode {
	switch param := p.(type) {
	case *ast.ColumnEncryptionKeyNameParameter:
		node := jsonNode{
			"$type":         "ColumnEncryptionKeyNameParameter",
			"ParameterKind": param.ParameterKind,
		}
		if param.Name != nil {
			node["Name"] = identifierToJSON(param.Name)
		}
		return addSpan(node, frag(p))
	case *ast.ColumnEncryptionTypeParameter:
		return addSpan(jsonNode{
			"$type":          "ColumnEncryptionTypeParameter",
			"EncryptionType": param.EncryptionType,
			"ParameterKind":  param.ParameterKind,
		}, frag(p))
	case *ast.ColumnEncryptionAlgorithmParameter:
		node := jsonNode{
			"$type":         "ColumnEncryptionAlgorithmParameter",
			"ParameterKind": param.ParameterKind,
		}
		if param.EncryptionAlgorithm != nil {
			node["EncryptionAlgorithm"] = scalarExpressionToJSON(param.EncryptionAlgorithm)
		}
		return addSpan(node, frag(p))
	default:
		return addSpan(jsonNode{"$type": "Unknown"}, frag(p))
	}
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func contractMessageToJSON(m *ast.ContractMessage) jsonNode {
	node := jsonNode{
		"$type":  "ContractMessage",
		"SentBy": m.SentBy,
	}
	if m.Name != nil {
		node["Name"] = identifierToJSON(m.Name)
	}
	return addSpan(node, frag(m))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func alterDatabaseSetStatementToJSON(s *ast.AlterDatabaseSetStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseSetStatement",
	}
	if s.Termination != nil {
		termNode := jsonNode{
			"$type":             "AlterDatabaseTermination",
			"ImmediateRollback": s.Termination.ImmediateRollback,
			"NoWait":            s.Termination.NoWait,
		}
		if s.Termination.RollbackAfter != nil {
			termNode["RollbackAfter"] = scalarExpressionToJSON(s.Termination.RollbackAfter)
		}
		node["Termination"] = addSpan(termNode, frag(s.Termination))
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = databaseOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	node["WithManualCutover"] = s.WithManualCutover
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = s.UseCurrent
	return addSpan(node, frag(s))
}

func databaseOptionToJSON(opt ast.DatabaseOption) jsonNode {
	switch o := opt.(type) {
	case *ast.AcceleratedDatabaseRecoveryDatabaseOption:
		return addSpan(jsonNode{
			"$type":       "AcceleratedDatabaseRecoveryDatabaseOption",
			"OptionKind":  o.OptionKind,
			"OptionState": o.OptionState,
		}, frag(opt))
	case *ast.OnOffDatabaseOption:
		return addSpan(jsonNode{
			"$type":       "OnOffDatabaseOption",
			"OptionKind":  o.OptionKind,
			"OptionState": o.OptionState,
		}, frag(opt))
	case *ast.AutomaticTuningDatabaseOption:
		node := jsonNode{
			"$type": "AutomaticTuningDatabaseOption",
		}
		if o.AutomaticTuningState != "" {
			node["AutomaticTuningState"] = o.AutomaticTuningState
		}
		if len(o.Options) > 0 {
			opts := make([]jsonNode, len(o.Options))
			for i, subOpt := range o.Options {
				opts[i] = automaticTuningOptionToJSON(subOpt)
			}
			node["Options"] = opts
		}
		node["OptionKind"] = o.OptionKind
		return addSpan(node, frag(opt))
	case *ast.DelayedDurabilityDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "DelayedDurabilityDatabaseOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.AutoCreateStatisticsDatabaseOption:
		node := jsonNode{
			"$type":          "AutoCreateStatisticsDatabaseOption",
			"HasIncremental": o.HasIncremental,
		}
		if o.IncrementalState != "" {
			node["IncrementalState"] = o.IncrementalState
		} else {
			node["IncrementalState"] = "NotSet"
		}
		node["OptionState"] = o.OptionState
		node["OptionKind"] = o.OptionKind
		return addSpan(node, frag(opt))
	case *ast.MaxSizeDatabaseOption:
		node := jsonNode{
			"$type": "MaxSizeDatabaseOption",
		}
		if o.MaxSize != nil {
			node["MaxSize"] = scalarExpressionToJSON(o.MaxSize)
		}
		if o.Units != "" {
			node["Units"] = o.Units
		}
		if o.OptionKind != "" {
			node["OptionKind"] = o.OptionKind
		}
		return addSpan(node, frag(opt))
	case *ast.LiteralDatabaseOption:
		node := jsonNode{
			"$type": "LiteralDatabaseOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		if o.OptionKind != "" {
			node["OptionKind"] = o.OptionKind
		}
		return addSpan(node, frag(opt))
	case *ast.ElasticPoolSpecification:
		node := jsonNode{
			"$type": "ElasticPoolSpecification",
		}
		if o.ElasticPoolName != nil {
			node["ElasticPoolName"] = identifierToJSON(o.ElasticPoolName)
		}
		if o.OptionKind != "" {
			node["OptionKind"] = o.OptionKind
		}
		return addSpan(node, frag(opt))
	case *ast.RemoteDataArchiveDatabaseOption:
		node := jsonNode{
			"$type":       "RemoteDataArchiveDatabaseOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		if len(o.Settings) > 0 {
			settings := make([]jsonNode, len(o.Settings))
			for i, setting := range o.Settings {
				settings[i] = remoteDataArchiveDbSettingToJSON(setting)
			}
			node["Settings"] = settings
		}
		return addSpan(node, frag(opt))
	case *ast.ChangeTrackingDatabaseOption:
		node := jsonNode{
			"$type":       "ChangeTrackingDatabaseOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		if len(o.Details) > 0 {
			details := make([]jsonNode, len(o.Details))
			for i, detail := range o.Details {
				details[i] = changeTrackingOptionDetailToJSON(detail)
			}
			node["Details"] = details
		}
		return addSpan(node, frag(opt))
	case *ast.RecoveryDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "RecoveryDatabaseOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.CursorDefaultDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "CursorDefaultDatabaseOption",
			"IsLocal":    o.IsLocal,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.SimpleDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "DatabaseOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.GenericDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "DatabaseOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.PageVerifyDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "PageVerifyDatabaseOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.PartnerDatabaseOption:
		node := jsonNode{
			"$type":         "PartnerDatabaseOption",
			"PartnerOption": o.PartnerOption,
			"OptionKind":    o.OptionKind,
		}
		if o.PartnerServer != nil {
			node["PartnerServer"] = scalarExpressionToJSON(o.PartnerServer)
		}
		if o.Timeout != nil {
			node["Timeout"] = scalarExpressionToJSON(o.Timeout)
		}
		return addSpan(node, frag(opt))
	case *ast.WitnessDatabaseOption:
		node := jsonNode{
			"$type":      "WitnessDatabaseOption",
			"IsOff":      o.IsOff,
			"OptionKind": o.OptionKind,
		}
		if o.WitnessServer != nil {
			node["WitnessServer"] = scalarExpressionToJSON(o.WitnessServer)
		}
		return addSpan(node, frag(opt))
	case *ast.ParameterizationDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "ParameterizationDatabaseOption",
			"IsSimple":   o.IsSimple,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.ContainmentDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "ContainmentDatabaseOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.IdentifierDatabaseOption:
		node := jsonNode{
			"$type":      "IdentifierDatabaseOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = identifierToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.HadrDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "HadrDatabaseOption",
			"HadrOption": o.HadrOption,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.HadrAvailabilityGroupDatabaseOption:
		node := jsonNode{
			"$type":      "HadrAvailabilityGroupDatabaseOption",
			"HadrOption": o.HadrOption,
			"OptionKind": o.OptionKind,
		}
		if o.GroupName != nil {
			node["GroupName"] = identifierToJSON(o.GroupName)
		}
		return addSpan(node, frag(opt))
	case *ast.FileStreamDatabaseOption:
		node := jsonNode{
			"$type":      "FileStreamDatabaseOption",
			"OptionKind": o.OptionKind,
		}
		if o.NonTransactedAccess != "" {
			node["NonTransactedAccess"] = o.NonTransactedAccess
		}
		if o.DirectoryName != nil {
			node["DirectoryName"] = scalarExpressionToJSON(o.DirectoryName)
		}
		return addSpan(node, frag(opt))
	case *ast.TargetRecoveryTimeDatabaseOption:
		node := jsonNode{
			"$type":      "TargetRecoveryTimeDatabaseOption",
			"OptionKind": o.OptionKind,
			"Unit":       o.Unit,
		}
		if o.RecoveryTime != nil {
			node["RecoveryTime"] = scalarExpressionToJSON(o.RecoveryTime)
		}
		return addSpan(node, frag(opt))
	case *ast.QueryStoreDatabaseOption:
		node := jsonNode{
			"$type":    "QueryStoreDatabaseOption",
			"Clear":    o.Clear,
			"ClearAll": o.ClearAll,
		}
		if o.OptionState != "" {
			node["OptionState"] = o.OptionState
		} else {
			node["OptionState"] = "NotSet"
		}
		if len(o.Options) > 0 {
			opts := make([]jsonNode, len(o.Options))
			for i, subOpt := range o.Options {
				opts[i] = queryStoreOptionToJSON(subOpt)
			}
			node["Options"] = opts
		}
		node["OptionKind"] = o.OptionKind
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownDatabaseOption"}, frag(opt))
	}
}

func automaticTuningOptionToJSON(opt ast.AutomaticTuningOption) jsonNode {
	switch o := opt.(type) {
	case *ast.AutomaticTuningCreateIndexOption:
		return addSpan(jsonNode{
			"$type":      "AutomaticTuningCreateIndexOption",
			"OptionKind": o.OptionKind,
			"Value":      o.Value,
		}, frag(opt))
	case *ast.AutomaticTuningDropIndexOption:
		return addSpan(jsonNode{
			"$type":      "AutomaticTuningDropIndexOption",
			"OptionKind": o.OptionKind,
			"Value":      o.Value,
		}, frag(opt))
	case *ast.AutomaticTuningForceLastGoodPlanOption:
		return addSpan(jsonNode{
			"$type":      "AutomaticTuningForceLastGoodPlanOption",
			"OptionKind": o.OptionKind,
			"Value":      o.Value,
		}, frag(opt))
	case *ast.AutomaticTuningMaintainIndexOption:
		return addSpan(jsonNode{
			"$type":      "AutomaticTuningMaintainIndexOption",
			"OptionKind": o.OptionKind,
			"Value":      o.Value,
		}, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownAutomaticTuningOption"}, frag(opt))
	}
}

func queryStoreOptionToJSON(opt ast.QueryStoreOption) jsonNode {
	switch o := opt.(type) {
	case *ast.QueryStoreDesiredStateOption:
		return addSpan(jsonNode{
			"$type":                  "QueryStoreDesiredStateOption",
			"Value":                  o.Value,
			"OperationModeSpecified": o.OperationModeSpecified,
			"OptionKind":             o.OptionKind,
		}, frag(opt))
	case *ast.QueryStoreCapturePolicyOption:
		return addSpan(jsonNode{
			"$type":      "QueryStoreCapturePolicyOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.QueryStoreSizeCleanupPolicyOption:
		return addSpan(jsonNode{
			"$type":      "QueryStoreSizeCleanupPolicyOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.QueryStoreIntervalLengthOption:
		node := jsonNode{
			"$type":      "QueryStoreIntervalLengthOption",
			"OptionKind": o.OptionKind,
		}
		if o.StatsIntervalLength != nil {
			node["StatsIntervalLength"] = scalarExpressionToJSON(o.StatsIntervalLength)
		}
		return addSpan(node, frag(opt))
	case *ast.QueryStoreMaxStorageSizeOption:
		node := jsonNode{
			"$type":      "QueryStoreMaxStorageSizeOption",
			"OptionKind": o.OptionKind,
		}
		if o.MaxQdsSize != nil {
			node["MaxQdsSize"] = scalarExpressionToJSON(o.MaxQdsSize)
		}
		return addSpan(node, frag(opt))
	case *ast.QueryStoreMaxPlansPerQueryOption:
		node := jsonNode{
			"$type":      "QueryStoreMaxPlansPerQueryOption",
			"OptionKind": o.OptionKind,
		}
		if o.MaxPlansPerQuery != nil {
			node["MaxPlansPerQuery"] = scalarExpressionToJSON(o.MaxPlansPerQuery)
		}
		return addSpan(node, frag(opt))
	case *ast.QueryStoreTimeCleanupPolicyOption:
		node := jsonNode{
			"$type":      "QueryStoreTimeCleanupPolicyOption",
			"OptionKind": o.OptionKind,
		}
		if o.StaleQueryThreshold != nil {
			node["StaleQueryThreshold"] = scalarExpressionToJSON(o.StaleQueryThreshold)
		}
		return addSpan(node, frag(opt))
	case *ast.QueryStoreWaitStatsCaptureOption:
		return addSpan(jsonNode{
			"$type":       "QueryStoreWaitStatsCaptureOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}, frag(opt))
	case *ast.QueryStoreDataFlushIntervalOption:
		node := jsonNode{
			"$type":      "QueryStoreDataFlushIntervalOption",
			"OptionKind": o.OptionKind,
		}
		if o.FlushInterval != nil {
			node["FlushInterval"] = scalarExpressionToJSON(o.FlushInterval)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownQueryStoreOption"}, frag(opt))
	}
}

func remoteDataArchiveDbSettingToJSON(setting ast.RemoteDataArchiveDbSetting) jsonNode {
	switch s := setting.(type) {
	case *ast.RemoteDataArchiveDbServerSetting:
		node := jsonNode{
			"$type":       "RemoteDataArchiveDbServerSetting",
			"SettingKind": s.SettingKind,
		}
		if s.Server != nil {
			node["Server"] = scalarExpressionToJSON(s.Server)
		}
		return addSpan(node, frag(setting))
	case *ast.RemoteDataArchiveDbCredentialSetting:
		node := jsonNode{
			"$type":       "RemoteDataArchiveDbCredentialSetting",
			"SettingKind": s.SettingKind,
		}
		if s.Credential != nil {
			node["Credential"] = identifierToJSON(s.Credential)
		}
		return addSpan(node, frag(setting))
	case *ast.RemoteDataArchiveDbFederatedServiceAccountSetting:
		return addSpan(jsonNode{
			"$type":       "RemoteDataArchiveDbFederatedServiceAccountSetting",
			"IsOn":        s.IsOn,
			"SettingKind": s.SettingKind,
		}, frag(setting))
	default:
		return addSpan(jsonNode{"$type": "UnknownRemoteDataArchiveDbSetting"}, frag(setting))
	}
}

func changeTrackingOptionDetailToJSON(detail ast.ChangeTrackingOptionDetail) jsonNode {
	switch d := detail.(type) {
	case *ast.AutoCleanupChangeTrackingOptionDetail:
		return addSpan(jsonNode{
			"$type": "AutoCleanupChangeTrackingOptionDetail",
			"IsOn":  d.IsOn,
		}, frag(detail))
	case *ast.ChangeRetentionChangeTrackingOptionDetail:
		node := jsonNode{
			"$type": "ChangeRetentionChangeTrackingOptionDetail",
			"Unit":  d.Unit,
		}
		if d.RetentionPeriod != nil {
			node["RetentionPeriod"] = scalarExpressionToJSON(d.RetentionPeriod)
		}
		return addSpan(node, frag(detail))
	default:
		return addSpan(jsonNode{"$type": "UnknownChangeTrackingOptionDetail"}, frag(detail))
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
			options[i] = indexOptionToJSON(o)
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
	if idx.FilterPredicate != nil {
		node["FilterPredicate"] = booleanExpressionToJSON(idx.FilterPredicate)
	}
	if idx.OnFileGroupOrPartitionScheme != nil {
		node["OnFileGroupOrPartitionScheme"] = fileGroupOrPartitionSchemeToJSON(idx.OnFileGroupOrPartitionScheme)
	}
	if idx.FileStreamOn != nil {
		node["FileStreamOn"] = identifierOrValueExpressionToJSON(idx.FileStreamOn)
	}
	return addSpan(node, frag(idx))
}

func indexTypeToJSON(t *ast.IndexType) jsonNode {
	node := jsonNode{
		"$type": "IndexType",
	}
	if t.IndexTypeKind != "" {
		node["IndexTypeKind"] = t.IndexTypeKind
	}
	return addSpan(node, frag(t))
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
	return addSpan(node, frag(c))
}

func printStatementToJSON(s *ast.PrintStatement) jsonNode {
	node := jsonNode{
		"$type": "PrintStatement",
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	if s.On != nil {
		node["On"] = identifierToJSON(s.On)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	if s.WithCtesAndXmlNamespaces != nil {
		node["WithCtesAndXmlNamespaces"] = withCtesAndXmlNamespacesToJSON(s.WithCtesAndXmlNamespaces)
	}
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(h))
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
		return addSpan(node, frag(h))
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
		return addSpan(node, frag(h))
	case *ast.TableHintsOptimizerHint:
		node := jsonNode{
			"$type": "TableHintsOptimizerHint",
		}
		if hint.ObjectName != nil {
			node["ObjectName"] = schemaObjectNameToJSON(hint.ObjectName)
		}
		if len(hint.TableHints) > 0 {
			hints := make([]jsonNode, len(hint.TableHints))
			for i, h := range hint.TableHints {
				hints[i] = tableHintToJSON(h)
			}
			node["TableHints"] = hints
		}
		if hint.HintKind != "" {
			node["HintKind"] = hint.HintKind
		}
		return addSpan(node, frag(h))
	case *ast.UseHintList:
		node := jsonNode{
			"$type": "UseHintList",
		}
		if len(hint.Hints) > 0 {
			hints := make([]jsonNode, len(hint.Hints))
			for i, h := range hint.Hints {
				hints[i] = scalarExpressionToJSON(h)
			}
			node["Hints"] = hints
		}
		if hint.HintKind != "" {
			node["HintKind"] = hint.HintKind
		}
		return addSpan(node, frag(h))
	default:
		return addSpan(jsonNode{"$type": "UnknownOptimizerHint"}, frag(h))
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
	return addSpan(node, frag(p))
}

func queryExpressionToJSON(qe ast.QueryExpression) jsonNode {
	switch q := qe.(type) {
	case *ast.QuerySpecification:
		return addSpan(querySpecificationToJSON(q), frag(qe))
	case *ast.QueryParenthesisExpression:
		return addSpan(queryParenthesisExpressionToJSON(q), frag(qe))
	case *ast.BinaryQueryExpression:
		return addSpan(binaryQueryExpressionToJSON(q), frag(qe))
	default:
		return addSpan(jsonNode{"$type": "UnknownQueryExpression"}, frag(qe))
	}
}

func queryParenthesisExpressionToJSON(q *ast.QueryParenthesisExpression) jsonNode {
	node := jsonNode{
		"$type": "QueryParenthesisExpression",
	}
	if q.QueryExpression != nil {
		node["QueryExpression"] = queryExpressionToJSON(q.QueryExpression)
	}
	return addSpan(node, frag(q))
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
	return addSpan(node, frag(q))
}

func querySpecificationToJSON(q *ast.QuerySpecification) jsonNode {
	node := jsonNode{
		"$type": "QuerySpecification",
	}
	if q.UniqueRowFilter != "" {
		node["UniqueRowFilter"] = q.UniqueRowFilter
	} else {
		node["UniqueRowFilter"] = "NotSpecified"
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
	if q.WindowClause != nil {
		node["WindowClause"] = windowClauseToJSON(q.WindowClause)
	}
	if q.OrderByClause != nil {
		node["OrderByClause"] = orderByClauseToJSON(q.OrderByClause)
	}
	if q.OffsetClause != nil {
		node["OffsetClause"] = offsetClauseToJSON(q.OffsetClause)
	}
	if q.ForClause != nil {
		node["ForClause"] = forClauseToJSON(q.ForClause)
	}
	return addSpan(node, frag(q))
}

func offsetClauseToJSON(oc *ast.OffsetClause) jsonNode {
	node := jsonNode{
		"$type": "OffsetClause",
	}
	if oc.OffsetExpression != nil {
		node["OffsetExpression"] = scalarExpressionToJSON(oc.OffsetExpression)
	}
	if oc.FetchExpression != nil {
		node["FetchExpression"] = scalarExpressionToJSON(oc.FetchExpression)
	}
	return addSpan(node, frag(oc))
}

func forClauseToJSON(fc ast.ForClause) jsonNode {
	switch f := fc.(type) {
	case *ast.BrowseForClause:
		return addSpan(jsonNode{"$type": "BrowseForClause"}, frag(fc))
	case *ast.ReadOnlyForClause:
		return addSpan(jsonNode{"$type": "ReadOnlyForClause"}, frag(fc))
	case *ast.UpdateForClause:
		node := jsonNode{"$type": "UpdateForClause"}
		if len(f.Columns) > 0 {
			cols := make([]jsonNode, len(f.Columns))
			for i, col := range f.Columns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["Columns"] = cols
		}
		return addSpan(node, frag(fc))
	case *ast.XmlForClause:
		node := jsonNode{"$type": "XmlForClause"}
		if len(f.Options) > 0 {
			opts := make([]jsonNode, len(f.Options))
			for i, opt := range f.Options {
				opts[i] = xmlForClauseOptionToJSON(opt)
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(fc))
	case *ast.JsonForClause:
		node := jsonNode{"$type": "JsonForClause"}
		if len(f.Options) > 0 {
			opts := make([]jsonNode, len(f.Options))
			for i, opt := range f.Options {
				opts[i] = jsonForClauseOptionToJSON(opt)
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(fc))
	default:
		return addSpan(jsonNode{"$type": "UnknownForClause"}, frag(fc))
	}
}

func xmlForClauseOptionToJSON(opt *ast.XmlForClauseOption) jsonNode {
	node := jsonNode{"$type": "XmlForClauseOption"}
	if opt.OptionKind != "" {
		node["OptionKind"] = opt.OptionKind
	}
	if opt.Value != nil {
		node["Value"] = stringLiteralToJSON(opt.Value)
	}
	return addSpan(node, frag(opt))
}

func jsonForClauseOptionToJSON(opt *ast.JsonForClauseOption) jsonNode {
	node := jsonNode{"$type": "JsonForClauseOption"}
	if opt.OptionKind != "" {
		node["OptionKind"] = opt.OptionKind
	}
	if opt.Value != nil {
		node["Value"] = stringLiteralToJSON(opt.Value)
	}
	return addSpan(node, frag(opt))
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
	return addSpan(node, frag(t))
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
		return addSpan(node, frag(elem))
	case *ast.SelectStarExpression:
		node := jsonNode{
			"$type": "SelectStarExpression",
		}
		if e.Qualifier != nil {
			node["Qualifier"] = multiPartIdentifierToJSON(e.Qualifier)
		}
		return addSpan(node, frag(elem))
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
			node["Variable"] = addSpan(varNode, frag(e.Variable))
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		if e.AssignmentKind != "" {
			node["AssignmentKind"] = e.AssignmentKind
		}
		return addSpan(node, frag(elem))
	default:
		return addSpan(jsonNode{"$type": "UnknownSelectElement"}, frag(elem))
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
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.RealLiteral:
		node := jsonNode{
			"$type": "RealLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return addSpan(node, frag(expr))
	case *ast.MoneyLiteral:
		node := jsonNode{
			"$type": "MoneyLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.IdentifierLiteral:
		node := jsonNode{
			"$type": "IdentifierLiteral",
		}
		if e.LiteralType != "" {
			node["LiteralType"] = e.LiteralType
		}
		if e.QuoteType != "" {
			node["QuoteType"] = e.QuoteType
		}
		if e.Value != "" {
			node["Value"] = e.Value
		}
		return addSpan(node, frag(expr))
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
		if e.WithinGroupClause != nil {
			node["WithinGroupClause"] = withinGroupClauseToJSON(e.WithinGroupClause)
		}
		if e.OverClause != nil {
			node["OverClause"] = overClauseToJSON(e.OverClause)
		}
		if len(e.IgnoreRespectNulls) > 0 {
			idents := make([]jsonNode, len(e.IgnoreRespectNulls))
			for i, ident := range e.IgnoreRespectNulls {
				idents[i] = identifierToJSON(ident)
			}
			node["IgnoreRespectNulls"] = idents
		}
		node["WithArrayWrapper"] = e.WithArrayWrapper
		if e.TrimOptions != nil {
			node["TrimOptions"] = identifierToJSON(e.TrimOptions)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		if len(e.JsonParameters) > 0 {
			params := make([]jsonNode, len(e.JsonParameters))
			for i, kv := range e.JsonParameters {
				params[i] = jsonNode{
					"$type":       "JsonKeyValue",
					"JsonKeyName": scalarExpressionToJSON(kv.JsonKeyName),
					"JsonValue":   scalarExpressionToJSON(kv.JsonValue),
				}
			}
			node["JsonParameters"] = params
		}
		if len(e.AbsentOrNullOnNull) > 0 {
			idents := make([]jsonNode, len(e.AbsentOrNullOnNull))
			for i, ident := range e.AbsentOrNullOnNull {
				idents[i] = identifierToJSON(ident)
			}
			node["AbsentOrNullOnNull"] = idents
		}
		return addSpan(node, frag(expr))
	case *ast.PartitionFunctionCall:
		node := jsonNode{
			"$type": "PartitionFunctionCall",
		}
		if e.DatabaseName != nil {
			node["DatabaseName"] = identifierToJSON(e.DatabaseName)
		}
		if e.SchemaName != nil {
			node["SchemaName"] = identifierToJSON(e.SchemaName)
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.NullIfExpression:
		node := jsonNode{
			"$type": "NullIfExpression",
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = scalarExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = scalarExpressionToJSON(e.SecondExpression)
		}
		return addSpan(node, frag(expr))
	case *ast.CoalesceExpression:
		node := jsonNode{
			"$type": "CoalesceExpression",
		}
		if len(e.Expressions) > 0 {
			exprs := make([]jsonNode, len(e.Expressions))
			for i, expr := range e.Expressions {
				exprs[i] = scalarExpressionToJSON(expr)
			}
			node["Expressions"] = exprs
		}
		return addSpan(node, frag(expr))
	case *ast.ParameterlessCall:
		node := jsonNode{
			"$type": "ParameterlessCall",
		}
		if e.ParameterlessCallType != "" {
			node["ParameterlessCallType"] = e.ParameterlessCallType
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return addSpan(node, frag(expr))
	case *ast.IdentityFunctionCall:
		node := jsonNode{
			"$type": "IdentityFunctionCall",
		}
		if e.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(e.DataType)
		}
		if e.Seed != nil {
			node["Seed"] = scalarExpressionToJSON(e.Seed)
		}
		if e.Increment != nil {
			node["Increment"] = scalarExpressionToJSON(e.Increment)
		}
		return addSpan(node, frag(expr))
	case *ast.LeftFunctionCall:
		node := jsonNode{
			"$type": "LeftFunctionCall",
		}
		if len(e.Parameters) > 0 {
			params := make([]jsonNode, len(e.Parameters))
			for i, p := range e.Parameters {
				params[i] = scalarExpressionToJSON(p)
			}
			node["Parameters"] = params
		}
		return addSpan(node, frag(expr))
	case *ast.RightFunctionCall:
		node := jsonNode{
			"$type": "RightFunctionCall",
		}
		if len(e.Parameters) > 0 {
			params := make([]jsonNode, len(e.Parameters))
			for i, p := range e.Parameters {
				params[i] = scalarExpressionToJSON(p)
			}
			node["Parameters"] = params
		}
		return addSpan(node, frag(expr))
	case *ast.AtTimeZoneCall:
		node := jsonNode{
			"$type": "AtTimeZoneCall",
		}
		if e.DateValue != nil {
			node["DateValue"] = scalarExpressionToJSON(e.DateValue)
		}
		if e.TimeZone != nil {
			node["TimeZone"] = scalarExpressionToJSON(e.TimeZone)
		}
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.NextValueForExpression:
		node := jsonNode{
			"$type": "NextValueForExpression",
		}
		if e.SequenceName != nil {
			node["SequenceName"] = schemaObjectNameToJSON(e.SequenceName)
		}
		if e.OverClause != nil {
			node["OverClause"] = overClauseToJSON(e.OverClause)
		}
		return addSpan(node, frag(expr))
	case *ast.IIfCall:
		node := jsonNode{
			"$type": "IIfCall",
		}
		if e.Predicate != nil {
			node["Predicate"] = booleanExpressionToJSON(e.Predicate)
		}
		if e.ThenExpression != nil {
			node["ThenExpression"] = scalarExpressionToJSON(e.ThenExpression)
		}
		if e.ElseExpression != nil {
			node["ElseExpression"] = scalarExpressionToJSON(e.ElseExpression)
		}
		return addSpan(node, frag(expr))
	case *ast.ParseCall:
		node := jsonNode{
			"$type": "ParseCall",
		}
		if e.StringValue != nil {
			node["StringValue"] = scalarExpressionToJSON(e.StringValue)
		}
		if e.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(e.DataType)
		}
		if e.Culture != nil {
			node["Culture"] = scalarExpressionToJSON(e.Culture)
		}
		return addSpan(node, frag(expr))
	case *ast.TryParseCall:
		node := jsonNode{
			"$type": "TryParseCall",
		}
		if e.StringValue != nil {
			node["StringValue"] = scalarExpressionToJSON(e.StringValue)
		}
		if e.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(e.DataType)
		}
		if e.Culture != nil {
			node["Culture"] = scalarExpressionToJSON(e.Culture)
		}
		return addSpan(node, frag(expr))
	case *ast.VariableReference:
		node := jsonNode{
			"$type": "VariableReference",
		}
		if e.Name != "" {
			node["Name"] = e.Name
		}
		return addSpan(node, frag(expr))
	case *ast.GlobalVariableExpression:
		node := jsonNode{
			"$type": "GlobalVariableExpression",
		}
		if e.Name != "" {
			node["Name"] = e.Name
		}
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.OdbcFunctionCall:
		node := jsonNode{
			"$type": "OdbcFunctionCall",
		}
		if e.Name != nil {
			node["Name"] = identifierToJSON(e.Name)
		}
		node["ParametersUsed"] = e.ParametersUsed
		if len(e.Parameters) > 0 {
			params := make([]jsonNode, len(e.Parameters))
			for i, param := range e.Parameters {
				params[i] = scalarExpressionToJSON(param)
			}
			node["Parameters"] = params
		}
		return addSpan(node, frag(expr))
	case *ast.OdbcConvertSpecification:
		node := jsonNode{
			"$type": "OdbcConvertSpecification",
		}
		if e.Identifier != nil {
			node["Identifier"] = identifierToJSON(e.Identifier)
		}
		return addSpan(node, frag(expr))
	case *ast.ExtractFromExpression:
		node := jsonNode{
			"$type": "ExtractFromExpression",
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		if e.ExtractedElement != nil {
			node["ExtractedElement"] = identifierToJSON(e.ExtractedElement)
		}
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.ParenthesisExpression:
		node := jsonNode{
			"$type": "ParenthesisExpression",
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		return addSpan(node, frag(expr))
	case *ast.ScalarSubquery:
		node := jsonNode{
			"$type": "ScalarSubquery",
		}
		if e.QueryExpression != nil {
			node["QueryExpression"] = queryExpressionToJSON(e.QueryExpression)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return addSpan(node, frag(expr))
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
				clauses[i] = addSpan(clause, frag(c))
			}
			node["WhenClauses"] = clauses
		}
		if e.ElseExpression != nil {
			node["ElseExpression"] = scalarExpressionToJSON(e.ElseExpression)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return addSpan(node, frag(expr))
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
				clauses[i] = addSpan(clause, frag(c))
			}
			node["WhenClauses"] = clauses
		}
		if e.ElseExpression != nil {
			node["ElseExpression"] = scalarExpressionToJSON(e.ElseExpression)
		}
		if e.Collation != nil {
			node["Collation"] = identifierToJSON(e.Collation)
		}
		return addSpan(node, frag(expr))
	case *ast.SourceDeclaration:
		node := jsonNode{
			"$type": "SourceDeclaration",
		}
		if e.Value != nil {
			node["Value"] = eventSessionObjectNameToJSON(e.Value)
		}
		return addSpan(node, frag(expr))
	default:
		return addSpan(jsonNode{"$type": "UnknownScalarExpression"}, frag(expr))
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
	return addSpan(node, frag(id))
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
	return addSpan(node, frag(mpi))
}

func eventSessionObjectNameToJSON(e *ast.EventSessionObjectName) jsonNode {
	node := jsonNode{
		"$type": "EventSessionObjectName",
	}
	if e.MultiPartIdentifier != nil {
		node["MultiPartIdentifier"] = multiPartIdentifierToJSON(e.MultiPartIdentifier)
	}
	return addSpan(node, frag(e))
}

func sourceDeclarationToJSON(s *ast.SourceDeclaration) jsonNode {
	node := jsonNode{
		"$type": "SourceDeclaration",
	}
	if s.Value != nil {
		node["Value"] = eventSessionObjectNameToJSON(s.Value)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(iove))
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
	return addSpan(node, frag(fc))
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
		if r.TableSampleClause != nil {
			node["TableSampleClause"] = tableSampleClauseToJSON(r.TableSampleClause)
		}
		if r.TemporalClause != nil {
			node["TemporalClause"] = temporalClauseToJSON(r.TemporalClause)
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
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
		} else {
			node["JoinHint"] = "None"
		}
		if r.FirstTableReference != nil {
			node["FirstTableReference"] = tableReferenceToJSON(r.FirstTableReference)
		}
		if r.SecondTableReference != nil {
			node["SecondTableReference"] = tableReferenceToJSON(r.SecondTableReference)
		}
		return addSpan(node, frag(ref))
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
		return addSpan(node, frag(ref))
	case *ast.JoinParenthesisTableReference:
		node := jsonNode{
			"$type": "JoinParenthesisTableReference",
		}
		if r.Join != nil {
			node["Join"] = tableReferenceToJSON(r.Join)
		}
		return addSpan(node, frag(ref))
	case *ast.OdbcQualifiedJoinTableReference:
		node := jsonNode{
			"$type": "OdbcQualifiedJoinTableReference",
		}
		if r.TableReference != nil {
			node["TableReference"] = tableReferenceToJSON(r.TableReference)
		}
		return addSpan(node, frag(ref))
	case *ast.VariableTableReference:
		node := jsonNode{
			"$type": "VariableTableReference",
		}
		if r.Variable != nil {
			node["Variable"] = scalarExpressionToJSON(r.Variable)
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.VariableMethodCallTableReference:
		node := jsonNode{
			"$type": "VariableMethodCallTableReference",
		}
		if r.Variable != nil {
			node["Variable"] = scalarExpressionToJSON(r.Variable)
		}
		if r.MethodName != nil {
			node["MethodName"] = identifierToJSON(r.MethodName)
		}
		if len(r.Parameters) > 0 {
			params := make([]jsonNode, len(r.Parameters))
			for i, p := range r.Parameters {
				params[i] = scalarExpressionToJSON(p)
			}
			node["Parameters"] = params
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
		return addSpan(node, frag(ref))
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
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		if len(r.Columns) > 0 {
			cols := make([]jsonNode, len(r.Columns))
			for i, c := range r.Columns {
				cols[i] = identifierToJSON(c)
			}
			node["Columns"] = cols
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.GlobalFunctionTableReference:
		node := jsonNode{
			"$type": "GlobalFunctionTableReference",
		}
		if r.Name != nil {
			node["Name"] = identifierToJSON(r.Name)
		}
		if len(r.Parameters) > 0 {
			params := make([]jsonNode, len(r.Parameters))
			for i, p := range r.Parameters {
				params[i] = scalarExpressionToJSON(p)
			}
			node["Parameters"] = params
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		if len(r.Columns) > 0 {
			cols := make([]jsonNode, len(r.Columns))
			for i, c := range r.Columns {
				cols[i] = identifierToJSON(c)
			}
			node["Columns"] = cols
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.OpenJsonTableReference:
		node := jsonNode{
			"$type": "OpenJsonTableReference",
		}
		if r.Variable != nil {
			node["Variable"] = scalarExpressionToJSON(r.Variable)
		}
		if r.RowPattern != nil {
			node["RowPattern"] = scalarExpressionToJSON(r.RowPattern)
		}
		if len(r.SchemaDeclarationItems) > 0 {
			items := make([]jsonNode, len(r.SchemaDeclarationItems))
			for i, item := range r.SchemaDeclarationItems {
				itemNode := jsonNode{
					"$type": "SchemaDeclarationItemOpenjson",
				}
				itemNode["AsJson"] = item.AsJson
				if item.ColumnDefinition != nil {
					colDef := jsonNode{
						"$type": "ColumnDefinitionBase",
					}
					if item.ColumnDefinition.ColumnIdentifier != nil {
						colDef["ColumnIdentifier"] = identifierToJSON(item.ColumnDefinition.ColumnIdentifier)
					}
					if item.ColumnDefinition.DataType != nil {
						colDef["DataType"] = dataTypeReferenceToJSON(item.ColumnDefinition.DataType)
					}
					if item.ColumnDefinition.Collation != nil {
						colDef["Collation"] = identifierToJSON(item.ColumnDefinition.Collation)
					}
					itemNode["ColumnDefinition"] = colDef
				}
				if item.Mapping != nil {
					itemNode["Mapping"] = scalarExpressionToJSON(item.Mapping)
				}
				items[i] = itemNode
			}
			node["SchemaDeclarationItems"] = items
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.BuiltInFunctionTableReference:
		node := jsonNode{
			"$type": "BuiltInFunctionTableReference",
		}
		if r.Name != nil {
			node["Name"] = identifierToJSON(r.Name)
		}
		if len(r.Parameters) > 0 {
			params := make([]jsonNode, len(r.Parameters))
			for i, p := range r.Parameters {
				params[i] = scalarExpressionToJSON(p)
			}
			node["Parameters"] = params
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		if len(r.Columns) > 0 {
			cols := make([]jsonNode, len(r.Columns))
			for i, c := range r.Columns {
				cols[i] = identifierToJSON(c)
			}
			node["Columns"] = cols
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.InlineDerivedTable:
		node := jsonNode{
			"$type": "InlineDerivedTable",
		}
		if len(r.RowValues) > 0 {
			rows := make([]jsonNode, len(r.RowValues))
			for i, row := range r.RowValues {
				rows[i] = rowValueToJSON(row)
			}
			node["RowValues"] = rows
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
		return addSpan(node, frag(ref))
	case *ast.DataModificationTableReference:
		node := jsonNode{
			"$type": "DataModificationTableReference",
		}
		if r.DataModificationSpecification != nil {
			node["DataModificationSpecification"] = dataModificationSpecificationToJSON(r.DataModificationSpecification)
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
		return addSpan(node, frag(ref))
	case *ast.ChangeTableChangesTableReference:
		node := jsonNode{
			"$type": "ChangeTableChangesTableReference",
		}
		if r.Target != nil {
			node["Target"] = schemaObjectNameToJSON(r.Target)
		}
		if r.SinceVersion != nil {
			node["SinceVersion"] = scalarExpressionToJSON(r.SinceVersion)
		}
		node["ForceSeek"] = r.ForceSeek
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
		return addSpan(node, frag(ref))
	case *ast.ChangeTableVersionTableReference:
		node := jsonNode{
			"$type": "ChangeTableVersionTableReference",
		}
		if r.Target != nil {
			node["Target"] = schemaObjectNameToJSON(r.Target)
		}
		if len(r.PrimaryKeyColumns) > 0 {
			cols := make([]jsonNode, len(r.PrimaryKeyColumns))
			for i, c := range r.PrimaryKeyColumns {
				cols[i] = identifierToJSON(c)
			}
			node["PrimaryKeyColumns"] = cols
		}
		if len(r.PrimaryKeyValues) > 0 {
			vals := make([]jsonNode, len(r.PrimaryKeyValues))
			for i, v := range r.PrimaryKeyValues {
				vals[i] = scalarExpressionToJSON(v)
			}
			node["PrimaryKeyValues"] = vals
		}
		node["ForceSeek"] = r.ForceSeek
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
		return addSpan(node, frag(ref))
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
		return addSpan(node, frag(ref))
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
		if len(r.WithColumns) > 0 {
			cols := make([]jsonNode, len(r.WithColumns))
			for i, c := range r.WithColumns {
				cols[i] = openRowsetColumnDefinitionToJSON(c)
			}
			node["WithColumns"] = cols
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
		return addSpan(node, frag(ref))
	case *ast.OpenRowsetCosmos:
		node := jsonNode{
			"$type": "OpenRowsetCosmos",
		}
		if len(r.Options) > 0 {
			opts := make([]jsonNode, len(r.Options))
			for i, o := range r.Options {
				opts[i] = openRowsetCosmosOptionToJSON(o)
			}
			node["Options"] = opts
		}
		if len(r.WithColumns) > 0 {
			cols := make([]jsonNode, len(r.WithColumns))
			for i, c := range r.WithColumns {
				cols[i] = openRowsetColumnDefinitionToJSON(c)
			}
			node["WithColumns"] = cols
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.OpenRowsetTableReference:
		node := jsonNode{
			"$type": "OpenRowsetTableReference",
		}
		if r.ProviderName != nil {
			node["ProviderName"] = scalarExpressionToJSON(r.ProviderName)
		}
		if r.ProviderString != nil {
			node["ProviderString"] = scalarExpressionToJSON(r.ProviderString)
		}
		if r.DataSource != nil {
			node["DataSource"] = scalarExpressionToJSON(r.DataSource)
		}
		if r.UserId != nil {
			node["UserId"] = scalarExpressionToJSON(r.UserId)
		}
		if r.Password != nil {
			node["Password"] = scalarExpressionToJSON(r.Password)
		}
		if r.Query != nil {
			node["Query"] = scalarExpressionToJSON(r.Query)
		}
		if r.Object != nil {
			node["Object"] = schemaObjectNameToJSON(r.Object)
		}
		if len(r.WithColumns) > 0 {
			cols := make([]jsonNode, len(r.WithColumns))
			for i, c := range r.WithColumns {
				cols[i] = openRowsetColumnDefinitionToJSON(c)
			}
			node["WithColumns"] = cols
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.AdHocTableReference:
		node := jsonNode{
			"$type": "AdHocTableReference",
		}
		if r.DataSource != nil {
			node["DataSource"] = adHocDataSourceToJSON(r.DataSource)
		}
		if r.Object != nil {
			objNode := jsonNode{
				"$type": "SchemaObjectNameOrValueExpression",
			}
			if r.Object.SchemaObjectName != nil {
				objNode["SchemaObjectName"] = schemaObjectNameToJSON(r.Object.SchemaObjectName)
			}
			if r.Object.ValueExpression != nil {
				objNode["ValueExpression"] = scalarExpressionToJSON(r.Object.ValueExpression)
			}
			node["Object"] = objNode
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.OpenXmlTableReference:
		node := jsonNode{
			"$type": "OpenXmlTableReference",
		}
		if r.Variable != nil {
			node["Variable"] = scalarExpressionToJSON(r.Variable)
		}
		if r.RowPattern != nil {
			node["RowPattern"] = scalarExpressionToJSON(r.RowPattern)
		}
		if r.Flags != nil {
			node["Flags"] = scalarExpressionToJSON(r.Flags)
		}
		if len(r.SchemaDeclarationItems) > 0 {
			items := make([]jsonNode, len(r.SchemaDeclarationItems))
			for i, item := range r.SchemaDeclarationItems {
				items[i] = schemaDeclarationItemToJSON(item)
			}
			node["SchemaDeclarationItems"] = items
		}
		if r.TableName != nil {
			node["TableName"] = schemaObjectNameToJSON(r.TableName)
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.OpenQueryTableReference:
		node := jsonNode{
			"$type": "OpenQueryTableReference",
		}
		if r.LinkedServer != nil {
			node["LinkedServer"] = identifierToJSON(r.LinkedServer)
		}
		if r.Query != nil {
			node["Query"] = scalarExpressionToJSON(r.Query)
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.PredictTableReference:
		node := jsonNode{
			"$type": "PredictTableReference",
		}
		if r.ModelVariable != nil {
			node["ModelVariable"] = scalarExpressionToJSON(r.ModelVariable)
		}
		if r.DataSource != nil {
			node["DataSource"] = tableReferenceToJSON(r.DataSource)
		}
		if r.RunTime != nil {
			node["RunTime"] = identifierToJSON(r.RunTime)
		}
		if len(r.SchemaDeclarationItems) > 0 {
			items := make([]jsonNode, len(r.SchemaDeclarationItems))
			for i, item := range r.SchemaDeclarationItems {
				items[i] = schemaDeclarationItemToJSON(item)
			}
			node["SchemaDeclarationItems"] = items
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.PivotedTableReference:
		node := jsonNode{
			"$type": "PivotedTableReference",
		}
		if r.TableReference != nil {
			node["TableReference"] = tableReferenceToJSON(r.TableReference)
		}
		if len(r.InColumns) > 0 {
			cols := make([]jsonNode, len(r.InColumns))
			for i, col := range r.InColumns {
				cols[i] = identifierToJSON(col)
			}
			node["InColumns"] = cols
		}
		if r.PivotColumn != nil {
			node["PivotColumn"] = columnReferenceExpressionToJSON(r.PivotColumn)
		}
		if len(r.ValueColumns) > 0 {
			cols := make([]jsonNode, len(r.ValueColumns))
			for i, col := range r.ValueColumns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["ValueColumns"] = cols
		}
		if r.AggregateFunctionIdentifier != nil {
			node["AggregateFunctionIdentifier"] = multiPartIdentifierToJSON(r.AggregateFunctionIdentifier)
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.UnpivotedTableReference:
		node := jsonNode{
			"$type": "UnpivotedTableReference",
		}
		if r.TableReference != nil {
			node["TableReference"] = tableReferenceToJSON(r.TableReference)
		}
		if len(r.InColumns) > 0 {
			cols := make([]jsonNode, len(r.InColumns))
			for i, col := range r.InColumns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["InColumns"] = cols
		}
		if r.PivotColumn != nil {
			node["PivotColumn"] = identifierToJSON(r.PivotColumn)
		}
		if r.ValueColumn != nil {
			node["ValueColumn"] = identifierToJSON(r.ValueColumn)
		}
		if r.NullHandling != "" && r.NullHandling != "None" {
			node["NullHandling"] = r.NullHandling
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.QueryDerivedTable:
		node := jsonNode{
			"$type": "QueryDerivedTable",
		}
		if r.QueryExpression != nil {
			node["QueryExpression"] = queryExpressionToJSON(r.QueryExpression)
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
		return addSpan(node, frag(ref))
	case *ast.FullTextTableReference:
		node := jsonNode{
			"$type": "FullTextTableReference",
		}
		if r.FullTextFunctionType != "" {
			node["FullTextFunctionType"] = r.FullTextFunctionType
		}
		if r.TableName != nil {
			node["TableName"] = schemaObjectNameToJSON(r.TableName)
		}
		if len(r.Columns) > 0 {
			cols := make([]jsonNode, len(r.Columns))
			for i, col := range r.Columns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["Columns"] = cols
		}
		if r.SearchCondition != nil {
			node["SearchCondition"] = scalarExpressionToJSON(r.SearchCondition)
		}
		if r.TopN != nil {
			node["TopN"] = scalarExpressionToJSON(r.TopN)
		}
		if r.Language != nil {
			node["Language"] = scalarExpressionToJSON(r.Language)
		}
		if r.PropertyName != nil {
			node["PropertyName"] = scalarExpressionToJSON(r.PropertyName)
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	case *ast.SemanticTableReference:
		node := jsonNode{
			"$type": "SemanticTableReference",
		}
		if r.SemanticFunctionType != "" {
			node["SemanticFunctionType"] = r.SemanticFunctionType
		}
		if r.TableName != nil {
			node["TableName"] = schemaObjectNameToJSON(r.TableName)
		}
		if len(r.Columns) > 0 {
			cols := make([]jsonNode, len(r.Columns))
			for i, col := range r.Columns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["Columns"] = cols
		}
		if r.SourceKey != nil {
			node["SourceKey"] = scalarExpressionToJSON(r.SourceKey)
		}
		if r.MatchedColumn != nil {
			node["MatchedColumn"] = columnReferenceExpressionToJSON(r.MatchedColumn)
		}
		if r.MatchedKey != nil {
			node["MatchedKey"] = scalarExpressionToJSON(r.MatchedKey)
		}
		if r.Alias != nil {
			node["Alias"] = identifierToJSON(r.Alias)
		}
		node["ForPath"] = r.ForPath
		return addSpan(node, frag(ref))
	default:
		return addSpan(jsonNode{"$type": "UnknownTableReference"}, frag(ref))
	}
}

func schemaDeclarationItemToJSON(item *ast.SchemaDeclarationItem) jsonNode {
	node := jsonNode{
		"$type": "SchemaDeclarationItem",
	}
	if item.ColumnDefinition != nil {
		node["ColumnDefinition"] = columnDefinitionBaseToJSON(item.ColumnDefinition)
	}
	if item.Mapping != nil {
		node["Mapping"] = scalarExpressionToJSON(item.Mapping)
	}
	return addSpan(node, frag(item))
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
	return addSpan(node, frag(son))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.BooleanParenthesisExpression:
		node := jsonNode{
			"$type": "BooleanParenthesisExpression",
		}
		if e.Expression != nil {
			node["Expression"] = booleanExpressionToJSON(e.Expression)
		}
		return addSpan(node, frag(expr))
	case *ast.BooleanNotExpression:
		node := jsonNode{
			"$type": "BooleanNotExpression",
		}
		if e.Expression != nil {
			node["Expression"] = booleanExpressionToJSON(e.Expression)
		}
		return addSpan(node, frag(expr))
	case *ast.UpdateCall:
		node := jsonNode{
			"$type": "UpdateCall",
		}
		if e.Identifier != nil {
			node["Identifier"] = identifierToJSON(e.Identifier)
		}
		return addSpan(node, frag(expr))
	case *ast.TSEqualCall:
		node := jsonNode{
			"$type": "TSEqualCall",
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = scalarExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = scalarExpressionToJSON(e.SecondExpression)
		}
		return addSpan(node, frag(expr))
	case *ast.BooleanIsNullExpression:
		node := jsonNode{
			"$type": "BooleanIsNullExpression",
		}
		node["IsNot"] = e.IsNot
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		return addSpan(node, frag(expr))
	case *ast.DistinctPredicate:
		node := jsonNode{
			"$type": "DistinctPredicate",
		}
		if e.FirstExpression != nil {
			node["FirstExpression"] = scalarExpressionToJSON(e.FirstExpression)
		}
		if e.SecondExpression != nil {
			node["SecondExpression"] = scalarExpressionToJSON(e.SecondExpression)
		}
		node["IsNot"] = e.IsNot
		return addSpan(node, frag(expr))
	case *ast.SubqueryComparisonPredicate:
		node := jsonNode{
			"$type":                           "SubqueryComparisonPredicate",
			"ComparisonType":                  e.ComparisonType,
			"SubqueryComparisonPredicateType": e.SubqueryComparisonPredicateType,
		}
		if e.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(e.Expression)
		}
		if e.Subquery != nil {
			node["Subquery"] = scalarExpressionToJSON(e.Subquery)
		}
		return addSpan(node, frag(expr))
	case *ast.BooleanInExpression:
		node := jsonNode{
			"$type": "InPredicate",
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
			node["Subquery"] = jsonNode{
				"$type":           "ScalarSubquery",
				"QueryExpression": queryExpressionToJSON(e.Subquery),
			}
		}
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
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
		return addSpan(node, frag(expr))
	case *ast.EventDeclarationCompareFunctionParameter:
		node := jsonNode{
			"$type": "EventDeclarationCompareFunctionParameter",
		}
		if e.Name != nil {
			node["Name"] = eventSessionObjectNameToJSON(e.Name)
		}
		if e.SourceDeclaration != nil {
			node["SourceDeclaration"] = sourceDeclarationToJSON(e.SourceDeclaration)
		}
		if e.EventValue != nil {
			node["EventValue"] = scalarExpressionToJSON(e.EventValue)
		}
		return addSpan(node, frag(expr))
	case *ast.SourceDeclaration:
		return addSpan(sourceDeclarationToJSON(e), frag(expr))
	case *ast.GraphMatchPredicate:
		node := jsonNode{
			"$type": "GraphMatchPredicate",
		}
		if e.Expression != nil {
			node["Expression"] = graphMatchExpressionToJSON(e.Expression)
		}
		return addSpan(node, frag(expr))
	case *ast.FullTextPredicate:
		node := jsonNode{
			"$type": "FullTextPredicate",
		}
		if e.FullTextFunctionType != "" {
			node["FullTextFunctionType"] = e.FullTextFunctionType
		}
		if len(e.Columns) > 0 {
			cols := make([]jsonNode, len(e.Columns))
			for i, col := range e.Columns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["Columns"] = cols
		}
		if e.Value != nil {
			node["Value"] = scalarExpressionToJSON(e.Value)
		}
		if e.PropertyName != nil {
			node["PropertyName"] = scalarExpressionToJSON(e.PropertyName)
		}
		if e.LanguageTerm != nil {
			node["LanguageTerm"] = scalarExpressionToJSON(e.LanguageTerm)
		}
		return addSpan(node, frag(expr))
	case *ast.ExistsPredicate:
		node := jsonNode{
			"$type": "ExistsPredicate",
		}
		if e.Subquery != nil {
			node["Subquery"] = jsonNode{
				"$type":           "ScalarSubquery",
				"QueryExpression": queryExpressionToJSON(e.Subquery),
			}
		}
		return addSpan(node, frag(expr))
	case *ast.GraphMatchCompositeExpression:
		// GraphMatchCompositeExpression can appear as a BooleanExpression in chained patterns
		node := jsonNode{
			"$type": "GraphMatchCompositeExpression",
		}
		if e.LeftNode != nil {
			node["LeftNode"] = graphMatchNodeExpressionToJSON(e.LeftNode)
		}
		if e.Edge != nil {
			node["Edge"] = identifierToJSON(e.Edge)
		}
		if e.RightNode != nil {
			node["RightNode"] = graphMatchNodeExpressionToJSON(e.RightNode)
		}
		node["ArrowOnRight"] = e.ArrowOnRight
		return addSpan(node, frag(expr))
	default:
		return addSpan(jsonNode{"$type": "UnknownBooleanExpression"}, frag(expr))
	}
}

// graphMatchContext tracks seen node pointers for $ref support
type graphMatchContext struct {
	seenNodes map[*ast.GraphMatchNodeExpression]bool
}

func newGraphMatchContext() *graphMatchContext {
	return &graphMatchContext{
		seenNodes: make(map[*ast.GraphMatchNodeExpression]bool),
	}
}

func graphMatchExpressionToJSON(expr ast.GraphMatchExpression) jsonNode {
	ctx := newGraphMatchContext()
	return addSpan(graphMatchExpressionToJSONWithContext(expr, ctx), frag(expr))
}

func graphMatchExpressionToJSONWithContext(expr ast.GraphMatchExpression, ctx *graphMatchContext) jsonNode {
	switch e := expr.(type) {
	case *ast.GraphMatchCompositeExpression:
		node := jsonNode{
			"$type": "GraphMatchCompositeExpression",
		}
		if e.LeftNode != nil {
			node["LeftNode"] = graphMatchNodeExpressionToJSONWithContext(e.LeftNode, ctx)
		}
		if e.Edge != nil {
			node["Edge"] = identifierToJSON(e.Edge)
		}
		if e.RightNode != nil {
			node["RightNode"] = graphMatchNodeExpressionToJSONWithContext(e.RightNode, ctx)
		}
		node["ArrowOnRight"] = e.ArrowOnRight
		return addSpan(node, frag(e))
	case *ast.GraphMatchNodeExpression:
		return graphMatchNodeExpressionToJSONWithContext(e, ctx)
	case *ast.BooleanBinaryExpression:
		// Chained patterns produce BooleanBinaryExpression with And
		return booleanBinaryExpressionToJSONWithGraphContext(e, ctx)
	case *ast.GraphMatchRecursivePredicate:
		node := jsonNode{
			"$type": "GraphMatchRecursivePredicate",
		}
		if e.Function != "" {
			node["Function"] = e.Function
		}
		if e.OuterNodeExpression != nil {
			node["OuterNodeExpression"] = graphMatchNodeExpressionToJSONWithContext(e.OuterNodeExpression, ctx)
		}
		if len(e.Expression) > 0 {
			exprs := make([]jsonNode, len(e.Expression))
			for i, expr := range e.Expression {
				exprs[i] = graphMatchExpressionToJSONWithContext(expr, ctx)
			}
			node["Expression"] = exprs
		}
		if e.RecursiveQuantifier != nil {
			node["RecursiveQuantifier"] = graphRecursiveMatchQuantifierToJSON(e.RecursiveQuantifier)
		}
		node["AnchorOnLeft"] = e.AnchorOnLeft
		return node
	case *ast.GraphMatchLastNodePredicate:
		node := jsonNode{
			"$type": "GraphMatchLastNodePredicate",
		}
		if e.LeftExpression != nil {
			node["LeftExpression"] = graphMatchNodeExpressionToJSONWithContext(e.LeftExpression, ctx)
		}
		if e.RightExpression != nil {
			node["RightExpression"] = graphMatchNodeExpressionToJSONWithContext(e.RightExpression, ctx)
		}
		return node
	default:
		return jsonNode{"$type": "UnknownGraphMatchExpression"}
	}
}

func booleanBinaryExpressionToJSONWithGraphContext(e *ast.BooleanBinaryExpression, ctx *graphMatchContext) jsonNode {
	node := jsonNode{
		"$type": "BooleanBinaryExpression",
	}
	if e.BinaryExpressionType != "" {
		node["BinaryExpressionType"] = e.BinaryExpressionType
	}
	if e.FirstExpression != nil {
		// Check if first expression is a graph match expression type
		switch firstExpr := e.FirstExpression.(type) {
		case *ast.GraphMatchCompositeExpression:
			node["FirstExpression"] = graphMatchExpressionToJSONWithContext(firstExpr, ctx)
		case *ast.BooleanBinaryExpression:
			// Could be nested chained patterns - check if it contains graph match expressions
			node["FirstExpression"] = booleanBinaryExpressionToJSONWithGraphContext(firstExpr, ctx)
		case *ast.GraphMatchRecursivePredicate:
			node["FirstExpression"] = graphMatchExpressionToJSONWithContext(firstExpr, ctx)
		case *ast.GraphMatchLastNodePredicate:
			node["FirstExpression"] = graphMatchExpressionToJSONWithContext(firstExpr, ctx)
		default:
			node["FirstExpression"] = booleanExpressionToJSON(e.FirstExpression)
		}
	}
	if e.SecondExpression != nil {
		// Check if second expression is a graph match expression type
		switch secondExpr := e.SecondExpression.(type) {
		case *ast.GraphMatchCompositeExpression:
			node["SecondExpression"] = graphMatchExpressionToJSONWithContext(secondExpr, ctx)
		case *ast.BooleanBinaryExpression:
			// Could be nested chained patterns - check if it contains graph match expressions
			node["SecondExpression"] = booleanBinaryExpressionToJSONWithGraphContext(secondExpr, ctx)
		case *ast.GraphMatchRecursivePredicate:
			node["SecondExpression"] = graphMatchExpressionToJSONWithContext(secondExpr, ctx)
		case *ast.GraphMatchLastNodePredicate:
			node["SecondExpression"] = graphMatchExpressionToJSONWithContext(secondExpr, ctx)
		default:
			node["SecondExpression"] = booleanExpressionToJSON(e.SecondExpression)
		}
	}
	return addSpan(node, frag(e))
}

func graphMatchNodeExpressionToJSON(expr *ast.GraphMatchNodeExpression) jsonNode {
	node := jsonNode{
		"$type": "GraphMatchNodeExpression",
	}
	if expr.Node != nil {
		node["Node"] = identifierToJSON(expr.Node)
	}
	node["UsesLastNode"] = expr.UsesLastNode
	return addSpan(node, frag(expr))
}

func graphMatchNodeExpressionToJSONWithContext(expr *ast.GraphMatchNodeExpression, ctx *graphMatchContext) jsonNode {
	// Check if we've seen this exact pointer before
	if ctx.seenNodes[expr] {
		// This node pointer has been seen before, use $ref
		return jsonNode{"$ref": "GraphMatchNodeExpression"}
	}
	ctx.seenNodes[expr] = true
	return graphMatchNodeExpressionToJSON(expr)
}

func graphRecursiveMatchQuantifierToJSON(q *ast.GraphRecursiveMatchQuantifier) jsonNode {
	node := jsonNode{
		"$type": "GraphRecursiveMatchQuantifier",
	}
	node["IsPlusSign"] = q.IsPlusSign
	if q.LowerLimit != nil {
		node["LowerLimit"] = scalarExpressionToJSON(q.LowerLimit)
	}
	if q.UpperLimit != nil {
		node["UpperLimit"] = scalarExpressionToJSON(q.UpperLimit)
	}
	return addSpan(node, frag(q))
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
	return addSpan(node, frag(gbc))
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
		return addSpan(node, frag(spec))
	case *ast.RollupGroupingSpecification:
		node := jsonNode{
			"$type": "RollupGroupingSpecification",
		}
		if len(s.Arguments) > 0 {
			args := make([]jsonNode, len(s.Arguments))
			for i, arg := range s.Arguments {
				args[i] = groupingSpecificationToJSON(arg)
			}
			node["Arguments"] = args
		}
		return addSpan(node, frag(spec))
	case *ast.CubeGroupingSpecification:
		node := jsonNode{
			"$type": "CubeGroupingSpecification",
		}
		if len(s.Arguments) > 0 {
			args := make([]jsonNode, len(s.Arguments))
			for i, arg := range s.Arguments {
				args[i] = groupingSpecificationToJSON(arg)
			}
			node["Arguments"] = args
		}
		return addSpan(node, frag(spec))
	case *ast.CompositeGroupingSpecification:
		node := jsonNode{
			"$type": "CompositeGroupingSpecification",
		}
		if len(s.Items) > 0 {
			items := make([]jsonNode, len(s.Items))
			for i, item := range s.Items {
				items[i] = groupingSpecificationToJSON(item)
			}
			node["Items"] = items
		}
		return addSpan(node, frag(spec))
	case *ast.GrandTotalGroupingSpecification:
		return addSpan(jsonNode{
			"$type": "GrandTotalGroupingSpecification",
		}, frag(spec))
	case *ast.GroupingSetsGroupingSpecification:
		node := jsonNode{
			"$type": "GroupingSetsGroupingSpecification",
		}
		if len(s.Arguments) > 0 {
			args := make([]jsonNode, len(s.Arguments))
			for i, arg := range s.Arguments {
				args[i] = groupingSpecificationToJSON(arg)
			}
			node["Sets"] = args
		}
		return addSpan(node, frag(spec))
	default:
		return addSpan(jsonNode{"$type": "UnknownGroupingSpecification"}, frag(spec))
	}
}

func havingClauseToJSON(hc *ast.HavingClause) jsonNode {
	node := jsonNode{
		"$type": "HavingClause",
	}
	if hc.SearchCondition != nil {
		node["SearchCondition"] = booleanExpressionToJSON(hc.SearchCondition)
	}
	return addSpan(node, frag(hc))
}

func windowClauseToJSON(wc *ast.WindowClause) jsonNode {
	node := jsonNode{
		"$type": "WindowClause",
	}
	if len(wc.WindowDefinition) > 0 {
		defs := make([]jsonNode, len(wc.WindowDefinition))
		for i, def := range wc.WindowDefinition {
			defs[i] = windowDefinitionToJSON(def)
		}
		node["WindowDefinition"] = defs
	}
	return addSpan(node, frag(wc))
}

func windowDefinitionToJSON(wd *ast.WindowDefinition) jsonNode {
	node := jsonNode{
		"$type": "WindowDefinition",
	}
	if wd.WindowName != nil {
		node["WindowName"] = identifierToJSON(wd.WindowName)
	}
	if wd.RefWindowName != nil {
		node["RefWindowName"] = identifierToJSON(wd.RefWindowName)
	}
	if len(wd.Partitions) > 0 {
		partitions := make([]jsonNode, len(wd.Partitions))
		for i, p := range wd.Partitions {
			partitions[i] = scalarExpressionToJSON(p)
		}
		node["Partitions"] = partitions
	}
	if wd.OrderByClause != nil {
		node["OrderByClause"] = orderByClauseToJSON(wd.OrderByClause)
	}
	return addSpan(node, frag(wd))
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
	return addSpan(node, frag(obc))
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
	return addSpan(node, frag(ewso))
}

func withinGroupClauseToJSON(wg *ast.WithinGroupClause) jsonNode {
	node := jsonNode{
		"$type": "WithinGroupClause",
	}
	if wg.OrderByClause != nil {
		node["OrderByClause"] = orderByClauseToJSON(wg.OrderByClause)
	}
	node["HasGraphPath"] = wg.HasGraphPath
	return addSpan(node, frag(wg))
}

func overClauseToJSON(oc *ast.OverClause) jsonNode {
	node := jsonNode{
		"$type": "OverClause",
	}
	if oc.WindowName != nil {
		node["WindowName"] = identifierToJSON(oc.WindowName)
	}
	if len(oc.Partitions) > 0 {
		partitions := make([]jsonNode, len(oc.Partitions))
		for i, p := range oc.Partitions {
			partitions[i] = scalarExpressionToJSON(p)
		}
		node["Partitions"] = partitions
	}
	if oc.OrderByClause != nil {
		node["OrderByClause"] = orderByClauseToJSON(oc.OrderByClause)
	}
	if oc.WindowFrameClause != nil {
		node["WindowFrameClause"] = windowFrameClauseToJSON(oc.WindowFrameClause)
	}
	return addSpan(node, frag(oc))
}

func windowFrameClauseToJSON(wfc *ast.WindowFrameClause) jsonNode {
	node := jsonNode{
		"$type":           "WindowFrameClause",
		"WindowFrameType": wfc.WindowFrameType,
	}
	if wfc.Top != nil {
		node["Top"] = windowDelimiterToJSON(wfc.Top)
	}
	if wfc.Bottom != nil {
		node["Bottom"] = windowDelimiterToJSON(wfc.Bottom)
	}
	return addSpan(node, frag(wfc))
}

func windowDelimiterToJSON(wd *ast.WindowDelimiter) jsonNode {
	node := jsonNode{
		"$type":               "WindowDelimiter",
		"WindowDelimiterType": wd.WindowDelimiterType,
	}
	if wd.OffsetValue != nil {
		node["OffsetValue"] = scalarExpressionToJSON(wd.OffsetValue)
	}
	return addSpan(node, frag(wd))
}

// ======================= New Statement JSON Functions =======================

func tableSampleClauseToJSON(tsc *ast.TableSampleClause) jsonNode {
	node := jsonNode{
		"$type":  "TableSampleClause",
		"System": tsc.System,
	}
	if tsc.SampleNumber != nil {
		node["SampleNumber"] = scalarExpressionToJSON(tsc.SampleNumber)
	}
	node["TableSampleClauseOption"] = tsc.TableSampleClauseOption
	if tsc.RepeatSeed != nil {
		node["RepeatSeed"] = scalarExpressionToJSON(tsc.RepeatSeed)
	}
	return addSpan(node, frag(tsc))
}

func temporalClauseToJSON(tc *ast.TemporalClause) jsonNode {
	node := jsonNode{
		"$type": "TemporalClause",
	}
	if tc.TemporalClauseType != "" {
		node["TemporalClauseType"] = tc.TemporalClauseType
	}
	if tc.StartTime != nil {
		node["StartTime"] = scalarExpressionToJSON(tc.StartTime)
	}
	if tc.EndTime != nil {
		node["EndTime"] = scalarExpressionToJSON(tc.EndTime)
	}
	return addSpan(node, frag(tc))
}

func tableHintToJSON(h ast.TableHintType) jsonNode {
	switch th := h.(type) {
	case *ast.TableHint:
		node := jsonNode{
			"$type": "TableHint",
		}
		if th.HintKind != "" {
			node["HintKind"] = th.HintKind
		}
		return addSpan(node, frag(h))
	case *ast.IndexTableHint:
		node := jsonNode{
			"$type": "IndexTableHint",
		}
		if len(th.IndexValues) > 0 {
			values := make([]jsonNode, len(th.IndexValues))
			for i, v := range th.IndexValues {
				values[i] = identifierOrValueExpressionToJSON(v)
			}
			node["IndexValues"] = values
		}
		if th.HintKind != "" {
			node["HintKind"] = th.HintKind
		}
		return addSpan(node, frag(h))
	case *ast.LiteralTableHint:
		node := jsonNode{
			"$type": "LiteralTableHint",
		}
		if th.Value != nil {
			node["Value"] = scalarExpressionToJSON(th.Value)
		}
		if th.HintKind != "" {
			node["HintKind"] = th.HintKind
		}
		return addSpan(node, frag(h))
	case *ast.ForceSeekTableHint:
		node := jsonNode{
			"$type": "ForceSeekTableHint",
		}
		if th.IndexValue != nil {
			node["IndexValue"] = identifierOrValueExpressionToJSON(th.IndexValue)
		}
		if len(th.ColumnValues) > 0 {
			cols := make([]jsonNode, len(th.ColumnValues))
			for i, c := range th.ColumnValues {
				cols[i] = columnReferenceExpressionToJSON(c)
			}
			node["ColumnValues"] = cols
		}
		if th.HintKind != "" {
			node["HintKind"] = th.HintKind
		}
		return addSpan(node, frag(h))
	default:
		return addSpan(jsonNode{"$type": "TableHint"}, frag(h))
	}
}

func insertStatementToJSON(s *ast.InsertStatement) jsonNode {
	node := jsonNode{
		"$type": "InsertStatement",
	}
	if s.InsertSpecification != nil {
		node["InsertSpecification"] = insertSpecificationToJSON(s.InsertSpecification)
	}
	if s.WithCtesAndXmlNamespaces != nil {
		node["WithCtesAndXmlNamespaces"] = withCtesAndXmlNamespacesToJSON(s.WithCtesAndXmlNamespaces)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return addSpan(node, frag(s))
}

func dataModificationSpecificationToJSON(spec ast.DataModificationSpecification) jsonNode {
	switch s := spec.(type) {
	case *ast.InsertSpecification:
		return addSpan(insertSpecificationToJSON(s), frag(spec))
	case *ast.UpdateSpecification:
		return addSpan(updateSpecificationToJSON(s), frag(spec))
	case *ast.DeleteSpecification:
		return addSpan(deleteSpecificationToJSON(s), frag(spec))
	case *ast.MergeSpecification:
		return addSpan(mergeSpecificationToJSON(s), frag(spec))
	default:
		return addSpan(jsonNode{"$type": "UnknownDataModificationSpecification"}, frag(spec))
	}
}

func insertSpecificationToJSON(spec *ast.InsertSpecification) jsonNode {
	node := jsonNode{
		"$type":        "InsertSpecification",
		"InsertOption": "None",
	}
	if spec.InsertOption != "" {
		node["InsertOption"] = spec.InsertOption
	}
	if spec.InsertSource != nil {
		node["InsertSource"] = insertSourceToJSON(spec.InsertSource)
	}
	if spec.Target != nil {
		node["Target"] = tableReferenceToJSON(spec.Target)
	}
	if spec.TopRowFilter != nil {
		node["TopRowFilter"] = topRowFilterToJSON(spec.TopRowFilter)
	}
	if spec.OutputClause != nil {
		node["OutputClause"] = outputClauseToJSON(spec.OutputClause)
	}
	if spec.OutputIntoClause != nil {
		node["OutputIntoClause"] = outputIntoClauseToJSON(spec.OutputIntoClause)
	}
	if len(spec.Columns) > 0 {
		cols := make([]jsonNode, len(spec.Columns))
		for i, c := range spec.Columns {
			cols[i] = scalarExpressionToJSON(c)
		}
		node["Columns"] = cols
	}
	return addSpan(node, frag(spec))
}

func outputClauseToJSON(oc *ast.OutputClause) jsonNode {
	node := jsonNode{
		"$type": "OutputClause",
	}
	if len(oc.SelectColumns) > 0 {
		cols := make([]jsonNode, len(oc.SelectColumns))
		for i, c := range oc.SelectColumns {
			cols[i] = selectElementToJSON(c)
		}
		node["SelectColumns"] = cols
	}
	return addSpan(node, frag(oc))
}

func outputIntoClauseToJSON(oic *ast.OutputIntoClause) jsonNode {
	node := jsonNode{
		"$type": "OutputIntoClause",
	}
	if len(oic.SelectColumns) > 0 {
		cols := make([]jsonNode, len(oic.SelectColumns))
		for i, c := range oic.SelectColumns {
			cols[i] = selectElementToJSON(c)
		}
		node["SelectColumns"] = cols
	}
	if oic.IntoTable != nil {
		node["IntoTable"] = tableReferenceToJSON(oic.IntoTable)
	}
	if len(oic.IntoTableColumns) > 0 {
		cols := make([]jsonNode, len(oic.IntoTableColumns))
		for i, c := range oic.IntoTableColumns {
			cols[i] = columnReferenceExpressionToJSON(c)
		}
		node["IntoTableColumns"] = cols
	}
	return addSpan(node, frag(oic))
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
		return addSpan(node, frag(src))
	case *ast.SelectInsertSource:
		node := jsonNode{
			"$type": "SelectInsertSource",
		}
		if s.Select != nil {
			node["Select"] = queryExpressionToJSON(s.Select)
		}
		return addSpan(node, frag(src))
	case *ast.ExecuteInsertSource:
		node := jsonNode{
			"$type": "ExecuteInsertSource",
		}
		if s.Execute != nil {
			node["Execute"] = executeSpecificationToJSON(s.Execute)
		}
		return addSpan(node, frag(src))
	default:
		return addSpan(jsonNode{"$type": "UnknownInsertSource"}, frag(src))
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
	return addSpan(node, frag(rv))
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
	return addSpan(node, frag(spec))
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
		if e.AdHocDataSource != nil {
			node["AdHocDataSource"] = adHocDataSourceToJSON(e.AdHocDataSource)
		}
		return addSpan(node, frag(entity))
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
		return addSpan(node, frag(entity))
	default:
		return addSpan(jsonNode{"$type": "UnknownExecutableEntity"}, frag(entity))
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
	return addSpan(node, frag(prn))
}

func procedureReferenceToJSON(pr *ast.ProcedureReference) jsonNode {
	node := jsonNode{
		"$type": "ProcedureReference",
	}
	if pr.Name != nil {
		node["Name"] = schemaObjectNameToJSON(pr.Name)
	}
	if pr.Number != nil {
		node["Number"] = scalarExpressionToJSON(pr.Number)
	}
	return addSpan(node, frag(pr))
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
	return addSpan(node, frag(ep))
}

func adHocDataSourceToJSON(ds *ast.AdHocDataSource) jsonNode {
	node := jsonNode{
		"$type": "AdHocDataSource",
	}
	if ds.ProviderName != nil {
		node["ProviderName"] = scalarExpressionToJSON(ds.ProviderName)
	}
	if ds.InitString != nil {
		node["InitString"] = scalarExpressionToJSON(ds.InitString)
	}
	return addSpan(node, frag(ds))
}

func updateStatementToJSON(s *ast.UpdateStatement) jsonNode {
	node := jsonNode{
		"$type": "UpdateStatement",
	}
	if s.UpdateSpecification != nil {
		node["UpdateSpecification"] = updateSpecificationToJSON(s.UpdateSpecification)
	}
	if s.WithCtesAndXmlNamespaces != nil {
		node["WithCtesAndXmlNamespaces"] = withCtesAndXmlNamespacesToJSON(s.WithCtesAndXmlNamespaces)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return addSpan(node, frag(s))
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
	if spec.TopRowFilter != nil {
		node["TopRowFilter"] = topRowFilterToJSON(spec.TopRowFilter)
	}
	if spec.FromClause != nil {
		node["FromClause"] = fromClauseToJSON(spec.FromClause)
	}
	if spec.WhereClause != nil {
		node["WhereClause"] = whereClauseToJSON(spec.WhereClause)
	}
	if spec.OutputClause != nil {
		node["OutputClause"] = outputClauseToJSON(spec.OutputClause)
	}
	if spec.OutputIntoClause != nil {
		node["OutputIntoClause"] = outputIntoClauseToJSON(spec.OutputIntoClause)
	}
	return addSpan(node, frag(spec))
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
		return addSpan(node, frag(sc))
	case *ast.FunctionCallSetClause:
		node := jsonNode{
			"$type": "FunctionCallSetClause",
		}
		if c.MutatorFunction != nil {
			node["MutatorFunction"] = scalarExpressionToJSON(c.MutatorFunction)
		}
		return addSpan(node, frag(sc))
	default:
		return addSpan(jsonNode{"$type": "UnknownSetClause"}, frag(sc))
	}
}

func deleteStatementToJSON(s *ast.DeleteStatement) jsonNode {
	node := jsonNode{
		"$type": "DeleteStatement",
	}
	if s.DeleteSpecification != nil {
		node["DeleteSpecification"] = deleteSpecificationToJSON(s.DeleteSpecification)
	}
	if s.WithCtesAndXmlNamespaces != nil {
		node["WithCtesAndXmlNamespaces"] = withCtesAndXmlNamespacesToJSON(s.WithCtesAndXmlNamespaces)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return addSpan(node, frag(s))
}

func mergeStatementToJSON(s *ast.MergeStatement) jsonNode {
	node := jsonNode{
		"$type": "MergeStatement",
	}
	if s.MergeSpecification != nil {
		node["MergeSpecification"] = mergeSpecificationToJSON(s.MergeSpecification)
	}
	if s.WithCtesAndXmlNamespaces != nil {
		node["WithCtesAndXmlNamespaces"] = withCtesAndXmlNamespacesToJSON(s.WithCtesAndXmlNamespaces)
	}
	if len(s.OptimizerHints) > 0 {
		hints := make([]jsonNode, len(s.OptimizerHints))
		for i, h := range s.OptimizerHints {
			hints[i] = optimizerHintToJSON(h)
		}
		node["OptimizerHints"] = hints
	}
	return addSpan(node, frag(s))
}

func mergeSpecificationToJSON(spec *ast.MergeSpecification) jsonNode {
	node := jsonNode{
		"$type": "MergeSpecification",
	}
	if spec.TableAlias != nil {
		node["TableAlias"] = identifierToJSON(spec.TableAlias)
	}
	if spec.TableReference != nil {
		node["TableReference"] = tableReferenceToJSON(spec.TableReference)
	}
	if spec.SearchCondition != nil {
		node["SearchCondition"] = booleanExpressionToJSON(spec.SearchCondition)
	}
	if len(spec.ActionClauses) > 0 {
		clauses := make([]jsonNode, len(spec.ActionClauses))
		for i, c := range spec.ActionClauses {
			clauses[i] = mergeActionClauseToJSON(c)
		}
		node["ActionClauses"] = clauses
	}
	if spec.Target != nil {
		node["Target"] = tableReferenceToJSON(spec.Target)
	}
	if spec.OutputClause != nil {
		node["OutputClause"] = outputClauseToJSON(spec.OutputClause)
	}
	if spec.TopRowFilter != nil {
		node["TopRowFilter"] = topRowFilterToJSON(spec.TopRowFilter)
	}
	return addSpan(node, frag(spec))
}

func mergeActionClauseToJSON(c *ast.MergeActionClause) jsonNode {
	node := jsonNode{
		"$type":     "MergeActionClause",
		"Condition": c.Condition,
	}
	if c.SearchCondition != nil {
		node["SearchCondition"] = booleanExpressionToJSON(c.SearchCondition)
	}
	if c.Action != nil {
		node["Action"] = mergeActionToJSON(c.Action)
	}
	return addSpan(node, frag(c))
}

func mergeActionToJSON(a ast.MergeAction) jsonNode {
	switch action := a.(type) {
	case *ast.DeleteMergeAction:
		return addSpan(jsonNode{"$type": "DeleteMergeAction"}, frag(a))
	case *ast.UpdateMergeAction:
		node := jsonNode{"$type": "UpdateMergeAction"}
		if len(action.SetClauses) > 0 {
			clauses := make([]jsonNode, len(action.SetClauses))
			for i, sc := range action.SetClauses {
				clauses[i] = setClauseToJSON(sc)
			}
			node["SetClauses"] = clauses
		}
		return addSpan(node, frag(a))
	case *ast.InsertMergeAction:
		node := jsonNode{"$type": "InsertMergeAction"}
		if len(action.Columns) > 0 {
			cols := make([]jsonNode, len(action.Columns))
			for i, col := range action.Columns {
				cols[i] = columnReferenceExpressionToJSON(col)
			}
			node["Columns"] = cols
		}
		if action.Source != nil {
			node["Source"] = insertSourceToJSON(action.Source)
		}
		return addSpan(node, frag(a))
	default:
		return addSpan(jsonNode{"$type": "UnknownMergeAction"}, frag(a))
	}
}

func withCtesAndXmlNamespacesToJSON(w *ast.WithCtesAndXmlNamespaces) jsonNode {
	node := jsonNode{
		"$type": "WithCtesAndXmlNamespaces",
	}
	if w.XmlNamespaces != nil {
		node["XmlNamespaces"] = xmlNamespacesToJSON(w.XmlNamespaces)
	}
	if len(w.CommonTableExpressions) > 0 {
		ctes := make([]jsonNode, len(w.CommonTableExpressions))
		for i, cte := range w.CommonTableExpressions {
			ctes[i] = commonTableExpressionToJSON(cte)
		}
		node["CommonTableExpressions"] = ctes
	}
	if w.ChangeTrackingContext != nil {
		node["ChangeTrackingContext"] = scalarExpressionToJSON(w.ChangeTrackingContext)
	}
	return addSpan(node, frag(w))
}

func commonTableExpressionToJSON(cte *ast.CommonTableExpression) jsonNode {
	node := jsonNode{
		"$type": "CommonTableExpression",
	}
	if cte.ExpressionName != nil {
		node["ExpressionName"] = identifierToJSON(cte.ExpressionName)
	}
	if len(cte.Columns) > 0 {
		cols := make([]jsonNode, len(cte.Columns))
		for i, col := range cte.Columns {
			cols[i] = identifierToJSON(col)
		}
		node["Columns"] = cols
	}
	if cte.QueryExpression != nil {
		node["QueryExpression"] = queryExpressionToJSON(cte.QueryExpression)
	}
	return addSpan(node, frag(cte))
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
	if spec.TopRowFilter != nil {
		node["TopRowFilter"] = topRowFilterToJSON(spec.TopRowFilter)
	}
	if spec.OutputClause != nil {
		node["OutputClause"] = outputClauseToJSON(spec.OutputClause)
	}
	if spec.OutputIntoClause != nil {
		node["OutputIntoClause"] = outputIntoClauseToJSON(spec.OutputIntoClause)
	}
	return addSpan(node, frag(spec))
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
	return addSpan(node, frag(wc))
}

func cursorIdToJSON(cid *ast.CursorId) jsonNode {
	node := jsonNode{
		"$type": "CursorId",
	}
	node["IsGlobal"] = cid.IsGlobal
	if cid.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(cid.Name)
	}
	return addSpan(node, frag(cid))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(elem))
}

func declareTableVariableStatementToJSON(s *ast.DeclareTableVariableStatement) jsonNode {
	node := jsonNode{
		"$type": "DeclareTableVariableStatement",
	}
	if s.Body != nil {
		node["Body"] = declareTableVariableBodyToJSON(s.Body)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(body))
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
	return addSpan(node, frag(dt))
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
	return addSpan(node, frag(s))
}

func cursorDefinitionToJSON(cd *ast.CursorDefinition) jsonNode {
	node := jsonNode{
		"$type": "CursorDefinition",
	}
	if len(cd.Options) > 0 {
		opts := make([]jsonNode, len(cd.Options))
		for i, opt := range cd.Options {
			opts[i] = addSpan(jsonNode{
				"$type":      "CursorOption",
				"OptionKind": opt.OptionKind,
			}, frag(opt))
		}
		node["Options"] = opts
	}
	if cd.Select != nil {
		node["Select"] = selectStatementToJSON(cd.Select)
	}
	return addSpan(node, frag(cd))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func beginEndBlockStatementToJSON(s *ast.BeginEndBlockStatement) jsonNode {
	node := jsonNode{
		"$type": "BeginEndBlockStatement",
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(o))
	case *ast.LiteralAtomicBlockOption:
		node := jsonNode{
			"$type":      "LiteralAtomicBlockOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return addSpan(node, frag(o))
	case *ast.OnOffAtomicBlockOption:
		return addSpan(jsonNode{
			"$type":       "OnOffAtomicBlockOption",
			"OptionState": opt.OptionState,
			"OptionKind":  opt.OptionKind,
		}, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownAtomicBlockOption"}, frag(o))
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
	return addSpan(node, frag(sl))
}

func beginDialogStatementToJSON(s *ast.BeginDialogStatement) jsonNode {
	node := jsonNode{
		"$type":          "BeginDialogStatement",
		"IsConversation": s.IsConversation,
	}
	if s.Handle != nil {
		node["Handle"] = scalarExpressionToJSON(s.Handle)
	}
	if s.InitiatorServiceName != nil {
		node["InitiatorServiceName"] = identifierOrValueExpressionToJSON(s.InitiatorServiceName)
	}
	if s.TargetServiceName != nil {
		node["TargetServiceName"] = scalarExpressionToJSON(s.TargetServiceName)
	}
	if s.ContractName != nil {
		node["ContractName"] = identifierOrValueExpressionToJSON(s.ContractName)
	}
	if s.InstanceSpec != nil {
		node["InstanceSpec"] = scalarExpressionToJSON(s.InstanceSpec)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = dialogOptionToJSON(o)
		}
		node["Options"] = options
	}
	return addSpan(node, frag(s))
}

func dialogOptionToJSON(o ast.DialogOption) jsonNode {
	switch opt := o.(type) {
	case *ast.ScalarExpressionDialogOption:
		node := jsonNode{
			"$type":      "ScalarExpressionDialogOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return addSpan(node, frag(o))
	case *ast.OnOffDialogOption:
		return addSpan(jsonNode{
			"$type":       "OnOffDialogOption",
			"OptionState": opt.OptionState,
			"OptionKind":  opt.OptionKind,
		}, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownDialogOption"}, frag(o))
	}
}

func beginConversationTimerStatementToJSON(s *ast.BeginConversationTimerStatement) jsonNode {
	node := jsonNode{
		"$type": "BeginConversationTimerStatement",
	}
	if s.Handle != nil {
		node["Handle"] = scalarExpressionToJSON(s.Handle)
	}
	if s.Timeout != nil {
		node["Timeout"] = scalarExpressionToJSON(s.Timeout)
	}
	return addSpan(node, frag(s))
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
	if len(s.ViewOptions) > 0 {
		opts := make([]jsonNode, len(s.ViewOptions))
		for i, opt := range s.ViewOptions {
			opts[i] = viewOptionToJSON(opt)
		}
		node["ViewOptions"] = opts
	}
	if s.SelectStatement != nil {
		node["SelectStatement"] = selectStatementToJSON(s.SelectStatement)
	}
	node["WithCheckOption"] = s.WithCheckOption
	node["IsMaterialized"] = s.IsMaterialized
	return addSpan(node, frag(s))
}

func createOrAlterViewStatementToJSON(s *ast.CreateOrAlterViewStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateOrAlterViewStatement",
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
	if len(s.ViewOptions) > 0 {
		opts := make([]jsonNode, len(s.ViewOptions))
		for i, opt := range s.ViewOptions {
			opts[i] = viewOptionToJSON(opt)
		}
		node["ViewOptions"] = opts
	}
	if s.SelectStatement != nil {
		node["SelectStatement"] = selectStatementToJSON(s.SelectStatement)
	}
	node["WithCheckOption"] = s.WithCheckOption
	node["IsMaterialized"] = s.IsMaterialized
	return addSpan(node, frag(s))
}

func alterViewStatementToJSON(s *ast.AlterViewStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterViewStatement",
	}
	node["IsRebuild"] = s.IsRebuild
	node["IsDisable"] = s.IsDisable
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
	if len(s.ViewOptions) > 0 {
		opts := make([]jsonNode, len(s.ViewOptions))
		for i, opt := range s.ViewOptions {
			opts[i] = viewOptionToJSON(opt)
		}
		node["ViewOptions"] = opts
	}
	if s.SelectStatement != nil {
		node["SelectStatement"] = selectStatementToJSON(s.SelectStatement)
	}
	node["WithCheckOption"] = s.WithCheckOption
	node["IsMaterialized"] = s.IsMaterialized
	return addSpan(node, frag(s))
}

func viewOptionToJSON(opt ast.ViewOption) jsonNode {
	switch o := opt.(type) {
	case *ast.ViewStatementOption:
		return addSpan(jsonNode{
			"$type":      "ViewOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.ViewDistributionOption:
		node := jsonNode{
			"$type":      "ViewDistributionOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			switch v := o.Value.(type) {
			case *ast.ViewHashDistributionPolicy:
				valueNode := jsonNode{
					"$type": "ViewHashDistributionPolicy",
				}
				if v.DistributionColumn != nil {
					valueNode["DistributionColumn"] = identifierToJSON(v.DistributionColumn)
				}
				if len(v.DistributionColumns) > 0 {
					cols := make([]jsonNode, len(v.DistributionColumns))
					for i, c := range v.DistributionColumns {
						// First column is same as DistributionColumn, use $ref
						if i == 0 && v.DistributionColumn != nil {
							cols[i] = jsonNode{"$ref": "Identifier"}
						} else {
							cols[i] = identifierToJSON(c)
						}
					}
					valueNode["DistributionColumns"] = cols
				}
				node["Value"] = addSpan(valueNode, frag(v))
			case *ast.ViewRoundRobinDistributionPolicy:
				node["Value"] = addSpan(jsonNode{
					"$type": "ViewRoundRobinDistributionPolicy",
				}, frag(v))
			}
		}
		return addSpan(node, frag(opt))
	case *ast.ViewForAppendOption:
		return addSpan(jsonNode{
			"$type":      "ViewForAppendOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownViewOption"}, frag(opt))
	}
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
	return addSpan(node, frag(s))
}

func executeStatementToJSON(s *ast.ExecuteStatement) jsonNode {
	node := jsonNode{
		"$type": "ExecuteStatement",
	}
	if s.ExecuteSpecification != nil {
		node["ExecuteSpecification"] = executeSpecificationToJSON(s.ExecuteSpecification)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = executeOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
}

func executeOptionToJSON(opt ast.ExecuteOptionType) jsonNode {
	switch o := opt.(type) {
	case *ast.ExecuteOption:
		return addSpan(jsonNode{
			"$type":      "ExecuteOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.ResultSetsExecuteOption:
		node := jsonNode{
			"$type":                "ResultSetsExecuteOption",
			"ResultSetsOptionKind": o.ResultSetsOptionKind,
			"OptionKind":           o.OptionKind,
		}
		if len(o.Definitions) > 0 {
			defs := make([]jsonNode, len(o.Definitions))
			for i, def := range o.Definitions {
				defs[i] = resultSetDefinitionToJSON(def)
			}
			node["Definitions"] = defs
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{}, frag(opt))
	}
}

func resultSetDefinitionToJSON(def ast.ResultSetDefinitionType) jsonNode {
	switch d := def.(type) {
	case *ast.ResultSetDefinition:
		return addSpan(jsonNode{
			"$type":         "ResultSetDefinition",
			"ResultSetType": d.ResultSetType,
		}, frag(def))
	case *ast.InlineResultSetDefinition:
		node := jsonNode{
			"$type":         "InlineResultSetDefinition",
			"ResultSetType": d.ResultSetType,
		}
		if len(d.ResultColumnDefinitions) > 0 {
			cols := make([]jsonNode, len(d.ResultColumnDefinitions))
			for i, col := range d.ResultColumnDefinitions {
				cols[i] = resultColumnDefinitionToJSON(col)
			}
			node["ResultColumnDefinitions"] = cols
		}
		return addSpan(node, frag(def))
	case *ast.SchemaObjectResultSetDefinition:
		node := jsonNode{
			"$type":         "SchemaObjectResultSetDefinition",
			"ResultSetType": d.ResultSetType,
		}
		if d.Name != nil {
			node["Name"] = schemaObjectNameToJSON(d.Name)
		}
		return addSpan(node, frag(def))
	default:
		return addSpan(jsonNode{}, frag(def))
	}
}

func resultColumnDefinitionToJSON(col *ast.ResultColumnDefinition) jsonNode {
	node := jsonNode{
		"$type": "ResultColumnDefinition",
	}
	if col.ColumnDefinition != nil {
		colDefNode := jsonNode{
			"$type": "ColumnDefinitionBase",
		}
		if col.ColumnDefinition.ColumnIdentifier != nil {
			colDefNode["ColumnIdentifier"] = identifierToJSON(col.ColumnDefinition.ColumnIdentifier)
		}
		if col.ColumnDefinition.DataType != nil {
			colDefNode["DataType"] = dataTypeReferenceToJSON(col.ColumnDefinition.DataType)
		}
		node["ColumnDefinition"] = colDefNode
	}
	if col.Nullable != nil {
		node["Nullable"] = jsonNode{
			"$type":    "NullableConstraintDefinition",
			"Nullable": col.Nullable.Nullable,
		}
	}
	return addSpan(node, frag(col))
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
	return addSpan(node, frag(s))
}

func executeContextToJSON(c *ast.ExecuteContext) jsonNode {
	node := jsonNode{
		"$type": "ExecuteContext",
		"Kind":  c.Kind,
	}
	if c.Principal != nil {
		node["Principal"] = scalarExpressionToJSON(c.Principal)
	}
	return addSpan(node, frag(c))
}

func returnStatementToJSON(s *ast.ReturnStatement) jsonNode {
	node := jsonNode{
		"$type": "ReturnStatement",
	}
	if s.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(s.Expression)
	}
	return addSpan(node, frag(s))
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
	astStart := p.curTok

	// Consume TABLE
	p.nextToken()

	stmt := &ast.CreateTableStatement{}

	// Parse table name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.SchemaObjectName = name

	// Check for AS FILETABLE
	if p.curTok.Type == TokenAs {
		p.nextToken() // consume AS
		if strings.ToUpper(p.curTok.Literal) == "FILETABLE" {
			stmt.AsFileTable = true
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "NODE" {
			stmt.AsNode = true
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "EDGE" {
			stmt.AsEdge = true
			p.nextToken()
		}
	}

	// Check for ON, TEXTIMAGE_ON, FILESTREAM_ON, WITH clauses (for AS FILETABLE)
	if p.curTok.Type != TokenLParen {
		spanV1, spanErr1 := p.parseCreateTableOptions(stmt)
		return spanned(p, spanV1, astStart), spanErr1
	}
	p.nextToken()

	// Check if this is a CTAS column list (just column names) or regular table definition
	// CTAS columns: (col1, col2) - identifier followed by comma or )
	// Regular: (col1 INT, col2 VARCHAR(50)) - identifier followed by data type
	isCtasColumnList := false
	if p.curTok.Type == TokenIdent {
		// Check if next token is comma or rparen (CTAS column list)
		// Use peekTok directly instead of advancing to avoid lexer state issues
		if p.peekTok.Type == TokenComma || p.peekTok.Type == TokenRParen {
			isCtasColumnList = true
		}
	}

	if isCtasColumnList {
		// Parse CTAS column names
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			col := p.parseIdentifier()
			stmt.CtasColumns = append(stmt.CtasColumns, col)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	} else {
		stmt.Definition = &ast.TableDefinition{}

		// Parse column definitions and table constraints
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			upperLit := strings.ToUpper(p.curTok.Literal)

			// Check for table-level constraints
			if upperLit == "CONSTRAINT" {
				constraint, err := p.parseNamedTableConstraint()
				if err != nil {
					p.skipToEndOfStatement()
					return spanned(p, stmt, astStart), nil
				}
				if constraint != nil {
					stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
				}
			} else if upperLit == "PRIMARY" || upperLit == "UNIQUE" || upperLit == "FOREIGN" || upperLit == "CHECK" {
				constraint, err := p.parseUnnamedTableConstraint()
				if err != nil {
					p.skipToEndOfStatement()
					return spanned(p, stmt, astStart), nil
				}
				if constraint != nil {
					stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
				}
			} else if upperLit == "PERIOD" {
				// Parse PERIOD FOR SYSTEM_TIME
				p.nextToken() // consume PERIOD
				if strings.ToUpper(p.curTok.Literal) == "FOR" {
					p.nextToken() // consume FOR
				}
				if strings.ToUpper(p.curTok.Literal) == "SYSTEM_TIME" {
					p.nextToken() // consume SYSTEM_TIME
				}
				// Expect (
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
				}
				// Parse start column
				periodStart := p.curTok
				startCol := p.parseIdentifier()
				// Expect comma
				if p.curTok.Type == TokenComma {
					p.nextToken() // consume ,
				}
				// Parse end column
				endCol := p.parseIdentifier()
				// Expect )
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
				stp := &ast.SystemTimePeriodDefinition{
					StartTimeColumn: startCol,
					EndTimeColumn:   endCol,
				}
				// ScriptDom spans the period from the first column name
				// through the closing paren.
				p.spanFrom(periodStart, stp)
				stmt.Definition.SystemTimePeriod = stp
			} else if upperLit == "INDEX" {
				// Parse inline index definition
				indexDef, err := p.parseInlineIndexDefinition()
				if err != nil {
					p.skipToEndOfStatement()
					return spanned(p, stmt, astStart), nil
				}
				stmt.Definition.Indexes = append(stmt.Definition.Indexes, indexDef)
			} else if upperLit == "CONNECTION" {
				// Parse unnamed CONNECTION constraint for graph edge tables
				connectionTok := p.curTok
				p.nextToken() // consume CONNECTION
				constraint := &ast.GraphConnectionConstraintDefinition{}
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
						conn := &ast.GraphConnectionBetweenNodes{}
						// Parse FromNode
						fromNode, err := p.parseSchemaObjectName()
						if err != nil {
							p.skipToEndOfStatement()
							return spanned(p, stmt, astStart), nil
						}
						conn.FromNode = fromNode
						// Expect TO
						if strings.ToUpper(p.curTok.Literal) == "TO" {
							p.nextToken() // consume TO
						}
						// Parse ToNode
						toNode, err := p.parseSchemaObjectName()
						if err != nil {
							p.skipToEndOfStatement()
							return spanned(p, stmt, astStart), nil
						}
						conn.ToNode = toNode
						p.spanFromChild(conn, conn.FromNode)
						constraint.FromNodeToNodeList = append(constraint.FromNodeToNodeList, conn)
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
				// Check for ON DELETE CASCADE
				if p.curTok.Type == TokenOn && strings.ToUpper(p.peekTok.Literal) == "DELETE" {
					p.nextToken() // consume ON
					p.nextToken() // consume DELETE
					if strings.ToUpper(p.curTok.Literal) == "CASCADE" {
						constraint.DeleteAction = "Cascade"
						p.nextToken() // consume CASCADE
					} else if strings.ToUpper(p.curTok.Literal) == "NO" {
						p.nextToken() // consume NO
						if strings.ToUpper(p.curTok.Literal) == "ACTION" {
							constraint.DeleteAction = "NoAction"
							p.nextToken() // consume ACTION
						}
					}
				}
				p.spanFrom(connectionTok, constraint)
				stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
			} else {
				// Parse column definition
				colDef, err := p.parseColumnDefinition()
				if err != nil {
					p.skipToEndOfStatement()
					return spanned(p, stmt, astStart), nil
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
	}

	// Parse optional ON filegroup, TEXTIMAGE_ON, FILESTREAM_ON, and WITH clauses
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if p.curTok.Type == TokenOn {
			p.nextToken() // consume ON
			// Parse filegroup or partition scheme with optional columns
			fg, err := p.parseFileGroupOrPartitionScheme()
			if err != nil {
				return nil, err
			}
			stmt.OnFileGroupOrPartitionScheme = fg
		} else if upperLit == "TEXTIMAGE_ON" {
			p.nextToken() // consume TEXTIMAGE_ON
			// Parse filegroup identifier or string literal
			if p.curTok.Type == TokenString {
				value := p.curTok.Literal
				// Strip quotes from string literal
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				stmt.TextImageOn = &ast.IdentifierOrValueExpression{
					Value:           value,
					ValueExpression: p.strLit(value, false),
				}
				p.nextToken()
			} else {
				ident := p.parseIdentifier()
				stmt.TextImageOn = &ast.IdentifierOrValueExpression{
					Value:      ident.Value,
					Identifier: ident,
				}
			}
		} else if upperLit == "FILESTREAM_ON" {
			p.nextToken() // consume FILESTREAM_ON
			// Parse filegroup identifier or string literal
			if p.curTok.Type == TokenString {
				value := p.curTok.Literal
				// Strip quotes from string literal
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				stmt.FileStreamOn = &ast.IdentifierOrValueExpression{
					Value:           value,
					ValueExpression: p.strLit(value, false),
				}
				p.nextToken()
			} else {
				ident := p.parseIdentifier()
				stmt.FileStreamOn = &ast.IdentifierOrValueExpression{
					Value:      ident.Value,
					Identifier: ident,
				}
			}
		} else if p.curTok.Type == TokenWith {
			// Parse WITH clause with table options
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Parse table options
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optNameTok := p.curTok
					optionName := strings.ToUpper(p.curTok.Literal)
					p.nextToken() // consume option name

					if optionName == "DATA_COMPRESSION" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						opt, err := p.parseDataCompressionOption()
						if err != nil {
							break
						}
						stmt.Options = append(stmt.Options, &ast.TableDataCompressionOption{
							DataCompressionOption: opt,
							OptionKind:            "DataCompression",
						})
					} else if optionName == "XML_COMPRESSION" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						opt, err := p.parseXmlCompressionOption()
						if err != nil {
							break
						}
						stmt.Options = append(stmt.Options, &ast.TableXmlCompressionOption{
							XmlCompressionOption: opt,
							OptionKind:           "XmlCompression",
						})
					} else if optionName == "MEMORY_OPTIMIZED" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						stateTok := p.curTok
						stateUpper := strings.ToUpper(p.curTok.Literal)
						state := "On"
						if stateUpper == "OFF" {
							state = "Off"
						}
						p.nextToken() // consume ON/OFF
						mo := &ast.MemoryOptimizedTableOption{
							OptionKind:  "MemoryOptimized",
							OptionState: state,
						}
						p.tokSpan(mo, stateTok)
						stmt.Options = append(stmt.Options, mo)
					} else if optionName == "DURABILITY" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						valueUpper := strings.ToUpper(p.curTok.Literal)
						durabilityKind := "SchemaOnly"
						if valueUpper == "SCHEMA_AND_DATA" {
							durabilityKind = "SchemaAndData"
						}
						p.nextToken() // consume value
						stmt.Options = append(stmt.Options, &ast.DurabilityTableOption{
							OptionKind:                "Durability",
							DurabilityTableOptionKind: durabilityKind,
						})
					} else if optionName == "REMOTE_DATA_ARCHIVE" {
						opt, err := p.parseRemoteDataArchiveTableOption(false)
						if err != nil {
							return nil, err
						}
						stmt.Options = append(stmt.Options, opt)
					} else if optionName == "SYSTEM_VERSIONING" {
						opt, err := p.parseSystemVersioningTableOption(optNameTok)
						if err != nil {
							return nil, err
						}
						stmt.Options = append(stmt.Options, opt)
					} else if optionName == "LEDGER" {
						opt, err := p.parseLedgerTableOption()
						if err != nil {
							return nil, err
						}
						stmt.Options = append(stmt.Options, opt)
					} else if optionName == "CLUSTERED" {
						// Could be CLUSTERED INDEX or CLUSTERED COLUMNSTORE INDEX
						if strings.ToUpper(p.curTok.Literal) == "COLUMNSTORE" {
							p.nextToken() // consume COLUMNSTORE
							if p.curTok.Type == TokenIndex {
								p.nextToken() // consume INDEX
							}
							indexType := &ast.TableClusteredIndexType{
								ColumnStore: true,
							}
							// Check for ORDER(columns)
							if strings.ToUpper(p.curTok.Literal) == "ORDER" {
								p.nextToken() // consume ORDER
								if p.curTok.Type == TokenLParen {
									p.nextToken() // consume (
									firstColTok := p.curTok
									for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
										col := &ast.ColumnReferenceExpression{
											ColumnType: "Regular",
											MultiPartIdentifier: &ast.MultiPartIdentifier{
												Identifiers: []*ast.Identifier{p.parseIdentifier()},
												Count:       1,
											},
										}
										indexType.OrderedColumns = append(indexType.OrderedColumns, col)
										if p.curTok.Type == TokenComma {
											p.nextToken()
										} else {
											break
										}
									}
									if p.curTok.Type == TokenRParen {
										p.nextToken()
									}
									// ScriptDom spans the index type from the first
									// ordered column through the closing paren.
									p.spanFrom(firstColTok, indexType)
								}
							}
							stmt.Options = append(stmt.Options, &ast.TableIndexOption{
								Value:      indexType,
								OptionKind: "LockEscalation",
							})
						} else if p.curTok.Type == TokenIndex {
							p.nextToken() // consume INDEX
							// Parse column list
							indexType := &ast.TableClusteredIndexType{
								ColumnStore: false,
							}
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								firstColTok := p.curTok
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									colTok := p.curTok
									col := &ast.ColumnWithSortOrder{
										SortOrder: ast.SortOrderNotSpecified,
										Column: &ast.ColumnReferenceExpression{
											ColumnType: "Regular",
											MultiPartIdentifier: &ast.MultiPartIdentifier{
												Identifiers: []*ast.Identifier{p.parseIdentifier()},
												Count:       1,
											},
										},
									}
									// Parse optional ASC/DESC
									sortUpper := strings.ToUpper(p.curTok.Literal)
									if sortUpper == "ASC" {
										col.SortOrder = ast.SortOrderAscending
										p.nextToken()
									} else if sortUpper == "DESC" {
										col.SortOrder = ast.SortOrderDescending
										p.nextToken()
									}
									p.spanFrom(colTok, col)
									indexType.Columns = append(indexType.Columns, col)
									if p.curTok.Type == TokenComma {
										p.nextToken()
									} else {
										break
									}
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken()
								}
								// ScriptDom spans the index type from the first
								// column through the closing paren.
								p.spanFrom(firstColTok, indexType)
							}
							stmt.Options = append(stmt.Options, &ast.TableIndexOption{
								Value:      indexType,
								OptionKind: "LockEscalation",
							})
						}
					} else if optionName == "HEAP" {
						heapType := &ast.TableNonClusteredIndexType{}
						p.tokSpan(heapType, optNameTok)
						stmt.Options = append(stmt.Options, &ast.TableIndexOption{
							Value:      heapType,
							OptionKind: "LockEscalation",
						})
					} else if optionName == "DISTRIBUTION" {
						// Parse DISTRIBUTION = HASH(col1, col2, ...) or ROUND_ROBIN or REPLICATE
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						distTypeTok := p.curTok
						distTypeUpper := strings.ToUpper(p.curTok.Literal)
						if distTypeUpper == "HASH" {
							p.nextToken() // consume HASH
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								hashPolicy := &ast.TableHashDistributionPolicy{}
								// Parse column list
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									col := p.parseIdentifier()
									if hashPolicy.DistributionColumn == nil {
										hashPolicy.DistributionColumn = col
									}
									hashPolicy.DistributionColumns = append(hashPolicy.DistributionColumns, col)
									if p.curTok.Type == TokenComma {
										p.nextToken()
									} else {
										break
									}
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken()
								}
								p.spanFrom(distTypeTok, hashPolicy)
								distOpt := &ast.TableDistributionOption{
									OptionKind: "Distribution",
									Value:      hashPolicy,
								}
								// ScriptDom positions the option on the DISTRIBUTION keyword.
								p.tokSpan(distOpt, optNameTok)
								stmt.Options = append(stmt.Options, distOpt)
							}
						} else if distTypeUpper == "ROUND_ROBIN" {
							p.nextToken() // consume ROUND_ROBIN
							rrPolicy := &ast.TableRoundRobinDistributionPolicy{}
							p.tokSpan(rrPolicy, distTypeTok)
							distOpt := &ast.TableDistributionOption{
								OptionKind: "Distribution",
								Value:      rrPolicy,
							}
							p.tokSpan(distOpt, optNameTok)
							stmt.Options = append(stmt.Options, distOpt)
						} else if distTypeUpper == "REPLICATE" {
							p.nextToken() // consume REPLICATE
							repPolicy := &ast.TableReplicateDistributionPolicy{}
							p.tokSpan(repPolicy, distTypeTok)
							distOpt := &ast.TableDistributionOption{
								OptionKind: "Distribution",
								Value:      repPolicy,
							}
							p.tokSpan(distOpt, optNameTok)
							stmt.Options = append(stmt.Options, distOpt)
						} else {
							// Unknown distribution - skip for now
							p.nextToken()
						}
					} else if optionName == "PARTITION" {
						// Parse PARTITION(column RANGE [LEFT|RIGHT] FOR VALUES (v1, v2, ...))
						if p.curTok.Type == TokenLParen {
							p.nextToken() // consume (
							partOpt := &ast.TablePartitionOption{
								OptionKind:           "Partition",
								PartitionOptionSpecs: &ast.TablePartitionOptionSpecifications{},
							}
							// Parse partition column
							partOpt.PartitionColumn = p.parseIdentifier()
							// Expect RANGE keyword
							if strings.ToUpper(p.curTok.Literal) == "RANGE" {
								p.nextToken() // consume RANGE
								// Check for LEFT or RIGHT
								var dirTok Token
								hasDir := false
								rangeDir := strings.ToUpper(p.curTok.Literal)
								if rangeDir == "LEFT" {
									partOpt.PartitionOptionSpecs.Range = "Left"
									dirTok = p.curTok
									hasDir = true
									p.nextToken()
								} else if rangeDir == "RIGHT" {
									partOpt.PartitionOptionSpecs.Range = "Right"
									dirTok = p.curTok
									hasDir = true
									p.nextToken()
								} else {
									partOpt.PartitionOptionSpecs.Range = "NotSpecified"
								}
								// Expect FOR keyword
								if strings.ToUpper(p.curTok.Literal) == "FOR" {
									p.nextToken() // consume FOR
								}
								// Expect VALUES keyword
								if strings.ToUpper(p.curTok.Literal) == "VALUES" {
									p.nextToken() // consume VALUES
								}
								// Parse boundary values list
								if p.curTok.Type == TokenLParen {
									p.nextToken() // consume (
									for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
										val, _ := p.parseScalarExpression()
										if val != nil {
											partOpt.PartitionOptionSpecs.BoundaryValues = append(partOpt.PartitionOptionSpecs.BoundaryValues, val)
										}
										if p.curTok.Type == TokenComma {
											p.nextToken()
										} else {
											break
										}
									}
									if p.curTok.Type == TokenRParen {
										// ScriptDom spans the specifications from the
										// range direction (when present) through the
										// closing paren of the VALUES list; without a
										// direction only the closing paren is covered.
										if hasDir {
											p.spanTokens(partOpt.PartitionOptionSpecs, dirTok, p.curTok)
										} else {
											p.spanTokens(partOpt.PartitionOptionSpecs, p.curTok, p.curTok)
										}
										p.nextToken() // consume )
									}
								}
							}
							if p.curTok.Type == TokenRParen {
								p.nextToken() // consume )
							}
							// ScriptDom spans the option from the PARTITION keyword
							// through the closing paren.
							p.spanFrom(optNameTok, partOpt)
							stmt.Options = append(stmt.Options, partOpt)
						}
					} else {
						// Skip unknown option value
						if p.curTok.Type == TokenEquals {
							p.nextToken()
						}
						p.nextToken()
					}

					if p.curTok.Type == TokenComma {
						p.nextToken()
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}
		} else if p.curTok.Type == TokenAs {
			// Parse AS NODE, AS EDGE, or AS SELECT (CTAS)
			p.nextToken() // consume AS
			nodeOrEdge := strings.ToUpper(p.curTok.Literal)
			if nodeOrEdge == "NODE" {
				stmt.AsNode = true
				p.nextToken()
			} else if nodeOrEdge == "EDGE" {
				stmt.AsEdge = true
				p.nextToken()
			} else if p.curTok.Type == TokenSelect {
				// CTAS: CREATE TABLE ... AS SELECT
				selectStmt, err := p.parseSelectStatement()
				if err != nil {
					return nil, err
				}
				// The nested SELECT excludes the trailing semicolon, which
				// belongs to the CREATE TABLE statement.
				p.trimTrailingSemicolon(selectStmt)
				stmt.SelectStatement = selectStmt
			}
		} else if upperLit == "FEDERATED" {
			p.nextToken() // consume FEDERATED
			// Expect ON
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
			}
			// Expect (
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
			}
			// Parse distribution_name = column_name
			distributionName := p.parseIdentifier()
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			columnName := p.parseIdentifier()
			stmt.FederationScheme = &ast.FederationScheme{
				DistributionName: distributionName,
				ColumnName:       columnName,
			}
			// Expect )
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		} else {
			break
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseCreateTableOptions parses table options (ON, TEXTIMAGE_ON, FILESTREAM_ON, WITH) for tables without column definitions (like AS FILETABLE)
func (p *Parser) parseCreateTableOptions(stmt *ast.CreateTableStatement) (*ast.CreateTableStatement, error) {
	astStart := p.curTok

	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if p.curTok.Type == TokenOn {
			p.nextToken() // consume ON
			// Parse filegroup or partition scheme with optional columns
			fg, err := p.parseFileGroupOrPartitionScheme()
			if err != nil {
				return nil, err
			}
			stmt.OnFileGroupOrPartitionScheme = fg
		} else if upperLit == "TEXTIMAGE_ON" {
			p.nextToken() // consume TEXTIMAGE_ON
			// Parse filegroup identifier or string literal
			if p.curTok.Type == TokenString {
				value := p.curTok.Literal
				// Strip quotes from string literal
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				stmt.TextImageOn = &ast.IdentifierOrValueExpression{
					Value:           value,
					ValueExpression: p.strLit(value, false),
				}
				p.nextToken()
			} else {
				ident := p.parseIdentifier()
				stmt.TextImageOn = &ast.IdentifierOrValueExpression{
					Value:      ident.Value,
					Identifier: ident,
				}
			}
		} else if upperLit == "FILESTREAM_ON" {
			p.nextToken() // consume FILESTREAM_ON
			// Parse filegroup identifier or string literal
			if p.curTok.Type == TokenString {
				value := p.curTok.Literal
				// Strip quotes from string literal
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				stmt.FileStreamOn = &ast.IdentifierOrValueExpression{
					Value:           value,
					ValueExpression: p.strLit(value, false),
				}
				p.nextToken()
			} else {
				ident := p.parseIdentifier()
				stmt.FileStreamOn = &ast.IdentifierOrValueExpression{
					Value:      ident.Value,
					Identifier: ident,
				}
			}
		} else if p.curTok.Type == TokenWith {
			// Parse WITH clause with table options
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Parse table options
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optNameTok := p.curTok
					_ = optNameTok
					optionName := strings.ToUpper(p.curTok.Literal)
					p.nextToken() // consume option name

					if optionName == "DATA_COMPRESSION" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						opt, err := p.parseDataCompressionOption()
						if err != nil {
							break
						}
						stmt.Options = append(stmt.Options, &ast.TableDataCompressionOption{
							DataCompressionOption: opt,
							OptionKind:            "DataCompression",
						})
					} else if optionName == "FILETABLE_DIRECTORY" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						// Parse the directory name as a literal or NULL
						opt := &ast.FileTableDirectoryTableOption{
							OptionKind: "FileTableDirectory",
						}
						if strings.ToUpper(p.curTok.Literal) == "NULL" {
							nl := &ast.NullLiteral{
								LiteralType: "Null",
								Value:       "NULL",
							}
							p.tokSpan(nl, p.curTok)
							opt.Value = nl
							p.nextToken()
						} else if p.curTok.Type == TokenString {
							value := p.curTok.Literal
							if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
								value = value[1 : len(value)-1]
							}
							opt.Value = p.strLit(value, false)
							p.nextToken()
						} else {
							value := p.curTok.Literal
							opt.Value = p.strLit(value, false)
							p.nextToken()
						}
						// ScriptDom positions this option on its value.
						p.spanFromChild(opt, opt.Value)
						stmt.Options = append(stmt.Options, opt)
					} else if optionName == "FILETABLE_COLLATE_FILENAME" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						// Parse the collation name as an identifier
						collationName := p.parseIdentifier()
						stmt.Options = append(stmt.Options, &ast.FileTableCollateFileNameTableOption{
							OptionKind: "FileTableCollateFileName",
							Value:      collationName,
						})
					} else if optionName == "MEMORY_OPTIMIZED" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						stateTok := p.curTok
						stateUpper := strings.ToUpper(p.curTok.Literal)
						state := "On"
						if stateUpper == "OFF" {
							state = "Off"
						}
						p.nextToken() // consume ON/OFF
						mo := &ast.MemoryOptimizedTableOption{
							OptionKind:  "MemoryOptimized",
							OptionState: state,
						}
						p.tokSpan(mo, stateTok)
						stmt.Options = append(stmt.Options, mo)
					} else if optionName == "DURABILITY" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						valueUpper := strings.ToUpper(p.curTok.Literal)
						durabilityKind := "SchemaOnly"
						if valueUpper == "SCHEMA_AND_DATA" {
							durabilityKind = "SchemaAndData"
						}
						p.nextToken() // consume value
						stmt.Options = append(stmt.Options, &ast.DurabilityTableOption{
							OptionKind:                "Durability",
							DurabilityTableOptionKind: durabilityKind,
						})
					} else if optionName == "FILETABLE_PRIMARY_KEY_CONSTRAINT_NAME" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						constraintName := p.parseIdentifier()
						stmt.Options = append(stmt.Options, &ast.FileTableConstraintNameTableOption{
							OptionKind: "FileTablePrimaryKeyConstraintName",
							Value:      constraintName,
						})
					} else if optionName == "FILETABLE_STREAMID_UNIQUE_CONSTRAINT_NAME" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						constraintName := p.parseIdentifier()
						stmt.Options = append(stmt.Options, &ast.FileTableConstraintNameTableOption{
							OptionKind: "FileTableStreamIdUniqueConstraintName",
							Value:      constraintName,
						})
					} else if optionName == "FILETABLE_FULLPATH_UNIQUE_CONSTRAINT_NAME" {
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						constraintName := p.parseIdentifier()
						stmt.Options = append(stmt.Options, &ast.FileTableConstraintNameTableOption{
							OptionKind: "FileTableFullPathUniqueConstraintName",
							Value:      constraintName,
						})
					} else if optionName == "REMOTE_DATA_ARCHIVE" {
						opt, err := p.parseRemoteDataArchiveTableOption(false)
						if err != nil {
							return nil, err
						}
						stmt.Options = append(stmt.Options, opt)
					} else if optionName == "CLUSTERED" {
						// Could be CLUSTERED INDEX or CLUSTERED COLUMNSTORE INDEX
						if strings.ToUpper(p.curTok.Literal) == "COLUMNSTORE" {
							p.nextToken() // consume COLUMNSTORE
							if p.curTok.Type == TokenIndex {
								p.nextToken() // consume INDEX
							}
							indexType := &ast.TableClusteredIndexType{
								ColumnStore: true,
							}
							// Check for ORDER(columns)
							if strings.ToUpper(p.curTok.Literal) == "ORDER" {
								p.nextToken() // consume ORDER
								if p.curTok.Type == TokenLParen {
									p.nextToken() // consume (
									firstColTok := p.curTok
									for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
										col := &ast.ColumnReferenceExpression{
											ColumnType: "Regular",
											MultiPartIdentifier: &ast.MultiPartIdentifier{
												Identifiers: []*ast.Identifier{p.parseIdentifier()},
												Count:       1,
											},
										}
										indexType.OrderedColumns = append(indexType.OrderedColumns, col)
										if p.curTok.Type == TokenComma {
											p.nextToken()
										} else {
											break
										}
									}
									if p.curTok.Type == TokenRParen {
										p.nextToken()
									}
									// ScriptDom spans the index type from the first
									// ordered column through the closing paren.
									p.spanFrom(firstColTok, indexType)
								}
							}
							stmt.Options = append(stmt.Options, &ast.TableIndexOption{
								Value:      indexType,
								OptionKind: "LockEscalation",
							})
						} else if p.curTok.Type == TokenIndex {
							p.nextToken() // consume INDEX
							// Parse column list
							indexType := &ast.TableClusteredIndexType{
								ColumnStore: false,
							}
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								firstColTok := p.curTok
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									colTok := p.curTok
									col := &ast.ColumnWithSortOrder{
										SortOrder: ast.SortOrderNotSpecified,
										Column: &ast.ColumnReferenceExpression{
											ColumnType: "Regular",
											MultiPartIdentifier: &ast.MultiPartIdentifier{
												Identifiers: []*ast.Identifier{p.parseIdentifier()},
												Count:       1,
											},
										},
									}
									// Parse optional ASC/DESC
									sortUpper := strings.ToUpper(p.curTok.Literal)
									if sortUpper == "ASC" {
										col.SortOrder = ast.SortOrderAscending
										p.nextToken()
									} else if sortUpper == "DESC" {
										col.SortOrder = ast.SortOrderDescending
										p.nextToken()
									}
									p.spanFrom(colTok, col)
									indexType.Columns = append(indexType.Columns, col)
									if p.curTok.Type == TokenComma {
										p.nextToken()
									} else {
										break
									}
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken()
								}
								// ScriptDom spans the index type from the first
								// column through the closing paren.
								p.spanFrom(firstColTok, indexType)
							}
							stmt.Options = append(stmt.Options, &ast.TableIndexOption{
								Value:      indexType,
								OptionKind: "LockEscalation",
							})
						}
					} else if optionName == "HEAP" {
						heapType := &ast.TableNonClusteredIndexType{}
						p.tokSpan(heapType, optNameTok)
						stmt.Options = append(stmt.Options, &ast.TableIndexOption{
							Value:      heapType,
							OptionKind: "LockEscalation",
						})
					} else if optionName == "DISTRIBUTION" {
						// Parse DISTRIBUTION = HASH(col1, col2, ...) or ROUND_ROBIN or REPLICATE
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						distTypeTok := p.curTok
						distTypeUpper := strings.ToUpper(p.curTok.Literal)
						if distTypeUpper == "HASH" {
							p.nextToken() // consume HASH
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								hashPolicy := &ast.TableHashDistributionPolicy{}
								// Parse column list
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									col := p.parseIdentifier()
									if hashPolicy.DistributionColumn == nil {
										hashPolicy.DistributionColumn = col
									}
									hashPolicy.DistributionColumns = append(hashPolicy.DistributionColumns, col)
									if p.curTok.Type == TokenComma {
										p.nextToken()
									} else {
										break
									}
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken()
								}
								p.spanFrom(distTypeTok, hashPolicy)
								distOpt := &ast.TableDistributionOption{
									OptionKind: "Distribution",
									Value:      hashPolicy,
								}
								// ScriptDom positions the option on the DISTRIBUTION keyword.
								p.tokSpan(distOpt, optNameTok)
								stmt.Options = append(stmt.Options, distOpt)
							}
						} else if distTypeUpper == "ROUND_ROBIN" {
							p.nextToken() // consume ROUND_ROBIN
							rrPolicy := &ast.TableRoundRobinDistributionPolicy{}
							p.tokSpan(rrPolicy, distTypeTok)
							distOpt := &ast.TableDistributionOption{
								OptionKind: "Distribution",
								Value:      rrPolicy,
							}
							p.tokSpan(distOpt, optNameTok)
							stmt.Options = append(stmt.Options, distOpt)
						} else if distTypeUpper == "REPLICATE" {
							p.nextToken() // consume REPLICATE
							repPolicy := &ast.TableReplicateDistributionPolicy{}
							p.tokSpan(repPolicy, distTypeTok)
							distOpt := &ast.TableDistributionOption{
								OptionKind: "Distribution",
								Value:      repPolicy,
							}
							p.tokSpan(distOpt, optNameTok)
							stmt.Options = append(stmt.Options, distOpt)
						} else {
							// Unknown distribution - skip for now
							p.nextToken()
						}
					} else {
						// Skip unknown option value
						if p.curTok.Type == TokenEquals {
							p.nextToken()
						}
						p.nextToken()
					}

					if p.curTok.Type == TokenComma {
						p.nextToken()
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}
		} else if p.curTok.Type == TokenAs {
			// Parse AS NODE, AS EDGE, or AS SELECT (CTAS)
			p.nextToken() // consume AS
			nodeOrEdge := strings.ToUpper(p.curTok.Literal)
			if nodeOrEdge == "NODE" {
				stmt.AsNode = true
				p.nextToken()
			} else if nodeOrEdge == "EDGE" {
				stmt.AsEdge = true
				p.nextToken()
			} else if p.curTok.Type == TokenSelect {
				// CTAS: CREATE TABLE ... AS SELECT
				selectStmt, err := p.parseSelectStatement()
				if err != nil {
					return nil, err
				}
				// The nested SELECT excludes the trailing semicolon, which
				// belongs to the CREATE TABLE statement.
				p.trimTrailingSemicolon(selectStmt)
				stmt.SelectStatement = selectStmt
			}
		} else {
			break
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseRemoteDataArchiveTableOption parses REMOTE_DATA_ARCHIVE = ON/OFF (options...) for tables
// isAlterTable indicates if this is for ALTER TABLE SET (which uses RemoteDataArchiveAlterTableOption)
func (p *Parser) parseRemoteDataArchiveTableOption(isAlterTable bool) (ast.TableOption, error) {
	astStart := p.curTok

	// curTok should be = or (
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
	}

	// Parse ON, OFF, or OFF_WITHOUT_DATA_RECOVERY
	rdaStateTok := p.curTok
	rdaOption := "Enable"
	stateUpper := strings.ToUpper(p.curTok.Literal)
	if stateUpper == "ON" {
		rdaOption = "Enable"
		p.nextToken()
	} else if stateUpper == "OFF" {
		rdaOption = "Disable"
		p.nextToken()
	} else if stateUpper == "OFF_WITHOUT_DATA_RECOVERY" {
		rdaOption = "OffWithoutDataRecovery"
		p.nextToken()
	}

	var migrationState string
	var filterPredicate ast.ScalarExpression
	isMigrationStateSpecified := false
	isFilterPredicateSpecified := false

	// Parse options in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			switch optName {
			case "MIGRATION_STATE":
				isMigrationStateSpecified = true
				msUpper := strings.ToUpper(p.curTok.Literal)
				if msUpper == "PAUSED" {
					migrationState = "Paused"
				} else if msUpper == "OUTBOUND" {
					migrationState = "Outbound"
				} else if msUpper == "INBOUND" {
					migrationState = "Inbound"
				}
				p.nextToken()
			case "FILTER_PREDICATE":
				isFilterPredicateSpecified = true
				if strings.ToUpper(p.curTok.Literal) == "NULL" {
					// When FILTER_PREDICATE = NULL, filterPredicate stays nil
					p.nextToken()
				} else {
					// Parse function call like dbo.f1(c1)
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					filterPredicate = expr
				}
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}

		var rparenTok Token
		hasRParen := false
		if p.curTok.Type == TokenRParen {
			rparenTok = p.curTok
			hasRParen = true
			p.nextToken() // consume )
		}
		if isAlterTable {
			// ALTER TABLE SET form: spans from the ON/OFF state through the
			// closing paren.
			alterOpt := &ast.RemoteDataArchiveAlterTableOption{
				RdaTableOption:             rdaOption,
				MigrationState:             migrationState,
				IsMigrationStateSpecified:  isMigrationStateSpecified,
				FilterPredicate:            filterPredicate,
				IsFilterPredicateSpecified: isFilterPredicateSpecified,
				OptionKind:                 "RemoteDataArchive",
			}
			p.spanFrom(rdaStateTok, alterOpt)
			alterOpt.Frag().Pin()
			return alterOpt, nil
		}
		// CREATE TABLE form: ScriptDom positions the option on the closing
		// paren alone.
		createOpt := &ast.RemoteDataArchiveTableOption{
			RdaTableOption:  rdaOption,
			MigrationState:  migrationState,
			FilterPredicate: filterPredicate,
			OptionKind:      "RemoteDataArchive",
		}
		if hasRParen {
			p.tokSpan(createOpt, rparenTok)
			createOpt.Frag().Pin()
		}
		return createOpt, nil
	}

	if isAlterTable {
		return spanned(p, &ast.RemoteDataArchiveAlterTableOption{
			RdaTableOption:             rdaOption,
			MigrationState:             migrationState,
			IsMigrationStateSpecified:  isMigrationStateSpecified,
			FilterPredicate:            filterPredicate,
			IsFilterPredicateSpecified: isFilterPredicateSpecified,
			OptionKind:                 "RemoteDataArchive",
		}, astStart), nil
	}

	return spanned(p, &ast.RemoteDataArchiveTableOption{
		RdaTableOption:  rdaOption,
		MigrationState:  migrationState,
		FilterPredicate: filterPredicate,
		OptionKind:      "RemoteDataArchive",
	}, astStart), nil
}

// parseMergeStatement parses a MERGE statement
func (p *Parser) parseMergeStatement() (*ast.MergeStatement, error) {
	astStart := p.curTok

	// Consume MERGE
	p.nextToken()

	stmt := &ast.MergeStatement{
		MergeSpecification: &ast.MergeSpecification{},
	}

	// Check for TOP clause
	if p.curTok.Type == TokenTop {
		top, err := p.parseTopRowFilter()
		if err != nil {
			return nil, err
		}
		stmt.MergeSpecification.TopRowFilter = top
	}

	// Optional INTO keyword
	if p.curTok.Type == TokenInto {
		p.nextToken()
	}

	// Parse target table
	target, err := p.parseSingleTableReference()
	if err != nil {
		return nil, err
	}
	// If target has an alias, move it to TableAlias (ScriptDOM convention)
	if ntr, ok := target.(*ast.NamedTableReference); ok && ntr.Alias != nil {
		stmt.MergeSpecification.TableAlias = ntr.Alias
		ntr.Alias = nil
		// The target's span shrinks to the object name alone.
		if ntr.SchemaObject != nil && ntr.SchemaObject.Frag().HasSpan() {
			*ntr.Frag() = *ntr.SchemaObject.Frag()
		}
	}
	stmt.MergeSpecification.Target = target

	// Expect USING
	if strings.ToUpper(p.curTok.Literal) == "USING" {
		p.nextToken()
	}

	// Parse source table reference (may be parenthesized join or subquery)
	sourceRef, err := p.parseMergeSourceTableReference()
	if err != nil {
		return nil, err
	}
	stmt.MergeSpecification.TableReference = sourceRef

	// Expect ON
	if p.curTok.Type == TokenOn {
		p.nextToken()
	}

	// Parse ON condition - check for MATCH predicate
	if strings.ToUpper(p.curTok.Literal) == "MATCH" {
		matchPred, err := p.parseGraphMatchPredicate()
		if err != nil {
			return nil, err
		}
		stmt.MergeSpecification.SearchCondition = matchPred
	} else {
		cond, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		stmt.MergeSpecification.SearchCondition = cond
	}

	// Parse WHEN clauses
	for strings.ToUpper(p.curTok.Literal) == "WHEN" {
		clause, err := p.parseMergeActionClause()
		if err != nil {
			return nil, err
		}
		stmt.MergeSpecification.ActionClauses = append(stmt.MergeSpecification.ActionClauses, clause)
	}

	// Parse optional OUTPUT clause
	if strings.ToUpper(p.curTok.Literal) == "OUTPUT" {
		output, _, err := p.parseOutputClause()
		if err != nil {
			return nil, err
		}
		stmt.MergeSpecification.OutputClause = output
	}

	// The merge specification spans from MERGE through the OUTPUT clause,
	// excluding any OPTION clause and trailing semicolon.
	p.spanFrom(astStart, stmt.MergeSpecification)
	p.trimTrailingSemicolon(stmt.MergeSpecification)

	// Parse optional OPTION clause
	if strings.ToUpper(p.curTok.Literal) == "OPTION" {
		hints, err := p.parseOptionClause()
		if err != nil {
			return nil, err
		}
		stmt.OptimizerHints = hints
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseMergeSpecification parses a MERGE specification (used in DataModificationTableReference)
func (p *Parser) parseMergeSpecification() (*ast.MergeSpecification, error) {
	astStart := p.curTok

	// Consume MERGE
	p.nextToken()

	spec := &ast.MergeSpecification{}

	// Optional INTO keyword
	if strings.ToUpper(p.curTok.Literal) == "INTO" {
		p.nextToken()
	}

	// Parse target table
	target, err := p.parseSingleTableReference()
	if err != nil {
		return nil, err
	}
	// If target has an alias, move it to TableAlias (ScriptDOM convention)
	if ntr, ok := target.(*ast.NamedTableReference); ok && ntr.Alias != nil {
		spec.TableAlias = ntr.Alias
		ntr.Alias = nil
		// The target's span shrinks to the object name alone.
		if ntr.SchemaObject != nil && ntr.SchemaObject.Frag().HasSpan() {
			*ntr.Frag() = *ntr.SchemaObject.Frag()
		}
	}
	spec.Target = target

	// Expect USING
	if strings.ToUpper(p.curTok.Literal) == "USING" {
		p.nextToken()
	}

	// Parse source table reference (may be parenthesized join or subquery)
	sourceRef, err := p.parseMergeSourceTableReference()
	if err != nil {
		return nil, err
	}
	spec.TableReference = sourceRef

	// Expect ON
	if p.curTok.Type == TokenOn {
		p.nextToken()
	}

	// Parse ON condition - check for MATCH predicate
	if strings.ToUpper(p.curTok.Literal) == "MATCH" {
		matchPred, err := p.parseGraphMatchPredicate()
		if err != nil {
			return nil, err
		}
		spec.SearchCondition = matchPred
	} else {
		cond, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		spec.SearchCondition = cond
	}

	// Parse WHEN clauses
	for strings.ToUpper(p.curTok.Literal) == "WHEN" {
		clause, err := p.parseMergeActionClause()
		if err != nil {
			return nil, err
		}
		spec.ActionClauses = append(spec.ActionClauses, clause)
	}

	// Parse optional OUTPUT clause
	if strings.ToUpper(p.curTok.Literal) == "OUTPUT" {
		output, _, err := p.parseOutputClause()
		if err != nil {
			return nil, err
		}
		spec.OutputClause = output
	}

	return spanned(p, spec, astStart), nil
}

// parseMergeSourceTableReference parses the source table reference in a MERGE statement
func (p *Parser) parseMergeSourceTableReference() (ast.TableReference, error) {
	astStart := p.curTok

	// Check for parenthesized expression
	if p.curTok.Type == TokenLParen {
		// Check if this is a derived table (subquery) or a join
		if p.peekTok.Type == TokenSelect {
			// This is a derived table like (SELECT ...) AS alias
			spanV2, spanErr2 := p.parseDerivedTableReference()
			return spanned(p, spanV2, astStart), spanErr2
		}
		p.nextToken() // consume (
		// Parse the inner join expression
		inner, err := p.parseMergeJoinTableReference()
		if err != nil {
			return nil, err
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
		return spanned(p, &ast.JoinParenthesisTableReference{Join: inner}, astStart), nil
	}
	spanV3, spanErr3 := p.parseSingleTableReference()
	return spanned(p, spanV3, astStart), spanErr3
}

// parseMergeJoinTableReference parses a table reference which may include joins
func (p *Parser) parseMergeJoinTableReference() (ast.TableReference, error) {
	astStart := p.curTok

	left, err := p.parseSingleTableReference()
	if err != nil {
		return nil, err
	}

	// Check for JOIN
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "JOIN" || upperLit == "INNER" || upperLit == "LEFT" || upperLit == "RIGHT" || upperLit == "FULL" || upperLit == "CROSS" {
			join := &ast.QualifiedJoin{
				FirstTableReference: left,
				JoinHint:            "None",
			}

			// Parse join type
			switch upperLit {
			case "INNER", "JOIN":
				join.QualifiedJoinType = "Inner"
				if upperLit == "INNER" {
					p.nextToken() // consume INNER
				}
			case "LEFT":
				join.QualifiedJoinType = "LeftOuter"
				p.nextToken() // consume LEFT
				if strings.ToUpper(p.curTok.Literal) == "OUTER" {
					p.nextToken() // consume OUTER
				}
			case "RIGHT":
				join.QualifiedJoinType = "RightOuter"
				p.nextToken() // consume RIGHT
				if strings.ToUpper(p.curTok.Literal) == "OUTER" {
					p.nextToken() // consume OUTER
				}
			case "FULL":
				join.QualifiedJoinType = "FullOuter"
				p.nextToken() // consume FULL
				if strings.ToUpper(p.curTok.Literal) == "OUTER" {
					p.nextToken() // consume OUTER
				}
			case "CROSS":
				join.QualifiedJoinType = "CrossJoin"
				p.nextToken() // consume CROSS
			}

			// Consume JOIN keyword if present
			if strings.ToUpper(p.curTok.Literal) == "JOIN" {
				p.nextToken()
			}

			// Parse the right side of the join
			right, err := p.parseSingleTableReference()
			if err != nil {
				return nil, err
			}
			join.SecondTableReference = right

			// Parse ON condition
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				cond, err := p.parseBooleanExpression()
				if err != nil {
					return nil, err
				}
				join.SearchCondition = cond
			}

			left = join
		} else {
			break
		}
	}

	return spanned(p, left, astStart), nil
}

// parseGraphMatchPredicate parses MATCH (node-edge->node) graph pattern
func (p *Parser) parseGraphMatchPredicate() (*ast.GraphMatchPredicate, error) {
	astStart := p.curTok

	// Consume MATCH
	p.nextToken()

	pred := &ast.GraphMatchPredicate{}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after MATCH, got %s", p.curTok.Literal)
	}
	p.nextToken()
	// ScriptDom spans the predicate over the pattern expression and closing
	// paren, excluding the MATCH( prefix.
	patternTok := p.curTok

	// Parse the graph pattern expression (may be multiple composites joined by AND)
	expr, err := p.parseGraphMatchAndExpression()
	if err != nil {
		return nil, err
	}
	pred.Expression = expr

	// Expect )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	p.spanFrom(patternTok, pred)
	pred.Frag().Pin()
	_ = astStart
	return pred, nil
}

// parseGraphMatchAndExpression parses graph match expressions connected by AND
// Note: AND inside chains is now handled by parseGraphMatchChainedExpression
func (p *Parser) parseGraphMatchAndExpression() (ast.GraphMatchExpression, error) {
	astStart := p.curTok

	spanV4, spanErr4 := p.parseGraphMatchChainedExpression()
	return spanned(p, spanV4, astStart), spanErr4
}

// parseGraphMatchChainedExpression parses a chain like A-(B)->C-(D)->E
// Also handles AND which continues the chain but starts a fresh node
// Also handles SHORTEST_PATH and LAST_NODE functions
func (p *Parser) parseGraphMatchChainedExpression() (ast.GraphMatchExpression, error) {
	astStart := p.curTok

	// Check for SHORTEST_PATH or LAST_NODE at the start
	var first ast.GraphMatchExpression
	var rightNode *ast.GraphMatchNodeExpression
	var err error

	if strings.ToUpper(p.curTok.Literal) == "SHORTEST_PATH" {
		first, err = p.parseGraphMatchShortestPath()
		if err != nil {
			return nil, err
		}
		rightNode = nil
	} else if strings.ToUpper(p.curTok.Literal) == "LAST_NODE" {
		// Check if this is LAST_NODE(x) = LAST_NODE(y) comparison
		var leftNode *ast.GraphMatchNodeExpression
		first, leftNode, err = p.parseGraphMatchLastNodeComparison()
		if err != nil {
			return nil, err
		}
		if first == nil {
			// Not a comparison - LAST_NODE is part of a pattern
			// Use the parsed node as the left node of a composite
			first, rightNode, err = p.parseGraphMatchSingleComposite(leftNode)
			if err != nil {
				return nil, err
			}
		} else {
			rightNode = nil
		}
	} else {
		// Parse first composite pattern
		first, rightNode, err = p.parseGraphMatchSingleComposite(nil)
		if err != nil {
			return nil, err
		}
	}

	var result ast.GraphMatchExpression = first

	// Check for continuation - if the right node is followed by -, <, or AND, it's a chain
	for p.curTok.Type == TokenMinus || p.curTok.Type == TokenLessThan || p.curTok.Type == TokenAnd {
		// If AND, continue chain but start with fresh node (not chaining from previous)
		if p.curTok.Type == TokenAnd {
			p.nextToken() // consume AND
			rightNode = nil

			// Check for SHORTEST_PATH or LAST_NODE after AND
			if strings.ToUpper(p.curTok.Literal) == "SHORTEST_PATH" {
				next, err := p.parseGraphMatchShortestPath()
				if err != nil {
					return nil, err
				}
				result = &ast.BooleanBinaryExpression{
					BinaryExpressionType: "And",
					FirstExpression:      result.(ast.BooleanExpression),
					SecondExpression:     next,
				}
				continue
			} else if strings.ToUpper(p.curTok.Literal) == "LAST_NODE" {
				next, leftNode, err := p.parseGraphMatchLastNodeComparison()
				if err != nil {
					return nil, err
				}
				if next == nil {
					// LAST_NODE is part of a pattern, parse composite using leftNode
					next, rightNode, err = p.parseGraphMatchSingleComposite(leftNode)
					if err != nil {
						return nil, err
					}
				}
				result = &ast.BooleanBinaryExpression{
					BinaryExpressionType: "And",
					FirstExpression:      result.(ast.BooleanExpression),
					SecondExpression:     next.(ast.BooleanExpression),
				}
				continue
			}
		}

		// The previous right node becomes the left node of the next composite (nil if after AND)
		next, nextRightNode, err := p.parseGraphMatchSingleComposite(rightNode)
		if err != nil {
			return nil, err
		}

		// Wrap in BooleanBinaryExpression with And
		result = &ast.BooleanBinaryExpression{
			BinaryExpressionType: "And",
			FirstExpression:      result.(ast.BooleanExpression),
			SecondExpression:     next,
		}
		rightNode = nextRightNode
	}

	return spanned(p, result, astStart), nil
}

// parseGraphMatchShortestPath parses SHORTEST_PATH(pattern+) or SHORTEST_PATH(pattern{min,max})
func (p *Parser) parseGraphMatchShortestPath() (*ast.GraphMatchRecursivePredicate, error) {
	astStart := p.curTok

	pred := &ast.GraphMatchRecursivePredicate{
		Function: "ShortestPath",
	}

	p.nextToken() // consume SHORTEST_PATH

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after SHORTEST_PATH, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Determine if anchor is on left or right
	// Pattern: N (-(E)->N2)+ means anchor N is on left
	// Pattern: (N-(E)->)+N2 means anchor N2 is on right
	// Check if we have ( immediately or an identifier first

	if p.curTok.Type == TokenLParen {
		// Anchor on right: (pattern)+ N2
		pred.AnchorOnLeft = false
		p.nextToken() // consume inner (

		// Parse the recursive pattern(s)
		pred.Expression = []*ast.GraphMatchCompositeExpression{}
		var prevRightNode *ast.GraphMatchNodeExpression
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			comp, rightNode, err := p.parseGraphMatchSingleComposite(prevRightNode)
			if err != nil {
				return nil, err
			}
			// Store the right node for chaining to the next composite's left node
			prevRightNode = rightNode
			// For right anchor, the composite doesn't have a right node set explicitly
			// Clear it as it's implicit (next iteration's left node or the terminal OuterNodeExpression)
			comp.RightNode = nil
			pred.Expression = append(pred.Expression, comp)

			// Check for continuation within the recursive pattern
			if p.curTok.Type != TokenMinus && p.curTok.Type != TokenLessThan {
				break
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}

		// Parse quantifier: + or {min,max}
		pred.RecursiveQuantifier = &ast.GraphRecursiveMatchQuantifier{}
		if p.curTok.Type == TokenPlus {
			pred.RecursiveQuantifier.IsPlusSign = true
			p.nextToken() // consume +
		} else if p.curTok.Type == TokenLBrace {
			pred.RecursiveQuantifier.IsPlusSign = false
			p.nextToken() // consume {
			pred.RecursiveQuantifier.LowerLimit, _ = p.parseScalarExpression()
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
				pred.RecursiveQuantifier.UpperLimit, _ = p.parseScalarExpression()
			}
			if p.curTok.Type == TokenRBrace {
				p.nextToken() // consume }
			}
		}

		// Parse outer node (anchor on right)
		pred.OuterNodeExpression = p.parseGraphMatchNodeExpr()
	} else {
		// Anchor on left: N (pattern)+
		pred.AnchorOnLeft = true

		// Parse outer node (anchor)
		pred.OuterNodeExpression = p.parseGraphMatchNodeExpr()

		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after anchor node in SHORTEST_PATH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume inner (

		// Parse the recursive pattern(s) - left node is nil (implied from anchor)
		pred.Expression = []*ast.GraphMatchCompositeExpression{}
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			comp, _, err := p.parseGraphMatchSingleComposite(nil)
			if err != nil {
				return nil, err
			}
			// For left anchor, the first composite doesn't have a left node set explicitly
			// Clear it as it's implied from the anchor
			comp.LeftNode = nil
			pred.Expression = append(pred.Expression, comp)

			// Check for continuation within the recursive pattern
			if p.curTok.Type != TokenMinus && p.curTok.Type != TokenLessThan {
				break
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}

		// Parse quantifier: + or {min,max}
		pred.RecursiveQuantifier = &ast.GraphRecursiveMatchQuantifier{}
		if p.curTok.Type == TokenPlus {
			pred.RecursiveQuantifier.IsPlusSign = true
			p.nextToken() // consume +
		} else if p.curTok.Type == TokenLBrace {
			pred.RecursiveQuantifier.IsPlusSign = false
			p.nextToken() // consume {
			pred.RecursiveQuantifier.LowerLimit, _ = p.parseScalarExpression()
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
				pred.RecursiveQuantifier.UpperLimit, _ = p.parseScalarExpression()
			}
			if p.curTok.Type == TokenRBrace {
				p.nextToken() // consume }
			}
		}
	}

	// Consume closing ) of SHORTEST_PATH
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return spanned(p, pred, astStart), nil
}

// parseGraphMatchNodeExpr parses a node expression which may be LAST_NODE(x) or just x
func (p *Parser) parseGraphMatchNodeExpr() *ast.GraphMatchNodeExpression {
	astStart := p.curTok

	node := &ast.GraphMatchNodeExpression{}

	if strings.ToUpper(p.curTok.Literal) == "LAST_NODE" {
		node.UsesLastNode = true
		p.nextToken() // consume LAST_NODE
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			node.Node = p.parseIdentifier()
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	} else {
		node.Node = p.parseIdentifier()
	}

	return spanned(p, node, astStart)
}

// parseGraphMatchLastNodeComparison parses LAST_NODE(x) = LAST_NODE(y) predicate
// Returns (predicate, nil, nil) if it is a comparison
// Returns (nil, leftNode, nil) if LAST_NODE is not followed by = (the leftNode can be used for composite)
func (p *Parser) parseGraphMatchLastNodeComparison() (ast.GraphMatchExpression, *ast.GraphMatchNodeExpression, error) {
	// Parse LAST_NODE(x)
	left := p.parseGraphMatchNodeExpr()

	// Check for = comparison
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		right := p.parseGraphMatchNodeExpr()

		// Return a predicate expression
		return &ast.GraphMatchLastNodePredicate{
			LeftExpression:  left,
			RightExpression: right,
		}, nil, nil
	}

	// Not a comparison - this LAST_NODE is part of a pattern like LAST_NODE(N) - (E) -> N2
	// Return the parsed node so it can be used as the left node of a composite
	return nil, left, nil
}

// parseGraphMatchSingleComposite parses a single Node-(Edge)->Node pattern
// leftNode is provided when chaining (the previous right node becomes the left node)
// Returns the composite and the right node (for potential chaining)
func (p *Parser) parseGraphMatchSingleComposite(leftNode *ast.GraphMatchNodeExpression) (*ast.GraphMatchCompositeExpression, *ast.GraphMatchNodeExpression, error) {
	composite := &ast.GraphMatchCompositeExpression{}

	// Check if pattern starts with arrow (no explicit left node)
	// This happens in recursive patterns like -(E)->N inside SHORTEST_PATH
	startsWithArrow := p.curTok.Type == TokenLessThan || p.curTok.Type == TokenMinus
	arrowOnRight := true

	if startsWithArrow {
		// Pattern starts with arrow - left node is implicit
		if p.curTok.Type == TokenLessThan {
			arrowOnRight = false
			p.nextToken() // consume <
			if p.curTok.Type == TokenMinus {
				p.nextToken() // consume -
			}
		} else if p.curTok.Type == TokenMinus {
			p.nextToken() // consume -
		}
		// Use provided leftNode if any, otherwise leave it nil
		composite.LeftNode = leftNode
	} else {
		// Pattern starts with node identifier
		if leftNode != nil {
			composite.LeftNode = leftNode
		} else {
			ln := &ast.GraphMatchNodeExpression{
				Node: p.parseIdentifier(),
			}
			if ln.Node != nil && ln.Node.Frag().HasSpan() {
				*ln.Frag() = *ln.Node.Frag()
			}
			composite.LeftNode = ln
		}

		// Now check for arrow direction at the start: <- or -
		if p.curTok.Type == TokenLessThan {
			arrowOnRight = false
			p.nextToken() // consume <
			if p.curTok.Type == TokenMinus {
				p.nextToken() // consume -
			}
		} else if p.curTok.Type == TokenMinus {
			p.nextToken() // consume -
		}
	}

	// Parse edge - may be in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		composite.Edge = p.parseIdentifier()
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	} else {
		composite.Edge = p.parseIdentifier()
	}

	// Check for arrow direction at the end: - > or -> means arrow on right
	if p.curTok.Type == TokenMinus {
		p.nextToken() // consume -
		if p.curTok.Type == TokenGreaterThan {
			arrowOnRight = true
			p.nextToken() // consume >
		}
	}
	composite.ArrowOnRight = arrowOnRight

	// Parse right node (only if there's an identifier - in recursive patterns the right node may be implicit)
	var rightNode *ast.GraphMatchNodeExpression
	if p.curTok.Type == TokenIdent || strings.ToUpper(p.curTok.Literal) == "LAST_NODE" {
		rightNodeTok := p.curTok
		rightNode = &ast.GraphMatchNodeExpression{}
		if strings.ToUpper(p.curTok.Literal) == "LAST_NODE" {
			rightNode.UsesLastNode = true
			p.nextToken() // consume LAST_NODE
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				rightNode.Node = p.parseIdentifier()
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			p.spanFrom(rightNodeTok, rightNode)
		} else {
			rightNode.Node = p.parseIdentifier()
			if rightNode.Node != nil && rightNode.Node.Frag().HasSpan() {
				*rightNode.Frag() = *rightNode.Node.Frag()
			}
		}
		composite.RightNode = rightNode
	}

	// The composite spans from its left node through its right node.
	if composite.LeftNode != nil && composite.LeftNode.Frag().HasSpan() {
		p.spanFromChild(composite, composite.LeftNode)
	}

	return composite, rightNode, nil
}

// parseMergeActionClause parses a WHEN clause in a MERGE statement
func (p *Parser) parseMergeActionClause() (*ast.MergeActionClause, error) {
	astStart := p.curTok

	// Consume WHEN
	p.nextToken()

	clause := &ast.MergeActionClause{}

	// Parse condition: MATCHED, NOT MATCHED, NOT MATCHED BY SOURCE, NOT MATCHED BY TARGET
	if strings.ToUpper(p.curTok.Literal) == "MATCHED" {
		clause.Condition = "Matched"
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NOT" {
		p.nextToken() // consume NOT
		if strings.ToUpper(p.curTok.Literal) == "MATCHED" {
			p.nextToken() // consume MATCHED
			if strings.ToUpper(p.curTok.Literal) == "BY" {
				p.nextToken() // consume BY
				byWhat := strings.ToUpper(p.curTok.Literal)
				if byWhat == "SOURCE" {
					clause.Condition = "NotMatchedBySource"
					p.nextToken()
				} else if byWhat == "TARGET" {
					clause.Condition = "NotMatchedByTarget"
					p.nextToken()
				}
			} else {
				clause.Condition = "NotMatched"
			}
		}
	}

	// Optional AND condition
	if strings.ToUpper(p.curTok.Literal) == "AND" {
		p.nextToken()
		cond, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		clause.SearchCondition = cond
	}

	// Expect THEN
	if strings.ToUpper(p.curTok.Literal) == "THEN" {
		p.nextToken()
	}

	// Parse action: DELETE, UPDATE SET, INSERT
	actionTok := p.curTok
	actionWord := strings.ToUpper(p.curTok.Literal)
	if actionWord == "DELETE" {
		p.nextToken()
		delAction := &ast.DeleteMergeAction{}
		p.tokSpan(delAction, actionTok)
		clause.Action = delAction
	} else if actionWord == "UPDATE" {
		p.nextToken() // consume UPDATE
		if strings.ToUpper(p.curTok.Literal) == "SET" {
			p.nextToken() // consume SET
		}
		action := &ast.UpdateMergeAction{}
		// Parse SET clauses
		for {
			setClause, err := p.parseSetClause()
			if err != nil {
				break
			}
			action.SetClauses = append(action.SetClauses, setClause)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		clause.Action = action
	} else if actionWord == "INSERT" {
		p.nextToken() // consume INSERT
		action := &ast.InsertMergeAction{}

		// Check for DEFAULT VALUES first
		if p.curTok.Type == TokenDefault {
			defaultTok := p.curTok
			p.nextToken() // consume DEFAULT
			if strings.ToUpper(p.curTok.Literal) == "VALUES" {
				p.nextToken() // consume VALUES
			}
			vis := &ast.ValuesInsertSource{IsDefaultValues: true}
			p.spanFrom(defaultTok, vis)
			action.Source = vis
			clause.Action = action
		} else {
			// Parse optional column list
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					// Check for pseudo columns $ACTION and $CUID
					if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "$") {
						pseudoCol := strings.ToUpper(p.curTok.Literal)
						if pseudoCol == "$ACTION" {
							pc := &ast.ColumnReferenceExpression{
								ColumnType: "PseudoColumnAction",
							}
							p.tokSpan(pc, p.curTok)
							action.Columns = append(action.Columns, pc)
						} else if pseudoCol == "$CUID" {
							pc := &ast.ColumnReferenceExpression{
								ColumnType: "PseudoColumnCuid",
							}
							p.tokSpan(pc, p.curTok)
							action.Columns = append(action.Columns, pc)
						}
						p.nextToken()
					} else {
						col := &ast.ColumnReferenceExpression{
							ColumnType: "Regular",
							MultiPartIdentifier: &ast.MultiPartIdentifier{
								Identifiers: []*ast.Identifier{p.parseIdentifier()},
								Count:       1,
							},
						}
						action.Columns = append(action.Columns, col)
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
			}
			// Parse VALUES
			if strings.ToUpper(p.curTok.Literal) == "VALUES" {
				valuesTok := p.curTok
				p.nextToken()
				source := &ast.ValuesInsertSource{IsDefaultValues: false}
				if p.curTok.Type == TokenLParen {
					rowTok := p.curTok
					p.nextToken() // consume (
					rowValue := &ast.RowValue{}
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
						val, err := p.parseScalarExpression()
						if err != nil {
							break
						}
						rowValue.ColumnValues = append(rowValue.ColumnValues, val)
						if p.curTok.Type == TokenComma {
							p.nextToken()
						} else {
							break
						}
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken()
					}
					p.spanFrom(rowTok, rowValue)
					source.RowValues = append(source.RowValues, rowValue)
				}
				p.spanFrom(valuesTok, source)
				action.Source = source
			}
			clause.Action = action
		}
	}

	// ScriptDom positions the merge action clause over its AND search
	// condition (when present) and action tokens, excluding WHEN ... THEN.
	if a, ok := clause.Action.(spannable); ok {
		if !a.Frag().HasSpan() {
			p.spanFrom(actionTok, a)
		}
		af := a.Frag()
		if af.HasSpan() {
			clause.SetSpan(af.StartOffset, af.FragmentLength, af.StartLine, af.StartColumn)
			if sc, ok := clause.SearchCondition.(spannable); ok && sc.Frag().HasSpan() {
				scf := sc.Frag()
				clause.SetSpan(scf.StartOffset, af.EndOffset()-scf.StartOffset, scf.StartLine, scf.StartColumn)
			}
		}
	}
	_ = astStart
	return clause, nil
}

func (p *Parser) parseDataCompressionOption() (*ast.DataCompressionOption, error) {
	astStart := p.curTok

	opt := &ast.DataCompressionOption{
		OptionKind: "DataCompression",
	}

	// Parse compression level: NONE, ROW, PAGE, COLUMNSTORE, COLUMNSTORE_ARCHIVE
	levelStr := strings.ToUpper(p.curTok.Literal)
	switch levelStr {
	case "NONE":
		opt.CompressionLevel = "None"
	case "ROW":
		opt.CompressionLevel = "Row"
	case "PAGE":
		opt.CompressionLevel = "Page"
	case "COLUMNSTORE":
		opt.CompressionLevel = "ColumnStore"
	case "COLUMNSTORE_ARCHIVE":
		opt.CompressionLevel = "ColumnStoreArchive"
	default:
		opt.CompressionLevel = levelStr
	}
	p.nextToken()

	// Parse optional ON PARTITIONS clause
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
			p.nextToken() // consume PARTITIONS
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Parse partition ranges
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					pr := &ast.CompressionPartitionRange{}

					// Parse From
					if p.curTok.Type == TokenNumber {
						pr.From = p.intLitFromToken(p.curTok)
						p.nextToken()
					}

					// Check for TO
					if strings.ToUpper(p.curTok.Literal) == "TO" {
						p.nextToken() // consume TO
						if p.curTok.Type == TokenNumber {
							pr.To = p.intLitFromToken(p.curTok)
							p.nextToken()
						}
					}

					opt.PartitionRanges = append(opt.PartitionRanges, pr)

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
	}

	return spanned(p, opt, astStart), nil
}

func (p *Parser) parseXmlCompressionOption() (*ast.XmlCompressionOption, error) {
	astStart := p.curTok

	opt := &ast.XmlCompressionOption{
		OptionKind: "XmlCompression",
	}

	// Parse ON or OFF
	levelStr := strings.ToUpper(p.curTok.Literal)
	if levelStr == "ON" {
		opt.IsCompressed = "On"
	} else {
		opt.IsCompressed = "Off"
	}
	p.nextToken()

	// Parse optional ON PARTITIONS clause
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
			p.nextToken() // consume PARTITIONS
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Parse partition ranges
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					pr := &ast.CompressionPartitionRange{}

					// Parse From
					if p.curTok.Type == TokenNumber {
						pr.From = p.intLitFromToken(p.curTok)
						p.nextToken()
					}

					// Check for TO
					if strings.ToUpper(p.curTok.Literal) == "TO" {
						p.nextToken() // consume TO
						if p.curTok.Type == TokenNumber {
							pr.To = p.intLitFromToken(p.curTok)
							p.nextToken()
						}
					}

					opt.PartitionRanges = append(opt.PartitionRanges, pr)

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
	}

	return spanned(p, opt, astStart), nil
}

func (p *Parser) parseColumnDefinition() (*ast.ColumnDefinition, error) {
	astStart := p.curTok

	col := &ast.ColumnDefinition{}

	// Parse column name (parseIdentifier already calls nextToken)
	col.ColumnIdentifier = p.parseIdentifier()

	// Check for computed column (AS expression)
	if strings.ToUpper(p.curTok.Literal) == "AS" {
		p.nextToken() // consume AS
		// Parse computed column expression
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		col.ComputedColumnExpression = expr
		// Check for PERSISTED
		if strings.ToUpper(p.curTok.Literal) == "PERSISTED" {
			col.IsPersisted = true
			p.nextToken() // consume PERSISTED
		}
		// Fall through to parse constraints (NOT NULL, CHECK, FOREIGN KEY, etc.)
	} else {
		// Parse data type - be lenient if no data type is provided
		// First check if this looks like a constraint keyword (column without explicit type)
		upperLit := strings.ToUpper(p.curTok.Literal)
		isConstraintKeyword := p.curTok.Type == TokenNot || p.curTok.Type == TokenNull ||
			upperLit == "UNIQUE" || upperLit == "PRIMARY" || upperLit == "CHECK" ||
			upperLit == "DEFAULT" || upperLit == "CONSTRAINT" || upperLit == "IDENTITY" ||
			upperLit == "REFERENCES" || upperLit == "FOREIGN" || upperLit == "ROWGUIDCOL" ||
			p.curTok.Type == TokenComma || p.curTok.Type == TokenRParen

		if !isConstraintKeyword {
			dataType, err := p.parseDataTypeReference()
			if err != nil {
				// Lenient: return column definition without data type
				return spanned(p, col, astStart), nil
			}
			col.DataType = dataType
		}

		// Parse optional IDENTITY specification
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "IDENTITY" {
			identityTok := p.curTok
			p.nextToken() // consume IDENTITY
			identityOpts := &ast.IdentityOptions{}

			// Check for optional (seed, increment)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (

				// Parse seed - use parseScalarExpression to handle +/- signs and various literals
				seed, err := p.parseScalarExpression()
				if err == nil {
					identityOpts.IdentitySeed = seed
				}

				// Expect comma
				if p.curTok.Type == TokenComma {
					p.nextToken() // consume ,

					// Parse increment
					increment, err := p.parseScalarExpression()
					if err == nil {
						identityOpts.IdentityIncrement = increment
					}
				}

				// Expect closing paren
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			p.spanFrom(identityTok, identityOpts)

			// Check for NOT FOR REPLICATION
			if p.curTok.Type == TokenNot {
				notTok2 := p.curTok
				p.nextToken() // consume NOT
				if strings.ToUpper(p.curTok.Literal) == "FOR" {
					p.nextToken() // consume FOR
					if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
						p.nextToken() // consume REPLICATION
						identityOpts.NotForReplication = true
						p.spanFrom(identityTok, identityOpts)
					}
				} else if p.curTok.Type == TokenNull {
					// NOT NULL after IDENTITY - handle it here since NOT was already consumed
					nc := &ast.NullableConstraintDefinition{Nullable: false}
					p.spanTokens(nc, notTok2, p.curTok)
					p.nextToken() // consume NULL
					col.Constraints = append(col.Constraints, nc)
				}
			}

			col.IdentityOptions = identityOpts
		}
	} // end of else block for non-computed columns

	// Parse column constraints (NULL, NOT NULL, UNIQUE, PRIMARY KEY, DEFAULT, CHECK, CONSTRAINT)
	var constraintName *ast.Identifier
	var constraintTok Token
	// capEnd caps the column's span when its last clause is an inline index
	// whose trailing keywords ScriptDom excludes from both fragments.
	capEnd := 0
	capPrevEnd := -1
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)

		if upperLit == "GENERATED" {
			p.nextToken() // consume GENERATED
			if strings.ToUpper(p.curTok.Literal) == "ALWAYS" {
				p.nextToken() // consume ALWAYS
			}
			if p.curTok.Type == TokenAs {
				p.nextToken() // consume AS
			}
			// Parse the generated type: ROW START/END, SUSER_SID, SUSER_SNAME, etc.
			genType := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			if genType == "ROW" {
				// Parse START or END
				startEnd := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				if startEnd == "START" {
					col.GeneratedAlways = "RowStart"
				} else if startEnd == "END" {
					col.GeneratedAlways = "RowEnd"
				}
			} else if genType == "SUSER_SID" {
				startEnd := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				if startEnd == "START" {
					col.GeneratedAlways = "UserIdStart"
				} else if startEnd == "END" {
					col.GeneratedAlways = "UserIdEnd"
				}
			} else if genType == "SUSER_SNAME" {
				startEnd := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				if startEnd == "START" {
					col.GeneratedAlways = "UserNameStart"
				} else if startEnd == "END" {
					col.GeneratedAlways = "UserNameEnd"
				}
			} else if genType == "TRANSACTION_ID" {
				startEnd := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				if startEnd == "START" {
					col.GeneratedAlways = "TransactionIdStart"
				} else if startEnd == "END" {
					col.GeneratedAlways = "TransactionIdEnd"
				}
			} else if genType == "SEQUENCE_NUMBER" {
				startEnd := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				if startEnd == "START" {
					col.GeneratedAlways = "SequenceNumberStart"
				} else if startEnd == "END" {
					col.GeneratedAlways = "SequenceNumberEnd"
				}
			}
		} else if p.curTok.Type == TokenNot {
			notTok := p.curTok
			p.nextToken() // consume NOT
			if p.curTok.Type == TokenNull {
				nc := &ast.NullableConstraintDefinition{Nullable: false}
				p.spanTokens(nc, notTok, p.curTok)
				p.nextToken() // consume NULL
				col.Constraints = append(col.Constraints, nc)
			}
		} else if p.curTok.Type == TokenNull {
			nc := &ast.NullableConstraintDefinition{Nullable: true}
			p.tokSpan(nc, p.curTok)
			p.nextToken() // consume NULL
			col.Constraints = append(col.Constraints, nc)
		} else if upperLit == "UNIQUE" {
			uqTok := p.curTok
			p.nextToken() // consume UNIQUE
			constraint := &ast.UniqueConstraintDefinition{
				IsPrimaryKey:         false,
				ConstraintIdentifier: constraintName,
			}
			constraintName = nil // clear for next constraint
			// Parse optional CLUSTERED/NONCLUSTERED
			if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
				constraint.Clustered = true
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
				constraint.Clustered = false
				p.nextToken()
				// Check for HASH suffix
				if strings.ToUpper(p.curTok.Literal) == "HASH" {
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
					p.nextToken()
				} else {
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
				}
			}
			// Parse optional column list (column ASC, column DESC, ...)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					colWithSort := p.parseColumnWithSortOrder()
					constraint.Columns = append(constraint.Columns, colWithSort)
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
			// Parse WITH (index_options)
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
				constraint.IndexOptions = p.parseConstraintIndexOptions()
			}
			// Parse ON filegroup/partition_scheme
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				fg, _ := p.parseFileGroupOrPartitionScheme()
				constraint.OnFileGroupOrPartitionScheme = fg
			}
			// Parse NOT ENFORCED (Azure Synapse) - but only if next token is ENFORCED
			if p.curTok.Type == TokenNot && strings.ToUpper(p.peekTok.Literal) == "ENFORCED" {
				p.nextToken() // consume NOT
				p.nextToken() // consume ENFORCED
				enforced := false
				constraint.IsEnforced = &enforced
			}
			p.spanFrom(uqTok, constraint)
			// A named constraint spans from its CONSTRAINT keyword.
			if constraint.ConstraintIdentifier != nil {
				p.respanStart(constraint, constraintTok)
			}
			col.Constraints = append(col.Constraints, constraint)
		} else if upperLit == "PRIMARY" {
			uqTok := p.curTok
			p.nextToken() // consume PRIMARY
			if p.curTok.Type == TokenKey {
				p.nextToken() // consume KEY
			}
			constraint := &ast.UniqueConstraintDefinition{
				IsPrimaryKey:         true,
				ConstraintIdentifier: constraintName,
			}
			constraintName = nil // clear for next constraint
			// Parse optional CLUSTERED/NONCLUSTERED
			if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
				constraint.Clustered = true
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
				constraint.Clustered = false
				p.nextToken()
				// Check for HASH suffix
				if strings.ToUpper(p.curTok.Literal) == "HASH" {
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
					p.nextToken()
				} else {
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
				}
			}
			// Parse optional column list (column ASC, column DESC, ...)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					cwsTok := p.curTok
					colRef := &ast.ColumnReferenceExpression{
						ColumnType: "Regular",
						MultiPartIdentifier: &ast.MultiPartIdentifier{
							Identifiers: []*ast.Identifier{p.parseIdentifier()},
							Count:       1,
						},
					}
					sortOrder := ast.SortOrderNotSpecified
					if strings.ToUpper(p.curTok.Literal) == "ASC" {
						sortOrder = ast.SortOrderAscending
						p.nextToken()
					} else if strings.ToUpper(p.curTok.Literal) == "DESC" {
						sortOrder = ast.SortOrderDescending
						p.nextToken()
					}
					cwsCol := &ast.ColumnWithSortOrder{
						Column:    colRef,
						SortOrder: sortOrder,
					}
					p.spanFrom(cwsTok, cwsCol)
					constraint.Columns = append(constraint.Columns, cwsCol)
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
			// Parse WITH (index_options)
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
				constraint.IndexOptions = p.parseConstraintIndexOptions()
			}
			// Parse ON filegroup/partition_scheme
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				fg, _ := p.parseFileGroupOrPartitionScheme()
				constraint.OnFileGroupOrPartitionScheme = fg
			}
			// Parse NOT ENFORCED (Azure Synapse) - but only if next token is ENFORCED
			if p.curTok.Type == TokenNot && strings.ToUpper(p.peekTok.Literal) == "ENFORCED" {
				p.nextToken() // consume NOT
				p.nextToken() // consume ENFORCED
				enforced := false
				constraint.IsEnforced = &enforced
			}
			p.spanFrom(uqTok, constraint)
			// A named constraint spans from its CONSTRAINT keyword.
			if constraint.ConstraintIdentifier != nil {
				p.respanStart(constraint, constraintTok)
			}
			col.Constraints = append(col.Constraints, constraint)
		} else if p.curTok.Type == TokenDefault {
			defTok := p.curTok
			p.nextToken() // consume DEFAULT
			defaultConstraint := &ast.DefaultConstraintDefinition{
				ConstraintIdentifier: constraintName,
			}
			constraintName = nil // clear for next constraint

			// Parse the default expression
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			defaultConstraint.Expression = expr
			// Parse optional WITH VALUES
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				if strings.ToUpper(p.curTok.Literal) == "VALUES" {
					p.nextToken() // consume VALUES
					defaultConstraint.WithValues = true
				}
			}
			p.spanFrom(defTok, defaultConstraint)
			if defaultConstraint.ConstraintIdentifier != nil {
				p.respanStart(defaultConstraint, constraintTok)
			}
			col.DefaultConstraint = defaultConstraint
		} else if upperLit == "CHECK" {
			checkTok := p.curTok
			p.nextToken() // consume CHECK
			notForReplication := false
			// Check for NOT FOR REPLICATION (comes before the condition)
			if p.curTok.Type == TokenNot {
				p.nextToken() // consume NOT
				if strings.ToUpper(p.curTok.Literal) == "FOR" {
					p.nextToken() // consume FOR
					if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
						p.nextToken() // consume REPLICATION
						notForReplication = true
					}
				}
			}
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				cond, err := p.parseBooleanExpression()
				if err != nil {
					return nil, err
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
				ck := &ast.CheckConstraintDefinition{
					CheckCondition:       cond,
					ConstraintIdentifier: constraintName,
					NotForReplication:    notForReplication,
				}
				p.spanFrom(checkTok, ck)
				if ck.ConstraintIdentifier != nil {
					p.respanStart(ck, constraintTok)
				}
				col.Constraints = append(col.Constraints, ck)
				constraintName = nil // clear for next constraint
			}
		} else if upperLit == "FOREIGN" {
			// Parse FOREIGN KEY constraint for column
			constraint, err := p.parseForeignKeyConstraint()
			if err != nil {
				return nil, err
			}
			constraint.ConstraintIdentifier = constraintName
			if constraintName != nil {
				p.respanStart(constraint, constraintTok)
			}
			constraintName = nil
			col.Constraints = append(col.Constraints, constraint)
		} else if upperLit == "REFERENCES" {
			// Parse inline REFERENCES constraint (shorthand for FOREIGN KEY)
			p.nextToken() // consume REFERENCES
			constraint := &ast.ForeignKeyConstraintDefinition{
				ConstraintIdentifier: constraintName,
				DeleteAction:         "NotSpecified",
				UpdateAction:         "NotSpecified",
			}
			constraintName = nil
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
			// Parse ON DELETE and ON UPDATE actions
			for {
				actionUpperLit := strings.ToUpper(p.curTok.Literal)
				if actionUpperLit == "ON" {
					p.nextToken() // consume ON
					actionType := strings.ToUpper(p.curTok.Literal)
					p.nextToken() // consume DELETE or UPDATE

					action := p.parseForeignKeyAction()
					if actionType == "DELETE" {
						constraint.DeleteAction = action
					} else if actionType == "UPDATE" {
						constraint.UpdateAction = action
					}
				} else if actionUpperLit == "NOT" {
					p.nextToken() // consume NOT
					if strings.ToUpper(p.curTok.Literal) == "FOR" {
						p.nextToken() // consume FOR
						if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
							p.nextToken() // consume REPLICATION
							constraint.NotForReplication = true
						}
					}
				} else {
					break
				}
			}
			col.Constraints = append(col.Constraints, constraint)
		} else if upperLit == "CONSTRAINT" {
			constraintTok = p.curTok
			p.nextToken() // consume CONSTRAINT
			// Parse and save constraint name for next constraint
			constraintName = p.parseIdentifier()
			// Continue to parse actual constraint in next iteration
			continue
		} else if upperLit == "COLLATE" {
			p.nextToken() // consume COLLATE
			col.Collation = p.parseIdentifier()
		} else if upperLit == "INDEX" {
			idxTok := p.curTok
			p.nextToken() // consume INDEX
			indexDef := &ast.IndexDefinition{
				IndexType: &ast.IndexType{},
			}
			// Parse index name (skip if it's CLUSTERED/NONCLUSTERED/UNIQUE)
			idxUpper := strings.ToUpper(p.curTok.Literal)
			if p.curTok.Type == TokenIdent && idxUpper != "CLUSTERED" && idxUpper != "NONCLUSTERED" && idxUpper != "UNIQUE" && p.curTok.Type != TokenLParen {
				indexDef.Name = p.parseIdentifier()
			}
			// ScriptDom spans the inline index through its last concrete
			// clause; bare CLUSTERED/NONCLUSTERED keywords do not extend it.
			p.spanFrom(idxTok, indexDef)
			// Parse optional UNIQUE
			if strings.ToUpper(p.curTok.Literal) == "UNIQUE" {
				indexDef.Unique = true
				p.nextToken()
			}
			// Parse optional CLUSTERED/NONCLUSTERED [HASH]
			if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
				indexDef.IndexType.IndexTypeKind = "Clustered"
				p.nextToken()
				// Check for HASH
				if strings.ToUpper(p.curTok.Literal) == "HASH" {
					indexDef.IndexType.IndexTypeKind = "ClusteredHash"
					p.tokSpan(indexDef.IndexType, p.curTok)
					p.nextToken()
					p.spanFrom(idxTok, indexDef)
				}
			} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
				indexDef.IndexType.IndexTypeKind = "NonClustered"
				p.nextToken()
				// Check for HASH
				if strings.ToUpper(p.curTok.Literal) == "HASH" {
					indexDef.IndexType.IndexTypeKind = "NonClusteredHash"
					p.tokSpan(indexDef.IndexType, p.curTok)
					p.nextToken()
					p.spanFrom(idxTok, indexDef)
				}
			} else if strings.ToUpper(p.curTok.Literal) == "HASH" {
				// Standalone HASH is treated as NonClusteredHash
				indexDef.IndexType.IndexTypeKind = "NonClusteredHash"
				p.tokSpan(indexDef.IndexType, p.curTok)
				p.nextToken()
				p.spanFrom(idxTok, indexDef)
			}
			// Parse optional column list: (col1 [ASC|DESC], ...)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					cwsTok := p.curTok
					colWithSort := &ast.ColumnWithSortOrder{
						SortOrder: ast.SortOrderNotSpecified,
					}
					// Parse column name
					colRef := &ast.ColumnReferenceExpression{
						ColumnType: "Regular",
						MultiPartIdentifier: &ast.MultiPartIdentifier{
							Identifiers: []*ast.Identifier{p.parseIdentifier()},
						},
					}
					colRef.MultiPartIdentifier.Count = len(colRef.MultiPartIdentifier.Identifiers)
					colWithSort.Column = colRef

					// Parse optional ASC/DESC
					if strings.ToUpper(p.curTok.Literal) == "ASC" {
						colWithSort.SortOrder = ast.SortOrderAscending
						p.nextToken()
					} else if strings.ToUpper(p.curTok.Literal) == "DESC" {
						colWithSort.SortOrder = ast.SortOrderDescending
						p.nextToken()
					}
					p.spanFrom(cwsTok, colWithSort)
					indexDef.Columns = append(indexDef.Columns, colWithSort)

					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
				p.spanFrom(idxTok, indexDef)
			}
			// Parse optional WITH (index_options)
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
						optStart := p.curTok
						lenBefore := len(indexDef.IndexOptions)
						optionName := strings.ToUpper(p.curTok.Literal)
						p.nextToken() // consume option name
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if optionName == "BUCKET_COUNT" {
							opt := &ast.IndexExpressionOption{
								OptionKind: "BucketCount",
								Expression: p.intLitFromToken(p.curTok),
							}
							indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
							p.nextToken()
						} else if optionName == "DATA_COMPRESSION" {
							// Parse DATA_COMPRESSION = level [ON PARTITIONS(...)]
							compressionLevel := "None"
							levelUpper := strings.ToUpper(p.curTok.Literal)
							switch levelUpper {
							case "NONE":
								compressionLevel = "None"
							case "ROW":
								compressionLevel = "Row"
							case "PAGE":
								compressionLevel = "Page"
							case "COLUMNSTORE":
								compressionLevel = "ColumnStore"
							case "COLUMNSTORE_ARCHIVE":
								compressionLevel = "ColumnStoreArchive"
							}
							p.nextToken() // consume compression level
							opt := &ast.DataCompressionOption{
								CompressionLevel: compressionLevel,
								OptionKind:       "DataCompression",
							}
							// Check for optional ON PARTITIONS(range)
							if p.curTok.Type == TokenOn {
								p.nextToken() // consume ON
								if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
									p.nextToken() // consume PARTITIONS
									if p.curTok.Type == TokenLParen {
										p.nextToken() // consume (
										for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
											partRange := &ast.CompressionPartitionRange{}
											partRange.From = p.intLitFromToken(p.curTok)
											p.nextToken()
											if strings.ToUpper(p.curTok.Literal) == "TO" {
												p.nextToken() // consume TO
												partRange.To = p.intLitFromToken(p.curTok)
												p.nextToken()
											}
											opt.PartitionRanges = append(opt.PartitionRanges, partRange)
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
							}
							indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
						} else if optionName == "PAD_INDEX" || optionName == "STATISTICS_NORECOMPUTE" ||
							optionName == "ALLOW_ROW_LOCKS" || optionName == "ALLOW_PAGE_LOCKS" ||
							optionName == "DROP_EXISTING" || optionName == "SORT_IN_TEMPDB" ||
							optionName == "OPTIMIZE_FOR_SEQUENTIAL_KEY" {
							// ON/OFF options
							stateUpper := strings.ToUpper(p.curTok.Literal)
							optState := "On"
							if stateUpper == "OFF" {
								optState = "Off"
							}
							p.nextToken()
							optKind := map[string]string{
								"PAD_INDEX":                   "PadIndex",
								"STATISTICS_NORECOMPUTE":      "StatisticsNoRecompute",
								"ALLOW_ROW_LOCKS":             "AllowRowLocks",
								"ALLOW_PAGE_LOCKS":            "AllowPageLocks",
								"DROP_EXISTING":               "DropExisting",
								"SORT_IN_TEMPDB":              "SortInTempDB",
								"OPTIMIZE_FOR_SEQUENTIAL_KEY": "OptimizeForSequentialKey",
							}[optionName]
							indexDef.IndexOptions = append(indexDef.IndexOptions, &ast.IndexStateOption{
								OptionKind:  optKind,
								OptionState: optState,
							})
						} else if optionName == "IGNORE_DUP_KEY" {
							stateUpper := strings.ToUpper(p.curTok.Literal)
							optState := "On"
							if stateUpper == "OFF" {
								optState = "Off"
							}
							p.nextToken()
							opt := &ast.IgnoreDupKeyIndexOption{
								OptionKind:  "IgnoreDupKey",
								OptionState: optState,
							}
							// Check for optional (SUPPRESS_MESSAGES = ON/OFF)
							if optState == "On" && p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								if strings.ToUpper(p.curTok.Literal) == "SUPPRESS_MESSAGES" {
									p.nextToken() // consume SUPPRESS_MESSAGES
									if p.curTok.Type == TokenEquals {
										p.nextToken() // consume =
									}
									suppressVal := strings.ToUpper(p.curTok.Literal) == "ON"
									opt.SuppressMessagesOption = &suppressVal
									p.nextToken() // consume ON/OFF
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken() // consume )
								}
							}
							indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
						} else if optionName == "FILLFACTOR" || optionName == "MAXDOP" {
							// Integer expression options
							optKind := "FillFactor"
							if optionName == "MAXDOP" {
								optKind = "MaxDop"
							}
							indexDef.IndexOptions = append(indexDef.IndexOptions, &ast.IndexExpressionOption{
								OptionKind: optKind,
								Expression: p.intLitFromToken(p.curTok),
							})
							p.nextToken()
						} else {
							// Skip other options
							p.nextToken()
						}
						if len(indexDef.IndexOptions) > lenBefore {
							last := indexDef.IndexOptions[len(indexDef.IndexOptions)-1]
							if o, ok := last.(spannable); ok {
								p.spanFrom(optStart, o)
							}
							if eo, ok := last.(*ast.IndexExpressionOption); ok && eo.OptionKind == "BucketCount" {
								p.spanFromChild(eo, eo.Expression)
							}
							// ScriptDom excludes the trailing time unit from COMPRESSION_DELAY.
							if co, ok := last.(*ast.CompressionDelayIndexOption); ok {
								if c, ok2 := any(co.Expression).(spannable); ok2 && c.Frag().HasSpan() && co.Frag().HasSpan() {
									co.Frag().FragmentLength = c.Frag().EndOffset() - co.Frag().StartOffset
								}
							}
						}
						if p.curTok.Type == TokenComma {
							p.nextToken()
						} else {
							break
						}
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
					p.spanFrom(idxTok, indexDef)
				}
			}
			// Parse optional ON filegroup for inline index
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				fg, _ := p.parseFileGroupOrPartitionScheme()
				indexDef.OnFileGroupOrPartitionScheme = fg
				p.spanFrom(idxTok, indexDef)
			}
			// Parse optional FILESTREAM_ON for inline index
			if strings.ToUpper(p.curTok.Literal) == "FILESTREAM_ON" {
				p.nextToken() // consume FILESTREAM_ON
				ident := p.parseIdentifier()
				indexDef.FileStreamOn = &ast.IdentifierOrValueExpression{
					Value:      ident.Value,
					Identifier: ident,
				}
				p.spanFrom(idxTok, indexDef)
			}
			// Parse optional INCLUDE clause
			if strings.ToUpper(p.curTok.Literal) == "INCLUDE" {
				p.nextToken() // consume INCLUDE
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
						indexDef.IncludeColumns = append(indexDef.IncludeColumns, colRef)

						if p.curTok.Type == TokenComma {
							p.nextToken()
						} else {
							break
						}
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
					p.spanFrom(idxTok, indexDef)
				}
			}
			// Parse optional WHERE clause for filtered index
			if p.curTok.Type == TokenWhere {
				p.nextToken() // consume WHERE
				filterExpr, err := p.parseBooleanExpression()
				if err != nil {
					return nil, err
				}
				indexDef.FilterPredicate = filterExpr
				p.spanFrom(idxTok, indexDef)
			}
			if indexDef.Frag().HasSpan() {
				capEnd = indexDef.Frag().EndOffset()
				capPrevEnd = p.prevEndByte
			}
			col.Index = indexDef
		} else if upperLit == "SPARSE" {
			if col.StorageOptions == nil {
				col.StorageOptions = &ast.ColumnStorageOptions{}
			}
			col.StorageOptions.SparseOption = "Sparse"
			// ScriptDom positions storage options on the last storage keyword.
			p.tokSpan(col.StorageOptions, p.curTok)
			p.nextToken() // consume SPARSE
		} else if upperLit == "FILESTREAM" {
			if col.StorageOptions == nil {
				col.StorageOptions = &ast.ColumnStorageOptions{}
			}
			col.StorageOptions.IsFileStream = true
			p.tokSpan(col.StorageOptions, p.curTok)
			p.nextToken() // consume FILESTREAM
		} else if upperLit == "COLUMN_SET" {
			p.nextToken() // consume COLUMN_SET
			// Expect FOR ALL_SPARSE_COLUMNS
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
				if strings.ToUpper(p.curTok.Literal) == "ALL_SPARSE_COLUMNS" {
					if col.StorageOptions == nil {
						col.StorageOptions = &ast.ColumnStorageOptions{}
					}
					col.StorageOptions.SparseOption = "ColumnSetForAllSparseColumns"
					p.tokSpan(col.StorageOptions, p.curTok)
					p.nextToken() // consume ALL_SPARSE_COLUMNS
				}
			}
		} else if upperLit == "ROWGUIDCOL" {
			p.nextToken() // consume ROWGUIDCOL
			col.IsRowGuidCol = true
		} else if upperLit == "HIDDEN" {
			// ScriptDom excludes a trailing HIDDEN keyword from the column's
			// span (it still counts when later clauses follow). An earlier,
			// narrower cap (e.g. an encryption clause) stays in effect.
			hidStart, _, _ := p.srcMap.at(p.prevEndByte)
			if !(capPrevEnd == p.prevEndByte && capEnd > 0 && capEnd < hidStart) {
				capEnd = hidStart
			}
			p.nextToken() // consume HIDDEN
			capPrevEnd = p.prevEndByte
			col.IsHidden = true
		} else if upperLit == "MASKED" {
			p.nextToken() // consume MASKED
			col.IsMasked = true
			// Parse optional WITH clause for masking function
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					if strings.ToUpper(p.curTok.Literal) == "FUNCTION" {
						p.nextToken() // consume FUNCTION
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if p.curTok.Type == TokenString {
							maskFunc, err := p.parseStringLiteral()
							if err == nil {
								col.MaskingFunction = maskFunc
							}
						}
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}
			}
			// ScriptDom ends the column's span at the masking function
			// string, excluding the closing paren.
			if mf, ok := any(col.MaskingFunction).(spannable); ok && mf.Frag().HasSpan() {
				capEnd = mf.Frag().EndOffset()
				capPrevEnd = p.prevEndByte
			}
		} else if upperLit == "ENCRYPTED" {
			p.nextToken() // consume ENCRYPTED
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
			}
			// Parse encryption specification: (COLUMN_ENCRYPTION_KEY = key1, ENCRYPTION_TYPE = ..., ALGORITHM = ...)
			if p.curTok.Type == TokenLParen {
				encSpec, err := p.parseColumnEncryptionSpecification()
				if err == nil {
					col.Encryption = encSpec
					// A trailing encryption clause caps the column's span at
					// the specification's (keyword-narrowed) end.
					if encSpec.Frag().HasSpan() {
						capEnd = encSpec.Frag().EndOffset()
						capPrevEnd = p.prevEndByte
					}
				}
			}
		} else if upperLit == "IDENTITY" && col.IdentityOptions == nil {
			// IDENTITY can appear after DEFAULT or other constraints
			identityTok := p.curTok
			p.nextToken() // consume IDENTITY
			identityOpts := &ast.IdentityOptions{}

			// Check for optional (seed, increment)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (

				// Parse seed
				seed, err := p.parseScalarExpression()
				if err == nil {
					identityOpts.IdentitySeed = seed
				}

				// Expect comma
				if p.curTok.Type == TokenComma {
					p.nextToken() // consume ,

					// Parse increment
					increment, err := p.parseScalarExpression()
					if err == nil {
						identityOpts.IdentityIncrement = increment
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

			p.spanFrom(identityTok, identityOpts)
			col.IdentityOptions = identityOpts
		} else {
			break
		}
	}

	spanned(p, col, astStart)
	if f := col.Frag(); capEnd > 0 && capPrevEnd == p.prevEndByte &&
		f.HasSpan() && capEnd > f.StartOffset && capEnd < f.EndOffset() {
		f.FragmentLength = capEnd - f.StartOffset
	}
	return col, nil
}

// parseNamedTableConstraint parses a CONSTRAINT name ... table constraint
func (p *Parser) parseNamedTableConstraint() (ast.TableConstraint, error) {
	astStart := p.curTok

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
		return spanned(p, constraint, astStart), nil
	} else if upperLit == "UNIQUE" {
		constraint, err := p.parseUniqueConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return spanned(p, constraint, astStart), nil
	} else if upperLit == "FOREIGN" {
		constraint, err := p.parseForeignKeyConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return spanned(p, constraint, astStart), nil
	} else if upperLit == "CHECK" {
		constraint, err := p.parseCheckConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return spanned(p, constraint, astStart), nil
	} else if upperLit == "CONNECTION" {
		constraint, err := p.parseConnectionConstraint()
		if err != nil {
			return nil, err
		}
		constraint.ConstraintIdentifier = constraintName
		return spanned(p, constraint, astStart), nil
	}

	return nil, nil
}

// parseUnnamedTableConstraint parses an unnamed table constraint (PRIMARY KEY, UNIQUE, FOREIGN KEY, CHECK)
func (p *Parser) parseUnnamedTableConstraint() (ast.TableConstraint, error) {
	astStart := p.curTok

	upperLit := strings.ToUpper(p.curTok.Literal)

	if upperLit == "PRIMARY" {
		spanV5, spanErr5 := p.parsePrimaryKeyConstraint()
		return spanned(p, spanV5, astStart), spanErr5
	} else if upperLit == "UNIQUE" {
		spanV6, spanErr6 := p.parseUniqueConstraint()
		return spanned(p, spanV6, astStart), spanErr6
	} else if upperLit == "FOREIGN" {
		spanV7, spanErr7 := p.parseForeignKeyConstraint()
		return spanned(p, spanV7, astStart), spanErr7
	} else if upperLit == "CHECK" {
		spanV8, spanErr8 := p.parseCheckConstraint()
		return spanned(p, spanV8, astStart), spanErr8
	}

	return nil, nil
}

// parsePrimaryKeyConstraint parses PRIMARY KEY CLUSTERED/NONCLUSTERED (columns)
func (p *Parser) parsePrimaryKeyConstraint() (*ast.UniqueConstraintDefinition, error) {
	astStart := p.curTok

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

	// Parse WITH (index_options) or WITH option = value
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		constraint.IndexOptions = p.parseConstraintIndexOptions()
	}

	// Parse ON filegroup/partition_scheme
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fg, _ := p.parseFileGroupOrPartitionScheme()
		constraint.OnFileGroupOrPartitionScheme = fg
	}

	// Parse NOT ENFORCED (Azure Synapse) - but only if next token is ENFORCED
	if p.curTok.Type == TokenNot && strings.ToUpper(p.peekTok.Literal) == "ENFORCED" {
		p.nextToken() // consume NOT
		p.nextToken() // consume ENFORCED
		enforced := false
		constraint.IsEnforced = &enforced
	}

	return spanned(p, constraint, astStart), nil
}

// parseUniqueConstraint parses UNIQUE CLUSTERED/NONCLUSTERED (columns)
func (p *Parser) parseUniqueConstraint() (*ast.UniqueConstraintDefinition, error) {
	astStart := p.curTok

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

	// Parse WITH (index_options) or WITH option = value
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		constraint.IndexOptions = p.parseConstraintIndexOptions()
	}

	// Parse ON filegroup/partition_scheme
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fg, _ := p.parseFileGroupOrPartitionScheme()
		constraint.OnFileGroupOrPartitionScheme = fg
	}

	// Parse NOT ENFORCED (Azure Synapse) - but only if next token is ENFORCED
	if p.curTok.Type == TokenNot && strings.ToUpper(p.peekTok.Literal) == "ENFORCED" {
		p.nextToken() // consume NOT
		p.nextToken() // consume ENFORCED
		enforced := false
		constraint.IsEnforced = &enforced
	}

	return spanned(p, constraint, astStart), nil
}

// parseConstraintIndexOptions parses index options for constraints
// Handles both WITH (option = value, ...) and WITH option = value formats
func (p *Parser) parseConstraintIndexOptions() []ast.IndexOption {
	var options []ast.IndexOption

	// Check if we have parenthesized options
	hasParens := p.curTok.Type == TokenLParen
	if hasParens {
		p.nextToken() // consume (
	}

	for {
		if hasParens && p.curTok.Type == TokenRParen {
			break
		}
		if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon {
			break
		}
		// Stop if we hit ON (for ON filegroup clause)
		if p.curTok.Type == TokenOn {
			break
		}
		// Stop if we hit a comma that's part of table definition (not option list)
		if !hasParens && p.curTok.Type == TokenComma {
			break
		}
		// Stop if we hit closing paren that's part of table definition
		if !hasParens && p.curTok.Type == TokenRParen {
			break
		}
		// Stop if we hit a keyword that starts a new constraint
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "CONSTRAINT" || upperLit == "PRIMARY" || upperLit == "UNIQUE" ||
			upperLit == "FOREIGN" || upperLit == "CHECK" || upperLit == "DEFAULT" ||
			upperLit == "INDEX" {
			break
		}

		optTok := p.curTok
		lenBefore := len(options)
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		// Handle deprecated standalone options (no value, just skip them)
		// These are deprecated SQL Server options that don't produce AST nodes
		if optionName == "SORTED_DATA" || optionName == "SORTED_DATA_REORG" {
			// Skip these deprecated options - they don't produce IndexOption nodes
			continue
		}

		// Check for = sign
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		// Check for ON/OFF or value
		valueToken := p.curTok
		valueStr := strings.ToUpper(valueToken.Literal)
		p.nextToken()

		if optionName == "IGNORE_DUP_KEY" {
			opt := &ast.IgnoreDupKeyIndexOption{
				OptionKind:  "IgnoreDupKey",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			}
			// Check for optional (SUPPRESS_MESSAGES = ON/OFF)
			if valueStr == "ON" && p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				if strings.ToUpper(p.curTok.Literal) == "SUPPRESS_MESSAGES" {
					p.nextToken() // consume SUPPRESS_MESSAGES
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					suppressVal := strings.ToUpper(p.curTok.Literal) == "ON"
					opt.SuppressMessagesOption = &suppressVal
					p.nextToken() // consume ON/OFF
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			options = append(options, opt)
		} else if valueStr == "ON" || valueStr == "OFF" {
			opt := &ast.IndexStateOption{
				OptionKind:  p.getIndexOptionKind(optionName),
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			}
			options = append(options, opt)
		} else {
			// Expression option like FILLFACTOR = 34
			opt := &ast.IndexExpressionOption{
				OptionKind: p.getIndexOptionKind(optionName),
				Expression: p.intLitFromToken(valueToken),
			}
			options = append(options, opt)
		}

		if len(options) > lenBefore {
			last := options[len(options)-1]
			if o, ok := last.(spannable); ok {
				p.spanFrom(optTok, o)
			}
			if eo, ok := last.(*ast.IndexExpressionOption); ok && eo.OptionKind == "BucketCount" {
				p.spanFromChild(eo, eo.Expression)
			}
		}

		if p.curTok.Type == TokenComma {
			if hasParens {
				// Inside parentheses, consume comma and continue parsing options
				p.nextToken()
			} else {
				// Without parentheses, the comma separates constraints, not options
				// Don't consume it - let the outer parser handle it
				break
			}
		} else if !hasParens {
			// Before breaking, check if current token is a deprecated standalone option
			// that should be skipped. These options can appear after other options.
			nextUpperLit := strings.ToUpper(p.curTok.Literal)
			if nextUpperLit == "SORTED_DATA" || nextUpperLit == "SORTED_DATA_REORG" {
				p.nextToken() // consume the deprecated option
				// Continue the loop to potentially find more options or ON/comma
				continue
			}
			break
		}
	}

	if hasParens && p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return options
}

// parseForeignKeyConstraint parses FOREIGN KEY (columns) REFERENCES table (columns)
func (p *Parser) parseForeignKeyConstraint() (*ast.ForeignKeyConstraintDefinition, error) {
	astStart := p.curTok

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

	// Parse ON DELETE and ON UPDATE actions
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "ON" {
			p.nextToken() // consume ON
			actionType := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume DELETE or UPDATE

			// Parse action: NO ACTION, CASCADE, SET NULL, SET DEFAULT
			action := p.parseForeignKeyAction()

			if actionType == "DELETE" {
				constraint.DeleteAction = action
			} else if actionType == "UPDATE" {
				constraint.UpdateAction = action
			}
		} else if upperLit == "NOT" {
			// NOT FOR REPLICATION
			p.nextToken() // consume NOT
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
				if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
					p.nextToken() // consume REPLICATION
					constraint.NotForReplication = true
				}
			}
		} else {
			break
		}
	}

	return spanned(p, constraint, astStart), nil
}

// parseForeignKeyAction parses CASCADE, NO ACTION, SET NULL, SET DEFAULT
func (p *Parser) parseForeignKeyAction() string {
	upperLit := strings.ToUpper(p.curTok.Literal)

	switch upperLit {
	case "CASCADE":
		p.nextToken()
		return "Cascade"
	case "NO":
		p.nextToken() // consume NO
		if strings.ToUpper(p.curTok.Literal) == "ACTION" {
			p.nextToken() // consume ACTION
		}
		return "NoAction"
	case "SET":
		p.nextToken() // consume SET
		setType := strings.ToUpper(p.curTok.Literal)
		p.nextToken() // consume NULL or DEFAULT
		if setType == "NULL" {
			return "SetNull"
		} else if setType == "DEFAULT" {
			return "SetDefault"
		}
		return "NotSpecified"
	default:
		return "NotSpecified"
	}
}

// parseCheckConstraint parses CHECK (expression) or CHECK NOT FOR REPLICATION (expression)
func (p *Parser) parseCheckConstraint() (*ast.CheckConstraintDefinition, error) {
	astStart := p.curTok

	// Consume CHECK
	p.nextToken()

	constraint := &ast.CheckConstraintDefinition{}

	// Check for NOT FOR REPLICATION (comes before the condition)
	if p.curTok.Type == TokenNot {
		p.nextToken() // consume NOT
		if strings.ToUpper(p.curTok.Literal) == "FOR" {
			p.nextToken() // consume FOR
			if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
				p.nextToken() // consume REPLICATION
				constraint.NotForReplication = true
			}
		}
	}

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

	return spanned(p, constraint, astStart), nil
}

// parseConnectionConstraint parses CONNECTION (node1 TO node2, ...)
func (p *Parser) parseConnectionConstraint() (*ast.GraphConnectionConstraintDefinition, error) {
	astStart := p.curTok

	// Consume CONNECTION
	p.nextToken()

	constraint := &ast.GraphConnectionConstraintDefinition{}

	// Parse connection pairs
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			conn := &ast.GraphConnectionBetweenNodes{}

			// Parse FromNode
			fromNode, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, err
			}
			conn.FromNode = fromNode

			// Expect TO
			if strings.ToUpper(p.curTok.Literal) == "TO" {
				p.nextToken() // consume TO
			}

			// Parse ToNode
			toNode, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, err
			}
			conn.ToNode = toNode
			p.spanFromChild(conn, conn.FromNode)
			constraint.FromNodeToNodeList = append(constraint.FromNodeToNodeList, conn)

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

	// Check for ON DELETE CASCADE
	if p.curTok.Type == TokenOn && strings.ToUpper(p.peekTok.Literal) == "DELETE" {
		p.nextToken() // consume ON
		p.nextToken() // consume DELETE
		if strings.ToUpper(p.curTok.Literal) == "CASCADE" {
			constraint.DeleteAction = "Cascade"
			p.nextToken() // consume CASCADE
		} else if strings.ToUpper(p.curTok.Literal) == "NO" {
			p.nextToken() // consume NO
			if strings.ToUpper(p.curTok.Literal) == "ACTION" {
				constraint.DeleteAction = "NoAction"
				p.nextToken() // consume ACTION
			}
		}
	}

	return spanned(p, constraint, astStart), nil
}

// parseColumnWithSortOrder parses a column name with optional ASC/DESC sort order
func (p *Parser) parseColumnWithSortOrder() *ast.ColumnWithSortOrder {
	astStart := p.curTok

	col := &ast.ColumnWithSortOrder{
		SortOrder: ast.SortOrderNotSpecified,
	}

	// Check for graph pseudo-columns
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "$NODE_ID" {
		col.Column = &ast.ColumnReferenceExpression{
			ColumnType: "PseudoColumnGraphNodeId",
		}
		p.tokSpan(col.Column, p.curTok)
		p.nextToken()
	} else if upperLit == "$EDGE_ID" {
		col.Column = &ast.ColumnReferenceExpression{
			ColumnType: "PseudoColumnGraphEdgeId",
		}
		p.tokSpan(col.Column, p.curTok)
		p.nextToken()
	} else if upperLit == "$FROM_ID" {
		col.Column = &ast.ColumnReferenceExpression{
			ColumnType: "PseudoColumnGraphFromId",
		}
		p.tokSpan(col.Column, p.curTok)
		p.nextToken()
	} else if upperLit == "$TO_ID" {
		col.Column = &ast.ColumnReferenceExpression{
			ColumnType: "PseudoColumnGraphToId",
		}
		p.tokSpan(col.Column, p.curTok)
		p.nextToken()
	} else {
		// Parse regular column name
		ident := p.parseIdentifier()
		col.Column = &ast.ColumnReferenceExpression{
			ColumnType: "Regular",
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Count:       1,
				Identifiers: []*ast.Identifier{ident},
			},
		}
	}

	// Parse optional ASC/DESC
	upperLit = strings.ToUpper(p.curTok.Literal)
	if upperLit == "ASC" {
		col.SortOrder = ast.SortOrderAscending
		p.nextToken()
	} else if upperLit == "DESC" {
		col.SortOrder = ast.SortOrderDescending
		p.nextToken()
	}

	return spanned(p, col, astStart)
}

func (p *Parser) parseGrantStatement() (*ast.GrantStatement, error) {
	astStart := p.curTok

	// Consume GRANT
	p.nextToken()

	stmt := &ast.GrantStatement{}

	// Parse permission(s)
	perm := &ast.Permission{}
	permStart := p.curTok
	for p.curTok.Type != TokenTo && p.curTok.Type != TokenOn && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenCreate ||
			p.curTok.Type == TokenProcedure || p.curTok.Type == TokenView ||
			p.curTok.Type == TokenSelect || p.curTok.Type == TokenInsert ||
			p.curTok.Type == TokenUpdate || p.curTok.Type == TokenDelete ||
			p.curTok.Type == TokenAlter || p.curTok.Type == TokenExecute ||
			p.curTok.Type == TokenDrop || p.curTok.Type == TokenExternal ||
			p.curTok.Type == TokenAll || p.curTok.Type == TokenExec ||
			p.curTok.Type == TokenDatabase || p.curTok.Type == TokenTable ||
			p.curTok.Type == TokenFunction || p.curTok.Type == TokenBackup ||
			p.curTok.Type == TokenDefault || p.curTok.Type == TokenTrigger ||
			p.curTok.Type == TokenSchema || p.curTok.Type == TokenMaster ||
			p.curTok.Type == TokenKey || p.curTok.Type == TokenEncryption {
			perm.Identifiers = append(perm.Identifiers, p.spanIdent(p.curTok.Literal, "NotQuoted"))
			p.nextToken()
		} else if p.curTok.Type == TokenLParen {
			// Column list for permission (e.g., SELECT (c1, c2))
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := p.parseIdentifier()
				perm.Columns = append(perm.Columns, col)
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
			p.spanFrom(permStart, perm)
		} else if p.curTok.Type == TokenComma {
			stmt.Permissions = append(stmt.Permissions, perm)
			perm = &ast.Permission{}
			p.nextToken()
			permStart = p.curTok
		} else {
			break
		}
	}
	if len(perm.Identifiers) > 0 {
		stmt.Permissions = append(stmt.Permissions, perm)
	}

	// Check for ON clause (SecurityTargetObject)
	if p.curTok.Type == TokenOn {
		onTok := p.curTok
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
				stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
			} else if strings.ToUpper(p.curTok.Literal) == "STOPLIST" {
				p.nextToken() // consume STOPLIST
				stmt.SecurityTargetObject.ObjectKind = "FullTextStopList"
			} else {
				stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
			}
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

		// Parse object name
		if p.curTok.Type == TokenColonColon {
			p.nextToken() // consume ::
		}

		// Parse object name as multi-part identifier
		// This handles both "OBJECT::name" and plain "..name" syntax
		if p.curTok.Type == TokenDot || p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			stmt.SecurityTargetObject.ObjectName = &ast.SecurityTargetObjectName{}
			multiPart := &ast.MultiPartIdentifier{}
			for {
				// Handle double dots (e.g., ..t1) by adding empty identifier
				if p.curTok.Type == TokenDot {
					multiPart.Identifiers = append(multiPart.Identifiers, p.spanIdent("", "NotQuoted"))
				} else {
					id := p.parseIdentifier()
					multiPart.Identifiers = append(multiPart.Identifiers, id)
				}
				if p.curTok.Type == TokenDot {
					p.nextToken() // consume .
				} else {
					break
				}
			}
			multiPart.Count = len(multiPart.Identifiers)
			stmt.SecurityTargetObject.ObjectName.MultiPartIdentifier = multiPart
		}

		// Parse optional column list (c1, c2, ...)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := p.parseIdentifier()
				stmt.SecurityTargetObject.Columns = append(stmt.SecurityTargetObject.Columns, col)
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
		// ScriptDom spans the target from the ON keyword.
		p.spanFrom(onTok, stmt.SecurityTargetObject)
	}

	// Expect TO
	if p.curTok.Type == TokenTo {
		p.nextToken()
	}

	// Parse principal(s)
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		prTok := p.curTok
		principal := &ast.SecurityPrincipal{}
		if p.curTok.Type == TokenPublic {
			principal.PrincipalType = "Public"
			p.nextToken()
		} else if p.curTok.Type == TokenNull {
			principal.PrincipalType = "Null"
			p.nextToken()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			principal.PrincipalType = "Identifier"
			// parseIdentifier already calls nextToken()
			principal.Identifier = p.parseIdentifier()
		} else {
			break
		}
		p.spanFrom(prTok, principal)
		stmt.Principals = append(stmt.Principals, principal)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Check for WITH GRANT OPTION
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "GRANT" {
			p.nextToken() // consume GRANT
			if strings.ToUpper(p.curTok.Literal) == "OPTION" {
				p.nextToken() // consume OPTION
			}
		}
		stmt.WithGrantOption = true
	}

	// Check for AS clause
	if strings.ToUpper(p.curTok.Literal) == "AS" {
		p.nextToken() // consume AS
		stmt.AsClause = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseRevokeStatement() (*ast.RevokeStatement, error) {
	astStart := p.curTok

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
	permStart := p.curTok
	for p.curTok.Type != TokenTo && p.curTok.Type != TokenOn && p.curTok.Type != TokenEOF && strings.ToUpper(p.curTok.Literal) != "FROM" {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenCreate ||
			p.curTok.Type == TokenProcedure || p.curTok.Type == TokenView ||
			p.curTok.Type == TokenSelect || p.curTok.Type == TokenInsert ||
			p.curTok.Type == TokenUpdate || p.curTok.Type == TokenDelete ||
			p.curTok.Type == TokenAlter || p.curTok.Type == TokenExecute ||
			p.curTok.Type == TokenDrop || p.curTok.Type == TokenExternal ||
			p.curTok.Type == TokenAll || p.curTok.Type == TokenExec ||
			p.curTok.Type == TokenDatabase || p.curTok.Type == TokenTable ||
			p.curTok.Type == TokenFunction || p.curTok.Type == TokenBackup ||
			p.curTok.Type == TokenDefault || p.curTok.Type == TokenTrigger ||
			p.curTok.Type == TokenSchema {
			perm.Identifiers = append(perm.Identifiers, p.spanIdent(p.curTok.Literal, "NotQuoted"))
			p.nextToken()
		} else if p.curTok.Type == TokenLParen {
			// Parse column list for permission
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := p.parseIdentifier()
				perm.Columns = append(perm.Columns, col)
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
			p.spanFrom(permStart, perm)
		} else if p.curTok.Type == TokenComma {
			stmt.Permissions = append(stmt.Permissions, perm)
			perm = &ast.Permission{}
			p.nextToken()
			permStart = p.curTok
		} else {
			break
		}
	}
	if len(perm.Identifiers) > 0 {
		stmt.Permissions = append(stmt.Permissions, perm)
	}

	// Check for ON clause (SecurityTargetObject)
	if p.curTok.Type == TokenOn {
		onTok := p.curTok
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
				stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
			} else if strings.ToUpper(p.curTok.Literal) == "STOPLIST" {
				p.nextToken()
				stmt.SecurityTargetObject.ObjectKind = "FullTextStopList"
			} else {
				stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
			}
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

		// Parse object name
		if p.curTok.Type == TokenColonColon {
			p.nextToken() // consume ::
		}

		// Parse object name as multi-part identifier
		// This handles both "OBJECT::name" and plain "..name" syntax
		if p.curTok.Type == TokenDot || p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			stmt.SecurityTargetObject.ObjectName = &ast.SecurityTargetObjectName{}
			multiPart := &ast.MultiPartIdentifier{}
			for {
				// Handle double dots (e.g., ..t1) by adding empty identifier
				if p.curTok.Type == TokenDot {
					multiPart.Identifiers = append(multiPart.Identifiers, p.spanIdent("", "NotQuoted"))
				} else {
					id := p.parseIdentifier()
					multiPart.Identifiers = append(multiPart.Identifiers, id)
				}
				if p.curTok.Type == TokenDot {
					p.nextToken() // consume .
				} else {
					break
				}
			}
			multiPart.Count = len(multiPart.Identifiers)
			stmt.SecurityTargetObject.ObjectName.MultiPartIdentifier = multiPart
		}

		// Parse optional column list (c1, c2, ...)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := p.parseIdentifier()
				stmt.SecurityTargetObject.Columns = append(stmt.SecurityTargetObject.Columns, col)
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
		// ScriptDom spans the target from the ON keyword.
		p.spanFrom(onTok, stmt.SecurityTargetObject)
	}

	// Expect TO or FROM
	if p.curTok.Type == TokenTo || strings.ToUpper(p.curTok.Literal) == "FROM" {
		p.nextToken()
	}

	// Parse principal(s)
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && strings.ToUpper(p.curTok.Literal) != "CASCADE" && strings.ToUpper(p.curTok.Literal) != "AS" {
		prTok := p.curTok
		principal := &ast.SecurityPrincipal{}
		if p.curTok.Type == TokenPublic {
			principal.PrincipalType = "Public"
			p.nextToken()
		} else if p.curTok.Type == TokenNull {
			principal.PrincipalType = "Null"
			p.nextToken()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			principal.PrincipalType = "Identifier"
			principal.Identifier = p.parseIdentifier()
		} else {
			break
		}
		p.spanFrom(prTok, principal)
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

	// Check for AS
	if strings.ToUpper(p.curTok.Literal) == "AS" {
		p.nextToken() // consume AS
		stmt.AsClause = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseDenyStatement() (*ast.DenyStatement, error) {
	astStart := p.curTok

	// Consume DENY
	p.nextToken()

	stmt := &ast.DenyStatement{}

	// Parse permission(s)
	perm := &ast.Permission{}
	permStart := p.curTok
	for p.curTok.Type != TokenTo && p.curTok.Type != TokenOn && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenCreate ||
			p.curTok.Type == TokenProcedure || p.curTok.Type == TokenView ||
			p.curTok.Type == TokenSelect || p.curTok.Type == TokenInsert ||
			p.curTok.Type == TokenUpdate || p.curTok.Type == TokenDelete ||
			p.curTok.Type == TokenAlter || p.curTok.Type == TokenExecute ||
			p.curTok.Type == TokenDrop || p.curTok.Type == TokenExternal ||
			p.curTok.Type == TokenAll || p.curTok.Type == TokenExec ||
			p.curTok.Type == TokenDatabase || p.curTok.Type == TokenTable ||
			p.curTok.Type == TokenFunction || p.curTok.Type == TokenBackup ||
			p.curTok.Type == TokenDefault || p.curTok.Type == TokenTrigger ||
			p.curTok.Type == TokenSchema || p.curTok.Type == TokenMaster ||
			p.curTok.Type == TokenKey || p.curTok.Type == TokenEncryption {
			perm.Identifiers = append(perm.Identifiers, p.spanIdent(p.curTok.Literal, "NotQuoted"))
			p.nextToken()
		} else if p.curTok.Type == TokenLParen {
			// Column list for permission (e.g., SELECT (c1, c2))
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := p.parseIdentifier()
				perm.Columns = append(perm.Columns, col)
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
			p.spanFrom(permStart, perm)
		} else if p.curTok.Type == TokenComma {
			stmt.Permissions = append(stmt.Permissions, perm)
			perm = &ast.Permission{}
			p.nextToken()
			permStart = p.curTok
		} else {
			break
		}
	}
	if len(perm.Identifiers) > 0 {
		stmt.Permissions = append(stmt.Permissions, perm)
	}

	// Check for ON clause (SecurityTargetObject)
	if p.curTok.Type == TokenOn {
		onTok := p.curTok
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
				stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
			} else if strings.ToUpper(p.curTok.Literal) == "STOPLIST" {
				p.nextToken()
				stmt.SecurityTargetObject.ObjectKind = "FullTextStopList"
			} else {
				stmt.SecurityTargetObject.ObjectKind = "FullTextCatalog"
			}
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

		// Parse object name
		if p.curTok.Type == TokenColonColon {
			p.nextToken() // consume ::
		}

		// Parse object name as multi-part identifier
		// This handles both "OBJECT::name" and plain "..name" syntax
		if p.curTok.Type == TokenDot || p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			stmt.SecurityTargetObject.ObjectName = &ast.SecurityTargetObjectName{}
			multiPart := &ast.MultiPartIdentifier{}
			for {
				// Handle double dots (e.g., ..t1) by adding empty identifier
				if p.curTok.Type == TokenDot {
					multiPart.Identifiers = append(multiPart.Identifiers, p.spanIdent("", "NotQuoted"))
				} else {
					id := p.parseIdentifier()
					multiPart.Identifiers = append(multiPart.Identifiers, id)
				}
				if p.curTok.Type == TokenDot {
					p.nextToken() // consume .
				} else {
					break
				}
			}
			multiPart.Count = len(multiPart.Identifiers)
			stmt.SecurityTargetObject.ObjectName.MultiPartIdentifier = multiPart
		}

		// Parse optional column list (c1, c2, ...)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := p.parseIdentifier()
				stmt.SecurityTargetObject.Columns = append(stmt.SecurityTargetObject.Columns, col)
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
		// ScriptDom spans the target from the ON keyword.
		p.spanFrom(onTok, stmt.SecurityTargetObject)
	}

	// Expect TO
	if p.curTok.Type == TokenTo {
		p.nextToken()
	}

	// Parse principal(s)
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && strings.ToUpper(p.curTok.Literal) != "CASCADE" && strings.ToUpper(p.curTok.Literal) != "AS" {
		prTok := p.curTok
		principal := &ast.SecurityPrincipal{}
		if p.curTok.Type == TokenPublic {
			principal.PrincipalType = "Public"
			p.nextToken()
		} else if p.curTok.Type == TokenNull {
			principal.PrincipalType = "Null"
			p.nextToken()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			principal.PrincipalType = "Identifier"
			principal.Identifier = p.parseIdentifier()
		} else {
			break
		}
		p.spanFrom(prTok, principal)
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

	// Check for AS clause
	if strings.ToUpper(p.curTok.Literal) == "AS" {
		p.nextToken() // consume AS
		stmt.AsClause = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
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
	if s.FileStreamOn != nil {
		node["FileStreamOn"] = identifierOrValueExpressionToJSON(s.FileStreamOn)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = tableOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	if s.FederationScheme != nil {
		node["FederationScheme"] = federationSchemeToJSON(s.FederationScheme)
	}
	if s.SelectStatement != nil {
		node["SelectStatement"] = selectStatementToJSON(s.SelectStatement)
	}
	if len(s.CtasColumns) > 0 {
		cols := make([]jsonNode, len(s.CtasColumns))
		for i, col := range s.CtasColumns {
			cols[i] = identifierToJSON(col)
		}
		node["CtasColumns"] = cols
	}
	return addSpan(node, frag(s))
}

func federationSchemeToJSON(fs *ast.FederationScheme) jsonNode {
	node := jsonNode{
		"$type": "FederationScheme",
	}
	if fs.DistributionName != nil {
		node["DistributionName"] = identifierToJSON(fs.DistributionName)
	}
	if fs.ColumnName != nil {
		node["ColumnName"] = identifierToJSON(fs.ColumnName)
	}
	return addSpan(node, frag(fs))
}

func tableOptionToJSON(opt ast.TableOption) jsonNode {
	switch o := opt.(type) {
	case *ast.TableDataCompressionOption:
		node := jsonNode{
			"$type":      "TableDataCompressionOption",
			"OptionKind": o.OptionKind,
		}
		if o.DataCompressionOption != nil {
			node["DataCompressionOption"] = dataCompressionOptionToJSON(o.DataCompressionOption)
		}
		return addSpan(node, frag(opt))
	case *ast.TableXmlCompressionOption:
		node := jsonNode{
			"$type":      "TableXmlCompressionOption",
			"OptionKind": o.OptionKind,
		}
		if o.XmlCompressionOption != nil {
			node["XmlCompressionOption"] = xmlCompressionOptionToJSON(o.XmlCompressionOption)
		}
		return addSpan(node, frag(opt))
	case *ast.TableIndexOption:
		node := jsonNode{
			"$type":      "TableIndexOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = tableIndexTypeToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.TableDistributionOption:
		node := jsonNode{
			"$type":      "TableDistributionOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = tableDistributionPolicyToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.TablePartitionOption:
		node := jsonNode{
			"$type":      "TablePartitionOption",
			"OptionKind": o.OptionKind,
		}
		if o.PartitionColumn != nil {
			node["PartitionColumn"] = identifierToJSON(o.PartitionColumn)
		}
		if o.PartitionOptionSpecs != nil {
			node["PartitionOptionSpecs"] = tablePartitionOptionSpecsToJSON(o.PartitionOptionSpecs)
		}
		return addSpan(node, frag(opt))
	case *ast.SystemVersioningTableOption:
		return addSpan(systemVersioningTableOptionToJSON(o), frag(opt))
	case *ast.LedgerTableOption:
		return addSpan(ledgerTableOptionToJSON(o), frag(opt))
	case *ast.MemoryOptimizedTableOption:
		return addSpan(jsonNode{
			"$type":       "MemoryOptimizedTableOption",
			"OptionKind":  o.OptionKind,
			"OptionState": o.OptionState,
		}, frag(opt))
	case *ast.DurabilityTableOption:
		return addSpan(jsonNode{
			"$type":                     "DurabilityTableOption",
			"OptionKind":                o.OptionKind,
			"DurabilityTableOptionKind": o.DurabilityTableOptionKind,
		}, frag(opt))
	case *ast.FileTableDirectoryTableOption:
		node := jsonNode{
			"$type":      "FileTableDirectoryTableOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.FileTableCollateFileNameTableOption:
		node := jsonNode{
			"$type":      "FileTableCollateFileNameTableOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = identifierToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.FileTableConstraintNameTableOption:
		node := jsonNode{
			"$type":      "FileTableConstraintNameTableOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = identifierToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.LockEscalationTableOption:
		return addSpan(jsonNode{
			"$type":      "LockEscalationTableOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.FileStreamOnTableOption:
		node := jsonNode{
			"$type":      "FileStreamOnTableOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = identifierOrValueExpressionToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.RemoteDataArchiveTableOption:
		node := jsonNode{
			"$type":          "RemoteDataArchiveTableOption",
			"RdaTableOption": o.RdaTableOption,
			"MigrationState": o.MigrationState,
			"OptionKind":     o.OptionKind,
		}
		if o.FilterPredicate != nil {
			node["FilterPredicate"] = scalarExpressionToJSON(o.FilterPredicate)
		}
		return addSpan(node, frag(opt))
	case *ast.RemoteDataArchiveAlterTableOption:
		node := jsonNode{
			"$type":                      "RemoteDataArchiveAlterTableOption",
			"RdaTableOption":             o.RdaTableOption,
			"MigrationState":             o.MigrationState,
			"IsMigrationStateSpecified":  o.IsMigrationStateSpecified,
			"IsFilterPredicateSpecified": o.IsFilterPredicateSpecified,
			"OptionKind":                 o.OptionKind,
		}
		if o.FilterPredicate != nil {
			node["FilterPredicate"] = scalarExpressionToJSON(o.FilterPredicate)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownTableOption"}, frag(opt))
	}
}

func dataCompressionOptionToJSON(opt *ast.DataCompressionOption) jsonNode {
	node := jsonNode{
		"$type":            "DataCompressionOption",
		"CompressionLevel": opt.CompressionLevel,
		"OptionKind":       opt.OptionKind,
	}
	if len(opt.PartitionRanges) > 0 {
		ranges := make([]jsonNode, len(opt.PartitionRanges))
		for i, pr := range opt.PartitionRanges {
			ranges[i] = compressionPartitionRangeToJSON(pr)
		}
		node["PartitionRanges"] = ranges
	}
	return addSpan(node, frag(opt))
}

func xmlCompressionOptionToJSON(opt *ast.XmlCompressionOption) jsonNode {
	node := jsonNode{
		"$type":        "XmlCompressionOption",
		"IsCompressed": opt.IsCompressed,
		"OptionKind":   opt.OptionKind,
	}
	if len(opt.PartitionRanges) > 0 {
		ranges := make([]jsonNode, len(opt.PartitionRanges))
		for i, pr := range opt.PartitionRanges {
			ranges[i] = compressionPartitionRangeToJSON(pr)
		}
		node["PartitionRanges"] = ranges
	}
	return addSpan(node, frag(opt))
}

func tableDistributionPolicyToJSON(policy ast.TableDistributionPolicy) jsonNode {
	switch p := policy.(type) {
	case *ast.TableHashDistributionPolicy:
		node := jsonNode{
			"$type": "TableHashDistributionPolicy",
		}
		if p.DistributionColumn != nil {
			node["DistributionColumn"] = identifierToJSON(p.DistributionColumn)
		}
		if len(p.DistributionColumns) > 0 {
			cols := make([]jsonNode, len(p.DistributionColumns))
			for i, c := range p.DistributionColumns {
				// First column is same as DistributionColumn, use $ref
				if i == 0 && p.DistributionColumn != nil {
					cols[i] = jsonNode{"$ref": "Identifier"}
				} else {
					cols[i] = identifierToJSON(c)
				}
			}
			node["DistributionColumns"] = cols
		}
		return addSpan(node, frag(policy))
	case *ast.TableRoundRobinDistributionPolicy:
		return addSpan(jsonNode{
			"$type": "TableRoundRobinDistributionPolicy",
		}, frag(policy))
	case *ast.TableReplicateDistributionPolicy:
		return addSpan(jsonNode{
			"$type": "TableReplicateDistributionPolicy",
		}, frag(policy))
	default:
		return addSpan(jsonNode{"$type": "UnknownDistributionPolicy"}, frag(policy))
	}
}

func tablePartitionOptionSpecsToJSON(specs *ast.TablePartitionOptionSpecifications) jsonNode {
	node := jsonNode{
		"$type": "TablePartitionOptionSpecifications",
		"Range": specs.Range,
	}
	if len(specs.BoundaryValues) > 0 {
		vals := make([]jsonNode, len(specs.BoundaryValues))
		for i, v := range specs.BoundaryValues {
			vals[i] = scalarExpressionToJSON(v)
		}
		node["BoundaryValues"] = vals
	}
	return addSpan(node, frag(specs))
}

func tableIndexTypeToJSON(t ast.TableIndexType) jsonNode {
	switch v := t.(type) {
	case *ast.TableClusteredIndexType:
		node := jsonNode{
			"$type":       "TableClusteredIndexType",
			"ColumnStore": v.ColumnStore,
		}
		if len(v.Columns) > 0 {
			cols := make([]jsonNode, len(v.Columns))
			for i, c := range v.Columns {
				cols[i] = columnWithSortOrderToJSON(c)
			}
			node["Columns"] = cols
		}
		if len(v.OrderedColumns) > 0 {
			cols := make([]jsonNode, len(v.OrderedColumns))
			for i, c := range v.OrderedColumns {
				cols[i] = columnReferenceExpressionToJSON(c)
			}
			node["OrderedColumns"] = cols
		}
		return addSpan(node, frag(t))
	case *ast.TableNonClusteredIndexType:
		return addSpan(jsonNode{
			"$type": "TableNonClusteredIndexType",
		}, frag(t))
	default:
		return addSpan(jsonNode{"$type": "UnknownTableIndexType"}, frag(t))
	}
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
	return addSpan(node, frag(pr))
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
	if t.SystemTimePeriod != nil {
		node["SystemTimePeriod"] = systemTimePeriodDefinitionToJSON(t.SystemTimePeriod)
	}
	return addSpan(node, frag(t))
}

func systemTimePeriodDefinitionToJSON(s *ast.SystemTimePeriodDefinition) jsonNode {
	return addSpan(jsonNode{
		"$type":           "SystemTimePeriodDefinition",
		"StartTimeColumn": identifierToJSON(s.StartTimeColumn),
		"EndTimeColumn":   identifierToJSON(s.EndTimeColumn),
	}, frag(s))
}

func tableConstraintToJSON(c ast.TableConstraint) jsonNode {
	switch constraint := c.(type) {
	case *ast.UniqueConstraintDefinition:
		return addSpan(uniqueConstraintToJSON(constraint), frag(c))
	case *ast.CheckConstraintDefinition:
		return addSpan(checkConstraintToJSON(constraint), frag(c))
	case *ast.ForeignKeyConstraintDefinition:
		return addSpan(foreignKeyConstraintToJSON(constraint), frag(c))
	case *ast.GraphConnectionConstraintDefinition:
		return addSpan(graphConnectionConstraintToJSON(constraint), frag(c))
	case *ast.DefaultConstraintDefinition:
		return addSpan(defaultConstraintToJSON(constraint), frag(c))
	default:
		return addSpan(jsonNode{"$type": "UnknownTableConstraint"}, frag(c))
	}
}

func graphConnectionConstraintToJSON(c *ast.GraphConnectionConstraintDefinition) jsonNode {
	node := jsonNode{
		"$type": "GraphConnectionConstraintDefinition",
	}
	if len(c.FromNodeToNodeList) > 0 {
		connections := make([]jsonNode, len(c.FromNodeToNodeList))
		for i, conn := range c.FromNodeToNodeList {
			connNode := jsonNode{
				"$type": "GraphConnectionBetweenNodes",
			}
			if conn.FromNode != nil {
				connNode["FromNode"] = schemaObjectNameToJSON(conn.FromNode)
			}
			if conn.ToNode != nil {
				connNode["ToNode"] = schemaObjectNameToJSON(conn.ToNode)
			}
			connections[i] = addSpan(connNode, frag(conn))
		}
		node["FromNodeToNodeList"] = connections
	}
	deleteAction := c.DeleteAction
	if deleteAction == "" {
		deleteAction = "NotSpecified"
	}
	node["DeleteAction"] = deleteAction
	if c.ConstraintIdentifier != nil {
		node["ConstraintIdentifier"] = identifierToJSON(c.ConstraintIdentifier)
	}
	return addSpan(node, frag(c))
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
	// Output IsEnforced if it's explicitly set
	if c.IsEnforced != nil {
		node["IsEnforced"] = *c.IsEnforced
	}
	return addSpan(node, frag(c))
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
	if c.GeneratedAlways != "" {
		node["GeneratedAlways"] = c.GeneratedAlways
	}
	if c.StorageOptions != nil {
		node["StorageOptions"] = columnStorageOptionsToJSON(c.StorageOptions)
	}
	if c.ComputedColumnExpression != nil {
		node["ComputedColumnExpression"] = scalarExpressionToJSON(c.ComputedColumnExpression)
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
	if c.Collation != nil {
		node["Collation"] = identifierToJSON(c.Collation)
	}
	if c.Index != nil {
		node["Index"] = indexDefinitionToJSON(c.Index)
	}
	if c.Encryption != nil {
		node["Encryption"] = columnEncryptionDefinitionToJSON(c.Encryption)
	}
	if c.MaskingFunction != nil {
		node["MaskingFunction"] = scalarExpressionToJSON(c.MaskingFunction)
	}
	return addSpan(node, frag(c))
}

func columnStorageOptionsToJSON(o *ast.ColumnStorageOptions) jsonNode {
	sparseOption := o.SparseOption
	if sparseOption == "" {
		sparseOption = "None"
	}
	return addSpan(jsonNode{
		"$type":        "ColumnStorageOptions",
		"IsFileStream": o.IsFileStream,
		"SparseOption": sparseOption,
	}, frag(o))
}

func defaultConstraintToJSON(d *ast.DefaultConstraintDefinition) jsonNode {
	node := jsonNode{
		"$type":      "DefaultConstraintDefinition",
		"WithValues": d.WithValues,
	}
	if d.ConstraintIdentifier != nil {
		node["ConstraintIdentifier"] = identifierToJSON(d.ConstraintIdentifier)
	}
	if d.Expression != nil {
		node["Expression"] = scalarExpressionToJSON(d.Expression)
	}
	if d.Column != nil {
		node["Column"] = identifierToJSON(d.Column)
	}
	return addSpan(node, frag(d))
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
	return addSpan(node, frag(i))
}

func constraintDefinitionToJSON(c ast.ConstraintDefinition) jsonNode {
	switch constraint := c.(type) {
	case *ast.NullableConstraintDefinition:
		return addSpan(jsonNode{
			"$type":    "NullableConstraintDefinition",
			"Nullable": constraint.Nullable,
		}, frag(c))
	case *ast.UniqueConstraintDefinition:
		return addSpan(uniqueConstraintToJSON(constraint), frag(c))
	case *ast.CheckConstraintDefinition:
		return addSpan(checkConstraintToJSON(constraint), frag(c))
	case *ast.ForeignKeyConstraintDefinition:
		return addSpan(foreignKeyConstraintToJSON(constraint), frag(c))
	default:
		return addSpan(jsonNode{"$type": "UnknownConstraint"}, frag(c))
	}
}

func uniqueConstraintToJSON(c *ast.UniqueConstraintDefinition) jsonNode {
	node := jsonNode{
		"$type":        "UniqueConstraintDefinition",
		"IsPrimaryKey": c.IsPrimaryKey,
	}
	// Output Clustered if it's true, or if IndexType is NonClustered (not Hash variants)
	if c.Clustered {
		node["Clustered"] = c.Clustered
	} else if c.IndexType != nil && (c.IndexType.IndexTypeKind == "NonClustered" || c.IndexType.IndexTypeKind == "Clustered") {
		node["Clustered"] = c.Clustered
	}
	// Output IsEnforced if it's explicitly set
	if c.IsEnforced != nil {
		node["IsEnforced"] = *c.IsEnforced
	}
	if len(c.Columns) > 0 {
		cols := make([]jsonNode, len(c.Columns))
		for i, col := range c.Columns {
			cols[i] = columnWithSortOrderToJSON(col)
		}
		node["Columns"] = cols
	}
	if len(c.IndexOptions) > 0 {
		opts := make([]jsonNode, len(c.IndexOptions))
		for i, opt := range c.IndexOptions {
			opts[i] = indexOptionToJSON(opt)
		}
		node["IndexOptions"] = opts
	}
	if c.OnFileGroupOrPartitionScheme != nil {
		node["OnFileGroupOrPartitionScheme"] = fileGroupOrPartitionSchemeToJSON(c.OnFileGroupOrPartitionScheme)
	}
	if c.IndexType != nil {
		node["IndexType"] = indexTypeToJSON(c.IndexType)
	}
	if c.ConstraintIdentifier != nil {
		node["ConstraintIdentifier"] = identifierToJSON(c.ConstraintIdentifier)
	}
	return addSpan(node, frag(c))
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
	return addSpan(node, frag(c))
}

func dataTypeReferenceToJSON(d ast.DataTypeReference) jsonNode {
	switch dt := d.(type) {
	case *ast.SqlDataTypeReference:
		return addSpan(sqlDataTypeReferenceToJSON(dt), frag(d))
	case *ast.XmlDataTypeReference:
		return addSpan(xmlDataTypeReferenceToJSON(dt), frag(d))
	case *ast.UserDataTypeReference:
		return addSpan(userDataTypeReferenceToJSON(dt), frag(d))
	default:
		return addSpan(jsonNode{"$type": "UnknownDataType"}, frag(d))
	}
}

func userDataTypeReferenceToJSON(dt *ast.UserDataTypeReference) jsonNode {
	node := jsonNode{
		"$type": "UserDataTypeReference",
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
	return addSpan(node, frag(dt))
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
	return addSpan(node, frag(dt))
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
	if s.AsClause != nil {
		node["AsClause"] = identifierToJSON(s.AsClause)
	}
	return addSpan(node, frag(s))
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
	if s.AsClause != nil {
		node["AsClause"] = identifierToJSON(s.AsClause)
	}
	return addSpan(node, frag(s))
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
	if s.AsClause != nil {
		node["AsClause"] = identifierToJSON(s.AsClause)
	}
	return addSpan(node, frag(s))
}

func securityTargetObjectToJSON(s *ast.SecurityTargetObject) jsonNode {
	node := jsonNode{
		"$type":      "SecurityTargetObject",
		"ObjectKind": s.ObjectKind,
	}
	if s.ObjectName != nil {
		node["ObjectName"] = securityTargetObjectNameToJSON(s.ObjectName)
	}
	if len(s.Columns) > 0 {
		cols := make([]jsonNode, len(s.Columns))
		for i, c := range s.Columns {
			cols[i] = identifierToJSON(c)
		}
		node["Columns"] = cols
	}
	return addSpan(node, frag(s))
}

func securityTargetObjectNameToJSON(s *ast.SecurityTargetObjectName) jsonNode {
	node := jsonNode{
		"$type": "SecurityTargetObjectName",
	}
	if s.MultiPartIdentifier != nil {
		node["MultiPartIdentifier"] = multiPartIdentifierToJSON(s.MultiPartIdentifier)
	}
	return addSpan(node, frag(s))
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
	if len(p.Columns) > 0 {
		cols := make([]jsonNode, len(p.Columns))
		for i, col := range p.Columns {
			cols[i] = identifierToJSON(col)
		}
		node["Columns"] = cols
	}
	return addSpan(node, frag(p))
}

func securityPrincipalToJSON(p *ast.SecurityPrincipal) jsonNode {
	node := jsonNode{
		"$type":         "SecurityPrincipal",
		"PrincipalType": p.PrincipalType,
	}
	if p.Identifier != nil {
		node["Identifier"] = identifierToJSON(p.Identifier)
	}
	return addSpan(node, frag(p))
}

func predicateSetStatementToJSON(s *ast.PredicateSetStatement) jsonNode {
	return addSpan(jsonNode{
		"$type":   "PredicateSetStatement",
		"Options": s.Options,
		"IsOn":    s.IsOn,
	}, frag(s))
}

func setStatisticsStatementToJSON(s *ast.SetStatisticsStatement) jsonNode {
	return addSpan(jsonNode{
		"$type":   "SetStatisticsStatement",
		"Options": s.Options,
		"IsOn":    s.IsOn,
	}, frag(s))
}

func setRowCountStatementToJSON(s *ast.SetRowCountStatement) jsonNode {
	node := jsonNode{
		"$type": "SetRowCountStatement",
	}
	if s.NumberRows != nil {
		node["NumberRows"] = scalarExpressionToJSON(s.NumberRows)
	}
	return addSpan(node, frag(s))
}

func setOffsetsStatementToJSON(s *ast.SetOffsetsStatement) jsonNode {
	return addSpan(jsonNode{
		"$type":   "SetOffsetsStatement",
		"Options": s.Options,
		"IsOn":    s.IsOn,
	}, frag(s))
}

func setCommandStatementToJSON(s *ast.SetCommandStatement) jsonNode {
	node := jsonNode{
		"$type": "SetCommandStatement",
	}
	if len(s.Commands) > 0 {
		cmds := make([]jsonNode, len(s.Commands))
		for i, cmd := range s.Commands {
			cmds[i] = setCommandToJSON(cmd)
		}
		node["Commands"] = cmds
	}
	return addSpan(node, frag(s))
}

func setCommandToJSON(cmd ast.SetCommand) jsonNode {
	switch c := cmd.(type) {
	case *ast.SetFipsFlaggerCommand:
		return addSpan(jsonNode{
			"$type":           "SetFipsFlaggerCommand",
			"ComplianceLevel": c.ComplianceLevel,
		}, frag(cmd))
	case *ast.GeneralSetCommand:
		node := jsonNode{
			"$type":       "GeneralSetCommand",
			"CommandType": c.CommandType,
		}
		if c.Parameter != nil {
			node["Parameter"] = scalarExpressionToJSON(c.Parameter)
		}
		return addSpan(node, frag(cmd))
	default:
		return addSpan(jsonNode{"$type": "UnknownSetCommand"}, frag(cmd))
	}
}

func setTransactionIsolationLevelStatementToJSON(s *ast.SetTransactionIsolationLevelStatement) jsonNode {
	return addSpan(jsonNode{
		"$type": "SetTransactionIsolationLevelStatement",
		"Level": s.Level,
	}, frag(s))
}

func setTextSizeStatementToJSON(s *ast.SetTextSizeStatement) jsonNode {
	node := jsonNode{
		"$type": "SetTextSizeStatement",
	}
	if s.TextSize != nil {
		node["TextSize"] = scalarExpressionToJSON(s.TextSize)
	}
	return addSpan(node, frag(s))
}

func setIdentityInsertStatementToJSON(s *ast.SetIdentityInsertStatement) jsonNode {
	node := jsonNode{
		"$type": "SetIdentityInsertStatement",
		"IsOn":  s.IsOn,
	}
	if s.Table != nil {
		node["Table"] = schemaObjectNameToJSON(s.Table)
	}
	return addSpan(node, frag(s))
}

func setErrorLevelStatementToJSON(s *ast.SetErrorLevelStatement) jsonNode {
	node := jsonNode{
		"$type": "SetErrorLevelStatement",
	}
	if s.Level != nil {
		node["Level"] = scalarExpressionToJSON(s.Level)
	}
	return addSpan(node, frag(s))
}

func commitTransactionStatementToJSON(s *ast.CommitTransactionStatement) jsonNode {
	node := jsonNode{
		"$type":                   "CommitTransactionStatement",
		"DelayedDurabilityOption": s.DelayedDurabilityOption,
	}
	if s.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func rollbackTransactionStatementToJSON(s *ast.RollbackTransactionStatement) jsonNode {
	node := jsonNode{
		"$type": "RollbackTransactionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func saveTransactionStatementToJSON(s *ast.SaveTransactionStatement) jsonNode {
	node := jsonNode{
		"$type": "SaveTransactionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierOrValueExpressionToJSON(s.Name)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func useStatementToJSON(s *ast.UseStatement) jsonNode {
	node := jsonNode{
		"$type": "UseStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	return addSpan(node, frag(s))
}

func killStatementToJSON(s *ast.KillStatement) jsonNode {
	node := jsonNode{
		"$type":          "KillStatement",
		"WithStatusOnly": s.WithStatusOnly,
	}
	if s.Parameter != nil {
		node["Parameter"] = scalarExpressionToJSON(s.Parameter)
	}
	return addSpan(node, frag(s))
}

func killStatsJobStatementToJSON(s *ast.KillStatsJobStatement) jsonNode {
	node := jsonNode{
		"$type": "KillStatsJobStatement",
	}
	if s.JobId != nil {
		node["JobId"] = scalarExpressionToJSON(s.JobId)
	}
	return addSpan(node, frag(s))
}

func killQueryNotificationSubscriptionStatementToJSON(s *ast.KillQueryNotificationSubscriptionStatement) jsonNode {
	node := jsonNode{
		"$type": "KillQueryNotificationSubscriptionStatement",
		"All":   s.All,
	}
	if s.SubscriptionId != nil {
		node["SubscriptionId"] = scalarExpressionToJSON(s.SubscriptionId)
	}
	return addSpan(node, frag(s))
}

func closeSymmetricKeyStatementToJSON(s *ast.CloseSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CloseSymmetricKeyStatement",
		"All":   s.All,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func closeMasterKeyStatementToJSON(s *ast.CloseMasterKeyStatement) jsonNode {
	return addSpan(jsonNode{
		"$type": "CloseMasterKeyStatement",
	}, frag(s))
}

func openMasterKeyStatementToJSON(s *ast.OpenMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "OpenMasterKeyStatement",
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return addSpan(node, frag(s))
}

func openSymmetricKeyStatementToJSON(s *ast.OpenSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "OpenSymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.DecryptionMechanism != nil {
		node["DecryptionMechanism"] = cryptoMechanismToJSON(s.DecryptionMechanism)
	}
	return addSpan(node, frag(s))
}

func checkpointStatementToJSON(s *ast.CheckpointStatement) jsonNode {
	node := jsonNode{
		"$type": "CheckpointStatement",
	}
	if s.Duration != nil {
		node["Duration"] = scalarExpressionToJSON(s.Duration)
	}
	return addSpan(node, frag(s))
}

func reconfigureStatementToJSON(s *ast.ReconfigureStatement) jsonNode {
	return addSpan(jsonNode{
		"$type":        "ReconfigureStatement",
		"WithOverride": s.WithOverride,
	}, frag(s))
}

func shutdownStatementToJSON(s *ast.ShutdownStatement) jsonNode {
	return addSpan(jsonNode{
		"$type":      "ShutdownStatement",
		"WithNoWait": s.WithNoWait,
	}, frag(s))
}

func setUserStatementToJSON(s *ast.SetUserStatement) jsonNode {
	node := jsonNode{
		"$type":       "SetUserStatement",
		"WithNoReset": s.WithNoReset,
	}
	if s.UserName != nil {
		node["UserName"] = scalarExpressionToJSON(s.UserName)
	}
	return addSpan(node, frag(s))
}

func lineNoStatementToJSON(s *ast.LineNoStatement) jsonNode {
	node := jsonNode{
		"$type": "LineNoStatement",
	}
	if s.LineNo != nil {
		node["LineNo"] = scalarExpressionToJSON(s.LineNo)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	if c.Collation != nil {
		node["Collation"] = identifierToJSON(c.Collation)
	}
	return addSpan(node, frag(c))
}

func goToStatementToJSON(s *ast.GoToStatement) jsonNode {
	node := jsonNode{
		"$type": "GoToStatement",
	}
	if s.LabelName != nil {
		node["LabelName"] = identifierToJSON(s.LabelName)
	}
	return addSpan(node, frag(s))
}

func labelStatementToJSON(s *ast.LabelStatement) jsonNode {
	return addSpan(jsonNode{
		"$type": "LabelStatement",
		"Value": s.Value,
	}, frag(s))
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
	return addSpan(node, frag(s))
}

func createMasterKeyStatementToJSON(s *ast.CreateMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateMasterKeyStatement",
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
		node["Where"] = scalarExpressionToJSON(s.Where)
	}
	node["IsConversationGroupIdWhere"] = s.IsConversationGroupIdWhere
	return addSpan(node, frag(s))
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
	if v.Alias != nil {
		node["Alias"] = identifierToJSON(v.Alias)
	}
	node["ForPath"] = v.ForPath
	return addSpan(node, frag(v))
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
	return addSpan(node, frag(s))
}

func alterMasterKeyStatementToJSON(s *ast.AlterMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type":  "AlterMasterKeyStatement",
		"Option": s.Option,
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(a))
	case *ast.DropMemberAlterRoleAction:
		node := jsonNode{
			"$type": "DropMemberAlterRoleAction",
		}
		if action.Member != nil {
			node["Member"] = identifierToJSON(action.Member)
		}
		return addSpan(node, frag(a))
	case *ast.RenameAlterRoleAction:
		node := jsonNode{
			"$type": "RenameAlterRoleAction",
		}
		if action.NewName != nil {
			node["NewName"] = identifierToJSON(action.NewName)
		}
		return addSpan(node, frag(a))
	default:
		return addSpan(jsonNode{"$type": "UnknownAlterRoleAction"}, frag(a))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func createAvailabilityGroupStatementToJSON(s *ast.CreateAvailabilityGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateAvailabilityGroupStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = availabilityGroupOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	if len(s.Databases) > 0 {
		dbs := make([]jsonNode, len(s.Databases))
		for i, db := range s.Databases {
			dbs[i] = identifierToJSON(db)
		}
		node["Databases"] = dbs
	}
	if len(s.Replicas) > 0 {
		reps := make([]jsonNode, len(s.Replicas))
		for i, rep := range s.Replicas {
			reps[i] = availabilityReplicaToJSON(rep)
		}
		node["Replicas"] = reps
	}
	return addSpan(node, frag(s))
}

func availabilityGroupOptionToJSON(opt ast.AvailabilityGroupOption) jsonNode {
	switch o := opt.(type) {
	case *ast.LiteralAvailabilityGroupOption:
		node := jsonNode{
			"$type": "LiteralAvailabilityGroupOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		node["OptionKind"] = o.OptionKind
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownAvailabilityGroupOption"}, frag(opt))
	}
}

func alterAvailabilityGroupStatementToJSON(s *ast.AlterAvailabilityGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterAvailabilityGroupStatement",
	}
	if s.StatementType != "" {
		node["AlterAvailabilityGroupStatementType"] = s.StatementType
	}
	if s.Action != nil {
		node["Action"] = availabilityGroupActionToJSON(s.Action)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Databases) > 0 {
		dbs := make([]jsonNode, len(s.Databases))
		for i, db := range s.Databases {
			dbs[i] = identifierToJSON(db)
		}
		node["Databases"] = dbs
	}
	if len(s.Replicas) > 0 {
		reps := make([]jsonNode, len(s.Replicas))
		for i, rep := range s.Replicas {
			reps[i] = availabilityReplicaToJSON(rep)
		}
		node["Replicas"] = reps
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = availabilityGroupOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
}

func availabilityGroupActionToJSON(action ast.AvailabilityGroupAction) jsonNode {
	switch a := action.(type) {
	case *ast.AlterAvailabilityGroupAction:
		return addSpan(jsonNode{
			"$type":      "AlterAvailabilityGroupAction",
			"ActionType": a.ActionType,
		}, frag(action))
	case *ast.AlterAvailabilityGroupFailoverAction:
		node := jsonNode{
			"$type":      "AlterAvailabilityGroupFailoverAction",
			"ActionType": a.ActionType,
		}
		if len(a.Options) > 0 {
			opts := make([]jsonNode, len(a.Options))
			for i, opt := range a.Options {
				optNode := jsonNode{
					"$type":      "AlterAvailabilityGroupFailoverOption",
					"OptionKind": opt.OptionKind,
				}
				if opt.Value != nil {
					optNode["Value"] = scalarExpressionToJSON(opt.Value)
				}
				opts[i] = optNode
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(action))
	default:
		return addSpan(jsonNode{"$type": "UnknownAvailabilityGroupAction"}, frag(action))
	}
}

func availabilityReplicaToJSON(rep *ast.AvailabilityReplica) jsonNode {
	node := jsonNode{
		"$type": "AvailabilityReplica",
	}
	if rep.ServerName != nil {
		node["ServerName"] = stringLiteralToJSON(rep.ServerName)
	}
	if len(rep.Options) > 0 {
		opts := make([]jsonNode, len(rep.Options))
		for i, opt := range rep.Options {
			opts[i] = availabilityReplicaOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(rep))
}

func availabilityReplicaOptionToJSON(opt ast.AvailabilityReplicaOption) jsonNode {
	switch o := opt.(type) {
	case *ast.AvailabilityModeReplicaOption:
		return addSpan(jsonNode{
			"$type":      "AvailabilityModeReplicaOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.FailoverModeReplicaOption:
		return addSpan(jsonNode{
			"$type":      "FailoverModeReplicaOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.LiteralReplicaOption:
		node := jsonNode{
			"$type": "LiteralReplicaOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		node["OptionKind"] = o.OptionKind
		return addSpan(node, frag(opt))
	case *ast.PrimaryRoleReplicaOption:
		return addSpan(jsonNode{
			"$type":            "PrimaryRoleReplicaOption",
			"AllowConnections": o.AllowConnections,
			"OptionKind":       o.OptionKind,
		}, frag(opt))
	case *ast.SecondaryRoleReplicaOption:
		return addSpan(jsonNode{
			"$type":            "SecondaryRoleReplicaOption",
			"AllowConnections": o.AllowConnections,
			"OptionKind":       o.OptionKind,
		}, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownReplicaOption"}, frag(opt))
	}
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
	return addSpan(node, frag(s))
}

func alterServerAuditStatementToJSON(s *ast.AlterServerAuditStatement) jsonNode {
	node := jsonNode{
		"$type":       "AlterServerAuditStatement",
		"RemoveWhere": s.RemoveWhere,
	}
	if s.NewName != nil {
		node["NewName"] = identifierToJSON(s.NewName)
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
	return addSpan(node, frag(s))
}

func createServerAuditSpecificationStatementToJSON(s *ast.CreateServerAuditSpecificationStatement) jsonNode {
	node := jsonNode{
		"$type":      "CreateServerAuditSpecificationStatement",
		"AuditState": s.AuditState,
	}
	if len(s.Parts) > 0 {
		parts := make([]jsonNode, len(s.Parts))
		for i, p := range s.Parts {
			parts[i] = auditSpecificationPartToJSON(p)
		}
		node["Parts"] = parts
	}
	if s.SpecificationName != nil {
		node["SpecificationName"] = identifierToJSON(s.SpecificationName)
	}
	if s.AuditName != nil {
		node["AuditName"] = identifierToJSON(s.AuditName)
	}
	return addSpan(node, frag(s))
}

func alterServerAuditSpecificationStatementToJSON(s *ast.AlterServerAuditSpecificationStatement) jsonNode {
	node := jsonNode{
		"$type":      "AlterServerAuditSpecificationStatement",
		"AuditState": s.AuditState,
	}
	if len(s.Parts) > 0 {
		parts := make([]jsonNode, len(s.Parts))
		for i, p := range s.Parts {
			parts[i] = auditSpecificationPartToJSON(p)
		}
		node["Parts"] = parts
	}
	if s.SpecificationName != nil {
		node["SpecificationName"] = identifierToJSON(s.SpecificationName)
	}
	if s.AuditName != nil {
		node["AuditName"] = identifierToJSON(s.AuditName)
	}
	return addSpan(node, frag(s))
}

func createDatabaseAuditSpecificationStatementToJSON(s *ast.CreateDatabaseAuditSpecificationStatement) jsonNode {
	node := jsonNode{
		"$type":      "CreateDatabaseAuditSpecificationStatement",
		"AuditState": s.AuditState,
	}
	if len(s.Parts) > 0 {
		parts := make([]jsonNode, len(s.Parts))
		for i, p := range s.Parts {
			parts[i] = auditSpecificationPartToJSON(p)
		}
		node["Parts"] = parts
	}
	if s.SpecificationName != nil {
		node["SpecificationName"] = identifierToJSON(s.SpecificationName)
	}
	if s.AuditName != nil {
		node["AuditName"] = identifierToJSON(s.AuditName)
	}
	return addSpan(node, frag(s))
}

func alterDatabaseAuditSpecificationStatementToJSON(s *ast.AlterDatabaseAuditSpecificationStatement) jsonNode {
	node := jsonNode{
		"$type":      "AlterDatabaseAuditSpecificationStatement",
		"AuditState": s.AuditState,
	}
	if len(s.Parts) > 0 {
		parts := make([]jsonNode, len(s.Parts))
		for i, p := range s.Parts {
			parts[i] = auditSpecificationPartToJSON(p)
		}
		node["Parts"] = parts
	}
	if s.SpecificationName != nil {
		node["SpecificationName"] = identifierToJSON(s.SpecificationName)
	}
	if s.AuditName != nil {
		node["AuditName"] = identifierToJSON(s.AuditName)
	}
	return addSpan(node, frag(s))
}

func auditSpecificationPartToJSON(p *ast.AuditSpecificationPart) jsonNode {
	node := jsonNode{
		"$type":  "AuditSpecificationPart",
		"IsDrop": p.IsDrop,
	}
	if p.Details != nil {
		node["Details"] = auditSpecificationDetailToJSON(p.Details)
	}
	return addSpan(node, frag(p))
}

func auditSpecificationDetailToJSON(d ast.AuditSpecificationDetail) jsonNode {
	switch detail := d.(type) {
	case *ast.AuditActionGroupReference:
		return addSpan(jsonNode{
			"$type": "AuditActionGroupReference",
			"Group": detail.Group,
		}, frag(d))
	case *ast.AuditActionSpecification:
		node := jsonNode{
			"$type": "AuditActionSpecification",
		}
		if len(detail.Actions) > 0 {
			actions := make([]jsonNode, len(detail.Actions))
			for i, a := range detail.Actions {
				actions[i] = addSpan(jsonNode{
					"$type":      "DatabaseAuditAction",
					"ActionKind": a.ActionKind,
				}, frag(a))
			}
			node["Actions"] = actions
		}
		if len(detail.Principals) > 0 {
			principals := make([]jsonNode, len(detail.Principals))
			for i, p := range detail.Principals {
				principalNode := jsonNode{
					"$type":         "SecurityPrincipal",
					"PrincipalType": p.PrincipalType,
				}
				if p.Identifier != nil {
					principalNode["Identifier"] = identifierToJSON(p.Identifier)
				}
				principals[i] = addSpan(principalNode, frag(p))
			}
			node["Principals"] = principals
		}
		if detail.TargetObject != nil {
			node["TargetObject"] = securityTargetObjectToJSON(detail.TargetObject)
		}
		return addSpan(node, frag(d))
	default:
		return addSpan(jsonNode{}, frag(d))
	}
}

func dropServerAuditStatementToJSON(s *ast.DropServerAuditStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropServerAuditStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropServerAuditSpecificationStatementToJSON(s *ast.DropServerAuditSpecificationStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropServerAuditSpecificationStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropDatabaseAuditSpecificationStatementToJSON(s *ast.DropDatabaseAuditSpecificationStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropDatabaseAuditSpecificationStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(t))
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
		return addSpan(node, frag(o))
	case *ast.MaxSizeAuditTargetOption:
		node := jsonNode{
			"$type":       "MaxSizeAuditTargetOption",
			"IsUnlimited": opt.IsUnlimited,
			"Unit":        opt.Unit,
			"OptionKind":  opt.OptionKind,
		}
		if opt.Size != nil {
			node["Size"] = scalarExpressionToJSON(opt.Size)
		}
		return addSpan(node, frag(o))
	case *ast.MaxRolloverFilesAuditTargetOption:
		node := jsonNode{
			"$type":       "MaxRolloverFilesAuditTargetOption",
			"IsUnlimited": opt.IsUnlimited,
			"OptionKind":  opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return addSpan(node, frag(o))
	case *ast.OnOffAuditTargetOption:
		return addSpan(jsonNode{
			"$type":      "OnOffAuditTargetOption",
			"Value":      opt.Value,
			"OptionKind": opt.OptionKind,
		}, frag(o))
	case *ast.RetentionDaysAuditTargetOption:
		node := jsonNode{
			"$type":      "RetentionDaysAuditTargetOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Days != nil {
			node["Days"] = scalarExpressionToJSON(opt.Days)
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownAuditTargetOption"}, frag(o))
	}
}

func auditOptionToJSON(o ast.AuditOption) jsonNode {
	switch opt := o.(type) {
	case *ast.OnFailureAuditOption:
		return addSpan(jsonNode{
			"$type":           "OnFailureAuditOption",
			"OnFailureAction": opt.OnFailureAction,
			"OptionKind":      opt.OptionKind,
		}, frag(o))
	case *ast.QueueDelayAuditOption:
		node := jsonNode{
			"$type":      "QueueDelayAuditOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Delay != nil {
			node["Delay"] = scalarExpressionToJSON(opt.Delay)
		}
		return addSpan(node, frag(o))
	case *ast.StateAuditOption:
		return addSpan(jsonNode{
			"$type":      "StateAuditOption",
			"Value":      opt.Value,
			"OptionKind": opt.OptionKind,
		}, frag(o))
	case *ast.AuditGuidAuditOption:
		node := jsonNode{
			"$type":      "AuditGuidAuditOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Guid != nil {
			node["Guid"] = scalarExpressionToJSON(opt.Guid)
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownAuditOption"}, frag(o))
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(o))
	case *ast.OnOffRemoteServiceBindingOption:
		return addSpan(jsonNode{
			"$type":       "OnOffRemoteServiceBindingOption",
			"OptionState": opt.OptionState,
			"OptionKind":  opt.OptionKind,
		}, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownRemoteServiceBindingOption"}, frag(o))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func alterServerConfigurationSoftNumaOptionToJSON(o *ast.AlterServerConfigurationSoftNumaOption) jsonNode {
	node := jsonNode{
		"$type":      "AlterServerConfigurationSoftNumaOption",
		"OptionKind": o.OptionKind,
	}
	if o.OptionValue != nil {
		node["OptionValue"] = onOffOptionValueToJSON(o.OptionValue)
	}
	return addSpan(node, frag(o))
}

func onOffOptionValueToJSON(o *ast.OnOffOptionValue) jsonNode {
	return addSpan(jsonNode{
		"$type":       "OnOffOptionValue",
		"OptionState": o.OptionState,
	}, frag(o))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(o))
}

func alterServerConfigurationExternalAuthenticationOptionToJSON(o *ast.AlterServerConfigurationExternalAuthenticationOption) jsonNode {
	node := jsonNode{
		"$type":      "AlterServerConfigurationExternalAuthenticationOption",
		"OptionKind": o.OptionKind,
	}
	if o.OptionValue != nil {
		node["OptionValue"] = literalOptionValueToJSON(o.OptionValue)
	}
	return addSpan(node, frag(o))
}

func literalOptionValueToJSON(o *ast.LiteralOptionValue) jsonNode {
	node := jsonNode{
		"$type": "LiteralOptionValue",
	}
	if o.Value != nil {
		node["Value"] = scalarExpressionToJSON(o.Value)
	}
	return addSpan(node, frag(o))
}

func alterServerConfigurationSetDiagnosticsLogStatementToJSON(s *ast.AlterServerConfigurationSetDiagnosticsLogStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationSetDiagnosticsLogStatement",
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			switch opt := o.(type) {
			case *ast.AlterServerConfigurationDiagnosticsLogOption:
				optNode := jsonNode{
					"$type":      "AlterServerConfigurationDiagnosticsLogOption",
					"OptionKind": opt.OptionKind,
				}
				if opt.OptionValue != nil {
					switch v := opt.OptionValue.(type) {
					case *ast.OnOffOptionValue:
						optNode["OptionValue"] = onOffOptionValueToJSON(v)
					case *ast.LiteralOptionValue:
						optNode["OptionValue"] = literalOptionValueToJSON(v)
					}
				}
				options[i] = optNode
			case *ast.AlterServerConfigurationDiagnosticsLogMaxSizeOption:
				optNode := jsonNode{
					"$type":      "AlterServerConfigurationDiagnosticsLogMaxSizeOption",
					"SizeUnit":   opt.SizeUnit,
					"OptionKind": opt.OptionKind,
				}
				if opt.OptionValue != nil {
					optNode["OptionValue"] = literalOptionValueToJSON(opt.OptionValue)
				}
				options[i] = optNode
			}
		}
		node["Options"] = options
	}
	return addSpan(node, frag(s))
}

func alterServerConfigurationSetFailoverClusterPropertyStatementToJSON(s *ast.AlterServerConfigurationSetFailoverClusterPropertyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationSetFailoverClusterPropertyStatement",
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			optNode := jsonNode{
				"$type":      "AlterServerConfigurationFailoverClusterPropertyOption",
				"OptionKind": o.OptionKind,
			}
			if o.OptionValue != nil {
				optNode["OptionValue"] = literalOptionValueToJSON(o.OptionValue)
			}
			options[i] = optNode
		}
		node["Options"] = options
	}
	return addSpan(node, frag(s))
}

func alterServerConfigurationSetBufferPoolExtensionStatementToJSON(s *ast.AlterServerConfigurationSetBufferPoolExtensionStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationSetBufferPoolExtensionStatement",
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			optNode := jsonNode{
				"$type": "AlterServerConfigurationBufferPoolExtensionContainerOption",
			}
			if len(o.Suboptions) > 0 {
				suboptions := make([]jsonNode, len(o.Suboptions))
				for j, sub := range o.Suboptions {
					switch s := sub.(type) {
					case *ast.AlterServerConfigurationBufferPoolExtensionOption:
						subNode := jsonNode{
							"$type":      "AlterServerConfigurationBufferPoolExtensionOption",
							"OptionKind": s.OptionKind,
						}
						if s.OptionValue != nil {
							subNode["OptionValue"] = literalOptionValueToJSON(s.OptionValue)
						}
						suboptions[j] = subNode
					case *ast.AlterServerConfigurationBufferPoolExtensionSizeOption:
						subNode := jsonNode{
							"$type":      "AlterServerConfigurationBufferPoolExtensionSizeOption",
							"SizeUnit":   s.SizeUnit,
							"OptionKind": s.OptionKind,
						}
						if s.OptionValue != nil {
							subNode["OptionValue"] = literalOptionValueToJSON(s.OptionValue)
						}
						suboptions[j] = subNode
					}
				}
				optNode["Suboptions"] = suboptions
			}
			optNode["OptionKind"] = o.OptionKind
			if o.OptionValue != nil {
				optNode["OptionValue"] = onOffOptionValueToJSON(o.OptionValue)
			}
			options[i] = optNode
		}
		node["Options"] = options
	}
	return addSpan(node, frag(s))
}

func alterServerConfigurationSetHadrClusterStatementToJSON(s *ast.AlterServerConfigurationSetHadrClusterStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServerConfigurationSetHadrClusterStatement",
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			optNode := jsonNode{
				"$type":      "AlterServerConfigurationHadrClusterOption",
				"OptionKind": o.OptionKind,
			}
			if o.OptionValue != nil {
				optNode["OptionValue"] = literalOptionValueToJSON(o.OptionValue)
			}
			optNode["IsLocal"] = o.IsLocal
			options[i] = optNode
		}
		node["Options"] = options
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(r))
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
	return addSpan(node, frag(s))
}

func createProcedureStatementToJSON(s *ast.CreateProcedureStatement) jsonNode {
	node := jsonNode{
		"$type":            "CreateProcedureStatement",
		"IsForReplication": s.IsForReplication,
	}
	if s.ProcedureReference != nil {
		node["ProcedureReference"] = procedureReferenceToJSON(s.ProcedureReference)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			options[i] = procedureOptionToJSON(opt)
		}
		node["Options"] = options
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = procedureParameterToJSON(p)
		}
		node["Parameters"] = params
	}
	if s.MethodSpecifier != nil {
		node["MethodSpecifier"] = methodSpecifierToJSON(s.MethodSpecifier)
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return addSpan(node, frag(s))
}

func createOrAlterProcedureStatementToJSON(s *ast.CreateOrAlterProcedureStatement) jsonNode {
	node := jsonNode{
		"$type":            "CreateOrAlterProcedureStatement",
		"IsForReplication": s.IsForReplication,
	}
	if s.ProcedureReference != nil {
		node["ProcedureReference"] = procedureReferenceToJSON(s.ProcedureReference)
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			options[i] = procedureOptionToJSON(opt)
		}
		node["Options"] = options
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = procedureParameterToJSON(p)
		}
		node["Parameters"] = params
	}
	if s.MethodSpecifier != nil {
		node["MethodSpecifier"] = methodSpecifierToJSON(s.MethodSpecifier)
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return addSpan(node, frag(s))
}

func procedureOptionToJSON(opt ast.ProcedureOptionBase) jsonNode {
	switch o := opt.(type) {
	case *ast.ProcedureOption:
		return addSpan(jsonNode{
			"$type":      "ProcedureOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.ExecuteAsProcedureOption:
		node := jsonNode{
			"$type":      "ExecuteAsProcedureOption",
			"OptionKind": o.OptionKind,
		}
		if o.ExecuteAs != nil {
			node["ExecuteAs"] = executeAsClauseToJSON(o.ExecuteAs)
		}
		return addSpan(node, frag(opt))
	}
	return addSpan(jsonNode{}, frag(opt))
}

func executeAsClauseToJSON(e *ast.ExecuteAsClause) jsonNode {
	node := jsonNode{
		"$type":           "ExecuteAsClause",
		"ExecuteAsOption": e.ExecuteAsOption,
	}
	if e.Literal != nil {
		node["Literal"] = stringLiteralToJSON(e.Literal)
	}
	return addSpan(node, frag(e))
}

func methodSpecifierToJSON(m *ast.MethodSpecifier) jsonNode {
	node := jsonNode{
		"$type": "MethodSpecifier",
	}
	if m.AssemblyName != nil {
		node["AssemblyName"] = identifierToJSON(m.AssemblyName)
	}
	if m.ClassName != nil {
		node["ClassName"] = identifierToJSON(m.ClassName)
	}
	if m.MethodName != nil {
		node["MethodName"] = identifierToJSON(m.MethodName)
	}
	return addSpan(node, frag(m))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(p))
}

func nullableConstraintToJSON(n *ast.NullableConstraintDefinition) jsonNode {
	return addSpan(jsonNode{
		"$type":    "NullableConstraintDefinition",
		"Nullable": n.Nullable,
	}, frag(n))
}

// parseRestoreStatement parses a RESTORE statement
func (p *Parser) parseRestoreStatement() (ast.Statement, error) {
	astStart := p.curTok

	// Consume RESTORE
	p.nextToken()

	// Check for SERVICE MASTER KEY
	if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
		spanV9, spanErr9 := p.parseRestoreServiceMasterKeyStatement()
		return spanned(p, spanV9, astStart), spanErr9
	}

	// Check for MASTER KEY
	if strings.ToUpper(p.curTok.Literal) == "MASTER" {
		spanV10, spanErr10 := p.parseRestoreMasterKeyStatement()
		return spanned(p, spanV10, astStart), spanErr10
	}

	stmt := &ast.RestoreStatement{}
	hasDatabaseName := true

	// Parse restore kind (DATABASE, LOG, etc.)
	switch strings.ToUpper(p.curTok.Literal) {
	case "DATABASE":
		stmt.Kind = "Database"
		p.nextToken()
	case "LOG":
		stmt.Kind = "TransactionLog"
		p.nextToken()
	case "FILELISTONLY":
		stmt.Kind = "FileListOnly"
		p.nextToken()
		hasDatabaseName = false
	case "VERIFYONLY":
		stmt.Kind = "VerifyOnly"
		p.nextToken()
		hasDatabaseName = false
	case "LABELONLY":
		stmt.Kind = "LabelOnly"
		p.nextToken()
		hasDatabaseName = false
	case "REWINDONLY":
		stmt.Kind = "RewindOnly"
		p.nextToken()
		hasDatabaseName = false
	case "HEADERONLY":
		stmt.Kind = "HeaderOnly"
		p.nextToken()
		hasDatabaseName = false
	default:
		stmt.Kind = "Database"
	}

	// Parse database name (only for DATABASE and LOG kinds)
	if hasDatabaseName {
		dbName := &ast.IdentifierOrValueExpression{}
		if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
			// Variable reference
			varRef := p.spanVarRef(p.curTok.Literal)
			p.nextToken()
			dbName.Value = varRef.Name
			dbName.ValueExpression = varRef
		} else {
			ident := p.parseIdentifier()
			dbName.Value = ident.Value
			dbName.Identifier = ident
		}
		stmt.DatabaseName = dbName

		// Parse optional FILE = or FILEGROUP = before FROM
		for strings.ToUpper(p.curTok.Literal) == "FILE" || strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
			itemKind := "Files"
			if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
				itemKind = "FileGroups"
			}
			p.nextToken() // consume FILE/FILEGROUP
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after FILE/FILEGROUP, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =

			fileInfo := &ast.BackupRestoreFileInfo{ItemKind: itemKind}
			// Parse the file name
			var item ast.ScalarExpression
			if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
				// Strip surrounding quotes
				val := p.curTok.Literal
				if len(val) >= 2 && ((val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"')) {
					val = val[1 : len(val)-1]
				}
				item = p.strLit(val, p.curTok.Type == TokenNationalString)
				p.nextToken()
			} else if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				item = p.spanVarRef(p.curTok.Literal)
				p.nextToken()
			} else {
				ident := p.parseIdentifier()
				item = &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{ident},
						Count:       1,
					},
				}
			}
			fileInfo.Items = append(fileInfo.Items, item)
			stmt.Files = append(stmt.Files, fileInfo)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
	}
	// Check for optional FROM clause
	if strings.ToUpper(p.curTok.Literal) != "FROM" {
		// No FROM clause - check for WITH clause
		if p.curTok.Type == TokenWith {
			goto parseWithClause
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return spanned(p, stmt, astStart), nil
	}
	p.nextToken()

	// Parse devices
	for {
		device := &ast.DeviceInfo{DeviceType: "None"}

		// Check for device type
		switch strings.ToUpper(p.curTok.Literal) {
		case "TAPE":
			device.DeviceType = "Tape"
			p.nextToken()
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after TAPE, got %s", p.curTok.Literal)
			}
			p.nextToken()
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
		if device.DeviceType == "Disk" || device.DeviceType == "URL" || device.DeviceType == "Tape" {
			// For DISK, URL, and TAPE, use PhysicalDevice with the string literal directly
			if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
				// Strip the surrounding quotes from the literal
				val := p.curTok.Literal
				if len(val) >= 2 && ((val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"')) {
					val = val[1 : len(val)-1]
				}
				strLit := p.strLit(val, p.curTok.Type == TokenNationalString)
				device.PhysicalDevice = strLit
				p.nextToken()
			} else if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				varRef := p.spanVarRef(p.curTok.Literal)
				device.PhysicalDevice = varRef
				p.nextToken()
			}
		} else {
			// For other device types, use LogicalDevice
			deviceName := &ast.IdentifierOrValueExpression{}
			if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				varRef := p.spanVarRef(p.curTok.Literal)
				p.nextToken()
				deviceName.Value = varRef.Name
				deviceName.ValueExpression = varRef
			} else if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
				val := p.curTok.Literal
				if len(val) >= 2 && ((val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"')) {
					val = val[1 : len(val)-1]
				}
				strLit := p.strLit(val, p.curTok.Type == TokenNationalString)
				deviceName.Value = strLit.Value
				deviceName.ValueExpression = strLit
				p.nextToken()
			} else {
				ident := p.parseIdentifier()
				deviceName.Value = ident.Value
				deviceName.Identifier = ident
			}
			device.LogicalDevice = deviceName
		}
		stmt.Devices = append(stmt.Devices, device)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Parse WITH clause
parseWithClause:
	if p.curTok.Type == TokenWith {
		p.nextToken()

		for {
			restoreOptTok := p.curTok
			nOptsBefore := len(stmt.Options)
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

			case "MOVE":
				// MOVE 'logical_file_name' TO 'os_file_name'
				opt := &ast.MoveRestoreOption{OptionKind: "Move"}
				// Parse logical file name
				expr, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				opt.LogicalFileName = expr
				// Expect TO
				if strings.ToUpper(p.curTok.Literal) != "TO" {
					return nil, fmt.Errorf("expected TO after logical file name, got %s", p.curTok.Literal)
				}
				p.nextToken()
				// Parse OS file name
				osExpr, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				opt.OSFileName = osExpr
				stmt.Options = append(stmt.Options, opt)

			case "STOPATMARK", "STOPBEFOREMARK":
				opt := &ast.StopRestoreOption{
					OptionKind: "StopAt",
					IsStopAt:   optionName == "STOPATMARK",
				}
				if optionName == "STOPBEFOREMARK" {
					opt.OptionKind = "Stop"
				}
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					opt.Mark = expr
				}
				// Check for AFTER clause
				if strings.ToUpper(p.curTok.Literal) == "AFTER" {
					p.nextToken()
					afterExpr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					opt.After = afterExpr
				}
				stmt.Options = append(stmt.Options, opt)

			case "STANDBY":
				opt := &ast.ScalarExpressionRestoreOption{
					OptionKind: "Standby",
				}
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					opt.Value = expr
				}
				stmt.Options = append(stmt.Options, opt)

			case "FILE", "MEDIANAME", "MEDIAPASSWORD", "PASSWORD", "STOPAT":
				// Options that take a scalar expression value
				optKind := optionName
				switch optionName {
				case "MEDIANAME":
					optKind = "MediaName"
				case "MEDIAPASSWORD":
					optKind = "MediaPassword"
				case "PASSWORD":
					optKind = "Password"
				case "STOPAT":
					optKind = "StopAt"
				case "FILE":
					optKind = "File"
				}
				opt := &ast.ScalarExpressionRestoreOption{OptionKind: optKind}
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					opt.Value = expr
				}
				stmt.Options = append(stmt.Options, opt)

			case "STATS":
				// STATS can optionally have a value: STATS or STATS = 10
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					stmt.Options = append(stmt.Options, &ast.ScalarExpressionRestoreOption{
						OptionKind: "Stats",
						Value:      expr,
					})
				} else {
					stmt.Options = append(stmt.Options, &ast.SimpleRestoreOption{OptionKind: "Stats"})
				}

			case "ENABLE_BROKER", "ERROR_BROKER_CONVERSATIONS", "NEW_BROKER",
				"KEEP_REPLICATION", "RESTRICTED_USER",
				"KEEP_TEMPORAL_RETENTION", "NOREWIND", "NOUNLOAD",
				"RECOVERY", "NORECOVERY", "REPLACE", "RESTART", "REWIND",
				"UNLOAD", "CHECKSUM", "NO_CHECKSUM", "STOP_ON_ERROR",
				"CONTINUE_AFTER_ERROR":
				// Map option names to proper casing
				optKind := optionName
				switch optionName {
				case "ENABLE_BROKER":
					optKind = "EnableBroker"
				case "ERROR_BROKER_CONVERSATIONS":
					optKind = "ErrorBrokerConversations"
				case "NEW_BROKER":
					optKind = "NewBroker"
				case "KEEP_REPLICATION":
					optKind = "KeepReplication"
				case "RESTRICTED_USER":
					optKind = "RestrictedUser"
				case "KEEP_TEMPORAL_RETENTION":
					optKind = "KeepTemporalRetention"
				case "NOREWIND":
					optKind = "NoRewind"
				case "NOUNLOAD":
					optKind = "NoUnload"
				case "RECOVERY":
					optKind = "Recovery"
				case "NORECOVERY":
					optKind = "NoRecovery"
				case "REPLACE":
					optKind = "Replace"
				case "RESTART":
					optKind = "Restart"
				case "REWIND":
					optKind = "Rewind"
				case "UNLOAD":
					optKind = "Unload"
				case "CHECKSUM":
					optKind = "Checksum"
				case "NO_CHECKSUM":
					optKind = "NoChecksum"
				case "STOP_ON_ERROR":
					optKind = "StopOnError"
				case "CONTINUE_AFTER_ERROR":
					optKind = "ContinueAfterError"
				}
				stmt.Options = append(stmt.Options, &ast.SimpleRestoreOption{OptionKind: optKind})

			default:
				// Generic option with optional value
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

			for _, o := range stmt.Options[nOptsBefore:] {
				switch o.(type) {
				case *ast.StopRestoreOption, *ast.MoveRestoreOption, *ast.ScalarExpressionRestoreOption:
					// ScriptDom spans these options over their value
					// expressions only (derived from children).
					continue
				}
				if s, ok := any(o).(spannable); ok && !s.Frag().HasSpan() {
					p.spanFrom(restoreOptTok, s)
				}
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

	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseRestoreServiceMasterKeyStatement() (*ast.RestoreServiceMasterKeyStatement, error) {
	astStart := p.curTok

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

	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseRestoreMasterKeyStatement() (*ast.RestoreMasterKeyStatement, error) {
	astStart := p.curTok

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

	return spanned(p, stmt, astStart), nil
}

// parseCreateUserStatement parses a CREATE USER statement
func (p *Parser) parseCreateUserStatement() (*ast.CreateUserStatement, error) {
	astStart := p.curTok

	// Consume USER
	p.nextToken()

	stmt := &ast.CreateUserStatement{}

	// Parse user name
	stmt.Name = p.parseIdentifier()

	// Check for login option
	if strings.ToUpper(p.curTok.Literal) == "FOR" || strings.ToUpper(p.curTok.Literal) == "FROM" {
		p.nextToken()

		loginOption := &ast.UserLoginOption{}

		switch strings.ToUpper(p.curTok.Literal) {
		case "LOGIN":
			loginOption.UserLoginOptionType = "Login"
			p.nextToken()
			loginOption.Identifier = p.parseIdentifier()
		case "CERTIFICATE":
			loginOption.UserLoginOptionType = "Certificate"
			p.nextToken()
			loginOption.Identifier = p.parseIdentifier()
		case "ASYMMETRIC":
			p.nextToken() // consume ASYMMETRIC
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			loginOption.UserLoginOptionType = "AsymmetricKey"
			loginOption.Identifier = p.parseIdentifier()
		case "EXTERNAL":
			p.nextToken() // consume EXTERNAL
			if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
				p.nextToken() // consume PROVIDER
			}
			loginOption.UserLoginOptionType = "External"
		}
		// ScriptDom positions the login option on its identifier.
		if loginOption.Identifier != nil && loginOption.Identifier.Frag().HasSpan() {
			*loginOption.Frag() = *loginOption.Identifier.Frag()
		}
		stmt.UserLoginOption = loginOption
	} else if strings.ToUpper(p.curTok.Literal) == "WITHOUT" {
		p.nextToken() // consume WITHOUT
		wlOpt := &ast.UserLoginOption{
			UserLoginOptionType: "WithoutLogin",
		}
		if p.curTok.Type == TokenLogin {
			// ScriptDom positions WITHOUT LOGIN on the LOGIN keyword.
			p.tokSpan(wlOpt, p.curTok)
			p.nextToken() // consume LOGIN
		}
		stmt.UserLoginOption = wlOpt
	}

	// Parse WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken()

		for {
			optTok := p.curTok
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
			switch o := opt.(type) {
			case *ast.IdentifierPrincipalOption:
				// ScriptDom spans this option on its identifier value.
				p.spanFromChild(o, o.Identifier)
			case spannable:
				p.spanFrom(optTok, o)
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

	return spanned(p, stmt, astStart), nil
}

// parseCreateAggregateStatement parses a CREATE AGGREGATE statement
func (p *Parser) parseCreateAggregateStatement() (*ast.CreateAggregateStatement, error) {
	astStart := p.curTok

	// Consume AGGREGATE
	p.nextToken()

	stmt := &ast.CreateAggregateStatement{}

	// Parse aggregate name
	name, _ := p.parseSchemaObjectName()
	stmt.Name = name

	// Check for ( (optional for lenient parsing)
	if p.curTok.Type != TokenLParen {
		p.skipToEndOfStatement()
		return spanned(p, stmt, astStart), nil
	}
	p.nextToken()

	// Parse parameters
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		paramStart := p.curTok
		param := &ast.ProcedureParameter{
			IsVarying: false,
			Modifier:  "None",
		}

		// Parse parameter name
		if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
			param.VariableName = p.spanIdent(p.curTok.Literal, "NotQuoted")
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

		// Check for NULL or NOT NULL
		if p.curTok.Type == TokenNull {
			param.Nullable = &ast.NullableConstraintDefinition{Nullable: true}
			p.tokSpan(param.Nullable, p.curTok)
			p.nextToken()
		} else if p.curTok.Type == TokenNot {
			notTok := p.curTok
			p.nextToken() // consume NOT
			if p.curTok.Type == TokenNull {
				param.Nullable = &ast.NullableConstraintDefinition{Nullable: false}
				p.spanTokens(param.Nullable, notTok, p.curTok)
				p.nextToken()
			}
		}

		p.spanFrom(paramStart, param)
		p.spanFrom(paramStart, param)
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

	return spanned(p, stmt, astStart), nil
}

// parseCreateColumnStoreIndexStatement parses a CREATE COLUMNSTORE INDEX statement
func (p *Parser) parseCreateColumnStoreIndexStatement() (*ast.CreateColumnStoreIndexStatement, error) {
	astStart := p.curTok

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
			// Check for graph pseudo-columns
			upperLit := strings.ToUpper(p.curTok.Literal)
			var colRef *ast.ColumnReferenceExpression
			if upperLit == "$NODE_ID" {
				colRef = &ast.ColumnReferenceExpression{
					ColumnType: "PseudoColumnGraphNodeId",
				}
				p.tokSpan(colRef, p.curTok)
				p.nextToken()
			} else if upperLit == "$EDGE_ID" {
				colRef = &ast.ColumnReferenceExpression{
					ColumnType: "PseudoColumnGraphEdgeId",
				}
				p.tokSpan(colRef, p.curTok)
				p.nextToken()
			} else if upperLit == "$FROM_ID" {
				colRef = &ast.ColumnReferenceExpression{
					ColumnType: "PseudoColumnGraphFromId",
				}
				p.tokSpan(colRef, p.curTok)
				p.nextToken()
			} else if upperLit == "$TO_ID" {
				colRef = &ast.ColumnReferenceExpression{
					ColumnType: "PseudoColumnGraphToId",
				}
				p.tokSpan(colRef, p.curTok)
				p.nextToken()
			} else {
				colRef = &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{p.parseIdentifier()},
					},
				}
				colRef.MultiPartIdentifier.Count = len(colRef.MultiPartIdentifier.Identifiers)
			}
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

				optStart := p.curTok
				lenBefore := len(stmt.IndexOptions)
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

				case "SORT_IN_TEMPDB", "DROP_EXISTING":
					optKind := "SortInTempDB"
					if optName == "DROP_EXISTING" {
						optKind = "DropExisting"
					}
					p.nextToken() // consume option name
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
						OptionKind:  optKind,
						OptionState: state,
					})

				case "MAXDOP":
					p.nextToken() // consume MAXDOP
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					stmt.IndexOptions = append(stmt.IndexOptions, &ast.IndexExpressionOption{
						OptionKind: "MaxDop",
						Expression: expr,
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

				case "DATA_COMPRESSION":
					p.nextToken() // consume DATA_COMPRESSION
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					level := strings.ToUpper(p.curTok.Literal)
					compressionLevel := "None"
					switch level {
					case "COLUMNSTORE":
						compressionLevel = "ColumnStore"
					case "COLUMNSTORE_ARCHIVE":
						compressionLevel = "ColumnStoreArchive"
					case "PAGE":
						compressionLevel = "Page"
					case "ROW":
						compressionLevel = "Row"
					case "NONE":
						compressionLevel = "None"
					}
					p.nextToken() // consume compression level
					opt := &ast.DataCompressionOption{
						CompressionLevel: compressionLevel,
						OptionKind:       "DataCompression",
					}
					// Check for optional ON PARTITIONS(range)
					if p.curTok.Type == TokenOn {
						p.nextToken() // consume ON
						if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
							p.nextToken() // consume PARTITIONS
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									partRange := &ast.CompressionPartitionRange{}
									partRange.From = p.intLitFromToken(p.curTok)
									p.nextToken()
									if strings.ToUpper(p.curTok.Literal) == "TO" {
										p.nextToken() // consume TO
										partRange.To = p.intLitFromToken(p.curTok)
										p.nextToken()
									}
									opt.PartitionRanges = append(opt.PartitionRanges, partRange)
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
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)

				case "ONLINE":
					p.nextToken() // consume ONLINE
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					valueStr := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					onlineOpt := &ast.OnlineIndexOption{
						OptionKind:  "Online",
						OptionState: "On",
					}
					if valueStr == "OFF" {
						onlineOpt.OptionState = "Off"
					}
					// Check for optional (WAIT_AT_LOW_PRIORITY (...))
					if valueStr == "ON" && p.curTok.Type == TokenLParen {
						p.nextToken() // consume (
						if strings.ToUpper(p.curTok.Literal) == "WAIT_AT_LOW_PRIORITY" {
							waitLpTok := p.curTok
							p.nextToken() // consume WAIT_AT_LOW_PRIORITY
							lowPriorityOpt := &ast.OnlineIndexLowPriorityLockWaitOption{}
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									lpSubTok := p.curTok
									subOptName := strings.ToUpper(p.curTok.Literal)
									if subOptName == "MAX_DURATION" {
										p.nextToken() // consume MAX_DURATION
										if p.curTok.Type == TokenEquals {
											p.nextToken() // consume =
										}
										durVal, _ := p.parsePrimaryExpression()
										unit := ""
										if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
											unit = "Minutes"
											p.nextToken()
										} else if strings.ToUpper(p.curTok.Literal) == "SECONDS" {
											unit = "Seconds"
											p.nextToken()
										}
										maxDurOpt := &ast.LowPriorityLockWaitMaxDurationOption{
											MaxDuration: durVal,
											Unit:        unit,
											OptionKind:  "MaxDuration",
										}
										p.spanFrom(lpSubTok, maxDurOpt)
										lowPriorityOpt.Options = append(lowPriorityOpt.Options, maxDurOpt)
									} else if subOptName == "ABORT_AFTER_WAIT" {
										p.nextToken() // consume ABORT_AFTER_WAIT
										if p.curTok.Type == TokenEquals {
											p.nextToken() // consume =
										}
										abortType := "None"
										switch strings.ToUpper(p.curTok.Literal) {
										case "NONE":
											abortType = "None"
										case "SELF":
											abortType = "Self"
										case "BLOCKERS":
											abortType = "Blockers"
										}
										p.nextToken()
										abortOpt := &ast.LowPriorityLockWaitAbortAfterWaitOption{
											AbortAfterWait: abortType,
											OptionKind:     "AbortAfterWait",
										}
										p.spanFrom(lpSubTok, abortOpt)
										lowPriorityOpt.Options = append(lowPriorityOpt.Options, abortOpt)
									} else {
										break
									}
									if p.curTok.Type == TokenComma {
										p.nextToken()
									}
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken() // consume ) for WAIT_AT_LOW_PRIORITY options
								}
							}
							p.spanFrom(waitLpTok, lowPriorityOpt)
							onlineOpt.LowPriorityLockWaitOption = lowPriorityOpt
						}
						if p.curTok.Type == TokenRParen {
							p.nextToken() // consume ) for ONLINE option
						}
					}
					stmt.IndexOptions = append(stmt.IndexOptions, onlineOpt)

				default:
					// Skip unknown options
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken()
						p.nextToken() // skip value
					}
				}
				if len(stmt.IndexOptions) > lenBefore {
					last := stmt.IndexOptions[len(stmt.IndexOptions)-1]
					if o, ok := last.(spannable); ok {
						p.spanFrom(optStart, o)
					}
					if eo, ok := last.(*ast.IndexExpressionOption); ok && eo.OptionKind == "BucketCount" {
						p.spanFromChild(eo, eo.Expression)
					}
					// ScriptDom excludes the trailing time unit from COMPRESSION_DELAY.
					if co, ok := last.(*ast.CompressionDelayIndexOption); ok {
						if c, ok2 := any(co.Expression).(spannable); ok2 && c.Frag().HasSpan() && co.Frag().HasSpan() {
							co.Frag().FragmentLength = c.Frag().EndOffset() - co.Frag().StartOffset
						}
					}
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken()
			}
		}
	}

	// Parse optional ON filegroup/partition scheme
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fgStart := p.curTok
		fgps := &ast.FileGroupOrPartitionScheme{
			Name: &ast.IdentifierOrValueExpression{
				Identifier: p.parseIdentifier(),
			},
		}
		fgps.Name.Value = fgps.Name.Identifier.Value
		// Check for partition columns
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				fgps.PartitionSchemeColumns = append(fgps.PartitionSchemeColumns, p.parseIdentifier())
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
		p.spanFrom(fgStart, fgps)
		stmt.OnFileGroupOrPartitionScheme = fgps
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseAlterFunctionStatement parses an ALTER FUNCTION statement
func (p *Parser) parseAlterFunctionStatement() (*ast.AlterFunctionStatement, error) {
	astStart := p.curTok

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
			paramStart := p.curTok
			param := &ast.ProcedureParameter{
				IsVarying: false,
				Modifier:  "None",
			}

			// Parse parameter name
			if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				param.VariableName = p.spanIdent(p.curTok.Literal, "NotQuoted")
				p.nextToken()
			}

			// Skip optional AS keyword (e.g., @param AS int)
			if p.curTok.Type == TokenAs {
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

			p.spanFrom(paramStart, param)
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
				p.tokSpan(opt, p.curTok)
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

		// Parse optional WITH clause for function options
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			for {
				upperOpt := strings.ToUpper(p.curTok.Literal)
				switch upperOpt {
				case "INLINE":
					inlineTok := p.curTok
					p.nextToken() // consume INLINE
					// Expect = ON|OFF
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					optState := strings.ToUpper(p.curTok.Literal)
					state := "On"
					if optState == "OFF" {
						state = "Off"
					}
					p.nextToken() // consume ON/OFF
					inlineOpt := &ast.InlineFunctionOption{
						OptionKind:  "Inline",
						OptionState: state,
					}
					p.spanFrom(inlineTok, inlineOpt)
					stmt.Options = append(stmt.Options, inlineOpt)
				case "ENCRYPTION", "SCHEMABINDING", "NATIVE_COMPILATION", "CALLED":
					optNameTok2 := p.curTok
					optKind := capitalizeFirst(strings.ToLower(p.curTok.Literal))
					p.nextToken()
					// Handle CALLED ON NULL INPUT
					var inputTok Token
					if optKind == "Called" {
						for strings.ToUpper(p.curTok.Literal) == "ON" || strings.ToUpper(p.curTok.Literal) == "NULL" || strings.ToUpper(p.curTok.Literal) == "INPUT" {
							if strings.ToUpper(p.curTok.Literal) == "INPUT" {
								inputTok = p.curTok
							}
							p.nextToken()
						}
						optKind = "CalledOnNullInput"
					}
					fo := &ast.FunctionOption{OptionKind: optKind}
					if inputTok.Literal != "" {
						p.tokSpan(fo, inputTok)
					} else {
						p.tokSpan(fo, optNameTok2)
					}
					stmt.Options = append(stmt.Options, fo)
				case "RETURNS":
					// Handle RETURNS NULL ON NULL INPUT
					var inputTok Token
					for strings.ToUpper(p.curTok.Literal) == "RETURNS" || strings.ToUpper(p.curTok.Literal) == "NULL" || strings.ToUpper(p.curTok.Literal) == "ON" || strings.ToUpper(p.curTok.Literal) == "INPUT" {
						if strings.ToUpper(p.curTok.Literal) == "INPUT" {
							inputTok = p.curTok
						}
						p.nextToken()
					}
					fo := &ast.FunctionOption{OptionKind: "ReturnsNullOnNullInput"}
					p.tokSpan(fo, inputTok)
					stmt.Options = append(stmt.Options, fo)
				default:
					// Unknown option - skip it
					if p.curTok.Type == TokenIdent {
						p.nextToken()
					}
				}

				if p.curTok.Type == TokenComma {
					p.nextToken() // consume comma
				} else {
					break
				}
			}
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
		// A function body's final statement excludes its trailing semicolon.
		if n := len(stmtList.Statements); n > 0 {
			if s, ok := any(stmtList.Statements[n-1]).(spannable); ok {
				p.trimTrailingSemicolon(s)
				s.Frag().Pin()
			}
		}
		spanStatementList(stmtList)
		stmt.StatementList = stmtList
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseFunctionStatementList parses the body of a function
func (p *Parser) parseFunctionStatementList() (*ast.StatementList, error) {
	astStart := p.curTok

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

	return spanned(p, stmtList, astStart), nil
}

// parseAlterTriggerStatement parses an ALTER TRIGGER statement
func (p *Parser) parseAlterTriggerStatement() (*ast.AlterTriggerStatement, error) {
	astStart := p.curTok

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
	triggerObjTok := p.curTok
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
	// ScriptDom spans the trigger object over the ON target tokens.
	p.spanFrom(triggerObjTok, triggerObject)
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
		p.tokSpan(action, p.curTok)
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
				p.tokSpan(action.EventTypeGroup, p.curTok)
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
	spanStatementList(stmtList)
	stmt.StatementList = stmtList

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseAlterIndexStatement() (*ast.AlterIndexStatement, error) {
	astStart := p.curTok

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
		// Parse SET options (SET (...))
		if p.curTok.Type == TokenLParen {
			p.nextToken()
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optStart := p.curTok
				lenBefore := len(stmt.IndexOptions)
				optionName := strings.ToUpper(p.curTok.Literal)
				p.nextToken()

				if p.curTok.Type == TokenEquals {
					p.nextToken()
					valueStr := p.curTok.Literal
					valueTok := p.curTok
					valueUpper := strings.ToUpper(valueStr)
					p.nextToken()

					if optionName == "COMPRESSION_DELAY" {
						// Parse COMPRESSION_DELAY = value [MINUTE|MINUTES]
						timeUnit := "Unitless"
						nextUpper := strings.ToUpper(p.curTok.Literal)
						if nextUpper == "MINUTE" {
							timeUnit = "Minute"
							p.nextToken()
						} else if nextUpper == "MINUTES" {
							timeUnit = "Minutes"
							p.nextToken()
						}
						opt := &ast.CompressionDelayIndexOption{
							OptionKind: "CompressionDelay",
							Expression: p.intLitFromToken(valueTok),
							TimeUnit:   timeUnit,
						}
						stmt.IndexOptions = append(stmt.IndexOptions, opt)
					} else if valueUpper == "ON" || valueUpper == "OFF" {
						if optionName == "IGNORE_DUP_KEY" {
							opt := &ast.IgnoreDupKeyIndexOption{
								OptionKind:  "IgnoreDupKey",
								OptionState: p.capitalizeFirst(strings.ToLower(valueUpper)),
							}
							// Check for (SUPPRESS_MESSAGES = ON/OFF)
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								if strings.ToUpper(p.curTok.Literal) == "SUPPRESS_MESSAGES" {
									p.nextToken() // consume SUPPRESS_MESSAGES
									if p.curTok.Type == TokenEquals {
										p.nextToken() // consume =
									}
									suppressVal := strings.ToUpper(p.curTok.Literal)
									if suppressVal == "ON" {
										opt.SuppressMessagesOption = boolPtr(true)
									} else if suppressVal == "OFF" {
										opt.SuppressMessagesOption = boolPtr(false)
									}
									p.nextToken()
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken() // consume )
								}
							}
							stmt.IndexOptions = append(stmt.IndexOptions, opt)
						} else {
							opt := &ast.IndexStateOption{
								OptionKind:  p.getIndexOptionKind(optionName),
								OptionState: p.capitalizeFirst(strings.ToLower(valueUpper)),
							}
							stmt.IndexOptions = append(stmt.IndexOptions, opt)
						}
					} else {
						opt := &ast.IndexExpressionOption{
							OptionKind: p.getIndexOptionKind(optionName),
							Expression: p.intLitFromToken(valueTok),
						}
						stmt.IndexOptions = append(stmt.IndexOptions, opt)
					}
				}

				if len(stmt.IndexOptions) > lenBefore {
					last := stmt.IndexOptions[len(stmt.IndexOptions)-1]
					if o, ok := last.(spannable); ok {
						p.spanFrom(optStart, o)
					}
					if eo, ok := last.(*ast.IndexExpressionOption); ok && eo.OptionKind == "BucketCount" {
						p.spanFromChild(eo, eo.Expression)
					}
					// ScriptDom excludes the trailing time unit from COMPRESSION_DELAY.
					if co, ok := last.(*ast.CompressionDelayIndexOption); ok {
						if c, ok2 := any(co.Expression).(spannable); ok2 && c.Frag().HasSpan() && co.Frag().HasSpan() {
							co.Frag().FragmentLength = c.Frag().EndOffset() - co.Frag().StartOffset
						}
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
		partValTok := p.curTok
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
		p.spanFrom(partValTok, stmt.Partition)
	}

	// Parse WITH clause if present
	if p.curTok.Type == TokenWith {
		p.nextToken()
		// Check for XMLNAMESPACES
		if strings.ToUpper(p.curTok.Literal) == "XMLNAMESPACES" {
			stmt.XmlNamespaces = p.parseXmlNamespaces()
		} else if p.curTok.Type == TokenLParen {
			p.nextToken()

			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optStart := p.curTok
				lenBefore := len(stmt.IndexOptions)
				optionName := strings.ToUpper(p.curTok.Literal)
				p.nextToken()

				// Handle WAIT_AT_LOW_PRIORITY (...) - no equals sign
				if optionName == "WAIT_AT_LOW_PRIORITY" && p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					waitOpt := &ast.WaitAtLowPriorityOption{
						OptionKind: "WaitAtLowPriority",
					}
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
						lpSubTok := p.curTok
						subOptName := strings.ToUpper(p.curTok.Literal)
						if subOptName == "MAX_DURATION" {
							p.nextToken() // consume MAX_DURATION
							if p.curTok.Type == TokenEquals {
								p.nextToken() // consume =
							}
							durVal, _ := p.parsePrimaryExpression()
							unit := ""
							if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
								unit = "Minutes"
								p.nextToken()
							}
							maxDurOpt := &ast.LowPriorityLockWaitMaxDurationOption{
								MaxDuration: durVal,
								Unit:        unit,
								OptionKind:  "MaxDuration",
							}
							p.spanFrom(lpSubTok, maxDurOpt)
							waitOpt.Options = append(waitOpt.Options, maxDurOpt)
						} else if subOptName == "ABORT_AFTER_WAIT" {
							p.nextToken() // consume ABORT_AFTER_WAIT
							if p.curTok.Type == TokenEquals {
								p.nextToken() // consume =
							}
							abortType := "None"
							switch strings.ToUpper(p.curTok.Literal) {
							case "NONE":
								abortType = "None"
							case "SELF":
								abortType = "Self"
							case "BLOCKERS":
								abortType = "Blockers"
							}
							p.nextToken()
							abortOpt := &ast.LowPriorityLockWaitAbortAfterWaitOption{
								AbortAfterWait: abortType,
								OptionKind:     "AbortAfterWait",
							}
							p.spanFrom(lpSubTok, abortOpt)
							waitOpt.Options = append(waitOpt.Options, abortOpt)
						} else {
							break
						}
						if p.curTok.Type == TokenComma {
							p.nextToken()
						}
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
					stmt.IndexOptions = append(stmt.IndexOptions, waitOpt)
				} else if p.curTok.Type == TokenEquals {
					p.nextToken()
					valueStr := strings.ToUpper(p.curTok.Literal)
					valueTok := p.curTok
					p.nextToken()

					// Handle MAX_DURATION = value [MINUTES] as top-level option
					if optionName == "MAX_DURATION" {
						unit := ""
						if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
							unit = "Minutes"
							p.nextToken()
						}
						opt := &ast.MaxDurationOption{
							MaxDuration: p.intLitFromToken(valueTok),
							Unit:        unit,
							OptionKind:  "MaxDuration",
						}
						stmt.IndexOptions = append(stmt.IndexOptions, opt)
					} else if valueStr == "ON" || valueStr == "OFF" {
						// Determine if it's a state option (ON/OFF) or expression option
						if optionName == "IGNORE_DUP_KEY" {
							opt := &ast.IgnoreDupKeyIndexOption{
								OptionKind:  "IgnoreDupKey",
								OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
							}
							stmt.IndexOptions = append(stmt.IndexOptions, opt)
						} else if optionName == "ONLINE" {
							// Handle ONLINE = ON (WAIT_AT_LOW_PRIORITY (...))
							onlineOpt := &ast.OnlineIndexOption{
								OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
								OptionKind:  "Online",
							}
							// Check for optional (WAIT_AT_LOW_PRIORITY (...))
							if valueStr == "ON" && p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								if strings.ToUpper(p.curTok.Literal) == "WAIT_AT_LOW_PRIORITY" {
									waitLpTok := p.curTok
									p.nextToken() // consume WAIT_AT_LOW_PRIORITY
									lowPriorityOpt := &ast.OnlineIndexLowPriorityLockWaitOption{}
									if p.curTok.Type == TokenLParen {
										p.nextToken() // consume (
										for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
											lpSubTok := p.curTok
											subOptName := strings.ToUpper(p.curTok.Literal)
											if subOptName == "MAX_DURATION" {
												p.nextToken() // consume MAX_DURATION
												if p.curTok.Type == TokenEquals {
													p.nextToken() // consume =
												}
												durVal, _ := p.parsePrimaryExpression()
												unit := ""
												if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
													unit = "Minutes"
													p.nextToken()
												} else if strings.ToUpper(p.curTok.Literal) == "SECONDS" {
													unit = "Seconds"
													p.nextToken()
												}
												maxDurOpt := &ast.LowPriorityLockWaitMaxDurationOption{
													MaxDuration: durVal,
													Unit:        unit,
													OptionKind:  "MaxDuration",
												}
												p.spanFrom(lpSubTok, maxDurOpt)
												lowPriorityOpt.Options = append(lowPriorityOpt.Options, maxDurOpt)
											} else if subOptName == "ABORT_AFTER_WAIT" {
												p.nextToken() // consume ABORT_AFTER_WAIT
												if p.curTok.Type == TokenEquals {
													p.nextToken() // consume =
												}
												abortType := "None"
												switch strings.ToUpper(p.curTok.Literal) {
												case "NONE":
													abortType = "None"
												case "SELF":
													abortType = "Self"
												case "BLOCKERS":
													abortType = "Blockers"
												}
												p.nextToken()
												abortOpt := &ast.LowPriorityLockWaitAbortAfterWaitOption{
													AbortAfterWait: abortType,
													OptionKind:     "AbortAfterWait",
												}
												p.spanFrom(lpSubTok, abortOpt)
												lowPriorityOpt.Options = append(lowPriorityOpt.Options, abortOpt)
											} else {
												break
											}
											if p.curTok.Type == TokenComma {
												p.nextToken()
											}
										}
										if p.curTok.Type == TokenRParen {
											p.nextToken() // consume ) for WAIT_AT_LOW_PRIORITY options
										}
									}
									p.spanFrom(waitLpTok, lowPriorityOpt)
									onlineOpt.LowPriorityLockWaitOption = lowPriorityOpt
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken() // consume ) for ONLINE option
								}
							}
							stmt.IndexOptions = append(stmt.IndexOptions, onlineOpt)
						} else {
							opt := &ast.IndexStateOption{
								OptionKind:  p.getIndexOptionKind(optionName),
								OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
							}
							stmt.IndexOptions = append(stmt.IndexOptions, opt)
						}
					} else if optionName == "DATA_COMPRESSION" {
						// Handle DATA_COMPRESSION = level [ON PARTITIONS (...)]
						compressionLevel := "None"
						switch valueStr {
						case "COLUMNSTORE":
							compressionLevel = "ColumnStore"
						case "COLUMNSTORE_ARCHIVE":
							compressionLevel = "ColumnStoreArchive"
						case "PAGE":
							compressionLevel = "Page"
						case "ROW":
							compressionLevel = "Row"
						case "NONE":
							compressionLevel = "None"
						}
						opt := &ast.DataCompressionOption{
							CompressionLevel: compressionLevel,
							OptionKind:       "DataCompression",
						}
						// Check for optional ON PARTITIONS(range)
						if p.curTok.Type == TokenOn {
							p.nextToken() // consume ON
							if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
								p.nextToken() // consume PARTITIONS
								if p.curTok.Type == TokenLParen {
									p.nextToken() // consume (
									for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
										partRange := &ast.CompressionPartitionRange{}
										partRange.From = p.intLitFromToken(p.curTok)
										p.nextToken()
										if strings.ToUpper(p.curTok.Literal) == "TO" {
											p.nextToken() // consume TO
											partRange.To = p.intLitFromToken(p.curTok)
											p.nextToken()
										}
										opt.PartitionRanges = append(opt.PartitionRanges, partRange)
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
						}
						stmt.IndexOptions = append(stmt.IndexOptions, opt)
					} else {
						// Expression option like FILLFACTOR = 80
						opt := &ast.IndexExpressionOption{
							OptionKind: p.getIndexOptionKind(optionName),
							Expression: p.intLitFromToken(valueTok),
						}
						stmt.IndexOptions = append(stmt.IndexOptions, opt)
					}
				}

				if len(stmt.IndexOptions) > lenBefore {
					last := stmt.IndexOptions[len(stmt.IndexOptions)-1]
					if o, ok := last.(spannable); ok {
						p.spanFrom(optStart, o)
					}
					if eo, ok := last.(*ast.IndexExpressionOption); ok && eo.OptionKind == "BucketCount" {
						p.spanFromChild(eo, eo.Expression)
					}
					// ScriptDom excludes the trailing time unit from COMPRESSION_DELAY.
					if co, ok := last.(*ast.CompressionDelayIndexOption); ok {
						if c, ok2 := any(co.Expression).(spannable); ok2 && c.Frag().HasSpan() && co.Frag().HasSpan() {
							co.Frag().FragmentLength = c.Frag().EndOffset() - co.Frag().StartOffset
						}
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
	}

	// Parse FOR clause (selective XML index paths)
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		stmt.AlterIndexType = "UpdateSelectiveXmlPaths"
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				path := &ast.SelectiveXmlIndexPromotedPath{}
				actionWord := strings.ToUpper(p.curTok.Literal)
				if actionWord == "ADD" || actionWord == "REMOVE" {
					p.nextToken() // consume add/remove
				}
				// Parse path name
				pathNameTok := p.curTok
				path.Name = p.parseIdentifier()

				// Check for = 'path'
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
					pathLit, _ := p.parseStringLiteral()
					path.Path = pathLit
				}

				// Check for AS XQUERY 'type'
				if p.curTok.Type == TokenAs {
					p.nextToken() // consume AS
					if strings.ToUpper(p.curTok.Literal) == "XQUERY" {
						p.nextToken() // consume XQUERY
						xqDataType, _ := p.parseStringLiteral()
						path.XQueryDataType = xqDataType
					}
				}

				// The path's span runs from its name through its last child
				// clause (MAXLENGTH includes its closing paren); a trailing
				// SINGLETON keyword is excluded.
				p.spanFrom(pathNameTok, path)

				// Check for MAXLENGTH(n) or SINGLETON
				for {
					upperLit := strings.ToUpper(p.curTok.Literal)
					if upperLit == "MAXLENGTH" {
						p.nextToken() // consume MAXLENGTH
						if p.curTok.Type == TokenLParen {
							p.nextToken() // consume (
							path.MaxLength = p.intLitFromToken(p.curTok)
							p.nextToken() // consume number
							if p.curTok.Type == TokenRParen {
								p.nextToken() // consume )
							}
						}
						p.spanFrom(pathNameTok, path)
					} else if upperLit == "SINGLETON" {
						path.IsSingleton = true
						p.nextToken()
					} else {
						break
					}
				}

				stmt.PromotedPaths = append(stmt.PromotedPaths, path)

				if p.curTok.Type == TokenComma {
					p.nextToken()
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseXmlNamespaces parses WITH XMLNAMESPACES clause
func (p *Parser) parseXmlNamespaces() *ast.XmlNamespaces {
	p.nextToken() // consume XMLNAMESPACES
	xmlNs := &ast.XmlNamespaces{}

	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		// ScriptDom spans XMLNAMESPACES from the first element through the
		// closing paren.
		firstElemTok := p.curTok
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			elemTok := p.curTok
			elem := &ast.XmlNamespacesAliasElement{}
			// Parse string literal (namespace URI)
			strLit, _ := p.parseStringLiteral()
			elem.String = strLit

			// Expect AS
			if p.curTok.Type == TokenAs {
				p.nextToken() // consume AS
				elem.Identifier = p.parseIdentifier()
			}

			p.spanFrom(elemTok, elem)
			xmlNs.XmlNamespacesElements = append(xmlNs.XmlNamespacesElements, elem)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
		p.spanFrom(firstElemTok, xmlNs)
	}

	return xmlNs
}

func (p *Parser) getIndexOptionKind(optionName string) string {
	optionMap := map[string]string{
		"BUCKET_COUNT":                "BucketCount",
		"PAD_INDEX":                   "PadIndex",
		"FILLFACTOR":                  "FillFactor",
		"SORT_IN_TEMPDB":              "SortInTempDB",
		"IGNORE_DUP_KEY":              "IgnoreDupKey",
		"STATISTICS_NORECOMPUTE":      "StatisticsNoRecompute",
		"STATISTICS_INCREMENTAL":      "StatisticsIncremental",
		"DROP_EXISTING":               "DropExisting",
		"ONLINE":                      "Online",
		"ALLOW_ROW_LOCKS":             "AllowRowLocks",
		"ALLOW_PAGE_LOCKS":            "AllowPageLocks",
		"MAXDOP":                      "MaxDop",
		"DATA_COMPRESSION":            "DataCompression",
		"RESUMABLE":                   "Resumable",
		"MAX_DURATION":                "MaxDuration",
		"WAIT_AT_LOW_PRIORITY":        "WaitAtLowPriority",
		"OPTIMIZE_FOR_SEQUENTIAL_KEY": "OptimizeForSequentialKey",
		"COMPRESS_ALL_ROW_GROUPS":     "CompressAllRowGroups",
		"COMPRESSION_DELAY":           "CompressionDelay",
		"LOB_COMPACTION":              "LobCompaction",
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
	astStart := p.curTok

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
			paramStart := p.curTok
			param := &ast.ProcedureParameter{
				IsVarying: false,
				Modifier:  "None",
			}

			// Parse parameter name
			if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				param.VariableName = p.spanIdent(p.curTok.Literal, "NotQuoted")
				p.nextToken()
			}

			// Skip optional AS keyword (e.g., @param AS int)
			if p.curTok.Type == TokenAs {
				p.nextToken()
			}

			// Parse data type if present
			if p.curTok.Type != TokenRParen && p.curTok.Type != TokenComma {
				dataType, err := p.parseDataTypeReference()
				if err != nil {
					return nil, err
				}
				param.DataType = dataType
			}

			// Check for default value (= expr)
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
				defaultValue, err := p.parseScalarExpression()
				if err == nil {
					param.Value = defaultValue
				}
			}

			// Check for NULL/NOT NULL nullability
			if p.curTok.Type == TokenNull {
				param.Nullable = &ast.NullableConstraintDefinition{Nullable: true}
				p.tokSpan(param.Nullable, p.curTok)
				p.nextToken()
			} else if p.curTok.Type == TokenNot {
				notTok := p.curTok
				p.nextToken() // consume NOT
				if p.curTok.Type == TokenNull {
					param.Nullable = &ast.NullableConstraintDefinition{Nullable: false}
					p.spanTokens(param.Nullable, notTok, p.curTok)
					p.nextToken()
				}
			}

			// Check for READONLY modifier
			if strings.ToUpper(p.curTok.Literal) == "READONLY" {
				param.Modifier = "ReadOnly"
				p.nextToken()
			}

			p.spanFrom(paramStart, param)
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
		return spanned(p, stmt, astStart), nil
	}
	p.nextToken()

	// Check if RETURNS @varName TABLE or RETURNS TABLE (table-valued function)
	var tableVarName *ast.Identifier
	var tvfStartTok Token
	hasTvfStart := false
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		// Parse variable name for multi-statement table-valued function
		tvfStartTok = p.curTok
		hasTvfStart = true
		tableVarName = &ast.Identifier{
			Value:     p.curTok.Literal,
			QuoteType: "NotQuoted",
		}
		p.tokSpan(tableVarName, p.curTok)
		p.nextToken()
	}

	if strings.ToUpper(p.curTok.Literal) == "TABLE" {
		p.nextToken()

		// Check for column definitions in parentheses
		if p.curTok.Type == TokenLParen {
			p.nextToken()
			if !hasTvfStart {
				// Inline form: the return type spans from the first column
				// definition through the closing paren.
				tvfStartTok = p.curTok
				hasTvfStart = true
			}
			tableReturnType := &ast.TableValuedFunctionReturnType{
				DeclareTableVariableBody: &ast.DeclareTableVariableBody{
					VariableName: tableVarName,
					AsDefined:    false,
					Definition: &ast.TableDefinition{
						ColumnDefinitions: []*ast.ColumnDefinition{},
					},
				},
			}

			// Parse column definitions
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				colDefTok := p.curTok
				colDef := &ast.ColumnDefinition{
					IsPersisted:  false,
					IsRowGuidCol: false,
					IsHidden:     false,
					IsMasked:     false,
				}

				// Parse column name
				colDef.ColumnIdentifier = p.parseIdentifier()

				// Parse data type
				if p.curTok.Type != TokenRParen && p.curTok.Type != TokenComma {
					dataType, err := p.parseDataTypeReference()
					if err != nil {
						return nil, err
					}
					colDef.DataType = dataType
				}

				// Parse column constraints (PRIMARY KEY, NOT NULL, NULL, etc.)
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenComma && p.curTok.Type != TokenEOF {
					constrTok := p.curTok
					upperLit := strings.ToUpper(p.curTok.Literal)
					if upperLit == "PRIMARY" {
						p.nextToken() // consume PRIMARY
						if strings.ToUpper(p.curTok.Literal) == "KEY" {
							p.nextToken() // consume KEY
						}
						pkc := &ast.UniqueConstraintDefinition{
							IsPrimaryKey: true,
						}
						p.spanFrom(constrTok, pkc)
						colDef.Constraints = append(colDef.Constraints, pkc)
					} else if upperLit == "NOT" {
						p.nextToken() // consume NOT
						if p.curTok.Type == TokenNull {
							p.nextToken() // consume NULL
							nc := &ast.NullableConstraintDefinition{
								Nullable: false,
							}
							p.spanFrom(constrTok, nc)
							colDef.Constraints = append(colDef.Constraints, nc)
						}
					} else if p.curTok.Type == TokenNull {
						p.nextToken() // consume NULL
						nc := &ast.NullableConstraintDefinition{
							Nullable: true,
						}
						p.spanFrom(constrTok, nc)
						colDef.Constraints = append(colDef.Constraints, nc)
					} else if upperLit == "UNIQUE" {
						p.nextToken() // consume UNIQUE
						uc := &ast.UniqueConstraintDefinition{
							IsPrimaryKey: false,
						}
						p.spanFrom(constrTok, uc)
						colDef.Constraints = append(colDef.Constraints, uc)
					} else {
						break
					}
				}

				p.spanFrom(colDefTok, colDef)
				tableReturnType.DeclareTableVariableBody.Definition.ColumnDefinitions = append(
					tableReturnType.DeclareTableVariableBody.Definition.ColumnDefinitions,
					colDef,
				)

				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}

			if p.curTok.Type == TokenRParen {
				p.nextToken()
			}

			p.spanFrom(tvfStartTok, tableReturnType)
			p.spanFrom(tvfStartTok, tableReturnType.DeclareTableVariableBody)
			stmt.ReturnType = tableReturnType
		} else {
			// Simple RETURNS TABLE without column definitions
			stmt.ReturnType = &ast.TableValuedFunctionReturnType{}
		}

		// Parse optional WITH clause for function options
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			p.parseFunctionOptions(stmt)
		}

		// Parse optional ORDER clause for CLR table-valued functions
		if strings.ToUpper(p.curTok.Literal) == "ORDER" {
			orderTok := p.curTok
			p.nextToken() // consume ORDER
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				orderHint := &ast.OrderBulkInsertOption{
					OptionKind: "Order",
					IsUnique:   false,
				}

				// Parse columns with sort order
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					cwsTok := p.curTok
					colWithSort := &ast.ColumnWithSortOrder{
						Column: &ast.ColumnReferenceExpression{
							ColumnType: "Regular",
							MultiPartIdentifier: &ast.MultiPartIdentifier{
								Count:       1,
								Identifiers: []*ast.Identifier{p.parseIdentifier()},
							},
						},
						SortOrder: ast.SortOrderNotSpecified,
					}

					// Check for ASC/DESC
					upperSort := strings.ToUpper(p.curTok.Literal)
					if upperSort == "ASC" {
						colWithSort.SortOrder = ast.SortOrderAscending
						p.nextToken()
					} else if upperSort == "DESC" {
						colWithSort.SortOrder = ast.SortOrderDescending
						p.nextToken()
					}

					p.spanFrom(cwsTok, colWithSort)
					orderHint.Columns = append(orderHint.Columns, colWithSort)

					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}

				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}

				p.spanFrom(orderTok, orderHint)
				stmt.OrderHint = orderHint
			}
		}

		// Parse AS
		if p.curTok.Type == TokenAs {
			p.nextToken()
		}

		// Check for EXTERNAL NAME (CLR function)
		if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
			p.nextToken() // consume EXTERNAL
			if strings.ToUpper(p.curTok.Literal) == "NAME" {
				p.nextToken() // consume NAME
			}

			// Parse assembly.class.method
			stmt.MethodSpecifier = &ast.MethodSpecifier{}
			stmt.MethodSpecifier.AssemblyName = p.parseIdentifier()
			if p.curTok.Type == TokenDot {
				p.nextToken()
				stmt.MethodSpecifier.ClassName = p.parseIdentifier()
			}
			if p.curTok.Type == TokenDot {
				p.nextToken()
				stmt.MethodSpecifier.MethodName = p.parseIdentifier()
			}
		} else if strings.ToUpper(p.curTok.Literal) == "RETURN" {
			// Inline table-valued function: RETURN SELECT...
			p.nextToken()
			selectStmt, err := p.parseStatement()
			if err != nil {
				return nil, err
			}
			if sel, ok := selectStmt.(*ast.SelectStatement); ok {
				stmt.ReturnType = &ast.SelectFunctionReturnType{
					SelectStatement: sel,
				}
			}
		} else {
			// Multi-statement table-valued function: BEGIN ... END
			stmtList, err := p.parseFunctionStatementList()
			if err != nil {
				p.skipToEndOfStatement()
				return spanned(p, stmt, astStart), nil
			}
			// A function body's final statement excludes its trailing
			// semicolon, which belongs to the CREATE FUNCTION statement.
			if n := len(stmtList.Statements); n > 0 {
				if s, ok := any(stmtList.Statements[n-1]).(spannable); ok {
					p.trimTrailingSemicolon(s)
					s.Frag().Pin()
				}
			}
			spanStatementList(stmtList)
			stmt.StatementList = stmtList
		}
	} else {
		// Scalar function - parse return type
		returnDataType, err := p.parseDataTypeReference()
		if err != nil {
			p.skipToEndOfStatement()
			return spanned(p, stmt, astStart), nil
		}
		stmt.ReturnType = &ast.ScalarFunctionReturnType{
			DataType: returnDataType,
		}

		// Parse optional WITH clause for function options
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			p.parseFunctionOptions(stmt)
		}

		// Parse AS
		if p.curTok.Type == TokenAs {
			p.nextToken()
		}

		// Check for EXTERNAL NAME (CLR scalar function)
		if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
			p.nextToken() // consume EXTERNAL
			if strings.ToUpper(p.curTok.Literal) == "NAME" {
				p.nextToken() // consume NAME
			}

			// Parse assembly.class.method
			stmt.MethodSpecifier = &ast.MethodSpecifier{}
			stmt.MethodSpecifier.AssemblyName = p.parseIdentifier()
			if p.curTok.Type == TokenDot {
				p.nextToken()
				stmt.MethodSpecifier.ClassName = p.parseIdentifier()
			}
			if p.curTok.Type == TokenDot {
				p.nextToken()
				stmt.MethodSpecifier.MethodName = p.parseIdentifier()
			}
		} else {
			// Parse statement list
			stmtList, err := p.parseFunctionStatementList()
			if err != nil {
				p.skipToEndOfStatement()
				return spanned(p, stmt, astStart), nil
			}
			// A function body's final statement excludes its trailing
			// semicolon, which belongs to the CREATE FUNCTION statement.
			if n := len(stmtList.Statements); n > 0 {
				if s, ok := any(stmtList.Statements[n-1]).(spannable); ok {
					p.trimTrailingSemicolon(s)
					s.Frag().Pin()
				}
			}
			spanStatementList(stmtList)
			stmt.StatementList = stmtList
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseFunctionOptions parses function WITH options
func (p *Parser) parseFunctionOptions(stmt *ast.CreateFunctionStatement) {
	for {
		upperOpt := strings.ToUpper(p.curTok.Literal)
		switch upperOpt {
		case "INLINE":
			inlineTok := p.curTok
			p.nextToken() // consume INLINE
			// Expect = ON|OFF
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			optState := strings.ToUpper(p.curTok.Literal)
			state := "On"
			if optState == "OFF" {
				state = "Off"
			}
			p.nextToken() // consume ON/OFF
			inlineOpt := &ast.InlineFunctionOption{
				OptionKind:  "Inline",
				OptionState: state,
			}
			p.spanFrom(inlineTok, inlineOpt)
			stmt.Options = append(stmt.Options, inlineOpt)
		case "ENCRYPTION", "SCHEMABINDING", "NATIVE_COMPILATION":
			var optKind string
			switch upperOpt {
			case "ENCRYPTION":
				optKind = "Encryption"
			case "SCHEMABINDING":
				optKind = "SchemaBinding"
			case "NATIVE_COMPILATION":
				optKind = "NativeCompilation"
			}
			fo := &ast.FunctionOption{OptionKind: optKind}
			p.tokSpan(fo, p.curTok)
			p.nextToken()
			stmt.Options = append(stmt.Options, fo)
		case "CALLED":
			p.nextToken() // consume CALLED
			// Handle CALLED ON NULL INPUT
			var inputTok Token
			for strings.ToUpper(p.curTok.Literal) == "ON" || strings.ToUpper(p.curTok.Literal) == "NULL" || strings.ToUpper(p.curTok.Literal) == "INPUT" {
				if strings.ToUpper(p.curTok.Literal) == "INPUT" {
					inputTok = p.curTok
				}
				p.nextToken()
			}
			fo := &ast.FunctionOption{OptionKind: "CalledOnNullInput"}
			// ScriptDom spans this option on the INPUT keyword.
			p.tokSpan(fo, inputTok)
			stmt.Options = append(stmt.Options, fo)
		case "RETURNS":
			// Handle RETURNS NULL ON NULL INPUT
			var inputTok Token
			for strings.ToUpper(p.curTok.Literal) == "RETURNS" || strings.ToUpper(p.curTok.Literal) == "NULL" || strings.ToUpper(p.curTok.Literal) == "ON" || strings.ToUpper(p.curTok.Literal) == "INPUT" {
				if strings.ToUpper(p.curTok.Literal) == "INPUT" {
					inputTok = p.curTok
				}
				p.nextToken()
			}
			fo := &ast.FunctionOption{OptionKind: "ReturnsNullOnNullInput"}
			p.tokSpan(fo, inputTok)
			stmt.Options = append(stmt.Options, fo)
		case "EXECUTE":
			execTok := p.curTok
			p.nextToken() // consume EXECUTE
			if p.curTok.Type == TokenAs {
				p.nextToken() // consume AS
			}
			execAsOpt := &ast.ExecuteAsFunctionOption{
				OptionKind: "ExecuteAs",
				ExecuteAs:  &ast.ExecuteAsClause{},
			}
			// ScriptDom spans EXECUTE AS options on the EXECUTE keyword.
			p.tokSpan(execAsOpt, execTok)
			p.tokSpan(execAsOpt.ExecuteAs, execTok)
			upperOption := strings.ToUpper(p.curTok.Literal)
			switch upperOption {
			case "CALLER":
				execAsOpt.ExecuteAs.ExecuteAsOption = "Caller"
				p.nextToken()
			case "SELF":
				execAsOpt.ExecuteAs.ExecuteAsOption = "Self"
				p.nextToken()
			case "OWNER":
				execAsOpt.ExecuteAs.ExecuteAsOption = "Owner"
				p.nextToken()
			default:
				// String literal for user name
				if p.curTok.Type == TokenString {
					execAsOpt.ExecuteAs.ExecuteAsOption = "String"
					value := p.curTok.Literal
					// Strip quotes
					if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
						value = value[1 : len(value)-1]
					}
					execAsOpt.ExecuteAs.Literal = p.strLit(value, false)
					p.nextToken()
				}
			}
			if execAsOpt.ExecuteAs.ExecuteAsOption == "String" {
				// The string form spans through the literal.
				p.spanFrom(execTok, execAsOpt)
				p.spanFrom(execTok, execAsOpt.ExecuteAs)
			}
			stmt.Options = append(stmt.Options, execAsOpt)
		default:
			// Unknown option or end of options - break out
			if p.curTok.Type == TokenIdent && upperOpt != "ORDER" && upperOpt != "AS" {
				p.nextToken()
			} else {
				return
			}
		}

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		} else {
			break
		}
	}
}

// parseCreateOrAlterFunctionStatement parses a CREATE OR ALTER FUNCTION statement
func (p *Parser) parseCreateOrAlterFunctionStatement() (*ast.CreateOrAlterFunctionStatement, error) {
	astStart := p.curTok

	// Consume FUNCTION
	p.nextToken()

	stmt := &ast.CreateOrAlterFunctionStatement{}

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
			paramStart := p.curTok
			param := &ast.ProcedureParameter{
				IsVarying: false,
				Modifier:  "None",
			}

			// Parse parameter name
			if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				param.VariableName = p.spanIdent(p.curTok.Literal, "NotQuoted")
				p.nextToken()
			}

			// Skip optional AS keyword (e.g., @param AS int)
			if p.curTok.Type == TokenAs {
				p.nextToken()
			}

			// Parse data type if present
			if p.curTok.Type != TokenRParen && p.curTok.Type != TokenComma {
				dataType, err := p.parseDataTypeReference()
				if err != nil {
					return nil, err
				}
				param.DataType = dataType
			}

			// Check for default value (= expr)
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
				defaultValue, err := p.parseScalarExpression()
				if err == nil {
					param.Value = defaultValue
				}
			}

			// Check for NULL/NOT NULL nullability
			if p.curTok.Type == TokenNull {
				param.Nullable = &ast.NullableConstraintDefinition{Nullable: true}
				p.tokSpan(param.Nullable, p.curTok)
				p.nextToken()
			} else if p.curTok.Type == TokenNot {
				notTok := p.curTok
				p.nextToken() // consume NOT
				if p.curTok.Type == TokenNull {
					param.Nullable = &ast.NullableConstraintDefinition{Nullable: false}
					p.spanTokens(param.Nullable, notTok, p.curTok)
					p.nextToken()
				}
			}

			// Check for READONLY modifier
			if strings.ToUpper(p.curTok.Literal) == "READONLY" {
				param.Modifier = "ReadOnly"
				p.nextToken()
			}

			p.spanFrom(paramStart, param)
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
		return spanned(p, stmt, astStart), nil
	}
	p.nextToken()

	// Parse return type
	returnDataType, err := p.parseDataTypeReference()
	if err != nil {
		p.skipToEndOfStatement()
		return spanned(p, stmt, astStart), nil
	}
	stmt.ReturnType = &ast.ScalarFunctionReturnType{
		DataType: returnDataType,
	}

	// Parse optional WITH clause for function options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		for {
			upperOpt := strings.ToUpper(p.curTok.Literal)
			switch upperOpt {
			case "INLINE":
				inlineTok := p.curTok
				p.nextToken() // consume INLINE
				// Expect = ON|OFF
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				optState := strings.ToUpper(p.curTok.Literal)
				state := "On"
				if optState == "OFF" {
					state = "Off"
				}
				p.nextToken() // consume ON/OFF
				inlineOpt := &ast.InlineFunctionOption{
					OptionKind:  "Inline",
					OptionState: state,
				}
				p.spanFrom(inlineTok, inlineOpt)
				stmt.Options = append(stmt.Options, inlineOpt)
			case "ENCRYPTION", "SCHEMABINDING", "NATIVE_COMPILATION", "CALLED":
				var optKind string
				switch strings.ToUpper(p.curTok.Literal) {
				case "ENCRYPTION":
					optKind = "Encryption"
				case "SCHEMABINDING":
					optKind = "SchemaBinding"
				case "NATIVE_COMPILATION":
					optKind = "NativeCompilation"
				case "CALLED":
					optKind = "CalledOnNullInput"
				}
				p.nextToken()
				// Handle CALLED ON NULL INPUT - skip additional tokens
				if optKind == "CalledOnNullInput" {
					for strings.ToUpper(p.curTok.Literal) == "ON" || strings.ToUpper(p.curTok.Literal) == "NULL" || strings.ToUpper(p.curTok.Literal) == "INPUT" {
						p.nextToken()
					}
				}
				stmt.Options = append(stmt.Options, &ast.FunctionOption{
					OptionKind: optKind,
				})
			case "RETURNS":
				// Handle RETURNS NULL ON NULL INPUT
				var inputTok Token
				for strings.ToUpper(p.curTok.Literal) == "RETURNS" || strings.ToUpper(p.curTok.Literal) == "NULL" || strings.ToUpper(p.curTok.Literal) == "ON" || strings.ToUpper(p.curTok.Literal) == "INPUT" {
					if strings.ToUpper(p.curTok.Literal) == "INPUT" {
						inputTok = p.curTok
					}
					p.nextToken()
				}
				fo := &ast.FunctionOption{OptionKind: "ReturnsNullOnNullInput"}
				p.tokSpan(fo, inputTok)
				stmt.Options = append(stmt.Options, fo)
			case "EXECUTE":
				execTok := p.curTok
				p.nextToken() // consume EXECUTE
				if p.curTok.Type == TokenAs {
					p.nextToken() // consume AS
				}
				execAsOpt := &ast.ExecuteAsFunctionOption{
					OptionKind: "ExecuteAs",
					ExecuteAs:  &ast.ExecuteAsClause{},
				}
				// ScriptDom spans EXECUTE AS options on the EXECUTE keyword.
				p.tokSpan(execAsOpt, execTok)
				p.tokSpan(execAsOpt.ExecuteAs, execTok)
				upperOption := strings.ToUpper(p.curTok.Literal)
				switch upperOption {
				case "CALLER":
					execAsOpt.ExecuteAs.ExecuteAsOption = "Caller"
					p.nextToken()
				case "SELF":
					execAsOpt.ExecuteAs.ExecuteAsOption = "Self"
					p.nextToken()
				case "OWNER":
					execAsOpt.ExecuteAs.ExecuteAsOption = "Owner"
					p.nextToken()
				default:
					// String literal for user name
					if p.curTok.Type == TokenString {
						execAsOpt.ExecuteAs.ExecuteAsOption = "String"
						value := p.curTok.Literal
						// Strip quotes
						if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
							value = value[1 : len(value)-1]
						}
						execAsOpt.ExecuteAs.Literal = p.strLit(value, false)
						p.nextToken()
					}
				}
				if execAsOpt.ExecuteAs.ExecuteAsOption == "String" {
					// The string form spans through the literal.
					p.spanFrom(execTok, execAsOpt)
					p.spanFrom(execTok, execAsOpt.ExecuteAs)
				}
				stmt.Options = append(stmt.Options, execAsOpt)
			default:
				// Unknown option - skip it
				if p.curTok.Type == TokenIdent {
					p.nextToken()
				}
			}

			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma
			} else {
				break
			}
		}
	}

	// Parse AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse statement list
	stmtList, err := p.parseFunctionStatementList()
	if err != nil {
		p.skipToEndOfStatement()
		return spanned(p, stmt, astStart), nil
	}
	// A function body's final statement excludes its trailing semicolon.
	if n := len(stmtList.Statements); n > 0 {
		if s, ok := any(stmtList.Statements[n-1]).(spannable); ok {
			p.trimTrailingSemicolon(s)
			s.Frag().Pin()
		}
	}
	spanStatementList(stmtList)
	stmt.StatementList = stmtList

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// parseCreateOrAlterProcedureStatement parses a CREATE OR ALTER PROCEDURE statement
func (p *Parser) parseCreateOrAlterProcedureStatement() (*ast.CreateOrAlterProcedureStatement, error) {
	astStart := p.curTok

	// Parse as regular CREATE PROCEDURE, then convert to CreateOrAlter type
	stmt, err := p.parseCreateProcedureStatement()
	if err != nil {
		return nil, err
	}
	return spanned(p, &ast.CreateOrAlterProcedureStatement{
		ProcedureReference: stmt.ProcedureReference,
		Parameters:         stmt.Parameters,
		StatementList:      stmt.StatementList,
		IsForReplication:   stmt.IsForReplication,
		Options:            stmt.Options,
		MethodSpecifier:    stmt.MethodSpecifier,
	}, astStart), nil
}

// parseCreateOrAlterViewStatement parses a CREATE OR ALTER VIEW statement
func (p *Parser) parseCreateOrAlterViewStatement() (*ast.CreateOrAlterViewStatement, error) {
	astStart := p.curTok

	// Parse as regular CREATE VIEW, then convert to CreateOrAlter type
	stmt, err := p.parseCreateViewStatement()
	if err != nil {
		return nil, err
	}
	return spanned(p, &ast.CreateOrAlterViewStatement{
		SchemaObjectName: stmt.SchemaObjectName,
		Columns:          stmt.Columns,
		SelectStatement:  stmt.SelectStatement,
		WithCheckOption:  stmt.WithCheckOption,
		ViewOptions:      stmt.ViewOptions,
		IsMaterialized:   stmt.IsMaterialized,
	}, astStart), nil
}

// parseCreateOrAlterTriggerStatement parses a CREATE OR ALTER TRIGGER statement
func (p *Parser) parseCreateOrAlterTriggerStatement() (*ast.CreateOrAlterTriggerStatement, error) {
	astStart := p.curTok

	// Parse as regular CREATE TRIGGER, then convert to CreateOrAlter type
	stmt, err := p.parseCreateTriggerStatement()
	if err != nil {
		return nil, err
	}
	return spanned(p, &ast.CreateOrAlterTriggerStatement{
		Name:                stmt.Name,
		TriggerObject:       stmt.TriggerObject,
		TriggerType:         stmt.TriggerType,
		TriggerActions:      stmt.TriggerActions,
		Options:             stmt.Options,
		WithAppend:          stmt.WithAppend,
		IsNotForReplication: stmt.IsNotForReplication,
		MethodSpecifier:     stmt.MethodSpecifier,
		StatementList:       stmt.StatementList,
	}, astStart), nil
}

// parseCreateTriggerStatement parses a CREATE TRIGGER statement
func (p *Parser) parseCreateTriggerStatement() (*ast.CreateTriggerStatement, error) {
	astStart := p.curTok

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
	triggerObjTok := p.curTok
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
	// ScriptDom spans the trigger object over the ON target tokens.
	p.spanFrom(triggerObjTok, triggerObject)
	stmt.TriggerObject = triggerObject

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		for {
			optName := strings.ToUpper(p.curTok.Literal)
			switch optName {
			case "NATIVE_COMPILATION":
				sopt := &ast.TriggerOption{OptionKind: "NativeCompile"}
				p.tokSpan(sopt, p.curTok)
				stmt.Options = append(stmt.Options, sopt)
				p.nextToken()
			case "SCHEMABINDING":
				sopt := &ast.TriggerOption{OptionKind: "SchemaBinding"}
				p.tokSpan(sopt, p.curTok)
				stmt.Options = append(stmt.Options, sopt)
				p.nextToken()
			case "ENCRYPTION":
				sopt := &ast.TriggerOption{OptionKind: "Encryption"}
				p.tokSpan(sopt, p.curTok)
				stmt.Options = append(stmt.Options, sopt)
				p.nextToken()
			case "EXECUTE":
				execTok := p.curTok
				p.nextToken() // consume EXECUTE
				if p.curTok.Type == TokenAs {
					p.nextToken() // consume AS
				}
				execAsClause := &ast.ExecuteAsClause{}
				switch strings.ToUpper(p.curTok.Literal) {
				case "CALLER":
					execAsClause.ExecuteAsOption = "Caller"
					p.nextToken()
				case "SELF":
					execAsClause.ExecuteAsOption = "Self"
					p.nextToken()
				case "OWNER":
					execAsClause.ExecuteAsOption = "Owner"
					p.nextToken()
				default:
					// Check for string literal (e.g., EXECUTE AS 'dbo')
					if p.curTok.Type == TokenString {
						strLit, err := p.parseStringLiteral()
						if err != nil {
							return nil, err
						}
						execAsClause.ExecuteAsOption = "String"
						execAsClause.Literal = strLit
					} else {
						// User name
						execAsClause.ExecuteAsOption = "User"
						p.nextToken()
					}
				}
				execTrigOpt := &ast.ExecuteAsTriggerOption{
					OptionKind:      "ExecuteAsClause",
					ExecuteAsClause: execAsClause,
				}
				// ScriptDom spans keyword forms on EXECUTE alone but the
				// string form through the literal.
				if execAsClause.ExecuteAsOption == "String" {
					p.spanFrom(execTok, execTrigOpt)
					p.spanFrom(execTok, execAsClause)
				} else {
					p.tokSpan(execTrigOpt, execTok)
					p.tokSpan(execAsClause, execTok)
				}
				stmt.Options = append(stmt.Options, execTrigOpt)
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
		p.tokSpan(action, p.curTok)
		actionType := strings.ToUpper(p.curTok.Literal)

		// Check for empty action type (lenient parsing for incomplete statements)
		if actionType == "" || p.curTok.Type == TokenAs {
			break
		}

		// Check for NOT FOR REPLICATION
		if actionType == "NOT" {
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
				// Convert action type to proper case (e.g., DENY_DATABASE -> DenyDatabase)
				eventType := convertEventTypeCase(actionType)
				action.EventTypeGroup = &ast.EventTypeContainer{
					EventType: eventType,
				}
				p.tokSpan(action.EventTypeGroup, p.curTok)
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

	// Parse NOT FOR REPLICATION
	if strings.ToUpper(p.curTok.Literal) == "NOT" {
		p.nextToken() // consume NOT
		if strings.ToUpper(p.curTok.Literal) == "FOR" {
			p.nextToken() // consume FOR
		}
		if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
			p.nextToken() // consume REPLICATION
			stmt.IsNotForReplication = true
		}
	}

	// Parse AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Skip leading semicolons
	for p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	// Check for EXTERNAL NAME (CLR trigger)
	if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
		p.nextToken() // consume EXTERNAL
		if strings.ToUpper(p.curTok.Literal) == "NAME" {
			p.nextToken() // consume NAME
		}
		// Parse assembly.class.method
		stmt.MethodSpecifier = &ast.MethodSpecifier{}
		stmt.MethodSpecifier.AssemblyName = p.parseIdentifier()
		if p.curTok.Type == TokenDot {
			p.nextToken()
			stmt.MethodSpecifier.ClassName = p.parseIdentifier()
		}
		if p.curTok.Type == TokenDot {
			p.nextToken()
			stmt.MethodSpecifier.MethodName = p.parseIdentifier()
		}
		// Skip optional semicolons
		for p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return spanned(p, stmt, astStart), nil
	}

	// Parse statement list (all statements until GO/EOF)
	stmtList := &ast.StatementList{}
	for p.curTok.Type != TokenEOF {
		// Check for GO or end of batch
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			break
		}

		// Skip semicolons between statements
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
			continue
		}

		innerStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if innerStmt != nil {
			stmtList.Statements = append(stmtList.Statements, innerStmt)
		}
	}
	spanStatementList(stmtList)
	stmt.StatementList = stmtList

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return spanned(p, stmt, astStart), nil
}

// convertEventTypeCase converts an event type like "DENY_DATABASE" to "DenyDatabase"
func convertEventTypeCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
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
	if len(s.Files) > 0 {
		files := make([]jsonNode, len(s.Files))
		for i, f := range s.Files {
			files[i] = backupRestoreFileInfoToJSON(f)
		}
		node["Files"] = files
	}
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = restoreOptionToJSON(o)
		}
		node["Options"] = options
	}
	return addSpan(node, frag(s))
}

func backupRestoreFileInfoToJSON(f *ast.BackupRestoreFileInfo) jsonNode {
	node := jsonNode{
		"$type":    "BackupRestoreFileInfo",
		"ItemKind": f.ItemKind,
	}
	if len(f.Items) > 0 {
		items := make([]jsonNode, len(f.Items))
		for i, item := range f.Items {
			items[i] = scalarExpressionToJSON(item)
		}
		node["Items"] = items
	}
	return addSpan(node, frag(f))
}

func backupDatabaseStatementToJSON(s *ast.BackupDatabaseStatement) jsonNode {
	node := jsonNode{
		"$type": "BackupDatabaseStatement",
	}
	if len(s.Files) > 0 {
		files := make([]jsonNode, len(s.Files))
		for i, f := range s.Files {
			files[i] = backupRestoreFileInfoToJSON(f)
		}
		node["Files"] = files
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierOrValueExpressionToJSON(s.DatabaseName)
	}
	if len(s.MirrorToClauses) > 0 {
		clauses := make([]jsonNode, len(s.MirrorToClauses))
		for i, c := range s.MirrorToClauses {
			clauses[i] = mirrorToClauseToJSON(c)
		}
		node["MirrorToClauses"] = clauses
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
			options[i] = backupOptionBaseToJSON(o)
		}
		node["Options"] = options
	}
	return addSpan(node, frag(s))
}

func mirrorToClauseToJSON(c *ast.MirrorToClause) jsonNode {
	node := jsonNode{
		"$type": "MirrorToClause",
	}
	if len(c.Devices) > 0 {
		devices := make([]jsonNode, len(c.Devices))
		for i, d := range c.Devices {
			devices[i] = deviceInfoToJSON(d)
		}
		node["Devices"] = devices
	}
	return addSpan(node, frag(c))
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
			options[i] = backupOptionBaseToJSON(o)
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func backupMasterKeyStatementToJSON(s *ast.BackupMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "BackupMasterKeyStatement",
	}
	if s.File != nil {
		node["File"] = scalarExpressionToJSON(s.File)
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func backupOptionBaseToJSON(o ast.BackupOptionBase) jsonNode {
	switch opt := o.(type) {
	case *ast.BackupOption:
		return addSpan(backupOptionToJSON(opt), frag(o))
	case *ast.BackupEncryptionOption:
		return addSpan(backupEncryptionOptionToJSON(opt), frag(o))
	default:
		return addSpan(jsonNode{"$type": "Unknown"}, frag(o))
	}
}

func backupOptionToJSON(o *ast.BackupOption) jsonNode {
	node := jsonNode{
		"$type":      "BackupOption",
		"OptionKind": o.OptionKind,
	}
	if o.Value != nil {
		node["Value"] = scalarExpressionToJSON(o.Value)
	}
	return addSpan(node, frag(o))
}

func backupEncryptionOptionToJSON(o *ast.BackupEncryptionOption) jsonNode {
	node := jsonNode{
		"$type":      "BackupEncryptionOption",
		"Algorithm":  o.Algorithm,
		"OptionKind": o.OptionKind,
	}
	if o.Encryptor != nil {
		node["Encryptor"] = cryptoMechanismToJSON(o.Encryptor)
	}
	return addSpan(node, frag(o))
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
	return addSpan(node, frag(d))
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
		return addSpan(node, frag(o))
	case *ast.GeneralSetCommandRestoreOption:
		node := jsonNode{
			"$type":      "GeneralSetCommandRestoreOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.OptionValue != nil {
			node["OptionValue"] = scalarExpressionToJSON(opt.OptionValue)
		}
		return addSpan(node, frag(o))
	case *ast.SimpleRestoreOption:
		return addSpan(jsonNode{
			"$type":      "RestoreOption",
			"OptionKind": opt.OptionKind,
		}, frag(o))
	case *ast.StopRestoreOption:
		node := jsonNode{
			"$type":      "StopRestoreOption",
			"OptionKind": opt.OptionKind,
			"IsStopAt":   opt.IsStopAt,
		}
		if opt.Mark != nil {
			node["Mark"] = scalarExpressionToJSON(opt.Mark)
		}
		if opt.After != nil {
			node["After"] = scalarExpressionToJSON(opt.After)
		}
		return addSpan(node, frag(o))
	case *ast.ScalarExpressionRestoreOption:
		node := jsonNode{
			"$type":      "ScalarExpressionRestoreOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return addSpan(node, frag(o))
	case *ast.MoveRestoreOption:
		node := jsonNode{
			"$type":      "MoveRestoreOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.LogicalFileName != nil {
			node["LogicalFileName"] = scalarExpressionToJSON(opt.LogicalFileName)
		}
		if opt.OSFileName != nil {
			node["OSFileName"] = scalarExpressionToJSON(opt.OSFileName)
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownRestoreOption"}, frag(o))
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
	return addSpan(node, frag(f))
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
	return addSpan(node, frag(s))
}

func userLoginOptionToJSON(u *ast.UserLoginOption) jsonNode {
	node := jsonNode{
		"$type":               "UserLoginOption",
		"UserLoginOptionType": u.UserLoginOptionType,
	}
	if u.Identifier != nil {
		node["Identifier"] = identifierToJSON(u.Identifier)
	}
	return addSpan(node, frag(u))
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
		return addSpan(node, frag(o))
	case *ast.IdentifierPrincipalOption:
		node := jsonNode{
			"$type":      "IdentifierPrincipalOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Identifier != nil {
			node["Identifier"] = identifierToJSON(opt.Identifier)
		}
		return addSpan(node, frag(o))
	case *ast.PasswordAlterPrincipalOption:
		node := jsonNode{
			"$type":      "PasswordAlterPrincipalOption",
			"OptionKind": opt.OptionKind,
			"MustChange": opt.MustChange,
			"Unlock":     opt.Unlock,
			"Hashed":     opt.Hashed,
		}
		if opt.Password != nil {
			node["Password"] = scalarExpressionToJSON(opt.Password)
		}
		if opt.OldPassword != nil {
			node["OldPassword"] = stringLiteralToJSON(opt.OldPassword)
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownUserOption"}, frag(o))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(a))
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
	if s.OnFileGroupOrPartitionScheme != nil {
		node["OnFileGroupOrPartitionScheme"] = fileGroupOrPartitionSchemeToJSON(s.OnFileGroupOrPartitionScheme)
	}
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(opt))
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
		return addSpan(node, frag(opt))
	case *ast.IndexStateOption:
		return addSpan(jsonNode{
			"$type":       "IndexStateOption",
			"OptionKind":  o.OptionKind,
			"OptionState": o.OptionState,
		}, frag(opt))
	case *ast.IndexExpressionOption:
		node := jsonNode{
			"$type":      "IndexExpressionOption",
			"OptionKind": o.OptionKind,
		}
		if o.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(o.Expression)
		}
		return addSpan(node, frag(opt))
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
		return addSpan(node, frag(opt))
	case *ast.OnlineIndexOption:
		node := jsonNode{
			"$type":       "OnlineIndexOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		if o.LowPriorityLockWaitOption != nil {
			node["LowPriorityLockWaitOption"] = onlineIndexLowPriorityLockWaitOptionToJSON(o.LowPriorityLockWaitOption)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownIndexOption"}, frag(opt))
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(opt))
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
		return addSpan(node, frag(opt))
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
		return addSpan(node, frag(opt))
	case *ast.CellsPerObjectSpatialIndexOption:
		node := jsonNode{
			"$type": "CellsPerObjectSpatialIndexOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownSpatialIndexOption"}, frag(opt))
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
	return addSpan(node, frag(p))
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
	return addSpan(node, frag(p))
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
			opts[i] = functionOptionBaseToJSON(o)
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
	return addSpan(node, frag(s))
}

func functionOptionToJSON(o *ast.FunctionOption) jsonNode {
	return addSpan(jsonNode{
		"$type":      "FunctionOption",
		"OptionKind": o.OptionKind,
	}, frag(o))
}

func inlineFunctionOptionToJSON(o *ast.InlineFunctionOption) jsonNode {
	return addSpan(jsonNode{
		"$type":       "InlineFunctionOption",
		"OptionState": o.OptionState,
		"OptionKind":  o.OptionKind,
	}, frag(o))
}

func functionOptionBaseToJSON(o ast.FunctionOptionBase) jsonNode {
	switch opt := o.(type) {
	case *ast.FunctionOption:
		return addSpan(functionOptionToJSON(opt), frag(o))
	case *ast.InlineFunctionOption:
		return addSpan(inlineFunctionOptionToJSON(opt), frag(o))
	case *ast.ExecuteAsFunctionOption:
		return addSpan(executeAsFunctionOptionToJSON(opt), frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownFunctionOption"}, frag(o))
	}
}

func executeAsFunctionOptionToJSON(o *ast.ExecuteAsFunctionOption) jsonNode {
	node := jsonNode{
		"$type":      "ExecuteAsFunctionOption",
		"OptionKind": o.OptionKind,
	}
	if o.ExecuteAs != nil {
		node["ExecuteAs"] = executeAsClauseToJSON(o.ExecuteAs)
	}
	return addSpan(node, frag(o))
}

func orderBulkInsertOptionToJSON(o *ast.OrderBulkInsertOption) jsonNode {
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
	return addSpan(node, frag(o))
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
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = functionOptionBaseToJSON(o)
		}
		node["Options"] = options
	}
	if s.OrderHint != nil {
		node["OrderHint"] = orderBulkInsertOptionToJSON(s.OrderHint)
	}
	if s.MethodSpecifier != nil {
		node["MethodSpecifier"] = methodSpecifierToJSON(s.MethodSpecifier)
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return addSpan(node, frag(s))
}

func createOrAlterFunctionStatementToJSON(s *ast.CreateOrAlterFunctionStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateOrAlterFunctionStatement",
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
	if len(s.Options) > 0 {
		options := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			options[i] = functionOptionBaseToJSON(o)
		}
		node["Options"] = options
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(r))
	case *ast.SelectFunctionReturnType:
		node := jsonNode{
			"$type": "SelectFunctionReturnType",
		}
		if rt.SelectStatement != nil {
			node["SelectStatement"] = selectStatementToJSON(rt.SelectStatement)
		}
		return addSpan(node, frag(r))
	case *ast.TableValuedFunctionReturnType:
		node := jsonNode{
			"$type": "TableValuedFunctionReturnType",
		}
		if rt.DeclareTableVariableBody != nil {
			node["DeclareTableVariableBody"] = declareTableVariableBodyToJSON(rt.DeclareTableVariableBody)
		}
		return addSpan(node, frag(r))
	default:
		return addSpan(jsonNode{"$type": "UnknownFunctionReturnType"}, frag(r))
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
	return addSpan(node, frag(s))
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
	if s.MethodSpecifier != nil {
		node["MethodSpecifier"] = methodSpecifierToJSON(s.MethodSpecifier)
	}
	if s.StatementList != nil {
		node["StatementList"] = statementListToJSON(s.StatementList)
	}
	return addSpan(node, frag(s))
}

func createOrAlterTriggerStatementToJSON(s *ast.CreateOrAlterTriggerStatement) jsonNode {
	node := jsonNode{
		"$type":               "CreateOrAlterTriggerStatement",
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(o))
	case *ast.ExecuteAsTriggerOption:
		node := jsonNode{
			"$type":      "ExecuteAsTriggerOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.ExecuteAsClause != nil {
			execClause := jsonNode{
				"$type":           "ExecuteAsClause",
				"ExecuteAsOption": opt.ExecuteAsClause.ExecuteAsOption,
			}
			if opt.ExecuteAsClause.Literal != nil {
				execClause["Literal"] = stringLiteralToJSON(opt.ExecuteAsClause.Literal)
			}
			node["ExecuteAsClause"] = addSpan(execClause, frag(opt.ExecuteAsClause))
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownTriggerOption"}, frag(o))
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
	return addSpan(node, frag(t))
}

func triggerActionToJSON(a *ast.TriggerAction) jsonNode {
	node := jsonNode{
		"$type":             "TriggerAction",
		"TriggerActionType": a.TriggerActionType,
	}
	if a.EventTypeGroup != nil {
		node["EventTypeGroup"] = addSpan(jsonNode{
			"$type":     "EventTypeContainer",
			"EventType": a.EventTypeGroup.EventType,
		}, frag(a.EventTypeGroup))
	}
	return addSpan(node, frag(a))
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
	return addSpan(node, frag(s))
}

func endConversationStatementToJSON(s *ast.EndConversationStatement) jsonNode {
	node := jsonNode{
		"$type":       "EndConversationStatement",
		"WithCleanup": s.WithCleanup,
	}
	if s.Conversation != nil {
		node["Conversation"] = scalarExpressionToJSON(s.Conversation)
	}
	if s.ErrorCode != nil {
		node["ErrorCode"] = scalarExpressionToJSON(s.ErrorCode)
	}
	if s.ErrorDescription != nil {
		node["ErrorDescription"] = scalarExpressionToJSON(s.ErrorDescription)
	}
	return addSpan(node, frag(s))
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
	if len(s.PromotedPaths) > 0 {
		paths := make([]jsonNode, len(s.PromotedPaths))
		for i, p := range s.PromotedPaths {
			paths[i] = selectiveXmlIndexPromotedPathToJSON(p)
		}
		node["PromotedPaths"] = paths
	}
	if s.XmlNamespaces != nil {
		node["XmlNamespaces"] = xmlNamespacesToJSON(s.XmlNamespaces)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if len(s.IndexOptions) > 0 {
		opts := make([]jsonNode, len(s.IndexOptions))
		for i, opt := range s.IndexOptions {
			opts[i] = indexOptionToJSON(opt)
		}
		node["IndexOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func selectiveXmlIndexPromotedPathToJSON(p *ast.SelectiveXmlIndexPromotedPath) jsonNode {
	node := jsonNode{
		"$type": "SelectiveXmlIndexPromotedPath",
	}
	if p.Name != nil {
		node["Name"] = identifierToJSON(p.Name)
	}
	if p.Path != nil {
		node["Path"] = stringLiteralToJSON(p.Path)
	}
	if p.XQueryDataType != nil {
		node["XQueryDataType"] = stringLiteralToJSON(p.XQueryDataType)
	}
	if p.SQLDataType != nil {
		node["SQLDataType"] = sqlDataTypeReferenceToJSON(p.SQLDataType)
	}
	if p.MaxLength != nil {
		node["MaxLength"] = scalarExpressionToJSON(p.MaxLength)
	}
	node["IsSingleton"] = p.IsSingleton
	return addSpan(node, frag(p))
}

func xmlNamespacesToJSON(x *ast.XmlNamespaces) jsonNode {
	node := jsonNode{
		"$type": "XmlNamespaces",
	}
	if len(x.XmlNamespacesElements) > 0 {
		elems := make([]jsonNode, len(x.XmlNamespacesElements))
		for i, e := range x.XmlNamespacesElements {
			elems[i] = xmlNamespacesElementToJSON(e)
		}
		node["XmlNamespacesElements"] = elems
	}
	return addSpan(node, frag(x))
}

func xmlNamespacesElementToJSON(e ast.XmlNamespacesElement) jsonNode {
	switch elem := e.(type) {
	case *ast.XmlNamespacesAliasElement:
		return addSpan(xmlNamespacesAliasElementToJSON(elem), frag(e))
	case *ast.XmlNamespacesDefaultElement:
		return addSpan(xmlNamespacesDefaultElementToJSON(elem), frag(e))
	default:
		return addSpan(jsonNode{}, frag(e))
	}
}

func xmlNamespacesAliasElementToJSON(e *ast.XmlNamespacesAliasElement) jsonNode {
	node := jsonNode{
		"$type": "XmlNamespacesAliasElement",
	}
	if e.Identifier != nil {
		node["Identifier"] = identifierToJSON(e.Identifier)
	}
	if e.String != nil {
		node["String"] = stringLiteralToJSON(e.String)
	}
	return addSpan(node, frag(e))
}

func xmlNamespacesDefaultElementToJSON(e *ast.XmlNamespacesDefaultElement) jsonNode {
	node := jsonNode{
		"$type": "XmlNamespacesDefaultElement",
	}
	if e.String != nil {
		node["String"] = stringLiteralToJSON(e.String)
	}
	return addSpan(node, frag(e))
}

func partitionSpecifierToJSON(p *ast.PartitionSpecifier) jsonNode {
	node := jsonNode{
		"$type": "PartitionSpecifier",
		"All":   p.All,
	}
	if p.Number != nil {
		node["Number"] = scalarExpressionToJSON(p.Number)
	}
	return addSpan(node, frag(p))
}

func indexOptionToJSON(opt ast.IndexOption) jsonNode {
	switch o := opt.(type) {
	case *ast.IndexStateOption:
		return addSpan(jsonNode{
			"$type":       "IndexStateOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}, frag(opt))
	case *ast.IndexExpressionOption:
		return addSpan(jsonNode{
			"$type":      "IndexExpressionOption",
			"OptionKind": o.OptionKind,
			"Expression": scalarExpressionToJSON(o.Expression),
		}, frag(opt))
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
		return addSpan(node, frag(opt))
	case *ast.IgnoreDupKeyIndexOption:
		node := jsonNode{
			"$type":       "IgnoreDupKeyIndexOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		if o.SuppressMessagesOption != nil {
			node["SuppressMessagesOption"] = *o.SuppressMessagesOption
		}
		return addSpan(node, frag(opt))
	case *ast.OnlineIndexOption:
		node := jsonNode{
			"$type":       "OnlineIndexOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		if o.LowPriorityLockWaitOption != nil {
			node["LowPriorityLockWaitOption"] = onlineIndexLowPriorityLockWaitOptionToJSON(o.LowPriorityLockWaitOption)
		}
		return addSpan(node, frag(opt))
	case *ast.CompressionDelayIndexOption:
		node := jsonNode{
			"$type":      "CompressionDelayIndexOption",
			"OptionKind": o.OptionKind,
			"TimeUnit":   o.TimeUnit,
		}
		if o.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(o.Expression)
		}
		return addSpan(node, frag(opt))
	case *ast.MaxDurationOption:
		node := jsonNode{
			"$type":      "MaxDurationOption",
			"OptionKind": o.OptionKind,
		}
		if o.MaxDuration != nil {
			node["MaxDuration"] = scalarExpressionToJSON(o.MaxDuration)
		}
		if o.Unit != "" {
			node["Unit"] = o.Unit
		}
		return addSpan(node, frag(opt))
	case *ast.XmlCompressionOption:
		node := jsonNode{
			"$type":        "XmlCompressionOption",
			"IsCompressed": o.IsCompressed,
			"OptionKind":   o.OptionKind,
		}
		if len(o.PartitionRanges) > 0 {
			ranges := make([]jsonNode, len(o.PartitionRanges))
			for i, r := range o.PartitionRanges {
				ranges[i] = compressionPartitionRangeToJSON(r)
			}
			node["PartitionRanges"] = ranges
		}
		return addSpan(node, frag(opt))
	case *ast.WaitAtLowPriorityOption:
		node := jsonNode{
			"$type":      "WaitAtLowPriorityOption",
			"OptionKind": o.OptionKind,
		}
		if len(o.Options) > 0 {
			options := make([]jsonNode, len(o.Options))
			for i, opt := range o.Options {
				options[i] = lowPriorityLockWaitOptionToJSON(opt)
			}
			node["Options"] = options
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownIndexOption"}, frag(opt))
	}
}

func convertUserOptionKind(name string) string {
	// Convert option names to the expected format
	optionMap := map[string]string{
		"OBJECT_ID":        "Object_ID",
		"DEFAULT_SCHEMA":   "DefaultSchema",
		"DEFAULT_LANGUAGE": "DefaultLanguage",
		"SID":              "Sid",
		"PASSWORD":         "Password",
		"NAME":             "Name",
		"LOGIN":            "Login",
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func dropIndexStatementToJSON(s *ast.DropIndexStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropIndexStatement",
		"IsIfExists": s.IsIfExists,
	}
	if len(s.DropIndexClauses) > 0 {
		clauses := make([]jsonNode, len(s.DropIndexClauses))
		for i, clause := range s.DropIndexClauses {
			clauses[i] = dropIndexClauseToJSON(clause)
		}
		node["DropIndexClauses"] = clauses
	}
	return addSpan(node, frag(s))
}

func dropIndexClauseToJSON(c *ast.DropIndexClause) jsonNode {
	// If we have an Object (ON clause), use DropIndexClause type
	if c.Object != nil {
		node := jsonNode{
			"$type": "DropIndexClause",
		}
		if c.Index != nil {
			node["Index"] = identifierToJSON(c.Index)
		}
		node["Object"] = schemaObjectNameToJSON(c.Object)
		if len(c.Options) > 0 {
			options := make([]jsonNode, len(c.Options))
			for i, opt := range c.Options {
				options[i] = dropIndexOptionToJSON(opt)
			}
			node["Options"] = options
		}
		return addSpan(node, frag(c))
	}

	// Otherwise use BackwardsCompatibleDropIndexClause for backwards-compatible syntax
	node := jsonNode{
		"$type": "BackwardsCompatibleDropIndexClause",
	}
	if c.LegacyIndex != nil {
		node["Index"] = childObjectNameToJSON(c.LegacyIndex)
	} else if c.Index != nil {
		// Just index name without object - use identifier
		node["Index"] = identifierToJSON(c.Index)
	}
	return addSpan(node, frag(c))
}

// childObjectNameToJSON converts a SchemaObjectName to a ChildObjectName JSON format
// where the BaseIdentifier is the parent and the last identifier becomes ChildIdentifier
func childObjectNameToJSON(s *ast.SchemaObjectName) jsonNode {
	node := jsonNode{
		"$type": "ChildObjectName",
		"Count": s.Count,
	}

	// For ChildObjectName: BaseIdentifier is the parent, ChildIdentifier is the actual child
	if s.Count >= 2 {
		// For 2-part: BaseIdentifier.ChildIdentifier
		if s.SchemaIdentifier != nil {
			node["BaseIdentifier"] = identifierToJSON(s.SchemaIdentifier)
		}
		if s.BaseIdentifier != nil {
			node["ChildIdentifier"] = identifierToJSON(s.BaseIdentifier)
		}
	}

	// For 3+ parts: add DatabaseIdentifier/SchemaIdentifier
	if s.Count >= 3 {
		if s.DatabaseIdentifier != nil {
			node["SchemaIdentifier"] = identifierToJSON(s.DatabaseIdentifier)
		}
	}

	if s.Count >= 4 {
		if s.ServerIdentifier != nil {
			node["DatabaseIdentifier"] = identifierToJSON(s.ServerIdentifier)
		}
	}

	// Add identifiers array
	if len(s.Identifiers) > 0 {
		idents := make([]jsonNode, len(s.Identifiers))
		for i, id := range s.Identifiers {
			idents[i] = jsonNode{"$ref": "Identifier"}
			_ = id
		}
		node["Identifiers"] = idents
	}

	return addSpan(node, frag(s))
}

func dropIndexOptionToJSON(opt ast.DropIndexOption) jsonNode {
	switch o := opt.(type) {
	case *ast.OnlineIndexOption:
		node := jsonNode{
			"$type":       "OnlineIndexOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		if o.LowPriorityLockWaitOption != nil {
			node["LowPriorityLockWaitOption"] = onlineIndexLowPriorityLockWaitOptionToJSON(o.LowPriorityLockWaitOption)
		}
		return addSpan(node, frag(opt))
	case *ast.MoveToDropIndexOption:
		node := jsonNode{
			"$type":      "MoveToDropIndexOption",
			"OptionKind": o.OptionKind,
		}
		if o.MoveTo != nil {
			node["MoveTo"] = fileGroupOrPartitionSchemeToJSON(o.MoveTo)
		}
		return addSpan(node, frag(opt))
	case *ast.FileStreamOnDropIndexOption:
		node := jsonNode{
			"$type":      "FileStreamOnDropIndexOption",
			"OptionKind": o.OptionKind,
		}
		if o.FileStreamOn != nil {
			node["FileStreamOn"] = identifierOrValueExpressionToJSON(o.FileStreamOn)
		}
		return addSpan(node, frag(opt))
	case *ast.DataCompressionOption:
		return addSpan(jsonNode{
			"$type":            "DataCompressionOption",
			"CompressionLevel": o.CompressionLevel,
			"OptionKind":       o.OptionKind,
		}, frag(opt))
	case *ast.WaitAtLowPriorityOption:
		node := jsonNode{
			"$type":      "WaitAtLowPriorityOption",
			"OptionKind": o.OptionKind,
		}
		if len(o.Options) > 0 {
			options := make([]jsonNode, len(o.Options))
			for i, opt := range o.Options {
				options[i] = lowPriorityLockWaitOptionToJSON(opt)
			}
			node["Options"] = options
		}
		return addSpan(node, frag(opt))
	case *ast.IndexExpressionOption:
		node := jsonNode{
			"$type":      "IndexExpressionOption",
			"OptionKind": o.OptionKind,
		}
		if o.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(o.Expression)
		}
		return addSpan(node, frag(opt))
	}
	return addSpan(jsonNode{}, frag(opt))
}

func dropStatisticsStatementToJSON(s *ast.DropStatisticsStatement) jsonNode {
	node := jsonNode{
		"$type": "DropStatisticsStatement",
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = childObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func dropPartitionFunctionStatementToJSON(s *ast.DropPartitionFunctionStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropPartitionFunctionStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropPartitionSchemeStatementToJSON(s *ast.DropPartitionSchemeStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropPartitionSchemeStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropApplicationRoleStatementToJSON(s *ast.DropApplicationRoleStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropApplicationRoleStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropCertificateStatementToJSON(s *ast.DropCertificateStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropCertificateStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropMasterKeyStatementToJSON(s *ast.DropMasterKeyStatement) jsonNode {
	return addSpan(jsonNode{
		"$type": "DropMasterKeyStatement",
	}, frag(s))
}

func dropXmlSchemaCollectionStatementToJSON(s *ast.DropXmlSchemaCollectionStatement) jsonNode {
	node := jsonNode{
		"$type": "DropXmlSchemaCollectionStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropContractStatementToJSON(s *ast.DropContractStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropContractStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropEndpointStatementToJSON(s *ast.DropEndpointStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropEndpointStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropMessageTypeStatementToJSON(s *ast.DropMessageTypeStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropMessageTypeStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropQueueStatementToJSON(s *ast.DropQueueStatement) jsonNode {
	node := jsonNode{
		"$type": "DropQueueStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropRemoteServiceBindingStatementToJSON(s *ast.DropRemoteServiceBindingStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropRemoteServiceBindingStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropRouteStatementToJSON(s *ast.DropRouteStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropRouteStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropServiceStatementToJSON(s *ast.DropServiceStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropServiceStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropEventNotificationStatementToJSON(s *ast.DropEventNotificationStatement) jsonNode {
	node := jsonNode{
		"$type": "DropEventNotificationStatement",
	}
	if len(s.Notifications) > 0 {
		notifications := make([]jsonNode, len(s.Notifications))
		for i, n := range s.Notifications {
			notifications[i] = identifierToJSON(n)
		}
		node["Notifications"] = notifications
	}
	if s.Scope != nil {
		scope := jsonNode{
			"$type":  "EventNotificationObjectScope",
			"Target": s.Scope.Target,
		}
		if s.Scope.QueueName != nil {
			scope["QueueName"] = schemaObjectNameToJSON(s.Scope.QueueName)
		}
		node["Scope"] = scope
	}
	return addSpan(node, frag(s))
}

func dropEventSessionStatementToJSON(s *ast.DropEventSessionStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropEventSessionStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.SessionScope != "" {
		node["SessionScope"] = s.SessionScope
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func dropExternalDataSourceStatementToJSON(s *ast.DropExternalDataSourceStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalDataSourceStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropExternalFileFormatStatementToJSON(s *ast.DropExternalFileFormatStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalFileFormatStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func dropExternalResourcePoolStatementToJSON(s *ast.DropExternalResourcePoolStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalResourcePoolStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropExternalModelStatementToJSON(s *ast.DropExternalModelStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropExternalModelStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropWorkloadGroupStatementToJSON(s *ast.DropWorkloadGroupStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropWorkloadGroupStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropWorkloadClassifierStatementToJSON(s *ast.DropWorkloadClassifierStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropWorkloadClassifierStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(opt))
	case *ast.ClassifierMemberNameOption:
		node := jsonNode{
			"$type":      "ClassifierMemberNameOption",
			"OptionType": o.OptionType,
		}
		if o.MemberName != nil {
			node["MemberName"] = stringLiteralToJSON(o.MemberName)
		}
		return addSpan(node, frag(opt))
	case *ast.ClassifierWlmContextOption:
		node := jsonNode{
			"$type":      "ClassifierWlmContextOption",
			"OptionType": o.OptionType,
		}
		if o.WlmContext != nil {
			node["WlmContext"] = stringLiteralToJSON(o.WlmContext)
		}
		return addSpan(node, frag(opt))
	case *ast.ClassifierStartTimeOption:
		node := jsonNode{
			"$type":      "ClassifierStartTimeOption",
			"OptionType": o.OptionType,
		}
		if o.Time != nil {
			node["Time"] = wlmTimeLiteralToJSON(o.Time)
		}
		return addSpan(node, frag(opt))
	case *ast.ClassifierEndTimeOption:
		node := jsonNode{
			"$type":      "ClassifierEndTimeOption",
			"OptionType": o.OptionType,
		}
		if o.Time != nil {
			node["Time"] = wlmTimeLiteralToJSON(o.Time)
		}
		return addSpan(node, frag(opt))
	case *ast.ClassifierWlmLabelOption:
		node := jsonNode{
			"$type":      "ClassifierWlmLabelOption",
			"OptionType": o.OptionType,
		}
		if o.WlmLabel != nil {
			node["WlmLabel"] = stringLiteralToJSON(o.WlmLabel)
		}
		return addSpan(node, frag(opt))
	case *ast.ClassifierImportanceOption:
		return addSpan(jsonNode{
			"$type":      "ClassifierImportanceOption",
			"Importance": o.Importance,
			"OptionType": o.OptionType,
		}, frag(opt))
	default:
		return addSpan(jsonNode{}, frag(opt))
	}
}

func wlmTimeLiteralToJSON(t *ast.WlmTimeLiteral) jsonNode {
	node := jsonNode{
		"$type": "WlmTimeLiteral",
	}
	if t.TimeString != nil {
		node["TimeString"] = stringLiteralToJSON(t.TimeString)
	}
	return addSpan(node, frag(t))
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
		return addSpan(node, frag(param))
	case *ast.WorkloadGroupImportanceParameter:
		return addSpan(jsonNode{
			"$type":          "WorkloadGroupImportanceParameter",
			"ParameterType":  param.ParameterType,
			"ParameterValue": param.ParameterValue,
		}, frag(param))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func sequenceOptionToJSON(opt interface{}) jsonNode {
	switch o := opt.(type) {
	case *ast.SequenceOption:
		return addSpan(jsonNode{
			"$type":      "SequenceOption",
			"OptionKind": o.OptionKind,
			"NoValue":    o.NoValue,
		}, frag(o))
	case *ast.ScalarExpressionSequenceOption:
		node := jsonNode{
			"$type":      "ScalarExpressionSequenceOption",
			"OptionKind": o.OptionKind,
			"NoValue":    o.NoValue,
		}
		if o.OptionValue != nil {
			node["OptionValue"] = scalarExpressionToJSON(o.OptionValue)
		}
		return addSpan(node, frag(o))
	case *ast.DataTypeSequenceOption:
		node := jsonNode{
			"$type":      "DataTypeSequenceOption",
			"OptionKind": o.OptionKind,
			"NoValue":    o.NoValue,
		}
		if o.DataType != nil {
			node["DataType"] = dataTypeReferenceToJSON(o.DataType)
		}
		return addSpan(node, frag(o))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(l))
}

func dbccOptionToJSON(o *ast.DbccOption) jsonNode {
	return addSpan(jsonNode{
		"$type":      "DbccOption",
		"OptionKind": o.OptionKind,
	}, frag(o))
}

func dropTypeStatementToJSON(s *ast.DropTypeStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropTypeStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func dropUserStatementToJSON(s *ast.DropUserStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropUserStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropRoleStatementToJSON(s *ast.DropRoleStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropRoleStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropAssemblyStatementToJSON(s *ast.DropAssemblyStatement) jsonNode {
	node := jsonNode{
		"$type":            "DropAssemblyStatement",
		"WithNoDependents": s.WithNoDependents,
	}
	if len(s.Objects) > 0 {
		objects := make([]jsonNode, len(s.Objects))
		for i, obj := range s.Objects {
			objects[i] = schemaObjectNameToJSON(obj)
		}
		node["Objects"] = objects
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func dropAsymmetricKeyStatementToJSON(s *ast.DropAsymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type":             "DropAsymmetricKeyStatement",
		"RemoveProviderKey": s.RemoveProviderKey,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func dropSymmetricKeyStatementToJSON(s *ast.DropSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type":             "DropSymmetricKeyStatement",
		"RemoveProviderKey": s.RemoveProviderKey,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func alterTableFileTableNamespaceStatementToJSON(s *ast.AlterTableFileTableNamespaceStatement) jsonNode {
	node := jsonNode{
		"$type":    "AlterTableFileTableNamespaceStatement",
		"IsEnable": s.IsEnable,
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func tableSwitchOptionToJSON(opt ast.TableSwitchOption) jsonNode {
	switch o := opt.(type) {
	case *ast.TruncateTargetTableSwitchOption:
		return addSpan(jsonNode{
			"$type":          "TruncateTargetTableSwitchOption",
			"TruncateTarget": o.TruncateTarget,
			"OptionKind":     o.OptionKind,
		}, frag(opt))
	case *ast.LowPriorityLockWaitTableSwitchOption:
		node := jsonNode{
			"$type":      "LowPriorityLockWaitTableSwitchOption",
			"OptionKind": o.OptionKind,
		}
		if len(o.Options) > 0 {
			opts := make([]jsonNode, len(o.Options))
			for i, subOpt := range o.Options {
				opts[i] = lowPriorityLockWaitOptionToJSON(subOpt)
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownSwitchOption"}, frag(opt))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func alterTableRebuildStatementToJSON(s *ast.AlterTableRebuildStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterTableRebuildStatement",
	}
	if s.Partition != nil {
		node["Partition"] = partitionSpecifierToJSON(s.Partition)
	}
	if len(s.IndexOptions) > 0 {
		var opts []jsonNode
		for _, opt := range s.IndexOptions {
			opts = append(opts, indexOptionToJSON(opt))
		}
		node["IndexOptions"] = opts
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return addSpan(node, frag(s))
}

func alterTableAlterPartitionStatementToJSON(s *ast.AlterTableAlterPartitionStatement) jsonNode {
	node := jsonNode{
		"$type":   "AlterTableAlterPartitionStatement",
		"IsSplit": s.IsSplit,
	}
	if s.BoundaryValue != nil {
		node["BoundaryValue"] = scalarExpressionToJSON(s.BoundaryValue)
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return addSpan(node, frag(s))
}

func alterTableChangeTrackingStatementToJSON(s *ast.AlterTableChangeTrackingModificationStatement) jsonNode {
	node := jsonNode{
		"$type":               "AlterTableChangeTrackingModificationStatement",
		"IsEnable":            s.IsEnable,
		"TrackColumnsUpdated": s.TrackColumnsUpdated,
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(o))
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
	return addSpan(node, frag(r))
}

func ledgerTableOptionToJSON(o *ast.LedgerTableOption) jsonNode {
	node := jsonNode{
		"$type":       "LedgerTableOption",
		"OptionState": o.OptionState,
		"AppendOnly":  o.AppendOnly,
	}
	if o.LedgerViewOption != nil {
		node["LedgerViewOption"] = ledgerViewOptionToJSON(o.LedgerViewOption)
	}
	node["OptionKind"] = o.OptionKind
	return addSpan(node, frag(o))
}

func ledgerViewOptionToJSON(o *ast.LedgerViewOption) jsonNode {
	node := jsonNode{
		"$type": "LedgerViewOption",
	}
	if o.ViewName != nil {
		node["ViewName"] = schemaObjectNameToJSON(o.ViewName)
	}
	if o.TransactionIdColumnName != nil {
		node["TransactionIdColumnName"] = identifierToJSON(o.TransactionIdColumnName)
	}
	if o.SequenceNumberColumnName != nil {
		node["SequenceNumberColumnName"] = identifierToJSON(o.SequenceNumberColumnName)
	}
	if o.OperationTypeColumnName != nil {
		node["OperationTypeColumnName"] = identifierToJSON(o.OperationTypeColumnName)
	}
	if o.OperationTypeDescColumnName != nil {
		node["OperationTypeDescColumnName"] = identifierToJSON(o.OperationTypeDescColumnName)
	}
	node["OptionKind"] = o.OptionKind
	return addSpan(node, frag(o))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(opt))
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(opt))
	case *ast.ExternalFileFormatLiteralOption:
		node := jsonNode{
			"$type":      "ExternalFileFormatLiteralOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.ExternalFileFormatUseDefaultTypeOption:
		return addSpan(jsonNode{
			"$type":                            "ExternalFileFormatUseDefaultTypeOption",
			"ExternalFileFormatUseDefaultType": o.ExternalFileFormatUseDefaultType,
			"OptionKind":                       o.OptionKind,
		}, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownExternalFileFormatOption"}, frag(opt))
	}
}

func createExternalTableStatementToJSON(s *ast.CreateExternalTableStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalTableStatement",
	}
	if s.SchemaObjectName != nil {
		node["SchemaObjectName"] = schemaObjectNameToJSON(s.SchemaObjectName)
	}
	if len(s.ColumnDefinitions) > 0 {
		cols := make([]jsonNode, len(s.ColumnDefinitions))
		for i, col := range s.ColumnDefinitions {
			cols[i] = externalTableColumnDefinitionToJSON(col)
		}
		node["ColumnDefinitions"] = cols
	}
	if s.DataSource != nil {
		node["DataSource"] = identifierToJSON(s.DataSource)
	}
	if len(s.ExternalTableOptions) > 0 {
		opts := make([]jsonNode, len(s.ExternalTableOptions))
		for i, opt := range s.ExternalTableOptions {
			opts[i] = externalTableOptionItemToJSON(opt)
		}
		node["ExternalTableOptions"] = opts
	}
	if s.SelectStatement != nil {
		node["SelectStatement"] = selectStatementToJSON(s.SelectStatement)
	}
	return addSpan(node, frag(s))
}

func externalTableOptionItemToJSON(opt ast.ExternalTableOptionItem) jsonNode {
	switch o := opt.(type) {
	case *ast.ExternalTableLiteralOrIdentifierOption:
		return addSpan(externalTableLiteralOrIdentifierOptionToJSON(o), frag(opt))
	case *ast.ExternalTableRejectTypeOption:
		return addSpan(externalTableRejectTypeOptionToJSON(o), frag(opt))
	case *ast.ExternalTableDistributionOption:
		return addSpan(externalTableDistributionOptionToJSON(o), frag(opt))
	default:
		return addSpan(jsonNode{}, frag(opt))
	}
}

func externalTableRejectTypeOptionToJSON(opt *ast.ExternalTableRejectTypeOption) jsonNode {
	return addSpan(jsonNode{
		"$type":      "ExternalTableRejectTypeOption",
		"Value":      opt.Value,
		"OptionKind": opt.OptionKind,
	}, frag(opt))
}

func externalTableDistributionOptionToJSON(opt *ast.ExternalTableDistributionOption) jsonNode {
	node := jsonNode{
		"$type":      "ExternalTableDistributionOption",
		"OptionKind": opt.OptionKind,
	}
	if opt.Value != nil {
		switch v := opt.Value.(type) {
		case *ast.ExternalTableShardedDistributionPolicy:
			policyNode := jsonNode{
				"$type": "ExternalTableShardedDistributionPolicy",
			}
			if v.ShardingColumn != nil {
				policyNode["ShardingColumn"] = identifierToJSON(v.ShardingColumn)
			}
			node["Value"] = addSpan(policyNode, frag(v))
		case *ast.ExternalTableRoundRobinDistributionPolicy:
			node["Value"] = addSpan(jsonNode{
				"$type": "ExternalTableRoundRobinDistributionPolicy",
			}, frag(v))
		case *ast.ExternalTableReplicatedDistributionPolicy:
			node["Value"] = addSpan(jsonNode{
				"$type": "ExternalTableReplicatedDistributionPolicy",
			}, frag(v))
		}
	}
	return addSpan(node, frag(opt))
}

func externalTableColumnDefinitionToJSON(col *ast.ExternalTableColumnDefinition) jsonNode {
	node := jsonNode{
		"$type": "ExternalTableColumnDefinition",
	}
	if col.ColumnDefinition != nil {
		node["ColumnDefinition"] = columnDefinitionBaseToJSON(col.ColumnDefinition)
	}
	if col.NullableConstraint != nil {
		node["NullableConstraint"] = nullableConstraintToJSON(col.NullableConstraint)
	}
	return addSpan(node, frag(col))
}

func externalTableLiteralOrIdentifierOptionToJSON(opt *ast.ExternalTableLiteralOrIdentifierOption) jsonNode {
	node := jsonNode{
		"$type": "ExternalTableLiteralOrIdentifierOption",
	}
	if opt.Value != nil {
		node["Value"] = identifierOrValueExpressionToJSON(opt.Value)
	}
	if opt.OptionKind != "" {
		node["OptionKind"] = opt.OptionKind
	}
	return addSpan(node, frag(opt))
}

func createExternalLanguageStatementToJSON(s *ast.CreateExternalLanguageStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalLanguageStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ExternalLanguageFiles) > 0 {
		files := make([]jsonNode, len(s.ExternalLanguageFiles))
		for i, f := range s.ExternalLanguageFiles {
			files[i] = externalLanguageFileOptionToJSON(f)
		}
		node["ExternalLanguageFiles"] = files
	}
	return addSpan(node, frag(s))
}

func externalLanguageFileOptionToJSON(f *ast.ExternalLanguageFileOption) jsonNode {
	node := jsonNode{
		"$type": "ExternalLanguageFileOption",
	}
	if f.Content != nil {
		node["Content"] = scalarExpressionToJSON(f.Content)
	}
	if f.FileName != nil {
		node["FileName"] = scalarExpressionToJSON(f.FileName)
	}
	if f.Platform != nil {
		node["Platform"] = identifierToJSON(f.Platform)
	}
	if f.Parameters != nil {
		node["Parameters"] = scalarExpressionToJSON(f.Parameters)
	}
	if f.EnvironmentVariables != nil {
		node["EnvironmentVariables"] = scalarExpressionToJSON(f.EnvironmentVariables)
	}
	return addSpan(node, frag(f))
}

func createExternalLibraryStatementToJSON(s *ast.CreateExternalLibraryStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalLibraryStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Language != nil {
		node["Language"] = scalarExpressionToJSON(s.Language)
	}
	if len(s.ExternalLibraryFiles) > 0 {
		files := make([]jsonNode, len(s.ExternalLibraryFiles))
		for i, f := range s.ExternalLibraryFiles {
			files[i] = externalLibraryFileOptionToJSON(f)
		}
		node["ExternalLibraryFiles"] = files
	}
	return addSpan(node, frag(s))
}

func externalLibraryFileOptionToJSON(f *ast.ExternalLibraryFileOption) jsonNode {
	node := jsonNode{
		"$type": "ExternalLibraryFileOption",
	}
	if f.Content != nil {
		node["Content"] = scalarExpressionToJSON(f.Content)
	}
	if f.Platform != nil {
		node["Platform"] = identifierToJSON(f.Platform)
	}
	return addSpan(node, frag(f))
}

func createEventSessionStatementToJSON(s *ast.CreateEventSessionStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateEventSessionStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.SessionScope != "" {
		node["SessionScope"] = s.SessionScope
	}
	if len(s.EventDeclarations) > 0 {
		events := make([]jsonNode, len(s.EventDeclarations))
		for i, e := range s.EventDeclarations {
			events[i] = eventDeclarationToJSON(e)
		}
		node["EventDeclarations"] = events
	}
	if len(s.TargetDeclarations) > 0 {
		targets := make([]jsonNode, len(s.TargetDeclarations))
		for i, t := range s.TargetDeclarations {
			targets[i] = targetDeclarationToJSON(t)
		}
		node["TargetDeclarations"] = targets
	}
	if len(s.SessionOptions) > 0 {
		opts := make([]jsonNode, len(s.SessionOptions))
		for i, o := range s.SessionOptions {
			opts[i] = sessionOptionToJSON(o)
		}
		node["SessionOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func alterEventSessionStatementToJSON(s *ast.AlterEventSessionStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterEventSessionStatement",
	}
	if s.StatementType != "" {
		node["StatementType"] = s.StatementType
	}
	// DropEventDeclarations comes before Name in JSON
	if len(s.DropEventDeclarations) > 0 {
		events := make([]jsonNode, len(s.DropEventDeclarations))
		for i, e := range s.DropEventDeclarations {
			events[i] = eventSessionObjectNameToJSON(e)
		}
		node["DropEventDeclarations"] = events
	}
	// DropTargetDeclarations comes before Name in JSON
	if len(s.DropTargetDeclarations) > 0 {
		targets := make([]jsonNode, len(s.DropTargetDeclarations))
		for i, t := range s.DropTargetDeclarations {
			targets[i] = eventSessionObjectNameToJSON(t)
		}
		node["DropTargetDeclarations"] = targets
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.SessionScope != "" {
		node["SessionScope"] = s.SessionScope
	}
	if len(s.EventDeclarations) > 0 {
		events := make([]jsonNode, len(s.EventDeclarations))
		for i, e := range s.EventDeclarations {
			events[i] = eventDeclarationToJSON(e)
		}
		node["EventDeclarations"] = events
	}
	if len(s.TargetDeclarations) > 0 {
		targets := make([]jsonNode, len(s.TargetDeclarations))
		for i, t := range s.TargetDeclarations {
			targets[i] = targetDeclarationToJSON(t)
		}
		node["TargetDeclarations"] = targets
	}
	if len(s.SessionOptions) > 0 {
		opts := make([]jsonNode, len(s.SessionOptions))
		for i, o := range s.SessionOptions {
			opts[i] = sessionOptionToJSON(o)
		}
		node["SessionOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func alterAuthorizationStatementToJSON(s *ast.AlterAuthorizationStatement) jsonNode {
	node := jsonNode{
		"$type":         "AlterAuthorizationStatement",
		"ToSchemaOwner": s.ToSchemaOwner,
	}
	if s.SecurityTargetObject != nil {
		node["SecurityTargetObject"] = securityTargetObjectToJSON(s.SecurityTargetObject)
	}
	if s.PrincipalName != nil {
		node["PrincipalName"] = identifierToJSON(s.PrincipalName)
	}
	return addSpan(node, frag(s))
}

func eventDeclarationToJSON(e *ast.EventDeclaration) jsonNode {
	node := jsonNode{
		"$type": "EventDeclaration",
	}
	if e.ObjectName != nil {
		node["ObjectName"] = eventSessionObjectNameToJSON(e.ObjectName)
	}
	if len(e.EventDeclarationSetParameters) > 0 {
		params := make([]jsonNode, len(e.EventDeclarationSetParameters))
		for i, p := range e.EventDeclarationSetParameters {
			params[i] = eventDeclarationSetParameterToJSON(p)
		}
		node["EventDeclarationSetParameters"] = params
	}
	if len(e.EventDeclarationActionParameters) > 0 {
		actions := make([]jsonNode, len(e.EventDeclarationActionParameters))
		for i, a := range e.EventDeclarationActionParameters {
			actions[i] = eventSessionObjectNameToJSON(a)
		}
		node["EventDeclarationActionParameters"] = actions
	}
	if e.EventDeclarationPredicateParameter != nil {
		node["EventDeclarationPredicateParameter"] = booleanExpressionToJSON(e.EventDeclarationPredicateParameter)
	}
	return addSpan(node, frag(e))
}

func targetDeclarationToJSON(t *ast.TargetDeclaration) jsonNode {
	node := jsonNode{
		"$type": "TargetDeclaration",
	}
	if t.ObjectName != nil {
		node["ObjectName"] = eventSessionObjectNameToJSON(t.ObjectName)
	}
	if len(t.TargetDeclarationParameters) > 0 {
		params := make([]jsonNode, len(t.TargetDeclarationParameters))
		for i, p := range t.TargetDeclarationParameters {
			params[i] = eventDeclarationSetParameterToJSON(p)
		}
		node["TargetDeclarationParameters"] = params
	}
	return addSpan(node, frag(t))
}

func eventDeclarationSetParameterToJSON(p *ast.EventDeclarationSetParameter) jsonNode {
	node := jsonNode{
		"$type": "EventDeclarationSetParameter",
	}
	if p.EventField != nil {
		node["EventField"] = identifierToJSON(p.EventField)
	}
	if p.EventValue != nil {
		node["EventValue"] = scalarExpressionToJSON(p.EventValue)
	}
	return addSpan(node, frag(p))
}

func sessionOptionToJSON(o ast.SessionOption) jsonNode {
	switch opt := o.(type) {
	case *ast.LiteralSessionOption:
		node := jsonNode{
			"$type": "LiteralSessionOption",
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		if opt.Unit != "" {
			node["Unit"] = opt.Unit
		}
		if opt.OptionKind != "" {
			node["OptionKind"] = opt.OptionKind
		}
		return addSpan(node, frag(o))
	case *ast.OnOffSessionOption:
		node := jsonNode{
			"$type": "OnOffSessionOption",
		}
		if opt.OptionState != "" {
			node["OptionState"] = opt.OptionState
		}
		if opt.OptionKind != "" {
			node["OptionKind"] = opt.OptionKind
		}
		return addSpan(node, frag(o))
	case *ast.EventRetentionSessionOption:
		node := jsonNode{
			"$type": "EventRetentionSessionOption",
		}
		if opt.Value != "" {
			node["Value"] = opt.Value
		}
		if opt.OptionKind != "" {
			node["OptionKind"] = opt.OptionKind
		}
		return addSpan(node, frag(o))
	case *ast.MaxDispatchLatencySessionOption:
		node := jsonNode{
			"$type":      "MaxDispatchLatencySessionOption",
			"IsInfinite": opt.IsInfinite,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		if opt.OptionKind != "" {
			node["OptionKind"] = opt.OptionKind
		}
		return addSpan(node, frag(o))
	case *ast.MemoryPartitionSessionOption:
		node := jsonNode{
			"$type": "MemoryPartitionSessionOption",
		}
		if opt.Value != "" {
			node["Value"] = opt.Value
		}
		if opt.OptionKind != "" {
			node["OptionKind"] = opt.OptionKind
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownSessionOption"}, frag(o))
	}
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
	return addSpan(node, frag(s))
}

func insertBulkColumnDefinitionToJSON(c *ast.InsertBulkColumnDefinition) jsonNode {
	node := jsonNode{
		"$type": "InsertBulkColumnDefinition",
	}
	if c.Column != nil {
		node["Column"] = columnDefinitionBaseToJSON(c.Column)
	}
	// Always include NullNotNull - use "NotSpecified" if empty
	nullNotNull := c.NullNotNull
	if nullNotNull == "" || nullNotNull == "Unspecified" {
		nullNotNull = "NotSpecified"
	}
	node["NullNotNull"] = nullNotNull
	return addSpan(node, frag(c))
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
	if c.Collation != nil {
		node["Collation"] = identifierToJSON(c.Collation)
	}
	return addSpan(node, frag(c))
}

// normalizeRowsetOptionsJSON normalizes a JSON string for ROWSET_OPTIONS
// by removing whitespace and uppercasing keys to match ScriptDOM behavior
func normalizeRowsetOptionsJSON(jsonStr string) string {
	// Parse and re-serialize the JSON to normalize it
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If parsing fails, return as-is
		return jsonStr
	}

	// Normalize keys to uppercase and values
	normalized := normalizeJSONObject(data)

	// Re-serialize without extra whitespace
	result, err := json.Marshal(normalized)
	if err != nil {
		return jsonStr
	}
	return string(result)
}

// normalizeJSONObject recursively normalizes JSON object keys to uppercase
func normalizeJSONObject(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range data {
		upperKey := strings.ToUpper(k)
		switch val := v.(type) {
		case map[string]interface{}:
			result[upperKey] = normalizeJSONObject(val)
		case []interface{}:
			result[upperKey] = normalizeJSONArray(val)
		default:
			result[upperKey] = v
		}
	}
	return result
}

// normalizeJSONArray recursively normalizes JSON array values
func normalizeJSONArray(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))
	for i, v := range data {
		switch val := v.(type) {
		case map[string]interface{}:
			result[i] = normalizeJSONObject(val)
		case []interface{}:
			result[i] = normalizeJSONArray(val)
		case string:
			// Uppercase string values in arrays
			result[i] = strings.ToUpper(val)
		default:
			result[i] = v
		}
	}
	return result
}

func bulkInsertOptionToJSON(opt ast.BulkInsertOption) jsonNode {
	switch o := opt.(type) {
	case *ast.BulkInsertOptionBase:
		return addSpan(jsonNode{
			"$type":      "BulkInsertOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.LiteralBulkInsertOption:
		node := jsonNode{
			"$type":      "LiteralBulkInsertOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			// For RowsetOptions, normalize the JSON string value
			if o.OptionKind == "RowsetOptions" {
				if strLit, ok := o.Value.(*ast.StringLiteral); ok {
					normalizedValue := normalizeRowsetOptionsJSON(strLit.Value)
					normalizedLit := &ast.StringLiteral{
						LiteralType:   strLit.LiteralType,
						IsNational:    strLit.IsNational,
						IsLargeObject: strLit.IsLargeObject,
						Value:         normalizedValue,
					}
					node["Value"] = scalarExpressionToJSON(normalizedLit)
				} else {
					node["Value"] = scalarExpressionToJSON(o.Value)
				}
			} else {
				node["Value"] = scalarExpressionToJSON(o.Value)
			}
		}
		return addSpan(node, frag(opt))
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
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownBulkInsertOption"}, frag(opt))
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
	return addSpan(node, frag(s))
}

func copyStatementToJSON(s *ast.CopyStatement) jsonNode {
	node := jsonNode{
		"$type": "CopyStatement",
	}
	if len(s.From) > 0 {
		from := make([]jsonNode, len(s.From))
		for i, f := range s.From {
			from[i] = scalarExpressionToJSON(f)
		}
		node["From"] = from
	}
	if s.Into != nil {
		node["Into"] = schemaObjectNameToJSON(s.Into)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = copyOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
}

func copyOptionToJSON(o *ast.CopyOption) jsonNode {
	node := jsonNode{
		"$type": "CopyOption",
		"Kind":  normalizeCopyOptionKind(o.Kind),
	}
	if o.Value != nil {
		node["Value"] = copyOptionValueToJSON(o.Value)
	}
	return addSpan(node, frag(o))
}

// normalizeCopyOptionKind converts option names to PascalCase
func normalizeCopyOptionKind(kind string) string {
	// Map common option names
	optionMap := map[string]string{
		"FILE_TYPE":            "File_Type",
		"FIELDTERMINATOR":      "FieldTerminator",
		"ROWTERMINATOR":        "RowTerminator",
		"FIELDQUOTE":           "FieldQuote",
		"DATEFORMAT":           "DateFormat",
		"ENCODING":             "Encoding",
		"MAXERRORS":            "MaxErrors",
		"ERRORFILE":            "ErrorFile",
		"FIRSTROW":             "FirstRow",
		"CREDENTIAL":           "Credential",
		"IDENTITY_INSERT":      "Identity_Insert",
		"COMPRESSION":          "Compression",
		"FILE_FORMAT":          "File_Format",
		"ERRORFILE_CREDENTIAL": "ErrorFileCredential",
		"COLUMNOPTIONS":        "ColumnOptions",
	}
	upper := strings.ToUpper(kind)
	if mapped, ok := optionMap[upper]; ok {
		return mapped
	}
	return kind
}

func copyOptionValueToJSON(v ast.CopyOptionValue) jsonNode {
	switch val := v.(type) {
	case *ast.SingleValueTypeCopyOption:
		node := jsonNode{
			"$type": "SingleValueTypeCopyOption",
		}
		if val.SingleValue != nil {
			node["SingleValue"] = identifierOrValueExpressionToJSON(val.SingleValue)
		}
		return addSpan(node, frag(v))
	case *ast.CopyCredentialOption:
		node := jsonNode{
			"$type": "CopyCredentialOption",
		}
		if val.Identity != nil {
			node["Identity"] = scalarExpressionToJSON(val.Identity)
		}
		if val.Secret != nil {
			node["Secret"] = scalarExpressionToJSON(val.Secret)
		}
		return addSpan(node, frag(v))
	case *ast.ListTypeCopyOption:
		node := jsonNode{
			"$type": "ListTypeCopyOption",
		}
		if len(val.Options) > 0 {
			opts := make([]jsonNode, len(val.Options))
			for i, opt := range val.Options {
				opts[i] = copyColumnOptionToJSON(opt)
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(v))
	}
	return nil
}

func copyColumnOptionToJSON(c *ast.CopyColumnOption) jsonNode {
	node := jsonNode{
		"$type": "CopyColumnOption",
	}
	if c.ColumnName != nil {
		node["ColumnName"] = identifierToJSON(c.ColumnName)
	}
	if c.DefaultValue != nil {
		node["DefaultValue"] = scalarExpressionToJSON(c.DefaultValue)
	}
	if c.FieldNumber != nil {
		node["FieldNumber"] = scalarExpressionToJSON(c.FieldNumber)
	}
	return addSpan(node, frag(c))
}

func alterUserStatementToJSON(s *ast.AlterUserStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterUserStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.UserOptions) > 0 {
		options := make([]jsonNode, len(s.UserOptions))
		for i, o := range s.UserOptions {
			options[i] = userOptionToJSON(o)
		}
		node["UserOptions"] = options
	}
	return addSpan(node, frag(s))
}

func alterRouteStatementToJSON(s *ast.AlterRouteStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterRouteStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.RouteOptions) > 0 {
		opts := make([]jsonNode, len(s.RouteOptions))
		for i, opt := range s.RouteOptions {
			opts[i] = routeOptionToJSON(opt)
		}
		node["RouteOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func alterSearchPropertyListStatementToJSON(s *ast.AlterSearchPropertyListStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterSearchPropertyListStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Action != nil {
		node["Action"] = searchPropertyListActionToJSON(s.Action)
	}
	return addSpan(node, frag(s))
}

func searchPropertyListActionToJSON(a ast.SearchPropertyListAction) jsonNode {
	switch action := a.(type) {
	case *ast.AddSearchPropertyListAction:
		node := jsonNode{
			"$type": "AddSearchPropertyListAction",
		}
		if action.PropertyName != nil {
			node["PropertyName"] = stringLiteralToJSON(action.PropertyName)
		}
		if action.Guid != nil {
			node["Guid"] = stringLiteralToJSON(action.Guid)
		}
		if action.Id != nil {
			node["Id"] = scalarExpressionToJSON(action.Id)
		}
		if action.Description != nil {
			node["Description"] = stringLiteralToJSON(action.Description)
		}
		return addSpan(node, frag(a))
	case *ast.DropSearchPropertyListAction:
		node := jsonNode{
			"$type": "DropSearchPropertyListAction",
		}
		if action.PropertyName != nil {
			node["PropertyName"] = stringLiteralToJSON(action.PropertyName)
		}
		return addSpan(node, frag(a))
	default:
		return addSpan(jsonNode{"$type": "UnknownSearchPropertyListAction"}, frag(a))
	}
}

func alterAssemblyStatementToJSON(s *ast.AlterAssemblyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterAssemblyStatement",
	}
	// Include IsDropAll if there are any files/params/options, or if it's true
	if s.IsDropAll || len(s.DropFiles) > 0 || len(s.AddFiles) > 0 || len(s.Parameters) > 0 || len(s.Options) > 0 {
		node["IsDropAll"] = s.IsDropAll
	}
	if len(s.DropFiles) > 0 {
		files := make([]jsonNode, len(s.DropFiles))
		for i, f := range s.DropFiles {
			files[i] = stringLiteralToJSON(f)
		}
		node["DropFiles"] = files
	}
	if len(s.AddFiles) > 0 {
		files := make([]jsonNode, len(s.AddFiles))
		for i, f := range s.AddFiles {
			files[i] = addFileSpecToJSON(f)
		}
		node["AddFiles"] = files
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = scalarExpressionToJSON(p)
		}
		node["Parameters"] = params
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			opts[i] = assemblyOptionToJSON(o)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
}

func assemblyOptionToJSON(o ast.AssemblyOptionBase) jsonNode {
	switch opt := o.(type) {
	case *ast.AssemblyOption:
		return addSpan(jsonNode{
			"$type":      "AssemblyOption",
			"OptionKind": opt.OptionKind,
		}, frag(o))
	case *ast.OnOffAssemblyOption:
		return addSpan(jsonNode{
			"$type":       "OnOffAssemblyOption",
			"OptionKind":  opt.OptionKind,
			"OptionState": opt.OptionState,
		}, frag(o))
	case *ast.PermissionSetAssemblyOption:
		return addSpan(jsonNode{
			"$type":               "PermissionSetAssemblyOption",
			"OptionKind":          opt.OptionKind,
			"PermissionSetOption": opt.PermissionSetOption,
		}, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownAssemblyOption"}, frag(o))
	}
}

func addFileSpecToJSON(f *ast.AddFileSpec) jsonNode {
	node := jsonNode{
		"$type": "AddFileSpec",
	}
	if f.File != nil {
		node["File"] = scalarExpressionToJSON(f.File)
	}
	if f.FileName != nil {
		node["FileName"] = stringLiteralToJSON(f.FileName)
	}
	return addSpan(node, frag(f))
}

func alterEndpointStatementToJSON(s *ast.AlterEndpointStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterEndpointStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.State != "" {
		node["State"] = s.State
	}
	if s.Affinity != nil {
		node["Affinity"] = endpointAffinityToJSON(s.Affinity)
	}
	if s.Protocol != "" {
		node["Protocol"] = s.Protocol
	}
	if len(s.ProtocolOptions) > 0 {
		opts := make([]jsonNode, len(s.ProtocolOptions))
		for i, opt := range s.ProtocolOptions {
			opts[i] = endpointProtocolOptionToJSON(opt)
		}
		node["ProtocolOptions"] = opts
	}
	if s.EndpointType != "" {
		node["EndpointType"] = s.EndpointType
	}
	if len(s.PayloadOptions) > 0 {
		opts := make([]jsonNode, len(s.PayloadOptions))
		for i, opt := range s.PayloadOptions {
			opts[i] = payloadOptionToJSON(opt)
		}
		node["PayloadOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func endpointAffinityToJSON(a *ast.EndpointAffinity) jsonNode {
	node := jsonNode{
		"$type": "EndpointAffinity",
	}
	if a.Kind != "" {
		node["Kind"] = a.Kind
	}
	if a.Value != nil {
		node["Value"] = scalarExpressionToJSON(a.Value)
	}
	return addSpan(node, frag(a))
}

func endpointProtocolOptionToJSON(opt ast.EndpointProtocolOption) jsonNode {
	switch o := opt.(type) {
	case *ast.LiteralEndpointProtocolOption:
		node := jsonNode{
			"$type": "LiteralEndpointProtocolOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.ListenerIPEndpointProtocolOption:
		node := jsonNode{
			"$type": "ListenerIPEndpointProtocolOption",
			"IsAll": o.IsAll,
		}
		if o.IPv4PartOne != nil {
			node["IPv4PartOne"] = ipv4ToJSON(o.IPv4PartOne)
		}
		if o.IPv4PartTwo != nil {
			node["IPv4PartTwo"] = ipv4ToJSON(o.IPv4PartTwo)
		}
		if o.IPv6 != nil {
			node["IPv6"] = scalarExpressionToJSON(o.IPv6)
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.AuthenticationEndpointProtocolOption:
		node := jsonNode{
			"$type":               "AuthenticationEndpointProtocolOption",
			"AuthenticationTypes": o.AuthenticationTypes,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.PortsEndpointProtocolOption:
		node := jsonNode{
			"$type":     "PortsEndpointProtocolOption",
			"PortTypes": o.PortTypes,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.CompressionEndpointProtocolOption:
		node := jsonNode{
			"$type":     "CompressionEndpointProtocolOption",
			"IsEnabled": o.IsEnabled,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownProtocolOption"}, frag(opt))
	}
}

func ipv4ToJSON(ip *ast.IPv4) jsonNode {
	node := jsonNode{
		"$type": "IPv4",
	}
	if ip.OctetOne != nil {
		node["OctetOne"] = scalarExpressionToJSON(ip.OctetOne)
	}
	if ip.OctetTwo != nil {
		node["OctetTwo"] = scalarExpressionToJSON(ip.OctetTwo)
	}
	if ip.OctetThree != nil {
		node["OctetThree"] = scalarExpressionToJSON(ip.OctetThree)
	}
	if ip.OctetFour != nil {
		node["OctetFour"] = scalarExpressionToJSON(ip.OctetFour)
	}
	return addSpan(node, frag(ip))
}

func payloadOptionToJSON(opt ast.PayloadOption) jsonNode {
	switch o := opt.(type) {
	case *ast.SoapMethod:
		node := jsonNode{
			"$type": "SoapMethod",
		}
		if o.Alias != nil {
			node["Alias"] = stringLiteralToJSON(o.Alias)
		}
		if o.Namespace != nil {
			node["Namespace"] = stringLiteralToJSON(o.Namespace)
		}
		if o.Action != "" {
			node["Action"] = o.Action
		}
		if o.Name != nil {
			node["Name"] = stringLiteralToJSON(o.Name)
		}
		if o.Format != "" {
			node["Format"] = o.Format
		}
		if o.Schema != "" {
			node["Schema"] = o.Schema
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.EnabledDisabledPayloadOption:
		node := jsonNode{
			"$type":     "EnabledDisabledPayloadOption",
			"IsEnabled": o.IsEnabled,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.AuthenticationPayloadOption:
		node := jsonNode{
			"$type":               "AuthenticationPayloadOption",
			"Protocol":            o.Protocol,
			"TryCertificateFirst": o.TryCertificateFirst,
		}
		if o.Certificate != nil {
			node["Certificate"] = identifierToJSON(o.Certificate)
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.EncryptionPayloadOption:
		node := jsonNode{
			"$type":             "EncryptionPayloadOption",
			"EncryptionSupport": o.EncryptionSupport,
			"AlgorithmPartOne":  o.AlgorithmPartOne,
			"AlgorithmPartTwo":  o.AlgorithmPartTwo,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.RolePayloadOption:
		node := jsonNode{
			"$type": "RolePayloadOption",
			"Role":  o.Role,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.LiteralPayloadOption:
		node := jsonNode{
			"$type": "LiteralPayloadOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.SchemaPayloadOption:
		node := jsonNode{
			"$type":      "SchemaPayloadOption",
			"IsStandard": o.IsStandard,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.CharacterSetPayloadOption:
		node := jsonNode{
			"$type": "CharacterSetPayloadOption",
			"IsSql": o.IsSql,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.SessionTimeoutPayloadOption:
		node := jsonNode{
			"$type":   "SessionTimeoutPayloadOption",
			"IsNever": o.IsNever,
		}
		if o.Timeout != nil {
			node["Timeout"] = scalarExpressionToJSON(o.Timeout)
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.WsdlPayloadOption:
		node := jsonNode{
			"$type":  "WsdlPayloadOption",
			"IsNone": o.IsNone,
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	case *ast.LoginTypePayloadOption:
		node := jsonNode{
			"$type":     "LoginTypePayloadOption",
			"IsWindows": o.IsWindows,
		}
		if o.Kind != "" {
			node["Kind"] = o.Kind
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownPayloadOption"}, frag(opt))
	}
}

func alterServiceStatementToJSON(s *ast.AlterServiceStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterServiceStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.QueueName != nil {
		node["QueueName"] = schemaObjectNameToJSON(s.QueueName)
	}
	if len(s.ServiceContracts) > 0 {
		contracts := make([]jsonNode, len(s.ServiceContracts))
		for i, c := range s.ServiceContracts {
			contracts[i] = serviceContractToJSON(c)
		}
		node["ServiceContracts"] = contracts
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func alterAsymmetricKeyStatementToJSON(s *ast.AlterAsymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterAsymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.AttestedBy != nil {
		node["AttestedBy"] = scalarExpressionToJSON(s.AttestedBy)
	}
	if s.Kind != "" {
		node["Kind"] = s.Kind
	}
	if s.EncryptionPassword != nil {
		node["EncryptionPassword"] = scalarExpressionToJSON(s.EncryptionPassword)
	}
	if s.DecryptionPassword != nil {
		node["DecryptionPassword"] = scalarExpressionToJSON(s.DecryptionPassword)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
			opts[i] = addSpan(optNode, frag(opt))
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
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
			opts[i] = addSpan(optNode, frag(opt))
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
}

func createFullTextStopListStatementToJSON(s *ast.CreateFullTextStopListStatement) jsonNode {
	node := jsonNode{
		"$type":            "CreateFullTextStopListStatement",
		"IsSystemStopList": s.IsSystemStopList,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.SourceStopListName != nil {
		node["SourceStopListName"] = identifierToJSON(s.SourceStopListName)
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	return addSpan(node, frag(s))
}

func alterFullTextStopListStatementToJSON(s *ast.AlterFullTextStopListStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterFullTextStopListStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Action != nil {
		node["Action"] = fullTextStopListActionToJSON(s.Action)
	}
	return addSpan(node, frag(s))
}

func fullTextStopListActionToJSON(a *ast.FullTextStopListAction) jsonNode {
	node := jsonNode{
		"$type": "FullTextStopListAction",
		"IsAdd": a.IsAdd,
		"IsAll": a.IsAll,
	}
	if a.StopWord != nil {
		node["StopWord"] = stringLiteralToJSON(a.StopWord)
	}
	if a.LanguageTerm != nil {
		node["LanguageTerm"] = identifierOrValueExpressionToJSON(a.LanguageTerm)
	}
	return addSpan(node, frag(a))
}

func dropFullTextStopListStatementToJSON(s *ast.DropFullTextStopListStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropFullTextStopListStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropFullTextCatalogStatementToJSON(s *ast.DropFullTextCatalogStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropFullTextCatalogStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func dropFulltextIndexStatementToJSON(s *ast.DropFulltextIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "DropFullTextIndexStatement",
	}
	if s.TableName != nil {
		node["TableName"] = schemaObjectNameToJSON(s.TableName)
	}
	return addSpan(node, frag(s))
}

func alterFulltextIndexStatementToJSON(s *ast.AlterFulltextIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterFullTextIndexStatement",
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if s.Action != nil {
		node["Action"] = alterFullTextIndexActionToJSON(s.Action)
	}
	return addSpan(node, frag(s))
}

func alterFullTextIndexActionToJSON(a ast.AlterFullTextIndexActionOption) jsonNode {
	switch action := a.(type) {
	case *ast.SimpleAlterFullTextIndexAction:
		return addSpan(jsonNode{
			"$type":      "SimpleAlterFullTextIndexAction",
			"ActionKind": action.ActionKind,
		}, frag(a))
	case *ast.AddAlterFullTextIndexAction:
		node := jsonNode{
			"$type":            "AddAlterFullTextIndexAction",
			"WithNoPopulation": action.WithNoPopulation,
		}
		if len(action.Columns) > 0 {
			cols := make([]jsonNode, len(action.Columns))
			for i, col := range action.Columns {
				cols[i] = fullTextIndexColumnToJSON(col)
			}
			node["Columns"] = cols
		}
		return addSpan(node, frag(a))
	case *ast.DropAlterFullTextIndexAction:
		node := jsonNode{
			"$type":            "DropAlterFullTextIndexAction",
			"WithNoPopulation": action.WithNoPopulation,
		}
		if len(action.Columns) > 0 {
			cols := make([]jsonNode, len(action.Columns))
			for i, col := range action.Columns {
				cols[i] = identifierToJSON(col)
			}
			node["Columns"] = cols
		}
		return addSpan(node, frag(a))
	case *ast.SetStopListAlterFullTextIndexAction:
		node := jsonNode{
			"$type":            "SetStopListAlterFullTextIndexAction",
			"WithNoPopulation": action.WithNoPopulation,
		}
		if action.StopListOption != nil {
			node["StopListOption"] = stopListFullTextIndexOptionToJSON(action.StopListOption)
		}
		return addSpan(node, frag(a))
	case *ast.SetSearchPropertyListAlterFullTextIndexAction:
		node := jsonNode{
			"$type":            "SetSearchPropertyListAlterFullTextIndexAction",
			"WithNoPopulation": action.WithNoPopulation,
		}
		if action.SearchPropertyListOption != nil {
			node["SearchPropertyListOption"] = searchPropertyListFullTextIndexOptionToJSON(action.SearchPropertyListOption)
		}
		return addSpan(node, frag(a))
	case *ast.AlterColumnAlterFullTextIndexAction:
		node := jsonNode{
			"$type":            "AlterColumnAlterFullTextIndexAction",
			"WithNoPopulation": action.WithNoPopulation,
		}
		if action.Column != nil {
			node["Column"] = fullTextIndexColumnToJSON(action.Column)
		}
		return addSpan(node, frag(a))
	}
	return nil
}

func stopListFullTextIndexOptionToJSON(opt *ast.StopListFullTextIndexOption) jsonNode {
	node := jsonNode{
		"$type":      "StopListFullTextIndexOption",
		"IsOff":      opt.IsOff,
		"OptionKind": opt.OptionKind,
	}
	if opt.StopListName != nil {
		node["StopListName"] = identifierToJSON(opt.StopListName)
	}
	return addSpan(node, frag(opt))
}

func searchPropertyListFullTextIndexOptionToJSON(opt *ast.SearchPropertyListFullTextIndexOption) jsonNode {
	node := jsonNode{
		"$type":      "SearchPropertyListFullTextIndexOption",
		"IsOff":      opt.IsOff,
		"OptionKind": opt.OptionKind,
	}
	if opt.PropertyListName != nil {
		node["PropertyListName"] = identifierToJSON(opt.PropertyListName)
	}
	return addSpan(node, frag(opt))
}

func fullTextIndexColumnToJSON(col *ast.FullTextIndexColumn) jsonNode {
	node := jsonNode{
		"$type":                "FullTextIndexColumn",
		"StatisticalSemantics": col.StatisticalSemantics,
	}
	if col.Name != nil {
		node["Name"] = identifierToJSON(col.Name)
	}
	if col.TypeColumn != nil {
		node["TypeColumn"] = identifierToJSON(col.TypeColumn)
	}
	if col.LanguageTerm != nil {
		node["LanguageTerm"] = identifierOrValueExpressionToJSON(col.LanguageTerm)
	}
	return addSpan(node, frag(col))
}

func alterSymmetricKeyStatementToJSON(s *ast.AlterSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterSymmetricKeyStatement",
	}
	// Only include IsAdd when there are encrypting mechanisms (meaning an action was specified)
	if len(s.EncryptingMechanisms) > 0 {
		node["IsAdd"] = s.IsAdd
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.EncryptingMechanisms) > 0 {
		mechs := make([]jsonNode, len(s.EncryptingMechanisms))
		for i, m := range s.EncryptingMechanisms {
			mechs[i] = cryptoMechanismToJSON(m)
		}
		node["EncryptingMechanisms"] = mechs
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func createDatabaseStatementToJSON(s *ast.CreateDatabaseStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateDatabaseStatement",
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if s.Containment != nil {
		node["Containment"] = containmentDatabaseOptionToJSON(s.Containment)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = createDatabaseOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	if len(s.FileGroups) > 0 {
		fgs := make([]jsonNode, len(s.FileGroups))
		for i, fg := range s.FileGroups {
			fgs[i] = fileGroupDefinitionToJSON(fg)
		}
		node["FileGroups"] = fgs
	}
	if len(s.LogOn) > 0 {
		logs := make([]jsonNode, len(s.LogOn))
		for i, fd := range s.LogOn {
			logs[i] = fileDeclarationToJSON(fd)
		}
		node["LogOn"] = logs
	}
	// Always output AttachMode
	node["AttachMode"] = s.AttachMode
	if s.CopyOf != nil {
		node["CopyOf"] = multiPartIdentifierToJSON(s.CopyOf)
	}
	if s.Collation != nil {
		node["Collation"] = identifierToJSON(s.Collation)
	}
	if s.DatabaseSnapshot != nil {
		node["DatabaseSnapshot"] = identifierToJSON(s.DatabaseSnapshot)
	}
	return addSpan(node, frag(s))
}

func containmentDatabaseOptionToJSON(c *ast.ContainmentDatabaseOption) jsonNode {
	return addSpan(jsonNode{
		"$type":      "ContainmentDatabaseOption",
		"Value":      c.Value,
		"OptionKind": c.OptionKind,
	}, frag(c))
}

func createDatabaseEncryptionKeyStatementToJSON(s *ast.CreateDatabaseEncryptionKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateDatabaseEncryptionKeyStatement",
	}
	if s.Encryptor != nil {
		node["Encryptor"] = cryptoMechanismToJSON(s.Encryptor)
	}
	if s.Algorithm != "" {
		node["Algorithm"] = s.Algorithm
	}
	return addSpan(node, frag(s))
}

func alterDatabaseEncryptionKeyStatementToJSON(s *ast.AlterDatabaseEncryptionKeyStatement) jsonNode {
	node := jsonNode{
		"$type":      "AlterDatabaseEncryptionKeyStatement",
		"Regenerate": s.Regenerate,
	}
	if s.Encryptor != nil {
		node["Encryptor"] = cryptoMechanismToJSON(s.Encryptor)
	}
	if s.Algorithm != "" {
		node["Algorithm"] = s.Algorithm
	}
	return addSpan(node, frag(s))
}

func dropDatabaseEncryptionKeyStatementToJSON(s *ast.DropDatabaseEncryptionKeyStatement) jsonNode {
	return addSpan(jsonNode{
		"$type": "DropDatabaseEncryptionKeyStatement",
	}, frag(s))
}

func fileGroupDefinitionToJSON(fg *ast.FileGroupDefinition) jsonNode {
	node := jsonNode{
		"$type": "FileGroupDefinition",
	}
	if fg.Name != nil {
		node["Name"] = identifierToJSON(fg.Name)
	}
	if len(fg.FileDeclarations) > 0 {
		decls := make([]jsonNode, len(fg.FileDeclarations))
		for i, fd := range fg.FileDeclarations {
			decls[i] = fileDeclarationToJSON(fd)
		}
		node["FileDeclarations"] = decls
	}
	node["IsDefault"] = fg.IsDefault
	node["ContainsFileStream"] = fg.ContainsFileStream
	node["ContainsMemoryOptimizedData"] = fg.ContainsMemoryOptimizedData
	return addSpan(node, frag(fg))
}

func fileDeclarationToJSON(fd *ast.FileDeclaration) jsonNode {
	node := jsonNode{
		"$type": "FileDeclaration",
	}
	if len(fd.Options) > 0 {
		opts := make([]jsonNode, len(fd.Options))
		for i, opt := range fd.Options {
			opts[i] = fileDeclarationOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	node["IsPrimary"] = fd.IsPrimary
	return addSpan(node, frag(fd))
}

func fileDeclarationOptionToJSON(opt ast.FileDeclarationOption) jsonNode {
	switch o := opt.(type) {
	case *ast.SimpleFileDeclarationOption:
		return addSpan(jsonNode{
			"$type":      "FileDeclarationOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.NameFileDeclarationOption:
		node := jsonNode{
			"$type":      "NameFileDeclarationOption",
			"IsNewName":  o.IsNewName,
			"OptionKind": o.OptionKind,
		}
		if o.LogicalFileName != nil {
			node["LogicalFileName"] = identifierOrValueExpressionToJSON(o.LogicalFileName)
		}
		return addSpan(node, frag(opt))
	case *ast.FileNameFileDeclarationOption:
		node := jsonNode{
			"$type":      "FileNameFileDeclarationOption",
			"OptionKind": o.OptionKind,
		}
		if o.OSFileName != nil {
			node["OSFileName"] = stringLiteralToJSON(o.OSFileName)
		}
		return addSpan(node, frag(opt))
	case *ast.SizeFileDeclarationOption:
		node := jsonNode{
			"$type":      "SizeFileDeclarationOption",
			"Units":      o.Units,
			"OptionKind": o.OptionKind,
		}
		if o.Size != nil {
			node["Size"] = scalarExpressionToJSON(o.Size)
		}
		return addSpan(node, frag(opt))
	case *ast.MaxSizeFileDeclarationOption:
		node := jsonNode{
			"$type":      "MaxSizeFileDeclarationOption",
			"Units":      o.Units,
			"Unlimited":  o.Unlimited,
			"OptionKind": o.OptionKind,
		}
		if o.MaxSize != nil {
			node["MaxSize"] = scalarExpressionToJSON(o.MaxSize)
		}
		return addSpan(node, frag(opt))
	case *ast.FileGrowthFileDeclarationOption:
		node := jsonNode{
			"$type":      "FileGrowthFileDeclarationOption",
			"Units":      o.Units,
			"OptionKind": o.OptionKind,
		}
		if o.GrowthIncrement != nil {
			node["GrowthIncrement"] = scalarExpressionToJSON(o.GrowthIncrement)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "FileDeclarationOption"}, frag(opt))
	}
}

func createDatabaseOptionToJSON(opt ast.CreateDatabaseOption) jsonNode {
	switch o := opt.(type) {
	case *ast.OnOffDatabaseOption:
		return addSpan(jsonNode{
			"$type":       "OnOffDatabaseOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}, frag(opt))
	case *ast.IdentifierDatabaseOption:
		node := jsonNode{
			"$type":      "IdentifierDatabaseOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = identifierToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	case *ast.MaxSizeDatabaseOption:
		node := jsonNode{
			"$type": "MaxSizeDatabaseOption",
		}
		if o.MaxSize != nil {
			node["MaxSize"] = scalarExpressionToJSON(o.MaxSize)
		}
		if o.Units != "" {
			node["Units"] = o.Units
		}
		if o.OptionKind != "" {
			node["OptionKind"] = o.OptionKind
		}
		return addSpan(node, frag(opt))
	case *ast.LiteralDatabaseOption:
		node := jsonNode{
			"$type": "LiteralDatabaseOption",
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		if o.OptionKind != "" {
			node["OptionKind"] = o.OptionKind
		}
		return addSpan(node, frag(opt))
	case *ast.ElasticPoolSpecification:
		node := jsonNode{
			"$type": "ElasticPoolSpecification",
		}
		if o.ElasticPoolName != nil {
			node["ElasticPoolName"] = identifierToJSON(o.ElasticPoolName)
		}
		if o.OptionKind != "" {
			node["OptionKind"] = o.OptionKind
		}
		return addSpan(node, frag(opt))
	case *ast.SimpleDatabaseOption:
		return addSpan(jsonNode{
			"$type":      "DatabaseOption",
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.FileStreamDatabaseOption:
		node := jsonNode{
			"$type":      "FileStreamDatabaseOption",
			"OptionKind": o.OptionKind,
		}
		if o.NonTransactedAccess != "" {
			node["NonTransactedAccess"] = o.NonTransactedAccess
		}
		if o.DirectoryName != nil {
			node["DirectoryName"] = scalarExpressionToJSON(o.DirectoryName)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "CreateDatabaseOption"}, frag(opt))
	}
}

func createLoginStatementToJSON(s *ast.CreateLoginStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateLoginStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Source != nil {
		node["Source"] = createLoginSourceToJSON(s.Source)
	}
	return addSpan(node, frag(s))
}

func createLoginSourceToJSON(s ast.CreateLoginSource) jsonNode {
	switch src := s.(type) {
	case *ast.ExternalCreateLoginSource:
		node := jsonNode{
			"$type": "ExternalCreateLoginSource",
		}
		if len(src.Options) > 0 {
			opts := make([]jsonNode, len(src.Options))
			for i, opt := range src.Options {
				opts[i] = principalOptionToJSON(opt)
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(s))
	case *ast.PasswordCreateLoginSource:
		node := jsonNode{
			"$type":      "PasswordCreateLoginSource",
			"Hashed":     src.Hashed,
			"MustChange": src.MustChange,
		}
		if src.Password != nil {
			node["Password"] = scalarExpressionToJSON(src.Password)
		}
		if len(src.Options) > 0 {
			opts := make([]jsonNode, len(src.Options))
			for i, opt := range src.Options {
				opts[i] = principalOptionToJSON(opt)
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(s))
	case *ast.WindowsCreateLoginSource:
		node := jsonNode{
			"$type": "WindowsCreateLoginSource",
		}
		if len(src.Options) > 0 {
			opts := make([]jsonNode, len(src.Options))
			for i, opt := range src.Options {
				opts[i] = principalOptionToJSON(opt)
			}
			node["Options"] = opts
		}
		return addSpan(node, frag(s))
	case *ast.CertificateCreateLoginSource:
		node := jsonNode{
			"$type": "CertificateCreateLoginSource",
		}
		if src.Certificate != nil {
			node["Certificate"] = identifierToJSON(src.Certificate)
		}
		if src.Credential != nil {
			node["Credential"] = identifierToJSON(src.Credential)
		}
		return addSpan(node, frag(s))
	case *ast.AsymmetricKeyCreateLoginSource:
		node := jsonNode{
			"$type": "AsymmetricKeyCreateLoginSource",
		}
		if src.Key != nil {
			node["Key"] = identifierToJSON(src.Key)
		}
		if src.Credential != nil {
			node["Credential"] = identifierToJSON(src.Credential)
		}
		return addSpan(node, frag(s))
	default:
		return addSpan(jsonNode{}, frag(s))
	}
}

func principalOptionToJSON(o ast.PrincipalOption) jsonNode {
	switch opt := o.(type) {
	case *ast.LiteralPrincipalOption:
		node := jsonNode{
			"$type":      "LiteralPrincipalOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		return addSpan(node, frag(o))
	case *ast.IdentifierPrincipalOption:
		node := jsonNode{
			"$type":      "IdentifierPrincipalOption",
			"OptionKind": opt.OptionKind,
		}
		if opt.Identifier != nil {
			node["Identifier"] = identifierToJSON(opt.Identifier)
		}
		return addSpan(node, frag(o))
	case *ast.OnOffPrincipalOption:
		return addSpan(jsonNode{
			"$type":       "OnOffPrincipalOption",
			"OptionKind":  opt.OptionKind,
			"OptionState": opt.OptionState,
		}, frag(o))
	case *ast.PrincipalOptionSimple:
		return addSpan(jsonNode{
			"$type":      "PrincipalOption",
			"OptionKind": opt.OptionKind,
		}, frag(o))
	case *ast.PasswordAlterPrincipalOption:
		node := jsonNode{
			"$type":      "PasswordAlterPrincipalOption",
			"OptionKind": opt.OptionKind,
			"MustChange": opt.MustChange,
			"Unlock":     opt.Unlock,
			"Hashed":     opt.Hashed,
		}
		if opt.Password != nil {
			node["Password"] = scalarExpressionToJSON(opt.Password)
		}
		if opt.OldPassword != nil {
			node["OldPassword"] = stringLiteralToJSON(opt.OldPassword)
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{}, frag(o))
	}
}

func alterLoginEnableDisableStatementToJSON(s *ast.AlterLoginEnableDisableStatement) jsonNode {
	node := jsonNode{
		"$type":    "AlterLoginEnableDisableStatement",
		"IsEnable": s.IsEnable,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func alterLoginOptionsStatementToJSON(s *ast.AlterLoginOptionsStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterLoginOptionsStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = principalOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
}

func dropLoginStatementToJSON(s *ast.DropLoginStatement) jsonNode {
	node := jsonNode{
		"$type":      "DropLoginStatement",
		"IsIfExists": s.IsIfExists,
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func createIndexStatementToJSON(s *ast.CreateIndexStatement) jsonNode {
	node := jsonNode{
		"$type":                  "CreateIndexStatement",
		"Translated80SyntaxTo90": s.Translated80SyntaxTo90,
		"Unique":                 s.Unique,
	}
	if s.Clustered != nil {
		node["Clustered"] = *s.Clustered
	}
	if len(s.Columns) > 0 {
		cols := make([]jsonNode, len(s.Columns))
		for i, col := range s.Columns {
			cols[i] = columnWithSortOrderToJSON(col)
		}
		node["Columns"] = cols
	}
	if len(s.IncludeColumns) > 0 {
		cols := make([]jsonNode, len(s.IncludeColumns))
		for i, col := range s.IncludeColumns {
			cols[i] = columnReferenceExpressionToJSON(col)
		}
		node["IncludeColumns"] = cols
	}
	if s.OnFileGroupOrPartitionScheme != nil {
		node["OnFileGroupOrPartitionScheme"] = fileGroupOrPartitionSchemeToJSON(s.OnFileGroupOrPartitionScheme)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if len(s.IndexOptions) > 0 {
		opts := make([]jsonNode, len(s.IndexOptions))
		for i, opt := range s.IndexOptions {
			opts[i] = indexOptionToJSON(opt)
		}
		node["IndexOptions"] = opts
	}
	if s.FilterPredicate != nil {
		node["FilterPredicate"] = booleanExpressionToJSON(s.FilterPredicate)
	}
	if s.FileStreamOn != nil {
		node["FileStreamOn"] = identifierOrValueExpressionToJSON(s.FileStreamOn)
	}
	return addSpan(node, frag(s))
}

func createAsymmetricKeyStatementToJSON(s *ast.CreateAsymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateAsymmetricKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.KeySource != nil {
		node["KeySource"] = encryptionSourceToJSON(s.KeySource)
	}
	if s.EncryptionAlgorithm != "" {
		node["EncryptionAlgorithm"] = s.EncryptionAlgorithm
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Password != nil {
		node["Password"] = scalarExpressionToJSON(s.Password)
	}
	return addSpan(node, frag(s))
}

func encryptionSourceToJSON(source ast.EncryptionSource) interface{} {
	switch s := source.(type) {
	case *ast.ProviderEncryptionSource:
		return providerEncryptionSourceToJSON(s)
	case *ast.AssemblyEncryptionSource:
		node := jsonNode{
			"$type": "AssemblyEncryptionSource",
		}
		if s.Assembly != nil {
			node["Assembly"] = identifierToJSON(s.Assembly)
		}
		return addSpan(node, frag(s))
	case *ast.FileEncryptionSource:
		node := jsonNode{
			"$type":        "FileEncryptionSource",
			"IsExecutable": s.IsExecutable,
		}
		if s.File != nil {
			node["File"] = stringLiteralToJSON(s.File)
		}
		return addSpan(node, frag(s))
	default:
		return nil
	}
}

func providerEncryptionSourceToJSON(s *ast.ProviderEncryptionSource) jsonNode {
	node := jsonNode{
		"$type": "ProviderEncryptionSource",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.KeyOptions) > 0 {
		options := make([]interface{}, len(s.KeyOptions))
		for i, opt := range s.KeyOptions {
			options[i] = keyOptionToJSON(opt)
		}
		node["KeyOptions"] = options
	}
	return addSpan(node, frag(s))
}

func keyOptionToJSON(opt ast.KeyOption) interface{} {
	switch o := opt.(type) {
	case *ast.AlgorithmKeyOption:
		return addSpan(jsonNode{
			"$type":      "AlgorithmKeyOption",
			"Algorithm":  o.Algorithm,
			"OptionKind": o.OptionKind,
		}, frag(o))
	case *ast.ProviderKeyNameKeyOption:
		node := jsonNode{
			"$type":      "ProviderKeyNameKeyOption",
			"OptionKind": o.OptionKind,
		}
		if o.KeyName != nil {
			node["KeyName"] = scalarExpressionToJSON(o.KeyName)
		}
		return addSpan(node, frag(o))
	case *ast.CreationDispositionKeyOption:
		return addSpan(jsonNode{
			"$type":       "CreationDispositionKeyOption",
			"IsCreateNew": o.IsCreateNew,
			"OptionKind":  o.OptionKind,
		}, frag(o))
	case *ast.KeySourceKeyOption:
		node := jsonNode{
			"$type":      "KeySourceKeyOption",
			"OptionKind": o.OptionKind,
		}
		if o.PassPhrase != nil {
			node["PassPhrase"] = scalarExpressionToJSON(o.PassPhrase)
		}
		return addSpan(node, frag(o))
	case *ast.IdentityValueKeyOption:
		node := jsonNode{
			"$type":      "IdentityValueKeyOption",
			"OptionKind": o.OptionKind,
		}
		if o.IdentityPhrase != nil {
			node["IdentityPhrase"] = scalarExpressionToJSON(o.IdentityPhrase)
		}
		return addSpan(node, frag(o))
	default:
		return nil
	}
}

func createSymmetricKeyStatementToJSON(s *ast.CreateSymmetricKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSymmetricKeyStatement",
	}
	if len(s.KeyOptions) > 0 {
		opts := make([]interface{}, len(s.KeyOptions))
		for i, opt := range s.KeyOptions {
			opts[i] = keyOptionToJSON(opt)
		}
		node["KeyOptions"] = opts
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Provider != nil {
		node["Provider"] = identifierToJSON(s.Provider)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.EncryptingMechanisms) > 0 {
		mechs := make([]jsonNode, len(s.EncryptingMechanisms))
		for i, mech := range s.EncryptingMechanisms {
			mechs[i] = cryptoMechanismToJSON(mech)
		}
		node["EncryptingMechanisms"] = mechs
	}
	return addSpan(node, frag(s))
}

func cryptoMechanismToJSON(mech *ast.CryptoMechanism) jsonNode {
	node := jsonNode{
		"$type":               "CryptoMechanism",
		"CryptoMechanismType": mech.CryptoMechanismType,
	}
	if mech.Identifier != nil {
		node["Identifier"] = identifierToJSON(mech.Identifier)
	}
	if mech.PasswordOrSignature != nil {
		node["PasswordOrSignature"] = scalarExpressionToJSON(mech.PasswordOrSignature)
	}
	return addSpan(node, frag(mech))
}

func createCertificateStatementToJSON(s *ast.CreateCertificateStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateCertificateStatement",
	}
	if s.CertificateSource != nil {
		node["CertificateSource"] = encryptionSourceToJSON(s.CertificateSource)
	}
	if len(s.CertificateOptions) > 0 {
		options := make([]jsonNode, len(s.CertificateOptions))
		for i, opt := range s.CertificateOptions {
			options[i] = addSpan(jsonNode{
				"$type": "CertificateOption",
				"Kind":  opt.Kind,
				"Value": stringLiteralToJSON(opt.Value),
			}, frag(opt))
		}
		node["CertificateOptions"] = options
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.ActiveForBeginDialog != "" {
		node["ActiveForBeginDialog"] = s.ActiveForBeginDialog
	}
	if s.PrivateKeyPath != nil {
		node["PrivateKeyPath"] = stringLiteralToJSON(s.PrivateKeyPath)
	}
	if s.EncryptionPassword != nil {
		node["EncryptionPassword"] = stringLiteralToJSON(s.EncryptionPassword)
	}
	if s.DecryptionPassword != nil {
		node["DecryptionPassword"] = stringLiteralToJSON(s.DecryptionPassword)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func createServiceStatementToJSON(s *ast.CreateServiceStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateServiceStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.QueueName != nil {
		node["QueueName"] = schemaObjectNameToJSON(s.QueueName)
	}
	if len(s.ServiceContracts) > 0 {
		contracts := make([]jsonNode, len(s.ServiceContracts))
		for i, c := range s.ServiceContracts {
			contracts[i] = serviceContractToJSON(c)
		}
		node["ServiceContracts"] = contracts
	}
	return addSpan(node, frag(s))
}

func serviceContractToJSON(c *ast.ServiceContract) jsonNode {
	node := jsonNode{
		"$type": "ServiceContract",
	}
	if c.Name != nil {
		node["Name"] = identifierToJSON(c.Name)
	}
	node["Action"] = c.Action
	return addSpan(node, frag(c))
}

func createQueueStatementToJSON(s *ast.CreateQueueStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateQueueStatement",
	}
	if s.OnFileGroup != nil {
		node["OnFileGroup"] = identifierOrValueExpressionToJSON(s.OnFileGroup)
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
	return addSpan(node, frag(s))
}

func queueOptionToJSON(opt ast.QueueOption) jsonNode {
	switch o := opt.(type) {
	case *ast.QueueStateOption:
		node := jsonNode{
			"$type":       "QueueStateOption",
			"OptionState": o.OptionState,
			"OptionKind":  o.OptionKind,
		}
		return addSpan(node, frag(opt))
	case *ast.QueueOptionSimple:
		node := jsonNode{
			"$type":      "QueueOption",
			"OptionKind": o.OptionKind,
		}
		return addSpan(node, frag(opt))
	case *ast.QueueProcedureOption:
		node := jsonNode{
			"$type":      "QueueProcedureOption",
			"OptionKind": o.OptionKind,
		}
		if o.OptionValue != nil {
			node["OptionValue"] = schemaObjectNameToJSON(o.OptionValue)
		}
		return addSpan(node, frag(opt))
	case *ast.QueueValueOption:
		node := jsonNode{
			"$type":      "QueueValueOption",
			"OptionKind": o.OptionKind,
		}
		if o.OptionValue != nil {
			node["OptionValue"] = scalarExpressionToJSON(o.OptionValue)
		}
		return addSpan(node, frag(opt))
	case *ast.QueueExecuteAsOption:
		node := jsonNode{
			"$type":      "QueueExecuteAsOption",
			"OptionKind": o.OptionKind,
		}
		if o.OptionValue != nil {
			node["OptionValue"] = executeAsClauseToJSON(o.OptionValue)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "QueueOption"}, frag(opt))
	}
}

func createRouteStatementToJSON(s *ast.CreateRouteStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateRouteStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.RouteOptions) > 0 {
		opts := make([]jsonNode, len(s.RouteOptions))
		for i, opt := range s.RouteOptions {
			opts[i] = routeOptionToJSON(opt)
		}
		node["RouteOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func routeOptionToJSON(opt *ast.RouteOption) jsonNode {
	node := jsonNode{
		"$type":      "RouteOption",
		"OptionKind": opt.OptionKind,
	}
	if opt.Literal != nil {
		node["Literal"] = scalarExpressionToJSON(opt.Literal)
	}
	return addSpan(node, frag(opt))
}

func createEndpointStatementToJSON(s *ast.CreateEndpointStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateEndpointStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.State != "" {
		node["State"] = s.State
	}
	if s.Affinity != nil {
		node["Affinity"] = endpointAffinityToJSON(s.Affinity)
	}
	if s.Protocol != "" {
		node["Protocol"] = s.Protocol
	}
	if len(s.ProtocolOptions) > 0 {
		opts := make([]jsonNode, len(s.ProtocolOptions))
		for i, opt := range s.ProtocolOptions {
			opts[i] = endpointProtocolOptionToJSON(opt)
		}
		node["ProtocolOptions"] = opts
	}
	if s.EndpointType != "" {
		node["EndpointType"] = s.EndpointType
	}
	if len(s.PayloadOptions) > 0 {
		opts := make([]jsonNode, len(s.PayloadOptions))
		for i, opt := range s.PayloadOptions {
			opts[i] = payloadOptionToJSON(opt)
		}
		node["PayloadOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func createAssemblyStatementToJSON(s *ast.CreateAssemblyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateAssemblyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, param := range s.Parameters {
			params[i] = scalarExpressionToJSON(param)
		}
		node["Parameters"] = params
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = assemblyOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func applicationRoleOptionToJSON(opt *ast.ApplicationRoleOption) jsonNode {
	node := jsonNode{
		"$type":      "ApplicationRoleOption",
		"OptionKind": opt.OptionKind,
	}
	if opt.Value != nil {
		node["Value"] = identifierOrValueExpressionToJSON(opt.Value)
	}
	return addSpan(node, frag(opt))
}

func createFulltextCatalogStatementToJSON(s *ast.CreateFulltextCatalogStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateFulltextCatalogStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func createFulltextIndexStatementToJSON(s *ast.CreateFulltextIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateFullTextIndexStatement",
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if len(s.FullTextIndexColumns) > 0 {
		cols := make([]jsonNode, len(s.FullTextIndexColumns))
		for i, col := range s.FullTextIndexColumns {
			cols[i] = fullTextIndexColumnToJSON(col)
		}
		node["FullTextIndexColumns"] = cols
	}
	if s.KeyIndexName != nil {
		node["KeyIndexName"] = identifierToJSON(s.KeyIndexName)
	}
	if s.CatalogAndFileGroup != nil {
		node["CatalogAndFileGroup"] = fullTextCatalogAndFileGroupToJSON(s.CatalogAndFileGroup)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = fullTextIndexOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	return addSpan(node, frag(s))
}

func fullTextCatalogAndFileGroupToJSON(cfg *ast.FullTextCatalogAndFileGroup) jsonNode {
	node := jsonNode{
		"$type":            "FullTextCatalogAndFileGroup",
		"FileGroupIsFirst": cfg.FileGroupIsFirst,
	}
	if cfg.CatalogName != nil {
		node["CatalogName"] = identifierToJSON(cfg.CatalogName)
	}
	if cfg.FileGroupName != nil {
		node["FileGroupName"] = identifierToJSON(cfg.FileGroupName)
	}
	return addSpan(node, frag(cfg))
}

func fullTextIndexOptionToJSON(opt ast.FullTextIndexOption) jsonNode {
	switch o := opt.(type) {
	case *ast.ChangeTrackingFullTextIndexOption:
		return addSpan(jsonNode{
			"$type":      "ChangeTrackingFullTextIndexOption",
			"Value":      o.Value,
			"OptionKind": o.OptionKind,
		}, frag(opt))
	case *ast.StopListFullTextIndexOption:
		return addSpan(stopListFullTextIndexOptionToJSON(o), frag(opt))
	case *ast.SearchPropertyListFullTextIndexOption:
		return addSpan(searchPropertyListFullTextIndexOptionToJSON(o), frag(opt))
	}
	return nil
}

func createRemoteServiceBindingStatementToJSON(s *ast.CreateRemoteServiceBindingStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateRemoteServiceBindingStatement",
	}
	if s.Service != nil {
		node["Service"] = scalarExpressionToJSON(s.Service)
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func createTypeStatementToJSON(s *ast.CreateTypeStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateTypeStatement",
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func createTypeTableStatementToJSON(s *ast.CreateTypeTableStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateTypeTableStatement",
	}
	if s.Definition != nil {
		node["Definition"] = tableDefinitionToJSON(s.Definition)
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			opts[i] = tableOptionToJSON(o)
		}
		node["Options"] = opts
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	return addSpan(node, frag(s))
}

func createXmlIndexStatementToJSON(s *ast.CreateXmlIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateXmlIndexStatement",
	}
	node["Primary"] = s.Primary
	if s.XmlColumn != nil {
		node["XmlColumn"] = identifierToJSON(s.XmlColumn)
	}
	if s.SecondaryXmlIndexName != nil {
		node["SecondaryXmlIndexName"] = identifierToJSON(s.SecondaryXmlIndexName)
	}
	if s.SecondaryXmlIndexType != "" {
		node["SecondaryXmlIndexType"] = s.SecondaryXmlIndexType
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if len(s.IndexOptions) > 0 {
		opts := make([]jsonNode, len(s.IndexOptions))
		for i, opt := range s.IndexOptions {
			opts[i] = indexOptionToJSON(opt)
		}
		node["IndexOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func createSelectiveXmlIndexStatementToJSON(s *ast.CreateSelectiveXmlIndexStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateSelectiveXmlIndexStatement",
	}
	node["IsSecondary"] = s.IsSecondary
	if s.XmlColumn != nil {
		node["XmlColumn"] = identifierToJSON(s.XmlColumn)
	}
	if s.UsingXmlIndexName != nil {
		node["UsingXmlIndexName"] = identifierToJSON(s.UsingXmlIndexName)
	}
	if s.PathName != nil {
		node["PathName"] = identifierToJSON(s.PathName)
	}
	if len(s.PromotedPaths) > 0 {
		paths := make([]jsonNode, len(s.PromotedPaths))
		for i, path := range s.PromotedPaths {
			paths[i] = selectiveXmlIndexPromotedPathToJSON(path)
		}
		node["PromotedPaths"] = paths
	}
	if s.XmlNamespaces != nil {
		node["XmlNamespaces"] = xmlNamespacesToJSON(s.XmlNamespaces)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.OnName != nil {
		node["OnName"] = schemaObjectNameToJSON(s.OnName)
	}
	if len(s.IndexOptions) > 0 {
		opts := make([]jsonNode, len(s.IndexOptions))
		for i, opt := range s.IndexOptions {
			opts[i] = indexOptionToJSON(opt)
		}
		node["IndexOptions"] = opts
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(p))
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
	return addSpan(node, frag(s))
}

func eventNotificationObjectScopeToJSON(s *ast.EventNotificationObjectScope) jsonNode {
	node := jsonNode{
		"$type":  "EventNotificationObjectScope",
		"Target": s.Target,
	}
	if s.QueueName != nil {
		node["QueueName"] = schemaObjectNameToJSON(s.QueueName)
	}
	return addSpan(node, frag(s))
}

func eventTypeGroupContainerToJSON(c ast.EventTypeGroupContainer) jsonNode {
	switch v := c.(type) {
	case *ast.EventTypeContainer:
		return addSpan(jsonNode{
			"$type":     "EventTypeContainer",
			"EventType": v.EventType,
		}, frag(c))
	case *ast.EventGroupContainer:
		return addSpan(jsonNode{
			"$type":      "EventGroupContainer",
			"EventGroup": v.EventGroup,
		}, frag(c))
	default:
		return addSpan(jsonNode{"$type": "Unknown"}, frag(c))
	}
}

func alterDatabaseAddFileStatementToJSON(s *ast.AlterDatabaseAddFileStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseAddFileStatement",
	}
	// Check if we have any complete file declarations (with options)
	hasCompleteDeclarations := false
	for _, fd := range s.FileDeclarations {
		if len(fd.Options) > 0 {
			hasCompleteDeclarations = true
			break
		}
	}
	if hasCompleteDeclarations {
		decls := make([]jsonNode, len(s.FileDeclarations))
		for i, fd := range s.FileDeclarations {
			decls[i] = fileDeclarationToJSON(fd)
		}
		node["FileDeclarations"] = decls
	}
	if s.FileGroup != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroup)
	}
	// Only include IsLog/UseCurrent if we have complete declarations
	if hasCompleteDeclarations {
		node["IsLog"] = s.IsLog
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	if hasCompleteDeclarations {
		node["UseCurrent"] = s.UseCurrent
	}
	return addSpan(node, frag(s))
}

func alterDatabaseAddFileGroupStatementToJSON(s *ast.AlterDatabaseAddFileGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseAddFileGroupStatement",
	}
	if s.FileGroupName != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroupName)
	}
	node["ContainsFileStream"] = s.ContainsFileStream
	node["ContainsMemoryOptimizedData"] = s.ContainsMemoryOptimizedData
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = s.UseCurrent
	return addSpan(node, frag(s))
}

func alterDatabaseModifyFileStatementToJSON(s *ast.AlterDatabaseModifyFileStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseModifyFileStatement",
	}
	if s.FileDeclaration != nil {
		node["FileDeclaration"] = fileDeclarationToJSON(s.FileDeclaration)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = s.UseCurrent
	return addSpan(node, frag(s))
}

func alterDatabaseModifyFileGroupStatementToJSON(s *ast.AlterDatabaseModifyFileGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseModifyFileGroupStatement",
	}
	if s.FileGroupName != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroupName)
	}
	if s.NewFileGroupName != nil {
		node["NewFileGroupName"] = identifierToJSON(s.NewFileGroupName)
	}
	node["MakeDefault"] = s.MakeDefault
	if s.UpdatabilityOption != "" {
		node["UpdatabilityOption"] = s.UpdatabilityOption
	} else {
		node["UpdatabilityOption"] = "None"
	}
	if s.Termination != nil {
		node["Termination"] = alterDatabaseTerminationToJSON(s.Termination)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = s.UseCurrent
	return addSpan(node, frag(s))
}

func alterDatabaseTerminationToJSON(t *ast.AlterDatabaseTermination) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseTermination",
	}
	node["ImmediateRollback"] = t.ImmediateRollback
	if t.RollbackAfter != nil {
		node["RollbackAfter"] = scalarExpressionToJSON(t.RollbackAfter)
	}
	node["NoWait"] = t.NoWait
	return addSpan(node, frag(t))
}

func alterDatabaseModifyNameStatementToJSON(s *ast.AlterDatabaseModifyNameStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseModifyNameStatement",
	}
	if s.NewName != nil {
		node["NewDatabaseName"] = identifierToJSON(s.NewName)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = false
	return addSpan(node, frag(s))
}

func alterDatabaseRemoveFileStatementToJSON(s *ast.AlterDatabaseRemoveFileStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseRemoveFileStatement",
	}
	if s.FileName != nil {
		node["File"] = identifierToJSON(s.FileName)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = false
	return addSpan(node, frag(s))
}

func alterDatabaseRemoveFileGroupStatementToJSON(s *ast.AlterDatabaseRemoveFileGroupStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseRemoveFileGroupStatement",
	}
	if s.FileGroupName != nil {
		node["FileGroup"] = identifierToJSON(s.FileGroupName)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = s.UseCurrent
	return addSpan(node, frag(s))
}

func alterDatabaseCollateStatementToJSON(s *ast.AlterDatabaseCollateStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseCollateStatement",
	}
	if s.Collation != nil {
		node["Collation"] = identifierToJSON(s.Collation)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = false
	return addSpan(node, frag(s))
}

func alterDatabaseRebuildLogStatementToJSON(s *ast.AlterDatabaseRebuildLogStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseRebuildLogStatement",
	}
	if s.FileDeclaration != nil {
		node["FileDeclaration"] = fileDeclarationToJSON(s.FileDeclaration)
	}
	if s.DatabaseName != nil {
		node["DatabaseName"] = identifierToJSON(s.DatabaseName)
	}
	node["UseCurrent"] = s.UseCurrent
	return addSpan(node, frag(s))
}

func alterDatabaseScopedConfigurationClearStatementToJSON(s *ast.AlterDatabaseScopedConfigurationClearStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseScopedConfigurationClearStatement",
	}
	if s.Option != nil {
		node["Option"] = databaseConfigurationClearOptionToJSON(s.Option)
	}
	node["Secondary"] = s.Secondary
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(o))
}

func alterDatabaseScopedConfigurationSetStatementToJSON(s *ast.AlterDatabaseScopedConfigurationSetStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterDatabaseScopedConfigurationSetStatement",
	}
	if s.Option != nil {
		node["Option"] = databaseConfigurationSetOptionToJSON(s.Option)
	}
	node["Secondary"] = s.Secondary
	return addSpan(node, frag(s))
}

func databaseConfigurationSetOptionToJSON(o ast.DatabaseConfigurationSetOption) jsonNode {
	switch opt := o.(type) {
	case *ast.MaxDopConfigurationOption:
		node := jsonNode{
			"$type":   "MaxDopConfigurationOption",
			"Primary": opt.Primary,
		}
		if opt.Value != nil {
			node["Value"] = scalarExpressionToJSON(opt.Value)
		}
		node["OptionKind"] = opt.OptionKind
		return addSpan(node, frag(o))
	case *ast.OnOffPrimaryConfigurationOption:
		return addSpan(jsonNode{
			"$type":       "OnOffPrimaryConfigurationOption",
			"OptionState": opt.OptionState,
			"OptionKind":  opt.OptionKind,
		}, frag(o))
	case *ast.GenericConfigurationOption:
		node := jsonNode{
			"$type": "GenericConfigurationOption",
		}
		if opt.GenericOptionState != nil {
			node["GenericOptionState"] = identifierOrScalarExpressionToJSON(opt.GenericOptionState)
		}
		node["OptionKind"] = opt.OptionKind
		if opt.GenericOptionKind != nil {
			node["GenericOptionKind"] = identifierToJSON(opt.GenericOptionKind)
		}
		return addSpan(node, frag(o))
	default:
		return addSpan(jsonNode{"$type": "UnknownDatabaseConfigurationSetOption"}, frag(o))
	}
}

func identifierOrScalarExpressionToJSON(i *ast.IdentifierOrScalarExpression) jsonNode {
	node := jsonNode{
		"$type": "IdentifierOrScalarExpression",
	}
	if i.Identifier != nil {
		node["Identifier"] = identifierToJSON(i.Identifier)
	}
	if i.ScalarExpression != nil {
		node["ScalarExpression"] = scalarExpressionToJSON(i.ScalarExpression)
	}
	return addSpan(node, frag(i))
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
	return addSpan(node, frag(s))
}

func createResourcePoolStatementToJSON(s *ast.CreateResourcePoolStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateResourcePoolStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ResourcePoolParameters) > 0 {
		params := make([]jsonNode, len(s.ResourcePoolParameters))
		for i, param := range s.ResourcePoolParameters {
			params[i] = resourcePoolParameterToJSON(param)
		}
		node["ResourcePoolParameters"] = params
	}
	return addSpan(node, frag(s))
}

func alterResourcePoolStatementToJSON(s *ast.AlterResourcePoolStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterResourcePoolStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ResourcePoolParameters) > 0 {
		params := make([]jsonNode, len(s.ResourcePoolParameters))
		for i, param := range s.ResourcePoolParameters {
			params[i] = resourcePoolParameterToJSON(param)
		}
		node["ResourcePoolParameters"] = params
	}
	return addSpan(node, frag(s))
}

func dropResourcePoolStatementToJSON(s *ast.DropResourcePoolStatement) jsonNode {
	node := jsonNode{
		"$type": "DropResourcePoolStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func alterExternalResourcePoolStatementToJSON(s *ast.AlterExternalResourcePoolStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterExternalResourcePoolStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ExternalResourcePoolParameters) > 0 {
		params := make([]jsonNode, len(s.ExternalResourcePoolParameters))
		for i, param := range s.ExternalResourcePoolParameters {
			params[i] = externalResourcePoolParameterToJSON(param)
		}
		node["ExternalResourcePoolParameters"] = params
	}
	return addSpan(node, frag(s))
}

func createExternalResourcePoolStatementToJSON(s *ast.CreateExternalResourcePoolStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateExternalResourcePoolStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ExternalResourcePoolParameters) > 0 {
		params := make([]jsonNode, len(s.ExternalResourcePoolParameters))
		for i, param := range s.ExternalResourcePoolParameters {
			params[i] = externalResourcePoolParameterToJSON(param)
		}
		node["ExternalResourcePoolParameters"] = params
	}
	return addSpan(node, frag(s))
}

func externalResourcePoolParameterToJSON(p *ast.ExternalResourcePoolParameter) jsonNode {
	node := jsonNode{
		"$type": "ExternalResourcePoolParameter",
	}
	if p.ParameterType != "" {
		node["ParameterType"] = p.ParameterType
	}
	if p.ParameterValue != nil {
		node["ParameterValue"] = scalarExpressionToJSON(p.ParameterValue)
	}
	if p.AffinitySpecification != nil {
		node["AffinitySpecification"] = externalResourcePoolAffinitySpecificationToJSON(p.AffinitySpecification)
	}
	return addSpan(node, frag(p))
}

func externalResourcePoolAffinitySpecificationToJSON(s *ast.ExternalResourcePoolAffinitySpecification) jsonNode {
	node := jsonNode{
		"$type": "ExternalResourcePoolAffinitySpecification",
	}
	if s.AffinityType != "" {
		node["AffinityType"] = s.AffinityType
	}
	node["IsAuto"] = s.IsAuto
	if len(s.PoolAffinityRanges) > 0 {
		ranges := make([]jsonNode, len(s.PoolAffinityRanges))
		for i, r := range s.PoolAffinityRanges {
			ranges[i] = literalRangeToJSON(r)
		}
		node["PoolAffinityRanges"] = ranges
	}
	return addSpan(node, frag(s))
}

func resourcePoolParameterToJSON(p *ast.ResourcePoolParameter) jsonNode {
	node := jsonNode{
		"$type": "ResourcePoolParameter",
	}
	if p.ParameterType != "" {
		node["ParameterType"] = p.ParameterType
	}
	if p.ParameterValue != nil {
		node["ParameterValue"] = scalarExpressionToJSON(p.ParameterValue)
	}
	if p.AffinitySpecification != nil {
		node["AffinitySpecification"] = resourcePoolAffinitySpecificationToJSON(p.AffinitySpecification)
	}
	return addSpan(node, frag(p))
}

func resourcePoolAffinitySpecificationToJSON(s *ast.ResourcePoolAffinitySpecification) jsonNode {
	node := jsonNode{
		"$type": "ResourcePoolAffinitySpecification",
	}
	if s.AffinityType != "" {
		node["AffinityType"] = s.AffinityType
	}
	node["IsAuto"] = s.IsAuto
	if len(s.PoolAffinityRanges) > 0 {
		ranges := make([]jsonNode, len(s.PoolAffinityRanges))
		for i, r := range s.PoolAffinityRanges {
			ranges[i] = literalRangeToJSON(r)
		}
		node["PoolAffinityRanges"] = ranges
	}
	return addSpan(node, frag(s))
}

func literalRangeToJSON(r *ast.LiteralRange) jsonNode {
	node := jsonNode{
		"$type": "LiteralRange",
	}
	if r.From != nil {
		node["From"] = scalarExpressionToJSON(r.From)
	}
	if r.To != nil {
		node["To"] = scalarExpressionToJSON(r.To)
	}
	return addSpan(node, frag(r))
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
	return addSpan(node, frag(s))
}

func createColumnMasterKeyStatementToJSON(s *ast.CreateColumnMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateColumnMasterKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.Parameters) > 0 {
		params := make([]jsonNode, len(s.Parameters))
		for i, p := range s.Parameters {
			params[i] = columnMasterKeyParameterToJSON(p)
		}
		node["Parameters"] = params
	}
	return addSpan(node, frag(s))
}

func columnMasterKeyParameterToJSON(p ast.ColumnMasterKeyParameter) jsonNode {
	switch param := p.(type) {
	case *ast.ColumnMasterKeyStoreProviderNameParameter:
		node := jsonNode{
			"$type": "ColumnMasterKeyStoreProviderNameParameter",
		}
		if param.Name != nil {
			node["Name"] = scalarExpressionToJSON(param.Name)
		}
		node["ParameterKind"] = param.ParameterKind
		return addSpan(node, frag(p))
	case *ast.ColumnMasterKeyPathParameter:
		node := jsonNode{
			"$type": "ColumnMasterKeyPathParameter",
		}
		if param.Path != nil {
			node["Path"] = scalarExpressionToJSON(param.Path)
		}
		node["ParameterKind"] = param.ParameterKind
		return addSpan(node, frag(p))
	case *ast.ColumnMasterKeyEnclaveComputationsParameter:
		node := jsonNode{
			"$type": "ColumnMasterKeyEnclaveComputationsParameter",
		}
		if param.Signature != nil {
			node["Signature"] = scalarExpressionToJSON(param.Signature)
		}
		node["ParameterKind"] = param.ParameterKind
		return addSpan(node, frag(p))
	default:
		return addSpan(jsonNode{"$type": "UnknownColumnMasterKeyParameter"}, frag(p))
	}
}

func dropColumnMasterKeyStatementToJSON(s *ast.DropColumnMasterKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "DropColumnMasterKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func createColumnEncryptionKeyStatementToJSON(s *ast.CreateColumnEncryptionKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateColumnEncryptionKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ColumnEncryptionKeyValues) > 0 {
		values := make([]jsonNode, len(s.ColumnEncryptionKeyValues))
		for i, v := range s.ColumnEncryptionKeyValues {
			values[i] = columnEncryptionKeyValueToJSON(v)
		}
		node["ColumnEncryptionKeyValues"] = values
	}
	return addSpan(node, frag(s))
}

func alterColumnEncryptionKeyStatementToJSON(s *ast.AlterColumnEncryptionKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterColumnEncryptionKeyStatement",
	}
	if s.AlterType != "" {
		node["AlterType"] = s.AlterType
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ColumnEncryptionKeyValues) > 0 {
		values := make([]jsonNode, len(s.ColumnEncryptionKeyValues))
		for i, v := range s.ColumnEncryptionKeyValues {
			values[i] = columnEncryptionKeyValueToJSON(v)
		}
		node["ColumnEncryptionKeyValues"] = values
	}
	return addSpan(node, frag(s))
}

func dropColumnEncryptionKeyStatementToJSON(s *ast.DropColumnEncryptionKeyStatement) jsonNode {
	node := jsonNode{
		"$type": "DropColumnEncryptionKeyStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func columnEncryptionKeyValueToJSON(v *ast.ColumnEncryptionKeyValue) jsonNode {
	node := jsonNode{
		"$type": "ColumnEncryptionKeyValue",
	}
	if len(v.Parameters) > 0 {
		params := make([]jsonNode, len(v.Parameters))
		for i, p := range v.Parameters {
			params[i] = columnEncryptionKeyValueParameterToJSON(p)
		}
		node["Parameters"] = params
	}
	return addSpan(node, frag(v))
}

func columnEncryptionKeyValueParameterToJSON(p ast.ColumnEncryptionKeyValueParameter) jsonNode {
	switch param := p.(type) {
	case *ast.ColumnMasterKeyNameParameter:
		node := jsonNode{
			"$type": "ColumnMasterKeyNameParameter",
		}
		if param.Name != nil {
			node["Name"] = identifierToJSON(param.Name)
		}
		node["ParameterKind"] = param.ParameterKind
		return addSpan(node, frag(p))
	case *ast.ColumnEncryptionAlgorithmNameParameter:
		node := jsonNode{
			"$type": "ColumnEncryptionAlgorithmNameParameter",
		}
		if param.Algorithm != nil {
			node["Algorithm"] = scalarExpressionToJSON(param.Algorithm)
		}
		node["ParameterKind"] = param.ParameterKind
		return addSpan(node, frag(p))
	case *ast.EncryptedValueParameter:
		node := jsonNode{
			"$type": "EncryptedValueParameter",
		}
		if param.Value != nil {
			node["Value"] = scalarExpressionToJSON(param.Value)
		}
		node["ParameterKind"] = param.ParameterKind
		return addSpan(node, frag(p))
	default:
		return addSpan(jsonNode{"$type": "UnknownColumnEncryptionKeyValueParameter"}, frag(p))
	}
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
	return addSpan(node, frag(s))
}

func dropCryptographicProviderStatementToJSON(s *ast.DropCryptographicProviderStatement) jsonNode {
	node := jsonNode{
		"$type": "DropCryptographicProviderStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func createBrokerPriorityStatementToJSON(s *ast.CreateBrokerPriorityStatement) jsonNode {
	node := jsonNode{
		"$type": "CreateBrokerPriorityStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.BrokerPriorityParameters) > 0 {
		params := make([]jsonNode, len(s.BrokerPriorityParameters))
		for i, p := range s.BrokerPriorityParameters {
			params[i] = brokerPriorityParameterToJSON(p)
		}
		node["BrokerPriorityParameters"] = params
	}
	return addSpan(node, frag(s))
}

func alterBrokerPriorityStatementToJSON(s *ast.AlterBrokerPriorityStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterBrokerPriorityStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.BrokerPriorityParameters) > 0 {
		params := make([]jsonNode, len(s.BrokerPriorityParameters))
		for i, p := range s.BrokerPriorityParameters {
			params[i] = brokerPriorityParameterToJSON(p)
		}
		node["BrokerPriorityParameters"] = params
	}
	return addSpan(node, frag(s))
}

func dropBrokerPriorityStatementToJSON(s *ast.DropBrokerPriorityStatement) jsonNode {
	node := jsonNode{
		"$type": "DropBrokerPriorityStatement",
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	node["IsIfExists"] = s.IsIfExists
	return addSpan(node, frag(s))
}

func brokerPriorityParameterToJSON(p *ast.BrokerPriorityParameter) jsonNode {
	node := jsonNode{
		"$type": "BrokerPriorityParameter",
	}
	if p.IsDefaultOrAny != "" {
		node["IsDefaultOrAny"] = p.IsDefaultOrAny
	}
	if p.ParameterType != "" {
		node["ParameterType"] = p.ParameterType
	}
	if p.ParameterValue != nil {
		node["ParameterValue"] = identifierOrValueExpressionToJSON(p.ParameterValue)
	}
	return addSpan(node, frag(p))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
		return addSpan(node, frag(ct))
	case *ast.ExpressionCallTarget:
		node := jsonNode{
			"$type": "ExpressionCallTarget",
		}
		if t.Expression != nil {
			node["Expression"] = scalarExpressionToJSON(t.Expression)
		}
		return addSpan(node, frag(ct))
	case *ast.UserDefinedTypeCallTarget:
		node := jsonNode{
			"$type": "UserDefinedTypeCallTarget",
		}
		if t.SchemaObjectName != nil {
			node["SchemaObjectName"] = schemaObjectNameToJSON(t.SchemaObjectName)
		}
		return addSpan(node, frag(ct))
	default:
		return addSpan(jsonNode{}, frag(ct))
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
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, o := range s.Options {
			opts[i] = procedureOptionToJSON(o)
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
	return addSpan(node, frag(s))
}

func alterExternalDataSourceStatementToJSON(s *ast.AlterExternalDataSourceStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterExternalDataSourceStatement",
	}
	if s.PreviousPushDownOption != "" {
		node["PreviousPushDownOption"] = s.PreviousPushDownOption
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.DataSourceType != "" {
		node["DataSourceType"] = s.DataSourceType
	}
	if s.Location != nil {
		node["Location"] = scalarExpressionToJSON(s.Location)
	}
	if len(s.ExternalDataSourceOptions) > 0 {
		opts := make([]jsonNode, len(s.ExternalDataSourceOptions))
		for i, o := range s.ExternalDataSourceOptions {
			opts[i] = externalDataSourceOptionToJSON(o)
		}
		node["ExternalDataSourceOptions"] = opts
	}
	return addSpan(node, frag(s))
}

func alterExternalLanguageStatementToJSON(s *ast.AlterExternalLanguageStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterExternalLanguageStatement",
	}
	if s.Platform != nil {
		node["Platform"] = identifierToJSON(s.Platform)
	}
	if s.Operation != nil {
		node["Operation"] = identifierToJSON(s.Operation)
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if len(s.ExternalLanguageFiles) > 0 {
		files := make([]jsonNode, len(s.ExternalLanguageFiles))
		for i, f := range s.ExternalLanguageFiles {
			files[i] = externalLanguageFileOptionToJSON(f)
		}
		node["ExternalLanguageFiles"] = files
	}
	return addSpan(node, frag(s))
}

func alterExternalLibraryStatementToJSON(s *ast.AlterExternalLibraryStatement) jsonNode {
	node := jsonNode{
		"$type": "AlterExternalLibraryStatement",
	}
	if s.Owner != nil {
		node["Owner"] = identifierToJSON(s.Owner)
	}
	if s.Name != nil {
		node["Name"] = identifierToJSON(s.Name)
	}
	if s.Language != nil {
		node["Language"] = stringLiteralToJSON(s.Language)
	}
	if len(s.ExternalLibraryFiles) > 0 {
		files := make([]jsonNode, len(s.ExternalLibraryFiles))
		for i, f := range s.ExternalLibraryFiles {
			files[i] = externalLibraryFileOptionToJSON(f)
		}
		node["ExternalLibraryFiles"] = files
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(f))
}

func openCursorStatementToJSON(s *ast.OpenCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "OpenCursorStatement",
	}
	if s.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(s.Cursor)
	}
	return addSpan(node, frag(s))
}

func closeCursorStatementToJSON(s *ast.CloseCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "CloseCursorStatement",
	}
	if s.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(s.Cursor)
	}
	return addSpan(node, frag(s))
}

func deallocateCursorStatementToJSON(s *ast.DeallocateCursorStatement) jsonNode {
	node := jsonNode{
		"$type": "DeallocateCursorStatement",
	}
	if s.Cursor != nil {
		node["Cursor"] = cursorIdToJSON(s.Cursor)
	}
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
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
	return addSpan(node, frag(s))
}

func statisticsOptionToJSON(opt ast.StatisticsOption) jsonNode {
	switch o := opt.(type) {
	case *ast.SimpleStatisticsOption:
		return addSpan(simpleStatisticsOptionToJSON(o), frag(opt))
	case *ast.LiteralStatisticsOption:
		return addSpan(literalStatisticsOptionToJSON(o), frag(opt))
	case *ast.OnOffStatisticsOption:
		return addSpan(onOffStatisticsOptionToJSON(o), frag(opt))
	case *ast.ResampleStatisticsOption:
		return addSpan(resampleStatisticsOptionToJSON(o), frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownStatisticsOption"}, frag(opt))
	}
}

func simpleStatisticsOptionToJSON(o *ast.SimpleStatisticsOption) jsonNode {
	node := jsonNode{
		"$type": "StatisticsOption",
	}
	if o.OptionKind != "" {
		node["OptionKind"] = o.OptionKind
	}
	return addSpan(node, frag(o))
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
	return addSpan(node, frag(o))
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
	return addSpan(node, frag(o))
}

func resampleStatisticsOptionToJSON(o *ast.ResampleStatisticsOption) jsonNode {
	node := jsonNode{
		"$type": "ResampleStatisticsOption",
	}
	if len(o.Partitions) > 0 {
		partitions := make([]jsonNode, len(o.Partitions))
		for i, p := range o.Partitions {
			partitions[i] = statisticsPartitionRangeToJSON(p)
		}
		node["Partitions"] = partitions
	}
	if o.OptionKind != "" {
		node["OptionKind"] = o.OptionKind
	}
	return addSpan(node, frag(o))
}

func statisticsPartitionRangeToJSON(r *ast.StatisticsPartitionRange) jsonNode {
	node := jsonNode{
		"$type": "StatisticsPartitionRange",
	}
	if r.From != nil {
		node["From"] = scalarExpressionToJSON(r.From)
	}
	if r.To != nil {
		node["To"] = scalarExpressionToJSON(r.To)
	}
	return addSpan(node, frag(r))
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
	return addSpan(node, frag(s))
}

func declareCursorDefinitionToJSON(d *ast.CursorDefinition) jsonNode {
	node := jsonNode{
		"$type": "CursorDefinition",
	}
	if len(d.Options) > 0 {
		opts := make([]jsonNode, len(d.Options))
		for i, o := range d.Options {
			opts[i] = addSpan(jsonNode{
				"$type":      "CursorOption",
				"OptionKind": o.OptionKind,
			}, frag(o))
		}
		node["Options"] = opts
	}
	if d.Select != nil {
		node["Select"] = selectStatementToJSON(d.Select)
	}
	return addSpan(node, frag(d))
}

func addSignatureStatementToJSON(s *ast.AddSignatureStatement) jsonNode {
	node := jsonNode{
		"$type":     "AddSignatureStatement",
		"IsCounter": s.IsCounter,
	}
	node["ElementKind"] = s.ElementKind
	if s.Element != nil {
		node["Element"] = schemaObjectNameToJSON(s.Element)
	}
	if len(s.Cryptos) > 0 {
		cryptos := make([]jsonNode, len(s.Cryptos))
		for i, c := range s.Cryptos {
			cryptos[i] = cryptoMechanismToJSON(c)
		}
		node["Cryptos"] = cryptos
	}
	return addSpan(node, frag(s))
}

func dropSignatureStatementToJSON(s *ast.DropSignatureStatement) jsonNode {
	node := jsonNode{
		"$type":     "DropSignatureStatement",
		"IsCounter": s.IsCounter,
	}
	node["ElementKind"] = s.ElementKind
	if s.Element != nil {
		node["Element"] = schemaObjectNameToJSON(s.Element)
	}
	if len(s.Cryptos) > 0 {
		cryptos := make([]jsonNode, len(s.Cryptos))
		for i, c := range s.Cryptos {
			cryptos[i] = cryptoMechanismToJSON(c)
		}
		node["Cryptos"] = cryptos
	}
	return addSpan(node, frag(s))
}

func addSensitivityClassificationStatementToJSON(s *ast.AddSensitivityClassificationStatement) jsonNode {
	node := jsonNode{
		"$type": "AddSensitivityClassificationStatement",
	}
	if len(s.Options) > 0 {
		opts := make([]jsonNode, len(s.Options))
		for i, opt := range s.Options {
			opts[i] = sensitivityClassificationOptionToJSON(opt)
		}
		node["Options"] = opts
	}
	if len(s.Columns) > 0 {
		cols := make([]jsonNode, len(s.Columns))
		for i, col := range s.Columns {
			cols[i] = columnReferenceExpressionToJSON(col)
		}
		node["Columns"] = cols
	}
	return addSpan(node, frag(s))
}

func dropSensitivityClassificationStatementToJSON(s *ast.DropSensitivityClassificationStatement) jsonNode {
	node := jsonNode{
		"$type": "DropSensitivityClassificationStatement",
	}
	if len(s.Columns) > 0 {
		cols := make([]jsonNode, len(s.Columns))
		for i, col := range s.Columns {
			cols[i] = columnReferenceExpressionToJSON(col)
		}
		node["Columns"] = cols
	}
	return addSpan(node, frag(s))
}

func sensitivityClassificationOptionToJSON(opt *ast.SensitivityClassificationOption) jsonNode {
	node := jsonNode{
		"$type": "SensitivityClassificationOption",
		"Type":  opt.Type,
	}
	if opt.Value != nil {
		node["Value"] = scalarExpressionToJSON(opt.Value)
	}
	return addSpan(node, frag(opt))
}

func openRowsetCosmosOptionToJSON(opt ast.OpenRowsetCosmosOption) jsonNode {
	switch o := opt.(type) {
	case *ast.LiteralOpenRowsetCosmosOption:
		node := jsonNode{
			"$type":      "LiteralOpenRowsetCosmosOption",
			"OptionKind": o.OptionKind,
		}
		if o.Value != nil {
			node["Value"] = scalarExpressionToJSON(o.Value)
		}
		return addSpan(node, frag(opt))
	default:
		return addSpan(jsonNode{"$type": "UnknownOpenRowsetCosmosOption"}, frag(opt))
	}
}

func openRowsetColumnDefinitionToJSON(col *ast.OpenRowsetColumnDefinition) jsonNode {
	node := jsonNode{
		"$type": "OpenRowsetColumnDefinition",
	}
	if col.JsonPath != nil {
		node["JsonPath"] = scalarExpressionToJSON(col.JsonPath)
	}
	if col.ColumnOrdinal != nil {
		node["ColumnOrdinal"] = scalarExpressionToJSON(col.ColumnOrdinal)
	}
	if col.ColumnIdentifier != nil {
		node["ColumnIdentifier"] = identifierToJSON(col.ColumnIdentifier)
	}
	if col.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(col.DataType)
	}
	if col.Collation != nil {
		node["Collation"] = identifierToJSON(col.Collation)
	}
	return addSpan(node, frag(col))
}

func createSecurityPolicyStatementToJSON(s *ast.CreateSecurityPolicyStatement) jsonNode {
	node := jsonNode{
		"$type":             "CreateSecurityPolicyStatement",
		"NotForReplication": s.NotForReplication,
		"ActionType":        s.ActionType,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if len(s.SecurityPolicyOptions) > 0 {
		opts := make([]jsonNode, len(s.SecurityPolicyOptions))
		for i, opt := range s.SecurityPolicyOptions {
			opts[i] = securityPolicyOptionToJSON(opt)
		}
		node["SecurityPolicyOptions"] = opts
	}
	if len(s.SecurityPredicateActions) > 0 {
		actions := make([]jsonNode, len(s.SecurityPredicateActions))
		for i, action := range s.SecurityPredicateActions {
			actions[i] = securityPredicateActionToJSON(action)
		}
		node["SecurityPredicateActions"] = actions
	}
	return addSpan(node, frag(s))
}

func alterSecurityPolicyStatementToJSON(s *ast.AlterSecurityPolicyStatement) jsonNode {
	// Determine ActionType based on statement contents
	actionType := "Alter"
	if len(s.SecurityPredicateActions) > 0 {
		actionType = "AlterPredicates"
	} else if len(s.SecurityPolicyOptions) > 0 {
		actionType = "AlterState"
	} else if s.NotForReplicationModified {
		actionType = "AlterReplication"
	}

	node := jsonNode{
		"$type":             "AlterSecurityPolicyStatement",
		"NotForReplication": s.NotForReplication,
		"ActionType":        actionType,
	}
	if s.Name != nil {
		node["Name"] = schemaObjectNameToJSON(s.Name)
	}
	if len(s.SecurityPolicyOptions) > 0 {
		opts := make([]jsonNode, len(s.SecurityPolicyOptions))
		for i, opt := range s.SecurityPolicyOptions {
			opts[i] = securityPolicyOptionToJSON(opt)
		}
		node["SecurityPolicyOptions"] = opts
	}
	if len(s.SecurityPredicateActions) > 0 {
		actions := make([]jsonNode, len(s.SecurityPredicateActions))
		for i, action := range s.SecurityPredicateActions {
			actions[i] = securityPredicateActionToJSON(action)
		}
		node["SecurityPredicateActions"] = actions
	}
	return addSpan(node, frag(s))
}

func securityPolicyOptionToJSON(opt *ast.SecurityPolicyOption) jsonNode {
	return addSpan(jsonNode{
		"$type":       "SecurityPolicyOption",
		"OptionKind":  opt.OptionKind,
		"OptionState": opt.OptionState,
	}, frag(opt))
}

func securityPredicateActionToJSON(action *ast.SecurityPredicateAction) jsonNode {
	node := jsonNode{
		"$type":                      "SecurityPredicateAction",
		"ActionType":                 action.ActionType,
		"SecurityPredicateType":      action.SecurityPredicateType,
		"SecurityPredicateOperation": action.SecurityPredicateOperation,
	}
	if action.FunctionCall != nil {
		node["FunctionCall"] = scalarExpressionToJSON(action.FunctionCall)
	}
	if action.TargetObjectName != nil {
		node["TargetObjectName"] = schemaObjectNameToJSON(action.TargetObjectName)
	}
	return addSpan(node, frag(action))
}
