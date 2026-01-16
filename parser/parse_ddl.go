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
	case "ASYMMETRIC":
		return p.parseDropAsymmetricKeyStatement()
	case "SYMMETRIC":
		return p.parseDropSymmetricKeyStatement()
	case "SIGNATURE":
		return p.parseDropSignatureStatement(false)
	case "COUNTER":
		p.nextToken() // consume COUNTER
		if strings.ToUpper(p.curTok.Literal) != "SIGNATURE" {
			return nil, fmt.Errorf("expected SIGNATURE after COUNTER, got %s", p.curTok.Literal)
		}
		return p.parseDropSignatureStatement(true)
	case "SENSITIVITY":
		return p.parseDropSensitivityClassificationStatement()
	case "FULLTEXT":
		return p.parseDropFulltextStatement()
	case "BROKER":
		return p.parseDropBrokerPriorityStatement()
	case "RESOURCE":
		return p.parseDropResourcePoolStatement()
	case "LOGIN":
		return p.parseDropLoginStatement()
	case "PARTITION":
		return p.parseDropPartitionStatement()
	case "APPLICATION":
		return p.parseDropApplicationRoleStatement()
	case "CERTIFICATE":
		return p.parseDropCertificateStatement()
	case "CREDENTIAL":
		return p.parseDropCredentialStatement(false)
	case "MASTER":
		return p.parseDropMasterKeyStatement()
	case "XML":
		return p.parseDropXmlSchemaCollectionStatement()
	case "CONTRACT":
		return p.parseDropContractStatement()
	case "ENDPOINT":
		return p.parseDropEndpointStatement()
	case "MESSAGE":
		return p.parseDropMessageTypeStatement()
	case "QUEUE":
		return p.parseDropQueueStatement()
	case "REMOTE":
		return p.parseDropRemoteServiceBindingStatement()
	case "ROUTE":
		return p.parseDropRouteStatement()
	case "SERVICE":
		return p.parseDropServiceStatement()
	case "EVENT":
		return p.parseDropEventNotificationStatement()
	}

	// Handle LOGIN token explicitly
	if p.curTok.Type == TokenLogin {
		return p.parseDropLoginStatement()
	}

	return nil, fmt.Errorf("unexpected token after DROP: %s", p.curTok.Literal)
}

func (p *Parser) parseDropFulltextStatement() (ast.Statement, error) {
	// Consume FULLTEXT
	p.nextToken()

	keyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume CATALOG/INDEX/STOPLIST

	switch keyword {
	case "STOPLIST":
		stmt := &ast.DropFullTextStopListStatement{
			Name:       p.parseIdentifier(),
			IsIfExists: false,
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	case "CATALOG":
		stmt := &ast.DropFullTextCatalogStatement{
			Name: p.parseIdentifier(),
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	case "INDEX":
		// DROP FULLTEXT INDEX ON table
		if p.curTok.Type == TokenOn {
			p.nextToken() // consume ON
		}
		name, _ := p.parseSchemaObjectName()
		stmt := &ast.DropFulltextIndexStatement{
			TableName: name,
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	return nil, fmt.Errorf("unexpected token after DROP FULLTEXT: %s", keyword)
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

func (p *Parser) parseDropResourcePoolStatement() (*ast.DropResourcePoolStatement, error) {
	// Consume RESOURCE
	p.nextToken()

	// Expect POOL
	if strings.ToUpper(p.curTok.Literal) != "POOL" {
		return nil, fmt.Errorf("expected POOL after RESOURCE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.DropResourcePoolStatement{}

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

func (p *Parser) parseDropLoginStatement() (*ast.DropLoginStatement, error) {
	// Consume LOGIN
	p.nextToken()

	stmt := &ast.DropLoginStatement{
		Name: p.parseIdentifier(),
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

	// Check for WITH NO DEPENDENTS
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "NO" {
			p.nextToken() // consume NO
			if strings.ToUpper(p.curTok.Literal) == "DEPENDENTS" {
				p.nextToken() // consume DEPENDENTS
				stmt.WithNoDependents = true
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropAsymmetricKeyStatement() (*ast.DropAsymmetricKeyStatement, error) {
	// Consume ASYMMETRIC
	p.nextToken()

	// Expect KEY
	if strings.ToUpper(p.curTok.Literal) == "KEY" {
		p.nextToken()
	}

	stmt := &ast.DropAsymmetricKeyStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse key name
	stmt.Name = p.parseIdentifier()

	// Check for REMOVE PROVIDER KEY
	if strings.ToUpper(p.curTok.Literal) == "REMOVE" {
		p.nextToken() // consume REMOVE
		if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
			p.nextToken() // consume PROVIDER
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			stmt.RemoveProviderKey = true
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropSymmetricKeyStatement() (*ast.DropSymmetricKeyStatement, error) {
	// Consume SYMMETRIC
	p.nextToken()

	// Expect KEY
	if strings.ToUpper(p.curTok.Literal) == "KEY" {
		p.nextToken()
	}

	stmt := &ast.DropSymmetricKeyStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF, got %s", p.curTok.Literal)
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

	// Parse key name
	stmt.Name = p.parseIdentifier()

	// Check for REMOVE PROVIDER KEY
	if strings.ToUpper(p.curTok.Literal) == "REMOVE" {
		p.nextToken() // consume REMOVE
		if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
			p.nextToken() // consume PROVIDER
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			stmt.RemoveProviderKey = true
		}
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

	// Check for DATABASE AUDIT SPECIFICATION
	if strings.ToUpper(p.curTok.Literal) == "AUDIT" {
		p.nextToken() // consume AUDIT
		if strings.ToUpper(p.curTok.Literal) == "SPECIFICATION" {
			p.nextToken() // consume SPECIFICATION
		}
		stmt := &ast.DropDatabaseAuditSpecificationStatement{}
		// Check for IF EXISTS
		if strings.ToUpper(p.curTok.Literal) == "IF" {
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "EXISTS" {
				p.nextToken()
				stmt.IsIfExists = true
			}
		}
		stmt.Name = p.parseIdentifier()
		return stmt, nil
	}

	// Check for DATABASE ENCRYPTION KEY
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
		p.nextToken() // consume ENCRYPTION
		if p.curTok.Type == TokenKey {
			p.nextToken() // consume KEY
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return &ast.DropDatabaseEncryptionKeyStatement{}, nil
	}

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

	// Check for IF EXISTS
	if strings.ToUpper(p.curTok.Literal) == "IF" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
			return nil, fmt.Errorf("expected EXISTS after IF")
		}
		p.nextToken()
		stmt.IsIfExists = true
	}

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

func (p *Parser) parseDropServerRoleStatement() (ast.Statement, error) {
	// Consume SERVER
	p.nextToken()

	// Check if it's ROLE or AUDIT
	switch strings.ToUpper(p.curTok.Literal) {
	case "ROLE":
		p.nextToken()
		stmt := &ast.DropServerRoleStatement{}
		stmt.Name = p.parseIdentifier()
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	case "AUDIT":
		p.nextToken()
		// Check if next token is SPECIFICATION
		if strings.ToUpper(p.curTok.Literal) == "SPECIFICATION" {
			p.nextToken()
			stmt := &ast.DropServerAuditSpecificationStatement{}
			stmt.Name = p.parseIdentifier()
			// Skip optional semicolon
			if p.curTok.Type == TokenSemicolon {
				p.nextToken()
			}
			return stmt, nil
		}
		stmt := &ast.DropServerAuditStatement{}
		stmt.Name = p.parseIdentifier()
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	default:
		return nil, fmt.Errorf("expected ROLE or AUDIT after SERVER, got %s", p.curTok.Literal)
	}
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

	stmt := &ast.DropTriggerStatement{
		TriggerScope: "Normal", // Default to Normal for regular DROP TRIGGER
	}

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
			clause.Index = indexName
			clause.Object = tableName
		} else if p.curTok.Type == TokenDot {
			// Old backwards-compatible syntax: table.index
			p.nextToken() // consume dot
			childName := p.parseIdentifier()
			clause.LegacyIndex = &ast.SchemaObjectName{
				SchemaIdentifier: indexName,
				BaseIdentifier:   childName,
				Count:            2,
				Identifiers:      []*ast.Identifier{indexName, childName},
			}
		} else {
			// Just index name without ON or dot
			clause.Index = indexName
		}

		// Parse WITH options if present
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				clause.Options = p.parseDropIndexOptions()
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
		}

		stmt.DropIndexClauses = append(stmt.DropIndexClauses, clause)

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

func (p *Parser) parseDropIndexOptions() []ast.DropIndexOption {
	var options []ast.DropIndexOption

	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		switch upperLit {
		case "ONLINE":
			p.nextToken() // consume ONLINE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			optState := "Off"
			if strings.ToUpper(p.curTok.Literal) == "ON" {
				optState = "On"
			}
			p.nextToken()
			onlineOpt := &ast.OnlineIndexOption{
				OptionState: optState,
				OptionKind:  "Online",
			}
			// Check for optional (WAIT_AT_LOW_PRIORITY (...))
			if optState == "On" && p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				if strings.ToUpper(p.curTok.Literal) == "WAIT_AT_LOW_PRIORITY" {
					p.nextToken() // consume WAIT_AT_LOW_PRIORITY
					lowPriorityOpt := &ast.OnlineIndexLowPriorityLockWaitOption{}
					if p.curTok.Type == TokenLParen {
						p.nextToken() // consume (
						for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
							optName := strings.ToUpper(p.curTok.Literal)
							if optName == "MAX_DURATION" {
								p.nextToken() // consume MAX_DURATION
								if p.curTok.Type == TokenEquals {
									p.nextToken() // consume =
								}
								durVal, _ := p.parsePrimaryExpression()
								unit := "Minutes"
								if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
									p.nextToken()
								} else if strings.ToUpper(p.curTok.Literal) == "SECONDS" {
									unit = "Seconds"
									p.nextToken()
								}
								lowPriorityOpt.Options = append(lowPriorityOpt.Options, &ast.LowPriorityLockWaitMaxDurationOption{
									MaxDuration: durVal,
									Unit:        unit,
									OptionKind:  "MaxDuration",
								})
							} else if optName == "ABORT_AFTER_WAIT" {
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
								lowPriorityOpt.Options = append(lowPriorityOpt.Options, &ast.LowPriorityLockWaitAbortAfterWaitOption{
									AbortAfterWait: abortType,
									OptionKind:     "AbortAfterWait",
								})
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
					}
					onlineOpt.LowPriorityLockWaitOption = lowPriorityOpt
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			options = append(options, onlineOpt)
		case "MOVE":
			p.nextToken() // consume MOVE
			if strings.ToUpper(p.curTok.Literal) == "TO" {
				p.nextToken() // consume TO
			}
			moveTo := &ast.FileGroupOrPartitionScheme{}
			// Parse filegroup name
			fgName := p.parseIdentifier()
			moveTo.Name = &ast.IdentifierOrValueExpression{
				Value:      fgName.Value,
				Identifier: fgName,
			}
			// Check for partition columns
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				var cols []*ast.Identifier
				for {
					cols = append(cols, p.parseIdentifier())
					if p.curTok.Type != TokenComma {
						break
					}
					p.nextToken()
				}
				moveTo.PartitionSchemeColumns = cols
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			options = append(options, &ast.MoveToDropIndexOption{
				MoveTo:     moveTo,
				OptionKind: "MoveTo",
			})
		case "FILESTREAM_ON":
			p.nextToken() // consume FILESTREAM_ON
			ident := p.parseIdentifier()
			options = append(options, &ast.FileStreamOnDropIndexOption{
				FileStreamOn: &ast.IdentifierOrValueExpression{
					Value:      ident.Value,
					Identifier: ident,
				},
				OptionKind: "FileStreamOn",
			})
		case "DATA_COMPRESSION":
			p.nextToken() // consume DATA_COMPRESSION
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			level := "None"
			upperLevel := strings.ToUpper(p.curTok.Literal)
			switch upperLevel {
			case "ROW":
				level = "Row"
			case "PAGE":
				level = "Page"
			case "NONE":
				level = "None"
			}
			p.nextToken()
			options = append(options, &ast.DataCompressionOption{
				CompressionLevel: level,
				OptionKind:       "DataCompression",
			})
		case "MAXDOP":
			p.nextToken() // consume MAXDOP
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			var expr ast.ScalarExpression
			if p.curTok.Type == TokenNumber {
				expr = &ast.IntegerLiteral{
					LiteralType: "Integer",
					Value:       p.curTok.Literal,
				}
				p.nextToken()
			}
			options = append(options, &ast.IndexExpressionOption{
				Expression: expr,
				OptionKind: "MaxDop",
			})
		case "WAIT_AT_LOW_PRIORITY":
			p.nextToken() // consume WAIT_AT_LOW_PRIORITY
			waitOpt := &ast.WaitAtLowPriorityOption{
				OptionKind: "WaitAtLowPriority",
			}
			// Parse nested options inside parentheses
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for {
					optName := strings.ToUpper(p.curTok.Literal)
					if optName == "MAX_DURATION" {
						p.nextToken() // consume MAX_DURATION
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						maxDur := &ast.LowPriorityLockWaitMaxDurationOption{
							OptionKind: "MaxDuration",
						}
						// Parse integer value
						if p.curTok.Type == TokenNumber {
							maxDur.MaxDuration = &ast.IntegerLiteral{
								LiteralType: "Integer",
								Value:       p.curTok.Literal,
							}
							p.nextToken()
						}
						// Parse unit: MINUTES or SECONDS
						unitUpper := strings.ToUpper(p.curTok.Literal)
						if unitUpper == "MINUTES" {
							maxDur.Unit = "Minutes"
							p.nextToken()
						} else if unitUpper == "SECONDS" {
							maxDur.Unit = "Seconds"
							p.nextToken()
						}
						waitOpt.Options = append(waitOpt.Options, maxDur)
					} else if optName == "ABORT_AFTER_WAIT" {
						p.nextToken() // consume ABORT_AFTER_WAIT
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						abortOpt := &ast.LowPriorityLockWaitAbortAfterWaitOption{
							OptionKind: "AbortAfterWait",
						}
						abortValue := strings.ToUpper(p.curTok.Literal)
						switch abortValue {
						case "NONE":
							abortOpt.AbortAfterWait = "None"
						case "SELF":
							abortOpt.AbortAfterWait = "Self"
						case "BLOCKERS":
							abortOpt.AbortAfterWait = "Blockers"
						}
						p.nextToken()
						waitOpt.Options = append(waitOpt.Options, abortOpt)
					} else {
						break
					}
					if p.curTok.Type == TokenComma {
						p.nextToken() // consume comma
					} else {
						break
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			options = append(options, waitOpt)
		default:
			// Unknown option, skip
			p.nextToken()
		}

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		} else if p.curTok.Type == TokenRParen {
			break
		} else if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon {
			break
		}
	}

	return options
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

	stmt := &ast.DropSchemaStatement{
		DropBehavior: "None",
	}

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

	// Check for CASCADE or RESTRICT
	switch strings.ToUpper(p.curTok.Literal) {
	case "CASCADE":
		stmt.DropBehavior = "Cascade"
		p.nextToken()
	case "RESTRICT":
		stmt.DropBehavior = "Restrict"
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropPartitionStatement() (ast.Statement, error) {
	// Consume PARTITION
	p.nextToken()

	keyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	switch keyword {
	case "FUNCTION":
		stmt := &ast.DropPartitionFunctionStatement{
			Name: p.parseIdentifier(),
		}
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	case "SCHEME":
		stmt := &ast.DropPartitionSchemeStatement{
			Name: p.parseIdentifier(),
		}
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	return nil, fmt.Errorf("expected FUNCTION or SCHEME after PARTITION, got %s", keyword)
}

func (p *Parser) parseDropApplicationRoleStatement() (*ast.DropApplicationRoleStatement, error) {
	// Consume APPLICATION
	p.nextToken()
	// Consume ROLE
	if strings.ToUpper(p.curTok.Literal) == "ROLE" {
		p.nextToken()
	}

	stmt := &ast.DropApplicationRoleStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropCertificateStatement() (*ast.DropCertificateStatement, error) {
	// Consume CERTIFICATE
	p.nextToken()

	stmt := &ast.DropCertificateStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropMasterKeyStatement() (*ast.DropMasterKeyStatement, error) {
	// Consume MASTER
	p.nextToken()
	// Consume KEY
	if strings.ToUpper(p.curTok.Literal) == "KEY" {
		p.nextToken()
	}

	stmt := &ast.DropMasterKeyStatement{}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropXmlSchemaCollectionStatement() (*ast.DropXmlSchemaCollectionStatement, error) {
	// Consume XML
	p.nextToken()
	// Consume SCHEMA
	if strings.ToUpper(p.curTok.Literal) == "SCHEMA" {
		p.nextToken()
	}
	// Consume COLLECTION
	if strings.ToUpper(p.curTok.Literal) == "COLLECTION" {
		p.nextToken()
	}

	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}

	stmt := &ast.DropXmlSchemaCollectionStatement{
		Name: name,
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropContractStatement() (*ast.DropContractStatement, error) {
	// Consume CONTRACT
	p.nextToken()

	stmt := &ast.DropContractStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropEndpointStatement() (*ast.DropEndpointStatement, error) {
	// Consume ENDPOINT
	p.nextToken()

	stmt := &ast.DropEndpointStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropMessageTypeStatement() (*ast.DropMessageTypeStatement, error) {
	// Consume MESSAGE
	p.nextToken()
	// Consume TYPE
	if strings.ToUpper(p.curTok.Literal) == "TYPE" {
		p.nextToken()
	}

	stmt := &ast.DropMessageTypeStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropQueueStatement() (*ast.DropQueueStatement, error) {
	// Consume QUEUE
	p.nextToken()

	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}

	stmt := &ast.DropQueueStatement{
		Name: name,
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropRemoteServiceBindingStatement() (*ast.DropRemoteServiceBindingStatement, error) {
	// Consume REMOTE
	p.nextToken()
	// Consume SERVICE
	if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
		p.nextToken()
	}
	// Consume BINDING
	if strings.ToUpper(p.curTok.Literal) == "BINDING" {
		p.nextToken()
	}

	stmt := &ast.DropRemoteServiceBindingStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropRouteStatement() (*ast.DropRouteStatement, error) {
	// Consume ROUTE
	p.nextToken()

	stmt := &ast.DropRouteStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropServiceStatement() (*ast.DropServiceStatement, error) {
	// Consume SERVICE
	p.nextToken()

	stmt := &ast.DropServiceStatement{
		Name: p.parseIdentifier(),
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropEventNotificationStatement() (*ast.DropEventNotificationStatement, error) {
	// Consume EVENT
	p.nextToken()
	// Consume NOTIFICATION
	if strings.ToUpper(p.curTok.Literal) == "NOTIFICATION" {
		p.nextToken()
	}

	stmt := &ast.DropEventNotificationStatement{}

	// Parse notification names (comma-separated)
	for {
		stmt.Notifications = append(stmt.Notifications, p.parseIdentifier())
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}
		break
	}

	// Parse ON clause
	if p.curTok.Type == TokenOn {
		p.nextToken()
	}

	scope := &ast.EventNotificationObjectScope{}
	switch strings.ToUpper(p.curTok.Literal) {
	case "SERVER":
		scope.Target = "Server"
		p.nextToken()
	case "DATABASE":
		scope.Target = "Database"
		p.nextToken()
	case "QUEUE":
		scope.Target = "Queue"
		p.nextToken()
		queueName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		scope.QueueName = queueName
	}
	stmt.Scope = scope

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
	case TokenView:
		return p.parseAlterViewStatement()
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
		case "BROKER":
			return p.parseAlterBrokerPriorityStatement()
		case "FEDERATION":
			return p.parseAlterFederationStatement()
		case "WORKLOAD":
			return p.parseAlterWorkloadGroupStatement()
		case "SEQUENCE":
			return p.parseAlterSequenceStatement()
		case "SEARCH":
			return p.parseAlterSearchPropertyListStatement()
		case "AVAILABILITY":
			return p.parseAlterAvailabilityGroupStatement()
		case "MATERIALIZED":
			return p.parseAlterMaterializedViewStatement()
		}
		return nil, fmt.Errorf("unexpected token after ALTER: %s", p.curTok.Literal)
	default:
		return nil, fmt.Errorf("unexpected token after ALTER: %s", p.curTok.Literal)
	}
}

func (p *Parser) parseAlterResourceGovernorStatement() (ast.Statement, error) {
	// Consume RESOURCE
	p.nextToken()

	// Check if this is RESOURCE POOL or RESOURCE GOVERNOR
	if strings.ToUpper(p.curTok.Literal) == "POOL" {
		return p.parseAlterResourcePoolStatement()
	}

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

	// Check for DATABASE AUDIT SPECIFICATION
	if strings.ToUpper(p.curTok.Literal) == "AUDIT" {
		p.nextToken() // consume AUDIT
		if strings.ToUpper(p.curTok.Literal) == "SPECIFICATION" {
			p.nextToken() // consume SPECIFICATION
			return p.parseAlterDatabaseAuditSpecificationStatement()
		}
	}

	// Check for DATABASE ENCRYPTION KEY
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
		return p.parseAlterDatabaseEncryptionKeyStatement()
	}

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
		// SCOPED is actually a database name, treat it as such
		dbName := &ast.Identifier{Value: "SCOPED", QuoteType: "NotQuoted"}
		// Check for COLLATE
		if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
			p.nextToken() // consume COLLATE
			stmt := &ast.AlterDatabaseCollateStatement{
				DatabaseName: dbName,
				Collation:    p.parseIdentifier(),
			}
			p.skipToEndOfStatement()
			return stmt, nil
		}
		// Fall through to skip rest
		p.skipToEndOfStatement()
		return &ast.AlterDatabaseSetStatement{DatabaseName: dbName}, nil
	}

	// Parse database name followed by various commands
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket || p.curTok.Type == TokenCurrent {
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
			if strings.ToUpper(p.curTok.Literal) == "REBUILD" {
				return p.parseAlterDatabaseRebuildLogStatement(dbName)
			}
			if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
				p.nextToken() // consume COLLATE
				stmt := &ast.AlterDatabaseCollateStatement{
					DatabaseName: dbName,
					Collation:    p.parseIdentifier(),
				}
				p.skipToEndOfStatement()
				return stmt, nil
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

func (p *Parser) parseAlterDatabaseEncryptionKeyStatement() (*ast.AlterDatabaseEncryptionKeyStatement, error) {
	// curTok is ENCRYPTION
	p.nextToken() // consume ENCRYPTION

	// Consume KEY
	if p.curTok.Type == TokenKey {
		p.nextToken()
	}

	stmt := &ast.AlterDatabaseEncryptionKeyStatement{
		Algorithm: "None", // Default when not specified
	}

	// Check for REGENERATE
	if strings.ToUpper(p.curTok.Literal) == "REGENERATE" {
		stmt.Regenerate = true
		p.nextToken() // consume REGENERATE
	}

	// WITH ALGORITHM = ...
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
	}

	if strings.ToUpper(p.curTok.Literal) == "ALGORITHM" {
		p.nextToken() // consume ALGORITHM
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		stmt.Algorithm = normalizeAlgorithmName(p.curTok.Literal)
		p.nextToken()
	}

	// ENCRYPTION BY SERVER CERTIFICATE|ASYMMETRIC KEY name
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
		p.nextToken() // consume ENCRYPTION
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken() // consume SERVER
		}

		mechanism := &ast.CryptoMechanism{}
		mechType := strings.ToUpper(p.curTok.Literal)
		if mechType == "CERTIFICATE" {
			p.nextToken()
			mechanism.CryptoMechanismType = "Certificate"
			mechanism.Identifier = p.parseIdentifier()
		} else if mechType == "ASYMMETRIC" {
			p.nextToken()
			if p.curTok.Type == TokenKey {
				p.nextToken() // consume KEY
			}
			mechanism.CryptoMechanismType = "AsymmetricKey"
			mechanism.Identifier = p.parseIdentifier()
		}
		stmt.Encryptor = mechanism
	}

	// Skip to end of statement
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterDatabaseSetStatement(dbName *ast.Identifier) (*ast.AlterDatabaseSetStatement, error) {
	// Consume SET
	p.nextToken()

	stmt := &ast.AlterDatabaseSetStatement{}
	// Check if this is ALTER DATABASE CURRENT SET
	if dbName != nil && strings.ToUpper(dbName.Value) == "CURRENT" {
		stmt.UseCurrent = true
	} else {
		stmt.DatabaseName = dbName
	}

	// Parse options
	for {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		switch optionName {
		// Simple database options without ON/OFF
		case "ONLINE", "OFFLINE":
			opt := &ast.SimpleDatabaseOption{OptionKind: capitalizeFirst(strings.ToLower(optionName))}
			stmt.Options = append(stmt.Options, opt)
		case "SINGLE_USER":
			opt := &ast.SimpleDatabaseOption{OptionKind: "SingleUser"}
			stmt.Options = append(stmt.Options, opt)
		case "RESTRICTED_USER":
			opt := &ast.SimpleDatabaseOption{OptionKind: "RestrictedUser"}
			stmt.Options = append(stmt.Options, opt)
		case "MULTI_USER":
			opt := &ast.SimpleDatabaseOption{OptionKind: "MultiUser"}
			stmt.Options = append(stmt.Options, opt)
		case "READ_ONLY":
			opt := &ast.SimpleDatabaseOption{OptionKind: "ReadOnly"}
			stmt.Options = append(stmt.Options, opt)
		case "READ_WRITE":
			opt := &ast.SimpleDatabaseOption{OptionKind: "ReadWrite"}
			stmt.Options = append(stmt.Options, opt)
		case "RECOVERY":
			// Expect FULL, BULK_LOGGED, or SIMPLE
			recoveryType := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			recoveryValue := "Full"
			switch recoveryType {
			case "BULK_LOGGED":
				recoveryValue = "BulkLogged"
			case "SIMPLE":
				recoveryValue = "Simple"
			}
			opt := &ast.RecoveryDatabaseOption{OptionKind: "Recovery", Value: recoveryValue}
			stmt.Options = append(stmt.Options, opt)
		case "CURSOR_CLOSE_ON_COMMIT":
			// Expects ON/OFF
			optionValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "CursorCloseOnCommit",
				OptionState: capitalizeFirst(optionValue),
			}
			stmt.Options = append(stmt.Options, opt)
		case "CURSOR_DEFAULT":
			// Expects LOCAL or GLOBAL
			cursorValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.CursorDefaultDatabaseOption{
				OptionKind: "CursorDefault",
				IsLocal:    cursorValue == "LOCAL",
			}
			stmt.Options = append(stmt.Options, opt)
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
		case "AUTOMATIC_TUNING":
			opt := &ast.AutomaticTuningDatabaseOption{
				OptionKind:           "AutomaticTuning",
				AutomaticTuningState: "NotSet",
			}
			// Check for = INHERIT/CUSTOM/AUTO or (sub-options)
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
				stateVal := strings.ToUpper(p.curTok.Literal)
				opt.AutomaticTuningState = capitalizeFirst(stateVal)
				p.nextToken()
			}
			// Parse optional sub-options in parentheses
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					subOptName := strings.ToUpper(p.curTok.Literal)
					p.nextToken() // consume option name
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					subOptValue := capitalizeFirst(strings.ToUpper(p.curTok.Literal))
					p.nextToken() // consume value
					switch subOptName {
					case "CREATE_INDEX":
						opt.Options = append(opt.Options, &ast.AutomaticTuningCreateIndexOption{
							OptionKind: "Create_Index",
							Value:      subOptValue,
						})
					case "DROP_INDEX":
						opt.Options = append(opt.Options, &ast.AutomaticTuningDropIndexOption{
							OptionKind: "Drop_Index",
							Value:      subOptValue,
						})
					case "FORCE_LAST_GOOD_PLAN":
						opt.Options = append(opt.Options, &ast.AutomaticTuningForceLastGoodPlanOption{
							OptionKind: "Force_Last_Good_Plan",
							Value:      subOptValue,
						})
					case "MAINTAIN_INDEX":
						opt.Options = append(opt.Options, &ast.AutomaticTuningMaintainIndexOption{
							OptionKind: "Maintain_Index",
							Value:      subOptValue,
						})
					}
					if p.curTok.Type == TokenComma {
						p.nextToken()
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
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
		case "AUTO_CREATE_STATISTICS":
			// Parse ON/OFF and optional (INCREMENTAL = ON/OFF)
			optionValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.AutoCreateStatisticsDatabaseOption{
				OptionKind:  "AutoCreateStatistics",
				OptionState: capitalizeFirst(optionValue),
			}
			// Check for (INCREMENTAL = ON/OFF)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "INCREMENTAL" {
					p.nextToken() // consume INCREMENTAL
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
						incState := strings.ToUpper(p.curTok.Literal)
						p.nextToken() // consume ON/OFF
						opt.HasIncremental = true
						opt.IncrementalState = capitalizeFirst(incState)
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			stmt.Options = append(stmt.Options, opt)
		case "REMOTE_DATA_ARCHIVE":
			rdaOpt, err := p.parseRemoteDataArchiveOption()
			if err != nil {
				return nil, err
			}
			stmt.Options = append(stmt.Options, rdaOpt)
		case "COMPATIBILITY_LEVEL":
			// Parse = value
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after COMPATIBILITY_LEVEL")
			}
			p.nextToken() // consume =
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			opt := &ast.LiteralDatabaseOption{
				OptionKind: "CompatibilityLevel",
				Value:      val,
			}
			stmt.Options = append(stmt.Options, opt)
		case "CHANGE_TRACKING":
			ctOpt, err := p.parseChangeTrackingOption()
			if err != nil {
				return nil, err
			}
			stmt.Options = append(stmt.Options, ctOpt)
		case "EMERGENCY":
			opt := &ast.GenericDatabaseOption{OptionKind: "Emergency"}
			stmt.Options = append(stmt.Options, opt)
		case "ERROR_BROKER_CONVERSATIONS":
			opt := &ast.GenericDatabaseOption{OptionKind: "ErrorBrokerConversations"}
			stmt.Options = append(stmt.Options, opt)
		case "ENABLE_BROKER":
			opt := &ast.GenericDatabaseOption{OptionKind: "EnableBroker"}
			stmt.Options = append(stmt.Options, opt)
		case "DISABLE_BROKER":
			opt := &ast.GenericDatabaseOption{OptionKind: "DisableBroker"}
			stmt.Options = append(stmt.Options, opt)
		case "NEW_BROKER":
			opt := &ast.GenericDatabaseOption{OptionKind: "NewBroker"}
			stmt.Options = append(stmt.Options, opt)
		case "PAGE_VERIFY":
			// PAGE_VERIFY CHECKSUM | NONE | TORN_PAGE_DETECTION
			verifyValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			value := "None"
			switch verifyValue {
			case "CHECKSUM":
				value = "Checksum"
			case "TORN_PAGE_DETECTION":
				value = "TornPageDetection"
			}
			opt := &ast.PageVerifyDatabaseOption{
				OptionKind: "PageVerify",
				Value:      value,
			}
			stmt.Options = append(stmt.Options, opt)
		case "PARTNER":
			opt, err := p.parsePartnerDatabaseOption()
			if err != nil {
				return nil, err
			}
			stmt.Options = append(stmt.Options, opt)
		case "WITNESS":
			opt, err := p.parseWitnessDatabaseOption()
			if err != nil {
				return nil, err
			}
			stmt.Options = append(stmt.Options, opt)
		case "PARAMETERIZATION":
			// PARAMETERIZATION SIMPLE | FORCED
			paramValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.ParameterizationDatabaseOption{
				OptionKind: "Parameterization",
				IsSimple:   paramValue == "SIMPLE",
			}
			stmt.Options = append(stmt.Options, opt)
		case "CONTAINMENT":
			// CONTAINMENT = NONE | PARTIAL
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			containmentValue := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			value := "None"
			if containmentValue == "PARTIAL" {
				value = "Partial"
			}
			opt := &ast.ContainmentDatabaseOption{
				OptionKind: "Containment",
				Value:      value,
			}
			stmt.Options = append(stmt.Options, opt)
		case "TRANSFORM_NOISE_WORDS":
			// TRANSFORM_NOISE_WORDS = ON/OFF
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "TransformNoiseWords",
				OptionState: capitalizeFirst(state),
			}
			stmt.Options = append(stmt.Options, opt)
		case "DEFAULT_LANGUAGE":
			// DEFAULT_LANGUAGE = identifier | integer
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			if p.curTok.Type == TokenNumber {
				opt := &ast.LiteralDatabaseOption{
					OptionKind: "DefaultLanguage",
					Value: &ast.IntegerLiteral{
						LiteralType: "Integer",
						Value:       p.curTok.Literal,
					},
				}
				stmt.Options = append(stmt.Options, opt)
				p.nextToken()
			} else {
				opt := &ast.IdentifierDatabaseOption{
					OptionKind: "DefaultLanguage",
					Value:      p.parseIdentifier(),
				}
				stmt.Options = append(stmt.Options, opt)
			}
		case "DEFAULT_FULLTEXT_LANGUAGE":
			// DEFAULT_FULLTEXT_LANGUAGE = identifier | integer
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			if p.curTok.Type == TokenNumber {
				opt := &ast.LiteralDatabaseOption{
					OptionKind: "DefaultFullTextLanguage",
					Value: &ast.IntegerLiteral{
						LiteralType: "Integer",
						Value:       p.curTok.Literal,
					},
				}
				stmt.Options = append(stmt.Options, opt)
				p.nextToken()
			} else {
				opt := &ast.IdentifierDatabaseOption{
					OptionKind: "DefaultFullTextLanguage",
					Value:      p.parseIdentifier(),
				}
				stmt.Options = append(stmt.Options, opt)
			}
		case "TWO_DIGIT_YEAR_CUTOFF":
			// TWO_DIGIT_YEAR_CUTOFF = integer
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			opt := &ast.LiteralDatabaseOption{
				OptionKind: "TwoDigitYearCutoff",
				Value: &ast.IntegerLiteral{
					LiteralType: "Integer",
					Value:       p.curTok.Literal,
				},
			}
			stmt.Options = append(stmt.Options, opt)
			p.nextToken()
		case "HADR":
			// HADR {SUSPEND|RESUME|OFF|AVAILABILITY GROUP = name}
			hadrOpt := strings.ToUpper(p.curTok.Literal)
			switch hadrOpt {
			case "SUSPEND":
				p.nextToken()
				stmt.Options = append(stmt.Options, &ast.HadrDatabaseOption{
					HadrOption: "Suspend",
					OptionKind: "Hadr",
				})
			case "RESUME":
				p.nextToken()
				stmt.Options = append(stmt.Options, &ast.HadrDatabaseOption{
					HadrOption: "Resume",
					OptionKind: "Hadr",
				})
			case "OFF":
				p.nextToken()
				stmt.Options = append(stmt.Options, &ast.HadrDatabaseOption{
					HadrOption: "Off",
					OptionKind: "Hadr",
				})
			case "AVAILABILITY":
				p.nextToken() // consume AVAILABILITY
				if strings.ToUpper(p.curTok.Literal) == "GROUP" {
					p.nextToken() // consume GROUP
				}
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				groupName := p.parseIdentifier()
				stmt.Options = append(stmt.Options, &ast.HadrAvailabilityGroupDatabaseOption{
					GroupName:  groupName,
					HadrOption: "AvailabilityGroup",
					OptionKind: "Hadr",
				})
			default:
				// Unknown HADR option
				p.nextToken()
			}
		case "FILESTREAM":
			// FILESTREAM(NON_TRANSACTED_ACCESS=OFF|READ_ONLY|FULL, DIRECTORY_NAME='...')
			opt := &ast.FileStreamDatabaseOption{
				OptionKind: "FileStream",
			}
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					subOpt := strings.ToUpper(p.curTok.Literal)
					p.nextToken() // consume option name
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					switch subOpt {
					case "NON_TRANSACTED_ACCESS":
						accessVal := strings.ToUpper(p.curTok.Literal)
						p.nextToken()
						switch accessVal {
						case "OFF":
							opt.NonTransactedAccess = "Off"
						case "READ_ONLY":
							opt.NonTransactedAccess = "ReadOnly"
						case "FULL":
							opt.NonTransactedAccess = "Full"
						}
					case "DIRECTORY_NAME":
						// Can be a string literal or NULL
						if strings.ToUpper(p.curTok.Literal) == "NULL" {
							opt.DirectoryName = &ast.NullLiteral{
								LiteralType: "Null",
								Value:       p.curTok.Literal, // Preserve original case
							}
							p.nextToken()
						} else if p.curTok.Type == TokenString {
							opt.DirectoryName = &ast.StringLiteral{
								LiteralType:   "String",
								Value:         strings.Trim(p.curTok.Literal, "'"),
								IsNational:    false,
								IsLargeObject: false,
							}
							p.nextToken()
						}
					}
					if p.curTok.Type == TokenComma {
						p.nextToken()
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			stmt.Options = append(stmt.Options, opt)
		case "TARGET_RECOVERY_TIME":
			// TARGET_RECOVERY_TIME = N SECONDS|MINUTES
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			timeVal, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			unit := "Seconds"
			if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
				unit = "Minutes"
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "SECONDS" {
				p.nextToken()
			}
			trtOpt := &ast.TargetRecoveryTimeDatabaseOption{
				OptionKind:   "TargetRecoveryTime",
				RecoveryTime: timeVal,
				Unit:         unit,
			}
			stmt.Options = append(stmt.Options, trtOpt)
		case "QUERY_STORE":
			qsOpt, err := p.parseQueryStoreOption()
			if err != nil {
				return nil, err
			}
			stmt.Options = append(stmt.Options, qsOpt)
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

	// Parse optional termination clause: WITH NO_WAIT | WITH ROLLBACK AFTER N [SECONDS] | WITH ROLLBACK IMMEDIATE
	if p.curTok.Type == TokenWith || strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		term := &ast.AlterDatabaseTermination{}
		termKeyword := strings.ToUpper(p.curTok.Literal)
		if termKeyword == "NO_WAIT" {
			term.NoWait = true
			p.nextToken()
		} else if termKeyword == "ROLLBACK" {
			p.nextToken() // consume ROLLBACK
			rollbackType := strings.ToUpper(p.curTok.Literal)
			if rollbackType == "AFTER" {
				p.nextToken() // consume AFTER
				// Parse the number
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				term.RollbackAfter = val
				// Optional SECONDS keyword
				if strings.ToUpper(p.curTok.Literal) == "SECONDS" {
					p.nextToken()
				}
			} else if rollbackType == "IMMEDIATE" {
				term.ImmediateRollback = true
				p.nextToken()
			}
		}
		stmt.Termination = term
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseAlterDatabaseTermination parses the termination clause for ALTER DATABASE statements
// Forms: WITH NO_WAIT | WITH ROLLBACK AFTER N [SECONDS] | WITH ROLLBACK IMMEDIATE
func (p *Parser) parseAlterDatabaseTermination() *ast.AlterDatabaseTermination {
	if p.curTok.Type != TokenWith && strings.ToUpper(p.curTok.Literal) != "WITH" {
		return nil
	}
	p.nextToken() // consume WITH
	term := &ast.AlterDatabaseTermination{}
	termKeyword := strings.ToUpper(p.curTok.Literal)
	if termKeyword == "NO_WAIT" {
		term.NoWait = true
		p.nextToken()
	} else if termKeyword == "ROLLBACK" {
		p.nextToken() // consume ROLLBACK
		rollbackType := strings.ToUpper(p.curTok.Literal)
		if rollbackType == "AFTER" {
			p.nextToken() // consume AFTER
			// Parse the number
			val, _ := p.parseScalarExpression()
			term.RollbackAfter = val
			// Optional SECONDS keyword
			if strings.ToUpper(p.curTok.Literal) == "SECONDS" {
				p.nextToken()
			}
		} else if rollbackType == "IMMEDIATE" {
			term.ImmediateRollback = true
			p.nextToken()
		}
	}
	return term
}

// parseRemoteDataArchiveOption parses REMOTE_DATA_ARCHIVE option
// Forms:
//   - REMOTE_DATA_ARCHIVE = ON (options...)
//   - REMOTE_DATA_ARCHIVE = OFF
//   - REMOTE_DATA_ARCHIVE (options...) -- OptionState is "NotSet"
func (p *Parser) parseRemoteDataArchiveOption() (*ast.RemoteDataArchiveDatabaseOption, error) {
	opt := &ast.RemoteDataArchiveDatabaseOption{
		OptionKind:  "RemoteDataArchive",
		OptionState: "NotSet",
	}

	// Check for = ON/OFF or just (
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		stateVal := strings.ToUpper(p.curTok.Literal)
		opt.OptionState = capitalizeFirst(stateVal)
		p.nextToken() // consume ON/OFF
	}

	// Parse settings if we have (
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for {
			settingName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume setting name

			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after %s, got %s", settingName, p.curTok.Literal)
			}
			p.nextToken() // consume =

			switch settingName {
			case "SERVER":
				// Parse string literal
				server, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				setting := &ast.RemoteDataArchiveDbServerSetting{
					SettingKind: "Server",
					Server:      server,
				}
				opt.Settings = append(opt.Settings, setting)
			case "CREDENTIAL":
				// Parse identifier (may be bracketed)
				cred := p.parseIdentifier()
				setting := &ast.RemoteDataArchiveDbCredentialSetting{
					SettingKind: "Credential",
					Credential:  cred,
				}
				opt.Settings = append(opt.Settings, setting)
			case "FEDERATED_SERVICE_ACCOUNT":
				// Parse ON/OFF
				isOn := strings.ToUpper(p.curTok.Literal) == "ON"
				p.nextToken()
				setting := &ast.RemoteDataArchiveDbFederatedServiceAccountSetting{
					SettingKind: "FederatedServiceAccount",
					IsOn:        isOn,
				}
				opt.Settings = append(opt.Settings, setting)
			default:
				return nil, fmt.Errorf("unknown REMOTE_DATA_ARCHIVE setting: %s", settingName)
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after REMOTE_DATA_ARCHIVE settings, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	}

	return opt, nil
}

// parseChangeTrackingOption parses CHANGE_TRACKING option
// Forms:
//   - CHANGE_TRACKING = ON (options...)
//   - CHANGE_TRACKING = OFF
//   - CHANGE_TRACKING (options...) -- OptionState is "NotSet"
func (p *Parser) parseChangeTrackingOption() (*ast.ChangeTrackingDatabaseOption, error) {
	opt := &ast.ChangeTrackingDatabaseOption{
		OptionKind:  "ChangeTracking",
		OptionState: "NotSet",
	}

	// Check for = ON/OFF or just (
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		stateVal := strings.ToUpper(p.curTok.Literal)
		opt.OptionState = capitalizeFirst(stateVal)
		p.nextToken() // consume ON/OFF
	}

	// Parse details if we have (
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for {
			detailName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume detail name

			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after %s, got %s", detailName, p.curTok.Literal)
			}
			p.nextToken() // consume =

			switch detailName {
			case "AUTO_CLEANUP":
				// Parse ON/OFF
				isOn := strings.ToUpper(p.curTok.Literal) == "ON"
				p.nextToken()
				detail := &ast.AutoCleanupChangeTrackingOptionDetail{
					IsOn: isOn,
				}
				opt.Details = append(opt.Details, detail)
			case "CHANGE_RETENTION":
				// Parse value and unit (e.g., 100 HOURS, 3 DAYS, 5 MINUTES)
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				unit := ""
				unitVal := strings.ToUpper(p.curTok.Literal)
				switch unitVal {
				case "DAYS":
					unit = "Days"
				case "HOURS":
					unit = "Hours"
				case "MINUTES":
					unit = "Minutes"
				}
				if unit != "" {
					p.nextToken() // consume unit
				}
				detail := &ast.ChangeRetentionChangeTrackingOptionDetail{
					RetentionPeriod: val,
					Unit:            unit,
				}
				opt.Details = append(opt.Details, detail)
			default:
				return nil, fmt.Errorf("unknown CHANGE_TRACKING detail: %s", detailName)
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after CHANGE_TRACKING details, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	}

	return opt, nil
}

// parseQueryStoreOption parses QUERY_STORE database option
// Forms:
//   - QUERY_STORE = ON (options...)
//   - QUERY_STORE = OFF
//   - QUERY_STORE (options...)
//   - QUERY_STORE CLEAR [ALL]
func (p *Parser) parseQueryStoreOption() (*ast.QueryStoreDatabaseOption, error) {
	opt := &ast.QueryStoreDatabaseOption{
		OptionKind:  "QueryStore",
		OptionState: "NotSet",
	}

	// Check for = ON/OFF or CLEAR or just (
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		stateVal := strings.ToUpper(p.curTok.Literal)
		opt.OptionState = capitalizeFirst(stateVal)
		p.nextToken() // consume ON/OFF
	} else if strings.ToUpper(p.curTok.Literal) == "CLEAR" {
		p.nextToken() // consume CLEAR
		if strings.ToUpper(p.curTok.Literal) == "ALL" {
			opt.ClearAll = true
			p.nextToken() // consume ALL
		} else {
			opt.Clear = true
		}
		return opt, nil
	}

	// Parse options if we have (
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			switch optName {
			case "DESIRED_STATE":
				val := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				stateOpt := &ast.QueryStoreDesiredStateOption{
					OptionKind: "Desired_State",
				}
				switch val {
				case "READ_ONLY":
					stateOpt.Value = "ReadOnly"
				case "READ_WRITE":
					stateOpt.Value = "ReadWrite"
				case "OFF":
					stateOpt.Value = "Off"
				}
				opt.Options = append(opt.Options, stateOpt)
			case "OPERATION_MODE":
				val := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				stateOpt := &ast.QueryStoreDesiredStateOption{
					OptionKind:             "Desired_State",
					OperationModeSpecified: true,
				}
				switch val {
				case "READ_ONLY":
					stateOpt.Value = "ReadOnly"
				case "READ_WRITE":
					stateOpt.Value = "ReadWrite"
				case "OFF":
					stateOpt.Value = "Off"
				}
				opt.Options = append(opt.Options, stateOpt)
			case "QUERY_CAPTURE_MODE":
				val := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				captureOpt := &ast.QueryStoreCapturePolicyOption{
					OptionKind: "Query_Capture_Mode",
					Value:      val,
				}
				opt.Options = append(opt.Options, captureOpt)
			case "SIZE_BASED_CLEANUP_MODE":
				val := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				cleanupOpt := &ast.QueryStoreSizeCleanupPolicyOption{
					OptionKind: "Size_Based_Cleanup_Mode",
					Value:      val,
				}
				opt.Options = append(opt.Options, cleanupOpt)
			case "FLUSH_INTERVAL_SECONDS", "DATA_FLUSH_INTERVAL_SECONDS":
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				flushOpt := &ast.QueryStoreDataFlushIntervalOption{
					OptionKind:    "Flush_Interval_Seconds",
					FlushInterval: val,
				}
				opt.Options = append(opt.Options, flushOpt)
			case "INTERVAL_LENGTH_MINUTES":
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				intervalOpt := &ast.QueryStoreIntervalLengthOption{
					OptionKind:          "Interval_Length_Minutes",
					StatsIntervalLength: val,
				}
				opt.Options = append(opt.Options, intervalOpt)
			case "MAX_STORAGE_SIZE_MB":
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				storageOpt := &ast.QueryStoreMaxStorageSizeOption{
					OptionKind: "Current_Storage_Size_MB",
					MaxQdsSize: val,
				}
				opt.Options = append(opt.Options, storageOpt)
			case "MAX_PLANS_PER_QUERY":
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				plansOpt := &ast.QueryStoreMaxPlansPerQueryOption{
					OptionKind:       "Max_Plans_Per_Query",
					MaxPlansPerQuery: val,
				}
				opt.Options = append(opt.Options, plansOpt)
			case "CLEANUP_POLICY":
				// Expect (STALE_QUERY_THRESHOLD_DAYS = N)
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
						subOptName := strings.ToUpper(p.curTok.Literal)
						p.nextToken() // consume sub-option name
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if subOptName == "STALE_QUERY_THRESHOLD_DAYS" {
							val, err := p.parseScalarExpression()
							if err != nil {
								return nil, err
							}
							thresholdOpt := &ast.QueryStoreTimeCleanupPolicyOption{
								OptionKind:          "Stale_Query_Threshold_Days",
								StaleQueryThreshold: val,
							}
							opt.Options = append(opt.Options, thresholdOpt)
						}
						if p.curTok.Type == TokenComma {
							p.nextToken()
						}
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}
			case "WAIT_STATS_CAPTURE_MODE":
				val := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				waitOpt := &ast.QueryStoreWaitStatsCaptureOption{
					OptionKind:  "Wait_Stats_Capture_Mode",
					OptionState: capitalizeFirst(val),
				}
				opt.Options = append(opt.Options, waitOpt)
			default:
				// Skip unknown option
				if p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen {
					p.nextToken()
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
	}

	return opt, nil
}

// parsePartnerDatabaseOption parses PARTNER database mirroring option
func (p *Parser) parsePartnerDatabaseOption() (*ast.PartnerDatabaseOption, error) {
	opt := &ast.PartnerDatabaseOption{
		OptionKind: "Partner",
	}

	// Check if next token is = (PARTNER = 'server')
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		server, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		opt.PartnerServer = server
		opt.PartnerOption = "PartnerServer"
		return opt, nil
	}

	// Otherwise, parse partner action
	action := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	switch action {
	case "FAILOVER":
		opt.PartnerOption = "Failover"
	case "FORCE_SERVICE_ALLOW_DATA_LOSS":
		opt.PartnerOption = "ForceServiceAllowDataLoss"
	case "RESUME":
		opt.PartnerOption = "Resume"
	case "SUSPEND":
		opt.PartnerOption = "Suspend"
	case "SAFETY":
		// SAFETY FULL or SAFETY OFF
		safetyVal := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		if safetyVal == "FULL" {
			opt.PartnerOption = "SafetyFull"
		} else {
			opt.PartnerOption = "SafetyOff"
		}
	case "TIMEOUT":
		// TIMEOUT value
		opt.PartnerOption = "Timeout"
		val, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		opt.Timeout = val
	default:
		opt.PartnerOption = capitalizeFirst(strings.ToLower(action))
	}

	return opt, nil
}

// parseWitnessDatabaseOption parses WITNESS database mirroring option
func (p *Parser) parseWitnessDatabaseOption() (*ast.WitnessDatabaseOption, error) {
	opt := &ast.WitnessDatabaseOption{
		OptionKind: "Witness",
	}

	// Check if next token is = (WITNESS = 'server')
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		server, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		opt.WitnessServer = server
		return opt, nil
	}

	// Check for WITNESS OFF
	if strings.ToUpper(p.curTok.Literal) == "OFF" {
		opt.IsOff = true
		p.nextToken()
	}

	return opt, nil
}

func (p *Parser) parseAlterDatabaseAddStatement(dbName *ast.Identifier) (ast.Statement, error) {
	// Consume ADD
	p.nextToken()

	switch strings.ToUpper(p.curTok.Literal) {
	case "FILE":
		p.nextToken() // consume FILE
		stmt := &ast.AlterDatabaseAddFileStatement{
			DatabaseName: dbName,
			IsLog:        false,
		}
		// Parse file declarations
		decls, err := p.parseFileDeclarationList(false)
		if err != nil {
			return nil, err
		}
		stmt.FileDeclarations = decls
		// Parse TO FILEGROUP
		if strings.ToUpper(p.curTok.Literal) == "TO" {
			p.nextToken() // consume TO
			if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
				p.nextToken() // consume FILEGROUP
			}
			stmt.FileGroup = p.parseIdentifier()
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
			IsLog:        true,
		}
		// Parse file declarations
		decls, err := p.parseFileDeclarationList(false)
		if err != nil {
			return nil, err
		}
		stmt.FileDeclarations = decls
		// Parse TO FILEGROUP
		if strings.ToUpper(p.curTok.Literal) == "TO" {
			p.nextToken() // consume TO
			if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
				p.nextToken() // consume FILEGROUP
			}
			stmt.FileGroup = p.parseIdentifier()
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
		// Parse the file declaration (NAME = n1, NEWNAME = n2)
		decls, err := p.parseFileDeclarationList(false)
		if err != nil {
			return nil, err
		}
		if len(decls) > 0 {
			stmt.FileDeclaration = decls[0]
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
	filegroupLoop:
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
				break filegroupLoop
			}
		}
		// Parse optional WITH clause for termination
		if p.curTok.Type == TokenWith {
			stmt.Termination = p.parseAlterDatabaseTermination()
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

func (p *Parser) parseAlterDatabaseRebuildLogStatement(dbName *ast.Identifier) (*ast.AlterDatabaseRebuildLogStatement, error) {
	p.nextToken() // consume REBUILD

	stmt := &ast.AlterDatabaseRebuildLogStatement{
		DatabaseName: dbName,
	}

	// Expect LOG
	if strings.ToUpper(p.curTok.Literal) == "LOG" {
		p.nextToken() // consume LOG
	}

	// Check for optional ON clause with file declaration
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		decls, _ := p.parseFileDeclarationList(false)
		if len(decls) > 0 {
			stmt.FileDeclaration = decls[0]
		}
	}

	p.skipToEndOfStatement()
	return stmt, nil
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

	secondary := false
	// Check for FOR SECONDARY
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "SECONDARY" {
			secondary = true
			p.nextToken() // consume SECONDARY
		}
	}

	// Check for CLEAR or SET
	action := strings.ToUpper(p.curTok.Literal)
	if action == "CLEAR" {
		return p.parseAlterDatabaseScopedConfigurationClearStatement(secondary)
	} else if action == "SET" || p.curTok.Type == TokenSet {
		return p.parseAlterDatabaseScopedConfigurationSetStatement(secondary)
	}

	// Unknown action, skip to end
	p.skipToEndOfStatement()
	return &ast.AlterDatabaseScopedConfigurationClearStatement{Secondary: secondary}, nil
}

func (p *Parser) parseAlterDatabaseScopedConfigurationClearStatement(secondary bool) (ast.Statement, error) {
	p.nextToken() // consume CLEAR

	stmt := &ast.AlterDatabaseScopedConfigurationClearStatement{
		Secondary: secondary,
	}

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
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterDatabaseScopedConfigurationSetStatement(secondary bool) (ast.Statement, error) {
	p.nextToken() // consume SET

	stmt := &ast.AlterDatabaseScopedConfigurationSetStatement{
		Secondary: secondary,
	}

	optionNameOriginal := p.curTok.Literal // preserve original case for generic options
	optionName := strings.ToUpper(optionNameOriginal)
	p.nextToken() // consume option name

	// Expect =
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
	}

	switch optionName {
	case "MAXDOP":
		// MAXDOP = N | PRIMARY
		if strings.ToUpper(p.curTok.Literal) == "PRIMARY" {
			stmt.Option = &ast.MaxDopConfigurationOption{
				OptionKind: "MaxDop",
				Primary:    true,
			}
			p.nextToken()
		} else {
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Option = &ast.MaxDopConfigurationOption{
				OptionKind: "MaxDop",
				Value:      val,
				Primary:    false,
			}
		}
	case "LEGACY_CARDINALITY_ESTIMATION":
		state := p.parseOnOffPrimaryState()
		stmt.Option = &ast.OnOffPrimaryConfigurationOption{
			OptionKind:  "LegacyCardinalityEstimate",
			OptionState: state,
		}
	case "PARAMETER_SNIFFING":
		state := p.parseOnOffPrimaryState()
		stmt.Option = &ast.OnOffPrimaryConfigurationOption{
			OptionKind:  "ParameterSniffing",
			OptionState: state,
		}
	case "QUERY_OPTIMIZER_HOTFIXES":
		state := p.parseOnOffPrimaryState()
		stmt.Option = &ast.OnOffPrimaryConfigurationOption{
			OptionKind:  "QueryOptimizerHotFixes",
			OptionState: state,
		}
	default:
		// Handle generic options (like DW_COMPATIBILITY_LEVEL)
		optionKindIdent := &ast.Identifier{
			Value:     optionNameOriginal, // use original case
			QuoteType: "NotQuoted",
		}

		var state *ast.IdentifierOrScalarExpression
		// Check if value is a number, string, negative number, or identifier
		if p.curTok.Type == TokenNumber || p.curTok.Type == TokenString || p.curTok.Type == TokenMinus {
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			state = &ast.IdentifierOrScalarExpression{
				ScalarExpression: val,
			}
		} else {
			// It's an identifier (like ON, OFF, PRIMARY, or a custom value)
			state = &ast.IdentifierOrScalarExpression{
				Identifier: p.parseIdentifier(),
			}
		}

		stmt.Option = &ast.GenericConfigurationOption{
			OptionKind:         "MaxDop", // This seems odd but matches the expected output
			GenericOptionKind:  optionKindIdent,
			GenericOptionState: state,
		}
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseOnOffPrimaryState() string {
	state := strings.ToUpper(p.curTok.Literal)
	p.nextToken()
	switch state {
	case "ON":
		return "On"
	case "OFF":
		return "Off"
	case "PRIMARY":
		return "Primary"
	default:
		return capitalizeFirst(state)
	}
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
	case "DIAGNOSTICS":
		return p.parseAlterServerConfigurationSetDiagnosticsLogStatement()
	case "FAILOVER":
		return p.parseAlterServerConfigurationSetFailoverClusterPropertyStatement()
	case "BUFFER":
		return p.parseAlterServerConfigurationSetBufferPoolExtensionStatement()
	case "HADR":
		return p.parseAlterServerConfigurationSetHadrClusterStatement()
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

func (p *Parser) parseAlterServerConfigurationSetDiagnosticsLogStatement() (*ast.AlterServerConfigurationSetDiagnosticsLogStatement, error) {
	// Consume DIAGNOSTICS
	p.nextToken()

	// Expect LOG
	if strings.ToUpper(p.curTok.Literal) != "LOG" {
		return nil, fmt.Errorf("expected LOG after DIAGNOSTICS, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterServerConfigurationSetDiagnosticsLogStatement{}

	// Parse option(s)
	optionKind := strings.ToUpper(p.curTok.Literal)

	switch optionKind {
	case "ON":
		p.nextToken()
		stmt.Options = append(stmt.Options, &ast.AlterServerConfigurationDiagnosticsLogOption{
			OptionKind:  "OnOff",
			OptionValue: &ast.OnOffOptionValue{OptionState: "On"},
		})
	case "OFF":
		p.nextToken()
		stmt.Options = append(stmt.Options, &ast.AlterServerConfigurationDiagnosticsLogOption{
			OptionKind:  "OnOff",
			OptionValue: &ast.OnOffOptionValue{OptionState: "Off"},
		})
	case "MAX_SIZE":
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		var value ast.ScalarExpression
		sizeUnit := "Unspecified"
		if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
			value = &ast.DefaultLiteral{LiteralType: "Default", Value: p.curTok.Literal}
			p.nextToken()
		} else {
			value = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
			p.nextToken()
			// Check for size unit
			unitUpper := strings.ToUpper(p.curTok.Literal)
			if unitUpper == "KB" || unitUpper == "MB" || unitUpper == "GB" {
				sizeUnit = strings.ToUpper(unitUpper)
				p.nextToken()
			}
		}
		stmt.Options = append(stmt.Options, &ast.AlterServerConfigurationDiagnosticsLogMaxSizeOption{
			OptionKind:  "MaxSize",
			OptionValue: &ast.LiteralOptionValue{Value: value},
			SizeUnit:    sizeUnit,
		})
	case "MAX_FILES":
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		var value ast.ScalarExpression
		if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
			value = &ast.DefaultLiteral{LiteralType: "Default", Value: p.curTok.Literal}
			p.nextToken()
		} else {
			value = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
			p.nextToken()
		}
		stmt.Options = append(stmt.Options, &ast.AlterServerConfigurationDiagnosticsLogOption{
			OptionKind:  "MaxFiles",
			OptionValue: &ast.LiteralOptionValue{Value: value},
		})
	case "PATH":
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		var value ast.ScalarExpression
		if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
			value = &ast.DefaultLiteral{LiteralType: "Default", Value: p.curTok.Literal}
			p.nextToken()
		} else if p.curTok.Type == TokenString {
			strVal := p.curTok.Literal
			if len(strVal) >= 2 && strVal[0] == '\'' && strVal[len(strVal)-1] == '\'' {
				strVal = strVal[1 : len(strVal)-1]
			}
			value = &ast.StringLiteral{LiteralType: "String", Value: strVal}
			p.nextToken()
		}
		stmt.Options = append(stmt.Options, &ast.AlterServerConfigurationDiagnosticsLogOption{
			OptionKind:  "Path",
			OptionValue: &ast.LiteralOptionValue{Value: value},
		})
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterServerConfigurationSetFailoverClusterPropertyStatement() (*ast.AlterServerConfigurationSetFailoverClusterPropertyStatement, error) {
	// Consume FAILOVER
	p.nextToken()

	// Expect CLUSTER
	if strings.ToUpper(p.curTok.Literal) != "CLUSTER" {
		return nil, fmt.Errorf("expected CLUSTER after FAILOVER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect PROPERTY
	if strings.ToUpper(p.curTok.Literal) != "PROPERTY" {
		return nil, fmt.Errorf("expected PROPERTY after CLUSTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterServerConfigurationSetFailoverClusterPropertyStatement{}

	// Parse property name
	propertyName := p.curTok.Literal
	propertyNameUpper := strings.ToUpper(propertyName)
	p.nextToken()

	if p.curTok.Type == TokenEquals {
		p.nextToken()
	}

	// Map property names to OptionKind values
	optionKind := propertyName
	switch propertyNameUpper {
	case "VERBOSELOGGING":
		optionKind = "VerboseLogging"
	case "SQLDUMPERDUMPFLAGS":
		optionKind = "SqlDumperDumpFlags"
	case "SQLDUMPERDUMPPATH":
		optionKind = "SqlDumperDumpPath"
	case "SQLDUMPERDUMPTIMEOUT":
		optionKind = "SqlDumperDumpTimeout"
	case "FAILURECONDITIONLEVEL":
		optionKind = "FailureConditionLevel"
	case "HEALTHCHECKTIMEOUT":
		optionKind = "HealthCheckTimeout"
	}

	var value ast.ScalarExpression
	if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
		value = &ast.DefaultLiteral{LiteralType: "Default", Value: p.curTok.Literal}
		p.nextToken()
	} else if p.curTok.Type == TokenNumber {
		value = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	} else if p.curTok.Type == TokenBinary {
		value = &ast.BinaryLiteral{LiteralType: "Binary", Value: p.curTok.Literal}
		p.nextToken()
	} else if p.curTok.Type == TokenString {
		strVal := p.curTok.Literal
		if len(strVal) >= 2 && strVal[0] == '\'' && strVal[len(strVal)-1] == '\'' {
			strVal = strVal[1 : len(strVal)-1]
		}
		value = &ast.StringLiteral{LiteralType: "String", Value: strVal}
		p.nextToken()
	}

	stmt.Options = append(stmt.Options, &ast.AlterServerConfigurationFailoverClusterPropertyOption{
		OptionKind:  optionKind,
		OptionValue: &ast.LiteralOptionValue{Value: value},
	})

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterServerConfigurationSetBufferPoolExtensionStatement() (*ast.AlterServerConfigurationSetBufferPoolExtensionStatement, error) {
	// Consume BUFFER
	p.nextToken()

	// Expect POOL
	if strings.ToUpper(p.curTok.Literal) != "POOL" {
		return nil, fmt.Errorf("expected POOL after BUFFER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect EXTENSION
	if strings.ToUpper(p.curTok.Literal) != "EXTENSION" {
		return nil, fmt.Errorf("expected EXTENSION after POOL, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterServerConfigurationSetBufferPoolExtensionStatement{}

	// Parse ON or OFF
	stateUpper := strings.ToUpper(p.curTok.Literal)
	containerOption := &ast.AlterServerConfigurationBufferPoolExtensionContainerOption{
		OptionKind: "OnOff",
	}

	if stateUpper == "ON" {
		containerOption.OptionValue = &ast.OnOffOptionValue{OptionState: "On"}
		p.nextToken()

		// Check for parentheses with suboptions
		if p.curTok.Type == TokenLParen {
			p.nextToken()

			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optionKind := strings.ToUpper(p.curTok.Literal)
				p.nextToken()

				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}

				switch optionKind {
				case "FILENAME":
					strVal := p.curTok.Literal
					if len(strVal) >= 2 && strVal[0] == '\'' && strVal[len(strVal)-1] == '\'' {
						strVal = strVal[1 : len(strVal)-1]
					}
					containerOption.Suboptions = append(containerOption.Suboptions,
						&ast.AlterServerConfigurationBufferPoolExtensionOption{
							OptionKind:  "FileName",
							OptionValue: &ast.LiteralOptionValue{Value: &ast.StringLiteral{LiteralType: "String", Value: strVal}},
						})
					p.nextToken()
				case "SIZE":
					sizeVal := p.curTok.Literal
					p.nextToken()
					// Get size unit
					sizeUnit := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					containerOption.Suboptions = append(containerOption.Suboptions,
						&ast.AlterServerConfigurationBufferPoolExtensionSizeOption{
							OptionKind:  "Size",
							OptionValue: &ast.LiteralOptionValue{Value: &ast.IntegerLiteral{LiteralType: "Integer", Value: sizeVal}},
							SizeUnit:    sizeUnit,
						})
				}

				if p.curTok.Type == TokenComma {
					p.nextToken()
				}
			}

			if p.curTok.Type == TokenRParen {
				p.nextToken()
			}
		}
	} else if stateUpper == "OFF" {
		containerOption.OptionValue = &ast.OnOffOptionValue{OptionState: "Off"}
		p.nextToken()
	}

	stmt.Options = append(stmt.Options, containerOption)

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterServerConfigurationSetHadrClusterStatement() (*ast.AlterServerConfigurationSetHadrClusterStatement, error) {
	// Consume HADR
	p.nextToken()

	// Expect CLUSTER
	if strings.ToUpper(p.curTok.Literal) != "CLUSTER" {
		return nil, fmt.Errorf("expected CLUSTER after HADR, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect CONTEXT
	if strings.ToUpper(p.curTok.Literal) != "CONTEXT" {
		return nil, fmt.Errorf("expected CONTEXT after CLUSTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterServerConfigurationSetHadrClusterStatement{}

	if p.curTok.Type == TokenEquals {
		p.nextToken()
	}

	option := &ast.AlterServerConfigurationHadrClusterOption{
		OptionKind: "Context",
	}

	if strings.ToUpper(p.curTok.Literal) == "LOCAL" {
		option.IsLocal = true
		p.nextToken()
	} else if p.curTok.Type == TokenString {
		strVal := p.curTok.Literal
		if len(strVal) >= 2 && strVal[0] == '\'' && strVal[len(strVal)-1] == '\'' {
			strVal = strVal[1 : len(strVal)-1]
		}
		option.OptionValue = &ast.LiteralOptionValue{Value: &ast.StringLiteral{LiteralType: "String", Value: strVal}}
		p.nextToken()
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

	// Check for ENABLE/DISABLE TRIGGER, FILETABLE_NAMESPACE, or CHANGE_TRACKING
	if strings.ToUpper(p.curTok.Literal) == "ENABLE" || strings.ToUpper(p.curTok.Literal) == "DISABLE" {
		// Check if it's FILETABLE_NAMESPACE
		if strings.ToUpper(p.peekTok.Literal) == "FILETABLE_NAMESPACE" {
			return p.parseAlterTableFileTableNamespaceStatement(tableName)
		}
		// Check if it's CHANGE_TRACKING
		if strings.ToUpper(p.peekTok.Literal) == "CHANGE_TRACKING" {
			return p.parseAlterTableChangeTrackingStatement(tableName)
		}
		return p.parseAlterTableTriggerModificationStatement(tableName)
	}

	// Check for SWITCH
	if strings.ToUpper(p.curTok.Literal) == "SWITCH" {
		return p.parseAlterTableSwitchStatement(tableName)
	}

	// Check for WITH CHECK/NOCHECK ADD - this is an add with enforcement option
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "CHECK" || strings.ToUpper(p.curTok.Literal) == "NOCHECK" {
			enforcement := "Check"
			if strings.ToUpper(p.curTok.Literal) == "NOCHECK" {
				enforcement = "NoCheck"
			}
			p.nextToken() // consume CHECK/NOCHECK
			if strings.ToUpper(p.curTok.Literal) == "ADD" {
				stmt, err := p.parseAlterTableAddStatement(tableName)
				if err != nil {
					return nil, err
				}
				stmt.ExistingRowsCheckEnforcement = enforcement
				return stmt, nil
			}
			// It's a constraint modification statement (WITH CHECK/NOCHECK CONSTRAINT ...)
			return p.parseAlterTableConstraintModificationStatementAfterWith(tableName, enforcement)
		}
	}

	// Check for CHECK/NOCHECK CONSTRAINT
	if strings.ToUpper(p.curTok.Literal) == "CHECK" || strings.ToUpper(p.curTok.Literal) == "NOCHECK" {
		return p.parseAlterTableConstraintModificationStatement(tableName)
	}

	// Check for SET
	if strings.ToUpper(p.curTok.Literal) == "SET" {
		return p.parseAlterTableSetStatement(tableName)
	}

	// Check for REBUILD
	if strings.ToUpper(p.curTok.Literal) == "REBUILD" {
		return p.parseAlterTableRebuildStatement(tableName)
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

		// Check for IF EXISTS
		var isIfExists bool
		if strings.ToUpper(p.curTok.Literal) == "IF" {
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) != "EXISTS" {
				return nil, fmt.Errorf("expected EXISTS after IF")
			}
			p.nextToken()
			isIfExists = true
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
			IsIfExists:       isIfExists,
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

		case "WAIT_AT_LOW_PRIORITY":
			waitOpt, err := p.parseWaitAtLowPriorityOption()
			if err != nil {
				return nil, err
			}
			options = append(options, waitOpt)

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

func (p *Parser) parseWaitAtLowPriorityOption() (*ast.DropClusteredConstraintWaitAtLowPriorityLockOption, error) {
	// Consume WAIT_AT_LOW_PRIORITY
	p.nextToken()

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after WAIT_AT_LOW_PRIORITY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	opt := &ast.DropClusteredConstraintWaitAtLowPriorityLockOption{
		OptionKind: "MaxDop", // This seems to be the expected value based on test data
	}

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		optionName := strings.ToUpper(p.curTok.Literal)

		switch optionName {
		case "MAX_DURATION":
			p.nextToken() // consume MAX_DURATION
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after MAX_DURATION, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =

			maxDuration := &ast.LowPriorityLockWaitMaxDurationOption{
				OptionKind: "MaxDuration",
			}
			if p.curTok.Type != TokenNumber {
				return nil, fmt.Errorf("expected number after MAX_DURATION =, got %s", p.curTok.Literal)
			}
			maxDuration.MaxDuration = &ast.IntegerLiteral{
				LiteralType: "Integer",
				Value:       p.curTok.Literal,
			}
			p.nextToken() // consume number

			// Parse optional unit (MINUTES or SECONDS)
			unit := strings.ToUpper(p.curTok.Literal)
			if unit == "MINUTES" {
				maxDuration.Unit = "Minutes"
				p.nextToken() // consume unit
			} else if unit == "SECONDS" {
				maxDuration.Unit = "Seconds"
				p.nextToken() // consume unit
			}
			// If no unit is specified, leave Unit empty

			opt.Options = append(opt.Options, maxDuration)

		case "ABORT_AFTER_WAIT":
			p.nextToken() // consume ABORT_AFTER_WAIT
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after ABORT_AFTER_WAIT, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =

			abortOpt := &ast.LowPriorityLockWaitAbortAfterWaitOption{
				OptionKind: "AbortAfterWait",
			}
			abortValue := strings.ToUpper(p.curTok.Literal)
			switch abortValue {
			case "NONE":
				abortOpt.AbortAfterWait = "None"
			case "SELF":
				abortOpt.AbortAfterWait = "Self"
			case "BLOCKERS":
				abortOpt.AbortAfterWait = "Blockers"
			default:
				return nil, fmt.Errorf("expected NONE, SELF, or BLOCKERS after ABORT_AFTER_WAIT =, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume abort value

			opt.Options = append(opt.Options, abortOpt)

		default:
			return nil, fmt.Errorf("unexpected option in WAIT_AT_LOW_PRIORITY: %s", p.curTok.Literal)
		}

		// Check for comma
		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) to close WAIT_AT_LOW_PRIORITY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return opt, nil
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
		"BUCKET_COUNT":                 "BucketCount",
		"PAD_INDEX":                    "PadIndex",
		"FILLFACTOR":                   "FillFactor",
		"SORT_IN_TEMPDB":               "SortInTempDB",
		"IGNORE_DUP_KEY":               "IgnoreDupKey",
		"STATISTICS_NORECOMPUTE":       "StatisticsNoRecompute",
		"DROP_EXISTING":                "DropExisting",
		"ONLINE":                       "Online",
		"ALLOW_ROW_LOCKS":              "AllowRowLocks",
		"ALLOW_PAGE_LOCKS":             "AllowPageLocks",
		"MAXDOP":                       "MaxDop",
		"DATA_COMPRESSION":             "DataCompression",
		"COMPRESS_ALL_ROW_GROUPS":      "CompressAllRowGroups",
		"COMPRESSION_DELAY":            "CompressionDelay",
		"OPTIMIZE_FOR_SEQUENTIAL_KEY": "OptimizeForSequentialKey",
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

	// Check for ADD/DROP ROWGUIDCOL or ADD/DROP NOT FOR REPLICATION or ADD/DROP PERSISTED or ADD/DROP SPARSE
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "ADD" {
		p.nextToken() // consume ADD
		nextLit := strings.ToUpper(p.curTok.Literal)
		if nextLit == "ROWGUIDCOL" {
			stmt.AlterTableAlterColumnOption = "AddRowGuidCol"
			p.nextToken()
		} else if nextLit == "PERSISTED" {
			stmt.AlterTableAlterColumnOption = "AddPersisted"
			p.nextToken()
		} else if nextLit == "SPARSE" {
			stmt.AlterTableAlterColumnOption = "AddSparse"
			p.nextToken()
		} else if nextLit == "HIDDEN" {
			stmt.AlterTableAlterColumnOption = "AddHidden"
			p.nextToken()
		} else if nextLit == "NOT" {
			p.nextToken() // consume NOT
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
			}
			if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
				p.nextToken() // consume REPLICATION
			}
			stmt.AlterTableAlterColumnOption = "AddNotForReplication"
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	} else if upperLit == "DROP" {
		p.nextToken() // consume DROP
		nextLit := strings.ToUpper(p.curTok.Literal)
		if nextLit == "ROWGUIDCOL" {
			stmt.AlterTableAlterColumnOption = "DropRowGuidCol"
			p.nextToken()
		} else if nextLit == "PERSISTED" {
			stmt.AlterTableAlterColumnOption = "DropPersisted"
			p.nextToken()
		} else if nextLit == "SPARSE" {
			stmt.AlterTableAlterColumnOption = "DropSparse"
			p.nextToken()
		} else if nextLit == "HIDDEN" {
			stmt.AlterTableAlterColumnOption = "DropHidden"
			p.nextToken()
		} else if nextLit == "NOT" {
			p.nextToken() // consume NOT
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
			}
			if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
				p.nextToken() // consume REPLICATION
			}
			stmt.AlterTableAlterColumnOption = "DropNotForReplication"
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	// Parse data type - be lenient if no data type is provided
	// Use parseDataTypeReference to handle all data types including XML
	dataType, err := p.parseDataTypeReference()
	if err != nil {
		// Lenient: return statement without data type
		p.skipToEndOfStatement()
		return stmt, nil
	}
	stmt.DataType = dataType

	// Check for COLLATE
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		stmt.Collation = p.parseIdentifier()
	}

	// Parse optional SPARSE, FILESTREAM, COLUMN_SET FOR ALL_SPARSE_COLUMNS, HIDDEN, ENCRYPTED WITH, MASKED WITH
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "SPARSE" {
			if stmt.StorageOptions == nil {
				stmt.StorageOptions = &ast.ColumnStorageOptions{}
			}
			stmt.StorageOptions.SparseOption = "Sparse"
			p.nextToken()
		} else if upperLit == "FILESTREAM" {
			if stmt.StorageOptions == nil {
				stmt.StorageOptions = &ast.ColumnStorageOptions{}
			}
			stmt.StorageOptions.IsFileStream = true
			p.nextToken()
		} else if upperLit == "COLUMN_SET" {
			p.nextToken() // consume COLUMN_SET
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
			}
			if strings.ToUpper(p.curTok.Literal) == "ALL_SPARSE_COLUMNS" {
				p.nextToken() // consume ALL_SPARSE_COLUMNS
			}
			if stmt.StorageOptions == nil {
				stmt.StorageOptions = &ast.ColumnStorageOptions{}
			}
			stmt.StorageOptions.SparseOption = "ColumnSetForAllSparseColumns"
		} else if upperLit == "HIDDEN" {
			stmt.IsHidden = true
			p.nextToken()
		} else if upperLit == "ENCRYPTED" {
			p.nextToken() // consume ENCRYPTED
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
			}
			// Parse encryption specification: (COLUMN_ENCRYPTION_KEY = key1, ENCRYPTION_TYPE = ..., ALGORITHM = ...)
			if p.curTok.Type == TokenLParen {
				encSpec, err := p.parseColumnEncryptionSpecification()
				if err != nil {
					return nil, err
				}
				stmt.Encryption = encSpec
			}
		} else if upperLit == "MASKED" {
			stmt.IsMasked = true
			p.nextToken() // consume MASKED
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
			}
			// Parse masking function: (function = 'default()')
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				if strings.ToUpper(p.curTok.Literal) == "FUNCTION" {
					p.nextToken() // consume FUNCTION
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					if p.curTok.Type == TokenString {
						maskFunc, err := p.parseStringLiteral()
						if err != nil {
							return nil, err
						}
						stmt.MaskingFunction = maskFunc
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
		} else {
			break
		}
	}

	// Check for NULL/NOT NULL
	if strings.ToUpper(p.curTok.Literal) == "NULL" {
		stmt.AlterTableAlterColumnOption = "Null"
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NOT" {
		p.nextToken() // consume NOT
		if strings.ToUpper(p.curTok.Literal) == "NULL" {
			stmt.AlterTableAlterColumnOption = "NotNull"
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseColumnEncryptionSpecification() (*ast.ColumnEncryptionDefinition, error) {
	// curTok should be (
	p.nextToken() // consume (

	encDef := &ast.ColumnEncryptionDefinition{}

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		paramName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		switch paramName {
		case "COLUMN_ENCRYPTION_KEY":
			param := &ast.ColumnEncryptionKeyNameParameter{
				ParameterKind: "ColumnEncryptionKey",
				Name:          p.parseIdentifier(),
			}
			encDef.Parameters = append(encDef.Parameters, param)
		case "ENCRYPTION_TYPE":
			encType := strings.ToUpper(p.curTok.Literal)
			param := &ast.ColumnEncryptionTypeParameter{
				ParameterKind: "EncryptionType",
			}
			if encType == "DETERMINISTIC" {
				param.EncryptionType = "Deterministic"
			} else if encType == "RANDOMIZED" {
				param.EncryptionType = "Randomized"
			} else {
				param.EncryptionType = encType
			}
			p.nextToken()
			encDef.Parameters = append(encDef.Parameters, param)
		case "ALGORITHM":
			str, err := p.parseStringLiteral()
			if err != nil {
				return nil, err
			}
			param := &ast.ColumnEncryptionAlgorithmParameter{
				ParameterKind:       "Algorithm",
				EncryptionAlgorithm: str,
			}
			encDef.Parameters = append(encDef.Parameters, param)
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return encDef, nil
}

func (p *Parser) parseAlterTableAddStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableAddTableElementStatement, error) {
	// Consume ADD
	p.nextToken()

	stmt := &ast.AlterTableAddTableElementStatement{
		SchemaObjectName:             tableName,
		ExistingRowsCheckEnforcement: "NotSpecified",
		Definition:                   &ast.TableDefinition{},
	}

	// Loop to parse multiple elements separated by commas
	for {
		// Check if this is ADD CONSTRAINT
		if strings.ToUpper(p.curTok.Literal) == "CONSTRAINT" {
		p.nextToken() // consume CONSTRAINT
		// Parse constraint name
		constraintName := p.parseIdentifier()

		// Check what type of constraint follows
		upperLit := strings.ToUpper(p.curTok.Literal)

		switch upperLit {
		case "PRIMARY":
			p.nextToken() // consume PRIMARY
			if p.curTok.Type == TokenKey {
				p.nextToken() // consume KEY
			}
			constraint := &ast.UniqueConstraintDefinition{
				ConstraintIdentifier: constraintName,
				IsPrimaryKey:         true,
			}
			// Parse optional CLUSTERED/NONCLUSTERED/HASH
			for {
				upperOpt := strings.ToUpper(p.curTok.Literal)
				if upperOpt == "CLUSTERED" {
					constraint.Clustered = true
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
					p.nextToken()
				} else if upperOpt == "NONCLUSTERED" {
					constraint.Clustered = false
					p.nextToken()
					// Check for HASH suffix
					if strings.ToUpper(p.curTok.Literal) == "HASH" {
						constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
						p.nextToken()
					} else {
						constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
					}
				} else if upperOpt == "HASH" {
					// HASH without NONCLUSTERED
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
					p.nextToken()
				} else {
					break
				}
			}
		// Parse column list - only add constraint if we have a column list
			hasColumnsPK := false
			if p.curTok.Type == TokenLParen {
				hasColumnsPK = true
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					colRef := &ast.ColumnReferenceExpression{
						ColumnType: "Regular",
					}
					colName := p.parseIdentifier()
					colRef.MultiPartIdentifier = &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{colName},
						Count:       1,
					}
					sortOrder := ast.SortOrderNotSpecified
					upperSort := strings.ToUpper(p.curTok.Literal)
					if upperSort == "ASC" {
						sortOrder = ast.SortOrderAscending
						p.nextToken()
					} else if upperSort == "DESC" {
						sortOrder = ast.SortOrderDescending
						p.nextToken()
					}
					constraint.Columns = append(constraint.Columns, &ast.ColumnWithSortOrder{
						Column:    colRef,
						SortOrder: sortOrder,
					})
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
			// Parse WITH (index_options) or WITH option = value (no parentheses)
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				hasParen := p.curTok.Type == TokenLParen
				if hasParen {
					p.nextToken() // consume (
				}
				for {
					// Check for end conditions
					if hasParen && p.curTok.Type == TokenRParen {
						break
					}
					if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon {
						break
					}
					// Check for ON (filegroup) which means we're done with options
					if p.curTok.Type == TokenOn {
						break
					}
					// Check for keywords that start new constraints
					upperLiteral := strings.ToUpper(p.curTok.Literal)
					if upperLiteral == "CONSTRAINT" || upperLiteral == "PRIMARY" || upperLiteral == "UNIQUE" ||
						upperLiteral == "FOREIGN" || upperLiteral == "CHECK" || upperLiteral == "DEFAULT" ||
						upperLiteral == "INDEX" {
						break
					}

					optionName := upperLiteral
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					// Check for ON/OFF state options
					valueUpper := strings.ToUpper(p.curTok.Literal)
					if valueUpper == "ON" || valueUpper == "OFF" || p.curTok.Type == TokenOn {
						state := "On"
						if valueUpper == "OFF" {
							state = "Off"
						}
						p.nextToken() // consume ON/OFF
						option := &ast.IndexStateOption{
							OptionKind:  convertIndexOptionKind(optionName),
							OptionState: state,
						}
						constraint.IndexOptions = append(constraint.IndexOptions, option)
					} else {
						expr, _ := p.parseScalarExpression()
						option := &ast.IndexExpressionOption{
							OptionKind: convertIndexOptionKind(optionName),
							Expression: expr,
						}
						constraint.IndexOptions = append(constraint.IndexOptions, option)
					}
					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}
				if hasParen && p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}
			// Parse ON filegroup
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				fgIdent := p.parseIdentifier()
				fg := &ast.FileGroupOrPartitionScheme{
					Name: &ast.IdentifierOrValueExpression{
						Identifier: fgIdent,
						Value:      fgIdent.Value,
					},
				}
				constraint.OnFileGroupOrPartitionScheme = fg
			}
			// Parse NOT ENFORCED
			if strings.ToUpper(p.curTok.Literal) == "NOT" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "ENFORCED" {
					p.nextToken()
					f := false
					constraint.IsEnforced = &f
				}
			}
			// Only add constraint if we successfully parsed a column list
			if hasColumnsPK {
				stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
			}

		case "UNIQUE":
			p.nextToken() // consume UNIQUE
			constraint := &ast.UniqueConstraintDefinition{
				ConstraintIdentifier: constraintName,
				IsPrimaryKey:         false,
			}
			// Parse optional CLUSTERED/NONCLUSTERED/HASH
			for {
				upperOpt := strings.ToUpper(p.curTok.Literal)
				if upperOpt == "CLUSTERED" {
					constraint.Clustered = true
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
					p.nextToken()
				} else if upperOpt == "NONCLUSTERED" {
					constraint.Clustered = false
					p.nextToken()
					// Check for HASH suffix
					if strings.ToUpper(p.curTok.Literal) == "HASH" {
						constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
						p.nextToken()
					} else {
						constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
					}
				} else if upperOpt == "HASH" {
					// HASH without NONCLUSTERED
					constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
					p.nextToken()
				} else {
					break
				}
			}
			// Parse column list - only add constraint if we have a column list
			hasColumnsUQ := false
			if p.curTok.Type == TokenLParen {
				hasColumnsUQ = true
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					colRef := &ast.ColumnReferenceExpression{
						ColumnType: "Regular",
					}
					colName := p.parseIdentifier()
					colRef.MultiPartIdentifier = &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{colName},
						Count:       1,
					}
					sortOrder := ast.SortOrderNotSpecified
					upperSort := strings.ToUpper(p.curTok.Literal)
					if upperSort == "ASC" {
						sortOrder = ast.SortOrderAscending
						p.nextToken()
					} else if upperSort == "DESC" {
						sortOrder = ast.SortOrderDescending
						p.nextToken()
					}
					constraint.Columns = append(constraint.Columns, &ast.ColumnWithSortOrder{
						Column:    colRef,
						SortOrder: sortOrder,
					})
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
			// Parse WITH (index_options) or WITH option = value (no parentheses)
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				hasParen := p.curTok.Type == TokenLParen
				if hasParen {
					p.nextToken() // consume (
				}
				for {
					// Check for end conditions
					if hasParen && p.curTok.Type == TokenRParen {
						break
					}
					if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon {
						break
					}
					// Check for ON (filegroup) which means we're done with options
					if p.curTok.Type == TokenOn {
						break
					}
					// Check for keywords that start new constraints
					upperLiteral := strings.ToUpper(p.curTok.Literal)
					if upperLiteral == "CONSTRAINT" || upperLiteral == "PRIMARY" || upperLiteral == "UNIQUE" ||
						upperLiteral == "FOREIGN" || upperLiteral == "CHECK" || upperLiteral == "DEFAULT" ||
						upperLiteral == "INDEX" {
						break
					}

					optionName := upperLiteral
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					// Check for ON/OFF state options
					valueUpper := strings.ToUpper(p.curTok.Literal)
					if valueUpper == "ON" || valueUpper == "OFF" || p.curTok.Type == TokenOn {
						state := "On"
						if valueUpper == "OFF" {
							state = "Off"
						}
						p.nextToken() // consume ON/OFF
						option := &ast.IndexStateOption{
							OptionKind:  convertIndexOptionKind(optionName),
							OptionState: state,
						}
						constraint.IndexOptions = append(constraint.IndexOptions, option)
					} else {
						expr, _ := p.parseScalarExpression()
						option := &ast.IndexExpressionOption{
							OptionKind: convertIndexOptionKind(optionName),
							Expression: expr,
						}
						constraint.IndexOptions = append(constraint.IndexOptions, option)
					}
					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}
				if hasParen && p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}
			// Parse ON filegroup
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				fgIdent := p.parseIdentifier()
				fg := &ast.FileGroupOrPartitionScheme{
					Name: &ast.IdentifierOrValueExpression{
						Identifier: fgIdent,
						Value:      fgIdent.Value,
					},
				}
				constraint.OnFileGroupOrPartitionScheme = fg
			}
			// Parse NOT ENFORCED
			if strings.ToUpper(p.curTok.Literal) == "NOT" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "ENFORCED" {
					p.nextToken()
					f := false
					constraint.IsEnforced = &f
				}
			}
			// Only add constraint if we successfully parsed a column list
			if hasColumnsUQ {
				stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
			}

		case "FOREIGN":
			p.nextToken() // consume FOREIGN
			if p.curTok.Type == TokenKey {
				p.nextToken() // consume KEY
			}
			constraint := &ast.ForeignKeyConstraintDefinition{
				ConstraintIdentifier: constraintName,
			}
			// Parse optional column list
			hasReferences := false
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
				hasReferences = true
				p.nextToken()
				refName, err := p.parseSchemaObjectName()
				if err != nil {
					return nil, err
				}
				constraint.ReferenceTableName = refName
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
			// Parse ON DELETE/ON UPDATE actions
			for p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				actionType := strings.ToUpper(p.curTok.Literal)
				if actionType == "DELETE" || actionType == "UPDATE" {
					p.nextToken() // consume DELETE/UPDATE
					action := ""
					upperAction := strings.ToUpper(p.curTok.Literal)
					if upperAction == "CASCADE" {
						action = "Cascade"
						p.nextToken()
					} else if upperAction == "SET" {
						p.nextToken()
						if strings.ToUpper(p.curTok.Literal) == "NULL" {
							action = "SetNull"
							p.nextToken()
						} else if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
							action = "SetDefault"
							p.nextToken()
						}
					} else if upperAction == "NO" {
						p.nextToken()
						if strings.ToUpper(p.curTok.Literal) == "ACTION" {
							action = "NoAction"
							p.nextToken()
						}
					}
					if actionType == "DELETE" {
						constraint.DeleteAction = action
					} else {
						constraint.UpdateAction = action
					}
				} else {
					break
				}
			}
			// Parse NOT FOR REPLICATION
			if strings.ToUpper(p.curTok.Literal) == "NOT" &&
				strings.ToUpper(p.peekTok.Literal) == "FOR" {
				p.nextToken() // consume NOT
				p.nextToken() // consume FOR
				if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
					constraint.NotForReplication = true
					p.nextToken()
				}
			}
			// Parse NOT ENFORCED
			if strings.ToUpper(p.curTok.Literal) == "NOT" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "ENFORCED" {
					p.nextToken()
					f := false
					constraint.IsEnforced = &f
				}
			}
			// Only add constraint if we have references (columns are optional)
			if hasReferences {
				stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
			}

		case "CONNECTION":
			// Parse CONNECTION (node1 TO node2, ...)
			p.nextToken() // consume CONNECTION
			constraint := &ast.GraphConnectionConstraintDefinition{
				ConstraintIdentifier: constraintName,
			}
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
			stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)

		case "DEFAULT":
			// CONSTRAINT name DEFAULT value FOR column [WITH VALUES]
			p.nextToken() // consume DEFAULT
			constraint := &ast.DefaultConstraintDefinition{
				ConstraintIdentifier: constraintName,
			}
			expr, err := p.parseScalarExpression()
			if err != nil {
				// Invalid/incomplete DEFAULT constraint - skip it
				break
			}
			constraint.Expression = expr
			// Check for FOR column
			if strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
				constraint.Column = p.parseIdentifier()
			}
			// Check for WITH VALUES
			if p.curTok.Type == TokenWith && strings.ToUpper(p.peekTok.Literal) == "VALUES" {
				p.nextToken() // consume WITH
				p.nextToken() // consume VALUES
				constraint.WithValues = true
			}
			stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)

		case "CHECK":
			// CONSTRAINT name CHECK (condition)
			p.nextToken() // consume CHECK
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				expr, err := p.parseBooleanExpression()
				if err != nil {
					return nil, err
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
				constraint := &ast.CheckConstraintDefinition{
					ConstraintIdentifier: constraintName,
					CheckCondition:       expr,
				}
				stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
			}

		default:
			// Unknown constraint type - skip to end of statement
			p.skipToEndOfStatement()
		}
		// Continue to check for comma (don't return here - allow multiple constraints)
	} else if p.curTok.Type == TokenIndex {
		// ADD INDEX
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

					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}

					// Check for ON/OFF state options
					valueUpper := strings.ToUpper(p.curTok.Literal)
					if valueUpper == "ON" || valueUpper == "OFF" || p.curTok.Type == TokenOn {
						state := "On"
						if valueUpper == "OFF" {
							state = "Off"
						}
						p.nextToken() // consume ON/OFF
						option := &ast.IndexStateOption{
							OptionKind:  convertIndexOptionKind(optionName),
							OptionState: state,
						}
						indexDef.IndexOptions = append(indexDef.IndexOptions, option)
					} else {
						// Parse expression option value
						expr, err := p.parseScalarExpression()
						if err != nil {
							// Skip on error
							if p.curTok.Type == TokenComma {
								p.nextToken()
							}
							continue
						}
						option := &ast.IndexExpressionOption{
							OptionKind: convertIndexOptionKind(optionName),
							Expression: expr,
						}
						indexDef.IndexOptions = append(indexDef.IndexOptions, option)
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
			}
		}

		stmt.Definition.Indexes = append(stmt.Definition.Indexes, indexDef)
		} else if strings.ToUpper(p.curTok.Literal) == "CHECK" {
			// Table-level CHECK constraint without CONSTRAINT keyword
			p.nextToken() // consume CHECK
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				expr, err := p.parseBooleanExpression()
				if err != nil {
					p.skipToEndOfStatement()
				} else {
					if p.curTok.Type == TokenRParen {
						p.nextToken()
					}
					constraint := &ast.CheckConstraintDefinition{
						CheckCondition: expr,
					}
					stmt.Definition.TableConstraints = append(stmt.Definition.TableConstraints, constraint)
				}
			}
		} else {
			// Parse column definition (column_name data_type ...)
			colDef, err := p.parseColumnDefinition()
			if err != nil {
				return nil, err
			}
			stmt.Definition.ColumnDefinitions = append(stmt.Definition.ColumnDefinitions, colDef)
		}

		// Check for comma to continue parsing more elements
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

func (p *Parser) parseAlterTableConstraintModificationStatementAfterWith(tableName *ast.SchemaObjectName, existingEnforcement string) (*ast.AlterTableConstraintModificationStatement, error) {
	stmt := &ast.AlterTableConstraintModificationStatement{
		SchemaObjectName:             tableName,
		ExistingRowsCheckEnforcement: existingEnforcement,
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

func (p *Parser) parseAlterTableFileTableNamespaceStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableFileTableNamespaceStatement, error) {
	stmt := &ast.AlterTableFileTableNamespaceStatement{
		SchemaObjectName: tableName,
	}

	// Parse ENABLE or DISABLE
	if strings.ToUpper(p.curTok.Literal) == "ENABLE" {
		stmt.IsEnable = true
	} else {
		stmt.IsEnable = false
	}
	p.nextToken()

	// Consume FILETABLE_NAMESPACE
	if strings.ToUpper(p.curTok.Literal) != "FILETABLE_NAMESPACE" {
		return nil, fmt.Errorf("expected FILETABLE_NAMESPACE, got %s", p.curTok.Literal)
	}
	p.nextToken()

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
				} else if optionName == "WAIT_AT_LOW_PRIORITY" {
					opt := &ast.LowPriorityLockWaitTableSwitchOption{
						OptionKind: "LowPriorityLockWait",
					}

					// Expect (
					if p.curTok.Type == TokenLParen {
						p.nextToken()

						for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
							subOptName := strings.ToUpper(p.curTok.Literal)
							p.nextToken()

							if subOptName == "MAX_DURATION" {
								if p.curTok.Type == TokenEquals {
									p.nextToken()
								}
								// Parse the duration value
								durExpr, err := p.parseScalarExpression()
								if err != nil {
									return nil, err
								}
								subOpt := &ast.LowPriorityLockWaitMaxDurationOption{
									OptionKind:  "MaxDuration",
									MaxDuration: durExpr,
								}
								// Check for MINUTES
								if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
									subOpt.Unit = "Minutes"
									p.nextToken()
								}
								opt.Options = append(opt.Options, subOpt)
							} else if subOptName == "ABORT_AFTER_WAIT" {
								if p.curTok.Type == TokenEquals {
									p.nextToken()
								}
								value := p.curTok.Literal
								p.nextToken()
								// Convert to proper case
								abortValue := "None"
								switch strings.ToUpper(value) {
								case "NONE":
									abortValue = "None"
								case "SELF":
									abortValue = "Self"
								case "BLOCKERS":
									abortValue = "Blockers"
								}
								subOpt := &ast.LowPriorityLockWaitAbortAfterWaitOption{
									OptionKind:     "AbortAfterWait",
									AbortAfterWait: abortValue,
								}
								opt.Options = append(opt.Options, subOpt)
							}

							if p.curTok.Type == TokenComma {
								p.nextToken()
							}
						}

						if p.curTok.Type == TokenRParen {
							p.nextToken()
						}
					}

					stmt.Options = append(stmt.Options, opt)
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
		} else if optionName == "FILETABLE_DIRECTORY" {
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			// Parse the directory name as a literal or NULL
			opt := &ast.FileTableDirectoryTableOption{
				OptionKind: "FileTableDirectory",
			}
			if strings.ToUpper(p.curTok.Literal) == "NULL" {
				opt.Value = &ast.NullLiteral{
					LiteralType: "Null",
					Value:       "NULL",
				}
				p.nextToken()
			} else if p.curTok.Type == TokenString {
				value := p.curTok.Literal
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				opt.Value = &ast.StringLiteral{
					LiteralType:   "String",
					Value:         value,
					IsNational:    false,
					IsLargeObject: false,
				}
				p.nextToken()
			} else {
				value := p.curTok.Literal
				opt.Value = &ast.StringLiteral{
					LiteralType:   "String",
					Value:         value,
					IsNational:    false,
					IsLargeObject: false,
				}
				p.nextToken()
			}
			stmt.Options = append(stmt.Options, opt)
		} else if optionName == "LOCK_ESCALATION" {
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			valueUpper := strings.ToUpper(p.curTok.Literal)
			value := "Auto"
			if valueUpper == "TABLE" {
				value = "Table"
			} else if valueUpper == "DISABLE" {
				value = "Disable"
			}
			p.nextToken()
			stmt.Options = append(stmt.Options, &ast.LockEscalationTableOption{
				OptionKind: "LockEscalation",
				Value:      value,
			})
		} else if optionName == "FILESTREAM_ON" {
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			opt := &ast.FileStreamOnTableOption{
				OptionKind: "FileStreamOn",
			}
			if p.curTok.Type == TokenString {
				value := p.curTok.Literal
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				opt.Value = &ast.IdentifierOrValueExpression{
					Value: value,
					ValueExpression: &ast.StringLiteral{
						LiteralType:   "String",
						Value:         value,
						IsNational:    false,
						IsLargeObject: false,
					},
				}
				p.nextToken()
			} else {
				value := p.curTok.Literal
				opt.Value = &ast.IdentifierOrValueExpression{
					Value: value,
					Identifier: &ast.Identifier{
						Value:     value,
						QuoteType: "NotQuoted",
					},
				}
				p.nextToken()
			}
			stmt.Options = append(stmt.Options, opt)
		} else if optionName == "REMOTE_DATA_ARCHIVE" {
			rdaOpt, err := p.parseRemoteDataArchiveTableOption(true)
			if err != nil {
				return nil, err
			}
			stmt.Options = append(stmt.Options, rdaOpt)
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

func (p *Parser) parseAlterViewStatement() (*ast.AlterViewStatement, error) {
	// Consume VIEW
	p.nextToken()

	stmt := &ast.AlterViewStatement{}

	// Parse view name
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.SchemaObjectName = son

	// Check for column list
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			stmt.Columns = append(stmt.Columns, p.parseIdentifier())
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Check for WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken()
		// Parse view options (can be identifiers or keywords like ENCRYPTION)
		for p.curTok.Type != TokenAs && p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
			optName := strings.ToUpper(p.curTok.Literal)
			var optionKind string
			switch optName {
			case "ENCRYPTION":
				optionKind = "Encryption"
			case "SCHEMABINDING":
				optionKind = "SchemaBinding"
			case "VIEW_METADATA":
				optionKind = "ViewMetadata"
			default:
				optionKind = p.curTok.Literal
			}
			opt := &ast.ViewStatementOption{OptionKind: optionKind}
			stmt.ViewOptions = append(stmt.ViewOptions, opt)
			p.nextToken()
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
	}

	// Expect AS
	if p.curTok.Type != TokenAs {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse SELECT statement
	selStmt, err := p.parseSelectStatement()
	if err != nil {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	stmt.SelectStatement = selStmt

	// Check for WITH CHECK OPTION
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "CHECK" {
			p.nextToken()
			if p.curTok.Type == TokenOption {
				p.nextToken()
				stmt.WithCheckOption = true
			}
		}
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

func (p *Parser) parseAlterServerAuditStatement() (ast.Statement, error) {
	// AUDIT keyword should be current token, consume it
	p.nextToken()

	// Check if this is ALTER SERVER AUDIT SPECIFICATION
	if strings.ToUpper(p.curTok.Literal) == "SPECIFICATION" {
		return p.parseAlterServerAuditSpecificationStatement()
	}

	stmt := &ast.AlterServerAuditStatement{}

	// Parse audit name
	stmt.AuditName = p.parseIdentifier()

	// Check for MODIFY NAME
	if strings.ToUpper(p.curTok.Literal) == "MODIFY" {
		p.nextToken() // consume MODIFY
		if strings.ToUpper(p.curTok.Literal) == "NAME" {
			p.nextToken() // consume NAME
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			stmt.NewName = p.parseIdentifier()
			// Skip optional semicolon
			if p.curTok.Type == TokenSemicolon {
				p.nextToken()
			}
			return stmt, nil
		}
		return nil, fmt.Errorf("expected NAME after MODIFY, got %s", p.curTok.Literal)
	}

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

func (p *Parser) parseAlterLoginStatement() (ast.Statement, error) {
	// Consume LOGIN
	p.nextToken()

	// Parse login name
	name := p.parseIdentifier()

	// Check for ENABLE/DISABLE
	upper := strings.ToUpper(p.curTok.Literal)
	if upper == "ENABLE" {
		p.nextToken()
		return &ast.AlterLoginEnableDisableStatement{
			Name:     name,
			IsEnable: true,
		}, nil
	} else if upper == "DISABLE" {
		p.nextToken()
		return &ast.AlterLoginEnableDisableStatement{
			Name:     name,
			IsEnable: false,
		}, nil
	}

	// Check for ADD or DROP CREDENTIAL
	if p.curTok.Type == TokenAdd || p.curTok.Type == TokenDrop {
		stmt := &ast.AlterLoginAddDropCredentialStatement{
			Name:  name,
			IsAdd: p.curTok.Type == TokenAdd,
		}
		p.nextToken() // consume ADD/DROP

		// Expect CREDENTIAL
		if p.curTok.Type == TokenCredential {
			p.nextToken()
			stmt.CredentialName = p.parseIdentifier()
		}

		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	// Handle WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		// Check if we have valid options to parse
		upper := strings.ToUpper(p.curTok.Literal)
		if upper == "PASSWORD" || upper == "NO" || upper == "NAME" || upper == "DEFAULT_DATABASE" ||
			upper == "DEFAULT_LANGUAGE" || upper == "CHECK_POLICY" || upper == "CHECK_EXPIRATION" || upper == "CREDENTIAL" {
			return p.parseAlterLoginOptions(name)
		}
		// For incomplete statements like "alter login l1 with", fall back to old behavior
		p.skipToEndOfStatement()
		return &ast.AlterLoginAddDropCredentialStatement{Name: name, IsAdd: false}, nil
	}

	// Skip to end if we don't recognize the syntax
	p.skipToEndOfStatement()
	return &ast.AlterLoginAddDropCredentialStatement{Name: name, IsAdd: false}, nil
}

func (p *Parser) parseAlterLoginOptions(name *ast.Identifier) (*ast.AlterLoginOptionsStatement, error) {
	stmt := &ast.AlterLoginOptionsStatement{
		Name: name,
	}

	for {
		optName := strings.ToUpper(p.curTok.Literal)

		if optName == "PASSWORD" {
			p.nextToken() // consume PASSWORD
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			opt := &ast.PasswordAlterPrincipalOption{
				OptionKind: "Password",
			}

			// Parse password value
			if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
				val := p.curTok.Literal
				isNational := p.curTok.Type == TokenNationalString
				// Strip N prefix for national strings
				if isNational && len(val) > 0 && (val[0] == 'N' || val[0] == 'n') {
					val = val[1:]
				}
				// Strip surrounding quotes
				if len(val) >= 2 && val[0] == '\'' && val[len(val)-1] == '\'' {
					val = val[1 : len(val)-1]
				}
				opt.Password = &ast.StringLiteral{
					LiteralType:   "String",
					Value:         val,
					IsNational:    isNational,
					IsLargeObject: false,
				}
				p.nextToken()
			} else if p.curTok.Type == TokenBinary {
				opt.Password = &ast.BinaryLiteral{
					LiteralType:   "Binary",
					IsLargeObject: false,
					Value:         p.curTok.Literal,
				}
				p.nextToken()
			}

			// Parse optional flags
			for {
				upper := strings.ToUpper(p.curTok.Literal)
				if upper == "HASHED" {
					opt.Hashed = true
					p.nextToken()
				} else if upper == "MUST_CHANGE" {
					opt.MustChange = true
					p.nextToken()
				} else if upper == "UNLOCK" {
					opt.Unlock = true
					p.nextToken()
				} else if upper == "OLD_PASSWORD" {
					p.nextToken() // consume OLD_PASSWORD
					if p.curTok.Type == TokenEquals {
						p.nextToken()
					}
					if p.curTok.Type == TokenString {
						opt.OldPassword = p.parseStringLiteralValue()
						p.nextToken()
					}
				} else {
					break
				}
			}

			stmt.Options = append(stmt.Options, opt)
		} else if optName == "NO" && strings.ToUpper(p.peekTok.Literal) == "CREDENTIAL" {
			p.nextToken() // consume NO
			p.nextToken() // consume CREDENTIAL
			stmt.Options = append(stmt.Options, &ast.PrincipalOptionSimple{
				OptionKind: "NoCredential",
			})
		} else if optName == "NAME" {
			p.nextToken() // consume NAME
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			stmt.Options = append(stmt.Options, &ast.IdentifierPrincipalOption{
				OptionKind: "Name",
				Identifier: p.parseIdentifier(),
			})
		} else if optName == "DEFAULT_DATABASE" {
			p.nextToken() // consume DEFAULT_DATABASE
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			stmt.Options = append(stmt.Options, &ast.IdentifierPrincipalOption{
				OptionKind: "DefaultDatabase",
				Identifier: p.parseIdentifier(),
			})
		} else if optName == "DEFAULT_LANGUAGE" {
			p.nextToken() // consume DEFAULT_LANGUAGE
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			stmt.Options = append(stmt.Options, &ast.IdentifierPrincipalOption{
				OptionKind: "DefaultLanguage",
				Identifier: p.parseIdentifier(),
			})
		} else if optName == "CHECK_POLICY" {
			p.nextToken() // consume CHECK_POLICY
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			optState := "On"
			if strings.ToUpper(p.curTok.Literal) == "OFF" {
				optState = "Off"
			}
			p.nextToken()
			stmt.Options = append(stmt.Options, &ast.OnOffPrincipalOption{
				OptionKind:  "CheckPolicy",
				OptionState: optState,
			})
		} else if optName == "CHECK_EXPIRATION" {
			p.nextToken() // consume CHECK_EXPIRATION
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			optState := "On"
			if strings.ToUpper(p.curTok.Literal) == "OFF" {
				optState = "Off"
			}
			p.nextToken()
			stmt.Options = append(stmt.Options, &ast.OnOffPrincipalOption{
				OptionKind:  "CheckExpiration",
				OptionState: optState,
			})
		} else if optName == "CREDENTIAL" {
			p.nextToken() // consume CREDENTIAL
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			stmt.Options = append(stmt.Options, &ast.IdentifierPrincipalOption{
				OptionKind: "Credential",
				Identifier: p.parseIdentifier(),
			})
		} else {
			break
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

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

	// Parse WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken()

		for {
			optionName := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			// Handle PASSWORD specially for ALTER USER (can have OLD_PASSWORD)
			if optionName == "PASSWORD" {
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				passwordOpt := &ast.PasswordAlterPrincipalOption{
					OptionKind: "Password",
				}
				if p.curTok.Type == TokenString {
					passwordOpt.Password = p.parseStringLiteralValue()
					p.nextToken()
				}
				// Check for OLD_PASSWORD
				if strings.ToUpper(p.curTok.Literal) == "OLD_PASSWORD" {
					p.nextToken() // consume OLD_PASSWORD
					if p.curTok.Type == TokenEquals {
						p.nextToken()
					}
					if p.curTok.Type == TokenString {
						passwordOpt.OldPassword = p.parseStringLiteralValue()
						p.nextToken()
					}
				}
				stmt.UserOptions = append(stmt.UserOptions, passwordOpt)
			} else {
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}

				value, err := p.parseScalarExpression()
				if err != nil {
					break
				}

				// Check if value is a simple identifier
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
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
	}

	return stmt, nil
}

func (p *Parser) parseAlterRouteStatement() (*ast.AlterRouteStatement, error) {
	// Consume ROUTE
	p.nextToken()

	stmt := &ast.AlterRouteStatement{}

	// Parse route name
	stmt.Name = p.parseIdentifier()

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		stmt.RouteOptions = p.parseRouteOptions()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterAssemblyStatement() (*ast.AlterAssemblyStatement, error) {
	// Consume ASSEMBLY
	p.nextToken()

	stmt := &ast.AlterAssemblyStatement{}

	// Parse assembly name
	stmt.Name = p.parseIdentifier()

	// Parse clauses in any order
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		upperLit := strings.ToUpper(p.curTok.Literal)

		switch upperLit {
		case "FROM":
			p.nextToken() // consume FROM
			// Parse parameters (path literals)
			for {
				param, err := p.parseScalarExpression()
				if err != nil {
					break
				}
				stmt.Parameters = append(stmt.Parameters, param)
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}

		case "WITH":
			p.nextToken() // consume WITH
			// Parse options
		withLoop:
			for {
				optUpper := strings.ToUpper(p.curTok.Literal)
				switch optUpper {
				case "PERMISSION_SET":
					p.nextToken() // consume PERMISSION_SET
					if p.curTok.Type == TokenEquals {
						p.nextToken()
					}
					permSet := strings.ToUpper(p.curTok.Literal)
					opt := &ast.PermissionSetAssemblyOption{
						OptionKind: "PermissionSet",
					}
					switch permSet {
					case "SAFE":
						opt.PermissionSetOption = "Safe"
					case "EXTERNAL_ACCESS":
						opt.PermissionSetOption = "ExternalAccess"
					case "UNSAFE":
						opt.PermissionSetOption = "Unsafe"
					}
					p.nextToken()
					stmt.Options = append(stmt.Options, opt)

				case "VISIBILITY":
					p.nextToken() // consume VISIBILITY
					if p.curTok.Type == TokenEquals {
						p.nextToken()
					}
					stateUpper := strings.ToUpper(p.curTok.Literal)
					opt := &ast.OnOffAssemblyOption{
						OptionKind: "Visibility",
					}
					if stateUpper == "ON" {
						opt.OptionState = "On"
					} else {
						opt.OptionState = "Off"
					}
					p.nextToken()
					stmt.Options = append(stmt.Options, opt)

				case "UNCHECKED":
					p.nextToken() // consume UNCHECKED
					if strings.ToUpper(p.curTok.Literal) == "DATA" {
						p.nextToken() // consume DATA
					}
					stmt.Options = append(stmt.Options, &ast.AssemblyOption{
						OptionKind: "UncheckedData",
					})

				default:
					break withLoop
				}

				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}

		case "DROP":
			p.nextToken() // consume DROP
			if strings.ToUpper(p.curTok.Literal) == "FILE" {
				p.nextToken() // consume FILE
				if strings.ToUpper(p.curTok.Literal) == "ALL" {
					stmt.IsDropAll = true
					p.nextToken()
				} else {
					// Parse file names
					for {
						if p.curTok.Type == TokenString {
							value := p.curTok.Literal
							if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
								value = value[1 : len(value)-1]
							}
							stmt.DropFiles = append(stmt.DropFiles, &ast.StringLiteral{
								LiteralType:   "String",
								IsNational:    false,
								IsLargeObject: false,
								Value:         value,
							})
							p.nextToken()
						}
						if p.curTok.Type == TokenComma {
							p.nextToken()
						} else {
							break
						}
					}
				}
			}

		case "ADD":
			p.nextToken() // consume ADD
			if strings.ToUpper(p.curTok.Literal) == "FILE" {
				p.nextToken() // consume FILE
				if strings.ToUpper(p.curTok.Literal) == "FROM" {
					p.nextToken() // consume FROM
				}
				// Parse file specs
				for {
					fileSpec := &ast.AddFileSpec{}
					// Parse file (string or binary literal)
					file, err := p.parseScalarExpression()
					if err != nil {
						break
					}
					fileSpec.File = file

					// Check for AS 'filename'
					if p.curTok.Type == TokenAs {
						p.nextToken() // consume AS
						if p.curTok.Type == TokenString {
							value := p.curTok.Literal
							if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
								value = value[1 : len(value)-1]
							}
							fileSpec.FileName = &ast.StringLiteral{
								LiteralType:   "String",
								IsNational:    false,
								IsLargeObject: false,
								Value:         value,
							}
							p.nextToken()
						}
					}

					stmt.AddFiles = append(stmt.AddFiles, fileSpec)

					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}
			}

		default:
			// Unknown token - break out
			goto done
		}
	}

done:
	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAlterEndpointStatement() (*ast.AlterEndpointStatement, error) {
	// Consume ENDPOINT
	p.nextToken()

	stmt := &ast.AlterEndpointStatement{}
	hasOptions := false

	// Parse endpoint name
	stmt.Name = p.parseIdentifier()

	// Parse endpoint options (STATE, AFFINITY)
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		upper := strings.ToUpper(p.curTok.Literal)

		switch upper {
		case "STATE":
			hasOptions = true
			p.nextToken() // consume STATE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			stateUpper := strings.ToUpper(p.curTok.Literal)
			switch stateUpper {
			case "STARTED":
				stmt.State = "Started"
			case "STOPPED":
				stmt.State = "Stopped"
			case "DISABLED":
				stmt.State = "Disabled"
			}
			p.nextToken()

		case "AFFINITY":
			hasOptions = true
			p.nextToken() // consume AFFINITY
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			affinity := &ast.EndpointAffinity{}
			affinityUpper := strings.ToUpper(p.curTok.Literal)
			switch affinityUpper {
			case "NONE":
				affinity.Kind = "None"
				p.nextToken()
			case "ADMIN":
				affinity.Kind = "Admin"
				p.nextToken()
			default:
				// Integer affinity
				affinity.Kind = "Integer"
				if p.curTok.Type == TokenNumber {
					affinity.Value = &ast.IntegerLiteral{
						LiteralType: "Integer",
						Value:       p.curTok.Literal,
					}
					p.nextToken()
				}
			}
			stmt.Affinity = affinity

		case "AS":
			hasOptions = true
			p.nextToken() // consume AS
			// Protocol type (TCP, HTTP)
			protocolUpper := strings.ToUpper(p.curTok.Literal)
			switch protocolUpper {
			case "TCP":
				stmt.Protocol = "Tcp"
			case "HTTP":
				stmt.Protocol = "Http"
			}
			p.nextToken()
			// Parse protocol options (listener_port = value)
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optName := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					opt := &ast.LiteralEndpointProtocolOption{}
					switch optName {
					case "LISTENER_PORT":
						opt.Kind = "TcpListenerPort"
					case "LISTENER_IP":
						opt.Kind = "TcpListenerIP"
					default:
						opt.Kind = optName
					}
					if p.curTok.Type == TokenNumber {
						opt.Value = &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						}
						p.nextToken()
					} else if p.curTok.Type == TokenString {
						opt.Value = &ast.StringLiteral{
							LiteralType: "String",
							Value:       p.curTok.Literal,
						}
						p.nextToken()
					}
					stmt.ProtocolOptions = append(stmt.ProtocolOptions, opt)
					if p.curTok.Type == TokenComma {
						p.nextToken()
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}

		case "FOR":
			hasOptions = true
			p.nextToken() // consume FOR
			// Endpoint type (SOAP, SERVICE_BROKER, etc.)
			endpointTypeUpper := strings.ToUpper(p.curTok.Literal)
			switch endpointTypeUpper {
			case "SOAP":
				stmt.EndpointType = "Soap"
			case "SERVICE_BROKER":
				stmt.EndpointType = "ServiceBroker"
			case "DATABASE_MIRRORING":
				stmt.EndpointType = "DatabaseMirroring"
			case "TSQL":
				stmt.EndpointType = "Tsql"
			default:
				stmt.EndpointType = endpointTypeUpper
			}
			p.nextToken()
			// Parse payload options
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					actionUpper := strings.ToUpper(p.curTok.Literal)
					if actionUpper == "ADD" || actionUpper == "ALTER" || actionUpper == "DROP" {
						p.nextToken() // consume ADD/ALTER/DROP
						// Parse WEBMETHOD
						if strings.ToUpper(p.curTok.Literal) == "WEBMETHOD" {
							p.nextToken() // consume WEBMETHOD
							method := &ast.SoapMethod{
								Format: "NotSpecified",
								Schema: "NotSpecified",
							}
							switch actionUpper {
							case "ADD":
								method.Action = "Add"
								method.Kind = "WebMethod"
							case "ALTER":
								method.Action = "Alter"
								method.Kind = "WebMethod"
							case "DROP":
								method.Action = "Drop"
								method.Kind = "None"
							}
							// Parse alias (string literal)
							if p.curTok.Type == TokenString {
								method.Alias = p.parseStringLiteralValue()
								p.nextToken()
							}
							// Parse method options
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									optName := strings.ToUpper(p.curTok.Literal)
									p.nextToken()
									if p.curTok.Type == TokenEquals {
										p.nextToken() // consume =
									}
									if optName == "NAME" && p.curTok.Type == TokenString {
										method.Name = p.parseStringLiteralValue()
										p.nextToken()
									} else if optName == "FORMAT" {
										formatUpper := strings.ToUpper(p.curTok.Literal)
										switch formatUpper {
										case "ALL_RESULTS":
											method.Format = "AllResults"
										case "ROWSETS_ONLY":
											method.Format = "RowsetsOnly"
										case "NONE":
											method.Format = "None"
										default:
											method.Format = formatUpper
										}
										p.nextToken()
									} else if optName == "SCHEMA" {
										schemaUpper := strings.ToUpper(p.curTok.Literal)
										switch schemaUpper {
										case "DEFAULT":
											method.Schema = "Default"
										case "NONE":
											method.Schema = "None"
										case "STANDARD":
											method.Schema = "Standard"
										default:
											method.Schema = schemaUpper
										}
										p.nextToken()
									} else {
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
							stmt.PayloadOptions = append(stmt.PayloadOptions, method)
						}
					}
					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else if p.curTok.Type != TokenRParen {
						break
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}

		case ",":
			p.nextToken()

		default:
			// Unknown token, break out
			if hasOptions {
				// Set defaults for unspecified fields when options were parsed
				if stmt.State == "" {
					stmt.State = "NotSpecified"
				}
				if stmt.Protocol == "" {
					stmt.Protocol = "None"
				}
				if stmt.EndpointType == "" {
					stmt.EndpointType = "NotSpecified"
				}
			}
			return stmt, nil
		}
	}

	// Set defaults for unspecified fields when options were parsed
	if hasOptions {
		if stmt.State == "" {
			stmt.State = "NotSpecified"
		}
		if stmt.Protocol == "" {
			stmt.Protocol = "None"
		}
		if stmt.EndpointType == "" {
			stmt.EndpointType = "NotSpecified"
		}
	}

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

	// Check for ON QUEUE clause
	if p.curTok.Type == TokenOn && strings.ToUpper(p.peekTok.Literal) == "QUEUE" {
		p.nextToken() // consume ON
		p.nextToken() // consume QUEUE
		queueName, _ := p.parseSchemaObjectName()
		stmt.QueueName = queueName
	}

	// Check for contract modifications (ADD CONTRACT, DROP CONTRACT)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		var contracts []*ast.ServiceContract
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			action := "None"
			upperLit := strings.ToUpper(p.curTok.Literal)
			if upperLit == "ADD" {
				action = "Add"
				p.nextToken() // consume ADD
				if strings.ToUpper(p.curTok.Literal) == "CONTRACT" {
					p.nextToken() // consume CONTRACT
				}
			} else if upperLit == "DROP" {
				action = "Drop"
				p.nextToken() // consume DROP
				if strings.ToUpper(p.curTok.Literal) == "CONTRACT" {
					p.nextToken() // consume CONTRACT
				}
			}
			contract := &ast.ServiceContract{
				Name:   p.parseIdentifier(),
				Action: action,
			}
			contracts = append(contracts, contract)
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
		stmt.ServiceContracts = contracts
	}

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

	// Parse the action
	switch strings.ToUpper(p.curTok.Literal) {
	case "REMOVE":
		p.nextToken() // consume REMOVE
		switch strings.ToUpper(p.curTok.Literal) {
		case "PRIVATE":
			p.nextToken() // consume PRIVATE
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			stmt.Kind = "RemovePrivateKey"
		case "ATTESTED":
			p.nextToken() // consume ATTESTED
			if strings.ToUpper(p.curTok.Literal) == "OPTION" {
				p.nextToken() // consume OPTION
			}
			stmt.Kind = "RemoveAttestedOption"
		}
	case "ATTESTED":
		p.nextToken() // consume ATTESTED
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		attestedBy, _ := p.parseStringLiteral()
		stmt.AttestedBy = attestedBy
		stmt.Kind = "AttestedBy"
	case "WITH":
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "PRIVATE" {
			p.nextToken() // consume PRIVATE
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
		}
		stmt.Kind = "WithPrivateKey"
		// Parse (ENCRYPTION BY PASSWORD = '...', DECRYPTION BY PASSWORD = '...')
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				switch strings.ToUpper(p.curTok.Literal) {
				case "ENCRYPTION":
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
					pwd, _ := p.parseStringLiteral()
					stmt.EncryptionPassword = pwd
				case "DECRYPTION":
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
					pwd, _ := p.parseStringLiteral()
					stmt.DecryptionPassword = pwd
				default:
					p.nextToken()
				}
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

	// Check CATALOG, INDEX, or STOPLIST
	keyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	if keyword == "STOPLIST" {
		return p.parseAlterFulltextStopListStatement()
	}

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

	// Parse action (if any)
	action := p.tryParseAlterFullTextIndexAction()
	stmt.Action = action

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseAlterFulltextStopListStatement() (*ast.AlterFullTextStopListStatement, error) {
	stmt := &ast.AlterFullTextStopListStatement{
		Name: p.parseIdentifier(),
	}

	action := &ast.FullTextStopListAction{}

	// Parse ADD or DROP
	actionLit := strings.ToUpper(p.curTok.Literal)
	if actionLit == "ADD" {
		action.IsAdd = true
		p.nextToken() // consume ADD
	} else if actionLit == "DROP" {
		action.IsAdd = false
		p.nextToken() // consume DROP
	}

	// Check for ALL
	if strings.ToUpper(p.curTok.Literal) == "ALL" {
		action.IsAll = true
		p.nextToken() // consume ALL
	} else if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
		// Parse stopword
		strLit, _ := p.parseStringLiteral()
		action.StopWord = strLit
	}

	// Parse LANGUAGE term
	if p.curTok.Type == TokenLanguage || strings.ToUpper(p.curTok.Literal) == "LANGUAGE" {
		p.nextToken() // consume LANGUAGE
		action.LanguageTerm, _ = p.parseIdentifierOrValueExpression()
	}

	stmt.Action = action

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) tryParseAlterFullTextIndexAction() ast.AlterFullTextIndexActionOption {
	actionLit := strings.ToUpper(p.curTok.Literal)

	switch actionLit {
	case "ENABLE":
		p.nextToken()
		return &ast.SimpleAlterFullTextIndexAction{ActionKind: "Enable"}
	case "DISABLE":
		p.nextToken()
		return &ast.SimpleAlterFullTextIndexAction{ActionKind: "Disable"}
	case "SET":
		p.nextToken() // consume SET
		if strings.ToUpper(p.curTok.Literal) == "CHANGE_TRACKING" {
			// Parse CHANGE_TRACKING = MANUAL/AUTO/OFF
			p.nextToken() // consume CHANGE_TRACKING
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			trackingLit := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			switch trackingLit {
			case "MANUAL":
				return &ast.SimpleAlterFullTextIndexAction{ActionKind: "SetChangeTrackingManual"}
			case "AUTO":
				return &ast.SimpleAlterFullTextIndexAction{ActionKind: "SetChangeTrackingAuto"}
			case "OFF":
				return &ast.SimpleAlterFullTextIndexAction{ActionKind: "SetChangeTrackingOff"}
			}
		} else if strings.ToUpper(p.curTok.Literal) == "STOPLIST" {
			// Parse SET STOPLIST OFF | SYSTEM | name [WITH NO POPULATION]
			p.nextToken() // consume STOPLIST
			// Handle optional = sign
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			action := &ast.SetStopListAlterFullTextIndexAction{
				StopListOption: &ast.StopListFullTextIndexOption{
					OptionKind: "StopList",
				},
			}
			if strings.ToUpper(p.curTok.Literal) == "OFF" {
				action.StopListOption.IsOff = true
				p.nextToken()
			} else {
				action.StopListOption.IsOff = false
				action.StopListOption.StopListName = p.parseIdentifier()
			}
			// Check for WITH NO POPULATION
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				if strings.ToUpper(p.curTok.Literal) == "NO" {
					p.nextToken() // consume NO
					if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
						p.nextToken() // consume POPULATION
						action.WithNoPopulation = true
					}
				}
			}
			return action
		}
		return nil
	case "START":
		p.nextToken() // consume START
		popType := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
			p.nextToken()
		}
		switch popType {
		case "FULL":
			return &ast.SimpleAlterFullTextIndexAction{ActionKind: "StartFullPopulation"}
		case "INCREMENTAL":
			return &ast.SimpleAlterFullTextIndexAction{ActionKind: "StartIncrementalPopulation"}
		case "UPDATE":
			return &ast.SimpleAlterFullTextIndexAction{ActionKind: "StartUpdatePopulation"}
		}
		return nil
	case "STOP":
		p.nextToken() // consume STOP
		if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
			p.nextToken()
		}
		return &ast.SimpleAlterFullTextIndexAction{ActionKind: "StopPopulation"}
	case "PAUSE":
		p.nextToken() // consume PAUSE
		if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
			p.nextToken()
		}
		return &ast.SimpleAlterFullTextIndexAction{ActionKind: "PausePopulation"}
	case "RESUME":
		p.nextToken() // consume RESUME
		if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
			p.nextToken()
		}
		return &ast.SimpleAlterFullTextIndexAction{ActionKind: "ResumePopulation"}
	case "ADD":
		action, _ := p.parseAddAlterFullTextIndexAction()
		return action
	case "DROP":
		action, _ := p.parseDropAlterFullTextIndexAction()
		return action
	}

	// No action found
	return nil
}

func (p *Parser) parseAddAlterFullTextIndexAction() (*ast.AddAlterFullTextIndexAction, error) {
	p.nextToken() // consume ADD

	action := &ast.AddAlterFullTextIndexAction{}

	// Parse (column list)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			col := &ast.FullTextIndexColumn{}
			col.Name = p.parseIdentifier()

			// Check for TYPE COLUMN
			if strings.ToUpper(p.curTok.Literal) == "TYPE" {
				p.nextToken() // consume TYPE
				if strings.ToUpper(p.curTok.Literal) == "COLUMN" {
					p.nextToken() // consume COLUMN
				}
				col.TypeColumn = p.parseIdentifier()
			}

			// Check for LANGUAGE
			if strings.ToUpper(p.curTok.Literal) == "LANGUAGE" {
				p.nextToken() // consume LANGUAGE
				col.LanguageTerm = &ast.IdentifierOrValueExpression{}
				if p.curTok.Type == TokenNumber {
					col.LanguageTerm.Value = p.curTok.Literal
					col.LanguageTerm.ValueExpression = &ast.IntegerLiteral{Value: p.curTok.Literal, LiteralType: "Integer"}
					p.nextToken()
				} else if p.curTok.Type == TokenString {
					// Strip quotes from string literal
					val := p.curTok.Literal
					if len(val) >= 2 && (val[0] == '\'' || val[0] == '"') {
						val = val[1 : len(val)-1]
					}
					col.LanguageTerm.Value = val
					col.LanguageTerm.ValueExpression = &ast.StringLiteral{Value: val, LiteralType: "String"}
					p.nextToken()
				}
			}

			// StatisticalSemantics defaults to false

			action.Columns = append(action.Columns, col)

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

	// Check for WITH NO POPULATION
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "NO" {
			p.nextToken() // consume NO
			if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
				p.nextToken() // consume POPULATION
				action.WithNoPopulation = true
			}
		}
	}

	return action, nil
}

func (p *Parser) parseDropAlterFullTextIndexAction() (*ast.DropAlterFullTextIndexAction, error) {
	p.nextToken() // consume DROP

	action := &ast.DropAlterFullTextIndexAction{}

	// Parse (column list)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			action.Columns = append(action.Columns, p.parseIdentifier())

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

	// Check for WITH NO POPULATION
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "NO" {
			p.nextToken() // consume NO
			if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
				p.nextToken() // consume POPULATION
				action.WithNoPopulation = true
			}
		}
	}

	return action, nil
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

	// Parse ADD or DROP
	hasAction := false
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "ADD" {
		stmt.IsAdd = true
		hasAction = true
		p.nextToken()
	} else if upperLit == "DROP" {
		stmt.IsAdd = false
		hasAction = true
		p.nextToken()
	}

	// Only parse ENCRYPTION BY and mechanisms if there was an ADD or DROP
	if hasAction {
		// Expect ENCRYPTION
		if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
			p.nextToken() // consume ENCRYPTION
		}

		// Expect BY
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}

		// Parse encrypting mechanisms
		for {
			mechType := strings.ToUpper(p.curTok.Literal)
			mechanism := &ast.CryptoMechanism{}
			parsed := true

			switch mechType {
			case "PASSWORD":
				p.nextToken()
				mechanism.CryptoMechanismType = "Password"
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				pwd, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				mechanism.PasswordOrSignature = pwd

			case "CERTIFICATE":
				p.nextToken()
				mechanism.CryptoMechanismType = "Certificate"
				mechanism.Identifier = p.parseIdentifier()

			case "SYMMETRIC":
				p.nextToken()
				if p.curTok.Type == TokenKey {
					p.nextToken() // consume KEY
				}
				mechanism.CryptoMechanismType = "SymmetricKey"
				mechanism.Identifier = p.parseIdentifier()

			case "ASYMMETRIC":
				p.nextToken()
				if p.curTok.Type == TokenKey {
					p.nextToken() // consume KEY
				}
				mechanism.CryptoMechanismType = "AsymmetricKey"
				mechanism.Identifier = p.parseIdentifier()

			default:
				parsed = false
			}

			if !parsed {
				break
			}

			stmt.EncryptingMechanisms = append(stmt.EncryptingMechanisms, mechanism)

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

	// Parse WITH options (like RECOMPILE, ENCRYPTION, NATIVE_COMPILATION, etc.)
	if p.curTok.Type == TokenWith {
		p.nextToken()
		for {
			if strings.ToUpper(p.curTok.Literal) == "FOR" || p.curTok.Type == TokenAs || p.curTok.Type == TokenEOF {
				break
			}
			upperLit := strings.ToUpper(p.curTok.Literal)
			if upperLit == "RECOMPILE" {
				stmt.Options = append(stmt.Options, &ast.ProcedureOption{OptionKind: "Recompile"})
				p.nextToken()
			} else if upperLit == "ENCRYPTION" {
				stmt.Options = append(stmt.Options, &ast.ProcedureOption{OptionKind: "Encryption"})
				p.nextToken()
			} else if upperLit == "NATIVE_COMPILATION" {
				stmt.Options = append(stmt.Options, &ast.ProcedureOption{OptionKind: "NativeCompilation"})
				p.nextToken()
			} else if upperLit == "SCHEMABINDING" {
				stmt.Options = append(stmt.Options, &ast.ProcedureOption{OptionKind: "SchemaBinding"})
				p.nextToken()
			} else if upperLit == "EXECUTE" {
				p.nextToken() // consume EXECUTE
				if p.curTok.Type == TokenAs {
					p.nextToken() // consume AS
				}
				executeAsOpt := &ast.ExecuteAsProcedureOption{
					OptionKind: "ExecuteAs",
					ExecuteAs:  &ast.ExecuteAsClause{},
				}
				upperOption := strings.ToUpper(p.curTok.Literal)
				if upperOption == "CALLER" {
					executeAsOpt.ExecuteAs.ExecuteAsOption = "Caller"
					p.nextToken()
				} else if upperOption == "SELF" {
					executeAsOpt.ExecuteAs.ExecuteAsOption = "Self"
					p.nextToken()
				} else if upperOption == "OWNER" {
					executeAsOpt.ExecuteAs.ExecuteAsOption = "Owner"
					p.nextToken()
				} else if p.curTok.Type == TokenString {
					executeAsOpt.ExecuteAs.ExecuteAsOption = "String"
					value := p.curTok.Literal
					// Strip quotes
					if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
						value = value[1 : len(value)-1]
					}
					executeAsOpt.ExecuteAs.Literal = &ast.StringLiteral{
						LiteralType:   "String",
						IsNational:    false,
						IsLargeObject: false,
						Value:         value,
					}
					p.nextToken()
				}
				stmt.Options = append(stmt.Options, executeAsOpt)
			} else if upperLit == "REPLICATION" {
				stmt.IsForReplication = true
				p.nextToken()
			} else {
				p.nextToken()
			}
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else if p.curTok.Type == TokenAs || strings.ToUpper(p.curTok.Literal) == "FOR" || p.curTok.Type == TokenEOF {
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
	case "RESOURCE":
		return p.parseAlterExternalResourcePoolStatement()
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

	stmt := &ast.AlterExternalDataSourceStatement{
		DataSourceType:         "HADOOP",
		PreviousPushDownOption: "ON",
	}

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

		// Handle LOCATION as a separate field
		if optName == "LOCATION" {
			if p.curTok.Type == TokenString {
				strLit, _ := p.parseStringLiteral()
				stmt.Location = strLit
			} else {
				p.nextToken()
			}
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
			continue
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

	// Parse optional AUTHORIZATION
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Parse operation (SET, ADD, REMOVE)
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "SET" || upperLit == "ADD" || upperLit == "REMOVE" {
		stmt.Operation = p.parseIdentifier()

		if upperLit == "REMOVE" {
			// REMOVE PLATFORM <platform>
			if strings.ToUpper(p.curTok.Literal) == "PLATFORM" {
				p.nextToken() // consume PLATFORM
				stmt.Platform = p.parseIdentifier()
			}
		} else {
			// SET or ADD - parse file options
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				fileOption := &ast.ExternalLanguageFileOption{}
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					switch strings.ToUpper(p.curTok.Literal) {
					case "CONTENT":
						p.nextToken() // consume CONTENT
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						expr, _ := p.parseScalarExpression()
						fileOption.Content = expr
					case "FILE_NAME":
						p.nextToken() // consume FILE_NAME
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						expr, _ := p.parseScalarExpression()
						fileOption.FileName = expr
					case "PLATFORM":
						p.nextToken() // consume PLATFORM
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						fileOption.Platform = p.parseIdentifier()
					case "PARAMETERS":
						p.nextToken() // consume PARAMETERS
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						expr, _ := p.parseScalarExpression()
						fileOption.Parameters = expr
					case "ENVIRONMENT_VARIABLES":
						p.nextToken() // consume ENVIRONMENT_VARIABLES
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						expr, _ := p.parseScalarExpression()
						fileOption.EnvironmentVariables = expr
					default:
						p.nextToken()
					}
					if p.curTok.Type == TokenComma {
						p.nextToken()
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
				stmt.ExternalLanguageFiles = append(stmt.ExternalLanguageFiles, fileOption)
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

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
			var currentFileOption *ast.ExternalLibraryFileOption
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
						currentFileOption = &ast.ExternalLibraryFileOption{
							Content: content,
						}
						stmt.ExternalLibraryFiles = append(stmt.ExternalLibraryFiles, currentFileOption)
					} else if optName == "PLATFORM" && currentFileOption != nil {
						// PLATFORM is an identifier, not a string
						currentFileOption.Platform = p.parseIdentifier()
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

func (p *Parser) parseAlterExternalResourcePoolStatement() (*ast.AlterExternalResourcePoolStatement, error) {
	// Consume RESOURCE
	p.nextToken()

	// Expect POOL
	if strings.ToUpper(p.curTok.Literal) != "POOL" {
		return nil, fmt.Errorf("expected POOL, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterExternalResourcePoolStatement{}

	// Parse pool name
	stmt.Name = p.parseIdentifier()

	// Expect WITH
	if p.curTok.Type != TokenWith && strings.ToUpper(p.curTok.Literal) != "WITH" {
		return nil, fmt.Errorf("expected WITH, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected (, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse parameters
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		paramName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		param := &ast.ExternalResourcePoolParameter{}

		switch paramName {
		case "MAX_CPU_PERCENT":
			param.ParameterType = "MaxCpuPercent"
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			param.ParameterValue = val
		case "MAX_MEMORY_PERCENT":
			param.ParameterType = "MaxMemoryPercent"
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			param.ParameterValue = val
		case "MAX_PROCESSES":
			param.ParameterType = "MaxProcesses"
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			param.ParameterValue = val
		case "AFFINITY":
			param.ParameterType = "Affinity"
			affinitySpec := &ast.ExternalResourcePoolAffinitySpecification{}

			// Parse CPU or NUMANODE
			affinityType := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			if affinityType == "CPU" {
				affinitySpec.AffinityType = "Cpu"
			} else if affinityType == "NUMANODE" {
				affinitySpec.AffinityType = "NumaNode"
			}

			// Expect =
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}

			// Check for AUTO or range list
			if strings.ToUpper(p.curTok.Literal) == "AUTO" {
				affinitySpec.IsAuto = true
				p.nextToken()
			} else {
				// Parse range list: (1) or (1 TO 5, 6 TO 7)
				if p.curTok.Type == TokenLParen {
					p.nextToken()
				}
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					fromVal, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					rangeItem := &ast.LiteralRange{From: fromVal}

					// Check for TO
					if strings.ToUpper(p.curTok.Literal) == "TO" {
						p.nextToken()
						toVal, err := p.parseScalarExpression()
						if err != nil {
							return nil, err
						}
						rangeItem.To = toVal
					}

					affinitySpec.PoolAffinityRanges = append(affinitySpec.PoolAffinityRanges, rangeItem)

					// Check for comma
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
			param.AffinitySpecification = affinitySpec
		}

		stmt.ExternalResourcePoolParameters = append(stmt.ExternalResourcePoolParameters, param)

		// Check for comma
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
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

// convertOptionKind converts a SQL option name (e.g., "OPTIMIZED_LOCKING") to its OptionKind form (e.g., "OptimizedLocking")
func convertOptionKind(optionName string) string {
	// Handle special cases with specific capitalization
	switch optionName {
	case "VARDECIMAL_STORAGE_FORMAT":
		return "VarDecimalStorageFormat"
	case "ARITHABORT":
		return "ArithAbort"
	case "NUMERIC_ROUNDABORT":
		return "NumericRoundAbort"
	case "DB_CHAINING":
		return "DBChaining"
	case "AUTO_UPDATE_STATISTICS_ASYNC":
		return "AutoUpdateStatisticsAsync"
	case "DATE_CORRELATION_OPTIMIZATION":
		return "DateCorrelationOptimization"
	case "ALLOW_SNAPSHOT_ISOLATION":
		return "AllowSnapshotIsolation"
	case "READ_COMMITTED_SNAPSHOT":
		return "ReadCommittedSnapshot"
	case "SUPPLEMENTAL_LOGGING":
		return "SupplementalLogging"
	}

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

		// Check if EXTERNAL comes first
		if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
			p.nextToken() // consume EXTERNAL
			stmt.ExternalPoolName = p.parseIdentifier()

			// Check for comma and internal pool
			if p.curTok.Type == TokenComma {
				p.nextToken()
				stmt.PoolName = p.parseIdentifier()
			}
		} else {
			// Internal pool first
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

func (p *Parser) parseAddStatement() (ast.Statement, error) {
	// Consume ADD
	p.nextToken()

	upper := strings.ToUpper(p.curTok.Literal)
	switch upper {
	case "SIGNATURE":
		return p.parseAddSignatureStatement(false)
	case "COUNTER":
		p.nextToken() // consume COUNTER
		if strings.ToUpper(p.curTok.Literal) != "SIGNATURE" {
			return nil, fmt.Errorf("expected SIGNATURE after COUNTER, got %s", p.curTok.Literal)
		}
		return p.parseAddSignatureStatement(true)
	case "SENSITIVITY":
		return p.parseAddSensitivityClassificationStatement()
	}

	return nil, fmt.Errorf("unexpected token after ADD: %s", p.curTok.Literal)
}

func (p *Parser) parseAddSignatureStatement(isCounter bool) (*ast.AddSignatureStatement, error) {
	// Consume SIGNATURE
	p.nextToken()

	stmt := &ast.AddSignatureStatement{
		IsCounter:   isCounter,
		ElementKind: "NotSpecified",
	}

	// Expect TO
	if strings.ToUpper(p.curTok.Literal) != "TO" {
		return nil, fmt.Errorf("expected TO after SIGNATURE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse element kind if present (OBJECT::, ASSEMBLY::, DATABASE::)
	stmt.ElementKind, stmt.Element = p.parseSignatureElement()

	// Expect BY
	if strings.ToUpper(p.curTok.Literal) != "BY" {
		return nil, fmt.Errorf("expected BY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse crypto mechanisms
	cryptos, err := p.parseSignatureCryptoMechanisms()
	if err != nil {
		return nil, err
	}
	stmt.Cryptos = cryptos

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDropSignatureStatement(isCounter bool) (*ast.DropSignatureStatement, error) {
	// Consume SIGNATURE
	p.nextToken()

	stmt := &ast.DropSignatureStatement{
		IsCounter:   isCounter,
		ElementKind: "NotSpecified",
	}

	// Expect FROM
	if strings.ToUpper(p.curTok.Literal) != "FROM" {
		return nil, fmt.Errorf("expected FROM after SIGNATURE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse element kind if present (OBJECT::, ASSEMBLY::, DATABASE::)
	stmt.ElementKind, stmt.Element = p.parseSignatureElement()

	// Expect BY
	if strings.ToUpper(p.curTok.Literal) != "BY" {
		return nil, fmt.Errorf("expected BY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse crypto mechanisms
	cryptos, err := p.parseSignatureCryptoMechanisms()
	if err != nil {
		return nil, err
	}
	stmt.Cryptos = cryptos

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseAddSensitivityClassificationStatement() (*ast.AddSensitivityClassificationStatement, error) {
	// Consume SENSITIVITY
	p.nextToken()

	if strings.ToUpper(p.curTok.Literal) != "CLASSIFICATION" {
		return nil, fmt.Errorf("expected CLASSIFICATION after SENSITIVITY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume CLASSIFICATION

	if strings.ToUpper(p.curTok.Literal) != "TO" {
		return nil, fmt.Errorf("expected TO after CLASSIFICATION, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume TO

	stmt := &ast.AddSensitivityClassificationStatement{}

	// Parse column references (comma-separated)
	for {
		colRef := p.parseColumnReferenceForSensitivity()
		stmt.Columns = append(stmt.Columns, colRef)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	// Parse WITH clause
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH

		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		// Parse options
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			opt := &ast.SensitivityClassificationOption{}

			// Parse option type
			optType := strings.ToUpper(p.curTok.Literal)
			switch optType {
			case "LABEL":
				opt.Type = "Label"
			case "LABEL_ID":
				opt.Type = "LabelId"
			case "INFORMATION_TYPE":
				opt.Type = "InformationType"
			case "INFORMATION_TYPE_ID":
				opt.Type = "InformationTypeId"
			case "RANK":
				opt.Type = "Rank"
			default:
				return nil, fmt.Errorf("unexpected sensitivity classification option: %s", p.curTok.Literal)
			}
			p.nextToken() // consume option type

			// Expect =
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			// Parse value
			if p.curTok.Type == TokenString {
				value := p.curTok.Literal
				// Remove quotes
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				opt.Value = &ast.StringLiteral{
					LiteralType:   "String",
					IsNational:    false,
					IsLargeObject: false,
					Value:         value,
				}
				p.nextToken()
			} else {
				// Identifier literal (for RANK = HIGH, etc.)
				opt.Value = &ast.IdentifierLiteral{
					LiteralType: "Identifier",
					QuoteType:   "NotQuoted",
					Value:       strings.ToUpper(p.curTok.Literal),
				}
				p.nextToken()
			}

			stmt.Options = append(stmt.Options, opt)

			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	return stmt, nil
}

func (p *Parser) parseDropSensitivityClassificationStatement() (*ast.DropSensitivityClassificationStatement, error) {
	// Consume SENSITIVITY
	p.nextToken()

	if strings.ToUpper(p.curTok.Literal) != "CLASSIFICATION" {
		return nil, fmt.Errorf("expected CLASSIFICATION after SENSITIVITY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume CLASSIFICATION

	if strings.ToUpper(p.curTok.Literal) != "FROM" {
		return nil, fmt.Errorf("expected FROM after CLASSIFICATION, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume FROM

	stmt := &ast.DropSensitivityClassificationStatement{}

	// Parse column references (comma-separated)
	for {
		colRef := p.parseColumnReferenceForSensitivity()
		stmt.Columns = append(stmt.Columns, colRef)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	return stmt, nil
}

func (p *Parser) parseColumnReferenceForSensitivity() *ast.ColumnReferenceExpression {
	colRef := &ast.ColumnReferenceExpression{
		ColumnType: "Regular",
	}

	var identifiers []*ast.Identifier
	for {
		ident := p.parseIdentifier()
		identifiers = append(identifiers, ident)

		if p.curTok.Type == TokenDot {
			p.nextToken() // consume .
		} else {
			break
		}
	}

	colRef.MultiPartIdentifier = &ast.MultiPartIdentifier{
		Count:       len(identifiers),
		Identifiers: identifiers,
	}

	return colRef
}

func (p *Parser) parseSignatureElement() (string, *ast.SchemaObjectName) {
	// Check for element kind prefix (OBJECT::, ASSEMBLY::, DATABASE::)
	elementKind := "NotSpecified"

	upper := strings.ToUpper(p.curTok.Literal)
	if upper == "OBJECT" || upper == "ASSEMBLY" || upper == "DATABASE" {
		// Look ahead for ::
		if p.peekTok.Type == TokenColonColon {
			switch upper {
			case "OBJECT":
				elementKind = "Object"
			case "ASSEMBLY":
				elementKind = "Assembly"
			case "DATABASE":
				elementKind = "Database"
			}
			p.nextToken() // consume kind
			p.nextToken() // consume ::
		}
	}

	// Parse the element name
	element, _ := p.parseSchemaObjectName()

	return elementKind, element
}

func (p *Parser) parseSignatureCryptoMechanisms() ([]*ast.CryptoMechanism, error) {
	var cryptos []*ast.CryptoMechanism

	for {
		crypto, err := p.parseSignatureCryptoMechanism()
		if err != nil {
			return nil, err
		}
		if crypto != nil {
			cryptos = append(cryptos, crypto)
		}

		// Check for comma to continue
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}
		break
	}

	return cryptos, nil
}

func (p *Parser) parseSignatureCryptoMechanism() (*ast.CryptoMechanism, error) {
	crypto := &ast.CryptoMechanism{}

	upper := strings.ToUpper(p.curTok.Literal)

	switch upper {
	case "CERTIFICATE":
		crypto.CryptoMechanismType = "Certificate"
		p.nextToken()
		crypto.Identifier = p.parseIdentifier()
	case "ASYMMETRIC":
		p.nextToken() // consume ASYMMETRIC
		if strings.ToUpper(p.curTok.Literal) != "KEY" {
			return nil, fmt.Errorf("expected KEY after ASYMMETRIC, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume KEY
		crypto.CryptoMechanismType = "AsymmetricKey"
		crypto.Identifier = p.parseIdentifier()
	case "PASSWORD":
		crypto.CryptoMechanismType = "Password"
		p.nextToken() // consume PASSWORD
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			crypto.PasswordOrSignature = val
		}
	default:
		return nil, nil
	}

	// Check for WITH PASSWORD = or WITH SIGNATURE =
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		optUpper := strings.ToUpper(p.curTok.Literal)
		if optUpper == "PASSWORD" || optUpper == "SIGNATURE" {
			p.nextToken() // consume PASSWORD/SIGNATURE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				crypto.PasswordOrSignature = val
			}
		}
	}

	return crypto, nil
}

func (p *Parser) parseAlterSearchPropertyListStatement() (*ast.AlterSearchPropertyListStatement, error) {
	// Consume SEARCH
	p.nextToken()
	// Consume PROPERTY
	if strings.ToUpper(p.curTok.Literal) == "PROPERTY" {
		p.nextToken()
	}
	// Consume LIST
	if strings.ToUpper(p.curTok.Literal) == "LIST" {
		p.nextToken()
	}

	stmt := &ast.AlterSearchPropertyListStatement{}

	// Parse the list name
	stmt.Name = p.parseIdentifier()

	// Parse action: ADD or DROP
	actionType := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume ADD or DROP

	switch actionType {
	case "ADD":
		addAction := &ast.AddSearchPropertyListAction{}
		// Parse property name (string literal)
		if p.curTok.Type == TokenString {
			value := p.curTok.Literal
			if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
				value = value[1 : len(value)-1]
			}
			addAction.PropertyName = &ast.StringLiteral{
				LiteralType:   "String",
				IsNational:    false,
				IsLargeObject: false,
				Value:         value,
			}
			p.nextToken()
		}
		// Parse WITH clause
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Parse options
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optUpper := strings.ToUpper(p.curTok.Literal)
					switch optUpper {
					case "PROPERTY_SET_GUID":
						p.nextToken() // consume PROPERTY_SET_GUID
						if p.curTok.Type == TokenEquals {
							p.nextToken()
						}
						if p.curTok.Type == TokenString {
							value := p.curTok.Literal
							if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
								value = value[1 : len(value)-1]
							}
							addAction.Guid = &ast.StringLiteral{
								LiteralType:   "String",
								IsNational:    false,
								IsLargeObject: false,
								Value:         value,
							}
							p.nextToken()
						}
					case "PROPERTY_INT_ID":
						p.nextToken() // consume PROPERTY_INT_ID
						if p.curTok.Type == TokenEquals {
							p.nextToken()
						}
						if p.curTok.Type == TokenNumber {
							addAction.Id = &ast.IntegerLiteral{
								LiteralType: "Integer",
								Value:       p.curTok.Literal,
							}
							p.nextToken()
						}
					case "PROPERTY_DESCRIPTION":
						p.nextToken() // consume PROPERTY_DESCRIPTION
						if p.curTok.Type == TokenEquals {
							p.nextToken()
						}
						if p.curTok.Type == TokenString {
							value := p.curTok.Literal
							if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
								value = value[1 : len(value)-1]
							}
							addAction.Description = &ast.StringLiteral{
								LiteralType:   "String",
								IsNational:    false,
								IsLargeObject: false,
								Value:         value,
							}
							p.nextToken()
						}
					default:
						p.nextToken() // skip unknown option
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
		stmt.Action = addAction

	case "DROP":
		dropAction := &ast.DropSearchPropertyListAction{}
		// Parse property name (string literal)
		if p.curTok.Type == TokenString {
			value := p.curTok.Literal
			if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
				value = value[1 : len(value)-1]
			}
			dropAction.PropertyName = &ast.StringLiteral{
				LiteralType:   "String",
				IsNational:    false,
				IsLargeObject: false,
				Value:         value,
			}
			p.nextToken()
		}
		stmt.Action = dropAction
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCreateResourcePoolStatement parses CREATE RESOURCE POOL statement
func (p *Parser) parseCreateResourcePoolStatement() (*ast.CreateResourcePoolStatement, error) {
	// We've already consumed CREATE RESOURCE
	// Consume POOL
	if strings.ToUpper(p.curTok.Literal) == "POOL" {
		p.nextToken()
	}

	stmt := &ast.CreateResourcePoolStatement{}

	// Parse pool name
	stmt.Name = p.parseIdentifier()

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		params, err := p.parseResourcePoolParameters()
		if err != nil {
			return nil, err
		}
		stmt.ResourcePoolParameters = params
	}

	return stmt, nil
}

// parseAlterResourcePoolStatement parses ALTER RESOURCE POOL statement
func (p *Parser) parseAlterResourcePoolStatement() (*ast.AlterResourcePoolStatement, error) {
	// Consume POOL (we've already consumed ALTER RESOURCE)
	p.nextToken()

	stmt := &ast.AlterResourcePoolStatement{}

	// Parse pool name
	stmt.Name = p.parseIdentifier()

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		params, err := p.parseResourcePoolParameters()
		if err != nil {
			return nil, err
		}
		stmt.ResourcePoolParameters = params
	}

	return stmt, nil
}

// parseResourcePoolParameters parses resource pool parameters within WITH (...)
func (p *Parser) parseResourcePoolParameters() ([]*ast.ResourcePoolParameter, error) {
	var params []*ast.ResourcePoolParameter

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param, err := p.parseResourcePoolParameter()
		if err != nil {
			return nil, err
		}
		if param != nil {
			params = append(params, param)
		}

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return params, nil
}

// parseResourcePoolParameter parses a single resource pool parameter
func (p *Parser) parseResourcePoolParameter() (*ast.ResourcePoolParameter, error) {
	paramName := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume parameter name

	param := &ast.ResourcePoolParameter{}

	switch paramName {
	case "MIN_CPU_PERCENT":
		param.ParameterType = "MinCpuPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "MAX_CPU_PERCENT":
		param.ParameterType = "MaxCpuPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "CAP_CPU_PERCENT":
		param.ParameterType = "CapCpuPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "MIN_MEMORY_PERCENT":
		param.ParameterType = "MinMemoryPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "MAX_MEMORY_PERCENT":
		param.ParameterType = "MaxMemoryPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "TARGET_MEMORY_PERCENT":
		param.ParameterType = "TargetMemoryPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "MIN_IO_PERCENT":
		param.ParameterType = "MinIoPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "MAX_IO_PERCENT":
		param.ParameterType = "MaxIoPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "CAP_IO_PERCENT":
		param.ParameterType = "CapIoPercent"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "MIN_IOPS_PER_VOLUME":
		param.ParameterType = "MinIopsPerVolume"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "MAX_IOPS_PER_VOLUME":
		param.ParameterType = "MaxIopsPerVolume"
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		param.ParameterValue = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
		p.nextToken()
	case "AFFINITY":
		param.ParameterType = "Affinity"
		affSpec, err := p.parseResourcePoolAffinitySpecification()
		if err != nil {
			return nil, err
		}
		param.AffinitySpecification = affSpec
	default:
		// Skip unknown parameter
		return nil, nil
	}

	return param, nil
}

// parseResourcePoolAffinitySpecification parses AFFINITY SCHEDULER/NUMANODE specification
func (p *Parser) parseResourcePoolAffinitySpecification() (*ast.ResourcePoolAffinitySpecification, error) {
	spec := &ast.ResourcePoolAffinitySpecification{}

	// Parse SCHEDULER or NUMANODE
	affinityType := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	switch affinityType {
	case "SCHEDULER":
		spec.AffinityType = "Scheduler"
	case "NUMANODE":
		spec.AffinityType = "NumaNode"
	default:
		return nil, fmt.Errorf("expected SCHEDULER or NUMANODE after AFFINITY, got %s", affinityType)
	}

	// Expect =
	if p.curTok.Type == TokenEquals {
		p.nextToken()
	}

	// Check for AUTO or range list
	if strings.ToUpper(p.curTok.Literal) == "AUTO" {
		spec.IsAuto = true
		p.nextToken()
	} else if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		spec.IsAuto = false

		// Parse range list
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			lr := &ast.LiteralRange{}

			// Parse 'from' value
			lr.From = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
			p.nextToken()

			// Check for TO
			if strings.ToUpper(p.curTok.Literal) == "TO" {
				p.nextToken() // consume TO
				lr.To = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
				p.nextToken()
			}

			spec.PoolAffinityRanges = append(spec.PoolAffinityRanges, lr)

			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	return spec, nil
}

func (p *Parser) parseAlterBrokerPriorityStatement() (*ast.AlterBrokerPriorityStatement, error) {
	// Consume BROKER
	p.nextToken()

	// Consume PRIORITY
	if strings.ToUpper(p.curTok.Literal) == "PRIORITY" {
		p.nextToken()
	}

	stmt := &ast.AlterBrokerPriorityStatement{}

	// Parse priority name
	stmt.Name = p.parseIdentifier()

	// Parse FOR CONVERSATION
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "CONVERSATION" {
			p.nextToken() // consume CONVERSATION
		}
	}

	// Parse SET (parameters)
	if strings.ToUpper(p.curTok.Literal) == "SET" {
		p.nextToken() // consume SET
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			stmt.BrokerPriorityParameters = p.parseBrokerPriorityParameters()
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseDropBrokerPriorityStatement() (*ast.DropBrokerPriorityStatement, error) {
	// Consume BROKER
	p.nextToken()

	// Consume PRIORITY
	if strings.ToUpper(p.curTok.Literal) == "PRIORITY" {
		p.nextToken()
	}

	stmt := &ast.DropBrokerPriorityStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken() // consume IF
		if strings.ToUpper(p.curTok.Literal) == "EXISTS" {
			stmt.IsIfExists = true
			p.nextToken() // consume EXISTS
		}
	}

	// Parse priority name
	stmt.Name = p.parseIdentifier()

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterTableRebuildStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableRebuildStatement, error) {
	stmt := &ast.AlterTableRebuildStatement{
		SchemaObjectName: tableName,
	}

	// Consume REBUILD
	p.nextToken()

	// Check for PARTITION
	if strings.ToUpper(p.curTok.Literal) == "PARTITION" {
		p.nextToken() // consume PARTITION
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		stmt.Partition = &ast.PartitionSpecifier{}
		if strings.ToUpper(p.curTok.Literal) == "ALL" {
			stmt.Partition.All = true
			p.nextToken()
		} else if p.curTok.Type == TokenNumber {
			stmt.Partition.All = false
			stmt.Partition.Number = &ast.IntegerLiteral{
				LiteralType: "Integer",
				Value:       p.curTok.Literal,
			}
			p.nextToken()
		}
	}

	// Check for WITH
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optionName := strings.ToUpper(p.curTok.Literal)
				p.nextToken() // consume option name
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				switch optionName {
				case "MAXDOP":
					opt := &ast.IndexExpressionOption{
						OptionKind: "MaxDop",
						Expression: &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						},
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
					p.nextToken()
				case "SORT_IN_TEMPDB":
					stateUpper := strings.ToUpper(p.curTok.Literal)
					state := "On"
					if stateUpper == "OFF" {
						state = "Off"
					}
					p.nextToken()
					opt := &ast.IndexStateOption{
						OptionKind:  "SortInTempDB",
						OptionState: state,
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
				case "PAD_INDEX":
					stateUpper := strings.ToUpper(p.curTok.Literal)
					state := "On"
					if stateUpper == "OFF" {
						state = "Off"
					}
					p.nextToken()
					opt := &ast.IndexStateOption{
						OptionKind:  "PadIndex",
						OptionState: state,
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
				case "FILLFACTOR":
					opt := &ast.IndexExpressionOption{
						OptionKind: "FillFactor",
						Expression: &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						},
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
					p.nextToken()
				case "ONLINE":
					stateUpper := strings.ToUpper(p.curTok.Literal)
					state := "On"
					if stateUpper == "OFF" {
						state = "Off"
					}
					p.nextToken()
					opt := &ast.OnlineIndexOption{
						OptionKind:  "Online",
						OptionState: state,
					}
					// Check for (WAIT_AT_LOW_PRIORITY ...)
					if p.curTok.Type == TokenLParen {
						p.nextToken() // consume (
						if strings.ToUpper(p.curTok.Literal) == "WAIT_AT_LOW_PRIORITY" {
							p.nextToken() // consume WAIT_AT_LOW_PRIORITY
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								lwOpt := &ast.OnlineIndexLowPriorityLockWaitOption{}
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									lwOptName := strings.ToUpper(p.curTok.Literal)
									p.nextToken()
									if p.curTok.Type == TokenEquals {
										p.nextToken()
									}
									if lwOptName == "MAX_DURATION" {
										maxDurOpt := &ast.LowPriorityLockWaitMaxDurationOption{
											OptionKind:  "MaxDuration",
											MaxDuration: &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal},
										}
										p.nextToken()
										// Check for MINUTES
										if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
											maxDurOpt.Unit = "Minutes"
											p.nextToken()
										}
										lwOpt.Options = append(lwOpt.Options, maxDurOpt)
									} else if lwOptName == "ABORT_AFTER_WAIT" {
										abortVal := strings.ToUpper(p.curTok.Literal)
										var abortAfterWait string
										switch abortVal {
										case "NONE":
											abortAfterWait = "None"
										case "SELF":
											abortAfterWait = "Self"
										case "BLOCKERS":
											abortAfterWait = "Blockers"
										default:
											abortAfterWait = abortVal
										}
										p.nextToken()
										lwOpt.Options = append(lwOpt.Options, &ast.LowPriorityLockWaitAbortAfterWaitOption{
											OptionKind:     "AbortAfterWait",
											AbortAfterWait: abortAfterWait,
										})
									}
									if p.curTok.Type == TokenComma {
										p.nextToken()
									}
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken() // consume inner )
								}
								opt.LowPriorityLockWaitOption = lwOpt
							}
						}
						if p.curTok.Type == TokenRParen {
							p.nextToken() // consume outer )
						}
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
				case "DATA_COMPRESSION":
					compLevel := strings.ToUpper(p.curTok.Literal)
					var compressionLevel string
					switch compLevel {
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
					default:
						compressionLevel = compLevel
					}
					p.nextToken()
					opt := &ast.DataCompressionOption{
						OptionKind:       "DataCompression",
						CompressionLevel: compressionLevel,
					}
					// Check for ON PARTITIONS (...)
					if p.curTok.Type == TokenOn {
						p.nextToken() // consume ON
						if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
							p.nextToken() // consume PARTITIONS
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									pr := &ast.CompressionPartitionRange{}
									if p.curTok.Type == TokenNumber {
										pr.From = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
										p.nextToken()
									}
									// Check for TO range
									if strings.ToUpper(p.curTok.Literal) == "TO" {
										p.nextToken() // consume TO
										if p.curTok.Type == TokenNumber {
											pr.To = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
											p.nextToken()
										}
									}
									opt.PartitionRanges = append(opt.PartitionRanges, pr)
									if p.curTok.Type == TokenComma {
										p.nextToken()
									}
								}
								if p.curTok.Type == TokenRParen {
									p.nextToken() // consume )
								}
							}
						}
					}
					stmt.IndexOptions = append(stmt.IndexOptions, opt)
				default:
					// Skip unknown options
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
	}

	return stmt, nil
}

func (p *Parser) parseAlterTableChangeTrackingStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableChangeTrackingModificationStatement, error) {
	stmt := &ast.AlterTableChangeTrackingModificationStatement{
		SchemaObjectName:    tableName,
		TrackColumnsUpdated: "NotSet",
	}

	// Parse ENABLE or DISABLE
	if strings.ToUpper(p.curTok.Literal) == "ENABLE" {
		stmt.IsEnable = true
	}
	p.nextToken() // consume ENABLE/DISABLE

	// Consume CHANGE_TRACKING
	p.nextToken()

	// Check for WITH
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optionName := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				if optionName == "TRACK_COLUMNS_UPDATED" {
					valueUpper := strings.ToUpper(p.curTok.Literal)
					if valueUpper == "ON" {
						stmt.TrackColumnsUpdated = "On"
					} else if valueUpper == "OFF" {
						stmt.TrackColumnsUpdated = "Off"
					}
					p.nextToken()
				} else {
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
	}

	return stmt, nil
}

func (p *Parser) parseAlterAvailabilityGroupStatement() (*ast.AlterAvailabilityGroupStatement, error) {
	// Consume AVAILABILITY
	p.nextToken()

	// Expect GROUP
	if strings.ToUpper(p.curTok.Literal) != "GROUP" {
		return nil, fmt.Errorf("expected GROUP after AVAILABILITY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterAvailabilityGroupStatement{}

	// Parse group name
	stmt.Name = p.parseIdentifier()

	// Determine the action type
	actionKeyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	switch actionKeyword {
	case "JOIN":
		stmt.StatementType = "Action"
		stmt.Action = &ast.AlterAvailabilityGroupAction{ActionType: "Join"}
	case "ADD":
		// ADD DATABASE or ADD REPLICA
		nextKeyword := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		if nextKeyword == "DATABASE" {
			stmt.StatementType = "AddDatabase"
			stmt.Databases = p.parseIdentifierList()
		} else if nextKeyword == "REPLICA" {
			stmt.StatementType = "AddReplica"
			// Expect ON
			if strings.ToUpper(p.curTok.Literal) == "ON" {
				p.nextToken()
			}
			stmt.Replicas = p.parseAvailabilityReplicas()
		}
	case "REMOVE":
		// REMOVE DATABASE or REMOVE REPLICA
		nextKeyword := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		if nextKeyword == "DATABASE" {
			stmt.StatementType = "RemoveDatabase"
			stmt.Databases = p.parseIdentifierList()
		} else if nextKeyword == "REPLICA" {
			stmt.StatementType = "RemoveReplica"
			// Expect ON
			if strings.ToUpper(p.curTok.Literal) == "ON" {
				p.nextToken()
			}
			stmt.Replicas = p.parseAvailabilityReplicasServerOnly()
		}
	case "MODIFY":
		// MODIFY REPLICA
		nextKeyword := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		if nextKeyword == "REPLICA" {
			stmt.StatementType = "ModifyReplica"
			// Expect ON
			if strings.ToUpper(p.curTok.Literal) == "ON" {
				p.nextToken()
			}
			stmt.Replicas = p.parseAvailabilityReplicas()
		}
	case "SET":
		stmt.StatementType = "Set"
		// Parse SET options
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optName := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				if optName == "REQUIRED_COPIES_TO_COMMIT" {
					val, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					stmt.Options = append(stmt.Options, &ast.LiteralAvailabilityGroupOption{
						OptionKind: "RequiredCopiesToCommit",
						Value:      val,
					})
				} else {
					// Skip unknown options
					if p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen {
						p.nextToken()
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
	case "FAILOVER":
		stmt.StatementType = "Action"
		action := &ast.AlterAvailabilityGroupFailoverAction{ActionType: "Failover"}
		// Check for WITH clause
		if p.curTok.Type == TokenWith || strings.ToUpper(p.curTok.Literal) == "WITH" {
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optName := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken()
					}
					if optName == "TARGET" {
						val, err := p.parseScalarExpression()
						if err != nil {
							return nil, err
						}
						action.Options = append(action.Options, &ast.AlterAvailabilityGroupFailoverOption{
							OptionKind: "Target",
							Value:      val,
						})
					} else {
						// Skip unknown options
						if p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen {
							p.nextToken()
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
		stmt.Action = action
	case "FORCE_FAILOVER_ALLOW_DATA_LOSS":
		stmt.StatementType = "Action"
		stmt.Action = &ast.AlterAvailabilityGroupAction{ActionType: "ForceFailoverAllowDataLoss"}
	case "ONLINE":
		stmt.StatementType = "Action"
		stmt.Action = &ast.AlterAvailabilityGroupAction{ActionType: "Online"}
	case "OFFLINE":
		stmt.StatementType = "Action"
		stmt.Action = &ast.AlterAvailabilityGroupAction{ActionType: "Offline"}
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

// parseIdentifierList parses a comma-separated list of identifiers
func (p *Parser) parseIdentifierList() []*ast.Identifier {
	var ids []*ast.Identifier
	for {
		ids = append(ids, p.parseIdentifier())
		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}
	return ids
}

// parseAvailabilityReplicas parses replica definitions with full options
func (p *Parser) parseAvailabilityReplicas() []*ast.AvailabilityReplica {
	var replicas []*ast.AvailabilityReplica
	for {
		replica := &ast.AvailabilityReplica{}

		// Parse server name (string literal)
		if p.curTok.Type == TokenString {
			replica.ServerName, _ = p.parseStringLiteral()
		}

		// Parse WITH clause for replica options
		if p.curTok.Type == TokenWith || strings.ToUpper(p.curTok.Literal) == "WITH" {
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optName := strings.ToUpper(p.curTok.Literal)
					p.nextToken() // consume option name

					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}

					switch optName {
					case "AVAILABILITY_MODE":
						modeStr := strings.ToUpper(p.curTok.Literal)
						p.nextToken()
						// Handle SYNCHRONOUS_COMMIT or ASYNCHRONOUS_COMMIT
						if p.curTok.Type == TokenIdent && strings.HasPrefix(strings.ToUpper(p.curTok.Literal), "_") {
							modeStr += strings.ToUpper(p.curTok.Literal)
							p.nextToken()
						}
						var mode string
						switch modeStr {
						case "SYNCHRONOUS_COMMIT":
							mode = "SynchronousCommit"
						case "ASYNCHRONOUS_COMMIT":
							mode = "AsynchronousCommit"
						default:
							mode = modeStr
						}
						replica.Options = append(replica.Options, &ast.AvailabilityModeReplicaOption{
							OptionKind: "AvailabilityMode",
							Value:      mode,
						})
					case "FAILOVER_MODE":
						modeStr := strings.ToUpper(p.curTok.Literal)
						p.nextToken()
						var mode string
						switch modeStr {
						case "AUTOMATIC":
							mode = "Automatic"
						case "MANUAL":
							mode = "Manual"
						default:
							mode = modeStr
						}
						replica.Options = append(replica.Options, &ast.FailoverModeReplicaOption{
							OptionKind: "FailoverMode",
							Value:      mode,
						})
					case "ENDPOINT_URL":
						val, _ := p.parseScalarExpression()
						replica.Options = append(replica.Options, &ast.LiteralReplicaOption{
							OptionKind: "EndpointUrl",
							Value:      val,
						})
					case "SESSION_TIMEOUT":
						val, _ := p.parseScalarExpression()
						replica.Options = append(replica.Options, &ast.LiteralReplicaOption{
							OptionKind: "SessionTimeout",
							Value:      val,
						})
					case "APPLY_DELAY":
						val, _ := p.parseScalarExpression()
						replica.Options = append(replica.Options, &ast.LiteralReplicaOption{
							OptionKind: "ApplyDelay",
							Value:      val,
						})
					case "PRIMARY_ROLE":
						if p.curTok.Type == TokenLParen {
							p.nextToken() // consume (
							for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
								innerOpt := strings.ToUpper(p.curTok.Literal)
								p.nextToken()
								if p.curTok.Type == TokenEquals {
									p.nextToken()
								}
								if innerOpt == "ALLOW_CONNECTIONS" {
									connMode := strings.ToUpper(p.curTok.Literal)
									p.nextToken()
									var mode string
									switch connMode {
									case "READ_WRITE":
										mode = "ReadWrite"
									case "ALL":
										mode = "All"
									default:
										mode = connMode
									}
									replica.Options = append(replica.Options, &ast.PrimaryRoleReplicaOption{
										OptionKind:       "PrimaryRole",
										AllowConnections: mode,
									})
								}
								if p.curTok.Type == TokenComma {
									p.nextToken()
								}
							}
							if p.curTok.Type == TokenRParen {
								p.nextToken()
							}
						}
					case "SECONDARY_ROLE":
						if p.curTok.Type == TokenLParen {
							p.nextToken() // consume (
							for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
								innerOpt := strings.ToUpper(p.curTok.Literal)
								p.nextToken()
								if p.curTok.Type == TokenEquals {
									p.nextToken()
								}
								if innerOpt == "ALLOW_CONNECTIONS" {
									connMode := strings.ToUpper(p.curTok.Literal)
									p.nextToken()
									var mode string
									switch connMode {
									case "NO":
										mode = "No"
									case "READ_ONLY":
										mode = "ReadOnly"
									case "ALL":
										mode = "All"
									default:
										mode = connMode
									}
									replica.Options = append(replica.Options, &ast.SecondaryRoleReplicaOption{
										OptionKind:       "SecondaryRole",
										AllowConnections: mode,
									})
								}
								if p.curTok.Type == TokenComma {
									p.nextToken()
								}
							}
							if p.curTok.Type == TokenRParen {
								p.nextToken()
							}
						}
					default:
						// Skip unknown options
						if p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen {
							p.nextToken()
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

		replicas = append(replicas, replica)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		} else {
			break
		}
	}
	return replicas
}

// parseAvailabilityReplicasServerOnly parses replicas with only server names (for REMOVE REPLICA)
func (p *Parser) parseAvailabilityReplicasServerOnly() []*ast.AvailabilityReplica {
	var replicas []*ast.AvailabilityReplica
	for {
		replica := &ast.AvailabilityReplica{}
		if p.curTok.Type == TokenString {
			replica.ServerName, _ = p.parseStringLiteral()
		}
		replicas = append(replicas, replica)
		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}
	return replicas
}
