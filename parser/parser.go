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
	case TokenSemicolon:
		p.nextToken()
		return nil, nil
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
	case TokenIdent:
		// Check if it's a variable reference (starts with @)
		if strings.HasPrefix(p.curTok.Literal, "@") {
			name := p.curTok.Literal
			p.nextToken()
			return &ast.VariableReference{Name: name}, nil
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

func whereClauseToJSON(wc *ast.WhereClause) jsonNode {
	node := jsonNode{
		"$type": "WhereClause",
	}
	if wc.SearchCondition != nil {
		node["SearchCondition"] = booleanExpressionToJSON(wc.SearchCondition)
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
