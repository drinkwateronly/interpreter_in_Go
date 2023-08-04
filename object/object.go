package object

import (
	"Monkey_1/ast"
	"bytes"
	"fmt"
	"strings"
)

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	STRING_OBJ       = "STRING"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
)

// 每个值有不同表现形式，因此使用 Object 接口会比使用多个字段的结构体简洁
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Integer ##############################################
type Integer struct {
	Value int64 // 使用宿主语言的原生类型
}

func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// Boolean ##############################################
type Boolean struct {
	Value bool // 使用宿主语言的原生类型
}

func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

// Null #################################################
type Null struct{}

func (n *Null) Inspect() string { return "nil" }

func (n *Null) Type() ObjectType { return NULL_OBJ }

// ReturnValue #################################################
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

// Error #################################################
type Error struct {
	Message string // 只返回了错误信息，无法返回行号列号
}

func (e *Error) Inspect() string { return "ERROR: " + e.Message }

func (e *Error) Type() ObjectType { return ERROR_OBJ }

// Function #################################################
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment // 目前理解为作用域
}

func (f *Function) Inspect() string {
	// 返回函数的字面值
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// String #################################################
type String struct {
	Value string // 只返回了错误信息，无法返回行号列号
}

func (s *String) Inspect() string { return s.Value }

func (s *String) Type() ObjectType { return STRING_OBJ }

// ------
type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }

func (b *Builtin) Inspect() string { return "builtin function" }

// ArrayLiteral
type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }

func (a *Array) Inspect() string {
	var out bytes.Buffer
	var elements []string
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ","))
	out.WriteString("]")
	return out.String()
}
