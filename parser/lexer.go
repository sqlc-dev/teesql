package parser

import (
	"encoding/binary"
	"strings"
	"unicode"
	"unicode/utf16"
)

// TokenType represents the type of a token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenIdent
	TokenNumber
	TokenString
	TokenNationalString
	TokenBinary
	TokenStar
	TokenComma
	TokenDot
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenSemicolon
	TokenEquals
	TokenLessThan
	TokenGreaterThan
	TokenPlus
	TokenMinus
	TokenSlash
	TokenModulo

	// Keywords
	TokenSelect
	TokenFrom
	TokenWhere
	TokenAnd
	TokenOr
	TokenAs
	TokenOption
	TokenAll
	TokenDistinct
	TokenPrint
	TokenThrow
	TokenAlter
	TokenTable
	TokenDrop
	TokenIndex
	TokenRevert
	TokenWith
	TokenCookie
	TokenDatabase
	TokenScoped
	TokenCredential
	TokenTop
	TokenPercent
	TokenTies
	TokenInto
	TokenGroup
	TokenBy
	TokenHaving
	TokenOrder
	TokenAsc
	TokenDesc
	TokenUnion
	TokenExcept
	TokenIntersect
	TokenCross
	TokenJoin
	TokenInner
	TokenLeft
	TokenRight
	TokenFull
	TokenOuter
	TokenOn
	TokenRollup
	TokenCube
	TokenNotEqual
	TokenLessOrEqual
	TokenGreaterOrEqual
	TokenNot
	TokenLBrace
	TokenRBrace
	TokenLeftShift
	TokenRightShift
	TokenPipe           // |
	TokenDoublePipe     // ||
	TokenConcatEquals   // ||=
	TokenBitwiseAnd     // &
	TokenPlusEquals     // +=
	TokenMinusEquals    // -=
	TokenStarEquals     // *=
	TokenSlashEquals    // /=
	TokenModuloEquals   // %=
	TokenAndEquals      // &=
	TokenOrEquals       // |=
	TokenXorEquals      // ^=
	TokenCaret          // ^

	// DML Keywords
	TokenInsert
	TokenUpdate
	TokenDelete
	TokenSet
	TokenValues
	TokenDefault
	TokenNull
	TokenIs
	TokenIn
	TokenLike
	TokenBetween
	TokenEscape
	TokenExec
	TokenExecute
	TokenOver

	// DDL Keywords
	TokenCreate
	TokenView
	TokenSchema
	TokenProcedure
	TokenFunction
	TokenTrigger
	TokenAuthorization

	// Control flow keywords
	TokenDeclare
	TokenIf
	TokenElse
	TokenCase
	TokenWhen
	TokenThen
	TokenWhile
	TokenBegin
	TokenEnd
	TokenReturn
	TokenBreak
	TokenContinue
	TokenGoto
	TokenTry
	TokenCatch

	// Additional keywords
	TokenCurrent
	TokenOf
	TokenCursor
	TokenOpenRowset
	TokenHoldlock
	TokenNowait
	TokenFast
	TokenMaxdop

	// Security keywords
	TokenGrant
	TokenRevoke
	TokenDeny
	TokenTo
	TokenPublic

	// Transaction keywords
	TokenCommit
	TokenRollback
	TokenSave
	TokenTransaction
	TokenTran
	TokenWork

	// Additional keywords
	TokenWaitfor
	TokenDelay
	TokenTime
	TokenMaster
	TokenKey
	TokenEncryption
	TokenPassword
	TokenLabel
	TokenRaiserror
	TokenReadtext
	TokenWritetext
	TokenUpdatetext
	TokenTruncate
	TokenColon
	TokenColonColon
	TokenMove
	TokenConversation
	TokenDialog
	TokenGet
	TokenUse
	TokenKill
	TokenCheckpoint
	TokenReconfigure
	TokenOverride
	TokenShutdown
	TokenSetuser
	TokenLineno
	TokenStatusonly
	TokenNoreset
	TokenSend
	TokenMessage
	TokenTyp
	TokenReceive
	TokenLogin
	TokenAdd
	TokenUser
	TokenCaller
	TokenNoRevert
	TokenExternal
	TokenLanguage
	TokenRestore
	TokenBackup
	TokenFilestream
	TokenReturns
	TokenClose
	TokenOpen
	TokenSymmetric
	TokenStats
	TokenJob
	TokenQuery
	TokenNotification
	TokenSubscription
	TokenDecryption
	TokenAsymmetric
	TokenCertificate
	TokenDbcc
)

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

// Lexer tokenizes T-SQL input.
type Lexer struct {
	input   string
	pos     int
	readPos int
	ch      byte
}

// NewLexer creates a new Lexer for the given input.
func NewLexer(input string) *Lexer {
	// Handle UTF-16 LE BOM (0xFF 0xFE) - convert to UTF-8
	if len(input) >= 2 && input[0] == 0xFF && input[1] == 0xFE {
		input = utf16LEToUTF8(input[2:])
	}
	// Skip UTF-8 BOM if present
	if len(input) >= 3 && input[0] == 0xEF && input[1] == 0xBB && input[2] == 0xBF {
		input = input[3:]
	}
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// utf16LEToUTF8 converts a UTF-16 LE string to UTF-8
func utf16LEToUTF8(data string) string {
	// Convert byte string to []uint16
	if len(data)%2 != 0 {
		data = data[:len(data)-1] // Truncate odd byte
	}
	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.LittleEndian.Uint16([]byte(data[i*2 : i*2+2]))
	}
	// Decode UTF-16 to runes
	runes := utf16.Decode(u16s)
	return string(runes)
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespaceAndComments()

	tok := Token{Pos: l.pos}

	switch l.ch {
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
	case '*':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenStarEquals
			tok.Literal = "*="
			l.readChar()
		} else {
			tok.Type = TokenStar
			tok.Literal = "*"
			l.readChar()
		}
	case ',':
		tok.Type = TokenComma
		tok.Literal = ","
		l.readChar()
	case '.':
		tok.Type = TokenDot
		tok.Literal = "."
		l.readChar()
	case '(':
		tok.Type = TokenLParen
		tok.Literal = "("
		l.readChar()
	case ')':
		tok.Type = TokenRParen
		tok.Literal = ")"
		l.readChar()
	case '[':
		tok = l.readBracketedIdentifier()
	case ']':
		tok.Type = TokenRBracket
		tok.Literal = "]"
		l.readChar()
	case ';':
		tok.Type = TokenSemicolon
		tok.Literal = ";"
		l.readChar()
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok.Type = TokenColonColon
			tok.Literal = "::"
			l.readChar()
		} else {
			tok.Type = TokenColon
			tok.Literal = ":"
			l.readChar()
		}
	case '=':
		tok.Type = TokenEquals
		tok.Literal = "="
		l.readChar()
	case '<':
		if l.peekChar() == '>' {
			l.readChar()
			tok.Type = TokenNotEqual
			tok.Literal = "<>"
			l.readChar()
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenLessOrEqual
			tok.Literal = "<="
			l.readChar()
		} else if l.peekChar() == '<' {
			l.readChar()
			tok.Type = TokenLeftShift
			tok.Literal = "<<"
			l.readChar()
		} else {
			tok.Type = TokenLessThan
			tok.Literal = "<"
			l.readChar()
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenGreaterOrEqual
			tok.Literal = ">="
			l.readChar()
		} else if l.peekChar() == '>' {
			l.readChar()
			tok.Type = TokenRightShift
			tok.Literal = ">>"
			l.readChar()
		} else {
			tok.Type = TokenGreaterThan
			tok.Literal = ">"
			l.readChar()
		}
	case '{':
		tok.Type = TokenLBrace
		tok.Literal = "{"
		l.readChar()
	case '}':
		tok.Type = TokenRBrace
		tok.Literal = "}"
		l.readChar()
	case '+':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenPlusEquals
			tok.Literal = "+="
			l.readChar()
		} else {
			tok.Type = TokenPlus
			tok.Literal = "+"
			l.readChar()
		}
	case '-':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenMinusEquals
			tok.Literal = "-="
			l.readChar()
		} else {
			tok.Type = TokenMinus
			tok.Literal = "-"
			l.readChar()
		}
	case '/':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenSlashEquals
			tok.Literal = "/="
			l.readChar()
		} else {
			tok.Type = TokenSlash
			tok.Literal = "/"
			l.readChar()
		}
	case '%':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenModuloEquals
			tok.Literal = "%="
			l.readChar()
		} else {
			tok.Type = TokenModulo
			tok.Literal = "%"
			l.readChar()
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar() // consume first |
			if l.peekChar() == '=' {
				l.readChar() // consume second |
				tok.Type = TokenConcatEquals
				tok.Literal = "||="
				l.readChar() // consume =
			} else {
				tok.Type = TokenDoublePipe
				tok.Literal = "||"
				l.readChar() // consume second |
			}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenOrEquals
			tok.Literal = "|="
			l.readChar()
		} else {
			tok.Type = TokenPipe
			tok.Literal = "|"
			l.readChar()
		}
	case '&':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenAndEquals
			tok.Literal = "&="
			l.readChar()
		} else {
			tok.Type = TokenBitwiseAnd
			tok.Literal = "&"
			l.readChar()
		}
	case '^':
		if l.peekChar() == '=' {
			l.readChar()
			tok.Type = TokenXorEquals
			tok.Literal = "^="
			l.readChar()
		} else {
			tok.Type = TokenCaret
			tok.Literal = "^"
			l.readChar()
		}
	case '\'':
		tok = l.readString()
	case '"':
		tok = l.readDoubleQuotedIdentifier()
	default:
		// Handle $ only if followed by a letter (for pseudo-columns like $ROWGUID)
		if l.ch == '$' && isLetter(l.peekChar()) {
			tok = l.readIdentifier()
		} else if isLetter(l.ch) || l.ch == '_' || l.ch == '@' || l.ch == '#' {
			tok = l.readIdentifier()
		} else if l.ch >= 0x80 {
			// Check for Unicode letter at start of identifier
			r, _ := l.peekRune()
			if unicode.IsLetter(r) {
				tok = l.readIdentifier()
			} else {
				tok.Type = TokenError
				tok.Literal = string(l.ch)
				l.readChar()
			}
		} else if isDigit(l.ch) {
			tok = l.readNumber()
		} else {
			tok.Type = TokenError
			tok.Literal = string(l.ch)
			l.readChar()
		}
	}

	return tok
}

// isWhitespace checks if the current position contains whitespace.
// T-SQL treats many control characters and Unicode spaces as whitespace.
func (l *Lexer) isWhitespace() bool {
	if l.ch == 0 {
		return false
	}
	// ASCII whitespace and control characters (0x01-0x20 range, excluding 0x00)
	// T-SQL treats most ASCII control characters as whitespace
	if l.ch <= 0x20 {
		return true
	}
	// Check for multi-byte UTF-8 whitespace sequences
	if l.ch >= 0x80 {
		// Try to decode rune at current position
		r, _ := l.peekRune()
		// unicode.IsSpace covers most whitespace, but T-SQL also treats
		// Zero Width Space (U+200B) as whitespace
		return unicode.IsSpace(r) || r == 0x200B
	}
	return false
}

// peekRune returns the rune at the current position without advancing.
func (l *Lexer) peekRune() (rune, int) {
	if l.pos >= len(l.input) {
		return 0, 0
	}
	// Fast path for ASCII
	if l.input[l.pos] < 0x80 {
		return rune(l.input[l.pos]), 1
	}
	// Decode UTF-8
	r, size := decodeRuneAt(l.input, l.pos)
	return r, size
}

// decodeRuneAt decodes a UTF-8 rune at the given position.
func decodeRuneAt(s string, pos int) (rune, int) {
	if pos >= len(s) {
		return 0, 0
	}
	b := s[pos]
	if b < 0x80 {
		return rune(b), 1
	}
	// 2-byte sequence
	if b&0xE0 == 0xC0 && pos+1 < len(s) {
		return rune(b&0x1F)<<6 | rune(s[pos+1]&0x3F), 2
	}
	// 3-byte sequence
	if b&0xF0 == 0xE0 && pos+2 < len(s) {
		return rune(b&0x0F)<<12 | rune(s[pos+1]&0x3F)<<6 | rune(s[pos+2]&0x3F), 3
	}
	// 4-byte sequence
	if b&0xF8 == 0xF0 && pos+3 < len(s) {
		return rune(b&0x07)<<18 | rune(s[pos+1]&0x3F)<<12 | rune(s[pos+2]&0x3F)<<6 | rune(s[pos+3]&0x3F), 4
	}
	return rune(b), 1
}

// skipWhitespaceChar advances past one whitespace character (which may be multi-byte).
func (l *Lexer) skipWhitespaceChar() {
	if l.ch < 0x80 {
		l.readChar()
		return
	}
	// Multi-byte UTF-8: advance by rune size
	_, size := l.peekRune()
	for i := 0; i < size; i++ {
		l.readChar()
	}
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// Skip whitespace (including Unicode whitespace)
		for l.ch != 0 && l.isWhitespace() {
			l.skipWhitespaceChar()
		}

		// Skip line comments (-- ...)
		if l.ch == '-' && l.peekChar() == '-' {
			for l.ch != 0 && l.ch != '\n' {
				l.readChar()
			}
			continue
		}

		// Skip block comments (/* ... */)
		if l.ch == '/' && l.peekChar() == '*' {
			l.readChar() // skip /
			l.readChar() // skip *
			for l.ch != 0 {
				if l.ch == '*' && l.peekChar() == '/' {
					l.readChar() // skip *
					l.readChar() // skip /
					break
				}
				l.readChar()
			}
			continue
		}

		break
	}
}

// isIdentifierChar checks if the current position is a valid identifier character.
// Handles both ASCII and Unicode letters.
func (l *Lexer) isIdentifierChar(first bool) bool {
	if l.ch == 0 {
		return false
	}
	// ASCII fast path
	if l.ch < 0x80 {
		if isLetter(l.ch) || l.ch == '_' || l.ch == '@' || l.ch == '#' {
			return true
		}
		if !first && (isDigit(l.ch) || l.ch == '$') {
			return true
		}
		// $ is valid at start only when followed by a letter (pseudo-columns like $ROWGUID)
		// But in an identifier context, $ is valid inside the identifier
		if l.ch == '$' {
			return true
		}
		return false
	}
	// UTF-8: decode rune and check if it's a letter
	r, _ := l.peekRune()
	return unicode.IsLetter(r)
}

// advanceIdentifierChar advances past the current identifier character (which may be multi-byte).
func (l *Lexer) advanceIdentifierChar() {
	if l.ch < 0x80 {
		l.readChar()
		return
	}
	// Multi-byte UTF-8: advance by rune size
	_, size := l.peekRune()
	for i := 0; i < size; i++ {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() Token {
	startPos := l.pos
	first := true
	for l.isIdentifierChar(first) {
		l.advanceIdentifierChar()
		first = false
	}
	literal := l.input[startPos:l.pos]

	// Handle N'...' national string literals
	if (literal == "N" || literal == "n") && l.ch == '\'' {
		return l.readNationalString(startPos)
	}

	return Token{
		Type:    lookupKeyword(literal),
		Literal: literal,
		Pos:     startPos,
	}
}

func (l *Lexer) readBracketedIdentifier() Token {
	startPos := l.pos
	l.readChar() // skip opening [
	for l.ch != 0 {
		if l.ch == ']' {
			if l.peekChar() == ']' {
				// Escaped bracket ]], consume both and continue
				l.readChar()
				l.readChar()
				continue
			}
			// Closing bracket
			l.readChar() // skip closing ]
			break
		}
		l.readChar()
	}
	return Token{
		Type:    TokenIdent,
		Literal: l.input[startPos:l.pos],
		Pos:     startPos,
	}
}

func (l *Lexer) readDoubleQuotedIdentifier() Token {
	startPos := l.pos
	l.readChar() // skip opening "
	for l.ch != 0 {
		if l.ch == '"' {
			if l.peekChar() == '"' {
				// Escaped quote "", consume both and continue
				l.readChar()
				l.readChar()
				continue
			}
			// Closing quote
			l.readChar() // skip closing "
			break
		}
		l.readChar()
	}
	return Token{
		Type:    TokenIdent,
		Literal: l.input[startPos:l.pos],
		Pos:     startPos,
	}
}

func (l *Lexer) readString() Token {
	startPos := l.pos
	l.readChar() // skip opening quote
	for l.ch != 0 {
		if l.ch == '\'' {
			if l.peekChar() == '\'' {
				// Escaped quote
				l.readChar()
				l.readChar()
				continue
			}
			break
		}
		l.readChar()
	}
	if l.ch == '\'' {
		l.readChar() // skip closing quote
	}
	return Token{
		Type:    TokenString,
		Literal: l.input[startPos:l.pos],
		Pos:     startPos,
	}
}

func (l *Lexer) readNationalString(startPos int) Token {
	// startPos already points to 'N', now we're at the opening quote
	l.readChar() // skip opening quote
	for l.ch != 0 {
		if l.ch == '\'' {
			if l.peekChar() == '\'' {
				// Escaped quote
				l.readChar()
				l.readChar()
				continue
			}
			break
		}
		l.readChar()
	}
	if l.ch == '\'' {
		l.readChar() // skip closing quote
	}
	return Token{
		Type:    TokenNationalString,
		Literal: l.input[startPos:l.pos],
		Pos:     startPos,
	}
}

func (l *Lexer) readNumber() Token {
	startPos := l.pos

	// Check for binary literal (0x...)
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		l.readChar() // consume 0
		l.readChar() // consume x
		for isHexDigit(l.ch) {
			l.readChar()
		}
		return Token{
			Type:    TokenBinary,
			Literal: l.input[startPos:l.pos],
			Pos:     startPos,
		}
	}

	for isDigit(l.ch) {
		l.readChar()
	}
	// Handle decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return Token{
		Type:    TokenNumber,
		Literal: l.input[startPos:l.pos],
		Pos:     startPos,
	}
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isLetter(ch byte) bool {
	// Only ASCII letters - don't treat UTF-8 leading bytes as letters
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

var keywords = map[string]TokenType{
	"SELECT":        TokenSelect,
	"FROM":          TokenFrom,
	"WHERE":         TokenWhere,
	"AND":           TokenAnd,
	"OR":            TokenOr,
	"AS":            TokenAs,
	"OPTION":        TokenOption,
	"ALL":           TokenAll,
	"DISTINCT":      TokenDistinct,
	"PRINT":         TokenPrint,
	"THROW":         TokenThrow,
	"ALTER":         TokenAlter,
	"TABLE":         TokenTable,
	"DROP":          TokenDrop,
	"INDEX":         TokenIndex,
	"REVERT":        TokenRevert,
	"WITH":          TokenWith,
	"COOKIE":        TokenCookie,
	"DATABASE":      TokenDatabase,
	"SCOPED":        TokenScoped,
	"CREDENTIAL":    TokenCredential,
	"TOP":           TokenTop,
	"PERCENT":       TokenPercent,
	"TIES":          TokenTies,
	"INTO":          TokenInto,
	"GROUP":         TokenGroup,
	"BY":            TokenBy,
	"HAVING":        TokenHaving,
	"ORDER":         TokenOrder,
	"ASC":           TokenAsc,
	"DESC":          TokenDesc,
	"UNION":         TokenUnion,
	"EXCEPT":        TokenExcept,
	"INTERSECT":     TokenIntersect,
	"CROSS":         TokenCross,
	"JOIN":          TokenJoin,
	"INNER":         TokenInner,
	"LEFT":          TokenLeft,
	"RIGHT":         TokenRight,
	"FULL":          TokenFull,
	"OUTER":         TokenOuter,
	"ON":            TokenOn,
	"ROLLUP":        TokenRollup,
	"CUBE":          TokenCube,
	"NOT":           TokenNot,
	"INSERT":        TokenInsert,
	"UPDATE":        TokenUpdate,
	"DELETE":        TokenDelete,
	"SET":           TokenSet,
	"VALUES":        TokenValues,
	"DEFAULT":       TokenDefault,
	"NULL":          TokenNull,
	"IS":            TokenIs,
	"IN":            TokenIn,
	"LIKE":          TokenLike,
	"BETWEEN":       TokenBetween,
	"ESCAPE":        TokenEscape,
	"EXEC":          TokenExec,
	"EXECUTE":       TokenExecute,
	"OVER":          TokenOver,
	"CREATE":        TokenCreate,
	"VIEW":          TokenView,
	"SCHEMA":        TokenSchema,
	"PROCEDURE":     TokenProcedure,
	"PROC":          TokenProcedure,
	"FUNCTION":      TokenFunction,
	"TRIGGER":       TokenTrigger,
	"AUTHORIZATION": TokenAuthorization,
	"DECLARE":       TokenDeclare,
	"IF":            TokenIf,
	"ELSE":          TokenElse,
	"CASE":          TokenCase,
	"WHEN":          TokenWhen,
	"THEN":          TokenThen,
	"WHILE":         TokenWhile,
	"BEGIN":         TokenBegin,
	"END":           TokenEnd,
	"RETURN":        TokenReturn,
	"BREAK":         TokenBreak,
	"CONTINUE":      TokenContinue,
	"GOTO":          TokenGoto,
	"TRY":           TokenTry,
	"CATCH":         TokenCatch,
	"CURRENT":       TokenCurrent,
	"OF":            TokenOf,
	"CURSOR":        TokenCursor,
	"OPENROWSET":    TokenOpenRowset,
	"HOLDLOCK":      TokenHoldlock,
	"NOWAIT":        TokenNowait,
	"FAST":          TokenFast,
	"MAXDOP":        TokenMaxdop,
	"GRANT":         TokenGrant,
	"REVOKE":        TokenRevoke,
	"DENY":          TokenDeny,
	"TO":            TokenTo,
	"PUBLIC":        TokenPublic,
	"COMMIT":        TokenCommit,
	"ROLLBACK":      TokenRollback,
	"SAVE":          TokenSave,
	"TRANSACTION":   TokenTransaction,
	"TRAN":          TokenTran,
	"WORK":          TokenWork,
	"WAITFOR":       TokenWaitfor,
	"DELAY":         TokenDelay,
	"TIME":          TokenTime,
	"MASTER":        TokenMaster,
	"KEY":           TokenKey,
	"ENCRYPTION":    TokenEncryption,
	"PASSWORD":      TokenPassword,
	"RAISERROR":     TokenRaiserror,
	"READTEXT":      TokenReadtext,
	"WRITETEXT":     TokenWritetext,
	"UPDATETEXT":    TokenUpdatetext,
	"TRUNCATE":      TokenTruncate,
	"MOVE":          TokenMove,
	"CONVERSATION":  TokenConversation,
	"DIALOG":        TokenDialog,
	"GET":           TokenGet,
	"USE":           TokenUse,
	"KILL":          TokenKill,
	"CHECKPOINT":    TokenCheckpoint,
	"RECONFIGURE":   TokenReconfigure,
	"OVERRIDE":      TokenOverride,
	"SHUTDOWN":      TokenShutdown,
	"SETUSER":       TokenSetuser,
	"LINENO":        TokenLineno,
	"STATUSONLY":    TokenStatusonly,
	"NORESET":       TokenNoreset,
	"SEND":          TokenSend,
	"MESSAGE":       TokenMessage,
	"TYPE":          TokenTyp,
	"RECEIVE":       TokenReceive,
	"LOGIN":         TokenLogin,
	"ADD":           TokenAdd,
	"USER":          TokenUser,
	"CALLER":        TokenCaller,
	"NOREVERT":      TokenNoRevert,
	"EXTERNAL":      TokenExternal,
	"LANGUAGE":      TokenLanguage,
	"RESTORE":       TokenRestore,
	"BACKUP":        TokenBackup,
	"FILESTREAM":    TokenFilestream,
	"RETURNS":       TokenReturns,
	"CLOSE":         TokenClose,
	"OPEN":          TokenOpen,
	"SYMMETRIC":     TokenSymmetric,
	"STATS":         TokenStats,
	"JOB":           TokenJob,
	"QUERY":         TokenQuery,
	"NOTIFICATION":  TokenNotification,
	"SUBSCRIPTION":  TokenSubscription,
	"DECRYPTION":    TokenDecryption,
	"ASYMMETRIC":    TokenAsymmetric,
	"CERTIFICATE":   TokenCertificate,
	"DBCC":          TokenDbcc,
}

func lookupKeyword(ident string) TokenType {
	if tok, ok := keywords[strings.ToUpper(ident)]; ok {
		return tok
	}
	return TokenIdent
}
