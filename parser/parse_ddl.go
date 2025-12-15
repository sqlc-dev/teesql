// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"strings"

	"github.com/kyleconroy/teesql/ast"
)

func (p *Parser) parseRevertStatement() (*ast.RevertStatement, error) {
	// Consume REVERT
	p.nextToken()

	stmt := &ast.RevertStatement{}

	// Check for WITH COOKIE = expression
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		if p.curTok.Type != TokenCookie {
			return nil, fmt.Errorf("expected COOKIE after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume COOKIE

		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after COOKIE, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume =

		cookie, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.Cookie = cookie
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropStatement() (ast.Statement, error) {
	// Consume DROP
	p.nextToken()

	// Check what type of DROP statement this is
	if p.curTok.Type == TokenDatabase {
		return p.parseDropDatabaseStatement()
	}

	if p.curTok.Type == TokenExternal {
		return p.parseDropExternalStatement()
	}

	if p.curTok.Type == TokenTable {
		return p.parseDropTableStatement()
	}

	if p.curTok.Type == TokenIndex {
		return p.parseDropIndexStatement()
	}

	// Handle keyword-based DROP statements
	switch strings.ToUpper(p.curTok.Literal) {
	case "SEQUENCE":
		return p.parseDropSequenceStatement()
	case "SEARCH":
		return p.parseDropSearchPropertyListStatement()
	case "SERVER":
		return p.parseDropServerRoleStatement()
	case "AVAILABILITY":
		return p.parseDropAvailabilityGroupStatement()
	case "FEDERATION":
		return p.parseDropFederationStatement()
	case "VIEW":
		return p.parseDropViewStatement()
	case "PROCEDURE", "PROC":
		return p.parseDropProcedureStatement()
	case "FUNCTION":
		return p.parseDropFunctionStatement()
	case "TRIGGER":
		return p.parseDropTriggerStatement()
	case "STATISTICS":
		return p.parseDropStatisticsStatement()
	case "DEFAULT":
		return p.parseDropDefaultStatement()
	case "RULE":
		return p.parseDropRuleStatement()
	case "SCHEMA":
		return p.parseDropSchemaStatement()
	}

	return nil, fmt.Errorf("unexpected token after DROP: %s", p.curTok.Literal)
}

func (p *Parser) parseDropExternalStatement() (ast.Statement, error) {
	// Consume EXTERNAL
	p.nextToken()

	if p.curTok.Type == TokenLanguage {
		return p.parseDropExternalLanguageStatement()
	}

	if strings.ToUpper(p.curTok.Literal) == "LIBRARY" {
		return p.parseDropExternalLibraryStatement()
	}

	return nil, fmt.Errorf("unexpected token after EXTERNAL: %s", p.curTok.Literal)
}

func (p *Parser) parseDropExternalLanguageStatement() (*ast.DropExternalLanguageStatement, error) {
	// Consume LANGUAGE
	p.nextToken()

	stmt := &ast.DropExternalLanguageStatement{}

	// Parse language name
	stmt.Name = p.parseIdentifier()

	// Check for AUTHORIZATION
	if p.curTok.Type == TokenAuthorization {
		p.nextToken()
		stmt.Authorization = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropExternalLibraryStatement() (*ast.DropExternalLibraryStatement, error) {
	// Consume LIBRARY
	p.nextToken()

	stmt := &ast.DropExternalLibraryStatement{}

	// Parse library name
	stmt.Name = p.parseIdentifier()

	// Check for AUTHORIZATION
	if p.curTok.Type == TokenAuthorization {
		p.nextToken()
		stmt.Owner = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropDatabaseStatement() (ast.Statement, error) {
	// Consume DATABASE
	p.nextToken()

	// Check for DATABASE SCOPED CREDENTIAL
	if p.curTok.Type == TokenScoped {
		p.nextToken() // consume SCOPED

		if p.curTok.Type == TokenCredential {
			return p.parseDropCredentialStatement(true)
		}

		return nil, fmt.Errorf("unexpected token after SCOPED: %s", p.curTok.Literal)
	}

	// Plain DROP DATABASE statement
	stmt := &ast.DropDatabaseStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse database names (comma-separated)
	for {
		stmt.Databases = append(stmt.Databases, p.parseIdentifier())

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropCredentialStatement(isDatabaseScoped bool) (*ast.DropCredentialStatement, error) {
	// Consume CREDENTIAL
	p.nextToken()

	stmt := &ast.DropCredentialStatement{
		IsDatabaseScoped: isDatabaseScoped,
		IsIfExists:       false,
	}

	// Parse credential name
	if p.curTok.Type != TokenIdent {
		return nil, fmt.Errorf("expected identifier, got %s", p.curTok.Literal)
	}

	stmt.Name = &ast.Identifier{
		Value:     p.curTok.Literal,
		QuoteType: "NotQuoted",
	}
	p.nextToken()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropSequenceStatement() (*ast.DropSequenceStatement, error) {
	// Consume SEQUENCE
	p.nextToken()

	stmt := &ast.DropSequenceStatement{}

	// Parse comma-separated list of schema object names
	for {
		name, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, name)

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

func (p *Parser) parseDropSearchPropertyListStatement() (*ast.DropSearchPropertyListStatement, error) {
	// Consume SEARCH
	p.nextToken()

	// Expect PROPERTY
	if strings.ToUpper(p.curTok.Literal) != "PROPERTY" {
		return nil, fmt.Errorf("expected PROPERTY after SEARCH, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect LIST
	if strings.ToUpper(p.curTok.Literal) != "LIST" {
		return nil, fmt.Errorf("expected LIST after PROPERTY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropSearchPropertyListStatement{}
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropServerRoleStatement() (*ast.DropServerRoleStatement, error) {
	// Consume SERVER
	p.nextToken()

	// Expect ROLE
	if strings.ToUpper(p.curTok.Literal) != "ROLE" {
		return nil, fmt.Errorf("expected ROLE after SERVER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropServerRoleStatement{}
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropAvailabilityGroupStatement() (*ast.DropAvailabilityGroupStatement, error) {
	// Consume AVAILABILITY
	p.nextToken()

	// Expect GROUP
	if strings.ToUpper(p.curTok.Literal) != "GROUP" {
		return nil, fmt.Errorf("expected GROUP after AVAILABILITY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropAvailabilityGroupStatement{}
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropFederationStatement() (*ast.DropFederationStatement, error) {
	// Consume FEDERATION
	p.nextToken()

	stmt := &ast.DropFederationStatement{}
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropTableStatement() (*ast.DropTableStatement, error) {
	// Consume TABLE
	p.nextToken()

	stmt := &ast.DropTableStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse table names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropViewStatement() (*ast.DropViewStatement, error) {
	// Consume VIEW
	p.nextToken()

	stmt := &ast.DropViewStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse view names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropProcedureStatement() (*ast.DropProcedureStatement, error) {
	// Consume PROCEDURE or PROC
	p.nextToken()

	stmt := &ast.DropProcedureStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse procedure names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropFunctionStatement() (*ast.DropFunctionStatement, error) {
	// Consume FUNCTION
	p.nextToken()

	stmt := &ast.DropFunctionStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse function names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropTriggerStatement() (*ast.DropTriggerStatement, error) {
	// Consume TRIGGER
	p.nextToken()

	stmt := &ast.DropTriggerStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse trigger names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Check for ON DATABASE or ON ALL SERVER
	if p.curTok.Type == TokenOn {
		p.nextToken()
		if p.curTok.Type == TokenDatabase {
			stmt.TriggerScope = "Database"
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "ALL" {
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "SERVER" {
				stmt.TriggerScope = "AllServer"
				p.nextToken()
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropIndexStatement() (*ast.DropIndexStatement, error) {
	// Consume INDEX
	p.nextToken()

	stmt := &ast.DropIndexStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse index clauses (comma-separated)
	for {
		clause := &ast.DropIndexClause{}
		idx, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		clause.Index = idx
		stmt.Indexes = append(stmt.Indexes, clause)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropStatisticsStatement() (*ast.DropStatisticsStatement, error) {
	// Consume STATISTICS
	p.nextToken()

	stmt := &ast.DropStatisticsStatement{}

	// Parse statistic names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropDefaultStatement() (*ast.DropDefaultStatement, error) {
	// Consume DEFAULT
	p.nextToken()

	stmt := &ast.DropDefaultStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse default names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropRuleStatement() (*ast.DropRuleStatement, error) {
	// Consume RULE
	p.nextToken()

	stmt := &ast.DropRuleStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse rule names (comma-separated)
	for {
		obj, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, obj)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropSchemaStatement() (*ast.DropSchemaStatement, error) {
	// Consume SCHEMA
	p.nextToken()

	stmt := &ast.DropSchemaStatement{}

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse schema name
	schema, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Schema = schema

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterStatement() (ast.Statement, error) {
	// Consume ALTER
	p.nextToken()

	// Check what type of ALTER statement this is
	switch p.curTok.Type {
	case TokenTable:
		return p.parseAlterTableStatement()
	case TokenMaster:
		return p.parseAlterMasterKeyStatement()
	case TokenSchema:
		return p.parseAlterSchemaStatement()
	case TokenLogin:
		return p.parseAlterLoginStatement()
	case TokenMessage:
		return p.parseAlterMessageTypeStatement()
	case TokenDatabase:
		return p.parseAlterDatabaseStatement()
	case TokenFunction:
		return p.parseAlterFunctionStatement()
	case TokenTrigger:
		return p.parseAlterTriggerStatement()
	case TokenIndex:
		return p.parseAlterIndexStatement()
	case TokenUser:
		return p.parseAlterUserStatement()
	case TokenAsymmetric:
		return p.parseAlterAsymmetricKeyStatement()
	case TokenSymmetric:
		return p.parseAlterSymmetricKeyStatement()
	case TokenCertificate:
		return p.parseAlterCertificateStatement()
	case TokenCredential:
		return p.parseAlterCredentialStatement()
	case TokenIdent:
		// Handle keywords that are not reserved tokens
		switch strings.ToUpper(p.curTok.Literal) {
		case "ROLE":
			return p.parseAlterRoleStatement()
		case "SERVER":
			return p.parseAlterServerConfigurationStatement()
		case "REMOTE":
			return p.parseAlterRemoteServiceBindingStatement()
		case "XML":
			return p.parseAlterXmlSchemaCollectionStatement()
		case "ROUTE":
			return p.parseAlterRouteStatement()
		case "ASSEMBLY":
			return p.parseAlterAssemblyStatement()
		case "ENDPOINT":
			return p.parseAlterEndpointStatement()
		case "SERVICE":
			return p.parseAlterServiceStatement()
		case "CERTIFICATE":
			return p.parseAlterCertificateStatement()
		case "APPLICATION":
			return p.parseAlterApplicationRoleStatement()
		case "ASYMMETRIC":
			return p.parseAlterAsymmetricKeyStatement()
		case "QUEUE":
			return p.parseAlterQueueStatement()
		case "PARTITION":
			return p.parseAlterPartitionStatement()
		case "FULLTEXT":
			return p.parseAlterFulltextStatement()
		case "SYMMETRIC":
			return p.parseAlterSymmetricKeyStatement()
		case "CREDENTIAL":
			return p.parseAlterCredentialStatement()
		case "SERVICE_MASTER_KEY":
			return p.parseAlterServiceMasterKeyStatement()
		}
		return nil, fmt.Errorf("unexpected token after ALTER: %s", p.curTok.Literal)
	default:
		return nil, fmt.Errorf("unexpected token after ALTER: %s", p.curTok.Literal)
	}
}

func (p *Parser) parseAlterDatabaseStatement() (ast.Statement, error) {
	// Consume DATABASE
	p.nextToken()

	// Check for SCOPED CREDENTIAL
	if p.curTok.Type == TokenScoped {
		p.nextToken() // consume SCOPED
		if p.curTok.Type == TokenCredential {
			return p.parseAlterDatabaseScopedCredentialStatement()
		}
	}

	// Parse database name followed by various commands
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
		dbName := p.parseIdentifier()

		switch p.curTok.Type {
		case TokenSet:
			return p.parseAlterDatabaseSetStatement(dbName)
		case TokenAdd:
			return p.parseAlterDatabaseAddStatement(dbName)
		default:
			// Check for MODIFY or REMOVE
			if strings.ToUpper(p.curTok.Literal) == "MODIFY" {
				return p.parseAlterDatabaseModifyStatement(dbName)
			}
			if strings.ToUpper(p.curTok.Literal) == "REMOVE" {
				return p.parseAlterDatabaseRemoveStatement(dbName)
			}
		}
		// Lenient - skip rest of statement
		p.skipToEndOfStatement()
		return &ast.AlterDatabaseSetStatement{DatabaseName: dbName}, nil
	}

	// Lenient: skip unknown database names (like $(tempdb) SQLCMD variables)
	p.skipToEndOfStatement()
	return &ast.AlterDatabaseSetStatement{}, nil
}

func (p *Parser) parseAlterDatabaseSetStatement(dbName *ast.Identifier) (*ast.AlterDatabaseSetStatement, error) {
	// Consume SET
	p.nextToken()

	stmt := &ast.AlterDatabaseSetStatement{
		DatabaseName: dbName,
	}

	// Parse options
	for {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		switch optionName {
		case "ACCELERATED_DATABASE_RECOVERY":
			// Expect = for this option
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after %s, got %s", optionName, p.curTok.Literal)
			}
			p.nextToken()
			optionValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.AcceleratedDatabaseRecoveryDatabaseOption{
				OptionKind:  "AcceleratedDatabaseRecovery",
				OptionState: capitalizeFirst(optionValue),
			}
			stmt.Options = append(stmt.Options, opt)
		case "TEMPORAL_HISTORY_RETENTION":
			// This option uses ON/OFF directly without =
			optionValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "TemporalHistoryRetention",
				OptionState: capitalizeFirst(optionValue),
			}
			stmt.Options = append(stmt.Options, opt)
		}

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

func (p *Parser) parseAlterDatabaseAddStatement(dbName *ast.Identifier) (ast.Statement, error) {
	// Consume ADD
	p.nextToken()

	switch strings.ToUpper(p.curTok.Literal) {
	case "FILE":
		p.nextToken() // consume FILE
		stmt := &ast.AlterDatabaseAddFileStatement{
			DatabaseName: dbName,
		}
		p.skipToEndOfStatement()
		return stmt, nil
	case "LOG":
		p.nextToken() // consume LOG
		if strings.ToUpper(p.curTok.Literal) == "FILE" {
			p.nextToken() // consume FILE
		}
		stmt := &ast.AlterDatabaseAddFileStatement{
			DatabaseName: dbName,
		}
		p.skipToEndOfStatement()
		return stmt, nil
	case "FILEGROUP":
		p.nextToken() // consume FILEGROUP
		stmt := &ast.AlterDatabaseAddFileGroupStatement{
			DatabaseName:  dbName,
			FileGroupName: p.parseIdentifier(),
		}
		p.skipToEndOfStatement()
		return stmt, nil
	default:
		p.skipToEndOfStatement()
		return &ast.AlterDatabaseSetStatement{DatabaseName: dbName}, nil
	}
}

func (p *Parser) parseAlterDatabaseModifyStatement(dbName *ast.Identifier) (ast.Statement, error) {
	// Consume MODIFY
	p.nextToken()

	switch strings.ToUpper(p.curTok.Literal) {
	case "FILE":
		p.nextToken() // consume FILE
		stmt := &ast.AlterDatabaseModifyFileStatement{
			DatabaseName: dbName,
		}
		p.skipToEndOfStatement()
		return stmt, nil
	case "FILEGROUP":
		p.nextToken() // consume FILEGROUP
		stmt := &ast.AlterDatabaseModifyFileGroupStatement{
			DatabaseName:  dbName,
			FileGroupName: p.parseIdentifier(),
		}
		p.skipToEndOfStatement()
		return stmt, nil
	case "NAME":
		p.nextToken() // consume NAME
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		stmt := &ast.AlterDatabaseModifyNameStatement{
			DatabaseName: dbName,
			NewName:      p.parseIdentifier(),
		}
		p.skipToEndOfStatement()
		return stmt, nil
	default:
		p.skipToEndOfStatement()
		return &ast.AlterDatabaseSetStatement{DatabaseName: dbName}, nil
	}
}

func (p *Parser) parseAlterDatabaseRemoveStatement(dbName *ast.Identifier) (ast.Statement, error) {
	// Consume REMOVE
	p.nextToken()

	switch strings.ToUpper(p.curTok.Literal) {
	case "FILE":
		p.nextToken() // consume FILE
		stmt := &ast.AlterDatabaseRemoveFileStatement{
			DatabaseName: dbName,
			FileName:     p.parseIdentifier(),
		}
		p.skipToEndOfStatement()
		return stmt, nil
	case "FILEGROUP":
		p.nextToken() // consume FILEGROUP
		stmt := &ast.AlterDatabaseRemoveFileGroupStatement{
			DatabaseName:  dbName,
			FileGroupName: p.parseIdentifier(),
		}
		p.skipToEndOfStatement()
		return stmt, nil
	default:
		p.skipToEndOfStatement()
		return &ast.AlterDatabaseSetStatement{DatabaseName: dbName}, nil
	}
}

func (p *Parser) parseAlterDatabaseScopedCredentialStatement() (*ast.AlterCredentialStatement, error) {
	// Consume CREDENTIAL
	p.nextToken()

	stmt := &ast.AlterCredentialStatement{
		IsDatabaseScoped: true,
	}

	// Parse credential name
	stmt.Name = p.parseIdentifier()

	// Check for WITH (optional for lenient parsing)
	if p.curTok.Type != TokenWith {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse options
	for {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after %s, got %s", optionName, p.curTok.Literal)
		}
		p.nextToken()

		// Parse value
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		switch optionName {
		case "IDENTITY":
			stmt.Identity = expr
		case "SECRET":
			stmt.Secret = expr
		}

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

func (p *Parser) parseAlterServerConfigurationStatement() (ast.Statement, error) {
	// Consume SERVER
	p.nextToken()

	// Expect CONFIGURATION
	if strings.ToUpper(p.curTok.Literal) != "CONFIGURATION" {
		return nil, fmt.Errorf("expected CONFIGURATION after SERVER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect SET
	if p.curTok.Type != TokenSet {
		return nil, fmt.Errorf("expected SET after CONFIGURATION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check what type of SET it is
	if strings.ToUpper(p.curTok.Literal) == "SOFTNUMA" {
		return p.parseAlterServerConfigurationSetSoftNumaStatement()
	}

	return nil, fmt.Errorf("unexpected token after SET: %s", p.curTok.Literal)
}

func (p *Parser) parseAlterServerConfigurationSetSoftNumaStatement() (*ast.AlterServerConfigurationSetSoftNumaStatement, error) {
	// Consume SOFTNUMA
	p.nextToken()

	stmt := &ast.AlterServerConfigurationSetSoftNumaStatement{}

	// Parse ON or OFF
	optionState := strings.ToUpper(p.curTok.Literal)
	if optionState != "ON" && optionState != "OFF" {
		return nil, fmt.Errorf("expected ON or OFF after SOFTNUMA, got %s", p.curTok.Literal)
	}
	p.nextToken()

	option := &ast.AlterServerConfigurationSoftNumaOption{
		OptionKind: "OnOff",
		OptionValue: &ast.OnOffOptionValue{
			OptionState: capitalizeFirst(optionState),
		},
	}
	stmt.Options = append(stmt.Options, option)

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func (p *Parser) parseAlterMessageTypeStatement() (*ast.AlterMessageTypeStatement, error) {
	// Consume MESSAGE
	p.nextToken()

	// Expect TYPE
	if strings.ToUpper(p.curTok.Literal) != "TYPE" {
		return nil, fmt.Errorf("expected TYPE after MESSAGE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterMessageTypeStatement{}

	// Parse message type name
	stmt.Name = p.parseIdentifier()

	// Check for VALIDATION (optional for lenient parsing)
	if strings.ToUpper(p.curTok.Literal) != "VALIDATION" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after VALIDATION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse validation method
	validationMethod := strings.ToUpper(p.curTok.Literal)
	switch validationMethod {
	case "EMPTY":
		stmt.ValidationMethod = "Empty"
		p.nextToken()
	case "NONE":
		stmt.ValidationMethod = "None"
		p.nextToken()
	case "WELL_FORMED_XML":
		stmt.ValidationMethod = "WellFormedXml"
		p.nextToken()
	case "VALID_XML":
		stmt.ValidationMethod = "ValidXml"
		p.nextToken()
		// Expect WITH SCHEMA COLLECTION
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			if strings.ToUpper(p.curTok.Literal) == "SCHEMA" {
				p.nextToken() // consume SCHEMA
				if strings.ToUpper(p.curTok.Literal) == "COLLECTION" {
					p.nextToken() // consume COLLECTION
					collName, err := p.parseSchemaObjectName()
					if err != nil {
						return nil, err
					}
					stmt.XmlSchemaCollectionName = collName
				}
			}
		}
	default:
		return nil, fmt.Errorf("unexpected validation method: %s", p.curTok.Literal)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterTableStatement() (ast.Statement, error) {
	// Consume TABLE
	p.nextToken()

	// Parse table name
	tableName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}

	// Check what kind of ALTER TABLE statement this is
	if p.curTok.Type == TokenDrop {
		return p.parseAlterTableDropStatement(tableName)
	}

	// Check for ALTER INDEX
	if p.curTok.Type == TokenAlter && p.peekTok.Type == TokenIndex {
		return p.parseAlterTableAlterIndexStatement(tableName)
	}

	// Check for ALTER COLUMN
	if p.curTok.Type == TokenAlter && strings.ToUpper(p.peekTok.Literal) == "COLUMN" {
		return p.parseAlterTableAlterColumnStatement(tableName)
	}

	// Check for ADD
	if strings.ToUpper(p.curTok.Literal) == "ADD" {
		return p.parseAlterTableAddStatement(tableName)
	}

	// Check for ENABLE/DISABLE TRIGGER
	if strings.ToUpper(p.curTok.Literal) == "ENABLE" || strings.ToUpper(p.curTok.Literal) == "DISABLE" {
		return p.parseAlterTableTriggerModificationStatement(tableName)
	}

	// Check for SWITCH
	if strings.ToUpper(p.curTok.Literal) == "SWITCH" {
		return p.parseAlterTableSwitchStatement(tableName)
	}

	// Check for WITH CHECK/NOCHECK or CHECK/NOCHECK CONSTRAINT
	if strings.ToUpper(p.curTok.Literal) == "WITH" || strings.ToUpper(p.curTok.Literal) == "CHECK" || strings.ToUpper(p.curTok.Literal) == "NOCHECK" {
		return p.parseAlterTableConstraintModificationStatement(tableName)
	}

	return nil, fmt.Errorf("unexpected token in ALTER TABLE: %s", p.curTok.Literal)
}

func (p *Parser) parseAlterTableDropStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableDropTableElementStatement, error) {
	// Consume DROP
	p.nextToken()

	stmt := &ast.AlterTableDropTableElementStatement{
		SchemaObjectName: tableName,
	}

	// Parse multiple elements separated by commas
	// Format: DROP [COLUMN] name, [CONSTRAINT] name, [INDEX] name, ...
	var currentElementType string = "NotSpecified"

	for {
		// Check for element type keyword
		switch {
		case strings.ToUpper(p.curTok.Literal) == "COLUMN":
			currentElementType = "Column"
			p.nextToken()
		case strings.ToUpper(p.curTok.Literal) == "CONSTRAINT":
			currentElementType = "Constraint"
			p.nextToken()
		case p.curTok.Type == TokenIndex:
			currentElementType = "Index"
			p.nextToken()
		}

		// Parse the element name
		if p.curTok.Type != TokenIdent && p.curTok.Type != TokenLBracket {
			if len(stmt.AlterTableDropTableElements) > 0 {
				break
			}
			return nil, fmt.Errorf("expected identifier, got %s", p.curTok.Literal)
		}

		element := &ast.AlterTableDropTableElement{
			TableElementType: currentElementType,
			Name:             p.parseIdentifier(),
			IsIfExists:       false,
		}
		stmt.AlterTableDropTableElements = append(stmt.AlterTableDropTableElements, element)

		// After adding an element, reset type to NotSpecified for next element
		// unless another type keyword is found
		currentElementType = "NotSpecified"

		// Check for comma to continue or end
		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
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

func (p *Parser) parseAlterTableAlterIndexStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableAlterIndexStatement, error) {
	// Consume ALTER
	p.nextToken()

	// Consume INDEX
	p.nextToken()

	stmt := &ast.AlterTableAlterIndexStatement{
		SchemaObjectName: tableName,
	}

	// Parse index name
	stmt.IndexIdentifier = p.parseIdentifier()

	// Parse operation type (REBUILD, DISABLE, etc.)
	operation := strings.ToUpper(p.curTok.Literal)
	switch operation {
	case "REBUILD":
		stmt.AlterIndexType = "Rebuild"
		p.nextToken()
	case "DISABLE":
		stmt.AlterIndexType = "Disable"
		p.nextToken()
	case "REORGANIZE":
		stmt.AlterIndexType = "Reorganize"
		p.nextToken()
	default:
		return nil, fmt.Errorf("expected REBUILD, DISABLE, or REORGANIZE after index name, got %s", p.curTok.Literal)
	}

	// Parse optional WITH clause for options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			// Parse option name
			optionName := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after option name, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =

			// Parse option value
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			option := &ast.IndexExpressionOption{
				OptionKind: convertIndexOptionKind(optionName),
				Expression: expr,
			}
			stmt.IndexOptions = append(stmt.IndexOptions, option)

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

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func convertIndexOptionKind(name string) string {
	optionMap := map[string]string{
		"BUCKET_COUNT":           "BucketCount",
		"PAD_INDEX":              "PadIndex",
		"FILLFACTOR":             "FillFactor",
		"SORT_IN_TEMPDB":         "SortInTempDB",
		"IGNORE_DUP_KEY":         "IgnoreDupKey",
		"STATISTICS_NORECOMPUTE": "StatisticsNoRecompute",
		"DROP_EXISTING":          "DropExisting",
		"ONLINE":                 "Online",
		"ALLOW_ROW_LOCKS":        "AllowRowLocks",
		"ALLOW_PAGE_LOCKS":       "AllowPageLocks",
		"MAXDOP":                 "MaxDop",
		"DATA_COMPRESSION":       "DataCompression",
	}
	if mapped, ok := optionMap[name]; ok {
		return mapped
	}
	return name
}

func (p *Parser) parseAlterTableAlterColumnStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableAlterColumnStatement, error) {
	// Consume ALTER
	p.nextToken()
	// Consume COLUMN
	p.nextToken()

	stmt := &ast.AlterTableAlterColumnStatement{
		SchemaObjectName:            tableName,
		AlterTableAlterColumnOption: "NoOptionDefined",
	}

	// Parse column name
	stmt.ColumnIdentifier = p.parseIdentifier()

	// Parse data type - be lenient if no data type is provided
	dataType, err := p.parseDataType()
	if err != nil {
		// Lenient: return statement without data type
		p.skipToEndOfStatement()
		return stmt, nil
	}
	stmt.DataType = dataType

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterTableAddStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableAddTableElementStatement, error) {
	// Consume ADD
	p.nextToken()

	stmt := &ast.AlterTableAddTableElementStatement{
		SchemaObjectName:             tableName,
		ExistingRowsCheckEnforcement: "NotSpecified",
		Definition:                   &ast.TableDefinition{},
	}

	// Check if this is ADD CONSTRAINT
	if strings.ToUpper(p.curTok.Literal) == "CONSTRAINT" {
		p.nextToken() // consume CONSTRAINT
		// Parse constraint name
		constraintName := p.parseIdentifier()
		_ = constraintName // We'll use this later when we implement full constraint support
		// Skip to end of statement (lenient parsing for incomplete constraints)
		p.skipToEndOfStatement()
		return stmt, nil
	}

	// Check if this is ADD INDEX
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX

		indexDef := &ast.IndexDefinition{}

		// Parse index name
		indexDef.Name = p.parseIdentifier()

		// Parse optional UNIQUE, CLUSTERED, NONCLUSTERED, HASH keywords
		var indexTypeKind string
		for {
			switch strings.ToUpper(p.curTok.Literal) {
			case "UNIQUE":
				indexDef.Unique = true
				p.nextToken()
				continue
			case "CLUSTERED":
				indexTypeKind = "Clustered"
				p.nextToken()
				continue
			case "NONCLUSTERED":
				// Check for HASH suffix
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "HASH" {
					indexTypeKind = "NonClusteredHash"
					p.nextToken()
				} else {
					indexTypeKind = "NonClustered"
				}
				continue
			case "HASH":
				if indexTypeKind == "" {
					indexTypeKind = "NonClusteredHash"
				}
				p.nextToken()
				continue
			}
			break
		}

		if indexTypeKind != "" {
			indexDef.IndexType = &ast.IndexType{
				IndexTypeKind: indexTypeKind,
			}
		}

		// Parse column list (c1, c2, ...)
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

				col := &ast.ColumnWithSortOrder{
					Column:    colRef,
					SortOrder: ast.SortOrderNotSpecified,
				}

				// Check for ASC/DESC
				switch strings.ToUpper(p.curTok.Literal) {
				case "ASC":
					col.SortOrder = ast.SortOrderAscending
					p.nextToken()
				case "DESC":
					col.SortOrder = ast.SortOrderDescending
					p.nextToken()
				}

				indexDef.Columns = append(indexDef.Columns, col)

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

		// Parse optional WITH clause
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH

			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (

				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					// Parse option name
					optionName := strings.ToUpper(p.curTok.Literal)
					p.nextToken()

					if p.curTok.Type != TokenEquals {
						return nil, fmt.Errorf("expected = after option name, got %s", p.curTok.Literal)
					}
					p.nextToken() // consume =

					// Parse option value
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}

					option := &ast.IndexExpressionOption{
						OptionKind: convertIndexOptionKind(optionName),
						Expression: expr,
					}
					indexDef.IndexOptions = append(indexDef.IndexOptions, option)

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

		stmt.Definition.Indexes = append(stmt.Definition.Indexes, indexDef)
	} else {
		// Parse column definition (column_name data_type ...)
		colDef, err := p.parseColumnDefinition()
		if err != nil {
			return nil, err
		}
		stmt.Definition.ColumnDefinitions = append(stmt.Definition.ColumnDefinitions, colDef)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterTableTriggerModificationStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableTriggerModificationStatement, error) {
	stmt := &ast.AlterTableTriggerModificationStatement{
		SchemaObjectName: tableName,
	}

	// Parse ENABLE or DISABLE
	if strings.ToUpper(p.curTok.Literal) == "ENABLE" {
		stmt.TriggerEnforcement = "Enable"
	} else {
		stmt.TriggerEnforcement = "Disable"
	}
	p.nextToken()

	// Expect TRIGGER keyword
	if strings.ToUpper(p.curTok.Literal) != "TRIGGER" {
		return nil, fmt.Errorf("expected TRIGGER after %s, got %s", stmt.TriggerEnforcement, p.curTok.Literal)
	}
	p.nextToken()

	// Check for ALL or trigger names
	if strings.ToUpper(p.curTok.Literal) == "ALL" {
		stmt.All = true
		p.nextToken()
	} else {
		stmt.All = false
		// Parse trigger names (comma-separated)
		for {
			stmt.TriggerNames = append(stmt.TriggerNames, p.parseIdentifier())

			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume comma
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterTableSwitchStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableSwitchStatement, error) {
	stmt := &ast.AlterTableSwitchStatement{
		SchemaObjectName: tableName,
	}

	// Consume SWITCH
	p.nextToken()

	// Check for PARTITION clause on source
	if strings.ToUpper(p.curTok.Literal) == "PARTITION" {
		p.nextToken()
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.SourcePartition = expr
	}

	// Expect TO
	if strings.ToUpper(p.curTok.Literal) != "TO" {
		return nil, fmt.Errorf("expected TO, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse target table name
	targetTable, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.TargetTable = targetTable

	// Check for PARTITION clause on target
	if strings.ToUpper(p.curTok.Literal) == "PARTITION" {
		p.nextToken()
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.TargetPartition = expr
	}

	// Check for WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if p.curTok.Type == TokenLParen {
			p.nextToken()

			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optionName := strings.ToUpper(p.curTok.Literal)
				p.nextToken()

				if optionName == "TRUNCATE_TARGET" {
					if p.curTok.Type == TokenEquals {
						p.nextToken()
						value := strings.ToUpper(p.curTok.Literal)
						p.nextToken()
						opt := &ast.TruncateTargetTableSwitchOption{
							TruncateTarget: value == "ON",
							OptionKind:     "TruncateTarget",
						}
						stmt.Options = append(stmt.Options, opt)
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

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterTableConstraintModificationStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableConstraintModificationStatement, error) {
	stmt := &ast.AlterTableConstraintModificationStatement{
		SchemaObjectName:             tableName,
		ExistingRowsCheckEnforcement: "NotSpecified",
	}

	// Check for WITH CHECK/NOCHECK
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "CHECK" {
			stmt.ExistingRowsCheckEnforcement = "Check"
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "NOCHECK" {
			stmt.ExistingRowsCheckEnforcement = "NoCheck"
			p.nextToken()
		}
	}

	// Expect CHECK or NOCHECK
	if strings.ToUpper(p.curTok.Literal) == "CHECK" {
		stmt.ConstraintEnforcement = "Check"
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NOCHECK" {
		stmt.ConstraintEnforcement = "NoCheck"
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected CHECK or NOCHECK, got %s", p.curTok.Literal)
	}

	// Expect CONSTRAINT
	if strings.ToUpper(p.curTok.Literal) != "CONSTRAINT" {
		return nil, fmt.Errorf("expected CONSTRAINT, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for ALL or constraint names
	if strings.ToUpper(p.curTok.Literal) == "ALL" {
		stmt.All = true
		p.nextToken()
	} else {
		stmt.All = false
		// Parse constraint names (comma-separated)
		for {
			stmt.ConstraintNames = append(stmt.ConstraintNames, p.parseIdentifier())
			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume comma
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterRoleStatement() (*ast.AlterRoleStatement, error) {
	// Consume ROLE
	p.nextToken()

	stmt := &ast.AlterRoleStatement{}

	// Parse role name
	stmt.Name = p.parseIdentifier()

	// Parse action: ADD MEMBER, DROP MEMBER, or WITH NAME =
	switch strings.ToUpper(p.curTok.Literal) {
	case "ADD":
		p.nextToken() // consume ADD
		if strings.ToUpper(p.curTok.Literal) != "MEMBER" {
			return nil, fmt.Errorf("expected MEMBER after ADD, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume MEMBER
		action := &ast.AddMemberAlterRoleAction{}
		action.Member = p.parseIdentifier()
		stmt.Action = action

	case "DROP":
		p.nextToken() // consume DROP
		if strings.ToUpper(p.curTok.Literal) != "MEMBER" {
			return nil, fmt.Errorf("expected MEMBER after DROP, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume MEMBER
		action := &ast.DropMemberAlterRoleAction{}
		action.Member = p.parseIdentifier()
		stmt.Action = action

	case "WITH":
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) != "NAME" {
			return nil, fmt.Errorf("expected NAME after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume NAME
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after NAME, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume =
		action := &ast.RenameAlterRoleAction{}
		action.NewName = p.parseIdentifier()
		stmt.Action = action

	default:
		// Handle incomplete statement
		p.skipToEndOfStatement()
		return stmt, nil
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterRemoteServiceBindingStatement() (*ast.AlterRemoteServiceBindingStatement, error) {
	// Consume REMOTE
	p.nextToken()

	// Expect SERVICE
	if strings.ToUpper(p.curTok.Literal) != "SERVICE" {
		return nil, fmt.Errorf("expected SERVICE after REMOTE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect BINDING
	if strings.ToUpper(p.curTok.Literal) != "BINDING" {
		return nil, fmt.Errorf("expected BINDING after SERVICE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterRemoteServiceBindingStatement{}

	// Parse binding name
	stmt.Name = p.parseIdentifier()

	// Check for WITH (optional for lenient parsing)
	if p.curTok.Type != TokenWith {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse options
	for {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after %s, got %s", optionName, p.curTok.Literal)
		}
		p.nextToken()

		switch optionName {
		case "USER":
			opt := &ast.UserRemoteServiceBindingOption{
				OptionKind: "User",
				User:       p.parseIdentifier(),
			}
			stmt.Options = append(stmt.Options, opt)
		case "ANONYMOUS":
			optState := strings.ToUpper(p.curTok.Literal)
			var state string
			if optState == "ON" {
				state = "On"
			} else {
				state = "Off"
			}
			p.nextToken()
			opt := &ast.OnOffRemoteServiceBindingOption{
				OptionKind:  "Anonymous",
				OptionState: state,
			}
			stmt.Options = append(stmt.Options, opt)
		}

		// Check for comma
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

func (p *Parser) parseAlterXmlSchemaCollectionStatement() (*ast.AlterXmlSchemaCollectionStatement, error) {
	// Consume XML
	p.nextToken()

	// Expect SCHEMA
	if strings.ToUpper(p.curTok.Literal) != "SCHEMA" {
		return nil, fmt.Errorf("expected SCHEMA after XML, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect COLLECTION
	if strings.ToUpper(p.curTok.Literal) != "COLLECTION" {
		return nil, fmt.Errorf("expected COLLECTION after SCHEMA, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterXmlSchemaCollectionStatement{}

	// Parse collection name (can be one or two parts)
	name, _ := p.parseSchemaObjectName()
	stmt.Name = name

	// Check for ADD (optional for lenient parsing)
	if strings.ToUpper(p.curTok.Literal) != "ADD" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse expression (variable or string literal)
	expr, err := p.parseScalarExpression()
	if err != nil {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	stmt.Expression = expr

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateXmlSchemaCollectionStatement() (*ast.CreateXmlSchemaCollectionStatement, error) {
	// Consume XML
	p.nextToken()

	// Expect SCHEMA
	if strings.ToUpper(p.curTok.Literal) != "SCHEMA" {
		return nil, fmt.Errorf("expected SCHEMA after XML, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect COLLECTION
	if strings.ToUpper(p.curTok.Literal) != "COLLECTION" {
		return nil, fmt.Errorf("expected COLLECTION after SCHEMA, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.CreateXmlSchemaCollectionStatement{}

	// Parse collection name (can be one or two parts)
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Expect AS
	if p.curTok.Type != TokenAs {
		return nil, fmt.Errorf("expected AS after collection name, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse expression (variable or string literal)
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Expression = expr

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateSearchPropertyListStatement() (*ast.CreateSearchPropertyListStatement, error) {
	// Consume SEARCH
	p.nextToken()

	// Expect PROPERTY
	if strings.ToUpper(p.curTok.Literal) != "PROPERTY" {
		return nil, fmt.Errorf("expected PROPERTY after SEARCH, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect LIST
	if strings.ToUpper(p.curTok.Literal) != "LIST" {
		return nil, fmt.Errorf("expected LIST after PROPERTY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.CreateSearchPropertyListStatement{}

	// Parse name
	stmt.Name = p.parseIdentifier()

	// Check for optional FROM clause
	if strings.ToUpper(p.curTok.Literal) == "FROM" {
		p.nextToken()
		// Parse source property list name (can be one or two parts)
		multiPart := &ast.MultiPartIdentifier{}
		for {
			id := p.parseIdentifier()
			multiPart.Identifiers = append(multiPart.Identifiers, id)
			if p.curTok.Type == TokenDot {
				p.nextToken()
			} else {
				break
			}
		}
		multiPart.Count = len(multiPart.Identifiers)
		stmt.SourceSearchPropertyList = multiPart
	}

	// Check for optional AUTHORIZATION clause
	if p.curTok.Type == TokenAuthorization {
		p.nextToken()
		stmt.Owner = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterMasterKeyStatement() (*ast.AlterMasterKeyStatement, error) {
	// Consume MASTER
	p.nextToken()

	// Expect KEY
	if p.curTok.Type != TokenKey {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterMasterKeyStatement{}

	// Check for FORCE or operation
	if strings.ToUpper(p.curTok.Literal) == "FORCE" {
		p.nextToken() // consume FORCE
		if strings.ToUpper(p.curTok.Literal) != "REGENERATE" {
			return nil, fmt.Errorf("expected REGENERATE after FORCE, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume REGENERATE
		stmt.Option = "ForceRegenerate"
	} else if strings.ToUpper(p.curTok.Literal) == "REGENERATE" {
		p.nextToken() // consume REGENERATE
		stmt.Option = "Regenerate"
	} else if strings.ToUpper(p.curTok.Literal) == "ADD" {
		p.nextToken() // consume ADD
		if p.curTok.Type != TokenEncryption {
			return nil, fmt.Errorf("expected ENCRYPTION after ADD, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ENCRYPTION
		if p.curTok.Type != TokenBy {
			return nil, fmt.Errorf("expected BY after ENCRYPTION, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume BY

		if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
			p.nextToken() // consume SERVICE
			p.nextToken() // consume MASTER
			p.nextToken() // consume KEY
			stmt.Option = "AddEncryptionByServiceMasterKey"
		} else if p.curTok.Type == TokenPassword {
			stmt.Option = "AddEncryptionByPassword"
			p.nextToken() // consume PASSWORD
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after PASSWORD, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			password, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Password = password
		} else {
			return nil, fmt.Errorf("expected PASSWORD or SERVICE after BY, got %s", p.curTok.Literal)
		}
	} else if strings.ToUpper(p.curTok.Literal) == "DROP" {
		p.nextToken() // consume DROP
		if p.curTok.Type != TokenEncryption {
			return nil, fmt.Errorf("expected ENCRYPTION after DROP, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ENCRYPTION
		if p.curTok.Type != TokenBy {
			return nil, fmt.Errorf("expected BY after ENCRYPTION, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume BY

		if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
			p.nextToken() // consume SERVICE
			p.nextToken() // consume MASTER
			p.nextToken() // consume KEY
			stmt.Option = "DropEncryptionByServiceMasterKey"
		} else if p.curTok.Type == TokenPassword {
			stmt.Option = "DropEncryptionByPassword"
			p.nextToken() // consume PASSWORD
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after PASSWORD, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			password, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Password = password
		} else {
			return nil, fmt.Errorf("expected PASSWORD or SERVICE after BY, got %s", p.curTok.Literal)
		}
	} else {
		return nil, fmt.Errorf("unexpected token in ALTER MASTER KEY: %s", p.curTok.Literal)
	}

	// Handle WITH ENCRYPTION BY PASSWORD for REGENERATE
	if stmt.Option == "Regenerate" || stmt.Option == "ForceRegenerate" {
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			if p.curTok.Type != TokenEncryption {
				return nil, fmt.Errorf("expected ENCRYPTION after WITH, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume ENCRYPTION
			if p.curTok.Type != TokenBy {
				return nil, fmt.Errorf("expected BY after ENCRYPTION, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume BY
			if p.curTok.Type != TokenPassword {
				return nil, fmt.Errorf("expected PASSWORD after BY, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume PASSWORD
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after PASSWORD, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			password, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Password = password
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterSchemaStatement() (*ast.AlterSchemaStatement, error) {
	// Consume SCHEMA
	p.nextToken()

	stmt := &ast.AlterSchemaStatement{}

	// Parse schema name
	stmt.Name = p.parseIdentifier()

	// Expect TRANSFER
	if strings.ToUpper(p.curTok.Literal) != "TRANSFER" {
		return nil, fmt.Errorf("expected TRANSFER after schema name, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume TRANSFER

	// Check for optional object kind (TYPE::, OBJECT::, XML SCHEMA COLLECTION::)
	stmt.ObjectKind = "NotSpecified"
	if strings.ToUpper(p.curTok.Literal) == "TYPE" {
		p.nextToken() // consume TYPE
		if p.curTok.Type != TokenColonColon {
			return nil, fmt.Errorf("expected :: after TYPE, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ::
		stmt.ObjectKind = "Type"
	} else if strings.ToUpper(p.curTok.Literal) == "OBJECT" {
		p.nextToken() // consume OBJECT
		if p.curTok.Type != TokenColonColon {
			return nil, fmt.Errorf("expected :: after OBJECT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ::
		stmt.ObjectKind = "Object"
	} else if strings.ToUpper(p.curTok.Literal) == "XML" {
		p.nextToken() // consume XML
		p.nextToken() // consume SCHEMA
		p.nextToken() // consume COLLECTION
		if p.curTok.Type != TokenColonColon {
			return nil, fmt.Errorf("expected :: after XML SCHEMA COLLECTION, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ::
		stmt.ObjectKind = "XmlSchemaCollection"
	}

	// Parse object name
	objectName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.ObjectName = objectName

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterLoginStatement() (*ast.AlterLoginAddDropCredentialStatement, error) {
	// Consume LOGIN
	p.nextToken()

	stmt := &ast.AlterLoginAddDropCredentialStatement{}

	// Parse login name
	stmt.Name = p.parseIdentifier()

	// Check for ADD or DROP - if not present, skip to end
	if p.curTok.Type == TokenAdd {
		stmt.IsAdd = true
		p.nextToken() // consume ADD
	} else if p.curTok.Type == TokenDrop {
		stmt.IsAdd = false
		p.nextToken() // consume DROP
	} else {
		// Handle incomplete statement
		p.skipToEndOfStatement()
		return stmt, nil
	}

	// Expect CREDENTIAL
	if p.curTok.Type != TokenCredential {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse credential name
	stmt.CredentialName = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterUserStatement() (*ast.AlterUserStatement, error) {
	// Consume USER
	p.nextToken()

	stmt := &ast.AlterUserStatement{}

	// Parse user name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterRouteStatement() (*ast.AlterRouteStatement, error) {
	// Consume ROUTE
	p.nextToken()

	stmt := &ast.AlterRouteStatement{}

	// Parse route name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterAssemblyStatement() (*ast.AlterAssemblyStatement, error) {
	// Consume ASSEMBLY
	p.nextToken()

	stmt := &ast.AlterAssemblyStatement{}

	// Parse assembly name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterEndpointStatement() (*ast.AlterEndpointStatement, error) {
	// Consume ENDPOINT
	p.nextToken()

	stmt := &ast.AlterEndpointStatement{}

	// Parse endpoint name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterServiceStatement() (*ast.AlterServiceStatement, error) {
	// Consume SERVICE
	p.nextToken()

	stmt := &ast.AlterServiceStatement{}

	// Parse service name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterCertificateStatement() (*ast.AlterCertificateStatement, error) {
	// Consume CERTIFICATE
	p.nextToken()

	stmt := &ast.AlterCertificateStatement{}

	// Parse certificate name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterApplicationRoleStatement() (*ast.AlterApplicationRoleStatement, error) {
	// Consume APPLICATION
	p.nextToken()
	// Consume ROLE
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "ROLE" {
		p.nextToken()
	}

	stmt := &ast.AlterApplicationRoleStatement{}

	// Parse role name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterAsymmetricKeyStatement() (*ast.AlterAsymmetricKeyStatement, error) {
	// Consume ASYMMETRIC
	p.nextToken()
	// Consume KEY
	if p.curTok.Type == TokenKey {
		p.nextToken()
	}

	stmt := &ast.AlterAsymmetricKeyStatement{}

	// Parse key name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterQueueStatement() (*ast.AlterQueueStatement, error) {
	// Consume QUEUE
	p.nextToken()

	stmt := &ast.AlterQueueStatement{}

	// Parse queue name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterPartitionStatement() (ast.Statement, error) {
	// Consume PARTITION
	p.nextToken()

	// Check SCHEME or FUNCTION
	keyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	if keyword == "SCHEME" {
		stmt := &ast.AlterPartitionSchemeStatement{}
		stmt.Name = p.parseIdentifier()
		p.skipToEndOfStatement()
		return stmt, nil
	}

	stmt := &ast.AlterPartitionFunctionStatement{}
	stmt.Name = p.parseIdentifier()
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterFulltextStatement() (ast.Statement, error) {
	// Consume FULLTEXT
	p.nextToken()

	// Check CATALOG or INDEX
	keyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	if keyword == "CATALOG" {
		stmt := &ast.AlterFulltextCatalogStatement{}
		stmt.Name = p.parseIdentifier()
		p.skipToEndOfStatement()
		return stmt, nil
	}

	stmt := &ast.AlterFulltextIndexStatement{}
	// Parse ON table_name
	if p.curTok.Type == TokenOn {
		p.nextToken()
		name, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.OnName = name
	}
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterSymmetricKeyStatement() (*ast.AlterSymmetricKeyStatement, error) {
	// Consume SYMMETRIC
	p.nextToken()
	// Consume KEY
	if p.curTok.Type == TokenKey {
		p.nextToken()
	}

	stmt := &ast.AlterSymmetricKeyStatement{}

	// Parse key name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterCredentialStatement() (*ast.AlterCredentialStatement, error) {
	// CREDENTIAL was already consumed, but it's handled differently here
	// This gets called from the TokenIdent case
	p.nextToken() // consume CREDENTIAL

	stmt := &ast.AlterCredentialStatement{}

	// Parse credential name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterServiceMasterKeyStatement() (*ast.AlterServiceMasterKeyStatement, error) {
	// SERVICE_MASTER_KEY was matched as an identifier
	p.nextToken() // consume SERVICE_MASTER_KEY

	stmt := &ast.AlterServiceMasterKeyStatement{}

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

// skipToEndOfStatement skips tokens until end of statement (semicolon, EOF, or next statement keyword)
func (p *Parser) skipToEndOfStatement() {
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		// Check for statement keywords that indicate start of next statement
		if p.isStatementKeyword(p.curTok.Type) {
			return
		}
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			return
		}
		p.nextToken()
	}
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
}

func (p *Parser) isStatementKeyword(t TokenType) bool {
	switch t {
	case TokenSelect, TokenInsert, TokenUpdate, TokenDelete, TokenCreate, TokenAlter, TokenDrop,
		TokenDeclare, TokenExec, TokenExecute, TokenIf, TokenWhile, TokenBegin, TokenEnd,
		TokenPrint, TokenThrow, TokenGrant, TokenRevoke, TokenReturn, TokenBreak, TokenContinue,
		TokenGoto, TokenWaitfor, TokenBackup, TokenRestore, TokenUse:
		return true
	}
	return false
}

