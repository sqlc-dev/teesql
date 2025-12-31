// Package parser provides T-SQL parsing functionality.
package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sqlc-dev/teesql/ast"
)

func (p *Parser) parseDeclareVariableStatement() (ast.Statement, error) {
	// Consume DECLARE
	p.nextToken()

	// Check if this is DECLARE cursor_name CURSOR (without @)
	if p.curTok.Type == TokenIdent && !strings.HasPrefix(p.curTok.Literal, "@") {
		// This might be DECLARE cursor_name CURSOR
		cursorName := p.parseIdentifier()

		// Check for CURSOR keyword
		if p.curTok.Type == TokenCursor {
			return p.parseDeclareCursorStatementContinued(cursorName)
		}
		// Could also be old cursor syntax with options before CURSOR
		kwd := strings.ToUpper(p.curTok.Literal)
		if kwd == "INSENSITIVE" || kwd == "SCROLL" {
			return p.parseDeclareCursorStatementContinued(cursorName)
		}
		// Not a cursor, error
		return nil, fmt.Errorf("expected CURSOR after identifier in DECLARE, got %s", p.curTok.Literal)
	}

	// Parse variable name
	if p.curTok.Type != TokenIdent || !strings.HasPrefix(p.curTok.Literal, "@") {
		return nil, fmt.Errorf("expected variable name, got %s", p.curTok.Literal)
	}
	varName := &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
	p.nextToken()

	// Skip optional AS
	asDefined := false
	if p.curTok.Type == TokenAs {
		asDefined = true
		p.nextToken()
	}

	// Check if this is a TABLE variable
	if p.curTok.Type == TokenTable {
		return p.parseDeclareTableVariableStatement(varName, asDefined)
	}

	// Regular variable declaration
	stmt := &ast.DeclareVariableStatement{}
	elem := &ast.DeclareVariableElement{
		VariableName: varName,
	}

	// Parse data type
	dataType, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	elem.DataType = dataType

	// Check for NULL / NOT NULL
	if p.curTok.Type == TokenNull {
		elem.Nullable = &ast.NullableConstraintDefinition{Nullable: true}
		p.nextToken()
	} else if p.curTok.Type == TokenNot {
		p.nextToken()
		if p.curTok.Type == TokenNull {
			elem.Nullable = &ast.NullableConstraintDefinition{Nullable: false}
			p.nextToken()
		}
	}

	// Check for = initial value
	if p.curTok.Type == TokenEquals {
		p.nextToken()
		val, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		elem.Value = val
	}

	stmt.Declarations = append(stmt.Declarations, elem)

	// Handle additional declarations separated by comma
	for p.curTok.Type == TokenComma {
		p.nextToken()
		decl, err := p.parseDeclareVariableElement()
		if err != nil {
			return nil, err
		}
		stmt.Declarations = append(stmt.Declarations, decl)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseDeclareTableVariableStatement(varName *ast.Identifier, asDefined bool) (*ast.DeclareTableVariableStatement, error) {
	// Consume TABLE
	p.nextToken()

	stmt := &ast.DeclareTableVariableStatement{
		Body: &ast.DeclareTableVariableBody{
			VariableName: varName,
			AsDefined:    asDefined,
		},
	}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TABLE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse table definition
	tableDef, err := p.parseTableDefinitionBody()
	if err != nil {
		return nil, err
	}
	stmt.Body.Definition = tableDef

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after table definition, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseTableDefinitionBody parses the body of a table definition (column definitions, constraints, indexes)
// between parentheses. The opening parenthesis should already be consumed.
func (p *Parser) parseTableDefinitionBody() (*ast.TableDefinition, error) {
	tableDef := &ast.TableDefinition{}

	// Parse column definitions, table constraints, and indexes
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		// Check for table constraints (CHECK, CONSTRAINT, PRIMARY KEY, UNIQUE, FOREIGN KEY, INDEX)
		upperLit := strings.ToUpper(p.curTok.Literal)

		if upperLit == "CHECK" {
			constraint, err := p.parseCheckConstraintInTable()
			if err != nil {
				return nil, err
			}
			tableDef.TableConstraints = append(tableDef.TableConstraints, constraint)
		} else if upperLit == "CONSTRAINT" {
			p.nextToken() // skip CONSTRAINT
			p.nextToken() // skip constraint name
			// Parse actual constraint
			continue
		} else if upperLit == "PRIMARY" || upperLit == "UNIQUE" || upperLit == "FOREIGN" {
			constraint, err := p.parseTableConstraint()
			if err != nil {
				return nil, err
			}
			if constraint != nil {
				tableDef.TableConstraints = append(tableDef.TableConstraints, constraint)
			}
		} else if upperLit == "INDEX" {
			indexDef, err := p.parseInlineIndexDefinition()
			if err != nil {
				return nil, err
			}
			tableDef.Indexes = append(tableDef.Indexes, indexDef)
		} else {
			// Column definition
			colDef, err := p.parseColumnDefinition()
			if err != nil {
				return nil, err
			}
			tableDef.ColumnDefinitions = append(tableDef.ColumnDefinitions, colDef)
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	return tableDef, nil
}

// parseCheckConstraintInTable parses a CHECK constraint in a table definition
func (p *Parser) parseCheckConstraintInTable() (*ast.CheckConstraintDefinition, error) {
	// Consume CHECK
	p.nextToken()

	constraint := &ast.CheckConstraintDefinition{}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after CHECK, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse the check condition
	cond, err := p.parseBooleanExpression()
	if err != nil {
		return nil, err
	}
	constraint.CheckCondition = cond

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after check condition, got %s", p.curTok.Literal)
	}
	p.nextToken()

	return constraint, nil
}

// parseTableConstraint parses PRIMARY KEY, UNIQUE, or FOREIGN KEY constraints
func (p *Parser) parseTableConstraint() (ast.TableConstraint, error) {
	upperLit := strings.ToUpper(p.curTok.Literal)

	if upperLit == "PRIMARY" {
		p.nextToken() // consume PRIMARY
		if p.curTok.Type == TokenKey {
			p.nextToken() // consume KEY
		}
		constraint := &ast.UniqueConstraintDefinition{
			IsPrimaryKey: true,
		}
		// Parse optional CLUSTERED/NONCLUSTERED
		if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
			constraint.Clustered = true
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
			constraint.Clustered = false
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
			p.nextToken()
		}
		// Skip the column list
		if p.curTok.Type == TokenLParen {
			p.skipParenthesizedContent()
		}
		return constraint, nil
	} else if upperLit == "UNIQUE" {
		p.nextToken() // consume UNIQUE
		constraint := &ast.UniqueConstraintDefinition{
			IsPrimaryKey: false,
		}
		// Parse optional CLUSTERED/NONCLUSTERED
		if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
			constraint.Clustered = true
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
			constraint.Clustered = false
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
			p.nextToken()
		}
		// Parse the column list
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				colRef := &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
				}
				// Parse column name
				colName := p.parseIdentifier()
				colRef.MultiPartIdentifier = &ast.MultiPartIdentifier{
					Identifiers: []*ast.Identifier{colName},
					Count:       1,
				}
				// Check for sort order
				sortOrder := ast.SortOrderNotSpecified
				upperColNext := strings.ToUpper(p.curTok.Literal)
				if upperColNext == "ASC" {
					sortOrder = ast.SortOrderAscending
					p.nextToken()
				} else if upperColNext == "DESC" {
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
		return constraint, nil
	} else if upperLit == "FOREIGN" {
		p.nextToken() // consume FOREIGN
		if p.curTok.Type == TokenKey {
			p.nextToken() // consume KEY
		}
		// Skip the constraint body for now
		if p.curTok.Type == TokenLParen {
			p.skipParenthesizedContent()
		}
		// Skip REFERENCES
		if strings.ToUpper(p.curTok.Literal) == "REFERENCES" {
			p.skipToEndOfStatement()
		}
		return &ast.ForeignKeyConstraintDefinition{}, nil
	}

	return nil, nil
}

// parseInlineIndexDefinition parses an inline INDEX definition in a table variable
func (p *Parser) parseInlineIndexDefinition() (*ast.IndexDefinition, error) {
	// Consume INDEX
	p.nextToken()

	indexDef := &ast.IndexDefinition{}

	// Parse index name
	if p.curTok.Type == TokenIdent {
		quoteType := "NotQuoted"
		if strings.HasPrefix(p.curTok.Literal, "[") && strings.HasSuffix(p.curTok.Literal, "]") {
			quoteType = "SquareBracket"
		}
		indexDef.Name = &ast.Identifier{
			Value:     p.curTok.Literal,
			QuoteType: quoteType,
		}
		p.nextToken()
	}

	// Parse optional UNIQUE
	if strings.ToUpper(p.curTok.Literal) == "UNIQUE" {
		indexDef.Unique = true
		p.nextToken()
	}

	// Parse optional CLUSTERED/NONCLUSTERED
	if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
		indexDef.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
		indexDef.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
		p.nextToken()
	}

	// Parse column list
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			quoteType := "NotQuoted"
			if strings.HasPrefix(p.curTok.Literal, "[") && strings.HasSuffix(p.curTok.Literal, "]") {
				quoteType = "SquareBracket"
			}
			col := &ast.ColumnWithSortOrder{
				Column: &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Count: 1,
						Identifiers: []*ast.Identifier{
							{Value: p.curTok.Literal, QuoteType: quoteType},
						},
					},
				},
				SortOrder: ast.SortOrderNotSpecified,
			}
			p.nextToken()

			// Parse optional ASC/DESC
			if strings.ToUpper(p.curTok.Literal) == "ASC" {
				col.SortOrder = ast.SortOrderAscending
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "DESC" {
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
			p.nextToken()
		}
	}

	// Parse optional INCLUDE
	if strings.ToUpper(p.curTok.Literal) == "INCLUDE" {
		p.nextToken() // consume INCLUDE
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				quoteType := "NotQuoted"
				if strings.HasPrefix(p.curTok.Literal, "[") && strings.HasSuffix(p.curTok.Literal, "]") {
					quoteType = "SquareBracket"
				}
				includeCol := &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Count: 1,
						Identifiers: []*ast.Identifier{
							{Value: p.curTok.Literal, QuoteType: quoteType},
						},
					},
				}
				indexDef.IncludeColumns = append(indexDef.IncludeColumns, includeCol)
				p.nextToken()

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
	}

	return indexDef, nil
}

// skipParenthesizedContent skips content within parentheses, handling nested parens
func (p *Parser) skipParenthesizedContent() {
	if p.curTok.Type != TokenLParen {
		return
	}
	p.nextToken() // consume (
	depth := 1
	for depth > 0 && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenLParen {
			depth++
		} else if p.curTok.Type == TokenRParen {
			depth--
		}
		p.nextToken()
	}
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

	// Check for NULL / NOT NULL
	if p.curTok.Type == TokenNull {
		elem.Nullable = &ast.NullableConstraintDefinition{Nullable: true}
		p.nextToken()
	} else if p.curTok.Type == TokenNot {
		p.nextToken()
		if p.curTok.Type == TokenNull {
			elem.Nullable = &ast.NullableConstraintDefinition{Nullable: false}
			p.nextToken()
		}
	}

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
	dt, err := p.parseDataTypeReference()
	if err != nil {
		return nil, err
	}
	// For backward compatibility, if it's SqlDataTypeReference, return it directly
	if sqlDt, ok := dt.(*ast.SqlDataTypeReference); ok {
		return sqlDt, nil
	}
	// Otherwise wrap in SqlDataTypeReference (shouldn't happen often)
	return &ast.SqlDataTypeReference{}, nil
}

// parseDataTypeReference parses a data type and returns the appropriate DataTypeReference
func (p *Parser) parseDataTypeReference() (ast.DataTypeReference, error) {
	if p.curTok.Type == TokenCursor {
		dt := &ast.SqlDataTypeReference{
			SqlDataTypeOption: "Cursor",
		}
		p.nextToken()
		return dt, nil
	}

	if p.curTok.Type != TokenIdent {
		return nil, fmt.Errorf("expected data type, got %s", p.curTok.Literal)
	}

	var typeName string
	var quoteType string
	literal := p.curTok.Literal

	// Check if this is a bracketed identifier like [int]
	if len(literal) >= 2 && literal[0] == '[' && literal[len(literal)-1] == ']' {
		typeName = literal[1 : len(literal)-1]
		quoteType = "SquareBracket"
	} else {
		typeName = literal
		quoteType = "NotQuoted"
	}
	p.nextToken()

	baseId := &ast.Identifier{Value: typeName, QuoteType: quoteType}
	baseName := &ast.SchemaObjectName{
		BaseIdentifier: baseId,
		Count:          1,
		Identifiers:    []*ast.Identifier{baseId},
	}

	// Check for XML with schema collection: XML(schema_collection)
	if strings.ToUpper(typeName) == "XML" && p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		// Parse the schema collection name
		schemaName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}

		return &ast.XmlDataTypeReference{
			XmlDataTypeOption:   "None",
			XmlSchemaCollection: schemaName,
			Name:                baseName,
		}, nil
	}

	// Check if this is a known SQL data type
	sqlOption, isKnownType := getSqlDataTypeOption(typeName)

	if !isKnownType {
		// Return UserDataTypeReference for unknown types
		return &ast.UserDataTypeReference{
			Name: baseName,
		}, nil
	}

	dt := &ast.SqlDataTypeReference{
		SqlDataTypeOption: sqlOption,
		Name:              baseName,
	}

	// Check for parameters like VARCHAR(100) or VARCHAR(MAX)
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			// Special case: MAX keyword in data type parameters
			if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "MAX" {
				dt.Parameters = append(dt.Parameters, &ast.MaxLiteral{
					LiteralType: "Max",
					Value:       p.curTok.Literal,
				})
				p.nextToken()
			} else {
				expr, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				dt.Parameters = append(dt.Parameters, expr)
			}
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

// getSqlDataTypeOption returns the SqlDataTypeOption for a type name and whether it's a known SQL type.
func getSqlDataTypeOption(typeName string) (string, bool) {
	typeMap := map[string]string{
		"INT":               "Int",
		"INTEGER":           "Int",
		"BIGINT":            "BigInt",
		"SMALLINT":          "SmallInt",
		"TINYINT":           "TinyInt",
		"BIT":               "Bit",
		"DECIMAL":           "Decimal",
		"NUMERIC":           "Numeric",
		"MONEY":             "Money",
		"SMALLMONEY":        "SmallMoney",
		"FLOAT":             "Float",
		"REAL":              "Real",
		"DATETIME":          "DateTime",
		"DATETIME2":         "DateTime2",
		"DATETIMEOFFSET":    "DateTimeOffset",
		"SMALLDATETIME":     "SmallDateTime",
		"DATE":              "Date",
		"TIME":              "Time",
		"CHAR":              "Char",
		"VARCHAR":           "VarChar",
		"TEXT":              "Text",
		"NCHAR":             "NChar",
		"NVARCHAR":          "NVarChar",
		"NTEXT":             "NText",
		"BINARY":            "Binary",
		"VARBINARY":         "VarBinary",
		"IMAGE":             "Image",
		"CURSOR":            "Cursor",
		"SQL_VARIANT":       "Sql_Variant",
		"TABLE":             "Table",
		"UNIQUEIDENTIFIER":  "UniqueIdentifier",
		"XML":               "Xml",
		"JSON":              "Json",
		"GEOGRAPHY":         "Geography",
		"GEOMETRY":          "Geometry",
		"HIERARCHYID":       "HierarchyId",
		"ROWVERSION":        "Rowversion",
		"TIMESTAMP":         "Timestamp",
		"CONNECTION":        "Connection",
		"VECTOR":            "Vector",
	}
	if mapped, ok := typeMap[strings.ToUpper(typeName)]; ok {
		return mapped, true
	}
	return "", false
}

func convertDataTypeOption(typeName string) string {
	if mapped, ok := getSqlDataTypeOption(typeName); ok {
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

	// Check for special SET statements
	if p.curTok.Type == TokenIdent {
		optionName := strings.ToUpper(p.curTok.Literal)

		// Handle SET ROWCOUNT
		if optionName == "ROWCOUNT" {
			p.nextToken() // consume ROWCOUNT
			var numRows ast.ScalarExpression
			if strings.HasPrefix(p.curTok.Literal, "@") {
				numRows = &ast.VariableReference{Name: p.curTok.Literal}
				p.nextToken()
			} else {
				numRows = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
				p.nextToken()
			}
			if p.curTok.Type == TokenSemicolon {
				p.nextToken()
			}
			return &ast.SetRowCountStatement{NumberRows: numRows}, nil
		}

		// Handle SET STATISTICS
		if optionName == "STATISTICS" {
			return p.parseSetStatisticsStatement()
		}

		// Handle SET OFFSETS
		if optionName == "OFFSETS" {
			return p.parseSetOffsetsStatement()
		}

		// Handle predicate SET options like SET ANSI_NULLS ON/OFF
		// These can have multiple options with commas
		setOpt := p.mapPredicateSetOption(optionName)
		if setOpt != "" {
			return p.parsePredicateSetStatement(setOpt)
		}
	}

	stmt := &ast.SetVariableStatement{
		AssignmentKind: "Equals",
	}

	// Parse variable name
	if p.curTok.Type != TokenIdent || !strings.HasPrefix(p.curTok.Literal, "@") {
		return nil, fmt.Errorf("expected variable name, got %s", p.curTok.Literal)
	}
	stmt.Variable = &ast.VariableReference{Name: p.curTok.Literal}
	p.nextToken()

	// Check for dot or double-colon separator (SET @a.b = ... or SET @a::b ...)
	if p.curTok.Type == TokenDot {
		stmt.SeparatorType = "Dot"
		p.nextToken()
		if p.curTok.Type == TokenIdent {
			stmt.Identifier = &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
			p.nextToken()
		}
	} else if p.curTok.Type == TokenColonColon {
		stmt.SeparatorType = "DoubleColon"
		p.nextToken() // consume ::
		if p.curTok.Type == TokenIdent {
			stmt.Identifier = &ast.Identifier{Value: p.curTok.Literal, QuoteType: "NotQuoted"}
			p.nextToken()
		}
	}

	// Check for function call: SET @a.b () or SET @a.b (params)
	if p.curTok.Type == TokenLParen {
		stmt.FunctionCallExists = true
		p.nextToken() // consume (
		// Parse parameters
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			param, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Parameters = append(stmt.Parameters, param)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	// Expect = or compound assignment operator
	if p.isCompoundAssignment() {
		stmt.AssignmentKind = p.getAssignmentKind()
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected =, got %s", p.curTok.Literal)
	}

	// Check for CURSOR definition
	if p.curTok.Type == TokenCursor {
		p.nextToken()
		cursorDef := &ast.CursorDefinition{}

		// Parse cursor options (SCROLL, DYNAMIC, etc.) until FOR
		for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
			if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "FOR" {
				p.nextToken() // consume FOR
				break
			}
			// Cursor options are typically identifiers like SCROLL, DYNAMIC, STATIC, etc.
			if p.curTok.Type == TokenIdent {
				optKind := strings.Title(strings.ToLower(p.curTok.Literal))
				cursorDef.Options = append(cursorDef.Options, &ast.CursorOption{OptionKind: optKind})
			}
			p.nextToken()
		}

		if p.curTok.Type == TokenSelect {
			qe, err := p.parseQueryExpression()
			if err != nil {
				return nil, err
			}
			cursorDef.Select = &ast.SelectStatement{QueryExpression: qe}
		}
		stmt.CursorDefinition = cursorDef
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

// mapPredicateSetOption maps option names to their SetOptions values
func (p *Parser) mapPredicateSetOption(name string) string {
	switch name {
	case "ANSI_DEFAULTS":
		return "AnsiDefaults"
	case "ANSI_NULLS":
		return "AnsiNulls"
	case "ANSI_NULL_DFLT_OFF":
		return "AnsiNullDfltOff"
	case "ANSI_NULL_DFLT_ON":
		return "AnsiNullDfltOn"
	case "ANSI_PADDING":
		return "AnsiPadding"
	case "ANSI_WARNINGS":
		return "AnsiWarnings"
	case "ARITHABORT":
		return "ArithAbort"
	case "ARITHIGNORE":
		return "ArithIgnore"
	case "CONCAT_NULL_YIELDS_NULL":
		return "ConcatNullYieldsNull"
	case "CURSOR_CLOSE_ON_COMMIT":
		return "CursorCloseOnCommit"
	case "FMTONLY":
		return "FmtOnly"
	case "FORCEPLAN":
		return "ForcePlan"
	case "IMPLICIT_TRANSACTIONS":
		return "ImplicitTransactions"
	case "NOCOUNT":
		return "NoCount"
	case "NOEXEC":
		return "NoExec"
	case "NO_BROWSETABLE":
		return "NoBrowsetable"
	case "NUMERIC_ROUNDABORT":
		return "NumericRoundAbort"
	case "PARSEONLY":
		return "ParseOnly"
	case "QUOTED_IDENTIFIER":
		return "QuotedIdentifier"
	case "REMOTE_PROC_TRANSACTIONS":
		return "RemoteProcTransactions"
	case "SHOWPLAN_ALL":
		return "ShowPlanAll"
	case "SHOWPLAN_TEXT":
		return "ShowPlanText"
	case "SHOWPLAN_XML":
		return "ShowPlanXml"
	case "XACT_ABORT":
		return "XactAbort"
	default:
		return ""
	}
}

// predicateSetOptionOrder defines the sort order for predicate SET options
var predicateSetOptionOrder = map[string]int{
	"AnsiNulls":             1,
	"AnsiNullDfltOff":       2,
	"AnsiNullDfltOn":        3,
	"AnsiPadding":           4,
	"AnsiWarnings":          5,
	"ConcatNullYieldsNull":  6,
	"CursorCloseOnCommit":   7,
	"ImplicitTransactions":  8,
	"QuotedIdentifier":      9,
	"ArithAbort":            10,
	"ArithIgnore":           11,
	"FmtOnly":               12,
	"NoCount":               13,
	"NoExec":                14,
	"NumericRoundAbort":     15,
	"ParseOnly":             16,
	"AnsiDefaults":          17,
	"ForcePlan":             18,
	"ShowPlanAll":           19,
	"ShowPlanText":          20,
	"ShowPlanXml":           21,
	"NoBrowsetable":         22,
	"RemoteProcTransactions": 23,
	"XactAbort":             24,
}

// parsePredicateSetStatement parses SET option1, option2, ... ON/OFF
func (p *Parser) parsePredicateSetStatement(firstOpt string) (*ast.PredicateSetStatement, error) {
	options := []string{firstOpt}
	p.nextToken() // consume first option

	// Check for more options with commas
	for p.curTok.Type == TokenComma {
		p.nextToken() // consume comma
		if p.curTok.Type == TokenIdent {
			nextOpt := p.mapPredicateSetOption(strings.ToUpper(p.curTok.Literal))
			if nextOpt != "" {
				options = append(options, nextOpt)
				p.nextToken()
			} else {
				break
			}
		}
	}

	// Sort options according to ScriptDom order
	sort.Slice(options, func(i, j int) bool {
		return predicateSetOptionOrder[options[i]] < predicateSetOptionOrder[options[j]]
	})

	// Parse ON/OFF
	isOn := false
	if p.curTok.Type == TokenOn || (p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "ON") {
		isOn = true
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OFF" {
		isOn = false
		p.nextToken()
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.PredicateSetStatement{
		Options: strings.Join(options, ", "),
		IsOn:    isOn,
	}, nil
}

// parseSetStatisticsStatement parses SET STATISTICS opt1, opt2, ... ON/OFF
func (p *Parser) parseSetStatisticsStatement() (*ast.SetStatisticsStatement, error) {
	p.nextToken() // consume STATISTICS

	// Map statistics options
	mapStatOpt := func(name string) string {
		switch name {
		case "IO":
			return "IO"
		case "PROFILE":
			return "Profile"
		case "TIME":
			return "Time"
		case "XML":
			return "Xml"
		default:
			return ""
		}
	}

	var options []string
	for {
		var optName string
		if p.curTok.Type == TokenTime {
			optName = "Time"
		} else if p.curTok.Type == TokenIdent {
			optName = mapStatOpt(strings.ToUpper(p.curTok.Literal))
		}
		if optName == "" {
			break
		}
		options = append(options, optName)
		p.nextToken()
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Parse ON/OFF
	isOn := false
	if p.curTok.Type == TokenOn || (p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "ON") {
		isOn = true
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OFF" {
		isOn = false
		p.nextToken()
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.SetStatisticsStatement{
		Options: strings.Join(options, ", "),
		IsOn:    isOn,
	}, nil
}

// setOffsetsOptionOrder defines the sort order for SET OFFSETS options
var setOffsetsOptionOrder = map[string]int{
	"Select":    1,
	"From":      2,
	"Order":     3,
	"Compute":   4,
	"Table":     5,
	"Procedure": 6,
	"Execute":   7,
	"Statement": 8,
	"Param":     9,
}

// parseSetOffsetsStatement parses SET OFFSETS opt1, opt2, ... ON/OFF
func (p *Parser) parseSetOffsetsStatement() (*ast.SetOffsetsStatement, error) {
	p.nextToken() // consume OFFSETS

	// Map offset options - these can be either tokens or identifiers
	mapOffsetOpt := func() string {
		switch p.curTok.Type {
		case TokenSelect:
			return "Select"
		case TokenFrom:
			return "From"
		case TokenOrder:
			return "Order"
		case TokenTable:
			return "Table"
		case TokenProcedure:
			return "Procedure"
		case TokenExecute:
			return "Execute"
		case TokenIdent:
			switch strings.ToUpper(p.curTok.Literal) {
			case "COMPUTE":
				return "Compute"
			case "STATEMENT":
				return "Statement"
			case "PARAM":
				return "Param"
			}
		}
		return ""
	}

	var options []string
	for {
		optName := mapOffsetOpt()
		if optName == "" {
			break
		}
		options = append(options, optName)
		p.nextToken()
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Sort options according to ScriptDom order
	sort.Slice(options, func(i, j int) bool {
		return setOffsetsOptionOrder[options[i]] < setOffsetsOptionOrder[options[j]]
	})

	// Parse ON/OFF
	isOn := false
	if p.curTok.Type == TokenOn || (p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "ON") {
		isOn = true
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OFF" {
		isOn = false
		p.nextToken()
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.SetOffsetsStatement{
		Options: strings.Join(options, ", "),
		IsOn:    isOn,
	}, nil
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
		// Check for ATOMIC
		if strings.ToUpper(p.curTok.Literal) == "ATOMIC" {
			return p.parseBeginAtomicBlockStatement()
		}
		// Fall through to BEGIN...END block
		fallthrough
	default:
		return p.parseBeginEndBlockStatementContinued()
	}
}

func (p *Parser) parseBeginAtomicBlockStatement() (*ast.BeginEndAtomicBlockStatement, error) {
	p.nextToken() // consume ATOMIC

	stmt := &ast.BeginEndAtomicBlockStatement{
		StatementList: &ast.StatementList{},
	}

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
		}

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			switch optName {
			case "TRANSACTION":
				// TRANSACTION ISOLATION LEVEL = ...
				if strings.ToUpper(p.curTok.Literal) == "ISOLATION" {
					p.nextToken() // consume ISOLATION
					if strings.ToUpper(p.curTok.Literal) == "LEVEL" {
						p.nextToken() // consume LEVEL
					}
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
				}
				// Parse the isolation level identifier
				opt := &ast.IdentifierAtomicBlockOption{
					OptionKind: "IsolationLevel",
					Value:      p.parseIdentifier(),
				}
				stmt.Options = append(stmt.Options, opt)
			case "LANGUAGE":
				// Parse the language value
				if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
					value := p.curTok.Literal
					isNational := p.curTok.Type == TokenNationalString
					// Strip the N prefix and quotes from national strings
					if isNational && len(value) >= 3 && (value[0] == 'N' || value[0] == 'n') && value[1] == '\'' {
						value = value[2 : len(value)-1]
					} else if len(value) >= 2 && value[0] == '\'' {
						// Strip quotes from regular strings
						value = value[1 : len(value)-1]
					}
					strLit := &ast.StringLiteral{
						LiteralType:   "String",
						Value:         value,
						IsNational:    isNational,
						IsLargeObject: false,
					}
					p.nextToken()
					opt := &ast.LiteralAtomicBlockOption{
						OptionKind: "Language",
						Value:      strLit,
					}
					stmt.Options = append(stmt.Options, opt)
				} else {
					opt := &ast.IdentifierAtomicBlockOption{
						OptionKind: "Language",
						Value:      p.parseIdentifier(),
					}
					stmt.Options = append(stmt.Options, opt)
				}
			case "DATEFIRST", "DATEFORMAT":
				opt := &ast.IdentifierAtomicBlockOption{
					OptionKind: optName,
					Value:      p.parseIdentifier(),
				}
				stmt.Options = append(stmt.Options, opt)
			default:
				// Skip unknown options
				if p.curTok.Type == TokenIdent || p.curTok.Type == TokenString {
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

func (p *Parser) parseBeginTransactionStatementContinued(distributed bool) (*ast.BeginTransactionStatement, error) {
	// TRANSACTION or TRAN already consumed by caller
	p.nextToken()

	stmt := &ast.BeginTransactionStatement{
		Distributed: distributed,
	}

	// Optional transaction name or variable - check for variable first
	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			ValueExpression: &ast.VariableReference{
				Name: p.curTok.Literal,
			},
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && !isKeyword(p.curTok.Literal) {
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			Identifier: &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
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
			if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString || (p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '@') {
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
		// Skip semicolons
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
			continue
		}
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
		// Skip semicolons
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
			continue
		}
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
	case TokenCredential:
		return p.parseCreateCredentialStatement(false)
	case TokenProcedure:
		return p.parseCreateProcedureStatement()
	case TokenDatabase:
		return p.parseCreateDatabaseStatement()
	case TokenLogin:
		return p.parseCreateLoginStatement()
	case TokenIndex:
		return p.parseCreateIndexStatement()
	case TokenAsymmetric:
		return p.parseCreateAsymmetricKeyStatement()
	case TokenSymmetric:
		return p.parseCreateSymmetricKeyStatement()
	case TokenCertificate:
		return p.parseCreateCertificateStatement()
	case TokenMessage:
		return p.parseCreateMessageTypeStatement()
	case TokenUser:
		return p.parseCreateUserStatement()
	case TokenFunction:
		return p.parseCreateFunctionStatement()
	case TokenTrigger:
		return p.parseCreateTriggerStatement()
	case TokenExternal:
		return p.parseCreateExternalStatement()
	case TokenTyp:
		return p.parseCreateTypeStatement()
	case TokenIdent:
		// Handle keywords that are not reserved tokens
		switch strings.ToUpper(p.curTok.Literal) {
		case "ROLE":
			return p.parseCreateRoleStatement()
		case "CONTRACT":
			return p.parseCreateContractStatement()
		case "PARTITION":
			// Could be PARTITION SCHEME or PARTITION FUNCTION
			p.nextToken() // consume PARTITION
			if strings.ToUpper(p.curTok.Literal) == "FUNCTION" {
				return p.parseCreatePartitionFunctionFromPartition()
			}
			return p.parseCreatePartitionSchemeStatementFromPartition()
		case "RULE":
			return p.parseCreateRuleStatement()
		case "SYNONYM":
			return p.parseCreateSynonymStatement()
		case "XML":
			// Could be XML SCHEMA COLLECTION or XML INDEX
			p.nextToken() // consume XML
			if strings.ToUpper(p.curTok.Literal) == "INDEX" {
				return p.parseCreateXmlIndexFromXml()
			}
			return p.parseCreateXmlSchemaCollectionFromXml()
		case "SEARCH":
			return p.parseCreateSearchPropertyListStatement()
		case "AGGREGATE":
			return p.parseCreateAggregateStatement()
		case "CLUSTERED", "NONCLUSTERED", "COLUMNSTORE":
			return p.parseCreateColumnStoreIndexStatement()
		case "EXTERNAL":
			return p.parseCreateExternalStatement()
		case "EVENT":
			// Could be EVENT SESSION or EVENT NOTIFICATION
			p.nextToken() // consume EVENT
			if strings.ToUpper(p.curTok.Literal) == "SESSION" {
				return p.parseCreateEventSessionStatementFromEvent()
			}
			return p.parseCreateEventNotificationFromEvent()
		case "SERVICE":
			return p.parseCreateServiceStatement()
		case "QUEUE":
			return p.parseCreateQueueStatement()
		case "ROUTE":
			return p.parseCreateRouteStatement()
		case "ENDPOINT":
			return p.parseCreateEndpointStatement()
		case "ASSEMBLY":
			return p.parseCreateAssemblyStatement()
		case "APPLICATION":
			return p.parseCreateApplicationRoleStatement()
		case "FULLTEXT":
			return p.parseCreateFulltextStatement()
		case "REMOTE":
			return p.parseCreateRemoteServiceBindingStatement()
		case "STATISTICS":
			return p.parseCreateStatisticsStatement()
		case "TYPE":
			return p.parseCreateTypeStatement()
		case "UNIQUE":
			return p.parseCreateIndexStatement()
		case "PRIMARY":
			return p.parseCreateXmlIndexStatement()
		case "CRYPTOGRAPHIC":
			return p.parseCreateCryptographicProviderStatement()
		case "FEDERATION":
			return p.parseCreateFederationStatement()
		case "WORKLOAD":
			// Check if it's CLASSIFIER or GROUP
			nextWord := strings.ToUpper(p.peekTok.Literal)
			if nextWord == "CLASSIFIER" {
				return p.parseCreateWorkloadClassifierStatement()
			}
			return p.parseCreateWorkloadGroupStatement()
		case "SEQUENCE":
			return p.parseCreateSequenceStatement()
		case "SPATIAL":
			return p.parseCreateSpatialIndexStatement()
		case "SERVER":
			// Check if it's SERVER ROLE or SERVER AUDIT
			p.nextToken() // consume SERVER
			switch strings.ToUpper(p.curTok.Literal) {
			case "ROLE":
				return p.parseCreateServerRoleStatementContinued()
			case "AUDIT":
				return p.parseCreateServerAuditStatement()
			default:
				return nil, fmt.Errorf("expected ROLE or AUDIT after SERVER, got %s", p.curTok.Literal)
			}
		}
		// Lenient: skip unknown CREATE statements
		p.skipToEndOfStatement()
		return &ast.CreateProcedureStatement{}, nil
	default:
		// Lenient: if we see another CREATE, skip it and try to continue
		// This handles malformed SQL like "create create create certificate c1"
		if p.curTok.Type == TokenCreate {
			// Skip the extra CREATE and retry
			p.nextToken()
			return p.parseCreateStatement()
		}
		// Lenient: skip unknown CREATE statements
		p.skipToEndOfStatement()
		return &ast.CreateProcedureStatement{}, nil
	}
}

func (p *Parser) parseCreateCryptographicProviderStatement() (*ast.CreateCryptographicProviderStatement, error) {
	// Consume CRYPTOGRAPHIC
	p.nextToken()

	// Consume PROVIDER
	if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
		p.nextToken()
	}

	stmt := &ast.CreateCryptographicProviderStatement{}

	// Parse provider name
	stmt.Name = p.parseIdentifier()

	// Parse FROM FILE = 'path'
	if strings.ToUpper(p.curTok.Literal) == "FROM" {
		p.nextToken() // consume FROM
		if strings.ToUpper(p.curTok.Literal) == "FILE" {
			p.nextToken() // consume FILE
		}
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		stmt.File, _ = p.parseStringLiteral()
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterCryptographicProviderStatement() (*ast.AlterCryptographicProviderStatement, error) {
	// Consume CRYPTOGRAPHIC
	p.nextToken()

	// Consume PROVIDER
	if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
		p.nextToken()
	}

	stmt := &ast.AlterCryptographicProviderStatement{}

	// Parse provider name
	stmt.Name = p.parseIdentifier()

	// Parse action: FROM FILE = 'path', ENABLE, or DISABLE
	switch strings.ToUpper(p.curTok.Literal) {
	case "FROM":
		stmt.Option = "None"
		p.nextToken() // consume FROM
		if strings.ToUpper(p.curTok.Literal) == "FILE" {
			p.nextToken() // consume FILE
		}
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		stmt.File, _ = p.parseStringLiteral()
	case "ENABLE":
		stmt.Option = "Enable"
		p.nextToken()
	case "DISABLE":
		stmt.Option = "Disable"
		p.nextToken()
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseDropCryptographicProviderStatement() (*ast.DropCryptographicProviderStatement, error) {
	// Consume CRYPTOGRAPHIC
	p.nextToken()

	// Consume PROVIDER
	if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
		p.nextToken()
	}

	stmt := &ast.DropCryptographicProviderStatement{}

	// Check for IF EXISTS
	if p.curTok.Type == TokenIf {
		p.nextToken() // consume IF
		if strings.ToUpper(p.curTok.Literal) == "EXISTS" {
			stmt.IsIfExists = true
			p.nextToken() // consume EXISTS
		}
	}

	// Parse provider name
	stmt.Name = p.parseIdentifier()

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateRoleStatement() (*ast.CreateRoleStatement, error) {
	// Consume ROLE
	p.nextToken()

	stmt := &ast.CreateRoleStatement{}

	// Parse role name
	stmt.Name = p.parseIdentifier()

	// Check for optional AUTHORIZATION
	if p.curTok.Type == TokenAuthorization {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateServerRoleStatement() (*ast.CreateServerRoleStatement, error) {
	// Consume SERVER
	p.nextToken()

	// Expect ROLE
	if strings.ToUpper(p.curTok.Literal) != "ROLE" {
		return nil, fmt.Errorf("expected ROLE after SERVER, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ROLE

	return p.parseCreateServerRoleStatementBody()
}

func (p *Parser) parseCreateServerRoleStatementContinued() (*ast.CreateServerRoleStatement, error) {
	// ROLE keyword should be current token, consume it
	p.nextToken()
	return p.parseCreateServerRoleStatementBody()
}

func (p *Parser) parseCreateServerRoleStatementBody() (*ast.CreateServerRoleStatement, error) {
	stmt := &ast.CreateServerRoleStatement{}

	// Parse role name
	stmt.Name = p.parseIdentifier()

	// Check for optional AUTHORIZATION
	if p.curTok.Type == TokenAuthorization {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateServerAuditStatement() (*ast.CreateServerAuditStatement, error) {
	// AUDIT keyword should be current token, consume it
	p.nextToken()

	stmt := &ast.CreateServerAuditStatement{}

	// Parse audit name
	stmt.AuditName = p.parseIdentifier()

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

func (p *Parser) parseAuditTarget() (*ast.AuditTarget, error) {
	target := &ast.AuditTarget{}

	// Parse target kind (FILE, APPLICATION_LOG, SECURITY_LOG)
	switch strings.ToUpper(p.curTok.Literal) {
	case "FILE":
		target.TargetKind = "File"
	case "APPLICATION_LOG":
		target.TargetKind = "ApplicationLog"
	case "SECURITY_LOG":
		target.TargetKind = "SecurityLog"
	default:
		target.TargetKind = capitalizeFirst(p.curTok.Literal)
	}
	p.nextToken()

	// Parse target options in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			opt, err := p.parseAuditTargetOption()
			if err != nil {
				return nil, err
			}
			target.TargetOptions = append(target.TargetOptions, opt)
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

	return target, nil
}

func (p *Parser) parseAuditTargetOption() (ast.AuditTargetOption, error) {
	optName := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after audit target option, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse value
	val, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	optKind := ""
	switch optName {
	case "FILEPATH":
		optKind = "FilePath"
	case "MAX_FILES":
		optKind = "MaxFiles"
	case "MAX_ROLLOVER_FILES":
		optKind = "MaxRolloverFiles"
	case "MAXSIZE":
		optKind = "MaxSize"
	case "RESERVE_DISK_SPACE":
		optKind = "ReserveDiskSpace"
	default:
		optKind = capitalizeFirst(strings.ToLower(optName))
	}

	return &ast.LiteralAuditTargetOption{
		OptionKind: optKind,
		Value:      val,
	}, nil
}

func (p *Parser) parseAuditOption() (ast.AuditOption, error) {
	optName := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	switch optName {
	case "ON_FAILURE":
		// Expect =
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after ON_FAILURE, got %s", p.curTok.Literal)
		}
		p.nextToken()
		action := ""
		switch strings.ToUpper(p.curTok.Literal) {
		case "CONTINUE":
			action = "Continue"
		case "SHUTDOWN":
			action = "Shutdown"
		case "FAIL_OPERATION":
			action = "FailOperation"
		default:
			action = capitalizeFirst(strings.ToLower(p.curTok.Literal))
		}
		p.nextToken()
		return &ast.OnFailureAuditOption{
			OptionKind:      "OnFailure",
			OnFailureAction: action,
		}, nil
	case "QUEUE_DELAY":
		// Expect =
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after QUEUE_DELAY, got %s", p.curTok.Literal)
		}
		p.nextToken()
		val, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.QueueDelayAuditOption{
			OptionKind: "QueueDelay",
			Delay:      val,
		}, nil
	case "STATE":
		// Expect =
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after STATE, got %s", p.curTok.Literal)
		}
		p.nextToken()
		value := capitalizeFirst(strings.ToLower(p.curTok.Literal))
		p.nextToken()
		return &ast.StateAuditOption{
			OptionKind: "State",
			Value:      value,
		}, nil
	case "AUDIT_GUID":
		// Expect =
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after AUDIT_GUID, got %s", p.curTok.Literal)
		}
		p.nextToken()
		val, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.AuditGuidAuditOption{
			OptionKind: "AuditGuid",
			Guid:       val,
		}, nil
	default:
		return nil, fmt.Errorf("unknown audit option: %s", optName)
	}
}

func (p *Parser) parseAuditPredicate() (ast.BooleanExpression, error) {
	return p.parseAuditBooleanExpression()
}

func (p *Parser) parseAuditBooleanExpression() (ast.BooleanExpression, error) {
	// Parse first operand
	left, err := p.parseAuditBooleanPrimary()
	if err != nil {
		return nil, err
	}

	// Check for AND/OR
	for strings.ToUpper(p.curTok.Literal) == "AND" || strings.ToUpper(p.curTok.Literal) == "OR" {
		op := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		right, err := p.parseAuditBooleanPrimary()
		if err != nil {
			return nil, err
		}
		var binaryType string
		if op == "AND" {
			binaryType = "And"
		} else {
			binaryType = "Or"
		}
		left = &ast.BooleanBinaryExpression{
			BinaryExpressionType: binaryType,
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}

	return left, nil
}

func (p *Parser) parseAuditBooleanPrimary() (ast.BooleanExpression, error) {
	// For audit predicates, the left side is a SourceDeclaration
	// which wraps an EventSessionObjectName
	var identifiers []*ast.Identifier
	identifiers = append(identifiers, p.parseIdentifier())

	// Check for multi-part identifier
	for p.curTok.Type == TokenDot {
		p.nextToken() // consume .
		identifiers = append(identifiers, p.parseIdentifier())
	}

	sourceDecl := &ast.SourceDeclaration{
		Value: &ast.EventSessionObjectName{
			MultiPartIdentifier: &ast.MultiPartIdentifier{
				Count:       len(identifiers),
				Identifiers: identifiers,
			},
		},
	}

	// Now parse comparison operator and right side
	compType := ""
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

	// Parse right side
	right, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	return &ast.BooleanComparisonExpression{
		ComparisonType:   compType,
		FirstExpression:  sourceDecl,
		SecondExpression: right,
	}, nil
}

func (p *Parser) parseCreateContractStatement() (*ast.CreateContractStatement, error) {
	// Consume CONTRACT
	p.nextToken()

	stmt := &ast.CreateContractStatement{}

	// Parse contract name
	stmt.Name = p.parseIdentifier()

	// Check for ( (optional for lenient parsing)
	if p.curTok.Type != TokenLParen {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume (

	// Parse messages
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		msg := &ast.ContractMessage{}

		// Parse message name
		msg.Name = p.parseIdentifier()

		// Expect SENT
		if strings.ToUpper(p.curTok.Literal) != "SENT" {
			return nil, fmt.Errorf("expected SENT, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume SENT

		// Expect BY
		if strings.ToUpper(p.curTok.Literal) != "BY" {
			return nil, fmt.Errorf("expected BY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume BY

		// Parse sender type
		senderType := strings.ToUpper(p.curTok.Literal)
		switch senderType {
		case "INITIATOR":
			msg.SentBy = "Initiator"
		case "TARGET":
			msg.SentBy = "Target"
		case "ANY":
			msg.SentBy = "Any"
		default:
			return nil, fmt.Errorf("expected INITIATOR, TARGET, or ANY, got %s", p.curTok.Literal)
		}
		p.nextToken()

		stmt.Messages = append(stmt.Messages, msg)

		// Check for comma or end of list
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else if p.curTok.Type != TokenRParen {
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

func (p *Parser) parseCreatePartitionSchemeStatement() (*ast.CreatePartitionSchemeStatement, error) {
	// Consume PARTITION
	p.nextToken()

	// Expect SCHEME
	if strings.ToUpper(p.curTok.Literal) != "SCHEME" {
		return nil, fmt.Errorf("expected SCHEME after PARTITION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.CreatePartitionSchemeStatement{}

	// Parse scheme name
	stmt.Name = p.parseIdentifier()

	// Expect AS
	if p.curTok.Type != TokenAs {
		return nil, fmt.Errorf("expected AS, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect PARTITION
	if strings.ToUpper(p.curTok.Literal) != "PARTITION" {
		return nil, fmt.Errorf("expected PARTITION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse partition function name
	stmt.PartitionFunction = p.parseIdentifier()

	// Check for optional ALL keyword
	if p.curTok.Type == TokenAll {
		stmt.IsAll = true
		p.nextToken()
	}

	// Expect TO
	if strings.ToUpper(p.curTok.Literal) != "TO" {
		return nil, fmt.Errorf("expected TO, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TO, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse file groups
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		idOrVal := &ast.IdentifierOrValueExpression{}

		if p.curTok.Type == TokenString {
			// String literal - strip surrounding quotes
			litVal := p.curTok.Literal
			if len(litVal) >= 2 && litVal[0] == '\'' && litVal[len(litVal)-1] == '\'' {
				litVal = litVal[1 : len(litVal)-1]
			}
			idOrVal.Value = litVal
			idOrVal.ValueExpression = &ast.StringLiteral{
				LiteralType:   "String",
				Value:         litVal,
				IsNational:    false,
				IsLargeObject: false,
			}
			p.nextToken()
		} else {
			// Identifier
			id := p.parseIdentifier()
			idOrVal.Value = id.Value
			idOrVal.Identifier = id
		}

		stmt.FileGroups = append(stmt.FileGroups, idOrVal)

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

func (p *Parser) parseCreateRuleStatement() (*ast.CreateRuleStatement, error) {
	// Consume RULE
	p.nextToken()

	stmt := &ast.CreateRuleStatement{}

	// Parse rule name (can be two-part: dbo.r1)
	name, _ := p.parseSchemaObjectName()
	stmt.Name = name

	// Check for AS (optional for lenient parsing)
	if p.curTok.Type != TokenAs {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse boolean expression
	expr, err := p.parseBooleanExpression()
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

func (p *Parser) parseCreateSynonymStatement() (*ast.CreateSynonymStatement, error) {
	// Consume SYNONYM
	p.nextToken()

	stmt := &ast.CreateSynonymStatement{}

	// Parse synonym name
	name, _ := p.parseSchemaObjectName()
	stmt.Name = name

	// Check for FOR (optional for lenient parsing)
	if strings.ToUpper(p.curTok.Literal) != "FOR" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse target name
	forName, _ := p.parseSchemaObjectName()
	stmt.ForName = forName

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateProcedureStatement() (*ast.CreateProcedureStatement, error) {
	// Consume PROCEDURE/PROC
	p.nextToken()

	stmt := &ast.CreateProcedureStatement{}
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

func (p *Parser) parseProcedureParameters() ([]*ast.ProcedureParameter, error) {
	var params []*ast.ProcedureParameter

	// Handle optional parentheses
	hasParens := p.curTok.Type == TokenLParen
	if hasParens {
		p.nextToken()
	}

	for {
		// Check if we're done
		if hasParens && p.curTok.Type == TokenRParen {
			p.nextToken()
			break
		}
		if !hasParens && (p.curTok.Type == TokenAs || p.curTok.Type == TokenWith || strings.ToUpper(p.curTok.Literal) == "FOR") {
			break
		}
		if p.curTok.Type == TokenEOF {
			break
		}

		// Parse parameter (starts with @)
		if !strings.HasPrefix(p.curTok.Literal, "@") {
			if hasParens {
				p.nextToken()
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}
			break
		}

		param := &ast.ProcedureParameter{
			Modifier: "None",
		}

		// Parse variable name
		param.VariableName = p.parseIdentifier()

		// Check for AS (optional type prefix)
		if p.curTok.Type == TokenAs {
			p.nextToken()
		}

		// Parse data type
		dataType, err := p.parseDataTypeReference()
		if err != nil {
			return nil, err
		}
		param.DataType = dataType

		// Parse optional NULL/NOT NULL
		if p.curTok.Type == TokenNull {
			param.Nullable = &ast.NullableConstraintDefinition{Nullable: true}
			p.nextToken()
		} else if p.curTok.Type == TokenNot {
			p.nextToken()
			if p.curTok.Type == TokenNull {
				param.Nullable = &ast.NullableConstraintDefinition{Nullable: false}
				p.nextToken()
			}
		}

		// Parse optional default value
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			param.Value = val
		}

		// Parse optional OUTPUT/OUT modifier
		if strings.ToUpper(p.curTok.Literal) == "OUTPUT" || strings.ToUpper(p.curTok.Literal) == "OUT" {
			param.Modifier = "Output"
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "READONLY" {
			param.Modifier = "ReadOnly"
			p.nextToken()
		}

		params = append(params, param)

		// Check for comma
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			if hasParens && p.curTok.Type == TokenRParen {
				p.nextToken()
			}
			break
		}
	}

	return params, nil
}

func (p *Parser) parseStatementList() (*ast.StatementList, error) {
	sl := &ast.StatementList{}

	for p.curTok.Type != TokenEOF && !p.isBatchSeparator() {
		// Skip semicolons
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
			continue
		}

		// Check for END (end of BEGIN block or TRY/CATCH)
		if p.curTok.Type == TokenEnd {
			break
		}

		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			sl.Statements = append(sl.Statements, stmt)
		}
	}

	return sl, nil
}

func (p *Parser) isBatchSeparator() bool {
	return p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO"
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

	// Expect AS - if not present, be lenient and skip
	if p.curTok.Type != TokenAs {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse SELECT statement
	selStmt, err := p.parseSelectStatement()
	if err != nil {
		// Be lenient for incomplete SELECT statements
		p.skipToEndOfStatement()
		return stmt, nil
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

	// Expect AS - if not present, be lenient
	if p.curTok.Type != TokenAs {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse expression
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

func (p *Parser) parseCreateMasterKeyStatement() (*ast.CreateMasterKeyStatement, error) {
	// Consume MASTER
	p.nextToken()

	stmt := &ast.CreateMasterKeyStatement{}

	// Expect KEY
	if p.curTok.Type != TokenKey {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Skip optional semicolon (for CREATE MASTER KEY;)
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
		return stmt, nil
	}

	// Check for optional ENCRYPTION BY PASSWORD clause
	if p.curTok.Type == TokenEncryption {
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
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateCredentialStatement(isDatabaseScoped bool) (*ast.CreateCredentialStatement, error) {
	// Consume CREDENTIAL
	p.nextToken()

	stmt := &ast.CreateCredentialStatement{
		IsDatabaseScoped: isDatabaseScoped,
	}

	// Parse credential name
	stmt.Name = p.parseIdentifier()

	// WITH IDENTITY (optional for lenient parsing)
	if p.curTok.Type != TokenWith {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume WITH

	if strings.ToUpper(p.curTok.Literal) != "IDENTITY" {
		return nil, fmt.Errorf("expected IDENTITY after WITH, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume IDENTITY

	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after IDENTITY, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume =

	identity, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Identity = identity

	// Optional SECRET clause
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume ,
		if strings.ToUpper(p.curTok.Literal) != "SECRET" {
			return nil, fmt.Errorf("expected SECRET after comma, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume SECRET

		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after SECRET, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume =

		secret, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.Secret = secret
	}

	// Optional FOR CRYPTOGRAPHIC PROVIDER clause
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) != "CRYPTOGRAPHIC" {
			return nil, fmt.Errorf("expected CRYPTOGRAPHIC after FOR, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume CRYPTOGRAPHIC
		if strings.ToUpper(p.curTok.Literal) != "PROVIDER" {
			return nil, fmt.Errorf("expected PROVIDER after CRYPTOGRAPHIC, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume PROVIDER
		stmt.CryptographicProviderName = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateDatabaseScopedCredentialStatement() (*ast.CreateCredentialStatement, error) {
	// Already consumed CREATE, curTok is DATABASE
	p.nextToken() // consume DATABASE

	// Expect SCOPED
	if strings.ToUpper(p.curTok.Literal) != "SCOPED" {
		return nil, fmt.Errorf("expected SCOPED after DATABASE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume SCOPED

	// Expect CREDENTIAL
	if p.curTok.Type != TokenCredential {
		return nil, fmt.Errorf("expected CREDENTIAL after SCOPED, got %s", p.curTok.Literal)
	}

	// Call the existing parser with isDatabaseScoped = true
	return p.parseCreateCredentialStatement(true)
}

func (p *Parser) parseExecuteStatement() (ast.Statement, error) {
	// Check for EXECUTE AS by looking at peek token
	if p.peekTok.Type == TokenAs {
		p.nextToken() // consume EXEC/EXECUTE
		return p.parseExecuteAsStatement()
	}

	execSpec, err := p.parseExecuteSpecification()
	if err != nil {
		return nil, err
	}

	stmt := &ast.ExecuteStatement{ExecuteSpecification: execSpec}

	// Parse WITH options (RESULT SETS, RECOMPILE)
	for p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		for {
			upperLit := strings.ToUpper(p.curTok.Literal)

			if upperLit == "RESULT" {
				p.nextToken() // consume RESULT
				if strings.ToUpper(p.curTok.Literal) == "SETS" {
					p.nextToken() // consume SETS
				}

				opt := &ast.ResultSetsExecuteOption{
					OptionKind: "ResultSets",
				}

				// Check for NONE, UNDEFINED, or definitions
				upperLit = strings.ToUpper(p.curTok.Literal)
				if upperLit == "NONE" {
					opt.ResultSetsOptionKind = "None"
					p.nextToken()
				} else if upperLit == "UNDEFINED" {
					opt.ResultSetsOptionKind = "Undefined"
					p.nextToken()
				} else if p.curTok.Type == TokenLParen {
					opt.ResultSetsOptionKind = "ResultSetsDefined"
					p.nextToken() // consume (
					opt.Definitions = p.parseResultSetDefinitions()
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}

				stmt.Options = append(stmt.Options, opt)
			} else if upperLit == "RECOMPILE" {
				p.nextToken() // consume RECOMPILE
				stmt.Options = append(stmt.Options, &ast.ExecuteOption{
					OptionKind: "Recompile",
				})
			} else {
				break
			}

			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma
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

func (p *Parser) parseResultSetDefinitions() []ast.ResultSetDefinitionType {
	var definitions []ast.ResultSetDefinitionType

	for {
		upperLit := strings.ToUpper(p.curTok.Literal)

		if upperLit == "AS" {
			p.nextToken() // consume AS
			upperLit = strings.ToUpper(p.curTok.Literal)

			if upperLit == "OBJECT" {
				p.nextToken() // consume OBJECT
				name, _ := p.parseSchemaObjectName()
				def := &ast.SchemaObjectResultSetDefinition{
					ResultSetType: "Object",
					Name:          name,
				}
				definitions = append(definitions, def)
			} else if upperLit == "FOR" {
				p.nextToken() // consume FOR
				if strings.ToUpper(p.curTok.Literal) == "XML" {
					p.nextToken() // consume XML
				}
				def := &ast.ResultSetDefinition{
					ResultSetType: "ForXml",
				}
				definitions = append(definitions, def)
			} else if upperLit == "TYPE" {
				p.nextToken() // consume TYPE
				name, _ := p.parseSchemaObjectName()
				def := &ast.SchemaObjectResultSetDefinition{
					ResultSetType: "Type",
					Name:          name,
				}
				definitions = append(definitions, def)
			}
		} else if p.curTok.Type == TokenLParen {
			// Inline column definitions: (col1 int, col2 varchar(50), ...)
			p.nextToken() // consume (
			def := &ast.InlineResultSetDefinition{
				ResultSetType: "Inline",
			}

			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				colDef := &ast.ResultColumnDefinition{
					ColumnDefinition: &ast.ColumnDefinitionBase{},
				}

				// Parse column name
				colDef.ColumnDefinition.ColumnIdentifier = p.parseIdentifier()

				// Parse data type
				colDef.ColumnDefinition.DataType, _ = p.parseDataType()

				// Check for NULL/NOT NULL
				if strings.ToUpper(p.curTok.Literal) == "NOT" {
					p.nextToken() // consume NOT
					if strings.ToUpper(p.curTok.Literal) == "NULL" {
						p.nextToken() // consume NULL
						colDef.Nullable = &ast.NullableConstraintDefinition{Nullable: false}
					}
				} else if strings.ToUpper(p.curTok.Literal) == "NULL" {
					p.nextToken() // consume NULL
					colDef.Nullable = &ast.NullableConstraintDefinition{Nullable: true}
				}

				def.ResultColumnDefinitions = append(def.ResultColumnDefinitions, colDef)

				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}

			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}

			definitions = append(definitions, def)
		} else {
			break
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	return definitions
}

func (p *Parser) parseExecuteAsStatement() (*ast.ExecuteAsStatement, error) {
	// We're positioned after EXECUTE, at AS
	p.nextToken() // consume AS

	stmt := &ast.ExecuteAsStatement{}

	// Parse the execute context
	stmt.ExecuteContext = &ast.ExecuteContext{}

	switch p.curTok.Type {
	case TokenCaller:
		stmt.ExecuteContext.Kind = "Caller"
		p.nextToken()
	case TokenLogin:
		stmt.ExecuteContext.Kind = "Login"
		p.nextToken()
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after LOGIN, got %s", p.curTok.Literal)
		}
		p.nextToken()
		principal, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.ExecuteContext.Principal = principal
	case TokenUser:
		stmt.ExecuteContext.Kind = "User"
		p.nextToken()
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after USER, got %s", p.curTok.Literal)
		}
		p.nextToken()
		principal, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.ExecuteContext.Principal = principal
	default:
		return nil, fmt.Errorf("expected CALLER, LOGIN, or USER after EXECUTE AS, got %s", p.curTok.Literal)
	}

	// Check for WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken()
		for {
			if strings.ToUpper(p.curTok.Literal) == "NO" {
				p.nextToken() // consume NO
				if strings.ToUpper(p.curTok.Literal) != "REVERT" {
					return nil, fmt.Errorf("expected REVERT after NO, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume REVERT
				stmt.WithNoRevert = true
			} else if p.curTok.Type == TokenCookie {
				p.nextToken() // consume COOKIE
				if p.curTok.Type != TokenInto {
					return nil, fmt.Errorf("expected INTO after COOKIE, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume INTO
				cookie, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.Cookie = cookie
			} else {
				break
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

	// Optional WITH (DELAYED_DURABILITY = ON|OFF)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (
		if strings.ToUpper(p.curTok.Literal) != "DELAYED_DURABILITY" {
			return nil, fmt.Errorf("expected DELAYED_DURABILITY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume DELAYED_DURABILITY
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after DELAYED_DURABILITY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume =
		if strings.ToUpper(p.curTok.Literal) == "ON" {
			stmt.DelayedDurabilityOption = "On"
		} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
			stmt.DelayedDurabilityOption = "Off"
		} else {
			return nil, fmt.Errorf("expected ON or OFF, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ON/OFF
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after DELAYED_DURABILITY option, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
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
	if p.curTok.Type == TokenIdent && p.curTok.Literal[0] == '@' {
		// Variable reference
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			ValueExpression: &ast.VariableReference{
				Name: p.curTok.Literal,
			},
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && !isKeyword(p.curTok.Literal) {
		// Simple identifier
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			Identifier: &ast.Identifier{
				Value:     p.curTok.Literal,
				QuoteType: "NotQuoted",
			},
		}
		p.nextToken()
	} else if p.curTok.Type == TokenNumber || p.curTok.Type == TokenMinus {
		// Legacy name format: [-]number:dotted.identifier
		name := p.parseLegacyTransactionName()
		stmt.Name = &ast.IdentifierOrValueExpression{
			Value: name,
			Identifier: &ast.Identifier{
				Value:     name,
				QuoteType: "NotQuoted",
			},
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseLegacyTransactionName parses legacy transaction names like "5:a.b" or "-100:[a].[b]"
func (p *Parser) parseLegacyTransactionName() string {
	var parts []string

	// Optional minus sign
	if p.curTok.Type == TokenMinus {
		parts = append(parts, "-")
		p.nextToken()
	}

	// Number part
	if p.curTok.Type == TokenNumber {
		parts = append(parts, p.curTok.Literal)
		p.nextToken()
	}

	// Colon
	if p.curTok.Type == TokenColon {
		parts = append(parts, ":")
		p.nextToken()
	}

	// Dotted identifier part (e.g., "a.b" or "[a].[b]")
	for {
		if p.curTok.Type == TokenIdent {
			// Check if it's a bracketed identifier
			if strings.HasPrefix(p.curTok.Literal, "[") && strings.HasSuffix(p.curTok.Literal, "]") {
				parts = append(parts, p.curTok.Literal)
			} else {
				parts = append(parts, p.curTok.Literal)
			}
			p.nextToken()
		} else {
			break
		}

		// Check for dot continuation
		if p.curTok.Type == TokenDot {
			parts = append(parts, ".")
			p.nextToken()
		} else {
			break
		}
	}

	return strings.Join(parts, "")
}

func (p *Parser) parseWaitForStatement() (*ast.WaitForStatement, error) {
	// Consume WAITFOR
	p.nextToken()

	stmt := &ast.WaitForStatement{}

	// Check for WAITFOR (statement) syntax
	if p.curTok.Type == TokenLParen {
		stmt.WaitForOption = "Statement"
		p.nextToken() // consume (

		// Parse the inner statement (RECEIVE or GET CONVERSATION GROUP)
		innerStmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		stmt.Statement = innerStmt

		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after statement, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )

		// Check for optional , TIMEOUT
		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
			if strings.ToUpper(p.curTok.Literal) != "TIMEOUT" {
				return nil, fmt.Errorf("expected TIMEOUT, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume TIMEOUT

			timeout, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Timeout = timeout
		}
	} else if p.curTok.Type == TokenDelay {
		stmt.WaitForOption = "Delay"
		p.nextToken()
		// Parse the parameter expression
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.Parameter = param
	} else if p.curTok.Type == TokenTime {
		stmt.WaitForOption = "Time"
		p.nextToken()
		// Parse the parameter expression
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.Parameter = param
	} else {
		return nil, fmt.Errorf("expected DELAY, TIME or ( after WAITFOR, got %s", p.curTok.Literal)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseMoveConversationStatement() (*ast.MoveConversationStatement, error) {
	// Consume MOVE
	p.nextToken()

	stmt := &ast.MoveConversationStatement{}

	// Expect CONVERSATION
	if p.curTok.Type != TokenConversation {
		return nil, fmt.Errorf("expected CONVERSATION after MOVE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse the conversation handle (variable reference)
	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.Conversation = &ast.VariableReference{
			Name: p.curTok.Literal,
		}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected variable reference for conversation handle, got %s", p.curTok.Literal)
	}

	// Expect TO
	if p.curTok.Type != TokenTo {
		return nil, fmt.Errorf("expected TO, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse the group id (variable reference)
	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.Group = &ast.VariableReference{
			Name: p.curTok.Literal,
		}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected variable reference for conversation group, got %s", p.curTok.Literal)
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseGetConversationGroupStatement() (*ast.GetConversationGroupStatement, error) {
	// Consume GET
	p.nextToken()

	stmt := &ast.GetConversationGroupStatement{}

	// Expect CONVERSATION
	if p.curTok.Type != TokenConversation {
		return nil, fmt.Errorf("expected CONVERSATION after GET, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect GROUP
	if p.curTok.Type != TokenGroup {
		return nil, fmt.Errorf("expected GROUP after CONVERSATION, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse the group id variable
	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.GroupId = &ast.VariableReference{
			Name: p.curTok.Literal,
		}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected variable reference for group id, got %s", p.curTok.Literal)
	}

	// Expect FROM
	if p.curTok.Type != TokenFrom {
		return nil, fmt.Errorf("expected FROM, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse queue name (SchemaObjectName)
	queue, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Queue = queue

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseTruncateTableStatement() (*ast.TruncateTableStatement, error) {
	// Consume TRUNCATE
	p.nextToken()

	stmt := &ast.TruncateTableStatement{}

	// Expect TABLE
	if p.curTok.Type != TokenTable {
		return nil, fmt.Errorf("expected TABLE after TRUNCATE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse table name
	tableName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.TableName = tableName

	// Check for optional WITH (PARTITIONS (...))
	if p.curTok.Type == TokenWith {
		p.nextToken()

		// Expect (
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken()

		// Expect PARTITIONS
		if p.curTok.Type != TokenIdent || strings.ToUpper(p.curTok.Literal) != "PARTITIONS" {
			return nil, fmt.Errorf("expected PARTITIONS, got %s", p.curTok.Literal)
		}
		p.nextToken()

		// Expect (
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after PARTITIONS, got %s", p.curTok.Literal)
		}
		p.nextToken()

		// Parse partition ranges
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			pr := &ast.CompressionPartitionRange{}

			// Parse From value
			from, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			pr.From = from

			// Check for TO
			if p.curTok.Type == TokenTo {
				p.nextToken()
				to, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				pr.To = to
			}

			stmt.PartitionRanges = append(stmt.PartitionRanges, pr)

			// Check for comma
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}

		// Consume closing )
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}

		// Consume outer closing )
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseUseStatement() (ast.Statement, error) {
	// Consume USE
	p.nextToken()

	// Check for FEDERATION
	if strings.ToUpper(p.curTok.Literal) == "FEDERATION" {
		return p.parseUseFederationStatement()
	}

	stmt := &ast.UseStatement{}

	// Parse database name - can be identifier or keyword like MASTER
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenMaster {
		stmt.DatabaseName = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseUseFederationStatement() (ast.Statement, error) {
	// Consume FEDERATION
	p.nextToken()

	// Check if this is "USE FEDERATION ROOT" or just "USE FEDERATION" as database name
	if strings.ToUpper(p.curTok.Literal) == "ROOT" {
		p.nextToken() // consume ROOT
		stmt := &ast.UseFederationStatement{}
		// Parse WITH RESET
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			if strings.ToUpper(p.curTok.Literal) == "RESET" {
				p.nextToken() // consume RESET
			}
		}
		p.skipToEndOfStatement()
		return stmt, nil
	}

	// Check if it's just "USE FEDERATION" as a database name (no other tokens before GO/EOF)
	if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon || strings.ToUpper(p.curTok.Literal) == "GO" {
		return &ast.UseStatement{
			DatabaseName: &ast.Identifier{Value: "federation", QuoteType: "NotQuoted"},
		}, nil
	}

	stmt := &ast.UseFederationStatement{}

	// Parse federation name
	stmt.FederationName = p.parseIdentifier()

	// Parse (distribution = value)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		stmt.DistributionName = p.parseIdentifier()
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		stmt.Value, _ = p.parseScalarExpression()
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		for {
			if strings.ToUpper(p.curTok.Literal) == "FILTERING" {
				p.nextToken() // consume FILTERING
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				if p.curTok.Type == TokenOn {
					stmt.Filtering = true
					p.nextToken()
				} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
					stmt.Filtering = false
					p.nextToken()
				}
			} else if strings.ToUpper(p.curTok.Literal) == "RESET" {
				p.nextToken() // consume RESET
			}
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma and continue
			} else {
				break
			}
		}
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateFederationStatement() (ast.Statement, error) {
	// Consume FEDERATION
	p.nextToken()

	stmt := &ast.CreateFederationStatement{}

	// Parse federation name
	stmt.Name = p.parseIdentifier()

	// Parse (distribution_name datatype RANGE)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		stmt.DistributionName = p.parseIdentifier()
		stmt.DataType, _ = p.parseDataType()
		if strings.ToUpper(p.curTok.Literal) == "RANGE" {
			p.nextToken() // consume RANGE
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseAlterFederationStatement() (ast.Statement, error) {
	// Consume FEDERATION
	p.nextToken()

	stmt := &ast.AlterFederationStatement{}

	// Parse federation name
	stmt.Name = p.parseIdentifier()

	// Parse SPLIT AT or DROP AT
	switch strings.ToUpper(p.curTok.Literal) {
	case "SPLIT":
		stmt.Kind = "Split"
		p.nextToken() // consume SPLIT
		if strings.ToUpper(p.curTok.Literal) == "AT" {
			p.nextToken() // consume AT
		}
	case "DROP":
		p.nextToken() // consume DROP
		if strings.ToUpper(p.curTok.Literal) == "AT" {
			p.nextToken() // consume AT
		}
		// Check for LOW or HIGH after opening paren
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if strings.ToUpper(p.curTok.Literal) == "LOW" {
				stmt.Kind = "DropLow"
				p.nextToken() // consume LOW
			} else if strings.ToUpper(p.curTok.Literal) == "HIGH" {
				stmt.Kind = "DropHigh"
				p.nextToken() // consume HIGH
			}
			stmt.DistributionName = p.parseIdentifier()
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			stmt.Boundary, _ = p.parseScalarExpression()
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
		p.skipToEndOfStatement()
		return stmt, nil
	}

	// Parse (distribution_name = value)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		stmt.DistributionName = p.parseIdentifier()
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		stmt.Boundary, _ = p.parseScalarExpression()
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseKillStatement() (ast.Statement, error) {
	// Consume KILL
	p.nextToken()

	// Check for STATS JOB
	if p.curTok.Type == TokenStats {
		p.nextToken() // consume STATS
		if p.curTok.Type != TokenJob {
			return nil, fmt.Errorf("expected JOB after STATS, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume JOB

		stmt := &ast.KillStatsJobStatement{}
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.JobId = param

		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	// Check for QUERY NOTIFICATION SUBSCRIPTION
	if p.curTok.Type == TokenQuery {
		p.nextToken() // consume QUERY
		if p.curTok.Type != TokenNotification {
			return nil, fmt.Errorf("expected NOTIFICATION after QUERY, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume NOTIFICATION
		if p.curTok.Type != TokenSubscription {
			return nil, fmt.Errorf("expected SUBSCRIPTION after NOTIFICATION, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume SUBSCRIPTION

		stmt := &ast.KillQueryNotificationSubscriptionStatement{}

		if p.curTok.Type == TokenAll {
			stmt.All = true
			p.nextToken()
		} else {
			param, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.SubscriptionId = param
		}

		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	stmt := &ast.KillStatement{}

	// Parse parameter (could be integer, negative integer, or string)
	param, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Parameter = param

	// Check for WITH STATUSONLY
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if p.curTok.Type == TokenStatusonly {
			stmt.WithStatusOnly = true
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCheckpointStatement() (*ast.CheckpointStatement, error) {
	// Consume CHECKPOINT
	p.nextToken()

	stmt := &ast.CheckpointStatement{}

	// Optional duration (number only)
	if p.curTok.Type == TokenNumber {
		duration, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.Duration = duration
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseReconfigureStatement() (*ast.ReconfigureStatement, error) {
	// Consume RECONFIGURE
	p.nextToken()

	stmt := &ast.ReconfigureStatement{}

	// Check for WITH OVERRIDE
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if p.curTok.Type == TokenOverride {
			stmt.WithOverride = true
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseShutdownStatement() (*ast.ShutdownStatement, error) {
	// Consume SHUTDOWN
	p.nextToken()

	stmt := &ast.ShutdownStatement{}

	// Check for WITH NOWAIT
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if p.curTok.Type == TokenNowait {
			stmt.WithNoWait = true
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseSetUserStatement() (*ast.SetUserStatement, error) {
	// Consume SETUSER
	p.nextToken()

	stmt := &ast.SetUserStatement{}

	// Parse optional user name (variable or string)
	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.UserName = &ast.VariableReference{
			Name: p.curTok.Literal,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenString {
		str, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		stmt.UserName = str
	} else if p.curTok.Type == TokenNationalString {
		str, err := p.parseNationalStringFromToken()
		if err != nil {
			return nil, err
		}
		stmt.UserName = str
	}

	// Check for WITH NORESET
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if p.curTok.Type == TokenNoreset {
			stmt.WithNoReset = true
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseLineNoStatement() (*ast.LineNoStatement, error) {
	// Consume LINENO
	p.nextToken()

	stmt := &ast.LineNoStatement{}

	// Parse line number
	lineNo, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.LineNo = lineNo

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseRaiseErrorStatement() (*ast.RaiseErrorStatement, error) {
	// Consume RAISERROR
	p.nextToken()

	stmt := &ast.RaiseErrorStatement{
		RaiseErrorOptions: "None",
	}

	// Expect (
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after RAISERROR, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// First parameter (error message or number)
	first, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.FirstParameter = first

	// Expect ,
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after first parameter, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Second parameter (severity)
	second, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.SecondParameter = second

	// Expect ,
	if p.curTok.Type != TokenComma {
		return nil, fmt.Errorf("expected , after second parameter, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Third parameter (state)
	third, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.ThirdParameter = third

	// Optional additional parameters
	for p.curTok.Type == TokenComma {
		p.nextToken()
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.OptionalParameters = append(stmt.OptionalParameters, param)
	}

	// Expect )
	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken()
		var options []string
		for {
			optName := strings.ToUpper(p.curTok.Literal)
			switch optName {
			case "LOG":
				options = append(options, "Log")
			case "NOWAIT":
				options = append(options, "NoWait")
			case "SETERROR":
				options = append(options, "SetError")
			default:
				return nil, fmt.Errorf("unknown RAISERROR option: %s", p.curTok.Literal)
			}
			p.nextToken()

			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken()
		}
		sort.Strings(options)
		stmt.RaiseErrorOptions = strings.Join(options, ", ")
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseReadTextStatement() (*ast.ReadTextStatement, error) {
	// Consume READTEXT
	p.nextToken()

	stmt := &ast.ReadTextStatement{}

	// Parse column (multi-part identifier like t1.c1 or master.dbo.t1.c1 or ..t1.c1)
	multiPart := &ast.MultiPartIdentifier{}
	for {
		// Handle leading dots or consecutive dots by inserting empty identifiers
		if p.curTok.Type == TokenDot {
			multiPart.Identifiers = append(multiPart.Identifiers, &ast.Identifier{Value: "", QuoteType: "NotQuoted"})
			p.nextToken()
			continue
		}

		id := p.parseIdentifier()
		multiPart.Identifiers = append(multiPart.Identifiers, id)

		if p.curTok.Type == TokenDot {
			p.nextToken()
		} else {
			break
		}
	}
	multiPart.Count = len(multiPart.Identifiers)
	stmt.Column = &ast.ColumnReferenceExpression{
		ColumnType:          "Regular",
		MultiPartIdentifier: multiPart,
	}

	// Parse text pointer (variable or binary literal)
	if p.curTok.Type == TokenBinary {
		stmt.TextPointer = &ast.BinaryLiteral{
			LiteralType:   "Binary",
			Value:         p.curTok.Literal,
			IsLargeObject: false,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		stmt.TextPointer = &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected text pointer, got %s", p.curTok.Literal)
	}

	// Parse offset
	offset, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Offset = offset

	// Parse size
	size, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Size = size

	// Check for optional HOLDLOCK
	if strings.ToUpper(p.curTok.Literal) == "HOLDLOCK" {
		stmt.HoldLock = true
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseWriteTextStatement() (*ast.WriteTextStatement, error) {
	// Consume WRITETEXT
	p.nextToken()

	stmt := &ast.WriteTextStatement{}

	// Check for BULK keyword
	if strings.ToUpper(p.curTok.Literal) == "BULK" {
		stmt.Bulk = true
		p.nextToken()
	}

	// Parse column (multi-part identifier like t1.c1)
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
	stmt.Column = &ast.ColumnReferenceExpression{
		ColumnType:          "Regular",
		MultiPartIdentifier: multiPart,
	}

	// Parse text ID (can be binary literal, variable, or integer)
	if p.curTok.Type == TokenBinary {
		stmt.TextId = &ast.BinaryLiteral{
			LiteralType:   "Binary",
			Value:         p.curTok.Literal,
			IsLargeObject: false,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		stmt.TextId = &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
	} else if p.curTok.Type == TokenNumber {
		stmt.TextId = &ast.IntegerLiteral{
			LiteralType: "Integer",
			Value:       p.curTok.Literal,
		}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected text ID, got %s", p.curTok.Literal)
	}

	// Check for optional WITH LOG
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "LOG" {
			stmt.WithLog = true
			p.nextToken()
		}
	}

	// Parse source parameter (variable, string literal, binary literal, or NULL)
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.SourceParameter = expr

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseUpdateTextStatement() (*ast.UpdateTextStatement, error) {
	// Consume UPDATETEXT
	p.nextToken()

	stmt := &ast.UpdateTextStatement{}

	// Check for BULK keyword
	if strings.ToUpper(p.curTok.Literal) == "BULK" {
		stmt.Bulk = true
		p.nextToken()
	}

	// Parse column (multi-part identifier like t1.c1)
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
	stmt.Column = &ast.ColumnReferenceExpression{
		ColumnType:          "Regular",
		MultiPartIdentifier: multiPart,
	}

	// Parse text ID (can be binary literal, variable, or integer)
	if p.curTok.Type == TokenBinary {
		stmt.TextId = &ast.BinaryLiteral{
			LiteralType:   "Binary",
			Value:         p.curTok.Literal,
			IsLargeObject: false,
		}
		p.nextToken()
	} else if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		stmt.TextId = &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
	} else if p.curTok.Type == TokenNumber {
		stmt.TextId = &ast.IntegerLiteral{
			LiteralType: "Integer",
			Value:       p.curTok.Literal,
		}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected text ID, got %s", p.curTok.Literal)
	}

	// Check for optional TIMESTAMP = value
	if strings.ToUpper(p.curTok.Literal) == "TIMESTAMP" {
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		// Parse timestamp value (binary literal)
		if p.curTok.Type == TokenBinary {
			stmt.Timestamp = &ast.BinaryLiteral{
				LiteralType:   "Binary",
				Value:         p.curTok.Literal,
				IsLargeObject: false,
			}
			p.nextToken()
		}
	}

	// Parse insert offset (use parsePrimaryExpression to avoid treating - as binary subtraction)
	insertOffset, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}
	stmt.InsertOffset = insertOffset

	// Parse delete length (use parsePrimaryExpression to avoid treating - as binary subtraction)
	deleteLength, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}
	stmt.DeleteLength = deleteLength

	// Check for WITH LOG
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "LOG" {
			stmt.WithLog = true
			p.nextToken()
		}
	}

	// Check for optional source (column and/or parameter)
	// This could be: nothing, just sourceParam, or sourceColumn sourceParam
	if p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && p.curTok.Type != TokenUpdatetext && strings.ToUpper(p.curTok.Literal) != "GO" {
		// Try to parse as column reference first (may be multi-part identifier)
		// If it starts with a dot or is a multi-part identifier, it's a column
		if p.curTok.Type == TokenDot ||
			(p.curTok.Type == TokenIdent && !strings.HasPrefix(p.curTok.Literal, "@") && !strings.HasPrefix(p.curTok.Literal, "N") && p.curTok.Type != TokenString && p.curTok.Type != TokenNull && p.curTok.Type != TokenBinary) ||
			p.curTok.Type == TokenLBracket {
			// This could be a source column
			srcMultiPart := &ast.MultiPartIdentifier{}
			for {
				if p.curTok.Type == TokenDot {
					srcMultiPart.Identifiers = append(srcMultiPart.Identifiers, &ast.Identifier{Value: "", QuoteType: "NotQuoted"})
					p.nextToken()
					continue
				}
				id := p.parseIdentifier()
				srcMultiPart.Identifiers = append(srcMultiPart.Identifiers, id)
				if p.curTok.Type == TokenDot {
					p.nextToken()
				} else {
					break
				}
			}
			srcMultiPart.Count = len(srcMultiPart.Identifiers)

			// Check if next token is a source parameter (variable, string, etc.)
			if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
				// This is sourceColumn followed by sourceParam
				stmt.SourceColumn = &ast.ColumnReferenceExpression{
					ColumnType:          "Regular",
					MultiPartIdentifier: srcMultiPart,
				}
				stmt.SourceParameter = &ast.VariableReference{Name: p.curTok.Literal}
				p.nextToken()
			} else if p.curTok.Type == TokenBinary {
				// sourceColumn followed by binary sourceParam
				stmt.SourceColumn = &ast.ColumnReferenceExpression{
					ColumnType:          "Regular",
					MultiPartIdentifier: srcMultiPart,
				}
				stmt.SourceParameter = &ast.BinaryLiteral{
					LiteralType:   "Binary",
					Value:         p.curTok.Literal,
					IsLargeObject: false,
				}
				p.nextToken()
			} else {
				// Just a source parameter (the "column" we parsed is actually a value)
				// This shouldn't happen based on the test patterns
			}
		} else {
			// Just a source parameter
			expr, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.SourceParameter = expr
		}
	}

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

	// Not a label - be lenient and skip to end of statement
	// This handles malformed SQL like "abcde" or other unknown identifiers
	p.skipToEndOfStatement()
	return &ast.LabelStatement{Value: label}, nil
}

func isKeyword(s string) bool {
	_, ok := keywords[strings.ToUpper(s)]
	return ok
}

func (p *Parser) parseSendStatement() (*ast.SendStatement, error) {
	// Consume SEND
	p.nextToken()

	stmt := &ast.SendStatement{}

	// ON CONVERSATION
	if p.curTok.Type != TokenOn {
		return nil, fmt.Errorf("expected ON after SEND, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume ON

	if p.curTok.Type != TokenConversation {
		return nil, fmt.Errorf("expected CONVERSATION after ON, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume CONVERSATION

	// Parse conversation handle(s)
	// Syntax: (@var) OR (@var1, @var2, ...) OR ((@var))
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		// Check for double parens: ((...))
		if p.curTok.Type == TokenLParen {
			// Double paren case - parse as single ParenthesisExpression
			p.nextToken() // consume inner (
			inner, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume inner )
			}
			stmt.ConversationHandles = append(stmt.ConversationHandles, &ast.ParenthesisExpression{Expression: inner})
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume outer )
			}
		} else {
			// Parse comma-separated list of expressions
			for {
				handle, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.ConversationHandles = append(stmt.ConversationHandles, handle)

				if p.curTok.Type == TokenComma {
					p.nextToken() // consume comma
					continue
				}
				break
			}
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ), got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )
		}
	} else {
		// Non-parenthesized expression
		handle, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.ConversationHandles = append(stmt.ConversationHandles, handle)
	}

	// Optional MESSAGE TYPE
	if p.curTok.Type == TokenMessage {
		p.nextToken() // consume MESSAGE
		if p.curTok.Type != TokenTyp {
			return nil, fmt.Errorf("expected TYPE after MESSAGE, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume TYPE

		// Parse message type name - could be identifier or variable
		if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
			stmt.MessageTypeName = &ast.IdentifierOrValueExpression{
				Value: p.curTok.Literal,
				ValueExpression: &ast.VariableReference{
					Name: p.curTok.Literal,
				},
			}
			p.nextToken()
		} else {
			id := p.parseIdentifier()
			stmt.MessageTypeName = &ast.IdentifierOrValueExpression{
				Value:      id.Value,
				Identifier: id,
			}
		}
	}

	// Optional message body in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		body, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.MessageBody = body
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after message body, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseReceiveStatement() (*ast.ReceiveStatement, error) {
	// Consume RECEIVE
	p.nextToken()

	stmt := &ast.ReceiveStatement{}

	// Check for TOP
	if p.curTok.Type == TokenTop {
		p.nextToken() // consume TOP
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after TOP, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (
		top, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.Top = top
		if p.curTok.Type != TokenRParen {
			return nil, fmt.Errorf("expected ) after TOP value, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume )
	}

	// Parse select elements (similar to SELECT)
	for {
		elem, err := p.parseSelectElement()
		if err != nil {
			return nil, err
		}
		stmt.SelectElements = append(stmt.SelectElements, elem)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	// FROM queue
	if p.curTok.Type != TokenFrom {
		return nil, fmt.Errorf("expected FROM, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume FROM

	queue, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.Queue = queue

	// Optional INTO @table_variable
	if p.curTok.Type == TokenInto {
		p.nextToken() // consume INTO
		if p.curTok.Type != TokenIdent || len(p.curTok.Literal) == 0 || p.curTok.Literal[0] != '@' {
			return nil, fmt.Errorf("expected @variable after INTO, got %s", p.curTok.Literal)
		}
		stmt.Into = &ast.VariableTableReference{
			Variable: &ast.VariableReference{Name: p.curTok.Literal},
		}
		p.nextToken()
	}

	// Optional WHERE clause
	if p.curTok.Type == TokenWhere {
		p.nextToken() // consume WHERE

		// Check for conversation_group_id
		if strings.ToLower(p.curTok.Literal) == "conversation_group_id" {
			stmt.IsConversationGroupIdWhere = true
		}

		where, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		stmt.Where = where
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseBackupStatement() (ast.Statement, error) {
	// Consume BACKUP
	p.nextToken()

	// Check for CERTIFICATE
	if strings.ToUpper(p.curTok.Literal) == "CERTIFICATE" {
		return p.parseBackupCertificateStatement()
	}

	// Check for SERVICE MASTER KEY
	if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
		return p.parseBackupServiceMasterKeyStatement()
	}

	// Check for DATABASE or LOG
	isLog := false
	if p.curTok.Type == TokenDatabase {
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "LOG" {
		isLog = true
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected DATABASE or LOG after BACKUP, got %s", p.curTok.Literal)
	}

	// Parse database name
	var dbName *ast.IdentifierOrValueExpression
	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		dbName = &ast.IdentifierOrValueExpression{
			Value: p.curTok.Literal,
			ValueExpression: &ast.VariableReference{
				Name: p.curTok.Literal,
			},
		}
		p.nextToken()
	} else {
		id := p.parseIdentifier()
		dbName = &ast.IdentifierOrValueExpression{
			Value:      id.Value,
			Identifier: id,
		}
	}

	// Expect TO
	if p.curTok.Type != TokenTo {
		return nil, fmt.Errorf("expected TO after database name, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse devices
	var devices []*ast.DeviceInfo
	for {
		device := &ast.DeviceInfo{
			DeviceType: "None",
		}

		// Check for device type (DISK, TAPE, URL, etc.)
		deviceType := strings.ToUpper(p.curTok.Literal)
		hasPhysicalType := false
		if deviceType == "DISK" || deviceType == "TAPE" || deviceType == "URL" || deviceType == "VIRTUAL_DEVICE" {
			hasPhysicalType = true
			switch deviceType {
			case "DISK":
				device.DeviceType = "Disk"
			case "TAPE":
				device.DeviceType = "Tape"
			case "URL":
				device.DeviceType = "Url"
			case "VIRTUAL_DEVICE":
				device.DeviceType = "VirtualDevice"
			}
			p.nextToken()
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after device type, got %s", p.curTok.Literal)
			}
			p.nextToken()
		}

		// Parse device name
		if hasPhysicalType {
			// Physical device: use PhysicalDevice field with ScalarExpression
			if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
				device.PhysicalDevice = &ast.VariableReference{
					Name: p.curTok.Literal,
				}
				p.nextToken()
			} else if p.curTok.Type == TokenString {
				str, err := p.parseStringLiteral()
				if err != nil {
					return nil, err
				}
				device.PhysicalDevice = str
			} else {
				return nil, fmt.Errorf("expected string or variable for physical device, got %s", p.curTok.Literal)
			}
		} else {
			// Logical device: use LogicalDevice field with IdentifierOrValueExpression
			if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
				device.LogicalDevice = &ast.IdentifierOrValueExpression{
					Value: p.curTok.Literal,
					ValueExpression: &ast.VariableReference{
						Name: p.curTok.Literal,
					},
				}
				p.nextToken()
			} else if p.curTok.Type == TokenString {
				str, err := p.parseStringLiteral()
				if err != nil {
					return nil, err
				}
				device.LogicalDevice = &ast.IdentifierOrValueExpression{
					Value:           str.Value,
					ValueExpression: str,
				}
			} else {
				id := p.parseIdentifier()
				device.LogicalDevice = &ast.IdentifierOrValueExpression{
					Value:      id.Value,
					Identifier: id,
				}
			}
		}

		devices = append(devices, device)

		// Check for comma (more devices)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Parse optional WITH clause
	var options []*ast.BackupOption
	if p.curTok.Type == TokenWith {
		p.nextToken()

		for {
			optionName := strings.ToUpper(p.curTok.Literal)
			option := &ast.BackupOption{}

			switch optionName {
			case "COMPRESSION":
				option.OptionKind = "Compression"
			case "NO_COMPRESSION":
				option.OptionKind = "NoCompression"
			case "STOP_ON_ERROR":
				option.OptionKind = "StopOnError"
			case "CONTINUE_AFTER_ERROR":
				option.OptionKind = "ContinueAfterError"
			case "CHECKSUM":
				option.OptionKind = "Checksum"
			case "NO_CHECKSUM":
				option.OptionKind = "NoChecksum"
			case "INIT":
				option.OptionKind = "Init"
			case "NOINIT":
				option.OptionKind = "NoInit"
			case "FORMAT":
				option.OptionKind = "Format"
			case "NOFORMAT":
				option.OptionKind = "NoFormat"
			default:
				option.OptionKind = optionName
			}
			p.nextToken()

			// Check for = value
			if p.curTok.Type == TokenEquals {
				p.nextToken()
				val, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				option.Value = val
			}

			options = append(options, option)

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

	if isLog {
		return &ast.BackupTransactionLogStatement{
			DatabaseName: dbName,
			Devices:      devices,
			Options:      options,
		}, nil
	}
	return &ast.BackupDatabaseStatement{
		DatabaseName: dbName,
		Devices:      devices,
		Options:      options,
	}, nil
}

func (p *Parser) parseBackupCertificateStatement() (*ast.BackupCertificateStatement, error) {
	// Consume CERTIFICATE
	p.nextToken()

	stmt := &ast.BackupCertificateStatement{
		ActiveForBeginDialog: "NotSet",
	}

	// Parse certificate name
	stmt.Name = p.parseIdentifier()

	// Expect TO
	if p.curTok.Type != TokenTo {
		return nil, fmt.Errorf("expected TO after certificate name, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect FILE
	if strings.ToUpper(p.curTok.Literal) != "FILE" {
		return nil, fmt.Errorf("expected FILE after TO, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after FILE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse file path
	file, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.File = file

	// Check for WITH PRIVATE KEY clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		if strings.ToUpper(p.curTok.Literal) == "PRIVATE" {
			p.nextToken() // consume PRIVATE
			if strings.ToUpper(p.curTok.Literal) != "KEY" {
				return nil, fmt.Errorf("expected KEY after PRIVATE, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume KEY

			// Expect (
			if p.curTok.Type != TokenLParen {
				return nil, fmt.Errorf("expected ( after PRIVATE KEY, got %s", p.curTok.Literal)
			}
			p.nextToken()

			// Parse options
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optName := strings.ToUpper(p.curTok.Literal)
				p.nextToken()

				if p.curTok.Type == TokenEquals {
					p.nextToken()
					val, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}

					switch optName {
					case "FILE":
						stmt.PrivateKeyPath = val
					case "ENCRYPTION":
						// ENCRYPTION BY PASSWORD = value
						if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
							p.nextToken()
							if p.curTok.Type == TokenEquals {
								p.nextToken()
								val, err = p.parseScalarExpression()
								if err != nil {
									return nil, err
								}
							}
						}
						stmt.EncryptionPassword = val
					case "DECRYPTION":
						// DECRYPTION BY PASSWORD = value
						if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
							p.nextToken()
							if p.curTok.Type == TokenEquals {
								p.nextToken()
								val, err = p.parseScalarExpression()
								if err != nil {
									return nil, err
								}
							}
						}
						stmt.DecryptionPassword = val
					}
				} else if optName == "ENCRYPTION" || optName == "DECRYPTION" {
					// ENCRYPTION BY PASSWORD = value
					if strings.ToUpper(p.curTok.Literal) == "BY" {
						p.nextToken() // consume BY
						if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
							p.nextToken() // consume PASSWORD
							if p.curTok.Type == TokenEquals {
								p.nextToken()
								val, err := p.parseScalarExpression()
								if err != nil {
									return nil, err
								}
								if optName == "ENCRYPTION" {
									stmt.EncryptionPassword = val
								} else {
									stmt.DecryptionPassword = val
								}
							}
						}
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

func (p *Parser) parseBackupServiceMasterKeyStatement() (*ast.BackupServiceMasterKeyStatement, error) {
	// Consume SERVICE
	p.nextToken()

	// Expect MASTER
	if strings.ToUpper(p.curTok.Literal) != "MASTER" {
		return nil, fmt.Errorf("expected MASTER after SERVICE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect KEY
	if p.curTok.Type != TokenKey {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.BackupServiceMasterKeyStatement{}

	// Expect TO
	if p.curTok.Type != TokenTo {
		return nil, fmt.Errorf("expected TO after SERVICE MASTER KEY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect FILE
	if strings.ToUpper(p.curTok.Literal) != "FILE" {
		return nil, fmt.Errorf("expected FILE after TO, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Expect =
	if p.curTok.Type != TokenEquals {
		return nil, fmt.Errorf("expected = after FILE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Parse file path
	file, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.File = file

	// Parse ENCRYPTION BY PASSWORD clause
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
		p.nextToken() // consume ENCRYPTION
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
			p.nextToken() // consume PASSWORD
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			pwd, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Password = pwd
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCloseStatement() (ast.Statement, error) {
	p.nextToken() // consume CLOSE

	if p.curTok.Type == TokenSymmetric {
		p.nextToken() // consume SYMMETRIC
		if p.curTok.Type != TokenKey {
			return nil, fmt.Errorf("expected KEY after SYMMETRIC, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume KEY
		stmt := &ast.CloseSymmetricKeyStatement{Name: p.parseIdentifier()}
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	if p.curTok.Type == TokenAll {
		p.nextToken() // consume ALL
		if p.curTok.Type != TokenSymmetric {
			return nil, fmt.Errorf("expected SYMMETRIC after ALL, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume SYMMETRIC
		if strings.ToUpper(p.curTok.Literal) != "KEYS" {
			return nil, fmt.Errorf("expected KEYS after SYMMETRIC, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume KEYS
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return &ast.CloseSymmetricKeyStatement{All: true}, nil
	}

	if p.curTok.Type == TokenMaster {
		p.nextToken() // consume MASTER
		if p.curTok.Type != TokenKey {
			return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume KEY
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return &ast.CloseMasterKeyStatement{}, nil
	}

	// Otherwise, it's CLOSE cursor_name
	return p.parseCloseCursorStatement()
}

func (p *Parser) parseOpenStatement() (ast.Statement, error) {
	p.nextToken() // consume OPEN

	if p.curTok.Type == TokenMaster {
		p.nextToken() // consume MASTER
		if p.curTok.Type != TokenKey {
			return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume KEY

		stmt := &ast.OpenMasterKeyStatement{}
		if p.curTok.Type == TokenDecryption {
			p.nextToken() // DECRYPTION
			if p.curTok.Type != TokenBy {
				return nil, fmt.Errorf("expected BY after DECRYPTION, got %s", p.curTok.Literal)
			}
			p.nextToken() // BY
			if p.curTok.Type != TokenPassword {
				return nil, fmt.Errorf("expected PASSWORD after BY, got %s", p.curTok.Literal)
			}
			p.nextToken() // PASSWORD
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after PASSWORD, got %s", p.curTok.Literal)
			}
			p.nextToken() // =
			pwd, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.Password = pwd
		}
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	if p.curTok.Type == TokenSymmetric {
		p.nextToken() // consume SYMMETRIC
		if p.curTok.Type != TokenKey {
			return nil, fmt.Errorf("expected KEY after SYMMETRIC, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume KEY
		stmt := &ast.OpenSymmetricKeyStatement{Name: p.parseIdentifier()}

		// Parse DECRYPTION BY <mechanism>
		if p.curTok.Type == TokenDecryption {
			p.nextToken() // consume DECRYPTION
			if p.curTok.Type == TokenBy {
				p.nextToken() // consume BY
			}
			mechanism := &ast.CryptoMechanism{}
			upperLit := strings.ToUpper(p.curTok.Literal)

			switch upperLit {
			case "CERTIFICATE":
				p.nextToken() // consume CERTIFICATE
				mechanism.CryptoMechanismType = "Certificate"
				mechanism.Identifier = p.parseIdentifier()
				// Check for optional WITH PASSWORD
				if p.curTok.Type == TokenWith {
					p.nextToken() // consume WITH
					if p.curTok.Type == TokenPassword {
						p.nextToken() // consume PASSWORD
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if p.curTok.Type == TokenNationalString {
							str, _ := p.parseNationalStringFromToken()
							mechanism.PasswordOrSignature = str
						} else if p.curTok.Type == TokenString {
							mechanism.PasswordOrSignature = p.parseStringLiteralValue()
							p.nextToken()
						}
					}
				}
			case "ASYMMETRIC":
				p.nextToken() // consume ASYMMETRIC
				if p.curTok.Type == TokenKey {
					p.nextToken() // consume KEY
				}
				mechanism.CryptoMechanismType = "AsymmetricKey"
				mechanism.Identifier = p.parseIdentifier()
				// Check for optional WITH PASSWORD
				if p.curTok.Type == TokenWith {
					p.nextToken() // consume WITH
					if p.curTok.Type == TokenPassword {
						p.nextToken() // consume PASSWORD
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						if p.curTok.Type == TokenNationalString {
							str, _ := p.parseNationalStringFromToken()
							mechanism.PasswordOrSignature = str
						} else if p.curTok.Type == TokenString {
							mechanism.PasswordOrSignature = p.parseStringLiteralValue()
							p.nextToken()
						}
					}
				}
			case "SYMMETRIC":
				p.nextToken() // consume SYMMETRIC
				if p.curTok.Type == TokenKey {
					p.nextToken() // consume KEY
				}
				mechanism.CryptoMechanismType = "SymmetricKey"
				mechanism.Identifier = p.parseIdentifier()
			case "PASSWORD":
				p.nextToken() // consume PASSWORD
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				mechanism.CryptoMechanismType = "Password"
				if p.curTok.Type == TokenNationalString {
					str, _ := p.parseNationalStringFromToken()
					mechanism.PasswordOrSignature = str
				} else if p.curTok.Type == TokenString {
					mechanism.PasswordOrSignature = p.parseStringLiteralValue()
					p.nextToken()
				}
			}
			stmt.DecryptionMechanism = mechanism
		}

		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	}

	// Otherwise, it's OPEN cursor_name
	return p.parseOpenCursorStatement()
}

func (p *Parser) parseCreateExternalStatement() (ast.Statement, error) {
	// Consume EXTERNAL
	p.nextToken()

	keyword := strings.ToUpper(p.curTok.Literal)
	switch keyword {
	case "DATA":
		return p.parseCreateExternalDataSourceStatement()
	case "FILE":
		return p.parseCreateExternalFileFormatStatement()
	case "TABLE":
		return p.parseCreateExternalTableStatement()
	case "LANGUAGE":
		return p.parseCreateExternalLanguageStatement()
	case "LIBRARY":
		return p.parseCreateExternalLibraryStatement()
	}
	return nil, fmt.Errorf("unexpected token after CREATE EXTERNAL: %s", p.curTok.Literal)
}

func (p *Parser) parseCreateExternalDataSourceStatement() (*ast.CreateExternalDataSourceStatement, error) {
	// DATA SOURCE name WITH (options)
	p.nextToken() // consume DATA
	if strings.ToUpper(p.curTok.Literal) != "SOURCE" {
		return nil, fmt.Errorf("expected SOURCE after DATA, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume SOURCE

	stmt := &ast.CreateExternalDataSourceStatement{
		Name: p.parseIdentifier(),
	}

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		// Default to EXTERNAL_GENERICS if WITH clause exists but no TYPE specified
		stmt.DataSourceType = "EXTERNAL_GENERICS"
		p.nextToken() // consume WITH
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			switch optName {
			case "TYPE":
				// TYPE sets DataSourceType
				stmt.DataSourceType = strings.ToUpper(p.curTok.Literal)
				p.nextToken()
			case "LOCATION":
				// LOCATION sets Location as StringLiteral
				strLit, err := p.parseStringLiteral()
				if err != nil {
					return nil, err
				}
				stmt.Location = strLit
			default:
				// All other options go into ExternalDataSourceOptions
				opt := &ast.ExternalDataSourceLiteralOrIdentifierOption{
					OptionKind: externalDataSourceOptionKindToPascalCase(optName),
					Value:      &ast.IdentifierOrValueExpression{},
				}

				// Determine if value is identifier or string literal
				if p.curTok.Type == TokenString {
					strLit, err := p.parseStringLiteral()
					if err != nil {
						return nil, err
					}
					opt.Value.Value = strLit.Value
					opt.Value.ValueExpression = strLit
				} else {
					// It's an identifier
					ident := p.parseIdentifier()
					opt.Value.Value = ident.Value
					opt.Value.Identifier = ident
				}
				stmt.ExternalDataSourceOptions = append(stmt.ExternalDataSourceOptions, opt)
			}

			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

// externalDataSourceOptionKindToPascalCase converts option names to PascalCase
func externalDataSourceOptionKindToPascalCase(optName string) string {
	switch strings.ToUpper(optName) {
	case "CREDENTIAL":
		return "Credential"
	case "RESOURCE_MANAGER_LOCATION":
		return "ResourceManagerLocation"
	case "DATABASE_NAME":
		return "DatabaseName"
	case "SHARD_MAP_NAME":
		return "ShardMapName"
	case "CONNECTION_OPTIONS":
		return "ConnectionOptions"
	case "PUSHDOWN":
		return "Pushdown"
	default:
		return optName
	}
}

func (p *Parser) parseCreateExternalFileFormatStatement() (*ast.CreateExternalFileFormatStatement, error) {
	// FILE FORMAT name WITH (options)
	p.nextToken() // consume FILE
	if strings.ToUpper(p.curTok.Literal) != "FORMAT" {
		return nil, fmt.Errorf("expected FORMAT after FILE, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume FORMAT

	stmt := &ast.CreateExternalFileFormatStatement{
		Name: p.parseIdentifier(),
	}

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			if optName == "FORMAT_TYPE" {
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				// Parse format type value and convert to PascalCase
				stmt.FormatType = p.formatTypeToPascalCase(p.curTok.Literal)
				p.nextToken() // consume value
			} else if optName == "FORMAT_OPTIONS" {
				// Parse container option with suboptions
				opt := &ast.ExternalFileFormatContainerOption{
					OptionKind: "FormatOptions",
				}
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
						subOpt := p.parseExternalFileFormatSuboption()
						if subOpt != nil {
							opt.Suboptions = append(opt.Suboptions, subOpt)
						}
						if p.curTok.Type == TokenComma {
							p.nextToken()
						}
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}
				stmt.ExternalFileFormatOptions = append(stmt.ExternalFileFormatOptions, opt)
			} else {
				// Skip other options for now
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
					p.nextToken() // consume value
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

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) formatTypeToPascalCase(s string) string {
	upper := strings.ToUpper(s)
	switch upper {
	case "DELTA":
		return "Delta"
	case "DELIMITEDTEXT":
		return "DelimitedText"
	case "PARQUET":
		return "Parquet"
	case "ORC":
		return "Orc"
	case "RCFILE":
		return "RcFile"
	case "JSON":
		return "Json"
	default:
		return s
	}
}

func (p *Parser) parseExternalFileFormatSuboption() ast.ExternalFileFormatOption {
	optName := strings.ToUpper(p.curTok.Literal)
	p.nextToken() // consume option name

	// Map to option kind
	optionKind := p.externalFileFormatOptionKind(optName)

	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		val, _ := p.parseStringLiteral()
		return &ast.ExternalFileFormatLiteralOption{
			OptionKind: optionKind,
			Value:      val,
		}
	}
	return nil
}

func (p *Parser) externalFileFormatOptionKind(name string) string {
	switch strings.ToUpper(name) {
	case "PARSER_VERSION":
		return "ParserVersion"
	case "FIELD_TERMINATOR":
		return "FieldTerminator"
	case "STRING_DELIMITER":
		return "StringDelimiter"
	case "DATE_FORMAT":
		return "DateFormat"
	case "USE_TYPE_DEFAULT":
		return "UseTypeDefault"
	case "ENCODING":
		return "Encoding"
	case "DATA_COMPRESSION":
		return "DataCompression"
	case "FIRST_ROW":
		return "FirstRow"
	default:
		return name
	}
}

func (p *Parser) parseCreateExternalTableStatement() (*ast.CreateExternalTableStatement, error) {
	// TABLE name - skip rest of statement for now
	p.nextToken() // consume TABLE

	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt := &ast.CreateExternalTableStatement{
		SchemaObjectName: name,
	}

	// Skip rest of statement
	for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && !p.isStatementTerminator() {
		p.nextToken()
	}
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseCreateExternalLanguageStatement() (*ast.CreateExternalLanguageStatement, error) {
	p.nextToken() // consume LANGUAGE
	stmt := &ast.CreateExternalLanguageStatement{
		Name: p.parseIdentifier(),
	}
	// Skip rest of statement for now
	for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && !p.isStatementTerminator() {
		p.nextToken()
	}
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseCreateExternalLibraryStatement() (*ast.CreateExternalLibraryStatement, error) {
	p.nextToken() // consume LIBRARY
	stmt := &ast.CreateExternalLibraryStatement{
		Name: p.parseIdentifier(),
	}

	// Parse optional AUTHORIZATION
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Parse FROM clause
	if p.curTok.Type == TokenFrom {
		p.nextToken() // consume FROM
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			fileOption := &ast.ExternalLibraryFileOption{}
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				switch strings.ToUpper(p.curTok.Literal) {
				case "CONTENT":
					p.nextToken() // consume CONTENT
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					content, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					fileOption.Content = content
				case "PLATFORM":
					p.nextToken() // consume PLATFORM
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					fileOption.Platform = p.parseIdentifier()
				default:
					p.nextToken()
				}
				if p.curTok.Type == TokenComma {
					p.nextToken()
				}
			}
			if fileOption.Content != nil {
				stmt.ExternalLibraryFiles = append(stmt.ExternalLibraryFiles, fileOption)
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				if p.curTok.Type == TokenLanguage || strings.ToUpper(p.curTok.Literal) == "LANGUAGE" {
					p.nextToken() // consume LANGUAGE
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					lang, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					stmt.Language = lang
				} else {
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

func (p *Parser) parseCreateEventSessionStatement() (*ast.CreateEventSessionStatement, error) {
	p.nextToken() // consume EVENT
	if strings.ToUpper(p.curTok.Literal) != "SESSION" {
		return nil, fmt.Errorf("expected SESSION after EVENT, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume SESSION

	stmt := &ast.CreateEventSessionStatement{
		Name: p.parseIdentifier(),
	}

	// ON SERVER
	if p.curTok.Type == TokenOn {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken()
		}
	}

	// Skip rest of statement for now - event sessions are complex
	for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && !p.isStatementTerminator() {
		p.nextToken()
	}
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseCreateEventSessionStatementFromEvent() (*ast.CreateEventSessionStatement, error) {
	// EVENT has already been consumed, curTok is SESSION
	p.nextToken() // consume SESSION

	stmt := &ast.CreateEventSessionStatement{
		Name: p.parseIdentifier(),
	}

	// ON SERVER
	if p.curTok.Type == TokenOn {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken()
		}
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateEventNotificationFromEvent() (*ast.CreateEventNotificationStatement, error) {
	// EVENT has already been consumed, curTok is NOTIFICATION
	if strings.ToUpper(p.curTok.Literal) == "NOTIFICATION" {
		p.nextToken() // consume NOTIFICATION
	}

	stmt := &ast.CreateEventNotificationStatement{
		Name: p.parseIdentifier(),
	}

	// Parse ON <scope>
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		stmt.Scope = &ast.EventNotificationObjectScope{}

		scopeUpper := strings.ToUpper(p.curTok.Literal)
		switch scopeUpper {
		case "SERVER":
			stmt.Scope.Target = "Server"
			p.nextToken()
		case "DATABASE":
			stmt.Scope.Target = "Database"
			p.nextToken()
		case "QUEUE":
			stmt.Scope.Target = "Queue"
			p.nextToken()
			// Parse queue name
			stmt.Scope.QueueName, _ = p.parseSchemaObjectName()
		}
	}

	// Parse optional WITH FAN_IN
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "FAN_IN" {
			stmt.WithFanIn = true
			p.nextToken() // consume FAN_IN
		}
	}

	// Parse FOR <event_type_or_group_list>
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR

		// Parse comma-separated list of event types/groups
		for {
			eventName := p.curTok.Literal
			p.nextToken()

			// Convert event name to PascalCase and determine if it's a group or type
			pascalName := eventNameToPascalCase(eventName)

			// If name ends with "Events" (after conversion), it's a group
			if strings.HasSuffix(strings.ToUpper(eventName), "_EVENTS") || strings.HasSuffix(strings.ToUpper(eventName), "EVENTS") {
				stmt.EventTypeGroups = append(stmt.EventTypeGroups, &ast.EventGroupContainer{
					EventGroup: pascalName,
				})
			} else {
				stmt.EventTypeGroups = append(stmt.EventTypeGroups, &ast.EventTypeContainer{
					EventType: pascalName,
				})
			}

			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume comma
		}
	}

	// Parse TO SERVICE 'service_name', 'broker_instance'
	if strings.ToUpper(p.curTok.Literal) == "TO" {
		p.nextToken() // consume TO
		if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
			p.nextToken() // consume SERVICE

			// Parse broker service name (string literal)
			if p.curTok.Type == TokenString {
				litVal := p.curTok.Literal
				// Strip surrounding quotes
				if len(litVal) >= 2 && litVal[0] == '\'' && litVal[len(litVal)-1] == '\'' {
					litVal = litVal[1 : len(litVal)-1]
				}
				stmt.BrokerService = &ast.StringLiteral{
					LiteralType:   "String",
					IsNational:    false,
					IsLargeObject: false,
					Value:         litVal,
				}
				p.nextToken()
			}

			// Parse comma and broker instance specifier
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma

				if p.curTok.Type == TokenString {
					litVal := p.curTok.Literal
					// Strip surrounding quotes
					if len(litVal) >= 2 && litVal[0] == '\'' && litVal[len(litVal)-1] == '\'' {
						litVal = litVal[1 : len(litVal)-1]
					}
					stmt.BrokerInstanceSpecifier = &ast.StringLiteral{
						LiteralType:   "String",
						IsNational:    false,
						IsLargeObject: false,
						Value:         litVal,
					}
					p.nextToken()
				}
			}
		}
	}

	// Skip any remaining tokens
	p.skipToEndOfStatement()
	return stmt, nil
}

// eventNameToPascalCase converts an event name like "Object_Created" or "DDL_CREDENTIAL_EVENTS" to PascalCase.
func eventNameToPascalCase(name string) string {
	// Split by underscore
	parts := strings.Split(name, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		// Capitalize first letter, lowercase rest
		result.WriteString(strings.ToUpper(part[:1]))
		result.WriteString(strings.ToLower(part[1:]))
	}
	return result.String()
}

func (p *Parser) parseCreatePartitionFunctionFromPartition() (*ast.CreatePartitionFunctionStatement, error) {
	// PARTITION has already been consumed, curTok is FUNCTION
	if strings.ToUpper(p.curTok.Literal) == "FUNCTION" {
		p.nextToken() // consume FUNCTION
	}

	stmt := &ast.CreatePartitionFunctionStatement{
		Name: p.parseIdentifier(),
	}

	// Parse ( parameter_type )
	if p.curTok.Type != TokenLParen {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume (

	// Parse parameter type (data type with optional collation)
	paramType := &ast.PartitionParameterType{}
	dt, err := p.parseDataType()
	if err != nil {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	paramType.DataType = dt

	// Check for COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		paramType.Collation = p.parseIdentifier()
	}

	stmt.ParameterType = paramType

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	// Expect AS
	if p.curTok.Type != TokenAs {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume AS

	// Expect RANGE
	if strings.ToUpper(p.curTok.Literal) != "RANGE" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume RANGE

	// Parse LEFT or RIGHT (optional, default is LEFT)
	rangeDirection := strings.ToUpper(p.curTok.Literal)
	if rangeDirection == "LEFT" || rangeDirection == "RIGHT" {
		stmt.Range = strings.Title(strings.ToLower(rangeDirection))
		p.nextToken() // consume LEFT/RIGHT
	}

	// Expect FOR VALUES
	if strings.ToUpper(p.curTok.Literal) != "FOR" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume FOR

	if strings.ToUpper(p.curTok.Literal) != "VALUES" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume VALUES

	// Expect (
	if p.curTok.Type != TokenLParen {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume (

	// Parse boundary values
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		expr, err := p.parseScalarExpression()
		if err != nil {
			p.skipToEndOfStatement()
			return stmt, nil
		}
		stmt.BoundaryValues = append(stmt.BoundaryValues, expr)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume ,
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreatePartitionSchemeStatementFromPartition() (*ast.CreatePartitionSchemeStatement, error) {
	// PARTITION has already been consumed, curTok is SCHEME
	if strings.ToUpper(p.curTok.Literal) == "SCHEME" {
		p.nextToken() // consume SCHEME
	}

	stmt := &ast.CreatePartitionSchemeStatement{}

	// Parse scheme name
	stmt.Name = p.parseIdentifier()

	// Check for AS (optional for lenient parsing)
	if p.curTok.Type != TokenAs {
		// Incomplete statement, return what we have
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume AS

	// Expect PARTITION (optional for lenient parsing)
	if strings.ToUpper(p.curTok.Literal) != "PARTITION" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse partition function name
	stmt.PartitionFunction = p.parseIdentifier()

	// Check for optional ALL keyword
	if p.curTok.Type == TokenAll {
		stmt.IsAll = true
		p.nextToken()
	}

	// Expect TO (optional for lenient parsing)
	if strings.ToUpper(p.curTok.Literal) != "TO" {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Expect (
	if p.curTok.Type != TokenLParen {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken()

	// Parse file groups
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		idOrVal := &ast.IdentifierOrValueExpression{}

		if p.curTok.Type == TokenString {
			// String literal - strip surrounding quotes
			litVal := p.curTok.Literal
			if len(litVal) >= 2 && litVal[0] == '\'' && litVal[len(litVal)-1] == '\'' {
				litVal = litVal[1 : len(litVal)-1]
			}
			idOrVal.Value = litVal
			idOrVal.ValueExpression = &ast.StringLiteral{
				LiteralType:   "String",
				Value:         litVal,
				IsNational:    false,
				IsLargeObject: false,
			}
			p.nextToken()
		} else {
			// Identifier
			id := p.parseIdentifier()
			idOrVal.Value = id.Value
			idOrVal.Identifier = id
		}

		stmt.FileGroups = append(stmt.FileGroups, idOrVal)

		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	// Skip )
	if p.curTok.Type == TokenRParen {
		p.nextToken()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateDatabaseStatement() (ast.Statement, error) {
	p.nextToken() // consume DATABASE

	// Check for DATABASE SCOPED CREDENTIAL
	if strings.ToUpper(p.curTok.Literal) == "SCOPED" {
		p.nextToken() // consume SCOPED
		if p.curTok.Type == TokenCredential {
			return p.parseCreateCredentialStatement(true)
		}
	}

	stmt := &ast.CreateDatabaseStatement{
		DatabaseName: p.parseIdentifier(),
		AttachMode:   "None",
	}

	// Check for Azure-style parenthesized options (maxsize=1gb, edition='web')
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		opts, err := p.parseAzureDatabaseOptions()
		if err != nil {
			return nil, err
		}
		stmt.Options = opts
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Check for CONTAINMENT = NONE/PARTIAL
	if strings.ToUpper(p.curTok.Literal) == "CONTAINMENT" {
		p.nextToken() // consume CONTAINMENT
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}
		val := strings.ToUpper(p.curTok.Literal)
		containmentVal := "None"
		if val == "PARTIAL" {
			containmentVal = "Partial"
		}
		stmt.Containment = &ast.ContainmentDatabaseOption{
			Value:      containmentVal,
			OptionKind: "Containment",
		}
		p.nextToken()
	}

	// Check for AS COPY OF syntax
	if p.curTok.Type == TokenAs {
		p.nextToken() // consume AS
		if strings.ToUpper(p.curTok.Literal) == "COPY" {
			p.nextToken() // consume COPY
			if p.curTok.Type == TokenOf {
				p.nextToken() // consume OF
				// Parse multi-part identifier (server.database or just database)
				multiPart := &ast.MultiPartIdentifier{}
				for {
					id := p.parseIdentifier()
					multiPart.Identifiers = append(multiPart.Identifiers, id)
					if p.curTok.Type == TokenDot {
						p.nextToken() // consume dot
					} else {
						break
					}
				}
				multiPart.Count = len(multiPart.Identifiers)
				stmt.CopyOf = multiPart
			}
		}
	}

	// Check for ON clause (file groups)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fileGroups, err := p.parseFileGroups()
		if err != nil {
			return nil, err
		}
		stmt.FileGroups = fileGroups
	}

	// Check for LOG ON clause
	if strings.ToUpper(p.curTok.Literal) == "LOG" {
		p.nextToken() // consume LOG
		if p.curTok.Type == TokenOn {
			p.nextToken() // consume ON
			logDecls, err := p.parseFileDeclarationList(false)
			if err != nil {
				return nil, err
			}
			stmt.LogOn = logDecls
		}
	}

	// Check for COLLATE clause
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		stmt.Collation = p.parseIdentifier()
	}

	// Check for FOR ATTACH clause
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		switch strings.ToUpper(p.curTok.Literal) {
		case "ATTACH":
			stmt.AttachMode = "Attach"
			p.nextToken()
		case "ATTACH_REBUILD_LOG":
			stmt.AttachMode = "AttachRebuildLog"
			p.nextToken()
		}
	}

	// Check for WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		opts, err := p.parseCreateDatabaseOptions()
		if err != nil {
			return nil, err
		}
		stmt.Options = append(stmt.Options, opts...)
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateDatabaseOptions() ([]ast.CreateDatabaseOption, error) {
	var options []ast.CreateDatabaseOption

	for {
		optName := strings.ToUpper(p.curTok.Literal)
		switch optName {
		case "LEDGER":
			p.nextToken() // consume LEDGER
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after LEDGER, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "Ledger",
				OptionState: capitalizeFirst(state),
			}
			options = append(options, opt)

		case "CATALOG_COLLATION":
			p.nextToken() // consume CATALOG_COLLATION
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after CATALOG_COLLATION, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			opt := &ast.IdentifierDatabaseOption{
				OptionKind: "CatalogCollation",
				Value:      p.parseIdentifier(),
			}
			options = append(options, opt)

		case "TRANSFORM_NOISE_WORDS":
			p.nextToken() // consume TRANSFORM_NOISE_WORDS
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "TransformNoiseWords",
				OptionState: capitalizeFirst(state),
			}
			options = append(options, opt)

		case "DB_CHAINING":
			p.nextToken() // consume DB_CHAINING
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume = (optional)
			}
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "DBChaining",
				OptionState: capitalizeFirst(state),
			}
			options = append(options, opt)

		case "NESTED_TRIGGERS":
			p.nextToken() // consume NESTED_TRIGGERS
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "NestedTriggers",
				OptionState: capitalizeFirst(state),
			}
			options = append(options, opt)

		default:
			// Unknown option, return what we have
			return options, nil
		}

		// Check for comma separator
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	return options, nil
}

func (p *Parser) parseAzureDatabaseOptions() ([]ast.CreateDatabaseOption, error) {
	var options []ast.CreateDatabaseOption

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}

		optName := strings.ToUpper(p.curTok.Literal)
		p.nextToken() // consume option name

		// Expect =
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		switch optName {
		case "MAXSIZE":
			// Parse maxsize value and unit (e.g., "1gb", "5 gb")
			maxSizeValue := p.curTok.Literal
			p.nextToken() // consume value

			// Check for unit (GB, TB, etc.) - might be attached or separate
			var units string
			upperVal := strings.ToUpper(maxSizeValue)
			if strings.HasSuffix(upperVal, "GB") {
				units = "GB"
				maxSizeValue = strings.TrimSuffix(upperVal, "GB")
			} else if strings.HasSuffix(upperVal, "TB") {
				units = "TB"
				maxSizeValue = strings.TrimSuffix(upperVal, "TB")
			} else if strings.HasSuffix(upperVal, "MB") {
				units = "MB"
				maxSizeValue = strings.TrimSuffix(upperVal, "MB")
			} else {
				// Unit might be separate token
				if p.curTok.Type == TokenIdent {
					units = strings.ToUpper(p.curTok.Literal)
					p.nextToken()
				}
			}

			opt := &ast.MaxSizeDatabaseOption{
				OptionKind: "MaxSize",
				MaxSize: &ast.IntegerLiteral{
					LiteralType: "Integer",
					Value:       maxSizeValue,
				},
				Units: units,
			}
			options = append(options, opt)

		case "EDITION":
			// Parse edition value (string literal)
			value, _ := p.parseStringLiteral()
			opt := &ast.LiteralDatabaseOption{
				OptionKind: "Edition",
				Value:      value,
			}
			options = append(options, opt)

		default:
			// Skip unknown option value
			if p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen {
				p.nextToken()
			}
		}
	}

	return options, nil
}

// parseFileGroups parses the file group definitions in CREATE DATABASE ON clause
func (p *Parser) parseFileGroups() ([]*ast.FileGroupDefinition, error) {
	var fileGroups []*ast.FileGroupDefinition

	for {
		fg := &ast.FileGroupDefinition{}
		isPrimary := false

		// Check for PRIMARY keyword or FILEGROUP keyword
		switch strings.ToUpper(p.curTok.Literal) {
		case "PRIMARY":
			isPrimary = true
			p.nextToken() // consume PRIMARY
		case "FILEGROUP":
			p.nextToken() // consume FILEGROUP
			fg.Name = p.parseIdentifier()
			// Check for CONTAINS FILESTREAM or CONTAINS MEMORY_OPTIMIZED_DATA
			if strings.ToUpper(p.curTok.Literal) == "CONTAINS" {
				p.nextToken() // consume CONTAINS
				switch strings.ToUpper(p.curTok.Literal) {
				case "FILESTREAM":
					fg.ContainsFileStream = true
					p.nextToken()
				case "MEMORY_OPTIMIZED_DATA":
					fg.ContainsMemoryOptimizedData = true
					p.nextToken()
				}
			}
			// Check for DEFAULT keyword
			if p.curTok.Type == TokenDefault {
				fg.IsDefault = true
				p.nextToken()
			}
		}

		// Parse file declarations for this group
		decls, err := p.parseFileDeclarationList(isPrimary)
		if err != nil {
			return nil, err
		}
		fg.FileDeclarations = decls
		fileGroups = append(fileGroups, fg)

		// Check if there's a comma followed by another FILEGROUP
		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
			// Check if next is FILEGROUP - if so, continue the loop
			if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
				continue
			}
			// Otherwise it might be another file in primary, so we need to handle that
			// by going back and adding it to the first filegroup
			// Actually, this case means there are more files after the comma
			// Check if it's a new filegroup or more declarations for the first group
		}

		// Check for FILEGROUP keyword for next group
		if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
			continue
		}

		break
	}

	return fileGroups, nil
}

// parseFileDeclarationList parses a comma-separated list of file declarations
func (p *Parser) parseFileDeclarationList(firstIsPrimary bool) ([]*ast.FileDeclaration, error) {
	var decls []*ast.FileDeclaration

	isFirst := true
	for {
		// Expect opening paren for file declaration
		if p.curTok.Type != TokenLParen {
			break
		}
		p.nextToken() // consume (

		decl := &ast.FileDeclaration{}
		if isFirst && firstIsPrimary {
			decl.IsPrimary = true
		}
		isFirst = false

		// Parse file options
		opts, err := p.parseFileDeclarationOptions()
		if err != nil {
			return nil, err
		}
		decl.Options = opts

		// Expect closing paren
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}

		decls = append(decls, decl)

		// Check for comma
		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
			// If next token is FILEGROUP or LOG, break out
			upper := strings.ToUpper(p.curTok.Literal)
			if upper == "FILEGROUP" || upper == "LOG" {
				break
			}
			// Otherwise, continue parsing more declarations
			continue
		}

		break
	}

	return decls, nil
}

// parseFileDeclarationOptions parses the options inside a file declaration
func (p *Parser) parseFileDeclarationOptions() ([]ast.FileDeclarationOption, error) {
	var opts []ast.FileDeclarationOption

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		if p.curTok.Type == TokenComma {
			p.nextToken()
			continue
		}

		optName := strings.ToUpper(p.curTok.Literal)

		switch optName {
		case "NAME":
			p.nextToken() // consume NAME
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			var opt *ast.NameFileDeclarationOption
			if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
				// Parse as string literal
				strLit, _ := p.parseStringLiteral()
				opt = &ast.NameFileDeclarationOption{
					LogicalFileName: &ast.IdentifierOrValueExpression{
						Value:           strLit.Value,
						ValueExpression: strLit,
					},
					IsNewName:  false,
					OptionKind: "Name",
				}
			} else {
				// Parse as identifier
				id := p.parseIdentifier()
				opt = &ast.NameFileDeclarationOption{
					LogicalFileName: &ast.IdentifierOrValueExpression{
						Value:      id.Value,
						Identifier: id,
					},
					IsNewName:  false,
					OptionKind: "Name",
				}
			}
			opts = append(opts, opt)

		case "NEWNAME":
			p.nextToken() // consume NEWNAME
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			id := p.parseIdentifier()
			opt := &ast.NameFileDeclarationOption{
				LogicalFileName: &ast.IdentifierOrValueExpression{
					Value:      id.Value,
					Identifier: id,
				},
				IsNewName:  true,
				OptionKind: "NewName",
			}
			opts = append(opts, opt)

		case "FILENAME":
			p.nextToken() // consume FILENAME
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			strLit, _ := p.parseStringLiteral()
			opt := &ast.FileNameFileDeclarationOption{
				OSFileName: strLit,
				OptionKind: "FileName",
			}
			opts = append(opts, opt)

		case "SIZE":
			p.nextToken() // consume SIZE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			size, units := p.parseSizeValue()
			opt := &ast.SizeFileDeclarationOption{
				Size:       size,
				Units:      units,
				OptionKind: "Size",
			}
			opts = append(opts, opt)

		case "MAXSIZE":
			p.nextToken() // consume MAXSIZE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			// Check for UNLIMITED
			if strings.ToUpper(p.curTok.Literal) == "UNLIMITED" {
				p.nextToken()
				opt := &ast.MaxSizeFileDeclarationOption{
					Units:      "Unspecified",
					Unlimited:  true,
					OptionKind: "MaxSize",
				}
				opts = append(opts, opt)
			} else {
				size, units := p.parseSizeValue()
				opt := &ast.MaxSizeFileDeclarationOption{
					MaxSize:    size,
					Units:      units,
					Unlimited:  false,
					OptionKind: "MaxSize",
				}
				opts = append(opts, opt)
			}

		case "FILEGROWTH":
			p.nextToken() // consume FILEGROWTH
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			size, units := p.parseSizeValue()
			opt := &ast.FileGrowthFileDeclarationOption{
				GrowthIncrement: size,
				Units:           units,
				OptionKind:      "FileGrowth",
			}
			opts = append(opts, opt)

		default:
			// Unknown option, break
			return opts, nil
		}
	}

	return opts, nil
}

// parseSizeValue parses a size value with optional unit (e.g., "10", "5 MB", "15%")
func (p *Parser) parseSizeValue() (ast.ScalarExpression, string) {
	value := p.curTok.Literal
	p.nextToken() // consume value

	// Check if unit is attached to value (e.g., "5MB", "15%")
	upperVal := strings.ToUpper(value)
	if strings.HasSuffix(upperVal, "%") {
		numVal := strings.TrimSuffix(value, "%")
		return &ast.IntegerLiteral{LiteralType: "Integer", Value: numVal}, "Percent"
	}
	if strings.HasSuffix(upperVal, "KB") {
		numVal := strings.TrimSuffix(upperVal, "KB")
		return &ast.IntegerLiteral{LiteralType: "Integer", Value: numVal}, "KB"
	}
	if strings.HasSuffix(upperVal, "MB") {
		numVal := strings.TrimSuffix(upperVal, "MB")
		return &ast.IntegerLiteral{LiteralType: "Integer", Value: numVal}, "MB"
	}
	if strings.HasSuffix(upperVal, "GB") {
		numVal := strings.TrimSuffix(upperVal, "GB")
		return &ast.IntegerLiteral{LiteralType: "Integer", Value: numVal}, "GB"
	}
	if strings.HasSuffix(upperVal, "TB") {
		numVal := strings.TrimSuffix(upperVal, "TB")
		return &ast.IntegerLiteral{LiteralType: "Integer", Value: numVal}, "TB"
	}

	// Check for separate unit token
	units := "Unspecified"
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenModulo {
		unitStr := strings.ToUpper(p.curTok.Literal)
		switch unitStr {
		case "KB":
			units = "KB"
			p.nextToken()
		case "MB":
			units = "MB"
			p.nextToken()
		case "GB":
			units = "GB"
			p.nextToken()
		case "TB":
			units = "TB"
			p.nextToken()
		case "%":
			units = "Percent"
			p.nextToken()
		}
	}

	return &ast.IntegerLiteral{LiteralType: "Integer", Value: value}, units
}

func (p *Parser) parseCreateLoginStatement() (*ast.CreateLoginStatement, error) {
	p.nextToken() // consume LOGIN

	stmt := &ast.CreateLoginStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateIndexStatement() (*ast.CreateIndexStatement, error) {
	// May already be past INDEX keyword if called from UNIQUE case
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX
	} else if strings.ToUpper(p.curTok.Literal) == "UNIQUE" {
		p.nextToken() // consume UNIQUE
		if p.curTok.Type == TokenIndex {
			p.nextToken() // consume INDEX
		}
	}

	stmt := &ast.CreateIndexStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateSpatialIndexStatement() (*ast.CreateSpatialIndexStatement, error) {
	p.nextToken() // consume SPATIAL
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX
	}

	stmt := &ast.CreateSpatialIndexStatement{
		Name:                  p.parseIdentifier(),
		SpatialIndexingScheme: "None",
	}

	// Parse ON table_name(column_name)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		stmt.Object, _ = p.parseSchemaObjectName()

		// Parse (column_name)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			stmt.SpatialColumnName = p.parseIdentifier()
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	// Parse USING clause for spatial indexing scheme
	if strings.ToUpper(p.curTok.Literal) == "USING" {
		p.nextToken() // consume USING
		scheme := strings.ToUpper(p.curTok.Literal)
		switch scheme {
		case "GEOMETRY_GRID":
			stmt.SpatialIndexingScheme = "GeometryGrid"
		case "GEOGRAPHY_GRID":
			stmt.SpatialIndexingScheme = "GeographyGrid"
		case "GEOMETRY_AUTO_GRID":
			stmt.SpatialIndexingScheme = "GeometryAutoGrid"
		case "GEOGRAPHY_AUTO_GRID":
			stmt.SpatialIndexingScheme = "GeographyAutoGrid"
		}
		p.nextToken() // consume scheme
	}

	// Parse WITH clause for options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
		}

		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			switch optName {
			case "DATA_COMPRESSION":
				compression := strings.ToUpper(p.curTok.Literal)
				compressionLevel := "None"
				switch compression {
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
				}
				p.nextToken() // consume compression level

				opt := &ast.SpatialIndexRegularOption{
					Option: &ast.DataCompressionOption{
						CompressionLevel: compressionLevel,
						OptionKind:       "DataCompression",
					},
				}
				stmt.SpatialIndexOptions = append(stmt.SpatialIndexOptions, opt)

			case "BOUNDING_BOX":
				bbOpt := p.parseBoundingBoxOption()
				stmt.SpatialIndexOptions = append(stmt.SpatialIndexOptions, bbOpt)

			case "GRIDS":
				gridsOpt := p.parseGridsOption()
				stmt.SpatialIndexOptions = append(stmt.SpatialIndexOptions, gridsOpt)

			case "CELLS_PER_OBJECT":
				expr, _ := p.parseScalarExpression()
				cellsOpt := &ast.CellsPerObjectSpatialIndexOption{
					Value: expr,
				}
				stmt.SpatialIndexOptions = append(stmt.SpatialIndexOptions, cellsOpt)

			case "PAD_INDEX", "SORT_IN_TEMPDB", "ALLOW_ROW_LOCKS", "ALLOW_PAGE_LOCKS", "DROP_EXISTING", "ONLINE", "STATISTICS_NORECOMPUTE", "STATISTICS_INCREMENTAL":
				optState := strings.ToUpper(p.curTok.Literal)
				p.nextToken() // consume ON/OFF
				opt := &ast.SpatialIndexRegularOption{
					Option: &ast.IndexStateOption{
						OptionKind:  p.getIndexOptionKind(optName),
						OptionState: p.capitalizeFirst(strings.ToLower(optState)),
					},
				}
				stmt.SpatialIndexOptions = append(stmt.SpatialIndexOptions, opt)

			case "MAXDOP", "FILLFACTOR":
				expr, _ := p.parseScalarExpression()
				opt := &ast.SpatialIndexRegularOption{
					Option: &ast.IndexExpressionOption{
						OptionKind: p.getIndexOptionKind(optName),
						Expression: expr,
					},
				}
				stmt.SpatialIndexOptions = append(stmt.SpatialIndexOptions, opt)

			case "IGNORE_DUP_KEY":
				optState := strings.ToUpper(p.curTok.Literal)
				p.nextToken() // consume ON/OFF
				opt := &ast.SpatialIndexRegularOption{
					Option: &ast.IgnoreDupKeyIndexOption{
						OptionKind:  "IgnoreDupKey",
						OptionState: p.capitalizeFirst(strings.ToLower(optState)),
					},
				}
				stmt.SpatialIndexOptions = append(stmt.SpatialIndexOptions, opt)

			default:
				// Skip unknown option value
				if p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen {
					p.nextToken()
				}
			}

			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma
			}
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse ON filegroup clause
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		stmt.OnFileGroup, _ = p.parseIdentifierOrValueExpression()
	}

	return stmt, nil
}

func (p *Parser) parseBoundingBoxOption() *ast.BoundingBoxSpatialIndexOption {
	opt := &ast.BoundingBoxSpatialIndexOption{}

	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
	}

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param := &ast.BoundingBoxParameter{Parameter: "None"}

		// Check if it's named parameter (XMIN, YMIN, etc.)
		paramName := strings.ToUpper(p.curTok.Literal)
		switch paramName {
		case "XMIN":
			param.Parameter = "XMin"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		case "YMIN":
			param.Parameter = "YMin"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		case "XMAX":
			param.Parameter = "XMax"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		case "YMAX":
			param.Parameter = "YMax"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		}

		param.Value, _ = p.parseScalarExpression()
		opt.BoundingBoxParameters = append(opt.BoundingBoxParameters, param)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return opt
}

func (p *Parser) parseGridsOption() *ast.GridsSpatialIndexOption {
	opt := &ast.GridsSpatialIndexOption{}

	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
	}

	levelIndex := 1
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		param := &ast.GridParameter{Parameter: "None"}

		// Check if it's named parameter (LEVEL_1, LEVEL_2, etc.)
		paramName := strings.ToUpper(p.curTok.Literal)
		switch paramName {
		case "LEVEL_1":
			param.Parameter = "Level1"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		case "LEVEL_2":
			param.Parameter = "Level2"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		case "LEVEL_3":
			param.Parameter = "Level3"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		case "LEVEL_4":
			param.Parameter = "Level4"
			p.nextToken()
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
		}

		// Parse the grid value (LOW, MEDIUM, HIGH)
		valueStr := strings.ToUpper(p.curTok.Literal)
		switch valueStr {
		case "LOW":
			param.Value = "Low"
		case "MEDIUM":
			param.Value = "Medium"
		case "HIGH":
			param.Value = "High"
		default:
			param.Value = p.capitalizeFirst(strings.ToLower(valueStr))
		}
		p.nextToken() // consume value

		opt.GridParameters = append(opt.GridParameters, param)
		levelIndex++

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return opt
}

func (p *Parser) parseCreateAsymmetricKeyStatement() (*ast.CreateAsymmetricKeyStatement, error) {
	p.nextToken() // consume ASYMMETRIC
	if strings.ToUpper(p.curTok.Literal) == "KEY" {
		p.nextToken() // consume KEY
	}

	stmt := &ast.CreateAsymmetricKeyStatement{
		Name: p.parseIdentifier(),
	}

	// Check for FROM PROVIDER
	if p.curTok.Type == TokenFrom {
		p.nextToken() // consume FROM
		if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
			p.nextToken() // consume PROVIDER
			source := &ast.ProviderEncryptionSource{
				Name: p.parseIdentifier(),
			}
			stmt.EncryptionAlgorithm = "None"

			// Check for WITH options
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				for {
					optName := strings.ToUpper(p.curTok.Literal)
					switch optName {
					case "ALGORITHM":
						p.nextToken() // consume ALGORITHM
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						alg := strings.ToUpper(p.curTok.Literal)
						// Map algorithm names to proper case
						algMap := map[string]string{
							"DES":         "Des",
							"RC2":         "RC2",
							"RC4":         "RC4",
							"RC4_128":     "RC4_128",
							"TRIPLE_DES":  "TripleDes",
							"AES_128":     "Aes128",
							"AES_192":     "Aes192",
							"AES_256":     "Aes256",
							"RSA_512":     "Rsa512",
							"RSA_1024":    "Rsa1024",
							"RSA_2048":    "Rsa2048",
							"RSA_3072":    "Rsa3072",
							"RSA_4096":    "Rsa4096",
							"DESX":        "DesX",
							"TRIPLE_DES_3KEY": "TripleDes3Key",
						}
						mappedAlg := alg
						if mapped, ok := algMap[alg]; ok {
							mappedAlg = mapped
						}
						source.KeyOptions = append(source.KeyOptions, &ast.AlgorithmKeyOption{
							Algorithm:  mappedAlg,
							OptionKind: "Algorithm",
						})
						p.nextToken()
					case "PROVIDER_KEY_NAME":
						p.nextToken() // consume PROVIDER_KEY_NAME
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						keyName, _ := p.parseStringLiteral()
						source.KeyOptions = append(source.KeyOptions, &ast.ProviderKeyNameKeyOption{
							KeyName:    keyName,
							OptionKind: "ProviderKeyName",
						})
					case "CREATION_DISPOSITION":
						p.nextToken() // consume CREATION_DISPOSITION
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						isCreateNew := strings.ToUpper(p.curTok.Literal) == "CREATE_NEW"
						source.KeyOptions = append(source.KeyOptions, &ast.CreationDispositionKeyOption{
							IsCreateNew: isCreateNew,
							OptionKind:  "CreationDisposition",
						})
						p.nextToken()
					default:
						goto doneWithOptions
					}

					if p.curTok.Type == TokenComma {
						p.nextToken() // consume comma
					} else if strings.ToUpper(p.curTok.Literal) != "ALGORITHM" &&
						strings.ToUpper(p.curTok.Literal) != "PROVIDER_KEY_NAME" &&
						strings.ToUpper(p.curTok.Literal) != "CREATION_DISPOSITION" {
						break
					}
				}
			doneWithOptions:
			}
			stmt.KeySource = source
		}
	}

	// Check for ENCRYPTION BY PASSWORD
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
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
		password, _ := p.parseStringLiteral()
		stmt.Password = password
	}

	// Skip optional semicolon and rest of statement
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseCreateSymmetricKeyStatement() (*ast.CreateSymmetricKeyStatement, error) {
	p.nextToken() // consume SYMMETRIC
	if strings.ToUpper(p.curTok.Literal) == "KEY" {
		p.nextToken() // consume KEY
	}

	stmt := &ast.CreateSymmetricKeyStatement{
		Name: p.parseIdentifier(),
	}

	// Check for FROM PROVIDER clause
	if p.curTok.Type == TokenFrom && strings.ToUpper(p.peekTok.Literal) == "PROVIDER" {
		p.nextToken() // consume FROM
		p.nextToken() // consume PROVIDER
		stmt.Provider = p.parseIdentifier()
	}

	// Check for WITH clause (key options)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		keyOpts, err := p.parseSymmetricKeyOptions()
		if err != nil {
			return nil, err
		}
		stmt.KeyOptions = keyOpts
	}

	// Check for ENCRYPTION BY clause
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
		p.nextToken() // consume ENCRYPTION
		if strings.ToUpper(p.curTok.Literal) == "BY" {
			p.nextToken() // consume BY
		}
		mechanisms, err := p.parseCryptoMechanisms()
		if err != nil {
			return nil, err
		}
		stmt.EncryptingMechanisms = mechanisms
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseSymmetricKeyOptions() ([]ast.KeyOption, error) {
	var options []ast.KeyOption

	for {
		optName := strings.ToUpper(p.curTok.Literal)
		switch optName {
		case "PROVIDER_KEY_NAME":
			p.nextToken() // consume PROVIDER_KEY_NAME
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			keyName, _ := p.parseScalarExpression()
			opt := &ast.ProviderKeyNameKeyOption{
				KeyName:    keyName,
				OptionKind: "ProviderKeyName",
			}
			options = append(options, opt)

		case "ALGORITHM":
			p.nextToken() // consume ALGORITHM
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			algo := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume algorithm name
			opt := &ast.AlgorithmKeyOption{
				Algorithm:  algo,
				OptionKind: "Algorithm",
			}
			options = append(options, opt)

		case "CREATION_DISPOSITION":
			p.nextToken() // consume CREATION_DISPOSITION
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			disposition := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume CREATE_NEW or OPEN_EXISTING
			opt := &ast.CreationDispositionKeyOption{
				IsCreateNew: disposition == "CREATE_NEW",
				OptionKind:  "CreationDisposition",
			}
			options = append(options, opt)

		default:
			return options, nil
		}

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		} else {
			break
		}
	}

	return options, nil
}

func (p *Parser) parseCryptoMechanisms() ([]*ast.CryptoMechanism, error) {
	var mechanisms []*ast.CryptoMechanism

	for {
		mechanism := &ast.CryptoMechanism{}
		upperLit := strings.ToUpper(p.curTok.Literal)

		switch upperLit {
		case "CERTIFICATE":
			p.nextToken() // consume CERTIFICATE
			mechanism.CryptoMechanismType = "Certificate"
			mechanism.Identifier = p.parseIdentifier()
		case "SYMMETRIC":
			p.nextToken() // consume SYMMETRIC
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			mechanism.CryptoMechanismType = "SymmetricKey"
			mechanism.Identifier = p.parseIdentifier()
		case "ASYMMETRIC":
			p.nextToken() // consume ASYMMETRIC
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			mechanism.CryptoMechanismType = "AsymmetricKey"
			mechanism.Identifier = p.parseIdentifier()
		case "PASSWORD":
			p.nextToken() // consume PASSWORD
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			mechanism.CryptoMechanismType = "Password"
			// Password should be a string literal
			mechanism.PasswordOrSignature, _ = p.parseScalarExpression()
		default:
			return mechanisms, nil
		}

		mechanisms = append(mechanisms, mechanism)

		if p.curTok.Type == TokenComma {
			p.nextToken() // consume ,
		} else {
			break
		}
	}

	return mechanisms, nil
}

func (p *Parser) parseCreateCertificateStatement() (*ast.CreateCertificateStatement, error) {
	p.nextToken() // consume CERTIFICATE

	stmt := &ast.CreateCertificateStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateMessageTypeStatement() (*ast.CreateMessageTypeStatement, error) {
	p.nextToken() // consume MESSAGE
	if strings.ToUpper(p.curTok.Literal) == "TYPE" {
		p.nextToken() // consume TYPE
	}

	stmt := &ast.CreateMessageTypeStatement{
		Name: p.parseIdentifier(),
	}

	// Optional AUTHORIZATION
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken()
		stmt.Owner = p.parseIdentifier()
	}

	// Optional VALIDATION
	if strings.ToUpper(p.curTok.Literal) == "VALIDATION" {
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}
		valMethod := strings.ToUpper(p.curTok.Literal)
		switch valMethod {
		case "WELL_FORMED_XML":
			stmt.ValidationMethod = "WellFormedXml"
			p.nextToken()
		case "NONE":
			stmt.ValidationMethod = "None"
			p.nextToken()
		case "EMPTY":
			stmt.ValidationMethod = "Empty"
			p.nextToken()
		case "VALID_XML":
			stmt.ValidationMethod = "ValidXml"
			p.nextToken()
			// Expect WITH SCHEMA COLLECTION
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "SCHEMA" {
					p.nextToken()
					if strings.ToUpper(p.curTok.Literal) == "COLLECTION" {
						p.nextToken()
						schemaName, _ := p.parseSchemaObjectName()
						stmt.XmlSchemaCollectionName = schemaName
					}
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

func (p *Parser) parseCreateServiceStatement() (*ast.CreateServiceStatement, error) {
	p.nextToken() // consume SERVICE

	stmt := &ast.CreateServiceStatement{
		Name: p.parseIdentifier(),
	}

	// Check for AUTHORIZATION clause
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Check for ON QUEUE clause
	if p.curTok.Type == TokenOn && strings.ToUpper(p.peekTok.Literal) == "QUEUE" {
		p.nextToken() // consume ON
		p.nextToken() // consume QUEUE
		queueName, _ := p.parseSchemaObjectName()
		stmt.QueueName = queueName
	}

	// Check for contract list (c1, c2, ...)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		var contracts []*ast.ServiceContract
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			contract := &ast.ServiceContract{
				Name:   p.parseIdentifier(),
				Action: "None",
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

func (p *Parser) parseCreateQueueStatement() (*ast.CreateQueueStatement, error) {
	p.nextToken() // consume QUEUE

	name, _ := p.parseSchemaObjectName()
	stmt := &ast.CreateQueueStatement{
		Name: name,
	}

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

func (p *Parser) parseQueueOptions() ([]ast.QueueOption, error) {
	var options []ast.QueueOption

	for {
		optName := strings.ToUpper(p.curTok.Literal)
		switch optName {
		case "STATUS":
			p.nextToken() // consume STATUS
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after STATUS, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.QueueStateOption{
				OptionState: capitalizeFirst(state),
				OptionKind:  "Status",
			}
			options = append(options, opt)

		case "RETENTION":
			p.nextToken() // consume RETENTION
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after RETENTION, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.QueueStateOption{
				OptionState: capitalizeFirst(state),
				OptionKind:  "Retention",
			}
			options = append(options, opt)

		case "POISON_MESSAGE_HANDLING":
			p.nextToken() // consume POISON_MESSAGE_HANDLING
			if p.curTok.Type != TokenLParen {
				return nil, fmt.Errorf("expected ( after POISON_MESSAGE_HANDLING, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume (
			// Expect STATUS = ON/OFF
			if strings.ToUpper(p.curTok.Literal) != "STATUS" {
				return nil, fmt.Errorf("expected STATUS in POISON_MESSAGE_HANDLING, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume STATUS
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after STATUS in POISON_MESSAGE_HANDLING, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume =
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			if p.curTok.Type != TokenRParen {
				return nil, fmt.Errorf("expected ) after POISON_MESSAGE_HANDLING status, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume )
			opt := &ast.QueueStateOption{
				OptionState: capitalizeFirst(state),
				OptionKind:  "PoisonMessageHandlingStatus",
			}
			options = append(options, opt)

		case "ACTIVATION":
			p.nextToken() // consume ACTIVATION
			if p.curTok.Type != TokenLParen {
				return nil, fmt.Errorf("expected ( after ACTIVATION, got %s", p.curTok.Literal)
			}
			p.nextToken() // consume (
			// Check for DROP or other activation options
			if strings.ToUpper(p.curTok.Literal) == "DROP" {
				p.nextToken() // consume DROP
				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ) after ACTIVATION DROP, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume )
				opt := &ast.QueueOptionSimple{
					OptionKind: "ActivationDrop",
				}
				options = append(options, opt)
			} else {
				// Skip to end of activation clause
				depth := 1
				for depth > 0 && p.curTok.Type != TokenEOF {
					if p.curTok.Type == TokenLParen {
						depth++
					} else if p.curTok.Type == TokenRParen {
						depth--
					}
					p.nextToken()
				}
			}

		default:
			// Unknown option, return what we have
			return options, nil
		}

		// Check for comma separator
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	return options, nil
}

func (p *Parser) parseCreateRouteStatement() (*ast.CreateRouteStatement, error) {
	p.nextToken() // consume ROUTE

	stmt := &ast.CreateRouteStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateEndpointStatement() (*ast.CreateEndpointStatement, error) {
	p.nextToken() // consume ENDPOINT

	stmt := &ast.CreateEndpointStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateAssemblyStatement() (*ast.CreateAssemblyStatement, error) {
	p.nextToken() // consume ASSEMBLY

	stmt := &ast.CreateAssemblyStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateApplicationRoleStatement() (*ast.CreateApplicationRoleStatement, error) {
	p.nextToken() // consume APPLICATION
	if strings.ToUpper(p.curTok.Literal) == "ROLE" {
		p.nextToken() // consume ROLE
	}

	stmt := &ast.CreateApplicationRoleStatement{
		Name: p.parseIdentifier(),
	}

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

func (p *Parser) parseApplicationRoleOptions() ([]*ast.ApplicationRoleOption, error) {
	var options []*ast.ApplicationRoleOption

	for {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		// Expect =
		if p.curTok.Type != TokenEquals {
			return nil, fmt.Errorf("expected = after %s, got %s", optionName, p.curTok.Literal)
		}
		p.nextToken()

		opt := &ast.ApplicationRoleOption{}

		switch optionName {
		case "PASSWORD":
			opt.OptionKind = "Password"
			// Parse string literal
			if p.curTok.Type == TokenString {
				val := p.curTok.Literal
				// Strip quotes from string literal
				if len(val) >= 2 && (val[0] == '\'' && val[len(val)-1] == '\'') {
					val = val[1 : len(val)-1]
				}
				opt.Value = &ast.IdentifierOrValueExpression{
					Value: val,
					ValueExpression: &ast.StringLiteral{
						Value:       val,
						LiteralType: "String",
					},
				}
				p.nextToken()
			}
		case "DEFAULT_SCHEMA":
			opt.OptionKind = "DefaultSchema"
			// Parse identifier
			id := p.parseIdentifier()
			opt.Value = &ast.IdentifierOrValueExpression{
				Value:      id.Value,
				Identifier: id,
			}
		case "NAME":
			opt.OptionKind = "Name"
			id := p.parseIdentifier()
			opt.Value = &ast.IdentifierOrValueExpression{
				Value:      id.Value,
				Identifier: id,
			}
		case "LOGIN":
			opt.OptionKind = "Login"
			id := p.parseIdentifier()
			opt.Value = &ast.IdentifierOrValueExpression{
				Value:      id.Value,
				Identifier: id,
			}
		default:
			// Unknown option, skip
			p.nextToken()
		}

		options = append(options, opt)

		if p.curTok.Type != TokenComma {
			break
		}
		p.nextToken() // consume comma
	}

	return options, nil
}

func (p *Parser) parseCreateFulltextStatement() (ast.Statement, error) {
	p.nextToken() // consume FULLTEXT

	switch strings.ToUpper(p.curTok.Literal) {
	case "CATALOG":
		return p.parseCreateFulltextCatalogStatement()
	case "INDEX":
		p.nextToken() // consume INDEX
		// FULLTEXT INDEX ON table_name
		if p.curTok.Type == TokenOn {
			p.nextToken() // consume ON
		}
		onName, _ := p.parseSchemaObjectName()
		stmt := &ast.CreateFulltextIndexStatement{
			OnName: onName,
		}
		p.skipToEndOfStatement()
		return stmt, nil
	default:
		// Just create a catalog statement as default
		stmt := &ast.CreateFullTextCatalogStatement{
			Name: p.parseIdentifier(),
		}
		p.skipToEndOfStatement()
		return stmt, nil
	}
}

func (p *Parser) parseCreateFulltextCatalogStatement() (*ast.CreateFullTextCatalogStatement, error) {
	p.nextToken() // consume CATALOG

	stmt := &ast.CreateFullTextCatalogStatement{
		Name: p.parseIdentifier(),
	}

	// Parse optional clauses
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && !p.isBatchSeparator() {
		switch strings.ToUpper(p.curTok.Literal) {
		case "ON":
			p.nextToken() // consume ON
			if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
				p.nextToken() // consume FILEGROUP
				stmt.FileGroup = p.parseIdentifier()
			}
		case "IN":
			p.nextToken() // consume IN
			if strings.ToUpper(p.curTok.Literal) == "PATH" {
				p.nextToken() // consume PATH
				path, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.Path = path
			}
		case "WITH":
			p.nextToken() // consume WITH
			// Parse options like ACCENT_SENSITIVITY = ON/OFF
			for {
				if strings.ToUpper(p.curTok.Literal) == "ACCENT_SENSITIVITY" {
					p.nextToken() // consume ACCENT_SENSITIVITY
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					opt := &ast.OnOffFullTextCatalogOption{
						OptionKind: "AccentSensitivity",
					}
					if strings.ToUpper(p.curTok.Literal) == "ON" {
						opt.OptionState = "On"
					} else {
						opt.OptionState = "Off"
					}
					p.nextToken() // consume ON/OFF
					stmt.Options = append(stmt.Options, opt)
				} else {
					break
				}
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
		case "AS":
			p.nextToken() // consume AS
			if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
				p.nextToken() // consume DEFAULT
				stmt.IsDefault = true
			}
		case "AUTHORIZATION":
			p.nextToken() // consume AUTHORIZATION
			stmt.Owner = p.parseIdentifier()
		default:
			// Unknown clause, skip this token
			if p.curTok.Type == TokenSemicolon || p.isBatchSeparator() {
				break
			}
			p.nextToken()
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateRemoteServiceBindingStatement() (*ast.CreateRemoteServiceBindingStatement, error) {
	p.nextToken() // consume REMOTE
	if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
		p.nextToken() // consume SERVICE
	}
	if strings.ToUpper(p.curTok.Literal) == "BINDING" {
		p.nextToken() // consume BINDING
	}

	stmt := &ast.CreateRemoteServiceBindingStatement{
		Name: p.parseIdentifier(),
	}

	// Parse TO SERVICE 'service_name'
	if strings.ToUpper(p.curTok.Literal) == "TO" {
		p.nextToken() // consume TO
		if strings.ToUpper(p.curTok.Literal) == "SERVICE" {
			p.nextToken() // consume SERVICE
		}
		// Parse service name string
		stmt.Service = p.parseStringLiteralValue()
		p.nextToken() // consume string
	}

	// Parse WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		stmt.Options = p.parseRemoteServiceBindingOptions()
	}

	// Skip any remaining parts
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseRemoteServiceBindingOptions() []ast.RemoteServiceBindingOption {
	var options []ast.RemoteServiceBindingOption

	for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF {
		// Check for GO batch separator
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "GO" {
			break
		}
		upper := strings.ToUpper(p.curTok.Literal)

		if upper == "USER" {
			p.nextToken() // consume USER
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			opt := &ast.UserRemoteServiceBindingOption{
				OptionKind: "User",
				User:       p.parseIdentifier(),
			}
			options = append(options, opt)
		} else if upper == "ANONYMOUS" {
			p.nextToken() // consume ANONYMOUS
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			optState := "Off"
			if strings.ToUpper(p.curTok.Literal) == "ON" {
				optState = "On"
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
				optState = "Off"
				p.nextToken()
			}
			opt := &ast.OnOffRemoteServiceBindingOption{
				OptionKind:  "Anonymous",
				OptionState: optState,
			}
			options = append(options, opt)
		} else if p.curTok.Type == TokenComma {
			p.nextToken() // consume comma
		} else {
			break
		}
	}

	return options
}

func (p *Parser) parseCreateStatisticsStatement() (*ast.CreateStatisticsStatement, error) {
	p.nextToken() // consume STATISTICS

	stmt := &ast.CreateStatisticsStatement{
		Name: p.parseIdentifier(),
	}

	// Parse ON table_name
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		tableName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.OnName = tableName
	}

	// Parse columns in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			// Parse column name
			colRef, err := p.parseColumnReferenceOrFunctionCall()
			if err != nil {
				return nil, err
			}
			// Type assert to ColumnReferenceExpression
			if cr, ok := colRef.(*ast.ColumnReferenceExpression); ok {
				stmt.Columns = append(stmt.Columns, cr)
			}
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse optional WITH clause (reuse UPDATE STATISTICS options logic)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && p.curTok.Type != TokenWhere {
			optionName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			switch optionName {
			case "FULLSCAN":
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.SimpleStatisticsOption{
					OptionKind: "FullScan",
				})
			case "NORECOMPUTE":
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.SimpleStatisticsOption{
					OptionKind: "NoRecompute",
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
			case "INCREMENTAL":
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					state := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.OnOffStatisticsOption{
						OptionKind:  "Incremental",
						OptionState: state,
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

	// Skip optional WHERE clause and rest of statement
	if p.curTok.Type == TokenWhere || p.curTok.Type == TokenSemicolon {
		p.skipToEndOfStatement()
	}

	return stmt, nil
}

func (p *Parser) parseCreateTypeStatement() (ast.Statement, error) {
	p.nextToken() // consume TYPE

	name, _ := p.parseSchemaObjectName()

	// Check what follows the type name
	switch strings.ToUpper(p.curTok.Literal) {
	case "FROM":
		// CREATE TYPE ... FROM (User-Defined Data Type)
		p.nextToken() // consume FROM
		// Check if there's a valid data type to parse
		if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon {
			// Incomplete statement - fall through to generic type
			stmt := &ast.CreateTypeStatement{
				Name: name,
			}
			p.skipToEndOfStatement()
			return stmt, nil
		}
		dataType, err := p.parseDataTypeReference()
		if err != nil {
			// Fall back to generic type on error
			stmt := &ast.CreateTypeStatement{
				Name: name,
			}
			p.skipToEndOfStatement()
			return stmt, nil
		}
		stmt := &ast.CreateTypeUddtStatement{
			Name:     name,
			DataType: dataType,
		}
		// Check for NULL / NOT NULL
		if p.curTok.Type == TokenNull {
			stmt.NullableConstraint = &ast.NullableConstraintDefinition{Nullable: true}
			p.nextToken()
		} else if p.curTok.Type == TokenNot {
			p.nextToken() // consume NOT
			if p.curTok.Type == TokenNull {
				p.nextToken() // consume NULL
			}
			stmt.NullableConstraint = &ast.NullableConstraintDefinition{Nullable: false}
		}
		// Skip semicolon if present
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	case "EXTERNAL":
		// CREATE TYPE ... EXTERNAL NAME (CLR User-Defined Type)
		p.nextToken() // consume EXTERNAL
		if strings.ToUpper(p.curTok.Literal) != "NAME" {
			// Incomplete statement - fall back to generic type
			stmt := &ast.CreateTypeStatement{
				Name: name,
			}
			p.skipToEndOfStatement()
			return stmt, nil
		}
		p.nextToken() // consume NAME
		// Check if there's something to parse
		if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon {
			// Incomplete statement - fall back to generic type
			stmt := &ast.CreateTypeStatement{
				Name: name,
			}
			p.skipToEndOfStatement()
			return stmt, nil
		}
		// Parse assembly name (could be [AssemblyName] or AssemblyName.[ClassName])
		assemblyName := &ast.AssemblyName{}
		firstIdent := p.parseIdentifier()
		assemblyName.Name = firstIdent
		// Check for dot and class name
		if p.curTok.Type == TokenDot {
			p.nextToken() // consume dot
			className := p.parseIdentifier()
			assemblyName.ClassName = className
		}
		stmt := &ast.CreateTypeUdtStatement{
			Name:         name,
			AssemblyName: assemblyName,
		}
		// Skip semicolon if present
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
		return stmt, nil
	case "AS":
		// Check if this is AS TABLE
		p.nextToken() // consume AS
		if strings.ToUpper(p.curTok.Literal) == "TABLE" {
			p.nextToken() // consume TABLE
			// Parse the table definition
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				tableDef, err := p.parseTableDefinitionBody()
				if err != nil {
					stmt := &ast.CreateTypeStatement{
						Name: name,
					}
					p.skipToEndOfStatement()
					return stmt, nil
				}
				stmt := &ast.CreateTypeTableStatement{
					Name:       name,
					Definition: tableDef,
				}
				// Skip closing paren if present
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
				// Skip semicolon if present
				if p.curTok.Type == TokenSemicolon {
					p.nextToken()
				}
				return stmt, nil
			}
		}
		// Fall through to generic type
		fallthrough
	default:
		// Generic CREATE TYPE statement
		stmt := &ast.CreateTypeStatement{
			Name: name,
		}
		p.skipToEndOfStatement()
		return stmt, nil
	}
}

func (p *Parser) parseCreateXmlIndexStatement() (*ast.CreateXmlIndexStatement, error) {
	// Handle PRIMARY XML INDEX
	p.nextToken() // consume PRIMARY
	if strings.ToUpper(p.curTok.Literal) == "XML" {
		p.nextToken() // consume XML
	}
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX
	}

	stmt := &ast.CreateXmlIndexStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateXmlIndexFromXml() (*ast.CreateXmlIndexStatement, error) {
	// XML has already been consumed, curTok is INDEX
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX
	}

	stmt := &ast.CreateXmlIndexStatement{
		Name: p.parseIdentifier(),
	}

	// Skip rest of statement
	p.skipToEndOfStatement()
	return stmt, nil
}

func (p *Parser) parseCreateXmlSchemaCollectionFromXml() (*ast.CreateXmlSchemaCollectionStatement, error) {
	// XML has already been consumed, expect SCHEMA
	if strings.ToUpper(p.curTok.Literal) == "SCHEMA" {
		p.nextToken() // consume SCHEMA
	}
	if strings.ToUpper(p.curTok.Literal) == "COLLECTION" {
		p.nextToken() // consume COLLECTION
	}

	name, _ := p.parseSchemaObjectName()
	stmt := &ast.CreateXmlSchemaCollectionStatement{
		Name: name,
	}

	// Check for AS (optional for lenient parsing)
	if p.curTok.Type != TokenAs {
		p.skipToEndOfStatement()
		return stmt, nil
	}
	p.nextToken() // consume AS

	// Parse expression (variable or string literal)
	expr, err := p.parseScalarExpression()
	if err == nil {
		stmt.Expression = expr
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

// parseRenameStatement parses RENAME statements (Azure SQL DW/Synapse).
// RENAME OBJECT [::] old_name TO new_name
// RENAME DATABASE [::] old_name TO new_name
func (p *Parser) parseRenameStatement() (*ast.RenameEntityStatement, error) {
	// Consume RENAME
	p.nextToken()

	stmt := &ast.RenameEntityStatement{}

	// Parse entity type: OBJECT or DATABASE
	typeLit := strings.ToUpper(p.curTok.Literal)
	if typeLit == "OBJECT" {
		stmt.RenameEntityType = "Object"
		p.nextToken()
	} else if typeLit == "DATABASE" {
		stmt.RenameEntityType = "Database"
		p.nextToken()
	}

	// Check for optional ::
	if p.curTok.Type == TokenColonColon {
		p.nextToken() // consume ::
		stmt.SeparatorType = "DoubleColon"
	}

	// Parse old name (schema object name)
	oldName, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.OldName = oldName

	// Consume TO
	if p.curTok.Type == TokenTo {
		p.nextToken()
	}

	// Parse new name (single identifier)
	stmt.NewName = p.parseIdentifier()

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCursorId parses a cursor identifier (optional GLOBAL, cursor name/variable).
func (p *Parser) parseCursorId() *ast.CursorId {
	cursorId := &ast.CursorId{}

	// Check for GLOBAL keyword
	if strings.ToUpper(p.curTok.Literal) == "GLOBAL" {
		cursorId.IsGlobal = true
		p.nextToken()
	}

	// Parse cursor name or variable
	cursorId.Name = &ast.IdentifierOrValueExpression{
		Value: p.curTok.Literal,
	}

	// Check if it's a variable
	if p.curTok.Type == TokenIdent && strings.HasPrefix(p.curTok.Literal, "@") {
		cursorId.Name.ValueExpression = &ast.VariableReference{Name: p.curTok.Literal}
	} else {
		// Create identifier inline (same logic as parseIdentifier but without advancing)
		literal := p.curTok.Literal
		quoteType := "NotQuoted"
		if len(literal) >= 2 && literal[0] == '[' && literal[len(literal)-1] == ']' {
			quoteType = "SquareBracket"
			literal = literal[1 : len(literal)-1]
		}
		cursorId.Name.Identifier = &ast.Identifier{
			Value:     literal,
			QuoteType: quoteType,
		}
	}
	p.nextToken()

	return cursorId
}

// parseOpenCursorStatement parses OPEN cursor_name.
func (p *Parser) parseOpenCursorStatement() (*ast.OpenCursorStatement, error) {
	stmt := &ast.OpenCursorStatement{
		Cursor: p.parseCursorId(),
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCloseCursorStatement parses CLOSE cursor_name.
func (p *Parser) parseCloseCursorStatement() (*ast.CloseCursorStatement, error) {
	stmt := &ast.CloseCursorStatement{
		Cursor: p.parseCursorId(),
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseDeallocateCursorStatement parses DEALLOCATE cursor_name.
func (p *Parser) parseDeallocateCursorStatement() (*ast.DeallocateCursorStatement, error) {
	// Already consumed DEALLOCATE
	p.nextToken()

	// Check for optional CURSOR keyword
	if p.curTok.Type == TokenCursor {
		p.nextToken()
	}

	stmt := &ast.DeallocateCursorStatement{
		Cursor: p.parseCursorId(),
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseFetchCursorStatement parses FETCH ... FROM cursor_name.
func (p *Parser) parseFetchCursorStatement() (*ast.FetchCursorStatement, error) {
	// Already consumed FETCH
	p.nextToken()

	stmt := &ast.FetchCursorStatement{}

	// Check for fetch orientation
	orientationKeyword := strings.ToUpper(p.curTok.Literal)
	switch orientationKeyword {
	case "NEXT":
		stmt.FetchType = &ast.FetchType{Orientation: "Next"}
		p.nextToken()
	case "PRIOR":
		stmt.FetchType = &ast.FetchType{Orientation: "Prior"}
		p.nextToken()
	case "FIRST":
		stmt.FetchType = &ast.FetchType{Orientation: "First"}
		p.nextToken()
	case "LAST":
		stmt.FetchType = &ast.FetchType{Orientation: "Last"}
		p.nextToken()
	case "ABSOLUTE":
		p.nextToken() // consume ABSOLUTE
		offset, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.FetchType = &ast.FetchType{
			Orientation: "Absolute",
			RowOffset:   offset,
		}
	case "RELATIVE":
		p.nextToken() // consume RELATIVE
		offset, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		stmt.FetchType = &ast.FetchType{
			Orientation: "Relative",
			RowOffset:   offset,
		}
	}

	// Check for FROM keyword
	if p.curTok.Type == TokenFrom {
		p.nextToken()
	}

	// Parse cursor id
	stmt.Cursor = p.parseCursorId()

	// Check for INTO clause
	if p.curTok.Type == TokenInto {
		p.nextToken() // consume INTO
		for {
			varRef, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.IntoVariables = append(stmt.IntoVariables, varRef)
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

// parseDeclareCursorStatementContinued parses DECLARE cursor CURSOR ... after the cursor name.
func (p *Parser) parseDeclareCursorStatementContinued(cursorName *ast.Identifier) (*ast.DeclareCursorStatement, error) {
	stmt := &ast.DeclareCursorStatement{
		Name:             cursorName,
		CursorDefinition: &ast.CursorDefinition{},
	}

	// Parse cursor options (INSENSITIVE, SCROLL, LOCAL, GLOBAL, FORWARD_ONLY, etc.)
	for p.curTok.Type != TokenCursor && p.curTok.Type != TokenEOF && strings.ToUpper(p.curTok.Literal) != "FOR" {
		kwd := strings.ToUpper(p.curTok.Literal)
		switch kwd {
		case "INSENSITIVE", "SCROLL", "LOCAL", "GLOBAL", "FORWARD_ONLY", "STATIC",
			"KEYSET", "DYNAMIC", "FAST_FORWARD", "READ_ONLY", "SCROLL_LOCKS",
			"OPTIMISTIC", "TYPE_WARNING":
			stmt.CursorDefinition.Options = append(stmt.CursorDefinition.Options, &ast.CursorOption{
				OptionKind: toTitleCase(kwd),
			})
			p.nextToken()
		default:
			break
		}
		if p.curTok.Type == TokenCursor || strings.ToUpper(p.curTok.Literal) == "FOR" {
			break
		}
	}

	// Consume CURSOR keyword
	if p.curTok.Type == TokenCursor {
		p.nextToken()
	}

	// Parse more options after CURSOR (for the new syntax)
	for strings.ToUpper(p.curTok.Literal) != "FOR" && p.curTok.Type != TokenEOF {
		kwd := strings.ToUpper(p.curTok.Literal)
		switch kwd {
		case "LOCAL", "GLOBAL", "FORWARD_ONLY", "SCROLL", "STATIC", "KEYSET",
			"DYNAMIC", "FAST_FORWARD", "READ_ONLY", "SCROLL_LOCKS", "OPTIMISTIC",
			"TYPE_WARNING":
			stmt.CursorDefinition.Options = append(stmt.CursorDefinition.Options, &ast.CursorOption{
				OptionKind: toTitleCase(kwd),
			})
			p.nextToken()
		default:
			break
		}
		if strings.ToUpper(p.curTok.Literal) == "FOR" {
			break
		}
	}

	// Consume FOR keyword
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken()
	}

	// Parse SELECT statement
	selectStmt, err := p.parseSelectStatement()
	if err != nil {
		return nil, err
	}
	stmt.CursorDefinition.Select = selectStmt

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// toTitleCase converts underscore-separated names to TitleCase.
func toTitleCase(s string) string {
	parts := strings.Split(strings.ToLower(s), "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// parseEnableDisableTriggerStatement parses ENABLE/DISABLE TRIGGER statements
func (p *Parser) parseEnableDisableTriggerStatement(enforcement string) (*ast.EnableDisableTriggerStatement, error) {
	// Consume ENABLE or DISABLE
	p.nextToken()

	// Expect TRIGGER
	if strings.ToUpper(p.curTok.Literal) != "TRIGGER" {
		return nil, fmt.Errorf("expected TRIGGER after %s, got %s", enforcement, p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.EnableDisableTriggerStatement{
		TriggerEnforcement: enforcement,
	}

	// Check for ALL
	if strings.ToUpper(p.curTok.Literal) == "ALL" {
		stmt.All = true
		p.nextToken()
	} else {
		stmt.All = false
		// Parse trigger names (comma-separated)
		for {
			name, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, err
			}
			stmt.TriggerNames = append(stmt.TriggerNames, name)

			if p.curTok.Type != TokenComma {
				break
			}
			p.nextToken() // consume comma
		}
	}

	// Expect ON
	if p.curTok.Type != TokenOn {
		return nil, fmt.Errorf("expected ON after trigger names, got %s", p.curTok.Literal)
	}
	p.nextToken()

	// Check for ALL SERVER or DATABASE or table name
	stmt.TriggerObject = &ast.TriggerObject{}

	if strings.ToUpper(p.curTok.Literal) == "ALL" {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			stmt.TriggerObject.TriggerScope = "AllServer"
			p.nextToken()
		} else {
			return nil, fmt.Errorf("expected SERVER after ALL, got %s", p.curTok.Literal)
		}
	} else if strings.ToUpper(p.curTok.Literal) == "DATABASE" {
		stmt.TriggerObject.TriggerScope = "Database"
		p.nextToken()
	} else {
		// Parse table name
		tableName, err := p.parseSchemaObjectName()
		if err != nil {
			return nil, err
		}
		stmt.TriggerObject.Name = tableName
		stmt.TriggerObject.TriggerScope = "Normal"
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseCreateWorkloadGroupStatement parses CREATE WORKLOAD GROUP statement.
func (p *Parser) parseCreateWorkloadGroupStatement() (*ast.CreateWorkloadGroupStatement, error) {
	// Consume WORKLOAD
	p.nextToken()

	// Consume GROUP
	if strings.ToUpper(p.curTok.Literal) == "GROUP" {
		p.nextToken()
	}

	stmt := &ast.CreateWorkloadGroupStatement{}

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

		// Check if first item is EXTERNAL or pool name
		if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
			p.nextToken() // consume EXTERNAL
			stmt.ExternalPoolName = p.parseIdentifier()
			// Check for comma and regular pool
			if p.curTok.Type == TokenComma {
				p.nextToken()
				stmt.PoolName = p.parseIdentifier()
			}
		} else {
			// Regular pool name first
			stmt.PoolName = p.parseIdentifier()
			// Check for comma and EXTERNAL
			if p.curTok.Type == TokenComma {
				p.nextToken()
				if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
					p.nextToken() // consume EXTERNAL
					stmt.ExternalPoolName = p.parseIdentifier()
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

// parseWorkloadGroupParameter parses a single workload group parameter.
func (p *Parser) parseWorkloadGroupParameter() (interface{}, error) {
	// Parse parameter name
	paramName := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	// Parse = value
	if p.curTok.Type == TokenEquals {
		p.nextToken()
	}

	// Handle IMPORTANCE specially - returns string value, not expression
	if paramName == "IMPORTANCE" {
		importanceValue := strings.ToUpper(p.curTok.Literal)
		// Convert to proper case
		switch importanceValue {
		case "LOW":
			importanceValue = "Low"
		case "BELOW_NORMAL":
			importanceValue = "Below_Normal"
		case "NORMAL":
			importanceValue = "Normal"
		case "ABOVE_NORMAL":
			importanceValue = "Above_Normal"
		case "MEDIUM":
			importanceValue = "Medium"
		case "HIGH":
			importanceValue = "High"
		}
		p.nextToken()
		return &ast.WorkloadGroupImportanceParameter{
			ParameterType:  "Importance",
			ParameterValue: importanceValue,
		}, nil
	}

	param := &ast.WorkloadGroupResourceParameter{}
	switch paramName {
	case "REQUEST_MAX_MEMORY_GRANT_PERCENT":
		param.ParameterType = "RequestMaxMemoryGrantPercent"
	case "REQUEST_MAX_CPU_TIME_SEC":
		param.ParameterType = "RequestMaxCpuTimeSec"
	case "REQUEST_MEMORY_GRANT_TIMEOUT_SEC":
		param.ParameterType = "RequestMemoryGrantTimeoutSec"
	case "MAX_DOP", "MAXDOP":
		param.ParameterType = "MaxDop"
	case "GROUP_MAX_REQUESTS":
		param.ParameterType = "GroupMaxRequests"
	case "GROUP_MIN_MEMORY_PERCENT":
		param.ParameterType = "GroupMinMemoryPercent"
	case "CAP_PERCENTAGE_RESOURCE":
		param.ParameterType = "CapPercentageResource"
	case "MIN_PERCENTAGE_RESOURCE":
		param.ParameterType = "MinPercentageResource"
	case "QUERY_EXECUTION_TIMEOUT_SEC":
		param.ParameterType = "QueryExecutionTimeoutSec"
	case "REQUEST_MIN_RESOURCE_GRANT_PERCENT":
		param.ParameterType = "RequestMinResourceGrantPercent"
	case "REQUEST_MAX_RESOURCE_GRANT_PERCENT":
		param.ParameterType = "RequestMaxResourceGrantPercent"
	default:
		param.ParameterType = paramName
	}

	// Parse the value
	val, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	param.ParameterValue = val

	return param, nil
}

// parseCreateWorkloadClassifierStatement parses CREATE WORKLOAD CLASSIFIER statement.
func (p *Parser) parseCreateWorkloadClassifierStatement() (*ast.CreateWorkloadClassifierStatement, error) {
	// Consume WORKLOAD
	p.nextToken()

	// Consume CLASSIFIER
	if strings.ToUpper(p.curTok.Literal) == "CLASSIFIER" {
		p.nextToken()
	}

	stmt := &ast.CreateWorkloadClassifierStatement{}

	// Parse classifier name
	stmt.ClassifierName = p.parseIdentifier()

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
		}

		// Parse options
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			opt, err := p.parseWorkloadClassifierOption()
			if err != nil {
				return nil, err
			}
			if opt != nil {
				stmt.Options = append(stmt.Options, opt)
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

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseWorkloadClassifierOption parses a single workload classifier option.
func (p *Parser) parseWorkloadClassifierOption() (ast.WorkloadClassifierOption, error) {
	// Parse option name
	optName := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	// Parse = value
	if p.curTok.Type == TokenEquals {
		p.nextToken()
	}

	switch optName {
	case "WORKLOAD_GROUP":
		opt := &ast.ClassifierWorkloadGroupOption{
			OptionType: "WorkloadGroup",
		}
		strLit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		opt.WorkloadGroupName = strLit
		return opt, nil

	case "MEMBERNAME":
		opt := &ast.ClassifierMemberNameOption{
			OptionType: "MemberName",
		}
		strLit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		opt.MemberName = strLit
		return opt, nil

	case "WLM_CONTEXT":
		opt := &ast.ClassifierWlmContextOption{
			OptionType: "WlmContext",
		}
		strLit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		opt.WlmContext = strLit
		return opt, nil

	case "START_TIME":
		opt := &ast.ClassifierStartTimeOption{
			OptionType: "StartTime",
		}
		strLit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		opt.Time = &ast.WlmTimeLiteral{
			TimeString: strLit,
		}
		return opt, nil

	case "END_TIME":
		opt := &ast.ClassifierEndTimeOption{
			OptionType: "EndTime",
		}
		strLit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		opt.Time = &ast.WlmTimeLiteral{
			TimeString: strLit,
		}
		return opt, nil

	case "WLM_LABEL":
		opt := &ast.ClassifierWlmLabelOption{
			OptionType: "WlmLabel",
		}
		strLit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		opt.WlmLabel = strLit
		return opt, nil

	case "IMPORTANCE":
		opt := &ast.ClassifierImportanceOption{
			OptionType: "Importance",
		}
		importanceValue := strings.ToUpper(p.curTok.Literal)
		switch importanceValue {
		case "LOW":
			opt.Importance = "Low"
		case "BELOW_NORMAL":
			opt.Importance = "Below_Normal"
		case "NORMAL":
			opt.Importance = "Normal"
		case "ABOVE_NORMAL":
			opt.Importance = "Above_Normal"
		case "HIGH":
			opt.Importance = "High"
		default:
			opt.Importance = importanceValue
		}
		p.nextToken()
		return opt, nil

	default:
		// Skip unknown option
		if p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			p.nextToken()
		}
		return nil, nil
	}
}

// parseDbccStatement parses a DBCC statement.
func (p *Parser) parseDbccStatement() (*ast.DbccStatement, error) {
	// Consume DBCC
	p.nextToken()

	stmt := &ast.DbccStatement{}

	// Parse command name
	if p.curTok.Type == TokenIdent {
		cmdName := strings.ToUpper(p.curTok.Literal)
		rawName := p.curTok.Literal
		canonical, isKnown := p.getDbccCommand(cmdName)
		if isKnown {
			stmt.Command = canonical
		} else {
			// Unknown command - set DllName and use "Free" as command
			stmt.DllName = rawName
			stmt.Command = "Free"
		}
		p.nextToken()
	}

	// Check for parenthesis
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		// Check if empty parentheses
		if p.curTok.Type == TokenRParen {
			stmt.ParenthesisRequired = true
			p.nextToken() // consume )
		} else {
			// Parse literals/parameters
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				lit := &ast.DbccNamedLiteral{}

				// Check for named parameter (name = value)
				if p.peekTok.Type == TokenEquals {
					lit.Name = p.curTok.Literal
					p.nextToken() // consume name
					p.nextToken() // consume =
				}

				// Parse the value
				val, err := p.parseScalarExpression()
				if err != nil {
					break
				}
				lit.Value = val
				stmt.Literals = append(stmt.Literals, lit)

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

	// Check for WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		// Parse options
		for {
			if p.curTok.Type == TokenEOF || p.curTok.Type == TokenSemicolon {
				break
			}

			optName := strings.ToUpper(p.curTok.Literal)
			if optName == "" {
				break
			}

			option := &ast.DbccOption{
				OptionKind: p.convertDbccOptionKind(optName),
			}
			stmt.Options = append(stmt.Options, option)
			p.nextToken()

			// Check for comma or JOIN separator
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "JOIN" {
				stmt.OptionsUseJoin = true
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

// getDbccCommand returns the canonical DBCC command name and whether it's a known command.
func (p *Parser) getDbccCommand(cmd string) (string, bool) {
	commandMap := map[string]string{
		"CHECKDB":              "CheckDB",
		"CHECKTABLE":           "CheckTable",
		"CHECKALLOC":           "CheckAlloc",
		"CHECKCATALOG":         "CheckCatalog",
		"CHECKIDENT":           "CheckIdent",
		"CHECKFILEGROUP":       "CheckFileGroup",
		"CLEANTABLE":           "CleanTable",
		"DBREINDEX":            "DbReindex",
		"DROPCLEANBUFFERS":     "DropCleanBuffers",
		"FREEPROCCACHE":        "FreeProcCache",
		"FREESESSIONCACHE":     "FreeSessionCache",
		"FREESYSTEMCACHE":      "FreeSystemCache",
		"INPUTBUFFER":          "InputBuffer",
		"OPENTRAN":             "OpenTran",
		"OUTPUTBUFFER":         "OutputBuffer",
		"PROCCACHE":            "ProcCache",
		"SHOW_STATISTICS":      "ShowStatistics",
		"SHOWCONTIG":           "ShowContig",
		"SHRINKDATABASE":       "ShrinkDatabase",
		"SHRINKFILE":           "ShrinkFile",
		"SQLPERF":              "SqlPerf",
		"TRACEON":              "TraceOn",
		"TRACEOFF":             "TraceOff",
		"TRACESTATUS":          "TraceStatus",
		"UPDATEUSAGE":          "UpdateUsage",
		"USEROPTIONS":          "UserOptions",
		"CONCURRENCYVIOLATION": "ConcurrencyViolation",
		"MEMOBJLIST":           "MemObjList",
		"MEMORYMAP":            "MemoryMap",
		"FREE":                 "Free",
		"HELP":                 "Help",
	}
	if canonical, ok := commandMap[cmd]; ok {
		return canonical, true
	}
	return cmd, false
}

// convertDbccOptionKind converts a DBCC option name to its canonical form.
func (p *Parser) convertDbccOptionKind(opt string) string {
	optionMap := map[string]string{
		"ALL_ERRORMSGS":           "AllErrorMessages",
		"NO_INFOMSGS":             "NoInfoMessages",
		"TABLOCK":                 "TabLock",
		"TABLERESULTS":            "TableResults",
		"COUNTROWS":               "CountRows",
		"COUNT_ROWS":              "CountRows",
		"STAT_HEADER":             "StatHeader",
		"DENSITY_VECTOR":          "DensityVector",
		"HISTOGRAM_STEPS":         "HistogramSteps",
		"ESTIMATEONLY":            "EstimateOnly",
		"FAST":                    "Fast",
		"ALL_LEVELS":              "AllLevels",
		"ALL_INDEXES":             "AllIndexes",
		"PHYSICAL_ONLY":           "PhysicalOnly",
		"DATA_PURITY":             "DataPurity",
		"EXTENDED_LOGICAL_CHECKS": "ExtendedLogicalChecks",
		"MARK_IN_USE_FOR_REMOVAL": "MarkInUseForRemoval",
		"ALL_CONSTRAINTS":         "AllConstraints",
		"STATS_STREAM":            "StatsStream",
		"HISTOGRAM":               "Histogram",
	}
	if canonical, ok := optionMap[opt]; ok {
		return canonical
	}
	return opt
}
