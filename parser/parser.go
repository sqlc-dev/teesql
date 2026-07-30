// Package parser provides T-SQL parsing functionality.
package parser

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/sqlc-dev/teesql/ast"
)

// Parse parses T-SQL from the given reader and returns an AST Script.
func Parse(ctx context.Context, r io.Reader) (*ast.Script, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	// For now, return an empty script for empty input
	if len(data) == 0 {
		script := &ast.Script{}
		script.SetSpan(0, 0, 1, 1)
		return script, nil
	}

	p := newParser(string(data))
	return p.parseScript()
}

// Parser holds the parsing state.
type Parser struct {
	lexer   *Lexer
	curTok  Token
	peekTok Token

	srcMap *sourceMap
	// prevEndByte is the byte offset one past the end of the last consumed
	// token; node spans extend from their first token to this point.
	prevEndByte int
	// endBeforeSemiByte is the value prevEndByte had just before the most
	// recently consumed semicolon token. It lets callers trim a trailing
	// semicolon out of a sub-statement's span (e.g. the SELECT body of
	// CREATE VIEW, whose semicolon belongs to the outer statement only).
	endBeforeSemiByte int
}

func newParser(input string) *Parser {
	p := &Parser{lexer: NewLexer(input)}
	// The lexer may have transformed the input (BOM stripping, UTF-16
	// decoding), so build the source map from its view of the input.
	p.srcMap = newSourceMap(p.lexer.input)
	// Read two tokens to initialize curTok and peekTok
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	if p.curTok.Type == TokenSemicolon {
		p.endBeforeSemiByte = p.prevEndByte
	}
	if p.curTok.Type != TokenEOF {
		if end := p.curTok.Pos + len(p.curTok.Literal); end > p.prevEndByte {
			p.prevEndByte = end
		}
	}
	p.curTok = p.peekTok
	p.peekTok = p.lexer.NextToken()
}

// spannable is satisfied by every AST node via the embedded ast.Fragment.
type spannable interface {
	SetSpan(startOffset, fragmentLength, startLine, startColumn int)
	Frag() *ast.Fragment
}

// spanned records the source span of n as running from the given start token
// to the end of the last consumed token, then returns n unchanged. It is a
// no-op for nil nodes, non-node values, and nodes for which no tokens have
// been consumed. When both an inner and an outer parse function wrap the
// same node, the outer call wins: it sees at least as wide an extent.
func spanned[T any](p *Parser, n T, start Token) T {
	s, ok := any(n).(spannable)
	if !ok {
		return n
	}
	if v := reflect.ValueOf(any(n)); v.Kind() == reflect.Pointer && v.IsNil() {
		return n
	}
	if p.prevEndByte <= start.Pos {
		return n
	}
	su, sl, sc := p.srcMap.at(start.Pos)
	eu, _, _ := p.srcMap.at(p.prevEndByte)
	s.SetSpan(su, eu-su, sl, sc)
	return n
}

// spanFrom is the statement-form of spanned for use inside parse functions.
func (p *Parser) spanFrom(start Token, n spannable) {
	if p.prevEndByte <= start.Pos {
		return
	}
	su, sl, sc := p.srcMap.at(start.Pos)
	eu, _, _ := p.srcMap.at(p.prevEndByte)
	n.SetSpan(su, eu-su, sl, sc)
}

// spanStatementList sets a StatementList's span to cover its statements
// (first statement start through last statement end), matching ScriptDom.
func spanStatementList(sl *ast.StatementList) {
	if sl == nil || len(sl.Statements) == 0 {
		return
	}
	f, ok1 := any(sl.Statements[0]).(spannable)
	l, ok2 := any(sl.Statements[len(sl.Statements)-1]).(spannable)
	if !ok1 || !ok2 || !f.Frag().HasSpan() || !l.Frag().HasSpan() {
		return
	}
	ff, lf := f.Frag(), l.Frag()
	sl.SetSpan(ff.StartOffset, lf.EndOffset()-ff.StartOffset, ff.StartLine, ff.StartColumn)
}

// spanMaxLit builds a MaxLiteral carrying the span of the current token.
func (p *Parser) spanMaxLit() *ast.MaxLiteral {
	m := &ast.MaxLiteral{LiteralType: "Max", Value: p.curTok.Literal}
	p.tokSpan(m, p.curTok)
	return m
}

// spanFromChild spans n from the start of an already-spanned child node
// through the end of the last consumed token (for wrapper nodes whose
// trailing keywords are consumed after the child was parsed).
func (p *Parser) spanFromChild(n spannable, child any) {
	c, ok := child.(spannable)
	if !ok || c == nil {
		return
	}
	if v := reflect.ValueOf(child); v.Kind() == reflect.Pointer && v.IsNil() {
		return
	}
	cf := c.Frag()
	if !cf.HasSpan() {
		return
	}
	n.SetSpan(cf.StartOffset, cf.FragmentLength, cf.StartLine, cf.StartColumn)
	eu, _, _ := p.srcMap.at(p.prevEndByte)
	if eu > cf.EndOffset() {
		n.Frag().FragmentLength = eu - cf.StartOffset
	}
}

// tokSpan stamps a node with the span of a single token (used for leaf
// nodes such as identifiers built directly from the current token).
func (p *Parser) tokSpan(n spannable, tok Token) {
	if len(tok.Literal) == 0 {
		return
	}
	su, sl, sc := p.srcMap.at(tok.Pos)
	eu, _, _ := p.srcMap.at(tok.Pos + len(tok.Literal))
	n.SetSpan(su, eu-su, sl, sc)
}

// unspan clears any recorded span on a node. ScriptDom leaves certain
// synthesized nodes (normalized keywords stored as identifiers, defaulted
// options) without position information; unspan reproduces that.
func unspan[T any](n T) T {
	if s, ok := any(n).(spannable); ok {
		if v := reflect.ValueOf(any(n)); v.Kind() != reflect.Pointer || !v.IsNil() {
			*s.Frag() = ast.Fragment{}
		}
	}
	return n
}

// spanIdent builds an Identifier carrying the span of the current token.
// It must be called while p.curTok is still the identifier's token.
func (p *Parser) spanIdent(value, quoteType string) *ast.Identifier {
	id := &ast.Identifier{Value: value, QuoteType: quoteType}
	p.tokSpan(id, p.curTok)
	return id
}

// spanVarRef builds a VariableReference carrying the span of the current
// token. It must be called while p.curTok is still the variable's token.
func (p *Parser) spanVarRef(name string) *ast.VariableReference {
	v := &ast.VariableReference{Name: name}
	p.tokSpan(v, p.curTok)
	return v
}

// varRefFromToken builds a VariableReference from a previously captured token.
func (p *Parser) varRefFromToken(tok Token) *ast.VariableReference {
	v := &ast.VariableReference{Name: tok.Literal}
	p.tokSpan(v, tok)
	return v
}

// spanTokens stamps a node with the span running from the start of one
// token to the end of another (inclusive).
func (p *Parser) spanTokens(n spannable, start, end Token) {
	su, sl, sc := p.srcMap.at(start.Pos)
	eu, _, _ := p.srcMap.at(end.Pos + len(end.Literal))
	n.SetSpan(su, eu-su, sl, sc)
}

// optHint builds a plain OptimizerHint spanning a single token. ScriptDom
// positions keyword-only optimizer hints on their first token (e.g.
// CHECKCONSTRAINTS PLAN spans just CHECKCONSTRAINTS).
func (p *Parser) optHint(kind string, tok Token) *ast.OptimizerHint {
	h := &ast.OptimizerHint{HintKind: kind}
	p.tokSpan(h, tok)
	return h
}

// spanChildToToken spans n from an already-spanned child's start through the
// end of the given token.
func (p *Parser) spanChildToToken(n spannable, child spannable, end Token) {
	cf := child.Frag()
	if !cf.HasSpan() {
		return
	}
	eu, _, _ := p.srcMap.at(end.Pos + len(end.Literal))
	n.SetSpan(cf.StartOffset, eu-cf.StartOffset, cf.StartLine, cf.StartColumn)
}

// intLitFromToken builds an IntegerLiteral carrying the span of the given token.
func (p *Parser) intLitFromToken(tok Token) *ast.IntegerLiteral {
	l := &ast.IntegerLiteral{LiteralType: "Integer", Value: tok.Literal}
	p.tokSpan(l, tok)
	return l
}

// strLit builds a StringLiteral (with an already-stripped value) carrying the
// span of the current token. It must be called while p.curTok is still the
// string's token.
func (p *Parser) strLit(value string, isNational bool) *ast.StringLiteral {
	l := &ast.StringLiteral{
		LiteralType:   "String",
		Value:         value,
		IsNational:    isNational,
		IsLargeObject: false,
	}
	p.tokSpan(l, p.curTok)
	return l
}

// identFromToken builds an Identifier from the given token's literal with
// the token's source span, without consuming it.
func (p *Parser) identFromToken(tok Token) *ast.Identifier {
	literal := tok.Literal
	quoteType := "NotQuoted"
	if len(literal) >= 2 && literal[0] == '[' && literal[len(literal)-1] == ']' {
		quoteType = "SquareBracket"
		literal = strings.ReplaceAll(literal[1:len(literal)-1], "]]", "]")
	} else if len(literal) >= 2 && literal[0] == '"' && literal[len(literal)-1] == '"' {
		quoteType = "DoubleQuote"
		literal = strings.ReplaceAll(literal[1:len(literal)-1], "\"\"", "\"")
	}
	id := &ast.Identifier{Value: literal, QuoteType: quoteType}
	p.tokSpan(id, tok)
	return id
}

// trimTrailingSemicolon shrinks n's span so it no longer covers the most
// recently consumed semicolon (used where ScriptDom attributes the semicolon
// to an enclosing statement rather than the nested one).
func (p *Parser) trimTrailingSemicolon(n spannable) {
	f := n.Frag()
	if !f.HasSpan() {
		return
	}
	eu, _, _ := p.srcMap.at(p.endBeforeSemiByte)
	if eu > f.StartOffset && eu < f.EndOffset() {
		f.FragmentLength = eu - f.StartOffset
	}
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

	// The script fragment always covers the entire input, including any
	// leading or trailing trivia (matching ScriptDom).
	script.SetSpan(0, p.srcMap.u16Len(), 1, 1)

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

	// A batch spans from the first token of its first statement to the end
	// of its last statement; the GO separator belongs to no batch.
	if len(batch.Statements) > 0 {
		first, okF := any(batch.Statements[0]).(spannable)
		last, okL := any(batch.Statements[len(batch.Statements)-1]).(spannable)
		if okF && okL && first.Frag().HasSpan() && last.Frag().HasSpan() {
			ff, lf := first.Frag(), last.Frag()
			batch.SetSpan(ff.StartOffset, lf.EndOffset()-ff.StartOffset, ff.StartLine, ff.StartColumn)
		}
	}

	return batch, nil
}

func (p *Parser) parseStatement() (ast.Statement, error) {
	astStart := p.curTok

	stmt, err := p.parseStatementInner()
	if err != nil || stmt == nil {
		return stmt, err
	}
	// ScriptDom includes a statement's terminating semicolon in its span.
	// Most parse functions consume it themselves; consume it here for the
	// ones that do not, so every statement span covers its terminator.
	if p.curTok.Type == TokenSemicolon {
		p.nextToken()
	}
	return spanned(p, stmt, astStart), nil
}

func (p *Parser) parseStatementInner() (ast.Statement, error) {
	astStart := p.curTok

	switch p.curTok.Type {
	case TokenWith:
		spanV348, spanErr348 := p.parseWithStatement()
		return spanned(p, spanV348, astStart), spanErr348
	case TokenSelect, TokenLParen:
		spanV349, spanErr349 := p.parseSelectStatement()
		return spanned(p, spanV349, astStart), spanErr349
	case TokenInsert:
		spanV350, spanErr350 := p.parseInsertStatement()
		return spanned(p, spanV350, astStart), spanErr350
	case TokenUpdate:
		spanV351, spanErr351 := p.parseUpdateOrUpdateStatisticsStatement()
		return spanned(p, spanV351, astStart), spanErr351
	case TokenDelete:
		spanV352, spanErr352 := p.parseDeleteStatement()
		return spanned(p, spanV352, astStart), spanErr352
	case TokenDeclare:
		spanV353, spanErr353 := p.parseDeclareVariableStatement()
		return spanned(p, spanV353, astStart), spanErr353
	case TokenSet:
		spanV354, spanErr354 := p.parseSetVariableStatement()
		return spanned(p, spanV354, astStart), spanErr354
	case TokenIf:
		spanV355, spanErr355 := p.parseIfStatement()
		return spanned(p, spanV355, astStart), spanErr355
	case TokenWhile:
		spanV356, spanErr356 := p.parseWhileStatement()
		return spanned(p, spanV356, astStart), spanErr356
	case TokenBegin:
		spanV357, spanErr357 := p.parseBeginStatement()
		return spanned(p, spanV357, astStart), spanErr357
	case TokenCreate:
		spanV358, spanErr358 := p.parseCreateStatement()
		return spanned(p, spanV358, astStart), spanErr358
	case TokenExec, TokenExecute:
		spanV359, spanErr359 := p.parseExecuteStatement()
		return spanned(p, spanV359, astStart), spanErr359
	case TokenPrint:
		spanV360, spanErr360 := p.parsePrintStatement()
		return spanned(p, spanV360, astStart), spanErr360
	case TokenThrow:
		spanV361, spanErr361 := p.parseThrowStatement()
		return spanned(p, spanV361, astStart), spanErr361
	case TokenAlter:
		spanV362, spanErr362 := p.parseAlterStatement()
		return spanned(p, spanV362, astStart), spanErr362
	case TokenRevert:
		spanV363, spanErr363 := p.parseRevertStatement()
		return spanned(p, spanV363, astStart), spanErr363
	case TokenDrop:
		spanV364, spanErr364 := p.parseDropStatement()
		return spanned(p, spanV364, astStart), spanErr364
	case TokenReturn:
		spanV365, spanErr365 := p.parseReturnStatement()
		return spanned(p, spanV365, astStart), spanErr365
	case TokenBreak:
		spanV366, spanErr366 := p.parseBreakStatement()
		return spanned(p, spanV366, astStart), spanErr366
	case TokenContinue:
		spanV367, spanErr367 := p.parseContinueStatement()
		return spanned(p, spanV367, astStart), spanErr367
	case TokenGrant:
		return p.parseGrantStatement()
	case TokenRevoke:
		return p.parseRevokeStatement()
	case TokenDeny:
		return p.parseDenyStatement()
	case TokenCommit:
		spanV368, spanErr368 := p.parseCommitTransactionStatement()
		return spanned(p, spanV368, astStart), spanErr368
	case TokenRollback:
		spanV369, spanErr369 := p.parseRollbackTransactionStatement()
		return spanned(p, spanV369, astStart), spanErr369
	case TokenSave:
		spanV370, spanErr370 := p.parseSaveTransactionStatement()
		return spanned(p, spanV370, astStart), spanErr370
	case TokenWaitfor:
		spanV371, spanErr371 := p.parseWaitForStatement()
		return spanned(p, spanV371, astStart), spanErr371
	case TokenGoto:
		spanV372, spanErr372 := p.parseGotoStatement()
		return spanned(p, spanV372, astStart), spanErr372
	case TokenMove:
		spanV373, spanErr373 := p.parseMoveConversationStatement()
		return spanned(p, spanV373, astStart), spanErr373
	case TokenGet:
		spanV374, spanErr374 := p.parseGetConversationGroupStatement()
		return spanned(p, spanV374, astStart), spanErr374
	case TokenTruncate:
		spanV375, spanErr375 := p.parseTruncateTableStatement()
		return spanned(p, spanV375, astStart), spanErr375
	case TokenUse:
		spanV376, spanErr376 := p.parseUseStatement()
		return spanned(p, spanV376, astStart), spanErr376
	case TokenKill:
		spanV377, spanErr377 := p.parseKillStatement()
		return spanned(p, spanV377, astStart), spanErr377
	case TokenCheckpoint:
		spanV378, spanErr378 := p.parseCheckpointStatement()
		return spanned(p, spanV378, astStart), spanErr378
	case TokenReconfigure:
		spanV379, spanErr379 := p.parseReconfigureStatement()
		return spanned(p, spanV379, astStart), spanErr379
	case TokenShutdown:
		spanV380, spanErr380 := p.parseShutdownStatement()
		return spanned(p, spanV380, astStart), spanErr380
	case TokenSetuser:
		spanV381, spanErr381 := p.parseSetUserStatement()
		return spanned(p, spanV381, astStart), spanErr381
	case TokenLineno:
		spanV382, spanErr382 := p.parseLineNoStatement()
		return spanned(p, spanV382, astStart), spanErr382
	case TokenRaiserror:
		spanV383, spanErr383 := p.parseRaiseErrorStatement()
		return spanned(p, spanV383, astStart), spanErr383
	case TokenReadtext:
		spanV384, spanErr384 := p.parseReadTextStatement()
		return spanned(p, spanV384, astStart), spanErr384
	case TokenWritetext:
		spanV385, spanErr385 := p.parseWriteTextStatement()
		return spanned(p, spanV385, astStart), spanErr385
	case TokenUpdatetext:
		spanV386, spanErr386 := p.parseUpdateTextStatement()
		return spanned(p, spanV386, astStart), spanErr386
	case TokenSend:
		spanV387, spanErr387 := p.parseSendStatement()
		return spanned(p, spanV387, astStart), spanErr387
	case TokenReceive:
		spanV388, spanErr388 := p.parseReceiveStatement()
		return spanned(p, spanV388, astStart), spanErr388
	case TokenRestore:
		return p.parseRestoreStatement()
	case TokenBackup:
		spanV389, spanErr389 := p.parseBackupStatement()
		return spanned(p, spanV389, astStart), spanErr389
	case TokenClose:
		spanV390, spanErr390 := p.parseCloseStatement()
		return spanned(p, spanV390, astStart), spanErr390
	case TokenEnd:
		// Check for END CONVERSATION
		if p.peekTok.Type == TokenConversation {
			spanV391, spanErr391 := p.parseEndConversationStatement()
			return spanned(p, spanV391, astStart), spanErr391
		}
		return nil, fmt.Errorf("unexpected token: END")
	case TokenOpen:
		spanV392, spanErr392 := p.parseOpenStatement()
		return spanned(p, spanV392, astStart), spanErr392
	case TokenDbcc:
		spanV393, spanErr393 := p.parseDbccStatement()
		return spanned(p, spanV393, astStart), spanErr393
	case TokenAdd:
		spanV394, spanErr394 := p.parseAddStatement()
		return spanned(p, spanV394, astStart), spanErr394
	case TokenSemicolon:
		p.nextToken()
		return nil, nil
	case TokenIdent:
		// Check for BULK INSERT
		if strings.ToUpper(p.curTok.Literal) == "BULK" {
			p.nextToken() // consume BULK
			spanV395, spanErr395 := p.parseBulkInsertStatement()
			return spanned(p, spanV395, astStart), spanErr395
		}
		// Check for RENAME (Azure SQL DW/Synapse)
		if strings.ToUpper(p.curTok.Literal) == "RENAME" {
			spanV396, spanErr396 := p.parseRenameStatement()
			return spanned(p, spanV396, astStart), spanErr396
		}
		// Check for FETCH cursor
		if strings.ToUpper(p.curTok.Literal) == "FETCH" {
			spanV397, spanErr397 := p.parseFetchCursorStatement()
			return spanned(p, spanV397, astStart), spanErr397
		}
		// Check for DEALLOCATE cursor
		if strings.ToUpper(p.curTok.Literal) == "DEALLOCATE" {
			spanV398, spanErr398 := p.parseDeallocateCursorStatement()
			return spanned(p, spanV398, astStart), spanErr398
		}
		// Check for ENABLE TRIGGER
		if strings.ToUpper(p.curTok.Literal) == "ENABLE" {
			spanV399, spanErr399 := p.parseEnableDisableTriggerStatement("Enable")
			return spanned(p, spanV399, astStart), spanErr399
		}
		// Check for DISABLE TRIGGER
		if strings.ToUpper(p.curTok.Literal) == "DISABLE" {
			spanV400, spanErr400 := p.parseEnableDisableTriggerStatement("Disable")
			return spanned(p, spanV400, astStart), spanErr400
		}
		// Check for MERGE statement
		if strings.ToUpper(p.curTok.Literal) == "MERGE" {
			return p.parseMergeStatement()
		}
		// Check for COPY INTO statement
		if strings.ToUpper(p.curTok.Literal) == "COPY" {
			spanV401, spanErr401 := p.parseCopyStatement()
			return spanned(p, spanV401, astStart), spanErr401
		}
		// Check for label (identifier followed by colon)
		spanV402, spanErr402 := p.parseLabelOrError()
		return spanned(p, spanV402, astStart), spanErr402
	default:
		return nil, fmt.Errorf("unexpected token: %s", p.curTok.Literal)
	}
}
