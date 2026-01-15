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

	// Check for parenthesized WITH expression: (WITH ... SELECT ...)
	// Only handle this case specially, let normal parsing handle other cases
	if p.curTok.Type == TokenLParen && p.peekTok.Type == TokenWith {
		p.nextToken() // consume (
		// Parse WITH clause and SELECT statement
		withStmt, err := p.parseWithStatement()
		if err != nil {
			return nil, err
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		// Return the SelectStatement with its WithCtesAndXmlNamespaces
		if selStmt, ok := withStmt.(*ast.SelectStatement); ok {
			return selStmt, nil
		}
		return nil, fmt.Errorf("expected SELECT statement after WITH, got %T", withStmt)
	}

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

	// Parse OFFSET...FETCH clause after ORDER BY
	if strings.ToUpper(p.curTok.Literal) == "OFFSET" {
		oc, err := p.parseOffsetClause()
		if err != nil {
			return nil, nil, nil, err
		}
		if qs, ok := left.(*ast.QuerySpecification); ok {
			qs.OffsetClause = oc
		}
	}

	// Parse FOR clause (FOR BROWSE, FOR XML, FOR UPDATE, FOR READ ONLY)
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		forClause, err := p.parseForClause()
		if err != nil {
			return nil, nil, nil, err
		}
		// Attach to QuerySpecification
		if qs, ok := left.(*ast.QuerySpecification); ok {
			qs.ForClause = forClause
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

	// Parse optional WINDOW clause
	if strings.ToUpper(p.curTok.Literal) == "WINDOW" {
		windowClause, err := p.parseWindowClause()
		if err != nil {
			return nil, nil, nil, err
		}
		qs.WindowClause = windowClause
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

		// Check for subquery (SELECT ...) or parenthesized query expression starting with (
		if p.curTok.Type == TokenSelect || p.curTok.Type == TokenLParen {
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
		varRef := &ast.VariableReference{Name: varName}

		// Handle postfix operations (method calls, property access)
		expr, err := p.handlePostfixOperations(varRef)
		if err != nil {
			return nil, err
		}

		// Check if next token is a binary operator - if so, continue parsing the expression
		for p.curTok.Type == TokenPlus || p.curTok.Type == TokenMinus ||
			p.curTok.Type == TokenStar || p.curTok.Type == TokenSlash ||
			p.curTok.Type == TokenPercent || p.curTok.Type == TokenDoublePipe {
			// We have a variable followed by a binary operator, continue parsing
			var opType string
			switch p.curTok.Type {
			case TokenPlus:
				opType = "Add"
			case TokenMinus:
				opType = "Subtract"
			case TokenStar:
				opType = "Multiply"
			case TokenSlash:
				opType = "Divide"
			case TokenPercent:
				opType = "Modulo"
			case TokenDoublePipe:
				opType = "Add" // String concatenation
			}
			p.nextToken() // consume operator
			right, err := p.parsePrimaryExpression()
			if err != nil {
				return nil, err
			}
			expr = &ast.BinaryExpression{
				FirstExpression:      expr,
				SecondExpression:     right,
				BinaryExpressionType: opType,
			}
		}

		sse := &ast.SelectScalarExpression{Expression: expr}

		// Check for column alias
		if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '[' {
			alias := p.parseIdentifier()
			sse.ColumnName = &ast.IdentifierOrValueExpression{
				Value:      alias.Value,
				Identifier: alias,
			}
		} else if p.curTok.Type == TokenAs {
			p.nextToken()
			if p.curTok.Type == TokenString {
				// String literal alias: AS 'alias'
				str := p.parseStringLiteralValue()
				p.nextToken()
				sse.ColumnName = &ast.IdentifierOrValueExpression{
					Value:           str.Value,
					ValueExpression: str,
				}
			} else {
				alias := p.parseIdentifier()
				sse.ColumnName = &ast.IdentifierOrValueExpression{
					Value:      alias.Value,
					Identifier: alias,
				}
			}
		} else if p.curTok.Type == TokenIdent {
			upper := strings.ToUpper(p.curTok.Literal)
			if upper != "FROM" && upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "INTO" && upper != "UNION" && upper != "EXCEPT" && upper != "INTERSECT" && upper != "GO" {
				alias := p.parseIdentifier()
				sse.ColumnName = &ast.IdentifierOrValueExpression{
					Value:      alias.Value,
					Identifier: alias,
				}
			}
		}
		return sse, nil
	}

	// Check for equals-style alias: column1 = expr, [column1] = expr, 'string' = expr, N'string' = expr
	// This is detected by seeing identifier or string followed by =
	if p.peekTok.Type == TokenEquals {
		// We have alias = expr pattern
		var columnName *ast.IdentifierOrValueExpression

		if p.curTok.Type == TokenIdent {
			// identifier = expr
			alias := p.parseIdentifier()
			columnName = &ast.IdentifierOrValueExpression{
				Value:      alias.Value,
				Identifier: alias,
			}
		} else if p.curTok.Type == TokenString {
			// 'string' = expr
			str := p.parseStringLiteralValue()
			p.nextToken()
			columnName = &ast.IdentifierOrValueExpression{
				Value:           str.Value,
				ValueExpression: str,
			}
		} else if p.curTok.Type == TokenNationalString {
			// N'string' = expr
			str, _ := p.parseNationalStringFromToken()
			columnName = &ast.IdentifierOrValueExpression{
				Value:           str.Value,
				ValueExpression: str,
			}
		}

		if columnName != nil {
			p.nextToken() // consume =

			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			return &ast.SelectScalarExpression{
				Expression: expr,
				ColumnName: columnName,
			}, nil
		}
	}

	// Otherwise parse a scalar expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Check for qualified star: expression followed by .*
	// This happens when parseColumnReferenceOrFunctionCall stopped before consuming .*
	if p.curTok.Type == TokenDot && p.peekTok.Type == TokenStar {
		// Convert expression to qualified star
		if colRef, ok := expr.(*ast.ColumnReferenceExpression); ok {
			p.nextToken() // consume .
			p.nextToken() // consume *
			return &ast.SelectStarExpression{
				Qualifier: colRef.MultiPartIdentifier,
			}, nil
		}
	}

	// Check for COLLATE clause before creating SelectScalarExpression
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		collation := p.parseIdentifier()
		// Attach collation to the expression
		switch e := expr.(type) {
		case *ast.FunctionCall:
			e.Collation = collation
		case *ast.ColumnReferenceExpression:
			e.Collation = collation
		}
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
		if p.curTok.Type == TokenString {
			// String literal alias: AS 'alias'
			str := p.parseStringLiteralValue()
			p.nextToken()
			sse.ColumnName = &ast.IdentifierOrValueExpression{
				Value:           str.Value,
				ValueExpression: str,
			}
		} else if p.curTok.Type == TokenNationalString {
			// National string literal alias: AS N'alias'
			str, _ := p.parseNationalStringFromToken()
			sse.ColumnName = &ast.IdentifierOrValueExpression{
				Value:           str.Value,
				ValueExpression: str,
			}
		} else {
			alias := p.parseIdentifier()
			sse.ColumnName = &ast.IdentifierOrValueExpression{
				Value:      alias.Value,
				Identifier: alias,
			}
		}
	} else if p.curTok.Type == TokenString {
		// String literal alias without AS: expr 'alias'
		str := p.parseStringLiteralValue()
		p.nextToken()
		sse.ColumnName = &ast.IdentifierOrValueExpression{
			Value:           str.Value,
			ValueExpression: str,
		}
	} else if p.curTok.Type == TokenNationalString {
		// National string literal alias without AS: expr N'alias'
		str, _ := p.parseNationalStringFromToken()
		sse.ColumnName = &ast.IdentifierOrValueExpression{
			Value:           str.Value,
			ValueExpression: str,
		}
	} else if p.curTok.Type == TokenIdent {
		// Check if this is an alias (not a keyword that starts a new clause)
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "FROM" && upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "INTO" && upper != "UNION" && upper != "EXCEPT" && upper != "INTERSECT" && upper != "GO" && upper != "COLLATE" {
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
		TokenTyp, TokenScoped:
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
	left, err := p.parsePostfixExpression()
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

		right, err := p.parsePostfixExpression()
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

// parsePostfixExpression handles postfix operators like AT TIME ZONE
func (p *Parser) parsePostfixExpression() (ast.ScalarExpression, error) {
	expr, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	// Handle postfix operations: method calls, property access, AT TIME ZONE
	for {
		// Check for method/property access: expr.func() or expr.prop
		// The next token after the dot must be an identifier (plain or bracketed)
		if p.curTok.Type == TokenDot && (p.peekTok.Type == TokenIdent || (len(p.peekTok.Literal) > 0 && p.peekTok.Literal[0] == '[')) {
			p.nextToken() // consume dot

			if !p.isIdentifierToken() {
				return nil, fmt.Errorf("expected identifier after dot, got %s", p.curTok.Literal)
			}

			// Parse method/property name
			quoteType := "NotQuoted"
			name := p.curTok.Literal
			if len(name) >= 2 && name[0] == '[' && name[len(name)-1] == ']' {
				quoteType = "SquareBracket"
				name = name[1 : len(name)-1]
			}
			methodName := &ast.Identifier{Value: name, QuoteType: quoteType}
			p.nextToken()

			if p.curTok.Type == TokenLParen {
				// It's a method call: expr.func()
				p.nextToken() // consume (

				fc := &ast.FunctionCall{
					CallTarget:       &ast.ExpressionCallTarget{Expression: expr},
					FunctionName:     methodName,
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
					return nil, fmt.Errorf("expected ) after method call, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume )

				// Check for OVER clause
				if strings.ToUpper(p.curTok.Literal) == "OVER" {
					overClause, err := p.parseOverClause()
					if err != nil {
						return nil, err
					}
					fc.OverClause = overClause
				}

				expr = fc
			} else {
				// It's a property access: expr.prop
				expr = &ast.UserDefinedTypePropertyAccess{
					CallTarget: &ast.ExpressionCallTarget{
						Expression: expr,
					},
					PropertyName: methodName,
				}
			}
			continue
		}

		// Check for AT TIME ZONE - only if followed by "TIME"
		if strings.ToUpper(p.curTok.Literal) == "AT" && strings.ToUpper(p.peekTok.Literal) == "TIME" {
			p.nextToken() // consume AT
			p.nextToken() // consume TIME
			if strings.ToUpper(p.curTok.Literal) != "ZONE" {
				return nil, fmt.Errorf("expected ZONE after TIME, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume ZONE

			timezone, err := p.parsePrimaryExpression()
			if err != nil {
				return nil, err
			}

			expr = &ast.AtTimeZoneCall{
				DateValue: expr,
				TimeZone:  timezone,
			}
			continue
		}

		// No more postfix operations
		break
	}

	return expr, nil
}

// handlePostfixOperations handles method calls and property access on an existing expression
func (p *Parser) handlePostfixOperations(expr ast.ScalarExpression) (ast.ScalarExpression, error) {
	for {
		// Check for method/property access: expr.func() or expr.prop
		if p.curTok.Type == TokenDot && (p.peekTok.Type == TokenIdent || (len(p.peekTok.Literal) > 0 && p.peekTok.Literal[0] == '[')) {
			p.nextToken() // consume dot

			// Check for bracket-quoted identifier or regular identifier token
			isBracketQuoted := len(p.curTok.Literal) >= 2 && p.curTok.Literal[0] == '[' && p.curTok.Literal[len(p.curTok.Literal)-1] == ']'
			if !p.isIdentifierToken() && !isBracketQuoted {
				return nil, fmt.Errorf("expected identifier after dot, got %s", p.curTok.Literal)
			}

			// Parse method/property name
			quoteType := "NotQuoted"
			name := p.curTok.Literal
			if len(name) >= 2 && name[0] == '[' && name[len(name)-1] == ']' {
				quoteType = "SquareBracket"
				name = name[1 : len(name)-1]
			}
			methodName := &ast.Identifier{Value: name, QuoteType: quoteType}
			p.nextToken()

			if p.curTok.Type == TokenLParen {
				// It's a method call: expr.func()
				p.nextToken() // consume (

				fc := &ast.FunctionCall{
					CallTarget:       &ast.ExpressionCallTarget{Expression: expr},
					FunctionName:     methodName,
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
					return nil, fmt.Errorf("expected ) after method call, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume )

				// Check for OVER clause
				if strings.ToUpper(p.curTok.Literal) == "OVER" {
					overClause, err := p.parseOverClause()
					if err != nil {
						return nil, err
					}
					fc.OverClause = overClause
				}

				expr = fc
			} else {
				// It's a property access: expr.prop
				expr = &ast.UserDefinedTypePropertyAccess{
					CallTarget: &ast.ExpressionCallTarget{
						Expression: expr,
					},
					PropertyName: methodName,
				}
			}
			continue
		}

		// No more postfix operations
		break
	}

	return expr, nil
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
		if upper == "IDENTITY" && p.peekTok.Type == TokenLParen {
			return p.parseIdentityFunctionCall()
		}
		if upper == "IDENTITYCOL" {
			p.nextToken()
			return &ast.ColumnReferenceExpression{ColumnType: "IdentityCol"}, nil
		}
		if upper == "ROWGUIDCOL" {
			p.nextToken()
			return &ast.ColumnReferenceExpression{ColumnType: "RowGuidCol"}, nil
		}
		if upper == "$ACTION" {
			p.nextToken()
			return &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnAction"}, nil
		}
		if upper == "$IDENTITY" {
			p.nextToken()
			return &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnIdentity"}, nil
		}
		if upper == "$ROWGUID" {
			p.nextToken()
			return &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnRowGuid"}, nil
		}
		if upper == "$CUID" {
			p.nextToken()
			return &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnCuid"}, nil
		}
		// Check for NEXT VALUE FOR sequence expression
		if upper == "NEXT" && strings.ToUpper(p.peekTok.Literal) == "VALUE" {
			return p.parseNextValueForExpression()
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
	case TokenDot:
		// Multi-part identifier starting with empty parts (e.g., ..t1.c1)
		return p.parseColumnReferenceWithLeadingDots()
	case TokenMaster, TokenDatabase, TokenKey, TokenTable, TokenIndex,
		TokenSchema, TokenUser, TokenView, TokenTime:
		// Keywords that can be used as identifiers in column/table references
		return p.parseColumnReferenceOrFunctionCall()
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

// parseNextValueForExpression parses NEXT VALUE FOR sequence_name [OVER (...)]
func (p *Parser) parseNextValueForExpression() (*ast.NextValueForExpression, error) {
	p.nextToken() // consume NEXT
	p.nextToken() // consume VALUE

	// Expect FOR
	if strings.ToUpper(p.curTok.Literal) != "FOR" {
		return nil, fmt.Errorf("expected FOR after NEXT VALUE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume FOR

	expr := &ast.NextValueForExpression{}

	// Parse sequence name (may be multi-part: schema.sequence)
	seqName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	expr.SequenceName = seqName

	// Check for optional OVER clause
	if p.curTok.Type == TokenOver || strings.ToUpper(p.curTok.Literal) == "OVER" {
		overClause, err := p.parseOverClause()
		if err != nil {
			return nil, err
		}
		expr.OverClause = overClause
	}

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
	isNational := false

	// Check for national string (N'...')
	if p.curTok.Type == TokenNationalString {
		isNational = true
		// Strip the N prefix
		if len(raw) >= 3 && (raw[0] == 'N' || raw[0] == 'n') && raw[1] == '\'' {
			raw = raw[1:] // Remove the N, keep the rest including quotes
		}
	}
	p.nextToken()

	// Remove surrounding quotes and handle escaped quotes
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		inner := raw[1 : len(raw)-1]
		// Replace escaped quotes
		value := strings.ReplaceAll(inner, "''", "'")
		return &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    isNational,
			IsLargeObject: false,
			Value:         value,
		}, nil
	}

	return &ast.StringLiteral{
		LiteralType:   "String",
		IsNational:    isNational,
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

func (p *Parser) isIdentifierToken() bool {
	switch p.curTok.Type {
	case TokenIdent, TokenMaster, TokenDatabase, TokenKey, TokenTable, TokenIndex,
		TokenSchema, TokenUser, TokenView, TokenDefault, TokenTyp, TokenLanguage,
		TokenTime:
		return true
	default:
		return false
	}
}

func (p *Parser) parseColumnReferenceOrFunctionCall() (ast.ScalarExpression, error) {
	var identifiers []*ast.Identifier
	colType := "Regular"

	for {
		if !p.isIdentifierToken() {
			break
		}

		quoteType := "NotQuoted"
		literal := p.curTok.Literal
		upper := strings.ToUpper(literal)

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
		} else if upper == "IDENTITYCOL" || upper == "ROWGUIDCOL" {
			// IDENTITYCOL/ROWGUIDCOL at end of multi-part identifier sets column type
			// and is not included in the identifier list
			if upper == "IDENTITYCOL" {
				colType = "IdentityCol"
			} else {
				colType = "RowGuidCol"
			}
			p.nextToken()
			break
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
		// Check if this is a qualified star like d.* - if so, don't consume the dot
		// Let the caller handle the .* pattern
		if p.peekTok.Type == TokenStar {
			break
		}
		p.nextToken() // consume dot

		// Handle consecutive dots (empty parts in multi-part identifier)
		for p.curTok.Type == TokenDot {
			identifiers = append(identifiers, &ast.Identifier{Value: "", QuoteType: "NotQuoted"})
			p.nextToken() // consume dot
		}
	}

	// Check for $PARTITION function call: [db.]$PARTITION.func(args)
	if len(identifiers) >= 2 && p.curTok.Type == TokenLParen {
		// Check if $PARTITION is in the identifiers
		partitionIdx := -1
		for i, id := range identifiers {
			if strings.ToUpper(id.Value) == "$PARTITION" {
				partitionIdx = i
				break
			}
		}

		if partitionIdx >= 0 {
			// Build PartitionFunctionCall
			pfc := &ast.PartitionFunctionCall{}

			// DatabaseName comes before $PARTITION if present
			if partitionIdx == 1 {
				pfc.DatabaseName = identifiers[0]
			}

			// FunctionName comes after $PARTITION
			if partitionIdx+1 < len(identifiers) {
				pfc.FunctionName = identifiers[partitionIdx+1]
			}

			// Parse parameters
			p.nextToken() // consume (
			if p.curTok.Type != TokenRParen {
				for {
					param, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					pfc.Parameters = append(pfc.Parameters, param)

					if p.curTok.Type != TokenComma {
						break
					}
					p.nextToken() // consume comma
				}
			}

			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ) in $PARTITION function call, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )

			return pfc, nil
		}
	}

	// Check for :: (user-defined type method call or property access): a.b::func() or a::prop
	if p.curTok.Type == TokenColonColon && len(identifiers) > 0 {
		p.nextToken() // consume ::

		// Parse function/property name - can be regular identifier or bracket-quoted
		isBracketQuoted := len(p.curTok.Literal) >= 2 && p.curTok.Literal[0] == '[' && p.curTok.Literal[len(p.curTok.Literal)-1] == ']'
		if p.curTok.Type != TokenIdent && !isBracketQuoted {
			return nil, fmt.Errorf("expected identifier after ::, got %s", p.curTok.Literal)
		}
		nameValue := p.curTok.Literal
		quoteType := "NotQuoted"
		if isBracketQuoted {
			quoteType = "SquareBracket"
			nameValue = nameValue[1 : len(nameValue)-1]
		}
		name := &ast.Identifier{Value: nameValue, QuoteType: quoteType}
		p.nextToken()

		// Build SchemaObjectName from identifiers (filter out empty identifiers from leading dots)
		var nonEmptyIdents []*ast.Identifier
		for _, id := range identifiers {
			if id.Value != "" {
				nonEmptyIdents = append(nonEmptyIdents, id)
			}
		}
		schemaObjName := identifiersToSchemaObjectName(nonEmptyIdents)

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

	// If we have identifiers, build a column reference with them
	if len(identifiers) > 0 {
		return &ast.ColumnReferenceExpression{
			ColumnType: colType,
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Count:       len(identifiers),
				Identifiers: identifiers,
			},
		}, nil
	}

	// No identifiers means just IDENTITYCOL or ROWGUIDCOL (already handled in parsePrimaryExpression)
	// but handle the case anyway
	return &ast.ColumnReferenceExpression{
		ColumnType: colType,
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

func (p *Parser) parseColumnReferenceWithLeadingDots() (ast.ScalarExpression, error) {
	// Handle multi-part identifiers starting with dots like ..t1.c1 or .db..t1.c1
	var identifiers []*ast.Identifier

	// Add empty identifiers for leading dots
	for p.curTok.Type == TokenDot {
		identifiers = append(identifiers, &ast.Identifier{Value: "", QuoteType: "NotQuoted"})
		p.nextToken() // consume dot
	}

	// Now parse the remaining identifiers
	for p.isIdentifierToken() {
		quoteType := "NotQuoted"
		literal := p.curTok.Literal
		// Handle special column types
		upper := strings.ToUpper(literal)
		if upper == "IDENTITYCOL" || upper == "ROWGUIDCOL" {
			// Return with the proper column type
			colType := "IdentityCol"
			if upper == "ROWGUIDCOL" {
				colType = "RowGuidCol"
			}
			p.nextToken()
			return &ast.ColumnReferenceExpression{
				ColumnType: colType,
				MultiPartIdentifier: &ast.MultiPartIdentifier{
					Count:       len(identifiers),
					Identifiers: identifiers,
				},
			}, nil
		}
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
		// Check for qualified star
		if p.peekTok.Type == TokenStar {
			break
		}
		p.nextToken() // consume dot
	}

	// Don't consume .* here - let the caller (parseSelectElement) handle qualified stars

	// Check for :: (user-defined type method call or property access): .t::func() or .t::prop
	if p.curTok.Type == TokenColonColon && len(identifiers) > 0 {
		p.nextToken() // consume ::

		// Parse function/property name - can be regular identifier or bracket-quoted
		isBracketQuoted := len(p.curTok.Literal) >= 2 && p.curTok.Literal[0] == '[' && p.curTok.Literal[len(p.curTok.Literal)-1] == ']'
		if p.curTok.Type != TokenIdent && !isBracketQuoted {
			return nil, fmt.Errorf("expected identifier after ::, got %s", p.curTok.Literal)
		}
		nameValue := p.curTok.Literal
		quoteType := "NotQuoted"
		if isBracketQuoted {
			quoteType = "SquareBracket"
			nameValue = nameValue[1 : len(nameValue)-1]
		}
		name := &ast.Identifier{Value: nameValue, QuoteType: quoteType}
		p.nextToken()

		// Build SchemaObjectName from identifiers (filter out empty identifiers from leading dots)
		var nonEmptyIdents []*ast.Identifier
		for _, id := range identifiers {
			if id.Value != "" {
				nonEmptyIdents = append(nonEmptyIdents, id)
			}
		}
		schemaObjName := identifiersToSchemaObjectName(nonEmptyIdents)

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
				return nil, fmt.Errorf("expected ) in function call with ::, got %s", p.curTok.Literal)
			}
			p.nextToken()

			// Check for OVER clause or property access after method call
			return p.parsePostExpressionAccess(fc)
		}

		// Property access: .t::a
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

	// Check if this is a function call
	if p.curTok.Type == TokenLParen && len(identifiers) > 1 {
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

func (p *Parser) parseFunctionCallFromIdentifiers(identifiers []*ast.Identifier) (ast.ScalarExpression, error) {
	// Check for special functions that need custom handling
	if len(identifiers) == 1 {
		funcName := strings.ToUpper(identifiers[0].Value)
		switch funcName {
		case "IIF":
			return p.parseIIfCall()
		case "PARSE":
			return p.parseParseCall(false)
		case "TRY_PARSE":
			return p.parseParseCall(true)
		}
	}

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
	funcNameUpper := strings.ToUpper(fc.FunctionName.Value)

	// Special handling for TRIM function with LEADING/TRAILING/BOTH options
	if funcNameUpper == "TRIM" && p.curTok.Type != TokenRParen {
		// Check for LEADING, TRAILING, or BOTH keyword
		trimOpt := strings.ToUpper(p.curTok.Literal)
		if trimOpt == "LEADING" || trimOpt == "TRAILING" || trimOpt == "BOTH" {
			fc.TrimOptions = &ast.Identifier{Value: trimOpt, QuoteType: "NotQuoted"}
			p.nextToken()
		}
	}

	if p.curTok.Type != TokenRParen {
		for {
			param, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			fc.Parameters = append(fc.Parameters, param)

			// Special handling for TRIM function: FROM keyword acts as separator
			if funcNameUpper == "TRIM" && strings.ToUpper(p.curTok.Literal) == "FROM" {
				p.nextToken() // consume FROM
				continue
			}

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

		// Check for WITHIN GROUP clause for function calls (e.g., PERCENTILE_CONT)
		if fc, ok := expr.(*ast.FunctionCall); ok && strings.ToUpper(p.curTok.Literal) == "WITHIN" {
			p.nextToken() // consume WITHIN
			if strings.ToUpper(p.curTok.Literal) == "GROUP" {
				p.nextToken() // consume GROUP
			}

			if p.curTok.Type != TokenLParen {
				return nil, fmt.Errorf("expected ( after WITHIN GROUP, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume (

			// Parse ORDER BY clause
			withinGroup := &ast.WithinGroupClause{
				HasGraphPath: false,
			}

			if p.curTok.Type == TokenOrder {
				orderBy, err := p.parseOrderByClause()
				if err != nil {
					return nil, err
				}
				withinGroup.OrderByClause = orderBy
			}

			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ) in WITHIN GROUP clause, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )

			fc.WithinGroupClause = withinGroup
			continue // continue to check for more clauses like OVER
		}

		// Check for RESPECT NULLS or IGNORE NULLS for window functions
		if fc, ok := expr.(*ast.FunctionCall); ok {
			upperLit := strings.ToUpper(p.curTok.Literal)
			if upperLit == "RESPECT" || upperLit == "IGNORE" {
				// Parse RESPECT NULLS or IGNORE NULLS
				firstIdent := &ast.Identifier{
					Value:     strings.ToUpper(p.curTok.Literal),
					QuoteType: "NotQuoted",
				}
				p.nextToken() // consume RESPECT/IGNORE

				if strings.ToUpper(p.curTok.Literal) == "NULLS" {
					secondIdent := &ast.Identifier{
						Value:     strings.ToUpper(p.curTok.Literal),
						QuoteType: "NotQuoted",
					}
					p.nextToken() // consume NULLS
					fc.IgnoreRespectNulls = []*ast.Identifier{firstIdent, secondIdent}
				}
				continue // continue to check for OVER clause
			}
		}

		// Check for OVER clause for function calls
		if fc, ok := expr.(*ast.FunctionCall); ok && strings.ToUpper(p.curTok.Literal) == "OVER" {
			overClause, err := p.parseOverClause()
			if err != nil {
				return nil, err
			}
			fc.OverClause = overClause
		}

		// Check for COLLATE clause for function calls
		if fc, ok := expr.(*ast.FunctionCall); ok && strings.ToUpper(p.curTok.Literal) == "COLLATE" {
			p.nextToken() // consume COLLATE
			fc.Collation = p.parseIdentifier()
			continue
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

	// Check for PIVOT or UNPIVOT
	if strings.ToUpper(p.curTok.Literal) == "PIVOT" {
		pivoted, err := p.parsePivotedTableReference(left)
		if err != nil {
			return nil, err
		}
		left = pivoted
	} else if strings.ToUpper(p.curTok.Literal) == "UNPIVOT" {
		unpivoted, err := p.parseUnpivotedTableReference(left)
		if err != nil {
			return nil, err
		}
		left = unpivoted
	}

	// Check for JOINs
	for {
		// Check for CROSS JOIN or CROSS APPLY
		if p.curTok.Type == TokenCross {
			p.nextToken() // consume CROSS
			if p.curTok.Type == TokenJoin {
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
			} else if strings.ToUpper(p.curTok.Literal) == "APPLY" {
				p.nextToken() // consume APPLY

				right, err := p.parseSingleTableReference()
				if err != nil {
					return nil, err
				}

				left = &ast.UnqualifiedJoin{
					UnqualifiedJoinType:  "CrossApply",
					FirstTableReference:  left,
					SecondTableReference: right,
				}
				continue
			} else {
				return nil, fmt.Errorf("expected JOIN or APPLY after CROSS, got %s", p.curTok.Literal)
			}
		}

		// Check for OUTER APPLY
		if p.curTok.Type == TokenOuter && strings.ToUpper(p.peekTok.Literal) == "APPLY" {
			p.nextToken() // consume OUTER
			p.nextToken() // consume APPLY

			right, err := p.parseSingleTableReference()
			if err != nil {
				return nil, err
			}

			left = &ast.UnqualifiedJoin{
				UnqualifiedJoinType:  "OuterApply",
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
	// Check for derived table (parenthesized query)
	if p.curTok.Type == TokenLParen {
		return p.parseDerivedTableReference()
	}

	// Check for OPENROWSET
	if p.curTok.Type == TokenOpenRowset {
		return p.parseOpenRowset()
	}

	// Check for PREDICT
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PREDICT" {
		return p.parsePredictTableReference()
	}

	// Check for CHANGETABLE
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "CHANGETABLE" {
		return p.parseChangeTableReference()
	}

	// Check for full-text table functions (CONTAINSTABLE, FREETEXTTABLE)
	if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper == "CONTAINSTABLE" || upper == "FREETEXTTABLE" {
			return p.parseFullTextTableReference(upper)
		}
		// Check for semantic table functions
		if upper == "SEMANTICKEYPHRASETABLE" || upper == "SEMANTICSIMILARITYTABLE" || upper == "SEMANTICSIMILARITYDETAILSTABLE" {
			return p.parseSemanticTableReference(upper)
		}
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

	// Check for table-valued function (identifier followed by parentheses that's not a table hint)
	// Parse schema object name first, then check if it's followed by function call parentheses
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}

	// Check for function call (has parentheses with non-hint content)
	if p.curTok.Type == TokenLParen && !p.peekIsTableHint() {
		params, err := p.parseFunctionParameters()
		if err != nil {
			return nil, err
		}

		// Parse optional alias (AS alias or just alias) and optional column list
		var alias *ast.Identifier
		var columns []*ast.Identifier
		if p.curTok.Type == TokenAs {
			p.nextToken()
			alias = p.parseIdentifier()
		} else if p.curTok.Type == TokenIdent {
			upper := strings.ToUpper(p.curTok.Literal)
			if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" &&
				upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
				upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
				upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
				alias = p.parseIdentifier()
			}
		}
		// Check for column list: alias(c1, c2, ...)
		if alias != nil && p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				columns = append(columns, p.parseIdentifier())
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

		// Use GlobalFunctionTableReference for specific built-in global functions
		if son.Count == 1 && son.BaseIdentifier != nil {
			upper := strings.ToUpper(son.BaseIdentifier.Value)
			if upper == "STRING_SPLIT" || upper == "GENERATE_SERIES" {
				return &ast.GlobalFunctionTableReference{
					Name:       son.BaseIdentifier,
					Parameters: params,
					Alias:      alias,
					Columns:    columns,
					ForPath:    false,
				}, nil
			}
		}

		ref := &ast.SchemaObjectFunctionTableReference{
			SchemaObject: son,
			Parameters:   params,
			Alias:        alias,
			Columns:      columns,
			ForPath:      false,
		}
		return ref, nil
	}

	// It's a regular named table reference
	return p.parseNamedTableReferenceWithName(son)
}

// parseDerivedTableReference parses a derived table (parenthesized query) like (SELECT ...) AS alias
// or an inline derived table (VALUES clause) like (VALUES (...), (...)) AS alias(cols)
// or a data modification table reference (DML with OUTPUT) like (INSERT ... OUTPUT ...) AS alias
func (p *Parser) parseDerivedTableReference() (ast.TableReference, error) {
	p.nextToken() // consume (

	// Check for VALUES clause (inline derived table)
	if strings.ToUpper(p.curTok.Literal) == "VALUES" {
		return p.parseInlineDerivedTable()
	}

	// Check for DML statements (INSERT, UPDATE, DELETE, MERGE) as table sources
	if p.curTok.Type == TokenInsert {
		return p.parseDataModificationTableReference("INSERT")
	}
	if p.curTok.Type == TokenUpdate {
		return p.parseDataModificationTableReference("UPDATE")
	}
	if p.curTok.Type == TokenDelete {
		return p.parseDataModificationTableReference("DELETE")
	}
	if strings.ToUpper(p.curTok.Literal) == "MERGE" {
		return p.parseDataModificationTableReference("MERGE")
	}

	// Parse the query expression
	qe, err := p.parseQueryExpression()
	if err != nil {
		return nil, err
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after derived table query, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	ref := &ast.QueryDerivedTable{
		QueryExpression: qe,
		ForPath:         false,
	}

	// Parse optional alias (AS alias or just alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
		// Could be an alias without AS, but need to be careful not to consume keywords
		if p.curTok.Type == TokenIdent {
			upper := strings.ToUpper(p.curTok.Literal)
			if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" && upper != "USING" && upper != "WHEN" && upper != "OUTPUT" && upper != "PIVOT" && upper != "UNPIVOT" {
				ref.Alias = p.parseIdentifier()
			}
		} else {
			ref.Alias = p.parseIdentifier()
		}
	}

	return ref, nil
}

// parseDataModificationTableReference parses a DML statement used as a table source
// This is called after ( is consumed and the DML keyword is the current token
func (p *Parser) parseDataModificationTableReference(dmlType string) (*ast.DataModificationTableReference, error) {
	ref := &ast.DataModificationTableReference{
		ForPath: false,
	}

	var err error
	switch dmlType {
	case "INSERT":
		spec, parseErr := p.parseInsertSpecification()
		if parseErr != nil {
			return nil, parseErr
		}
		ref.DataModificationSpecification = spec
	case "UPDATE":
		spec, parseErr := p.parseUpdateSpecification()
		if parseErr != nil {
			return nil, parseErr
		}
		ref.DataModificationSpecification = spec
	case "DELETE":
		spec, parseErr := p.parseDeleteSpecification()
		if parseErr != nil {
			return nil, parseErr
		}
		ref.DataModificationSpecification = spec
	case "MERGE":
		spec, parseErr := p.parseMergeSpecification()
		if parseErr != nil {
			return nil, parseErr
		}
		ref.DataModificationSpecification = spec
	default:
		return nil, fmt.Errorf("unknown DML type: %s", dmlType)
	}
	if err != nil {
		return nil, err
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after data modification statement, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse required alias (AS alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" &&
			upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
			upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
			upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
			ref.Alias = p.parseIdentifier()
		}
	}

	return ref, nil
}

// parseInlineDerivedTable parses a VALUES clause used as a table source
// Called after ( is consumed and VALUES is the current token
func (p *Parser) parseInlineDerivedTable() (*ast.InlineDerivedTable, error) {
	p.nextToken() // consume VALUES

	ref := &ast.InlineDerivedTable{
		ForPath: false,
	}

	// Parse row values: (val1, val2), (val3, val4), ...
	for {
		if p.curTok.Type != TokenLParen {
			break
		}
		p.nextToken() // consume (

		row := &ast.RowValue{}
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			row.ColumnValues = append(row.ColumnValues, expr)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
		ref.RowValues = append(ref.RowValues, row)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume , between rows
		} else {
			break
		}
	}

	// Expect ) to close the VALUES clause
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after VALUES clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse optional alias: AS alias or just alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" &&
			upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
			upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
			upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
			ref.Alias = p.parseIdentifier()
		}
	}

	// Parse optional column list: alias(col1, col2, ...)
	if ref.Alias != nil && p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			ref.Columns = append(ref.Columns, p.parseIdentifier())
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

	return ref, nil
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

	// T-SQL supports two syntaxes for table hints:
	// 1. Old-style: table_name (nolock) AS alias - hints before alias, no WITH
	// 2. New-style: table_name AS alias WITH (hints) - alias before hints, WITH required

	// Check for old-style hints (without WITH keyword): table (nolock) as alias
	if p.curTok.Type == TokenLParen && p.peekIsTableHint() {
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
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" && upper != "USING" && upper != "WHEN" && upper != "OUTPUT" && upper != "PIVOT" && upper != "UNPIVOT" {
			ref.Alias = &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
			p.nextToken()
		}
	}

	// Check for new-style hints (with WITH keyword): alias WITH (hints)
	if p.curTok.Type == TokenWith && p.peekTok.Type == TokenLParen {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen && p.peekIsTableHint() {
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
					if p.isTableHintToken() {
						continue
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

// parseNamedTableReferenceWithName parses a named table reference when the schema object name has already been parsed
func (p *Parser) parseNamedTableReferenceWithName(son *ast.SchemaObjectName) (*ast.NamedTableReference, error) {
	ref := &ast.NamedTableReference{
		SchemaObject: son,
		ForPath:      false,
	}

	// T-SQL supports two syntaxes for table hints:
	// 1. Old-style: table_name (nolock) AS alias - hints before alias, no WITH
	// 2. New-style: table_name AS alias WITH (hints) - alias before hints, WITH required

	// Check for old-style hints (without WITH keyword): table (nolock) as alias
	if p.curTok.Type == TokenLParen && p.peekIsTableHint() {
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
				if p.isTableHintToken() {
					continue
				}
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Parse optional alias (AS alias or just alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
		if p.curTok.Type != TokenIdent && p.curTok.Type != TokenLBracket {
			return nil, fmt.Errorf("expected identifier after AS, got %s", p.curTok.Literal)
		}
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
		// Could be an alias without AS, but need to be careful not to consume keywords
		if p.curTok.Type == TokenIdent {
			upper := strings.ToUpper(p.curTok.Literal)
			if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" && upper != "USING" && upper != "WHEN" && upper != "OUTPUT" && upper != "PIVOT" && upper != "UNPIVOT" {
				ref.Alias = p.parseIdentifier()
			}
		} else {
			ref.Alias = p.parseIdentifier()
		}
	}

	// Check for new-style hints (with WITH keyword): alias WITH (hints)
	if p.curTok.Type == TokenWith && p.peekTok.Type == TokenLParen {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen && p.peekIsTableHint() {
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
					if p.isTableHintToken() {
						continue
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

// parseFullTextTableReference parses CONTAINSTABLE or FREETEXTTABLE
func (p *Parser) parseFullTextTableReference(funcType string) (*ast.FullTextTableReference, error) {
	ref := &ast.FullTextTableReference{
		ForPath: false,
	}
	if funcType == "CONTAINSTABLE" {
		ref.FullTextFunctionType = "Contains"
	} else {
		ref.FullTextFunctionType = "FreeText"
	}
	p.nextToken() // consume function name

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after %s, got %s", funcType, p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse table name
	tableName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	ref.TableName = tableName

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after table name, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse column specification - could be *, (columns), or PROPERTY(column, 'property')
	if p.curTok.Type == TokenStar {
		ref.Columns = []*ast.ColumnReferenceExpression{{ColumnType: "Wildcard"}}
		p.nextToken()
	} else if p.curTok.Type == TokenLParen {
		// Column list
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			if p.curTok.Type == TokenStar {
				ref.Columns = append(ref.Columns, &ast.ColumnReferenceExpression{ColumnType: "Wildcard"})
				p.nextToken()
			} else {
				col := p.parseIdentifier()
				ref.Columns = append(ref.Columns, &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{col},
						Count:       1,
					},
				})
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
	} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PROPERTY" {
		// PROPERTY(column, 'property_name')
		p.nextToken() // consume PROPERTY
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after PROPERTY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		// Parse column name
		col := p.parseIdentifier()
		ref.Columns = []*ast.ColumnReferenceExpression{{
			ColumnType: "Regular",
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Identifiers: []*ast.Identifier{col},
				Count:       1,
			},
		}}

		// Expect comma
		if p.curTok.Type != TokenComma {
			return nil, fmt.Errorf("expected , after column in PROPERTY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ,

		// Parse property name (string literal)
		propExpr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		ref.PropertyName = propExpr

		// Expect )
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after PROPERTY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	} else {
		// Single column
		col := p.parseIdentifier()
		ref.Columns = []*ast.ColumnReferenceExpression{{
			ColumnType: "Regular",
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Identifiers: []*ast.Identifier{col},
				Count:       1,
			},
		}}
	}

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after columns, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse search condition (string literal or expression)
	searchCond, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}
	ref.SearchCondition = searchCond

	// Parse optional LANGUAGE and top_n - can come in any order
	for p.curTok.Type == TokenComma {
		p.nextToken() // consume ,

		if p.curTok.Type == TokenLanguage {
			p.nextToken() // consume LANGUAGE
			langExpr, err := p.parsePrimaryExpression()
			if err != nil {
				return nil, err
			}
			ref.Language = langExpr
		} else {
			// top_n value
			topExpr, err := p.parsePrimaryExpression()
			if err != nil {
				return nil, err
			}
			ref.TopN = topExpr
		}
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after CONTAINSTABLE/FREETEXTTABLE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse optional alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
			ref.Alias = p.parseIdentifier()
		}
	}

	return ref, nil
}

// parseSemanticTableReference parses SEMANTICKEYPHRASETABLE, SEMANTICSIMILARITYTABLE, or SEMANTICSIMILARITYDETAILSTABLE
func (p *Parser) parseSemanticTableReference(funcType string) (*ast.SemanticTableReference, error) {
	ref := &ast.SemanticTableReference{
		ForPath: false,
	}
	switch funcType {
	case "SEMANTICKEYPHRASETABLE":
		ref.SemanticFunctionType = "SemanticKeyPhraseTable"
	case "SEMANTICSIMILARITYTABLE":
		ref.SemanticFunctionType = "SemanticSimilarityTable"
	case "SEMANTICSIMILARITYDETAILSTABLE":
		ref.SemanticFunctionType = "SemanticSimilarityDetailsTable"
	}
	p.nextToken() // consume function name

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after %s, got %s", funcType, p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse table name
	tableName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	ref.TableName = tableName

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after table name, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse column specification - could be *, (columns), or single column
	if p.curTok.Type == TokenStar {
		ref.Columns = []*ast.ColumnReferenceExpression{{ColumnType: "Wildcard"}}
		p.nextToken()
	} else if p.curTok.Type == TokenLParen {
		// Column list
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			if p.curTok.Type == TokenStar {
				ref.Columns = append(ref.Columns, &ast.ColumnReferenceExpression{ColumnType: "Wildcard"})
				p.nextToken()
			} else {
				col := p.parseIdentifier()
				ref.Columns = append(ref.Columns, &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{col},
						Count:       1,
					},
				})
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
	} else {
		// Single column
		col := p.parseIdentifier()
		ref.Columns = []*ast.ColumnReferenceExpression{{
			ColumnType: "Regular",
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Identifiers: []*ast.Identifier{col},
				Count:       1,
			},
		}}
	}

	// For SEMANTICSIMILARITYTABLE and SEMANTICKEYPHRASETABLE: optional source_key
	// For SEMANTICSIMILARITYDETAILSTABLE: source_key, matched_column, matched_key
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		// Parse source_key expression
		sourceKey, err := p.parseSimpleExpression()
		if err != nil {
			return nil, err
		}
		ref.SourceKey = sourceKey

		// For SEMANTICSIMILARITYDETAILSTABLE, parse matched_column and matched_key
		if funcType == "SEMANTICSIMILARITYDETAILSTABLE" {
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
				// Parse matched_column
				col := p.parseIdentifier()
				ref.MatchedColumn = &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{col},
						Count:       1,
					},
				}

				if p.curTok.Type == TokenComma {
					p.nextToken() // consume ,
					// Parse matched_key expression
					matchedKey, err := p.parseSimpleExpression()
					if err != nil {
						return nil, err
					}
					ref.MatchedKey = matchedKey
				}
			}
		}
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after semantic table function, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse optional alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
			ref.Alias = p.parseIdentifier()
		}
	}

	return ref, nil
}

// parseSimpleExpression parses a simple expression (including unary minus for negative numbers)
func (p *Parser) parseSimpleExpression() (ast.ScalarExpression, error) {
	if p.curTok.Type == TokenMinus {
		p.nextToken() // consume -
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpression{
			UnaryExpressionType: "Negative",
			Expression:          expr,
		}, nil
	}
	return p.parsePrimaryExpression()
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

	// SPATIAL_WINDOW_MAX_CELLS hint with value
	if hintName == "SPATIAL_WINDOW_MAX_CELLS" {
		hint := &ast.LiteralTableHint{
			HintKind: "SpatialWindowMaxCells",
		}
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		if p.curTok.Type == TokenNumber {
			hint.Value = &ast.IntegerLiteral{
				LiteralType: "Integer",
				Value:       p.curTok.Literal,
			}
			p.nextToken()
		}
		return hint, nil
	}

	// FORCESEEK hint with optional index and column list
	if hintName == "FORCESEEK" {
		hint := &ast.ForceSeekTableHint{
			HintKind: "ForceSeek",
		}
		// Check for optional parenthesis with index and columns
		if p.curTok.Type != TokenLParen {
			return hint, nil
		}
		p.nextToken() // consume (
		// Parse index value (identifier or number)
		if p.curTok.Type == TokenNumber {
			hint.IndexValue = &ast.IdentifierOrValueExpression{
				Value: p.curTok.Literal,
				ValueExpression: &ast.IntegerLiteral{
					LiteralType: "Integer",
					Value:       p.curTok.Literal,
				},
			}
			p.nextToken()
		} else if p.curTok.Type == TokenIdent {
			hint.IndexValue = &ast.IdentifierOrValueExpression{
				Value: p.curTok.Literal,
				Identifier: &ast.Identifier{
					Value:     p.curTok.Literal,
					QuoteType: "NotQuoted",
				},
			}
			p.nextToken()
		}
		// Parse optional column list
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col, _ := p.parseColumnReference()
				if col != nil {
					hint.ColumnValues = append(hint.ColumnValues, col)
				}
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else if p.curTok.Type != TokenRParen {
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
		// Consume outer )
		if p.curTok.Type == TokenRParen {
			p.nextToken()
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
	case "FORCESEEK":
		return "ForceSeek"
	case "FORCESCAN":
		return "ForceScan"
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
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "HINT" {
			p.nextToken() // consume HINT
			return p.parseUseHintList()
		}
		return &ast.OptimizerHint{HintKind: "Use"}, nil
	}

	// Handle keyword tokens that can be optimizer hints (ORDER, GROUP, MAXDOP, etc.)
	if p.curTok.Type == TokenOrder || p.curTok.Type == TokenGroup {
		hintKind := convertHintKind(p.curTok.Literal)
		firstWord := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		// Check for two-word hints like ORDER GROUP
		if (firstWord == "ORDER" || firstWord == "HASH" || firstWord == "MERGE" ||
			firstWord == "CONCAT" || firstWord == "LOOP" || firstWord == "FORCE") &&
			isSecondHintWordToken(p.curTok.Type) {
			secondWord := strings.ToUpper(p.curTok.Literal)
			if secondWord == "GROUP" || secondWord == "JOIN" || secondWord == "UNION" ||
				secondWord == "ORDER" {
				hintKind = hintKind + convertHintKind(p.curTok.Literal)
				p.nextToken()
			}
		}
		return &ast.OptimizerHint{HintKind: hintKind}, nil
	}

	// Handle MAXDOP keyword
	if p.curTok.Type == TokenMaxdop {
		p.nextToken() // consume MAXDOP
		// MAXDOP takes a numeric argument
		if p.curTok.Type == TokenNumber {
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: "MaxDop", Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: "MaxDop"}, nil
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

	// Handle FAST keyword
	if p.curTok.Type == TokenFast {
		p.nextToken() // consume FAST
		// FAST takes a numeric argument
		if p.curTok.Type == TokenNumber {
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: "Fast", Value: value}, nil
		}
		return &ast.OptimizerHint{HintKind: "Fast"}, nil
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

	case "NO_PERFORMANCE_SPOOL":
		p.nextToken() // consume NO_PERFORMANCE_SPOOL
		return &ast.OptimizerHint{HintKind: "NoPerformanceSpool"}, nil

	default:
		// Handle generic hints
		hintKind := convertHintKind(p.curTok.Literal)
		firstWord := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		// Check for two-word hints like ORDER GROUP, HASH GROUP, etc.
		if (firstWord == "ORDER" || firstWord == "HASH" || firstWord == "MERGE" ||
			firstWord == "CONCAT" || firstWord == "LOOP" || firstWord == "FORCE" ||
			firstWord == "KEEP" || firstWord == "ROBUST" || firstWord == "EXPAND" ||
			firstWord == "KEEPFIXED" || firstWord == "SHRINKDB" || firstWord == "ALTERCOLUMN" ||
			firstWord == "BYPASS") &&
			isSecondHintWordToken(p.curTok.Type) {
			secondWord := strings.ToUpper(p.curTok.Literal)
			if secondWord == "GROUP" || secondWord == "JOIN" || secondWord == "UNION" ||
				secondWord == "ORDER" || secondWord == "PLAN" || secondWord == "VIEWS" ||
				secondWord == "OPTIMIZER_QUEUE" {
				hintKind = hintKind + convertHintKind(p.curTok.Literal)
				p.nextToken()
			}
		}

		// Check if this is a literal hint with value (USEPLAN 2, etc.)
		if p.curTok.Type == TokenNumber {
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			return &ast.LiteralOptimizerHint{HintKind: hintKind, Value: value}, nil
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

func (p *Parser) parseUseHintList() (ast.OptimizerHintBase, error) {
	hint := &ast.UseHintList{
		HintKind: "Unspecified",
	}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after USE HINT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse hint string literals
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}

		if p.curTok.Type == TokenString {
			str := p.parseStringLiteralValue()
			p.nextToken()
			hint.Hints = append(hint.Hints, str)
		} else if p.curTok.Type == TokenNationalString {
			str, _ := p.parseNationalStringFromToken()
			hint.Hints = append(hint.Hints, str)
		} else {
			break
		}
	}

	// Expect )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return hint, nil
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
		"BYPASS":                       "Bypass",
		"OPTIMIZER_QUEUE":              "OptimizerQueue",
		"USEPLAN":                      "UsePlan",
		"SHRINKDB":                     "ShrinkDB",
		"ALTERCOLUMN":                  "AlterColumn",
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

// isSecondHintWordToken checks if a token can be a second word in a two-word optimizer hint
func isSecondHintWordToken(t TokenType) bool {
	return t == TokenIdent || t == TokenGroup || t == TokenJoin || t == TokenUnion || t == TokenOrder
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
		spec, err := p.parseGroupingSpecification()
		if err != nil {
			return nil, err
		}
		gbc.GroupingSpecifications = append(gbc.GroupingSpecifications, spec)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// Check for WITH ROLLUP or WITH CUBE (old syntax)
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

// parseGroupingSpecification parses a single grouping specification
func (p *Parser) parseGroupingSpecification() (ast.GroupingSpecification, error) {
	// Check for ROLLUP (...)
	if p.curTok.Type == TokenRollup {
		return p.parseRollupGroupingSpecification()
	}

	// Check for CUBE (...)
	if p.curTok.Type == TokenCube {
		return p.parseCubeGroupingSpecification()
	}

	// Check for GROUPING SETS (...)
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GROUPING" &&
		p.peekTok.Type == TokenIdent && strings.ToUpper(p.peekTok.Literal) == "SETS" {
		return p.parseGroupingSetsGroupingSpecification()
	}

	// Check for grand total () or composite grouping (c1, c2, ...)
	if p.curTok.Type == TokenLParen {
		// Check for empty parens () which is grand total
		if p.peekTok.Type == TokenRParen {
			p.nextToken() // consume (
			p.nextToken() // consume )
			return &ast.GrandTotalGroupingSpecification{}, nil
		}
		return p.parseCompositeGroupingSpecification()
	}

	// Regular expression grouping
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	spec := &ast.ExpressionGroupingSpecification{
		Expression:             expr,
		DistributedAggregation: false,
	}

	// Check for WITH (DISTRIBUTED_AGG) hint - only if next token is (
	// This distinguishes from WITH ROLLUP/CUBE at the end
	if p.curTok.Type == TokenWith && p.peekTok.Type == TokenLParen {
		p.nextToken() // consume WITH
		p.nextToken() // consume (
		if strings.ToUpper(p.curTok.Literal) == "DISTRIBUTED_AGG" {
			spec.DistributedAggregation = true
			p.nextToken() // consume DISTRIBUTED_AGG
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	return spec, nil
}

// parseRollupGroupingSpecification parses ROLLUP (c1, c2, ...)
func (p *Parser) parseRollupGroupingSpecification() (*ast.RollupGroupingSpecification, error) {
	p.nextToken() // consume ROLLUP

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after ROLLUP, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	spec := &ast.RollupGroupingSpecification{}

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		arg, err := p.parseGroupingSpecificationArgument()
		if err != nil {
			return nil, err
		}
		spec.Arguments = append(spec.Arguments, arg)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return spec, nil
}

// parseCubeGroupingSpecification parses CUBE (c1, c2, ...)
func (p *Parser) parseCubeGroupingSpecification() (*ast.CubeGroupingSpecification, error) {
	p.nextToken() // consume CUBE

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after CUBE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	spec := &ast.CubeGroupingSpecification{}

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		arg, err := p.parseGroupingSpecificationArgument()
		if err != nil {
			return nil, err
		}
		spec.Arguments = append(spec.Arguments, arg)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return spec, nil
}

// parseGroupingSetsGroupingSpecification parses GROUPING SETS (...)
func (p *Parser) parseGroupingSetsGroupingSpecification() (*ast.GroupingSetsGroupingSpecification, error) {
	p.nextToken() // consume GROUPING
	p.nextToken() // consume SETS

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after GROUPING SETS, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	spec := &ast.GroupingSetsGroupingSpecification{}

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		arg, err := p.parseGroupingSetsArgument()
		if err != nil {
			return nil, err
		}
		spec.Arguments = append(spec.Arguments, arg)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return spec, nil
}

// parseGroupingSetsArgument parses an argument inside GROUPING SETS which can be
// CUBE(...), ROLLUP(...), a column, or a parenthesized group
func (p *Parser) parseGroupingSetsArgument() (ast.GroupingSpecification, error) {
	// Check for CUBE
	if p.curTok.Type == TokenCube {
		return p.parseCubeGroupingSpecification()
	}

	// Check for ROLLUP
	if p.curTok.Type == TokenRollup {
		return p.parseRollupGroupingSpecification()
	}

	// Check for parenthesized group
	if p.curTok.Type == TokenLParen {
		// Check for empty parens () which is grand total
		if p.peekTok.Type == TokenRParen {
			p.nextToken() // consume (
			p.nextToken() // consume )
			return &ast.GrandTotalGroupingSpecification{}, nil
		}
		return p.parseGroupingSetsCompositeArgument()
	}

	// Regular expression (column reference or literal)
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionGroupingSpecification{
		Expression:             expr,
		DistributedAggregation: false,
	}, nil
}

// parseGroupingSetsCompositeArgument parses a parenthesized group inside GROUPING SETS
// which can contain CUBE, ROLLUP, columns, or a mix
func (p *Parser) parseGroupingSetsCompositeArgument() (ast.GroupingSpecification, error) {
	p.nextToken() // consume (

	// Check what's inside - might be CUBE, ROLLUP, or columns
	var items []ast.GroupingSpecification

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		var item ast.GroupingSpecification
		var err error

		if p.curTok.Type == TokenCube {
			item, err = p.parseCubeGroupingSpecification()
		} else if p.curTok.Type == TokenRollup {
			item, err = p.parseRollupGroupingSpecification()
		} else if p.curTok.Type == TokenLParen {
			// Check for empty parens () which is grand total
			if p.peekTok.Type == TokenRParen {
				p.nextToken() // consume (
				p.nextToken() // consume )
				item = &ast.GrandTotalGroupingSpecification{}
			} else {
				item, err = p.parseGroupingSetsCompositeArgument()
			}
		} else {
			// Expression
			expr, e := p.parseScalarExpression()
			if e != nil {
				return nil, e
			}
			item = &ast.ExpressionGroupingSpecification{
				Expression:             expr,
				DistributedAggregation: false,
			}
		}

		if err != nil {
			return nil, err
		}
		items = append(items, item)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return &ast.CompositeGroupingSpecification{Items: items}, nil
}

// parseGroupingSpecificationArgument parses an argument inside ROLLUP/CUBE which can be
// an expression or a composite grouping like (c2, c3)
func (p *Parser) parseGroupingSpecificationArgument() (ast.GroupingSpecification, error) {
	// Check for composite grouping (c1, c2)
	if p.curTok.Type == TokenLParen {
		return p.parseCompositeGroupingSpecification()
	}

	// Regular expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return &ast.ExpressionGroupingSpecification{
		Expression:             expr,
		DistributedAggregation: false,
	}, nil
}

// parseCompositeGroupingSpecification parses (c1, c2, ...)
func (p *Parser) parseCompositeGroupingSpecification() (*ast.CompositeGroupingSpecification, error) {
	p.nextToken() // consume (

	spec := &ast.CompositeGroupingSpecification{}

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		spec.Items = append(spec.Items, &ast.ExpressionGroupingSpecification{
			Expression:             expr,
			DistributedAggregation: false,
		})

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return spec, nil
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

// parseOffsetClause parses OFFSET n ROWS FETCH NEXT/FIRST m ROWS ONLY
func (p *Parser) parseOffsetClause() (*ast.OffsetClause, error) {
	// Consume OFFSET
	p.nextToken()

	oc := &ast.OffsetClause{}

	// Parse offset expression
	offsetExpr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	oc.OffsetExpression = offsetExpr

	// Skip ROWS/ROW keyword
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "ROWS" || upperLit == "ROW" {
		p.nextToken()
	}

	// Parse FETCH NEXT/FIRST m ROWS ONLY
	if strings.ToUpper(p.curTok.Literal) == "FETCH" {
		p.nextToken() // consume FETCH

		// Skip NEXT or FIRST
		upperLit = strings.ToUpper(p.curTok.Literal)
		if upperLit == "NEXT" || upperLit == "FIRST" {
			p.nextToken()
		}

		// Parse fetch expression
		fetchExpr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		oc.FetchExpression = fetchExpr

		// Skip ROWS/ROW keyword
		upperLit = strings.ToUpper(p.curTok.Literal)
		if upperLit == "ROWS" || upperLit == "ROW" {
			p.nextToken()
		}

		// Skip ONLY keyword
		if strings.ToUpper(p.curTok.Literal) == "ONLY" {
			p.nextToken()
		}
	}

	return oc, nil
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
	// Check for CONTAINS/FREETEXT predicates
	if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper == "CONTAINS" || upper == "FREETEXT" {
			return p.parseFullTextPredicate(upper)
		}
		if upper == "EXISTS" {
			return p.parseExistsPredicate()
		}
	}

	// Check for parenthesized expression - could be boolean or scalar subquery
	if p.curTok.Type == TokenLParen {
		// Peek ahead to see if it's a subquery (SELECT)
		if p.peekTok.Type == TokenSelect {
			// Parse as scalar subquery that will be used in a comparison
			p.nextToken() // consume (
			qe, err := p.parseQueryExpression()
			if err != nil {
				return nil, err
			}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )

			subquery := &ast.ScalarSubquery{QueryExpression: qe}

			// Now check for comparison operators
			if p.isComparisonOperator() {
				return p.parseComparisonAfterLeft(subquery)
			}
			// If no comparison, this might be used in other contexts
			// For now, treat it as an error if used standalone
			return nil, fmt.Errorf("scalar subquery must be followed by a comparison operator")
		}

		// Parse as parenthesized boolean expression
		p.nextToken() // consume (

		// Parse inner boolean expression
		inner, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}

		// Check if we got a placeholder for a scalar expression without comparison
		// This happens when parsing something like (XACT_STATE()) in: IF (XACT_STATE()) = -1
		if placeholder, ok := inner.(*ast.BooleanScalarPlaceholder); ok {
			// The inner content was a bare scalar expression
			// curTok should still be ) since we didn't consume it
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )

			// Wrap the scalar in a ParenthesisExpression
			parenExpr := &ast.ParenthesisExpression{Expression: placeholder.Scalar}

			// Check for comparison operators after the parenthesized expression
			if p.isComparisonOperator() {
				return p.parseComparisonAfterLeft(parenExpr)
			}

			// Check for IS NULL / IS NOT NULL
			if p.curTok.Type == TokenIs {
				return p.parseIsNullAfterLeft(parenExpr)
			}

			// Check for NOT before IN/LIKE/BETWEEN
			notDefined := false
			if p.curTok.Type == TokenNot {
				notDefined = true
				p.nextToken()
			}

			if p.curTok.Type == TokenIn {
				return p.parseInExpressionAfterLeft(parenExpr, notDefined)
			}
			if p.curTok.Type == TokenLike {
				return p.parseLikeExpressionAfterLeft(parenExpr, notDefined)
			}
			if p.curTok.Type == TokenBetween {
				return p.parseBetweenExpressionAfterLeft(parenExpr, notDefined)
			}

			if notDefined {
				return nil, fmt.Errorf("expected IN, LIKE, or BETWEEN after NOT, got %s", p.curTok.Literal)
			}

			// If no comparison follows, return error
			return nil, fmt.Errorf("expected comparison operator after parenthesized expression, got %s", p.curTok.Literal)
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

	// Check for IS NULL / IS NOT NULL / IS [NOT] DISTINCT FROM
	if p.curTok.Type == TokenIs {
		p.nextToken() // consume IS

		isNot := false
		if p.curTok.Type == TokenNot {
			isNot = true
			p.nextToken() // consume NOT
		}

		// Check for DISTINCT FROM
		if p.curTok.Type == TokenDistinct {
			p.nextToken() // consume DISTINCT
			if strings.ToUpper(p.curTok.Literal) != "FROM" {
				return nil, fmt.Errorf("expected FROM after DISTINCT, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume FROM

			// Special case: IS [NOT] DISTINCT FROM NULL becomes IS [NOT] NULL
			if p.curTok.Type == TokenNull {
				p.nextToken() // consume NULL
				// IS NOT DISTINCT FROM NULL = IS NULL (IsNot: false)
				// IS DISTINCT FROM NULL = IS NOT NULL (IsNot: true)
				return &ast.BooleanIsNullExpression{
					IsNot:      !isNot,
					Expression: left,
				}, nil
			}

			// Check for SOME/ANY/ALL (subquery)
			upperLit := strings.ToUpper(p.curTok.Literal)
			if upperLit == "SOME" || upperLit == "ANY" || upperLit == "ALL" {
				predicateType := "Any"
				if upperLit == "ALL" {
					predicateType = "All"
				}
				p.nextToken() // consume SOME/ANY/ALL

				if p.curTok.Type != TokenLParen {
					return nil, fmt.Errorf("expected ( after %s, got %s", upperLit, p.curTok.Literal)
				}
				p.nextToken() // consume (

				subqueryExpr, err := p.parseQueryExpression()
				if err != nil {
					return nil, err
				}

				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
				}
				p.nextToken() // consume )

				compType := "IsDistinctFrom"
				if isNot {
					compType = "IsNotDistinctFrom"
				}

				return &ast.SubqueryComparisonPredicate{
					Expression:                      left,
					ComparisonType:                  compType,
					Subquery:                        &ast.ScalarSubquery{QueryExpression: subqueryExpr},
					SubqueryComparisonPredicateType: predicateType,
				}, nil
			}

			// Parse the second expression
			secondExpr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			return &ast.DistinctPredicate{
				FirstExpression:  left,
				SecondExpression: secondExpr,
				IsNot:            isNot,
			}, nil
		}

		if p.curTok.Type != TokenNull {
			return nil, fmt.Errorf("expected NULL or DISTINCT after IS/IS NOT, got %s", p.curTok.Literal)
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
	case TokenRParen:
		// We're at ) without a comparison operator - this happens when parsing
		// a parenthesized scalar expression like (XACT_STATE()) in a boolean context.
		// Return a special marker that the caller can handle.
		return &ast.BooleanScalarPlaceholder{Scalar: left}, nil
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

// isComparisonOperator checks if the current token is a comparison operator
func (p *Parser) isComparisonOperator() bool {
	switch p.curTok.Type {
	case TokenEquals, TokenNotEqual, TokenLessThan, TokenGreaterThan,
		TokenLessOrEqual, TokenGreaterOrEqual:
		return true
	default:
		return false
	}
}

// parseComparisonAfterLeft parses a comparison expression after the left operand is already parsed
func (p *Parser) parseComparisonAfterLeft(left ast.ScalarExpression) (ast.BooleanExpression, error) {
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

// parseInExpressionAfterLeft parses an IN expression after the left operand is already parsed
func (p *Parser) parseInExpressionAfterLeft(left ast.ScalarExpression, notDefined bool) (ast.BooleanExpression, error) {
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

// parseLikeExpressionAfterLeft parses a LIKE expression after the left operand is already parsed
func (p *Parser) parseLikeExpressionAfterLeft(left ast.ScalarExpression, notDefined bool) (ast.BooleanExpression, error) {
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

// parseBetweenExpressionAfterLeft parses a BETWEEN expression after the left operand is already parsed
func (p *Parser) parseBetweenExpressionAfterLeft(left ast.ScalarExpression, notDefined bool) (ast.BooleanExpression, error) {
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

// finishParenthesizedBooleanExpression finishes parsing a parenthesized boolean expression
// after the initial comparison/expression has been parsed
func (p *Parser) finishParenthesizedBooleanExpression(inner ast.BooleanExpression) (ast.BooleanExpression, error) {
	// Check for AND/OR continuation
	for p.curTok.Type == TokenAnd || p.curTok.Type == TokenOr {
		op := p.curTok.Type
		p.nextToken()

		right, err := p.parseBooleanPrimaryExpression()
		if err != nil {
			return nil, err
		}

		if op == TokenAnd {
			inner = &ast.BooleanBinaryExpression{
				BinaryExpressionType: "And",
				FirstExpression:      inner,
				SecondExpression:     right,
			}
		} else {
			inner = &ast.BooleanBinaryExpression{
				BinaryExpressionType: "Or",
				FirstExpression:      inner,
				SecondExpression:     right,
			}
		}
	}

	// Expect closing parenthesis
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return &ast.BooleanParenthesisExpression{Expression: inner}, nil
}

// parseIsNullAfterLeft parses IS NULL / IS NOT NULL / IS [NOT] DISTINCT FROM after the left operand is already parsed
func (p *Parser) parseIsNullAfterLeft(left ast.ScalarExpression) (ast.BooleanExpression, error) {
	p.nextToken() // consume IS

	isNot := false
	if p.curTok.Type == TokenNot {
		isNot = true
		p.nextToken() // consume NOT
	}

	// Check for DISTINCT FROM
	if p.curTok.Type == TokenDistinct {
		p.nextToken() // consume DISTINCT
		if strings.ToUpper(p.curTok.Literal) != "FROM" {
			return nil, fmt.Errorf("expected FROM after DISTINCT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume FROM

		// Special case: IS [NOT] DISTINCT FROM NULL becomes IS [NOT] NULL
		if p.curTok.Type == TokenNull {
			p.nextToken() // consume NULL
			// IS NOT DISTINCT FROM NULL = IS NULL (IsNot: false)
			// IS DISTINCT FROM NULL = IS NOT NULL (IsNot: true)
			return &ast.BooleanIsNullExpression{
				IsNot:      !isNot,
				Expression: left,
			}, nil
		}

		// Parse the second expression
		secondExpr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		return &ast.DistinctPredicate{
			FirstExpression:  left,
			SecondExpression: secondExpr,
			IsNot:            isNot,
		}, nil
	}

	if p.curTok.Type != TokenNull {
		return nil, fmt.Errorf("expected NULL or DISTINCT after IS/IS NOT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume NULL

	return &ast.BooleanIsNullExpression{
		IsNot:      isNot,
		Expression: left,
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

// parseIdentityFunctionCall parses an IDENTITY function call: IDENTITY(data_type [, seed, increment])
func (p *Parser) parseIdentityFunctionCall() (ast.ScalarExpression, error) {
	p.nextToken() // consume IDENTITY
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after IDENTITY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse the data type
	dt, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}

	identity := &ast.IdentityFunctionCall{
		DataType: dt,
	}

	// Check for optional seed and increment
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		seed, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		identity.Seed = seed

		// Expect comma before increment
		if p.curTok.Type != TokenComma {
			return nil, fmt.Errorf("expected , before increment in IDENTITY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ,

		increment, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		identity.Increment = increment
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in IDENTITY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return identity, nil
}

// parsePredictTableReference parses PREDICT(...) in FROM clause
// PREDICT(MODEL = expression, DATA = table AS alias, RUNTIME=ident) WITH (columns) AS alias
func (p *Parser) parsePredictTableReference() (*ast.PredictTableReference, error) {
	p.nextToken() // consume PREDICT

	ref := &ast.PredictTableReference{
		ForPath: false,
	}

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after PREDICT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse arguments: MODEL = expr, DATA = table AS alias, RUNTIME = ident
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		argName := strings.ToUpper(p.curTok.Literal)
		p.nextToken() // consume argument name

		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		switch argName {
		case "MODEL":
			// MODEL can be a subquery or variable
			if p.curTok.Type == TokenLParen {
				// Subquery
				p.nextToken() // consume (
				qe, err := p.parseQueryExpression()
				if err != nil {
					return nil, err
				}
				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
				}
				p.nextToken() // consume )
				ref.ModelVariable = &ast.ScalarSubquery{QueryExpression: qe}
			} else if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				// Variable
				ref.ModelVariable = &ast.VariableReference{Name: p.curTok.Literal}
				p.nextToken()
			}
		case "DATA":
			// DATA = table AS alias
			son, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, err
			}
			dataSource := &ast.NamedTableReference{
				SchemaObject: son,
				ForPath:      false,
			}
			// Check for AS alias
			if p.curTok.Type == TokenAs {
				p.nextToken()
				dataSource.Alias = p.parseIdentifier()
			}
			ref.DataSource = dataSource
		case "RUNTIME":
			ref.RunTime = p.parseIdentifier()
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	// Parse optional WITH clause for output schema
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (

			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				item := &ast.SchemaDeclarationItem{
					ColumnDefinition: &ast.ColumnDefinitionBase{},
				}
				item.ColumnDefinition.ColumnIdentifier = p.parseIdentifier()

				// Parse data type
				dataType, err := p.parseDataTypeReference()
				if err != nil {
					return nil, err
				}
				item.ColumnDefinition.DataType = dataType

				ref.SchemaDeclarationItems = append(ref.SchemaDeclarationItems, item)

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

	// Parse optional AS alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	}

	return ref, nil
}

// parseForClause parses FOR BROWSE, FOR XML, FOR UPDATE, FOR READ ONLY clauses.
func (p *Parser) parseForClause() (ast.ForClause, error) {
	p.nextToken() // consume FOR

	keyword := strings.ToUpper(p.curTok.Literal)

	switch keyword {
	case "BROWSE":
		p.nextToken() // consume BROWSE
		return &ast.BrowseForClause{}, nil

	case "READ":
		p.nextToken() // consume READ
		if strings.ToUpper(p.curTok.Literal) == "ONLY" {
			p.nextToken() // consume ONLY
		}
		return &ast.ReadOnlyForClause{}, nil

	case "UPDATE":
		p.nextToken() // consume UPDATE
		clause := &ast.UpdateForClause{}

		// Check for OF column_list
		if strings.ToUpper(p.curTok.Literal) == "OF" {
			p.nextToken() // consume OF

			// Parse column list
			for {
				col, err := p.parseColumnReference()
				if err != nil {
					return nil, err
				}
				clause.Columns = append(clause.Columns, col)

				if p.curTok.Type != TokenComma {
					break
				}
				p.nextToken() // consume comma
			}
		}
		return clause, nil

	case "XML":
		p.nextToken() // consume XML
		return p.parseXmlForClause()

	case "JSON":
		p.nextToken() // consume JSON
		return p.parseJsonForClause()

	default:
		return nil, fmt.Errorf("unexpected token after FOR: %s", p.curTok.Literal)
	}
}

// parseXmlForClause parses FOR XML options.
func (p *Parser) parseXmlForClause() (*ast.XmlForClause, error) {
	clause := &ast.XmlForClause{}

	// Parse XML options separated by commas
	for {
		option, err := p.parseXmlForClauseOption()
		if err != nil {
			return nil, err
		}
		clause.Options = append(clause.Options, option)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return clause, nil
}

// parseXmlForClauseOption parses a single XML FOR clause option.
func (p *Parser) parseXmlForClauseOption() (*ast.XmlForClauseOption, error) {
	option := &ast.XmlForClauseOption{}

	keyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume the option keyword

	switch keyword {
	case "AUTO":
		option.OptionKind = "Auto"
	case "EXPLICIT":
		option.OptionKind = "Explicit"
	case "RAW":
		option.OptionKind = "Raw"
		// Check for optional element name: RAW ('name')
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if p.curTok.Type == TokenString {
				option.Value = p.parseStringLiteralValue()
				p.nextToken() // consume string
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	case "PATH":
		option.OptionKind = "Path"
		// Check for optional path name: PATH ('name')
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if p.curTok.Type == TokenString {
				option.Value = p.parseStringLiteralValue()
				p.nextToken() // consume string
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	case "ELEMENTS":
		// Check for XSINIL or ABSENT
		nextKeyword := strings.ToUpper(p.curTok.Literal)
		if nextKeyword == "XSINIL" {
			option.OptionKind = "ElementsXsiNil"
			p.nextToken() // consume XSINIL
		} else if nextKeyword == "ABSENT" {
			option.OptionKind = "ElementsAbsent"
			p.nextToken() // consume ABSENT
		} else {
			option.OptionKind = "Elements"
		}
	case "XMLDATA":
		option.OptionKind = "XmlData"
	case "XMLSCHEMA":
		option.OptionKind = "XmlSchema"
		// Check for optional namespace: XMLSCHEMA ('namespace')
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if p.curTok.Type == TokenString {
				option.Value = p.parseStringLiteralValue()
				p.nextToken() // consume string
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	case "ROOT":
		option.OptionKind = "Root"
		// Check for optional root name: ROOT ('name')
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if p.curTok.Type == TokenString {
				option.Value = p.parseStringLiteralValue()
				p.nextToken() // consume string
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	case "TYPE":
		option.OptionKind = "Type"
	case "BINARY":
		// BINARY BASE64
		if strings.ToUpper(p.curTok.Literal) == "BASE64" {
			option.OptionKind = "BinaryBase64"
			p.nextToken() // consume BASE64
		}
	default:
		option.OptionKind = keyword
	}

	return option, nil
}

// parseJsonForClause parses FOR JSON options.
func (p *Parser) parseJsonForClause() (*ast.JsonForClause, error) {
	clause := &ast.JsonForClause{}

	// Parse JSON options separated by commas
	for {
		option, err := p.parseJsonForClauseOption()
		if err != nil {
			return nil, err
		}
		clause.Options = append(clause.Options, option)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return clause, nil
}

// parseJsonForClauseOption parses a single JSON FOR clause option.
func (p *Parser) parseJsonForClauseOption() (*ast.JsonForClauseOption, error) {
	option := &ast.JsonForClauseOption{}

	keyword := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume the option keyword

	switch keyword {
	case "AUTO":
		option.OptionKind = "Auto"
	case "PATH":
		option.OptionKind = "Path"
	case "ROOT":
		option.OptionKind = "Root"
		// Check for optional root name: ROOT('name')
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if p.curTok.Type == TokenString {
				option.Value = p.parseStringLiteralValue()
				p.nextToken() // consume string
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	case "INCLUDE_NULL_VALUES":
		option.OptionKind = "IncludeNullValues"
	case "WITHOUT_ARRAY_WRAPPER":
		option.OptionKind = "WithoutArrayWrapper"
	default:
		option.OptionKind = keyword
	}

	return option, nil
}

// parseFullTextPredicate parses CONTAINS or FREETEXT predicates
func (p *Parser) parseFullTextPredicate(funcType string) (*ast.FullTextPredicate, error) {
	// Convert to PascalCase: "CONTAINS" -> "Contains", "FREETEXT" -> "FreeText"
	pascalType := strings.ToUpper(funcType[:1]) + strings.ToLower(funcType[1:])
	if funcType == "FREETEXT" {
		pascalType = "FreeText"
	}
	pred := &ast.FullTextPredicate{
		FullTextFunctionType: pascalType,
	}
	p.nextToken() // consume CONTAINS/FREETEXT

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after %s, got %s", funcType, p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse column specification: *, column, (columns), or PROPERTY(column, 'prop')
	if p.curTok.Type == TokenStar {
		pred.Columns = []*ast.ColumnReferenceExpression{{ColumnType: "Wildcard"}}
		p.nextToken() // consume *
	} else if p.curTok.Type == TokenLParen {
		// Column list
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			if p.curTok.Type == TokenStar {
				pred.Columns = append(pred.Columns, &ast.ColumnReferenceExpression{ColumnType: "Wildcard"})
				p.nextToken()
			} else {
				col := p.parseIdentifier()
				pred.Columns = append(pred.Columns, &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{col},
						Count:       1,
					},
				})
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
	} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PROPERTY" {
		// PROPERTY(column, 'property_name')
		p.nextToken() // consume PROPERTY
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after PROPERTY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		// Parse column name
		col := p.parseIdentifier()
		pred.Columns = []*ast.ColumnReferenceExpression{{
			ColumnType: "Regular",
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Identifiers: []*ast.Identifier{col},
				Count:       1,
			},
		}}

		// Expect comma
		if p.curTok.Type != TokenComma {
			return nil, fmt.Errorf("expected , after column in PROPERTY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ,

		// Parse property name (string literal)
		propExpr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		pred.PropertyName = propExpr

		// Expect )
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after PROPERTY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	} else {
		// Single column
		col := p.parseIdentifier()
		pred.Columns = []*ast.ColumnReferenceExpression{{
			ColumnType: "Regular",
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Identifiers: []*ast.Identifier{col},
				Count:       1,
			},
		}}
	}

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after columns in %s, got %s", funcType, p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse search value
	value, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}
	pred.Value = value

	// Parse optional LANGUAGE term
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		if p.curTok.Type == TokenLanguage {
			p.nextToken() // consume LANGUAGE
			langTerm, err := p.parsePrimaryExpression()
			if err != nil {
				return nil, err
			}
			pred.LanguageTerm = langTerm
		}
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after %s, got %s", funcType, p.curTok.Literal)
	}
	p.nextToken() // consume )

	return pred, nil
}

// parseExistsPredicate parses EXISTS (subquery)
func (p *Parser) parseExistsPredicate() (*ast.ExistsPredicate, error) {
	p.nextToken() // consume EXISTS

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after EXISTS, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse subquery
	subquery, err := p.parseQueryExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after EXISTS subquery, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return &ast.ExistsPredicate{Subquery: subquery}, nil
}

// parseIIfCall parses IIF(condition, true_value, false_value)
func (p *Parser) parseIIfCall() (*ast.IIfCall, error) {
	p.nextToken() // consume (

	// Parse boolean predicate
	pred, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after IIF condition, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse then expression
	thenExpr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after IIF then expression, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse else expression
	elseExpr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after IIF, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return &ast.IIfCall{
		Predicate:      pred,
		ThenExpression: thenExpr,
		ElseExpression: elseExpr,
	}, nil
}

// parseParseCall parses PARSE(string AS type [USING culture]) or TRY_PARSE(string AS type [USING culture])
func (p *Parser) parseParseCall(isTry bool) (ast.ScalarExpression, error) {
	p.nextToken() // consume (

	// Parse string value expression
	strVal, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Expect AS
	if strings.ToUpper(p.curTok.Literal) != "AS" {
		return nil, fmt.Errorf("expected AS after PARSE value, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume AS

	// Parse data type
	dataType, err := p.parseDataType()
	if err != nil {
		return nil, err
	}

	var culture ast.ScalarExpression

	// Check for USING culture
	if strings.ToUpper(p.curTok.Literal) == "USING" {
		p.nextToken() // consume USING
		culture, err = p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after PARSE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	if isTry {
		return &ast.TryParseCall{
			StringValue: strVal,
			DataType:    dataType,
			Culture:     culture,
		}, nil
	}
	return &ast.ParseCall{
		StringValue: strVal,
		DataType:    dataType,
		Culture:     culture,
	}, nil
}

// parseChangeTableReference parses CHANGETABLE(CHANGES ...) or CHANGETABLE(VERSION ...)
func (p *Parser) parseChangeTableReference() (ast.TableReference, error) {
	p.nextToken() // consume CHANGETABLE

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after CHANGETABLE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	upper := strings.ToUpper(p.curTok.Literal)
	if upper == "CHANGES" {
		return p.parseChangeTableChangesReference()
	} else if upper == "VERSION" {
		return p.parseChangeTableVersionReference()
	}

	return nil, fmt.Errorf("expected CHANGES or VERSION after CHANGETABLE(, got %s", p.curTok.Literal)
}

// parseChangeTableChangesReference parses CHANGETABLE(CHANGES table, version [, FORCESEEK])
func (p *Parser) parseChangeTableChangesReference() (*ast.ChangeTableChangesTableReference, error) {
	p.nextToken() // consume CHANGES

	ref := &ast.ChangeTableChangesTableReference{
		ForPath: false,
	}

	// Parse target table
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	ref.Target = son

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after table name in CHANGETABLE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse since version
	version, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	ref.SinceVersion = version

	// Check for optional FORCESEEK
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		if strings.ToUpper(p.curTok.Literal) == "FORCESEEK" {
			ref.ForceSeek = true
			p.nextToken()
		}
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after CHANGETABLE arguments, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse AS alias
	if p.curTok.Type != TokenAs {
		return nil, fmt.Errorf("expected AS after CHANGETABLE(...), got %s", p.curTok.Literal)
	}
	p.nextToken() // consume AS
	ref.Alias = p.parseIdentifier()

	// Check for column list: alias(c1, c2, ...)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			ref.Columns = append(ref.Columns, p.parseIdentifier())
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

	return ref, nil
}

// parseChangeTableVersionReference parses CHANGETABLE(VERSION table, (cols), (vals) [, FORCESEEK])
func (p *Parser) parseChangeTableVersionReference() (*ast.ChangeTableVersionTableReference, error) {
	p.nextToken() // consume VERSION

	ref := &ast.ChangeTableVersionTableReference{
		ForPath: false,
	}

	// Parse target table
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	ref.Target = son

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after table name in CHANGETABLE VERSION, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse primary key columns: (c1, c2, ...)
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( for primary key columns, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		ref.PrimaryKeyColumns = append(ref.PrimaryKeyColumns, p.parseIdentifier())
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}
	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after primary key columns, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse primary key values: (v1, v2, ...)
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( for primary key values, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		val, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		ref.PrimaryKeyValues = append(ref.PrimaryKeyValues, val)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}
	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	// Check for optional FORCESEEK
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		if strings.ToUpper(p.curTok.Literal) == "FORCESEEK" {
			ref.ForceSeek = true
			p.nextToken()
		}
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after CHANGETABLE VERSION arguments, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse AS alias
	if p.curTok.Type != TokenAs {
		return nil, fmt.Errorf("expected AS after CHANGETABLE(...), got %s", p.curTok.Literal)
	}
	p.nextToken() // consume AS
	ref.Alias = p.parseIdentifier()

	// Check for column list: alias(c1, c2, ...)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			ref.Columns = append(ref.Columns, p.parseIdentifier())
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

	return ref, nil
}

// parseOverClause parses an OVER clause after a function call
// Handles both: OVER Win1 and OVER (PARTITION BY c1 ORDER BY c2 ROWS ...)
func (p *Parser) parseOverClause() (*ast.OverClause, error) {
	// Current token should be OVER, consume it
	p.nextToken() // consume OVER

	overClause := &ast.OverClause{}

	// Check if it's just a window name (no parentheses)
	if p.curTok.Type != TokenLParen {
		// It's OVER WindowName
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			overClause.WindowName = p.parseIdentifier()
			return overClause, nil
		}
		return nil, fmt.Errorf("expected ( or window name after OVER, got %s", p.curTok.Literal)
	}

	p.nextToken() // consume (

	// Check if it starts with a window name reference
	// OVER (Win1 ORDER BY ...) or OVER (Win1 PARTITION BY ... )
	// This is tricky because we need to distinguish between Win1 (window name) and c1 (column name in PARTITION BY)
	if p.curTok.Type == TokenIdent && p.peekTok.Type != TokenComma && p.peekTok.Type != TokenRParen {
		upperPeek := strings.ToUpper(p.peekTok.Literal)
		if upperPeek != "BY" && upperPeek != "," {
			// Could be a window name reference if followed by ORDER, PARTITION, ROWS, RANGE, or )
			if upperPeek == "ORDER" || upperPeek == "PARTITION" || upperPeek == "ROWS" || upperPeek == "RANGE" || p.peekTok.Type == TokenRParen {
				overClause.WindowName = p.parseIdentifier()
			}
		}
	}

	// Parse PARTITION BY
	if strings.ToUpper(p.curTok.Literal) == "PARTITION" {
		p.nextToken() // consume PARTITION
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		// Parse partition expressions
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			if strings.ToUpper(p.curTok.Literal) == "ORDER" || strings.ToUpper(p.curTok.Literal) == "ROWS" || strings.ToUpper(p.curTok.Literal) == "RANGE" {
				break
			}
			partExpr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			overClause.Partitions = append(overClause.Partitions, partExpr)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
	}

	// Parse ORDER BY
	if p.curTok.Type == TokenOrder {
		orderBy, err := p.parseOrderByClause()
		if err != nil {
			return nil, err
		}
		overClause.OrderByClause = orderBy
	}

	// Parse window frame (ROWS/RANGE)
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "ROWS" || upperLit == "RANGE" {
		frameClause, err := p.parseWindowFrameClause()
		if err != nil {
			return nil, err
		}
		overClause.WindowFrameClause = frameClause
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in OVER clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return overClause, nil
}

// parseWindowFrameClause parses ROWS/RANGE ... BETWEEN ... AND ...
func (p *Parser) parseWindowFrameClause() (*ast.WindowFrameClause, error) {
	frame := &ast.WindowFrameClause{}

	// Parse ROWS or RANGE
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "ROWS" {
		frame.WindowFrameType = "Rows"
	} else if upperLit == "RANGE" {
		frame.WindowFrameType = "Range"
	} else {
		return nil, fmt.Errorf("expected ROWS or RANGE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse BETWEEN or single boundary
	if strings.ToUpper(p.curTok.Literal) == "BETWEEN" {
		p.nextToken() // consume BETWEEN
		top, err := p.parseWindowDelimiter()
		if err != nil {
			return nil, err
		}
		frame.Top = top

		if strings.ToUpper(p.curTok.Literal) != "AND" {
			return nil, fmt.Errorf("expected AND in ROWS BETWEEN, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume AND

		bottom, err := p.parseWindowDelimiter()
		if err != nil {
			return nil, err
		}
		frame.Bottom = bottom
	} else {
		// Single boundary (e.g., ROWS UNBOUNDED PRECEDING)
		top, err := p.parseWindowDelimiter()
		if err != nil {
			return nil, err
		}
		frame.Top = top
	}

	return frame, nil
}

// parseWindowDelimiter parses UNBOUNDED PRECEDING/FOLLOWING, CURRENT ROW, n PRECEDING/FOLLOWING
func (p *Parser) parseWindowDelimiter() (*ast.WindowDelimiter, error) {
	delim := &ast.WindowDelimiter{}

	upperLit := strings.ToUpper(p.curTok.Literal)

	if upperLit == "CURRENT" {
		p.nextToken() // consume CURRENT
		if strings.ToUpper(p.curTok.Literal) != "ROW" {
			return nil, fmt.Errorf("expected ROW after CURRENT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ROW
		delim.WindowDelimiterType = "CurrentRow"
	} else if upperLit == "UNBOUNDED" {
		p.nextToken() // consume UNBOUNDED
		upperDir := strings.ToUpper(p.curTok.Literal)
		if upperDir == "PRECEDING" {
			delim.WindowDelimiterType = "UnboundedPreceding"
		} else if upperDir == "FOLLOWING" {
			delim.WindowDelimiterType = "UnboundedFollowing"
		} else {
			return nil, fmt.Errorf("expected PRECEDING or FOLLOWING after UNBOUNDED, got %s", p.curTok.Literal)
		}
		p.nextToken()
	} else {
		// n PRECEDING or n FOLLOWING
		offset, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		delim.OffsetValue = offset

		upperDir := strings.ToUpper(p.curTok.Literal)
		if upperDir == "PRECEDING" {
			delim.WindowDelimiterType = "ValuePreceding"
		} else if upperDir == "FOLLOWING" {
			delim.WindowDelimiterType = "ValueFollowing"
		} else {
			return nil, fmt.Errorf("expected PRECEDING or FOLLOWING after value, got %s", p.curTok.Literal)
		}
		p.nextToken()
	}

	return delim, nil
}

// parseWindowClause parses WINDOW Win1 AS (...), Win2 AS (...)
func (p *Parser) parseWindowClause() (*ast.WindowClause, error) {
	p.nextToken() // consume WINDOW

	clause := &ast.WindowClause{}

	for {
		def := &ast.WindowDefinition{}

		// Parse window name
		def.WindowName = p.parseIdentifier()

		// Expect AS
		if strings.ToUpper(p.curTok.Literal) != "AS" {
			return nil, fmt.Errorf("expected AS after window name, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume AS

		// Expect (
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after AS in window definition, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		// Check if it references another window name
		if p.curTok.Type == TokenIdent {
			upperPeek := strings.ToUpper(p.peekTok.Literal)
			// It's a reference if followed by ) or PARTITION or ORDER
			if p.peekTok.Type == TokenRParen || upperPeek == "PARTITION" || upperPeek == "ORDER" {
				// Could be a window name reference
				if p.peekTok.Type == TokenRParen {
					// Just a window name reference: Win1 AS (Win2)
					def.RefWindowName = p.parseIdentifier()
				} else if upperPeek != "BY" {
					// Window name followed by more clauses
					def.RefWindowName = p.parseIdentifier()
				}
			}
		}

		// Parse PARTITION BY
		if strings.ToUpper(p.curTok.Literal) == "PARTITION" {
			p.nextToken() // consume PARTITION
			if strings.ToUpper(p.curTok.Literal) == "BY" {
				p.nextToken() // consume BY
			}
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				if strings.ToUpper(p.curTok.Literal) == "ORDER" {
					break
				}
				partExpr, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				def.Partitions = append(def.Partitions, partExpr)
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
		}

		// Parse ORDER BY
		if p.curTok.Type == TokenOrder {
			orderBy, err := p.parseOrderByClause()
			if err != nil {
				return nil, err
			}
			def.OrderByClause = orderBy
		}

		// Expect )
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) in window definition, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		clause.WindowDefinition = append(clause.WindowDefinition, def)

		// Check for comma (more window definitions)
		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume ,
	}

	return clause, nil
}

// parsePivotedTableReference parses PIVOT clause
// Syntax: table PIVOT (aggregate_func(columns) FOR pivot_column IN (value1, value2, ...)) AS alias
func (p *Parser) parsePivotedTableReference(tableRef ast.TableReference) (*ast.PivotedTableReference, error) {
	p.nextToken() // consume PIVOT

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after PIVOT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	pivoted := &ast.PivotedTableReference{
		TableReference: tableRef,
		ForPath:        false,
	}

	// Parse aggregate function identifier (may be multi-part like dbo.z1.MyAggregate)
	aggregateId := &ast.MultiPartIdentifier{}
	for {
		id := p.parseIdentifier()
		aggregateId.Identifiers = append(aggregateId.Identifiers, id)
		aggregateId.Count++
		if p.curTok.Type == TokenDot {
			p.nextToken() // consume .
		} else {
			break
		}
	}
	pivoted.AggregateFunctionIdentifier = aggregateId

	// Expect ( for aggregate function parameters
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( for aggregate function parameters, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse value columns (parameters to aggregate function)
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		col, err := p.parseColumnReference()
		if err != nil {
			return nil, err
		}
		pivoted.ValueColumns = append(pivoted.ValueColumns, col)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after aggregate function parameters, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Expect FOR keyword
	if strings.ToUpper(p.curTok.Literal) != "FOR" {
		return nil, fmt.Errorf("expected FOR in PIVOT clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume FOR

	// Parse pivot column
	col, err := p.parseColumnReference()
	if err != nil {
		return nil, err
	}
	pivoted.PivotColumn = col

	// Expect IN keyword
	if p.curTok.Type != TokenIn {
		return nil, fmt.Errorf("expected IN in PIVOT clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume IN

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after IN, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse IN columns (values)
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		id := p.parseIdentifier()
		pivoted.InColumns = append(pivoted.InColumns, id)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after IN values, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Expect ) to close PIVOT clause
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) to close PIVOT clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse required alias (AS alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
		pivoted.Alias = p.parseIdentifier()
	}

	return pivoted, nil
}

// parseUnpivotedTableReference parses UNPIVOT clause
func (p *Parser) parseUnpivotedTableReference(tableRef ast.TableReference) (*ast.UnpivotedTableReference, error) {
	p.nextToken() // consume UNPIVOT

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after UNPIVOT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	unpivoted := &ast.UnpivotedTableReference{
		TableReference: tableRef,
		NullHandling:   "None",
		ForPath:        false,
	}

	// Parse pivot value column
	unpivoted.PivotValue = p.parseIdentifier()

	// Expect FOR keyword
	if strings.ToUpper(p.curTok.Literal) != "FOR" {
		return nil, fmt.Errorf("expected FOR in UNPIVOT clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume FOR

	// Parse pivot column
	unpivoted.PivotColumn = p.parseIdentifier()

	// Expect IN keyword
	if p.curTok.Type != TokenIn {
		return nil, fmt.Errorf("expected IN in UNPIVOT clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume IN

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after IN, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse IN columns
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		col, err := p.parseColumnReference()
		if err != nil {
			return nil, err
		}
		unpivoted.InColumns = append(unpivoted.InColumns, col)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after IN columns, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Expect ) to close UNPIVOT clause
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) to close UNPIVOT clause, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse required alias (AS alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
		unpivoted.Alias = p.parseIdentifier()
	}

	return unpivoted, nil
}
