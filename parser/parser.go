// Package parser provides T-SQL parsing functionality.
package parser

import (
	"context"
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
		return p.parseUpdateOrUpdateStatisticsStatement()
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
	case TokenMove:
		return p.parseMoveConversationStatement()
	case TokenGet:
		return p.parseGetConversationGroupStatement()
	case TokenTruncate:
		return p.parseTruncateTableStatement()
	case TokenUse:
		return p.parseUseStatement()
	case TokenKill:
		return p.parseKillStatement()
	case TokenCheckpoint:
		return p.parseCheckpointStatement()
	case TokenReconfigure:
		return p.parseReconfigureStatement()
	case TokenShutdown:
		return p.parseShutdownStatement()
	case TokenSetuser:
		return p.parseSetUserStatement()
	case TokenLineno:
		return p.parseLineNoStatement()
	case TokenRaiserror:
		return p.parseRaiseErrorStatement()
	case TokenReadtext:
		return p.parseReadTextStatement()
	case TokenWritetext:
		return p.parseWriteTextStatement()
	case TokenUpdatetext:
		return p.parseUpdateTextStatement()
	case TokenSend:
		return p.parseSendStatement()
	case TokenReceive:
		return p.parseReceiveStatement()
	case TokenRestore:
		return p.parseRestoreStatement()
	case TokenBackup:
		return p.parseBackupStatement()
	case TokenClose:
		return p.parseCloseStatement()
	case TokenOpen:
		return p.parseOpenStatement()
	case TokenDbcc:
		return p.parseDbccStatement()
	case TokenSemicolon:
		p.nextToken()
		return nil, nil
	case TokenIdent:
		// Check for BULK INSERT
		if strings.ToUpper(p.curTok.Literal) == "BULK" {
			p.nextToken() // consume BULK
			return p.parseBulkInsertStatement()
		}
		// Check for RENAME (Azure SQL DW/Synapse)
		if strings.ToUpper(p.curTok.Literal) == "RENAME" {
			return p.parseRenameStatement()
		}
		// Check for FETCH cursor
		if strings.ToUpper(p.curTok.Literal) == "FETCH" {
			return p.parseFetchCursorStatement()
		}
		// Check for DEALLOCATE cursor
		if strings.ToUpper(p.curTok.Literal) == "DEALLOCATE" {
			return p.parseDeallocateCursorStatement()
		}
		// Check for ENABLE TRIGGER
		if strings.ToUpper(p.curTok.Literal) == "ENABLE" {
			return p.parseEnableDisableTriggerStatement("Enable")
		}
		// Check for DISABLE TRIGGER
		if strings.ToUpper(p.curTok.Literal) == "DISABLE" {
			return p.parseEnableDisableTriggerStatement("Disable")
		}
		// Check for label (identifier followed by colon)
		return p.parseLabelOrError()
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.curTok.Literal)
	}
}

