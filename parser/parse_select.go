// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"strings"

	"github.com/kyleconroy/teesql/ast"
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

	for p.curTok.Type == TokenPlus || p.curTok.Type == TokenMinus {
		var opType string
		if p.curTok.Type == TokenPlus {
			opType = "Add"
		} else {
			opType = "Subtract"
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
		return p.parseColumnReference()
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
		return &ast.ParenthesisExpression{Expression: expr}, nil
	case TokenCase:
		return p.parseCaseExpression()
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
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLabel {
			hintKind := convertHintKind(p.curTok.Literal)
			p.nextToken()

			// Check if this is a literal hint (LABEL = value, etc.)
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
				value, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				hints = append(hints, &ast.LiteralOptimizerHint{
					HintKind: hintKind,
					Value:    value,
				})
			} else {
				hints = append(hints, &ast.OptimizerHint{HintKind: hintKind})
			}
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

// ======================= New Statement Parsing Functions =======================

