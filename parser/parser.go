package parser

import (
	"Monkey_1/ast"
	"Monkey_1/lexer"
	"Monkey_1/token"
	"fmt"
	"strconv"
)

// 优先级
const (
	_int = iota // 空白标识符
	LOWEST
	EQUALS      // ==
	LESSGREATER // < or >
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// token与优先级映射
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	// 根据当前curToken.Type确定其优先级
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(expression ast.Expression) ast.Expression
)

type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token // 当前的token
	peekToken token.Token // 后一个token，辅助决策

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// 初始化，移动两次，cur为第0个token，peek为第1个token
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// 注册
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	// 注册前缀表达式
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	// 注册中缀表达式
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

// 移动curToken与peekToken
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram 对整个源代码进行解析，遍历所有的词法单元
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{} // 存放所有的ast节点

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement() // 每次解析一条表达式
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// parseStatement 对单个表达式进行解析
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()

	}
}

// parseLetStatement 解析let开头的语句
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	// let <标识符> = <表达式>
	// 期待下一个词法单元是标识符
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	// 创建一个ast的标识符节点
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	// 期待下一个词法单元是赋值=
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	// 循环剩余的节点，直到是分号;。目前来看可能有问题
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	// 返回&ast.LetStatement的ast节点
	return stmt
}

// parseReturnStatement 解析return开头的语句
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// 小工具1，判断当前token的token.TokenType
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// 小工具2，判断下一个token的token.TokenType
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg) // 记录error
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) { // 只有后一个的类型正确，才会前移词法单元
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// 在parseExpressionStatement中，最低的优先级会传递给parseExpression。
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpression 解析表达式
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// precedence 是前一个token的优先级

	// 首先匹配curToken.Type对应的 前缀解析函数
	prefix := p.prefixParseFns[p.curToken.Type]
	// 没有匹配到前缀
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// 匹配到了 前缀，当成 左表达式
	leftExp := prefix()

	// 直到下一个token是分号，或者下一个token优先级更高
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// 匹配到 peekToken.Type对应的 中缀解析函数
		infix := p.infixParseFns[p.peekToken.Type]
		// 没有匹配到
		if infix == nil {
			return leftExp
		}
		// 匹配到了，移动token
		p.nextToken()
		// 解析中缀表达式，当成 左表达式，同时前面匹配到的 前缀 就不是左表达式
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier 解析 标识符 表达式
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 解析 int 表达式
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parsePrefixExpression 解析前缀表达式如 -1 和 !X
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	// 传入的是左表达式
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// 为了提供更详细的 解析函数 未匹配信息
func (p *Parser) noPrefixParseFnError(tokenType token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", tokenType)
	p.errors = append(p.errors, msg)
}
