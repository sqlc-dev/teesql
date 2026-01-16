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
		// Parse optional CLUSTERED/NONCLUSTERED/HASH
		if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
			constraint.Clustered = true
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
			constraint.Clustered = false
			p.nextToken()
			// Check for HASH suffix
			if strings.ToUpper(p.curTok.Literal) == "HASH" {
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
				p.nextToken()
			} else {
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
			}
		} else if strings.ToUpper(p.curTok.Literal) == "HASH" {
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
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
		// Parse WITH (index_options)
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optionName := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					expr, _ := p.parseScalarExpression()
					option := &ast.IndexExpressionOption{
						OptionKind: p.getIndexOptionKind(optionName),
						Expression: expr,
					}
					constraint.IndexOptions = append(constraint.IndexOptions, option)
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
		return constraint, nil
	} else if upperLit == "UNIQUE" {
		p.nextToken() // consume UNIQUE
		constraint := &ast.UniqueConstraintDefinition{
			IsPrimaryKey: false,
		}
		// Parse optional CLUSTERED/NONCLUSTERED/HASH
		if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
			constraint.Clustered = true
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
			constraint.Clustered = false
			p.nextToken()
			// Check for HASH suffix
			if strings.ToUpper(p.curTok.Literal) == "HASH" {
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
				p.nextToken()
			} else {
				constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
			}
		} else if strings.ToUpper(p.curTok.Literal) == "HASH" {
			constraint.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
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
		// Parse WITH (index_options)
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optionName := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					expr, _ := p.parseScalarExpression()
					option := &ast.IndexExpressionOption{
						OptionKind: p.getIndexOptionKind(optionName),
						Expression: expr,
					}
					constraint.IndexOptions = append(constraint.IndexOptions, option)
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

	indexDef := &ast.IndexDefinition{
		IndexType: &ast.IndexType{}, // Default empty index type
	}

	// Parse index name
	if p.curTok.Type == TokenIdent {
		indexDef.Name = p.parseIdentifier()
	}

	// Parse optional UNIQUE
	if strings.ToUpper(p.curTok.Literal) == "UNIQUE" {
		indexDef.Unique = true
		p.nextToken()
	}

	// Parse optional CLUSTERED/NONCLUSTERED [HASH/COLUMNSTORE]
	if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
		indexDef.IndexType = &ast.IndexType{IndexTypeKind: "Clustered"}
		p.nextToken()
		// Check for HASH or COLUMNSTORE
		if strings.ToUpper(p.curTok.Literal) == "HASH" {
			indexDef.IndexType.IndexTypeKind = "ClusteredHash"
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "COLUMNSTORE" {
			indexDef.IndexType.IndexTypeKind = "ClusteredColumnStore"
			p.nextToken()
		}
	} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
		indexDef.IndexType = &ast.IndexType{IndexTypeKind: "NonClustered"}
		p.nextToken()
		// Check for HASH or COLUMNSTORE
		if strings.ToUpper(p.curTok.Literal) == "HASH" {
			indexDef.IndexType.IndexTypeKind = "NonClusteredHash"
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "COLUMNSTORE" {
			indexDef.IndexType.IndexTypeKind = "NonClusteredColumnStore"
			p.nextToken()
		}
	} else if strings.ToUpper(p.curTok.Literal) == "COLUMNSTORE" {
		// Implicit NONCLUSTERED COLUMNSTORE
		indexDef.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredColumnStore"}
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "HASH" {
		// Implicit NONCLUSTERED HASH
		indexDef.IndexType = &ast.IndexType{IndexTypeKind: "NonClusteredHash"}
		p.nextToken()
	}

	// Parse column list
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			colIdent := p.parseIdentifier()
			col := &ast.ColumnWithSortOrder{
				Column: &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Count: 1,
						Identifiers: []*ast.Identifier{colIdent},
					},
				},
				SortOrder: ast.SortOrderNotSpecified,
			}

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
				// Check for graph pseudo columns like $node_id, $edge_id, $from_id, $to_id
				upperLit := strings.ToUpper(p.curTok.Literal)
				if upperLit == "$NODE_ID" {
					includeCol := &ast.ColumnReferenceExpression{
						ColumnType: "PseudoColumnGraphNodeId",
					}
					indexDef.IncludeColumns = append(indexDef.IncludeColumns, includeCol)
					p.nextToken()
				} else if upperLit == "$EDGE_ID" {
					includeCol := &ast.ColumnReferenceExpression{
						ColumnType: "PseudoColumnGraphEdgeId",
					}
					indexDef.IncludeColumns = append(indexDef.IncludeColumns, includeCol)
					p.nextToken()
				} else if upperLit == "$FROM_ID" {
					includeCol := &ast.ColumnReferenceExpression{
						ColumnType: "PseudoColumnFromNodeId",
					}
					indexDef.IncludeColumns = append(indexDef.IncludeColumns, includeCol)
					p.nextToken()
				} else if upperLit == "$TO_ID" {
					includeCol := &ast.ColumnReferenceExpression{
						ColumnType: "PseudoColumnToNodeId",
					}
					indexDef.IncludeColumns = append(indexDef.IncludeColumns, includeCol)
					p.nextToken()
				} else {
					colIdent := p.parseIdentifier()
					includeCol := &ast.ColumnReferenceExpression{
						ColumnType: "Regular",
						MultiPartIdentifier: &ast.MultiPartIdentifier{
							Count:       1,
							Identifiers: []*ast.Identifier{colIdent},
						},
					}
					indexDef.IncludeColumns = append(indexDef.IncludeColumns, includeCol)
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
		}
	}

	// Parse optional WHERE clause for filtered indexes
	if strings.ToUpper(p.curTok.Literal) == "WHERE" {
		p.nextToken() // consume WHERE
		filterPredicate, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		indexDef.FilterPredicate = filterPredicate
	}

	// Parse optional WITH options
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
				// Parse option value
				switch optionName {
				case "BUCKET_COUNT":
					opt := &ast.IndexExpressionOption{
						OptionKind: "BucketCount",
						Expression: &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						},
					}
					indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
					p.nextToken()
				case "FILLFACTOR":
					opt := &ast.IndexExpressionOption{
						OptionKind: "FillFactor",
						Expression: &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						},
					}
					indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
					p.nextToken()
				case "PAD_INDEX", "STATISTICS_NORECOMPUTE", "ALLOW_ROW_LOCKS", "ALLOW_PAGE_LOCKS", "OPTIMIZE_FOR_SEQUENTIAL_KEY":
					optionKindMap := map[string]string{
						"PAD_INDEX":                   "PadIndex",
						"STATISTICS_NORECOMPUTE":      "StatisticsNoRecompute",
						"ALLOW_ROW_LOCKS":             "AllowRowLocks",
						"ALLOW_PAGE_LOCKS":            "AllowPageLocks",
						"OPTIMIZE_FOR_SEQUENTIAL_KEY": "OptimizeForSequentialKey",
					}
					state := strings.ToUpper(p.curTok.Literal)
					optState := "Off"
					if state == "ON" {
						optState = "On"
					}
					opt := &ast.IndexStateOption{
						OptionKind:  optionKindMap[optionName],
						OptionState: optState,
					}
					indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
					p.nextToken()
				case "IGNORE_DUP_KEY":
					state := strings.ToUpper(p.curTok.Literal)
					optState := "Off"
					if state == "ON" {
						optState = "On"
					}
					opt := &ast.IgnoreDupKeyIndexOption{
						OptionKind:  "IgnoreDupKey",
						OptionState: optState,
					}
					indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
					p.nextToken()
				case "DATA_COMPRESSION":
					compressionLevel := "None"
					levelStr := strings.ToUpper(p.curTok.Literal)
					switch levelStr {
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
					p.nextToken() // consume the compression level
					opt := &ast.DataCompressionOption{
						OptionKind:       "DataCompression",
						CompressionLevel: compressionLevel,
					}
					// Check for optional ON PARTITIONS(range)
					if p.curTok.Type == TokenOn {
						p.nextToken() // consume ON
						if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
							p.nextToken() // consume PARTITIONS
							if p.curTok.Type == TokenLParen {
								p.nextToken() // consume (
								for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
									pr := &ast.CompressionPartitionRange{}
									// Parse From value
									from := &ast.IntegerLiteral{
										LiteralType: "Integer",
										Value:       p.curTok.Literal,
									}
									pr.From = from
									p.nextToken()
									// Check for TO
									if strings.ToUpper(p.curTok.Literal) == "TO" {
										p.nextToken() // consume TO
										to := &ast.IntegerLiteral{
											LiteralType: "Integer",
											Value:       p.curTok.Literal,
										}
										pr.To = to
										p.nextToken()
									}
									opt.PartitionRanges = append(opt.PartitionRanges, pr)
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
					}
					indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
				case "COMPRESSION_DELAY":
					opt := &ast.CompressionDelayIndexOption{
						OptionKind: "CompressionDelay",
						Expression: &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						},
						TimeUnit: "Unitless",
					}
					p.nextToken() // consume the number
					// Check for optional MINUTE/MINUTES time unit
					upperLit := strings.ToUpper(p.curTok.Literal)
					if upperLit == "MINUTE" {
						opt.TimeUnit = "Minute"
						p.nextToken()
					} else if upperLit == "MINUTES" {
						opt.TimeUnit = "Minutes"
						p.nextToken()
					}
					indexDef.IndexOptions = append(indexDef.IndexOptions, opt)
				default:
					// Skip unknown options
					p.nextToken()
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
		}
	}

	// Parse optional ON filegroup clause
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fgName := p.curTok.Literal
		fg := &ast.FileGroupOrPartitionScheme{
			Name: &ast.IdentifierOrValueExpression{
				Value:      fgName,
				Identifier: p.parseIdentifier(),
			},
		}
		indexDef.OnFileGroupOrPartitionScheme = fg
	}

	// Parse optional FILESTREAM_ON clause
	if strings.ToUpper(p.curTok.Literal) == "FILESTREAM_ON" {
		p.nextToken() // consume FILESTREAM_ON
		fsName := p.curTok.Literal
		indexDef.FileStreamOn = &ast.IdentifierOrValueExpression{
			Value:      fsName,
			Identifier: p.parseIdentifier(),
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

	// Handle NATIONAL prefix (NATIONAL CHAR, NATIONAL TEXT, etc.)
	isNational := false
	if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "NATIONAL" {
		isNational = true
		p.nextToken() // consume NATIONAL
	}

	if p.curTok.Type != TokenIdent {
		return nil, fmt.Errorf("expected data type, got %s", p.curTok.Literal)
	}

	var typeName string
	var quoteType string
	literal := p.curTok.Literal

	// Check if this is a bracketed or quoted identifier
	if len(literal) >= 2 && literal[0] == '[' && literal[len(literal)-1] == ']' {
		typeName = literal[1 : len(literal)-1]
		quoteType = "SquareBracket"
	} else if len(literal) >= 2 && literal[0] == '"' && literal[len(literal)-1] == '"' {
		typeName = literal[1 : len(literal)-1]
		quoteType = "DoubleQuote"
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

	// Check for XML type - returns XmlDataTypeReference
	if strings.ToUpper(typeName) == "XML" {
		xmlRef := &ast.XmlDataTypeReference{
			XmlDataTypeOption: "None",
			Name:              baseName,
		}
		// Check for schema collection: XML(CONTENT|DOCUMENT schema_collection)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (

			// Check for CONTENT or DOCUMENT keyword
			upper := strings.ToUpper(p.curTok.Literal)
			if upper == "CONTENT" {
				xmlRef.XmlDataTypeOption = "Content"
				p.nextToken()
			} else if upper == "DOCUMENT" {
				xmlRef.XmlDataTypeOption = "Document"
				p.nextToken()
			}

			// Parse the schema collection name
			schemaName, err := p.parseSchemaObjectName()
			if err != nil {
				return nil, err
			}
			xmlRef.XmlSchemaCollection = schemaName

			if p.curTok.Type == TokenRParen {
				p.nextToken()
			}
		}

		return xmlRef, nil
	}

	// Check if this is a known SQL data type
	sqlOption, isKnownType := getSqlDataTypeOption(typeName)

	// Check for multi-word types: CHAR VARYING -> VarChar, DOUBLE PRECISION -> Float
	// Also handle BINARY VARYING -> VarBinary, CHARACTER VARYING -> VarChar
	// And NCHAR VARYING -> NVarChar, NCHARACTER VARYING -> NVarChar
	if upper := strings.ToUpper(typeName); upper == "CHAR" || upper == "CHARACTER" || upper == "DOUBLE" || upper == "BINARY" || upper == "NCHAR" || upper == "NCHARACTER" {
		nextUpper := strings.ToUpper(p.curTok.Literal)
		if (upper == "CHAR" || upper == "CHARACTER") && nextUpper == "VARYING" {
			sqlOption = "VarChar"
			isKnownType = true
			p.nextToken() // consume VARYING
		} else if upper == "BINARY" && nextUpper == "VARYING" {
			sqlOption = "VarBinary"
			isKnownType = true
			p.nextToken() // consume VARYING
		} else if (upper == "NCHAR" || upper == "NCHARACTER") && nextUpper == "VARYING" {
			sqlOption = "NVarChar"
			isKnownType = true
			p.nextToken() // consume VARYING
		} else if upper == "DOUBLE" && nextUpper == "PRECISION" {
			baseName.BaseIdentifier.Value = "FLOAT" // Use FLOAT for output
			sqlOption = "Float"
			isKnownType = true
			p.nextToken() // consume PRECISION
		}
	}

	// Apply NATIONAL prefix to convert to national types
	if isNational && isKnownType {
		switch sqlOption {
		case "Text":
			sqlOption = "NText"
		case "Char":
			sqlOption = "NChar"
		case "VarChar":
			sqlOption = "NVarChar"
		}
	}

	if !isKnownType {
		// Check for multi-part type name (e.g., dbo.mytype or sys.text)
		if p.curTok.Type == TokenDot {
			p.nextToken() // consume .
			// Get the next identifier
			nextIdent := p.parseIdentifier()
			// Schema.Type structure
			baseName.SchemaIdentifier = baseId
			baseName.BaseIdentifier = nextIdent
			baseName.Count = 2
			baseName.Identifiers = []*ast.Identifier{baseId, nextIdent}

			// Check for third part: database.schema.type
			if p.curTok.Type == TokenDot {
				p.nextToken() // consume .
				thirdIdent := p.parseIdentifier()
				// Database.Schema.Type structure
				baseName.DatabaseIdentifier = baseId
				baseName.SchemaIdentifier = nextIdent
				baseName.BaseIdentifier = thirdIdent
				baseName.Count = 3
				baseName.Identifiers = []*ast.Identifier{baseId, nextIdent, thirdIdent}
			}

			// Re-check if the base type (after schema) is a known SQL type
			// This handles cases like sys.int, sys.text, etc.
			baseTypeName := baseName.BaseIdentifier.Value
			baseOption, baseIsKnown := getSqlDataTypeOption(baseTypeName)

			// Handle multi-word types with schema prefix: sys.Char varying -> VarChar
			if baseUpper := strings.ToUpper(baseTypeName); baseUpper == "CHAR" || baseUpper == "CHARACTER" || baseUpper == "BINARY" || baseUpper == "NCHAR" || baseUpper == "NCHARACTER" {
				nextUpper := strings.ToUpper(p.curTok.Literal)
				if (baseUpper == "CHAR" || baseUpper == "CHARACTER") && nextUpper == "VARYING" {
					baseOption = "VarChar"
					baseIsKnown = true
					p.nextToken() // consume VARYING
				} else if baseUpper == "BINARY" && nextUpper == "VARYING" {
					baseOption = "VarBinary"
					baseIsKnown = true
					p.nextToken() // consume VARYING
				} else if (baseUpper == "NCHAR" || baseUpper == "NCHARACTER") && nextUpper == "VARYING" {
					baseOption = "NVarChar"
					baseIsKnown = true
					p.nextToken() // consume VARYING
				}
			}

			// Apply NATIONAL prefix for schema-qualified national types
			if isNational && baseIsKnown {
				switch baseOption {
				case "Text":
					baseOption = "NText"
				case "Char":
					baseOption = "NChar"
				case "VarChar":
					baseOption = "NVarChar"
				}
			}

			if baseIsKnown {
				// Special handling for XML type with schema prefix: sys.[xml](CONTENT schema_collection)
				if strings.ToUpper(baseName.BaseIdentifier.Value) == "XML" {
					xmlRef := &ast.XmlDataTypeReference{
						XmlDataTypeOption: "None",
						Name:              baseName,
					}
					// Check for schema collection: XML(CONTENT|DOCUMENT schema_collection)
					if p.curTok.Type == TokenLParen {
						p.nextToken() // consume (

						// Check for CONTENT or DOCUMENT keyword
						upper := strings.ToUpper(p.curTok.Literal)
						if upper == "CONTENT" {
							xmlRef.XmlDataTypeOption = "Content"
							p.nextToken()
						} else if upper == "DOCUMENT" {
							xmlRef.XmlDataTypeOption = "Document"
							p.nextToken()
						}

						// Parse the schema collection name
						schemaName, err := p.parseSchemaObjectName()
						if err != nil {
							return nil, err
						}
						xmlRef.XmlSchemaCollection = schemaName

						if p.curTok.Type == TokenRParen {
							p.nextToken()
						}
					}
					return xmlRef, nil
				}

				// Return SqlDataTypeReference for known types with schema prefix
				dt := &ast.SqlDataTypeReference{
					SqlDataTypeOption: baseOption,
					Name:              baseName,
				}
				// Handle parameters
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
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
						p.nextToken() // consume comma
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}
				return dt, nil
			}
		}

		userRef := &ast.UserDataTypeReference{
			Name: baseName,
		}

		// Check for parameters: mytype(10) or mytype(10, 20) or mytype(max)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				// Special case: MAX keyword
				if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "MAX" {
					userRef.Parameters = append(userRef.Parameters, &ast.MaxLiteral{
						LiteralType: "Max",
						Value:       p.curTok.Literal,
					})
					p.nextToken()
				} else {
					expr, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					userRef.Parameters = append(userRef.Parameters, expr)
				}
				if p.curTok.Type != TokenComma {
					break
				}
				p.nextToken() // consume comma
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}

		return userRef, nil
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
		"DEC":               "Decimal",
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
		"CHARACTER":         "Char",
		"VARCHAR":           "VarChar",
		"TEXT":              "Text",
		"NCHAR":             "NChar",
		"NCHARACTER":        "NChar",
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
	// Note: some options like LANGUAGE are keyword tokens, so we also check for those
	if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLanguage ||
		p.curTok.Type == TokenTransaction {
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

		// Handle SET TRANSACTION ISOLATION LEVEL
		if optionName == "TRANSACTION" {
			return p.parseSetTransactionIsolationLevel()
		}

		// Handle SET TEXTSIZE
		if optionName == "TEXTSIZE" {
			return p.parseSetTextSizeStatement()
		}

		// Handle SET IDENTITY_INSERT
		if optionName == "IDENTITY_INSERT" {
			return p.parseSetIdentityInsertStatement()
		}

		// Handle SET ERRLVL
		if optionName == "ERRLVL" {
			return p.parseSetErrorLevelStatement()
		}

		// Handle SET command statements (FIPS_FLAGGER, LANGUAGE, etc.)
		if p.isSetCommandOption(optionName) {
			return p.parseSetCommandStatement(optionName)
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

	// Statistics option order for sorting
	statisticsOptionOrder := map[string]int{
		"IO":      1,
		"Profile": 2,
		"Time":    3,
		"Xml":     4,
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

	// Sort options according to ScriptDom order
	sort.Slice(options, func(i, j int) bool {
		return statisticsOptionOrder[options[i]] < statisticsOptionOrder[options[j]]
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

// isSetCommandOption returns true if the option is a SET command option
func (p *Parser) isSetCommandOption(optName string) bool {
	switch optName {
	case "FIPS_FLAGGER", "QUERY_GOVERNOR_COST_LIMIT", "LANGUAGE", "DATEFORMAT",
		"DATEFIRST", "DEADLOCK_PRIORITY", "LOCK_TIMEOUT", "CONTEXT_INFO":
		return true
	}
	return false
}

// parseSetCommandStatement parses SET commands like FIPS_FLAGGER, LANGUAGE, etc.
func (p *Parser) parseSetCommandStatement(firstOpt string) (*ast.SetCommandStatement, error) {
	stmt := &ast.SetCommandStatement{}

	// Consume the first option name (already read in parseSetVariableStatement)
	p.nextToken()

	// Parse the first command
	cmd, err := p.parseSetCommand(firstOpt)
	if err != nil {
		return nil, err
	}
	stmt.Commands = append(stmt.Commands, cmd)

	// Parse additional commands separated by comma
	for p.curTok.Type == TokenComma {
		p.nextToken() // consume comma
		optName := strings.ToUpper(p.curTok.Literal)
		p.nextToken() // consume option name
		cmd, err := p.parseSetCommand(optName)
		if err != nil {
			return nil, err
		}
		stmt.Commands = append(stmt.Commands, cmd)
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

// parseSetCommand parses a single SET command
func (p *Parser) parseSetCommand(optName string) (ast.SetCommand, error) {
	switch optName {
	case "FIPS_FLAGGER":
		// Parse OFF, 'ENTRY', 'INTERMEDIATE', 'FULL'
		var level string
		if p.curTok.Type == TokenIdent && strings.ToUpper(p.curTok.Literal) == "OFF" {
			level = "Off"
			p.nextToken()
		} else if p.curTok.Type == TokenString {
			// Strip quotes from the value
			val := strings.Trim(p.curTok.Literal, "'\"")
			switch strings.ToUpper(val) {
			case "ENTRY":
				level = "Entry"
			case "INTERMEDIATE":
				level = "Intermediate"
			case "FULL":
				level = "Full"
			default:
				level = capitalizeFirst(strings.ToLower(val))
			}
			p.nextToken()
		}
		return &ast.SetFipsFlaggerCommand{ComplianceLevel: level}, nil

	case "QUERY_GOVERNOR_COST_LIMIT":
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.GeneralSetCommand{CommandType: "QueryGovernorCostLimit", Parameter: param}, nil

	case "LANGUAGE":
		param, err := p.parseSetCommandParameter()
		if err != nil {
			return nil, err
		}
		return &ast.GeneralSetCommand{CommandType: "Language", Parameter: param}, nil

	case "DATEFORMAT":
		param, err := p.parseSetCommandParameter()
		if err != nil {
			return nil, err
		}
		return &ast.GeneralSetCommand{CommandType: "DateFormat", Parameter: param}, nil

	case "DATEFIRST":
		param, err := p.parseSetCommandParameter()
		if err != nil {
			return nil, err
		}
		return &ast.GeneralSetCommand{CommandType: "DateFirst", Parameter: param}, nil

	case "DEADLOCK_PRIORITY":
		param, err := p.parseSetCommandParameter()
		if err != nil {
			return nil, err
		}
		return &ast.GeneralSetCommand{CommandType: "DeadlockPriority", Parameter: param}, nil

	case "LOCK_TIMEOUT":
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.GeneralSetCommand{CommandType: "LockTimeout", Parameter: param}, nil

	case "CONTEXT_INFO":
		param, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.GeneralSetCommand{CommandType: "ContextInfo", Parameter: param}, nil

	default:
		return nil, fmt.Errorf("unknown SET command: %s", optName)
	}
}

// parseSetCommandParameter parses parameters for SET commands that can be identifier, string or variable
func (p *Parser) parseSetCommandParameter() (ast.ScalarExpression, error) {
	if strings.HasPrefix(p.curTok.Literal, "@") {
		// Variable reference
		v := &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
		return v, nil
	} else if p.curTok.Type == TokenString {
		// String literal - strip quotes from value
		val := strings.Trim(p.curTok.Literal, "'\"")
		lit := &ast.StringLiteral{
			LiteralType:   "String",
			Value:         val,
			IsNational:    false,
			IsLargeObject: false,
		}
		p.nextToken()
		return lit, nil
	} else if p.curTok.Type == TokenIdent {
		// Identifier literal
		lit := &ast.IdentifierLiteral{
			LiteralType: "Identifier",
			QuoteType:   "NotQuoted",
			Value:       p.curTok.Literal,
		}
		p.nextToken()
		return lit, nil
	}
	return p.parseScalarExpression()
}

// parseSetTransactionIsolationLevel parses SET TRANSACTION ISOLATION LEVEL statement
func (p *Parser) parseSetTransactionIsolationLevel() (*ast.SetTransactionIsolationLevelStatement, error) {
	p.nextToken() // consume TRANSACTION

	// Skip ISOLATION LEVEL
	if strings.ToUpper(p.curTok.Literal) == "ISOLATION" {
		p.nextToken()
	}
	if strings.ToUpper(p.curTok.Literal) == "LEVEL" {
		p.nextToken()
	}

	// Parse level
	var level string
	firstWord := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	switch firstWord {
	case "READ":
		secondWord := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		if secondWord == "COMMITTED" {
			level = "ReadCommitted"
		} else if secondWord == "UNCOMMITTED" {
			level = "ReadUncommitted"
		}
	case "REPEATABLE":
		if strings.ToUpper(p.curTok.Literal) == "READ" {
			p.nextToken()
		}
		level = "RepeatableRead"
	case "SERIALIZABLE":
		level = "Serializable"
	case "SNAPSHOT":
		level = "Snapshot"
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.SetTransactionIsolationLevelStatement{Level: level}, nil
}

// parseSetTextSizeStatement parses SET TEXTSIZE statement
func (p *Parser) parseSetTextSizeStatement() (*ast.SetTextSizeStatement, error) {
	p.nextToken() // consume TEXTSIZE

	textSize, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.SetTextSizeStatement{TextSize: textSize}, nil
}

// parseSetIdentityInsertStatement parses SET IDENTITY_INSERT table ON/OFF
func (p *Parser) parseSetIdentityInsertStatement() (*ast.SetIdentityInsertStatement, error) {
	p.nextToken() // consume IDENTITY_INSERT

	// Parse table name
	tableName, _ := p.parseSchemaObjectName()

	// Parse ON/OFF
	isOn := false
	if p.curTok.Type == TokenOn || strings.ToUpper(p.curTok.Literal) == "ON" {
		isOn = true
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
		isOn = false
		p.nextToken()
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.SetIdentityInsertStatement{Table: tableName, IsOn: isOn}, nil
}

// parseSetErrorLevelStatement parses SET ERRLVL statement
func (p *Parser) parseSetErrorLevelStatement() (*ast.SetErrorLevelStatement, error) {
	p.nextToken() // consume ERRLVL

	level, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return &ast.SetErrorLevelStatement{Level: level}, nil
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
	case TokenDialog:
		return p.parseBeginDialogStatement()
	case TokenConversation:
		return p.parseBeginConversationTimerStatement()
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
				// Parse the isolation level identifier - may be multi-word like "READ COMMITTED"
				levelValue := strings.ToUpper(p.curTok.Literal)
				p.nextToken()
				// Check for two-word isolation levels
				nextWord := strings.ToUpper(p.curTok.Literal)
				if (levelValue == "READ" && (nextWord == "COMMITTED" || nextWord == "UNCOMMITTED")) ||
					(levelValue == "REPEATABLE" && nextWord == "READ") {
					levelValue = levelValue + " " + nextWord
					p.nextToken()
				}
				opt := &ast.IdentifierAtomicBlockOption{
					OptionKind: "IsolationLevel",
					Value: &ast.Identifier{
						Value:     levelValue,
						QuoteType: "NotQuoted",
					},
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
			case "DATEFIRST":
				// Parse as integer literal
				intLit := &ast.IntegerLiteral{
					LiteralType: "Integer",
					Value:       p.curTok.Literal,
				}
				p.nextToken()
				opt := &ast.LiteralAtomicBlockOption{
					OptionKind: "DateFirst",
					Value:      intLit,
				}
				stmt.Options = append(stmt.Options, opt)
			case "DATEFORMAT":
				// Parse as string literal
				value := p.curTok.Literal
				// Strip quotes if present
				if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
					value = value[1 : len(value)-1]
				}
				strLit := &ast.StringLiteral{
					LiteralType:   "String",
					Value:         value,
					IsNational:    false,
					IsLargeObject: false,
				}
				p.nextToken()
				opt := &ast.LiteralAtomicBlockOption{
					OptionKind: "DateFormat",
					Value:      strLit,
				}
				stmt.Options = append(stmt.Options, opt)
			case "DELAYED_DURABILITY":
				// Parse ON/OFF as OnOffAtomicBlockOption
				stateUpper := strings.ToUpper(p.curTok.Literal)
				optState := "Off"
				if stateUpper == "ON" {
					optState = "On"
				}
				p.nextToken()
				opt := &ast.OnOffAtomicBlockOption{
					OptionKind:  "DelayedDurability",
					OptionState: optState,
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
	for p.curTok.Type != TokenEOF {
		// Skip semicolons
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
			continue
		}
		// Check for END TRY (not END CONVERSATION)
		if p.curTok.Type == TokenEnd {
			if p.peekTok.Type == TokenConversation {
				// It's END CONVERSATION, parse it
				endConvStmt, err := p.parseEndConversationStatement()
				if err != nil {
					return nil, err
				}
				if endConvStmt != nil {
					stmt.TryStatements.Statements = append(stmt.TryStatements.Statements, endConvStmt)
				}
				continue
			}
			// It's END TRY, break
			break
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
	for p.curTok.Type != TokenEOF {
		// Skip semicolons
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
			continue
		}
		// Check for END CATCH (not END CONVERSATION)
		if p.curTok.Type == TokenEnd {
			if p.peekTok.Type == TokenConversation {
				// It's END CONVERSATION, parse it
				endConvStmt, err := p.parseEndConversationStatement()
				if err != nil {
					return nil, err
				}
				if endConvStmt != nil {
					stmt.CatchStatements.Statements = append(stmt.CatchStatements.Statements, endConvStmt)
				}
				continue
			}
			// It's END CATCH, break
			break
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

	// Parse statements until END (but not END CONVERSATION)
	for p.curTok.Type != TokenEOF {
		// Check for END (not END CONVERSATION)
		if p.curTok.Type == TokenEnd {
			if p.peekTok.Type == TokenConversation {
				// It's END CONVERSATION, parse it
				endConvStmt, err := p.parseEndConversationStatement()
				if err != nil {
					return nil, err
				}
				if endConvStmt != nil {
					stmt.StatementList.Statements = append(stmt.StatementList.Statements, endConvStmt)
				}
				continue
			}
			// It's END (block terminator), break
			break
		}
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

func (p *Parser) parseBeginDialogStatement() (*ast.BeginDialogStatement, error) {
	p.nextToken() // consume DIALOG

	stmt := &ast.BeginDialogStatement{}

	// Check for optional CONVERSATION keyword
	if p.curTok.Type == TokenConversation {
		stmt.IsConversation = true
		p.nextToken() // consume CONVERSATION
	}

	// Parse dialog handle (variable reference)
	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.Handle = &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected variable for dialog handle")
	}

	// Parse FROM SERVICE
	if p.curTok.Type != TokenFrom {
		return nil, fmt.Errorf("expected FROM after dialog handle")
	}
	p.nextToken() // consume FROM

	if strings.ToUpper(p.curTok.Literal) != "SERVICE" {
		return nil, fmt.Errorf("expected SERVICE after FROM")
	}
	p.nextToken() // consume SERVICE

	// Parse initiator service name (identifier)
	id := p.parseIdentifier()
	stmt.InitiatorServiceName = &ast.IdentifierOrValueExpression{
		Value:      id.Value,
		Identifier: id,
	}

	// Parse TO SERVICE
	if p.curTok.Type != TokenTo {
		return nil, fmt.Errorf("expected TO after initiator service name")
	}
	p.nextToken() // consume TO

	if strings.ToUpper(p.curTok.Literal) != "SERVICE" {
		return nil, fmt.Errorf("expected SERVICE after TO")
	}
	p.nextToken() // consume SERVICE

	// Parse target service name (string literal or variable)
	if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
		strLit, err := p.parseStringLiteral()
		if err != nil {
			return nil, err
		}
		stmt.TargetServiceName = strLit
	} else if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.TargetServiceName = &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected string literal or variable for target service name")
	}

	// Check for optional instance spec (after comma)
	if p.curTok.Type == TokenComma {
		p.nextToken() // consume comma
		if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
			strLit, err := p.parseStringLiteral()
			if err != nil {
				return nil, err
			}
			stmt.InstanceSpec = strLit
		} else if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
			stmt.InstanceSpec = &ast.VariableReference{Name: p.curTok.Literal}
			p.nextToken()
		}
	}

	// Parse optional ON CONTRACT
	if p.curTok.Type == TokenOn && strings.ToUpper(p.peekTok.Literal) == "CONTRACT" {
		p.nextToken() // consume ON
		p.nextToken() // consume CONTRACT
		id := p.parseIdentifier()
		stmt.ContractName = &ast.IdentifierOrValueExpression{
			Value:      id.Value,
			Identifier: id,
		}
	}

	// Parse optional WITH options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		for {
			optName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume option name

			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			switch optName {
			case "RELATED_CONVERSATION":
				if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
					stmt.Options = append(stmt.Options, &ast.ScalarExpressionDialogOption{
						Value:      &ast.VariableReference{Name: p.curTok.Literal},
						OptionKind: "RelatedConversation",
					})
					p.nextToken()
				}
			case "RELATED_CONVERSATION_GROUP":
				if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
					stmt.Options = append(stmt.Options, &ast.ScalarExpressionDialogOption{
						Value:      &ast.VariableReference{Name: p.curTok.Literal},
						OptionKind: "RelatedConversationGroup",
					})
					p.nextToken()
				}
			case "ENCRYPTION":
				optState := strings.ToUpper(p.curTok.Literal)
				if optState == "ON" {
					stmt.Options = append(stmt.Options, &ast.OnOffDialogOption{
						OptionState: "On",
						OptionKind:  "Encryption",
					})
				} else if optState == "OFF" {
					stmt.Options = append(stmt.Options, &ast.OnOffDialogOption{
						OptionState: "Off",
						OptionKind:  "Encryption",
					})
				}
				p.nextToken()
			case "LIFETIME":
				if p.curTok.Type == TokenNumber {
					stmt.Options = append(stmt.Options, &ast.ScalarExpressionDialogOption{
						Value: &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						},
						OptionKind: "Lifetime",
					})
					p.nextToken()
				}
			}

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

func (p *Parser) parseBeginConversationTimerStatement() (*ast.BeginConversationTimerStatement, error) {
	p.nextToken() // consume CONVERSATION

	// Expect TIMER
	if strings.ToUpper(p.curTok.Literal) != "TIMER" {
		return nil, fmt.Errorf("expected TIMER after BEGIN CONVERSATION")
	}
	p.nextToken() // consume TIMER

	stmt := &ast.BeginConversationTimerStatement{}

	// Parse handle in parentheses
	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after TIMER")
	}
	p.nextToken() // consume (

	if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
		stmt.Handle = &ast.VariableReference{Name: p.curTok.Literal}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected variable for conversation handle")
	}

	if p.curTok.Type != TokenRParen {
		return nil, fmt.Errorf("expected ) after handle")
	}
	p.nextToken() // consume )

	// Parse TIMEOUT = value
	if strings.ToUpper(p.curTok.Literal) != "TIMEOUT" {
		return nil, fmt.Errorf("expected TIMEOUT")
	}
	p.nextToken() // consume TIMEOUT

	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
	}

	if p.curTok.Type == TokenNumber {
		stmt.Timeout = &ast.IntegerLiteral{
			LiteralType: "Integer",
			Value:       p.curTok.Literal,
		}
		p.nextToken()
	} else {
		return nil, fmt.Errorf("expected integer for timeout value")
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

	// Parse statements until END (but not END CONVERSATION)
	for p.curTok.Type != TokenEOF {
		// Check for END (not END CONVERSATION)
		if p.curTok.Type == TokenEnd {
			if p.peekTok.Type == TokenConversation {
				// It's END CONVERSATION, parse it
				endConvStmt, err := p.parseEndConversationStatement()
				if err != nil {
					return nil, err
				}
				if endConvStmt != nil {
					stmt.StatementList.Statements = append(stmt.StatementList.Statements, endConvStmt)
				}
				continue
			}
			// It's END (block terminator), break
			break
		}
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
	case TokenOr:
		// Handle CREATE OR ALTER
		p.nextToken() // consume OR
		if strings.ToUpper(p.curTok.Literal) != "ALTER" {
			return nil, fmt.Errorf("expected ALTER after CREATE OR, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume ALTER
		switch p.curTok.Type {
		case TokenFunction:
			return p.parseCreateOrAlterFunctionStatement()
		case TokenProcedure:
			return p.parseCreateOrAlterProcedureStatement()
		case TokenView:
			return p.parseCreateOrAlterViewStatement()
		case TokenTrigger:
			return p.parseCreateOrAlterTriggerStatement()
		default:
			// Lenient: skip unknown CREATE OR ALTER statements
			p.skipToEndOfStatement()
			return &ast.CreateProcedureStatement{}, nil
		}
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
		case "CLUSTERED":
			// Check if next token is COLUMNSTORE or INDEX
			if p.peekTok.Type == TokenIdent && strings.ToUpper(p.peekTok.Literal) == "COLUMNSTORE" {
				return p.parseCreateColumnStoreIndexStatement()
			}
			// Otherwise it's CLUSTERED INDEX -> use parseCreateIndexStatement
			return p.parseCreateIndexStatement()
		case "NONCLUSTERED":
			// Check if next token is COLUMNSTORE or INDEX
			if p.peekTok.Type == TokenIdent && strings.ToUpper(p.peekTok.Literal) == "COLUMNSTORE" {
				return p.parseCreateColumnStoreIndexStatement()
			}
			// Otherwise it's NONCLUSTERED INDEX -> use parseCreateIndexStatement
			return p.parseCreateIndexStatement()
		case "COLUMNSTORE":
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
		case "SELECTIVE":
			return p.parseCreateSelectiveXmlIndexStatement()
		case "COLUMN":
			return p.parseCreateColumnMasterKeyStatement()
		case "CRYPTOGRAPHIC":
			return p.parseCreateCryptographicProviderStatement()
		case "BROKER":
			return p.parseCreateBrokerPriorityStatement()
		case "FEDERATION":
			return p.parseCreateFederationStatement()
		case "WORKLOAD":
			// Check if it's CLASSIFIER or GROUP
			nextWord := strings.ToUpper(p.peekTok.Literal)
			if nextWord == "CLASSIFIER" {
				return p.parseCreateWorkloadClassifierStatement()
			}
			return p.parseCreateWorkloadGroupStatement()
		case "RESOURCE":
			// Check if it's RESOURCE POOL or RESOURCE GOVERNOR
			p.nextToken() // consume RESOURCE
			if strings.ToUpper(p.curTok.Literal) == "POOL" {
				return p.parseCreateResourcePoolStatement()
			}
			// RESOURCE GOVERNOR not supported for CREATE
			p.skipToEndOfStatement()
			return &ast.CreateProcedureStatement{}, nil
		case "SEQUENCE":
			return p.parseCreateSequenceStatement()
		case "SPATIAL":
			return p.parseCreateSpatialIndexStatement()
		case "MATERIALIZED":
			return p.parseCreateMaterializedViewStatement()
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
		case "AVAILABILITY":
			return p.parseCreateAvailabilityGroupStatement()
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

func (p *Parser) parseCreateAvailabilityGroupStatement() (*ast.CreateAvailabilityGroupStatement, error) {
	// Consume AVAILABILITY
	p.nextToken()

	// Expect GROUP
	if strings.ToUpper(p.curTok.Literal) != "GROUP" {
		return nil, fmt.Errorf("expected GROUP after AVAILABILITY, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.CreateAvailabilityGroupStatement{}

	// Parse group name
	stmt.Name = p.parseIdentifier()

	// Parse WITH clause for group options
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
				case "REQUIRED_COPIES_TO_COMMIT":
					val, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					stmt.Options = append(stmt.Options, &ast.LiteralAvailabilityGroupOption{
						OptionKind: "RequiredCopiesToCommit",
						Value:      val,
					})
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

	// Parse FOR DATABASE clause
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "DATABASE" {
			p.nextToken() // consume DATABASE
			// Parse comma-separated database names
			for {
				stmt.Databases = append(stmt.Databases, p.parseIdentifier())
				if p.curTok.Type != TokenComma {
					break
				}
				p.nextToken() // consume comma
			}
		}
	}

	// Parse REPLICA ON clause
	if strings.ToUpper(p.curTok.Literal) == "REPLICA" {
		p.nextToken() // consume REPLICA
		if strings.ToUpper(p.curTok.Literal) == "ON" {
			p.nextToken() // consume ON
		}

		// Parse comma-separated replica definitions
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
							val, err := p.parseScalarExpression()
							if err != nil {
								return nil, err
							}
							replica.Options = append(replica.Options, &ast.LiteralReplicaOption{
								OptionKind: "EndpointUrl",
								Value:      val,
							})
						case "SESSION_TIMEOUT":
							val, err := p.parseScalarExpression()
							if err != nil {
								return nil, err
							}
							replica.Options = append(replica.Options, &ast.LiteralReplicaOption{
								OptionKind: "SessionTimeout",
								Value:      val,
							})
						case "APPLY_DELAY":
							val, err := p.parseScalarExpression()
							if err != nil {
								return nil, err
							}
							replica.Options = append(replica.Options, &ast.LiteralReplicaOption{
								OptionKind: "ApplyDelay",
								Value:      val,
							})
						case "PRIMARY_ROLE":
							// Parse (ALLOW_CONNECTIONS = ...)
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
							// Parse (ALLOW_CONNECTIONS = ...)
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

			stmt.Replicas = append(stmt.Replicas, replica)

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

func (p *Parser) parseCreateColumnMasterKeyStatement() (*ast.CreateColumnMasterKeyStatement, error) {
	// CREATE COLUMN MASTER KEY name WITH (options)
	// Already consumed CREATE COLUMN, now need to consume MASTER KEY
	p.nextToken() // consume COLUMN

	if strings.ToUpper(p.curTok.Literal) != "MASTER" {
		return nil, fmt.Errorf("expected MASTER after COLUMN, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume MASTER

	if strings.ToUpper(p.curTok.Literal) != "KEY" {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume KEY

	stmt := &ast.CreateColumnMasterKeyStatement{}

	// Parse key name
	stmt.Name = p.parseIdentifier()

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		if p.curTok.Type != TokenLParen {
			return nil, fmt.Errorf("expected ( after WITH, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume (

		// Parse parameters
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			paramName := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume parameter name

			switch paramName {
			case "KEY_STORE_PROVIDER_NAME":
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				value, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.Parameters = append(stmt.Parameters, &ast.ColumnMasterKeyStoreProviderNameParameter{
					Name:          value,
					ParameterKind: "KeyStoreProviderName",
				})
			case "KEY_PATH":
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				value, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.Parameters = append(stmt.Parameters, &ast.ColumnMasterKeyPathParameter{
					Path:          value,
					ParameterKind: "KeyPath",
				})
			case "ENCLAVE_COMPUTATIONS":
				// ENCLAVE_COMPUTATIONS ( SIGNATURE = value )
				if p.curTok.Type != TokenLParen {
					return nil, fmt.Errorf("expected ( after ENCLAVE_COMPUTATIONS, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume (

				// Parse SIGNATURE = value
				if strings.ToUpper(p.curTok.Literal) == "SIGNATURE" {
					p.nextToken() // consume SIGNATURE
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					value, err := p.parseScalarExpression()
					if err != nil {
						return nil, err
					}
					stmt.Parameters = append(stmt.Parameters, &ast.ColumnMasterKeyEnclaveComputationsParameter{
						Signature:     value,
						ParameterKind: "Signature",
					})
				}

				// Consume closing )
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			default:
				// Skip unknown parameter
				p.nextToken()
			}

			// Skip comma if present
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
		}

		// Consume closing )
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	// Skip any remaining tokens
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

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

func (p *Parser) parseCreateServerAuditStatement() (ast.Statement, error) {
	// AUDIT keyword should be current token, consume it
	p.nextToken()

	// Check if this is CREATE SERVER AUDIT SPECIFICATION
	if strings.ToUpper(p.curTok.Literal) == "SPECIFICATION" {
		return p.parseCreateServerAuditSpecificationStatement()
	}

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

func (p *Parser) parseCreateServerAuditSpecificationStatement() (*ast.CreateServerAuditSpecificationStatement, error) {
	// SPECIFICATION keyword should be current token, consume it
	p.nextToken()

	stmt := &ast.CreateServerAuditSpecificationStatement{
		AuditState: "NotSet",
	}

	// Parse specification name
	stmt.SpecificationName = p.parseIdentifier()

	// Parse FOR SERVER AUDIT audit_name
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken() // consume SERVER
		}
		if strings.ToUpper(p.curTok.Literal) == "AUDIT" {
			p.nextToken() // consume AUDIT
		}
		stmt.AuditName = p.parseIdentifier()
	}

	// Parse ADD/DROP parts
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "ADD" || upperLit == "DROP" {
			part := &ast.AuditSpecificationPart{
				IsDrop: upperLit == "DROP",
			}
			p.nextToken() // consume ADD/DROP
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Parse audit action group reference
				groupName := p.curTok.Literal
				part.Details = &ast.AuditActionGroupReference{
					Group: convertAuditGroupName(groupName),
				}
				p.nextToken() // consume group name
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			stmt.Parts = append(stmt.Parts, part)
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
				continue
			}
		}
		break
	}

	// Parse WITH (STATE = ON/OFF)
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if strings.ToUpper(p.curTok.Literal) == "STATE" {
				p.nextToken() // consume STATE
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				if strings.ToUpper(p.curTok.Literal) == "ON" {
					stmt.AuditState = "On"
				} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
					stmt.AuditState = "Off"
				}
				p.nextToken() // consume ON/OFF
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	return stmt, nil
}

func (p *Parser) parseAlterServerAuditSpecificationStatement() (*ast.AlterServerAuditSpecificationStatement, error) {
	// SPECIFICATION keyword should be current token, consume it
	p.nextToken()

	stmt := &ast.AlterServerAuditSpecificationStatement{
		AuditState: "NotSet",
	}

	// Parse specification name
	stmt.SpecificationName = p.parseIdentifier()

	// Parse FOR SERVER AUDIT audit_name
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken() // consume SERVER
		}
		if strings.ToUpper(p.curTok.Literal) == "AUDIT" {
			p.nextToken() // consume AUDIT
		}
		stmt.AuditName = p.parseIdentifier()
	}

	// Parse ADD/DROP parts
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "ADD" || upperLit == "DROP" {
			part := &ast.AuditSpecificationPart{
				IsDrop: upperLit == "DROP",
			}
			p.nextToken() // consume ADD/DROP
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Parse audit action group reference
				groupName := p.curTok.Literal
				part.Details = &ast.AuditActionGroupReference{
					Group: convertAuditGroupName(groupName),
				}
				p.nextToken() // consume group name
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			stmt.Parts = append(stmt.Parts, part)
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
				continue
			}
		}
		break
	}

	// Parse WITH (STATE = ON/OFF)
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if strings.ToUpper(p.curTok.Literal) == "STATE" {
				p.nextToken() // consume STATE
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				if strings.ToUpper(p.curTok.Literal) == "ON" {
					stmt.AuditState = "On"
				} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
					stmt.AuditState = "Off"
				}
				p.nextToken() // consume ON/OFF
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	return stmt, nil
}

func (p *Parser) parseCreateDatabaseAuditSpecificationStatement() (*ast.CreateDatabaseAuditSpecificationStatement, error) {
	stmt := &ast.CreateDatabaseAuditSpecificationStatement{
		AuditState: "NotSet",
	}

	// Parse specification name
	stmt.SpecificationName = p.parseIdentifier()

	// Parse FOR SERVER AUDIT audit_name
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken() // consume SERVER
		}
		if strings.ToUpper(p.curTok.Literal) == "AUDIT" {
			p.nextToken() // consume AUDIT
		}
		stmt.AuditName = p.parseIdentifier()
	}

	// Parse ADD/DROP parts
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "ADD" || upperLit == "DROP" {
			part, err := p.parseAuditSpecificationPart(upperLit == "DROP")
			if err != nil {
				return nil, err
			}
			stmt.Parts = append(stmt.Parts, part)
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
				continue
			}
		}
		break
	}

	// Parse WITH (STATE = ON/OFF)
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if strings.ToUpper(p.curTok.Literal) == "STATE" {
				p.nextToken() // consume STATE
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				if strings.ToUpper(p.curTok.Literal) == "ON" {
					stmt.AuditState = "On"
				} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
					stmt.AuditState = "Off"
				}
				p.nextToken() // consume ON/OFF
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	return stmt, nil
}

func (p *Parser) parseAlterDatabaseAuditSpecificationStatement() (*ast.AlterDatabaseAuditSpecificationStatement, error) {
	stmt := &ast.AlterDatabaseAuditSpecificationStatement{
		AuditState: "NotSet",
	}

	// Parse specification name
	stmt.SpecificationName = p.parseIdentifier()

	// Parse FOR SERVER AUDIT audit_name (optional in ALTER)
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "SERVER" {
			p.nextToken() // consume SERVER
		}
		if strings.ToUpper(p.curTok.Literal) == "AUDIT" {
			p.nextToken() // consume AUDIT
		}
		stmt.AuditName = p.parseIdentifier()
	}

	// Parse ADD/DROP parts
	for {
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "ADD" || upperLit == "DROP" {
			part, err := p.parseAuditSpecificationPart(upperLit == "DROP")
			if err != nil {
				return nil, err
			}
			stmt.Parts = append(stmt.Parts, part)
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume ,
				continue
			}
		}
		break
	}

	// Parse WITH (STATE = ON/OFF)
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			if strings.ToUpper(p.curTok.Literal) == "STATE" {
				p.nextToken() // consume STATE
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				if strings.ToUpper(p.curTok.Literal) == "ON" {
					stmt.AuditState = "On"
				} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
					stmt.AuditState = "Off"
				}
				p.nextToken() // consume ON/OFF
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	return stmt, nil
}

// convertAuditGroupName converts an audit group name to the expected format
func convertAuditGroupName(name string) string {
	// Map of audit group names to their expected format
	groupMap := map[string]string{
		"SUCCESSFUL_DATABASE_AUTHENTICATION_GROUP": "SuccessfulDatabaseAuthenticationGroup",
		"FAILED_DATABASE_AUTHENTICATION_GROUP":     "FailedDatabaseAuthenticationGroup",
		"DATABASE_LOGOUT_GROUP":                    "DatabaseLogoutGroup",
		"USER_CHANGE_PASSWORD_GROUP":               "UserChangePasswordGroup",
		"USER_DEFINED_AUDIT_GROUP":                 "UserDefinedAuditGroup",
		"DATABASE_PERMISSION_CHANGE_GROUP":         "DatabasePermissionChange",
		"SCHEMA_OBJECT_PERMISSION_CHANGE_GROUP":    "SchemaObjectPermissionChange",
		"DATABASE_ROLE_MEMBER_CHANGE_GROUP":        "DatabaseRoleMemberChange",
		"APPLICATION_ROLE_CHANGE_PASSWORD_GROUP":   "ApplicationRoleChangePassword",
		"SCHEMA_OBJECT_ACCESS_GROUP":               "SchemaObjectAccess",
		"BACKUP_RESTORE_GROUP":                     "BackupRestore",
		"DBCC_GROUP":                               "Dbcc",
		"AUDIT_CHANGE_GROUP":                       "AuditChange",
		"DATABASE_CHANGE_GROUP":                    "DatabaseChange",
		"DATABASE_OBJECT_CHANGE_GROUP":             "DatabaseObjectChange",
		"DATABASE_PRINCIPAL_CHANGE_GROUP":          "DatabasePrincipalChange",
		"SCHEMA_OBJECT_CHANGE_GROUP":               "SchemaObjectChange",
		"DATABASE_PRINCIPAL_IMPERSONATION_GROUP":   "DatabasePrincipalImpersonation",
		"DATABASE_OBJECT_OWNERSHIP_CHANGE_GROUP":   "DatabaseObjectOwnershipChange",
		"DATABASE_OWNERSHIP_CHANGE_GROUP":          "DatabaseOwnershipChange",
		"SCHEMA_OBJECT_OWNERSHIP_CHANGE_GROUP":     "SchemaObjectOwnershipChange",
		"DATABASE_OBJECT_PERMISSION_CHANGE_GROUP":  "DatabaseObjectPermissionChange",
		"DATABASE_OPERATION_GROUP":                 "DatabaseOperation",
		"DATABASE_OBJECT_ACCESS_GROUP":             "DatabaseObjectAccess",
		"BATCH_COMPLETED_GROUP":                    "BatchCompletedGroup",
		"BATCH_STARTED_GROUP":                      "BatchStartedGroup",
		"SUCCESSFUL_LOGIN_GROUP":                   "SuccessfulLogin",
		"LOGOUT_GROUP":                             "Logout",
		"SERVER_STATE_CHANGE_GROUP":                "ServerStateChange",
		"FAILED_LOGIN_GROUP":                       "FailedLogin",
		"LOGIN_CHANGE_PASSWORD_GROUP":              "LoginChangePassword",
		"SERVER_ROLE_MEMBER_CHANGE_GROUP":          "ServerRoleMemberChange",
		"SERVER_PRINCIPAL_IMPERSONATION_GROUP":     "ServerPrincipalImpersonation",
		"SERVER_OBJECT_OWNERSHIP_CHANGE_GROUP":     "ServerObjectOwnershipChange",
		"DATABASE_MIRRORING_LOGIN_GROUP":           "DatabaseMirroringLogin",
		"BROKER_LOGIN_GROUP":                       "BrokerLogin",
		"SERVER_PERMISSION_CHANGE_GROUP":           "ServerPermissionChange",
		"SERVER_OBJECT_PERMISSION_CHANGE_GROUP":    "ServerObjectPermissionChange",
		"SERVER_OPERATION_GROUP":                   "ServerOperation",
		"TRACE_CHANGE_GROUP":                       "TraceChange",
		"SERVER_OBJECT_CHANGE_GROUP":               "ServerObjectChange",
		"SERVER_PRINCIPAL_CHANGE_GROUP":            "ServerPrincipalChange",
	}
	if mapped, ok := groupMap[strings.ToUpper(name)]; ok {
		return mapped
	}
	return capitalizeFirst(strings.ToLower(strings.ReplaceAll(name, "_", " ")))
}

// isAuditAction checks if the given word is a database audit action
func isAuditAction(word string) bool {
	actions := map[string]bool{
		"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true,
		"EXECUTE": true, "RECEIVE": true, "REFERENCES": true,
	}
	return actions[word]
}

// convertAuditActionKind converts audit action to expected format
func convertAuditActionKind(action string) string {
	actionMap := map[string]string{
		"SELECT":     "Select",
		"INSERT":     "Insert",
		"UPDATE":     "Update",
		"DELETE":     "Delete",
		"EXECUTE":    "Execute",
		"RECEIVE":    "Receive",
		"REFERENCES": "References",
	}
	if mapped, ok := actionMap[action]; ok {
		return mapped
	}
	return capitalizeFirst(strings.ToLower(action))
}

// parseAuditSpecificationPart parses an ADD or DROP part of an audit specification
func (p *Parser) parseAuditSpecificationPart(isDrop bool) (*ast.AuditSpecificationPart, error) {
	part := &ast.AuditSpecificationPart{
		IsDrop: isDrop,
	}
	p.nextToken() // consume ADD/DROP

	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (

		// Check if it's an action specification (SELECT, INSERT, etc.) or an audit group
		firstWord := strings.ToUpper(p.curTok.Literal)
		if isAuditAction(firstWord) {
			// Parse action specification
			spec := &ast.AuditActionSpecification{}

			// Parse actions
			for {
				actionKind := convertAuditActionKind(strings.ToUpper(p.curTok.Literal))
				spec.Actions = append(spec.Actions, &ast.DatabaseAuditAction{ActionKind: actionKind})
				p.nextToken()
				if p.curTok.Type == TokenComma {
					p.nextToken()
					// Check if next is ON (end of actions) or another action
					if strings.ToUpper(p.curTok.Literal) == "ON" {
						break
					}
				} else {
					break
				}
			}

			// Parse ON object
			if strings.ToUpper(p.curTok.Literal) == "ON" {
				p.nextToken() // consume ON
				objIdent := p.parseIdentifier()
				spec.TargetObject = &ast.SecurityTargetObject{
					ObjectKind: "NotSpecified",
					ObjectName: &ast.SecurityTargetObjectName{
						MultiPartIdentifier: &ast.MultiPartIdentifier{
							Identifiers: []*ast.Identifier{objIdent},
							Count:       1,
						},
					},
				}
			}

			// Parse BY principals
			if strings.ToUpper(p.curTok.Literal) == "BY" {
				p.nextToken() // consume BY
				for {
					principal := &ast.SecurityPrincipal{}
					upper := strings.ToUpper(p.curTok.Literal)
					if upper == "PUBLIC" {
						principal.PrincipalType = "Public"
						p.nextToken()
					} else if upper == "NULL" {
						principal.PrincipalType = "Null"
						p.nextToken()
					} else {
						principal.PrincipalType = "Identifier"
						principal.Identifier = p.parseIdentifier()
					}
					spec.Principals = append(spec.Principals, principal)
					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}
			}
			part.Details = spec
		} else {
			// Parse audit action group reference
			groupName := p.curTok.Literal
			part.Details = &ast.AuditActionGroupReference{
				Group: convertAuditGroupName(groupName),
			}
			p.nextToken() // consume group name
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	return part, nil
}

func (p *Parser) parseAuditTarget() (*ast.AuditTarget, error) {
	target := &ast.AuditTarget{}

	// Parse target kind (FILE, APPLICATION_LOG, SECURITY_LOG, URL, EXTERNAL_MONITOR)
	switch strings.ToUpper(p.curTok.Literal) {
	case "FILE":
		target.TargetKind = "File"
	case "APPLICATION_LOG":
		target.TargetKind = "ApplicationLog"
	case "SECURITY_LOG":
		target.TargetKind = "SecurityLog"
	case "URL":
		target.TargetKind = "Url"
	case "EXTERNAL_MONITOR":
		target.TargetKind = "ExternalMonitor"
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

	switch optName {
	case "MAXSIZE":
		// Check for UNLIMITED
		if strings.ToUpper(p.curTok.Literal) == "UNLIMITED" {
			p.nextToken()
			return &ast.MaxSizeAuditTargetOption{
				OptionKind:  "MaxSize",
				IsUnlimited: true,
				Unit:        "Unspecified",
			}, nil
		}
		// Parse size value
		size, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		// Parse unit (MB, GB, TB)
		unit := "Unspecified"
		unitUpper := strings.ToUpper(p.curTok.Literal)
		if unitUpper == "MB" || unitUpper == "GB" || unitUpper == "TB" {
			unit = unitUpper
			p.nextToken()
		}
		return &ast.MaxSizeAuditTargetOption{
			OptionKind:  "MaxSize",
			Size:        size,
			Unit:        unit,
			IsUnlimited: false,
		}, nil

	case "MAX_ROLLOVER_FILES":
		// Check for UNLIMITED
		if strings.ToUpper(p.curTok.Literal) == "UNLIMITED" {
			p.nextToken()
			return &ast.MaxRolloverFilesAuditTargetOption{
				OptionKind:  "MaxRolloverFiles",
				IsUnlimited: true,
			}, nil
		}
		// Parse value
		val, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.MaxRolloverFilesAuditTargetOption{
			OptionKind:  "MaxRolloverFiles",
			Value:       val,
			IsUnlimited: false,
		}, nil

	case "RESERVE_DISK_SPACE":
		// Parse ON/OFF
		value := "Off"
		valUpper := strings.ToUpper(p.curTok.Literal)
		if valUpper == "ON" || p.curTok.Type == TokenOn {
			value = "On"
		}
		p.nextToken()
		return &ast.OnOffAuditTargetOption{
			OptionKind: "ReserveDiskSpace",
			Value:      value,
		}, nil

	case "RETENTION_DAYS":
		// Parse the number of days
		days, err := p.parseScalarExpression()
		if err != nil {
			return nil, err
		}
		return &ast.RetentionDaysAuditTargetOption{
			OptionKind: "RetentionDays",
			Days:       days,
		}, nil

	default:
		// Parse literal value (FILEPATH, etc.)
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
		default:
			optKind = capitalizeFirst(strings.ToLower(optName))
		}
		return &ast.LiteralAuditTargetOption{
			OptionKind: optKind,
			Value:      val,
		}, nil
	}
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

	// Parse WITH options (like RECOMPILE, ENCRYPTION, EXECUTE AS, etc.)
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

	// Parse optional FOR REPLICATION
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if strings.ToUpper(p.curTok.Literal) == "REPLICATION" {
			stmt.IsForReplication = true
			p.nextToken() // consume REPLICATION
		}
	}

	// Expect AS
	if p.curTok.Type == TokenAs {
		p.nextToken()
	}

	// Check for EXTERNAL NAME (CLR procedure)
	if strings.ToUpper(p.curTok.Literal) == "EXTERNAL" {
		p.nextToken() // consume EXTERNAL
		if strings.ToUpper(p.curTok.Literal) == "NAME" {
			p.nextToken() // consume NAME
		}
		// Parse assembly.class.method
		stmt.MethodSpecifier = &ast.MethodSpecifier{}
		stmt.MethodSpecifier.AssemblyName = p.parseIdentifier()
		if p.curTok.Type == TokenDot {
			p.nextToken()
			stmt.MethodSpecifier.ClassName = p.parseIdentifier()
		}
		if p.curTok.Type == TokenDot {
			p.nextToken()
			stmt.MethodSpecifier.MethodName = p.parseIdentifier()
		}
		return stmt, nil
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

		// Parse optional VARYING (for CURSOR type)
		if strings.ToUpper(p.curTok.Literal) == "VARYING" {
			param.IsVarying = true
			p.nextToken()
		}

		// Parse optional default value
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			val, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			// Convert single-identifier ColumnReferenceExpression to IdentifierLiteral
			// (e.g., for default values like false, true, null)
			if colRef, ok := val.(*ast.ColumnReferenceExpression); ok {
				if colRef.MultiPartIdentifier != nil && colRef.MultiPartIdentifier.Count == 1 &&
					len(colRef.MultiPartIdentifier.Identifiers) == 1 {
					ident := colRef.MultiPartIdentifier.Identifiers[0]
					val = &ast.IdentifierLiteral{
						LiteralType: "Identifier",
						QuoteType:   ident.QuoteType,
						Value:       ident.Value,
					}
				}
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

		// Check for END (end of BEGIN block or TRY/CATCH, or END CONVERSATION statement)
		if p.curTok.Type == TokenEnd {
			// Look ahead to check if it's END CONVERSATION (a statement)
			if p.peekTok.Type == TokenConversation {
				// It's END CONVERSATION statement, parse it
				stmt, err := p.parseEndConversationStatement()
				if err != nil {
					return nil, err
				}
				if stmt != nil {
					sl.Statements = append(sl.Statements, stmt)
				}
				continue
			}
			// Otherwise it's the end of a BEGIN block
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

	// Check for WITH CHECK OPTION
	if p.curTok.Type == TokenWith {
		p.nextToken()
		if strings.ToUpper(p.curTok.Literal) == "CHECK" {
			p.nextToken()
			if strings.ToUpper(p.curTok.Literal) == "OPTION" {
				p.nextToken()
				stmt.WithCheckOption = true
			}
		}
	}

	return stmt, nil
}

func (p *Parser) parseCreateMaterializedViewStatement() (*ast.CreateViewStatement, error) {
	// Consume MATERIALIZED
	p.nextToken()

	// Expect VIEW
	if p.curTok.Type != TokenView {
		return nil, fmt.Errorf("expected VIEW after MATERIALIZED, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.CreateViewStatement{
		IsMaterialized: true,
	}

	// Parse view name
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.SchemaObjectName = son

	// Parse WITH options for materialized view
	if p.curTok.Type == TokenWith || strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken()
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optionName := strings.ToUpper(p.curTok.Literal)
				p.nextToken()

				if optionName == "DISTRIBUTION" {
					// Parse DISTRIBUTION = HASH(col1, col2, ...) or DISTRIBUTION = ROUND_ROBIN
					if p.curTok.Type == TokenEquals {
						p.nextToken()
					}
					if strings.ToUpper(p.curTok.Literal) == "HASH" {
						p.nextToken()
						if p.curTok.Type == TokenLParen {
							p.nextToken()
							hashPolicy := &ast.ViewHashDistributionPolicy{}
							// Parse column list
							for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
								col := p.parseIdentifier()
								if hashPolicy.DistributionColumn == nil {
									hashPolicy.DistributionColumn = col
								}
								hashPolicy.DistributionColumns = append(hashPolicy.DistributionColumns, col)
								if p.curTok.Type == TokenComma {
									p.nextToken()
								} else {
									break
								}
							}
							if p.curTok.Type == TokenRParen {
								p.nextToken()
							}
							stmt.ViewOptions = append(stmt.ViewOptions, &ast.ViewDistributionOption{
								OptionKind: "Distribution",
								Value:      hashPolicy,
							})
						}
					} else if strings.ToUpper(p.curTok.Literal) == "ROUND_ROBIN" {
						p.nextToken() // consume ROUND_ROBIN
						stmt.ViewOptions = append(stmt.ViewOptions, &ast.ViewDistributionOption{
							OptionKind: "Distribution",
							Value:      &ast.ViewRoundRobinDistributionPolicy{},
						})
					}
				} else if optionName == "FOR_APPEND" {
					stmt.ViewOptions = append(stmt.ViewOptions, &ast.ViewForAppendOption{
						OptionKind: "ForAppend",
					})
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

	return stmt, nil
}

func (p *Parser) parseAlterMaterializedViewStatement() (*ast.AlterViewStatement, error) {
	// Consume MATERIALIZED
	p.nextToken()

	// Expect VIEW
	if p.curTok.Type != TokenView {
		return nil, fmt.Errorf("expected VIEW after MATERIALIZED, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.AlterViewStatement{
		IsMaterialized: true,
	}

	// Parse view name
	son, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt.SchemaObjectName = son

	// Parse REBUILD or DISABLE
	switch strings.ToUpper(p.curTok.Literal) {
	case "REBUILD":
		stmt.IsRebuild = true
		p.nextToken()
	case "DISABLE":
		stmt.IsDisable = true
		p.nextToken()
	}

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
		if p.curTok.Type == TokenCreate || p.curTok.Type == TokenGrant || p.curTok.Type == TokenDeny || p.curTok.Type == TokenRevoke {
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

func (p *Parser) parseCreateDatabaseEncryptionKeyStatement() (*ast.CreateDatabaseEncryptionKeyStatement, error) {
	// curTok is ENCRYPTION
	p.nextToken() // consume ENCRYPTION

	// Consume KEY
	if p.curTok.Type == TokenKey {
		p.nextToken()
	}

	stmt := &ast.CreateDatabaseEncryptionKeyStatement{}

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

	// Check for implicit procedure execution (identifier followed by parameters)
	// This happens at batch start where you can call a stored procedure without EXEC
	if p.isImplicitExecuteParameter() {
		return p.parseImplicitExecuteStatement(label)
	}

	// Not a label or implicit execute - be lenient and skip to end of statement
	// This handles malformed SQL like "abcde" or other unknown identifiers
	p.skipToEndOfStatement()
	return &ast.LabelStatement{Value: label}, nil
}

// isImplicitExecuteParameter checks if current token could be a parameter for implicit EXEC
func (p *Parser) isImplicitExecuteParameter() bool {
	switch p.curTok.Type {
	case TokenString, TokenNationalString, TokenNumber:
		return true
	case TokenIdent:
		// Variables (@var) or identifiers followed by comma/semicolon
		if strings.HasPrefix(p.curTok.Literal, "@") {
			return true
		}
		// DEFAULT keyword
		if strings.ToUpper(p.curTok.Literal) == "DEFAULT" {
			return true
		}
		// Regular identifier as parameter (like sp_addtype birthday, datetime)
		return true
	case TokenSemicolon, TokenEOF:
		// No parameters - could still be implicit exec
		return true
	default:
		return false
	}
}

// parseImplicitExecuteStatement parses an implicit EXEC statement (procedure call without EXEC keyword)
func (p *Parser) parseImplicitExecuteStatement(procName string) (ast.Statement, error) {
	// Build the SchemaObjectName from the procedure name
	// Use the same identifier pointer for both Identifiers array and BaseIdentifier
	// so that JSON marshaling can use $ref
	baseIdent := &ast.Identifier{Value: procName, QuoteType: "NotQuoted"}
	son := &ast.SchemaObjectName{
		Count:          1,
		Identifiers:    []*ast.Identifier{baseIdent},
		BaseIdentifier: baseIdent,
	}

	procRef := &ast.ExecutableProcedureReference{
		ProcedureReference: &ast.ProcedureReferenceName{
			ProcedureReference: &ast.ProcedureReference{
				Name: son,
			},
		},
	}

	// Parse parameters
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && !p.isStatementTerminator() {
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

	spec := &ast.ExecuteSpecification{
		ExecutableEntity: procRef,
	}

	stmt := &ast.ExecuteStatement{ExecuteSpecification: spec}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
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

	// Optional WHERE clause - RECEIVE uses simplified WHERE: CONVERSATION_GROUP_ID = value or CONVERSATION_HANDLE = value
	if p.curTok.Type == TokenWhere {
		p.nextToken() // consume WHERE

		// Check for conversation_group_id
		if strings.ToUpper(p.curTok.Literal) == "CONVERSATION_GROUP_ID" {
			stmt.IsConversationGroupIdWhere = true
			p.nextToken() // consume CONVERSATION_GROUP_ID
		} else if strings.ToUpper(p.curTok.Literal) == "CONVERSATION_HANDLE" {
			p.nextToken() // consume CONVERSATION_HANDLE
		}

		// Skip equals sign
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}

		// Parse the value (usually a variable reference)
		where, err := p.parseScalarExpression()
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

	// Check for MASTER KEY
	if strings.ToUpper(p.curTok.Literal) == "MASTER" {
		return p.parseBackupMasterKeyStatement()
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

	// Parse optional file specification (READ_WRITE_FILEGROUPS, FILE, FILEGROUP, etc.)
	var files []*ast.BackupRestoreFileInfo
	for {
		upperLiteral := strings.ToUpper(p.curTok.Literal)
		if upperLiteral == "READ_WRITE_FILEGROUPS" {
			files = append(files, &ast.BackupRestoreFileInfo{
				ItemKind: "ReadWriteFileGroups",
			})
			p.nextToken()
		} else if upperLiteral == "FILE" {
			p.nextToken()
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after FILE, got %s", p.curTok.Literal)
			}
			p.nextToken()
			fileInfo := &ast.BackupRestoreFileInfo{
				ItemKind: "Files",
			}
			// Check for parenthesized list: FILE = ('f1', 'f2')
			if p.curTok.Type == TokenLParen {
				p.nextToken()
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					expr, err := p.parsePrimaryExpression()
					if err != nil {
						return nil, err
					}
					fileInfo.Items = append(fileInfo.Items, expr)
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
				expr, err := p.parsePrimaryExpression()
				if err != nil {
					return nil, err
				}
				fileInfo.Items = append(fileInfo.Items, expr)
			}
			files = append(files, fileInfo)
		} else if upperLiteral == "FILEGROUP" {
			p.nextToken()
			if p.curTok.Type != TokenEquals {
				return nil, fmt.Errorf("expected = after FILEGROUP, got %s", p.curTok.Literal)
			}
			p.nextToken()
			fileInfo := &ast.BackupRestoreFileInfo{
				ItemKind: "FileGroups",
			}
			// Check for parenthesized list: FILEGROUP = ('fg1', 'fg2')
			if p.curTok.Type == TokenLParen {
				p.nextToken()
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					expr, err := p.parsePrimaryExpression()
					if err != nil {
						return nil, err
					}
					fileInfo.Items = append(fileInfo.Items, expr)
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
				expr, err := p.parsePrimaryExpression()
				if err != nil {
					return nil, err
				}
				fileInfo.Items = append(fileInfo.Items, expr)
			}
			files = append(files, fileInfo)
		} else {
			break
		}
		// Check for comma to continue with more files
		if p.curTok.Type == TokenComma {
			p.nextToken()
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

	// Parse optional MIRROR TO clause(s)
	var mirrorToClauses []*ast.MirrorToClause
	for strings.ToUpper(p.curTok.Literal) == "MIRROR" {
		p.nextToken() // consume MIRROR
		if p.curTok.Type != TokenTo {
			return nil, fmt.Errorf("expected TO after MIRROR, got %s", p.curTok.Literal)
		}
		p.nextToken() // consume TO

		mirrorClause := &ast.MirrorToClause{}
		// Parse mirror devices
		for {
			mirrorDevice := &ast.DeviceInfo{
				DeviceType: "None",
			}

			// Check for device type (DISK, TAPE, URL, etc.)
			mirrorDeviceType := strings.ToUpper(p.curTok.Literal)
			hasMirrorPhysicalType := false
			if mirrorDeviceType == "DISK" || mirrorDeviceType == "TAPE" || mirrorDeviceType == "URL" || mirrorDeviceType == "VIRTUAL_DEVICE" {
				hasMirrorPhysicalType = true
				switch mirrorDeviceType {
				case "DISK":
					mirrorDevice.DeviceType = "Disk"
				case "TAPE":
					mirrorDevice.DeviceType = "Tape"
				case "URL":
					mirrorDevice.DeviceType = "Url"
				case "VIRTUAL_DEVICE":
					mirrorDevice.DeviceType = "VirtualDevice"
				}
				p.nextToken()
				if p.curTok.Type != TokenEquals {
					return nil, fmt.Errorf("expected = after device type, got %s", p.curTok.Literal)
				}
				p.nextToken()
			}

			// Parse device name
			if hasMirrorPhysicalType {
				// Physical device: use PhysicalDevice field with ScalarExpression
				if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
					mirrorDevice.PhysicalDevice = &ast.VariableReference{
						Name: p.curTok.Literal,
					}
					p.nextToken()
				} else if p.curTok.Type == TokenString {
					str, err := p.parseStringLiteral()
					if err != nil {
						return nil, err
					}
					mirrorDevice.PhysicalDevice = str
				} else {
					return nil, fmt.Errorf("expected string or variable for physical device, got %s", p.curTok.Literal)
				}
			} else {
				// Logical device: use LogicalDevice field with IdentifierOrValueExpression
				if p.curTok.Type == TokenIdent && len(p.curTok.Literal) > 0 && p.curTok.Literal[0] == '@' {
					mirrorDevice.LogicalDevice = &ast.IdentifierOrValueExpression{
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
					mirrorDevice.LogicalDevice = &ast.IdentifierOrValueExpression{
						Value:           str.Value,
						ValueExpression: str,
					}
				} else {
					id := p.parseIdentifier()
					mirrorDevice.LogicalDevice = &ast.IdentifierOrValueExpression{
						Value:      id.Value,
						Identifier: id,
					}
				}
			}

			mirrorClause.Devices = append(mirrorClause.Devices, mirrorDevice)

			// Check for comma (more mirror devices)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
		mirrorToClauses = append(mirrorToClauses, mirrorClause)
	}

	// Parse optional WITH clause
	var options []ast.BackupOptionBase
	if p.curTok.Type == TokenWith {
		p.nextToken()

		for {
			optionName := strings.ToUpper(p.curTok.Literal)

			// Check for ENCRYPTION with parentheses
			if optionName == "ENCRYPTION" && p.peekTok.Type == TokenLParen {
				encOpt, err := p.parseBackupEncryptionOption()
				if err != nil {
					return nil, err
				}
				options = append(options, encOpt)
			} else {
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
				case "STATS":
					option.OptionKind = "Stats"
				case "BLOCKSIZE":
					option.OptionKind = "BlockSize"
				case "BUFFERCOUNT":
					option.OptionKind = "BufferCount"
				case "DESCRIPTION":
					option.OptionKind = "Description"
				case "DIFFERENTIAL":
					option.OptionKind = "Differential"
				case "EXPIREDATE":
					option.OptionKind = "ExpireDate"
				case "MEDIANAME":
					option.OptionKind = "MediaName"
				case "MEDIADESCRIPTION":
					option.OptionKind = "MediaDescription"
				case "RETAINDAYS":
					option.OptionKind = "RetainDays"
				case "SKIP":
					option.OptionKind = "Skip"
				case "NOSKIP":
					option.OptionKind = "NoSkip"
				case "REWIND":
					option.OptionKind = "Rewind"
				case "NOREWIND":
					option.OptionKind = "NoRewind"
				case "UNLOAD":
					option.OptionKind = "Unload"
				case "NOUNLOAD":
					option.OptionKind = "NoUnload"
				case "RESTART":
					option.OptionKind = "Restart"
				case "COPY_ONLY":
					option.OptionKind = "CopyOnly"
				case "NAME":
					option.OptionKind = "Name"
				case "MAXTRANSFERSIZE":
					option.OptionKind = "MaxTransferSize"
				case "NO_TRUNCATE":
					option.OptionKind = "NoTruncate"
				case "NORECOVERY":
					option.OptionKind = "NoRecovery"
				case "STANDBY":
					option.OptionKind = "Standby"
				case "NO_LOG":
					option.OptionKind = "NoLog"
				case "TRUNCATE_ONLY":
					option.OptionKind = "TruncateOnly"
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

	if isLog {
		return &ast.BackupTransactionLogStatement{
			DatabaseName: dbName,
			Devices:      devices,
			Options:      options,
		}, nil
	}
	return &ast.BackupDatabaseStatement{
		Files:           files,
		DatabaseName:    dbName,
		MirrorToClauses: mirrorToClauses,
		Devices:         devices,
		Options:         options,
	}, nil
}

func (p *Parser) parseBackupEncryptionOption() (*ast.BackupEncryptionOption, error) {
	// curTok is ENCRYPTION
	p.nextToken() // consume ENCRYPTION

	if p.curTok.Type != TokenLParen {
		return nil, fmt.Errorf("expected ( after ENCRYPTION, got %s", p.curTok.Literal)
	}
	p.nextToken() // consume (

	opt := &ast.BackupEncryptionOption{
		OptionKind: "None",
	}

	// Parse options inside parentheses
	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		optName := strings.ToUpper(p.curTok.Literal)
		p.nextToken()

		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		switch optName {
		case "ALGORITHM":
			// Parse algorithm type: AES_128, AES_192, AES_256, TRIPLE_DES_3KEY
			algName := strings.ToUpper(p.curTok.Literal)
			switch algName {
			case "AES_128":
				opt.Algorithm = "Aes128"
			case "AES_192":
				opt.Algorithm = "Aes192"
			case "AES_256":
				opt.Algorithm = "Aes256"
			case "TRIPLE_DES_3KEY":
				opt.Algorithm = "TripleDes3Key"
			default:
				opt.Algorithm = algName
			}
			p.nextToken()
		case "SERVER":
			// SERVER CERTIFICATE or SERVER ASYMMETRIC KEY
			mechType := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			opt.Encryptor = &ast.CryptoMechanism{}
			switch mechType {
			case "CERTIFICATE":
				opt.Encryptor.CryptoMechanismType = "Certificate"
			case "ASYMMETRIC":
				// Consume KEY
				if p.curTok.Type == TokenKey {
					p.nextToken()
				}
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}
				opt.Encryptor.CryptoMechanismType = "AsymmetricKey"
			}

			// Parse identifier
			opt.Encryptor.Identifier = p.parseIdentifier()
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return opt, nil
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

func (p *Parser) parseBackupMasterKeyStatement() (*ast.BackupMasterKeyStatement, error) {
	// Consume MASTER
	p.nextToken()

	// Expect KEY
	if p.curTok.Type != TokenKey {
		return nil, fmt.Errorf("expected KEY after MASTER, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.BackupMasterKeyStatement{}

	// Expect TO
	if p.curTok.Type != TokenTo {
		return nil, fmt.Errorf("expected TO after MASTER KEY, got %s", p.curTok.Literal)
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
	case "RESOURCE":
		return p.parseCreateExternalResourcePoolStatement()
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
				// Handle other options (SERDE_METHOD, DATA_COMPRESSION) as literal options
				optionKind := p.externalFileFormatOptionKind(optName)
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
					// Parse value (string literal or identifier like FALSE/TRUE)
					if p.curTok.Type == TokenString {
						val, _ := p.parseStringLiteral()
						stmt.ExternalFileFormatOptions = append(stmt.ExternalFileFormatOptions, &ast.ExternalFileFormatLiteralOption{
							OptionKind: optionKind,
							Value:      val,
						})
					} else {
						// Handle identifiers like FALSE, TRUE, etc.
						val := &ast.StringLiteral{
							LiteralType: "String",
							Value:       p.curTok.Literal,
						}
						p.nextToken()
						stmt.ExternalFileFormatOptions = append(stmt.ExternalFileFormatOptions, &ast.ExternalFileFormatLiteralOption{
							OptionKind: optionKind,
							Value:      val,
						})
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

		// Special handling for USE_TYPE_DEFAULT which uses ExternalFileFormatUseDefaultTypeOption
		if optName == "USE_TYPE_DEFAULT" {
			// Value is TRUE or FALSE (as identifier, not string)
			value := strings.ToUpper(p.curTok.Literal)
			defaultType := "False"
			if value == "TRUE" {
				defaultType = "True"
			}
			p.nextToken()
			return &ast.ExternalFileFormatUseDefaultTypeOption{
				OptionKind:                       optionKind,
				ExternalFileFormatUseDefaultType: defaultType,
			}
		}

		// Handle integer values for FIRST_ROW
		if optName == "FIRST_ROW" {
			val := &ast.IntegerLiteral{
				LiteralType: "Integer",
				Value:       p.curTok.Literal,
			}
			p.nextToken()
			return &ast.ExternalFileFormatLiteralOption{
				OptionKind: optionKind,
				Value:      val,
			}
		}

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
	case "SERDE_METHOD":
		return "SerDeMethod"
	default:
		return name
	}
}

func (p *Parser) parseCreateExternalTableStatement() (*ast.CreateExternalTableStatement, error) {
	p.nextToken() // consume TABLE

	name, err := p.parseSchemaObjectName()
	if err != nil {
		return nil, err
	}
	stmt := &ast.CreateExternalTableStatement{
		SchemaObjectName: name,
	}

	// Parse column definitions in parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			colDef, err := p.parseExternalTableColumnDefinition()
			if err != nil {
				return nil, err
			}
			stmt.ColumnDefinitions = append(stmt.ColumnDefinitions, colDef)
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume comma
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse WITH clause for options
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				optName := strings.ToUpper(p.curTok.Literal)
				p.nextToken() // consume option name

				// Expect =
				if p.curTok.Type == TokenEquals {
					p.nextToken() // consume =
				}

				switch optName {
				case "DATA_SOURCE":
					stmt.DataSource = p.parseIdentifier()
				case "REJECT_TYPE":
					opt := &ast.ExternalTableRejectTypeOption{
						OptionKind: "RejectType",
					}
					// VALUE or PERCENTAGE
					val := strings.ToUpper(p.curTok.Literal)
					switch val {
					case "VALUE":
						opt.Value = "Value"
					case "PERCENTAGE":
						opt.Value = "Percentage"
					default:
						opt.Value = val
					}
					p.nextToken() // consume value
					stmt.ExternalTableOptions = append(stmt.ExternalTableOptions, opt)
				case "REJECT_VALUE", "REJECT_SAMPLE_VALUE":
					opt := &ast.ExternalTableLiteralOrIdentifierOption{
						Value: &ast.IdentifierOrValueExpression{},
					}
					if optName == "REJECT_VALUE" {
						opt.OptionKind = "RejectValue"
					} else {
						opt.OptionKind = "RejectSampleValue"
					}
					// Parse numeric or integer literal
					if p.curTok.Type == TokenNumber {
						if strings.Contains(p.curTok.Literal, ".") {
							numLit := &ast.NumericLiteral{LiteralType: "Numeric", Value: p.curTok.Literal}
							opt.Value.Value = p.curTok.Literal
							opt.Value.ValueExpression = numLit
						} else {
							intLit := &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
							opt.Value.Value = p.curTok.Literal
							opt.Value.ValueExpression = intLit
						}
						p.nextToken()
					}
					stmt.ExternalTableOptions = append(stmt.ExternalTableOptions, opt)
				case "DISTRIBUTION":
					// Parse DISTRIBUTION = SHARDED(col), ROUND_ROBIN, or REPLICATE
					distVal := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					opt := &ast.ExternalTableDistributionOption{
						OptionKind: "Distribution",
					}
					if distVal == "SHARDED" {
						if p.curTok.Type == TokenLParen {
							p.nextToken() // consume (
							sharded := &ast.ExternalTableShardedDistributionPolicy{
								ShardingColumn: p.parseIdentifier(),
							}
							if p.curTok.Type == TokenRParen {
								p.nextToken() // consume )
							}
							opt.Value = sharded
						}
					} else if distVal == "ROUND_ROBIN" {
						opt.Value = &ast.ExternalTableRoundRobinDistributionPolicy{}
					} else if distVal == "REPLICATE" || distVal == "REPLICATED" {
						opt.Value = &ast.ExternalTableReplicatedDistributionPolicy{}
					}
					stmt.ExternalTableOptions = append(stmt.ExternalTableOptions, opt)
				case "LOCATION", "FILE_FORMAT", "TABLE_OPTIONS", "SCHEMA_NAME", "OBJECT_NAME", "REJECTED_ROW_LOCATION":
					opt := &ast.ExternalTableLiteralOrIdentifierOption{
						Value: &ast.IdentifierOrValueExpression{},
					}
					switch optName {
					case "LOCATION":
						opt.OptionKind = "Location"
					case "FILE_FORMAT":
						opt.OptionKind = "FileFormat"
					case "TABLE_OPTIONS":
						opt.OptionKind = "TableOptions"
					case "SCHEMA_NAME":
						opt.OptionKind = "SchemaName"
					case "OBJECT_NAME":
						opt.OptionKind = "ObjectName"
					case "REJECTED_ROW_LOCATION":
						opt.OptionKind = "RejectedRowLocation"
					}

					// Parse the value (can be identifier or string literal)
					if p.curTok.Type == TokenString {
						strLit := p.parseStringLiteralValue()
						p.nextToken() // consume string
						opt.Value.Value = strLit.Value
						opt.Value.ValueExpression = strLit
					} else if p.curTok.Type == TokenNationalString {
						strLit, _ := p.parseNationalStringFromToken()
						opt.Value.Value = strLit.Value
						opt.Value.ValueExpression = strLit
					} else if p.curTok.Type == TokenIdent || p.curTok.Type == TokenLBracket {
						ident := p.parseIdentifier()
						opt.Value.Value = ident.Value
						opt.Value.Identifier = ident
					}
					stmt.ExternalTableOptions = append(stmt.ExternalTableOptions, opt)
				default:
					// Skip unknown options
					for p.curTok.Type != TokenComma && p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
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
	}

	// Parse optional AS SELECT (CTAS syntax)
	if p.curTok.Type == TokenAs {
		p.nextToken() // consume AS
		selectStmt, err := p.parseSelectStatement()
		if err != nil {
			return nil, err
		}
		stmt.SelectStatement = selectStmt
	}

	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseExternalTableColumnDefinition() (*ast.ExternalTableColumnDefinition, error) {
	colDef := &ast.ExternalTableColumnDefinition{
		ColumnDefinition: &ast.ColumnDefinitionBase{},
	}

	// Parse column name
	colDef.ColumnDefinition.ColumnIdentifier = p.parseIdentifier()

	// Parse data type
	dt, err := p.parseDataType()
	if err != nil {
		return nil, err
	}
	colDef.ColumnDefinition.DataType = dt

	// Parse optional COLLATE
	if strings.ToUpper(p.curTok.Literal) == "COLLATE" {
		p.nextToken() // consume COLLATE
		colDef.ColumnDefinition.Collation = p.parseIdentifier()
	}

	// Parse optional NULL/NOT NULL
	if strings.ToUpper(p.curTok.Literal) == "NOT" {
		p.nextToken() // consume NOT
		if strings.ToUpper(p.curTok.Literal) == "NULL" {
			p.nextToken() // consume NULL
			colDef.NullableConstraint = &ast.NullableConstraintDefinition{
				Nullable: false,
			}
		}
	} else if strings.ToUpper(p.curTok.Literal) == "NULL" {
		p.nextToken() // consume NULL
		colDef.NullableConstraint = &ast.NullableConstraintDefinition{
			Nullable: true,
		}
	}

	return colDef, nil
}

func (p *Parser) parseCreateExternalLanguageStatement() (*ast.CreateExternalLanguageStatement, error) {
	p.nextToken() // consume LANGUAGE
	stmt := &ast.CreateExternalLanguageStatement{
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
		for {
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
			} else {
				break
			}
			if p.curTok.Type == TokenComma {
				p.nextToken() // consume , for multiple file options
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

func (p *Parser) parseCreateExternalResourcePoolStatement() (*ast.CreateExternalResourcePoolStatement, error) {
	// Consume RESOURCE
	p.nextToken()

	// Expect POOL
	if strings.ToUpper(p.curTok.Literal) != "POOL" {
		return nil, fmt.Errorf("expected POOL after RESOURCE, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.CreateExternalResourcePoolStatement{}

	// Parse pool name
	stmt.Name = p.parseIdentifier()

	// Check for optional WITH clause
	if p.curTok.Type == TokenWith || strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH

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

	// ON SERVER/DATABASE
	if p.curTok.Type == TokenOn {
		p.nextToken()
		scopeUpper := strings.ToUpper(p.curTok.Literal)
		if scopeUpper == "SERVER" {
			stmt.SessionScope = "Server"
			p.nextToken()
		} else if scopeUpper == "DATABASE" {
			stmt.SessionScope = "Database"
			p.nextToken()
		}
	}

	// Parse ADD EVENT/TARGET and WITH clauses
	for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && !p.isStatementTerminator() {
		upperLit := strings.ToUpper(p.curTok.Literal)

		if upperLit == "ADD" {
			p.nextToken()
			addType := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			if addType == "EVENT" {
				event := p.parseEventDeclaration()
				stmt.EventDeclarations = append(stmt.EventDeclarations, event)
			} else if addType == "TARGET" {
				target := p.parseTargetDeclaration()
				stmt.TargetDeclarations = append(stmt.TargetDeclarations, target)
			}
		} else if upperLit == "WITH" || p.curTok.Type == TokenWith {
			p.nextToken()
			if p.curTok.Type == TokenLParen {
				p.nextToken()
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					opt := p.parseSessionOption()
					if opt != nil {
						stmt.SessionOptions = append(stmt.SessionOptions, opt)
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
			}
		} else {
			p.nextToken()
		}
	}
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return stmt, nil
}

func (p *Parser) parseEventDeclaration() *ast.EventDeclaration {
	event := &ast.EventDeclaration{}

	// Parse package.event_name
	event.ObjectName = p.parseEventSessionObjectName()

	// Parse optional ( SET ... ACTION(...) WHERE ... )
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			upperLit := strings.ToUpper(p.curTok.Literal)
			if upperLit == "SET" {
				p.nextToken()
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					upperCheck := strings.ToUpper(p.curTok.Literal)
					if upperCheck == "ACTION" || upperCheck == "WHERE" {
						break
					}
					param := &ast.EventDeclarationSetParameter{
						EventField: p.parseIdentifier(),
					}
					if p.curTok.Type == TokenEquals {
						p.nextToken()
						param.EventValue, _ = p.parseScalarExpression()
					}
					event.EventDeclarationSetParameters = append(event.EventDeclarationSetParameters, param)
					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}
			} else if upperLit == "ACTION" {
				p.nextToken()
				if p.curTok.Type == TokenLParen {
					p.nextToken()
					for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
						actionName := p.parseEventSessionObjectName()
						event.EventDeclarationActionParameters = append(event.EventDeclarationActionParameters, actionName)
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
			} else if upperLit == "WHERE" {
				p.nextToken()
				event.EventDeclarationPredicateParameter = p.parseEventPredicate()
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	return event
}

func (p *Parser) parseTargetDeclaration() *ast.TargetDeclaration {
	target := &ast.TargetDeclaration{}

	// Parse package.target_name
	target.ObjectName = p.parseEventSessionObjectName()

	// Parse optional ( SET ... )
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
			if strings.ToUpper(p.curTok.Literal) == "SET" {
				p.nextToken()
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					param := &ast.EventDeclarationSetParameter{
						EventField: p.parseIdentifier(),
					}
					if p.curTok.Type == TokenEquals {
						p.nextToken()
						param.EventValue, _ = p.parseScalarExpression()
					}
					target.TargetDeclarationParameters = append(target.TargetDeclarationParameters, param)
					if p.curTok.Type == TokenComma {
						p.nextToken()
					} else {
						break
					}
				}
			} else {
				break
			}
		}
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
	}

	return target
}

func (p *Parser) parseEventSessionObjectName() *ast.EventSessionObjectName {
	var identifiers []*ast.Identifier

	for {
		if p.curTok.Type != TokenIdent && p.curTok.Type != TokenLBracket {
			break
		}
		identifiers = append(identifiers, p.parseIdentifier())
		if p.curTok.Type != TokenDot {
			break
		}
		p.nextToken() // consume dot
	}

	return &ast.EventSessionObjectName{
		MultiPartIdentifier: &ast.MultiPartIdentifier{
			Identifiers: identifiers,
			Count:       len(identifiers),
		},
	}
}

func (p *Parser) parseEventPredicate() ast.BooleanExpression {
	return p.parseEventPredicateOr()
}

func (p *Parser) parseEventPredicateOr() ast.BooleanExpression {
	left := p.parseEventPredicateAnd()
	for strings.ToUpper(p.curTok.Literal) == "OR" {
		p.nextToken()
		right := p.parseEventPredicateAnd()
		left = &ast.BooleanBinaryExpression{
			BinaryExpressionType: "Or",
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}
	return left
}

func (p *Parser) parseEventPredicateAnd() ast.BooleanExpression {
	left := p.parseEventPredicatePrimary()
	for strings.ToUpper(p.curTok.Literal) == "AND" {
		p.nextToken()
		right := p.parseEventPredicatePrimary()
		left = &ast.BooleanBinaryExpression{
			BinaryExpressionType: "And",
			FirstExpression:      left,
			SecondExpression:     right,
		}
	}
	return left
}

func (p *Parser) parseEventPredicatePrimary() ast.BooleanExpression {
	// Handle NOT operator
	if strings.ToUpper(p.curTok.Literal) == "NOT" {
		p.nextToken()
		inner := p.parseEventPredicatePrimary()
		return &ast.BooleanNotExpression{Expression: inner}
	}

	// Handle parentheses
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		expr := p.parseEventPredicateOr()
		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}
		return &ast.BooleanParenthesisExpression{Expression: expr}
	}

	// Parse [package].[function_or_field](...) or [package].[field] NOT LIKE 'pattern'
	name := p.parseEventSessionObjectName()

	// Check for function call
	if p.curTok.Type == TokenLParen {
		p.nextToken()
		// Parse function parameters
		var source *ast.SourceDeclaration
		var eventValue ast.ScalarExpression

		// First param is usually a source declaration
		sourceName := p.parseEventSessionObjectName()
		source = &ast.SourceDeclaration{Value: sourceName}

		if p.curTok.Type == TokenComma {
			p.nextToken()
			eventValue, _ = p.parseScalarExpression()
		}

		if p.curTok.Type == TokenRParen {
			p.nextToken()
		}

		return &ast.EventDeclarationCompareFunctionParameter{
			Name:              name,
			SourceDeclaration: source,
			EventValue:        eventValue,
		}
	}

	// Check for NOT LIKE or LIKE
	notLike := false
	if strings.ToUpper(p.curTok.Literal) == "NOT" {
		notLike = true
		p.nextToken()
	}

	if strings.ToUpper(p.curTok.Literal) == "LIKE" {
		p.nextToken()
		pattern, _ := p.parseScalarExpression()
		compType := "Like"
		if notLike {
			compType = "NotLike"
		}
		return &ast.BooleanComparisonExpression{
			ComparisonType:   compType,
			FirstExpression:  &ast.SourceDeclaration{Value: name},
			SecondExpression: pattern,
		}
	}

	// Handle comparison operators: =, !=, <>, <, >, <=, >=
	var compType string
	switch p.curTok.Type {
	case TokenEquals:
		compType = "Equals"
		p.nextToken()
	case TokenLessThan:
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			compType = "LessThanOrEqualTo"
		} else if p.curTok.Type == TokenGreaterThan {
			p.nextToken()
			compType = "NotEqualToBrackets"
		} else {
			compType = "LessThan"
		}
	case TokenGreaterThan:
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			compType = "GreaterThanOrEqualTo"
		} else {
			compType = "GreaterThan"
		}
	}

	if compType != "" {
		rightExpr, _ := p.parseScalarExpression()
		return &ast.BooleanComparisonExpression{
			ComparisonType:   compType,
			FirstExpression:  &ast.SourceDeclaration{Value: name},
			SecondExpression: rightExpr,
		}
	}

	// Check for != operator (exclamation equals)
	if p.curTok.Literal == "!" {
		p.nextToken()
		if p.curTok.Type == TokenEquals {
			p.nextToken()
			rightExpr, _ := p.parseScalarExpression()
			return &ast.BooleanComparisonExpression{
				ComparisonType:   "NotEqualToExclamation",
				FirstExpression:  &ast.SourceDeclaration{Value: name},
				SecondExpression: rightExpr,
			}
		}
	}

	// Fallback: return source declaration wrapped in something
	return &ast.SourceDeclaration{Value: name}
}

func (p *Parser) parseSessionOption() ast.SessionOption {
	optName := strings.ToUpper(p.curTok.Literal)
	p.nextToken()

	if p.curTok.Type == TokenEquals {
		p.nextToken()
	}

	switch optName {
	case "MAX_MEMORY", "MAX_EVENT_SIZE":
		value, _ := p.parseScalarExpression()
		unit := ""
		if strings.ToUpper(p.curTok.Literal) == "KB" || strings.ToUpper(p.curTok.Literal) == "MB" {
			unit = strings.ToUpper(p.curTok.Literal)
			p.nextToken()
		}
		return &ast.LiteralSessionOption{
			OptionKind: p.sessionOptionKind(optName),
			Value:      value,
			Unit:       unit,
		}
	case "EVENT_RETENTION_MODE":
		value := p.curTok.Literal
		p.nextToken()
		return &ast.EventRetentionSessionOption{
			OptionKind: "EventRetention",
			Value:      p.eventRetentionValue(value),
		}
	case "MAX_DISPATCH_LATENCY":
		value, _ := p.parseScalarExpression()
		// Check for SECONDS
		if strings.ToUpper(p.curTok.Literal) == "SECONDS" {
			p.nextToken()
		}
		return &ast.MaxDispatchLatencySessionOption{
			OptionKind: "MaxDispatchLatency",
			Value:      value,
			IsInfinite: false,
		}
	case "MEMORY_PARTITION_MODE":
		value := p.curTok.Literal
		p.nextToken()
		return &ast.MemoryPartitionSessionOption{
			OptionKind: "MemoryPartition",
			Value:      p.memoryPartitionValue(value),
		}
	case "TRACK_CAUSALITY", "STARTUP_STATE":
		stateUpper := strings.ToUpper(p.curTok.Literal)
		p.nextToken()
		state := "Off"
		if stateUpper == "ON" {
			state = "On"
		}
		return &ast.OnOffSessionOption{
			OptionKind:  p.sessionOptionKind(optName),
			OptionState: state,
		}
	default:
		// Skip unknown option value
		p.nextToken()
		return nil
	}
}

func (p *Parser) sessionOptionKind(name string) string {
	switch name {
	case "MAX_MEMORY":
		return "MaxMemory"
	case "MAX_EVENT_SIZE":
		return "MaxEventSize"
	case "TRACK_CAUSALITY":
		return "TrackCausality"
	case "STARTUP_STATE":
		return "StartUpState"
	default:
		return name
	}
}

func (p *Parser) eventRetentionValue(value string) string {
	switch strings.ToUpper(value) {
	case "ALLOW_SINGLE_EVENT_LOSS":
		return "AllowSingleEventLoss"
	case "ALLOW_MULTIPLE_EVENT_LOSS":
		return "AllowMultipleEventLoss"
	case "NO_EVENT_LOSS":
		return "NoEventLoss"
	default:
		return value
	}
}

func (p *Parser) memoryPartitionValue(value string) string {
	switch strings.ToUpper(value) {
	case "NONE":
		return "None"
	case "PER_CPU":
		return "PerCpu"
	case "PER_NODE":
		return "PerNode"
	default:
		return value
	}
}

func (p *Parser) parseCreateEventSessionStatementFromEvent() (*ast.CreateEventSessionStatement, error) {
	// EVENT has already been consumed, curTok is SESSION
	p.nextToken() // consume SESSION

	stmt := &ast.CreateEventSessionStatement{
		Name: p.parseIdentifier(),
	}

	// ON SERVER/DATABASE
	if p.curTok.Type == TokenOn {
		p.nextToken()
		scopeUpper := strings.ToUpper(p.curTok.Literal)
		if scopeUpper == "SERVER" {
			stmt.SessionScope = "Server"
			p.nextToken()
		} else if scopeUpper == "DATABASE" {
			stmt.SessionScope = "Database"
			p.nextToken()
		}
	}

	// Parse ADD EVENT/TARGET and WITH clauses
	for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && !p.isStatementTerminator() {
		upperLit := strings.ToUpper(p.curTok.Literal)

		if upperLit == "ADD" {
			p.nextToken()
			addType := strings.ToUpper(p.curTok.Literal)
			p.nextToken()

			if addType == "EVENT" {
				event := p.parseEventDeclaration()
				stmt.EventDeclarations = append(stmt.EventDeclarations, event)
			} else if addType == "TARGET" {
				target := p.parseTargetDeclaration()
				stmt.TargetDeclarations = append(stmt.TargetDeclarations, target)
			}
		} else if upperLit == "WITH" || p.curTok.Type == TokenWith {
			p.nextToken()
			if p.curTok.Type == TokenLParen {
				p.nextToken()
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					opt := p.parseSessionOption()
					if opt != nil {
						stmt.SessionOptions = append(stmt.SessionOptions, opt)
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
			}
		} else {
			p.nextToken()
		}
	}
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
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
			upperName := strings.ToUpper(eventName)

			// Determine if it's a group or a type
			// Groups are: names ending in "_EVENTS" or "EVENTS", or TRC_* or DDL_* prefixed names
			isGroup := strings.HasSuffix(upperName, "_EVENTS") ||
				strings.HasSuffix(upperName, "EVENTS") ||
				strings.HasPrefix(upperName, "TRC_") ||
				strings.HasPrefix(upperName, "DDL_")

			if isGroup {
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
// eventTypeNameMap maps uppercase event type names to their correct PascalCase equivalents
var eventTypeNameMap = map[string]string{
	// Audit events with DB (must be uppercase)
	"AUDIT_ADD_DB_USER_EVENT":                        "AuditAddDBUserEvent",
	"AUDIT_ADD_MEMBER_TO_DB_ROLE_EVENT":              "AuditAddMemberToDBRoleEvent",
	"AUDIT_ADDLOGIN_EVENT":                           "AuditAddLoginEvent",
	// Log events
	"ERRORLOG":                                       "ErrorLog",
	"EVENTLOG":                                       "EventLog",
	// OLEDB events
	"OLEDB_DATAREAD_EVENT":                           "OledbDataReadEvent",
	"OLEDB_QUERYINTERFACE_EVENT":                     "OledbQueryInterfaceEvent",
	// Showplan events
	"SHOWPLAN_ALL_FOR_QUERY_COMPILE":                 "ShowPlanAllForQueryCompile",
	"SHOWPLAN_XML_FOR_QUERY_COMPILE":                 "ShowPlanXmlForQueryCompile",
	"SHOWPLAN_XML":                                   "ShowPlanXml",
	"SHOWPLAN_XML_STATISTICS_PROFILE":                "ShowPlanXmlStatisticsProfile",
	// SP cache events
	"SP_CACHEINSERT":                                 "SpCacheInsert",
	"SP_CACHEMISS":                                   "SpCacheMiss",
	"SP_CACHEREMOVE":                                 "SpCacheRemove",
	// Recompile events
	"SQL_STMTRECOMPILE":                              "SqlStmtRecompile",
	// User configurable events
	"USERCONFIGURABLE_0":                             "UserConfigurable0",
	"USERCONFIGURABLE_1":                             "UserConfigurable1",
	"USERCONFIGURABLE_2":                             "UserConfigurable2",
	"USERCONFIGURABLE_3":                             "UserConfigurable3",
	"USERCONFIGURABLE_4":                             "UserConfigurable4",
	"USERCONFIGURABLE_5":                             "UserConfigurable5",
	"USERCONFIGURABLE_6":                             "UserConfigurable6",
	"USERCONFIGURABLE_7":                             "UserConfigurable7",
	"USERCONFIGURABLE_8":                             "UserConfigurable8",
	"USERCONFIGURABLE_9":                             "UserConfigurable9",
	// XQuery
	"XQUERY_STATIC_TYPE":                             "XQueryStaticType",
	// TSql
	"TRC_TSQL":                                       "TrcTSql",
}

func eventNameToPascalCase(name string) string {
	// Check if we have a specific mapping
	if mapped, ok := eventTypeNameMap[strings.ToUpper(name)]; ok {
		return mapped
	}

	// Split by underscore
	parts := strings.Split(name, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		// Special case: DB should be uppercase
		if strings.ToUpper(part) == "DB" {
			result.WriteString("DB")
		} else {
			// Capitalize first letter, lowercase rest
			result.WriteString(strings.ToUpper(part[:1]))
			result.WriteString(strings.ToLower(part[1:]))
		}
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

	// Check for DATABASE AUDIT SPECIFICATION
	if strings.ToUpper(p.curTok.Literal) == "AUDIT" {
		p.nextToken() // consume AUDIT
		if strings.ToUpper(p.curTok.Literal) == "SPECIFICATION" {
			p.nextToken() // consume SPECIFICATION
			return p.parseCreateDatabaseAuditSpecificationStatement()
		}
	}

	// Check for DATABASE ENCRYPTION KEY
	if strings.ToUpper(p.curTok.Literal) == "ENCRYPTION" {
		return p.parseCreateDatabaseEncryptionKeyStatement()
	}

	// Check for DATABASE SCOPED CREDENTIAL
	if p.curTok.Type == TokenScoped || strings.ToUpper(p.curTok.Literal) == "SCOPED" {
		// Look ahead to see if it's SCOPED CREDENTIAL
		if p.peekTok.Type == TokenCredential {
			p.nextToken() // consume SCOPED
			return p.parseCreateCredentialStatement(true)
		}
		// Otherwise SCOPED is the database name
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

				// Check for Azure-style options after COPY OF
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					opts, err := p.parseAzureDatabaseOptions()
					if err != nil {
						return nil, err
					}
					stmt.Options = append(stmt.Options, opts...)
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}
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
		case "ATTACH_FORCE_REBUILD_LOG":
			stmt.AttachMode = "AttachForceRebuildLog"
			p.nextToken()
		}
	}

	// Check for AS SNAPSHOT OF clause
	if strings.ToUpper(p.curTok.Literal) == "AS" {
		p.nextToken() // consume AS
		if strings.ToUpper(p.curTok.Literal) == "SNAPSHOT" {
			p.nextToken() // consume SNAPSHOT
			if strings.ToUpper(p.curTok.Literal) == "OF" {
				p.nextToken() // consume OF
			}
			stmt.DatabaseSnapshot = p.parseIdentifier()
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

		case "TRUSTWORTHY":
			p.nextToken() // consume TRUSTWORTHY
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume = (optional)
			}
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.OnOffDatabaseOption{
				OptionKind:  "Trustworthy",
				OptionState: capitalizeFirst(state),
			}
			options = append(options, opt)

		case "ENABLE_BROKER":
			p.nextToken() // consume ENABLE_BROKER
			opt := &ast.SimpleDatabaseOption{
				OptionKind: "EnableBroker",
			}
			options = append(options, opt)

		case "NEW_BROKER":
			p.nextToken() // consume NEW_BROKER
			opt := &ast.SimpleDatabaseOption{
				OptionKind: "NewBroker",
			}
			options = append(options, opt)

		case "ERROR_BROKER_CONVERSATIONS":
			p.nextToken() // consume ERROR_BROKER_CONVERSATIONS
			opt := &ast.SimpleDatabaseOption{
				OptionKind: "ErrorBrokerConversations",
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

		case "DEFAULT_LANGUAGE":
			p.nextToken() // consume DEFAULT_LANGUAGE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			// Can be identifier or integer
			if p.curTok.Type == TokenNumber {
				opt := &ast.LiteralDatabaseOption{
					OptionKind: "DefaultLanguage",
					Value: &ast.IntegerLiteral{
						LiteralType: "Integer",
						Value:       p.curTok.Literal,
					},
				}
				options = append(options, opt)
				p.nextToken()
			} else {
				opt := &ast.IdentifierDatabaseOption{
					OptionKind: "DefaultLanguage",
					Value:      p.parseIdentifier(),
				}
				options = append(options, opt)
			}

		case "DEFAULT_FULLTEXT_LANGUAGE":
			p.nextToken() // consume DEFAULT_FULLTEXT_LANGUAGE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			// Can be identifier or integer
			if p.curTok.Type == TokenNumber {
				opt := &ast.LiteralDatabaseOption{
					OptionKind: "DefaultFullTextLanguage",
					Value: &ast.IntegerLiteral{
						LiteralType: "Integer",
						Value:       p.curTok.Literal,
					},
				}
				options = append(options, opt)
				p.nextToken()
			} else {
				opt := &ast.IdentifierDatabaseOption{
					OptionKind: "DefaultFullTextLanguage",
					Value:      p.parseIdentifier(),
				}
				options = append(options, opt)
			}

		case "TWO_DIGIT_YEAR_CUTOFF":
			p.nextToken() // consume TWO_DIGIT_YEAR_CUTOFF
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
			options = append(options, opt)
			p.nextToken()

		case "RESTRICTED_USER":
			p.nextToken() // consume RESTRICTED_USER
			opt := &ast.SimpleDatabaseOption{
				OptionKind: "RestrictedUser",
			}
			options = append(options, opt)

		case "FILESTREAM":
			p.nextToken() // consume FILESTREAM
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

		case "SERVICE_OBJECTIVE":
			// Check for elastic_pool(name = [epool1]) syntax
			if strings.ToUpper(p.curTok.Literal) == "ELASTIC_POOL" {
				p.nextToken() // consume ELASTIC_POOL
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					// Parse NAME = [poolname]
					if strings.ToUpper(p.curTok.Literal) == "NAME" {
						p.nextToken() // consume NAME
						if p.curTok.Type == TokenEquals {
							p.nextToken() // consume =
						}
						poolName := p.parseIdentifier()
						opt := &ast.ElasticPoolSpecification{
							OptionKind:      "ServiceObjective",
							ElasticPoolName: poolName,
						}
						options = append(options, opt)
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}
			} else {
				// Parse service objective value (string literal)
				value, _ := p.parseStringLiteral()
				opt := &ast.LiteralDatabaseOption{
					OptionKind: "ServiceObjective",
					Value:      value,
				}
				options = append(options, opt)
			}

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

		case "OFFLINE":
			p.nextToken() // consume OFFLINE
			opt := &ast.SimpleFileDeclarationOption{
				OptionKind: "Offline",
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

	// Check for FROM clause
	if p.curTok.Type == TokenFrom {
		p.nextToken() // consume FROM

		upper := strings.ToUpper(p.curTok.Literal)
		switch upper {
		case "EXTERNAL":
			// FROM EXTERNAL PROVIDER
			p.nextToken() // consume EXTERNAL
			if strings.ToUpper(p.curTok.Literal) == "PROVIDER" {
				p.nextToken() // consume PROVIDER
			}

			source := &ast.ExternalCreateLoginSource{}

			// Parse WITH options
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				source.Options = p.parsePrincipalOptions()
			}

			stmt.Source = source

		case "WINDOWS":
			// FROM WINDOWS
			p.nextToken() // consume WINDOWS

			source := &ast.WindowsCreateLoginSource{}

			// Parse WITH options
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				source.Options = p.parsePrincipalOptions()
			}

			stmt.Source = source

		case "CERTIFICATE":
			// FROM CERTIFICATE certname
			p.nextToken() // consume CERTIFICATE

			source := &ast.CertificateCreateLoginSource{
				Certificate: p.parseIdentifier(),
			}

			// Parse WITH CREDENTIAL option
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				if p.curTok.Type == TokenCredential || strings.ToUpper(p.curTok.Literal) == "CREDENTIAL" {
					p.nextToken() // consume CREDENTIAL
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					source.Credential = p.parseIdentifier()
				}
			}

			stmt.Source = source

		case "ASYMMETRIC":
			// FROM ASYMMETRIC KEY keyname
			p.nextToken() // consume ASYMMETRIC
			if p.curTok.Type == TokenKey || strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}

			source := &ast.AsymmetricKeyCreateLoginSource{
				Key: p.parseIdentifier(),
			}

			// Parse WITH CREDENTIAL option
			if p.curTok.Type == TokenWith {
				p.nextToken() // consume WITH
				if p.curTok.Type == TokenCredential || strings.ToUpper(p.curTok.Literal) == "CREDENTIAL" {
					p.nextToken() // consume CREDENTIAL
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					source.Credential = p.parseIdentifier()
				}
			}

			stmt.Source = source
		}
	} else if p.curTok.Type == TokenWith {
		// WITH PASSWORD = '...'
		p.nextToken() // consume WITH

		source := &ast.PasswordCreateLoginSource{}

		// Parse PASSWORD = 'value' [HASHED] [MUST_CHANGE]
		if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
			p.nextToken() // consume PASSWORD
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}

			// Parse password value (string or binary)
			source.Password = p.parsePasswordValue()

			// Parse optional flags and other options
			for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF && strings.ToUpper(p.curTok.Literal) != "GO" {
				upper := strings.ToUpper(p.curTok.Literal)
				if upper == "HASHED" {
					source.Hashed = true
					p.nextToken()
				} else if upper == "MUST_CHANGE" {
					source.MustChange = true
					p.nextToken()
				} else if p.curTok.Type == TokenComma {
					p.nextToken()
					// Parse remaining options
					source.Options = append(source.Options, p.parsePrincipalOptions()...)
					break
				} else {
					break
				}
			}
		}

		stmt.Source = source
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parsePasswordValue() ast.ScalarExpression {
	if p.curTok.Type == TokenString {
		value := p.curTok.Literal
		isNational := false
		if len(value) > 0 && (value[0] == 'N' || value[0] == 'n') && len(value) > 1 && value[1] == '\'' {
			isNational = true
			value = value[2 : len(value)-1]
		} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = value[1 : len(value)-1]
		}
		p.nextToken()
		return &ast.StringLiteral{
			LiteralType:   "String",
			IsNational:    isNational,
			IsLargeObject: false,
			Value:         value,
		}
	} else if p.curTok.Type == TokenBinary {
		value := p.curTok.Literal
		p.nextToken()
		return &ast.BinaryLiteral{
			LiteralType:   "Binary",
			IsLargeObject: false,
			Value:         value,
		}
	}
	// Return nil if not a recognized password value
	return nil
}

func (p *Parser) parsePrincipalOptions() []ast.PrincipalOption {
	var options []ast.PrincipalOption

	for {
		optName := strings.ToUpper(p.curTok.Literal)
		p.nextToken() // consume option name

		// Expect =
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		switch optName {
		case "SID":
			// SID = 0x... (binary literal)
			if p.curTok.Type == TokenBinary {
				options = append(options, &ast.LiteralPrincipalOption{
					OptionKind: "Sid",
					Value: &ast.BinaryLiteral{
						LiteralType:   "Binary",
						IsLargeObject: false,
						Value:         p.curTok.Literal,
					},
				})
				p.nextToken()
			}
		case "TYPE":
			// TYPE = X or TYPE = [X] or TYPE = E
			options = append(options, &ast.IdentifierPrincipalOption{
				OptionKind: "Type",
				Identifier: p.parseIdentifier(),
			})
		case "DEFAULT_DATABASE":
			options = append(options, &ast.IdentifierPrincipalOption{
				OptionKind: "DefaultDatabase",
				Identifier: p.parseIdentifier(),
			})
		case "DEFAULT_LANGUAGE":
			options = append(options, &ast.IdentifierPrincipalOption{
				OptionKind: "DefaultLanguage",
				Identifier: p.parseIdentifier(),
			})
		case "CHECK_EXPIRATION":
			// CHECK_EXPIRATION = ON/OFF
			optState := "On"
			if strings.ToUpper(p.curTok.Literal) == "OFF" {
				optState = "Off"
			}
			options = append(options, &ast.OnOffPrincipalOption{
				OptionKind:  "CheckExpiration",
				OptionState: optState,
			})
			p.nextToken()
		case "CHECK_POLICY":
			// CHECK_POLICY = ON/OFF
			optState := "On"
			if strings.ToUpper(p.curTok.Literal) == "OFF" {
				optState = "Off"
			}
			options = append(options, &ast.OnOffPrincipalOption{
				OptionKind:  "CheckPolicy",
				OptionState: optState,
			})
			p.nextToken()
		case "CREDENTIAL":
			options = append(options, &ast.IdentifierPrincipalOption{
				OptionKind: "Credential",
				Identifier: p.parseIdentifier(),
			})
		default:
			// Unknown option, skip value
			if p.curTok.Type != TokenComma && p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF {
				p.nextToken()
			}
		}

		// Check for comma (more options)
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	return options
}

func (p *Parser) parseCreateIndexStatement() (*ast.CreateIndexStatement, error) {
	stmt := &ast.CreateIndexStatement{
		Translated80SyntaxTo90: false,
	}

	// Parse optional UNIQUE
	if strings.ToUpper(p.curTok.Literal) == "UNIQUE" {
		stmt.Unique = true
		p.nextToken() // consume UNIQUE
	}

	// Parse optional CLUSTERED/NONCLUSTERED
	if strings.ToUpper(p.curTok.Literal) == "CLUSTERED" {
		clustered := true
		stmt.Clustered = &clustered
		p.nextToken()
	} else if strings.ToUpper(p.curTok.Literal) == "NONCLUSTERED" {
		clustered := false
		stmt.Clustered = &clustered
		p.nextToken()
	}

	// Consume INDEX keyword
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX
	}

	// Parse index name
	stmt.Name = p.parseIdentifier()

	// Parse ON table_name(columns)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		stmt.OnName, _ = p.parseSchemaObjectName()

		// Parse column list (columns with sort order)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := p.parseColumnWithSortOrder()
				stmt.Columns = append(stmt.Columns, col)

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

	// Parse INCLUDE (columns)
	if strings.ToUpper(p.curTok.Literal) == "INCLUDE" {
		p.nextToken() // consume INCLUDE
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				colRef := &ast.ColumnReferenceExpression{
					ColumnType: "Regular",
					MultiPartIdentifier: &ast.MultiPartIdentifier{
						Count:       1,
						Identifiers: []*ast.Identifier{p.parseIdentifier()},
					},
				}
				stmt.IncludeColumns = append(stmt.IncludeColumns, colRef)

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

	// Parse WHERE filter_predicate
	if strings.ToUpper(p.curTok.Literal) == "WHERE" {
		p.nextToken() // consume WHERE
		filterPred, err := p.parseBooleanExpression()
		if err == nil {
			stmt.FilterPredicate = filterPred
		}
	}

	// Parse WITH (index options)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		// Check if this is SQL 80 style (no parentheses) or modern style (with parentheses)
		if p.curTok.Type == TokenLParen {
			stmt.IndexOptions = p.parseCreateIndexOptions()
		} else {
			// SQL 80 style - no parentheses around options
			stmt.Translated80SyntaxTo90 = true
			stmt.IndexOptions = p.parseCreateIndexOptions80Style()
		}
	}

	// Parse ON filegroup/partition_scheme
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fg, _ := p.parseFileGroupOrPartitionScheme()
		stmt.OnFileGroupOrPartitionScheme = fg
	}

	// Parse FILESTREAM_ON filegroup
	if strings.ToUpper(p.curTok.Literal) == "FILESTREAM_ON" {
		p.nextToken() // consume FILESTREAM_ON
		value := p.curTok.Literal
		stmt.FileStreamOn = &ast.IdentifierOrValueExpression{
			Value: value,
			Identifier: &ast.Identifier{
				Value:     value,
				QuoteType: "NotQuoted",
			},
		}
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateIndexOptions() []ast.IndexOption {
	var options []ast.IndexOption

	// Expect (
	if p.curTok.Type != TokenLParen {
		return options
	}
	p.nextToken() // consume (

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken() // consume option name

		// Expect =
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		// Parse value
		valueToken := p.curTok
		valueStr := strings.ToUpper(valueToken.Literal)
		p.nextToken() // consume value

		switch optionName {
		case "PAD_INDEX":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "PadIndex",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "FILLFACTOR":
			options = append(options, &ast.IndexExpressionOption{
				OptionKind: "FillFactor",
				Expression: &ast.IntegerLiteral{LiteralType: "Integer", Value: valueToken.Literal},
			})
		case "IGNORE_DUP_KEY":
			opt := &ast.IgnoreDupKeyIndexOption{
				OptionKind:  "IgnoreDupKey",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			}
			// Check for optional (SUPPRESS_MESSAGES = ON/OFF)
			if valueStr == "ON" && p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				if strings.ToUpper(p.curTok.Literal) == "SUPPRESS_MESSAGES" {
					p.nextToken() // consume SUPPRESS_MESSAGES
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					suppressVal := strings.ToUpper(p.curTok.Literal) == "ON"
					opt.SuppressMessagesOption = &suppressVal
					p.nextToken() // consume ON/OFF
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			}
			options = append(options, opt)
		case "DROP_EXISTING":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "DropExisting",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "STATISTICS_NORECOMPUTE":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "StatisticsNoRecompute",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "SORT_IN_TEMPDB":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "SortInTempDB",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "ONLINE":
			onlineOpt := &ast.OnlineIndexOption{
				OptionKind:  "Online",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			}
			// Check for optional (WAIT_AT_LOW_PRIORITY (...))
			if valueStr == "ON" && p.curTok.Type == TokenLParen {
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
		case "ALLOW_ROW_LOCKS":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "AllowRowLocks",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "ALLOW_PAGE_LOCKS":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "AllowPageLocks",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "MAXDOP":
			options = append(options, &ast.IndexExpressionOption{
				OptionKind: "MaxDop",
				Expression: &ast.IntegerLiteral{LiteralType: "Integer", Value: valueToken.Literal},
			})
		case "MAX_DURATION":
			// Parse MAX_DURATION = value [MINUTES]
			opt := &ast.MaxDurationOption{
				OptionKind:  "MaxDuration",
				MaxDuration: &ast.IntegerLiteral{LiteralType: "Integer", Value: valueToken.Literal},
			}
			// Check for optional MINUTES unit
			if strings.ToUpper(p.curTok.Literal) == "MINUTES" {
				opt.Unit = "Minutes"
				p.nextToken() // consume MINUTES
			}
			options = append(options, opt)
		case "DATA_COMPRESSION":
			// Parse DATA_COMPRESSION = level [ON PARTITIONS(range)]
			compressionLevel := "None"
			switch valueStr {
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
			opt := &ast.DataCompressionOption{
				CompressionLevel: compressionLevel,
				OptionKind:       "DataCompression",
			}
			// Check for optional ON PARTITIONS(range)
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
					p.nextToken() // consume PARTITIONS
					if p.curTok.Type == TokenLParen {
						p.nextToken() // consume (
						for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
							partRange := &ast.CompressionPartitionRange{}
							// Parse From value
							partRange.From = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
							p.nextToken()
							// Check for TO keyword indicating a range
							if strings.ToUpper(p.curTok.Literal) == "TO" {
								p.nextToken() // consume TO
								partRange.To = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
								p.nextToken()
							}
							opt.PartitionRanges = append(opt.PartitionRanges, partRange)
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
			}
			options = append(options, opt)
		case "XML_COMPRESSION":
			// Parse XML_COMPRESSION = ON/OFF [ON PARTITIONS(range)]
			isCompressed := "On"
			if valueStr == "OFF" {
				isCompressed = "Off"
			}
			opt := &ast.XmlCompressionOption{
				IsCompressed: isCompressed,
				OptionKind:   "XmlCompression",
			}
			// Check for optional ON PARTITIONS(range)
			if p.curTok.Type == TokenOn {
				p.nextToken() // consume ON
				if strings.ToUpper(p.curTok.Literal) == "PARTITIONS" {
					p.nextToken() // consume PARTITIONS
					if p.curTok.Type == TokenLParen {
						p.nextToken() // consume (
						for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
							partRange := &ast.CompressionPartitionRange{}
							// Parse From value
							partRange.From = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
							p.nextToken()
							// Check for TO keyword indicating a range
							if strings.ToUpper(p.curTok.Literal) == "TO" {
								p.nextToken() // consume TO
								partRange.To = &ast.IntegerLiteral{LiteralType: "Integer", Value: p.curTok.Literal}
								p.nextToken()
							}
							opt.PartitionRanges = append(opt.PartitionRanges, partRange)
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
			}
			options = append(options, opt)
		default:
			// Generic handling for other options
			if valueStr == "ON" || valueStr == "OFF" {
				options = append(options, &ast.IndexStateOption{
					OptionKind:  p.getIndexOptionKind(optionName),
					OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
				})
			} else {
				options = append(options, &ast.IndexExpressionOption{
					OptionKind: p.getIndexOptionKind(optionName),
					Expression: &ast.IntegerLiteral{LiteralType: "Integer", Value: valueToken.Literal},
				})
			}
		}

		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	if p.curTok.Type == TokenRParen {
		p.nextToken() // consume )
	}

	return options
}

// parseCreateIndexOptions80Style parses index options without parentheses (SQL 80 style)
// e.g., WITH FILLFACTOR = 23, PAD_INDEX
func (p *Parser) parseCreateIndexOptions80Style() []ast.IndexOption {
	var options []ast.IndexOption

	for {
		// Check if current token could be an index option
		upper := strings.ToUpper(p.curTok.Literal)
		if !p.isIndexOption80Style(upper) {
			break
		}

		optionName := upper
		p.nextToken() // consume option name

		var valueStr string
		var valueToken Token

		// Check if there's an = value
		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
			valueToken = p.curTok
			valueStr = strings.ToUpper(valueToken.Literal)
			p.nextToken() // consume value
		} else {
			// No value means this is a flag option that is ON
			valueStr = "ON"
		}

		switch optionName {
		case "PAD_INDEX":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "PadIndex",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "FILLFACTOR":
			options = append(options, &ast.IndexExpressionOption{
				OptionKind: "FillFactor",
				Expression: &ast.IntegerLiteral{LiteralType: "Integer", Value: valueToken.Literal},
			})
		case "IGNORE_DUP_KEY":
			// In SQL 80 style, IGNORE_DUP_KEY uses IndexStateOption
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "IgnoreDupKey",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "DROP_EXISTING":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "DropExisting",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "STATISTICS_NORECOMPUTE":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "StatisticsNoRecompute",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		case "SORT_IN_TEMPDB":
			options = append(options, &ast.IndexStateOption{
				OptionKind:  "SortInTempDB",
				OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
			})
		default:
			// Generic handling for other options
			if valueStr == "ON" || valueStr == "OFF" {
				options = append(options, &ast.IndexStateOption{
					OptionKind:  p.getIndexOptionKind(optionName),
					OptionState: p.capitalizeFirst(strings.ToLower(valueStr)),
				})
			} else if valueToken.Type == TokenNumber || valueToken.Type != 0 {
				options = append(options, &ast.IndexExpressionOption{
					OptionKind: p.getIndexOptionKind(optionName),
					Expression: &ast.IntegerLiteral{LiteralType: "Integer", Value: valueToken.Literal},
				})
			}
		}

		// Check for comma to continue parsing options
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	return options
}

// isIndexOption80Style checks if a token could be an index option in SQL 80 style
func (p *Parser) isIndexOption80Style(name string) bool {
	switch name {
	case "PAD_INDEX", "FILLFACTOR", "IGNORE_DUP_KEY", "DROP_EXISTING",
		"STATISTICS_NORECOMPUTE", "SORT_IN_TEMPDB":
		return true
	default:
		return false
	}
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

	// Check for AUTHORIZATION clause
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Check for FROM clause (FILE, EXECUTABLE FILE, ASSEMBLY, PROVIDER)
	if p.curTok.Type == TokenFrom {
		p.nextToken() // consume FROM
		fromType := strings.ToUpper(p.curTok.Literal)
		switch fromType {
		case "FILE":
			p.nextToken() // consume FILE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			file, _ := p.parseStringLiteral()
			stmt.KeySource = &ast.FileEncryptionSource{
				IsExecutable: false,
				File:         file,
			}
			stmt.EncryptionAlgorithm = "None"
		case "EXECUTABLE":
			p.nextToken() // consume EXECUTABLE
			if strings.ToUpper(p.curTok.Literal) == "FILE" {
				p.nextToken() // consume FILE
			}
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			file, _ := p.parseStringLiteral()
			stmt.KeySource = &ast.FileEncryptionSource{
				IsExecutable: true,
				File:         file,
			}
			stmt.EncryptionAlgorithm = "None"
		case "ASSEMBLY":
			p.nextToken() // consume ASSEMBLY
			assemblyName := p.parseIdentifier()
			stmt.KeySource = &ast.AssemblyEncryptionSource{
				Assembly: assemblyName,
			}
			stmt.EncryptionAlgorithm = "None"
		case "PROVIDER":
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
						mappedAlg := p.mapEncryptionAlgorithm(alg)
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
						goto doneWithProviderOptions
					}

					if p.curTok.Type == TokenComma {
						p.nextToken() // consume comma
					} else if strings.ToUpper(p.curTok.Literal) != "ALGORITHM" &&
						strings.ToUpper(p.curTok.Literal) != "PROVIDER_KEY_NAME" &&
						strings.ToUpper(p.curTok.Literal) != "CREATION_DISPOSITION" {
						break
					}
				}
			doneWithProviderOptions:
			}
			stmt.KeySource = source
		}
	}

	// Check for WITH ALGORITHM = ... (without FROM clause)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		if strings.ToUpper(p.curTok.Literal) == "ALGORITHM" {
			p.nextToken() // consume ALGORITHM
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			alg := strings.ToUpper(p.curTok.Literal)
			stmt.EncryptionAlgorithm = p.mapEncryptionAlgorithm(alg)
			p.nextToken()
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

// mapEncryptionAlgorithm maps SQL encryption algorithm names to proper case
func (p *Parser) mapEncryptionAlgorithm(alg string) string {
	algMap := map[string]string{
		"DES":             "Des",
		"RC2":             "RC2",
		"RC4":             "RC4",
		"RC4_128":         "RC4_128",
		"TRIPLE_DES":      "TripleDes",
		"AES_128":         "Aes128",
		"AES_192":         "Aes192",
		"AES_256":         "Aes256",
		"RSA_512":         "Rsa512",
		"RSA_1024":        "Rsa1024",
		"RSA_2048":        "Rsa2048",
		"RSA_3072":        "Rsa3072",
		"RSA_4096":        "Rsa4096",
		"DESX":            "DesX",
		"TRIPLE_DES_3KEY": "TripleDes3Key",
	}
	if mapped, ok := algMap[alg]; ok {
		return mapped
	}
	return alg
}

func (p *Parser) parseCreateSymmetricKeyStatement() (*ast.CreateSymmetricKeyStatement, error) {
	p.nextToken() // consume SYMMETRIC
	if strings.ToUpper(p.curTok.Literal) == "KEY" {
		p.nextToken() // consume KEY
	}

	stmt := &ast.CreateSymmetricKeyStatement{
		Name: p.parseIdentifier(),
	}

	// Check for AUTHORIZATION clause
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
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
			algo := normalizeAlgorithmName(p.curTok.Literal)
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

		case "KEY_SOURCE":
			p.nextToken() // consume KEY_SOURCE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			passPhrase, _ := p.parseScalarExpression()
			opt := &ast.KeySourceKeyOption{
				PassPhrase: passPhrase,
				OptionKind: "KeySource",
			}
			options = append(options, opt)

		case "IDENTITY_VALUE":
			p.nextToken() // consume IDENTITY_VALUE
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			identityPhrase, _ := p.parseScalarExpression()
			opt := &ast.IdentityValueKeyOption{
				IdentityPhrase: identityPhrase,
				OptionKind:     "IdentityValue",
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

// normalizeAlgorithmName converts algorithm names to their canonical ScriptDom form.
func normalizeAlgorithmName(name string) string {
	switch strings.ToUpper(name) {
	case "DES":
		return "Des"
	case "TRIPLE_DES":
		return "TripleDes"
	case "TRIPLE_DES_3KEY":
		return "TripleDes3Key"
	case "RC2":
		return "RC2"
	case "RC4":
		return "RC4"
	case "RC4_128":
		return "RC4_128"
	case "DESX":
		return "Desx"
	case "AES_128":
		return "Aes128"
	case "AES_192":
		return "Aes192"
	case "AES_256":
		return "Aes256"
	case "RSA_512":
		return "RSA_512"
	case "RSA_1024":
		return "RSA_1024"
	case "RSA_2048":
		return "RSA_2048"
	case "RSA_3072":
		return "RSA_3072"
	case "RSA_4096":
		return "RSA_4096"
	default:
		return strings.ToUpper(name)
	}
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
		Name:                 p.parseIdentifier(),
		ActiveForBeginDialog: "NotSet",
	}

	// Optional AUTHORIZATION
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken()
		stmt.Owner = p.parseIdentifier()
	}

	// Optional ENCRYPTION BY PASSWORD
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
			if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
				strLit, _ := p.parseStringLiteral()
				stmt.EncryptionPassword = strLit
			}
		}
	}

	// Optional FROM clause
	if p.curTok.Type == TokenFrom {
		p.nextToken() // consume FROM
		sourceType := strings.ToUpper(p.curTok.Literal)

		if sourceType == "ASSEMBLY" {
			p.nextToken() // consume ASSEMBLY
			stmt.CertificateSource = &ast.AssemblyEncryptionSource{
				Assembly: p.parseIdentifier(),
			}
		} else if sourceType == "FILE" || sourceType == "EXECUTABLE" {
			isExecutable := false
			if sourceType == "EXECUTABLE" {
				isExecutable = true
				p.nextToken() // consume EXECUTABLE
				// Next should be FILE
			}
			if strings.ToUpper(p.curTok.Literal) == "FILE" {
				p.nextToken() // consume FILE
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
					strLit, _ := p.parseStringLiteral()
					stmt.CertificateSource = &ast.FileEncryptionSource{
						IsExecutable: isExecutable,
						File:         strLit,
					}
				}
			}
		}
	}

	// Parse WITH clauses (can appear multiple times for different purposes)
	for p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		// Check if it's PRIVATE KEY or certificate options
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "PRIVATE" {
			p.nextToken() // consume PRIVATE
			if strings.ToUpper(p.curTok.Literal) == "KEY" {
				p.nextToken() // consume KEY
			}
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optName := strings.ToUpper(p.curTok.Literal)
					p.nextToken() // consume option name

					switch optName {
					case "FILE":
						if p.curTok.Type == TokenEquals {
							p.nextToken()
						}
						if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
							strLit, _ := p.parseStringLiteral()
							stmt.PrivateKeyPath = strLit
						}
					case "DECRYPTION":
						// DECRYPTION BY PASSWORD
						if strings.ToUpper(p.curTok.Literal) == "BY" {
							p.nextToken()
						}
						if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
							p.nextToken()
							if p.curTok.Type == TokenEquals {
								p.nextToken()
							}
							if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
								strLit, _ := p.parseStringLiteral()
								stmt.DecryptionPassword = strLit
							}
						}
					case "ENCRYPTION":
						// ENCRYPTION BY PASSWORD
						if strings.ToUpper(p.curTok.Literal) == "BY" {
							p.nextToken()
						}
						if strings.ToUpper(p.curTok.Literal) == "PASSWORD" {
							p.nextToken()
							if p.curTok.Type == TokenEquals {
								p.nextToken()
							}
							if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
								strLit, _ := p.parseStringLiteral()
								stmt.EncryptionPassword = strLit
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
		} else {
			// Certificate options: SUBJECT, START_DATE, EXPIRY_DATE
			for {
				optName := strings.ToUpper(p.curTok.Literal)
				if optName != "SUBJECT" && optName != "START_DATE" && optName != "EXPIRY_DATE" {
					break
				}
				p.nextToken() // consume option name
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
					strLit, _ := p.parseStringLiteral()
					kind := ""
					switch optName {
					case "SUBJECT":
						kind = "Subject"
					case "START_DATE":
						kind = "StartDate"
					case "EXPIRY_DATE":
						kind = "ExpiryDate"
					}
					stmt.CertificateOptions = append(stmt.CertificateOptions, &ast.CertificateOption{
						Kind:  kind,
						Value: strLit,
					})
				}
				if p.curTok.Type == TokenComma {
					p.nextToken()
				} else {
					break
				}
			}
		}
	}

	// Optional ACTIVE FOR BEGIN_DIALOG
	if strings.ToUpper(p.curTok.Literal) == "ACTIVE" {
		p.nextToken() // consume ACTIVE
		if strings.ToUpper(p.curTok.Literal) == "FOR" {
			p.nextToken() // consume FOR
		}
		if strings.ToUpper(p.curTok.Literal) == "BEGIN_DIALOG" {
			p.nextToken() // consume BEGIN_DIALOG
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}
			if strings.ToUpper(p.curTok.Literal) == "ON" {
				stmt.ActiveForBeginDialog = "On"
				p.nextToken()
			} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
				stmt.ActiveForBeginDialog = "Off"
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

	// Check for ON clause (filegroup)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fg, err := p.parseIdentifierOrValueExpression()
		if err != nil {
			return nil, err
		}
		stmt.OnFileGroup = fg
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

	// Check for ON clause after WITH (alternative syntax)
	if p.curTok.Type == TokenOn {
		p.nextToken() // consume ON
		fg, err := p.parseIdentifierOrValueExpression()
		if err != nil {
			return nil, err
		}
		stmt.OnFileGroup = fg
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

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
				// Parse activation options
				activationOpts, err := p.parseActivationOptions()
				if err != nil {
					return nil, err
				}
				options = append(options, activationOpts...)
				if p.curTok.Type != TokenRParen {
					return nil, fmt.Errorf("expected ) after ACTIVATION options, got %s", p.curTok.Literal)
				}
				p.nextToken() // consume )
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

func (p *Parser) parseActivationOptions() ([]ast.QueueOption, error) {
	var options []ast.QueueOption

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
		optName := strings.ToUpper(p.curTok.Literal)
		switch optName {
		case "STATUS":
			p.nextToken() // consume STATUS
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			state := strings.ToUpper(p.curTok.Literal)
			p.nextToken() // consume ON/OFF
			opt := &ast.QueueStateOption{
				OptionState: capitalizeFirst(state),
				OptionKind:  "ActivationStatus",
			}
			options = append(options, opt)

		case "PROCEDURE_NAME":
			p.nextToken() // consume PROCEDURE_NAME
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			procName, _ := p.parseSchemaObjectName()
			opt := &ast.QueueProcedureOption{
				OptionValue: procName,
				OptionKind:  "ActivationProcedureName",
			}
			options = append(options, opt)

		case "MAX_QUEUE_READERS":
			p.nextToken() // consume MAX_QUEUE_READERS
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			value, _ := p.parseScalarExpression()
			opt := &ast.QueueValueOption{
				OptionValue: value,
				OptionKind:  "ActivationMaxQueueReaders",
			}
			options = append(options, opt)

		case "EXECUTE":
			p.nextToken() // consume EXECUTE
			// Expect AS
			if strings.ToUpper(p.curTok.Literal) == "AS" {
				p.nextToken() // consume AS
			}
			execAs := &ast.ExecuteAsClause{}
			// Check for SELF, OWNER, or string
			execVal := strings.ToUpper(p.curTok.Literal)
			switch execVal {
			case "SELF":
				execAs.ExecuteAsOption = "Self"
				p.nextToken()
			case "OWNER":
				execAs.ExecuteAsOption = "Owner"
				p.nextToken()
			default:
				// String literal - 'username'
				if p.curTok.Type == TokenString {
					value := p.curTok.Literal
					// Remove quotes
					if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
						value = value[1 : len(value)-1]
					}
					execAs.ExecuteAsOption = "String"
					execAs.Literal = &ast.StringLiteral{
						LiteralType:   "String",
						IsNational:    false,
						IsLargeObject: false,
						Value:         value,
					}
					p.nextToken()
				}
			}
			opt := &ast.QueueExecuteAsOption{
				OptionValue: execAs,
				OptionKind:  "ActivationExecuteAs",
			}
			options = append(options, opt)

		default:
			return options, nil
		}

		// Check for comma separator
		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	return options, nil
}

func (p *Parser) parseCreateRouteStatement() (*ast.CreateRouteStatement, error) {
	p.nextToken() // consume ROUTE

	stmt := &ast.CreateRouteStatement{
		Name: p.parseIdentifier(),
	}

	// Parse optional AUTHORIZATION clause
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

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

func (p *Parser) parseRouteOptions() []*ast.RouteOption {
	var options []*ast.RouteOption

	for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF {
		optionName := strings.ToUpper(p.curTok.Literal)
		p.nextToken() // consume option name

		if p.curTok.Type == TokenEquals {
			p.nextToken() // consume =
		}

		var optionKind string
		switch optionName {
		case "BROKER_INSTANCE":
			optionKind = "BrokerInstance"
		case "SERVICE_NAME":
			optionKind = "ServiceName"
		case "LIFETIME":
			optionKind = "Lifetime"
		case "ADDRESS":
			optionKind = "Address"
		case "MIRROR_ADDRESS":
			optionKind = "MirrorAddress"
		default:
			// Unknown option, skip
			if p.curTok.Type == TokenComma {
				p.nextToken()
			}
			continue
		}

		// Parse literal value
		var literal ast.ScalarExpression
		if p.curTok.Type == TokenString {
			value := p.curTok.Literal
			// Strip quotes
			if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
				value = value[1 : len(value)-1]
			}
			literal = &ast.StringLiteral{
				LiteralType: "String",
				Value:       value,
			}
			p.nextToken()
		} else if p.curTok.Type == TokenNumber {
			literal = &ast.IntegerLiteral{
				LiteralType: "Integer",
				Value:       p.curTok.Literal,
			}
			p.nextToken()
		} else {
			// Unknown value, try to skip
			p.nextToken()
		}

		if literal != nil {
			options = append(options, &ast.RouteOption{
				OptionKind: optionKind,
				Literal:    literal,
			})
		}

		// Skip comma if present
		if p.curTok.Type == TokenComma {
			p.nextToken()
		} else {
			break
		}
	}

	return options
}

func (p *Parser) parseCreateEndpointStatement() (*ast.CreateEndpointStatement, error) {
	p.nextToken() // consume ENDPOINT

	stmt := &ast.CreateEndpointStatement{
		Name: p.parseIdentifier(),
	}
	hasOptions := false

	// Check for AUTHORIZATION immediately after name
	if p.curTok.Type == TokenAuthorization {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Parse endpoint options (STATE, AFFINITY, AS, FOR)
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
			protocolUpper := strings.ToUpper(p.curTok.Literal)
			switch protocolUpper {
			case "TCP":
				stmt.Protocol = "Tcp"
			case "HTTP":
				stmt.Protocol = "Http"
			}
			p.nextToken()
			// Parse protocol options
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
					optName := strings.ToUpper(p.curTok.Literal)
					p.nextToken()
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					if optName == "LISTENER_IP" {
						ipOpt := &ast.ListenerIPEndpointProtocolOption{
							Kind: "TcpListenerIP",
						}
						if strings.ToUpper(p.curTok.Literal) == "ALL" {
							ipOpt.IsAll = true
							p.nextToken()
						} else if p.curTok.Type == TokenLParen {
							p.nextToken() // consume (
							ipOpt.IPv4PartOne = p.parseIPv4Address()
							// Check for colon-separated second IP address
							if p.curTok.Type == TokenColon {
								p.nextToken() // consume :
								ipOpt.IPv4PartTwo = p.parseIPv4Address()
							}
							if p.curTok.Type == TokenRParen {
								p.nextToken() // consume )
							}
						}
						stmt.ProtocolOptions = append(stmt.ProtocolOptions, ipOpt)
					} else {
						opt := &ast.LiteralEndpointProtocolOption{}
						switch optName {
						case "LISTENER_PORT":
							opt.Kind = "TcpListenerPort"
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
					}
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
			endpointTypeUpper := strings.ToUpper(p.curTok.Literal)
			switch endpointTypeUpper {
			case "SOAP":
				stmt.EndpointType = "Soap"
			case "SERVICE_BROKER":
				stmt.EndpointType = "ServiceBroker"
			case "DATABASE_MIRRORING":
				stmt.EndpointType = "DatabaseMirroring"
			case "TSQL":
				stmt.EndpointType = "TSql"
			default:
				stmt.EndpointType = endpointTypeUpper
			}
			p.nextToken()
			// Parse payload options
			if p.curTok.Type == TokenLParen {
				p.nextToken() // consume (
				// Empty parentheses are ok
				if p.curTok.Type == TokenRParen {
					p.nextToken()
				}
			}

		case ",":
			p.nextToken()

		case "GO":
			// End of statement
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

		default:
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
	}

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

func (p *Parser) parseCreateAssemblyStatement() (*ast.CreateAssemblyStatement, error) {
	p.nextToken() // consume ASSEMBLY

	stmt := &ast.CreateAssemblyStatement{
		Name: p.parseIdentifier(),
	}

	// Check for AUTHORIZATION clause
	if p.curTok.Type == TokenAuthorization {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Parse FROM clause
	if strings.ToUpper(p.curTok.Literal) == "FROM" {
		p.nextToken() // consume FROM
		// Parse list of expressions (variable references, string literals, binary expressions)
		for {
			expr, err := p.parseScalarExpression()
			if err != nil {
				break
			}
			stmt.Parameters = append(stmt.Parameters, expr)
			if p.curTok.Type == TokenComma {
				p.nextToken()
			} else {
				break
			}
		}
	}

	// Parse WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH
		// Parse PERMISSION_SET = value
		if strings.ToUpper(p.curTok.Literal) == "PERMISSION_SET" {
			p.nextToken() // consume PERMISSION_SET
			if p.curTok.Type == TokenEquals {
				p.nextToken() // consume =
			}
			option := &ast.PermissionSetAssemblyOption{
				OptionKind: "PermissionSet",
			}
			switch strings.ToUpper(p.curTok.Literal) {
			case "SAFE":
				option.PermissionSetOption = "Safe"
			case "EXTERNAL_ACCESS":
				option.PermissionSetOption = "ExternalAccess"
			case "UNSAFE":
				option.PermissionSetOption = "Unsafe"
			}
			p.nextToken()
			stmt.Options = append(stmt.Options, option)
		}
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

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
	case "STOPLIST":
		return p.parseCreateFulltextStopListStatement()
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

		// Parse optional (column_list)
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				col := &ast.FullTextIndexColumn{}
				col.Name = p.parseIdentifier()

				// Parse optional TYPE COLUMN type_column_name
				if strings.ToUpper(p.curTok.Literal) == "TYPE" {
					p.nextToken() // consume TYPE
					if strings.ToUpper(p.curTok.Literal) == "COLUMN" {
						p.nextToken() // consume COLUMN
					}
					col.TypeColumn = p.parseIdentifier()
				}

				// Parse optional LANGUAGE language_term
				if p.curTok.Type == TokenLanguage {
					p.nextToken() // consume LANGUAGE
					col.LanguageTerm = &ast.IdentifierOrValueExpression{}
					if p.curTok.Type == TokenString {
						strLit, _ := p.parseStringLiteral()
						col.LanguageTerm.Value = strLit.Value
						col.LanguageTerm.ValueExpression = strLit
					} else if p.curTok.Type == TokenNumber {
						// Check for hex literal (0x...)
						if strings.HasPrefix(strings.ToLower(p.curTok.Literal), "0x") {
							lit := &ast.BinaryLiteral{
								LiteralType:   "Binary",
								IsLargeObject: false,
								Value:         p.curTok.Literal,
							}
							col.LanguageTerm.Value = p.curTok.Literal
							col.LanguageTerm.ValueExpression = lit
						} else {
							// Parse integer literal directly
							lit := &ast.IntegerLiteral{
								LiteralType: "Integer",
								Value:       p.curTok.Literal,
							}
							col.LanguageTerm.Value = p.curTok.Literal
							col.LanguageTerm.ValueExpression = lit
						}
						p.nextToken()
					} else if p.curTok.Type == TokenBinary {
						// Handle binary/hex literal
						lit := &ast.BinaryLiteral{
							LiteralType:   "Binary",
							IsLargeObject: false,
							Value:         p.curTok.Literal,
						}
						col.LanguageTerm.Value = p.curTok.Literal
						col.LanguageTerm.ValueExpression = lit
						p.nextToken()
					} else {
						col.LanguageTerm.Identifier = p.parseIdentifier()
						col.LanguageTerm.Value = col.LanguageTerm.Identifier.Value
					}
				}

				// Parse optional STATISTICAL_SEMANTICS
				if strings.ToUpper(p.curTok.Literal) == "STATISTICAL_SEMANTICS" {
					col.StatisticalSemantics = true
					p.nextToken()
				}

				stmt.FullTextIndexColumns = append(stmt.FullTextIndexColumns, col)

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

		// Parse KEY INDEX name
		if strings.ToUpper(p.curTok.Literal) == "KEY" {
			p.nextToken() // consume KEY
			if strings.ToUpper(p.curTok.Literal) == "INDEX" {
				p.nextToken() // consume INDEX
			}
			stmt.KeyIndexName = p.parseIdentifier()
		}

		// Parse ON clause for catalog/filegroup
		if p.curTok.Type == TokenOn {
			p.nextToken() // consume ON
			stmt.CatalogAndFileGroup = &ast.FullTextCatalogAndFileGroup{}

			if p.curTok.Type == TokenLParen {
				// (FILEGROUP fg, catalog) or (catalog, FILEGROUP fg) format
				p.nextToken() // consume (

				// Check first element
				if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
					p.nextToken() // consume FILEGROUP
					stmt.CatalogAndFileGroup.FileGroupName = p.parseIdentifier()
					stmt.CatalogAndFileGroup.FileGroupIsFirst = true

					// Check for comma and catalog
					if p.curTok.Type == TokenComma {
						p.nextToken() // consume comma
						stmt.CatalogAndFileGroup.CatalogName = p.parseIdentifier()
					}
				} else {
					// It's a catalog name first
					stmt.CatalogAndFileGroup.CatalogName = p.parseIdentifier()
					stmt.CatalogAndFileGroup.FileGroupIsFirst = false

					// Check for comma and filegroup
					if p.curTok.Type == TokenComma {
						p.nextToken() // consume comma
						if strings.ToUpper(p.curTok.Literal) == "FILEGROUP" {
							p.nextToken() // consume FILEGROUP
						}
						stmt.CatalogAndFileGroup.FileGroupName = p.parseIdentifier()
					}
				}

				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
			} else {
				// Just a catalog name without parentheses
				stmt.CatalogAndFileGroup.CatalogName = p.parseIdentifier()
				stmt.CatalogAndFileGroup.FileGroupIsFirst = false
			}
		}

		// Parse WITH clause
		if p.curTok.Type == TokenWith {
			p.nextToken() // consume WITH

			// Handle optional parentheses: WITH (option, option) vs WITH option
			hasParen := false
			if p.curTok.Type == TokenLParen {
				hasParen = true
				p.nextToken() // consume (
			}

			noPopulation := false
			for {
				optLit := strings.ToUpper(p.curTok.Literal)
				if optLit == "CHANGE_TRACKING" {
					p.nextToken() // consume CHANGE_TRACKING
					// Handle optional = sign
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					var trackingValue string
					if strings.ToUpper(p.curTok.Literal) == "MANUAL" {
						trackingValue = "Manual"
						p.nextToken()
					} else if strings.ToUpper(p.curTok.Literal) == "AUTO" {
						trackingValue = "Auto"
						p.nextToken()
					} else if strings.ToUpper(p.curTok.Literal) == "OFF" {
						trackingValue = "Off"
						p.nextToken()
					}
					// If we see NO POPULATION after CHANGE_TRACKING OFF, update the value
					if trackingValue == "Off" && noPopulation {
						trackingValue = "OffNoPopulation"
					}
					stmt.Options = append(stmt.Options, &ast.ChangeTrackingFullTextIndexOption{
						Value:      trackingValue,
						OptionKind: "ChangeTracking",
					})
				} else if optLit == "STOPLIST" {
					p.nextToken() // consume STOPLIST
					// Handle optional = sign
					if p.curTok.Type == TokenEquals {
						p.nextToken() // consume =
					}
					opt := &ast.StopListFullTextIndexOption{
						OptionKind: "StopList",
					}
					if strings.ToUpper(p.curTok.Literal) == "OFF" {
						opt.IsOff = true
						p.nextToken()
					} else if strings.ToUpper(p.curTok.Literal) == "SYSTEM" {
						opt.IsOff = false
						opt.StopListName = p.parseIdentifier()
					} else {
						opt.IsOff = false
						opt.StopListName = p.parseIdentifier()
					}
					stmt.Options = append(stmt.Options, opt)
				} else if optLit == "NO" {
					p.nextToken() // consume NO
					if strings.ToUpper(p.curTok.Literal) == "POPULATION" {
						p.nextToken() // consume POPULATION
						noPopulation = true
						// Update CHANGE_TRACKING OFF to OffNoPopulation
						for i, opt := range stmt.Options {
							if ctOpt, ok := opt.(*ast.ChangeTrackingFullTextIndexOption); ok && ctOpt.Value == "Off" {
								ctOpt.Value = "OffNoPopulation"
								stmt.Options[i] = ctOpt
							}
						}
					}
				} else if hasParen && p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
					break
				} else {
					break
				}

				if p.curTok.Type == TokenComma {
					p.nextToken() // consume comma
				} else if p.curTok.Type == TokenSemicolon || p.curTok.Type == TokenEOF {
					break
				} else if hasParen && p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
					break
				}
			}
		}

		// Skip optional semicolon
		if p.curTok.Type == TokenSemicolon {
			p.nextToken()
		}
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

func (p *Parser) parseCreateFulltextStopListStatement() (*ast.CreateFullTextStopListStatement, error) {
	p.nextToken() // consume STOPLIST

	stmt := &ast.CreateFullTextStopListStatement{
		Name:             p.parseIdentifier(),
		IsSystemStopList: false,
	}

	// Parse FROM clause
	if p.curTok.Type == TokenFrom {
		p.nextToken() // consume FROM

		// Check for SYSTEM STOPLIST
		if strings.ToUpper(p.curTok.Literal) == "SYSTEM" {
			p.nextToken() // consume SYSTEM
			if strings.ToUpper(p.curTok.Literal) == "STOPLIST" {
				p.nextToken() // consume STOPLIST
			}
			stmt.IsSystemStopList = true
		} else {
			// Parse schema.name or just name
			first := p.parseIdentifier()
			if p.curTok.Type == TokenDot {
				p.nextToken() // consume .
				stmt.DatabaseName = first
				stmt.SourceStopListName = p.parseIdentifier()
			} else {
				stmt.SourceStopListName = first
			}
		}
	}

	// Parse AUTHORIZATION clause
	if strings.ToUpper(p.curTok.Literal) == "AUTHORIZATION" {
		p.nextToken() // consume AUTHORIZATION
		stmt.Owner = p.parseIdentifier()
	}

	// Skip optional semicolon
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}

	return stmt, nil
}

func (p *Parser) parseCreateFulltextCatalogStatement() (*ast.CreateFullTextCatalogStatement, error) {
	p.nextToken() // consume CATALOG

	stmt := &ast.CreateFullTextCatalogStatement{
		Name: p.parseIdentifier(),
	}

	// Parse optional clauses
	for p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon && !p.isBatchSeparator() && !p.isStatementTerminator() {
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

	// Parse optional WHERE clause (filter predicate)
	if p.curTok.Type == TokenWhere {
		p.nextToken() // consume WHERE
		pred, err := p.parseBooleanExpression()
		if err != nil {
			return nil, err
		}
		stmt.FilterPredicate = pred
	}

	// Parse optional WITH clause (reuse UPDATE STATISTICS options logic)
	if p.curTok.Type == TokenWith {
		p.nextToken() // consume WITH

		for p.curTok.Type != TokenSemicolon && p.curTok.Type != TokenEOF {
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
			case "STATS_STREAM":
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}
				value, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.LiteralStatisticsOption{
					OptionKind: "StatsStream",
					Literal:    value,
				})
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
			case "AUTO_DROP":
				if p.curTok.Type == TokenEquals {
					p.nextToken()
					state := strings.ToUpper(p.curTok.Literal)
					optionState := "On"
					if state == "OFF" {
						optionState = "Off"
					}
					p.nextToken()
					stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.OnOffStatisticsOption{
						OptionKind:  "AutoDrop",
						OptionState: optionState,
					})
				} else {
					stmt.StatisticsOptions = append(stmt.StatisticsOptions, &ast.OnOffStatisticsOption{
						OptionKind:  "AutoDrop",
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

	// Skip any remaining tokens
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
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
				// Parse optional WITH clause for table options
				if p.curTok.Type == TokenWith {
					p.nextToken() // consume WITH
					if p.curTok.Type == TokenLParen {
						p.nextToken() // consume (
						// Parse options
						for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
							optUpper := strings.ToUpper(p.curTok.Literal)
							if optUpper == "MEMORY_OPTIMIZED" {
								p.nextToken() // consume MEMORY_OPTIMIZED
								if p.curTok.Type == TokenEquals {
									p.nextToken() // consume =
								}
								stateUpper := strings.ToUpper(p.curTok.Literal)
								state := "On"
								if stateUpper == "OFF" {
									state = "Off"
								}
								p.nextToken() // consume ON/OFF
								stmt.Options = append(stmt.Options, &ast.MemoryOptimizedTableOption{
									OptionKind:  "MemoryOptimized",
									OptionState: state,
								})
							} else {
								// Skip unknown option
								p.nextToken()
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
					}
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
		Primary:               true,
		SecondaryXmlIndexType: "NotSpecified",
		Name:                  p.parseIdentifier(),
	}

	// Parse ON table_name
	if strings.ToUpper(p.curTok.Literal) == "ON" {
		p.nextToken() // consume ON
		stmt.OnName, _ = p.parseSchemaObjectName()
	}

	// Parse (column)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		stmt.XmlColumn = p.parseIdentifier()
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse WITH (options) if present
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			// parseCreateIndexOptions expects to consume ( and ) itself
			stmt.IndexOptions = p.parseCreateIndexOptions()
		}
	}

	return stmt, nil
}

func (p *Parser) parseCreateXmlIndexFromXml() (ast.Statement, error) {
	// XML has already been consumed, curTok is INDEX
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX
	}

	name := p.parseIdentifier()
	var onName *ast.SchemaObjectName
	var xmlColumn *ast.Identifier

	// Parse ON table_name
	if strings.ToUpper(p.curTok.Literal) == "ON" {
		p.nextToken() // consume ON
		onName, _ = p.parseSchemaObjectName()
	}

	// Parse (column)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		xmlColumn = p.parseIdentifier()
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse USING XML INDEX name
	if strings.ToUpper(p.curTok.Literal) == "USING" {
		p.nextToken() // consume USING
		if strings.ToUpper(p.curTok.Literal) == "XML" {
			p.nextToken() // consume XML
		}
		if p.curTok.Type == TokenIndex {
			p.nextToken() // consume INDEX
		}
		usingName := p.parseIdentifier()
		if strings.ToUpper(p.curTok.Literal) == "FOR" {
			p.nextToken() // consume FOR
			// Check if this is a selective XML index (FOR followed by parenthesis with path names)
			// vs regular secondary XML index (FOR followed by VALUE|PATH|PROPERTY)
			if p.curTok.Type == TokenLParen {
				// This is a secondary selective XML index
				selectiveStmt := &ast.CreateSelectiveXmlIndexStatement{
					Name:              name,
					OnName:            onName,
					XmlColumn:         xmlColumn,
					IsSecondary:       true,
					UsingXmlIndexName: usingName,
				}
				p.nextToken() // consume (
				// Parse path name(s)
				if p.curTok.Type == TokenIdent {
					selectiveStmt.PathName = p.parseIdentifier()
				}
				if p.curTok.Type == TokenRParen {
					p.nextToken() // consume )
				}
				return selectiveStmt, nil
			}
			// Regular secondary XML index
			stmt := &ast.CreateXmlIndexStatement{
				Primary:               false,
				SecondaryXmlIndexType: "NotSpecified",
				Name:                  name,
				OnName:                onName,
				XmlColumn:             xmlColumn,
				SecondaryXmlIndexName: usingName,
			}
			switch strings.ToUpper(p.curTok.Literal) {
			case "VALUE":
				stmt.SecondaryXmlIndexType = "Value"
				p.nextToken()
			case "PATH":
				stmt.SecondaryXmlIndexType = "Path"
				p.nextToken()
			case "PROPERTY":
				stmt.SecondaryXmlIndexType = "Property"
				p.nextToken()
			}
			// Parse WITH (options) if present
			if strings.ToUpper(p.curTok.Literal) == "WITH" {
				p.nextToken() // consume WITH
				if p.curTok.Type == TokenLParen {
					stmt.IndexOptions = p.parseCreateIndexOptions()
				}
			}
			return stmt, nil
		}
	}

	// Non-secondary XML index
	stmt := &ast.CreateXmlIndexStatement{
		Primary:               false,
		SecondaryXmlIndexType: "NotSpecified",
		Name:                  name,
		OnName:                onName,
		XmlColumn:             xmlColumn,
	}

	// Parse WITH (options) if present
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			// parseCreateIndexOptions expects to consume ( and ) itself
			stmt.IndexOptions = p.parseCreateIndexOptions()
		}
	}

	return stmt, nil
}

func (p *Parser) parseCreateSelectiveXmlIndexStatement() (*ast.CreateSelectiveXmlIndexStatement, error) {
	// SELECTIVE has already been matched, consume it
	p.nextToken() // consume SELECTIVE
	if strings.ToUpper(p.curTok.Literal) == "XML" {
		p.nextToken() // consume XML
	}
	if p.curTok.Type == TokenIndex {
		p.nextToken() // consume INDEX
	}

	stmt := &ast.CreateSelectiveXmlIndexStatement{
		IsSecondary: false,
		Name:        p.parseIdentifier(),
	}

	// Parse ON table_name
	if strings.ToUpper(p.curTok.Literal) == "ON" {
		p.nextToken() // consume ON
		stmt.OnName, _ = p.parseSchemaObjectName()
	}

	// Parse (column)
	if p.curTok.Type == TokenLParen {
		p.nextToken() // consume (
		stmt.XmlColumn = p.parseIdentifier()
		if p.curTok.Type == TokenRParen {
			p.nextToken() // consume )
		}
	}

	// Parse optional WITH XMLNAMESPACES clause
	if strings.ToUpper(p.curTok.Literal) == "WITH" && strings.ToUpper(p.peekTok.Literal) == "XMLNAMESPACES" {
		p.nextToken() // consume WITH
		stmt.XmlNamespaces = p.parseXmlNamespaces()
	}

	// Parse FOR clause with paths
	if strings.ToUpper(p.curTok.Literal) == "FOR" {
		p.nextToken() // consume FOR
		if p.curTok.Type == TokenLParen {
			p.nextToken() // consume (
			for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF {
				path := p.parseSelectiveXmlIndexPath()
				if path != nil {
					stmt.PromotedPaths = append(stmt.PromotedPaths, path)
				}
				if p.curTok.Type == TokenComma {
					p.nextToken() // consume ,
				} else {
					break
				}
			}
			if p.curTok.Type == TokenRParen {
				p.nextToken() // consume )
			}
		}
	}

	// Parse WITH (options) if present
	if strings.ToUpper(p.curTok.Literal) == "WITH" {
		p.nextToken() // consume WITH
		if p.curTok.Type == TokenLParen {
			stmt.IndexOptions = p.parseCreateIndexOptions()
		}
	}

	return stmt, nil
}

func (p *Parser) parseSelectiveXmlIndexPath() *ast.SelectiveXmlIndexPromotedPath {
	path := &ast.SelectiveXmlIndexPromotedPath{}

	// Parse path name (identifier)
	path.Name = p.parseIdentifier()

	// Check for = 'path_value'
	if p.curTok.Type == TokenEquals {
		p.nextToken() // consume =
		if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
			path.Path, _ = p.parseStringLiteral()
		}
	}

	// Parse optional AS XQUERY/SQL clause
	if p.curTok.Type == TokenAs {
		p.nextToken() // consume AS
		upperLit := strings.ToUpper(p.curTok.Literal)
		if upperLit == "XQUERY" {
			p.nextToken() // consume XQUERY
			// Check for optional type or MAXLENGTH
			if p.curTok.Type == TokenString || p.curTok.Type == TokenNationalString {
				// XQuery type like 'xs:string' or 'node()'
				path.XQueryDataType, _ = p.parseStringLiteral()
			}
			// Check for MAXLENGTH
			if strings.ToUpper(p.curTok.Literal) == "MAXLENGTH" {
				p.nextToken() // consume MAXLENGTH
				if p.curTok.Type == TokenLParen {
					p.nextToken() // consume (
					if p.curTok.Type == TokenNumber {
						path.MaxLength = &ast.IntegerLiteral{
							LiteralType: "Integer",
							Value:       p.curTok.Literal,
						}
						p.nextToken() // consume number
					}
					if p.curTok.Type == TokenRParen {
						p.nextToken() // consume )
					}
				}
			}
			// Check for SINGLETON
			if strings.ToUpper(p.curTok.Literal) == "SINGLETON" {
				path.IsSingleton = true
				p.nextToken() // consume SINGLETON
			}
		} else if upperLit == "SQL" {
			p.nextToken() // consume SQL
			// Parse SQL data type
			dt, _ := p.parseDataTypeReference()
			if sdt, ok := dt.(*ast.SqlDataTypeReference); ok {
				path.SQLDataType = sdt
			}
			// Check for SINGLETON
			if strings.ToUpper(p.curTok.Literal) == "SINGLETON" {
				path.IsSingleton = true
				p.nextToken() // consume SINGLETON
			}
		}
	}

	return path
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

	// Parse SELECT statement (may have WITH clause for CTEs)
	var selectStmt *ast.SelectStatement
	if p.curTok.Type == TokenWith {
		// Parse WITH + SELECT statement
		withStmt, err := p.parseWithStatement()
		if err != nil {
			return nil, err
		}
		if sel, ok := withStmt.(*ast.SelectStatement); ok {
			selectStmt = sel
		} else {
			return nil, fmt.Errorf("expected SELECT statement after WITH in cursor definition")
		}
	} else {
		var err error
		selectStmt, err = p.parseSelectStatement()
		if err != nil {
			return nil, err
		}
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

// parseEndConversationStatement parses END CONVERSATION statements
func (p *Parser) parseEndConversationStatement() (*ast.EndConversationStatement, error) {
	// Consume END
	p.nextToken()

	// Expect CONVERSATION
	if strings.ToUpper(p.curTok.Literal) != "CONVERSATION" {
		return nil, fmt.Errorf("expected CONVERSATION after END, got %s", p.curTok.Literal)
	}
	p.nextToken()

	stmt := &ast.EndConversationStatement{}

	// Parse the conversation handle expression
	expr, err := p.parseScalarExpression()
	if err != nil {
		return nil, err
	}
	stmt.Conversation = expr

	// Check for WITH clause
	if p.curTok.Type == TokenWith {
		p.nextToken()

		if strings.ToUpper(p.curTok.Literal) == "CLEANUP" {
			stmt.WithCleanup = true
			p.nextToken()
		} else if strings.ToUpper(p.curTok.Literal) == "ERROR" {
			p.nextToken()

			// Expect =
			if p.curTok.Type == TokenEquals {
				p.nextToken()
			}

			// Parse error code
			errCode, err := p.parseScalarExpression()
			if err != nil {
				return nil, err
			}
			stmt.ErrorCode = errCode

			// Expect DESCRIPTION
			if strings.ToUpper(p.curTok.Literal) == "DESCRIPTION" {
				p.nextToken()

				// Expect =
				if p.curTok.Type == TokenEquals {
					p.nextToken()
				}

				// Parse error description
				errDesc, err := p.parseScalarExpression()
				if err != nil {
					return nil, err
				}
				stmt.ErrorDescription = errDesc
			}
		}
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

func (p *Parser) parseCreateBrokerPriorityStatement() (*ast.CreateBrokerPriorityStatement, error) {
	// Consume BROKER
	p.nextToken()

	// Consume PRIORITY
	if strings.ToUpper(p.curTok.Literal) == "PRIORITY" {
		p.nextToken()
	}

	stmt := &ast.CreateBrokerPriorityStatement{}

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

func (p *Parser) parseBrokerPriorityParameters() []*ast.BrokerPriorityParameter {
	var params []*ast.BrokerPriorityParameter

	for p.curTok.Type != TokenRParen && p.curTok.Type != TokenEOF && p.curTok.Type != TokenSemicolon {
		param := &ast.BrokerPriorityParameter{}

		// Get parameter type
		paramType := strings.ToUpper(p.curTok.Literal)
		switch paramType {
		case "PRIORITY_LEVEL":
			param.ParameterType = "PriorityLevel"
		case "CONTRACT_NAME":
			param.ParameterType = "ContractName"
		case "REMOTE_SERVICE_NAME":
			param.ParameterType = "RemoteServiceName"
		case "LOCAL_SERVICE_NAME":
			param.ParameterType = "LocalServiceName"
		default:
			param.ParameterType = paramType
		}
		p.nextToken() // consume parameter name

		// Consume = if present
		if p.curTok.Type == TokenEquals {
			p.nextToken()
		}

		// Parse value: DEFAULT, ANY, or an expression
		valLiteral := strings.ToUpper(p.curTok.Literal)
		if valLiteral == "DEFAULT" {
			param.IsDefaultOrAny = "Default"
			p.nextToken() // consume DEFAULT
		} else if valLiteral == "ANY" {
			param.IsDefaultOrAny = "Any"
			p.nextToken() // consume ANY
		} else {
			param.IsDefaultOrAny = "None"
			param.ParameterValue, _ = p.parseIdentifierOrValueExpression()
		}

		params = append(params, param)

		// Skip comma
		if p.curTok.Type == TokenComma {
			p.nextToken()
		}
	}

	return params
}
