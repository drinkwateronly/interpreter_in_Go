package object

import (
	"Monkey_1/ast"
	"bytes"
	"fmt"
	"hash/fnv"
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
	HASH_OBJ         = "HASH"
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

type HashKey struct {
	//Pairs map[Object]Object
	Type  ObjectType
	Value uint64
}

func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// todo: 通过缓存HashKey()⽅法的返回值来优化其性能
func (s *String) HashKey() HashKey {
	h := fnv.New64a() // 可能出现哈希碰撞，todo: 单链法和开放式寻址法
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }

func (h *Hash) Inspect() string {
	var out bytes.Buffer
	var pairs []string
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

type Hashable interface {
	HashKey() HashKey
}
