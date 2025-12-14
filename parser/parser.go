// Package parser provides T-SQL parsing functionality.
package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/kyleconroy/teesql/ast"
)

// Parse parses T-SQL from the given reader and returns an AST Script.
func Parse(ctx context.Context, r io.Reader) (*ast.Script, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	// For now, return an empty script for empty input
	if len(data) == 0 {
		return &ast.Script{}, nil
	}

	p := newParser(string(data))
	return p.parseScript()
}

// Parser holds the parsing state.
type Parser struct {
	lexer   *Lexer
	curTok  Token
	peekTok Token
}

func newParser(input string) *Parser {
	p := &Parser{lexer: NewLexer(input)}
	// Read two tokens to initialize curTok and peekTok
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.lexer.NextToken()
}

func (p *Parser) parseScript() (*ast.Script, error) {
	script := &ast.Script{}

	// Parse all batches (separated by GO)
	for p.curTok.Type != TokenEOF {
		batch, err := p.parseBatch()
		if err != nil {
			return nil, err
		}
		if batch != nil && len(batch.Statements) > 0 {
			script.Batches = append(script.Batches, batch)
		}
	}

	return script, nil
}

func (p *Parser) parseBatch() (*ast.Batch, error) {
	batch := &ast.Batch{}

	for p.curTok.Type != TokenEOF {
		// Stop at GO statements (batch separators)
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			p.nextToken()
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			batch.Statements = append(batch.Statements, stmt)
		}
	}

	return batch, nil
}

func (p *Parser) parseStatement() (ast.Statement, error) {
	switch p.curTok.Type {
	case TokenSelect, TokenLParen:
		return p.parseSelectStatement()
	case TokenInsert:
		return p.parseInsertStatement()
	case TokenUpdate:
		return p.parseUpdateStatement()
	case TokenDelete:
		return p.parseDeleteStatement()
	case TokenDeclare:
		return p.parseDeclareVariableStatement()
	case TokenSet:
		return p.parseSetVariableStatement()
	case TokenIf:
		return p.parseIfStatement()
	case TokenWhile:
		return p.parseWhileStatement()
	case TokenBegin:
		return p.parseBeginStatement()
	case TokenCreate:
		return p.parseCreateStatement()
	case TokenExec, TokenExecute:
		return p.parseExecuteStatement()
	case TokenPrint:
		return p.parsePrintStatement()
	case TokenThrow:
		return p.parseThrowStatement()
	case TokenAlter:
		return p.parseAlterStatement()
	case TokenRevert:
		return p.parseRevertStatement()
	case TokenDrop:
		return p.parseDropStatement()
	case TokenReturn:
		return p.parseReturnStatement()
	case TokenBreak:
		return p.parseBreakStatement()
	case TokenContinue:
		return p.parseContinueStatement()
	case TokenGrant:
		return p.parseGrantStatement()
	case TokenCommit:
		return p.parseCommitTransactionStatement()
	case TokenRollback:
		return p.parseRollbackTransactionStatement()
	case TokenSave:
		return p.parseSaveTransactionStatement()
	case TokenWaitfor:
		return p.parseWaitForStatement()
	case TokenGoto:
		return p.parseGotoStatement()
	case TokenSemicolon:
		p.nextToken()
		return nil, nil
	case TokenIdent:
		// Check for label (identifier followed by colon)
		return p.parseLabelOrError()
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.curTok.Literal)
	}
}

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
		return p.parseDropDatabaseScopedStatement()
	}

	return nil, fmt.Errorf("unexpected token after DROP: %s", p.curTok.Literal)
}

func (p *Parser) parseDropDatabaseScopedStatement() (ast.Statement, error) {
	// Consume DATABASE
	p.nextToken()

	if p.curTok.Type != TokenScoped {
		return nil, fmt.Errorf("expected SCOPED after DATABASE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume SCOPED

	if p.curTok.Type == TokenCredential {
		return p.parseDropCredentialStatement(true)
	}

	return nil, fmt.Errorf("unexpected token after SCOPED: %s", p.curTok.Literal)
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

func (p *Parser) parseAlterStatement() (ast.Statement, error) {
	// Consume ALTER
	p.nextToken()

	// Check what type of ALTER statement this is
	if p.curTok.Type == TokenTable {
		return p.parseAlterTableStatement()
	}

	return nil, fmt.Errorf("unexpected token after ALTER: %s", p.curTok.Literal)
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

	return nil, fmt.Errorf("unexpected token in ALTER TABLE: %s", p.curTok.Literal)
}

func (p *Parser) parseAlterTableDropStatement(tableName *ast.SchemaObjectName) (*ast.AlterTableDropTableElementStatement, error) {
	// Consume DROP
	p.nextToken()

	stmt := &ast.AlterTableDropTableElementStatement{
		SchemaObjectName: tableName,
	}

	// Parse the element type and name
	var elementType string
	switch p.curTok.Type {
	case TokenIndex:
		elementType = "Index"
		p.nextToken()
	default:
		return nil, fmt.Errorf("unexpected token after DROP: %s", p.curTok.Literal)
	}

	// Parse the element name
	if p.curTok.Type != TokenIdent {
		return nil, fmt.Errorf("expected identifier after %s, got %s", elementType, p.curTok.Literal)
	}

	element := &ast.AlterTableDropTableElement{
		TableElementType: elementType,
		Name: &ast.Identifier{
			Value:     p.curTok.Literal,
			QuoteType: "NotQuoted",
		},
		IsIfExists: false,
	}
	p.nextToken()

	stmt.AlterTableDropTableElements = append(stmt.AlterTableDropTableElements, element)

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parsePrintStatement() (*ast.PrintStatement, error) {
	// Consume PRINT
	p.nextToken()

	// Parse expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.PrintStatement{Expression: expr}, nil
}

func (p *Parser) parseThrowStatement() (*ast.ThrowStatement, error) {
	// Consume THROW
	p.nextToken()

	stmt := &ast.ThrowStatement{}

	// THROW can be used without arguments (re-throw)
	if p.curTok.Type == TokenSemicolon || p.curTok.Type == TokenEOF ||
		p.curTok.Type == TokenSelect || p.curTok.Type == TokenPrint || p.curTok.Type == TokenThrow {
		return stmt, nil
	}

	// Parse error number
	errNum, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.ErrorNumber = errNum

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected comma after error number, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse message
	msg, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Message = msg

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected comma after message, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse state
	state, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.State = state

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseSelectStatement() (*ast.SelectStatement, error) {
	stmt := &ast.SelectStatement{}

	// Parse query expression (handles UNION, parens, etc.)
	qe, into, err := p.parseQueryExpressionWithInto()
	if err != nil {
		return nil, err
	}
	stmt.QueryExpression = qe
	stmt.Into = into

	// Parse optional OPTION clause
	if p.curTok.Type == TokenOption {
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

	return stmt, nil
}

func (p *Parser) parseQueryExpression() (ast.QueryExpression, error) {
	qe, _, err := p.parseQueryExpressionWithInto()
	return qe, err
}

func (p *Parser) parseQueryExpressionWithInto() (ast.QueryExpression, *ast.SchemaObjectName, error) {
	// Parse primary query expression (could be SELECT or parenthesized)
	left, into, err := p.parsePrimaryQueryExpression()
	if err != nil {
		return nil, nil, err
	}

	// Track if we have any binary operations
	hasBinaryOp := false

	// Check for binary operations (UNION, EXCEPT, INTERSECT)
	for p.curTok.Type == TokenUnion || p.curTok.Type == TokenExcept || p.curTok.Type == TokenIntersect {
		hasBinaryOp = true
		var opType string
		switch p.curTok.Type {
		case TokenUnion:
			opType = "Union"
		case TokenExcept:
			opType = "Except"
		case TokenIntersect:
			opType = "Intersect"
		}
		p.nextToken()

		// Check for ALL
		all := false
		if p.curTok.Type == TokenAll {
			all = true
			p.nextToken()
		}

		// Parse the right side
		right, rightInto, err := p.parsePrimaryQueryExpression()
		if err != nil {
			return nil, nil, err
		}

		// INTO can only appear in the first query of a UNION
		if rightInto != nil && into == nil {
			into = rightInto
		}

		bqe := &ast.BinaryQueryExpression{
			BinaryQueryExpressionType: opType,
			All:                       all,
			FirstQueryExpression:      left,
			SecondQueryExpression:     right,
		}

		left = bqe
	}

	// Parse ORDER BY after all UNION operations
	if p.curTok.Type == TokenOrder {
		obc, err := p.parseOrderByClause()
		if err != nil {
			return nil, nil, err
		}

		if hasBinaryOp {
			// Attach to BinaryQueryExpression
			if bqe, ok := left.(*ast.BinaryQueryExpression); ok {
				bqe.OrderByClause = obc
			}
		} else {
			// Attach to QuerySpecification
			if qs, ok := left.(*ast.QuerySpecification); ok {
				qs.OrderByClause = obc
			}
		}
	}

	return left, into, nil
}

func (p *Parser) parsePrimaryQueryExpression() (ast.QueryExpression, *ast.SchemaObjectName, error) {
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		qe, into, err := p.parseQueryExpressionWithInto()
		if err != nil {
			return nil, nil, err
		}
		if p.curTok.Type != TokenRParen {
			return nil, nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
		return &ast.QueryParenthesisExpression{QueryExpression: qe}, into, nil
	}

	return p.parseQuerySpecificationWithInto()
}

func (p *Parser) parseQuerySpecificationWithInto() (*ast.QuerySpecification, *ast.SchemaObjectName, error) {
	qs, err := p.parseQuerySpecificationCore()
	if err != nil {
		return nil, nil, err
	}

	// Check for INTO clause after SELECT elements, before FROM
	var into *ast.SchemaObjectName
	if p.curTok.Type == TokenInto {
		p.nextToken() // consume INTO
		into, err = p.parseSchemaObjectName()
		if err != nil {
			return nil, nil, err
		}
	}

	// Parse optional FROM clause
	if p.curTok.Type == TokenFrom {
		fromClause, err := p.parseFromClause()
		if err != nil {
			return nil, nil, err
		}
		qs.FromClause = fromClause
	}

	// Parse optional WHERE clause
	if p.curTok.Type == TokenWhere {
		whereClause, err := p.parseWhereClause()
		if err != nil {
			return nil, nil, err
		}
		qs.WhereClause = whereClause
	}

	// Parse optional GROUP BY clause
	if p.curTok.Type == TokenGroup {
		groupByClause, err := p.parseGroupByClause()
		if err != nil {
			return nil, nil, err
		}
		qs.GroupByClause = groupByClause
	}

	// Parse optional HAVING clause
	if p.curTok.Type == TokenHaving {
		havingClause, err := p.parseHavingClause()
		if err != nil {
			return nil, nil, err
		}
		qs.HavingClause = havingClause
	}

	// Note: ORDER BY is parsed at the top level in parseQueryExpressionWithInto
	// to correctly handle UNION/EXCEPT/INTERSECT cases

	return qs, into, nil
}

func (p *Parser) parseQuerySpecificationCore() (*ast.QuerySpecification, error) {
	qs := &ast.QuerySpecification{
		UniqueRowFilter: "NotSpecified",
	}

	// Expect SELECT
	if p.curTok.Type != TokenSelect {
		return nil, fmt.Errorf("expected SELECT, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for ALL or DISTINCT
	if p.curTok.Type == TokenAll {
		qs.UniqueRowFilter = "All"
		p.nextToken()
	} else if p.curTok.Type == TokenDistinct {
		qs.UniqueRowFilter = "Distinct"
		p.nextToken()
	}

	// Check for TOP clause
	if p.curTok.Type == TokenTop {
		top, err := p.parseTopRowFilter()
		if err != nil {
			return nil, err
		}
		qs.TopRowFilter = top
	}

	// Parse select elements
	elements, err := p.parseSelectElements()
	if err != nil {
		return nil, err
	}
	qs.SelectElements = elements

	return qs, nil
}

func (p *Parser) parseTopRowFilter() (*ast.TopRowFilter, error) {
	// Consume TOP
	p.nextToken()

	top := &ast.TopRowFilter{}

	// Check for parenthesized expression
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		top.Expression = expr
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	} else {
		// Parse literal expression
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		top.Expression = expr
	}

	// Check for PERCENT
	if p.curTok.Type == TokenPercent {
		top.Percent = true
		p.nextToken()
	}

	// Check for WITH TIES
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenTies {
			top.WithTies = true
			p.nextToken()
		}
	}

	return top, nil
}

func (p *Parser) parseSelectElements() ([]ast.SelectElement, error) {
	var elements []ast.SelectElement

	for {
		elem, err := p.parseSelectElement()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return elements, nil
}

func (p *Parser) parseSelectElement() (ast.SelectElement, error) {
	// Check for *
	if p.curTok.Type == TokenStar {
		p.nextToken()
		return &ast.SelectStarExpression{}, nil
	}

	// Otherwise parse a scalar expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	sse := &ast.SelectScalarExpression{Expression: expr}

	// Check for column alias: [alias], AS alias, or just alias
	if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '[' {
		// Bracketed alias without AS
		alias := p.parseIdentifier()
		sse.ColumnName = &ast.IdentifierOrValueExpression{
			Value:      alias.Value,
			Identifier: alias,
		}
	} else if p.curTok.Type == TokenAs {
		p.nextToken() // consume AS
		alias := p.parseIdentifier()
		sse.ColumnName = &ast.IdentifierOrValueExpression{
			Value:      alias.Value,
			Identifier: alias,
		}
	} else if p.curTok.Type == TokenIdent {
		// Check if this is an alias (not a keyword that starts a new clause)
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "FROM" && upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "ORDER" && upper != "OPTION" && upper != "INTO" && upper != "UNION" && upper != "EXCEPT" && upper != "INTERSECT" && upper != "GO" {
			alias := p.parseIdentifier()
			sse.ColumnName = &ast.IdentifierOrValueExpression{
				Value:      alias.Value,
				Identifier: alias,
			}
		}
	}

	return sse, nil
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	literal := p.curTok.Literal
	quoteType := "NotQuoted"

	// Handle bracketed identifiers
	if len(literal) >= 2 && literal[0] == '[' && literal[len(literal)-1] == ']' {
		quoteType = "SquareBracket"
		literal = literal[1 : len(literal)-1]
	}

	id := &ast.Identifier{
		Value:     literal,
		QuoteType: quoteType,
	}
	p.nextToken()
	return id
}

func (p *Parser) parseScalarExpression() (ast.ScalarExpression, error) {
	return p.parseAdditiveExpression()
}

func (p *Parser) parseAdditiveExpression() (ast.ScalarExpression, error) {
	left, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenPlus || p.curTok.Type == TokenMinus {
		var opType string
		if p.curTok.Type == TokenPlus {
			opType = "Add"
		} else {
			opType = "Subtract"
		}
		p.nextToken()

		right, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryExpression{
			BinaryExpressionType: opType,
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}

	return left, nil
}

func (p *Parser) parsePrimaryExpression() (ast.ScalarExpression, error) {
	switch p.curTok.Type {
	case TokenNull:
		p.nextToken()
		return &ast.NullLiteral{LiteralType: "Null", Value: "null"}, nil
	case TokenDefault:
		val := p.curTok.Literal
		p.nextToken()
		return &ast.DefaultLiteral{LiteralType: "Default", Value: val}, nil
	case TokenMinus:
		p.nextToken()
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{UnaryExpressionType: "Negative", Expression: expr}, nil
	case TokenPlus:
		p.nextToken()
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{UnaryExpressionType: "Positive", Expression: expr}, nil
	case TokenIdent:
		// Check if it's a variable reference (starts with @)
		if strings.HasPrefix(p.curTok.Literal, "@") {
			name := p.curTok.Literal
			p.nextToken()
			return &ast.VariableReference{Name: name}, nil
		}
		// Check for N-prefixed national string (N'...')
		if strings.ToUpper(p.curTok.Literal) == "N" && p.peekTok.Type == TokenString {
			p.nextToken() // consume N
			return p.parseNationalStringLiteral()
		}
		return p.parseColumnReference()
	case TokenNumber:
		val := p.curTok.Literal
		p.nextToken()
		// Check if it's a decimal number
		if strings.Contains(val, ".") {
			return &ast.NumericLiteral{LiteralType: "Numeric", Value: val}, nil
		}
		return &ast.IntegerLiteral{LiteralType: "Integer", Value: val}, nil
	case TokenString:
		return p.parseStringLiteral()
	case TokenLBrace:
		return p.parseOdbcLiteral()
	case TokenLParen:
		// Parenthesized expression or subquery
		p.nextToken()
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken()
		return &ast.ParenthesisExpression{Expression: expr}, nil
	default:
		return nil, fmt.Errorf("unexpected token in expression: %s", p.curTok.Literal)
	}
}

func (p *Parser) parseOdbcLiteral() (*ast.OdbcLiteral, error) {
	// Consume {
	p.nextToken()

	// Expect "guid" identifier
	if p.curTok.Type != TokenIdent || strings.ToLower(p.curTok.Literal) != "guid" {
		return nil, fmt.Errorf("expected guid in ODBC literal, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for N prefix for national string
	isNational := false
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "N" {
		isNational = true
		p.nextToken()
	}

	// Expect string literal
	if p.curTok.Type != TokenString {
		return nil, fmt.Errorf("expected string in ODBC literal, got %s", p.curTok.Literal)
	}

	raw := p.curTok.Literal
	p.nextToken()

	// Remove surrounding quotes
	value := raw
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		value = raw[1 : len(raw)-1]
	}

	// Consume }
	if p.curTok.Type != TokenRBrace {
		return nil, fmt.Errorf("expected } in ODBC literal, got %s", p.curTok.Literal)
	}
	p.nextToken()

	return &ast.OdbcLiteral{
		LiteralType:     "Odbc",
		OdbcLiteralType: "Guid",
		IsNational:      isNational,
		Value:           value,
	}, nil
}

func (p *Parser) parseStringLiteral() (*ast.StringLiteral, error) {
	raw := p.curTok.Literal
	p.nextToken()

	// Remove surrounding quotes and handle escaped quotes
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		inner := raw[1 : len(raw)-1]
		// Replace escaped quotes
		value := strings.ReplaceAll(inner, "''", "'")
		return &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    false,
			IsLargeObject: false,
			Value:         value,
		}, nil
	}

	return &ast.StringLiteral{
		LiteralType:   "String",
		IsNational:    false,
		IsLargeObject: false,
		Value:         raw,
	}, nil
}

func (p *Parser) parseNationalStringLiteral() (*ast.StringLiteral, error) {
	raw := p.curTok.Literal
	p.nextToken()

	// Remove surrounding quotes and handle escaped quotes
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		inner := raw[1 : len(raw)-1]
		// Replace escaped quotes
		value := strings.ReplaceAll(inner, "''", "'")
		return &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    true,
			IsLargeObject: false,
			Value:         value,
		}, nil
	}

	return &ast.StringLiteral{
		LiteralType:   "String",
		IsNational:    true,
		IsLargeObject: false,
		Value:         raw,
	}, nil
}

func (p *Parser) parseColumnReference() (*ast.ColumnReferenceExpression, error) {
	var identifiers []*ast.Identifier

	for {
		if p.curTok.Type != TokenIdent {
			break
		}

		id := &ast.Identifier{
			Value:     p.curTok.Literal,
			QuoteType: "NotQuoted",
		}
		identifiers = append(identifiers, id)
		p.nextToken()

		if p.curTok.Type != TokenDot {
			break
		}
		p.nextToken() // consume dot
	}

	return &ast.ColumnReferenceExpression{
		ColumnType: "Regular",
		MultiPartIdentifier: &ast.MultiPartIdentifier{
			Count:       len(identifiers),
			Identifiers: identifiers,
		},
	}, nil
}

func (p *Parser) parseFromClause() (*ast.FromClause, error) {
	// Consume FROM
	if p.curTok.Type != TokenFrom {
		return nil, fmt.Errorf("expected FROM, got %s", p.curTok.Literal)
	}
	p.nextToken()

	fc := &ast.FromClause{}

	// Parse table references
	for {
		ref, err := p.parseTableReference()
		if err != nil {
			return nil, err
		}
		fc.TableReferences = append(fc.TableReferences, ref)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return fc, nil
}

func (p *Parser) parseTableReference() (ast.TableReference, error) {
	// Parse the base table reference
	baseRef, err := p.parseNamedTableReference()
	if err != nil {
		return nil, err
	}
	var left ast.TableReference = baseRef

	// Check for JOINs
	for {
		// Check for CROSS JOIN
		if p.curTok.Type == TokenCross {
			p.nextToken() // consume CROSS
			if p.curTok.Type != TokenJoin {
				return nil, fmt.Errorf("expected JOIN after CROSS, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume JOIN

			right, err := p.parseNamedTableReference()
			if err != nil {
				return nil, err
			}

			left = &ast.UnqualifiedJoin{
				UnqualifiedJoinType:  "CrossJoin",
				FirstTableReference:  left,
				SecondTableReference: right,
			}
			continue
		}

		// Check for qualified JOINs (INNER, LEFT, RIGHT, FULL)
		joinType := ""
		if p.curTok.Type == TokenInner {
			joinType = "Inner"
			p.nextToken()
		} else if p.curTok.Type == TokenLeft {
			joinType = "LeftOuter"
			p.nextToken()
			if p.curTok.Type == TokenOuter {
				p.nextToken()
			}
		} else if p.curTok.Type == TokenRight {
			joinType = "RightOuter"
			p.nextToken()
			if p.curTok.Type == TokenOuter {
				p.nextToken()
			}
		} else if p.curTok.Type == TokenFull {
			joinType = "FullOuter"
			p.nextToken()
			if p.curTok.Type == TokenOuter {
				p.nextToken()
			}
		} else if p.curTok.Type == TokenJoin {
			joinType = "Inner"
		}

		if joinType == "" {
			break
		}

		if p.curTok.Type != TokenJoin {
			return nil, fmt.Errorf("expected JOIN, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume JOIN

		right, err := p.parseNamedTableReference()
		if err != nil {
			return nil, err
		}

		// Parse ON clause
		if p.curTok.Type != TokenOn {
			return nil, fmt.Errorf("expected ON after JOIN, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ON

		condition, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}

		left = &ast.QualifiedJoin{
			QualifiedJoinType:    joinType,
			FirstTableReference:  left,
			SecondTableReference: right,
			SearchCondition:      condition,
		}
	}

	return left, nil
}

func (p *Parser) parseNamedTableReference() (*ast.NamedTableReference, error) {
	ref := &ast.NamedTableReference{
		ForPath: false,
	}

	// Parse schema object name (potentially multi-part: db.schema.table)
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	ref.SchemaObject = son

	// Parse optional alias (AS alias or just alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
		if p.curTok.Type != TokenIdent {
			return nil, fmt.Errorf("expected identifier after AS, got %s", p.curTok.Literal)
		}
		ref.Alias = &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent {
		// Could be an alias without AS, but need to be careful not to consume keywords
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "ORDER" && upper != "OPTION" && upper != "GO" {
			ref.Alias = &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
			p.nextToken()
		}
	}

	return ref, nil
}

func (p *Parser) parseSchemaObjectName() (*ast.SchemaObjectName, error) {
	var identifiers []*ast.Identifier

	for {
		// Handle empty parts (e.g., myDb..table means myDb.<empty>.table)
		if p.curTok.Type == TokenDot {
			// Add an empty identifier for the missing part
			identifiers = append(identifiers, &ast.Identifier{
				Value:     "",
				QuoteType: "NotQuoted",
			})
			p.nextToken() // consume dot
			continue
		}

		if p.curTok.Type != TokenIdent {
			break
		}

		id := p.parseIdentifier()
		identifiers = append(identifiers, id)

		if p.curTok.Type != TokenDot {
			break
		}
		p.nextToken() // consume dot
	}

	if len(identifiers) == 0 {
		return nil, fmt.Errorf("expected identifier for schema object name")
	}

	// Filter out nil identifiers for the count and assignment
	var nonNilIdentifiers []*ast.Identifier
	for _, id := range identifiers {
		if id != nil {
			nonNilIdentifiers = append(nonNilIdentifiers, id)
		}
	}

	son := &ast.SchemaObjectName{
		Count:       len(identifiers),
		Identifiers: identifiers,
	}

	// Set the appropriate identifier fields based on count
	// server.database.schema.table (4 parts)
	// database.schema.table (3 parts)
	// schema.table (2 parts) - but with .., schema is nil
	// table (1 part)
	switch len(identifiers) {
	case 4:
		son.ServerIdentifier = identifiers[0]
		son.DatabaseIdentifier = identifiers[1]
		son.SchemaIdentifier = identifiers[2]
		son.BaseIdentifier = identifiers[3]
	case 3:
		son.DatabaseIdentifier = identifiers[0]
		son.SchemaIdentifier = identifiers[1]
		son.BaseIdentifier = identifiers[2]
	case 2:
		son.SchemaIdentifier = identifiers[0]
		son.BaseIdentifier = identifiers[1]
	case 1:
		son.BaseIdentifier = identifiers[0]
	}

	return son, nil
}

func (p *Parser) parseOptionClause() ([]*ast.OptimizerHint, error) {
	// Consume OPTION
	if p.curTok.Type != TokenOption {
		return nil, fmt.Errorf("expected OPTION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Consume (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected (, got %s", p.curTok.Literal)
	}
	p.nextToken()

	var hints []*ast.OptimizerHint

	// Parse hints
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenIdent {
			hintKind := convertHintKind(p.curTok.Literal)
			hints = append(hints, &ast.OptimizerHint{HintKind: hintKind})
			p.nextToken()
		} else if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			p.nextToken()
		}
	}

	// Consume )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return hints, nil
}

// convertHintKind converts hint identifiers to their canonical names
func convertHintKind(hint string) string {
	// Map common hint names
	hintMap := map[string]string{
		"IGNORE_NONCLUSTERED_COLUMNSTORE_INDEX": "IgnoreNonClusteredColumnStoreIndex",
	}
	upper := strings.ToUpper(hint)
	if mapped, ok := hintMap[upper]; ok {
		return mapped
	}
	return hint
}

func (p *Parser) parseWhereClause() (*ast.WhereClause, error) {
	// Consume WHERE
	p.nextToken()

	condition, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}

	return &ast.WhereClause{SearchCondition: condition}, nil
}

func (p *Parser) parseGroupByClause() (*ast.GroupByClause, error) {
	// Consume GROUP
	p.nextToken()

	if p.curTok.Type != TokenBy {
		return nil, fmt.Errorf("expected BY after GROUP, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume BY

	gbc := &ast.GroupByClause{
		GroupByOption: "None",
		All:           false,
	}

	// Check for ALL
	if p.curTok.Type == TokenAll {
		gbc.All = true
		p.nextToken()
	}

	// Parse grouping specifications
	for {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		spec := &ast.ExpressionGroupingSpecification{
			Expression:             expr,
			DistributedAggregation: false,
		}
		gbc.GroupingSpecifications = append(gbc.GroupingSpecifications, spec)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Check for WITH ROLLUP or WITH CUBE
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenRollup {
			gbc.GroupByOption = "Rollup"
			p.nextToken()
		} else if p.curTok.Type == TokenCube {
			gbc.GroupByOption = "Cube"
			p.nextToken()
		}
	}

	return gbc, nil
}

func (p *Parser) parseHavingClause() (*ast.HavingClause, error) {
	// Consume HAVING
	p.nextToken()

	condition, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}

	return &ast.HavingClause{SearchCondition: condition}, nil
}

func (p *Parser) parseOrderByClause() (*ast.OrderByClause, error) {
	// Consume ORDER
	p.nextToken()

	if p.curTok.Type != TokenBy {
		return nil, fmt.Errorf("expected BY after ORDER, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume BY

	obc := &ast.OrderByClause{}

	// Parse order by elements
	for {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		elem := &ast.ExpressionWithSortOrder{
			Expression: expr,
			SortOrder:  "NotSpecified",
		}

		// Check for ASC or DESC
		if p.curTok.Type == TokenAsc {
			elem.SortOrder = "Ascending"
			p.nextToken()
		} else if p.curTok.Type == TokenDesc {
			elem.SortOrder = "Descending"
			p.nextToken()
		}

		obc.OrderByElements = append(obc.OrderByElements, elem)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return obc, nil
}

func (p *Parser) parseBooleanExpression() (ast.BooleanExpression, error) {
	return p.parseBooleanOrExpression()
}

func (p *Parser) parseBooleanOrExpression() (ast.BooleanExpression, error) {
	left, err := p.parseBooleanAndExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenOr {
		p.nextToken() // consume OR

		right, err := p.parseBooleanAndExpression()
		if err != nil {
			return nil, err
		}

		left = &ast.BooleanBinaryExpression{
			BinaryExpressionType: "Or",
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}

	return left, nil
}

func (p *Parser) parseBooleanAndExpression() (ast.BooleanExpression, error) {
	left, err := p.parseBooleanPrimaryExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenAnd {
		p.nextToken() // consume AND

		right, err := p.parseBooleanPrimaryExpression()
		if err != nil {
			return nil, err
		}

		left = &ast.BooleanBinaryExpression{
			BinaryExpressionType: "And",
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}

	return left, nil
}

func (p *Parser) parseBooleanPrimaryExpression() (ast.BooleanExpression, error) {
	// Parse left scalar expression
	left, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Check for comparison operator
	var compType string
	switch p.curTok.Type {
	case TokenEquals:
		compType = "Equals"
	case TokenNotEqual:
		compType = "NotEqualToBrackets"
	case TokenLessThan:
		compType = "LessThan"
	case TokenGreaterThan:
		compType = "GreaterThan"
	case TokenLessOrEqual:
		compType = "LessThanOrEqualTo"
	case TokenGreaterOrEqual:
		compType = "GreaterThanOrEqualTo"
	default:
		return nil, fmt.Errorf("expected comparison operator, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse right scalar expression
	right, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return &ast.BooleanComparisonExpression{
		ComparisonType:   compType,
		FirstExpression:  left,
		SecondExpression: right,
	}, nil
}

// ======================= New Statement Parsing Functions =======================

func (p *Parser) parseInsertStatement() (*ast.InsertStatement, error) {
	// Consume INSERT
	p.nextToken()

	stmt := &ast.InsertStatement{
		InsertSpecification: &ast.InsertSpecification{
			InsertOption: "None",
		},
	}

	// Check for INTO or OVER
	if p.curTok.Type == TokenInto {
		stmt.InsertSpecification.InsertOption = "Into"
		p.nextToken()
	} else if p.curTok.Type == TokenOver {
		stmt.InsertSpecification.InsertOption = "Over"
		p.nextToken()
	}

	// Parse target
	target, err := p.parseDMLTarget()
	if err != nil {
		return nil, err
	}
	stmt.InsertSpecification.Target = target

	// Parse optional column list
	if p.curTok.Type == TokenLParen {
		cols, err := p.parseColumnList()
		if err != nil {
			return nil, err
		}
		stmt.InsertSpecification.Columns = cols
	}

	// Parse insert source
	source, err := p.parseInsertSource()
	if err != nil {
		return nil, err
	}
	stmt.InsertSpecification.InsertSource = source

	// Parse optional OPTION clause
	if p.curTok.Type == TokenOption {
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

	return stmt, nil
}

func (p *Parser) parseDMLTarget() (ast.TableReference, error) {
	// Check for variable
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		name := p.curTok.Literal
		p.nextToken()
		return &ast.VariableTableReference{
			Variable: &ast.VariableReference{Name: name},
			ForPath:  false,
		}, nil
	}

	// Check for OPENROWSET
	if p.curTok.Type == TokenOpenRowset {
		return p.parseOpenRowset()
	}

	// Parse schema object name
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}

	// Check for function call (has parentheses)
	if p.curTok.Type == TokenLParen {
		params, err := p.parseFunctionParameters()
		if err != nil {
			return nil, err
		}
		return &ast.SchemaObjectFunctionTableReference{
			SchemaObject: son,
			Parameters:   params,
			ForPath:      false,
		}, nil
	}

	ref := &ast.NamedTableReference{
		SchemaObject: son,
		ForPath:      false,
	}

	// Check for table hints WITH (...)
	if p.curTok.Type == TokenWith {
		hints, err := p.parseTableHints()
		if err != nil {
			return nil, err
		}
		ref.TableHints = hints
	}

	return ref, nil
}

func (p *Parser) parseOpenRowset() (*ast.InternalOpenRowset, error) {
	// Consume OPENROWSET
	p.nextToken()

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after OPENROWSET, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse identifier
	if p.curTok.Type != TokenIdent {
		return nil, fmt.Errorf("expected identifier in OPENROWSET, got %s", p.curTok.Literal)
	}
	id := &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
	p.nextToken()

	var varArgs []ast.ScalarExpression
	for p.curTok.Type == TokenComma {
		p.nextToken()
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		varArgs = append(varArgs, expr)
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in OPENROWSET, got %s", p.curTok.Literal)
	}
	p.nextToken()

	return &ast.InternalOpenRowset{
		Identifier: id,
		VarArgs:    varArgs,
		ForPath:    false,
	}, nil
}

func (p *Parser) parseFunctionParameters() ([]ast.ScalarExpression, error) {
	// Consume (
	p.nextToken()

	var params []ast.ScalarExpression
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		params = append(params, expr)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
	}
	p.nextToken()

	return params, nil
}

func (p *Parser) parseTableHints() ([]*ast.TableHint, error) {
	// Consume WITH
	p.nextToken()

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
	}
	p.nextToken()

	var hints []*ast.TableHint
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenHoldlock || p.curTok.Type == TokenNowait {
			hintKind := convertTableHintKind(p.curTok.Literal)
			hints = append(hints, &ast.TableHint{HintKind: hintKind})
			p.nextToken()
		}
		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return hints, nil
}

func convertTableHintKind(hint string) string {
	hintMap := map[string]string{
		"HOLDLOCK": "HoldLock",
		"NOWAIT":   "NoWait",
		"NOLOCK":   "NoLock",
		"UPDLOCK":  "UpdLock",
		"XLOCK":    "XLock",
	}
	if mapped, ok := hintMap[strings.ToUpper(hint)]; ok {
		return mapped
	}
	return hint
}

func (p *Parser) parseColumnList() ([]*ast.ColumnReferenceExpression, error) {
	// Consume (
	p.nextToken()

	var cols []*ast.ColumnReferenceExpression
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		col, err := p.parseMultiPartIdentifierAsColumn()
		if err != nil {
			return nil, err
		}
		cols = append(cols, col)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
	}
	p.nextToken()

	return cols, nil
}

func (p *Parser) parseMultiPartIdentifierAsColumn() (*ast.ColumnReferenceExpression, error) {
	var identifiers []*ast.Identifier

	for {
		// Handle empty parts (e.g., ..a means two empty parts then a)
		if p.curTok.Type == TokenDot {
			identifiers = append(identifiers, &ast.Identifier{Value: "", QuoteType: "NotQuoted"})
			p.nextToken()
			continue
		}

		if p.curTok.Type != TokenIdent {
			break
		}

		id := p.parseIdentifier()
		identifiers = append(identifiers, id)

		if p.curTok.Type != TokenDot {
			break
		}
		p.nextToken()
	}

	return &ast.ColumnReferenceExpression{
		ColumnType: "Regular",
		MultiPartIdentifier: &ast.MultiPartIdentifier{
			Count:       len(identifiers),
			Identifiers: identifiers,
		},
	}, nil
}

func (p *Parser) parseInsertSource() (ast.InsertSource, error) {
	// Check for DEFAULT VALUES
	if p.curTok.Type == TokenDefault {
		p.nextToken()
		if p.curTok.Type == TokenValues {
			p.nextToken()
			return &ast.ValuesInsertSource{IsDefaultValues: true}, nil
		}
		return nil, fmt.Errorf("expected VALUES after DEFAULT, got %s", p.curTok.Literal)
	}

	// Check for VALUES (...)
	if p.curTok.Type == TokenValues {
		return p.parseValuesInsertSource()
	}

	// Check for EXEC/EXECUTE
	if p.curTok.Type == TokenExec || p.curTok.Type == TokenExecute {
		return p.parseExecuteInsertSource()
	}

	// Otherwise it's a SELECT
	qe, err := p.parseQueryExpression()
	if err != nil {
		return nil, err
	}
	return &ast.SelectInsertSource{Select: qe}, nil
}

func (p *Parser) parseValuesInsertSource() (*ast.ValuesInsertSource, error) {
	// Consume VALUES
	p.nextToken()

	source := &ast.ValuesInsertSource{IsDefaultValues: false}

	// Parse row values
	for {
		if p.curTok.Type != TokenLParen {
			break
		}
		row, err := p.parseRowValue()
		if err != nil {
			return nil, err
		}
		source.RowValues = append(source.RowValues, row)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	return source, nil
}

func (p *Parser) parseRowValue() (*ast.RowValue, error) {
	// Consume (
	p.nextToken()

	row := &ast.RowValue{}
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		row.ColumnValues = append(row.ColumnValues, expr)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
	}
	p.nextToken()

	return row, nil
}

func (p *Parser) parseExecuteInsertSource() (*ast.ExecuteInsertSource, error) {
	execSpec, err := p.parseExecuteSpecification()
	if err != nil {
		return nil, err
	}
	return &ast.ExecuteInsertSource{Execute: execSpec}, nil
}

func (p *Parser) parseExecuteSpecification() (*ast.ExecuteSpecification, error) {
	// Consume EXEC/EXECUTE
	p.nextToken()

	spec := &ast.ExecuteSpecification{}

	// Check for return variable assignment @var =
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		varName := p.curTok.Literal
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			spec.Variable = &ast.VariableReference{Name: varName}
			p.nextToken()
		} else {
			// It's actually the procedure variable
			spec.ExecutableEntity = &ast.ExecutableProcedureReference{
				ProcedureReference: &ast.ProcedureReferenceName{
					ProcedureVariable: &ast.VariableReference{Name: varName},
				},
			}
			return spec, nil
		}
	}

	// Parse procedure reference
	procRef := &ast.ExecutableProcedureReference{}

	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		// Procedure variable
		procRef.ProcedureReference = &ast.ProcedureReferenceName{
			ProcedureVariable: &ast.VariableReference{Name: p.curTok.Literal},
		}
		p.nextToken()
	} else {
		// Procedure name
		son, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		procRef.ProcedureReference = &ast.ProcedureReferenceName{
			ProcedureReference: &ast.ProcedureReference{Name: son},
		}
	}

	// Parse parameters
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon &&
		p.curTok.Type != TokenOption && !p.isStatementTerminator() {
		param, err := p.parseExecuteParameter()
		if err != nil {
			break
		}
		procRef.Parameters = append(procRef.Parameters, param)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	spec.ExecutableEntity = procRef
	return spec, nil
}

func (p *Parser) parseExecuteParameter() (*ast.ExecuteParameter, error) {
	param := &ast.ExecuteParameter{IsOutput: false}

	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	param.ParameterValue = expr

	return param, nil
}

func (p *Parser) isStatementTerminator() bool {
	switch p.curTok.Type {
	case TokenSelect, TokenInsert, TokenUpdate, TokenDelete, TokenDeclare,
		TokenIf, TokenWhile, TokenBegin, TokenEnd, TokenCreate, TokenAlter,
		TokenDrop, TokenExec, TokenExecute, TokenPrint, TokenThrow:
		return true
	}
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
		return true
	}
	return false
}

func (p *Parser) parseUpdateStatement() (*ast.UpdateStatement, error) {
	// Consume UPDATE
	p.nextToken()

	stmt := &ast.UpdateStatement{
		UpdateSpecification: &ast.UpdateSpecification{},
	}

	// Parse target
	target, err := p.parseDMLTarget()
	if err != nil {
		return nil, err
	}
	stmt.UpdateSpecification.Target = target

	// Expect SET
	if p.curTok.Type != TokenSet {
		return nil, fmt.Errorf("expected SET, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse SET clauses
	setClauses, err := p.parseSetClauses()
	if err != nil {
		return nil, err
	}
	stmt.UpdateSpecification.SetClauses = setClauses

	// Parse optional FROM clause
	if p.curTok.Type == TokenFrom {
		fromClause, err := p.parseFromClause()
		if err != nil {
			return nil, err
		}
		stmt.UpdateSpecification.FromClause = fromClause
	}

	// Parse optional WHERE clause
	if p.curTok.Type == TokenWhere {
		whereClause, err := p.parseWhereClause()
		if err != nil {
			return nil, err
		}
		stmt.UpdateSpecification.WhereClause = whereClause
	}

	// Parse optional OPTION clause
	if p.curTok.Type == TokenOption {
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

	return stmt, nil
}

func (p *Parser) parseSetClauses() ([]ast.SetClause, error) {
	var clauses []ast.SetClause

	for {
		clause, err := p.parseAssignmentSetClause()
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	return clauses, nil
}

func (p *Parser) parseAssignmentSetClause() (*ast.AssignmentSetClause, error) {
	clause := &ast.AssignmentSetClause{AssignmentKind: "Equals"}

	// Could be @var = col = value, @var = value, or col = value
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		varName := p.curTok.Literal
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			clause.Variable = &ast.VariableReference{Name: varName}
			p.nextToken()

			// Check if next is column = value (SET @a = col = value)
			if p.curTok.Type == TokenIdent && !strings.HasPrefix(p.curTok.Literal, "@") {
				// Could be @a = col = value or @a = expr
				savedTok := p.curTok
				col, err := p.parseMultiPartIdentifierAsColumn()
				if err != nil {
					return nil, err
				}
				if p.curTok.Type == TokenEquals {
					clause.Column = col
					p.nextToken()
					val, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					clause.NewValue = val
					return clause, nil
				}
				// Restore and parse as expression - need different approach
				// The column was actually the value expression
				_ = savedTok
				clause.NewValue = &ast.ColumnReferenceExpression{
					ColumnType:          col.ColumnType,
					MultiPartIdentifier: col.MultiPartIdentifier,
				}
				return clause, nil
			}

			// Just @var = value
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			clause.NewValue = val
			return clause, nil
		}
	}

	// col = value
	col, err := p.parseMultiPartIdentifierAsColumn()
	if err != nil {
		return nil, err
	}
	clause.Column = col

	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected =, got %s", p.curTok.Literal)
	}
	p.nextToken()

	val, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	clause.NewValue = val

	return clause, nil
}

func (p *Parser) parseDeleteStatement() (*ast.DeleteStatement, error) {
	// Consume DELETE
	p.nextToken()

	stmt := &ast.DeleteStatement{
		DeleteSpecification: &ast.DeleteSpecification{},
	}

	// Skip optional FROM
	if p.curTok.Type == TokenFrom {
		p.nextToken()
	}

	// Parse target
	target, err := p.parseDMLTarget()
	if err != nil {
		return nil, err
	}
	stmt.DeleteSpecification.Target = target

	// Parse optional FROM clause
	if p.curTok.Type == TokenFrom {
		fromClause, err := p.parseFromClause()
		if err != nil {
			return nil, err
		}
		stmt.DeleteSpecification.FromClause = fromClause
	}

	// Parse optional WHERE clause
	if p.curTok.Type == TokenWhere {
		whereClause, err := p.parseDeleteWhereClause()
		if err != nil {
			return nil, err
		}
		stmt.DeleteSpecification.WhereClause = whereClause
	}

	// Parse optional OPTION clause
	if p.curTok.Type == TokenOption {
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

	return stmt, nil
}

func (p *Parser) parseDeleteWhereClause() (*ast.WhereClause, error) {
	// Consume WHERE
	p.nextToken()

	// Check for CURRENT OF cursor_name
	if p.curTok.Type == TokenCurrent {
		p.nextToken()
		if p.curTok.Type != TokenOf {
			return nil, fmt.Errorf("expected OF after CURRENT, got %s", p.curTok.Literal)
		}
		p.nextToken()

		// Parse cursor name
		cursorName := p.curTok.Literal
		p.nextToken()

		return &ast.WhereClause{
			Cursor: &ast.CursorId{
				IsGlobal: false,
				Name: &ast.IdentifierOrValueExpression{
					Value: cursorName,
					Identifier: &ast.Identifier{
						Value:     cursorName,
						QuoteType: "NotQuoted",
					},
				},
			},
		}, nil
	}

	condition, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}

	return &ast.WhereClause{SearchCondition: condition}, nil
}

func (p *Parser) parseDeclareVariableStatement() (*ast.DeclareVariableStatement, error) {
	// Consume DECLARE
	p.nextToken()

	stmt := &ast.DeclareVariableStatement{}

	for {
		decl, err := p.parseDeclareVariableElement()
		if err != nil {
			return nil, err
		}
		stmt.Declarations = append(stmt.Declarations, decl)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDeclareVariableElement() (*ast.DeclareVariableElement, error) {
	elem := &ast.DeclareVariableElement{}

	// Parse variable name
	if p.curTok.Type != TokenIdent || !strings.HasPrefix(p.curTok.Literal, "@") {
		return nil, fmt.Errorf("expected variable name, got %s", p.curTok.Literal)
	}
	elem.VariableName = &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
	p.nextToken()

	// Skip optional AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse data type
	dataType, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	elem.DataType = dataType

	// Check for = initial value
	if p.curTok.Type == TokenEquals {
		p.nextToken()
		val, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		elem.Value = val
	}

	return elem, nil
}

func (p *Parser) parseDataType() (*ast.SqlDataTypeReference, error) {
	dt := &ast.SqlDataTypeReference{}

	if p.curTok.Type == TokenCursor {
		dt.SqlDataTypeOption = "Cursor"
		p.nextToken()
		return dt, nil
	}

	if p.curTok.Type != TokenIdent {
		return nil, fmt.Errorf("expected data type, got %s", p.curTok.Literal)
	}

	typeName := p.curTok.Literal
	dt.SqlDataTypeOption = convertDataTypeOption(typeName)
	baseId := &ast.Identifier{Value: typeName, QuoteType: "NotQuoted"}
	dt.Name = &ast.SchemaObjectName{
		BaseIdentifier: baseId,
		Count:          1,
		Identifiers:    []*ast.Identifier{baseId},
	}
	p.nextToken()

	// Check for parameters like VARCHAR(100)
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			dt.Parameters = append(dt.Parameters, expr)
			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken()
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	return dt, nil
}

func convertDataTypeOption(typeName string) string {
	typeMap := map[string]string{
		"INT":       "Int",
		"INTEGER":   "Int",
		"BIGINT":    "BigInt",
		"SMALLINT":  "SmallInt",
		"TINYINT":   "TinyInt",
		"BIT":       "Bit",
		"DECIMAL":   "Decimal",
		"NUMERIC":   "Numeric",
		"MONEY":     "Money",
		"SMALLMONEY": "SmallMoney",
		"FLOAT":     "Float",
		"REAL":      "Real",
		"DATETIME":  "DateTime",
		"DATETIME2": "DateTime2",
		"DATE":      "Date",
		"TIME":      "Time",
		"CHAR":      "Char",
		"VARCHAR":   "VarChar",
		"TEXT":      "Text",
		"NCHAR":     "NChar",
		"NVARCHAR":  "NVarChar",
		"NTEXT":     "NText",
		"BINARY":    "Binary",
		"VARBINARY": "VarBinary",
		"IMAGE":     "Image",
		"CURSOR":    "Cursor",
		"SQL_VARIANT": "Sql_Variant",
		"TABLE":     "Table",
		"UNIQUEIDENTIFIER": "UniqueIdentifier",
		"XML":       "Xml",
	}
	if mapped, ok := typeMap[strings.ToUpper(typeName)]; ok {
		return mapped
	}
	// Return with first letter capitalized
	if len(typeName) > 0 {
		return strings.ToUpper(typeName[:1]) + strings.ToLower(typeName[1:])
	}
	return typeName
}

func (p *Parser) parseSetVariableStatement() (ast.Statement, error) {
	// Consume SET
	p.nextToken()

	// Check for predicate SET options like SET ANSI_NULLS ON/OFF
	if p.curTok.Type == TokenIdent {
		optionName := strings.ToUpper(p.curTok.Literal)
		var setOpt ast.SetOptions
		switch optionName {
		case "ANSI_NULLS":
			setOpt = ast.SetOptionsAnsiNulls
		case "ANSI_PADDING":
			setOpt = ast.SetOptionsAnsiPadding
		case "ANSI_WARNINGS":
			setOpt = ast.SetOptionsAnsiWarnings
		case "ARITHABORT":
			setOpt = ast.SetOptionsArithAbort
		case "ARITHIGNORE":
			setOpt = ast.SetOptionsArithIgnore
		case "CONCAT_NULL_YIELDS_NULL":
			setOpt = ast.SetOptionsConcatNullYieldsNull
		case "CURSOR_CLOSE_ON_COMMIT":
			setOpt = ast.SetOptionsCursorCloseOnCommit
		case "FMTONLY":
			setOpt = ast.SetOptionsFmtOnly
		case "FORCEPLAN":
			setOpt = ast.SetOptionsForceplan
		case "IMPLICIT_TRANSACTIONS":
			setOpt = ast.SetOptionsImplicitTransactions
		case "NOCOUNT":
			setOpt = ast.SetOptionsNoCount
		case "NOEXEC":
			setOpt = ast.SetOptionsNoExec
		case "NUMERIC_ROUNDABORT":
			setOpt = ast.SetOptionsNumericRoundAbort
		case "PARSEONLY":
			setOpt = ast.SetOptionsParseOnly
		case "QUOTED_IDENTIFIER":
			setOpt = ast.SetOptionsQuotedIdentifier
		case "REMOTE_PROC_TRANSACTIONS":
			setOpt = ast.SetOptionsRemoteProcTransactions
		case "SHOWPLAN_ALL":
			setOpt = ast.SetOptionsShowplanAll
		case "SHOWPLAN_TEXT":
			setOpt = ast.SetOptionsShowplanText
		case "SHOWPLAN_XML":
			setOpt = ast.SetOptionsShowplanXml
		case "STATISTICS_IO":
			setOpt = ast.SetOptionsStatisticsIo
		case "STATISTICS_PROFILE":
			setOpt = ast.SetOptionsStatisticsProfile
		case "STATISTICS_TIME":
			setOpt = ast.SetOptionsStatisticsTime
		case "STATISTICS_XML":
			setOpt = ast.SetOptionsStatisticsXml
		case "XACT_ABORT":
			setOpt = ast.SetOptionsXactAbort
		}
		if setOpt != "" {
			p.nextToken() // consume option name
			isOn := false
			// ON is tokenized as TokenOn, not TokenIdent
			if p.curTok.Type == TokenOn || (p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "ON") {
				isOn = true
				p.nextToken()
			} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OFF" {
				isOn = false
				p.nextToken()
			}
			// Skip optional semicolon
			if p.curTok.Type == TokenSemicolon {
				p.nextToken()
			}
			return &ast.PredicateSetStatement{
				Options: setOpt,
				IsOn:    isOn,
			}, nil
		}
	}

	stmt := &ast.SetVariableStatement{
		AssignmentKind: "Equals",
		SeparatorType:  "Equals",
	}

	// Parse variable name
	if p.curTok.Type != TokenIdent || !strings.HasPrefix(p.curTok.Literal, "@") {
		return nil, fmt.Errorf("expected variable name, got %s", p.curTok.Literal)
	}
	stmt.Variable = &ast.VariableReference{Name: p.curTok.Literal}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected =, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for CURSOR definition
	if p.curTok.Type == TokenCursor {
		p.nextToken()
		// Parse cursor options and FOR SELECT
		// For now, simplified - skip to FOR
		for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
			if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken()
				break
			}
			p.nextToken()
		}
		if p.curTok.Type == TokenSelect {
			qe, err := p.parseQueryExpression()
			if err != nil {
				return nil, err
			}
			stmt.CursorDefinition = &ast.CursorDefinition{Select: qe}
		}
	} else {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.Expression = expr
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseIfStatement() (*ast.IfStatement, error) {
	// Consume IF
	p.nextToken()

	stmt := &ast.IfStatement{}

	// Parse predicate
	pred, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}
	stmt.Predicate = pred

	// Parse THEN statement
	thenStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	stmt.ThenStatement = thenStmt

	// Check for ELSE
	if p.curTok.Type == TokenElse {
		p.nextToken()
		elseStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmt.ElseStatement = elseStmt
	}

	return stmt, nil
}

func (p *Parser) parseWhileStatement() (*ast.WhileStatement, error) {
	// Consume WHILE
	p.nextToken()

	stmt := &ast.WhileStatement{}

	// Parse predicate
	pred, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}
	stmt.Predicate = pred

	// Parse body statement
	bodyStmt, err := p.parseStatement()
	if err != nil {
		return nil, err
	}
	stmt.Statement = bodyStmt

	return stmt, nil
}

func (p *Parser) parseBeginStatement() (ast.Statement, error) {
	// Peek at what follows BEGIN
	p.nextToken() // consume BEGIN

	switch p.curTok.Type {
	case TokenTransaction, TokenTran:
		return p.parseBeginTransactionStatementContinued(false)
	case TokenTry:
		return p.parseTryCatchStatement()
	case TokenIdent:
		// Check for DISTRIBUTED
		if strings.ToUpper(p.curTok.Literal) == "DISTRIBUTED" {
			p.nextToken() // consume DISTRIBUTED
			if p.curTok.Type == TokenTransaction || p.curTok.Type == TokenTran {
				return p.parseBeginTransactionStatementContinued(true)
			}
			return nil, fmt.Errorf("expected TRANSACTION after DISTRIBUTED, got %s", p.curTok.Literal)
		}
		// Fall through to BEGIN...END block
		fallthrough
	default:
		return p.parseBeginEndBlockStatementContinued()
	}
}

func (p *Parser) parseBeginTransactionStatementContinued(distributed bool) (*ast.BeginTransactionStatement, error) {
	// TRANSACTION or TRAN already consumed by caller
	p.nextToken()

	stmt := &ast.BeginTransactionStatement{
		Distributed: distributed,
	}

	// Optional transaction name or variable
	if p.curTok.Type == TokenIdent && !isKeyword(p.curTok.Literal) {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			Identifier: &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
			},
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			ValueExpression: &ast.VariableReference{
				Name: p.curTok.Literal,
			},
		}
		p.nextToken()
	}

	// Check for WITH MARK
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "MARK" {
			stmt.MarkDefined = true
			p.nextToken() // consume MARK
			// Optional mark description
			if p.curTok.Type == TokenString || (p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '@') {
				desc, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.MarkDescription = desc
			}
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseTryCatchStatement() (*ast.TryCatchStatement, error) {
	// TRY already seen, consume it
	p.nextToken()

	stmt := &ast.TryCatchStatement{
		TryStatements: &ast.StatementList{},
	}

	// Parse statements until END TRY
	for p.curTok.Type != TokenEnd && p.curTok.Type != TokenEOF {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmt.TryStatements.Statements = append(stmt.TryStatements.Statements, s)
		}
	}

	// Consume END TRY
	if p.curTok.Type == TokenEnd {
		p.nextToken() // consume END
		if p.curTok.Type == TokenTry {
			p.nextToken() // consume TRY
		}
	}

	// Expect BEGIN CATCH
	if p.curTok.Type == TokenBegin {
		p.nextToken() // consume BEGIN
		if p.curTok.Type == TokenCatch {
			p.nextToken() // consume CATCH
		}
	}

	stmt.CatchStatements = &ast.StatementList{}

	// Parse catch statements until END CATCH
	for p.curTok.Type != TokenEnd && p.curTok.Type != TokenEOF {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmt.CatchStatements.Statements = append(stmt.CatchStatements.Statements, s)
		}
	}

	// Consume END CATCH
	if p.curTok.Type == TokenEnd {
		p.nextToken() // consume END
		if p.curTok.Type == TokenCatch {
			p.nextToken() // consume CATCH
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseBeginEndBlockStatementContinued() (*ast.BeginEndBlockStatement, error) {
	stmt := &ast.BeginEndBlockStatement{
		StatementList: &ast.StatementList{},
	}

	// Parse statements until END
	for p.curTok.Type != TokenEnd && p.curTok.Type != TokenEOF {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmt.StatementList.Statements = append(stmt.StatementList.Statements, s)
		}
	}

	// Consume END
	if p.curTok.Type == TokenEnd {
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseBeginEndBlockStatement() (*ast.BeginEndBlockStatement, error) {
	// Consume BEGIN
	p.nextToken()

	stmt := &ast.BeginEndBlockStatement{
		StatementList: &ast.StatementList{},
	}

	// Parse statements until END
	for p.curTok.Type != TokenEnd && p.curTok.Type != TokenEOF {
		s, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s != nil {
			stmt.StatementList.Statements = append(stmt.StatementList.Statements, s)
		}
	}

	// Consume END
	if p.curTok.Type == TokenEnd {
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateStatement() (ast.Statement, error) {
	// Consume CREATE
	p.nextToken()

	switch p.curTok.Type {
	case TokenTable:
		return p.parseCreateTableStatement()
	case TokenView:
		return p.parseCreateViewStatement()
	case TokenSchema:
		return p.parseCreateSchemaStatement()
	case TokenDefault:
		return p.parseCreateDefaultStatement()
	case TokenMaster:
		return p.parseCreateMasterKeyStatement()
	default:
		return nil, fmt.Errorf("unexpected token after CREATE: %s", p.curTok.Literal)
	}
}

func (p *Parser) parseCreateViewStatement() (*ast.CreateViewStatement, error) {
	// Consume VIEW
	p.nextToken()

	stmt := &ast.CreateViewStatement{}

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
			if p.curTok.Type == TokenIdent {
				stmt.Columns = append(stmt.Columns, &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"})
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

	// Check for WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken()
		// Parse view options
		for p.curTok.Type == TokenIdent {
			opt := ast.ViewOption{OptionKind: p.curTok.Literal}
			stmt.ViewOptions = append(stmt.ViewOptions, opt)
			p.nextToken()
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
	}

	// Expect AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse SELECT statement
	selStmt, err := p.parseSelectStatement()
	if err != nil {
		return nil, err
	}
	stmt.SelectStatement = selStmt

	return stmt, nil
}

func (p *Parser) parseCreateSchemaStatement() (*ast.CreateSchemaStatement, error) {
	// Consume SCHEMA
	p.nextToken()

	stmt := &ast.CreateSchemaStatement{}

	// Parse schema name (can be bracketed) or AUTHORIZATION
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
		stmt.Name = p.parseIdentifier()
	}

	// Check for AUTHORIZATION
	if p.curTok.Type == TokenAuthorization {
		p.nextToken()
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			stmt.Owner = p.parseIdentifier()
		}
	}

	// Parse schema elements (CREATE TABLE, CREATE VIEW, GRANT)
	stmt.StatementList = &ast.StatementList{}
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		// Check for GO (batch separator)
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			break
		}
		// Parse schema element statements
		if p.curTok.Type == TokenCreate || p.curTok.Type == TokenGrant {
			elemStmt, err := p.parseStatement()
			if err != nil {
				break
			}
			if elemStmt != nil {
				stmt.StatementList.Statements = append(stmt.StatementList.Statements, elemStmt)
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

func (p *Parser) parseCreateDefaultStatement() (*ast.CreateDefaultStatement, error) {
	// Consume DEFAULT
	p.nextToken()

	stmt := &ast.CreateDefaultStatement{}

	// Parse default name
	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Name = name

	// Expect AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Parse expression
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

func (p *Parser) parseCreateMasterKeyStatement() (*ast.CreateMasterKeyStatement, error) {
	// Consume MASTER
	p.nextToken()

	stmt := &ast.CreateMasterKeyStatement{}

	// Expect KEY
	if p.curTok.Type != TokenKey {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect ENCRYPTION
	if p.curTok.Type != TokenEncryption {
		return nil, fmt.Errorf("expected ENCRYPTION after KEY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect BY
	if p.curTok.Type != TokenBy {
		return nil, fmt.Errorf("expected BY after ENCRYPTION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect PASSWORD
	if p.curTok.Type != TokenPassword {
		return nil, fmt.Errorf("expected PASSWORD after BY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after PASSWORD, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse password expression
	password, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Password = password

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseExecuteStatement() (*ast.ExecuteStatement, error) {
	execSpec, err := p.parseExecuteSpecification()
	if err != nil {
		return nil, err
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.ExecuteStatement{ExecuteSpecification: execSpec}, nil
}

func (p *Parser) parseReturnStatement() (*ast.ReturnStatement, error) {
	// Consume RETURN
	p.nextToken()

	stmt := &ast.ReturnStatement{}

	// Check for expression
	if p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && !p.isStatementTerminator() {
		expr, err := p.parseScalarExpression()
		if err == nil {
			stmt.Expression = expr
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseBreakStatement() (*ast.BreakStatement, error) {
	// Consume BREAK
	p.nextToken()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.BreakStatement{}, nil
}

func (p *Parser) parseContinueStatement() (*ast.ContinueStatement, error) {
	// Consume CONTINUE
	p.nextToken()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.ContinueStatement{}, nil
}

func (p *Parser) parseCommitTransactionStatement() (*ast.CommitTransactionStatement, error) {
	// Consume COMMIT
	p.nextToken()

	stmt := &ast.CommitTransactionStatement{
		DelayedDurabilityOption: "NotSet",
	}

	// Skip optional WORK, TRAN, or TRANSACTION
	if p.curTok.Type == TokenWork || p.curTok.Type == TokenTran || p.curTok.Type == TokenTransaction {
		p.nextToken()
	}

	// Optional transaction name or variable
	if p.curTok.Type == TokenIdent && !isKeyword(p.curTok.Literal) && p.curTok.Literal[0] != '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			Identifier: &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
			},
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			ValueExpression: &ast.VariableReference{
				Name: p.curTok.Literal,
			},
		}
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseRollbackTransactionStatement() (*ast.RollbackTransactionStatement, error) {
	// Consume ROLLBACK
	p.nextToken()

	stmt := &ast.RollbackTransactionStatement{}

	// Skip optional WORK, TRAN, or TRANSACTION
	if p.curTok.Type == TokenWork || p.curTok.Type == TokenTran || p.curTok.Type == TokenTransaction {
		p.nextToken()
	}

	// Optional transaction name or variable
	if p.curTok.Type == TokenIdent && !isKeyword(p.curTok.Literal) && p.curTok.Literal[0] != '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			Identifier: &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
			},
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			ValueExpression: &ast.VariableReference{
				Name: p.curTok.Literal,
			},
		}
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseSaveTransactionStatement() (*ast.SaveTransactionStatement, error) {
	// Consume SAVE
	p.nextToken()

	stmt := &ast.SaveTransactionStatement{}

	// Skip optional TRAN or TRANSACTION
	if p.curTok.Type == TokenTran || p.curTok.Type == TokenTransaction {
		p.nextToken()
	}

	// Optional transaction name or variable
	if p.curTok.Type == TokenIdent && !isKeyword(p.curTok.Literal) && p.curTok.Literal[0] != '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			Identifier: &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
			},
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			ValueExpression: &ast.VariableReference{
				Name: p.curTok.Literal,
			},
		}
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseWaitForStatement() (*ast.WaitForStatement, error) {
	// Consume WAITFOR
	p.nextToken()

	stmt := &ast.WaitForStatement{}

	// Expect DELAY or TIME
	if p.curTok.Type == TokenDelay {
		stmt.WaitForOption = "Delay"
		p.nextToken()
	} else if p.curTok.Type == TokenTime {
		stmt.WaitForOption = "Time"
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected DELAY or TIME after WAITFOR, got %s", p.curTok.Literal)
	}

	// Parse the parameter expression
	param, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Parameter = param

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseGotoStatement() (*ast.GoToStatement, error) {
	// Consume GOTO
	p.nextToken()

	stmt := &ast.GoToStatement{}

	// Expect label name
	if p.curTok.Type == TokenIdent {
		stmt.LabelName = &ast.Identifier{
			Value:     p.curTok.Literal,
			QuoteType: "NotQuoted",
		}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected label name after GOTO, got %s", p.curTok.Literal)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseLabelOrError() (ast.Statement, error) {
	// Check if this is a label (identifier followed by colon)
	label := p.curTok.Literal
	p.nextToken()

	// Check if followed by colon - then it's a label
	if p.curTok.Type == TokenColon {
		p.nextToken() // consume the colon
		return &ast.LabelStatement{Value: label + ":"}, nil
	}

	// Not a label - return error
	return nil, fmt.Errorf("unexpected identifier: %s", label)
}

func isKeyword(s string) bool {
	_, ok := keywords[strings.ToUpper(s)]
	return ok
}

// ======================= End New Statement Parsing Functions =======================

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
	case *ast.ExecuteStatement:
		return executeStatementToJSON(s)
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
	case *ast.RevertStatement:
		return revertStatementToJSON(s)
	case *ast.DropCredentialStatement:
		return dropCredentialStatementToJSON(s)
	case *ast.CreateTableStatement:
		return createTableStatementToJSON(s)
	case *ast.GrantStatement:
		return grantStatementToJSON(s)
	case *ast.PredicateSetStatement:
		return predicateSetStatementToJSON(s)
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
	case *ast.GoToStatement:
		return goToStatementToJSON(s)
	case *ast.LabelStatement:
		return labelStatementToJSON(s)
	case *ast.CreateDefaultStatement:
		return createDefaultStatementToJSON(s)
	case *ast.CreateMasterKeyStatement:
		return createMasterKeyStatementToJSON(s)
	case *ast.TryCatchStatement:
		return tryCatchStatementToJSON(s)
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

func optimizerHintToJSON(h *ast.OptimizerHint) jsonNode {
	node := jsonNode{
		"$type": "OptimizerHint",
	}
	if h.HintKind != "" {
		node["HintKind"] = h.HintKind
	}
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

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after table name, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt.Definition = &ast.TableDefinition{}

	// Parse column definitions
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		colDef, err := p.parseColumnDefinition()
		if err != nil {
			return nil, err
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

	// Parse data type
	dataType, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	col.DataType = dataType

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
		"$type":            "CreateTableStatement",
		"SchemaObjectName": schemaObjectNameToJSON(s.SchemaObjectName),
		"AsEdge":           s.AsEdge,
		"AsFileTable":      s.AsFileTable,
		"AsNode":           s.AsNode,
		"Definition":       tableDefinitionToJSON(s.Definition),
	}
	return node
}

func tableDefinitionToJSON(t *ast.TableDefinition) jsonNode {
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
	if c.DataType != nil {
		node["DataType"] = dataTypeReferenceToJSON(c.DataType)
	}
	return node
}

func dataTypeReferenceToJSON(d ast.DataTypeReference) jsonNode {
	switch dt := d.(type) {
	case *ast.SqlDataTypeReference:
		return sqlDataTypeReferenceToJSON(dt)
	default:
		return jsonNode{"$type": "UnknownDataType"}
	}
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
