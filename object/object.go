package object

import "fmt"

type ObjectType string

const (
	INTEGER_OBJ      = "INTEGER"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
)

// 每个值有不同表现形式，因此使用接口会比使用多个字段的结构体简洁
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

func (rv *ReturnValue) Inspect() string { return rv.Inspect() }

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

// Error #################################################
type Error struct {
	Message string // 只返回了错误信息，无法返回行号列号
}

func (e *Error) Inspect() string { return "ERROR: " + e.Message }

func (e *Error) Type() ObjectType { return ERROR_OBJ }
