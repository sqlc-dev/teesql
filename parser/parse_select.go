// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/teesql/ast"
)

func (p *Parser) parsePrintStatement() (*ast.PrintStatement, error) {
	astStart := p.curTok

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

	return spanned(p, &ast.PrintStatement{Expression: expr}, astStart), nil
}

func (p *Parser) parseThrowStatement() (*ast.ThrowStatement, error) {
	astStart := p.curTok

	// Consume THROW
	p.nextToken()

	stmt := &ast.ThrowStatement{}

	// THROW can be used without arguments (re-throw)
	if p.curTok.Type == TokenSemicolon || p.curTok.Type == TokenEOF ||
		p.curTok.Type == TokenSelect || p.curTok.Type == TokenPrint || p.curTok.Type == TokenThrow {
		return spanned(p, stmt, astStart), nil
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

	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseSelectStatement() (*ast.SelectStatement, error) {
	astStart := p.curTok

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
			return spanned(p, selStmt, astStart), nil
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

	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseQueryExpression() (ast.QueryExpression, error) {
	astStart := p.curTok

	qe, _, _, err := p.parseQueryExpressionWithInto()
	return spanned(p, qe, astStart), err
}

func (p *Parser) parseQueryExpressionWithInto() (ast.QueryExpression, *ast.SchemaObjectName, *ast.Identifier, error) {
	astStart := p.curTok

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

	if s, ok := any(left).(spannable); ok {
		p.spanFrom(astStart, s)
	}

	return left, into, on, nil
}

func (p *Parser) parsePrimaryQueryExpression() (ast.QueryExpression, *ast.SchemaObjectName, *ast.Identifier, error) {
	astStart := p.curTok

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
		return spanned(p, &ast.QueryParenthesisExpression{QueryExpression: qe}, astStart), into, on, nil
	}

	qe, into, on, err := p.parseQuerySpecificationWithInto()
	if err == nil {
		if s, ok := any(qe).(spannable); ok {
			p.spanFrom(astStart, s)
		}
	}
	return qe, into, on, err
}

// parseRestOfBinaryQueryExpression parses binary query operations (UNION/INTERSECT/EXCEPT)
// starting with a left operand that's already been parsed.
func (p *Parser) parseRestOfBinaryQueryExpression(left ast.QueryExpression) (ast.QueryExpression, error) {
	astStart := p.curTok
	// The binary expression starts where its (already-parsed) left operand
	// starts, not at the operator token.
	leftStartsEarlier := false
	if l, ok := any(left).(spannable); ok && l.Frag().HasSpan() {
		leftStartsEarlier = true
	}

	// Check for binary operations (UNION, EXCEPT, INTERSECT)
	for p.curTok.Type == TokenUnion || p.curTok.Type == TokenExcept || p.curTok.Type == TokenIntersect {
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
		right, _, _, err := p.parsePrimaryQueryExpression()
		if err != nil {
			return nil, err
		}

		bqe := &ast.BinaryQueryExpression{
			BinaryQueryExpressionType: opType,
			All:                       all,
			FirstQueryExpression:      left,
			SecondQueryExpression:     right,
		}
		if leftStartsEarlier {
			p.spanFromChild(bqe, bqe.FirstQueryExpression)
		}

		left = bqe
	}

	if leftStartsEarlier {
		if l, ok := any(left).(spannable); ok && l.Frag().HasSpan() {
			return left, nil
		}
	}
	return spanned(p, left, astStart), nil
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
	astStart := p.curTok

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

	return spanned(p, qs, astStart), nil
}

func (p *Parser) parseTopRowFilter() (*ast.TopRowFilter, error) {
	astStart := p.curTok

	// Consume TOP
	p.nextToken()

	top := &ast.TopRowFilter{}

	// Check for parenthesized expression
	if p.curTok.Type == TokenLParen {
		parenStart := p.curTok
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
			ss := &ast.ScalarSubquery{QueryExpression: qe}
			p.spanFrom(parenStart, ss)
			top.Expression = ss
		} else {
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			// Wrap in ParenthesisExpression
			pe := &ast.ParenthesisExpression{Expression: expr}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )
			p.spanFrom(parenStart, pe)
			top.Expression = pe
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

	return spanned(p, top, astStart), nil
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
	astStart := p.curTok

	// Check for *
	if p.curTok.Type == TokenStar {
		p.nextToken()
		return spanned(p, &ast.SelectStarExpression{}, astStart), nil
	}

	// Check for variable assignment: @var = expr or @var ||= expr
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		varName := p.curTok.Literal
		varNameTok := p.curTok
		p.nextToken() // consume variable

		// Check if this is an assignment
		if p.isCompoundAssignment() {
			ssv := &ast.SelectSetVariable{
				Variable:       p.varRefFromToken(varNameTok),
				AssignmentKind: p.getAssignmentKind(),
			}
			p.nextToken() // consume assignment operator

			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			ssv.Expression = expr
			return spanned(p, ssv, astStart), nil
		}

		// Not an assignment, treat as regular scalar expression starting with variable
		var varExpr ast.ScalarExpression
		if strings.HasPrefix(varName, "@@") {
			gve := &ast.GlobalVariableExpression{Name: varName}
			p.tokSpan(gve, varNameTok)
			varExpr = gve
		} else {
			varExpr = p.varRefFromToken(varNameTok)
		}

		// Handle postfix operations (method calls, property access)
		expr, err := p.handlePostfixOperations(varExpr)
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
		return spanned(p, sse, astStart), nil
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

			return spanned(p, &ast.SelectScalarExpression{
				Expression: expr,
				ColumnName: columnName,
			}, astStart), nil
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
			return spanned(p, &ast.SelectStarExpression{
				Qualifier: colRef.MultiPartIdentifier,
			}, astStart), nil
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
			p.respanEnd(e)
		case *ast.ColumnReferenceExpression:
			e.Collation = collation
			p.respanEnd(e)
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

	// When the expression's recorded span is pinned and starts after the
	// element's first token (e.g. a leading dot excluded by ScriptDom),
	// the select element inherits the later start.
	if es, ok := expr.(spannable); ok && es.Frag().Pinned() && sse.ColumnName == nil {
		esf := es.Frag()
		if esf.HasSpan() {
			su, sl, sc := p.srcMap.at(astStart.Pos)
			_ = sl
			_ = sc
			if esf.StartOffset > su {
				p.spanFromChild(sse, expr)
				sse.Pin()
				return sse, nil
			}
		}
	}

	return spanned(p, sse, astStart), nil
}

func (p *Parser) parseIdentifier() *ast.Identifier {
	astStart := p.curTok

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

	id := p.spanIdent(literal, quoteType)
	p.nextToken()
	return spanned(p, id, astStart)
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
	astStart := p.curTok

	spanV1, spanErr1 := p.parseBitwiseXorExpression()
	return spanned(p, spanV1, astStart), spanErr1
}

// In T-SQL, bitwise operator precedence from lowest to highest is: ^ (XOR), | (OR), & (AND)
// This is different from C where it's: | (OR), ^ (XOR), & (AND)

func (p *Parser) parseBitwiseXorExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	left, err := p.parseBitwiseOrExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenCaret {
		p.nextToken()

		right, err := p.parseBitwiseOrExpression()
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryExpression{
			BinaryExpressionType: "BitwiseXor",
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}

	return spanned(p, left, astStart), nil
}

func (p *Parser) parseBitwiseOrExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	left, err := p.parseBitwiseAndExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenPipe {
		p.nextToken()

		right, err := p.parseBitwiseAndExpression()
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryExpression{
			BinaryExpressionType: "BitwiseOr",
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}

	return spanned(p, left, astStart), nil
}

func (p *Parser) parseBitwiseAndExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	left, err := p.parseShiftExpression()
	if err != nil {
		return nil, err
	}

	for p.curTok.Type == TokenBitwiseAnd {
		p.nextToken()

		right, err := p.parseShiftExpression()
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryExpression{
			BinaryExpressionType: "BitwiseAnd",
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}

	return spanned(p, left, astStart), nil
}

func (p *Parser) parseShiftExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

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

	return spanned(p, left, astStart), nil
}

func (p *Parser) parseAdditiveExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

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

	return spanned(p, left, astStart), nil
}

func (p *Parser) parseMultiplicativeExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

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

	return spanned(p, left, astStart), nil
}

// parsePostfixExpression handles postfix operators like AT TIME ZONE
func (p *Parser) parsePostfixExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

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
			methodName := p.spanIdent(name, quoteType)
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

				p.spanFromChild(fc, expr)
				expr = fc
			} else {
				// It's a property access: expr.prop
				propAccess := &ast.UserDefinedTypePropertyAccess{
					CallTarget: &ast.ExpressionCallTarget{
						Expression: expr,
					},
					PropertyName: methodName,
				}
				p.spanFromChild(propAccess, expr)
				expr = propAccess
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

	return spanned(p, expr, astStart), nil
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
			methodName := p.spanIdent(name, quoteType)
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

				p.spanFromChild(fc, expr)
				expr = fc
			} else {
				// It's a property access: expr.prop
				propAccess := &ast.UserDefinedTypePropertyAccess{
					CallTarget: &ast.ExpressionCallTarget{
						Expression: expr,
					},
					PropertyName: methodName,
				}
				p.spanFromChild(propAccess, expr)
				expr = propAccess
			}
			continue
		}

		// No more postfix operations
		break
	}

	return expr, nil
}

func (p *Parser) parsePrimaryExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	switch p.curTok.Type {
	case TokenNull:
		val := p.curTok.Literal
		p.nextToken()
		return spanned(p, &ast.NullLiteral{LiteralType: "Null", Value: val}, astStart), nil
	case TokenDefault:
		val := p.curTok.Literal
		p.nextToken()
		return spanned(p, &ast.DefaultLiteral{LiteralType: "Default", Value: val}, astStart), nil
	case TokenMinus:
		p.nextToken()
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		return spanned(p, &ast.UnaryExpression{UnaryExpressionType: "Negative", Expression: expr}, astStart), nil
	case TokenPlus:
		p.nextToken()
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		return spanned(p, &ast.UnaryExpression{UnaryExpressionType: "Positive", Expression: expr}, astStart), nil
	case TokenError:
		// Handle ~ (bitwise NOT) operator
		if p.curTok.Literal == "~" {
			p.nextToken()
			expr, err := p.parsePrimaryExpression()
			if err != nil {
				return nil, err
			}
			return spanned(p, &ast.UnaryExpression{UnaryExpressionType: "BitwiseNot", Expression: expr}, astStart), nil
		}
		return nil, fmt.Errorf("unexpected token in expression: %s", p.curTok.Literal)
	case TokenLeft:
		// LEFT can be a function name (string function)
		if p.peekTok.Type == TokenLParen {
			p.nextToken() // consume LEFT
			spanV2, spanErr2 := p.parseLeftFunctionCall()
			return spanned(p, spanV2, astStart), spanErr2
		}
		return nil, fmt.Errorf("unexpected token in expression: %s", p.curTok.Literal)
	case TokenRight:
		// RIGHT can be a function name (string function)
		if p.peekTok.Type == TokenLParen {
			p.nextToken() // consume RIGHT
			spanV3, spanErr3 := p.parseRightFunctionCall()
			return spanned(p, spanV3, astStart), spanErr3
		}
		return nil, fmt.Errorf("unexpected token in expression: %s", p.curTok.Literal)
	case TokenIdent:
		// Check if it's a global variable reference (starts with @@)
		if strings.HasPrefix(p.curTok.Literal, "@@") {
			name := p.curTok.Literal
			p.nextToken()
			return spanned(p, &ast.GlobalVariableExpression{Name: name}, astStart), nil
		}
		// Check if it's a variable reference (starts with @)
		if strings.HasPrefix(p.curTok.Literal, "@") {
			nameTok := p.curTok
			p.nextToken()
			return spanned(p, p.varRefFromToken(nameTok), astStart), nil
		}
		// Check for N-prefixed national string (N'...')
		if strings.ToUpper(p.curTok.Literal) == "N" && p.peekTok.Type == TokenString {
			p.nextToken() // consume N
			spanV4, spanErr4 := p.parseNationalStringLiteral()
			return spanned(p, spanV4, astStart), spanErr4
		}
		// Check for CAST/CONVERT special functions
		upper := strings.ToUpper(p.curTok.Literal)
		if upper == "CAST" && p.peekTok.Type == TokenLParen {
			spanV5, spanErr5 := p.parseCastCall()
			return spanned(p, spanV5, astStart), spanErr5
		}
		if upper == "CONVERT" && p.peekTok.Type == TokenLParen {
			spanV6, spanErr6 := p.parseConvertCall()
			return spanned(p, spanV6, astStart), spanErr6
		}
		if upper == "TRY_CAST" && p.peekTok.Type == TokenLParen {
			spanV7, spanErr7 := p.parseTryCastCall()
			return spanned(p, spanV7, astStart), spanErr7
		}
		if upper == "TRY_CONVERT" && p.peekTok.Type == TokenLParen {
			spanV8, spanErr8 := p.parseTryConvertCall()
			return spanned(p, spanV8, astStart), spanErr8
		}
		if upper == "NULLIF" && p.peekTok.Type == TokenLParen {
			spanV9, spanErr9 := p.parseNullIfExpression()
			return spanned(p, spanV9, astStart), spanErr9
		}
		if upper == "COALESCE" && p.peekTok.Type == TokenLParen {
			spanV10, spanErr10 := p.parseCoalesceExpression()
			return spanned(p, spanV10, astStart), spanErr10
		}
		if upper == "IDENTITY" && p.peekTok.Type == TokenLParen {
			spanV11, spanErr11 := p.parseIdentityFunctionCall()
			return spanned(p, spanV11, astStart), spanErr11
		}
		if upper == "IDENTITYCOL" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "IdentityCol"}, astStart), nil
		}
		if upper == "ROWGUIDCOL" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "RowGuidCol"}, astStart), nil
		}
		if upper == "$ACTION" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnAction"}, astStart), nil
		}
		if upper == "$IDENTITY" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnIdentity"}, astStart), nil
		}
		if upper == "$ROWGUID" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnRowGuid"}, astStart), nil
		}
		if upper == "$CUID" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnCuid"}, astStart), nil
		}
		if upper == "$NODE_ID" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnGraphNodeId"}, astStart), nil
		}
		if upper == "$EDGE_ID" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnGraphEdgeId"}, astStart), nil
		}
		if upper == "$FROM_ID" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnGraphFromId"}, astStart), nil
		}
		if upper == "$TO_ID" {
			p.nextToken()
			return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnGraphToId"}, astStart), nil
		}
		// Check for NEXT VALUE FOR sequence expression
		if upper == "NEXT" && strings.ToUpper(p.peekTok.Literal) == "VALUE" {
			spanV12, spanErr12 := p.parseNextValueForExpression()
			return spanned(p, spanV12, astStart), spanErr12
		}
		// Check for parameterless calls (USER, CURRENT_USER, etc.) without parentheses
		if p.peekTok.Type != TokenLParen {
			parameterlessType := getParameterlessCallType(upper)
			if parameterlessType != "" {
				p.nextToken()
				call := &ast.ParameterlessCall{ParameterlessCallType: parameterlessType}
				// Check for optional COLLATE clause
				if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
					p.nextToken() // consume COLLATE
					call.Collation = p.parseIdentifier()
				}
				return spanned(p, call, astStart), nil
			}
		}
		spanV13, spanErr13 := p.parseColumnReferenceOrFunctionCall()
		return spanned(p, spanV13, astStart), spanErr13
	case TokenNumber:
		val := p.curTok.Literal
		p.nextToken()
		// Check if it's scientific notation (real literal)
		if strings.ContainsAny(val, "eE") {
			return spanned(p, &ast.RealLiteral{LiteralType: "Real", Value: val}, astStart), nil
		}
		// Check if it's a decimal number
		if strings.Contains(val, ".") {
			return spanned(p, &ast.NumericLiteral{LiteralType: "Numeric", Value: val}, astStart), nil
		}
		// Large numbers beyond INT range should be NumericLiteral
		// INT range is -2,147,483,648 to 2,147,483,647
		if len(val) > 10 || (len(val) == 10 && val > "2147483647") {
			return spanned(p, &ast.NumericLiteral{LiteralType: "Numeric", Value: val}, astStart), nil
		}
		return spanned(p, &ast.IntegerLiteral{LiteralType: "Integer", Value: val}, astStart), nil
	case TokenMoney:
		val := p.curTok.Literal
		p.nextToken()
		return spanned(p, &ast.MoneyLiteral{LiteralType: "Money", Value: val}, astStart), nil
	case TokenBinary:
		val := p.curTok.Literal
		p.nextToken()
		return spanned(p, &ast.BinaryLiteral{LiteralType: "Binary", Value: val, IsLargeObject: false}, astStart), nil
	case TokenString:
		spanV14, spanErr14 := p.parseStringLiteral()
		return spanned(p, spanV14, astStart), spanErr14
	case TokenNationalString:
		spanV15, spanErr15 := p.parseNationalStringFromToken()
		return spanned(p, spanV15, astStart), spanErr15
	case TokenLBrace:
		spanV16, spanErr16 := p.parseOdbcLiteral()
		return spanned(p, spanV16, astStart), spanErr16
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
			ss := &ast.ScalarSubquery{QueryExpression: qe}
			// Check for optional COLLATE clause
			if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
				p.nextToken() // consume COLLATE
				ss.Collation = p.parseIdentifier()
			}
			return spanned(p, ss, astStart), nil
		}
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		// Check if next token is UNION/INTERSECT/EXCEPT - if so, we're actually inside
		// a query expression, not a scalar expression. This happens with nested parens
		// like ((SELECT ...) UNION SELECT ...) where the inner parens create a ScalarSubquery
		// but the outer expression is a binary query expression.
		if p.curTok.Type == TokenUnion || p.curTok.Type == TokenIntersect || p.curTok.Type == TokenExcept {
			// Convert the scalar subquery to a query parenthesis expression
			if ss, ok := expr.(*ast.ScalarSubquery); ok {
				qpe := &ast.QueryParenthesisExpression{QueryExpression: ss.QueryExpression}
				// The parenthesis expression keeps the subquery's span
				// (including the parens).
				if ss.Frag().HasSpan() {
					*qpe.Frag() = *ss.Frag()
				}
				qe, err := p.parseRestOfBinaryQueryExpression(qpe)
				if err != nil {
					return nil, err
				}
				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
				}
				p.nextToken()
				ss := &ast.ScalarSubquery{QueryExpression: qe}
				// Check for optional COLLATE clause
				if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
					p.nextToken() // consume COLLATE
					ss.Collation = p.parseIdentifier()
				}
				return spanned(p, ss, astStart), nil
			}
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken()
		// Check for property access after parenthesized expression: (c1).SomeProperty
		spanV17, spanErr17 := p.parsePostExpressionAccess(spanned(p, &ast.ParenthesisExpression{Expression: expr}, astStart))
		return spanned(p, spanV17, astStart), spanErr17
	case TokenCase:
		spanV18, spanErr18 := p.parseCaseExpression()
		return spanned(p, spanV18, astStart), spanErr18
	case TokenStar:
		// Wildcard column reference (e.g., * in count(*))
		p.nextToken()
		return spanned(p, &ast.ColumnReferenceExpression{ColumnType: "Wildcard"}, astStart), nil
	case TokenDot:
		// Multi-part identifier starting with empty parts (e.g., ..t1.c1)
		spanV19, spanErr19 := p.parseColumnReferenceWithLeadingDots()
		return spanned(p, spanV19, astStart), spanErr19
	case TokenMaster, TokenDatabase, TokenKey, TokenTable, TokenIndex,
		TokenSchema, TokenView, TokenTime:
		// Keywords that can be used as identifiers in column/table references
		spanV20, spanErr20 := p.parseColumnReferenceOrFunctionCall()
		return spanned(p, spanV20, astStart), spanErr20
	case TokenUser:
		// USER without parentheses is a ParameterlessCall
		if p.peekTok.Type != TokenLParen && p.peekTok.Type != TokenDot {
			p.nextToken()
			call := &ast.ParameterlessCall{ParameterlessCallType: "User"}
			// Check for optional COLLATE clause
			if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
				p.nextToken() // consume COLLATE
				call.Collation = p.parseIdentifier()
			}
			return spanned(p, call, astStart), nil
		}
		spanV21, spanErr21 := p.parseColumnReferenceOrFunctionCall()
		return spanned(p, spanV21, astStart), spanErr21
	default:
		return nil, fmt.Errorf("unexpected token in expression: %s", p.curTok.Literal)
	}
}

func (p *Parser) parseCaseExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	p.nextToken() // consume CASE

	// Check if it's a searched CASE (CASE WHEN ...) or simple CASE (CASE expr WHEN ...)
	if p.curTok.Type == TokenWhen {
		// Searched CASE expression
		spanV22, spanErr22 := p.parseSearchedCaseExpression()
		return spanned(p, spanV22, astStart), spanErr22
	}
	// Simple CASE expression
	spanV23, spanErr23 := p.parseSimpleCaseExpression()
	return spanned(p, spanV23, astStart), spanErr23
}

func (p *Parser) parseSearchedCaseExpression() (*ast.SearchedCaseExpression, error) {
	astStart := p.curTok

	expr := &ast.SearchedCaseExpression{}

	for p.curTok.Type == TokenWhen {
		whenTok := p.curTok
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

		wc := &ast.SearchedWhenClause{
			WhenExpression: when,
			ThenExpression: then,
		}
		p.spanFrom(whenTok, wc)
		expr.WhenClauses = append(expr.WhenClauses, wc)
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

	// Check for optional COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		expr.Collation = p.parseIdentifier()
	}

	return spanned(p, expr, astStart), nil
}

func (p *Parser) parseSimpleCaseExpression() (*ast.SimpleCaseExpression, error) {
	astStart := p.curTok

	expr := &ast.SimpleCaseExpression{}

	// Parse input expression
	input, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	expr.InputExpression = input

	for p.curTok.Type == TokenWhen {
		whenTok := p.curTok
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

		wc := &ast.SimpleWhenClause{
			WhenExpression: when,
			ThenExpression: then,
		}
		p.spanFrom(whenTok, wc)
		expr.WhenClauses = append(expr.WhenClauses, wc)
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

	// Check for optional COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		expr.Collation = p.parseIdentifier()
	}

	return spanned(p, expr, astStart), nil
}

// parseNextValueForExpression parses NEXT VALUE FOR sequence_name [OVER (...)]
func (p *Parser) parseNextValueForExpression() (*ast.NextValueForExpression, error) {
	astStart := p.curTok

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

	return spanned(p, expr, astStart), nil
}

func (p *Parser) parseOdbcLiteral() (ast.ScalarExpression, error) {
	astStart := p.curTok

	// Consume {
	p.nextToken()

	// Check what type of ODBC escape this is
	keyword := strings.ToUpper(p.curTok.Literal)

	// { FN function_name(...) } - ODBC scalar function
	if keyword == "FN" {
		p.nextToken() // consume FN
		spanV24, spanErr24 := p.parseOdbcFunctionCall()
		return spanned(p, spanV24, astStart), spanErr24
	}

	// Determine the ODBC literal type
	var odbcType string
	switch keyword {
	case "GUID":
		odbcType = "Guid"
	case "T":
		odbcType = "Time"
	case "D":
		odbcType = "Date"
	case "TS":
		odbcType = "Timestamp"
	default:
		return nil, fmt.Errorf("expected guid, fn, t, d, or ts in ODBC escape, got %s", p.curTok.Literal)
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

	return spanned(p, &ast.OdbcLiteral{
		LiteralType:     "Odbc",
		OdbcLiteralType: odbcType,
		IsNational:      isNational,
		Value:           value,
	}, astStart), nil
}

func (p *Parser) parseOdbcFunctionCall() (*ast.OdbcFunctionCall, error) {
	astStart := p.curTok

	// Parse function name
	name := p.parseIdentifier()

	call := &ast.OdbcFunctionCall{
		Name: unspan(name),
	}

	// Check for parentheses (parameters)
	if p.curTok.Type == TokenLParen {
		call.ParametersUsed = true
		p.nextToken() // consume (

		// Handle special extract function: extract(element FROM expr)
		if strings.ToLower(name.Value) == "extract" {
			// Parse the extracted element (like "hour", "minute", etc.)
			element := p.parseIdentifier()

			// Expect FROM keyword
			if p.curTok.Type != TokenFrom {
				return nil, fmt.Errorf("expected FROM in ODBC extract function, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume FROM

			// Parse the expression to extract from
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			call.Parameters = append(call.Parameters, &ast.ExtractFromExpression{
				ExtractedElement: element,
				Expression:       expr,
			})
		} else {
			// Parse parameters
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				// For ODBC convert function, the second parameter is a conversion specifier
				// like sql_int, sql_varchar, etc.
				if strings.ToLower(name.Value) == "convert" && len(call.Parameters) == 1 {
					// Second parameter of convert is an OdbcConvertSpecification
					spec := &ast.OdbcConvertSpecification{
						Identifier: p.parseIdentifier(),
					}
					call.Parameters = append(call.Parameters, spec)
				} else {
					param, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					call.Parameters = append(call.Parameters, param)
				}

				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
		}

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) in ODBC function call, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	}

	// Consume closing }
	if p.curTok.Type != TokenRBrace {
		return nil, fmt.Errorf("expected } in ODBC function call, got %s", p.curTok.Literal)
	}
	p.nextToken()

	return spanned(p, call, astStart), nil
}

func (p *Parser) parseStringLiteral() (*ast.StringLiteral, error) {
	astStart := p.curTok

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
		return spanned(p, &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    isNational,
			IsLargeObject: false,
			Value:         value,
		}, astStart), nil
	}

	return spanned(p, &ast.StringLiteral{
		LiteralType:   "String",
		IsNational:    isNational,
		IsLargeObject: false,
		Value:         raw,
	}, astStart), nil
}

// parseStringLiteralValue creates a StringLiteral from the current token without consuming it
func (p *Parser) parseStringLiteralValue() *ast.StringLiteral {
	raw := p.curTok.Literal

	// Remove surrounding quotes and handle escaped quotes
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		inner := raw[1 : len(raw)-1]
		// Replace escaped quotes
		value := strings.ReplaceAll(inner, "''", "'")
		return p.strLit(value, false)
	}

	return p.strLit(raw, false)
}

func (p *Parser) parseNationalStringLiteral() (*ast.StringLiteral, error) {
	astStart := p.curTok

	raw := p.curTok.Literal
	p.nextToken()

	// Remove surrounding quotes and handle escaped quotes
	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		inner := raw[1 : len(raw)-1]
		// Replace escaped quotes
		value := strings.ReplaceAll(inner, "''", "'")
		return spanned(p, &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    true,
			IsLargeObject: false,
			Value:         value,
		}, astStart), nil
	}

	return spanned(p, &ast.StringLiteral{
		LiteralType:   "String",
		IsNational:    true,
		IsLargeObject: false,
		Value:         raw,
	}, astStart), nil
}

func (p *Parser) parseNationalStringFromToken() (*ast.StringLiteral, error) {
	astStart := p.curTok

	// Token is N'...' combined - strip the N prefix and quotes
	raw := p.curTok.Literal
	p.nextToken()

	// Raw is like N'value' or n'value'
	if len(raw) >= 3 && (raw[0] == 'N' || raw[0] == 'n') && raw[1] == '\'' && raw[len(raw)-1] == '\'' {
		inner := raw[2 : len(raw)-1]
		// Replace escaped quotes
		value := strings.ReplaceAll(inner, "''", "'")
		return spanned(p, &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    true,
			IsLargeObject: false,
			Value:         value,
		}, astStart), nil
	}

	return spanned(p, &ast.StringLiteral{
		LiteralType:   "String",
		IsNational:    true,
		IsLargeObject: false,
		Value:         raw,
	}, astStart), nil
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
	astStart := p.curTok

	// Check for graph pseudo columns at the start
	upper := strings.ToUpper(p.curTok.Literal)
	pseudoType := getPseudoColumnType(upper)
	if pseudoType != "" && p.peekTok.Type != TokenDot {
		p.nextToken()
		return spanned(p, &ast.ColumnReferenceExpression{ColumnType: pseudoType}, astStart), nil
	}

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
		} else if pseudoType := getPseudoColumnType(upper); pseudoType != "" {
			// Pseudo columns like $ROWGUID, $IDENTITY at end of multi-part identifier
			// set column type and are not included in the identifier list
			colType = pseudoType
			p.nextToken()
			break
		}

		id := p.spanIdent(literal, quoteType)
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
			identifiers = append(identifiers, p.spanIdent("", "NotQuoted"))
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

			return spanned(p, pfc, astStart), nil
		}
	}

	// Check for :: (user-defined type method call or property access): a.b::func() or a::prop
	if p.curTok.Type == TokenColonColon && len(identifiers) > 0 {
		colonTok := p.curTok
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
		name := p.spanIdent(nameValue, quoteType)
		p.nextToken()

		// Build SchemaObjectName from identifiers (filter out empty identifiers from leading dots)
		var nonEmptyIdents []*ast.Identifier
		for _, id := range identifiers {
			if id.Value != "" {
				nonEmptyIdents = append(nonEmptyIdents, id)
			}
		}
		schemaObjName := identifiersToSchemaObjectName(nonEmptyIdents)

		// ScriptDom spans the call target through the :: token.
		udtTarget := &ast.UserDefinedTypeCallTarget{SchemaObjectName: schemaObjName}
		if len(nonEmptyIdents) > 0 {
			p.spanChildToToken(udtTarget, nonEmptyIdents[0], colonTok)
		} else if len(identifiers) > 0 {
			p.spanChildToToken(udtTarget, identifiers[0], colonTok)
		}

		// If followed by ( it's a method call, otherwise property access
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (

			fc := &ast.FunctionCall{
				CallTarget:       udtTarget,
				FunctionName:     name,
				UniqueRowFilter:  "NotSpecified",
				WithArrayWrapper: false,
			}
			fcStartTarget := udtTarget

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

			p.spanFromChild(fc, fcStartTarget)
			fc.Pin()

			// Check for OVER clause or property access after method call
			spanV25, spanErr25 := p.parsePostExpressionAccess(fc)
			return spanned(p, spanV25, astStart), spanErr25
		}

		// Property access: t::a
		propAccess := &ast.UserDefinedTypePropertyAccess{
			CallTarget:   udtTarget,
			PropertyName: name,
		}
		p.spanFromChild(propAccess, udtTarget)
		propAccess.Pin()

		// Check for COLLATE clause
		if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
			p.nextToken() // consume COLLATE
			propAccess.Collation = p.parseIdentifier()
			p.spanFromChild(propAccess, udtTarget)
		}

		// Check for chained property access
		spanV26, spanErr26 := p.parsePostExpressionAccess(propAccess)
		return spanned(p, spanV26, astStart), spanErr26
	}

	// If followed by ( it's a function call
	if p.curTok.Type == TokenLParen {
		spanV27, spanErr27 := p.parseFunctionCallFromIdentifiers(identifiers)
		return spanned(p, spanV27, astStart), spanErr27
	}

	// If we have identifiers, build a column reference with them
	if len(identifiers) > 0 {
		return spanned(p, &ast.ColumnReferenceExpression{
			ColumnType: colType,
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Count:       len(identifiers),
				Identifiers: identifiers,
			},
		}, astStart), nil
	}

	// No identifiers means just IDENTITYCOL or ROWGUIDCOL (already handled in parsePrimaryExpression)
	// but handle the case anyway
	return spanned(p, &ast.ColumnReferenceExpression{
		ColumnType: colType,
	}, astStart), nil
}

func (p *Parser) parseColumnReference() (*ast.ColumnReferenceExpression, error) {
	astStart := p.curTok

	var expr ast.ScalarExpression
	var err error

	// Handle leading dots (like .st.StandardCost)
	if p.curTok.Type == TokenDot {
		expr, err = p.parseColumnReferenceWithLeadingDots()
	} else {
		expr, err = p.parseColumnReferenceOrFunctionCall()
	}
	if err != nil {
		return nil, err
	}
	if colRef, ok := expr.(*ast.ColumnReferenceExpression); ok {
		return spanned(p, colRef, astStart), nil
	}
	// If we got a function call, wrap it in a column reference (shouldn't happen in this context)
	return nil, fmt.Errorf("expected column reference, got function call")
}

func (p *Parser) parseColumnReferenceWithLeadingDots() (ast.ScalarExpression, error) {
	astStart := p.curTok

	// Handle multi-part identifiers starting with dots like ..t1.c1 or .db..t1.c1
	var identifiers []*ast.Identifier

	// Add empty identifiers for leading dots
	for p.curTok.Type == TokenDot {
		identifiers = append(identifiers, p.spanIdent("", "NotQuoted"))
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
			return spanned(p, &ast.ColumnReferenceExpression{
				ColumnType: colType,
				MultiPartIdentifier: &ast.MultiPartIdentifier{
					Count:       len(identifiers),
					Identifiers: identifiers,
				},
			}, astStart), nil
		}
		// Handle bracketed identifiers
		if len(literal) >= 2 && literal[0] == '[' && literal[len(literal)-1] == ']' {
			quoteType = "SquareBracket"
			literal = literal[1 : len(literal)-1]
		}

		id := p.spanIdent(literal, quoteType)
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
		colonTok := p.curTok
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
		name := p.spanIdent(nameValue, quoteType)
		p.nextToken()

		// Build SchemaObjectName from identifiers (filter out empty identifiers from leading dots)
		var nonEmptyIdents []*ast.Identifier
		for _, id := range identifiers {
			if id.Value != "" {
				nonEmptyIdents = append(nonEmptyIdents, id)
			}
		}
		schemaObjName := identifiersToSchemaObjectName(nonEmptyIdents)

		// ScriptDom spans the call target through the :: token.
		udtTarget := &ast.UserDefinedTypeCallTarget{SchemaObjectName: schemaObjName}
		if len(nonEmptyIdents) > 0 {
			p.spanChildToToken(udtTarget, nonEmptyIdents[0], colonTok)
		} else if len(identifiers) > 0 {
			p.spanChildToToken(udtTarget, identifiers[0], colonTok)
		}

		// If followed by ( it's a method call, otherwise property access
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (

			fc := &ast.FunctionCall{
				CallTarget:       udtTarget,
				FunctionName:     name,
				UniqueRowFilter:  "NotSpecified",
				WithArrayWrapper: false,
			}
			fcStartTarget := udtTarget

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

			p.spanFromChild(fc, fcStartTarget)
			fc.Pin()

			// Check for OVER clause or property access after method call
			spanV28, spanErr28 := p.parsePostExpressionAccess(fc)
			return spanned(p, spanV28, astStart), spanErr28
		}

		// Property access: .t::a
		propAccess := &ast.UserDefinedTypePropertyAccess{
			CallTarget:   udtTarget,
			PropertyName: name,
		}
		p.spanFromChild(propAccess, udtTarget)
		propAccess.Pin()

		// Check for COLLATE clause
		if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
			p.nextToken() // consume COLLATE
			propAccess.Collation = p.parseIdentifier()
			p.spanFromChild(propAccess, udtTarget)
		}

		// Check for chained property access
		spanV29, spanErr29 := p.parsePostExpressionAccess(propAccess)
		return spanned(p, spanV29, astStart), spanErr29
	}

	// Check if this is a function call
	if p.curTok.Type == TokenLParen && len(identifiers) > 1 {
		spanV30, spanErr30 := p.parseFunctionCallFromIdentifiers(identifiers)
		return spanned(p, spanV30, astStart), spanErr30
	}

	return spanned(p, &ast.ColumnReferenceExpression{
		ColumnType: "Regular",
		MultiPartIdentifier: &ast.MultiPartIdentifier{
			Count:       len(identifiers),
			Identifiers: identifiers,
		},
	}, astStart), nil
}

func (p *Parser) parseFunctionCallFromIdentifiers(identifiers []*ast.Identifier) (ast.ScalarExpression, error) {
	astStart := p.curTok

	// Check for special functions that need custom handling
	if len(identifiers) == 1 {
		// copyNameSpan gives the synthesized function-name identifier the
		// span of the original name token.
		copyNameSpan := func(v any) {
			fc, ok := v.(*ast.FunctionCall)
			if ok && fc != nil && fc.FunctionName != nil && identifiers[0].Frag().HasSpan() {
				f := identifiers[0].Frag()
				fc.FunctionName.SetSpan(f.StartOffset, f.FragmentLength, f.StartLine, f.StartColumn)
			}
		}
		funcName := strings.ToUpper(identifiers[0].Value)
		switch funcName {
		case "IIF":
			spanV31, spanErr31 := p.parseIIfCall()
			return spanned(p, spanV31, astStart), spanErr31
		case "PARSE":
			spanV32, spanErr32 := p.parseParseCall(false)
			copyNameSpan(spanV32)
			return spanned(p, spanV32, astStart), spanErr32
		case "TRY_PARSE":
			spanV33, spanErr33 := p.parseParseCall(true)
			copyNameSpan(spanV33)
			return spanned(p, spanV33, astStart), spanErr33
		case "JSON_OBJECT":
			spanV34, spanErr34 := p.parseJsonObjectCall()
			copyNameSpan(spanV34)
			return spanned(p, spanV34, astStart), spanErr34
		case "JSON_ARRAY":
			spanV35, spanErr35 := p.parseJsonArrayCall()
			copyNameSpan(spanV35)
			return spanned(p, spanV35, astStart), spanErr35
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

	// Span the call from its first identifier through the closing paren so
	// wrappers built by parsePostExpressionAccess see the full extent.
	if len(identifiers) > 0 {
		p.spanFromChild(fc, identifiers[0])
	}

	// Check for OVER clause or property access after function call
	spanV36, spanErr36 := p.parsePostExpressionAccess(fc)
	return spanned(p, spanV36, astStart), spanErr36
}

// parsePostExpressionAccess handles chained property access (.PropertyName), COLLATE clauses, and OVER clauses
// after an expression (function call, parenthesized expression, or property access).
func (p *Parser) parsePostExpressionAccess(expr ast.ScalarExpression) (ast.ScalarExpression, error) {
	astStart := p.curTok

	// Loop to handle chained property access like .SomeProperty.AnotherProperty
	for {
		// Check for .PropertyName pattern (property access)
		if p.curTok.Type == TokenDot {
			p.nextToken() // consume .

			if p.curTok.Type != TokenIdent {
				return nil, fmt.Errorf("expected property name after ., got %s", p.curTok.Literal)
			}
			propName := p.spanIdent(p.curTok.Literal, "NotQuoted")
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

				p.spanFromChild(fc, expr)
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

			p.spanFromChild(propAccess, expr)
			expr = propAccess
			continue
		}

		// Check for WITHIN GROUP clause for function calls (e.g., PERCENTILE_CONT)
		if fc, ok := expr.(*ast.FunctionCall); ok && strings.ToUpper(p.curTok.Literal) == "WITHIN" {
			withinTok := p.curTok
			p.nextToken() // consume WITHIN
			if strings.ToUpper(p.curTok.Literal) == "GROUP" {
				p.nextToken() // consume GROUP
			}

			if p.curTok.Type != TokenLParen {
				return nil, fmt.Errorf("expected ( after WITHIN GROUP, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume (

			// Parse ORDER BY clause or GRAPH PATH
			withinGroup := &ast.WithinGroupClause{
				HasGraphPath: false,
			}

			// Check for GRAPH PATH (case insensitive)
			if strings.ToUpper(p.curTok.Literal) == "GRAPH" && strings.ToUpper(p.peekTok.Literal) == "PATH" {
				withinGroup.HasGraphPath = true
				p.nextToken() // consume GRAPH
				p.nextToken() // consume PATH
			} else if p.curTok.Type == TokenOrder {
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

			p.spanFrom(withinTok, withinGroup)
			fc.WithinGroupClause = withinGroup
			continue // continue to check for more clauses like OVER
		}

		// Check for RESPECT NULLS or IGNORE NULLS for window functions
		if fc, ok := expr.(*ast.FunctionCall); ok {
			upperLit := strings.ToUpper(p.curTok.Literal)
			if upperLit == "RESPECT" || upperLit == "IGNORE" {
				// Parse RESPECT NULLS or IGNORE NULLS
				firstIdent := &ast.Identifier{Value: strings.ToUpper(p.curTok.Literal), QuoteType: "NotQuoted"}
				p.nextToken() // consume RESPECT/IGNORE

				if strings.ToUpper(p.curTok.Literal) == "NULLS" {
					secondIdent := &ast.Identifier{Value: strings.ToUpper(p.curTok.Literal), QuoteType: "NotQuoted"}
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

	return spanned(p, expr, astStart), nil
}

func (p *Parser) parseFromClause() (*ast.FromClause, error) {
	astStart := p.curTok

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
			return spanned(p, fc, astStart), nil
		}
		fc.TableReferences = append(fc.TableReferences, ref)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return spanned(p, fc, astStart), nil
}

func (p *Parser) parseTableReference() (ast.TableReference, error) {
	astStart := p.curTok

	// Parse the base table reference
	baseRef, err := p.parseSingleTableReference()
	if err != nil {
		return nil, err
	}
	var left ast.TableReference = baseRef

	// Check for JOINs and PIVOT/UNPIVOT (which can appear after table refs and joins)
	for {
		// Check for PIVOT or UNPIVOT that applies to current left
		if strings.ToUpper(p.curTok.Literal) == "PIVOT" {
			pivoted, err := p.parsePivotedTableReference(left)
			if err != nil {
				return nil, err
			}
			left = pivoted
			continue
		} else if strings.ToUpper(p.curTok.Literal) == "UNPIVOT" {
			unpivoted, err := p.parseUnpivotedTableReference(left)
			if err != nil {
				return nil, err
			}
			left = unpivoted
			continue
		}

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

		// Check for LOCAL modifier (undocumented feature) and join hints
		// Syntax: INNER LOCAL MERGE JOIN - LOCAL is just skipped
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "LOCAL" {
			p.nextToken() // skip LOCAL
		}

		// Check for join hints (REMOTE, LOOP, HASH, MERGE, REDUCE, REPLICATE, REDISTRIBUTE)
		joinHint := ""
		if p.curTok.Type == TokenIdent {
			upper := strings.ToUpper(p.curTok.Literal)
			switch upper {
			case "REMOTE", "LOOP", "HASH", "MERGE", "REDUCE", "REPLICATE", "REDISTRIBUTE":
				joinHint = upper[:1] + strings.ToLower(upper[1:]) // "REMOTE" -> "Remote"
				p.nextToken()
			}
		}

		if p.curTok.Type != TokenJoin {
			return nil, fmt.Errorf("expected JOIN, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume JOIN

		right, err := p.parseSingleTableReference()
		if err != nil {
			return nil, err
		}

		// Check for nested join - if we see another join type instead of ON,
		// the right side is actually a join expression
		for p.isJoinKeyword() {
			nestedJoinType, nestedJoinHint := p.parseJoinTypeAndHint()
			if nestedJoinType == "" {
				break
			}

			if p.curTok.Type != TokenJoin {
				return nil, fmt.Errorf("expected JOIN, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume JOIN

			nestedRight, err := p.parseSingleTableReference()
			if err != nil {
				return nil, err
			}

			// Parse ON clause for nested join
			if p.curTok.Type != TokenOn {
				return nil, fmt.Errorf("expected ON after nested JOIN, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume ON

			nestedCondition, err := p.parseBooleanExpression()
			if err != nil {
				return nil, err
			}

			// Wrap right in a QualifiedJoin
			right = &ast.QualifiedJoin{
				QualifiedJoinType:    nestedJoinType,
				JoinHint:             nestedJoinHint,
				FirstTableReference:  right,
				SecondTableReference: nestedRight,
				SearchCondition:      nestedCondition,
			}
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
			JoinHint:             joinHint,
			FirstTableReference:  left,
			SecondTableReference: right,
			SearchCondition:      condition,
		}
	}

	return spanned(p, left, astStart), nil
}

// isJoinKeyword returns true if the current token starts a join clause
func (p *Parser) isJoinKeyword() bool {
	switch p.curTok.Type {
	case TokenInner, TokenLeft, TokenRight, TokenFull, TokenJoin:
		return true
	default:
		return false
	}
}

// parseJoinTypeAndHint parses the join type (INNER, LEFT OUTER, etc.) and optional hint (REMOTE, LOOP, etc.)
// Returns empty string for joinType if no join is found
func (p *Parser) parseJoinTypeAndHint() (joinType, joinHint string) {
	switch p.curTok.Type {
	case TokenInner:
		joinType = "Inner"
		p.nextToken()
	case TokenLeft:
		joinType = "LeftOuter"
		p.nextToken()
		if p.curTok.Type == TokenOuter {
			p.nextToken()
		}
	case TokenRight:
		joinType = "RightOuter"
		p.nextToken()
		if p.curTok.Type == TokenOuter {
			p.nextToken()
		}
	case TokenFull:
		joinType = "FullOuter"
		p.nextToken()
		if p.curTok.Type == TokenOuter {
			p.nextToken()
		}
	case TokenJoin:
		joinType = "Inner"
	default:
		return "", ""
	}

	// Check for LOCAL modifier (undocumented feature) and join hints
	// Syntax: INNER LOCAL MERGE JOIN - LOCAL is just skipped
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "LOCAL" {
		p.nextToken() // skip LOCAL
	}

	// Check for join hints (REMOTE, LOOP, HASH, MERGE, REDUCE, REPLICATE, REDISTRIBUTE)
	if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		switch upper {
		case "REMOTE", "LOOP", "HASH", "MERGE", "REDUCE", "REPLICATE", "REDISTRIBUTE":
			joinHint = upper[:1] + strings.ToLower(upper[1:])
			p.nextToken()
		}
	}

	return joinType, joinHint
}

func (p *Parser) parseSingleTableReference() (ast.TableReference, error) {
	astStart := p.curTok

	// Check for derived table (parenthesized query)
	if p.curTok.Type == TokenLParen {
		spanV37, spanErr37 := p.parseDerivedTableReference()
		return spanned(p, spanV37, astStart), spanErr37
	}

	// Check for ODBC outer join escape sequence: { OJ ... }
	if p.curTok.Type == TokenLBrace {
		spanV38, spanErr38 := p.parseOdbcQualifiedJoinTableReference()
		return spanned(p, spanV38, astStart), spanErr38
	}

	// Check for built-in function table reference (::fn_name(...))
	if p.curTok.Type == TokenColonColon {
		spanV39, spanErr39 := p.parseBuiltInFunctionTableReference()
		return spanned(p, spanV39, astStart), spanErr39
	}

	// Check for OPENROWSET
	if p.curTok.Type == TokenOpenRowset {
		spanV40, spanErr40 := p.parseOpenRowset()
		return spanned(p, spanV40, astStart), spanErr40
	}

	// Check for OPENDATASOURCE
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OPENDATASOURCE" {
		spanV41, spanErr41 := p.parseAdHocTableReference()
		return spanned(p, spanV41, astStart), spanErr41
	}

	// Check for PREDICT
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PREDICT" {
		spanV42, spanErr42 := p.parsePredictTableReference()
		return spanned(p, spanV42, astStart), spanErr42
	}

	// Check for CHANGETABLE
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "CHANGETABLE" {
		spanV43, spanErr43 := p.parseChangeTableReference()
		return spanned(p, spanV43, astStart), spanErr43
	}

	// Check for OPENXML
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OPENXML" {
		spanV44, spanErr44 := p.parseOpenXmlTableReference()
		return spanned(p, spanV44, astStart), spanErr44
	}

	// Check for OPENQUERY
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OPENQUERY" {
		spanV45, spanErr45 := p.parseOpenQueryTableReference()
		return spanned(p, spanV45, astStart), spanErr45
	}

	// Check for full-text table functions (CONTAINSTABLE, FREETEXTTABLE)
	if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper == "CONTAINSTABLE" || upper == "FREETEXTTABLE" {
			spanV46, spanErr46 := p.parseFullTextTableReference(upper)
			return spanned(p, spanV46, astStart), spanErr46
		}
		// Check for semantic table functions
		if upper == "SEMANTICKEYPHRASETABLE" || upper == "SEMANTICSIMILARITYTABLE" || upper == "SEMANTICSIMILARITYDETAILSTABLE" {
			spanV47, spanErr47 := p.parseSemanticTableReference(upper)
			return spanned(p, spanV47, astStart), spanErr47
		}
	}

	// Check for variable table reference or variable method call
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		nameTok := p.curTok
		p.nextToken()

		// Check for method call: @var.method(...)
		if p.curTok.Type == TokenDot {
			p.nextToken() // consume dot
			methodName := p.parseIdentifier()
			if p.curTok.Type != TokenLParen {
				return nil, fmt.Errorf("expected ( after variable method name")
			}
			params, err := p.parseFunctionParameters()
			if err != nil {
				return nil, err
			}

			// Parse optional alias and column list
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
				for {
					col := p.parseIdentifier()
					columns = append(columns, col)
					if p.curTok.Type != TokenComma {
						break
					}
					p.nextToken() // consume comma
				}
				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ) after column list")
				}
				p.nextToken() // consume )
			}

			return spanned(p, &ast.VariableMethodCallTableReference{
				Variable:   p.varRefFromToken(nameTok),
				MethodName: methodName,
				Parameters: params,
				Alias:      alias,
				Columns:    columns,
				ForPath:    false,
			}, astStart), nil
		}

		// Parse optional alias for variable table reference
		varRef := &ast.VariableTableReference{
			Variable: p.varRefFromToken(nameTok),
			ForPath:  false,
		}
		if p.curTok.Type == TokenAs {
			p.nextToken()
			varRef.Alias = p.parseIdentifier()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			if p.curTok.Type == TokenIdent {
				upper := strings.ToUpper(p.curTok.Literal)
				if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" &&
					upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
					upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
					upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
					varRef.Alias = p.parseIdentifier()
				}
			} else {
				varRef.Alias = p.parseIdentifier()
			}
		}
		return spanned(p, varRef, astStart), nil
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
				return spanned(p, &ast.GlobalFunctionTableReference{
					Name:       son.BaseIdentifier,
					Parameters: params,
					Alias:      alias,
					Columns:    columns,
					ForPath:    false,
				}, astStart), nil
			}
			// Handle OPENJSON specially
			if upper == "OPENJSON" {
				spanV48, spanErr48 := p.parseOpenJsonTableReference(params, alias)
				return spanned(p, spanV48, astStart), spanErr48
			}
		}

		ref := &ast.SchemaObjectFunctionTableReference{
			SchemaObject: son,
			Parameters:   params,
			Alias:        alias,
			Columns:      columns,
			ForPath:      false,
		}
		return spanned(p, ref, astStart), nil
	}

	// It's a regular named table reference
	spanV49, spanErr49 := p.parseNamedTableReferenceWithName(son)
	return spanned(p, spanV49, astStart), spanErr49
}

// parseOpenJsonTableReference parses OPENJSON function with optional WITH clause
func (p *Parser) parseOpenJsonTableReference(params []ast.ScalarExpression, alias *ast.Identifier) (ast.TableReference, error) {
	astStart := p.curTok

	ref := &ast.OpenJsonTableReference{
		ForPath: false,
		Alias:   alias,
	}

	// First parameter is the Variable (JSON expression)
	if len(params) > 0 {
		ref.Variable = params[0]
	}

	// Second parameter is the RowPattern (optional path expression)
	if len(params) > 1 {
		ref.RowPattern = params[1]
	}

	// Check for WITH clause (schema declaration)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after OPENJSON WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		// Parse schema declaration items
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			item, err := p.parseSchemaDeclarationItemOpenjson()
			if err != nil {
				return nil, err
			}
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

	// Parse optional alias after WITH clause
	if ref.Alias == nil {
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
	}

	return spanned(p, ref, astStart), nil
}

// parseSchemaDeclarationItemOpenjson parses a column definition in OPENJSON WITH clause
func (p *Parser) parseSchemaDeclarationItemOpenjson() (*ast.SchemaDeclarationItemOpenjson, error) {
	astStart := p.curTok

	item := &ast.SchemaDeclarationItemOpenjson{
		ColumnDefinition: &ast.ColumnDefinitionBase{},
	}

	// Parse column name
	item.ColumnDefinition.ColumnIdentifier = p.parseIdentifier()

	// Parse data type
	dataType, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}
	item.ColumnDefinition.DataType = dataType

	// Parse optional COLLATE
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		item.ColumnDefinition.Collation = p.parseIdentifier()
	}
	// ScriptDom spans the column definition over name, type and collation.
	p.spanFrom(astStart, item.ColumnDefinition)

	// Parse optional path mapping (string literal) or AS JSON
	if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
		mapping, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		item.Mapping = mapping
	}

	// ScriptDom's item span excludes a trailing AS JSON.
	p.spanFrom(astStart, item)
	item.Pin()

	// Parse optional AS JSON
	if p.curTok.Type == TokenAs {
		p.nextToken() // consume AS
		if strings.ToUpper(p.curTok.Literal) == "JSON" {
			item.AsJson = true
			p.nextToken() // consume JSON
		}
	}

	return item, nil
}

// parseOdbcQualifiedJoinTableReference parses ODBC outer join escape sequence: { OJ ... }
func (p *Parser) parseOdbcQualifiedJoinTableReference() (ast.TableReference, error) {
	astStart := p.curTok

	p.nextToken() // consume {

	// Expect OJ keyword
	if strings.ToUpper(p.curTok.Literal) != "OJ" {
		return nil, fmt.Errorf("expected OJ after {, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume OJ

	// Parse the inner table reference (which can be a join)
	innerRef, err := p.parseTableReference()
	if err != nil {
		return nil, err
	}

	// Expect closing brace
	if p.curTok.Type != TokenRBrace {
		return nil, fmt.Errorf("expected } in ODBC outer join, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume }

	return spanned(p, &ast.OdbcQualifiedJoinTableReference{
		TableReference: innerRef,
	}, astStart), nil
}

// parseDerivedTableReference parses a derived table (parenthesized query) like (SELECT ...) AS alias
// or an inline derived table (VALUES clause) like (VALUES (...), (...)) AS alias(cols)
// or a data modification table reference (DML with OUTPUT) like (INSERT ... OUTPUT ...) AS alias
func (p *Parser) parseDerivedTableReference() (ast.TableReference, error) {
	astStart := p.curTok

	p.nextToken() // consume (

	// Check for VALUES clause (inline derived table)
	if strings.ToUpper(p.curTok.Literal) == "VALUES" {
		spanV50, spanErr50 := p.parseInlineDerivedTable()
		return spanned(p, spanV50, astStart), spanErr50
	}

	// Check for DML statements (INSERT, UPDATE, DELETE, MERGE) as table sources
	if p.curTok.Type == TokenInsert {
		spanV51, spanErr51 := p.parseDataModificationTableReference("INSERT")
		return spanned(p, spanV51, astStart), spanErr51
	}
	if p.curTok.Type == TokenUpdate {
		spanV52, spanErr52 := p.parseDataModificationTableReference("UPDATE")
		return spanned(p, spanV52, astStart), spanErr52
	}
	if p.curTok.Type == TokenDelete {
		spanV53, spanErr53 := p.parseDataModificationTableReference("DELETE")
		return spanned(p, spanV53, astStart), spanErr53
	}
	if strings.ToUpper(p.curTok.Literal) == "MERGE" {
		spanV54, spanErr54 := p.parseDataModificationTableReference("MERGE")
		return spanned(p, spanV54, astStart), spanErr54
	}

	// Check if this is a query (starts with SELECT, WITH, or another parenthesis for nested query)
	// or a parenthesized table reference (e.g., (t1 JOIN t2 ON ...))
	if p.curTok.Type != TokenSelect && p.curTok.Type != TokenWith && p.curTok.Type != TokenLParen {
		// This is a parenthesized table reference (e.g., (t1 JOIN t2 ON ...))
		tableRef, err := p.parseTableReference()
		if err != nil {
			return nil, err
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after parenthesized table reference, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
		return spanned(p, &ast.JoinParenthesisTableReference{
			Join:    tableRef,
			ForPath: false,
		}, astStart), nil
	}

	// Handle nested parenthesis specially
	// This could be:
	// 1. Query parenthesis: ((SELECT ... UNION ...)) - nested query expression
	// 2. Join parenthesis: ((SELECT ...) AS t1 JOIN ...) - nested derived table with joins
	if p.curTok.Type == TokenLParen {
		// Recursively parse the nested content as a derived table reference
		innerRef, err := p.parseDerivedTableReference()
		if err != nil {
			return nil, err
		}

		// Check what we got and what follows
		switch ref := innerRef.(type) {
		case *ast.QueryDerivedTable:
			// If no alias and we're at ) or binary query operator, this is a query parenthesis
			if ref.Alias == nil && (p.curTok.Type == TokenRParen ||
				p.curTok.Type == TokenUnion || p.curTok.Type == TokenExcept || p.curTok.Type == TokenIntersect) {
				// Convert to QueryParenthesisExpression and continue with query expression parsing
				qe := &ast.QueryParenthesisExpression{QueryExpression: ref.QueryExpression}
				// The parenthesis expression covers the same source extent
				// as the derived table it was parsed as.
				if ref.HasSpan() {
					qe.SetSpan(ref.StartOffset, ref.FragmentLength, ref.StartLine, ref.StartColumn)
					qe.Pin()
				}

				// Check for binary operations (UNION, EXCEPT, INTERSECT)
				if p.curTok.Type == TokenUnion || p.curTok.Type == TokenExcept || p.curTok.Type == TokenIntersect {
					qe2, err := p.parseRestOfBinaryQueryExpression(qe)
					if err != nil {
						return nil, err
					}
					// Now expect ) and return as QueryDerivedTable
					if p.curTok.Type != TokenRParen {
						return nil, fmt.Errorf("expected ) after binary query expression, got %s", p.curTok.Literal)
					}
					p.nextToken() // consume )

					result := &ast.QueryDerivedTable{
						QueryExpression: qe2,
						ForPath:         false,
					}

					// Parse optional alias
					if p.curTok.Type == TokenAs {
						p.nextToken()
						result.Alias = p.parseIdentifier()
					} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
						if p.curTok.Type == TokenIdent {
							upper := strings.ToUpper(p.curTok.Literal)
							if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" && upper != "USING" && upper != "WHEN" && upper != "OUTPUT" && upper != "PIVOT" && upper != "UNPIVOT" {
								result.Alias = p.parseIdentifier()
							}
						} else {
							result.Alias = p.parseIdentifier()
						}
					}

					return spanned(p, result, astStart), nil
				}

				// Just closing paren - expect ) and return as QueryDerivedTable
				p.nextToken() // consume )

				result := &ast.QueryDerivedTable{
					QueryExpression: qe,
					ForPath:         false,
				}

				// Parse optional alias
				if p.curTok.Type == TokenAs {
					p.nextToken()
					result.Alias = p.parseIdentifier()
				} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
					if p.curTok.Type == TokenIdent {
						upper := strings.ToUpper(p.curTok.Literal)
						if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" && upper != "USING" && upper != "WHEN" && upper != "OUTPUT" && upper != "PIVOT" && upper != "UNPIVOT" {
							result.Alias = p.parseIdentifier()
						}
					} else {
						result.Alias = p.parseIdentifier()
					}
				}

				return spanned(p, result, astStart), nil
			}

			// Otherwise, this is a derived table that may be followed by JOINs
			// Fall through to handle as table reference
			innerRef = ref

		case *ast.JoinParenthesisTableReference:
			// Already a join parenthesis - it may be followed by more JOINs or just )
		}

		// Handle as a table reference that may be followed by JOINs
		var tableRef ast.TableReference = innerRef
		for {
			// Check for CROSS JOIN / CROSS APPLY
			if p.curTok.Type == TokenCross {
				p.nextToken() // consume CROSS
				if p.curTok.Type == TokenJoin {
					p.nextToken() // consume JOIN
					right, err := p.parseSingleTableReference()
					if err != nil {
						return nil, err
					}
					tableRef = &ast.UnqualifiedJoin{
						UnqualifiedJoinType:  "CrossJoin",
						FirstTableReference:  tableRef,
						SecondTableReference: right,
					}
					continue
				} else if strings.ToUpper(p.curTok.Literal) == "APPLY" {
					p.nextToken() // consume APPLY
					right, err := p.parseSingleTableReference()
					if err != nil {
						return nil, err
					}
					tableRef = &ast.UnqualifiedJoin{
						UnqualifiedJoinType:  "CrossApply",
						FirstTableReference:  tableRef,
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
				tableRef = &ast.UnqualifiedJoin{
					UnqualifiedJoinType:  "OuterApply",
					FirstTableReference:  tableRef,
					SecondTableReference: right,
				}
				continue
			}

			// Check for qualified JOINs
			if p.isJoinKeyword() {
				joinType, joinHint := p.parseJoinTypeAndHint()
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

				// Check for nested join
				for p.isJoinKeyword() {
					nestedJoinType, nestedJoinHint := p.parseJoinTypeAndHint()
					if nestedJoinType == "" {
						break
					}
					if p.curTok.Type != TokenJoin {
						return nil, fmt.Errorf("expected JOIN, got %s", p.curTok.Literal)
					}
					p.nextToken() // consume JOIN

					nestedRight, err := p.parseSingleTableReference()
					if err != nil {
						return nil, err
					}

					if p.curTok.Type != TokenOn {
						return nil, fmt.Errorf("expected ON after nested JOIN, got %s", p.curTok.Literal)
					}
					p.nextToken() // consume ON

					nestedCondition, err := p.parseBooleanExpression()
					if err != nil {
						return nil, err
					}

					right = &ast.QualifiedJoin{
						QualifiedJoinType:    nestedJoinType,
						JoinHint:             nestedJoinHint,
						FirstTableReference:  right,
						SecondTableReference: nestedRight,
						SearchCondition:      nestedCondition,
					}
				}

				if p.curTok.Type != TokenOn {
					return nil, fmt.Errorf("expected ON after JOIN, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume ON

				condition, err := p.parseBooleanExpression()
				if err != nil {
					return nil, err
				}

				tableRef = &ast.QualifiedJoin{
					QualifiedJoinType:    joinType,
					JoinHint:             joinHint,
					FirstTableReference:  tableRef,
					SecondTableReference: right,
					SearchCondition:      condition,
				}
				continue
			}

			break
		}

		// Expect closing )
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after nested table reference, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		return spanned(p, &ast.JoinParenthesisTableReference{
			Join:    tableRef,
			ForPath: false,
		}, astStart), nil
	}

	// Parse the query expression (for SELECT or WITH)
	qe, err := p.parseQueryExpression()
	if err != nil {
		return nil, err
	}

	// Check if this is a nested derived table inside a parenthesized table reference
	// e.g., ((SELECT * FROM t1) AS t10 INNER JOIN t2 ON ...)
	// In this case, we're not at ) but at AS because the inner query expression
	// consumed its own parens and we need to build a derived table then continue with JOINs
	if p.curTok.Type != TokenRParen {
		// Build the inner derived table from the query expression
		innerRef := &ast.QueryDerivedTable{
			QueryExpression: qe,
			ForPath:         false,
		}

		// Parse alias for the inner derived table
		if p.curTok.Type == TokenAs {
			p.nextToken()
			innerRef.Alias = p.parseIdentifier()
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			if p.curTok.Type == TokenIdent {
				upper := strings.ToUpper(p.curTok.Literal)
				if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" && upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" && upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" && upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" && upper != "USING" && upper != "WHEN" && upper != "OUTPUT" && upper != "PIVOT" && upper != "UNPIVOT" {
					innerRef.Alias = p.parseIdentifier()
				}
			} else {
				innerRef.Alias = p.parseIdentifier()
			}
		}

		// Parse optional column list for inner derived table
		if innerRef.Alias != nil && p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for {
				col := p.parseIdentifier()
				innerRef.Columns = append(innerRef.Columns, col)
				if p.curTok.Type != TokenComma {
					break
				}
				p.nextToken() // consume comma
			}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ) after column list")
			}
			p.nextToken() // consume )
		}

		// Now parse any JOINs that follow
		var tableRef ast.TableReference = innerRef
		for {
			// Check for CROSS JOIN / CROSS APPLY
			if p.curTok.Type == TokenCross {
				p.nextToken() // consume CROSS
				if p.curTok.Type == TokenJoin {
					p.nextToken() // consume JOIN
					right, err := p.parseSingleTableReference()
					if err != nil {
						return nil, err
					}
					tableRef = &ast.UnqualifiedJoin{
						UnqualifiedJoinType:  "CrossJoin",
						FirstTableReference:  tableRef,
						SecondTableReference: right,
					}
					continue
				} else if strings.ToUpper(p.curTok.Literal) == "APPLY" {
					p.nextToken() // consume APPLY
					right, err := p.parseSingleTableReference()
					if err != nil {
						return nil, err
					}
					tableRef = &ast.UnqualifiedJoin{
						UnqualifiedJoinType:  "CrossApply",
						FirstTableReference:  tableRef,
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
				tableRef = &ast.UnqualifiedJoin{
					UnqualifiedJoinType:  "OuterApply",
					FirstTableReference:  tableRef,
					SecondTableReference: right,
				}
				continue
			}

			// Check for qualified JOINs
			if p.isJoinKeyword() {
				joinType, joinHint := p.parseJoinTypeAndHint()
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

				// Check for nested join
				for p.isJoinKeyword() {
					nestedJoinType, nestedJoinHint := p.parseJoinTypeAndHint()
					if nestedJoinType == "" {
						break
					}
					if p.curTok.Type != TokenJoin {
						return nil, fmt.Errorf("expected JOIN, got %s", p.curTok.Literal)
					}
					p.nextToken() // consume JOIN

					nestedRight, err := p.parseSingleTableReference()
					if err != nil {
						return nil, err
					}

					if p.curTok.Type != TokenOn {
						return nil, fmt.Errorf("expected ON after nested JOIN, got %s", p.curTok.Literal)
					}
					p.nextToken() // consume ON

					nestedCondition, err := p.parseBooleanExpression()
					if err != nil {
						return nil, err
					}

					right = &ast.QualifiedJoin{
						QualifiedJoinType:    nestedJoinType,
						JoinHint:             nestedJoinHint,
						FirstTableReference:  right,
						SecondTableReference: nestedRight,
						SearchCondition:      nestedCondition,
					}
				}

				if p.curTok.Type != TokenOn {
					return nil, fmt.Errorf("expected ON after JOIN, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume ON

				condition, err := p.parseBooleanExpression()
				if err != nil {
					return nil, err
				}

				tableRef = &ast.QualifiedJoin{
					QualifiedJoinType:    joinType,
					JoinHint:             joinHint,
					FirstTableReference:  tableRef,
					SecondTableReference: right,
					SearchCondition:      condition,
				}
				continue
			}

			break
		}

		// Expect closing ) for outer paren
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after parenthesized join expression, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		return spanned(p, &ast.JoinParenthesisTableReference{
			Join:    tableRef,
			ForPath: false,
		}, astStart), nil
	}

	p.nextToken() // consume )

	ref := &ast.QueryDerivedTable{
		QueryExpression: qe,
		ForPath:         false,
	}

	// Check for FOR PATH (graph path table reference)
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "FOR" && strings.ToUpper(p.peekTok.Literal) == "PATH" {
		p.nextToken() // consume FOR
		p.nextToken() // consume PATH
		ref.ForPath = true
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

	// Parse optional column list: alias(c1, c2, ...)
	if ref.Alias != nil && p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for {
			col := p.parseIdentifier()
			ref.Columns = append(ref.Columns, col)
			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume comma
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after column list")
		}
		p.nextToken() // consume )
	}

	return spanned(p, ref, astStart), nil
}

// parseDataModificationTableReference parses a DML statement used as a table source
// This is called after ( is consumed and the DML keyword is the current token
func (p *Parser) parseDataModificationTableReference(dmlType string) (*ast.DataModificationTableReference, error) {
	astStart := p.curTok

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

	return spanned(p, ref, astStart), nil
}

// parseInlineDerivedTable parses a VALUES clause used as a table source
// Called after ( is consumed and VALUES is the current token
func (p *Parser) parseInlineDerivedTable() (*ast.InlineDerivedTable, error) {
	astStart := p.curTok

	p.nextToken() // consume VALUES

	ref := &ast.InlineDerivedTable{
		ForPath: false,
	}

	// Parse row values: (val1, val2), (val3, val4), ...
	for {
		if p.curTok.Type != TokenLParen {
			break
		}
		rowTok := p.curTok
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
		p.spanFrom(rowTok, row)
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

	return spanned(p, ref, astStart), nil
}

func (p *Parser) parseNamedTableReference() (*ast.NamedTableReference, error) {
	astStart := p.curTok

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

	// Check for naked HOLDLOCK/NOWAIT before alias: table HOLDLOCK, table2
	if p.curTok.Type == TokenHoldlock {
		hlHint2 := &ast.TableHint{HintKind: "HoldLock"}
		p.tokSpan(hlHint2, p.curTok)
		ref.TableHints = append(ref.TableHints, hlHint2)
		p.nextToken()
	}
	if p.curTok.Type == TokenNowait {
		nwHint2 := &ast.TableHint{HintKind: "Nowait"}
		p.tokSpan(nwHint2, p.curTok)
		ref.TableHints = append(ref.TableHints, nwHint2)
		p.nextToken()
	}

	// Parse optional alias (AS alias or just alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			ref.Alias = p.parseIdentifier()
		} else {
			return nil, fmt.Errorf("expected identifier after AS, got %s", p.curTok.Literal)
		}
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

	// Check for old-style hints AFTER alias: table alias (1) or table alias (nolock)
	// peekIsOldStyleIndexHint is safe to use here since we're after the alias
	if p.curTok.Type == TokenLParen && (p.peekIsTableHint() || p.peekIsOldStyleIndexHint()) {
		lparenTok := p.curTok
		oldNumeric := !p.peekIsTableHint() && p.peekIsOldStyleIndexHint()
		nHintsBefore := len(ref.TableHints)
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
		// ScriptDom spans an old-style numeric index hint over the
		// enclosing parentheses.
		if oldNumeric && len(ref.TableHints) == nHintsBefore+1 {
			if h, ok := ref.TableHints[nHintsBefore].(spannable); ok {
				p.spanFrom(lparenTok, h)
			}
		}
	}

	// Check for naked HOLDLOCK/NOWAIT after alias: table alias HOLDLOCK
	if p.curTok.Type == TokenHoldlock {
		hlHint := &ast.TableHint{HintKind: "HoldLock"}
		p.tokSpan(hlHint, p.curTok)
		ref.TableHints = append(ref.TableHints, hlHint)
		p.nextToken()
	}
	if p.curTok.Type == TokenNowait {
		nwHint := &ast.TableHint{HintKind: "Nowait"}
		p.tokSpan(nwHint, p.curTok)
		ref.TableHints = append(ref.TableHints, nwHint)
		p.nextToken()
	}

	// Check for new-style hints (with WITH keyword): alias WITH (hints)
	if p.curTok.Type == TokenWith && p.peekTok.Type == TokenLParen {
		p.nextToken() // consume WITH
		// In WITH context, numbers are valid index hints: WITH (0)
		if p.curTok.Type == TokenLParen && (p.peekIsTableHint() || p.peekIsOldStyleIndexHint()) {
			lparenTok2 := p.curTok
			oldNumeric2 := !p.peekIsTableHint() && p.peekIsOldStyleIndexHint()
			nHintsBefore2 := len(ref.TableHints)
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
			// ScriptDom spans an old-style numeric index hint over the
			// enclosing parentheses.
			if oldNumeric2 && len(ref.TableHints) == nHintsBefore2+1 {
				if h, ok := ref.TableHints[nHintsBefore2].(spannable); ok {
					p.spanFrom(lparenTok2, h)
				}
			}
		}
	}

	return spanned(p, ref, astStart), nil
}

// parseNamedTableReferenceWithName parses a named table reference when the schema object name has already been parsed
func (p *Parser) parseNamedTableReferenceWithName(son *ast.SchemaObjectName) (*ast.NamedTableReference, error) {
	astStart := p.curTok

	ref := &ast.NamedTableReference{
		SchemaObject: son,
		ForPath:      false,
	}

	// Parse FOR SYSTEM_TIME clause (temporal tables)
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "FOR" && strings.ToUpper(p.peekTok.Literal) == "SYSTEM_TIME" {
		temporal, err := p.parseTemporalClause()
		if err != nil {
			return nil, err
		}
		ref.TemporalClause = temporal
	}

	// Parse FOR PATH clause (graph database path references)
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "FOR" && strings.ToUpper(p.peekTok.Literal) == "PATH" {
		p.nextToken() // consume FOR
		p.nextToken() // consume PATH
		ref.ForPath = true
	}

	// Check for TABLESAMPLE before alias
	if strings.ToUpper(p.curTok.Literal) == "TABLESAMPLE" {
		tableSample, err := p.parseTableSampleClause()
		if err != nil {
			return nil, err
		}
		ref.TableSampleClause = tableSample
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

	// Check for naked HOLDLOCK/NOWAIT before alias: table HOLDLOCK, table2
	if p.curTok.Type == TokenHoldlock {
		hlHint2 := &ast.TableHint{HintKind: "HoldLock"}
		p.tokSpan(hlHint2, p.curTok)
		ref.TableHints = append(ref.TableHints, hlHint2)
		p.nextToken()
	}
	if p.curTok.Type == TokenNowait {
		nwHint2 := &ast.TableHint{HintKind: "Nowait"}
		p.tokSpan(nwHint2, p.curTok)
		ref.TableHints = append(ref.TableHints, nwHint2)
		p.nextToken()
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

	// Check for old-style hints AFTER alias: table alias (1) or table alias (nolock)
	// peekIsOldStyleIndexHint is safe to use here since we're after the alias
	if p.curTok.Type == TokenLParen && (p.peekIsTableHint() || p.peekIsOldStyleIndexHint()) {
		lparenTok := p.curTok
		oldNumeric := !p.peekIsTableHint() && p.peekIsOldStyleIndexHint()
		nHintsBefore := len(ref.TableHints)
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
		// ScriptDom spans an old-style numeric index hint over the
		// enclosing parentheses.
		if oldNumeric && len(ref.TableHints) == nHintsBefore+1 {
			if h, ok := ref.TableHints[nHintsBefore].(spannable); ok {
				p.spanFrom(lparenTok, h)
			}
		}
	}

	// Check for naked HOLDLOCK/NOWAIT after alias: table alias HOLDLOCK
	if p.curTok.Type == TokenHoldlock {
		hlHint := &ast.TableHint{HintKind: "HoldLock"}
		p.tokSpan(hlHint, p.curTok)
		ref.TableHints = append(ref.TableHints, hlHint)
		p.nextToken()
	}
	if p.curTok.Type == TokenNowait {
		nwHint := &ast.TableHint{HintKind: "Nowait"}
		p.tokSpan(nwHint, p.curTok)
		ref.TableHints = append(ref.TableHints, nwHint)
		p.nextToken()
	}

	// Check for TABLESAMPLE after alias (supports syntax: t1 AS alias TABLESAMPLE (...))
	if ref.TableSampleClause == nil && strings.ToUpper(p.curTok.Literal) == "TABLESAMPLE" {
		tableSample, err := p.parseTableSampleClause()
		if err != nil {
			return nil, err
		}
		ref.TableSampleClause = tableSample
	}

	// Check for old-style hints after TABLESAMPLE (without WITH keyword): alias TABLESAMPLE (...)(nolock)
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

	// Check for new-style hints (with WITH keyword): alias WITH (hints)
	if p.curTok.Type == TokenWith && p.peekTok.Type == TokenLParen {
		p.nextToken() // consume WITH
		// In WITH context, numbers are valid index hints: WITH (0)
		if p.curTok.Type == TokenLParen && (p.peekIsTableHint() || p.peekIsOldStyleIndexHint()) {
			lparenTok2 := p.curTok
			oldNumeric2 := !p.peekIsTableHint() && p.peekIsOldStyleIndexHint()
			nHintsBefore2 := len(ref.TableHints)
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
			// ScriptDom spans an old-style numeric index hint over the
			// enclosing parentheses.
			if oldNumeric2 && len(ref.TableHints) == nHintsBefore2+1 {
				if h, ok := ref.TableHints[nHintsBefore2].(spannable); ok {
					p.spanFrom(lparenTok2, h)
				}
			}
		}
	}

	return spanned(p, ref, astStart), nil
}

// parseTemporalClause parses a FOR SYSTEM_TIME clause for temporal tables
func (p *Parser) parseTemporalClause() (*ast.TemporalClause, error) {
	astStart := p.curTok
	entryEnd := p.prevEndByte

	clause := &ast.TemporalClause{}

	p.nextToken() // consume FOR
	p.nextToken() // consume SYSTEM_TIME

	upper := strings.ToUpper(p.curTok.Literal)
	switch upper {
	case "AS":
		// AS OF <time>
		p.nextToken() // consume AS
		if strings.ToUpper(p.curTok.Literal) != "OF" {
			return nil, fmt.Errorf("expected OF after AS, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume OF
		clause.TemporalClauseType = "AsOf"
		startTime, err := p.parseTemporalTimeValue()
		if err != nil {
			return nil, err
		}
		clause.StartTime = startTime
		p.spanFromChild(clause, startTime)
		clause.Pin()

	case "BETWEEN":
		// BETWEEN <start> AND <end>
		p.nextToken() // consume BETWEEN
		clause.TemporalClauseType = "Between"
		startTime, err := p.parseTemporalTimeValue()
		if err != nil {
			return nil, err
		}
		clause.StartTime = startTime
		if p.curTok.Type != TokenAnd {
			return nil, fmt.Errorf("expected AND, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume AND
		endTime, err := p.parseTemporalTimeValue()
		if err != nil {
			return nil, err
		}
		clause.EndTime = endTime
		p.spanFromChild(clause, startTime)
		clause.Pin()

	case "FROM":
		// FROM <start> TO <end>
		p.nextToken() // consume FROM
		clause.TemporalClauseType = "FromTo"
		startTime, err := p.parseTemporalTimeValue()
		if err != nil {
			return nil, err
		}
		clause.StartTime = startTime
		if strings.ToUpper(p.curTok.Literal) != "TO" {
			return nil, fmt.Errorf("expected TO, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume TO
		endTime, err := p.parseTemporalTimeValue()
		if err != nil {
			return nil, err
		}
		clause.EndTime = endTime
		p.spanFromChild(clause, startTime)
		clause.Pin()

	case "CONTAINED":
		// CONTAINED IN (<start>, <end>)
		p.nextToken() // consume CONTAINED
		if strings.ToUpper(p.curTok.Literal) != "IN" {
			return nil, fmt.Errorf("expected IN after CONTAINED, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume IN
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after CONTAINED IN, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (
		clause.TemporalClauseType = "ContainedIn"
		startTime, err := p.parseTemporalTimeValue()
		if err != nil {
			return nil, err
		}
		clause.StartTime = startTime
		if p.curTok.Type != TokenComma {
			return nil, fmt.Errorf("expected comma, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ,
		endTime, err := p.parseTemporalTimeValue()
		if err != nil {
			return nil, err
		}
		clause.EndTime = endTime
		p.spanFromChild(clause, startTime)
		clause.Pin()
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		// ScriptDom ends the enclosing table reference, FROM clause and
		// query specification before this closing parenthesis, so do not
		// let it advance the recorded end position.
		savedEnd := p.prevEndByte
		p.nextToken() // consume )
		p.prevEndByte = savedEnd

	case "ALL":
		// ALL has no source position in ScriptDom, and the enclosing table
		// reference, FROM clause and query specification end before the
		// whole FOR SYSTEM_TIME ALL clause.
		p.nextToken() // consume ALL
		clause.TemporalClauseType = "TemporalAll"
		p.prevEndByte = entryEnd

	default:
		return nil, fmt.Errorf("unexpected temporal clause type: %s", p.curTok.Literal)
	}

	if clause.Pinned() {
		return clause, nil
	}
	if clause.TemporalClauseType == "TemporalAll" {
		return clause, nil
	}
	return spanned(p, clause, astStart), nil
}

// parseTemporalTimeValue parses a time value in a temporal clause (string literal or variable)
func (p *Parser) parseTemporalTimeValue() (ast.ScalarExpression, error) {
	astStart := p.curTok

	if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
		lit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		return spanned(p, lit, astStart), nil
	}
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		varRef := p.spanVarRef(p.curTok.Literal)
		p.nextToken()
		return spanned(p, varRef, astStart), nil
	}
	return nil, fmt.Errorf("expected string literal or variable for temporal time, got %s", p.curTok.Literal)
}

// parseFullTextTableReference parses CONTAINSTABLE or FREETEXTTABLE
func (p *Parser) parseFullTextTableReference(funcType string) (*ast.FullTextTableReference, error) {
	astStart := p.curTok

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

	return spanned(p, ref, astStart), nil
}

// parseSemanticTableReference parses SEMANTICKEYPHRASETABLE, SEMANTICSIMILARITYTABLE, or SEMANTICSIMILARITYDETAILSTABLE
func (p *Parser) parseSemanticTableReference(funcType string) (*ast.SemanticTableReference, error) {
	astStart := p.curTok

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
		wc := &ast.ColumnReferenceExpression{ColumnType: "Wildcard"}
		p.tokSpan(wc, p.curTok)
		ref.Columns = []*ast.ColumnReferenceExpression{wc}
		p.nextToken()
	} else if p.curTok.Type == TokenLParen {
		// Column list
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			if p.curTok.Type == TokenStar {
				wc := &ast.ColumnReferenceExpression{ColumnType: "Wildcard"}
				p.tokSpan(wc, p.curTok)
				ref.Columns = append(ref.Columns, wc)
				p.nextToken()
			} else {
				colTok := p.curTok
				col := p.parseIdentifier()
				colRef := &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{col},
						Count:       1,
					},
				}
				p.spanFrom(colTok, colRef)
				ref.Columns = append(ref.Columns, colRef)
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

	return spanned(p, ref, astStart), nil
}

// parseSimpleExpression parses a simple expression (including unary minus for negative numbers)
func (p *Parser) parseSimpleExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	if p.curTok.Type == TokenMinus {
		p.nextToken() // consume -
		expr, err := p.parsePrimaryExpression()
		if err != nil {
			return nil, err
		}
		return spanned(p, &ast.UnaryExpression{
			UnaryExpressionType: "Negative",
			Expression:          expr,
		}, astStart), nil
	}
	spanV55, spanErr55 := p.parsePrimaryExpression()
	return spanned(p, spanV55, astStart), spanErr55
}

// parseTableHint parses a single table hint
func (p *Parser) parseTableHint() (ast.TableHintType, error) {
	astStart := p.curTok

	// Handle old-style numeric index hint (just a number like "0" or "1")
	if p.curTok.Type == TokenNumber {
		hint := &ast.IndexTableHint{
			HintKind: "Index",
			IndexValues: []*ast.IdentifierOrValueExpression{
				{
					Value:           p.curTok.Literal,
					ValueExpression: p.intLitFromToken(p.curTok),
				},
			},
		}
		p.nextToken()
		// Check for additional comma-separated values
		for p.curTok.Type == TokenComma {
			p.nextToken()
			if p.curTok.Type == TokenNumber {
				hint.IndexValues = append(hint.IndexValues, &ast.IdentifierOrValueExpression{
					Value:           p.curTok.Literal,
					ValueExpression: p.intLitFromToken(p.curTok),
				})
				p.nextToken()
			} else if p.curTok.Type == TokenIdent {
				hint.IndexValues = append(hint.IndexValues, &ast.IdentifierOrValueExpression{
					Value:      p.curTok.Literal,
					Identifier: p.spanIdent(p.curTok.Literal, "NotQuoted"),
				})
				p.nextToken()
			} else {
				break
			}
		}
		return spanned(p, hint, astStart), nil
	}

	hintName := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume hint name

	// INDEX hint with values
	if hintName == "INDEX" {
		hint := &ast.IndexTableHint{
			HintKind: "Index",
		}
		// Handle INDEX = value syntax (alternative to INDEX(value))
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			var iov *ast.IdentifierOrValueExpression
			if p.curTok.Type == TokenNumber {
				iov = &ast.IdentifierOrValueExpression{
					Value:           p.curTok.Literal,
					ValueExpression: p.intLitFromToken(p.curTok),
				}
				p.nextToken()
			} else if p.curTok.Type == TokenIdent {
				iov = &ast.IdentifierOrValueExpression{
					Value:      p.curTok.Literal,
					Identifier: p.spanIdent(p.curTok.Literal, "NotQuoted"),
				}
				p.nextToken()
			}
			if iov != nil {
				hint.IndexValues = append(hint.IndexValues, iov)
			}
			return spanned(p, hint, astStart), nil
		}
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				var iov *ast.IdentifierOrValueExpression
				if p.curTok.Type == TokenNumber {
					iov = &ast.IdentifierOrValueExpression{
						Value:           p.curTok.Literal,
						ValueExpression: p.intLitFromToken(p.curTok),
					}
					p.nextToken()
				} else if p.curTok.Type == TokenIdent {
					iov = &ast.IdentifierOrValueExpression{
						Value:      p.curTok.Literal,
						Identifier: p.spanIdent(p.curTok.Literal, "NotQuoted"),
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
		return spanned(p, hint, astStart), nil
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
			hint.Value = p.intLitFromToken(p.curTok)
			p.nextToken()
		}
		return spanned(p, hint, astStart), nil
	}

	// FORCESEEK hint with optional index and column list
	if hintName == "FORCESEEK" {
		hint := &ast.ForceSeekTableHint{
			HintKind: "ForceSeek",
		}
		// Check for optional parenthesis with index and columns
		if p.curTok.Type != TokenLParen {
			return spanned(p, hint, astStart), nil
		}
		p.nextToken() // consume (
		// Parse index value (identifier or number)
		if p.curTok.Type == TokenNumber {
			hint.IndexValue = &ast.IdentifierOrValueExpression{
				Value:           p.curTok.Literal,
				ValueExpression: p.intLitFromToken(p.curTok),
			}
			p.nextToken()
		} else if p.curTok.Type == TokenIdent {
			hint.IndexValue = &ast.IdentifierOrValueExpression{
				Value:      p.curTok.Literal,
				Identifier: p.spanIdent(p.curTok.Literal, "NotQuoted"),
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
		return spanned(p, hint, astStart), nil
	}

	// Map hint names to HintKind
	hintKind := getTableHintKind(hintName)
	if hintKind == "" {
		return nil, nil // Unknown hint
	}

	return spanned(p, &ast.TableHint{
		HintKind: hintKind,
	}, astStart), nil
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
	case "READCOMMITTEDLOCK":
		return "ReadCommittedLock"
	case "KEEPIDENTITY":
		return "KeepIdentity"
	case "KEEPDEFAULTS":
		return "KeepDefaults"
	case "IGNORE_CONSTRAINTS":
		return "IgnoreConstraints"
	case "IGNORE_TRIGGERS":
		return "IgnoreTriggers"
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
		case "HOLDLOCK", "NOLOCK", "PAGLOCK", "READCOMMITTED", "READCOMMITTEDLOCK", "READPAST",
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
		case "HOLDLOCK", "NOLOCK", "PAGLOCK", "READCOMMITTED", "READCOMMITTEDLOCK", "READPAST",
			"READUNCOMMITTED", "REPEATABLEREAD", "ROWLOCK", "SERIALIZABLE",
			"SNAPSHOT", "TABLOCK", "TABLOCKX", "UPDLOCK", "XLOCK", "NOWAIT",
			"INDEX", "FORCESEEK", "FORCESCAN", "KEEPIDENTITY", "KEEPDEFAULTS",
			"IGNORE_CONSTRAINTS", "IGNORE_TRIGGERS", "NOEXPAND", "SPATIAL_WINDOW_MAX_CELLS":
			return true
		}
	}
	return false
}

// peekIsOldStyleIndexHint checks if the peek token is a number (for old-style index hint like (0))
// This is only valid after an alias or table name, not for function calls
func (p *Parser) peekIsOldStyleIndexHint() bool {
	return p.peekTok.Type == TokenNumber
}

func (p *Parser) parseSchemaObjectName() (*ast.SchemaObjectName, error) {
	astStart := p.curTok

	var identifiers []*ast.Identifier

	for {
		// Handle empty parts (e.g., myDb..table means myDb.<empty>.table)
		if p.curTok.Type == TokenDot {
			// Add an empty identifier for the missing part
			identifiers = append(identifiers, p.spanIdent("", "NotQuoted"))
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

	return spanned(p, son, astStart), nil
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
	astStart := p.curTok

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
			h := &ast.LiteralOptimizerHint{HintKind: "UsePlan", Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "HINT" {
			p.nextToken() // consume HINT
			return p.parseUseHintList()
		}
		return p.optHint("Use", astStart), nil
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
		return p.optHint(hintKind, astStart), nil
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
			h := &ast.LiteralOptimizerHint{HintKind: "MaxDop", Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		return p.optHint("MaxDop", astStart), nil
	}

	// Handle TABLE HINT optimizer hint
	if p.curTok.Type == TokenTable {
		p.nextToken() // consume TABLE
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "HINT" {
			p.nextToken() // consume HINT
			spanV57, spanErr57 := p.parseTableHintsOptimizerHint()
			return spanned(p, spanV57, astStart), spanErr57
		}
		return p.optHint("Table", astStart), nil
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
			h := &ast.LiteralOptimizerHint{HintKind: "Fast", Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		return p.optHint("Fast", astStart), nil
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
				return p.optHint("ParameterizationSimple", astStart), nil
			} else if subUpper == "FORCED" {
				return p.optHint("ParameterizationForced", astStart), nil
			}
		}
		return p.optHint("Parameterization", astStart), nil

	case "MAXRECURSION":
		p.nextToken() // consume MAXRECURSION
		value, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		h := &ast.LiteralOptimizerHint{HintKind: "MaxRecursion", Value: value}
		p.spanFromChild(h, value)
		return h, nil

	case "OPTIMIZE":
		p.nextToken() // consume OPTIMIZE
		if p.curTok.Type == TokenIdent {
			subUpper := strings.ToUpper(p.curTok.Literal)
			if subUpper == "FOR" {
				p.nextToken() // consume FOR
				return p.parseOptimizeForHint()
			} else if subUpper == "CORRELATED" {
				p.nextToken() // consume CORRELATED
				hintTok := astStart
				if strings.ToUpper(p.curTok.Literal) == "UNION" {
					// ScriptDom positions this hint on the UNION token.
					hintTok = p.curTok
					p.nextToken() // consume UNION
					if strings.ToUpper(p.curTok.Literal) == "ALL" {
						p.nextToken() // consume ALL
					}
				}
				return p.optHint("OptimizeCorrelatedUnionAll", hintTok), nil
			}
		}
		return p.optHint("Optimize", astStart), nil

	case "CHECKCONSTRAINTS":
		p.nextToken() // consume CHECKCONSTRAINTS
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PLAN" {
			p.nextToken() // consume PLAN
			return p.optHint("CheckConstraintsPlan", astStart), nil
		}
		return p.optHint("CheckConstraints", astStart), nil

	case "LABEL":
		p.nextToken() // consume LABEL
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			h := &ast.LiteralOptimizerHint{HintKind: "Label", Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		return p.optHint("Label", astStart), nil

	case "MAX_GRANT_PERCENT":
		p.nextToken() // consume MAX_GRANT_PERCENT
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			h := &ast.LiteralOptimizerHint{HintKind: "MaxGrantPercent", Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		return p.optHint("MaxGrantPercent", astStart), nil

	case "MIN_GRANT_PERCENT":
		p.nextToken() // consume MIN_GRANT_PERCENT
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			h := &ast.LiteralOptimizerHint{HintKind: "MinGrantPercent", Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		return p.optHint("MinGrantPercent", astStart), nil

	case "FAST":
		p.nextToken() // consume FAST
		// FAST can take a numeric argument
		if p.curTok.Type == TokenNumber {
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			h := &ast.LiteralOptimizerHint{HintKind: "Fast", Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		return p.optHint("Fast", astStart), nil

	case "NO_PERFORMANCE_SPOOL":
		p.nextToken() // consume NO_PERFORMANCE_SPOOL
		return p.optHint("NoPerformanceSpool", astStart), nil

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
			h := &ast.LiteralOptimizerHint{HintKind: hintKind, Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}

		// Check if this is a literal hint (LABEL = value, etc.)
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			value, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			h := &ast.LiteralOptimizerHint{HintKind: hintKind, Value: value}
			p.spanFromChild(h, value)
			return h, nil
		}
		return p.optHint(hintKind, astStart), nil
	}
}

func (p *Parser) parseUseHintList() (ast.OptimizerHintBase, error) {
	astStart := p.curTok

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

	return spanned(p, hint, astStart), nil
}

func (p *Parser) parseTableHintsOptimizerHint() (ast.OptimizerHintBase, error) {
	astStart := p.curTok

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

	return spanned(p, hint, astStart), nil
}

func (p *Parser) parseOptimizeForHint() (ast.OptimizerHintBase, error) {
	astStart := p.curTok

	hint := &ast.OptimizeForOptimizerHint{
		HintKind:     "OptimizeFor",
		IsForUnknown: false,
	}

	// Check for UNKNOWN
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "UNKNOWN" {
		p.nextToken()
		hint.IsForUnknown = true
		return spanned(p, hint, astStart), nil
	}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after OPTIMIZE FOR, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// ScriptDom spans this hint from the first pair through the closing
	// paren, excluding "OPTIMIZE FOR (".
	astStart = p.curTok

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

	return spanned(p, hint, astStart), nil
}

func (p *Parser) parseVariableValuePair() (*ast.VariableValuePair, error) {
	astStart := p.curTok

	// Expect @variable (variables are TokenIdent starting with @)
	if p.curTok.Type != TokenIdent || !strings.HasPrefix(p.curTok.Literal, "@") {
		return nil, nil
	}

	pair := &ast.VariableValuePair{
		Variable:     p.spanVarRef(p.curTok.Literal),
		IsForUnknown: false,
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		// Could be UNKNOWN
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "UNKNOWN" {
			p.nextToken()
			pair.IsForUnknown = true
			return spanned(p, pair, astStart), nil
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

	return spanned(p, pair, astStart), nil
}

// convertHintKind converts hint identifiers to their canonical names
func convertHintKind(hint string) string {
	// Map common hint names
	hintMap := map[string]string{
		"IGNORE_NONCLUSTERED_COLUMNSTORE_INDEX": "IgnoreNonClusteredColumnStoreIndex",
		"LABEL":                                 "Label",
		"MAX_GRANT_PERCENT":                     "MaxGrantPercent",
		"MIN_GRANT_PERCENT":                     "MinGrantPercent",
		"NO_PERFORMANCE_SPOOL":                  "NoPerformanceSpool",
		"PARAMETERIZATION":                      "Parameterization",
		"RECOMPILE":                             "Recompile",
		"MAXRECURSION":                          "MaxRecursion",
		"KEEPFIXED":                             "KeepFixed",
		"KEEP":                                  "Keep",
		"EXPAND":                                "Expand",
		"VIEWS":                                 "Views",
		"BYPASS":                                "Bypass",
		"OPTIMIZER_QUEUE":                       "OptimizerQueue",
		"USEPLAN":                               "UsePlan",
		"SHRINKDB":                              "ShrinkDB",
		"ALTERCOLUMN":                           "AlterColumn",
		"HASH":                                  "Hash",
		"ORDER":                                 "Order",
		"GROUP":                                 "Group",
		"MERGE":                                 "Merge",
		"CONCAT":                                "Concat",
		"UNION":                                 "Union",
		"LOOP":                                  "Loop",
		"JOIN":                                  "Join",
		"FAST":                                  "Fast",
		"FORCE":                                 "Force",
		"ROBUST":                                "Robust",
		"PLAN":                                  "Plan",
		"USE":                                   "Use",
		"SIMPLE":                                "Simple",
		"FORCED":                                "Forced",
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
	astStart := p.curTok

	// Consume WHERE
	p.nextToken()

	condition, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}

	return spanned(p, &ast.WhereClause{SearchCondition: condition}, astStart), nil
}

func (p *Parser) parseGroupByClause() (*ast.GroupByClause, error) {
	astStart := p.curTok

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

	return spanned(p, gbc, astStart), nil
}

// parseGroupingSpecification parses a single grouping specification
func (p *Parser) parseGroupingSpecification() (ast.GroupingSpecification, error) {
	astStart := p.curTok

	// Check for ROLLUP (...)
	if p.curTok.Type == TokenRollup {
		spanV59, spanErr59 := p.parseRollupGroupingSpecification()
		return spanned(p, spanV59, astStart), spanErr59
	}

	// Check for CUBE (...)
	if p.curTok.Type == TokenCube {
		spanV60, spanErr60 := p.parseCubeGroupingSpecification()
		return spanned(p, spanV60, astStart), spanErr60
	}

	// Check for GROUPING SETS (...)
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GROUPING" &&
		p.peekTok.Type == TokenIdent && strings.ToUpper(p.peekTok.Literal) == "SETS" {
		spanV61, spanErr61 := p.parseGroupingSetsGroupingSpecification()
		return spanned(p, spanV61, astStart), spanErr61
	}

	// Check for grand total () or composite grouping (c1, c2, ...)
	if p.curTok.Type == TokenLParen {
		// Check for empty parens () which is grand total
		if p.peekTok.Type == TokenRParen {
			p.nextToken() // consume (
			p.nextToken() // consume )
			return spanned(p, &ast.GrandTotalGroupingSpecification{}, astStart), nil
		}
		spanV62, spanErr62 := p.parseCompositeGroupingSpecification()
		return spanned(p, spanV62, astStart), spanErr62
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

	return spanned(p, spec, astStart), nil
}

// parseRollupGroupingSpecification parses ROLLUP (c1, c2, ...)
func (p *Parser) parseRollupGroupingSpecification() (*ast.RollupGroupingSpecification, error) {
	astStart := p.curTok

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

	return spanned(p, spec, astStart), nil
}

// parseCubeGroupingSpecification parses CUBE (c1, c2, ...)
func (p *Parser) parseCubeGroupingSpecification() (*ast.CubeGroupingSpecification, error) {
	astStart := p.curTok

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

	return spanned(p, spec, astStart), nil
}

// parseGroupingSetsGroupingSpecification parses GROUPING SETS (...)
func (p *Parser) parseGroupingSetsGroupingSpecification() (*ast.GroupingSetsGroupingSpecification, error) {
	astStart := p.curTok

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

	return spanned(p, spec, astStart), nil
}

// parseGroupingSetsArgument parses an argument inside GROUPING SETS which can be
// CUBE(...), ROLLUP(...), a column, or a parenthesized group
func (p *Parser) parseGroupingSetsArgument() (ast.GroupingSpecification, error) {
	astStart := p.curTok

	// Check for CUBE
	if p.curTok.Type == TokenCube {
		spanV63, spanErr63 := p.parseCubeGroupingSpecification()
		return spanned(p, spanV63, astStart), spanErr63
	}

	// Check for ROLLUP
	if p.curTok.Type == TokenRollup {
		spanV64, spanErr64 := p.parseRollupGroupingSpecification()
		return spanned(p, spanV64, astStart), spanErr64
	}

	// Check for parenthesized group
	if p.curTok.Type == TokenLParen {
		// Check for empty parens () which is grand total
		if p.peekTok.Type == TokenRParen {
			p.nextToken() // consume (
			p.nextToken() // consume )
			return spanned(p, &ast.GrandTotalGroupingSpecification{}, astStart), nil
		}
		spanV65, spanErr65 := p.parseGroupingSetsCompositeArgument()
		return spanned(p, spanV65, astStart), spanErr65
	}

	// Regular expression (column reference or literal)
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return spanned(p, &ast.ExpressionGroupingSpecification{
		Expression:             expr,
		DistributedAggregation: false,
	}, astStart), nil
}

// parseGroupingSetsCompositeArgument parses a parenthesized group inside GROUPING SETS
// which can contain CUBE, ROLLUP, columns, or a mix
func (p *Parser) parseGroupingSetsCompositeArgument() (ast.GroupingSpecification, error) {
	astStart := p.curTok
	_ = astStart

	p.nextToken() // consume (
	firstItemTok := p.curTok

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

	// ScriptDom spans a composite inside GROUPING SETS over its contents,
	// excluding the enclosing parentheses.
	cgs := &ast.CompositeGroupingSpecification{Items: items}
	p.spanFrom(firstItemTok, cgs)
	cgs.Pin()

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return cgs, nil
}

// parseGroupingSpecificationArgument parses an argument inside ROLLUP/CUBE which can be
// an expression or a composite grouping like (c2, c3)
func (p *Parser) parseGroupingSpecificationArgument() (ast.GroupingSpecification, error) {
	astStart := p.curTok

	// Check for composite grouping (c1, c2)
	if p.curTok.Type == TokenLParen {
		spanV66, spanErr66 := p.parseCompositeGroupingSpecification()
		return spanned(p, spanV66, astStart), spanErr66
	}

	// Regular expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return spanned(p, &ast.ExpressionGroupingSpecification{
		Expression:             expr,
		DistributedAggregation: false,
	}, astStart), nil
}

// parseCompositeGroupingSpecification parses (c1, c2, ...)
func (p *Parser) parseCompositeGroupingSpecification() (*ast.CompositeGroupingSpecification, error) {
	astStart := p.curTok

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

	return spanned(p, spec, astStart), nil
}

func (p *Parser) parseHavingClause() (*ast.HavingClause, error) {
	astStart := p.curTok

	// Consume HAVING
	p.nextToken()

	condition, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}

	return spanned(p, &ast.HavingClause{SearchCondition: condition}, astStart), nil
}

func (p *Parser) parseOrderByClause() (*ast.OrderByClause, error) {
	astStart := p.curTok

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

		p.spanFromChild(elem, expr)

		obc.OrderByElements = append(obc.OrderByElements, elem)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return spanned(p, obc, astStart), nil
}

// parseOffsetClause parses OFFSET n ROWS FETCH NEXT/FIRST m ROWS ONLY
func (p *Parser) parseOffsetClause() (*ast.OffsetClause, error) {
	astStart := p.curTok

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

	return spanned(p, oc, astStart), nil
}

func (p *Parser) parseBooleanExpression() (ast.BooleanExpression, error) {
	astStart := p.curTok

	spanV67, spanErr67 := p.parseBooleanOrExpression()
	return spanned(p, spanV67, astStart), spanErr67
}

func (p *Parser) parseBooleanOrExpression() (ast.BooleanExpression, error) {
	astStart := p.curTok

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
		if bbe, ok := left.(*ast.BooleanBinaryExpression); ok {
			p.pinBinaryFromPinnedChild(bbe)
		}
	}

	return spanned(p, left, astStart), nil
}

func (p *Parser) parseBooleanAndExpression() (ast.BooleanExpression, error) {
	astStart := p.curTok

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
		if bbe, ok := left.(*ast.BooleanBinaryExpression); ok {
			p.pinBinaryFromPinnedChild(bbe)
		}
	}

	return spanned(p, left, astStart), nil
}

func (p *Parser) parseBooleanPrimaryExpression() (ast.BooleanExpression, error) {
	astStart := p.curTok

	// Check for NOT before other predicates (NOT EXISTS, NOT MATCH, etc.)
	if p.curTok.Type == TokenNot {
		p.nextToken() // consume NOT
		inner, err := p.parseBooleanPrimaryExpression()
		if err != nil {
			return nil, err
		}
		return spanned(p, &ast.BooleanNotExpression{Expression: inner}, astStart), nil
	}

	// Check for CONTAINS/FREETEXT predicates
	if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper == "CONTAINS" || upper == "FREETEXT" {
			spanV68, spanErr68 := p.parseFullTextPredicate(upper)
			return spanned(p, spanV68, astStart), spanErr68
		}
		if upper == "EXISTS" {
			spanV69, spanErr69 := p.parseExistsPredicate()
			return spanned(p, spanV69, astStart), spanErr69
		}
		if upper == "MATCH" {
			return p.parseGraphMatchPredicate()
		}
		if upper == "TSEQUAL" {
			spanV70, spanErr70 := p.parseTSEqualPredicate()
			return spanned(p, spanV70, astStart), spanErr70
		}
		if upper == "NOT" {
			// Handle NOT followed by MATCH, EXISTS, etc.
			p.nextToken() // consume NOT
			inner, err := p.parseBooleanPrimaryExpression()
			if err != nil {
				return nil, err
			}
			return spanned(p, &ast.BooleanNotExpression{Expression: inner}, astStart), nil
		}
	}

	// Check for UPDATE(column) predicate - used in triggers
	if p.curTok.Type == TokenUpdate && p.peekTok.Type == TokenLParen {
		p.nextToken() // consume UPDATE
		p.nextToken() // consume (

		// Parse the column identifier
		ident := p.parseIdentifier()

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		return spanned(p, &ast.UpdateCall{Identifier: ident}, astStart), nil
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
				spanV71, spanErr71 := p.parseComparisonAfterLeft(subquery)
				return spanned(p, spanV71, astStart), spanErr71
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
			p.spanFrom(astStart, parenExpr)

			// Check for comparison operators after the parenthesized expression
			if p.isComparisonOperator() {
				spanV72, spanErr72 := p.parseComparisonAfterLeft(parenExpr)
				return spanned(p, spanV72, astStart), spanErr72
			}

			// Check for IS NULL / IS NOT NULL
			if p.curTok.Type == TokenIs {
				spanV73, spanErr73 := p.parseIsNullAfterLeft(parenExpr)
				return spanned(p, spanV73, astStart), spanErr73
			}

			// Check for NOT before IN/LIKE/BETWEEN
			notDefined := false
			if p.curTok.Type == TokenNot {
				notDefined = true
				p.nextToken()
			}

			if p.curTok.Type == TokenIn {
				spanV74, spanErr74 := p.parseInExpressionAfterLeft(parenExpr, notDefined)
				return spanned(p, spanV74, astStart), spanErr74
			}
			if p.curTok.Type == TokenLike {
				spanV75, spanErr75 := p.parseLikeExpressionAfterLeft(parenExpr, notDefined)
				return spanned(p, spanV75, astStart), spanErr75
			}
			if p.curTok.Type == TokenBetween {
				spanV76, spanErr76 := p.parseBetweenExpressionAfterLeft(parenExpr, notDefined)
				return spanned(p, spanV76, astStart), spanErr76
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

		return spanned(p, &ast.BooleanParenthesisExpression{Expression: inner}, astStart), nil
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
				return spanned(p, &ast.BooleanIsNullExpression{
					IsNot:      !isNot,
					Expression: left,
				}, astStart), nil
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
				subqLParen := p.curTok
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

				subq := &ast.ScalarSubquery{QueryExpression: subqueryExpr}
				p.spanFrom(subqLParen, subq)
				return spanned(p, &ast.SubqueryComparisonPredicate{
					Expression:                      left,
					ComparisonType:                  compType,
					Subquery:                        subq,
					SubqueryComparisonPredicateType: predicateType,
				}, astStart), nil
			}

			// Parse the second expression
			secondExpr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			return spanned(p, &ast.DistinctPredicate{
				FirstExpression:  left,
				SecondExpression: secondExpr,
				IsNot:            isNot,
			}, astStart), nil
		}

		if p.curTok.Type != TokenNull {
			return nil, fmt.Errorf("expected NULL or DISTINCT after IS/IS NOT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume NULL

		return spanned(p, &ast.BooleanIsNullExpression{
			IsNot:      isNot,
			Expression: left,
		}, astStart), nil
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
			return spanned(p, &ast.BooleanInExpression{
				Expression: left,
				NotDefined: notDefined,
				Subquery:   subquery,
			}, astStart), nil
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
		return spanned(p, &ast.BooleanInExpression{
			Expression: left,
			NotDefined: notDefined,
			Values:     values,
		}, astStart), nil
	}

	// Check for LIKE expression
	if p.curTok.Type == TokenLike {
		p.nextToken() // consume LIKE

		pattern, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		var escapeExpr ast.ScalarExpression
		var odbcEscape bool
		if p.curTok.Type == TokenEscape {
			p.nextToken() // consume ESCAPE
			escapeExpr, err = p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
		} else if p.curTok.Type == TokenLBrace {
			// ODBC escape syntax: {ESCAPE 'x'}
			p.nextToken() // consume {
			if p.curTok.Type == TokenEscape {
				odbcEscape = true
				p.nextToken() // consume ESCAPE
				escapeExpr, err = p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				if p.curTok.Type != TokenRBrace {
					return nil, fmt.Errorf("expected }, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume }
			} else {
				return nil, fmt.Errorf("expected ESCAPE after {, got %s", p.curTok.Literal)
			}
		}

		return spanned(p, &ast.BooleanLikeExpression{
			FirstExpression:  left,
			SecondExpression: pattern,
			EscapeExpression: escapeExpr,
			NotDefined:       notDefined,
			OdbcEscape:       odbcEscape,
		}, astStart), nil
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
		return spanned(p, &ast.BooleanTernaryExpression{
			TernaryExpressionType: ternaryType,
			FirstExpression:       left,
			SecondExpression:      low,
			ThirdExpression:       high,
		}, astStart), nil
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
	case TokenError:
		// Handle T-SQL specific ! operators: !=, !<, !>
		if p.curTok.Literal == "!" {
			p.nextToken() // consume !
			switch p.curTok.Type {
			case TokenEquals:
				compType = "NotEqualToExclamation"
			case TokenLessThan:
				compType = "NotLessThan"
			case TokenGreaterThan:
				compType = "NotGreaterThan"
			default:
				return nil, fmt.Errorf("expected =, <, or > after !, got %s", p.curTok.Literal)
			}
		} else {
			return nil, fmt.Errorf("expected comparison operator, got %s", p.curTok.Literal)
		}
	case TokenRParen:
		// We're at ) without a comparison operator - this happens when parsing
		// a parenthesized scalar expression like (XACT_STATE()) in a boolean context.
		// Return a special marker that the caller can handle.
		return spanned(p, &ast.BooleanScalarPlaceholder{Scalar: left}, astStart), nil
	default:
		return nil, fmt.Errorf("expected comparison operator, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse right scalar expression
	right, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return spanned(p, &ast.BooleanComparisonExpression{
		ComparisonType:   compType,
		FirstExpression:  left,
		SecondExpression: right,
	}, astStart), nil
}

// isComparisonOperator checks if the current token is a comparison operator
func (p *Parser) isComparisonOperator() bool {
	switch p.curTok.Type {
	case TokenEquals, TokenNotEqual, TokenLessThan, TokenGreaterThan,
		TokenLessOrEqual, TokenGreaterOrEqual:
		return true
	case TokenError:
		// Handle T-SQL specific ! operators: !=, !<, !>
		if p.curTok.Literal == "!" {
			switch p.peekTok.Type {
			case TokenEquals, TokenLessThan, TokenGreaterThan:
				return true
			}
		}
		return false
	default:
		return false
	}
}

// parseComparisonAfterLeft parses a comparison expression after the left operand is already parsed
func (p *Parser) parseComparisonAfterLeft(left ast.ScalarExpression) (ast.BooleanExpression, error) {
	astStart := p.curTok

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
	case TokenError:
		// Handle T-SQL specific ! operators: !=, !<, !>
		if p.curTok.Literal == "!" {
			p.nextToken() // consume !
			switch p.curTok.Type {
			case TokenEquals:
				compType = "NotEqualToExclamation"
			case TokenLessThan:
				compType = "NotLessThan"
			case TokenGreaterThan:
				compType = "NotGreaterThan"
			default:
				return nil, fmt.Errorf("expected =, <, or > after !, got %s", p.curTok.Literal)
			}
		} else {
			return nil, fmt.Errorf("expected comparison operator, got %s", p.curTok.Literal)
		}
	default:
		return nil, fmt.Errorf("expected comparison operator, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse right scalar expression
	right, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return spanned(p, &ast.BooleanComparisonExpression{
		ComparisonType:   compType,
		FirstExpression:  left,
		SecondExpression: right,
	}, astStart), nil
}

// parseInExpressionAfterLeft parses an IN expression after the left operand is already parsed
func (p *Parser) parseInExpressionAfterLeft(left ast.ScalarExpression, notDefined bool) (ast.BooleanExpression, error) {
	astStart := p.curTok

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
		return spanned(p, &ast.BooleanInExpression{
			Expression: left,
			NotDefined: notDefined,
			Subquery:   subquery,
		}, astStart), nil
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
	return spanned(p, &ast.BooleanInExpression{
		Expression: left,
		NotDefined: notDefined,
		Values:     values,
	}, astStart), nil
}

// parseLikeExpressionAfterLeft parses a LIKE expression after the left operand is already parsed
func (p *Parser) parseLikeExpressionAfterLeft(left ast.ScalarExpression, notDefined bool) (ast.BooleanExpression, error) {
	astStart := p.curTok

	p.nextToken() // consume LIKE

	pattern, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	var escapeExpr ast.ScalarExpression
	var odbcEscape bool
	if p.curTok.Type == TokenEscape {
		p.nextToken() // consume ESCAPE
		escapeExpr, err = p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
	} else if p.curTok.Type == TokenLBrace {
		// ODBC escape syntax: {ESCAPE 'x'}
		p.nextToken() // consume {
		if p.curTok.Type == TokenEscape {
			odbcEscape = true
			p.nextToken() // consume ESCAPE
			escapeExpr, err = p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			if p.curTok.Type != TokenRBrace {
				return nil, fmt.Errorf("expected }, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume }
		} else {
			return nil, fmt.Errorf("expected ESCAPE after {, got %s", p.curTok.Literal)
		}
	}

	return spanned(p, &ast.BooleanLikeExpression{
		FirstExpression:  left,
		SecondExpression: pattern,
		EscapeExpression: escapeExpr,
		NotDefined:       notDefined,
		OdbcEscape:       odbcEscape,
	}, astStart), nil
}

// parseBetweenExpressionAfterLeft parses a BETWEEN expression after the left operand is already parsed
func (p *Parser) parseBetweenExpressionAfterLeft(left ast.ScalarExpression, notDefined bool) (ast.BooleanExpression, error) {
	astStart := p.curTok

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
	return spanned(p, &ast.BooleanTernaryExpression{
		TernaryExpressionType: ternaryType,
		FirstExpression:       left,
		SecondExpression:      low,
		ThirdExpression:       high,
	}, astStart), nil
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
	astStart := p.curTok

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
			return spanned(p, &ast.BooleanIsNullExpression{
				IsNot:      !isNot,
				Expression: left,
			}, astStart), nil
		}

		// Parse the second expression
		secondExpr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		return spanned(p, &ast.DistinctPredicate{
			FirstExpression:  left,
			SecondExpression: secondExpr,
			IsNot:            isNot,
		}, astStart), nil
	}

	if p.curTok.Type != TokenNull {
		return nil, fmt.Errorf("expected NULL or DISTINCT after IS/IS NOT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume NULL

	return spanned(p, &ast.BooleanIsNullExpression{
		IsNot:      isNot,
		Expression: left,
	}, astStart), nil
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
	astStart := p.curTok

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

	return spanned(p, cast, astStart), nil
}

// parseConvertCall parses a CONVERT expression: CONVERT(data_type, expression [, style])
func (p *Parser) parseConvertCall() (ast.ScalarExpression, error) {
	astStart := p.curTok

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

	return spanned(p, convert, astStart), nil
}

// parseTryCastCall parses a TRY_CAST expression
func (p *Parser) parseTryCastCall() (ast.ScalarExpression, error) {
	astStart := p.curTok

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

	return spanned(p, cast, astStart), nil
}

// parseTryConvertCall parses a TRY_CONVERT expression
func (p *Parser) parseTryConvertCall() (ast.ScalarExpression, error) {
	astStart := p.curTok

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

	return spanned(p, convert, astStart), nil
}

// parseNullIfExpression parses a NULLIF(expr1, expr2) expression
func (p *Parser) parseNullIfExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	p.nextToken() // consume NULLIF
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after NULLIF, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse first expression
	first, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , in NULLIF, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse second expression
	second, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in NULLIF, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return spanned(p, &ast.NullIfExpression{
		FirstExpression:  first,
		SecondExpression: second,
	}, astStart), nil
}

// parseCoalesceExpression parses a COALESCE(expr1, expr2, ...) expression
func (p *Parser) parseCoalesceExpression() (ast.ScalarExpression, error) {
	astStart := p.curTok

	p.nextToken() // consume COALESCE
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after COALESCE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	var expressions []ast.ScalarExpression

	// Parse expressions
	for {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		expressions = append(expressions, expr)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		} else {
			break
		}
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in COALESCE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return spanned(p, &ast.CoalesceExpression{
		Expressions: expressions,
	}, astStart), nil
}

// parseIdentityFunctionCall parses an IDENTITY function call: IDENTITY(data_type [, seed, increment])
func (p *Parser) parseIdentityFunctionCall() (ast.ScalarExpression, error) {
	astStart := p.curTok

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

	return spanned(p, identity, astStart), nil
}

// parseLeftFunctionCall parses LEFT(string, count)
func (p *Parser) parseLeftFunctionCall() (ast.ScalarExpression, error) {
	astStart := p.curTok

	// Already consumed LEFT, now on (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after LEFT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	var params []ast.ScalarExpression

	// Parse parameters
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		params = append(params, param)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in LEFT function, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return spanned(p, &ast.LeftFunctionCall{Parameters: params}, astStart), nil
}

// parseRightFunctionCall parses RIGHT(string, count)
func (p *Parser) parseRightFunctionCall() (ast.ScalarExpression, error) {
	astStart := p.curTok

	// Already consumed RIGHT, now on (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after RIGHT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	var params []ast.ScalarExpression

	// Parse parameters
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		params = append(params, param)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in RIGHT function, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return spanned(p, &ast.RightFunctionCall{Parameters: params}, astStart), nil
}

// parsePredictTableReference parses PREDICT(...) in FROM clause
// PREDICT(MODEL = expression, DATA = table AS alias, RUNTIME=ident) WITH (columns) AS alias
func (p *Parser) parsePredictTableReference() (*ast.PredictTableReference, error) {
	astStart := p.curTok

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
				parenTok := p.curTok
				p.nextToken() // consume (
				qe, err := p.parseQueryExpression()
				if err != nil {
					return nil, err
				}
				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
				}
				p.nextToken() // consume )
				ss := &ast.ScalarSubquery{QueryExpression: qe}
				p.spanFrom(parenTok, ss)
				ref.ModelVariable = ss
			} else if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				// Variable
				ref.ModelVariable = p.spanVarRef(p.curTok.Literal)
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

	return spanned(p, ref, astStart), nil
}

// parseForClause parses FOR BROWSE, FOR XML, FOR UPDATE, FOR READ ONLY clauses.
func (p *Parser) parseForClause() (ast.ForClause, error) {
	astStart := p.curTok

	p.nextToken() // consume FOR

	keyword := strings.ToUpper(p.curTok.Literal)

	switch keyword {
	case "BROWSE":
		p.nextToken() // consume BROWSE
		return spanned(p, &ast.BrowseForClause{}, astStart), nil

	case "READ":
		p.nextToken() // consume READ
		if strings.ToUpper(p.curTok.Literal) == "ONLY" {
			// ScriptDom spans FOR READ ONLY on the ONLY token.
			roClause := &ast.ReadOnlyForClause{}
			p.tokSpan(roClause, p.curTok)
			roClause.Pin()
			p.nextToken() // consume ONLY
			return roClause, nil
		}
		return spanned(p, &ast.ReadOnlyForClause{}, astStart), nil

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
		return spanned(p, clause, astStart), nil

	case "XML":
		p.nextToken() // consume XML
		spanV77, spanErr77 := p.parseXmlForClause()
		return spanned(p, spanV77, astStart), spanErr77

	case "JSON":
		p.nextToken() // consume JSON
		spanV78, spanErr78 := p.parseJsonForClause()
		return spanned(p, spanV78, astStart), spanErr78

	default:
		return nil, fmt.Errorf("unexpected token after FOR: %s", p.curTok.Literal)
	}
}

// parseXmlForClause parses FOR XML options.
func (p *Parser) parseXmlForClause() (*ast.XmlForClause, error) {
	astStart := p.curTok

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

	return spanned(p, clause, astStart), nil
}

// parseXmlForClauseOption parses a single XML FOR clause option.
func (p *Parser) parseXmlForClauseOption() (*ast.XmlForClauseOption, error) {
	astStart := p.curTok

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
		// Check for XSINIL or ABSENT; ScriptDom spans these forms on the
		// trailing keyword only.
		nextKeyword := strings.ToUpper(p.curTok.Literal)
		if nextKeyword == "XSINIL" {
			option.OptionKind = "ElementsXsiNil"
			astStart = p.curTok
			p.nextToken() // consume XSINIL
		} else if nextKeyword == "ABSENT" {
			option.OptionKind = "ElementsAbsent"
			astStart = p.curTok
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
		// BINARY BASE64 - ScriptDom spans this option on BASE64 only.
		if strings.ToUpper(p.curTok.Literal) == "BASE64" {
			option.OptionKind = "BinaryBase64"
			astStart = p.curTok
			p.nextToken() // consume BASE64
		}
	default:
		option.OptionKind = keyword
	}

	return spanned(p, option, astStart), nil
}

// parseJsonForClause parses FOR JSON options.
func (p *Parser) parseJsonForClause() (*ast.JsonForClause, error) {
	astStart := p.curTok

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

	return spanned(p, clause, astStart), nil
}

// parseJsonForClauseOption parses a single JSON FOR clause option.
func (p *Parser) parseJsonForClauseOption() (*ast.JsonForClauseOption, error) {
	astStart := p.curTok

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

	return spanned(p, option, astStart), nil
}

// getPseudoColumnType returns the ColumnType for pseudo columns like $identity, $action, etc.
// Returns empty string if not a pseudo column.
func getPseudoColumnType(value string) string {
	switch strings.ToUpper(value) {
	case "$IDENTITY":
		return "PseudoColumnIdentity"
	case "$ACTION":
		return "PseudoColumnAction"
	case "$ROWGUID":
		return "PseudoColumnRowGuid"
	case "$CUID":
		return "PseudoColumnCuid"
	case "$NODE_ID":
		return "PseudoColumnGraphNodeId"
	case "$EDGE_ID":
		return "PseudoColumnGraphEdgeId"
	case "$FROM_ID":
		return "PseudoColumnGraphFromId"
	case "$TO_ID":
		return "PseudoColumnGraphToId"
	default:
		return ""
	}
}

// getParameterlessCallType returns the ParameterlessCallType for keywords like USER, CURRENT_USER, etc.
func getParameterlessCallType(value string) string {
	switch value {
	case "USER":
		return "User"
	case "CURRENT_USER":
		return "CurrentUser"
	case "SESSION_USER":
		return "SessionUser"
	case "SYSTEM_USER":
		return "SystemUser"
	case "CURRENT_TIMESTAMP":
		return "CurrentTimestamp"
	case "CURRENT_DATE":
		return "CurrentDate"
	default:
		return ""
	}
}

// parseFullTextPredicate parses CONTAINS or FREETEXT predicates
func (p *Parser) parseFullTextPredicate(funcType string) (*ast.FullTextPredicate, error) {
	astStart := p.curTok

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
				// Check for pseudo column
				pseudoType := getPseudoColumnType(col.Value)
				if pseudoType != "" && p.curTok.Type != TokenDot {
					// Standalone pseudo column like $identity
					pred.Columns = append(pred.Columns, &ast.ColumnReferenceExpression{
						ColumnType: pseudoType,
					})
				} else if p.curTok.Type == TokenDot {
					// Check for table.column or table.*
					p.nextToken() // consume .
					if p.curTok.Type == TokenStar {
						// table.*
						p.nextToken() // consume *
						pred.Columns = append(pred.Columns, &ast.ColumnReferenceExpression{
							ColumnType: "Wildcard",
							MultiPartIdentifier: &ast.MultiPartIdentifier{
								Identifiers: []*ast.Identifier{col},
								Count:       1,
							},
						})
					} else {
						// table.column or table.$identity
						col2 := p.parseIdentifier()
						pseudoType2 := getPseudoColumnType(col2.Value)
						if pseudoType2 != "" {
							// table.$identity - pseudo column with table prefix
							pred.Columns = append(pred.Columns, &ast.ColumnReferenceExpression{
								ColumnType: pseudoType2,
								MultiPartIdentifier: &ast.MultiPartIdentifier{
									Identifiers: []*ast.Identifier{col},
									Count:       1,
								},
							})
						} else {
							pred.Columns = append(pred.Columns, &ast.ColumnReferenceExpression{
								ColumnType: "Regular",
								MultiPartIdentifier: &ast.MultiPartIdentifier{
									Identifiers: []*ast.Identifier{col, col2},
									Count:       2,
								},
							})
						}
					}
				} else {
					pred.Columns = append(pred.Columns, &ast.ColumnReferenceExpression{
						ColumnType: "Regular",
						MultiPartIdentifier: &ast.MultiPartIdentifier{
							Identifiers: []*ast.Identifier{col},
							Count:       1,
						},
					})
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
		// Single column or table.column or table.*
		col := p.parseIdentifier()
		// Check for pseudo column
		pseudoType := getPseudoColumnType(col.Value)
		if pseudoType != "" && p.curTok.Type != TokenDot {
			// Standalone pseudo column like $identity
			pred.Columns = []*ast.ColumnReferenceExpression{{
				ColumnType: pseudoType,
			}}
		} else if p.curTok.Type == TokenDot {
			// Check for table.column or table.*
			p.nextToken() // consume .
			if p.curTok.Type == TokenStar {
				// table.*
				p.nextToken() // consume *
				pred.Columns = []*ast.ColumnReferenceExpression{{
					ColumnType: "Wildcard",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Identifiers: []*ast.Identifier{col},
						Count:       1,
					},
				}}
			} else {
				// table.column or table.$identity
				col2 := p.parseIdentifier()
				pseudoType2 := getPseudoColumnType(col2.Value)
				if pseudoType2 != "" {
					// table.$identity - pseudo column with table prefix
					pred.Columns = []*ast.ColumnReferenceExpression{{
						ColumnType: pseudoType2,
						MultiPartIdentifier: &ast.MultiPartIdentifier{
							Identifiers: []*ast.Identifier{col},
							Count:       1,
						},
					}}
				} else {
					pred.Columns = []*ast.ColumnReferenceExpression{{
						ColumnType: "Regular",
						MultiPartIdentifier: &ast.MultiPartIdentifier{
							Identifiers: []*ast.Identifier{col, col2},
							Count:       2,
						},
					}}
				}
			}
		} else {
			pred.Columns = []*ast.ColumnReferenceExpression{{
				ColumnType: "Regular",
				MultiPartIdentifier: &ast.MultiPartIdentifier{
					Identifiers: []*ast.Identifier{col},
					Count:       1,
				},
			}}
		}
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

	return spanned(p, pred, astStart), nil
}

// parseTSEqualPredicate parses TSEQUAL(expr1, expr2)
func (p *Parser) parseTSEqualPredicate() (*ast.TSEqualCall, error) {
	astStart := p.curTok

	p.nextToken() // consume TSEQUAL

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TSEQUAL, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse first expression
	first, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Expect comma
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , in TSEQUAL, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ,

	// Parse second expression
	second, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after TSEQUAL, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return spanned(p, &ast.TSEqualCall{
		FirstExpression:  first,
		SecondExpression: second,
	}, astStart), nil
}

// parseExistsPredicate parses EXISTS (subquery)
func (p *Parser) parseExistsPredicate() (*ast.ExistsPredicate, error) {
	astStart := p.curTok

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

	return spanned(p, &ast.ExistsPredicate{Subquery: subquery}, astStart), nil
}

// parseIIfCall parses IIF(condition, true_value, false_value)
func (p *Parser) parseIIfCall() (*ast.IIfCall, error) {
	astStart := p.curTok

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

	return spanned(p, &ast.IIfCall{
		Predicate:      pred,
		ThenExpression: thenExpr,
		ElseExpression: elseExpr,
	}, astStart), nil
}

// parseParseCall parses PARSE(string AS type [USING culture]) or TRY_PARSE(string AS type [USING culture])
func (p *Parser) parseParseCall(isTry bool) (ast.ScalarExpression, error) {
	astStart := p.curTok

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
		return spanned(p, &ast.TryParseCall{
			StringValue: strVal,
			DataType:    dataType,
			Culture:     culture,
		}, astStart), nil
	}
	return spanned(p, &ast.ParseCall{
		StringValue: strVal,
		DataType:    dataType,
		Culture:     culture,
	}, astStart), nil
}

// parseJsonObjectCall parses JSON_OBJECT('key':value, 'key2':value2, ... [NULL|ABSENT ON NULL])
func (p *Parser) parseJsonObjectCall() (*ast.FunctionCall, error) {
	astStart := p.curTok

	fc := &ast.FunctionCall{
		FunctionName:     p.spanIdent("JSON_OBJECT", "NotQuoted"),
		UniqueRowFilter:  "NotSpecified",
		WithArrayWrapper: false,
	}

	p.nextToken() // consume (

	// Parse key-value pairs
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		// Check for NULL ON NULL or ABSENT ON NULL at start of loop
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "NULL" || upperLit == "ABSENT" {
			// Look ahead to see if this is "NULL ON NULL" or "ABSENT ON NULL"
			if p.peekIsOnNull() {
				fc.AbsentOrNullOnNull = append(fc.AbsentOrNullOnNull, &ast.Identifier{Value: upperLit, QuoteType: "NotQuoted"})
				p.nextToken() // consume NULL or ABSENT
				p.nextToken() // consume ON
				p.nextToken() // consume NULL
				continue
			}
		}

		// Parse key expression
		keyExpr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}

		// Check for : (JSON key-value separator)
		if p.curTok.Type == TokenColon {
			p.nextToken() // consume :

			// Parse value expression
			valueExpr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}

			jkv := &ast.JsonKeyValue{
				JsonKeyName: keyExpr,
				JsonValue:   valueExpr,
			}
			// The key-value pair spans from key through value.
			p.spanFromChild(jkv, keyExpr)
			fc.JsonParameters = append(fc.JsonParameters, jkv)
		} else {
			// Just a regular parameter without colon (shouldn't happen for JSON_OBJECT)
			fc.Parameters = append(fc.Parameters, keyExpr)
		}

		// After parsing a value, check for NULL ON NULL or ABSENT ON NULL
		postValueLit := strings.ToUpper(p.curTok.Literal)
		if postValueLit == "NULL" || postValueLit == "ABSENT" {
			if p.peekIsOnNull() {
				fc.AbsentOrNullOnNull = append(fc.AbsentOrNullOnNull, &ast.Identifier{Value: postValueLit, QuoteType: "NotQuoted"})
				p.nextToken() // consume NULL or ABSENT
				p.nextToken() // consume ON
				p.nextToken() // consume NULL
				// Continue to check for ) or comma
			}
		}

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in JSON_OBJECT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return spanned(p, fc, astStart), nil
}

// parseJsonArrayCall parses JSON_ARRAY(value1, value2, ... [NULL|ABSENT ON NULL])
func (p *Parser) parseJsonArrayCall() (*ast.FunctionCall, error) {
	astStart := p.curTok

	fc := &ast.FunctionCall{
		FunctionName:     p.spanIdent("JSON_ARRAY", "NotQuoted"),
		UniqueRowFilter:  "NotSpecified",
		WithArrayWrapper: false,
	}

	p.nextToken() // consume (

	// Parse array elements
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		// Check for NULL ON NULL or ABSENT ON NULL at start of loop
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "NULL" || upperLit == "ABSENT" {
			// Look ahead to see if this is "NULL ON NULL" or "ABSENT ON NULL"
			if p.peekIsOnNull() {
				fc.AbsentOrNullOnNull = append(fc.AbsentOrNullOnNull, &ast.Identifier{Value: upperLit, QuoteType: "NotQuoted"})
				p.nextToken() // consume NULL or ABSENT
				p.nextToken() // consume ON
				p.nextToken() // consume NULL
				continue
			}
		}

		// Parse value expression
		valueExpr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		fc.Parameters = append(fc.Parameters, valueExpr)

		// After parsing a value, check for NULL ON NULL or ABSENT ON NULL
		postValueLit := strings.ToUpper(p.curTok.Literal)
		if postValueLit == "NULL" || postValueLit == "ABSENT" {
			if p.peekIsOnNull() {
				fc.AbsentOrNullOnNull = append(fc.AbsentOrNullOnNull, &ast.Identifier{Value: postValueLit, QuoteType: "NotQuoted"})
				p.nextToken() // consume NULL or ABSENT
				p.nextToken() // consume ON
				p.nextToken() // consume NULL
				// Continue to check for ) or comma
			}
		}

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) in JSON_ARRAY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	return spanned(p, fc, astStart), nil
}

// peekIsOnNull checks if the next tokens are "ON NULL"
func (p *Parser) peekIsOnNull() bool {
	// Just check if the next token is ON
	// The caller will verify the NULL after ON when consuming
	return p.peekTok.Type == TokenOn
}

// parseChangeTableReference parses CHANGETABLE(CHANGES ...) or CHANGETABLE(VERSION ...)
func (p *Parser) parseChangeTableReference() (ast.TableReference, error) {
	astStart := p.curTok

	p.nextToken() // consume CHANGETABLE

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after CHANGETABLE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	upper := strings.ToUpper(p.curTok.Literal)
	if upper == "CHANGES" {
		spanV79, spanErr79 := p.parseChangeTableChangesReference()
		return spanned(p, spanV79, astStart), spanErr79
	} else if upper == "VERSION" {
		spanV80, spanErr80 := p.parseChangeTableVersionReference()
		return spanned(p, spanV80, astStart), spanErr80
	}

	return nil, fmt.Errorf("expected CHANGES or VERSION after CHANGETABLE(, got %s", p.curTok.Literal)
}

// parseChangeTableChangesReference parses CHANGETABLE(CHANGES table, version [, FORCESEEK])
func (p *Parser) parseChangeTableChangesReference() (*ast.ChangeTableChangesTableReference, error) {
	astStart := p.curTok

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

	return spanned(p, ref, astStart), nil
}

// parseChangeTableVersionReference parses CHANGETABLE(VERSION table, (cols), (vals) [, FORCESEEK])
func (p *Parser) parseChangeTableVersionReference() (*ast.ChangeTableVersionTableReference, error) {
	astStart := p.curTok

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

	return spanned(p, ref, astStart), nil
}

// parseOverClause parses an OVER clause after a function call
// Handles both: OVER Win1 and OVER (PARTITION BY c1 ORDER BY c2 ROWS ...)
func (p *Parser) parseOverClause() (*ast.OverClause, error) {
	astStart := p.curTok

	// Current token should be OVER, consume it
	p.nextToken() // consume OVER

	overClause := &ast.OverClause{}

	// Check if it's just a window name (no parentheses)
	if p.curTok.Type != TokenLParen {
		// It's OVER WindowName
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			overClause.WindowName = p.parseIdentifier()
			return spanned(p, overClause, astStart), nil
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

	return spanned(p, overClause, astStart), nil
}

// parseWindowFrameClause parses ROWS/RANGE ... BETWEEN ... AND ...
func (p *Parser) parseWindowFrameClause() (*ast.WindowFrameClause, error) {
	astStart := p.curTok

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

	spanned(p, frame, astStart)
	// The frame clause ends where its last delimiter ends (a CURRENT ROW
	// delimiter ends at CURRENT, excluding ROW).
	last := frame.Bottom
	if last == nil {
		last = frame.Top
	}
	if last != nil && last.Frag().HasSpan() && frame.Frag().HasSpan() {
		if e := last.Frag().EndOffset(); e < frame.Frag().EndOffset() {
			frame.Frag().FragmentLength = e - frame.Frag().StartOffset
		}
	}
	return frame, nil
}

// parseWindowDelimiter parses UNBOUNDED PRECEDING/FOLLOWING, CURRENT ROW, n PRECEDING/FOLLOWING
func (p *Parser) parseWindowDelimiter() (*ast.WindowDelimiter, error) {
	astStart := p.curTok

	delim := &ast.WindowDelimiter{}

	upperLit := strings.ToUpper(p.curTok.Literal)

	if upperLit == "CURRENT" {
		p.nextToken() // consume CURRENT
		if strings.ToUpper(p.curTok.Literal) != "ROW" {
			return nil, fmt.Errorf("expected ROW after CURRENT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ROW
		delim.WindowDelimiterType = "CurrentRow"
		// ScriptDom positions CURRENT ROW on the CURRENT keyword alone.
		p.tokSpan(delim, astStart)
		return delim, nil
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

	return spanned(p, delim, astStart), nil
}

// parseWindowClause parses WINDOW Win1 AS (...), Win2 AS (...)
func (p *Parser) parseWindowClause() (*ast.WindowClause, error) {
	astStart := p.curTok

	p.nextToken() // consume WINDOW

	clause := &ast.WindowClause{}
	firstDefTok := p.curTok

	for {
		def := &ast.WindowDefinition{}
		defStartTok := p.curTok

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

		// The definition spans the name through the closing parenthesis.
		p.spanFrom(defStartTok, def)
		clause.WindowDefinition = append(clause.WindowDefinition, def)

		// Check for comma (more window definitions)
		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume ,
	}

	// ScriptDom spans the WINDOW clause over its definitions, excluding
	// the WINDOW keyword itself.
	p.spanFrom(firstDefTok, clause)
	clause.Pin()
	_ = astStart
	return clause, nil
}

// parsePivotedTableReference parses PIVOT clause
// Syntax: table PIVOT (aggregate_func(columns) FOR pivot_column IN (value1, value2, ...)) AS alias
func (p *Parser) parsePivotedTableReference(tableRef ast.TableReference) (*ast.PivotedTableReference, error) {
	astStart := p.curTok

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

	return spanned(p, pivoted, astStart), nil
}

// parseUnpivotedTableReference parses UNPIVOT clause
func (p *Parser) parseUnpivotedTableReference(tableRef ast.TableReference) (*ast.UnpivotedTableReference, error) {
	astStart := p.curTok

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
	unpivoted.ValueColumn = p.parseIdentifier()

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

	return spanned(p, unpivoted, astStart), nil
}

// parseTableSampleClause parses a TABLESAMPLE clause
// Syntax: TABLESAMPLE [SYSTEM] (expression [PERCENT | ROWS]) [REPEATABLE (seed)]
func (p *Parser) parseTableSampleClause() (*ast.TableSampleClause, error) {
	astStart := p.curTok

	p.nextToken() // consume TABLESAMPLE

	clause := &ast.TableSampleClause{
		System:                  false,
		TableSampleClauseOption: "NotSpecified",
	}

	// Check for SYSTEM keyword
	if strings.ToUpper(p.curTok.Literal) == "SYSTEM" {
		clause.System = true
		p.nextToken() // consume SYSTEM
	}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TABLESAMPLE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse the sample expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	clause.SampleNumber = expr

	// Check for PERCENT or ROWS option
	upper := strings.ToUpper(p.curTok.Literal)
	if upper == "PERCENT" {
		clause.TableSampleClauseOption = "Percent"
		p.nextToken()
	} else if upper == "ROWS" {
		clause.TableSampleClauseOption = "Rows"
		p.nextToken()
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after TABLESAMPLE expression, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Check for REPEATABLE (seed)
	if strings.ToUpper(p.curTok.Literal) == "REPEATABLE" {
		p.nextToken() // consume REPEATABLE

		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after REPEATABLE, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		seed, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		clause.RepeatSeed = seed

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after REPEATABLE seed, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	}

	return spanned(p, clause, astStart), nil
}

// parseBuiltInFunctionTableReference parses a built-in function table reference
// Syntax: ::function_name(parameters) [AS alias [(column_list)]]
func (p *Parser) parseBuiltInFunctionTableReference() (*ast.BuiltInFunctionTableReference, error) {
	astStart := p.curTok

	p.nextToken() // consume ::

	ref := &ast.BuiltInFunctionTableReference{
		ForPath: false,
	}

	// Parse function name
	ref.Name = p.parseIdentifier()

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after built-in function name, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	// Parse parameters
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		ref.Parameters = append(ref.Parameters, param)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after built-in function parameters, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume )

	// Parse optional alias (AS alias or just alias)
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "WINDOW" && upper != "ORDER" &&
			upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
			upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
			upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" &&
			upper != "PIVOT" && upper != "UNPIVOT" {
			ref.Alias = p.parseIdentifier()
		}
	}

	// Check for column list: alias(c1, c2, ...)
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
			p.nextToken()
		}
	}

	return spanned(p, ref, astStart), nil
}

// parseAdHocTableReference parses OPENDATASOURCE('provider', 'connstr').'object'
func (p *Parser) parseAdHocTableReference() (*ast.AdHocTableReference, error) {
	astStart := p.curTok

	p.nextToken() // consume OPENDATASOURCE

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after OPENDATASOURCE")
	}
	p.nextToken() // consume (

	// Parse provider name (should be a string literal)
	providerNameExpr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	providerName, ok := providerNameExpr.(*ast.StringLiteral)
	if !ok {
		return nil, fmt.Errorf("expected string literal for provider name")
	}

	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after provider name")
	}
	p.nextToken() // consume ,

	// Parse init string (connection string)
	initStringExpr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	initString, ok := initStringExpr.(*ast.StringLiteral)
	if !ok {
		return nil, fmt.Errorf("expected string literal for init string")
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after init string")
	}
	p.nextToken() // consume )

	dataSource := &ast.AdHocDataSource{
		ProviderName: providerName,
		InitString:   initString,
	}

	// Expect dot followed by object
	if p.curTok.Type != TokenDot {
		return nil, fmt.Errorf("expected . after OPENDATASOURCE(), got %s", p.curTok.Literal)
	}
	p.nextToken() // consume .

	// Parse the object - could be a string or schema object name
	var obj *ast.SchemaObjectNameOrValueExpression
	if p.curTok.Type == TokenString {
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		obj = &ast.SchemaObjectNameOrValueExpression{
			ValueExpression: expr,
		}
	} else {
		son, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		obj = &ast.SchemaObjectNameOrValueExpression{
			SchemaObjectName: son,
		}
	}

	result := &ast.AdHocTableReference{
		DataSource: dataSource,
		Object:     obj,
		ForPath:    false,
	}

	// Parse optional alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		result.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "ORDER" &&
			upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
			upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
			upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
			result.Alias = p.parseIdentifier()
		}
	}

	return spanned(p, result, astStart), nil
}

func (p *Parser) parseOpenXmlTableReference() (*ast.OpenXmlTableReference, error) {
	astStart := p.curTok

	p.nextToken() // consume OPENXML

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after OPENXML")
	}
	p.nextToken() // consume (

	// Parse variable (e.g., @idoc)
	variable, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after variable")
	}
	p.nextToken() // consume ,

	// Parse row pattern (e.g., '/ROOT/Customer')
	rowPattern, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	result := &ast.OpenXmlTableReference{
		Variable:   variable,
		RowPattern: rowPattern,
		ForPath:    false,
	}

	// Optional flags parameter
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		flags, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		result.Flags = flags
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after OPENXML parameters")
	}
	p.nextToken() // consume )

	// Optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		if p.curTok.Type == TokenLParen {
			// WITH (schema declarations)
			p.nextToken() // consume (

			for {
				// Parse column definition with optional mapping
				item, err := p.parseSchemaDeclarationItem()
				if err != nil {
					return nil, err
				}
				result.SchemaDeclarationItems = append(result.SchemaDeclarationItems, item)

				if p.curTok.Type != TokenComma {
					break
				}
				p.nextToken() // consume ,
			}

			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ) after schema declarations")
			}
			p.nextToken() // consume )
		} else {
			// WITH table_name
			tableName, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, err
			}
			result.TableName = tableName
		}
	}

	// Optional AS alias or just alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		result.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "ORDER" &&
			upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
			upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
			upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
			result.Alias = p.parseIdentifier()
		}
	}

	return spanned(p, result, astStart), nil
}

func (p *Parser) parseSchemaDeclarationItem() (*ast.SchemaDeclarationItem, error) {
	astStart := p.curTok

	// Parse column name
	colName := p.parseIdentifier()

	// Parse data type
	dataType, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}

	colDef := &ast.ColumnDefinitionBase{
		ColumnIdentifier: colName,
		DataType:         dataType,
	}

	item := &ast.SchemaDeclarationItem{
		ColumnDefinition: colDef,
	}

	// Optional mapping (XPath expression as string literal)
	// e.g., "CustomerID VARCHAR (10) '../@CustomerID'"
	if p.curTok.Type == TokenString {
		mapping, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		item.Mapping = mapping
	}

	return spanned(p, item, astStart), nil
}

func (p *Parser) parseOpenQueryTableReference() (*ast.OpenQueryTableReference, error) {
	astStart := p.curTok

	p.nextToken() // consume OPENQUERY

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after OPENQUERY")
	}
	p.nextToken() // consume (

	// Parse linked server identifier
	linkedServer := p.parseIdentifier()

	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after linked server")
	}
	p.nextToken() // consume ,

	// Parse query (string literal)
	query, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after OPENQUERY parameters")
	}
	p.nextToken() // consume )

	result := &ast.OpenQueryTableReference{
		LinkedServer: linkedServer,
		Query:        query,
		ForPath:      false,
	}

	// Optional AS alias or just alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		result.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "WHERE" && upper != "GROUP" && upper != "HAVING" && upper != "ORDER" &&
			upper != "OPTION" && upper != "GO" && upper != "WITH" && upper != "ON" &&
			upper != "JOIN" && upper != "INNER" && upper != "LEFT" && upper != "RIGHT" &&
			upper != "FULL" && upper != "CROSS" && upper != "OUTER" && upper != "FOR" {
			result.Alias = p.parseIdentifier()
		}
	}

	return spanned(p, result, astStart), nil
}
