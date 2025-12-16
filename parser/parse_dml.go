// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"strings"

	"github.com/kyleconroy/teesql/ast"
)

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
		"CODEPAGE":            "Codepage",
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

	if p.curTok.Type == TokenString {
		// String literal
		value := p.curTok.Literal
		// Remove quotes
		if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}
		result.Value = value
		result.ValueExpression = &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    false,
			IsLargeObject: false,
			Value:         value,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenNumber {
		// Integer literal
		result.Value = p.curTok.Literal
		result.ValueExpression = &ast.IntegerLiteral{
			LiteralType: "Integer",
			Value:       p.curTok.Literal,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent {
		// Identifier
		result.Value = p.curTok.Literal
		result.Identifier = &ast.Identifier{
			Value:     p.curTok.Literal,
			QuoteType: "NotQuoted",
		}
		p.nextToken()
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

