package protogen

import "unicode"

type Lexer struct {
	input        string
	position     int
	readPosition int
	Line         int
	Col          int
	ch           rune
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, Line: 1, Col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.readPosition])
	}

	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.Line++
		l.Col = 0
	} else {
		l.Col++
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	switch l.ch {
	case '{':
		l.readChar()
		return Token{Type: TokenLBrace, Value: "{", Line: l.Line, Col: l.Col}
	case '}':
		l.readChar()
		return Token{Type: TokenRBrace, Value: "}", Line: l.Line, Col: l.Col}
	case '(':
		l.readChar()
		return Token{Type: TokenLParen, Value: "(", Line: l.Line, Col: l.Col}
	case ')':
		l.readChar()
		return Token{Type: TokenRParen, Value: ")", Line: l.Line, Col: l.Col}
	case '[':
		l.readChar()
		return Token{Type: TokenLBracket, Value: "[", Line: l.Line, Col: l.Col}
	case ']':
		l.readChar()
		return Token{Type: TokenRBracket, Value: "]", Line: l.Line, Col: l.Col}
	case '=':
		l.readChar()
		return Token{Type: TokenEqual, Value: "=", Line: l.Line, Col: l.Col}
	case ':':
		l.readChar()
		return Token{Type: TokenColon, Value: ":", Line: l.Line, Col: l.Col}
	case ',':
		l.readChar()
		return Token{Type: TokenComma, Value: ",", Line: l.Line, Col: l.Col}
	case '@':
		l.readChar()
		return Token{Type: TokenAt, Value: "@", Line: l.Line, Col: l.Col}
	case '?':
		l.readChar()
		return Token{Type: TokenOptional, Value: "?", Line: l.Line, Col: l.Col}
	case '"', '\'':
		return l.readString()
	case 0:
		return Token{Type: TokenEOF, Value: "", Line: l.Line, Col: l.Col}
	default:
		if l.ch == '/' {
			path := l.readPath()
			return Token{Type: TokenPath, Value: path, Line: l.Line, Col: l.Col}
		} else if isLetter(l.ch) {
			ident := l.readIdentifier()
			if isKeyword(ident) {
				return Token{Type: TokenKeyword, Value: ident, Line: l.Line, Col: l.Col}
			}
			return Token{Type: TokenIdent, Value: ident, Line: l.Line, Col: l.Col}
		} else if unicode.IsDigit(l.ch) {
			num := l.readNumber()
			return Token{Type: TokenNumber, Value: num, Line: l.Line, Col: l.Col}
		}
		illegal := l.ch
		l.readChar()
		return Token{Type: TokenIllegal, Value: string(illegal)}
	}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	pos := l.position
	for isLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '.' || l.ch == '_' || l.ch == '/' {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) readNumber() string {
	pos := l.position
	for unicode.IsDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) readString() Token {
	quote := l.ch
	l.readChar() // skip opening quote
	pos := l.position
	for l.ch != quote && l.ch != 0 {
		l.readChar()
	}
	str := l.input[pos:l.position]
	l.readChar() // skip closing quote
	return Token{Type: TokenString, Value: str, Line: l.Line, Col: l.Col}
}

func (l *Lexer) readPath() string {
	pos := l.position
	for l.ch == '/' || isLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '-' || l.ch == '_' || l.ch == ':' {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == '.' || ch == '/'
}

func isKeyword(ident string) bool {
	// Only treat truly reserved words as keywords, e.g. 'enum' or 'packet'.
	reserved := []string{"enum", "packet", "type"}
	for _, k := range reserved {
		if ident == k {
			return true
		}
	}
	return false
}

func (l *Lexer) Lex() []Token {
	var tokens []Token
	for tok := l.NextToken(); tok.Type != TokenEOF; tok = l.NextToken() {
		tokens = append(tokens, tok)
	}
	return tokens
}
