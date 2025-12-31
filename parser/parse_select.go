// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/teesql/ast"
)

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
	qe, into, on, err := p.parseQueryExpressionWithInto()
	if err != nil {
		return nil, err
	}
	stmt.QueryExpression = qe
	stmt.Into = into
	stmt.On = on

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
	qe, _, _, err := p.parseQueryExpressionWithInto()
	return qe, err
}

func (p *Parser) parseQueryExpressionWithInto() (ast.QueryExpression, *ast.SchemaObjectName, *ast.Identifier, error) {
	// Parse primary query expression (could be SELECT or parenthesized)
	left, into, on, err := p.parsePrimaryQueryExpression()
	if err != nil {
		return nil, nil, nil, err
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
		right, rightInto, rightOn, err := p.parsePrimaryQueryExpression()
		if err != nil {
			return nil, nil, nil, err
		}

		// INTO can only appear in the first query of a UNION
		if rightInto != nil && into == nil {
			into = rightInto
			on = rightOn
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
			return nil, nil, nil, err
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

	return left, into, on, nil
}

func (p *Parser) parsePrimaryQueryExpression() (ast.QueryExpression, *ast.SchemaObjectName, *ast.Identifier, error) {
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		qe, into, on, err := p.parseQueryExpressionWithInto()
		if err != nil {
			return nil, nil, nil, err
		}
		if p.curTok.Type != TokenRParen {
			return nil, nil, nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
		return &ast.QueryParenthesisExpression{QueryExpression: qe}, into, on, nil
	}

	return p.parseQuerySpecificationWithInto()
}

func (p *Parser) parseQuerySpecificationWithInto() (*ast.QuerySpecification, *ast.SchemaObjectName, *ast.Identifier, error) {
	qs, err := p.parseQuerySpecificationCore()
	if err != nil {
		return nil, nil, nil, err
	}

	// Check for INTO clause after SELECT elements, before FROM
	var into *ast.SchemaObjectName
	var on *ast.Identifier
	if p.curTok.Type == TokenInto {
		p.nextToken() // consume INTO
		into, err = p.parseSchemaObjectName()
		if err != nil {
			return nil, nil, nil, err
		}
		// Check for ON filegroup clause
		if strings.ToUpper(p.curTok.Literal) == "ON" {
			p.nextToken() // consume ON
			on = p.parseIdentifier()
		}
	}

	// Parse optional FROM clause
	if p.curTok.Type == TokenFrom {
		fromClause, err := p.parseFromClause()
		if err != nil {
			return nil, nil, nil, err
		}
		qs.FromClause = fromClause
	}

	// Parse optional WHERE clause
	if p.curTok.Type == TokenWhere {
		whereClause, err := p.parseWhereClause()
		if err != nil {
			return nil, nil, nil, err
		}
		qs.WhereClause = whereClause
	}

	// Parse optional GROUP BY clause
	if p.curTok.Type == TokenGroup {
		groupByClause, err := p.parseGroupByClause()
		if err != nil {
			return nil, nil, nil, err
		}
		qs.GroupByClause = groupByClause
	}

	// Parse optional HAVING clause
	if p.curTok.Type == TokenHaving {
		havingClause, err := p.parseHavingClause()
		if err != nil {
			return nil, nil, nil, err
		}
		qs.HavingClause = havingClause
	}

	// Note: ORDER BY is parsed at the top level in parseQueryExpressionWithInto
	// to correctly handle UNION/EXCEPT/INTERSECT cases

	return qs, into, on, nil
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

		// Check for subquery (SELECT ...)
		if p.curTok.Type == TokenSelect {
			qe, err := p.parseQueryExpression()
			if err != nil {
				return nil, err
			}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken()
			top.Expression = &ast.ScalarSubquery{QueryExpression: qe}
		} else {
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			// Wrap in ParenthesisExpression
			top.Expression = &ast.ParenthesisExpression{Expression: expr}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )
		}
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

	// Check for variable assignment: @var = expr or @var ||= expr
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		varName := p.curTok.Literal
		p.nextToken() // consume variable

		// Check if this is an assignment
		if p.isCompoundAssignment() {
			ssv := &ast.SelectSetVariable{
				Variable:       &ast.VariableReference{Name: varName},
				AssignmentKind: p.getAssignmentKind(),
			}
			p.nextToken() // consume assignment operator

			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			ssv.Expression = expr
			return ssv, nil
		}

		// Not an assignment, treat as regular scalar expression starting with variable
		// We need to "un-consume" the variable and let parseScalarExpression handle it
		// Create the variable reference and use it as the expression
		varRef := &ast.VariableReference{Name: varName}
		sse := &ast.SelectScalarExpression{Expression: varRef}

		// Check for column alias
		if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '[' {
			alias := p.parseIdentifier()
			sse.ColumnName = &ast.IdentifierOrValueExpression{
				Value:      alias.Value,
				Identifier: alias,
			}
		} else if p.curTok.Type == TokenAs {
			p.nextToken()
			alias := p.parseIdentifier()
			sse.ColumnName = &ast.IdentifierOrValueExpression{
				Value:      alias.Value,
				Identifier: alias,
			}
		} else if p.curTok.Type == TokenIdent {
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
		// Unescape ]] to ]
		literal = strings.ReplaceAll(literal, "]]", "]")
	} else if len(literal) >= 2 && literal[0] == '"' && literal[len(literal)-1] == '"' {
		// Handle double-quoted identifiers
		quoteType = "DoubleQuote"
		literal = literal[1 : len(literal)-1]
		// Unescape "" to "
		literal = strings.ReplaceAll(literal, "\"\"", "\"")
	}

	id := &ast.Identifier{
		Value:     literal,
		QuoteType: quoteType,
	}
	p.nextToken()
	return id
}

// isKeywordAsIdentifier returns true if the current token is a keyword that can be used as an identifier
func (p *Parser) isKeywordAsIdentifier() bool {
	// In T-SQL, many keywords can be used as identifiers in the right context
	// This includes database objects, table names, column names, etc.
	switch p.curTok.Type {
	case TokenMaster, TokenKey, TokenIndex, TokenLanguage,
		TokenUser, TokenSchema, TokenDatabase, TokenTable,
		TokenView, TokenProcedure, TokenFunction, TokenTrigger,
		TokenDefault, TokenMessage, TokenCredential, TokenCertificate, TokenLogin,
		TokenExternal, TokenSymmetric, TokenAsymmetric, TokenGroup,
		TokenAdd, TokenGrant, TokenRevoke, TokenBackup, TokenRestore,
		TokenQuery, TokenJob, TokenStats, TokenPassword, TokenTime, TokenDelay,
		TokenTyp:
		return true
	default:
		return false
	}
}

func (p *Parser) parseScalarExpression() (ast.ScalarExpression, error) {
	return p.parseShiftExpression()
}

func (p *Parser) parseShiftExpression() (ast.ScalarExpression, error) {
	left, err := p.parseAdditiveExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenLeftShift || p.curTok.Type == TokenRightShift {
		var opType string
		if p.curTok.Type == TokenLeftShift {
			opType = "LeftShift"
		} else {
			opType = "RightShift"
		}
		p.nextToken()

		right, err := p.parseAdditiveExpression()
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

func (p *Parser) parseAdditiveExpression() (ast.ScalarExpression, error) {
	left, err := p.parseMultiplicativeExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenPlus || p.curTok.Type == TokenMinus || p.curTok.Type == TokenDoublePipe {
		var opType string
		switch p.curTok.Type {
		case TokenPlus:
			opType = "Add"
		case TokenMinus:
			opType = "Subtract"
		case TokenDoublePipe:
			opType = "Concat"
		}
		p.nextToken()

		right, err := p.parseMultiplicativeExpression()
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

func (p *Parser) parseMultiplicativeExpression() (ast.ScalarExpression, error) {
	left, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenStar || p.curTok.Type == TokenSlash || p.curTok.Type == TokenModulo {
		var opType string
		switch p.curTok.Type {
		case TokenStar:
			opType = "Multiply"
		case TokenSlash:
			opType = "Divide"
		case TokenModulo:
			opType = "Modulo"
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
		val := p.curTok.Literal
		p.nextToken()
		return &ast.NullLiteral{LiteralType: "Null", Value: val}, nil
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
		// Check if it's a global variable reference (starts with @@)
		if strings.HasPrefix(p.curTok.Literal, "@@") {
			name := p.curTok.Literal
			p.nextToken()
			return &ast.GlobalVariableExpression{Name: name}, nil
		}
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
		// Check for CAST/CONVERT special functions
		upper := strings.ToUpper(p.curTok.Literal)
		if upper == "CAST" && p.peekTok.Type == TokenLParen {
			return p.parseCastCall()
		}
		if upper == "CONVERT" && p.peekTok.Type == TokenLParen {
			return p.parseConvertCall()
		}
		if upper == "TRY_CAST" && p.peekTok.Type == TokenLParen {
			return p.parseTryCastCall()
		}
		if upper == "TRY_CONVERT" && p.peekTok.Type == TokenLParen {
			return p.parseTryConvertCall()
		}
		return p.parseColumnReferenceOrFunctionCall()
	case TokenNumber:
		val := p.curTok.Literal
		p.nextToken()
		// Check if it's a decimal number
		if strings.Contains(val, ".") {
			return &ast.NumericLiteral{LiteralType: "Numeric", Value: val}, nil
		}
		return &ast.IntegerLiteral{LiteralType: "Integer", Value: val}, nil
	case TokenBinary:
		val := p.curTok.Literal
		p.nextToken()
		return &ast.BinaryLiteral{LiteralType: "Binary", Value: val, IsLargeObject: false}, nil
	case TokenString:
		return p.parseStringLiteral()
	case TokenNationalString:
		return p.parseNationalStringFromToken()
	case TokenLBrace:
		return p.parseOdbcLiteral()
	case TokenLParen:
		// Parenthesized expression or scalar subquery
		p.nextToken()
		// Check if it's a scalar subquery (starts with SELECT)
		if p.curTok.Type == TokenSelect {
			qe, err := p.parseQueryExpression()
			if err != nil {
				return nil, err
			}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken()
			return &ast.ScalarSubquery{QueryExpression: qe}, nil
		}
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken()
		// Check for property access after parenthesized expression: (c1).SomeProperty
		return p.parsePostExpressionAccess(&ast.ParenthesisExpression{Expression: expr})
	case TokenCase:
		return p.parseCaseExpression()
	case TokenStar:
		// Wildcard column reference (e.g., * in count(*))
		p.nextToken()
		return &ast.ColumnReferenceExpression{ColumnType: "Wildcard"}, nil
	default:
		return nil, fmt.Errorf("unexpected token in expression: %s", p.curTok.Literal)
	}
}

func (p *Parser) parseCaseExpression() (ast.ScalarExpression, error) {
	p.nextToken() // consume CASE

	// Check if it's a searched CASE (CASE WHEN ...) or simple CASE (CASE expr WHEN ...)
	if p.curTok.Type == TokenWhen {
		// Searched CASE expression
		return p.parseSearchedCaseExpression()
	}
	// Simple CASE expression
	return p.parseSimpleCaseExpression()
}

func (p *Parser) parseSearchedCaseExpression() (*ast.SearchedCaseExpression, error) {
	expr := &ast.SearchedCaseExpression{}

	for p.curTok.Type == TokenWhen {
		p.nextToken() // consume WHEN

		when, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}

		if p.curTok.Type != TokenThen {
			return nil, fmt.Errorf("expected THEN in CASE, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume THEN

		then, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		expr.WhenClauses = append(expr.WhenClauses, &ast.SearchedWhenClause{
			WhenExpression: when,
			ThenExpression: then,
		})
	}

	// Optional ELSE
	if p.curTok.Type == TokenElse {
		p.nextToken() // consume ELSE
		elseExpr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		expr.ElseExpression = elseExpr
	}

	if p.curTok.Type != TokenEnd {
		return nil, fmt.Errorf("expected END in CASE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume END

	return expr, nil
}

func (p *Parser) parseSimpleCaseExpression() (*ast.SimpleCaseExpression, error) {
	expr := &ast.SimpleCaseExpression{}

	// Parse input expression
	input, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	expr.InputExpression = input

	for p.curTok.Type == TokenWhen {
		p.nextToken() // consume WHEN

		when, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		if p.curTok.Type != TokenThen {
			return nil, fmt.Errorf("expected THEN in CASE, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume THEN

		then, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		expr.WhenClauses = append(expr.WhenClauses, &ast.SimpleWhenClause{
			WhenExpression: when,
			ThenExpression: then,
		})
	}

	// Optional ELSE
	if p.curTok.Type == TokenElse {
		p.nextToken() // consume ELSE
		elseExpr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		expr.ElseExpression = elseExpr
	}

	if p.curTok.Type != TokenEnd {
		return nil, fmt.Errorf("expected END in CASE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume END

	return expr, nil
}

func (p *Parser) parseOdbcLiteral() (*ast.OdbcLiteral, error) {
	// Consume {
	p.nextToken()

	// Expect "guid" identifier
	if p.curTok.Type != TokenIdent || strings.ToLower(p.curTok.Literal) != "guid" {
		return nil, fmt.Errorf("expected guid in ODBC literal, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for national string (either separate N token or combined N'...' token)
	isNational := false
	var raw string

	if p.curTok.Type == TokenNationalString {
		// Combined N'...' token from lexer
		isNational = true
		raw = p.curTok.Literal
		// Strip the N prefix
		if len(raw) >= 3 && (raw[0] == 'N' || raw[0] == 'n') && raw[1] == '\'' {
			raw = raw[1:] // Remove the N, keep the rest including quotes
		}
		p.nextToken()
	} else {
		// Check for separate N token followed by string
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "N" {
			isNational = true
			p.nextToken()
		}

		// Expect string literal
		if p.curTok.Type != TokenString {
			return nil, fmt.Errorf("expected string in ODBC literal, got %s", p.curTok.Literal)
		}

		raw = p.curTok.Literal
		p.nextToken()
	}

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

// parseStringLiteralValue creates a StringLiteral from the current token without consuming it
func (p *Parser) parseStringLiteralValue() *ast.StringLiteral {
	raw := p.curTok.Literal

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
		}
	}

	return &ast.StringLiteral{
		LiteralType:   "String",
		IsNational:    false,
		IsLargeObject: false,
		Value:         raw,
	}
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

func (p *Parser) parseNationalStringFromToken() (*ast.StringLiteral, error) {
	// Token is N'...' combined - strip the N prefix and quotes
	raw := p.curTok.Literal
	p.nextToken()

	// Raw is like N'value' or n'value'
	if len(raw) >= 3 && (raw[0] == 'N' || raw[0] == 'n') && raw[1] == '\'' && raw[len(raw)-1] == '\'' {
		inner := raw[2 : len(raw)-1]
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

func (p *Parser) parseColumnReferenceOrFunctionCall() (ast.ScalarExpression, error) {
	var identifiers []*ast.Identifier

	for {
		if p.curTok.Type != TokenIdent {
			break
		}

		quoteType := "NotQuoted"
		literal := p.curTok.Literal
		// Handle bracketed identifiers
		if len(literal) >= 2 && literal[0] == '[' && literal[len(literal)-1] == ']' {
			quoteType = "SquareBracket"
			literal = literal[1 : len(literal)-1]
		}

		id := &ast.Identifier{
			Value:     literal,
			QuoteType: quoteType,
		}
		identifiers = append(identifiers, id)
		p.nextToken()

		if p.curTok.Type != TokenDot {
			break
		}
		p.nextToken() // consume dot
	}

	// Check for :: (user-defined type method call or property access): a.b::func() or a::prop
	if p.curTok.Type == TokenColonColon && len(identifiers) > 0 {
		p.nextToken() // consume ::

		// Parse function/property name
		if p.curTok.Type != TokenIdent {
			return nil, fmt.Errorf("expected identifier after ::, got %s", p.curTok.Literal)
		}
		name := &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
		p.nextToken()

		// Build SchemaObjectName from identifiers
		schemaObjName := identifiersToSchemaObjectName(identifiers)

		// If followed by ( it's a method call, otherwise property access
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (

			fc := &ast.FunctionCall{
				CallTarget: &ast.UserDefinedTypeCallTarget{
					SchemaObjectName: schemaObjName,
				},
				FunctionName:     name,
				UniqueRowFilter:  "NotSpecified",
				WithArrayWrapper: false,
			}

			// Parse parameters
			if p.curTok.Type != TokenRParen {
				for {
					param, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					fc.Parameters = append(fc.Parameters, param)

					if p.curTok.Type != TokenComma {
						break
					}
					p.nextToken() // consume comma
				}
			}

			// Expect )
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ) in function call, got %s", p.curTok.Literal)
			}
			p.nextToken()

			// Check for OVER clause or property access after method call
			return p.parsePostExpressionAccess(fc)
		}

		// Property access: t::a
		propAccess := &ast.UserDefinedTypePropertyAccess{
			CallTarget: &ast.UserDefinedTypeCallTarget{
				SchemaObjectName: schemaObjName,
			},
			PropertyName: name,
		}

		// Check for COLLATE clause
		if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
			p.nextToken() // consume COLLATE
			propAccess.Collation = p.parseIdentifier()
		}

		// Check for chained property access
		return p.parsePostExpressionAccess(propAccess)
	}

	// If followed by ( it's a function call
	if p.curTok.Type == TokenLParen {
		return p.parseFunctionCallFromIdentifiers(identifiers)
	}

	return &ast.ColumnReferenceExpression{
		ColumnType: "Regular",
		MultiPartIdentifier: &ast.MultiPartIdentifier{
			Count:       len(identifiers),
			Identifiers: identifiers,
		},
	}, nil
}

func (p *Parser) parseColumnReference() (*ast.ColumnReferenceExpression, error) {
	expr, err := p.parseColumnReferenceOrFunctionCall()
	if err != nil {
		return nil, err
	}
	if colRef, ok := expr.(*ast.ColumnReferenceExpression); ok {
		return colRef, nil
	}
	// If we got a function call, wrap it in a column reference (shouldn't happen in this context)
	return nil, fmt.Errorf("expected column reference, got function call")
}

func (p *Parser) parseFunctionCallFromIdentifiers(identifiers []*ast.Identifier) (ast.ScalarExpression, error) {
	fc := &ast.FunctionCall{
		UniqueRowFilter:  "NotSpecified",
		WithArrayWrapper: false,
	}

	if len(identifiers) == 1 {
		// Simple function call: func()
		fc.FunctionName = identifiers[0]
	} else {
		// Function call with call target: schema.func() or db.schema.func()
		// The last identifier is the function name, the rest form the call target
		fc.FunctionName = identifiers[len(identifiers)-1]
		fc.CallTarget = &ast.MultiPartIdentifierCallTarget{
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Count:       len(identifiers) - 1,
				Identifiers: identifiers[:len(identifiers)-1],
			},
		}
	}

	// Consume (
	p.nextToken()

	// Check for ALL or DISTINCT
	if strings.ToUpper(p.curTok.Literal) == "ALL" {
		fc.UniqueRowFilter = "All"
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "DISTINCT" {
		fc.UniqueRowFilter = "Distinct"
		p.nextToken()
	}

	// Parse parameters
	if p.curTok.Type != TokenRParen {
		for {
			param, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			fc.Parameters = append(fc.Parameters, param)

			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume comma
		}
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in function call, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for OVER clause or property access after function call
	return p.parsePostExpressionAccess(fc)
}

// parsePostExpressionAccess handles chained property access (.PropertyName), COLLATE clauses, and OVER clauses
// after an expression (function call, parenthesized expression, or property access).
func (p *Parser) parsePostExpressionAccess(expr ast.ScalarExpression) (ast.ScalarExpression, error) {
	// Loop to handle chained property access like .SomeProperty.AnotherProperty
	for {
		// Check for .PropertyName pattern (property access)
		if p.curTok.Type == TokenDot {
			p.nextToken() // consume .

			if p.curTok.Type != TokenIdent {
				return nil, fmt.Errorf("expected property name after ., got %s", p.curTok.Literal)
			}
			propName := &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
			p.nextToken()

			// Check if it's a method call: .method()
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (

				fc := &ast.FunctionCall{
					CallTarget: &ast.ExpressionCallTarget{
						Expression: expr,
					},
					FunctionName:     propName,
					UniqueRowFilter:  "NotSpecified",
					WithArrayWrapper: false,
				}

				// Parse parameters
				if p.curTok.Type != TokenRParen {
					for {
						param, err := p.parseScalarExpression()
						if err != nil {
							return nil, err
						}
						fc.Parameters = append(fc.Parameters, param)

						if p.curTok.Type != TokenComma {
							break
						}
						p.nextToken() // consume comma
					}
				}

				// Expect )
				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ) in method call, got %s", p.curTok.Literal)
				}
				p.nextToken()

				expr = fc
				continue
			}

			// Property access: .PropertyName
			propAccess := &ast.UserDefinedTypePropertyAccess{
				CallTarget: &ast.ExpressionCallTarget{
					Expression: expr,
				},
				PropertyName: propName,
			}

			// Check for COLLATE clause
			if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
				p.nextToken() // consume COLLATE
				propAccess.Collation = p.parseIdentifier()
			}

			expr = propAccess
			continue
		}

		// Check for OVER clause for function calls
		if fc, ok := expr.(*ast.FunctionCall); ok && strings.ToUpper(p.curTok.Literal) == "OVER" {
			p.nextToken() // consume OVER

			if p.curTok.Type != TokenLParen {
				return nil, fmt.Errorf("expected ( after OVER, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume (

			// For now, just skip to closing paren (basic OVER() support)
			// TODO: Parse partition by, order by, and window frame
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ) in OVER clause, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )

			fc.OverClause = &ast.OverClause{}
		}

		break
	}

	return expr, nil
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
			// Lenient: if we can't parse a table reference, return what we have
			return fc, nil
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
	baseRef, err := p.parseSingleTableReference()
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

			right, err := p.parseSingleTableReference()
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

		right, err := p.parseSingleTableReference()
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

func (p *Parser) parseSingleTableReference() (ast.TableReference, error) {
	// Check for OPENROWSET
	if p.curTok.Type == TokenOpenRowset {
		return p.parseOpenRowset()
	}

	// Check for variable table reference
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		name := p.curTok.Literal
		p.nextToken()
		return &ast.VariableTableReference{
			Variable: &ast.VariableReference{Name: name},
			ForPath:  false,
		}, nil
	}

	return p.parseNamedTableReference()
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
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" {
			ref.Alias = &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
			p.nextToken()
		}
	}

	// Parse optional table hints WITH (hint, hint, ...) or old-style (hint, hint, ...)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
	}
	if p.curTok.Type == TokenLParen {
		// Check if this looks like hints (first token is a hint keyword)
		// Save position to peek
		if p.peekIsTableHint() {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				hint, err := p.parseTableHint()
				if err != nil {
					return nil, err
				}
				if hint != nil {
					ref.TableHints = append(ref.TableHints, hint)
				}
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else if p.curTok.Type != TokenRParen {
					// Check if the next token is a valid table hint (space-separated hints)
					if p.isTableHintToken() {
						continue // Continue parsing space-separated hints
					}
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken()
			}
		}
	}

	return ref, nil
}

// parseTableHint parses a single table hint
func (p *Parser) parseTableHint() (ast.TableHintType, error) {
	hintName := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume hint name

	// INDEX hint with values
	if hintName == "INDEX" {
		hint := &ast.IndexTableHint{
			HintKind: "Index",
		}
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				var iov *ast.IdentifierOrValueExpression
				if p.curTok.Type == TokenNumber {
					iov = &ast.IdentifierOrValueExpression{
						Value: p.curTok.Literal,
						ValueExpression: &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						},
					}
					p.nextToken()
				} else if p.curTok.Type == TokenIdent {
					iov = &ast.IdentifierOrValueExpression{
						Value: p.curTok.Literal,
						Identifier: &ast.Identifier{
							Value:     p.curTok.Literal,
							QuoteType: "NotQuoted",
						},
					}
					p.nextToken()
				}
				if iov != nil {
					hint.IndexValues = append(hint.IndexValues, iov)
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
		return hint, nil
	}

	// Map hint names to HintKind
	hintKind := getTableHintKind(hintName)
	if hintKind == "" {
		return nil, nil // Unknown hint
	}

	return &ast.TableHint{
		HintKind: hintKind,
	}, nil
}

// getTableHintKind maps SQL hint names to their AST HintKind values
func getTableHintKind(name string) string {
	switch name {
	case "HOLDLOCK":
		return "HoldLock"
	case "NOLOCK":
		return "NoLock"
	case "PAGLOCK":
		return "PagLock"
	case "READCOMMITTED":
		return "ReadCommitted"
	case "READPAST":
		return "ReadPast"
	case "READUNCOMMITTED":
		return "ReadUncommitted"
	case "REPEATABLEREAD":
		return "RepeatableRead"
	case "ROWLOCK":
		return "Rowlock"
	case "SERIALIZABLE":
		return "Serializable"
	case "SNAPSHOT":
		return "Snapshot"
	case "TABLOCK":
		return "TabLock"
	case "TABLOCKX":
		return "TabLockX"
	case "UPDLOCK":
		return "UpdLock"
	case "XLOCK":
		return "XLock"
	case "NOWAIT":
		return "NoWait"
	default:
		return ""
	}
}

// isTableHintToken checks if the current token is a valid table hint keyword
func (p *Parser) isTableHintToken() bool {
	// Check for keyword tokens that are table hints
	if p.curTok.Type == TokenHoldlock || p.curTok.Type == TokenNowait {
		return true
	}
	// Check for identifiers that are table hints
	if p.curTok.Type == TokenIdent {
		switch strings.ToUpper(p.curTok.Literal) {
		case "HOLDLOCK", "NOLOCK", "PAGLOCK", "READCOMMITTED", "READPAST",
			"READUNCOMMITTED", "REPEATABLEREAD", "ROWLOCK", "SERIALIZABLE",
			"SNAPSHOT", "TABLOCK", "TABLOCKX", "UPDLOCK", "XLOCK", "NOWAIT",
			"INDEX", "FORCESEEK", "FORCESCAN", "KEEPIDENTITY", "KEEPDEFAULTS",
			"IGNORE_CONSTRAINTS", "IGNORE_TRIGGERS", "NOEXPAND", "SPATIAL_WINDOW_MAX_CELLS":
			return true
		}
	}
	return false
}

// peekIsTableHint checks if the peek token (next token after current) is a valid table hint keyword
func (p *Parser) peekIsTableHint() bool {
	// Check for keyword tokens that are table hints
	if p.peekTok.Type == TokenHoldlock || p.peekTok.Type == TokenNowait || p.peekTok.Type == TokenIndex {
		return true
	}
	// Check for identifiers that are table hints
	if p.peekTok.Type == TokenIdent {
		switch strings.ToUpper(p.peekTok.Literal) {
		case "HOLDLOCK", "NOLOCK", "PAGLOCK", "READCOMMITTED", "READPAST",
			"READUNCOMMITTED", "REPEATABLEREAD", "ROWLOCK", "SERIALIZABLE",
			"SNAPSHOT", "TABLOCK", "TABLOCKX", "UPDLOCK", "XLOCK", "NOWAIT",
			"INDEX", "FORCESEEK", "FORCESCAN", "KEEPIDENTITY", "KEEPDEFAULTS",
			"IGNORE_CONSTRAINTS", "IGNORE_TRIGGERS", "NOEXPAND", "SPATIAL_WINDOW_MAX_CELLS":
			return true
		}
	}
	return false
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

		// Accept identifiers and bracketed identifiers, as well as keywords
		// that can be used as object names (like MASTER, KEY, etc.)
		if p.curTok.Type != TokenIdent && p.curTok.Type != TokenLBracket && !p.isKeywordAsIdentifier() {
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

func (p *Parser) parseOptionClause() ([]ast.OptimizerHintBase, error) {
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

	var hints []ast.OptimizerHintBase

	// Parse hints
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}

		hint, err := p.parseOptimizerHint()
		if err != nil {
			return nil, err
		}
		if hint != nil {
			hints = append(hints, hint)
		}
	}

	// Consume )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return hints, nil
}

func (p *Parser) parseOptimizerHint() (ast.OptimizerHintBase, error) {
	// Handle both identifiers and keywords that can appear as optimizer hints
	// USE is a keyword (TokenUse), so we need to handle it specially
	if p.curTok.Type == TokenUse {
		p.nextToken() // consume USE
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PLAN" {
			p.nextToken() // consume PLAN
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: "UsePlan", Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: "Use"}, nil
	}

	// Handle keyword tokens that can be optimizer hints (ORDER, GROUP, etc.)
	if p.curTok.Type == TokenOrder || p.curTok.Type == TokenGroup {
		hintKind := convertHintKind(p.curTok.Literal)
		firstWord := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		// Check for two-word hints like ORDER GROUP
		if (firstWord == "ORDER" || firstWord == "HASH" || firstWord == "MERGE" ||
			firstWord == "CONCAT" || firstWord == "LOOP" || firstWord == "FORCE") &&
			(p.curTok.Type == TokenIdent || p.curTok.Type == TokenGroup) {
			secondWord := strings.ToUpper(p.curTok.Literal)
			if secondWord == "GROUP" || secondWord == "JOIN" || secondWord == "UNION" ||
				secondWord == "ORDER" {
				hintKind = hintKind + convertHintKind(p.curTok.Literal)
				p.nextToken()
			}
		}
		return &ast.OptimizerHint{HintKind: hintKind}, nil
	}

	// Handle TABLE HINT optimizer hint
	if p.curTok.Type == TokenTable {
		p.nextToken() // consume TABLE
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "HINT" {
			p.nextToken() // consume HINT
			return p.parseTableHintsOptimizerHint()
		}
		return &ast.OptimizerHint{HintKind: "Table"}, nil
	}

	if p.curTok.Type != TokenIdent && p.curTok.Type != TokenLabel {
		// Skip unknown tokens to avoid infinite loop
		p.nextToken()
		return nil, nil
	}

	upper := strings.ToUpper(p.curTok.Literal)

	switch upper {
	case "PARAMETERIZATION":
		p.nextToken() // consume PARAMETERIZATION
		if p.curTok.Type == TokenIdent {
			subUpper := strings.ToUpper(p.curTok.Literal)
			p.nextToken()
			if subUpper == "SIMPLE" {
				return &ast.OptimizerHint{HintKind: "ParameterizationSimple"}, nil
			} else if subUpper == "FORCED" {
				return &ast.OptimizerHint{HintKind: "ParameterizationForced"}, nil
			}
		}
		return &ast.OptimizerHint{HintKind: "Parameterization"}, nil

	case "MAXRECURSION":
		p.nextToken() // consume MAXRECURSION
		value, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.LiteralOptimizerHint{HintKind: "MaxRecursion", Value: value}, nil

	case "OPTIMIZE":
		p.nextToken() // consume OPTIMIZE
		if p.curTok.Type == TokenIdent {
			subUpper := strings.ToUpper(p.curTok.Literal)
			if subUpper == "FOR" {
				p.nextToken() // consume FOR
				return p.parseOptimizeForHint()
			} else if subUpper == "CORRELATED" {
				p.nextToken() // consume CORRELATED
				if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "UNION" {
					p.nextToken() // consume UNION
					if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "ALL" {
						p.nextToken() // consume ALL
					}
				}
				return &ast.OptimizerHint{HintKind: "OptimizeCorrelatedUnionAll"}, nil
			}
		}
		return &ast.OptimizerHint{HintKind: "Optimize"}, nil

	case "CHECKCONSTRAINTS":
		p.nextToken() // consume CHECKCONSTRAINTS
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PLAN" {
			p.nextToken() // consume PLAN
			return &ast.OptimizerHint{HintKind: "CheckConstraintsPlan"}, nil
		}
		return &ast.OptimizerHint{HintKind: "CheckConstraints"}, nil

	case "LABEL":
		p.nextToken() // consume LABEL
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: "Label", Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: "Label"}, nil

	case "MAX_GRANT_PERCENT":
		p.nextToken() // consume MAX_GRANT_PERCENT
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: "MaxGrantPercent", Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: "MaxGrantPercent"}, nil

	case "MIN_GRANT_PERCENT":
		p.nextToken() // consume MIN_GRANT_PERCENT
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: "MinGrantPercent", Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: "MinGrantPercent"}, nil

	case "FAST":
		p.nextToken() // consume FAST
		// FAST can take a numeric argument
		if p.curTok.Type == TokenNumber {
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: "Fast", Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: "Fast"}, nil

	default:
		// Handle generic hints
		hintKind := convertHintKind(p.curTok.Literal)
		firstWord := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		// Check for two-word hints like ORDER GROUP, HASH GROUP, etc.
		if (firstWord == "ORDER" || firstWord == "HASH" || firstWord == "MERGE" ||
			firstWord == "CONCAT" || firstWord == "LOOP" || firstWord == "FORCE") &&
			p.curTok.Type == TokenIdent {
			secondWord := strings.ToUpper(p.curTok.Literal)
			if secondWord == "GROUP" || secondWord == "JOIN" || secondWord == "UNION" ||
				secondWord == "ORDER" {
				hintKind = hintKind + convertHintKind(p.curTok.Literal)
				p.nextToken()
			}
		}

		// Check if this is a literal hint (LABEL = value, etc.)
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: hintKind, Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: hintKind}, nil
	}
}

func (p *Parser) parseTableHintsOptimizerHint() (ast.OptimizerHintBase, error) {
	hint := &ast.TableHintsOptimizerHint{
		HintKind: "TableHints",
	}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TABLE HINT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse object name
	objectName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	hint.ObjectName = objectName

	// Expect comma
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume comma
	}

	// Parse table hints
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}

		tableHint, err := p.parseTableHint()
		if err != nil {
			return nil, err
		}
		if tableHint != nil {
			hint.TableHints = append(hint.TableHints, tableHint)
		}
	}

	// Consume )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return hint, nil
}

func (p *Parser) parseOptimizeForHint() (ast.OptimizerHintBase, error) {
	hint := &ast.OptimizeForOptimizerHint{
		HintKind:     "OptimizeFor",
		IsForUnknown: false,
	}

	// Check for UNKNOWN
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "UNKNOWN" {
		p.nextToken()
		hint.IsForUnknown = true
		return hint, nil
	}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after OPTIMIZE FOR, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse variable-value pairs
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}

		pair, err := p.parseVariableValuePair()
		if err != nil {
			return nil, err
		}
		if pair != nil {
			hint.Pairs = append(hint.Pairs, pair)
		}
	}

	// Consume )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return hint, nil
}

func (p *Parser) parseVariableValuePair() (*ast.VariableValuePair, error) {
	// Expect @variable (variables are TokenIdent starting with @)
	if p.curTok.Type != TokenIdent || !strings.HasPrefix(p.curTok.Literal, "@") {
		return nil, nil
	}

	pair := &ast.VariableValuePair{
		Variable: &ast.VariableReference{
			Name: p.curTok.Literal,
		},
		IsForUnknown: false,
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		// Could be UNKNOWN
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "UNKNOWN" {
			p.nextToken()
			pair.IsForUnknown = true
			return pair, nil
		}
		return nil, fmt.Errorf("expected = after variable, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume =

	// Parse the value
	value, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	pair.Value = value

	return pair, nil
}

// convertHintKind converts hint identifiers to their canonical names
func convertHintKind(hint string) string {
	// Map common hint names
	hintMap := map[string]string{
		"IGNORE_NONCLUSTERED_COLUMNSTORE_INDEX": "IgnoreNonClusteredColumnStoreIndex",
		"LABEL":                        "Label",
		"MAX_GRANT_PERCENT":            "MaxGrantPercent",
		"MIN_GRANT_PERCENT":            "MinGrantPercent",
		"NO_PERFORMANCE_SPOOL":         "NoPerformanceSpool",
		"PARAMETERIZATION":             "Parameterization",
		"RECOMPILE":                    "Recompile",
		"MAXRECURSION":                 "MaxRecursion",
		"KEEPFIXED":                    "KeepFixed",
		"KEEP":                         "Keep",
		"EXPAND":                       "Expand",
		"VIEWS":                        "Views",
		"HASH":                         "Hash",
		"ORDER":                        "Order",
		"GROUP":                        "Group",
		"MERGE":                        "Merge",
		"CONCAT":                       "Concat",
		"UNION":                        "Union",
		"LOOP":                         "Loop",
		"JOIN":                         "Join",
		"FAST":                         "Fast",
		"FORCE":                        "Force",
		"ROBUST":                       "Robust",
		"PLAN":                         "Plan",
		"USE":                          "Use",
		"SIMPLE":                       "Simple",
		"FORCED":                       "Forced",
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
	// Check for parenthesized boolean expression
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		// Parse inner boolean expression
		inner, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		return &ast.BooleanParenthesisExpression{Expression: inner}, nil
	}

	// Parse left scalar expression
	left, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Check for NOT before IN/LIKE/BETWEEN
	notDefined := false
	if p.curTok.Type == TokenNot {
		notDefined = true
		p.nextToken() // consume NOT
	}

	// Check for IS NULL / IS NOT NULL
	if p.curTok.Type == TokenIs {
		p.nextToken() // consume IS

		isNot := false
		if p.curTok.Type == TokenNot {
			isNot = true
			p.nextToken() // consume NOT
		}

		if p.curTok.Type != TokenNull {
			return nil, fmt.Errorf("expected NULL after IS/IS NOT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume NULL

		return &ast.BooleanIsNullExpression{
			IsNot:      isNot,
			Expression: left,
		}, nil
	}

	// Check for IN expression
	if p.curTok.Type == TokenIn {
		p.nextToken() // consume IN

		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after IN, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		// Check if it's a subquery or value list
		if p.curTok.Type == TokenSelect {
			subquery, err := p.parseQueryExpression()
			if err != nil {
				return nil, err
			}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )
			return &ast.BooleanInExpression{
				Expression: left,
				NotDefined: notDefined,
				Subquery:   subquery,
			}, nil
		}

		// Parse value list
		var values []ast.ScalarExpression
		for {
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			values = append(values, val)
			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume ,
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
		return &ast.BooleanInExpression{
			Expression: left,
			NotDefined: notDefined,
			Values:     values,
		}, nil
	}

	// Check for LIKE expression
	if p.curTok.Type == TokenLike {
		p.nextToken() // consume LIKE

		pattern, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		var escapeExpr ast.ScalarExpression
		if p.curTok.Type == TokenEscape {
			p.nextToken() // consume ESCAPE
			escapeExpr, err = p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
		}

		return &ast.BooleanLikeExpression{
			FirstExpression:  left,
			SecondExpression: pattern,
			EscapeExpression: escapeExpr,
			NotDefined:       notDefined,
		}, nil
	}

	// Check for BETWEEN expression
	if p.curTok.Type == TokenBetween {
		p.nextToken() // consume BETWEEN

		low, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		if p.curTok.Type != TokenAnd {
			return nil, fmt.Errorf("expected AND in BETWEEN, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume AND

		high, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		ternaryType := "Between"
		if notDefined {
			ternaryType = "NotBetween"
		}
		return &ast.BooleanTernaryExpression{
			TernaryExpressionType: ternaryType,
			FirstExpression:       left,
			SecondExpression:      low,
			ThirdExpression:       high,
		}, nil
	}

	// If we saw NOT but didn't get IN/LIKE/BETWEEN, error
	if notDefined {
		return nil, fmt.Errorf("expected IN, LIKE, or BETWEEN after NOT, got %s", p.curTok.Literal)
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

// identifiersToSchemaObjectName converts a slice of identifiers to a SchemaObjectName.
// For 1 identifier: BaseIdentifier
// For 2 identifiers: SchemaIdentifier.BaseIdentifier
// For 3 identifiers: DatabaseIdentifier.SchemaIdentifier.BaseIdentifier
// For 4 identifiers: ServerIdentifier.DatabaseIdentifier.SchemaIdentifier.BaseIdentifier
func identifiersToSchemaObjectName(identifiers []*ast.Identifier) *ast.SchemaObjectName {
	son := &ast.SchemaObjectName{
		Count:       len(identifiers),
		Identifiers: identifiers,
	}

	switch len(identifiers) {
	case 1:
		son.BaseIdentifier = identifiers[0]
	case 2:
		son.SchemaIdentifier = identifiers[0]
		son.BaseIdentifier = identifiers[1]
	case 3:
		son.DatabaseIdentifier = identifiers[0]
		son.SchemaIdentifier = identifiers[1]
		son.BaseIdentifier = identifiers[2]
	case 4:
		son.ServerIdentifier = identifiers[0]
		son.DatabaseIdentifier = identifiers[1]
		son.SchemaIdentifier = identifiers[2]
		son.BaseIdentifier = identifiers[3]
	}

	return son
}

// ======================= New Statement Parsing Functions =======================


// parseCastCall parses a CAST expression: CAST(expression AS data_type)
func (p *Parser) parseCastCall() (ast.ScalarExpression, error) {
	p.nextToken() // consume CAST
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after CAST, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse the expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Expect AS
	if p.curTok.Type != TokenAs {
		return nil, fmt.Errorf("expected AS in CAST, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume AS

	// Parse the data type
	dt, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in CAST, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	cast := &ast.CastCall{
		DataType:  dt,
		Parameter: expr,
	}

	// Check for COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		cast.Collation = p.parseIdentifier()
	}

	return cast, nil
}

// parseConvertCall parses a CONVERT expression: CONVERT(data_type, expression [, style])
func (p *Parser) parseConvertCall() (ast.ScalarExpression, error) {
	p.nextToken() // consume CONVERT
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after CONVERT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse the data type first
	dt, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , in CONVERT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse the expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	convert := &ast.ConvertCall{
		DataType:  dt,
		Parameter: expr,
	}

	// Check for optional style parameter
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		style, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		convert.Style = style
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in CONVERT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Check for COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		convert.Collation = p.parseIdentifier()
	}

	return convert, nil
}

// parseTryCastCall parses a TRY_CAST expression
func (p *Parser) parseTryCastCall() (ast.ScalarExpression, error) {
	p.nextToken() // consume TRY_CAST
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TRY_CAST, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse the expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Expect AS
	if p.curTok.Type != TokenAs {
		return nil, fmt.Errorf("expected AS in TRY_CAST, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume AS

	// Parse the data type
	dt, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in TRY_CAST, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	cast := &ast.TryCastCall{
		DataType:  dt,
		Parameter: expr,
	}

	// Check for COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		cast.Collation = p.parseIdentifier()
	}

	return cast, nil
}

// parseTryConvertCall parses a TRY_CONVERT expression
func (p *Parser) parseTryConvertCall() (ast.ScalarExpression, error) {
	p.nextToken() // consume TRY_CONVERT
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TRY_CONVERT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse the data type first
	dt, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , in TRY_CONVERT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse the expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	convert := &ast.TryConvertCall{
		DataType:  dt,
		Parameter: expr,
	}

	// Check for optional style parameter
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		style, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		convert.Style = style
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in TRY_CONVERT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Check for COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		convert.Collation = p.parseIdentifier()
	}

	return convert, nil
}
