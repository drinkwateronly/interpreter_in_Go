package ast

import (
	"Monkey_1/token"
	"bytes"
	"strings"
)

// Node 结点接口, 都需要实现 TokenLiteral() 方法，返回其关联的词法单元token的字面量literal
type Node interface {
	TokenLiteral() string // 仅用于测试
	String() string
}

type Statement interface {
	Node
	statementNode() // 实际无作用，仅用于区分Expression与Statement
}

type Expression interface {
	Node
	expressionNode() // 实际无作用，仅用于区分Expression与Statement
}

// Program ------------------------------------------
// Program 是 Node 接口的实现
type Program struct {
	Statements []Statement // 存放的是AST的根节点。注意，每一条语句可以构成一AST，所以是切片
}

// TokenLiteral 返回第一个AST的根节点的TokenLiteral()，仅用于只有一个Statement的情况
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return "" // 这表示AST并没有根节点，AST是空的，即源码是空的。
	}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// LetStatement ------------------------------------------
// LetStatement 是 Statement 接口的实现
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier // 保存绑定的标识符
	Value Expression  // 保存产生值的表达式/或者值本身，没有*的原因是，Expression是接口
}

func (ls *LetStatement) statementNode() {}

func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.Value) // = ls.Name.string()
	out.WriteString("=")
	if ls.Value != nil { // ?
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String() // `let <ls.Name.Value> = <ls.Value.String()>;`
}

// Identifier --------------------------------------
// Identifier 是 Expression 接口的实现
type Identifier struct {
	Token token.Token // the token.IDNET token
	Value string      // 和 Token.Literal 一样 是Identifier的命名
}

func (i *Identifier) expressionNode() {}

func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

func (i *Identifier) String() string { return i.Value }

// ReturnStatement --------------------------------------
// ReturnStatement 是 Statement 接口的实现
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String() // `return <rs.ReturnValue.String()>;`
}

// ExpressionStatement --------------------------------------
// ExpressionStatement 是 Statement 接口的实现
// 代表的是表达式语句，虽然Monkey中表达式不是语句，但被具体实现时封装成了语句（从字段即可看出）
type ExpressionStatement struct {
	Token      token.Token
	Expression Expression // 字段是实现Expression接口的ast节点（通俗来讲是表达式节点）
}

func (es *ExpressionStatement) statementNode() {}

func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() //
	}
	return ""
}

// IntegerLiteral --------------------------------------
// IntegerLiteral 是 Expression 接口的实现，是AST中的一个节点
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

func (il *IntegerLiteral) String() string { return il.Token.Literal }

// PrefixExpression --------------------------------------
// PrefixExpression 是 Expression 接口的实现，是AST中的一个节点
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression --------------------------------------
// InfixExpression 是 Expression 接口的实现，是AST中的一个节点
type InfixExpression struct {
	Token    token.Token // 运算符的词法单元 如 + - * /
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(ie.Operator)
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// Boolean --------------------------------------
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {}

func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

func (b *Boolean) String() string { return b.Token.Literal }

// IfExpression --------------------------------------
type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}

func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// BlockStatement --------------------------------------
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

func (bs *BlockStatement) String() string {
	out := bytes.Buffer{}
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// FunctionLiteral --------------------------------------
// 定义函数， 函数可以作为
type FunctionLiteral struct {
	Token      token.Token     // Token.TokenType = FUNCTION, Token.Literal = 函数名
	Parameters []*Identifier   // 参数列表，是标识符
	Body       *BlockStatement // 块语句
}

func (fl *FunctionLiteral) expressionNode() {}

func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

func (fl *FunctionLiteral) String() string {
	out := bytes.Buffer{}
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CallExpression --------------------------------------
// 调用函数
type CallExpression struct {
	Token     token.Token  // '('
	Function  Expression   // 函数的Identifier 节点
	Arguments []Expression // 参数列表
}

func (ce *CallExpression) expressionNode() {}

func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

func (ce *CallExpression) String() string {
	var out bytes.Buffer
	//args := []string{}
	// string is the set of all strings of 8-bit bytes, conventionally but not necessarily representing UTF-8-encoded text. A string may be empty, but not nil. Values of string type are immutable.
	var args []string
	for _, arg := range ce.Arguments {
		args = append(args, arg.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// StringLiteral --------------------------------------
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}

func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

func (sl *StringLiteral) String() string { return sl.Value }

// ArrayLiteral -------
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}

func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("[")
	var elements []string
	for _, element := range al.Elements {
		elements = append(elements, element.String())
	}
	out.WriteString(strings.Join(elements, ","))
	out.WriteString("]")
	return out.String()
}

// IndexExpression -------
type IndexExpression struct {
	Token           token.Token
	ArrayIdentifier Expression
	Index           Expression
}

func (ie *IndexExpression) expressionNode() {}

func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(") // ?
	out.WriteString(ie.ArrayIdentifier.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("]")
	out.WriteString(")")
	return out.String()
}

// HashLiteral
type HashLiteral struct {
	Token token.Token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode() {}

func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }

func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	var pairs []string
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ","))
	out.WriteString("}")
	return out.String()
}
