package token

type TokenType string // 虽然string并不及int或type高效，但可读性强

type Token struct {
	Type    TokenType // token的类型
	Literal string    // token的字面值
}

// TokenType constant
const (
	ILLEGAL = "ILLEGAL"
	EOF     = ""

	//  Identifiers + literals
	IDENT = "IDENT" // add, x, y
	INT   = "INT"   // 1 2 3

	// 运算符 Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="

	// 分隔符 Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"

	// 关键词 keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	STRING   = "STRING"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
}

// LookupIdent 查找keywords表格，用于区分用户自定标识符和Monkey关键字，因为他们都是字符串形式的
func LookupIdent(ident string) TokenType {
	// 如果标识符在关键字里，返回对应tokenType
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	// 如果不在关键字里，表明是用户自定标识符
	return IDENT
}
