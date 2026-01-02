// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/teesql/ast"
)

func (p *Parser) parseWithStatement() (ast.Statement, error) {
	// Consume WITH
	p.nextToken()

	withClause := &ast.WithCtesAndXmlNamespaces{}

	// Parse XMLNAMESPACES, CHANGE_TRACKING_CONTEXT or CTEs
	for {
		if strings.ToUpper(p.curTok.Literal) == "XMLNAMESPACES" {
			p.nextToken() // consume XMLNAMESPACES
			xmlNs := &ast.XmlNamespaces{}
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					// Check for DEFAULT element
					if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
						p.nextToken() // consume DEFAULT
						strLit, _ := p.parseStringLiteral()
						elem := &ast.XmlNamespacesDefaultElement{String: strLit}
						xmlNs.XmlNamespacesElements = append(xmlNs.XmlNamespacesElements, elem)
					} else {
						// Alias element: string AS identifier
						strLit, _ := p.parseStringLiteral()
						elem := &ast.XmlNamespacesAliasElement{String: strLit}
						if p.curTok.Type == TokenAs {
							p.nextToken() // consume AS
							elem.Identifier = p.parseIdentifier()
						}
						xmlNs.XmlNamespacesElements = append(xmlNs.XmlNamespacesElements, elem)
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
			withClause.XmlNamespaces = xmlNs
		} else if strings.ToUpper(p.curTok.Literal) == "CHANGE_TRACKING_CONTEXT" {
			p.nextToken() // consume CHANGE_TRACKING_CONTEXT
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				expr, _ := p.parseScalarExpression()
				withClause.ChangeTrackingContext = expr
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
		} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			// Parse CTE: name (columns) AS (query)
			cte := &ast.CommonTableExpression{
				ExpressionName: p.parseIdentifier(),
			}

			// Parse optional column list
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					cte.Columns = append(cte.Columns, p.parseIdentifier())
					if p.curTok.Type == TokenComma {
						p.nextToken()
					}
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}

			// Expect AS
			if p.curTok.Type == TokenAs {
				p.nextToken() // consume AS
			}

			// Parse query in parentheses
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				queryExpr, err := p.parseQueryExpression()
				if err != nil {
					return nil, err
				}
				cte.QueryExpression = queryExpr
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}

			withClause.CommonTableExpressions = append(withClause.CommonTableExpressions, cte)
		} else {
			break
		}

		// Check for comma (more CTEs or XMLNAMESPACES followed by CTEs)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Now dispatch to the appropriate statement parser
	switch p.curTok.Type {
	case TokenInsert:
		stmt, err := p.parseInsertStatement()
		if err != nil {
			return nil, err
		}
		if ins, ok := stmt.(*ast.InsertStatement); ok {
			ins.WithCtesAndXmlNamespaces = withClause
		}
		return stmt, nil
	case TokenUpdate:
		stmt, err := p.parseUpdateOrUpdateStatisticsStatement()
		if err != nil {
			return nil, err
		}
		if upd, ok := stmt.(*ast.UpdateStatement); ok {
			upd.WithCtesAndXmlNamespaces = withClause
		}
		return stmt, nil
	case TokenDelete:
		stmt, err := p.parseDeleteStatement()
		if err != nil {
			return nil, err
		}
		stmt.WithCtesAndXmlNamespaces = withClause
		return stmt, nil
	case TokenSelect:
		stmt, err := p.parseSelectStatement()
		if err != nil {
			return nil, err
		}
		stmt.WithCtesAndXmlNamespaces = withClause
		return stmt, nil
	}

	return nil, fmt.Errorf("expected INSERT, UPDATE, DELETE, or SELECT after WITH clause, got %s", p.curTok.Literal)
}

func (p *Parser) parseInsertStatement() (ast.Statement, error) {
	// Consume INSERT
	p.nextToken()

	// Check for INSERT BULK
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "BULK" {
		return p.parseInsertBulkStatement()
	}

	stmt := &ast.InsertStatement{
		InsertSpecification: &ast.InsertSpecification{
			InsertOption: "None",
		},
	}

	// Check for TOP clause
	if p.curTok.Type == TokenTop {
		top, err := p.parseTopRowFilter()
		if err != nil {
			return nil, err
		}
		stmt.InsertSpecification.TopRowFilter = top
	}

	// Check for INTO or OVER
	if p.curTok.Type == TokenInto {
		stmt.InsertSpecification.InsertOption = "Into"
		p.nextToken()
	} else if p.curTok.Type == TokenOver {
		stmt.InsertSpecification.InsertOption = "Over"
		p.nextToken()
	}

	// Parse target - use parseInsertTarget which doesn't treat () as function params
	target, err := p.parseInsertTarget()
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

	// Parse OUTPUT clauses (can have OUTPUT INTO followed by OUTPUT)
	for p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OUTPUT" {
		outputClause, outputIntoClause, err := p.parseOutputClause()
		if err != nil {
			return nil, err
		}
		if outputIntoClause != nil {
			stmt.InsertSpecification.OutputIntoClause = outputIntoClause
		}
		if outputClause != nil {
			stmt.InsertSpecification.OutputClause = outputClause
		}
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

	// Check for function call (has parentheses) - used by UPDATE/DELETE targets
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

// parseInsertTarget parses the target for INSERT statements.
// Unlike parseDMLTarget, it does NOT treat parentheses as function parameters
// because in INSERT statements, parentheses after the table name are column names.
func (p *Parser) parseInsertTarget() (ast.TableReference, error) {
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

	// For INSERT targets, parentheses are column names, not function parameters
	// So we don't parse them here - the caller handles the column list

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

	// Check for alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		ref.Alias = p.parseIdentifier()
	} else if p.curTok.Type == TokenIdent {
		// Alias without AS - but need to check it's not a keyword
		upper := strings.ToUpper(p.curTok.Literal)
		if upper != "SELECT" && upper != "VALUES" && upper != "DEFAULT" && upper != "OUTPUT" && upper != "EXEC" && upper != "EXECUTE" {
			ref.Alias = p.parseIdentifier()
		}
	}

	return ref, nil
}

func (p *Parser) parseOpenRowset() (ast.TableReference, error) {
	// Consume OPENROWSET
	p.nextToken()

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after OPENROWSET, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for BULK form
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "BULK" {
		return p.parseBulkOpenRowset()
	}

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

func (p *Parser) parseBulkOpenRowset() (*ast.BulkOpenRowset, error) {
	// We're positioned on BULK, consume it
	p.nextToken()

	result := &ast.BulkOpenRowset{
		ForPath: false,
	}

	// Parse data file(s) - could be a single string or parenthesized list
	if p.curTok.Type == TokenLParen {
		// Multiple data files
		p.nextToken()
		for {
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			result.DataFiles = append(result.DataFiles, expr)

			if p.curTok.Type == TokenComma {
				p.nextToken()
				// Allow trailing comma
				if p.curTok.Type == TokenRParen {
					break
				}
				continue
			}
			break
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after data files, got %s", p.curTok.Literal)
		}
		p.nextToken()
	} else {
		// Single data file
		expr, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		result.DataFiles = append(result.DataFiles, expr)
	}

	// Parse options (comma-separated)
	for p.curTok.Type == TokenComma {
		p.nextToken()
		opt, err := p.parseOpenRowsetBulkOption()
		if err != nil {
			return nil, err
		}
		result.Options = append(result.Options, opt)
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after OPENROWSET BULK, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse optional alias
	if p.curTok.Type == TokenAs {
		p.nextToken()
		if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
			result.Alias = p.parseIdentifier()
		}
	}

	// Parse optional column list (e.g., AS a(c1, c2))
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for {
			if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
				result.Columns = append(result.Columns, p.parseIdentifier())
			}
			if p.curTok.Type == TokenComma {
				p.nextToken()
				continue
			}
			break
		}
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after column list, got %s", p.curTok.Literal)
		}
		p.nextToken()
	}

	return result, nil
}

func (p *Parser) parseOpenRowsetBulkOption() (ast.BulkInsertOption, error) {
	upper := strings.ToUpper(p.curTok.Literal)

	// Handle simple options (SINGLE_BLOB, SINGLE_CLOB, SINGLE_NCLOB)
	switch upper {
	case "SINGLE_BLOB":
		p.nextToken()
		return &ast.BulkInsertOptionBase{OptionKind: "SingleBlob"}, nil
	case "SINGLE_CLOB":
		p.nextToken()
		return &ast.BulkInsertOptionBase{OptionKind: "SingleClob"}, nil
	case "SINGLE_NCLOB":
		p.nextToken()
		return &ast.BulkInsertOptionBase{OptionKind: "SingleNClob"}, nil
	}

	// Handle ORDER option
	if upper == "ORDER" {
		p.nextToken()
		return p.parseOpenRowsetOrderOption()
	}

	// Handle KEY=VALUE options
	optionKind := p.getOpenRowsetOptionKind(upper)
	p.nextToken()

	if p.curTok.Type == TokenEquals {
		p.nextToken()
		value, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.LiteralBulkInsertOption{
			OptionKind: optionKind,
			Value:      value,
		}, nil
	}

	return &ast.BulkInsertOptionBase{OptionKind: optionKind}, nil
}

func (p *Parser) getOpenRowsetOptionKind(name string) string {
	optionMap := map[string]string{
		"FORMATFILE":       "FormatFile",
		"FORMAT":           "DataFileFormat",
		"CODEPAGE":         "CodePage",
		"ROWS_PER_BATCH":   "RowsPerBatch",
		"LASTROW":          "LastRow",
		"FIRSTROW":         "FirstRow",
		"MAXERRORS":        "MaxErrors",
		"ERRORFILE":        "ErrorFile",
		"FIELDQUOTE":       "FieldQuote",
		"FIELDTERMINATOR":  "FieldTerminator",
		"ROWTERMINATOR":    "RowTerminator",
		"ESCAPECHAR":       "EscapeChar",
		"DATA_COMPRESSION": "DataCompression",
		"PARSER_VERSION":   "ParserVersion",
		"HEADER_ROW":       "HeaderRow",
		"DATAFILETYPE":     "DataFileType",
		"ROWSET_OPTIONS":   "RowsetOptions",
	}
	if kind, ok := optionMap[name]; ok {
		return kind
	}
	return name
}

func (p *Parser) parseOpenRowsetOrderOption() (*ast.OrderBulkInsertOption, error) {
	result := &ast.OrderBulkInsertOption{
		OptionKind: "Order",
	}

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after ORDER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse column list with sort order
	for {
		col := &ast.ColumnWithSortOrder{
			SortOrder: ast.SortOrderNotSpecified,
		}

		// Parse column reference
		colRef, err := p.parseMultiPartIdentifierAsColumn()
		if err != nil {
			return nil, err
		}
		col.Column = colRef

		// Check for ASC/DESC
		if p.curTok.Type == TokenAsc {
			col.SortOrder = ast.SortOrderAscending
			p.nextToken()
		} else if p.curTok.Type == TokenDesc {
			col.SortOrder = ast.SortOrderDescending
			p.nextToken()
		}

		result.Columns = append(result.Columns, col)

		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}
		break
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after ORDER columns, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for UNIQUE
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "UNIQUE" {
		result.IsUnique = true
		p.nextToken()
	}

	return result, nil
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

func (p *Parser) parseTableHints() ([]ast.TableHintType, error) {
	// Consume WITH
	p.nextToken()

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
	}
	p.nextToken()

	var hints []ast.TableHintType
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		hint, err := p.parseTableHint()
		if err != nil {
			return nil, err
		}
		if hint != nil {
			hints = append(hints, hint)
		}
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else if p.curTok.Type != TokenRParen {
			// Check if the next token is a valid table hint (space-separated hints)
			if p.curTok.Type == TokenIdent && isTableHintKeyword(strings.ToUpper(p.curTok.Literal)) {
				continue // Continue parsing space-separated hints
			}
			break
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return hints, nil
}

// isTableHintKeyword checks if a string is a valid table hint keyword
func isTableHintKeyword(name string) bool {
	switch name {
	case "HOLDLOCK", "NOLOCK", "PAGLOCK", "READCOMMITTED", "READPAST",
		"READUNCOMMITTED", "REPEATABLEREAD", "ROWLOCK", "SERIALIZABLE",
		"SNAPSHOT", "TABLOCK", "TABLOCKX", "UPDLOCK", "XLOCK", "NOWAIT",
		"INDEX", "FORCESEEK", "FORCESCAN", "KEEPIDENTITY", "KEEPDEFAULTS",
		"IGNORE_CONSTRAINTS", "IGNORE_TRIGGERS", "NOEXPAND", "SPATIAL_WINDOW_MAX_CELLS":
		return true
	}
	return false
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
		// Check for pseudo columns
		lit := p.curTok.Literal
		upperLit := strings.ToUpper(lit)
		if upperLit == "$ACTION" {
			cols = append(cols, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnAction"})
			p.nextToken()
		} else if upperLit == "$CUID" {
			cols = append(cols, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnCuid"})
			p.nextToken()
		} else if upperLit == "$ROWGUID" {
			cols = append(cols, &ast.ColumnReferenceExpression{ColumnType: "PseudoColumnRowGuid"})
			p.nextToken()
		} else {
			col, err := p.parseMultiPartIdentifierAsColumn()
			if err != nil {
				return nil, err
			}
			cols = append(cols, col)
		}

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

	// Check for EXECUTE ('string') form - ExecutableStringList
	if p.curTok.Type == TokenLParen {
		strList, err := p.parseExecutableStringList()
		if err != nil {
			return nil, err
		}
		spec.ExecutableEntity = strList

		// Parse optional AS USER/LOGIN context
		if p.curTok.Type == TokenAs {
			ctx, err := p.parseExecuteContextForSpec()
			if err != nil {
				return nil, err
			}
			spec.ExecuteContext = ctx
		}

		// Parse optional AT LinkedServer
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "AT" {
			p.nextToken()
			spec.LinkedServer = p.parseIdentifier()
		}

		return spec, nil
	}

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

	// Check for OPENDATASOURCE or OPENROWSET
	upperLit := strings.ToUpper(p.curTok.Literal)
	if upperLit == "OPENDATASOURCE" || upperLit == "OPENROWSET" {
		p.nextToken() // consume OPENDATASOURCE/OPENROWSET
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (

			// Parse provider name
			var providerName *ast.StringLiteral
			if p.curTok.Type == TokenString {
				providerName = p.parseStringLiteralValue()
				p.nextToken()
			}

			// Expect comma
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}

			// Parse init string
			var initString *ast.StringLiteral
			if p.curTok.Type == TokenString {
				initString = p.parseStringLiteralValue()
				p.nextToken()
			}

			// Expect )
			if p.curTok.Type == TokenRParen {
				p.nextToken()
			}

			procRef.AdHocDataSource = &ast.AdHocDataSource{
				ProviderName: providerName,
				InitString:   initString,
			}

			// Expect . and then schema.object.procedure name
			if p.curTok.Type == TokenDot {
				p.nextToken() // consume .
			}
		}
	}

	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		// Procedure variable
		procRef.ProcedureReference = &ast.ProcedureReferenceName{
			ProcedureVariable: &ast.VariableReference{Name: p.curTok.Literal},
		}
		p.nextToken()
	} else if p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		// Procedure name
		son, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		pr := &ast.ProcedureReference{Name: son}

		// Check for procedure number: ;number
		if p.curTok.Type == TokenSemicolon {
			p.nextToken() // consume ;
			if p.curTok.Type == TokenNumber {
				pr.Number = &ast.IntegerLiteral{
					LiteralType: "Integer",
					Value:       p.curTok.Literal,
				}
				p.nextToken()
			}
		}

		procRef.ProcedureReference = &ast.ProcedureReferenceName{
			ProcedureReference: pr,
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

func (p *Parser) parseExecutableStringList() (*ast.ExecutableStringList, error) {
	// We're positioned on (, consume it
	p.nextToken()

	strList := &ast.ExecutableStringList{}

	// Parse the string expressions (may be strings, variables, or concatenations with +)
	// Strings are added to Strings, other parameters (after comma) go to Parameters
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		isVariable := p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@")
		if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString || isVariable {
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			// Flatten concatenated expressions to individual parts for the Strings array
			p.flattenStringExpression(expr, &strList.Strings)
		} else {
			break
		}

		// Check for comma or closing paren
		if p.curTok.Type == TokenComma {
			p.nextToken()
			// After comma, we switch to parsing parameters
			break
		}
		if p.curTok.Type == TokenRParen {
			break
		}
	}

	// Parse parameters (after the first comma following strings)
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param, err := p.parseExecuteParameter()
		if err != nil {
			return nil, err
		}
		strList.Parameters = append(strList.Parameters, param)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after EXECUTE string list, got %s", p.curTok.Literal)
	}
	p.nextToken()

	return strList, nil
}

func (p *Parser) flattenStringExpression(expr ast.ScalarExpression, strings *[]ast.ScalarExpression) {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		// Recursively flatten for + concatenation
		p.flattenStringExpression(e.FirstExpression, strings)
		p.flattenStringExpression(e.SecondExpression, strings)
	default:
		*strings = append(*strings, expr)
	}
}

func (p *Parser) parseExecuteContextForSpec() (*ast.ExecuteContext, error) {
	// We're positioned on AS, consume it
	p.nextToken()

	ctx := &ast.ExecuteContext{}

	upper := strings.ToUpper(p.curTok.Literal)
	switch upper {
	case "USER":
		ctx.Kind = "User"
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			ctx.Principal = expr
		}
	case "LOGIN":
		ctx.Kind = "Login"
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			ctx.Principal = expr
		}
	case "CALLER":
		ctx.Kind = "Caller"
		p.nextToken()
	case "OWNER":
		ctx.Kind = "Owner"
		p.nextToken()
	case "SELF":
		ctx.Kind = "Self"
		p.nextToken()
	default:
		return nil, fmt.Errorf("expected USER, LOGIN, CALLER, OWNER, or SELF after AS, got %s", p.curTok.Literal)
	}

	return ctx, nil
}

func (p *Parser) parseExecuteParameter() (*ast.ExecuteParameter, error) {
	param := &ast.ExecuteParameter{IsOutput: false}

	// Check for DEFAULT keyword
	if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
		param.ParameterValue = &ast.DefaultLiteral{LiteralType: "Default", Value: "DEFAULT"}
		p.nextToken()
		return param, nil
	}

	// Check for named parameter: @name = value
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		varName := p.curTok.Literal
		p.nextToken()

		if p.curTok.Type == TokenEquals {
			// Named parameter
			p.nextToken() // consume =
			param.Variable = &ast.VariableReference{Name: varName}

			// Check for DEFAULT keyword as value
			if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
				param.ParameterValue = &ast.DefaultLiteral{LiteralType: "Default", Value: "DEFAULT"}
				p.nextToken()
			} else {
				// Parse the parameter value
				expr, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				param.ParameterValue = expr
			}
		} else {
			// Just a variable as value (not a named parameter)
			param.ParameterValue = &ast.VariableReference{Name: varName}
		}
	} else {
		// Check for bare identifier as IdentifierLiteral (e.g., EXEC sp_addtype birthday, datetime)
		// Only if it's not followed by . or ( which would indicate a column/function reference
		if p.curTok.Type == TokenIdent && !strings.HasPrefix(p.curTok.Literal, "@") {
			upper := strings.ToUpper(p.curTok.Literal)
			// Skip keywords that are expression starters
			isKeyword := upper == "NULL" || upper == "DEFAULT" || upper == "NOT" ||
				upper == "CASE" || upper == "EXISTS" || upper == "CAST" ||
				upper == "CONVERT" || upper == "COALESCE" || upper == "NULLIF"
			if !isKeyword && p.peekTok.Type != TokenDot && p.peekTok.Type != TokenLParen {
				// Plain identifier - treat as IdentifierLiteral
				quoteType := "NotQuoted"
				if strings.HasPrefix(p.curTok.Literal, "[") {
					quoteType = "SquareBracket"
				}
				param.ParameterValue = &ast.IdentifierLiteral{
					LiteralType: "Identifier",
					QuoteType:   quoteType,
					Value:       p.curTok.Literal,
				}
				p.nextToken()
			} else {
				// Regular value expression
				expr, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				param.ParameterValue = expr
			}
		} else {
			// Regular value expression
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			param.ParameterValue = expr
		}
	}

	// Check for OUTPUT modifier
	if strings.ToUpper(p.curTok.Literal) == "OUTPUT" || strings.ToUpper(p.curTok.Literal) == "OUT" {
		param.IsOutput = true
		p.nextToken()
	}

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

	// Check for TOP clause
	if p.curTok.Type == TokenTop {
		top, err := p.parseTopRowFilter()
		if err != nil {
			return nil, err
		}
		stmt.UpdateSpecification.TopRowFilter = top
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
		clause, err := p.parseSetClause()
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

func (p *Parser) parseSetClause() (ast.SetClause, error) {
	// First, try to detect if this is a function call set clause
	// e.g., SET a.b.c.d.func() or SET a.b.c.d.func(args)

	// Variables start with @ and are never function call set clauses
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		return p.parseAssignmentSetClause()
	}

	// Check for $ROWGUID pseudo-column - always assignment
	if p.curTok.Type == TokenIdent && strings.EqualFold(p.curTok.Literal, "$ROWGUID") {
		return p.parseAssignmentSetClause()
	}

	// Parse multi-part identifier and look ahead for ( or =
	identifiers := []*ast.Identifier{}
	for {
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
		return nil, fmt.Errorf("expected identifier in SET clause")
	}

	// If followed by ( it's a function call set clause
	if p.curTok.Type == TokenLParen {
		// The last identifier is the function name, the rest form the call target
		if len(identifiers) < 2 {
			// Need at least object.func()
			return nil, fmt.Errorf("expected at least 2 identifiers for function call SET clause")
		}

		funcName := identifiers[len(identifiers)-1]
		targetIds := identifiers[:len(identifiers)-1]

		p.nextToken() // consume (

		// Parse parameters
		var params []ast.ScalarExpression
		if p.curTok.Type != TokenRParen {
			for {
				param, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				params = append(params, param)

				if p.curTok.Type != TokenComma {
					break
				}
				p.nextToken()
			}
		}

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) in function call SET clause, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		fc := &ast.FunctionCall{
			CallTarget: &ast.MultiPartIdentifierCallTarget{
				MultiPartIdentifier: &ast.MultiPartIdentifier{
					Count:       len(targetIds),
					Identifiers: targetIds,
				},
			},
			FunctionName:     funcName,
			Parameters:       params,
			UniqueRowFilter:  "NotSpecified",
			WithArrayWrapper: false,
		}

		return &ast.FunctionCallSetClause{MutatorFunction: fc}, nil
	}

	// Otherwise, it's an assignment set clause
	// Convert identifiers to ColumnReferenceExpression
	clause := &ast.AssignmentSetClause{
		AssignmentKind: "Equals",
		Column: &ast.ColumnReferenceExpression{
			ColumnType: "Regular",
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Count:       len(identifiers),
				Identifiers: identifiers,
			},
		},
	}

	if p.isCompoundAssignment() {
		clause.AssignmentKind = p.getAssignmentKind()
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected =, got %s", p.curTok.Literal)
	}

	val, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	clause.NewValue = val

	return clause, nil
}

// isCompoundAssignment checks if the current token is a compound assignment operator
func (p *Parser) isCompoundAssignment() bool {
	switch p.curTok.Type {
	case TokenEquals, TokenConcatEquals, TokenPlusEquals, TokenMinusEquals,
		TokenStarEquals, TokenSlashEquals, TokenModuloEquals,
		TokenAndEquals, TokenOrEquals, TokenXorEquals:
		return true
	}
	return false
}

// getAssignmentKind returns the AssignmentKind for the current compound assignment token
func (p *Parser) getAssignmentKind() string {
	switch p.curTok.Type {
	case TokenEquals:
		return "Equals"
	case TokenConcatEquals:
		return "ConcatEquals"
	case TokenPlusEquals:
		return "AddEquals"
	case TokenMinusEquals:
		return "SubtractEquals"
	case TokenStarEquals:
		return "MultiplyEquals"
	case TokenSlashEquals:
		return "DivideEquals"
	case TokenModuloEquals:
		return "ModEquals"
	case TokenAndEquals:
		return "BitwiseAndEquals"
	case TokenOrEquals:
		return "BitwiseOrEquals"
	case TokenXorEquals:
		return "BitwiseXorEquals"
	}
	return "Equals"
}

func (p *Parser) parseAssignmentSetClause() (*ast.AssignmentSetClause, error) {
	clause := &ast.AssignmentSetClause{AssignmentKind: "Equals"}

	// Could be @var = col = value, @var = value, @var ||= value, or col = value, col ||= value
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		varName := p.curTok.Literal
		p.nextToken()
		if p.isCompoundAssignment() {
			clause.AssignmentKind = p.getAssignmentKind()
			clause.Variable = &ast.VariableReference{Name: varName}
			p.nextToken()

			// Check if next is column = value or column ||= value (SET @a = col = value)
			if p.curTok.Type == TokenIdent && !strings.HasPrefix(p.curTok.Literal, "@") {
				// Could be @a = col = value, @a = col ||= value or @a = expr
				savedTok := p.curTok
				col, err := p.parseMultiPartIdentifierAsColumn()
				if err != nil {
					return nil, err
				}
				if p.isCompoundAssignment() {
					clause.AssignmentKind = p.getAssignmentKind()
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

			// Just @var = value or @var ||= value
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			clause.NewValue = val
			return clause, nil
		}
	}

	// Check for $ROWGUID pseudo-column
	if p.curTok.Type == TokenIdent && strings.EqualFold(p.curTok.Literal, "$ROWGUID") {
		clause.Column = &ast.ColumnReferenceExpression{
			ColumnType: "PseudoColumnRowGuid",
		}
		p.nextToken()
	} else {
		// col = value or col ||= value
		col, err := p.parseMultiPartIdentifierAsColumn()
		if err != nil {
			return nil, err
		}
		clause.Column = col
	}

	if p.isCompoundAssignment() {
		clause.AssignmentKind = p.getAssignmentKind()
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected =, got %s", p.curTok.Literal)
	}

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

	// Parse optional TOP clause
	if p.curTok.Type == TokenTop {
		topRowFilter, err := p.parseTopRowFilter()
		if err != nil {
			return nil, err
		}
		stmt.DeleteSpecification.TopRowFilter = topRowFilter
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

	// Parse OUTPUT clauses (can have OUTPUT INTO followed by OUTPUT)
	for p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OUTPUT" {
		outputClause, outputIntoClause, err := p.parseOutputClause()
		if err != nil {
			return nil, err
		}
		if outputIntoClause != nil {
			stmt.DeleteSpecification.OutputIntoClause = outputIntoClause
		}
		if outputClause != nil {
			stmt.DeleteSpecification.OutputClause = outputClause
		}
	}

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

func (p *Parser) parseInsertBulkStatement() (*ast.InsertBulkStatement, error) {
	// Consume BULK
	p.nextToken()

	stmt := &ast.InsertBulkStatement{}

	// Parse table name
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.To = son

	// Parse optional column definitions (col type [NULL|NOT NULL], ...)
	if p.curTok.Type == TokenLParen {
		colDefs, err := p.parseInsertBulkColumnDefinitions()
		if err != nil {
			return nil, err
		}
		stmt.ColumnDefinitions = colDefs
	}

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		options, err := p.parseBulkInsertOptions()
		if err != nil {
			return nil, err
		}
		stmt.Options = options
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseInsertBulkColumnDefinitions() ([]*ast.InsertBulkColumnDefinition, error) {
	// Consume (
	p.nextToken()

	var colDefs []*ast.InsertBulkColumnDefinition
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		colDef, err := p.parseInsertBulkColumnDefinition()
		if err != nil {
			return nil, err
		}
		colDefs = append(colDefs, colDef)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return colDefs, nil
}

func (p *Parser) parseInsertBulkColumnDefinition() (*ast.InsertBulkColumnDefinition, error) {
	colDef := &ast.InsertBulkColumnDefinition{
		Column:      &ast.ColumnDefinitionBase{},
		NullNotNull: "Unspecified",
	}

	// Parse column name
	if p.curTok.Type != TokenIdent {
		return nil, fmt.Errorf("expected column name, got %s", p.curTok.Literal)
	}
	colDef.Column.ColumnIdentifier = p.parseIdentifier()

	// Check for data type or timestamp keyword
	if p.curTok.Type == TokenIdent {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "TIMESTAMP" {
			// timestamp is a special case - no data type
			p.nextToken()
		} else {
			// Parse data type
			dataType, err := p.parseDataTypeReference()
			if err != nil {
				return nil, err
			}
			colDef.Column.DataType = dataType
		}
	}

	// Check for NULL or NOT NULL
	if p.curTok.Type == TokenNull {
		colDef.NullNotNull = "Null"
		p.nextToken()
	} else if p.curTok.Type == TokenNot {
		p.nextToken()
		if p.curTok.Type == TokenNull {
			colDef.NullNotNull = "NotNull"
			p.nextToken()
		}
	}

	return colDef, nil
}

func (p *Parser) parseBulkInsertOptions() ([]ast.BulkInsertOption, error) {
	// Consume WITH
	p.nextToken()

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
	}
	p.nextToken()

	var options []ast.BulkInsertOption
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		opt, err := p.parseBulkInsertOption()
		if err != nil {
			return nil, err
		}
		options = append(options, opt)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return options, nil
}

func (p *Parser) parseBulkInsertOption() (ast.BulkInsertOption, error) {
	if p.curTok.Type != TokenIdent && p.curTok.Type != TokenOrder {
		return nil, fmt.Errorf("expected option name, got %s", p.curTok.Literal)
	}

	optionName := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	// Handle ORDER option specially
	if optionName == "ORDER" {
		return p.parseOrderBulkInsertOption()
	}

	// Map option names to OptionKind values
	optionKindMap := map[string]string{
		"CHECK_CONSTRAINTS":   "CheckConstraints",
		"FIRE_TRIGGERS":       "FireTriggers",
		"KEEPNULLS":           "KeepNulls",
		"TABLOCK":             "TabLock",
		"NO_TRIGGERS":         "NoTriggers",
		"KEEPIDENTITY":        "KeepIdentity",
		"INCLUDE_HIDDEN":      "IncludeHidden",
		"BATCHSIZE":           "BatchSize",
		"CODEPAGE":            "CodePage",
		"DATAFILETYPE":        "DataFileType",
		"FIELDTERMINATOR":     "FieldTerminator",
		"FIRSTROW":            "FirstRow",
		"FORMATFILE":          "FormatFile",
		"KILOBYTES_PER_BATCH": "KilobytesPerBatch",
		"LASTROW":             "LastRow",
		"MAXERRORS":           "MaxErrors",
		"ROWTERMINATOR":       "RowTerminator",
		"ROWS_PER_BATCH":      "RowsPerBatch",
		"ERRORFILE":           "ErrorFile",
		"FORMAT":              "DataFileFormat",
		"ESCAPECHAR":          "EscapeChar",
		"FIELDQUOTE":          "FieldQuote",
	}

	optionKind := optionKindMap[optionName]
	if optionKind == "" {
		optionKind = optionName
	}

	// Check for = value
	if p.curTok.Type == TokenEquals {
		p.nextToken()
		value, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.LiteralBulkInsertOption{
			OptionKind: optionKind,
			Value:      value,
		}, nil
	}

	// Simple option without value
	return &ast.BulkInsertOptionBase{
		OptionKind: optionKind,
	}, nil
}

func (p *Parser) parseOrderBulkInsertOption() (*ast.OrderBulkInsertOption, error) {
	opt := &ast.OrderBulkInsertOption{
		OptionKind: "Order",
	}

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after ORDER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		col, err := p.parseMultiPartIdentifierAsColumn()
		if err != nil {
			return nil, err
		}

		sortOrder := ast.SortOrderNotSpecified
		if p.curTok.Type == TokenAsc {
			sortOrder = ast.SortOrderAscending
			p.nextToken()
		} else if p.curTok.Type == TokenDesc {
			sortOrder = ast.SortOrderDescending
			p.nextToken()
		}

		opt.Columns = append(opt.Columns, &ast.ColumnWithSortOrder{
			Column:    col,
			SortOrder: sortOrder,
		})

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken()
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	return opt, nil
}

func (p *Parser) parseBulkInsertStatement() (*ast.BulkInsertStatement, error) {
	// BULK has already been consumed, now we expect INSERT
	if p.curTok.Type != TokenInsert {
		return nil, fmt.Errorf("expected INSERT after BULK, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.BulkInsertStatement{}

	// Parse table name
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.To = son

	// Expect FROM
	if p.curTok.Type != TokenFrom {
		return nil, fmt.Errorf("expected FROM, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse FROM expression (string or identifier)
	fromExpr, err := p.parseIdentifierOrValueExpression()
	if err != nil {
		return nil, err
	}
	stmt.From = fromExpr

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		options, err := p.parseBulkInsertOptions()
		if err != nil {
			return nil, err
		}
		stmt.Options = options
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseIdentifierOrValueExpression() (*ast.IdentifierOrValueExpression, error) {
	result := &ast.IdentifierOrValueExpression{}

	if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
		// String literal
		strLit, _ := p.parseStringLiteral()
		result.Value = strLit.Value
		result.ValueExpression = strLit
	} else if p.curTok.Type == TokenNumber {
		// Integer literal
		result.Value = p.curTok.Literal
		result.ValueExpression = &ast.IntegerLiteral{
			LiteralType: "Integer",
			Value:       p.curTok.Literal,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenBinary {
		// Binary/hex literal
		result.Value = p.curTok.Literal
		result.ValueExpression = &ast.BinaryLiteral{
			LiteralType:   "Binary",
			IsLargeObject: false,
			Value:         p.curTok.Literal,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent {
		// Identifier - use parseIdentifier to handle bracketed identifiers properly
		ident := p.parseIdentifier()
		result.Value = ident.Value
		result.Identifier = ident
	} else if p.curTok.Type == TokenEOF {
		// Handle incomplete statement - return empty identifier
		result.Value = ""
		result.Identifier = &ast.Identifier{
			Value:     "",
			QuoteType: "NotQuoted",
		}
	} else {
		return nil, fmt.Errorf("expected identifier or value, got %s", p.curTok.Literal)
	}

	return result, nil
}

// parseUpdateOrUpdateStatisticsStatement routes to UPDATE or UPDATE STATISTICS.
func (p *Parser) parseUpdateOrUpdateStatisticsStatement() (ast.Statement, error) {
	// Consume UPDATE
	p.nextToken()

	// Check for UPDATE STATISTICS
	if p.curTok.Type == TokenStats || strings.ToUpper(p.curTok.Literal) == "STATISTICS" {
		return p.parseUpdateStatisticsStatementContinued()
	}

	// Otherwise, parse normal UPDATE statement
	stmt := &ast.UpdateStatement{
		UpdateSpecification: &ast.UpdateSpecification{},
	}

	// Check for TOP clause
	if p.curTok.Type == TokenTop {
		top, err := p.parseTopRowFilter()
		if err != nil {
			return nil, err
		}
		stmt.UpdateSpecification.TopRowFilter = top
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

	// Parse OUTPUT clauses (can have OUTPUT INTO followed by OUTPUT)
	for p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OUTPUT" {
		outputClause, outputIntoClause, err := p.parseOutputClause()
		if err != nil {
			return nil, err
		}
		if outputIntoClause != nil {
			stmt.UpdateSpecification.OutputIntoClause = outputIntoClause
		}
		if outputClause != nil {
			stmt.UpdateSpecification.OutputClause = outputClause
		}
	}

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

// parseUpdateStatisticsStatementContinued parses UPDATE STATISTICS after consuming UPDATE.
func (p *Parser) parseUpdateStatisticsStatementContinued() (*ast.UpdateStatisticsStatement, error) {
	// Consume STATISTICS
	p.nextToken()

	stmt := &ast.UpdateStatisticsStatement{}

	// Parse table name
	schemaObjectName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.SchemaObjectName = schemaObjectName

	// Parse optional SubElements (stat/index names)
	// Can be either in parentheses: (c1, c2, c3) or a single identifier: st1
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			ident := p.parseIdentifier()
			stmt.SubElements = append(stmt.SubElements, ident)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	} else if p.curTok.Type == TokenIdent {
		// Single identifier without parentheses
		ident := p.parseIdentifier()
		stmt.SubElements = append(stmt.SubElements, ident)
	}

	// Parse optional WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF {
			optionName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			switch optionName {
			case "ALL":
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.SimpleStatisticsOption{
					OptionKind: "All",
				})
			case "FULLSCAN":
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.SimpleStatisticsOption{
					OptionKind: "FullScan",
				})
			case "NORECOMPUTE":
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.SimpleStatisticsOption{
					OptionKind: "NoRecompute",
				})
			case "COLUMNS":
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.SimpleStatisticsOption{
					OptionKind: "Columns",
				})
			case "INDEX":
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.SimpleStatisticsOption{
					OptionKind: "Index",
				})
			case "ROWCOUNT":
				// Parse = value
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				val := p.curTok.Literal
				p.nextToken()
				// Use NumericLiteral for very large numbers, IntegerLiteral otherwise
				var literal ast.ScalarExpression
				if len(val) > 18 { // Numbers > 18 digits are likely > MaxInt64
					literal = &ast.NumericLiteral{LiteralType: "Numeric", Value: val}
				} else {
					literal = &ast.IntegerLiteral{LiteralType: "Integer", Value: val}
				}
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.LiteralStatisticsOption{
					OptionKind: "RowCount",
					Literal:    literal,
				})
			case "PAGECOUNT":
				// Parse = value
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				val := p.curTok.Literal
				p.nextToken()
				// Use NumericLiteral for very large numbers, IntegerLiteral otherwise
				var literal ast.ScalarExpression
				if len(val) > 18 { // Numbers > 18 digits are likely > MaxInt64
					literal = &ast.NumericLiteral{LiteralType: "Numeric", Value: val}
				} else {
					literal = &ast.IntegerLiteral{LiteralType: "Integer", Value: val}
				}
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.LiteralStatisticsOption{
					OptionKind: "PageCount",
					Literal:    literal,
				})
			case "SAMPLE":
				// Parse number PERCENT/ROWS
				value, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				mode := strings.ToUpper(p.curTok.Literal)
				p.nextToken() // consume PERCENT or ROWS
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.LiteralStatisticsOption{
					OptionKind: "Sample" + strings.Title(strings.ToLower(mode)),
					Literal:    value,
				})
			case "RESAMPLE":
				resampleOpt := &ast.ResampleStatisticsOption{
					OptionKind: "Resample",
				}
				// Check for ON PARTITIONS
				if p.curTok.Type == TokenOn {
					p.nextToken() // consume ON
					if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
						p.nextToken() // consume PARTITIONS
						if p.curTok.Type == TokenLParen {
							p.nextToken() // consume (
							for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
								// Parse partition range: number or number TO number
								// Just parse the literal value directly
								fromVal := &ast.IntegerLiteral{
									LiteralType: "Integer",
									Value:       p.curTok.Literal,
								}
								p.nextToken() // consume the number
								partRange := &ast.StatisticsPartitionRange{
									From: fromVal,
								}
								// Check for TO (TokenTo)
								if p.curTok.Type == TokenTo {
									p.nextToken() // consume TO
									toVal := &ast.IntegerLiteral{
										LiteralType: "Integer",
										Value:       p.curTok.Literal,
									}
									p.nextToken() // consume the number
									partRange.To = toVal
								}
								resampleOpt.Partitions = append(resampleOpt.Partitions, partRange)
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
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, resampleOpt)
			case "INCREMENTAL":
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					state := strings.ToUpper(p.curTok.Literal)
					optionState := "On"
					if state == "OFF" {
						optionState = "Off"
					}
					p.nextToken()
					stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.OnOffStatisticsOption{
						OptionKind:  "Incremental",
						OptionState: optionState,
					})
				} else {
					stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.OnOffStatisticsOption{
						OptionKind:  "Incremental",
						OptionState: "On",
					})
				}
			default:
				// Unknown option, skip
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

	return stmt, nil
}

// parseOutputClause parses an OUTPUT clause (with optional INTO).
// Returns (outputClause, outputIntoClause, error).
// If INTO is present, outputIntoClause is set; otherwise outputClause is set.
func (p *Parser) parseOutputClause() (*ast.OutputClause, *ast.OutputIntoClause, error) {
	// Consume OUTPUT
	p.nextToken()

	// Parse select columns
	var selectColumns []ast.SelectElement
	for {
		elem, err := p.parseSelectElement()
		if err != nil {
			return nil, nil, err
		}
		selectColumns = append(selectColumns, elem)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Check for INTO
	if p.curTok.Type == TokenInto {
		p.nextToken() // consume INTO

		// Parse target table (variable or table name)
		var intoTable ast.TableReference
		if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
			name := p.curTok.Literal
			p.nextToken()
			intoTable = &ast.VariableTableReference{
				Variable: &ast.VariableReference{Name: name},
				ForPath:  false,
			}
		} else {
			son, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, nil, err
			}
			intoTable = &ast.NamedTableReference{
				SchemaObject: son,
				ForPath:      false,
			}
		}

		// Parse optional column list
		var intoColumns []*ast.ColumnReferenceExpression
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
				intoColumns = append(intoColumns, colRef)

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

		return nil, &ast.OutputIntoClause{
			SelectColumns:    selectColumns,
			IntoTable:        intoTable,
			IntoTableColumns: intoColumns,
		}, nil
	}

	return &ast.OutputClause{
		SelectColumns: selectColumns,
	}, nil, nil
}

