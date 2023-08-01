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
)

// 每个值有不同表现形式，因此使用接口会比使用多个字段的结构体简洁
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Environment ##############################################
type Environment struct {
	store map[string]Object
	outer *Environment
}

// 为什么不直接使⽤map，⽽是使⽤封装

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, obj Object) Object {
	e.store[name] = obj
	return obj
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

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

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
