package parser

import (
	"Monkey_1/ast"
	"Monkey_1/lexer"
	"Monkey_1/token"
	"fmt"
)

type Parser struct {
	l         *lexer.Lexer
	errors    []string
	curToken  token.Token // 当前的token
	peekToken token.Token // 后一个token，辅助决策
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// 初始化，移动两次，cur为第0个token，peek为第1个token
	p.nextToken()
	p.nextToken()
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
		return nil
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

//
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
