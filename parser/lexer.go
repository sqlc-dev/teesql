package parser

import (
	"strings"
	"unicode"
)

// TokenType represents the type of a token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenIdent
	TokenNumber
	TokenString
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
	l := &Lexer{input: input}
	l.readChar()
	return l
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
		tok.Type = TokenStar
		tok.Literal = "*"
		l.readChar()
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
		tok.Type = TokenPlus
		tok.Literal = "+"
		l.readChar()
	case '-':
		tok.Type = TokenMinus
		tok.Literal = "-"
		l.readChar()
	case '\'':
		tok = l.readString()
	default:
		if isLetter(l.ch) || l.ch == '_' || l.ch == '@' || l.ch == '#' {
			tok = l.readIdentifier()
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

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// Skip whitespace
		for l.ch != 0 && (l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r') {
			l.readChar()
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

func (l *Lexer) readIdentifier() Token {
	startPos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '@' || l.ch == '#' {
		l.readChar()
	}
	literal := l.input[startPos:l.pos]
	return Token{
		Type:    lookupKeyword(literal),
		Literal: literal,
		Pos:     startPos,
	}
}

func (l *Lexer) readBracketedIdentifier() Token {
	startPos := l.pos
	l.readChar() // skip opening [
	for l.ch != 0 && l.ch != ']' {
		l.readChar()
	}
	if l.ch == ']' {
		l.readChar() // skip closing ]
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

func (l *Lexer) readNumber() Token {
	startPos := l.pos
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

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch))
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

var keywords = map[string]TokenType{
	"SELECT":     TokenSelect,
	"FROM":       TokenFrom,
	"WHERE":      TokenWhere,
	"AND":        TokenAnd,
	"OR":         TokenOr,
	"AS":         TokenAs,
	"OPTION":     TokenOption,
	"ALL":        TokenAll,
	"DISTINCT":   TokenDistinct,
	"PRINT":      TokenPrint,
	"THROW":      TokenThrow,
	"ALTER":      TokenAlter,
	"TABLE":      TokenTable,
	"DROP":       TokenDrop,
	"INDEX":      TokenIndex,
	"REVERT":     TokenRevert,
	"WITH":       TokenWith,
	"COOKIE":     TokenCookie,
	"DATABASE":   TokenDatabase,
	"SCOPED":     TokenScoped,
	"CREDENTIAL": TokenCredential,
	"TOP":        TokenTop,
	"PERCENT":    TokenPercent,
	"TIES":       TokenTies,
	"INTO":       TokenInto,
	"GROUP":      TokenGroup,
	"BY":         TokenBy,
	"HAVING":     TokenHaving,
	"ORDER":      TokenOrder,
	"ASC":        TokenAsc,
	"DESC":       TokenDesc,
	"UNION":      TokenUnion,
	"EXCEPT":     TokenExcept,
	"INTERSECT":  TokenIntersect,
	"CROSS":      TokenCross,
	"JOIN":       TokenJoin,
	"INNER":      TokenInner,
	"LEFT":       TokenLeft,
	"RIGHT":      TokenRight,
	"FULL":       TokenFull,
	"OUTER":      TokenOuter,
	"ON":         TokenOn,
	"ROLLUP":     TokenRollup,
	"CUBE":       TokenCube,
	"NOT":        TokenNot,
}

func lookupKeyword(ident string) TokenType {
	if tok, ok := keywords[strings.ToUpper(ident)]; ok {
		return tok
	}
	return TokenIdent
}
