package lexer

import "Monkey_1/token"

type Lexer struct {
	input        string
	position     int  // 所输入字符串中的当前位置 （指向当前字符）
	readPosition int  // 所输入字符串中的当前 读取 位置 （指向当前字符 的 后一个字符）
	ch           byte // 当前正在查看的字符本身
}

// New 根据input的source code创建一个语法分析器
func New(input string) *Lexer {
	l := &Lexer{input: input}
	// 初始化l中的position、readPosition，分别为0和1
	l.readChar()
	return l
}

// readChar 每次调用时读取Lexer.input的当前字符 并将position & readPosition后移
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// peekChar 与 readChar 类似，但只窥探后面一个字符串，不移动position & readPosition
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

// newToken 根据tokenType和ch新建token，仅用于token的长度为1个字符的情况
func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// NextToken 每次调用时返回当前的token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	// 跳过空字符串，包括\n，直到l.ch为非空字符串
	l.skipWhitespace()

	// 根据当前l.ch，返回对应的token
	switch l.ch {
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
		// case '=' 和 case '!'相似性略高，若有更多的双字符token，可以考虑合并成一个函数 makeTwoCharToken
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch) // “==”
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			// 进入此，表示开头已经是字母字符了
			tok.Literal = l.readIdentifier()
			// 是字符串，并进一步区分是用户自定标识符还是关键词
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigital(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	// 检查后，字符指针移动
	l.readChar()
	return tok
}

// readIdentifier 读取标识符直到遇见非字母字符
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigital(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position] // 即position:l.position-1之间的字符串就是标识符
}

// readIdentifier 读取数字直到遇见非数字字符
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigital(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position] // 即position:l.position-1之间的字符串就是标识符
}

// skipWhitespace 跳过空白的字符，包括换行符
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isDigital(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
