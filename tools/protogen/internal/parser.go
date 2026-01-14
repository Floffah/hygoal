package protogen

import (
	"errors"
	"fmt"
)

// Parser holds the state for parsing
type Parser struct {
	lexer  *Lexer
	curTok Token
}

func NewParser(input string) *Parser {
	lex := NewLexer(input)
	return &Parser{lexer: lex, curTok: lex.NextToken()}
}

func (p *Parser) next() Token {
	p.curTok = p.lexer.NextToken()
	return p.curTok
}

func (p *Parser) current() Token {
	return p.curTok
}

func (p *Parser) expect(typ TokenType) bool {
	return p.curTok.Type == typ
}

func (p *Parser) getError(message string) error {
	return &ParserError{
		Message: message,
		Line:    p.curTok.Line,
		Col:     p.curTok.Col - len(p.curTok.Value),
	}
}

func (p *Parser) getErrorf(format string, args ...interface{}) error {
	return &ParserError{
		Message: fmt.Sprintf(format, args...),
		Line:    p.curTok.Line,
		Col:     p.curTok.Col - len(p.curTok.Value),
	}
}

type ParserError struct {
	Message string
	Line    int
	Col     int
}

func (e *ParserError) Error() string {
	return e.Message
}

func FormatParseError(err error, fileName string) string {
	var parseErr *ParserError
	if errors.As(err, &parseErr) {
		return fmt.Sprintf("%s:%d:%d: %s\n", fileName, parseErr.Line, parseErr.Col, parseErr.Message)
	}

	return err.Error()
}
