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
	_int        = iota // 空白标识符
	LOWEST             // 最低，任何表达式的开始都是最低优先级
	EQUALS             // ==
	LESSGREATER        // < or >
	SUM                // +
	PRODUCT            // *
	PREFIX             // -X or !X
	CALL               // myFunction(X)
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
	token.LPAREN:   CALL,
}

type (
	// prefixParseFn或infixParseFn都会遵循: 函数在开始解析表达式时，curToken必须是所关联的词法单元类型，
	// 返回分析的表达式结果时，curToken是当前表达式类型中的最后⼀个词法单元。
	prefixParseFn func() ast.Expression                          // 前缀表达式解析函数类型
	infixParseFn  func(expression ast.Expression) ast.Expression // 中缀表达式解析函数类型
)

// Parser 语法解析器
type Parser struct {
	l         *lexer.Lexer // 词法解析器，依靠词法解析器可以检查并移动当前和后一个token
	errors    []string     // 记录解析出现的所有的错误，不会因为出现一个错误停止解析
	curToken  token.Token  // 当前的token
	peekToken token.Token  // 后一个token，辅助决策

	prefixParseFns map[token.TokenType]prefixParseFn // 记录不同tokenType对应的前缀表达式解析函数
	infixParseFns  map[token.TokenType]infixParseFn  // 记录不同tokenType对应的中缀表达式解析函数
}

// 向 Parser 注册某个tokenType的前缀表达式
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// 向 Parser 注册某个tokenType的中缀表达式
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// New 创建语法解析器Parser，参数为词法解析器Lexer
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{}, // 记录所有错误
	}
	// 初始化，移动两次，cur为第0个token，peek为第1个token
	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// 注册前缀表达式
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)

	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	// 注册中缀表达式，都是infixExpression
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	p.registerInfix(token.LPAREN, p.parseCallExpression)
	return p
}

// ParseProgram 对整个源代码进行解析的入口，遍历所有的词法单元，返回*ast.Program，是源代码的ast树根节点，内部存了所有的AST
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{} // 存放所有的ast节点

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement() // 每次解析一条语句
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
	// 只有两个Statement语句，剩下的是表达式语句ExpressionStatement，实际上就是表达式
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
	// 创建一个ast的LetStatement节点
	letStmt := &ast.LetStatement{Token: p.curToken}
	// let <标识符> = <表达式>
	// 期待下一个token的类型是token.IDENT，会移动token，否则记录错误，不移动token
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	// 创建一个ast的标识符Identifier节点
	letStmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	// 期待下一个token是赋值=
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	// 解析表达式
	letStmt.Value = p.parseExpression(LOWEST)
	// 循环剩余的token，直到是分号，这是因为parseExpression不移动token。
	for p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	// 返回&ast.LetStatement的ast节点
	return letStmt
}

// parseReturnStatement 解析return开头的语句
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	// 创建一个ast的 ReturnStatement 节点
	returnStmt := &ast.ReturnStatement{Token: p.curToken}
	// 略过p.curToken，即return token
	p.nextToken()
	// 解析表达式
	returnStmt.ReturnValue = p.parseExpression(LOWEST)
	// 循环剩余的token，直到是分号，这是因为parseExpression不移动token。
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return returnStmt
}

// parseExpressionStatement 解析表达式语句，即除了Let和Return后的语句
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	//defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	// parseExpressionStatement的工作完全交给 parseExpression
	stmt.Expression = p.parseExpression(LOWEST)
	// 表达式语句需要在有无分号的情况下都有效
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// parseExpression 解析表达式，实际的作用是递归构造ast节点
func (p *Parser) parseExpression(precedence int) ast.Expression {
	// precedence 是前一个token的优先级
	//defer untrace(trace("parseExpression"))
	// 【1】首先匹配curToken.Type对应的 前缀解析函数
	prefixFn := p.prefixParseFns[p.curToken.Type]
	// 没有匹配到前缀解析函数，也就是没有 IDENT, INT, !, -等作为语句的开头，此时应该是抛出错误
	if prefixFn == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// 匹配到了前缀解析函数，调用并解析，解析出的节点当成左表达式
	leftExp := prefixFn()

	// 直到下一个token是分号，或者下一个token优先级更高，因此无法直接计算下一个token
	// 下⼀个运算符或词法单元的左约束能⼒是强于当前token的右约束能⼒
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// 【2】然后，匹配 peekToken.Type对应的 中缀解析函数
		infix := p.infixParseFns[p.peekToken.Type]
		// 没有匹配到中缀解析函数，则返回匹配到的lefxExp，作为左表达式的前缀
		if infix == nil {
			return leftExp
		}
		// 匹配到了，移动token
		p.nextToken()
		// 解析中缀表达式，当成 左表达式，同时前面匹配到的 前缀被覆盖
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parsePrefixExpression 解析前缀表达式如 -1 和 !X
func (p *Parser) parsePrefixExpression() ast.Expression {
	//defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	// 移动token，因为curToken是 - 或者 !
	p.nextToken()
	// 从下一个token开始解析，解析出的表达式作为右表达式
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

// parseInfixExpression 解析中缀表达式
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	//defer untrace(trace("parseInfixExpression"))
	// 传入的是左表达式，并初始化
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	// 保留当前解析的token的优先级，因为接下来要移动token
	precedence := p.curPrecedence()
	// 移动token
	p.nextToken()

	// 解析下一个token，其表达式作为右表达式
	expression.Right = p.parseExpression(precedence)

	return expression
}

// ------------------------------------------------------------------------------
// parseIdentifier 解析 标识符 表达式
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral 解析 int 表达式
func (p *Parser) parseIntegerLiteral() ast.Expression {
	//defer untrace(trace("parseIntegerLiteral"))

	lit := &ast.IntegerLiteral{Token: p.curToken}
	// strconv包：字符串和数值类型的相互转换
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	// 实际值
	lit.Value = value
	return lit
}

// parseBoolean 解析 bool
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseGroupedExpression 解析 带括号的分组表达式
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil // 如果下一个不是{
	}

	// expectPeek内部已经调用了一次nextToken()
	p.nextToken()
	// 此时curToken指向了(的后一个Token

	// 解析出条件语句
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil // 如果解析完后，下一个不是),即无法闭环
	}
	if !p.expectPeek(token.LBRACE) {
		return nil // )的下一个不是{,即没有进入Consequence语句
	}
	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken() // 是else
		if !p.expectPeek(token.LBRACE) {
			return nil // 但else后没跟{
		}
		expression.Alternative = p.parseBlockStatement()

	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	// 不 断 调 ⽤ parseStatement ， 直 到 遇 ⻅ 右 ⼤ 括 号 } 或 token.EOF
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}
	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// 参数列表为空
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}
	p.nextToken()
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}

// ----------------------------- tools -----------------------------
// 移动curToken与peekToken
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs 判断当前token的token.TokenType
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs 判断下一个token的token.TokenType
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek 判断下一个token是否是期待的token，如果不是会记录错误，是检查语法错误的一部分。
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) { // 后一个token的类型对应上，前移词法单元
		p.nextToken()
		return true
	} else {
		p.peekError(t) // 添加错误
		return false
	}
}

// peekPrecedence 查看下一个token的优先级
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	// 如果优先级表没有记录，那他就是最低的优先级
	return LOWEST
}

// curPrecedence 查看curToken的优先级
func (p *Parser) curPrecedence() int {
	// 根据当前curToken.Type确定其优先级
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// ---------------- errors -------------------

// Errors 返回Parser记录的Errors
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError 当期待的词法单元没匹配上，为Parser记录错误
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg) // 记录error
}

// noPrefixParseFnError 为了提供更详细的 解析函数 未匹配信息
func (p *Parser) noPrefixParseFnError(tokenType token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", tokenType)
	p.errors = append(p.errors, msg)
}
