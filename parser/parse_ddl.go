// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/teesql/ast"
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
	case "SECURITY":
		return p.parseDropSecurityPolicyStatement()
	case "WORKLOAD":
		return p.parseDropWorkloadStatement()
	case "TYPE":
		return p.parseDropTypeStatement()
	case "AGGREGATE":
		return p.parseDropAggregateStatement()
	case "SYNONYM":
		return p.parseDropSynonymStatement()
	case "USER":
		return p.parseDropUserStatement()
	case "ROLE":
		return p.parseDropRoleStatement()
	case "ASSEMBLY":
		return p.parseDropAssemblyStatement()
	case "CRYPTOGRAPHIC":
		return p.parseDropCryptographicProviderStatement()
	}

	return nil, fmt.Errorf("unexpected token after DROP: %s", p.curTok.Literal)
}

func (p *Parser) parseDropExternalStatement() (ast.Statement, error) {
	// Consume EXTERNAL
	p.nextToken()

	if p.curTok.Type == TokenLanguage {
		return p.parseDropExternalLanguageStatement()
	}

	switch strings.ToUpper(p.curTok.Literal) {
	case "LIBRARY":
		return p.parseDropExternalLibraryStatement()
	case "DATA":
		return p.parseDropExternalDataSourceStatement()
	case "FILE":
		return p.parseDropExternalFileFormatStatement()
	case "TABLE":
		return p.parseDropExternalTableStatement()
	case "RESOURCE":
		return p.parseDropExternalResourcePoolStatement()
	case "MODEL":
		return p.parseDropExternalModelStatement()
	}

	return nil, fmt.Errorf("unexpected token after EXTERNAL: %s", p.curTok.Literal)
}

func (p *Parser) parseDropExternalModelStatement() (*ast.DropExternalModelStatement, error) {
	// Consume MODEL
	p.nextToken()

	stmt := &ast.DropExternalModelStatement{}

	// Parse model name
	stmt.Name, _ = p.parseSchemaObjectName()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
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

func (p *Parser) parseDropExternalDataSourceStatement() (*ast.DropExternalDataSourceStatement, error) {
	// Consume DATA
	p.nextToken()

	// Expect SOURCE
	if strings.ToUpper(p.curTok.Literal) != "SOURCE" {
		return nil, fmt.Errorf("expected SOURCE after DATA, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropExternalDataSourceStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse data source name
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropExternalFileFormatStatement() (*ast.DropExternalFileFormatStatement, error) {
	// Consume FILE
	p.nextToken()

	// Expect FORMAT
	if strings.ToUpper(p.curTok.Literal) != "FORMAT" {
		return nil, fmt.Errorf("expected FORMAT after FILE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropExternalFileFormatStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse file format name
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropExternalTableStatement() (*ast.DropExternalTableStatement, error) {
	// Consume TABLE
	p.nextToken()

	stmt := &ast.DropExternalTableStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse table names (can be comma-separated)
	for {
		tableName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, tableName)

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

func (p *Parser) parseDropExternalResourcePoolStatement() (*ast.DropExternalResourcePoolStatement, error) {
	// Consume RESOURCE
	p.nextToken()

	// Expect POOL
	if strings.ToUpper(p.curTok.Literal) != "POOL" {
		return nil, fmt.Errorf("expected POOL after RESOURCE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropExternalResourcePoolStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse pool name
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropSecurityPolicyStatement() (*ast.DropSecurityPolicyStatement, error) {
	// Consume SECURITY
	p.nextToken()

	// Expect POLICY
	if strings.ToUpper(p.curTok.Literal) != "POLICY" {
		return nil, fmt.Errorf("expected POLICY after SECURITY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropSecurityPolicyStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse policy names (can be comma-separated)
	for {
		policyName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, policyName)

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

func (p *Parser) parseDropWorkloadStatement() (ast.Statement, error) {
	// Consume WORKLOAD
	p.nextToken()

	switch strings.ToUpper(p.curTok.Literal) {
	case "GROUP":
		return p.parseDropWorkloadGroupStatement()
	case "CLASSIFIER":
		return p.parseDropWorkloadClassifierStatement()
	}

	return nil, fmt.Errorf("expected GROUP or CLASSIFIER after WORKLOAD, got %s", p.curTok.Literal)
}

func (p *Parser) parseDropWorkloadGroupStatement() (*ast.DropWorkloadGroupStatement, error) {
	// Consume GROUP
	p.nextToken()

	stmt := &ast.DropWorkloadGroupStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse group name
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropWorkloadClassifierStatement() (*ast.DropWorkloadClassifierStatement, error) {
	// Consume CLASSIFIER
	p.nextToken()

	stmt := &ast.DropWorkloadClassifierStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse classifier name
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropTypeStatement() (*ast.DropTypeStatement, error) {
	// Consume TYPE
	p.nextToken()

	stmt := &ast.DropTypeStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse type name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropAggregateStatement() (*ast.DropAggregateStatement, error) {
	// Consume AGGREGATE
	p.nextToken()

	stmt := &ast.DropAggregateStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse aggregate names (can be comma-separated)
	for {
		name, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, name)

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

func (p *Parser) parseDropSynonymStatement() (*ast.DropSynonymStatement, error) {
	// Consume SYNONYM
	p.nextToken()

	stmt := &ast.DropSynonymStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse synonym names (can be comma-separated)
	for {
		name, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, name)

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

func (p *Parser) parseDropUserStatement() (*ast.DropUserStatement, error) {
	// Consume USER
	p.nextToken()

	stmt := &ast.DropUserStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse user name
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropRoleStatement() (*ast.DropRoleStatement, error) {
	// Consume ROLE
	p.nextToken()

	stmt := &ast.DropRoleStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse role name
	stmt.Name = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropAssemblyStatement() (*ast.DropAssemblyStatement, error) {
	// Consume ASSEMBLY
	p.nextToken()

	stmt := &ast.DropAssemblyStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse assembly names (can be comma-separated)
	for {
		name, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.Objects = append(stmt.Objects, name)

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

func (p *Parser) parseDropDatabaseStatement() (ast.Statement, error) {
	// Consume DATABASE
	p.nextToken()

	// Check for DATABASE SCOPED CREDENTIAL (look ahead to confirm)
	if p.curTok.Type == TokenScoped && p.peekTok.Type == TokenCredential {
		p.nextToken() // consume SCOPED
		return p.parseDropCredentialStatement(true)
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

		// Parse index name
		indexName := p.parseIdentifier()

		// Check for ON clause (new syntax: index ON table)
		if strings.ToUpper(p.curTok.Literal) == "ON" {
			p.nextToken() // consume ON

			// Parse table name
			tableName, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, err
			}
			clause.IndexName = indexName
			clause.Object = tableName
		} else if p.curTok.Type == TokenDot {
			// Old backwards-compatible syntax: table.index
			p.nextToken() // consume dot
			childName := p.parseIdentifier()
			clause.Index = &ast.SchemaObjectName{
				SchemaIdentifier: indexName,
				BaseIdentifier:   childName,
			}
		} else {
			// Just index name without ON or dot
			clause.IndexName = indexName
		}

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
	case TokenProcedure:
		return p.parseAlterProcedureStatement()
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
	case TokenExternal:
		return p.parseAlterExternalStatement()
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
		case "EXTERNAL":
			return p.parseAlterExternalStatement()
		case "RESOURCE":
			return p.parseAlterResourceGovernorStatement()
		case "CRYPTOGRAPHIC":
			return p.parseAlterCryptographicProviderStatement()
		case "FEDERATION":
			return p.parseAlterFederationStatement()
		case "WORKLOAD":
			return p.parseAlterWorkloadGroupStatement()
		case "SEQUENCE":
			return p.parseAlterSequenceStatement()
		}
		return nil, fmt.Errorf("unexpected token after ALTER: %s", p.curTok.Literal)
	default:
		return nil, fmt.Errorf("unexpected token after ALTER: %s", p.curTok.Literal)
	}
}

func (p *Parser) parseAlterResourceGovernorStatement() (ast.Statement, error) {
	// Consume RESOURCE
	p.nextToken()

	// Consume GOVERNOR
	if strings.ToUpper(p.curTok.Literal) == "GOVERNOR" {
		p.nextToken()
	}

	stmt := &ast.AlterResourceGovernorStatement{}

	switch strings.ToUpper(p.curTok.Literal) {
	case "DISABLE":
		stmt.Command = "Disable"
		p.nextToken()
	case "RECONFIGURE":
		stmt.Command = "Reconfigure"
		p.nextToken()
	case "RESET":
		p.nextToken() // consume RESET
		if strings.ToUpper(p.curTok.Literal) == "STATISTICS" {
			p.nextToken() // consume STATISTICS
		}
		stmt.Command = "ResetStatistics"
	case "WITH":
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
		}
		// Expect CLASSIFIER_FUNCTION = ...
		if strings.ToUpper(p.curTok.Literal) == "CLASSIFIER_FUNCTION" {
			stmt.Command = "ClassifierFunction"
			p.nextToken() // consume CLASSIFIER_FUNCTION
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			// Check for NULL or schema.function
			if p.curTok.Type == TokenNull {
				// ClassifierFunction stays nil
				p.nextToken()
			} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
				stmt.ClassifierFunction, _ = p.parseSchemaObjectName()
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterDatabaseStatement() (ast.Statement, error) {
	// Consume DATABASE
	p.nextToken()

	// Check for SCOPED CREDENTIAL or SCOPED CONFIGURATION
	if p.curTok.Type == TokenScoped {
		p.nextToken() // consume SCOPED
		if p.curTok.Type == TokenCredential {
			return p.parseAlterDatabaseScopedCredentialStatement()
		}
		// Check for CONFIGURATION
		if strings.ToUpper(p.curTok.Literal) == "CONFIGURATION" {
			return p.parseAlterDatabaseScopedConfigurationStatement()
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
		case "DELAYED_DURABILITY":
			// This option uses = with DISABLED/ALLOWED/FORCED values
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after %s, got %s", optionName, p.curTok.Literal)
			}
			p.nextToken()
			optionValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.DelayedDurabilityDatabaseOption{
				OptionKind: "DelayedDurability",
				Value:      capitalizeFirst(optionValue),
			}
			stmt.Options = append(stmt.Options, opt)
		default:
			// Handle generic options with = syntax (e.g., OPTIMIZED_LOCKING = ON)
			if p.curTok.Type == TokenEquals {
				p.nextToken()
				optionValue := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				// Handle parenthesized sub-options
				if p.curTok.Type == TokenLParen {
					p.skipToEndOfStatement()
					return stmt, nil
				}
				opt := &ast.OnOffDatabaseOption{
					OptionKind:  convertOptionKind(optionName),
					OptionState: capitalizeFirst(optionValue),
				}
				stmt.Options = append(stmt.Options, opt)
			} else if p.curTok.Type == TokenOn ||
				strings.ToUpper(p.curTok.Literal) == "ON" || strings.ToUpper(p.curTok.Literal) == "OFF" {
				// Handle options without = (e.g., ENCRYPTION ON)
				optionValue := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				opt := &ast.OnOffDatabaseOption{
					OptionKind:  convertOptionKind(optionName),
					OptionState: capitalizeFirst(optionValue),
				}
				stmt.Options = append(stmt.Options, opt)
			} else {
				// Skip unknown option syntax
				p.skipToEndOfStatement()
				return stmt, nil
			}
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
		// Check for CONTAINS FILESTREAM or CONTAINS MEMORY_OPTIMIZED_DATA
		if strings.ToUpper(p.curTok.Literal) == "CONTAINS" {
			p.nextToken() // consume CONTAINS
			switch strings.ToUpper(p.curTok.Literal) {
			case "FILESTREAM":
				stmt.ContainsFileStream = true
				p.nextToken()
			case "MEMORY_OPTIMIZED_DATA":
				stmt.ContainsMemoryOptimizedData = true
				p.nextToken()
			}
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

	// Check for Azure-style MODIFY (options) syntax
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		createOpts, err := p.parseAzureDatabaseOptions()
		if err != nil {
			return nil, err
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
		// Convert CreateDatabaseOption to DatabaseOption
		opts := make([]ast.DatabaseOption, len(createOpts))
		for i, o := range createOpts {
			opts[i] = o.(ast.DatabaseOption)
		}
		stmt := &ast.AlterDatabaseSetStatement{
			DatabaseName: dbName,
			Options:      opts,
		}
		p.skipToEndOfStatement()
		return stmt, nil
	}

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
		// Parse optional modifiers
		for {
			switch strings.ToUpper(p.curTok.Literal) {
			case "DEFAULT":
				stmt.MakeDefault = true
				p.nextToken()
			case "READONLY":
				stmt.UpdatabilityOption = "ReadOnlyOld"
				p.nextToken()
			case "READ_ONLY":
				stmt.UpdatabilityOption = "ReadOnly"
				p.nextToken()
			case "READWRITE":
				stmt.UpdatabilityOption = "ReadWriteOld"
				p.nextToken()
			case "READ_WRITE":
				stmt.UpdatabilityOption = "ReadWrite"
				p.nextToken()
			case "AUTOGROW_ALL_FILES":
				stmt.UpdatabilityOption = "AutogrowAllFiles"
				p.nextToken()
			case "AUTOGROW_SINGLE_FILE":
				stmt.UpdatabilityOption = "AutogrowSingleFile"
				p.nextToken()
			case "NAME":
				p.nextToken() // consume NAME
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				stmt.NewFileGroupName = p.parseIdentifier()
			default:
				p.skipToEndOfStatement()
				return stmt, nil
			}
		}
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

func (p *Parser) parseAlterDatabaseScopedConfigurationStatement() (ast.Statement, error) {
	// Consume CONFIGURATION
	p.nextToken()

	stmt := &ast.AlterDatabaseScopedConfigurationClearStatement{}

	// Check for FOR SECONDARY
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "SECONDARY" {
			stmt.Secondary = true
			p.nextToken() // consume SECONDARY
		}
	}

	// Check for CLEAR
	if strings.ToUpper(p.curTok.Literal) == "CLEAR" {
		p.nextToken() // consume CLEAR

		// Parse option (PROCEDURE_CACHE)
		optionKind := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		option := &ast.DatabaseConfigurationClearOption{}
		if optionKind == "PROCEDURE_CACHE" {
			option.OptionKind = "ProcedureCache"
		} else {
			option.OptionKind = optionKind
		}

		// Check for optional plan handle (binary literal)
		if p.curTok.Type == TokenBinary {
			option.PlanHandle = &ast.BinaryLiteral{
				LiteralType: "Binary",
				Value:       p.curTok.Literal,
			}
			p.nextToken()
		}

		stmt.Option = option
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterServerConfigurationStatement() (ast.Statement, error) {
	// Consume SERVER
	p.nextToken()

	// Check if it's ALTER SERVER ROLE, ALTER SERVER AUDIT, or ALTER SERVER CONFIGURATION
	switch strings.ToUpper(p.curTok.Literal) {
	case "ROLE":
		return p.parseAlterServerRoleStatement()
	case "AUDIT":
		return p.parseAlterServerAuditStatement()
	}

	// Expect CONFIGURATION
	if strings.ToUpper(p.curTok.Literal) != "CONFIGURATION" {
		return nil, fmt.Errorf("expected CONFIGURATION, ROLE, or AUDIT after SERVER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect SET
	if p.curTok.Type != TokenSet {
		return nil, fmt.Errorf("expected SET after CONFIGURATION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check what type of SET it is
	switch strings.ToUpper(p.curTok.Literal) {
	case "SOFTNUMA":
		return p.parseAlterServerConfigurationSetSoftNumaStatement()
	case "PROCESS":
		return p.parseAlterServerConfigurationSetProcessAffinityStatement()
	case "EXTERNAL":
		return p.parseAlterServerConfigurationSetExternalAuthenticationStatement()
	default:
		return nil, fmt.Errorf("unexpected token after SET: %s", p.curTok.Literal)
	}
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

func (p *Parser) parseAlterServerConfigurationSetExternalAuthenticationStatement() (*ast.AlterServerConfigurationSetExternalAuthenticationStatement, error) {
	// Consume EXTERNAL
	p.nextToken()

	// Expect AUTHENTICATION
	if strings.ToUpper(p.curTok.Literal) != "AUTHENTICATION" {
		return nil, fmt.Errorf("expected AUTHENTICATION after EXTERNAL, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterServerConfigurationSetExternalAuthenticationStatement{}

	// Parse ON or OFF
	optionState := strings.ToUpper(p.curTok.Literal)
	if optionState != "ON" && optionState != "OFF" {
		return nil, fmt.Errorf("expected ON or OFF after AUTHENTICATION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	containerOption := &ast.AlterServerConfigurationExternalAuthenticationContainerOption{
		OptionKind: "OnOff",
		OptionValue: &ast.OnOffOptionValue{
			OptionState: capitalizeFirst(optionState),
		},
	}

	// Check for suboptions in parentheses (only for ON)
	if optionState == "ON" && p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		// Parse suboptions
		for {
			suboption := &ast.AlterServerConfigurationExternalAuthenticationOption{}

			optionName := strings.ToUpper(p.curTok.Literal)
			switch optionName {
			case "USE_IDENTITY":
				suboption.OptionKind = "UseIdentity"
				p.nextToken()
			case "CREDENTIAL_NAME":
				suboption.OptionKind = "CredentialName"
				p.nextToken()

				// Expect =
				if p.curTok.Type != TokenEquals {
					return nil, fmt.Errorf("expected = after CREDENTIAL_NAME, got %s", p.curTok.Literal)
				}
				p.nextToken()

				// Parse string literal
				if p.curTok.Type != TokenString {
					return nil, fmt.Errorf("expected string literal for CREDENTIAL_NAME value, got %s", p.curTok.Literal)
				}
				strLit, err := p.parseStringLiteral()
				if err != nil {
					return nil, err
				}
				suboption.OptionValue = &ast.LiteralOptionValue{
					Value: strLit,
				}
			default:
				return nil, fmt.Errorf("unexpected option in EXTERNAL AUTHENTICATION: %s", p.curTok.Literal)
			}

			containerOption.Suboptions = append(containerOption.Suboptions, suboption)

			// Check for comma or closing paren
			if p.curTok.Type == TokenComma {
				p.nextToken()
				continue
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken()
				break
			}
			return nil, fmt.Errorf("expected , or ) in EXTERNAL AUTHENTICATION options, got %s", p.curTok.Literal)
		}
	}

	stmt.Options = append(stmt.Options, containerOption)

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterServerConfigurationSetProcessAffinityStatement() (*ast.AlterServerConfigurationStatement, error) {
	// Consume PROCESS
	p.nextToken()

	// Expect AFFINITY
	if strings.ToUpper(p.curTok.Literal) != "AFFINITY" {
		return nil, fmt.Errorf("expected AFFINITY after PROCESS, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterServerConfigurationStatement{}

	// Parse CPU or NUMANODE
	affinityType := strings.ToUpper(p.curTok.Literal)
	switch affinityType {
	case "CPU":
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			// Check for AUTO
			if strings.ToUpper(p.curTok.Literal) == "AUTO" {
				stmt.ProcessAffinity = "CpuAuto"
				p.nextToken()
			} else {
				// Parse ranges
				stmt.ProcessAffinity = "Cpu"
				ranges, err := p.parseProcessAffinityRanges()
				if err != nil {
					return nil, err
				}
				stmt.ProcessAffinityRanges = ranges
			}
		}
	case "NUMANODE":
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			stmt.ProcessAffinity = "NumaNode"
			ranges, err := p.parseProcessAffinityRanges()
			if err != nil {
				return nil, err
			}
			stmt.ProcessAffinityRanges = ranges
		}
	default:
		return nil, fmt.Errorf("expected CPU or NUMANODE after AFFINITY, got %s", p.curTok.Literal)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseProcessAffinityRanges() ([]*ast.ProcessAffinityRange, error) {
	var ranges []*ast.ProcessAffinityRange

	for {
		r := &ast.ProcessAffinityRange{}

		// Parse From value
		if p.curTok.Type != TokenNumber {
			return nil, fmt.Errorf("expected number in process affinity range, got %s", p.curTok.Literal)
		}
		r.From = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()

		// Check for TO
		if strings.ToUpper(p.curTok.Literal) == "TO" {
			p.nextToken()
			if p.curTok.Type != TokenNumber {
				return nil, fmt.Errorf("expected number after TO, got %s", p.curTok.Literal)
			}
			r.To = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
			p.nextToken()
		}

		ranges = append(ranges, r)

		// Check for comma
		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	return ranges, nil
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

	// Check for SET
	if strings.ToUpper(p.curTok.Literal) == "SET" {
		return p.parseAlterTableSetStatement(tableName)
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
	// Format: DROP [COLUMN] name [WITH (options)], [CONSTRAINT] name [WITH (options)], ...
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

		// Check for WITH clause
		if p.curTok.Type == TokenWith {
			options, err := p.parseDropClusteredConstraintOptions()
			if err != nil {
				return nil, err
			}
			element.DropClusteredConstraintOptions = options
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

func (p *Parser) parseDropClusteredConstraintOptions() ([]ast.DropClusteredConstraintOption, error) {
	// Consume WITH
	p.nextToken()

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	var options []ast.DropClusteredConstraintOption

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		optionName := strings.ToUpper(p.curTok.Literal)

		switch optionName {
		case "ONLINE":
			p.nextToken() // consume ONLINE
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after ONLINE, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			state := strings.ToUpper(p.curTok.Literal)
			var optionState string
			if state == "ON" {
				optionState = "On"
			} else if state == "OFF" {
				optionState = "Off"
			} else {
				return nil, fmt.Errorf("expected ON or OFF after ONLINE =, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume ON/OFF
			options = append(options, &ast.DropClusteredConstraintStateOption{
				OptionKind:  "Online",
				OptionState: optionState,
			})

		case "MOVE":
			p.nextToken() // consume MOVE
			if strings.ToUpper(p.curTok.Literal) != "TO" {
				return nil, fmt.Errorf("expected TO after MOVE, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume TO

			fg, err := p.parseFileGroupOrPartitionScheme()
			if err != nil {
				return nil, err
			}
			options = append(options, &ast.DropClusteredConstraintMoveOption{
				OptionKind:  "MoveTo",
				OptionValue: fg,
			})

		case "MAXDOP":
			p.nextToken() // consume MAXDOP
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after MAXDOP, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			if p.curTok.Type != TokenNumber {
				return nil, fmt.Errorf("expected number after MAXDOP =, got %s", p.curTok.Literal)
			}
			options = append(options, &ast.DropClusteredConstraintValueOption{
				OptionKind: "MaxDop",
				OptionValue: &ast.IntegerLiteral{
					LiteralType: "Integer",
					Value:       p.curTok.Literal,
				},
			})
			p.nextToken() // consume number

		default:
			return nil, fmt.Errorf("unexpected option in DROP WITH clause: %s", p.curTok.Literal)
		}

		// Check for comma or end of options
		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) to close WITH options, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return options, nil
}

func (p *Parser) parseFileGroupOrPartitionScheme() (*ast.FileGroupOrPartitionScheme, error) {
	fg := &ast.FileGroupOrPartitionScheme{}

	// Parse filegroup/partition scheme name (can be identifier or string literal)
	iove, err := p.parseIdentifierOrValueExpression()
	if err != nil {
		return nil, err
	}
	fg.Name = iove

	// Check for partition scheme columns (column1, column2, ...)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			if p.curTok.Type != TokenIdent && p.curTok.Type != TokenLBracket {
				return nil, fmt.Errorf("expected column identifier in partition scheme, got %s", p.curTok.Literal)
			}
			fg.PartitionSchemeColumns = append(fg.PartitionSchemeColumns, p.parseIdentifier())

			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma
			}
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) to close partition scheme columns, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	}

	return fg, nil
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

func (p *Parser) parseAlterTableSetStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableSetStatement, error) {
	stmt := &ast.AlterTableSetStatement{
		SchemaObjectName: tableName,
	}

	// Consume SET
	p.nextToken()

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after SET, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse options
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		if optionName == "SYSTEM_VERSIONING" {
			opt, err := p.parseSystemVersioningTableOption()
			if err != nil {
				return nil, err
			}
			stmt.Options = append(stmt.Options, opt)
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	// Consume )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseSystemVersioningTableOption() (*ast.SystemVersioningTableOption, error) {
	opt := &ast.SystemVersioningTableOption{
		OptionKind:              "LockEscalation",
		ConsistencyCheckEnabled: "NotSet",
	}

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after SYSTEM_VERSIONING, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse ON or OFF
	stateVal := strings.ToUpper(p.curTok.Literal)
	if stateVal == "ON" {
		opt.OptionState = "On"
	} else if stateVal == "OFF" {
		opt.OptionState = "Off"
	} else {
		return nil, fmt.Errorf("expected ON or OFF after =, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for optional sub-options in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken()

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			subOptName := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}

			switch subOptName {
			case "HISTORY_TABLE":
				histTable, err := p.parseSchemaObjectName()
				if err != nil {
					return nil, err
				}
				opt.HistoryTable = histTable

			case "DATA_CONSISTENCY_CHECK":
				checkVal := strings.ToUpper(p.curTok.Literal)
				if checkVal == "ON" {
					opt.ConsistencyCheckEnabled = "On"
				} else if checkVal == "OFF" {
					opt.ConsistencyCheckEnabled = "Off"
				}
				p.nextToken()

			case "HISTORY_RETENTION_PERIOD":
				retPeriod, err := p.parseRetentionPeriodDefinition()
				if err != nil {
					return nil, err
				}
				opt.RetentionPeriod = retPeriod
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}

		// Consume )
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	return opt, nil
}

func (p *Parser) parseRetentionPeriodDefinition() (*ast.RetentionPeriodDefinition, error) {
	ret := &ast.RetentionPeriodDefinition{}

	// Check for INFINITE
	if strings.ToUpper(p.curTok.Literal) == "INFINITE" {
		ret.IsInfinity = true
		ret.Units = "Day" // Default unit for INFINITE
		p.nextToken()
		return ret, nil
	}

	// Parse numeric duration
	ret.IsInfinity = false

	// Parse integer literal
	if p.curTok.Type == TokenNumber {
		lit := &ast.IntegerLiteral{
			LiteralType: "Integer",
			Value:       p.curTok.Literal,
		}
		ret.Duration = lit
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected number for retention period, got %s", p.curTok.Literal)
	}

	// Parse unit
	unitVal := strings.ToUpper(p.curTok.Literal)
	switch unitVal {
	case "DAY", "DAYS":
		ret.Units = "Day"
	case "WEEK", "WEEKS":
		ret.Units = "Week"
	case "MONTH":
		ret.Units = "Month"
	case "MONTHS":
		ret.Units = "Months"
	case "YEAR", "YEARS":
		ret.Units = "Year"
	default:
		return nil, fmt.Errorf("unexpected unit %s for retention period", unitVal)
	}
	p.nextToken()

	return ret, nil
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

func (p *Parser) parseAlterServerRoleStatement() (*ast.AlterServerRoleStatement, error) {
	// Consume ROLE
	p.nextToken()

	stmt := &ast.AlterServerRoleStatement{}

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

func (p *Parser) parseAlterServerAuditStatement() (*ast.AlterServerAuditStatement, error) {
	// AUDIT keyword should be current token, consume it
	p.nextToken()

	stmt := &ast.AlterServerAuditStatement{}

	// Parse audit name
	stmt.AuditName = p.parseIdentifier()

	// Check for REMOVE WHERE
	if strings.ToUpper(p.curTok.Literal) == "REMOVE" {
		p.nextToken() // consume REMOVE
		if strings.ToUpper(p.curTok.Literal) == "WHERE" {
			p.nextToken() // consume WHERE
			stmt.RemoveWhere = true
			// Skip optional semicolon
			if p.curTok.Type == TokenSemicolon {
				p.nextToken()
			}
			return stmt, nil
		}
		return nil, fmt.Errorf("expected WHERE after REMOVE, got %s", p.curTok.Literal)
	}

	// Parse TO clause (audit target)
	if strings.ToUpper(p.curTok.Literal) == "TO" {
		p.nextToken() // consume TO
		target, err := p.parseAuditTarget()
		if err != nil {
			return nil, err
		}
		stmt.AuditTarget = target
	}

	// Parse WITH clause (options)
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				opt, err := p.parseAuditOption()
				if err != nil {
					return nil, err
				}
				stmt.Options = append(stmt.Options, opt)
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

	// Parse WHERE clause (predicate)
	if strings.ToUpper(p.curTok.Literal) == "WHERE" {
		p.nextToken() // consume WHERE
		pred, err := p.parseAuditPredicate()
		if err != nil {
			return nil, err
		}
		stmt.PredicateExpression = pred
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

func (p *Parser) parseAlterServiceStatement() (ast.Statement, error) {
	// Consume SERVICE
	p.nextToken()

	// Check for SERVICE MASTER KEY <action>
	// Only treat as AlterServiceMasterKeyStatement if followed by REGENERATE, FORCE, or WITH
	if strings.ToUpper(p.curTok.Literal) == "MASTER" && strings.ToUpper(p.peekTok.Literal) == "KEY" {
		// Peek ahead to see if there's an action keyword
		// Save current position info
		curLit := p.curTok.Literal
		p.nextToken() // consume MASTER
		p.nextToken() // consume KEY

		nextKeyword := strings.ToUpper(p.curTok.Literal)
		if nextKeyword == "REGENERATE" || nextKeyword == "FORCE" || nextKeyword == "WITH" {
			return p.parseAlterServiceMasterKeyStatementBody()
		}

		// Not a valid SERVICE MASTER KEY statement - treat "master" as service name
		// KEY and following tokens will be skipped by skipToEndOfStatement
		stmt := &ast.AlterServiceStatement{}
		stmt.Name = &ast.Identifier{QuoteType: "NotQuoted", Value: curLit}
		p.skipToEndOfStatement()
		return stmt, nil
	}

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
	stmt.ActiveForBeginDialog = "NotSet"

	// Parse certificate name
	stmt.Name = p.parseIdentifier()

	// Check what kind of ALTER CERTIFICATE this is
	lit := strings.ToUpper(p.curTok.Literal)
	if lit == "REMOVE" {
		p.nextToken() // consume REMOVE
		nextLit := strings.ToUpper(p.curTok.Literal)
		if nextLit == "PRIVATE" {
			p.nextToken() // consume PRIVATE
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			stmt.Kind = "RemovePrivateKey"
		} else if nextLit == "ATTESTED" {
			p.nextToken() // consume ATTESTED
			if strings.ToUpper(p.curTok.Literal) == "OPTION" {
				p.nextToken() // consume OPTION
			}
			stmt.Kind = "RemoveAttestedOption"
		}
	} else if lit == "ATTESTED" {
		p.nextToken() // consume ATTESTED
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		stmt.Kind = "AttestedBy"
		if p.curTok.Type == TokenString {
			strLit, _ := p.parseStringLiteral()
			stmt.AttestedBy = strLit
		}
	} else if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		withLit := strings.ToUpper(p.curTok.Literal)
		if withLit == "ACTIVE" {
			p.nextToken() // consume ACTIVE
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
			}
			if strings.ToUpper(p.curTok.Literal) == "BEGIN_DIALOG" {
				p.nextToken() // consume BEGIN_DIALOG
			}
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			stmt.Kind = "WithActiveForBeginDialog"
			if p.curTok.Type == TokenOn {
				stmt.ActiveForBeginDialog = "On"
				p.nextToken()
			} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OFF" {
				stmt.ActiveForBeginDialog = "Off"
				p.nextToken()
			}
		} else if withLit == "PRIVATE" {
			p.nextToken() // consume PRIVATE
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			stmt.Kind = "WithPrivateKey"
			// Parse private key options (FILE = '...', DECRYPTION BY PASSWORD = '...', etc.)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optLit := strings.ToUpper(p.curTok.Literal)
					if optLit == "FILE" {
						p.nextToken() // consume FILE
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if p.curTok.Type == TokenString {
							strLit, _ := p.parseStringLiteral()
							stmt.PrivateKeyPath = strLit
						}
					} else if optLit == "DECRYPTION" {
						p.nextToken() // consume DECRYPTION
						if strings.ToUpper(p.curTok.Literal) == "BY" {
							p.nextToken() // consume BY
						}
						if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
							p.nextToken() // consume PASSWORD
						}
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if p.curTok.Type == TokenString {
							strLit, _ := p.parseStringLiteral()
							stmt.DecryptionPassword = strLit
						}
					} else if optLit == "ENCRYPTION" {
						p.nextToken() // consume ENCRYPTION
						if strings.ToUpper(p.curTok.Literal) == "BY" {
							p.nextToken() // consume BY
						}
						if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
							p.nextToken() // consume PASSWORD
						}
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if p.curTok.Type == TokenString {
							strLit, _ := p.parseStringLiteral()
							stmt.EncryptionPassword = strLit
						}
					} else {
						p.nextToken() // skip unknown option
					}
					if p.curTok.Type == TokenComma {
						p.nextToken() // consume comma
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

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

	// Optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken()
		opts, err := p.parseApplicationRoleOptions()
		if err != nil {
			return nil, err
		}
		stmt.ApplicationRoleOptions = opts
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

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

	// Check for WITH clause
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		opts, err := p.parseQueueOptions()
		if err != nil {
			return nil, err
		}
		stmt.QueueOptions = opts
	}

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
		// Parse NEXT USED [filegroup]
		if strings.ToUpper(p.curTok.Literal) == "NEXT" {
			p.nextToken() // consume NEXT
			if strings.ToUpper(p.curTok.Literal) == "USED" {
				p.nextToken() // consume USED
			}
			// Check for optional filegroup name (identifier or string)
			if p.curTok.Type == TokenIdent || p.curTok.Type == TokenString {
				stmt.FileGroup = &ast.IdentifierOrValueExpression{}
				if p.curTok.Type == TokenString {
					strLit, err := p.parseStringLiteral()
					if err == nil {
						stmt.FileGroup.Value = strLit.Value
						stmt.FileGroup.ValueExpression = strLit
					}
				} else {
					ident := p.parseIdentifier()
					stmt.FileGroup.Value = ident.Value
					stmt.FileGroup.Identifier = ident
				}
			}
		}
		p.skipToEndOfStatement()
		return stmt, nil
	}

	stmt := &ast.AlterPartitionFunctionStatement{}
	stmt.Name = p.parseIdentifier()
	// Consume ()
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}
	// Parse SPLIT or MERGE
	action := strings.ToUpper(p.curTok.Literal)
	if action == "SPLIT" {
		stmt.HasAction = true
		stmt.IsSplit = true
		p.nextToken()
	} else if action == "MERGE" {
		stmt.HasAction = true
		stmt.IsSplit = false
		p.nextToken()
	}
	// Parse optional RANGE (value)
	if strings.ToUpper(p.curTok.Literal) == "RANGE" {
		p.nextToken() // consume RANGE
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			boundary, err := p.parseScalarExpression()
			if err == nil {
				stmt.Boundary = boundary
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}
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

		// Parse action: REBUILD, REORGANIZE, AS DEFAULT
		actionLit := strings.ToUpper(p.curTok.Literal)
		if actionLit == "REBUILD" {
			stmt.Action = "Rebuild"
			p.nextToken()
			// Check for WITH ACCENT_SENSITIVITY = ON/OFF
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				if strings.ToUpper(p.curTok.Literal) == "ACCENT_SENSITIVITY" {
					p.nextToken() // consume ACCENT_SENSITIVITY
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					opt := &ast.OnOffFullTextCatalogOption{
						OptionKind: "AccentSensitivity",
					}
					if p.curTok.Type == TokenOn {
						opt.OptionState = "On"
						p.nextToken()
					} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OFF" {
						opt.OptionState = "Off"
						p.nextToken()
					}
					stmt.Options = append(stmt.Options, opt)
				}
			}
		} else if actionLit == "REORGANIZE" {
			stmt.Action = "Reorganize"
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "AS" {
			p.nextToken() // consume AS
			if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
				p.nextToken() // consume DEFAULT
			}
			stmt.Action = "AsDefault"
		}

		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
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

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		for {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			if p.curTok.Type != TokenEquals {
				break
			}
			p.nextToken() // consume =

			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			switch optName {
			case "IDENTITY":
				stmt.Identity = val
			case "SECRET":
				stmt.Secret = val
			}

			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume ,
		}
	}

	// Skip rest of statement
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterServiceMasterKeyStatement() (*ast.AlterServiceMasterKeyStatement, error) {
	// SERVICE_MASTER_KEY was matched as an identifier
	p.nextToken() // consume SERVICE_MASTER_KEY
	return p.parseAlterServiceMasterKeyStatementBody()
}

func (p *Parser) parseAlterServiceMasterKeyStatementBody() (*ast.AlterServiceMasterKeyStatement, error) {
	stmt := &ast.AlterServiceMasterKeyStatement{}

	// Parse the kind: REGENERATE, FORCE REGENERATE, WITH OLD_ACCOUNT/NEW_ACCOUNT
	switch strings.ToUpper(p.curTok.Literal) {
	case "REGENERATE":
		stmt.Kind = "Regenerate"
		p.nextToken()
	case "FORCE":
		p.nextToken() // consume FORCE
		if strings.ToUpper(p.curTok.Literal) == "REGENERATE" {
			stmt.Kind = "ForceRegenerate"
			p.nextToken()
		}
	case "WITH":
		p.nextToken() // consume WITH
		// Parse OLD_ACCOUNT or NEW_ACCOUNT
		switch strings.ToUpper(p.curTok.Literal) {
		case "OLD_ACCOUNT":
			stmt.Kind = "WithOldAccount"
			p.nextToken() // consume OLD_ACCOUNT
		case "NEW_ACCOUNT":
			stmt.Kind = "WithNewAccount"
			p.nextToken() // consume NEW_ACCOUNT
		}
		// Parse = 'account'
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		if p.curTok.Type == TokenString {
			account, err := p.parseStringLiteral()
			if err == nil {
				stmt.Account = account
			}
		}
		// Parse , OLD_PASSWORD/NEW_PASSWORD = 'password'
		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
		// Skip OLD_PASSWORD or NEW_PASSWORD
		if strings.ToUpper(p.curTok.Literal) == "OLD_PASSWORD" || strings.ToUpper(p.curTok.Literal) == "NEW_PASSWORD" {
			p.nextToken()
		}
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		if p.curTok.Type == TokenString {
			password, err := p.parseStringLiteral()
			if err == nil {
				stmt.Password = password
			}
		}
	}

	// Skip to end of statement
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

func (p *Parser) parseAlterProcedureStatement() (*ast.AlterProcedureStatement, error) {
	// Consume PROCEDURE/PROC
	p.nextToken()

	stmt := &ast.AlterProcedureStatement{}
	stmt.ProcedureReference = &ast.ProcedureReference{}

	// Parse procedure name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.ProcedureReference.Name = name

	// Parse optional parameters
	if p.curTok.Type == TokenLParen || (p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@")) {
		params, err := p.parseProcedureParameters()
		if err != nil {
			return nil, err
		}
		stmt.Parameters = params
	}

	// Skip WITH options (like RECOMPILE, ENCRYPTION, etc.)
	if p.curTok.Type == TokenWith {
		p.nextToken()
		for {
			if strings.ToUpper(p.curTok.Literal) == "FOR" || p.curTok.Type == TokenAs || p.curTok.Type == TokenEOF {
				break
			}
			if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
				stmt.IsForReplication = true
			}
			p.nextToken()
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
	}

	// Expect AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse statement list
	stmts, err := p.parseStatementList()
	if err != nil {
		return nil, err
	}
	stmt.StatementList = stmts

	return stmt, nil
}

func (p *Parser) parseAlterExternalStatement() (ast.Statement, error) {
	// Consume EXTERNAL
	p.nextToken()

	// Check what type of external statement
	upper := strings.ToUpper(p.curTok.Literal)
	switch upper {
	case "DATA":
		return p.parseAlterExternalDataSourceStatement()
	case "LANGUAGE":
		return p.parseAlterExternalLanguageStatement()
	case "LIBRARY":
		return p.parseAlterExternalLibraryStatement()
	default:
		// Skip to end of statement for unsupported external statements
		p.skipToEndOfStatement()
		return &ast.AlterExternalDataSourceStatement{}, nil
	}
}

func (p *Parser) parseAlterExternalDataSourceStatement() (*ast.AlterExternalDataSourceStatement, error) {
	// Consume DATA
	p.nextToken()

	// Expect SOURCE
	if strings.ToUpper(p.curTok.Literal) != "SOURCE" {
		p.skipToEndOfStatement()
		return &ast.AlterExternalDataSourceStatement{}, nil
	}
	p.nextToken()

	stmt := &ast.AlterExternalDataSourceStatement{}

	// Parse name
	stmt.Name = p.parseIdentifier()

	// Expect SET
	if p.curTok.Type == TokenSet {
		p.nextToken()
	}

	// Parse options
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		if p.isStatementKeyword(p.curTok.Type) {
			break
		}
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			break
		}

		optName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		// Expect =
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}

		opt := &ast.ExternalDataSourceLiteralOrIdentifierOption{
			OptionKind: externalDataSourceOptionKindToPascalCase(optName),
			Value:      &ast.IdentifierOrValueExpression{},
		}

		// Parse value
		if p.curTok.Type == TokenString {
			strLit, _ := p.parseStringLiteral()
			opt.Value.Value = strLit.Value
			opt.Value.ValueExpression = strLit
		} else if p.curTok.Type == TokenIdent {
			ident := p.parseIdentifier()
			opt.Value.Value = ident.Value
			opt.Value.Identifier = ident
		} else {
			p.nextToken()
		}

		stmt.ExternalDataSourceOptions = append(stmt.ExternalDataSourceOptions, opt)

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

func (p *Parser) parseAlterExternalLanguageStatement() (*ast.AlterExternalLanguageStatement, error) {
	// Consume LANGUAGE
	p.nextToken()

	stmt := &ast.AlterExternalLanguageStatement{}

	// Parse name
	stmt.Name = p.parseIdentifier()

	// Skip rest of statement
	p.skipToEndOfStatement()

	return stmt, nil
}

func (p *Parser) parseAlterExternalLibraryStatement() (*ast.AlterExternalLibraryStatement, error) {
	// Consume LIBRARY
	p.nextToken()

	stmt := &ast.AlterExternalLibraryStatement{}

	// Parse name
	stmt.Name = p.parseIdentifier()

	// Parse optional AUTHORIZATION clause
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Parse SET clause
	if strings.ToUpper(p.curTok.Literal) == "SET" {
		p.nextToken() // consume SET
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optName := strings.ToUpper(p.curTok.Literal)
				p.nextToken() // consume option name

				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =

					if optName == "CONTENT" {
						content, err := p.parseScalarExpression()
						if err != nil {
							return nil, err
						}
						stmt.ExternalLibraryFiles = append(stmt.ExternalLibraryFiles, &ast.ExternalLibraryFileOption{
							Content: content,
						})
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

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optName := strings.ToUpper(p.curTok.Literal)
				p.nextToken() // consume option name

				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =

					if optName == "LANGUAGE" && p.curTok.Type == TokenString {
						strLit, _ := p.parseStringLiteral()
						stmt.Language = strLit
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

// convertOptionKind converts a SQL option name (e.g., "OPTIMIZED_LOCKING") to its OptionKind form (e.g., "OptimizedLocking")
func convertOptionKind(optionName string) string {
	// Split by underscores and capitalize each word
	parts := strings.Split(optionName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// parseAlterWorkloadGroupStatement parses ALTER WORKLOAD GROUP statement.
func (p *Parser) parseAlterWorkloadGroupStatement() (*ast.AlterWorkloadGroupStatement, error) {
	// Consume WORKLOAD
	p.nextToken()

	// Consume GROUP
	if strings.ToUpper(p.curTok.Literal) == "GROUP" {
		p.nextToken()
	}

	stmt := &ast.AlterWorkloadGroupStatement{}

	// Parse group name
	stmt.Name = p.parseIdentifier()

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
		}

		// Parse parameters
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			param, err := p.parseWorkloadGroupParameter()
			if err != nil {
				return nil, err
			}
			stmt.WorkloadGroupParameters = append(stmt.WorkloadGroupParameters, param)
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

	// Parse USING clause (resource pool reference)
	if strings.ToUpper(p.curTok.Literal) == "USING" {
		p.nextToken() // consume USING
		stmt.PoolName = p.parseIdentifier()

		// Check for EXTERNAL
		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
		if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
			p.nextToken() // consume EXTERNAL
			stmt.ExternalPoolName = p.parseIdentifier()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseAlterSequenceStatement parses ALTER SEQUENCE statement.
func (p *Parser) parseAlterSequenceStatement() (*ast.AlterSequenceStatement, error) {
	// Consume SEQUENCE
	p.nextToken()

	stmt := &ast.AlterSequenceStatement{}

	// Parse sequence name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Parse sequence options
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && !p.isStatementKeyword(p.curTok.Type) {
		option, err := p.parseSequenceOption()
		if err != nil {
			break
		}
		if option == nil {
			break // Unrecognized option, stop parsing
		}
		stmt.SequenceOptions = append(stmt.SequenceOptions, option)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCreateSequenceStatement parses CREATE SEQUENCE statement.
func (p *Parser) parseCreateSequenceStatement() (*ast.CreateSequenceStatement, error) {
	// Consume SEQUENCE
	p.nextToken()

	stmt := &ast.CreateSequenceStatement{}

	// Parse sequence name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Parse sequence options
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && !p.isStatementKeyword(p.curTok.Type) {
		option, err := p.parseSequenceOption()
		if err != nil {
			break
		}
		if option == nil {
			break // Unrecognized option, stop parsing
		}
		stmt.SequenceOptions = append(stmt.SequenceOptions, option)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseSequenceOption parses a single sequence option.
func (p *Parser) parseSequenceOption() (interface{}, error) {
	optionName := strings.ToUpper(p.curTok.Literal)

	// Check for NO prefix
	isNo := false
	if optionName == "NO" {
		isNo = true
		p.nextToken()
		optionName = strings.ToUpper(p.curTok.Literal)
	}

	var optionKind string
	hasValue := true

	switch optionName {
	case "RESTART":
		optionKind = "Restart"
		p.nextToken()
		// Check for WITH value
		if strings.ToUpper(p.curTok.Literal) == "WITH" {
			p.nextToken()
			hasValue = true
		} else {
			hasValue = false
		}
	case "INCREMENT":
		optionKind = "Increment"
		p.nextToken()
		// Consume BY
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken()
		}
	case "MINVALUE":
		optionKind = "MinValue"
		p.nextToken()
		if isNo {
			hasValue = false
		}
	case "MAXVALUE":
		optionKind = "MaxValue"
		p.nextToken()
		if isNo {
			hasValue = false
		}
	case "CYCLE":
		optionKind = "Cycle"
		p.nextToken()
		// CYCLE is always a SequenceOption (not ScalarExpressionSequenceOption)
		return &ast.SequenceOption{
			OptionKind: optionKind,
			NoValue:    isNo,
		}, nil
	case "CACHE":
		optionKind = "Cache"
		p.nextToken()
		if isNo {
			hasValue = false
		} else {
			// Check if there's a numeric value following
			if p.curTok.Type == TokenNumber {
				// Has value
			} else {
				hasValue = false
			}
		}
	case "START":
		optionKind = "Start"
		p.nextToken()
		// Consume WITH
		if strings.ToUpper(p.curTok.Literal) == "WITH" {
			p.nextToken()
		}
	case "AS":
		p.nextToken()
		// Parse data type - use parseDataTypeReference to preserve UserDataTypeReference
		dataType, err := p.parseDataTypeReference()
		if err != nil {
			return nil, err
		}
		return &ast.DataTypeSequenceOption{
			OptionKind: "As",
			DataType:   dataType,
			NoValue:    false,
		}, nil
	default:
		return nil, nil
	}

	if isNo {
		// NO prefix means NoValue = true
		return &ast.SequenceOption{
			OptionKind: optionKind,
			NoValue:    true,
		}, nil
	}

	if !hasValue {
		return &ast.ScalarExpressionSequenceOption{
			OptionKind: optionKind,
			NoValue:    false,
		}, nil
	}

	// Parse the value
	val, err := p.parseScalarExpression()
	if err != nil {
		return &ast.ScalarExpressionSequenceOption{
			OptionKind: optionKind,
			NoValue:    false,
		}, nil
	}

	return &ast.ScalarExpressionSequenceOption{
		OptionKind:  optionKind,
		OptionValue: val,
		NoValue:     false,
	}, nil
}

