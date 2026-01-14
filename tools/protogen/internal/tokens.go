package protogen

type TokenType string

const (
	TokenKeyword  TokenType = "Keyword"
	TokenIdent    TokenType = "Ident"
	TokenString   TokenType = "String"
	TokenNumber   TokenType = "Number"
	TokenLBrace   TokenType = "LBrace"
	TokenRBrace   TokenType = "RBrace"
	TokenLParen   TokenType = "LParen"
	TokenRParen   TokenType = "RParen"
	TokenLBracket TokenType = "LBracket"
	TokenRBracket TokenType = "RBracket"
	TokenEqual    TokenType = "Equal"
	TokenColon    TokenType = "Colon"
	TokenComma    TokenType = "Comma"
	TokenAt       TokenType = "At"
	TokenOptional TokenType = "Optional"
	TokenEOF      TokenType = "EOF"
	TokenPath     TokenType = "Path"
	TokenIllegal  TokenType = "Illegal"
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}
